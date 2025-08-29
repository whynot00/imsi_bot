package observation

import (
	"context"
)

type Repository interface {
	GetByIMSI(ctx context.Context, imsi string) ([]Observation, error)
	GetByIMEI(ctx context.Context, imsi string) ([]Observation, error)
}
