package whitelist

import "context"

type Repository interface {
	Create(ctx context.Context, telegramID int64) error
	Touch(ctx context.Context, telegramID int64) (*User, error)
	Update(ctx context.Context, user *User) error
}
