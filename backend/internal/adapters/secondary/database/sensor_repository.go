// Package database contains database adapters (secondary adapters)
package database

import (
	"context"
	"sync"
	"time"

	"github.com/emiliogain/smart-home-backend/internal/domain/sensor"
	"github.com/emiliogain/smart-home-backend/internal/ports/secondary"
)

// sensorRepository implements secondary.SensorRepository using in-memory storage
// In a real implementation, this would use PostgreSQL, MongoDB, etc.
type sensorRepository struct {
	sensors    map[string]*sensor.Sensor
	sensorData map[string][]*sensor.Data
	mutex      sync.RWMutex
}

// NewSensorRepository creates a new in-memory sensor repository
func NewSensorRepository() secondary.SensorRepository {
	return &sensorRepository{
		sensors:    make(map[string]*sensor.Sensor),
		sensorData: make(map[string][]*sensor.Data),
	}
}

// Save saves a sensor
func (r *sensorRepository) Save(ctx context.Context, s *sensor.Sensor) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.sensors[s.ID()] = s
	return nil
}

// FindByID finds a sensor by ID
func (r *sensorRepository) FindByID(ctx context.Context, id string) (*sensor.Sensor, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	s, exists := r.sensors[id]
	if !exists {
		return nil, sensor.ErrSensorNotFound
	}
	return s, nil
}

// FindAll finds all sensors
func (r *sensorRepository) FindAll(ctx context.Context) ([]*sensor.Sensor, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	sensors := make([]*sensor.Sensor, 0, len(r.sensors))
	for _, s := range r.sensors {
		sensors = append(sensors, s)
	}
	return sensors, nil
}

// FindByType finds sensors by type
func (r *sensorRepository) FindByType(ctx context.Context, sensorType sensor.SensorType) ([]*sensor.Sensor, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var sensors []*sensor.Sensor
	for _, s := range r.sensors {
		if s.Type() == sensorType {
			sensors = append(sensors, s)
		}
	}
	return sensors, nil
}

// FindByLocation finds sensors by location
func (r *sensorRepository) FindByLocation(ctx context.Context, location string) ([]*sensor.Sensor, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var sensors []*sensor.Sensor
	for _, s := range r.sensors {
		if s.Location() == location {
			sensors = append(sensors, s)
		}
	}
	return sensors, nil
}

// Update updates a sensor
func (r *sensorRepository) Update(ctx context.Context, s *sensor.Sensor) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.sensors[s.ID()]; !exists {
		return sensor.ErrSensorNotFound
	}
	r.sensors[s.ID()] = s
	return nil
}

// Delete deletes a sensor
func (r *sensorRepository) Delete(ctx context.Context, id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.sensors[id]; !exists {
		return sensor.ErrSensorNotFound
	}
	delete(r.sensors, id)
	delete(r.sensorData, id)
	return nil
}

// Exists checks if a sensor exists
func (r *sensorRepository) Exists(ctx context.Context, id string) (bool, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	_, exists := r.sensors[id]
	return exists, nil
}

// SaveData saves sensor data
func (r *sensorRepository) SaveData(ctx context.Context, data *sensor.Data) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.sensorData[data.SensorID()] = append(r.sensorData[data.SensorID()], data)
	return nil
}

// FindDataByID finds sensor data by ID
func (r *sensorRepository) FindDataByID(ctx context.Context, id string) (*sensor.Data, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, dataList := range r.sensorData {
		for _, data := range dataList {
			if data.ID() == id {
				return data, nil
			}
		}
	}
	return nil, sensor.ErrDataNotFound
}

// FindDataBySensor finds data for a sensor within time range
func (r *sensorRepository) FindDataBySensor(ctx context.Context, sensorID string, from, to time.Time) ([]*sensor.Data, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	allData, exists := r.sensorData[sensorID]
	if !exists {
		return []*sensor.Data{}, nil
	}

	var filteredData []*sensor.Data
	for _, data := range allData {
		if data.Timestamp().After(from) && data.Timestamp().Before(to) {
			filteredData = append(filteredData, data)
		}
	}

	return filteredData, nil
}

// FindLatestData finds the latest data for a sensor
func (r *sensorRepository) FindLatestData(ctx context.Context, sensorID string) (*sensor.Data, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	allData, exists := r.sensorData[sensorID]
	if !exists || len(allData) == 0 {
		return nil, sensor.ErrDataNotFound
	}

	latest := allData[0]
	for _, data := range allData {
		if data.Timestamp().After(latest.Timestamp()) {
			latest = data
		}
	}

	return latest, nil
}

// FindDataByLocation finds data for all sensors in a location
func (r *sensorRepository) FindDataByLocation(ctx context.Context, location string, from, to time.Time) ([]*sensor.Data, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var allData []*sensor.Data

	// Find sensors in location
	for _, s := range r.sensors {
		if s.Location() == location {
			sensorData, exists := r.sensorData[s.ID()]
			if !exists {
				continue
			}

			// Filter by time range
			for _, data := range sensorData {
				if data.Timestamp().After(from) && data.Timestamp().Before(to) {
					allData = append(allData, data)
				}
			}
		}
	}

	return allData, nil
}

// FindAggregatedData finds aggregated data (simplified implementation)
func (r *sensorRepository) FindAggregatedData(ctx context.Context, sensorID string, from, to time.Time, interval time.Duration) ([]*sensor.Data, error) {
	// For this simple implementation, just return raw data
	// In a real database, this would use SQL aggregation functions
	return r.FindDataBySensor(ctx, sensorID, from, to)
}

// GetDataCount returns the count of data points
func (r *sensorRepository) GetDataCount(ctx context.Context, sensorID string, from, to time.Time) (int64, error) {
	data, err := r.FindDataBySensor(ctx, sensorID, from, to)
	if err != nil {
		return 0, err
	}
	return int64(len(data)), nil
}

// DeleteOldData deletes old sensor data
func (r *sensorRepository) DeleteOldData(ctx context.Context, before time.Time) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for sensorID, dataList := range r.sensorData {
		var kept []*sensor.Data
		for _, data := range dataList {
			if data.Timestamp().After(before) {
				kept = append(kept, data)
			}
		}
		r.sensorData[sensorID] = kept
	}

	return nil
}

// CalculateStorageSize calculates storage size for sensor data
func (r *sensorRepository) CalculateStorageSize(ctx context.Context, sensorID string) (int64, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	dataList, exists := r.sensorData[sensorID]
	if !exists {
		return 0, nil
	}

	// Rough calculation (in real implementation, would be more accurate)
	return int64(len(dataList) * 1024), nil // Assume 1KB per data point
}
