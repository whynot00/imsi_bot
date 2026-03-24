package repo

import (
	"context"
	"database/sql"
	"fmt"
)

// DeviceRepo handles upsert and lookup of devices.
type DeviceRepo struct {
	db *sql.DB
}

func NewDeviceRepo(db *sql.DB) *DeviceRepo {
	return &DeviceRepo{db: db}
}

// Upsert inserts a device or returns the existing id if (imsi, imei) already
// exists. Either imsi or imei may be empty but not both.
func (r *DeviceRepo) Upsert(ctx context.Context, imsi, imei string) (int64, error) {
	if imsi == "" && imei == "" {
		return 0, fmt.Errorf("device must have at least one of imsi or imei")
	}

	var id int64
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO devices (imsi, imei)
		VALUES (NULLIF($1, ''), NULLIF($2, ''))
		ON CONFLICT ON CONSTRAINT uq_devices_imsi_imei
		DO UPDATE SET imsi = EXCLUDED.imsi
		RETURNING id`,
		imsi, imei,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("upsert device: %w", err)
	}
	return id, nil
}

// FindByIMSI returns all devices matching the given IMSI.
func (r *DeviceRepo) FindByIMSI(ctx context.Context, imsi string) ([]Device, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, COALESCE(imsi, ''), COALESCE(imei, '')
		FROM devices
		WHERE imsi = $1`,
		imsi,
	)
	if err != nil {
		return nil, fmt.Errorf("find by imsi: %w", err)
	}
	defer rows.Close()
	return scanDevices(rows)
}

// FindByIMEI returns all devices matching the given IMEI.
func (r *DeviceRepo) FindByIMEI(ctx context.Context, imei string) ([]Device, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, COALESCE(imsi, ''), COALESCE(imei, '')
		FROM devices
		WHERE imei = $1`,
		imei,
	)
	if err != nil {
		return nil, fmt.Errorf("find by imei: %w", err)
	}
	defer rows.Close()
	return scanDevices(rows)
}

func scanDevices(rows *sql.Rows) ([]Device, error) {
	var out []Device
	for rows.Next() {
		var d Device
		if err := rows.Scan(&d.ID, &d.IMSI, &d.IMEI); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}
