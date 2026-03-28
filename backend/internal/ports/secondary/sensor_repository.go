// Package secondary defines the outgoing ports (secondary adapters implement these)
package secondary

import (
	"context"
	"time"

	"github.com/emiliogain/smart-home-backend/internal/domain/sensor"
)

// SensorRepository defines the secondary port for sensor persistence
// Database adapters will implement this interface
type SensorRepository interface {
	// Sensor persistence
	Save(ctx context.Context, sensor *sensor.Sensor) error
	FindByID(ctx context.Context, id string) (*sensor.Sensor, error)
	FindAll(ctx context.Context) ([]*sensor.Sensor, error)
	FindByType(ctx context.Context, sensorType sensor.SensorType) ([]*sensor.Sensor, error)
	FindByLocation(ctx context.Context, location string) ([]*sensor.Sensor, error)
	Update(ctx context.Context, sensor *sensor.Sensor) error
	Delete(ctx context.Context, id string) error
	Exists(ctx context.Context, id string) (bool, error)

	// Sensor data persistence
	SaveData(ctx context.Context, data *sensor.Data) error
	FindDataByID(ctx context.Context, id string) (*sensor.Data, error)
	FindDataBySensor(ctx context.Context, sensorID string, from, to time.Time) ([]*sensor.Data, error)
	FindLatestData(ctx context.Context, sensorID string) (*sensor.Data, error)
	FindDataByLocation(ctx context.Context, location string, from, to time.Time) ([]*sensor.Data, error)

	// Data aggregation and analytics
	FindAggregatedData(ctx context.Context, sensorID string, from, to time.Time, interval time.Duration) ([]*sensor.Data, error)
	GetDataCount(ctx context.Context, sensorID string, from, to time.Time) (int64, error)

	// Data maintenance
	DeleteOldData(ctx context.Context, before time.Time) error
	CalculateStorageSize(ctx context.Context, sensorID string) (int64, error)
}

// SensorEventPublisher defines the port for publishing sensor events
// Message queue adapters will implement this interface
type SensorEventPublisher interface {
	PublishSensorCreated(ctx context.Context, sensor *sensor.Sensor) error
	PublishSensorUpdated(ctx context.Context, sensor *sensor.Sensor) error
	PublishSensorDeleted(ctx context.Context, sensorID string) error
	PublishDataReceived(ctx context.Context, data *sensor.Data) error
	PublishAlert(ctx context.Context, sensorID string, alertType string, message string) error
}
