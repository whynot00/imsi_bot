package middleware

import (
	"log/slog"

	"github.com/whynot00/imsi_bot/internal/domain/whitelist"
)

type Middleware struct {
	whl whitelist.Repository
	log *slog.Logger
}

func Create(whlRepo whitelist.Repository) *Middleware {

	return &Middleware{
		whl: whlRepo,
		log: slog.Default(),
	}
}
