package whitelist

import (
	"time"
)

type User struct {
	ID           int64     `db:"id"`
	Username     string    `db:"username"`
	TelegramID   int64     `db:"telegram_id"`
	CreatedAt    time.Time `db:"created_at"`
	LastActivity time.Time `db:"last_activity"`
}
