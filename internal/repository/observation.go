package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	obs "github.com/whynot00/imsi_bot/internal/domain/observation"
)

type ObservationRepositrory struct {
	db *sqlx.DB
}

func NewObsRepository(db *sqlx.DB) obs.Repository {

	return &ObservationRepositrory{
		db: db,
	}
}

func (r *ObservationRepositrory) WriteBatch(ctx context.Context, obs []obs.Observation) error {

	query := `
		INSERT INTO entities (standart, operator, date, lon, lat, signal_strength, imei, imsi, hash)
		VALUES (:standart, :operator, :date, :lon, :lat, :signal_strength, :imei, :imsi, :hash)
		ON CONFLICT (hash) DO NOTHING;`

	if _, err := r.db.NamedExec(query, obs); err != nil {
		return handleErr(err)
	}

	return nil
}

func (r *ObservationRepositrory) GetByIMSI(ctx context.Context, imsi string) ([]obs.Observation, error) {
	query := `
        SELECT
            standart,
            operator,
            date,
            lon,
            lat,
            signal_strength,
            imei,
            imsi,
            hash
        FROM entities
        WHERE imsi = $1
    `
	var result []obs.Observation
	if err := r.db.SelectContext(ctx, &result, query, imsi); err != nil {
		return nil, handleErr(err)
	}

	return result, nil
}

func (r *ObservationRepositrory) GetByIMEI(ctx context.Context, imei string) ([]obs.Observation, error) {

	query := `
        SELECT
            standart,
            operator,
            date,
            lon,
            lat,
			signal_strength,
            imei,
            imsi,
            hash
        FROM entities
        WHERE imei = $1
    `
	var result []obs.Observation
	if err := r.db.SelectContext(ctx, &result, query, imei); err != nil {
		return nil, handleErr(err)
	}

	return result, nil
}
