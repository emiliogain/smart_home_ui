package sensor

import "errors"

// Domain errors for sensor operations
var (
	ErrInvalidID       = errors.New("invalid sensor ID")
	ErrInvalidName     = errors.New("invalid sensor name")
	ErrInvalidLocation = errors.New("invalid sensor location")
	ErrInvalidSensorID = errors.New("invalid sensor ID in data")
	ErrNoData          = errors.New("no sensor data provided")
	ErrInvalidQuality  = errors.New("quality must be between 0 and 1")
	ErrSensorNotFound  = errors.New("sensor not found")
	ErrDataNotFound    = errors.New("sensor data not found")
	ErrInvalidType     = errors.New("invalid sensor type")
)
