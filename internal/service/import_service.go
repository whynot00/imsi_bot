// Package service contains business logic that orchestrates repositories
// and the parser layer.
package service

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/whynot00/imsi-bot/internal/parser"
	"github.com/whynot00/imsi-bot/internal/repo"
)

const (
	locationTolerance = 1 * time.Hour
	flushSize         = 1000  // rows per batch INSERT
	txChunkSize       = 10000 // rows per transaction commit
)

// Progress reports live import state to the caller.
type Progress struct {
	Phase     string `json:"phase"` // "locations", "devices", "done"
	Devices   int    `json:"devices"`
	Locations int    `json:"locations"`
	Sightings int    `json:"sightings"`
	Skipped   int    `json:"skipped"`
}

// ProgressFunc is called periodically during import with current progress.
type ProgressFunc func(Progress)

// ImportResult summarises what was written to the database.
type ImportResult struct {
	Devices   int `json:"devices"`
	Locations int `json:"locations"`
	Sightings int `json:"sightings"`
	Skipped   int `json:"skipped"`
}

// ImportService orchestrates parsing + persistence for both file kinds.
type ImportService struct {
	db       *sql.DB
	devices  *repo.DeviceRepo
	parametr *repo.ParametrRepo
	rk       *repo.RKRepo
}

func NewImportService(
	db *sql.DB,
	devices *repo.DeviceRepo,
	parametr *repo.ParametrRepo,
	rk *repo.RKRepo,
) *ImportService {
	return &ImportService{db: db, devices: devices, parametr: parametr, rk: rk}
}

// ---------------------------------------------------------------------------
// Parametr import — 2-pass streaming
// ---------------------------------------------------------------------------

func (s *ImportService) ImportParametrFromCSV(ctx context.Context, data []byte, onProgress ProgressFunc) (*ImportResult, error) {
	out := &ImportResult{}

	report := func(phase string) {
		if onProgress != nil {
			onProgress(Progress{
				Phase:     phase,
				Devices:   out.Devices,
				Locations: out.Locations,
				Sightings: out.Sightings,
				Skipped:   out.Skipped,
			})
		}
	}

	// --- Pass 1: collect unique locations ---
	type locKey struct {
		SeenAt time.Time
		Lat    float64
		Lon    float64
	}
	locSeen := make(map[locKey]struct{})
	var locRows []struct {
		SeenAt time.Time
		Lat    float64
		Lon    float64
	}

	err := parser.ParseStream(bytes.NewReader(data), func(_ *parser.Device, l *parser.Location) error {
		if l == nil {
			return nil
		}
		lat, err := strconv.ParseFloat(l.Lat, 64)
		if err != nil {
			return nil
		}
		lon, err := strconv.ParseFloat(l.Lon, 64)
		if err != nil {
			return nil
		}
		seenAt, err := parseTime(l.Date)
		if err != nil {
			return nil
		}
		k := locKey{SeenAt: seenAt, Lat: lat, Lon: lon}
		if _, dup := locSeen[k]; dup {
			return nil
		}
		locSeen[k] = struct{}{}
		locRows = append(locRows, struct {
			SeenAt time.Time
			Lat    float64
			Lon    float64
		}{seenAt, lat, lon})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("pass 1: %w", err)
	}
	locSeen = nil // free memory
	out.Locations = len(locRows)
	report("locations")
	log.Printf("[parametr] pass 1 done: %d unique locations collected", len(locRows))

	// --- Insert locations in a separate transaction ---
	var insertedLocs []repo.LocationInserted
	err = repo.TxFunc(ctx, s.db, func(tx *sql.Tx) error {
		var err error
		insertedLocs, err = repo.BatchInsertLocations(ctx, tx, locRows)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("insert locations: %w", err)
	}
	out.Locations = len(insertedLocs)
	locRows = nil // free memory
	log.Printf("[parametr] %d locations inserted", out.Locations)

	// Sort for binary search
	sort.Slice(insertedLocs, func(i, j int) bool {
		return insertedLocs[i].SeenAt.Before(insertedLocs[j].SeenAt)
	})

	// --- Pass 2: devices + sightings in chunked transactions ---
	deviceCache := make(map[string]int64) // persists across tx chunks

	type sightingWithKey struct {
		repo.SightingParametrRow
		DeviceKey string
	}

	var tx *sql.Tx
	var pendingDevices [][2]string
	pendingDeviceKeys := make(map[string]struct{})
	var pendingSightings []sightingWithKey
	pendingSightingKeys := make(map[string]struct{})
	rowsInTx := 0

	beginTx := func() error {
		var err error
		tx, err = s.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}
		rowsInTx = 0
		return nil
	}

	flushBatch := func() error {
		// 1. Upsert devices
		if len(pendingDevices) > 0 {
			ids, err := repo.BatchUpsertDevices(ctx, tx, pendingDevices)
			if err != nil {
				return err
			}
			for k, id := range ids {
				deviceCache[k] = id
			}
			pendingDevices = pendingDevices[:0]
			pendingDeviceKeys = make(map[string]struct{})
		}
		// 2. Resolve device IDs and insert sightings
		if len(pendingSightings) > 0 {
			rows := make([]repo.SightingParametrRow, len(pendingSightings))
			for i, s := range pendingSightings {
				rows[i] = s.SightingParametrRow
				rows[i].DeviceID = deviceCache[s.DeviceKey]
			}
			n, err := repo.BatchInsertSightingsParametr(ctx, tx, rows)
			if err != nil {
				return err
			}
			out.Sightings += n
			pendingSightings = pendingSightings[:0]
			pendingSightingKeys = make(map[string]struct{})
		}
		return nil
	}

	commitAndReopen := func() error {
		if err := flushBatch(); err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit: %w", err)
		}
		log.Printf("[parametr] pass 2 progress: %d devices, %d sightings committed", out.Devices, out.Sightings)
		report("devices")
		return beginTx()
	}

	if err := beginTx(); err != nil {
		return nil, err
	}

	err = parser.ParseStream(bytes.NewReader(data), func(d *parser.Device, _ *parser.Location) error {
		if d == nil {
			return nil
		}
		seenAt, err := parseTime(d.Date)
		if err != nil {
			out.Skipped++
			return nil
		}

		deviceKey := d.IMSI + "|" + d.IMEI
		sightingKey := deviceKey + "|" + seenAt.String()

		if _, dup := pendingSightingKeys[sightingKey]; dup {
			out.Skipped++
			return nil
		}

		if _, cached := deviceCache[deviceKey]; !cached {
			if _, pending := pendingDeviceKeys[deviceKey]; !pending {
				pendingDevices = append(pendingDevices, [2]string{d.IMSI, d.IMEI})
				pendingDeviceKeys[deviceKey] = struct{}{}
				out.Devices++
			}
		}

		locID := nearestLocation(insertedLocs, seenAt, locationTolerance)

		pendingSightings = append(pendingSightings, sightingWithKey{
			SightingParametrRow: repo.SightingParametrRow{
				SeenAt:     seenAt,
				Standart:   d.Standart,
				Operator:   d.Operator,
				Event:      d.Event,
				LocationID: locID,
			},
			DeviceKey: deviceKey,
		})
		pendingSightingKeys[sightingKey] = struct{}{}
		rowsInTx++

		if len(pendingSightings) >= flushSize {
			if err := flushBatch(); err != nil {
				return err
			}
			report("devices")
		}
		if rowsInTx >= txChunkSize {
			return commitAndReopen()
		}
		return nil
	})
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("pass 2: %w", err)
	}

	// flush + commit remaining
	if err := flushBatch(); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("final commit: %w", err)
	}

	log.Printf("[parametr] done: devices=%d locations=%d sightings=%d skipped=%d",
		out.Devices, out.Locations, out.Sightings, out.Skipped)
	return out, nil
}

// ---------------------------------------------------------------------------
// RK import — single-pass streaming with chunked transactions
// ---------------------------------------------------------------------------

func (s *ImportService) ImportRKFromCSV(ctx context.Context, data []byte, onProgress ProgressFunc) (*ImportResult, error) {
	out := &ImportResult{}

	report := func(phase string) {
		if onProgress != nil {
			onProgress(Progress{
				Phase:     phase,
				Devices:   out.Devices,
				Locations: out.Locations,
				Sightings: out.Sightings,
				Skipped:   out.Skipped,
			})
		}
	}

	deviceCache := make(map[string]int64)

	type sightingWithKey struct {
		repo.SightingRKRow
		DeviceKey string
	}

	var tx *sql.Tx
	var pendingDevices [][2]string
	pendingDeviceKeys := make(map[string]struct{})
	var pendingSightings []sightingWithKey
	pendingSightingKeys := make(map[string]struct{})
	rowsInTx := 0

	beginTx := func() error {
		var err error
		tx, err = s.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}
		rowsInTx = 0
		return nil
	}

	flushBatch := func() error {
		if len(pendingDevices) > 0 {
			ids, err := repo.BatchUpsertDevices(ctx, tx, pendingDevices)
			if err != nil {
				return err
			}
			for k, id := range ids {
				deviceCache[k] = id
			}
			pendingDevices = pendingDevices[:0]
			pendingDeviceKeys = make(map[string]struct{})
		}
		if len(pendingSightings) > 0 {
			rows := make([]repo.SightingRKRow, len(pendingSightings))
			for i, s := range pendingSightings {
				rows[i] = s.SightingRKRow
				rows[i].DeviceID = deviceCache[s.DeviceKey]
			}
			n, err := repo.BatchInsertSightingsRK(ctx, tx, rows)
			if err != nil {
				return err
			}
			out.Sightings += n
			pendingSightings = pendingSightings[:0]
			pendingSightingKeys = make(map[string]struct{})
		}
		return nil
	}

	commitAndReopen := func() error {
		if err := flushBatch(); err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit: %w", err)
		}
		log.Printf("[rk] progress: %d devices, %d sightings committed", out.Devices, out.Sightings)
		report("devices")
		return beginTx()
	}

	if err := beginTx(); err != nil {
		return nil, err
	}

	err := parser.ParseRawStream(bytes.NewReader(data), func(d parser.RawDevice) error {
		seenAt, err := parseTime(d.Date)
		if err != nil {
			out.Skipped++
			return nil
		}
		lat, err := strconv.ParseFloat(d.Lat, 64)
		if err != nil {
			out.Skipped++
			return nil
		}
		lon, err := strconv.ParseFloat(d.Lon, 64)
		if err != nil {
			out.Skipped++
			return nil
		}
		signal, _ := strconv.Atoi(d.SignalStrength)

		deviceKey := d.IMSI + "|" + d.IMEI
		sightingKey := deviceKey + "|" + seenAt.String()

		if _, dup := pendingSightingKeys[sightingKey]; dup {
			out.Skipped++
			return nil
		}

		if _, cached := deviceCache[deviceKey]; !cached {
			if _, pending := pendingDeviceKeys[deviceKey]; !pending {
				pendingDevices = append(pendingDevices, [2]string{d.IMSI, d.IMEI})
				pendingDeviceKeys[deviceKey] = struct{}{}
				out.Devices++
			}
		}

		pendingSightings = append(pendingSightings, sightingWithKey{
			SightingRKRow: repo.SightingRKRow{
				SeenAt:   seenAt,
				Standart: d.Standart,
				Lat:      lat,
				Lon:      lon,
				Signal:   signal,
			},
			DeviceKey: deviceKey,
		})
		pendingSightingKeys[sightingKey] = struct{}{}
		rowsInTx++

		if len(pendingSightings) >= flushSize {
			if err := flushBatch(); err != nil {
				return err
			}
			report("devices")
		}
		if rowsInTx >= txChunkSize {
			return commitAndReopen()
		}
		return nil
	})
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := flushBatch(); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("final commit: %w", err)
	}

	log.Printf("[rk] done: devices=%d sightings=%d skipped=%d",
		out.Devices, out.Sightings, out.Skipped)
	return out, nil
}

// parseTime parses the date format used in both CSV kinds.
func parseTime(s string) (time.Time, error) {
	const layout = "02.01.2006 15:04:05"
	t, err := time.ParseInLocation(layout, s, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse time %q: %w", s, err)
	}
	return t, nil
}

// nearestLocation finds the closest location by time using binary search.
func nearestLocation(locs []repo.LocationInserted, t time.Time, tolerance time.Duration) *int64 {
	if len(locs) == 0 {
		return nil
	}
	i := sort.Search(len(locs), func(i int) bool {
		return !locs[i].SeenAt.Before(t)
	})

	bestIdx := -1
	bestDiff := tolerance + 1

	for _, ci := range []int{i - 1, i} {
		if ci < 0 || ci >= len(locs) {
			continue
		}
		diff := locs[ci].SeenAt.Sub(t)
		if diff < 0 {
			diff = -diff
		}
		if diff <= tolerance && diff < bestDiff {
			bestDiff = diff
			bestIdx = ci
		}
	}

	if bestIdx < 0 {
		return nil
	}
	id := locs[bestIdx].ID
	return &id
}
