package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/whynot00/imsi_bot/internal/config"
	"github.com/whynot00/imsi_bot/internal/datapump"
	"github.com/whynot00/imsi_bot/internal/repository"
	"github.com/whynot00/imsi_bot/pkg/postgres"
)

type flags struct {
	files []string
}

func main() {

	fl := parseFlags()
	if fl == nil || fl.files == nil {
		fmt.Println("Флаг -f должен быть указан")
		return
	}

	log.Println("[INFO] Files:")
	for _, f := range fl.files {
		fmt.Printf("\t\t\t- %s\n", f)
	}

	ctx := context.Background()
	cfg := config.Load()
	pg := postgres.Load(cfg.Postgers.Host, cfg.Postgers.Port, cfg.Postgers.User, cfg.Postgers.Password, cfg.Postgers.DB, cfg.Postgers.SSLMode)

	t := time.Now()
	pump := datapump.NewPump(1500, repository.NewObsRepository(pg))
	countRows, err := pump.Pump(ctx, fl.files...)
	if err != nil {
		panic(err)
	}

	log.Printf("[INFO] Inserted %d rows in %.2f seconds", countRows, time.Since(t).Seconds())
}

func parseFlags() *flags {

	flsRaw := flag.String("f", "", `Укажите путь до файлов (в кавычках можно передать несколько файлов "File1.csv File2.csv")`)

	flag.Parse()

	if *flsRaw == "" {
		return nil
	}

	return &flags{
		files: strings.Split(*flsRaw, " "),
	}
}
