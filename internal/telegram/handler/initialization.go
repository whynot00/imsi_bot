package handler

import (
	"log/slog"

	"github.com/jmoiron/sqlx"
	"github.com/whynot00/imsi_bot/internal/datapump"
	"github.com/whynot00/imsi_bot/internal/domain/observation"
	"github.com/whynot00/imsi_bot/internal/domain/whitelist"
)

const (
	CreateNewUserMessageKey = "create_new_user"
)

type Handler struct {
	obs observation.Repository
	whl whitelist.Repository

	pump *datapump.PumpMaster
	db   *sqlx.DB
	log  *slog.Logger
}

func Create(db *sqlx.DB, obsRepo observation.Repository, whlRepo whitelist.Repository) *Handler {

	return &Handler{
		obs: obsRepo,
		whl: whlRepo,

		pump: datapump.NewPump(500, obsRepo),
		db:   db,
		log:  slog.Default(),
	}
}
