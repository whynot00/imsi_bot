package parser

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

// ParseStream reads a semicolon-delimited source CSV from r and calls fn for
// every row. fn receives a *Device and/or *Location (one or both may be nil).
// This avoids loading the entire file into memory.
func ParseStream(r io.Reader, fn func(d *Device, l *Location) error) error {
	csvReader, err := newParametrReader(r)
	if err != nil {
		return err
	}

	headers, err := csvReader.Read()
	if err != nil {
		return fmt.Errorf("reading header: %w", err)
	}
	headers = sanitiseHeaders(headers)

	idx := buildIndex(headers)

	required := []string{"id", "date", "imei", "imsi", "event", "lat", "lon"}
	for _, f := range required {
		if _, ok := idx[f]; !ok {
			return fmt.Errorf("required column not found for field %q (check columnMap in types.go)", f)
		}
	}

	lineNum := 1
	for {
		lineNum++
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("line %d: %w", lineNum, err)
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

		var d *Device
		var l *Location

		if imsi != "" || (imei != "" && imei != "closed") {
			d = &Device{
				ID:       id,
				Date:     date,
				Standart: standart,
				Operator: operator,
				IMSI:     imsi,
				IMEI:     imei,
				Event:    event,
			}
		}

		if lat != "" && lat != "0" && lon != "" && lon != "0" {
			l = &Location{
				ID:   id,
				Date: date,
				Lat:  lat,
				Lon:  lon,
			}
		}

		if d != nil || l != nil {
			if err := fn(d, l); err != nil {
				return err
			}
		}
	}

	return nil
}

// Parse reads a semicolon-delimited source CSV from r and returns a
// ParseResult with two slices. It uses ParseStream internally.
func Parse(r io.Reader) (*ParseResult, error) {
	result := &ParseResult{}
	err := ParseStream(r, func(d *Device, l *Location) error {
		if d != nil {
			result.Devices = append(result.Devices, *d)
		}
		if l != nil {
			result.Locations = append(result.Locations, *l)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ParseRawStream reads a headerless semicolon-delimited EDM-format CSV and
// calls fn for each parsed row. This avoids loading the entire file into memory.
func ParseRawStream(r io.Reader, fn func(d RawDevice) error) error {
	csvReader, err := newRawReader(r)
	if err != nil {
		return err
	}

	lineNum := 0
	for {
		lineNum++
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("line %d: %w", lineNum, err)
		}
		if len(row) < 16 {
			continue
		}

		coords := strings.Fields(strings.TrimSpace(row[11]))
		var lat, lon string
		if len(coords) == 2 {
			lat = coords[0]
			lon = coords[1]
		}

		if err := fn(RawDevice{
			Date:           strings.TrimSpace(row[2]),
			Standart:       strings.TrimSpace(row[4]),
			Lat:            lat,
			Lon:            lon,
			SignalStrength: strings.TrimSpace(row[12]),
			IMSI:           strings.TrimSpace(row[14]),
			IMEI:           strings.TrimSpace(row[15]),
		}); err != nil {
			return err
		}
	}

	return nil
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

// crlfReplacer is a streaming reader that normalises \r\n and bare \r to \n.
type crlfReplacer struct {
	r    *bufio.Reader
	prev byte
}

func (c *crlfReplacer) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	j := 0
	for i := 0; i < n; i++ {
		b := p[i]
		if b == '\r' {
			p[j] = '\n'
			j++
		} else if b == '\n' && c.prev == '\r' {
			// skip \n after \r (already emitted \n)
		} else {
			p[j] = b
			j++
		}
		c.prev = b
	}
	return j, err
}

// newCSVReader wraps r with CRLF normalisation and returns a configured csv.Reader.
func newCSVReader(r io.Reader) *csv.Reader {
	cr := &crlfReplacer{r: bufio.NewReaderSize(r, 256*1024)}
	csvReader := csv.NewReader(cr)
	csvReader.Comma = ';'
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true
	csvReader.FieldsPerRecord = -1
	return csvReader
}

func newParametrReader(r io.Reader) (*csv.Reader, error) {
	return newCSVReader(r), nil
}

func newRawReader(r io.Reader) (*csv.Reader, error) {
	return newCSVReader(r), nil
}

// ParseRaw reads a headerless EDM-format CSV and returns a slice of RawDevice.
// It uses ParseRawStream internally.
func ParseRaw(r io.Reader) ([]RawDevice, error) {
	var result []RawDevice
	err := ParseRawStream(r, func(d RawDevice) error {
		result = append(result, d)
		return nil
	})
	if err != nil {
		return nil, err
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
