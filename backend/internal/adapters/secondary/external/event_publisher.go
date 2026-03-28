// Package external contains external service adapters (secondary adapters)
package external

import (
	"context"
	"fmt"

	"github.com/emiliogain/smart-home-backend/internal/domain/device"
	"github.com/emiliogain/smart-home-backend/internal/domain/sensor"
	"github.com/emiliogain/smart-home-backend/internal/ports/secondary"
)

// eventPublisher implements both SensorEventPublisher and DeviceEventPublisher
// In a real implementation, this would use RabbitMQ, Apache Kafka, etc.
type eventPublisher struct {
	// In real implementation, would have message queue connections
}

// NewEventPublisher creates a new event publisher
func NewEventPublisher() EventPublisher {
	return &eventPublisher{}
}

// EventPublisher combines sensor and device event publishing
type EventPublisher interface {
	secondary.SensorEventPublisher
	secondary.DeviceEventPublisher
}

// Sensor event publishing methods

// PublishSensorCreated publishes sensor created event
func (p *eventPublisher) PublishSensorCreated(ctx context.Context, s *sensor.Sensor) error {
	fmt.Printf("EVENT: Sensor created - ID: %s, Name: %s, Type: %s, Location: %s\n",
		s.ID(), s.Name(), s.Type(), s.Location())
	return nil
}

// PublishSensorUpdated publishes sensor updated event
func (p *eventPublisher) PublishSensorUpdated(ctx context.Context, s *sensor.Sensor) error {
	fmt.Printf("EVENT: Sensor updated - ID: %s, Name: %s\n", s.ID(), s.Name())
	return nil
}

// PublishSensorDeleted publishes sensor deleted event
func (p *eventPublisher) PublishSensorDeleted(ctx context.Context, sensorID string) error {
	fmt.Printf("EVENT: Sensor deleted - ID: %s\n", sensorID)
	return nil
}

// PublishDataReceived publishes sensor data received event
func (p *eventPublisher) PublishDataReceived(ctx context.Context, data *sensor.Data) error {
	fmt.Printf("EVENT: Data received - Sensor: %s, Values: %v\n",
		data.SensorID(), data.Values())
	return nil
}

// PublishAlert publishes sensor alert event
func (p *eventPublisher) PublishAlert(ctx context.Context, sensorID string, alertType string, message string) error {
	fmt.Printf("ALERT: Sensor %s - Type: %s, Message: %s\n",
		sensorID, alertType, message)
	return nil
}

// Device event publishing methods

// PublishDeviceCreated publishes device created event
func (p *eventPublisher) PublishDeviceCreated(ctx context.Context, d *device.Device) error {
	fmt.Printf("EVENT: Device created - ID: %s, Name: %s, Type: %s, Location: %s\n",
		d.ID(), d.Name(), d.Type(), d.Location())
	return nil
}

// PublishDeviceUpdated publishes device updated event
func (p *eventPublisher) PublishDeviceUpdated(ctx context.Context, d *device.Device) error {
	fmt.Printf("EVENT: Device updated - ID: %s, Name: %s\n", d.ID(), d.Name())
	return nil
}

// PublishDeviceDeleted publishes device deleted event
func (p *eventPublisher) PublishDeviceDeleted(ctx context.Context, deviceID string) error {
	fmt.Printf("EVENT: Device deleted - ID: %s\n", deviceID)
	return nil
}

// PublishCommandExecuted publishes command executed event
func (p *eventPublisher) PublishCommandExecuted(ctx context.Context, cmd *device.Command) error {
	fmt.Printf("EVENT: Command executed - Device: %s, Command: %s, Status: %s\n",
		cmd.DeviceID(), cmd.Command(), cmd.Status())
	return nil
}

// PublishDeviceStateChanged publishes device state changed event
func (p *eventPublisher) PublishDeviceStateChanged(ctx context.Context, deviceID string, oldState, newState *device.State) error {
	fmt.Printf("EVENT: Device state changed - ID: %s, Old Power: %s, New Power: %s\n",
		deviceID, oldState.PowerState, newState.PowerState)
	return nil
}

// PublishDeviceOffline publishes device offline event
func (p *eventPublisher) PublishDeviceOffline(ctx context.Context, deviceID string) error {
	fmt.Printf("EVENT: Device offline - ID: %s\n", deviceID)
	return nil
}

// PublishDeviceOnline publishes device online event
func (p *eventPublisher) PublishDeviceOnline(ctx context.Context, deviceID string) error {
	fmt.Printf("EVENT: Device online - ID: %s\n", deviceID)
	return nil
}
