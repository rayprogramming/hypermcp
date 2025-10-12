package hypermcp

import (
	"errors"
	"fmt"
)

// Common error variables for type checking.
var (
	// ErrInvalidConfig indicates that the server configuration is invalid.
	ErrInvalidConfig = errors.New("invalid configuration")

	// ErrServerNotRunning indicates an operation was attempted on a non-running server.
	ErrServerNotRunning = errors.New("server not running")

	// ErrShutdownTimeout indicates the server shutdown exceeded the timeout.
	ErrShutdownTimeout = errors.New("shutdown timeout exceeded")

	// ErrTransportNotSupported indicates the requested transport type is not implemented.
	ErrTransportNotSupported = errors.New("transport not supported")
)

// ConfigError wraps configuration validation errors with context.
type ConfigError struct {
	Err   error
	Field string
}

func (e *ConfigError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("config error in field %q: %v", e.Field, e.Err)
	}
	return fmt.Sprintf("config error: %v", e.Err)
}

func (e *ConfigError) Unwrap() error {
	return e.Err
}

// NewConfigError creates a new configuration error.
func NewConfigError(field string, err error) *ConfigError {
	return &ConfigError{
		Field: field,
		Err:   err,
	}
}

// TransportError wraps transport-related errors.
type TransportError struct {
	Err       error
	Transport TransportType
}

func (e *TransportError) Error() string {
	return fmt.Sprintf("transport error (%s): %v", e.Transport, e.Err)
}

func (e *TransportError) Unwrap() error {
	return e.Err
}

// NewTransportError creates a new transport error.
func NewTransportError(transport TransportType, err error) *TransportError {
	return &TransportError{
		Transport: transport,
		Err:       err,
	}
}
