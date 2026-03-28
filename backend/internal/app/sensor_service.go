// Package app contains the application services that orchestrate between domain and infrastructure
package app

import (
	"context"
	"fmt"
	"time"

	"github.com/emiliogain/smart-home-backend/internal/domain/sensor"
	"github.com/emiliogain/smart-home-backend/internal/ports/primary"
	"github.com/emiliogain/smart-home-backend/internal/ports/secondary"
)

// sensorService implements the primary.SensorService port
// It orchestrates between domain logic and secondary adapters
type sensorService struct {
	domainService  sensor.Service
	repository     secondary.SensorRepository
	eventPublisher secondary.SensorEventPublisher
}

// NewSensorService creates a new sensor application service
func NewSensorService(
	domainService *sensor.Service,
	repository secondary.SensorRepository,
	eventPublisher secondary.SensorEventPublisher,
) primary.SensorService {
	return &sensorService{
		domainService:  *domainService,
		repository:     repository,
		eventPublisher: eventPublisher,
	}
}

// CreateSensor creates a new sensor
func (s *sensorService) CreateSensor(ctx context.Context, req primary.CreateSensorRequest) (*sensor.Sensor, error) {
	// Generate ID (in real implementation, this might come from a ID generator service)
	id := generateID()

	// Create sensor using domain logic
	newSensor, err := sensor.NewSensor(id, req.Name, req.Type, req.Location)
	if err != nil {
		return nil, fmt.Errorf("create sensor: %w", err)
	}

	// Set optional fields
	if req.Description != "" {
		newSensor.UpdateDescription(req.Description)
	}

	if req.Config != nil {
		// Validate configuration using domain service
		if err := s.domainService.ValidateConfiguration(*req.Config); err != nil {
			return nil, fmt.Errorf("invalid config: %w", err)
		}
		newSensor.UpdateConfig(*req.Config)
	}

	// Save to repository
	if err := s.repository.Save(ctx, newSensor); err != nil {
		return nil, fmt.Errorf("save sensor: %w", err)
	}

	// Publish event
	if err := s.eventPublisher.PublishSensorCreated(ctx, newSensor); err != nil {
		// Log error but don't fail the operation
		// In real implementation, use structured logging
		fmt.Printf("Failed to publish sensor created event: %v\n", err)
	}

	return newSensor, nil
}

// GetSensor retrieves a sensor by ID
func (s *sensorService) GetSensor(ctx context.Context, id string) (*sensor.Sensor, error) {
	sensorEntity, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get sensor %s: %w", id, err)
	}
	return sensorEntity, nil
}

// GetAllSensors retrieves all sensors
func (s *sensorService) GetAllSensors(ctx context.Context) ([]*sensor.Sensor, error) {
	sensors, err := s.repository.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all sensors: %w", err)
	}
	return sensors, nil
}

// GetSensorsByType retrieves sensors by type
func (s *sensorService) GetSensorsByType(ctx context.Context, sensorType sensor.SensorType) ([]*sensor.Sensor, error) {
	sensors, err := s.repository.FindByType(ctx, sensorType)
	if err != nil {
		return nil, fmt.Errorf("get sensors by type %s: %w", sensorType, err)
	}
	return sensors, nil
}

// GetSensorsByLocation retrieves sensors by location
func (s *sensorService) GetSensorsByLocation(ctx context.Context, location string) ([]*sensor.Sensor, error) {
	sensors, err := s.repository.FindByLocation(ctx, location)
	if err != nil {
		return nil, fmt.Errorf("get sensors by location %s: %w", location, err)
	}
	return sensors, nil
}

// UpdateSensor updates an existing sensor
func (s *sensorService) UpdateSensor(ctx context.Context, id string, req primary.UpdateSensorRequest) (*sensor.Sensor, error) {
	// Get existing sensor
	sensorEntity, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find sensor %s: %w", id, err)
	}

	// Apply updates using domain logic
	if req.Name != nil {
		if err := sensorEntity.UpdateName(*req.Name); err != nil {
			return nil, fmt.Errorf("update name: %w", err)
		}
	}

	if req.Location != nil {
		if err := sensorEntity.UpdateLocation(*req.Location); err != nil {
			return nil, fmt.Errorf("update location: %w", err)
		}
	}

	if req.Description != nil {
		sensorEntity.UpdateDescription(*req.Description)
	}

	if req.Status != nil {
		sensorEntity.UpdateStatus(*req.Status)
	}

	if req.Config != nil {
		// Validate configuration
		if err := s.domainService.ValidateConfiguration(*req.Config); err != nil {
			return nil, fmt.Errorf("invalid config: %w", err)
		}
		sensorEntity.UpdateConfig(*req.Config)
	}

	// Save changes
	if err := s.repository.Update(ctx, sensorEntity); err != nil {
		return nil, fmt.Errorf("update sensor: %w", err)
	}

	// Publish event
	if err := s.eventPublisher.PublishSensorUpdated(ctx, sensorEntity); err != nil {
		fmt.Printf("Failed to publish sensor updated event: %v\n", err)
	}

	return sensorEntity, nil
}

// DeleteSensor deletes a sensor
func (s *sensorService) DeleteSensor(ctx context.Context, id string) error {
	// Check if sensor exists
	exists, err := s.repository.Exists(ctx, id)
	if err != nil {
		return fmt.Errorf("check sensor exists: %w", err)
	}
	if !exists {
		return sensor.ErrSensorNotFound
	}

	// Delete sensor
	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete sensor %s: %w", id, err)
	}

	// Publish event
	if err := s.eventPublisher.PublishSensorDeleted(ctx, id); err != nil {
		fmt.Printf("Failed to publish sensor deleted event: %v\n", err)
	}

	return nil
}

// SubmitSensorData submits new sensor data
func (s *sensorService) SubmitSensorData(ctx context.Context, sensorID string, req primary.SubmitSensorDataRequest) (*sensor.Data, error) {
	// Get sensor to validate data
	sensorEntity, err := s.repository.FindByID(ctx, sensorID)
	if err != nil {
		return nil, fmt.Errorf("find sensor %s: %w", sensorID, err)
	}

	// Validate data using domain service
	if err := s.domainService.ValidateDataValues(sensorEntity.Type(), req.Values); err != nil {
		return nil, fmt.Errorf("invalid data values: %w", err)
	}

	// Create data entity
	dataID := generateID()
	data, err := sensor.NewData(dataID, sensorID, req.Values)
	if err != nil {
		return nil, fmt.Errorf("create data: %w", err)
	}

	// Set optional fields
	if req.Timestamp != nil {
		data.SetTimestamp(*req.Timestamp)
	}

	if req.Quality != nil {
		if err := data.SetQuality(*req.Quality); err != nil {
			return nil, fmt.Errorf("set quality: %w", err)
		}
	} else {
		// Calculate quality using domain service
		quality := s.domainService.CalculateDataQuality(data, sensorEntity)
		data.SetQuality(quality)
	}

	// Save data
	if err := s.repository.SaveData(ctx, data); err != nil {
		return nil, fmt.Errorf("save data: %w", err)
	}

	// Update sensor last seen
	sensorEntity.MarkAsSeen()
	if err := s.repository.Update(ctx, sensorEntity); err != nil {
		fmt.Printf("Failed to update sensor last seen: %v\n", err)
	}

	// Check for alerts
	if shouldAlert, message := s.domainService.ShouldTriggerAlert(sensorEntity, data); shouldAlert {
		if err := s.eventPublisher.PublishAlert(ctx, sensorID, "threshold", message); err != nil {
			fmt.Printf("Failed to publish alert: %v\n", err)
		}
	}

	// Publish data received event
	if err := s.eventPublisher.PublishDataReceived(ctx, data); err != nil {
		fmt.Printf("Failed to publish data received event: %v\n", err)
	}

	return data, nil
}

// GetSensorData retrieves sensor data within time range
func (s *sensorService) GetSensorData(ctx context.Context, sensorID string, from, to time.Time) ([]*sensor.Data, error) {
	data, err := s.repository.FindDataBySensor(ctx, sensorID, from, to)
	if err != nil {
		return nil, fmt.Errorf("get sensor data: %w", err)
	}
	return data, nil
}

// GetLatestSensorData retrieves the latest sensor data
func (s *sensorService) GetLatestSensorData(ctx context.Context, sensorID string) (*sensor.Data, error) {
	data, err := s.repository.FindLatestData(ctx, sensorID)
	if err != nil {
		return nil, fmt.Errorf("get latest data: %w", err)
	}
	return data, nil
}

// GetSensorDataByLocation retrieves sensor data by location
func (s *sensorService) GetSensorDataByLocation(ctx context.Context, location string, from, to time.Time) ([]*sensor.Data, error) {
	data, err := s.repository.FindDataByLocation(ctx, location, from, to)
	if err != nil {
		return nil, fmt.Errorf("get data by location: %w", err)
	}
	return data, nil
}

// GetSensorReadings converts sensor data to readings format
func (s *sensorService) GetSensorReadings(ctx context.Context, sensorID string, limit int) ([]*sensor.Reading, error) {
	// Get sensor info
	sensorEntity, err := s.repository.FindByID(ctx, sensorID)
	if err != nil {
		return nil, fmt.Errorf("find sensor: %w", err)
	}

	// Get recent data (simplified - in real implementation, add proper pagination)
	to := time.Now()
	from := to.Add(-24 * time.Hour) // Last 24 hours
	dataPoints, err := s.repository.FindDataBySensor(ctx, sensorID, from, to)
	if err != nil {
		return nil, fmt.Errorf("get data: %w", err)
	}

	// Convert to readings
	readings := make([]*sensor.Reading, 0, len(dataPoints))
	for _, data := range dataPoints {
		for key, value := range data.Values() {
			reading := &sensor.Reading{
				SensorID:  sensorID,
				Type:      sensorEntity.Type(),
				Location:  sensorEntity.Location(),
				Timestamp: data.Timestamp(),
				Value:     value,
				Quality:   data.Quality(),
			}
			readings = append(readings, reading)
		}

		if len(readings) >= limit {
			break
		}
	}

	return readings, nil
}

// GetAggregatedData retrieves aggregated sensor data
func (s *sensorService) GetAggregatedData(ctx context.Context, sensorID string, from, to time.Time, interval time.Duration) (map[string]interface{}, error) {
	// Get raw data
	dataPoints, err := s.repository.FindDataBySensor(ctx, sensorID, from, to)
	if err != nil {
		return nil, fmt.Errorf("get data: %w", err)
	}

	// Use domain service to aggregate
	aggregated := s.domainService.AggregateData(dataPoints, interval)
	return aggregated, nil
}

// Placeholder implementations for real-time subscriptions
// In a real implementation, these would use WebSocket connections or message queues

func (s *sensorService) SubscribeToSensorData(ctx context.Context, sensorID string) (<-chan *sensor.Reading, error) {
	// TODO: Implement real-time subscription
	ch := make(chan *sensor.Reading)
	close(ch)
	return ch, nil
}

func (s *sensorService) SubscribeToLocationData(ctx context.Context, location string) (<-chan *sensor.Reading, error) {
	// TODO: Implement real-time subscription
	ch := make(chan *sensor.Reading)
	close(ch)
	return ch, nil
}

func (s *sensorService) UnsubscribeFromSensorData(ctx context.Context, sensorID string) error {
	// TODO: Implement unsubscribe
	return nil
}

// Helper function - in real implementation, this would be a proper ID generator
func generateID() string {
	return fmt.Sprintf("id_%d", time.Now().UnixNano())
}
