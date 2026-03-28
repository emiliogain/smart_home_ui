package device

import (
	"errors"
	"time"
)

// Service contains the core business logic for devices
// It has NO external dependencies - pure domain logic
type Service struct {
	// No infrastructure dependencies here!
}

// NewService creates a new device domain service
func NewService() *Service {
	return &Service{}
}

// ValidateCommand validates a device command before execution
func (s *Service) ValidateCommand(device *Device, command *Command) error {
	// Check if device supports the command
	if !device.CanExecuteCommand(command.Command()) {
		return ErrCommandNotSupported
	}

	// Check if device is online
	if device.Status() != StatusOnline {
		return ErrDeviceOffline
	}

	// Validate command parameters based on device type and command
	return s.validateCommandParameters(device, command)
}

// CanExecuteCommand checks if a command can be executed on a device
func (s *Service) CanExecuteCommand(device *Device, commandName string) bool {
	if device.Status() != StatusOnline {
		return false
	}

	return device.CanExecuteCommand(commandName)
}

// CalculateCommandTimeout calculates appropriate timeout for a command
func (s *Service) CalculateCommandTimeout(deviceType Type, command string) time.Duration {
	baseTimeout := 30 * time.Second

	switch deviceType {
	case TypeGarageDoor:
		return 60 * time.Second // Garage doors are slow
	case TypeThermostat:
		return 45 * time.Second // HVAC systems need time
	case TypeCamera:
		if command == "take_snapshot" {
			return 15 * time.Second
		}
		return 30 * time.Second
	case TypeLock:
		return 10 * time.Second // Locks should be fast
	default:
		return baseTimeout
	}
}

// ShouldRetryCommand determines if a failed command should be retried
func (s *Service) ShouldRetryCommand(command *Command, attemptCount int) bool {
	maxRetries := 3

	if attemptCount >= maxRetries {
		return false
	}

	// Don't retry certain commands
	nonRetryableCommands := map[string]bool{
		"unlock":          true, // Security sensitive
		"lock":            true, // Security sensitive
		"take_snapshot":   true, // One-time operations
		"start_recording": true,
	}

	if nonRetryableCommands[command.Command()] {
		return false
	}

	// Only retry if command timed out or device was temporarily offline
	return command.Status() == CommandTimeout
}

// GetRetryDelay calculates delay before retrying a command
func (s *Service) GetRetryDelay(attemptCount int) time.Duration {
	// Exponential backoff: 1s, 2s, 4s, 8s...
	baseDelay := time.Second
	return baseDelay * time.Duration(1<<attemptCount)
}

// ValidatePropertyUpdate validates a property update
func (s *Service) ValidatePropertyUpdate(device *Device, property string, value interface{}) error {
	// Check if device supports the property
	capabilities := device.Capabilities()

	var propDef *PropertyDefinition
	for _, prop := range capabilities.Properties {
		if prop.Name == property {
			propDef = &prop
			break
		}
	}

	if propDef == nil {
		return ErrPropertyNotSupported
	}

	// Check if property is read-only
	if propDef.ReadOnly {
		return errors.New("property is read-only")
	}

	// Validate value type and constraints
	return validatePropertyByDefinition(*propDef, value)
}

// ShouldTriggerEvent determines if a device state change should trigger an event
func (s *Service) ShouldTriggerEvent(device *Device, oldState, newState State) (bool, string) {
	// Power state changes always trigger events
	if oldState.PowerState != newState.PowerState {
		return true, "power_state_changed"
	}

	// Property changes for certain device types
	switch device.Type() {
	case TypeLock:
		if s.propertyChanged(oldState, newState, "locked") {
			return true, "lock_state_changed"
		}
	case TypeThermostat:
		if s.propertyChanged(oldState, newState, "temperature") {
			return true, "temperature_changed"
		}
	case TypeCamera:
		if s.propertyChanged(oldState, newState, "recording") {
			return true, "recording_state_changed"
		}
	}

	return false, ""
}

// CalculateDeviceHealth calculates a health score for the device
func (s *Service) CalculateDeviceHealth(device *Device, commandHistory []*Command) float64 {
	health := 1.0

	// Reduce health based on offline time
	if device.LastSeen() != nil {
		offlineTime := time.Since(*device.LastSeen())
		if offlineTime > time.Hour {
			health *= 0.8
		}
	}

	// Reduce health based on failed commands
	if len(commandHistory) > 0 {
		recentCommands := s.getRecentCommands(commandHistory, 24*time.Hour)
		failureRate := s.calculateFailureRate(recentCommands)
		health *= (1.0 - failureRate)
	}

	// Device type specific health factors
	health *= s.calculateTypeSpecificHealth(device)

	return health
}

// OptimizeCommandExecution suggests optimal parameters for command execution
func (s *Service) OptimizeCommandExecution(device *Device, command string) map[string]interface{} {
	optimizations := make(map[string]interface{})

	switch device.Type() {
	case TypeLight:
		if command == "turn_on" {
			// Suggest gradual brightness increase for better UX
			optimizations["gradual"] = true
			optimizations["duration"] = "2s"
		}
	case TypeThermostat:
		if command == "set_temperature" {
			// Suggest eco-friendly temperature ranges
			optimizations["eco_mode"] = true
		}
	case TypeFan:
		if command == "turn_on" {
			// Start at low speed for noise reduction
			optimizations["initial_speed"] = 30
		}
	}

	return optimizations
}

// Private helper methods

func (s *Service) validateCommandParameters(device *Device, command *Command) error {
	params := command.Parameters()

	switch command.Command() {
	case "set_brightness":
		if brightness, exists := params["brightness"]; exists {
			if b, ok := brightness.(float64); ok {
				if b < 0 || b > 100 {
					return errors.New("brightness must be between 0 and 100")
				}
			} else {
				return errors.New("brightness must be a number")
			}
		} else {
			return errors.New("brightness parameter required")
		}

	case "set_temperature":
		if temp, exists := params["temperature"]; exists {
			if t, ok := temp.(float64); ok {
				if t < -10 || t > 40 {
					return errors.New("temperature must be between -10 and 40")
				}
			} else {
				return errors.New("temperature must be a number")
			}
		} else {
			return errors.New("temperature parameter required")
		}

	case "set_volume":
		if volume, exists := params["volume"]; exists {
			if v, ok := volume.(float64); ok {
				if v < 0 || v > 100 {
					return errors.New("volume must be between 0 and 100")
				}
			} else {
				return errors.New("volume must be a number")
			}
		} else {
			return errors.New("volume parameter required")
		}
	}

	return nil
}

func (s *Service) propertyChanged(oldState, newState State, property string) bool {
	oldValue, oldExists := oldState.Properties[property]
	newValue, newExists := newState.Properties[property]

	if oldExists != newExists {
		return true
	}

	if !oldExists {
		return false
	}

	return oldValue != newValue
}

func (s *Service) getRecentCommands(commands []*Command, duration time.Duration) []*Command {
	cutoff := time.Now().Add(-duration)
	var recent []*Command

	for _, cmd := range commands {
		if cmd.CreatedAt().After(cutoff) {
			recent = append(recent, cmd)
		}
	}

	return recent
}

func (s *Service) calculateFailureRate(commands []*Command) float64 {
	if len(commands) == 0 {
		return 0
	}

	failed := 0
	for _, cmd := range commands {
		if cmd.Status() == CommandFailed || cmd.Status() == CommandTimeout {
			failed++
		}
	}

	return float64(failed) / float64(len(commands))
}

func (s *Service) calculateTypeSpecificHealth(device *Device) float64 {
	// Device type specific health calculations
	switch device.Type() {
	case TypeCamera:
		// Cameras need regular connectivity
		if device.Status() == StatusOnline {
			return 1.0
		}
		return 0.5
	case TypeLock:
		// Security devices are critical
		if device.Status() != StatusOnline {
			return 0.3
		}
		return 1.0
	default:
		return 1.0
	}
}
