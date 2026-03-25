package repo

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// ParametrRepo handles locations and sightings from parametr-format files.
type ParametrRepo struct {
	db *sql.DB
}

func NewParametrRepo(db *sql.DB) *ParametrRepo {
	return &ParametrRepo{db: db}
}

// InsertLocation upserts a location ping and returns its id.
func (r *ParametrRepo) InsertLocation(ctx context.Context, seenAt time.Time, lat, lon float64) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO locations_parametr (seen_at, lat, lon)
		VALUES ($1, $2, $3)
		ON CONFLICT ON CONSTRAINT uq_locations_parametr
		DO UPDATE SET seen_at = EXCLUDED.seen_at
		RETURNING id`,
		seenAt, lat, lon,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert location: %w", err)
	}
	return id, nil
}

// NearestLocation returns the id of the location closest in time to seenAt
// within the given tolerance window. Returns 0, nil if none found.
func (r *ParametrRepo) NearestLocation(ctx context.Context, seenAt time.Time, tolerance time.Duration) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
		SELECT id
		FROM locations_parametr
		WHERE seen_at BETWEEN $1 AND $2
		ORDER BY ABS(EXTRACT(EPOCH FROM (seen_at - $3)))
		LIMIT 1`,
		seenAt.Add(-tolerance),
		seenAt.Add(tolerance),
		seenAt,
	).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("nearest location: %w", err)
	}
	return id, nil
}

// InsertSighting upserts a sighting. locationID may be 0 (no nearby location).
func (r *ParametrRepo) InsertSighting(ctx context.Context, s SightingParametr) error {
	var locID *int64
	if s.LocationID != nil && *s.LocationID != 0 {
		locID = s.LocationID
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sightings_parametr (device_id, seen_at, standart, operator, event, location_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT ON CONSTRAINT uq_sightings_parametr DO NOTHING`,
		s.DeviceID, s.SeenAt,
		nullText(s.Standart), nullText(s.Operator), nullText(s.Event),
		locID,
	)
	if err != nil {
		return fmt.Errorf("insert sighting: %w", err)
	}
	return nil
}

// FindSightingsByDeviceID returns all parametr sightings for a device,
// ordered by time ascending.
func (r *ParametrRepo) FindSightingsByDeviceID(ctx context.Context, deviceID int64) ([]SightingParametr, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT 
			s.id, s.device_id, s.seen_at,
			COALESCE(s.standart, ''), 
			COALESCE(s.operator, ''), 
			COALESCE(s.event, ''),
			s.location_id,
			l.lat,
			l.lon
		FROM sightings_parametr s
		LEFT JOIN locations_parametr l ON l.id = s.location_id
		WHERE s.device_id = $1
		ORDER BY s.seen_at ASC`,
		deviceID,
	)
	if err != nil {
		return nil, fmt.Errorf("find sightings: %w", err)
	}
	defer rows.Close()

	var out []SightingParametr
	for rows.Next() {
		var s SightingParametr
		if err := rows.Scan(
			&s.ID, &s.DeviceID, &s.SeenAt,
			&s.Standart, &s.Operator, &s.Event,
			&s.LocationID, &s.Lat, &s.Lon,
		); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func nullText(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
