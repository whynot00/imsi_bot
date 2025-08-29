package handler

import (
	"log/slog"

	"github.com/jmoiron/sqlx"
	"github.com/whynot00/imsi_bot/internal/domain/observation"
	"github.com/whynot00/imsi_bot/internal/domain/whitelist"
)

type Handler struct {
	obs observation.Repository
	whl whitelist.Repository

	db  *sqlx.DB
	log *slog.Logger
}

func Create(db *sqlx.DB, obsRepo observation.Repository, whl whitelist.Repository) *Handler {

	return &Handler{
		obs: obsRepo,
		whl: whl,
		db:  db,
		log: slog.Default(),
	}
}
