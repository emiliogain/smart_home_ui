// Package sensor contains the core sensor domain logic
package sensor

import (
	"time"
)

// SensorType represents different types of sensors
type SensorType string

const (
	TypeTemperature SensorType = "temperature"
	TypeHumidity    SensorType = "humidity"
	TypePressure    SensorType = "pressure"
	TypeMotion      SensorType = "motion"
	TypeLight       SensorType = "light"
	TypeDoor        SensorType = "door"
	TypeWindow      SensorType = "window"
	TypeSmoke       SensorType = "smoke"
	TypeGas         SensorType = "gas"
	TypePower       SensorType = "power"
)

// Status represents the current status of a sensor
type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
	StatusError    Status = "error"
	StatusOffline  Status = "offline"
)

// Sensor represents the core sensor entity
type Sensor struct {
	id          string
	name        string
	sensorType  SensorType
	location    string
	description string
	status      Status
	lastSeen    *time.Time
	createdAt   time.Time
	updatedAt   time.Time
	config      Config
}

// Config holds sensor-specific configuration
type Config struct {
	ReportingInterval time.Duration
	Thresholds        map[string]interface{}
	Calibration       map[string]float64
	Enabled           bool
}

// Data represents data collected from a sensor
type Data struct {
	id        string
	sensorID  string
	timestamp time.Time
	values    map[string]interface{}
	quality   float64 // Data quality score 0-1
	createdAt time.Time
}

// Reading is a simplified view for real-time data
type Reading struct {
	SensorID  string
	Type      SensorType
	Location  string
	Timestamp time.Time
	Value     interface{}
	Unit      string
	Quality   float64
}

// NewSensor creates a new sensor with validation
func NewSensor(id, name string, sensorType SensorType, location string) (*Sensor, error) {
	if id == "" {
		return nil, ErrInvalidID
	}
	if name == "" {
		return nil, ErrInvalidName
	}
	if location == "" {
		return nil, ErrInvalidLocation
	}

	now := time.Now()
	return &Sensor{
		id:         id,
		name:       name,
		sensorType: sensorType,
		location:   location,
		status:     StatusActive,
		createdAt:  now,
		updatedAt:  now,
		config: Config{
			ReportingInterval: 30 * time.Second,
			Enabled:           true,
		},
	}, nil
}

// Getters (pure functions, no side effects)
func (s *Sensor) ID() string           { return s.id }
func (s *Sensor) Name() string         { return s.name }
func (s *Sensor) Type() SensorType     { return s.sensorType }
func (s *Sensor) Location() string     { return s.location }
func (s *Sensor) Description() string  { return s.description }
func (s *Sensor) Status() Status       { return s.status }
func (s *Sensor) LastSeen() *time.Time { return s.lastSeen }
func (s *Sensor) CreatedAt() time.Time { return s.createdAt }
func (s *Sensor) UpdatedAt() time.Time { return s.updatedAt }
func (s *Sensor) Config() Config       { return s.config }

// UpdateName updates the sensor name
func (s *Sensor) UpdateName(name string) error {
	if name == "" {
		return ErrInvalidName
	}
	s.name = name
	s.updatedAt = time.Now()
	return nil
}

// UpdateLocation updates the sensor location
func (s *Sensor) UpdateLocation(location string) error {
	if location == "" {
		return ErrInvalidLocation
	}
	s.location = location
	s.updatedAt = time.Now()
	return nil
}

// UpdateDescription updates the sensor description
func (s *Sensor) UpdateDescription(description string) {
	s.description = description
	s.updatedAt = time.Now()
}

// UpdateStatus updates the sensor status
func (s *Sensor) UpdateStatus(status Status) {
	s.status = status
	s.updatedAt = time.Now()
	if status == StatusActive {
		now := time.Now()
		s.lastSeen = &now
	}
}

// UpdateConfig updates the sensor configuration
func (s *Sensor) UpdateConfig(config Config) {
	s.config = config
	s.updatedAt = time.Now()
}

// MarkAsSeen updates the last seen timestamp
func (s *Sensor) MarkAsSeen() {
	now := time.Now()
	s.lastSeen = &now
	s.updatedAt = now
}

// IsOnline checks if the sensor is considered online
func (s *Sensor) IsOnline(timeout time.Duration) bool {
	if s.lastSeen == nil {
		return false
	}
	return time.Since(*s.lastSeen) <= timeout
}

// NewData creates new sensor data with validation
func NewData(id, sensorID string, values map[string]interface{}) (*Data, error) {
	if id == "" {
		return nil, ErrInvalidID
	}
	if sensorID == "" {
		return nil, ErrInvalidSensorID
	}
	if len(values) == 0 {
		return nil, ErrNoData
	}

	now := time.Now()
	return &Data{
		id:        id,
		sensorID:  sensorID,
		timestamp: now,
		values:    values,
		quality:   1.0, // Default to perfect quality
		createdAt: now,
	}, nil
}

// Getters for Data
func (d *Data) ID() string                     { return d.id }
func (d *Data) SensorID() string               { return d.sensorID }
func (d *Data) Timestamp() time.Time           { return d.timestamp }
func (d *Data) Values() map[string]interface{} { return d.values }
func (d *Data) Quality() float64               { return d.quality }
func (d *Data) CreatedAt() time.Time           { return d.createdAt }

// SetQuality sets the data quality score (0-1)
func (d *Data) SetQuality(quality float64) error {
	if quality < 0 || quality > 1 {
		return ErrInvalidQuality
	}
	d.quality = quality
	return nil
}

// SetTimestamp allows setting a custom timestamp for historical data
func (d *Data) SetTimestamp(timestamp time.Time) {
	d.timestamp = timestamp
}

// ReconstructSensor restores a sensor from persistence (e.g. PostgreSQL).
func ReconstructSensor(
	id, name string,
	sensorType SensorType,
	location, description string,
	status Status,
	lastSeen *time.Time,
	createdAt, updatedAt time.Time,
	config Config,
) *Sensor {
	return &Sensor{
		id:          id,
		name:        name,
		sensorType:  sensorType,
		location:    location,
		description: description,
		status:      status,
		lastSeen:    lastSeen,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		config:      config,
	}
}

// ReconstructData restores sensor data from persistence.
func ReconstructData(
	id, sensorID string,
	timestamp time.Time,
	values map[string]interface{},
	quality float64,
	createdAt time.Time,
) *Data {
	if values == nil {
		values = map[string]interface{}{}
	}
	return &Data{
		id:        id,
		sensorID:  sensorID,
		timestamp: timestamp,
		values:    values,
		quality:   quality,
		createdAt: createdAt,
	}
}
