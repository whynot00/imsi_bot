package repo

import (
	"context"
	"database/sql"
	"fmt"
)

// IPRepo manages the allowed_ips table.
type IPRepo struct {
	db *sql.DB
}

func NewIPRepo(db *sql.DB) *IPRepo {
	return &IPRepo{db: db}
}

// List returns all allowed IPs.
func (r *IPRepo) List(ctx context.Context) ([]AllowedIP, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, ip, created_at FROM allowed_ips ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list ips: %w", err)
	}
	defer rows.Close()

	var out []AllowedIP
	for rows.Next() {
		var a AllowedIP
		if err := rows.Scan(&a.ID, &a.IP, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// AllIPs returns just the IP strings (used by middleware).
func (r *IPRepo) AllIPs(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT ip FROM allowed_ips`)
	if err != nil {
		return nil, fmt.Errorf("all ips: %w", err)
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var ip string
		if err := rows.Scan(&ip); err != nil {
			return nil, err
		}
		out = append(out, ip)
	}
	return out, rows.Err()
}

// Create adds a new allowed IP. Returns true if created, false if already exists.
func (r *IPRepo) Create(ctx context.Context, ip string) (bool, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO allowed_ips (ip) VALUES ($1) ON CONFLICT (ip) DO NOTHING`,
		ip,
	)
	if err != nil {
		return false, fmt.Errorf("create ip: %w", err)
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// Delete removes an allowed IP by ID. Returns true if deleted.
func (r *IPRepo) Delete(ctx context.Context, id int64) (bool, error) {
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM allowed_ips WHERE id = $1`,
		id,
	)
	if err != nil {
		return false, fmt.Errorf("delete ip: %w", err)
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}
