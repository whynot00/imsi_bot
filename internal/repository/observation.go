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

func (r *ObservationRepositrory) WriteBatch(ctx context.Context, obs []obs.Observation) (int, error) {
	if len(obs) == 0 {
		return 0, nil
	}

	const q = `
		INSERT INTO entities (standart, operator, date, lon, lat, signal_strength, imei, imsi, hash)
		VALUES (:standart, :operator, :date, :lon, :lat, :signal_strength, :imei, :imsi, :hash)
		ON CONFLICT (hash) DO NOTHING
		RETURNING 1;`

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, handleErr(err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	rows, err := tx.NamedQuery(q, obs)
	if err != nil {
		return 0, handleErr(err)
	}
	defer rows.Close()

	inserted := 0
	for rows.Next() {
		var one int
		if err := rows.Scan(&one); err != nil {
			return 0, handleErr(err)
		}
		inserted++
	}
	if err := rows.Err(); err != nil {
		return 0, handleErr(err)
	}

	if err := tx.Commit(); err != nil {
		return 0, handleErr(err)
	}

	return inserted, nil
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
