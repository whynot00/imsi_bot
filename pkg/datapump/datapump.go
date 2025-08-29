package datapump

import (
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/whynot00/imsi_bot/pkg/datapump/internal/filter"
	"github.com/whynot00/imsi_bot/pkg/datapump/internal/reader"
)

func Pump(file string, pg *sqlx.DB) {
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	lineCh := make(chan []string)

	go filter.Parse(lineCh, 500, pg)

	reader.Load(f, lineCh)
}
