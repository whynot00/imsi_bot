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

// Create inserts a new user. Returns true if created, false if already exists.
func (r *UserRepo) Create(ctx context.Context, telegramID int64, username string) (bool, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, username) VALUES ($1, $2) ON CONFLICT (id) DO NOTHING`,
		telegramID, username,
	)
	if err != nil {
		return false, fmt.Errorf("create user: %w", err)
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// Delete removes a user by Telegram ID. Returns true if deleted.
func (r *UserRepo) Delete(ctx context.Context, telegramID int64) (bool, error) {
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM users WHERE id = $1`,
		telegramID,
	)
	if err != nil {
		return false, fmt.Errorf("delete user: %w", err)
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// List returns all users.
func (r *UserRepo) List(ctx context.Context) ([]User, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, COALESCE(username, ''), created_at FROM users ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var out []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}
