// Package primary defines the incoming ports (primary adapters drive these)
package primary

import (
	"context"
	"time"

	"github.com/emiliogain/smart-home-backend/internal/domain/device"
)

// DeviceService defines the primary port for device operations
// This is what the HTTP handlers and other primary adapters will use
type DeviceService interface {
	// Device lifecycle
	CreateDevice(ctx context.Context, req CreateDeviceRequest) (*device.Device, error)
	GetDevice(ctx context.Context, id string) (*device.Device, error)
	GetAllDevices(ctx context.Context) ([]*device.Device, error)
	GetDevicesByType(ctx context.Context, deviceType device.Type) ([]*device.Device, error)
	GetDevicesByLocation(ctx context.Context, location string) ([]*device.Device, error)
	UpdateDevice(ctx context.Context, id string, req UpdateDeviceRequest) (*device.Device, error)
	DeleteDevice(ctx context.Context, id string) error

	// Device control
	SendCommand(ctx context.Context, deviceID string, req SendCommandRequest) (*device.Command, error)
	GetCommand(ctx context.Context, commandID string) (*device.Command, error)
	GetDeviceCommands(ctx context.Context, deviceID string, limit int) ([]*device.Command, error)
	CancelCommand(ctx context.Context, commandID string) error

	// Device state
	UpdateDeviceProperty(ctx context.Context, deviceID string, property string, value interface{}) error
	GetDeviceState(ctx context.Context, deviceID string) (*device.State, error)

	// Device monitoring
	GetDeviceHealth(ctx context.Context, deviceID string) (float64, error)
	GetDeviceEvents(ctx context.Context, deviceID string, from, to time.Time) ([]*device.Event, error)

	// Real-time subscriptions
	SubscribeToDeviceEvents(ctx context.Context, deviceID string) (<-chan *device.Event, error)
	SubscribeToLocationEvents(ctx context.Context, location string) (<-chan *device.Event, error)
	UnsubscribeFromDeviceEvents(ctx context.Context, deviceID string) error
}

// DTOs for device service requests

type CreateDeviceRequest struct {
	Name         string               `json:"name" validate:"required"`
	Type         device.Type          `json:"type" validate:"required"`
	Location     string               `json:"location" validate:"required"`
	Description  string               `json:"description"`
	Capabilities *device.Capabilities `json:"capabilities"`
}

type UpdateDeviceRequest struct {
	Name        *string        `json:"name"`
	Location    *string        `json:"location"`
	Description *string        `json:"description"`
	Status      *device.Status `json:"status"`
}

type SendCommandRequest struct {
	Command    string                 `json:"command" validate:"required"`
	Parameters map[string]interface{} `json:"parameters"`
	Timeout    *time.Duration         `json:"timeout"`
}
