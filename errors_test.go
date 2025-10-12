package hypermcp

import (
	"context"
	"errors"
	"testing"
)

func TestConfigError(t *testing.T) {
	tests := []struct {
		err       error
		name      string
		field     string
		wantError string
	}{
		{
			name:      "with field",
			field:     "Name",
			err:       errors.New("cannot be empty"),
			wantError: `config error in field "Name": cannot be empty`,
		},
		{
			name:      "without field",
			field:     "",
			err:       errors.New("invalid configuration"),
			wantError: "config error: invalid configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewConfigError(tt.field, tt.err)
			if err.Error() != tt.wantError {
				t.Errorf("ConfigError.Error() = %q, want %q", err.Error(), tt.wantError)
			}

			// Test Unwrap
			if unwrapped := err.Unwrap(); unwrapped != tt.err {
				t.Errorf("ConfigError.Unwrap() = %v, want %v", unwrapped, tt.err)
			}

			// Test errors.Is
			if !errors.Is(err, tt.err) {
				t.Error("errors.Is() should match wrapped error")
			}
		})
	}
}

func TestTransportError(t *testing.T) {
	tests := []struct {
		err       error
		name      string
		transport TransportType
		wantError string
	}{
		{
			name:      "streamable http not supported",
			transport: TransportStreamableHTTP,
			err:       ErrTransportNotSupported,
			wantError: "transport error (streamable-http): transport not supported",
		},
		{
			name:      "unknown transport",
			transport: TransportType("unknown"),
			err:       errors.New("unknown transport type"),
			wantError: "transport error (unknown): unknown transport type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewTransportError(tt.transport, tt.err)
			if err.Error() != tt.wantError {
				t.Errorf("TransportError.Error() = %q, want %q", err.Error(), tt.wantError)
			}

			// Test Unwrap
			if unwrapped := err.Unwrap(); unwrapped != tt.err {
				t.Errorf("TransportError.Unwrap() = %v, want %v", unwrapped, tt.err)
			}

			// Test errors.Is
			if !errors.Is(err, tt.err) {
				t.Error("errors.Is() should match wrapped error")
			}
		})
	}
}

func TestErrorVariables(t *testing.T) {
	// Test that error variables are defined
	if ErrInvalidConfig == nil {
		t.Error("ErrInvalidConfig should not be nil")
	}
	if ErrServerNotRunning == nil {
		t.Error("ErrServerNotRunning should not be nil")
	}
	if ErrShutdownTimeout == nil {
		t.Error("ErrShutdownTimeout should not be nil")
	}
	if ErrTransportNotSupported == nil {
		t.Error("ErrTransportNotSupported should not be nil")
	}
}

func TestConfigValidationErrors(t *testing.T) {
	tests := []struct {
		wantError error
		config    Config
		name      string
	}{
		{
			name: "empty name",
			config: Config{
				Name:    "",
				Version: "1.0.0",
			},
			wantError: ErrInvalidConfig,
		},
		{
			name: "empty version",
			config: Config{
				Name:    "test",
				Version: "",
			},
			wantError: ErrInvalidConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.config, nil)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !errors.Is(err, tt.wantError) {
				t.Errorf("expected error to wrap %v, got %v", tt.wantError, err)
			}
		})
	}
}

func TestTransportErrors(t *testing.T) {
	tests := []struct {
		wantError error
		name      string
		transport TransportType
	}{
		{
			name:      "streamable http not supported",
			transport: TransportStreamableHTTP,
			wantError: ErrTransportNotSupported,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RunWithTransport(context.Background(), nil, tt.transport, nil)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !errors.Is(err, tt.wantError) {
				t.Errorf("expected error to wrap %v, got %v", tt.wantError, err)
			}
		})
	}
}
