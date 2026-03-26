package repo

import (
	"context"
	"database/sql"
	"fmt"
)

// SearchRepo provides cross-table search by IMSI or IMEI.
type SearchRepo struct {
	db       *sql.DB
	devices  *DeviceRepo
	parametr *ParametrRepo
	rk       *RKRepo
}

func NewSearchRepo(db *sql.DB) *SearchRepo {
	return &SearchRepo{
		db:       db,
		devices:  NewDeviceRepo(db),
		parametr: NewParametrRepo(db),
		rk:       NewRKRepo(db),
	}
}

// ByIMSI returns a device and all its sightings across both source types.
func (r *SearchRepo) ByIMSI(ctx context.Context, imsi string) (*DeviceResult, error) {
	return r.search(ctx, func() ([]Device, error) {
		return r.devices.FindByIMSI(ctx, imsi)
	})
}

// ByIMEI returns a device and all its sightings across both source types.
func (r *SearchRepo) ByIMEI(ctx context.Context, imei string) (*DeviceResult, error) {
	return r.search(ctx, func() ([]Device, error) {
		return r.devices.FindByIMEI(ctx, imei)
	})
}

func (r *SearchRepo) search(ctx context.Context, findDevice func() ([]Device, error)) (*DeviceResult, error) {
	devices, err := findDevice()
	if err != nil {
		return nil, err
	}
	if len(devices) == 0 {
		return nil, nil
	}

	// There should only ever be one device per IMSI/IMEI due to the unique
	// constraint, but we take the first defensively.
	dev := devices[0]

	sp, err := r.parametr.FindSightingsByDeviceID(ctx, dev.ID)
	if err != nil {
		return nil, fmt.Errorf("parametr sightings: %w", err)
	}

	sr, err := r.rk.FindSightingsByDeviceID(ctx, dev.ID)
	if err != nil {
		return nil, fmt.Errorf("rk sightings: %w", err)
	}

	return &DeviceResult{
		Device:            dev,
		SightingsParametr: sp,
		SightingsRK:       sr,
	}, nil
}

// WithLocation joins a sighting's location inline — handy for display.
func (r *SearchRepo) WithLocation(ctx context.Context, locationID int64) (*LocationParametr, error) {
	var l LocationParametr
	err := r.db.QueryRowContext(ctx, `
		SELECT id, seen_at, lat, lon
		FROM locations_parametr
		WHERE id = $1`,
		locationID,
	).Scan(&l.ID, &l.SeenAt, &l.Lat, &l.Lon)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get location: %w", err)
	}
	return &l, nil
}
