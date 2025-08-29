package errorx

import "errors"

var (
	ErrUserIsExists = errors.New("user is already exists")
	ErrNoRows       = errors.New("no rows in result set")
)

type ReqError struct {
	UserID  int64
	Request string
	Err     error
}
