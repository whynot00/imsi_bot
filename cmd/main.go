package main

import (
	"context"

	"github.com/whynot00/imsi_bot/internal/config"
	"github.com/whynot00/imsi_bot/internal/repository"
	"github.com/whynot00/imsi_bot/internal/telegram"
	"github.com/whynot00/imsi_bot/pkg/postgres"
)

func main() {
	ctx := context.Background()
	cfg := config.Load()

	pg := postgres.Load(cfg.Postgers.Host, cfg.Postgers.Port, cfg.Postgers.User, cfg.Postgers.Password, cfg.Postgers.DB, cfg.Postgers.SSLMode)

	bot := telegram.New(
		ctx,
		pg,
		cfg.Telegram.Token,
		repository.NewObsRepository(pg),
		repository.NewWhitelistRepository(pg),
	).InitRoutes()

	bot.Start(ctx)
}
