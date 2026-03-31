package app

import (
	"context"
	"fmt"
	"time"

	"github.com/emiliogain/smart-home-backend/internal/domain/sensor"
	"github.com/emiliogain/smart-home-backend/internal/ports/secondary"
)

// SensorService orchestrates sensor operations between the HTTP layer and persistence/fusion.
type SensorService struct {
	repo   secondary.SensorRepository
	fusion secondary.FusionPredictor
}

// NewSensorService creates a new sensor application service.
func NewSensorService(repo secondary.SensorRepository, fusion secondary.FusionPredictor) *SensorService {
	return &SensorService{repo: repo, fusion: fusion}
}

func (s *SensorService) CreateSensor(ctx context.Context, sr sensor.Sensor) error {
	sr.CreatedAt = time.Now()
	sr.UpdatedAt = sr.CreatedAt
	if sr.Status == "" {
		sr.Status = "active"
	}
	return s.repo.SaveSensor(ctx, sr)
}

func (s *SensorService) GetSensor(ctx context.Context, id string) (*sensor.Sensor, error) {
	return s.repo.GetSensor(ctx, id)
}

func (s *SensorService) ListSensors(ctx context.Context) ([]sensor.Sensor, error) {
	return s.repo.ListSensors(ctx)
}

// SaveReading persists a reading and runs the fusion model on the recent window.
func (s *SensorService) SaveReading(ctx context.Context, r sensor.Reading) (*secondary.FusionResult, error) {
	if r.Timestamp.IsZero() {
		r.Timestamp = time.Now()
	}

	if err := s.repo.SaveReading(ctx, r); err != nil {
		return nil, fmt.Errorf("save reading: %w", err)
	}

	// Fetch recent readings for fusion.
	readings, err := s.repo.GetLatestReadings(ctx, r.SensorID, 50)
	if err != nil {
		return nil, fmt.Errorf("get latest readings: %w", err)
	}

	result, err := s.fusion.Predict(ctx, readings)
	if err != nil {
		return nil, fmt.Errorf("fusion predict: %w", err)
	}

	// TODO: push result to frontend via WebSocket

	return result, nil
}

func (s *SensorService) GetLatestReadings(ctx context.Context, sensorID string, limit int) ([]sensor.Reading, error) {
	return s.repo.GetLatestReadings(ctx, sensorID, limit)
}
