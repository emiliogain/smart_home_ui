package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	ctxdomain "github.com/emiliogain/smart-home-backend/internal/domain/context"
	"github.com/emiliogain/smart-home-backend/internal/domain/sensor"
	"github.com/emiliogain/smart-home-backend/internal/ports/secondary"
)

// SensorService orchestrates sensor operations between the HTTP layer and persistence/fusion.
type SensorService struct {
	repo        secondary.SensorRepository
	fusion      secondary.FusionPredictor
	broadcaster secondary.EventBroadcaster

	mu          sync.RWMutex
	lastContext *ctxdomain.ContextUpdate
}

// NewSensorService creates a new sensor application service.
func NewSensorService(repo secondary.SensorRepository, fusion secondary.FusionPredictor, broadcaster secondary.EventBroadcaster) *SensorService {
	return &SensorService{repo: repo, fusion: fusion, broadcaster: broadcaster}
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

	// Fetch recent readings from ALL sensors for cross-sensor fusion.
	enriched, err := s.repo.GetAllLatestReadings(ctx, 10)
	if err != nil {
		return nil, fmt.Errorf("get all latest readings: %w", err)
	}

	window := buildSensorWindow(enriched)

	result, err := s.fusion.Predict(ctx, window)
	if err != nil {
		return nil, fmt.Errorf("fusion predict: %w", err)
	}

	// Build the context update and cache it.
	snapshot := buildSnapshot(enriched)
	update := ctxdomain.NewContextUpdate(
		ctxdomain.ContextType(result.Label),
		result.Confidence,
		snapshot,
	)

	s.mu.Lock()
	s.lastContext = &update
	s.mu.Unlock()

	// Push to connected frontend clients via WebSocket.
	if s.broadcaster != nil {
		_ = s.broadcaster.BroadcastContextUpdate(ctx, update)
	}

	return result, nil
}

// SaveReadingsBatch persists multiple readings and runs fusion ONCE after all are saved.
// This avoids intermediate context flicker when the simulator submits a batch of sensor readings.
func (s *SensorService) SaveReadingsBatch(ctx context.Context, readings []sensor.Reading) (*secondary.FusionResult, error) {
	for i := range readings {
		if readings[i].Timestamp.IsZero() {
			readings[i].Timestamp = time.Now()
		}
		if err := s.repo.SaveReading(ctx, readings[i]); err != nil {
			return nil, fmt.Errorf("save reading %s: %w", readings[i].SensorID, err)
		}
	}

	// Run fusion once after all readings are persisted.
	enriched, err := s.repo.GetAllLatestReadings(ctx, 10)
	if err != nil {
		return nil, fmt.Errorf("get all latest readings: %w", err)
	}

	window := buildSensorWindow(enriched)

	result, err := s.fusion.Predict(ctx, window)
	if err != nil {
		return nil, fmt.Errorf("fusion predict: %w", err)
	}

	snapshot := buildSnapshot(enriched)
	update := ctxdomain.NewContextUpdate(
		ctxdomain.ContextType(result.Label),
		result.Confidence,
		snapshot,
	)

	s.mu.Lock()
	s.lastContext = &update
	s.mu.Unlock()

	if s.broadcaster != nil {
		_ = s.broadcaster.BroadcastContextUpdate(ctx, update)
	}

	return result, nil
}

func (s *SensorService) GetLatestReadings(ctx context.Context, sensorID string, limit int) ([]sensor.Reading, error) {
	return s.repo.GetLatestReadings(ctx, sensorID, limit)
}

// GetCurrentContext returns the most recent fusion result, or nil if none.
func (s *SensorService) GetCurrentContext() *ctxdomain.ContextUpdate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastContext
}

func buildSensorWindow(readings []sensor.EnrichedReading) secondary.SensorWindow {
	w := secondary.SensorWindow{
		All:        readings,
		ByType:     make(map[sensor.SensorType][]sensor.EnrichedReading),
		ByLocation: make(map[string][]sensor.EnrichedReading),
	}
	for _, r := range readings {
		w.ByType[r.SensorType] = append(w.ByType[r.SensorType], r)
		w.ByLocation[r.Location] = append(w.ByLocation[r.Location], r)
	}
	return w
}

func buildSnapshot(readings []sensor.EnrichedReading) *ctxdomain.SensorSnapshot {
	// Deduplicate: keep the most recent reading per sensor.
	// Use the sensor Name as sensorId so the frontend can pattern-match
	// on human-readable names (e.g. "temp_living_room" → includes "temp", "living").
	seen := make(map[string]bool)
	var out []ctxdomain.SensorReading
	for _, r := range readings {
		key := r.SensorID
		if seen[key] {
			continue
		}
		seen[key] = true
		// Prefer the human-readable name; fall back to UUID if name is empty.
		sensorID := r.SensorName
		if sensorID == "" {
			sensorID = r.SensorID
		}
		out = append(out, ctxdomain.SensorReading{
			SensorID: sensorID,
			Value:    r.Value,
			Unit:     r.Unit,
			At:       r.Timestamp.UTC().Format(time.RFC3339),
		})
	}
	return &ctxdomain.SensorSnapshot{Readings: out}
}
