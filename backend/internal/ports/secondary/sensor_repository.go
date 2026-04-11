package secondary

import (
	"context"

	"github.com/emiliogain/smart-home-backend/internal/domain/sensor"
)

// SensorRepository defines persistence operations for sensors and their readings.
type SensorRepository interface {
	SaveSensor(ctx context.Context, s sensor.Sensor) error
	GetSensor(ctx context.Context, id string) (*sensor.Sensor, error)
	ListSensors(ctx context.Context) ([]sensor.Sensor, error)
	SaveReading(ctx context.Context, r sensor.Reading) error
	GetLatestReadings(ctx context.Context, sensorID string, limit int) ([]sensor.Reading, error)
	// GetAllLatestReadings returns the most recent readings across all sensors,
	// enriched with sensor type and location metadata.
	GetAllLatestReadings(ctx context.Context, perSensor int) ([]sensor.EnrichedReading, error)
}
