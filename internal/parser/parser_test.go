package parser

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// minimal valid header with all columns the parser requires.
const header = "Ключ;Категория;Дата/Время;Стандарт;Сеть;LAC;CID;IMSI источника (Сист. номер);IMEI источника (Номер терминала);Широта_1;Долгота_1;Событие"

func idx(s string) int {
	cols := strings.Split(header, ";")
	for i, c := range cols {
		if c == s {
			return i
		}
	}
	return -1
}

// row builds a semicolon-delimited row aligned to the header above.
// Pass a map of column name → value; missing columns default to "".
func row(fields map[string]string) string {
	cols := strings.Split(header, ";")
	vals := make([]string, len(cols))
	for i, c := range cols {
		vals[i] = fields[c]
	}
	return strings.Join(vals, ";")
}

func mustParse(t *testing.T, csv string) *ParseResult {
	t.Helper()
	r, err := Parse(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}
	return r
}

// ---------------------------------------------------------------------------
// empty / degenerate input
// ---------------------------------------------------------------------------

func TestParse_EmptyFile(t *testing.T) {
	_, err := Parse(strings.NewReader(""))
	if err == nil {
		t.Fatal("expected error on empty input, got nil")
	}
}

func TestParse_HeaderOnly(t *testing.T) {
	r := mustParse(t, header)
	if len(r.Devices) != 0 {
		t.Errorf("devices: want 0, got %d", len(r.Devices))
	}
	if len(r.Locations) != 0 {
		t.Errorf("locations: want 0, got %d", len(r.Locations))
	}
}

// ---------------------------------------------------------------------------
// device detection
// ---------------------------------------------------------------------------

func TestParse_DeviceWithIMSI(t *testing.T) {
	csv := header + "\n" + row(map[string]string{
		"Ключ":       "1",
		"Дата/Время": "07.10.2025 11:50:41",
		"Стандарт":   "UMTS",
		"Сеть":       "МТС : Россия",
		"IMSI источника (Сист. номер)":     "250018707381530",
		"IMEI источника (Номер терминала)": "closed",
		"Событие": "Регистрация",
	})
	r := mustParse(t, csv)

	if len(r.Devices) != 1 {
		t.Fatalf("devices: want 1, got %d", len(r.Devices))
	}
	d := r.Devices[0]
	if d.IMSI != "250018707381530" {
		t.Errorf("IMSI: want 250018707381530, got %q", d.IMSI)
	}
	if d.Standart != "UMTS" {
		t.Errorf("Standart: want UMTS, got %q", d.Standart)
	}
	if d.Operator != "МТС : Россия" {
		t.Errorf("Operator: want МТС : Россия, got %q", d.Operator)
	}
	if d.Event != "Регистрация" {
		t.Errorf("Event: want Регистрация, got %q", d.Event)
	}
}

func TestParse_DeviceWithIMEIOnly(t *testing.T) {
	csv := header + "\n" + row(map[string]string{
		"Ключ":       "2",
		"Дата/Время": "07.10.2025 11:51:00",
		"Стандарт":   "LTE",
		"IMEI источника (Номер терминала)": "353312113249201",
	})
	r := mustParse(t, csv)

	if len(r.Devices) != 1 {
		t.Fatalf("devices: want 1, got %d", len(r.Devices))
	}
	if r.Devices[0].IMEI != "353312113249201" {
		t.Errorf("IMEI: want 353312113249201, got %q", r.Devices[0].IMEI)
	}
}

func TestParse_ClosedIMEI_WithoutIMSI_Skipped(t *testing.T) {
	// IMEI="closed" and no IMSI → should NOT appear in Devices
	csv := header + "\n" + row(map[string]string{
		"Ключ":       "3",
		"Дата/Время": "07.10.2025 11:51:00",
		"IMEI источника (Номер терминала)": "closed",
	})
	r := mustParse(t, csv)
	if len(r.Devices) != 0 {
		t.Errorf("devices: want 0 (closed IMEI + no IMSI), got %d", len(r.Devices))
	}
}

func TestParse_ClosedIMEI_WithIMSI_Kept(t *testing.T) {
	// IMEI="closed" but IMSI present → should appear in Devices
	csv := header + "\n" + row(map[string]string{
		"Ключ":       "4",
		"Дата/Время": "07.10.2025 11:51:00",
		"IMSI источника (Сист. номер)":     "250018707381530",
		"IMEI источника (Номер терминала)": "closed",
	})
	r := mustParse(t, csv)
	if len(r.Devices) != 1 {
		t.Fatalf("devices: want 1, got %d", len(r.Devices))
	}
}

func TestParse_NoIMSINoIMEI_Skipped(t *testing.T) {
	csv := header + "\n" + row(map[string]string{
		"Ключ":       "5",
		"Дата/Время": "07.10.2025 11:35:43",
		"Широта_1":   "56.2818832397461",
		"Долгота_1":  "43.940975189209",
	})
	r := mustParse(t, csv)
	if len(r.Devices) != 0 {
		t.Errorf("devices: want 0, got %d", len(r.Devices))
	}
}

// ---------------------------------------------------------------------------
// location detection
// ---------------------------------------------------------------------------

func TestParse_LocationFromCoordinates(t *testing.T) {
	csv := header + "\n" + row(map[string]string{
		"Ключ":       "6",
		"Дата/Время": "07.10.2025 11:35:43",
		"Широта_1":   "56.2818832397461",
		"Долгота_1":  "43.940975189209",
	})
	r := mustParse(t, csv)

	if len(r.Locations) != 1 {
		t.Fatalf("locations: want 1, got %d", len(r.Locations))
	}
	l := r.Locations[0]
	if l.Lat != "56.2818832397461" {
		t.Errorf("Lat: want 56.2818832397461, got %q", l.Lat)
	}
	if l.Lon != "43.940975189209" {
		t.Errorf("Lon: want 43.940975189209, got %q", l.Lon)
	}
	if l.Date != "07.10.2025 11:35:43" {
		t.Errorf("Date: want 07.10.2025 11:35:43, got %q", l.Date)
	}
}

func TestParse_ZeroCoordinates_Skipped(t *testing.T) {
	csv := header + "\n" + row(map[string]string{
		"Ключ":      "7",
		"Широта_1":  "0",
		"Долгота_1": "0",
	})
	r := mustParse(t, csv)
	if len(r.Locations) != 0 {
		t.Errorf("locations: want 0 for zero coords, got %d", len(r.Locations))
	}
}

func TestParse_EmptyCoordinates_Skipped(t *testing.T) {
	csv := header + "\n" + row(map[string]string{
		"Ключ": "8",
	})
	r := mustParse(t, csv)
	if len(r.Locations) != 0 {
		t.Errorf("locations: want 0 for empty coords, got %d", len(r.Locations))
	}
}

// ---------------------------------------------------------------------------
// row appears in both slices
// ---------------------------------------------------------------------------

func TestParse_RowInBothSlices(t *testing.T) {
	// A row that has IMSI AND coordinates should appear in both Devices and Locations.
	csv := header + "\n" + row(map[string]string{
		"Ключ":       "9",
		"Дата/Время": "07.10.2025 12:00:00",
		"Стандарт":   "LTE",
		"IMSI источника (Сист. номер)": "250018706636519",
		"Широта_1":  "56.2818603515625",
		"Долгота_1": "43.94091796875",
	})
	r := mustParse(t, csv)

	if len(r.Devices) != 1 {
		t.Errorf("devices: want 1, got %d", len(r.Devices))
	}
	if len(r.Locations) != 1 {
		t.Errorf("locations: want 1, got %d", len(r.Locations))
	}
}

// ---------------------------------------------------------------------------
// multiple rows
// ---------------------------------------------------------------------------

func TestParse_MultipleRows(t *testing.T) {
	rows := []string{
		header,
		// device only
		row(map[string]string{
			"Ключ":       "10",
			"Дата/Время": "07.10.2025 11:50:41",
			"Стандарт":   "UMTS",
			"IMSI источника (Сист. номер)": "250018707381530",
			"Событие": "Регистрация",
		}),
		// location only
		row(map[string]string{
			"Ключ":       "11",
			"Дата/Время": "07.10.2025 11:35:43",
			"Широта_1":   "56.2818832397461",
			"Долгота_1":  "43.940975189209",
		}),
		// both
		row(map[string]string{
			"Ключ":       "12",
			"Дата/Время": "07.10.2025 12:00:00",
			"Стандарт":   "LTE",
			"IMSI источника (Сист. номер)": "250018706636519",
			"Широта_1":  "56.2818603515625",
			"Долгота_1": "43.94091796875",
		}),
		// neither — skipped in both
		row(map[string]string{
			"Ключ":       "13",
			"Дата/Время": "07.10.2025 11:36:00",
		}),
	}

	r := mustParse(t, strings.Join(rows, "\n"))

	if len(r.Devices) != 2 {
		t.Errorf("devices: want 2, got %d", len(r.Devices))
	}
	if len(r.Locations) != 2 {
		t.Errorf("locations: want 2, got %d", len(r.Locations))
	}
}

// ---------------------------------------------------------------------------
// line ending normalisation
// ---------------------------------------------------------------------------

func TestParse_CRLF(t *testing.T) {
	csv := strings.ReplaceAll(
		header+"\n"+row(map[string]string{
			"Ключ": "14",
			"IMSI источника (Сист. номер)": "250018707381530",
		}),
		"\n", "\r\n",
	)
	r := mustParse(t, csv)
	if len(r.Devices) != 1 {
		t.Errorf("devices: want 1 after CRLF normalisation, got %d", len(r.Devices))
	}
}

func TestParse_BareCR(t *testing.T) {
	csv := strings.ReplaceAll(
		header+"\n"+row(map[string]string{
			"Ключ": "15",
			"IMSI источника (Сист. номер)": "250011781682986",
		}),
		"\n", "\r",
	)
	r := mustParse(t, csv)
	if len(r.Devices) != 1 {
		t.Errorf("devices: want 1 after bare CR normalisation, got %d", len(r.Devices))
	}
}

// ---------------------------------------------------------------------------
// BOM stripping
// ---------------------------------------------------------------------------

func TestParse_UTF8BOM(t *testing.T) {
	// Prepend UTF-8 BOM to the header.
	csv := "\xef\xbb\xbf" + header + "\n" + row(map[string]string{
		"Ключ": "16",
		"IMSI источника (Сист. номер)": "250014887172632",
	})
	r := mustParse(t, csv)
	if len(r.Devices) != 1 {
		t.Fatalf("devices: want 1 after BOM strip, got %d", len(r.Devices))
	}
	if r.Devices[0].IMSI != "250014887172632" {
		t.Errorf("IMSI: want 250014887172632, got %q", r.Devices[0].IMSI)
	}
}

// ---------------------------------------------------------------------------
// field count variance
// ---------------------------------------------------------------------------

func TestParse_FewerFieldsThanHeader(t *testing.T) {
	// Row with fewer fields than the header must not crash.
	csv := header + "\n" + "17;Служебное сообщение"
	r := mustParse(t, csv)
	// No IMSI/IMEI, no coords → both slices empty, but no panic.
	if len(r.Devices) != 0 || len(r.Locations) != 0 {
		t.Errorf("want empty result for short row, got devices=%d locations=%d",
			len(r.Devices), len(r.Locations))
	}
}

// ---------------------------------------------------------------------------
// Write round-trip
// ---------------------------------------------------------------------------

func TestWriteDevices_RoundTrip(t *testing.T) {
	csv := header + "\n" + row(map[string]string{
		"Ключ":       "18",
		"Дата/Время": "07.10.2025 11:50:41",
		"Стандарт":   "GSM",
		"Сеть":       "МТС : Россия",
		"IMSI источника (Сист. номер)":     "250019124314298",
		"IMEI источника (Номер терминала)": "869573071376987",
		"Событие": "Запрос",
	})
	r := mustParse(t, csv)

	var buf strings.Builder
	if err := WriteDevices(&buf, r.Devices); err != nil {
		t.Fatalf("WriteDevices error: %v", err)
	}
	out := buf.String()

	for _, want := range []string{
		"id;date;standart;operator;imsi;imei;event",
		"250019124314298",
		"869573071376987",
		"GSM",
		"Запрос",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\ngot:\n%s", want, out)
		}
	}
}

func TestWriteLocations_RoundTrip(t *testing.T) {
	csv := header + "\n" + row(map[string]string{
		"Ключ":       "19",
		"Дата/Время": "07.10.2025 11:35:43",
		"Широта_1":   "56.2818832397461",
		"Долгота_1":  "43.940975189209",
	})
	r := mustParse(t, csv)

	var buf strings.Builder
	if err := WriteLocations(&buf, r.Locations); err != nil {
		t.Fatalf("WriteLocations error: %v", err)
	}
	out := buf.String()

	for _, want := range []string{
		"id;date;lat;lon",
		"56.2818832397461",
		"43.940975189209",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\ngot:\n%s", want, out)
		}
	}
}

// ---------------------------------------------------------------------------
// field values preserved exactly
// ---------------------------------------------------------------------------

func TestParse_FieldValuesPreserved(t *testing.T) {
	csv := header + "\n" + row(map[string]string{
		"Ключ":       "20",
		"Дата/Время": "19.12.2025 09:15:00",
		"Стандарт":   "LTE",
		"Сеть":       "Билайн : Россия",
		"IMSI источника (Сист. номер)":     "250990000000001",
		"IMEI источника (Номер терминала)": "490154203237518",
		"Событие":   "Обновление местоположения",
		"Широта_1":  "55.7558",
		"Долгота_1": "37.6173",
	})
	r := mustParse(t, csv)

	if len(r.Devices) != 1 {
		t.Fatalf("devices: want 1, got %d", len(r.Devices))
	}
	d := r.Devices[0]
	checks := map[string]string{
		"ID":       d.ID,
		"Date":     d.Date,
		"Standart": d.Standart,
		"Operator": d.Operator,
		"IMSI":     d.IMSI,
		"IMEI":     d.IMEI,
		"Event":    d.Event,
	}
	wants := map[string]string{
		"ID":       "20",
		"Date":     "19.12.2025 09:15:00",
		"Standart": "LTE",
		"Operator": "Билайн : Россия",
		"IMSI":     "250990000000001",
		"IMEI":     "490154203237518",
		"Event":    "Обновление местоположения",
	}
	for field, got := range checks {
		if got != wants[field] {
			t.Errorf("%s: want %q, got %q", field, wants[field], got)
		}
	}

	if len(r.Locations) != 1 {
		t.Fatalf("locations: want 1, got %d", len(r.Locations))
	}
	l := r.Locations[0]
	if l.Lat != "55.7558" {
		t.Errorf("Lat: want 55.7558, got %q", l.Lat)
	}
	if l.Lon != "37.6173" {
		t.Errorf("Lon: want 37.6173, got %q", l.Lon)
	}
}

// ---------------------------------------------------------------------------
// ParseRaw tests
// ---------------------------------------------------------------------------

const rawHeader = "" // no header in EDM format

// rawRow builds a 28-field semicolon row matching the EDM format.
// Only the fields we care about are filled; rest are empty.
func rawRow(date, standart, coords, signal, imsi, imei string) string {
	cols := make([]string, 28)
	cols[2] = date     // col 3
	cols[4] = standart // col 5
	cols[11] = coords  // col 12
	cols[12] = signal  // col 13
	cols[14] = imsi    // col 15
	cols[15] = imei    // col 16
	return strings.Join(cols, ";")
}

func mustParseRaw(t *testing.T, csv string) []RawDevice {
	t.Helper()
	r, err := ParseRaw(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("ParseRaw() unexpected error: %v", err)
	}
	return r
}

func TestParseRaw_BasicRow(t *testing.T) {
	csv := rawRow("29.08.2025 10:53:03", "2G", "56.312275 44.002998", "9", "250996527346851", "863591024427800")
	devices := mustParseRaw(t, csv)

	if len(devices) != 1 {
		t.Fatalf("want 1 device, got %d", len(devices))
	}
	d := devices[0]
	if d.Date != "29.08.2025 10:53:03" {
		t.Errorf("Date: want 29.08.2025 10:53:03, got %q", d.Date)
	}
	if d.Standart != "2G" {
		t.Errorf("Standart: want 2G, got %q", d.Standart)
	}
	if d.Lat != "56.312275" {
		t.Errorf("Lat: want 56.312275, got %q", d.Lat)
	}
	if d.Lon != "44.002998" {
		t.Errorf("Lon: want 44.002998, got %q", d.Lon)
	}
	if d.SignalStrength != "9" {
		t.Errorf("SignalStrength: want 9, got %q", d.SignalStrength)
	}
	if d.IMSI != "250996527346851" {
		t.Errorf("IMSI: want 250996527346851, got %q", d.IMSI)
	}
	if d.IMEI != "863591024427800" {
		t.Errorf("IMEI: want 863591024427800, got %q", d.IMEI)
	}
}

func TestParseRaw_MultipleRows(t *testing.T) {
	rows := strings.Join([]string{
		rawRow("29.08.2025 10:53:03", "2G", "56.312275 44.002998", "9", "250996527346851", "863591024427800"),
		rawRow("29.08.2025 10:53:04", "2G", "56.312275 44.002998", "11", "250996675057559", "864502035906230"),
		rawRow("29.08.2025 10:53:05", "3G", "56.312275 44.002998", "10", "250990285323826", "866557056021930"),
	}, "\n")

	devices := mustParseRaw(t, rows)
	if len(devices) != 3 {
		t.Fatalf("want 3 devices, got %d", len(devices))
	}
	if devices[2].Standart != "3G" {
		t.Errorf("row 3 Standart: want 3G, got %q", devices[2].Standart)
	}
}

func TestParseRaw_ShortRowSkipped(t *testing.T) {
	// Row with fewer than 16 fields must be skipped silently.
	csv := "EDM;;29.08.2025 10:53:03"
	devices := mustParseRaw(t, csv)
	if len(devices) != 0 {
		t.Errorf("want 0 for short row, got %d", len(devices))
	}
}

func TestParseRaw_EmptyCoords(t *testing.T) {
	csv := rawRow("29.08.2025 10:53:03", "2G", "", "9", "250996527346851", "863591024427800")
	devices := mustParseRaw(t, csv)
	if len(devices) != 1 {
		t.Fatalf("want 1, got %d", len(devices))
	}
	if devices[0].Lat != "" || devices[0].Lon != "" {
		t.Errorf("want empty lat/lon, got %q %q", devices[0].Lat, devices[0].Lon)
	}
}

func TestParseRaw_CRLF(t *testing.T) {
	rows := strings.ReplaceAll(
		strings.Join([]string{
			rawRow("29.08.2025 10:53:03", "2G", "56.312275 44.002998", "9", "250996527346851", "863591024427800"),
			rawRow("29.08.2025 10:53:04", "2G", "56.312275 44.002998", "11", "250996675057559", "864502035906230"),
		}, "\n"),
		"\n", "\r\n",
	)
	devices := mustParseRaw(t, rows)
	if len(devices) != 2 {
		t.Errorf("want 2 after CRLF normalisation, got %d", len(devices))
	}
}

func TestParseRaw_EmptyInput(t *testing.T) {
	devices := mustParseRaw(t, "")
	if len(devices) != 0 {
		t.Errorf("want 0 for empty input, got %d", len(devices))
	}
}

func TestWriteRawDevices_RoundTrip(t *testing.T) {
	csv := rawRow("29.08.2025 10:53:03", "2G", "56.312275 44.002998", "9", "250996527346851", "863591024427800")
	devices := mustParseRaw(t, csv)

	var buf strings.Builder
	if err := WriteRawDevices(&buf, devices); err != nil {
		t.Fatalf("WriteRawDevices error: %v", err)
	}
	out := buf.String()

	for _, want := range []string{
		"date;standart;lat;lon;signal_strength;imsi;imei",
		"56.312275",
		"44.002998",
		"250996527346851",
		"863591024427800",
		"2G",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\ngot:\n%s", want, out)
		}
	}
}
