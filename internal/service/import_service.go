// Package service contains business logic that orchestrates repositories
// and the parser layer.
package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/whynot00/imsi-bot/internal/parser"
	"github.com/whynot00/imsi-bot/internal/repo"
)

const locationTolerance = 30 * time.Second

// ImportResult summarises what was written to the database.
type ImportResult struct {
	Devices   int
	Locations int
	Sightings int
	Skipped   int
}

// ImportService orchestrates parsing + persistence for both file kinds.
type ImportService struct {
	devices  *repo.DeviceRepo
	parametr *repo.ParametrRepo
	rk       *repo.RKRepo
}

func NewImportService(
	devices *repo.DeviceRepo,
	parametr *repo.ParametrRepo,
	rk *repo.RKRepo,
) *ImportService {
	return &ImportService{devices: devices, parametr: parametr, rk: rk}
}

// ImportParametr persists a parsed parametr result.
// Locations are inserted first; each device sighting is then linked to the
// nearest location within locationTolerance.
func (s *ImportService) ImportParametr(ctx context.Context, result *parser.ParseResult) (*ImportResult, error) {
	out := &ImportResult{}

	// 1. Insert locations.
	for _, l := range result.Locations {
		lat, err := strconv.ParseFloat(l.Lat, 64)
		if err != nil {
			out.Skipped++
			continue
		}
		lon, err := strconv.ParseFloat(l.Lon, 64)
		if err != nil {
			out.Skipped++
			continue
		}
		seenAt, err := parseTime(l.Date)
		if err != nil {
			out.Skipped++
			continue
		}
		if _, err := s.parametr.InsertLocation(ctx, seenAt, lat, lon); err != nil {
			return nil, fmt.Errorf("insert location: %w", err)
		}
		out.Locations++
	}

	// 2. Insert devices + sightings.
	for _, d := range result.Devices {
		seenAt, err := parseTime(d.Date)
		if err != nil {
			out.Skipped++
			continue
		}

		deviceID, err := s.devices.Upsert(ctx, d.IMSI, d.IMEI)
		if err != nil {
			return nil, fmt.Errorf("upsert device: %w", err)
		}
		out.Devices++

		locID, err := s.parametr.NearestLocation(ctx, seenAt, locationTolerance)
		if err != nil {
			return nil, fmt.Errorf("nearest location: %w", err)
		}

		sp := repo.SightingParametr{
			DeviceID: deviceID,
			SeenAt:   seenAt,
			Standart: d.Standart,
			Operator: d.Operator,
			Event:    d.Event,
		}
		if locID != 0 {
			sp.LocationID = &locID
		}

		if err := s.parametr.InsertSighting(ctx, sp); err != nil {
			return nil, fmt.Errorf("insert sighting: %w", err)
		}
		out.Sightings++
	}

	return out, nil
}

// ImportRK persists a slice of parsed rk devices.
func (s *ImportService) ImportRK(ctx context.Context, devices []parser.RawDevice) (*ImportResult, error) {
	out := &ImportResult{}

	for _, d := range devices {
		seenAt, err := parseTime(d.Date)
		if err != nil {
			out.Skipped++
			continue
		}

		lat, err := strconv.ParseFloat(d.Lat, 64)
		if err != nil {
			out.Skipped++
			continue
		}
		lon, err := strconv.ParseFloat(d.Lon, 64)
		if err != nil {
			out.Skipped++
			continue
		}
		signal, _ := strconv.Atoi(d.SignalStrength)

		deviceID, err := s.devices.Upsert(ctx, d.IMSI, d.IMEI)
		if err != nil {
			return nil, fmt.Errorf("upsert device: %w", err)
		}
		out.Devices++

		if err := s.rk.InsertSighting(ctx, repo.SightingRK{
			DeviceID: deviceID,
			SeenAt:   seenAt,
			Standart: d.Standart,
			Lat:      lat,
			Lon:      lon,
			Signal:   signal,
		}); err != nil {
			return nil, fmt.Errorf("insert rk sighting: %w", err)
		}
		out.Sightings++
	}

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
