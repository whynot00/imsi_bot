package main

import (
	"context"

	"github.com/whynot00/imsi_bot/internal/config"
	"github.com/whynot00/imsi_bot/internal/datapump"
	"github.com/whynot00/imsi_bot/internal/repository"
	"github.com/whynot00/imsi_bot/pkg/postgres"
)

func main() {

	ctx := context.Background()
	cfg := config.Load()
	pg := postgres.Load(cfg.Postgers.Host, cfg.Postgers.Port, cfg.Postgers.User, cfg.Postgers.Password, cfg.Postgers.DB, cfg.Postgers.SSLMode)

	pump := datapump.NewPump(500, repository.NewObsRepository(pg))
	if err := pump.Pump(ctx, "data/uaz.csv"); err != nil {
		panic(err)
	}
}
