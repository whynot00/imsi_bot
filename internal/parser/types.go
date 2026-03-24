// Package parser reads semicolon-delimited СОРМ/IMSI-catcher CSV exports
// and converts them into a clean, normalised structure.
package parser

// columnMap maps source Russian headers to internal field names.
var columnMap = map[string]string{
	"Ключ":       "id",
	"Стандарт":   "standart",
	"Сеть":       "operator",
	"Дата/Время": "date",
	"LAC":        "lac",
	"CID":        "cid",
	"Широта_1":   "lat",
	"Долгота_1":  "lon",
	"Уровень":    "signal_strength",
	"IMEI источника (Номер терминала)": "imei",
	"IMSI источника (Сист. номер)":     "imsi",
	"Событие": "event",
}

// Device represents a registered subscriber — a row with IMSI or IMEI.
type Device struct {
	ID       string `json:"id"`
	Date     string `json:"date"`
	Standart string `json:"standart"`
	Operator string `json:"operator"`
	IMSI     string `json:"imsi"`
	IMEI     string `json:"imei"`
	Event    string `json:"event"`
}

// Location represents a device position ping — a row with coordinates.
type Location struct {
	ID   string `json:"id"`
	Date string `json:"date"`
	Lat  string `json:"lat"`
	Lon  string `json:"lon"`
}

// ParseResult is the top-level result returned by Parse.
type ParseResult struct {
	Devices   []Device
	Locations []Location
}

// RawDevice represents a row from the headerless EDM-format CSV.
type RawDevice struct {
	Date           string `json:"date"`
	Standart       string `json:"standart"`
	Lat            string `json:"lat"`
	Lon            string `json:"lon"`
	SignalStrength string `json:"signal_strength"`
	IMSI           string `json:"imsi"`
	IMEI           string `json:"imei"`
}
