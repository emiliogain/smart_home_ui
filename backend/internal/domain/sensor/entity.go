package sensor

import "time"

// SensorType represents the kind of sensor.
type SensorType string

const (
	TypeTemperature SensorType = "temperature"
	TypeHumidity    SensorType = "humidity"
	TypePressure    SensorType = "pressure"
	TypeMotion      SensorType = "motion"
	TypeLight       SensorType = "light"
)

// Sensor represents a physical sensor registered in the system.
type Sensor struct {
	ID        string
	Name      string
	Type      SensorType
	Location  string
	Status    string // "active", "inactive"
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Reading is a single data point captured by a sensor.
type Reading struct {
	ID        string
	SensorID  string
	Value     float64
	Unit      string
	Timestamp time.Time
}
