package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/whynot00/imsi_bot/internal/domain/whitelist"
)

type WhitelistRepository struct {
	db *sqlx.DB
}

func NewWhitelistRepository(db *sqlx.DB) whitelist.Repository {
	return &WhitelistRepository{db: db}
}

func (r *WhitelistRepository) Create(ctx context.Context, telegramID int64) error {

	query := `INSERT INTO users(telegram_id) VALUES ($1)`

	_, err := r.db.ExecContext(ctx, query, telegramID)
	if err != nil {
		return handleErr(err)
	}

	return nil
}

func (r *WhitelistRepository) Touch(ctx context.Context, telegramID int64) (*whitelist.User, error) {

	query := `UPDATE users SET last_activity = NOW() WHERE telegram_id = $1 RETURNING *`

	var user whitelist.User
	if err := r.db.QueryRowxContext(ctx, query, telegramID).StructScan(&user); err != nil {
		return nil, handleErr(err)
	}

	return &user, nil
}

func (r *WhitelistRepository) Update(ctx context.Context, user *whitelist.User) error {

	query := `UPDATE users SET username = $1 WHERE telegram_id = $2`

	if _, err := r.db.ExecContext(ctx, query, user.Username, user.TelegramID); err != nil {
		return handleErr(err)
	}

	return nil
}
