package repo

import (
	"context"
	"database/sql"
	"fmt"
)

// UserRepo checks whether a Telegram user is authorised.
type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

// Exists returns true if the given Telegram user_id is in the users table.
func (r *UserRepo) Exists(ctx context.Context, telegramID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`,
		telegramID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("user exists: %w", err)
	}
	return exists, nil
}
