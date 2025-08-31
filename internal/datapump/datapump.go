package datapump

import (
	"context"
	"encoding/csv"
	"io"
	"os"

	"github.com/whynot00/imsi_bot/internal/domain/observation"
)

type PumpMaster struct {
	batchSize int
	obs       observation.Repository

	batch []observation.Observation
}

func NewPump(batchSize int, obs observation.Repository) *PumpMaster {

	return &PumpMaster{
		batchSize: batchSize,
		obs:       obs,

		batch: make([]observation.Observation, 0, batchSize),
	}
}

func (p *PumpMaster) Pump(ctx context.Context, file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}

	return p.load(ctx, f)
}

func (p *PumpMaster) load(ctx context.Context, file io.Reader) error {

	r := csv.NewReader(file)

	r.Comma = ';'
	r.LazyQuotes = true

	for {
		line, err := r.Read()
		if err == io.EOF {
			return p.commitLast(ctx)

		}

		if err := p.parse(ctx, line); err != nil {
			return err
		}
	}
}
