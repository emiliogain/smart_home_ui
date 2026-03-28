// Package device contains the core device domain logic
package device

import (
	"time"
)

// Type represents different types of smart home devices
type Type string

const (
	TypeLight      Type = "light"
	TypeThermostat Type = "thermostat"
	TypeFan        Type = "fan"
	TypeSwitch     Type = "switch"
	TypeLock       Type = "lock"
	TypeCamera     Type = "camera"
	TypeSpeaker    Type = "speaker"
	TypeGarageDoor Type = "garage_door"
	TypeBlind      Type = "blind"
	TypeOutlet     Type = "outlet"
)

// Status represents the current status of a device
type Status string

const (
	StatusOnline  Status = "online"
	StatusOffline Status = "offline"
	StatusError   Status = "error"
)

// PowerState represents the power state of a device
type PowerState string

const (
	PowerOn  PowerState = "on"
	PowerOff PowerState = "off"
)

// Device represents the core device entity
type Device struct {
	id           string
	name         string
	deviceType   Type
	location     string
	description  string
	status       Status
	powerState   PowerState
	lastSeen     *time.Time
	createdAt    time.Time
	updatedAt    time.Time
	state        State
	capabilities Capabilities
}

// State holds the current state of the device
type State struct {
	PowerState PowerState
	Properties map[string]interface{}
	LastUpdate time.Time
}

// Capabilities describes what the device can do
type Capabilities struct {
	SupportedCommands []string
	Properties        []PropertyDefinition
	MaxConcurrentOps  int
}

// PropertyDefinition describes a device property
type PropertyDefinition struct {
	Name     string
	Type     string // "string", "number", "boolean", "enum"
	ReadOnly bool
	Min      *float64
	Max      *float64
	Options  []string // For enum types
	Unit     string
}

// Command represents a command to be sent to a device
type Command struct {
	id         string
	deviceID   string
	command    string
	parameters map[string]interface{}
	status     CommandStatus
	result     map[string]interface{}
	error      string
	createdAt  time.Time
	executedAt *time.Time
}

// CommandStatus represents the status of a device command
type CommandStatus string

const (
	CommandPending   CommandStatus = "pending"
	CommandExecuting CommandStatus = "executing"
	CommandCompleted CommandStatus = "completed"
	CommandFailed    CommandStatus = "failed"
	CommandTimeout   CommandStatus = "timeout"
)

// Event represents an event that occurred on a device
type Event struct {
	ID        string
	DeviceID  string
	Type      string
	Data      map[string]interface{}
	Timestamp time.Time
}

// NewDevice creates a new device with validation
func NewDevice(id, name string, deviceType Type, location string) (*Device, error) {
	if id == "" {
		return nil, ErrInvalidID
	}
	if name == "" {
		return nil, ErrInvalidName
	}
	if location == "" {
		return nil, ErrInvalidLocation
	}

	now := time.Now()
	return &Device{
		id:         id,
		name:       name,
		deviceType: deviceType,
		location:   location,
		status:     StatusOnline,
		powerState: PowerOff,
		createdAt:  now,
		updatedAt:  now,
		state: State{
			PowerState: PowerOff,
			Properties: make(map[string]interface{}),
			LastUpdate: now,
		},
		capabilities: Capabilities{
			SupportedCommands: getDefaultCommands(deviceType),
			Properties:        getDefaultProperties(deviceType),
			MaxConcurrentOps:  1,
		},
	}, nil
}

// Getters
func (d *Device) ID() string                 { return d.id }
func (d *Device) Name() string               { return d.name }
func (d *Device) Type() Type                 { return d.deviceType }
func (d *Device) Location() string           { return d.location }
func (d *Device) Description() string        { return d.description }
func (d *Device) Status() Status             { return d.status }
func (d *Device) PowerState() PowerState     { return d.powerState }
func (d *Device) LastSeen() *time.Time       { return d.lastSeen }
func (d *Device) CreatedAt() time.Time       { return d.createdAt }
func (d *Device) UpdatedAt() time.Time       { return d.updatedAt }
func (d *Device) State() State               { return d.state }
func (d *Device) Capabilities() Capabilities { return d.capabilities }

// UpdateName updates the device name
func (d *Device) UpdateName(name string) error {
	if name == "" {
		return ErrInvalidName
	}
	d.name = name
	d.updatedAt = time.Now()
	return nil
}

// UpdateLocation updates the device location
func (d *Device) UpdateLocation(location string) error {
	if location == "" {
		return ErrInvalidLocation
	}
	d.location = location
	d.updatedAt = time.Now()
	return nil
}

// UpdateDescription updates the device description
func (d *Device) UpdateDescription(description string) {
	d.description = description
	d.updatedAt = time.Now()
}

// UpdateStatus updates the device status
func (d *Device) UpdateStatus(status Status) {
	d.status = status
	d.updatedAt = time.Now()
	if status == StatusOnline {
		now := time.Now()
		d.lastSeen = &now
	}
}

// UpdatePowerState updates the device power state
func (d *Device) UpdatePowerState(powerState PowerState) {
	d.powerState = powerState
	d.state.PowerState = powerState
	d.state.LastUpdate = time.Now()
	d.updatedAt = time.Now()
}

// UpdateProperty updates a device property
func (d *Device) UpdateProperty(key string, value interface{}) error {
	// Validate property exists in capabilities
	if !d.hasProperty(key) {
		return ErrPropertyNotSupported
	}

	// Validate property value
	if err := d.validatePropertyValue(key, value); err != nil {
		return err
	}

	d.state.Properties[key] = value
	d.state.LastUpdate = time.Now()
	d.updatedAt = time.Now()
	return nil
}

// MarkAsSeen updates the last seen timestamp
func (d *Device) MarkAsSeen() {
	now := time.Now()
	d.lastSeen = &now
	d.updatedAt = now
}

// IsOnline checks if the device is considered online
func (d *Device) IsOnline(timeout time.Duration) bool {
	if d.lastSeen == nil {
		return false
	}
	return time.Since(*d.lastSeen) <= timeout
}

// CanExecuteCommand checks if device supports a command
func (d *Device) CanExecuteCommand(command string) bool {
	for _, cmd := range d.capabilities.SupportedCommands {
		if cmd == command {
			return true
		}
	}
	return false
}

// NewCommand creates a new command for the device
func NewCommand(id, deviceID, command string, parameters map[string]interface{}) (*Command, error) {
	if id == "" {
		return nil, ErrInvalidID
	}
	if deviceID == "" {
		return nil, ErrInvalidDeviceID
	}
	if command == "" {
		return nil, ErrInvalidCommand
	}

	return &Command{
		id:         id,
		deviceID:   deviceID,
		command:    command,
		parameters: parameters,
		status:     CommandPending,
		createdAt:  time.Now(),
	}, nil
}

// Getters for Command
func (c *Command) ID() string                         { return c.id }
func (c *Command) DeviceID() string                   { return c.deviceID }
func (c *Command) Command() string                    { return c.command }
func (c *Command) Parameters() map[string]interface{} { return c.parameters }
func (c *Command) Status() CommandStatus              { return c.status }
func (c *Command) Result() map[string]interface{}     { return c.result }
func (c *Command) Error() string                      { return c.error }
func (c *Command) CreatedAt() time.Time               { return c.createdAt }
func (c *Command) ExecutedAt() *time.Time             { return c.executedAt }

// UpdateStatus updates command status
func (c *Command) UpdateStatus(status CommandStatus) {
	c.status = status
	if status == CommandExecuting && c.executedAt == nil {
		now := time.Now()
		c.executedAt = &now
	}
}

// SetResult sets command result
func (c *Command) SetResult(result map[string]interface{}) {
	c.result = result
	c.status = CommandCompleted
}

// SetError sets command error
func (c *Command) SetError(err string) {
	c.error = err
	c.status = CommandFailed
}

// Private helper methods

func (d *Device) hasProperty(key string) bool {
	for _, prop := range d.capabilities.Properties {
		if prop.Name == key {
			return true
		}
	}
	return false
}

func (d *Device) validatePropertyValue(key string, value interface{}) error {
	for _, prop := range d.capabilities.Properties {
		if prop.Name == key {
			return validatePropertyByDefinition(prop, value)
		}
	}
	return ErrPropertyNotFound
}

func validatePropertyByDefinition(def PropertyDefinition, value interface{}) error {
	switch def.Type {
	case "boolean":
		if _, ok := value.(bool); !ok {
			return ErrInvalidPropertyType
		}
	case "number":
		if numValue, ok := value.(float64); ok {
			if def.Min != nil && numValue < *def.Min {
				return ErrValueTooSmall
			}
			if def.Max != nil && numValue > *def.Max {
				return ErrValueTooLarge
			}
		} else {
			return ErrInvalidPropertyType
		}
	case "string":
		if _, ok := value.(string); !ok {
			return ErrInvalidPropertyType
		}
	case "enum":
		strValue, ok := value.(string)
		if !ok {
			return ErrInvalidPropertyType
		}
		for _, option := range def.Options {
			if option == strValue {
				return nil
			}
		}
		return ErrInvalidEnumValue
	}
	return nil
}

func getDefaultCommands(deviceType Type) []string {
	switch deviceType {
	case TypeLight:
		return []string{"turn_on", "turn_off", "set_brightness", "set_color"}
	case TypeThermostat:
		return []string{"set_temperature", "set_mode"}
	case TypeFan:
		return []string{"turn_on", "turn_off", "set_speed"}
	case TypeSwitch:
		return []string{"turn_on", "turn_off"}
	case TypeLock:
		return []string{"lock", "unlock"}
	case TypeCamera:
		return []string{"start_recording", "stop_recording", "take_snapshot"}
	case TypeSpeaker:
		return []string{"play", "pause", "stop", "set_volume"}
	case TypeGarageDoor:
		return []string{"open", "close"}
	case TypeBlind:
		return []string{"open", "close", "set_position"}
	case TypeOutlet:
		return []string{"turn_on", "turn_off"}
	default:
		return []string{"turn_on", "turn_off"}
	}
}

func getDefaultProperties(deviceType Type) []PropertyDefinition {
	switch deviceType {
	case TypeLight:
		return []PropertyDefinition{
			{Name: "brightness", Type: "number", Min: ptrFloat64(0), Max: ptrFloat64(100), Unit: "%"},
			{Name: "color", Type: "string"},
		}
	case TypeThermostat:
		return []PropertyDefinition{
			{Name: "temperature", Type: "number", Min: ptrFloat64(-10), Max: ptrFloat64(40), Unit: "°C"},
			{Name: "mode", Type: "enum", Options: []string{"heat", "cool", "auto", "off"}},
		}
	case TypeFan:
		return []PropertyDefinition{
			{Name: "speed", Type: "number", Min: ptrFloat64(0), Max: ptrFloat64(100), Unit: "%"},
		}
	case TypeSpeaker:
		return []PropertyDefinition{
			{Name: "volume", Type: "number", Min: ptrFloat64(0), Max: ptrFloat64(100), Unit: "%"},
		}
	case TypeBlind:
		return []PropertyDefinition{
			{Name: "position", Type: "number", Min: ptrFloat64(0), Max: ptrFloat64(100), Unit: "%"},
		}
	default:
		return []PropertyDefinition{}
	}
}

func ptrFloat64(f float64) *float64 {
	return &f
}
