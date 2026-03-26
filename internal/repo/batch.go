package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const batchSize = 1000

// TxFunc runs fn inside a transaction. If fn returns an error the tx is rolled back.
func TxFunc(ctx context.Context, db *sql.DB, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// LocationInserted is a successfully inserted location row.
type LocationInserted struct {
	ID     int64
	SeenAt time.Time
}

// BatchUpsertDevices inserts devices in batches and returns a map of "imsi|imei" → device_id.
// Duplicates within the input must be removed by the caller.
func BatchUpsertDevices(ctx context.Context, tx *sql.Tx, pairs [][2]string) (map[string]int64, error) {
	result := make(map[string]int64, len(pairs))

	for i := 0; i < len(pairs); i += batchSize {
		end := i + batchSize
		if end > len(pairs) {
			end = len(pairs)
		}
		batch := pairs[i:end]

		var sb strings.Builder
		sb.WriteString(`INSERT INTO devices (imsi, imei) VALUES `)
		args := make([]any, 0, len(batch)*2)
		for j, p := range batch {
			if j > 0 {
				sb.WriteByte(',')
			}
			n := j * 2
			fmt.Fprintf(&sb, "(NULLIF($%d,''),NULLIF($%d,''))", n+1, n+2)
			args = append(args, p[0], p[1])
		}
		sb.WriteString(` ON CONFLICT ON CONSTRAINT uq_devices_imsi_imei DO UPDATE SET imsi = EXCLUDED.imsi RETURNING id`)

		rows, err := tx.QueryContext(ctx, sb.String(), args...)
		if err != nil {
			return nil, fmt.Errorf("batch upsert devices: %w", err)
		}
		j := 0
		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return nil, err
			}
			p := batch[j]
			key := p[0] + "|" + p[1]
			result[key] = id
			j++
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// BatchInsertLocations inserts locations in batches and returns the inserted rows.
// Duplicates within the input must be removed by the caller.
func BatchInsertLocations(ctx context.Context, tx *sql.Tx, locs []struct {
	SeenAt time.Time
	Lat    float64
	Lon    float64
}) ([]LocationInserted, error) {
	var result []LocationInserted

	for i := 0; i < len(locs); i += batchSize {
		end := i + batchSize
		if end > len(locs) {
			end = len(locs)
		}
		batch := locs[i:end]

		var sb strings.Builder
		sb.WriteString(`INSERT INTO locations_parametr (seen_at, lat, lon) VALUES `)
		args := make([]any, 0, len(batch)*3)
		for j, l := range batch {
			if j > 0 {
				sb.WriteByte(',')
			}
			n := j * 3
			fmt.Fprintf(&sb, "($%d,$%d,$%d)", n+1, n+2, n+3)
			args = append(args, l.SeenAt, l.Lat, l.Lon)
		}
		sb.WriteString(` ON CONFLICT ON CONSTRAINT uq_locations_parametr DO UPDATE SET seen_at = EXCLUDED.seen_at RETURNING id, seen_at`)

		rows, err := tx.QueryContext(ctx, sb.String(), args...)
		if err != nil {
			return nil, fmt.Errorf("batch insert locations: %w", err)
		}
		for rows.Next() {
			var li LocationInserted
			if err := rows.Scan(&li.ID, &li.SeenAt); err != nil {
				rows.Close()
				return nil, err
			}
			result = append(result, li)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// SightingParametrRow is a row ready for batch insert into sightings_parametr.
type SightingParametrRow struct {
	DeviceID   int64
	SeenAt     time.Time
	Standart   string
	Operator   string
	Event      string
	LocationID *int64
}

// BatchInsertSightingsParametr inserts sightings in batches. Returns count of inserted rows.
func BatchInsertSightingsParametr(ctx context.Context, tx *sql.Tx, sightings []SightingParametrRow) (int, error) {
	total := 0

	for i := 0; i < len(sightings); i += batchSize {
		end := i + batchSize
		if end > len(sightings) {
			end = len(sightings)
		}
		batch := sightings[i:end]

		var sb strings.Builder
		sb.WriteString(`INSERT INTO sightings_parametr (device_id, seen_at, standart, operator, event, location_id) VALUES `)
		args := make([]any, 0, len(batch)*6)
		for j, s := range batch {
			if j > 0 {
				sb.WriteByte(',')
			}
			n := j * 6
			fmt.Fprintf(&sb, "($%d,$%d,$%d,$%d,$%d,$%d)", n+1, n+2, n+3, n+4, n+5, n+6)
			var standart, operator, event *string
			if s.Standart != "" {
				standart = &s.Standart
			}
			if s.Operator != "" {
				operator = &s.Operator
			}
			if s.Event != "" {
				event = &s.Event
			}
			var locID *int64
			if s.LocationID != nil && *s.LocationID != 0 {
				locID = s.LocationID
			}
			args = append(args, s.DeviceID, s.SeenAt, standart, operator, event, locID)
		}
		sb.WriteString(` ON CONFLICT ON CONSTRAINT uq_sightings_parametr DO NOTHING`)

		res, err := tx.ExecContext(ctx, sb.String(), args...)
		if err != nil {
			return total, fmt.Errorf("batch insert sightings parametr: %w", err)
		}
		n, _ := res.RowsAffected()
		total += int(n)
	}

	return total, nil
}

// SightingRKRow is a row ready for batch insert into sightings_rk.
type SightingRKRow struct {
	DeviceID int64
	SeenAt   time.Time
	Standart string
	Lat      float64
	Lon      float64
	Signal   int
}

// BatchInsertSightingsRK inserts RK sightings in batches. Returns count of inserted rows.
func BatchInsertSightingsRK(ctx context.Context, tx *sql.Tx, sightings []SightingRKRow) (int, error) {
	total := 0

	for i := 0; i < len(sightings); i += batchSize {
		end := i + batchSize
		if end > len(sightings) {
			end = len(sightings)
		}
		batch := sightings[i:end]

		var sb strings.Builder
		sb.WriteString(`INSERT INTO sightings_rk (device_id, seen_at, standart, lat, lon, signal) VALUES `)
		args := make([]any, 0, len(batch)*6)
		for j, s := range batch {
			if j > 0 {
				sb.WriteByte(',')
			}
			n := j * 6
			fmt.Fprintf(&sb, "($%d,$%d,$%d,$%d,$%d,$%d)", n+1, n+2, n+3, n+4, n+5, n+6)
			var standart *string
			if s.Standart != "" {
				standart = &s.Standart
			}
			var lat, lon *float64
			if s.Lat != 0 {
				lat = &s.Lat
			}
			if s.Lon != 0 {
				lon = &s.Lon
			}
			var signal *int
			if s.Signal != 0 {
				signal = &s.Signal
			}
			args = append(args, s.DeviceID, s.SeenAt, standart, lat, lon, signal)
		}
		sb.WriteString(` ON CONFLICT ON CONSTRAINT uq_sightings_rk DO NOTHING`)

		res, err := tx.ExecContext(ctx, sb.String(), args...)
		if err != nil {
			return total, fmt.Errorf("batch insert sightings rk: %w", err)
		}
		n, _ := res.RowsAffected()
		total += int(n)
	}

	return total, nil
}
