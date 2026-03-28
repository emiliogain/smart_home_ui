package device

import "errors"

// Domain errors for device operations
var (
	ErrInvalidID             = errors.New("invalid device ID")
	ErrInvalidName           = errors.New("invalid device name")
	ErrInvalidLocation       = errors.New("invalid device location")
	ErrInvalidDeviceID       = errors.New("invalid device ID in command")
	ErrInvalidCommand        = errors.New("invalid command")
	ErrDeviceNotFound        = errors.New("device not found")
	ErrCommandNotFound       = errors.New("command not found")
	ErrInvalidType           = errors.New("invalid device type")
	ErrPropertyNotSupported  = errors.New("property not supported by device")
	ErrPropertyNotFound      = errors.New("property not found")
	ErrInvalidPropertyType   = errors.New("invalid property type")
	ErrValueTooSmall         = errors.New("value below minimum")
	ErrValueTooLarge         = errors.New("value above maximum")
	ErrInvalidEnumValue      = errors.New("invalid enum value")
	ErrCommandNotSupported   = errors.New("command not supported by device")
	ErrDeviceOffline         = errors.New("device is offline")
	ErrCommandTimeout        = errors.New("command execution timeout")
	ErrCommandAlreadyRunning = errors.New("command already running")
)
