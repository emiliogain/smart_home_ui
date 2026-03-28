// Package app contains the application services that orchestrate between domain and infrastructure
package app

import (
	"context"
	"fmt"
	"time"

	"github.com/emiliogain/smart-home-backend/internal/domain/device"
	"github.com/emiliogain/smart-home-backend/internal/ports/primary"
	"github.com/emiliogain/smart-home-backend/internal/ports/secondary"
)

// deviceService implements the primary.DeviceService port
// It orchestrates between domain logic and secondary adapters
type deviceService struct {
	domainService  device.Service
	repository     secondary.DeviceRepository
	controller     secondary.DeviceController
	eventPublisher secondary.DeviceEventPublisher
}

// NewDeviceService creates a new device application service
func NewDeviceService(
	domainService *device.Service,
	repository secondary.DeviceRepository,
	controller secondary.DeviceController,
	eventPublisher secondary.DeviceEventPublisher,
) primary.DeviceService {
	return &deviceService{
		domainService:  *domainService,
		repository:     repository,
		controller:     controller,
		eventPublisher: eventPublisher,
	}
}

// CreateDevice creates a new device
func (s *deviceService) CreateDevice(ctx context.Context, req primary.CreateDeviceRequest) (*device.Device, error) {
	// Generate ID
	id := generateID()

	// Create device using domain logic
	newDevice, err := device.NewDevice(id, req.Name, req.Type, req.Location)
	if err != nil {
		return nil, fmt.Errorf("create device: %w", err)
	}

	// Set optional fields
	if req.Description != "" {
		newDevice.UpdateDescription(req.Description)
	}

	// TODO: Set capabilities if provided in req.Capabilities

	// Save to repository
	if err := s.repository.Save(ctx, newDevice); err != nil {
		return nil, fmt.Errorf("save device: %w", err)
	}

	// Publish event
	if err := s.eventPublisher.PublishDeviceCreated(ctx, newDevice); err != nil {
		fmt.Printf("Failed to publish device created event: %v\n", err)
	}

	return newDevice, nil
}

// GetDevice retrieves a device by ID
func (s *deviceService) GetDevice(ctx context.Context, id string) (*device.Device, error) {
	deviceEntity, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get device %s: %w", id, err)
	}
	return deviceEntity, nil
}

// GetAllDevices retrieves all devices
func (s *deviceService) GetAllDevices(ctx context.Context) ([]*device.Device, error) {
	devices, err := s.repository.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all devices: %w", err)
	}
	return devices, nil
}

// GetDevicesByType retrieves devices by type
func (s *deviceService) GetDevicesByType(ctx context.Context, deviceType device.Type) ([]*device.Device, error) {
	devices, err := s.repository.FindByType(ctx, deviceType)
	if err != nil {
		return nil, fmt.Errorf("get devices by type %s: %w", deviceType, err)
	}
	return devices, nil
}

// GetDevicesByLocation retrieves devices by location
func (s *deviceService) GetDevicesByLocation(ctx context.Context, location string) ([]*device.Device, error) {
	devices, err := s.repository.FindByLocation(ctx, location)
	if err != nil {
		return nil, fmt.Errorf("get devices by location %s: %w", location, err)
	}
	return devices, nil
}

// UpdateDevice updates an existing device
func (s *deviceService) UpdateDevice(ctx context.Context, id string, req primary.UpdateDeviceRequest) (*device.Device, error) {
	// Get existing device
	deviceEntity, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find device %s: %w", id, err)
	}

	// Store old state for event comparison
	oldState := deviceEntity.State()

	// Apply updates using domain logic
	if req.Name != nil {
		if err := deviceEntity.UpdateName(*req.Name); err != nil {
			return nil, fmt.Errorf("update name: %w", err)
		}
	}

	if req.Location != nil {
		if err := deviceEntity.UpdateLocation(*req.Location); err != nil {
			return nil, fmt.Errorf("update location: %w", err)
		}
	}

	if req.Description != nil {
		deviceEntity.UpdateDescription(*req.Description)
	}

	if req.Status != nil {
		deviceEntity.UpdateStatus(*req.Status)
	}

	// Save changes
	if err := s.repository.Update(ctx, deviceEntity); err != nil {
		return nil, fmt.Errorf("update device: %w", err)
	}

	// Check if state change should trigger event
	newState := deviceEntity.State()
	if shouldTrigger, eventType := s.domainService.ShouldTriggerEvent(deviceEntity, oldState, newState); shouldTrigger {
		event := &device.Event{
			ID:        generateID(),
			DeviceID:  id,
			Type:      eventType,
			Data:      map[string]interface{}{"old_state": oldState, "new_state": newState},
			Timestamp: time.Now(),
		}

		if err := s.repository.SaveEvent(ctx, event); err != nil {
			fmt.Printf("Failed to save event: %v\n", err)
		}
	}

	// Publish event
	if err := s.eventPublisher.PublishDeviceUpdated(ctx, deviceEntity); err != nil {
		fmt.Printf("Failed to publish device updated event: %v\n", err)
	}

	return deviceEntity, nil
}

// DeleteDevice deletes a device
func (s *deviceService) DeleteDevice(ctx context.Context, id string) error {
	// Check if device exists
	exists, err := s.repository.Exists(ctx, id)
	if err != nil {
		return fmt.Errorf("check device exists: %w", err)
	}
	if !exists {
		return device.ErrDeviceNotFound
	}

	// Delete device
	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete device %s: %w", id, err)
	}

	// Publish event
	if err := s.eventPublisher.PublishDeviceDeleted(ctx, id); err != nil {
		fmt.Printf("Failed to publish device deleted event: %v\n", err)
	}

	return nil
}

// SendCommand sends a command to a device
func (s *deviceService) SendCommand(ctx context.Context, deviceID string, req primary.SendCommandRequest) (*device.Command, error) {
	// Get device
	deviceEntity, err := s.repository.FindByID(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("find device %s: %w", deviceID, err)
	}

	// Create command
	commandID := generateID()
	command, err := device.NewCommand(commandID, deviceID, req.Command, req.Parameters)
	if err != nil {
		return nil, fmt.Errorf("create command: %w", err)
	}

	// Validate command using domain service
	if err := s.domainService.ValidateCommand(deviceEntity, command); err != nil {
		return nil, fmt.Errorf("validate command: %w", err)
	}

	// Set timeout
	var timeout time.Duration
	if req.Timeout != nil {
		timeout = *req.Timeout
	} else {
		timeout = s.domainService.CalculateCommandTimeout(deviceEntity.Type(), req.Command)
	}

	// Save command as pending
	if err := s.repository.SaveCommand(ctx, command); err != nil {
		return nil, fmt.Errorf("save command: %w", err)
	}

	// Execute command via controller (async in real implementation)
	go func() {
		// Create context with timeout
		cmdCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// Update command status to executing
		command.UpdateStatus(device.CommandExecuting)
		s.repository.UpdateCommand(cmdCtx, command)

		// Send to device controller
		if err := s.controller.SendCommand(cmdCtx, deviceID, command); err != nil {
			command.SetError(err.Error())
		} else {
			command.SetResult(map[string]interface{}{"status": "success"})
		}

		// Update command in repository
		s.repository.UpdateCommand(context.Background(), command)

		// Publish command executed event
		s.eventPublisher.PublishCommandExecuted(context.Background(), command)
	}()

	return command, nil
}

// GetCommand retrieves a command by ID
func (s *deviceService) GetCommand(ctx context.Context, commandID string) (*device.Command, error) {
	command, err := s.repository.FindCommandByID(ctx, commandID)
	if err != nil {
		return nil, fmt.Errorf("get command %s: %w", commandID, err)
	}
	return command, nil
}

// GetDeviceCommands retrieves commands for a device
func (s *deviceService) GetDeviceCommands(ctx context.Context, deviceID string, limit int) ([]*device.Command, error) {
	commands, err := s.repository.FindCommandsByDevice(ctx, deviceID, limit)
	if err != nil {
		return nil, fmt.Errorf("get device commands: %w", err)
	}
	return commands, nil
}

// CancelCommand cancels a pending command
func (s *deviceService) CancelCommand(ctx context.Context, commandID string) error {
	// Get command
	command, err := s.repository.FindCommandByID(ctx, commandID)
	if err != nil {
		return fmt.Errorf("find command: %w", err)
	}

	// Check if command can be cancelled
	if command.Status() != device.CommandPending && command.Status() != device.CommandExecuting {
		return fmt.Errorf("command cannot be cancelled in status: %s", command.Status())
	}

	// Update command status
	command.SetError("cancelled by user")
	if err := s.repository.UpdateCommand(ctx, command); err != nil {
		return fmt.Errorf("update command: %w", err)
	}

	return nil
}

// UpdateDeviceProperty updates a device property
func (s *deviceService) UpdateDeviceProperty(ctx context.Context, deviceID string, property string, value interface{}) error {
	// Get device
	deviceEntity, err := s.repository.FindByID(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("find device: %w", err)
	}

	// Validate property update using domain service
	if err := s.domainService.ValidatePropertyUpdate(deviceEntity, property, value); err != nil {
		return fmt.Errorf("validate property: %w", err)
	}

	// Store old state
	oldState := deviceEntity.State()

	// Update property using domain logic
	if err := deviceEntity.UpdateProperty(property, value); err != nil {
		return fmt.Errorf("update property: %w", err)
	}

	// Save device state
	newState := deviceEntity.State()
	if err := s.repository.SaveState(ctx, deviceID, &newState); err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	// Update device
	if err := s.repository.Update(ctx, deviceEntity); err != nil {
		return fmt.Errorf("update device: %w", err)
	}

	// Publish state change event
	if err := s.eventPublisher.PublishDeviceStateChanged(ctx, deviceID, &oldState, &newState); err != nil {
		fmt.Printf("Failed to publish state change event: %v\n", err)
	}

	return nil
}

// GetDeviceState retrieves current device state
func (s *deviceService) GetDeviceState(ctx context.Context, deviceID string) (*device.State, error) {
	state, err := s.repository.FindState(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("get device state: %w", err)
	}
	return state, nil
}

// GetDeviceHealth calculates device health score
func (s *deviceService) GetDeviceHealth(ctx context.Context, deviceID string) (float64, error) {
	// Get device
	deviceEntity, err := s.repository.FindByID(ctx, deviceID)
	if err != nil {
		return 0, fmt.Errorf("find device: %w", err)
	}

	// Get command history
	commands, err := s.repository.GetCommandHistory(ctx, deviceID, 7) // Last 7 days
	if err != nil {
		return 0, fmt.Errorf("get command history: %w", err)
	}

	// Calculate health using domain service
	health := s.domainService.CalculateDeviceHealth(deviceEntity, commands)
	return health, nil
}

// GetDeviceEvents retrieves device events
func (s *deviceService) GetDeviceEvents(ctx context.Context, deviceID string, from, to time.Time) ([]*device.Event, error) {
	events, err := s.repository.FindEventsByDevice(ctx, deviceID, from, to)
	if err != nil {
		return nil, fmt.Errorf("get device events: %w", err)
	}
	return events, nil
}

// Placeholder implementations for real-time subscriptions

func (s *deviceService) SubscribeToDeviceEvents(ctx context.Context, deviceID string) (<-chan *device.Event, error) {
	// TODO: Implement real-time subscription
	ch := make(chan *device.Event)
	close(ch)
	return ch, nil
}

func (s *deviceService) SubscribeToLocationEvents(ctx context.Context, location string) (<-chan *device.Event, error) {
	// TODO: Implement real-time subscription
	ch := make(chan *device.Event)
	close(ch)
	return ch, nil
}

func (s *deviceService) UnsubscribeFromDeviceEvents(ctx context.Context, deviceID string) error {
	// TODO: Implement unsubscribe
	return nil
}
