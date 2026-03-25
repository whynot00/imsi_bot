// Package repo provides PostgreSQL repositories for storing and querying
// parsed IMSI-catcher data.
package repo

import "time"

// Device is the DB representation of a unique subscriber device.
type Device struct {
	ID   int64
	IMSI string
	IMEI string
}

// LocationParametr is a coordinate ping from a parametr-format file.
type LocationParametr struct {
	ID     int64
	SeenAt time.Time
	Lat    float64
	Lon    float64
}

// SightingParametr is a subscriber observation from a parametr-format file.
type SightingParametr struct {
	ID         int64
	DeviceID   int64
	SeenAt     time.Time
	Standart   string
	Operator   string
	Event      string
	LocationID *int64
	Lat        *float64
	Lon        *float64
}

// SightingRK is a subscriber observation from an rk-format file.
type SightingRK struct {
	ID       int64
	DeviceID int64
	SeenAt   time.Time
	Standart string
	Lat      float64
	Lon      float64
	Signal   int
}

// DeviceResult is returned by IMSI/IMEI search queries and joins
// all sightings for a device across both source types.
type DeviceResult struct {
	Device            Device
	SightingsParametr []SightingParametr
	SightingsRK       []SightingRK
}
