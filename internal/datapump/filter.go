package datapump

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/whynot00/imsi_bot/internal/domain/observation"
)

func (p *PumpMaster) parse(ctx context.Context, line []string) {

	obs := observation.Observation{
		Standart: line[standart],
		Operator: imsiOperators[line[opertorCode]],
		Date:     parseDate(line[date]),
		IMEI:     parseIMEIorIMSI(line[imei]),
		IMSI:     parseIMEIorIMSI(line[imsi]),
	}

	lat, lon := parseCoords(line[coords])

	obs.LAT = lat
	obs.LON = lon
	obs.Hash = setHash(obs.Date, lat, lon)
	obs.SignalStrength, _ = strconv.Atoi(line[signalStrength])

	if len(p.batch) < p.batchSize {
		p.batch = append(p.batch, obs)
	}

	if len(p.batch) == p.batchSize {

		p.obs.WriteBatch(ctx, p.batch)
		p.batch = []observation.Observation{}
	}

}

func (p *PumpMaster) commitLast(ctx context.Context) {

	if len(p.batch) != 0 {

		p.obs.WriteBatch(ctx, p.batch)
	}

}

func parseDate(date string) time.Time {

	t, _ := time.Parse("02.01.2006 15:04:05", date)
	return t
}

func parseCoords(coords string) (float64, float64) {

	cds := strings.Split(coords, " ")

	lat, _ := strconv.ParseFloat(cds[0], 64)
	lon, _ := strconv.ParseFloat(cds[1], 64)

	return lat, lon
}

func parseIMEIorIMSI(code string) string {

	return strings.TrimSpace(code)
}

func setHash(date time.Time, lat float64, lon float64) string {
	s := fmt.Sprintf("%s§%.6f§%.6f", date.UTC().Format(time.RFC3339Nano), lat, lon)

	sum := sha256.Sum256([]byte(s))

	return fmt.Sprintf("%x", sum)
}
