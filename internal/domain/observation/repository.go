package observation

import (
	"context"
)

type Repository interface {
	WriteBatch(ctx context.Context, obs []Observation) error
	GetByIMSI(ctx context.Context, imsi string) ([]Observation, error)
	GetByIMEI(ctx context.Context, imsi string) ([]Observation, error)
}
