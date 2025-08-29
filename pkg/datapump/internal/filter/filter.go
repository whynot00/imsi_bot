package filter

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/whynot00/imsi_bot/pkg/datapump/models"
)

const insertEntity = `
		INSERT INTO entities (standart, operator, date, lon, lat, signal_strength, imei, imsi, hash)
		VALUES (:standart, :operator, :date, :lon, :lat, :signal_strength, :imei, :imsi, :hash)
		ON CONFLICT (hash) DO NOTHING;`

func Parse(lineCh <-chan []string, batchSize int, pg *sqlx.DB) {

	batch := make([]models.Entity, 0, batchSize)

	for line := range lineCh {
		if line[models.Coords] == "" {
			continue
		}

		entity := models.Entity{
			Standart: line[models.Standart],
			Operator: models.IMSIOperators[line[models.OpertorCode]],
			Date:     parseDate(line[models.Date]),
			IMEI:     parseIMEIorIMSI(line[models.IMEI]),
			IMSI:     parseIMEIorIMSI(line[models.IMSI]),
		}

		lat, lon := parseCoords(line[models.Coords])

		entity.LAT = lat
		entity.LON = lon
		entity.Hash = setHash(entity.Date, lat, lon)
		entity.SignalStrength, _ = strconv.Atoi(line[models.SignalStrength])

		if len(batch) < batchSize {
			batch = append(batch, entity)
		}

		if len(batch) == batchSize {

			if err := writeBatch(pg, batch); err != nil {
				continue
			}

			batch = []models.Entity{}
		}

	}

	if len(batch) != 0 {
		if err := writeBatch(pg, batch); err != nil {

		}
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

func writeBatch(db *sqlx.DB, entities []models.Entity) error {

	_, err := db.NamedExec(insertEntity, entities)
	return err

}
