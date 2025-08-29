package repository

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
	"github.com/whynot00/imsi_bot/internal/domain/errorx"
)

func handleErr(err error) error {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {

		switch pqErr.Code {
		case "23505":
			return errorx.ErrUserIsExists
		}

	}

	if errors.Is(err, sql.ErrNoRows) {
		return errorx.ErrNoRows
	}

	return err
}
