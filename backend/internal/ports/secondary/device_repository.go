// Package secondary defines the outgoing ports (secondary adapters implement these)
package secondary

import (
	"context"
	"time"

	"github.com/emiliogain/smart-home-backend/internal/domain/device"
)

// DeviceRepository defines the secondary port for device persistence
// Database adapters will implement this interface
type DeviceRepository interface {
	// Device persistence
	Save(ctx context.Context, device *device.Device) error
	FindByID(ctx context.Context, id string) (*device.Device, error)
	FindAll(ctx context.Context) ([]*device.Device, error)
	FindByType(ctx context.Context, deviceType device.Type) ([]*device.Device, error)
	FindByLocation(ctx context.Context, location string) ([]*device.Device, error)
	Update(ctx context.Context, device *device.Device) error
	Delete(ctx context.Context, id string) error
	Exists(ctx context.Context, id string) (bool, error)

	// Device state persistence
	SaveState(ctx context.Context, deviceID string, state *device.State) error
	FindState(ctx context.Context, deviceID string) (*device.State, error)

	// Command persistence
	SaveCommand(ctx context.Context, command *device.Command) error
	FindCommandByID(ctx context.Context, id string) (*device.Command, error)
	FindCommandsByDevice(ctx context.Context, deviceID string, limit int) ([]*device.Command, error)
	FindPendingCommands(ctx context.Context) ([]*device.Command, error)
	UpdateCommand(ctx context.Context, command *device.Command) error
	DeleteCommand(ctx context.Context, id string) error

	// Event persistence
	SaveEvent(ctx context.Context, event *device.Event) error
	FindEventsByDevice(ctx context.Context, deviceID string, from, to time.Time) ([]*device.Event, error)
	FindEventsByLocation(ctx context.Context, location string, from, to time.Time) ([]*device.Event, error)

	// Analytics and monitoring
	GetCommandHistory(ctx context.Context, deviceID string, days int) ([]*device.Command, error)
	CalculateUptime(ctx context.Context, deviceID string, from, to time.Time) (time.Duration, error)
}

// DeviceController defines the port for communicating with physical devices
// Hardware communication adapters will implement this interface
type DeviceController interface {
	// Device communication
	SendCommand(ctx context.Context, deviceID string, command *device.Command) error
	GetDeviceStatus(ctx context.Context, deviceID string) (*device.State, error)
	DiscoverDevices(ctx context.Context) ([]*device.Device, error)

	// Real-time communication
	SubscribeToDeviceEvents(ctx context.Context, deviceID string) (<-chan *device.Event, error)
	UnsubscribeFromDeviceEvents(ctx context.Context, deviceID string) error

	// Device management
	PairDevice(ctx context.Context, deviceID string) error
	UnpairDevice(ctx context.Context, deviceID string) error
	ResetDevice(ctx context.Context, deviceID string) error
}

// DeviceEventPublisher defines the port for publishing device events
// Message queue adapters will implement this interface
type DeviceEventPublisher interface {
	PublishDeviceCreated(ctx context.Context, device *device.Device) error
	PublishDeviceUpdated(ctx context.Context, device *device.Device) error
	PublishDeviceDeleted(ctx context.Context, deviceID string) error
	PublishCommandExecuted(ctx context.Context, command *device.Command) error
	PublishDeviceStateChanged(ctx context.Context, deviceID string, oldState, newState *device.State) error
	PublishDeviceOffline(ctx context.Context, deviceID string) error
	PublishDeviceOnline(ctx context.Context, deviceID string) error
}
