package observation

import "time"

type Observation struct {
	Standart       string    `db:"standart"`
	Operator       string    `db:"operator"`
	Date           time.Time `db:"date"`
	LON            float64   `db:"lon"`
	LAT            float64   `db:"lat"`
	SignalStrength int       `db:"signal_strength"`
	IMEI           string    `db:"imei"`
	IMSI           string    `db:"imsi"`
	Hash           string    `db:"hash"`
}
