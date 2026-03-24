package parser

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

// Parse reads a semicolon-delimited source CSV from r and returns a
// ParseResult with two slices:
//   - Devices   — rows that have IMSI or IMEI (registered subscribers)
//   - Locations — rows that have coordinates (device position pings)
func Parse(r io.Reader) (*ParseResult, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}
	raw = bytes.ReplaceAll(raw, []byte("\r\n"), []byte("\n"))
	raw = bytes.ReplaceAll(raw, []byte("\r"), []byte("\n"))

	csvReader := csv.NewReader(bytes.NewReader(raw))
	csvReader.Comma = ';'
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true
	csvReader.FieldsPerRecord = -1

	headers, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("reading header: %w", err)
	}
	headers = sanitiseHeaders(headers)

	idx := buildIndex(headers)

	required := []string{"id", "date", "imei", "imsi", "event", "lat", "lon"}
	for _, f := range required {
		if _, ok := idx[f]; !ok {
			return nil, fmt.Errorf("required column not found for field %q (check columnMap in types.go)", f)
		}
	}

	result := &ParseResult{}

	lineNum := 1
	for {
		lineNum++
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}

		id := cell(row, idx["id"])
		date := cell(row, idx["date"])
		imsi := cell(row, idx["imsi"])
		imei := cell(row, idx["imei"])
		lat := cell(row, idx["lat"])
		lon := cell(row, idx["lon"])
		standart := cell(row, idx["standart"])
		operator := cell(row, idx["operator"])
		event := cell(row, idx["event"])

		if imsi != "" || (imei != "" && imei != "closed") {
			result.Devices = append(result.Devices, Device{
				ID:       id,
				Date:     date,
				Standart: standart,
				Operator: operator,
				IMSI:     imsi,
				IMEI:     imei,
				Event:    event,
			})
		}

		if lat != "" && lat != "0" && lon != "" && lon != "0" {
			result.Locations = append(result.Locations, Location{
				ID:   id,
				Date: date,
				Lat:  lat,
				Lon:  lon,
			})
		}
	}

	return result, nil
}

// WriteDevices serialises devices as semicolon-delimited CSV to w.
func WriteDevices(w io.Writer, devices []Device) error {
	cw := csv.NewWriter(w)
	cw.Comma = ';'
	_ = cw.Write([]string{"id", "date", "standart", "operator", "imsi", "imei", "event"})
	for _, d := range devices {
		_ = cw.Write([]string{d.ID, d.Date, d.Standart, d.Operator, d.IMSI, d.IMEI, d.Event})
	}
	cw.Flush()
	return cw.Error()
}

// WriteLocations serialises locations as semicolon-delimited CSV to w.
func WriteLocations(w io.Writer, locations []Location) error {
	cw := csv.NewWriter(w)
	cw.Comma = ';'
	_ = cw.Write([]string{"id", "date", "lat", "lon"})
	for _, l := range locations {
		_ = cw.Write([]string{l.ID, l.Date, l.Lat, l.Lon})
	}
	cw.Flush()
	return cw.Error()
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func sanitiseHeaders(hdrs []string) []string {
	out := make([]string, len(hdrs))
	for i, h := range hdrs {
		h = strings.ReplaceAll(h, "\uFEFF", "")
		h = strings.TrimPrefix(h, "\xef\xbb\xbf")
		h = strings.TrimSpace(h)
		out[i] = h
	}
	return out
}

func buildIndex(hdrs []string) map[string]int {
	idx := make(map[string]int, len(columnMap))
	for i, h := range hdrs {
		if field, ok := columnMap[h]; ok {
			idx[field] = i
		}
	}
	return idx
}

func cell(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

// ParseRaw reads a headerless semicolon-delimited EDM-format CSV and returns
// a slice of RawDevice. Each row is expected to have at least 16 fields.
//
// Column mapping (1-based):
//
//	3  — date/time
//	5  — standart (2G/3G/4G)
//	12 — coordinates "lat lon" (space-separated)
//	13 — signal strength
//	15 — IMSI
//	16 — IMEI
func ParseRaw(r io.Reader) ([]RawDevice, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}
	raw = bytes.ReplaceAll(raw, []byte("\r\n"), []byte("\n"))
	raw = bytes.ReplaceAll(raw, []byte("\r"), []byte("\n"))

	csvReader := csv.NewReader(bytes.NewReader(raw))
	csvReader.Comma = ';'
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true
	csvReader.FieldsPerRecord = -1

	var result []RawDevice
	lineNum := 0
	for {
		lineNum++
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}
		if len(row) < 16 {
			continue // skip malformed rows
		}

		coords := strings.Fields(strings.TrimSpace(row[11])) // col 12, 0-based=11
		var lat, lon string
		if len(coords) == 2 {
			lat = coords[0]
			lon = coords[1]
		}

		result = append(result, RawDevice{
			Date:           strings.TrimSpace(row[2]), // col 3
			Standart:       strings.TrimSpace(row[4]), // col 5
			Lat:            lat,
			Lon:            lon,
			SignalStrength: strings.TrimSpace(row[12]), // col 13
			IMSI:           strings.TrimSpace(row[14]), // col 15
			IMEI:           strings.TrimSpace(row[15]), // col 16
		})
	}

	return result, nil
}

// WriteRawDevices serialises RawDevice slice as semicolon-delimited CSV to w.
func WriteRawDevices(w io.Writer, devices []RawDevice) error {
	cw := csv.NewWriter(w)
	cw.Comma = ';'
	_ = cw.Write([]string{"date", "standart", "lat", "lon", "signal_strength", "imsi", "imei"})
	for _, d := range devices {
		_ = cw.Write([]string{d.Date, d.Standart, d.Lat, d.Lon, d.SignalStrength, d.IMSI, d.IMEI})
	}
	cw.Flush()
	return cw.Error()
}
