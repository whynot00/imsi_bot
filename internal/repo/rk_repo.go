package repo

import (
	"context"
	"database/sql"
	"fmt"
)

// RKRepo handles sightings from rk-format files.
type RKRepo struct {
	db *sql.DB
}

func NewRKRepo(db *sql.DB) *RKRepo {
	return &RKRepo{db: db}
}

// InsertSighting upserts a sighting from an rk-format file.
func (r *RKRepo) InsertSighting(ctx context.Context, s SightingRK) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sightings_rk (device_id, seen_at, standart, lat, lon, signal)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT ON CONSTRAINT uq_sightings_rk DO NOTHING`,
		s.DeviceID, s.SeenAt,
		nullText(s.Standart),
		nullFloat(s.Lat), nullFloat(s.Lon),
		nullInt(s.Signal),
	)
	if err != nil {
		return fmt.Errorf("insert rk sighting: %w", err)
	}
	return nil
}

// FindSightingsByDeviceID returns all rk sightings for a device,
// ordered by time ascending.
func (r *RKRepo) FindSightingsByDeviceID(ctx context.Context, deviceID int64) ([]SightingRK, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, device_id, seen_at,
		       COALESCE(standart, ''),
		       COALESCE(lat, 0), COALESCE(lon, 0),
		       COALESCE(signal, 0)
		FROM sightings_rk
		WHERE device_id = $1
		ORDER BY seen_at ASC`,
		deviceID,
	)
	if err != nil {
		return nil, fmt.Errorf("find rk sightings: %w", err)
	}
	defer rows.Close()

	var out []SightingRK
	for rows.Next() {
		var s SightingRK
		if err := rows.Scan(
			&s.ID, &s.DeviceID, &s.SeenAt,
			&s.Standart, &s.Lat, &s.Lon, &s.Signal,
		); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func nullFloat(f float64) *float64 {
	if f == 0 {
		return nil
	}
	return &f
}

func nullInt(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}
