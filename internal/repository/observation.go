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
		return nil, err
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
		return nil, err
	}
	return result, nil
}
