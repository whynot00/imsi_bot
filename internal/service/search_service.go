package service

import (
	"context"
	"fmt"

	"github.com/whynot00/imsi-bot/internal/repo"
)

// SearchResult is returned by both ByIMSI and ByIMEI.
type SearchResult struct {
	Device            repo.Device
	SightingsParametr []repo.SightingParametr
	SightingsRK       []repo.SightingRK
}

// SuggestResult is a single autocomplete suggestion.
type SuggestResult struct {
	Value string `json:"value"`
	Kind  string `json:"kind"` // "imsi" or "imei"
}

// SearchService provides search and autocomplete over the devices table.
type SearchService struct {
	search  *repo.SearchRepo
	devices *repo.DeviceRepo
}

func NewSearchService(search *repo.SearchRepo, devices *repo.DeviceRepo) *SearchService {
	return &SearchService{search: search, devices: devices}
}

// ByIMSI returns a device and all its sightings.
func (s *SearchService) ByIMSI(ctx context.Context, imsi string) (*SearchResult, error) {
	if imsi == "" {
		return nil, fmt.Errorf("imsi is required")
	}
	r, err := s.search.ByIMSI(ctx, imsi)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	return &SearchResult{
		Device:            r.Device,
		SightingsParametr: r.SightingsParametr,
		SightingsRK:       r.SightingsRK,
	}, nil
}

// ByIMEI returns a device and all its sightings.
func (s *SearchService) ByIMEI(ctx context.Context, imei string) (*SearchResult, error) {
	if imei == "" {
		return nil, fmt.Errorf("imei is required")
	}
	r, err := s.search.ByIMEI(ctx, imei)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	return &SearchResult{
		Device:            r.Device,
		SightingsParametr: r.SightingsParametr,
		SightingsRK:       r.SightingsRK,
	}, nil
}

// Suggest returns up to 10 IMSI or IMEI values matching the given prefix.
// kind must be "imsi" or "imei".
func (s *SearchService) Suggest(ctx context.Context, prefix, kind string) ([]SuggestResult, error) {
	if len(prefix) < 3 {
		return nil, nil // don't query on very short prefixes
	}

	var values []string
	var err error

	switch kind {
	case "imsi":
		values, err = s.devices.SuggestIMSI(ctx, prefix)
	case "imei":
		values, err = s.devices.SuggestIMEI(ctx, prefix)
	default:
		return nil, fmt.Errorf("kind must be imsi or imei")
	}
	if err != nil {
		return nil, err
	}

	out := make([]SuggestResult, len(values))
	for i, v := range values {
		out[i] = SuggestResult{Value: v, Kind: kind}
	}
	return out, nil
}
