// Package primary defines the incoming ports (primary adapters drive these)
package primary

import (
	"context"
	"time"

	"github.com/emiliogain/smart-home-backend/internal/domain/sensor"
)

// SensorService defines the primary port for sensor operations
// This is what the HTTP handlers and other primary adapters will use
type SensorService interface {
	// Sensor lifecycle
	CreateSensor(ctx context.Context, req CreateSensorRequest) (*sensor.Sensor, error)
	GetSensor(ctx context.Context, id string) (*sensor.Sensor, error)
	GetAllSensors(ctx context.Context) ([]*sensor.Sensor, error)
	GetSensorsByType(ctx context.Context, sensorType sensor.SensorType) ([]*sensor.Sensor, error)
	GetSensorsByLocation(ctx context.Context, location string) ([]*sensor.Sensor, error)
	UpdateSensor(ctx context.Context, id string, req UpdateSensorRequest) (*sensor.Sensor, error)
	DeleteSensor(ctx context.Context, id string) error

	// Sensor data operations
	SubmitSensorData(ctx context.Context, sensorID string, req SubmitSensorDataRequest) (*sensor.Data, error)
	GetSensorData(ctx context.Context, sensorID string, from, to time.Time) ([]*sensor.Data, error)
	GetLatestSensorData(ctx context.Context, sensorID string) (*sensor.Data, error)
	GetSensorDataByLocation(ctx context.Context, location string, from, to time.Time) ([]*sensor.Data, error)

	// Analytics and aggregation
	GetSensorReadings(ctx context.Context, sensorID string, limit int) ([]*sensor.Reading, error)
	GetAggregatedData(ctx context.Context, sensorID string, from, to time.Time, interval time.Duration) (map[string]interface{}, error)

	// Real-time subscriptions
	SubscribeToSensorData(ctx context.Context, sensorID string) (<-chan *sensor.Reading, error)
	SubscribeToLocationData(ctx context.Context, location string) (<-chan *sensor.Reading, error)
	UnsubscribeFromSensorData(ctx context.Context, sensorID string) error
}

// DTOs for sensor service requests

type CreateSensorRequest struct {
	Name        string            `json:"name" validate:"required"`
	Type        sensor.SensorType `json:"type" validate:"required"`
	Location    string            `json:"location" validate:"required"`
	Description string            `json:"description"`
	Config      *sensor.Config    `json:"config"`
}

type UpdateSensorRequest struct {
	Name        *string        `json:"name"`
	Location    *string        `json:"location"`
	Description *string        `json:"description"`
	Status      *sensor.Status `json:"status"`
	Config      *sensor.Config `json:"config"`
}

type SubmitSensorDataRequest struct {
	Values    map[string]interface{} `json:"values" validate:"required"`
	Timestamp *time.Time             `json:"timestamp"`
	Quality   *float64               `json:"quality"`
}
