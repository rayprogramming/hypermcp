package hypermcp

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap/zaptest"
)

func TestRunWithTransport_Stdio(t *testing.T) {
	// Skip this test - stdio transport captures stdout which interferes with test execution
	t.Skip("Skipping stdio transport test - it interferes with test output")

	logger := zaptest.NewLogger(t)
	cfg := Config{
		Name:         "test-server",
		Version:      "1.0.0",
		CacheEnabled: false,
	}

	srv, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// We can't actually run the server in a test (it would block),
	// but we can verify the function accepts the correct transport type
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately to prevent blocking

	// This will fail quickly due to canceled context, but that's expected
	_ = RunWithTransport(ctx, srv, TransportStdio, logger)
}

func TestRunWithTransport_StreamableHTTP(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := Config{
		Name:         "test-server",
		Version:      "1.0.0",
		CacheEnabled: false,
	}

	srv, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	ctx := context.Background()

	// Should return error for unimplemented transport
	err = RunWithTransport(ctx, srv, TransportStreamableHTTP, logger)
	if err == nil {
		t.Error("expected error for unimplemented transport")
	}

	// Check that it's the right error type
	if !errors.Is(err, ErrTransportNotSupported) {
		t.Errorf("expected ErrTransportNotSupported, got %v", err)
	}
}

func TestRunWithTransport_UnknownTransport(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := Config{
		Name:         "test-server",
		Version:      "1.0.0",
		CacheEnabled: false,
	}

	srv, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	ctx := context.Background()

	// Should return error for unknown transport
	err = RunWithTransport(ctx, srv, TransportType("unknown"), logger)
	if err == nil {
		t.Error("expected error for unknown transport")
	}
}

func TestTransportType_Constants(t *testing.T) {
	// Verify transport type constants are defined correctly
	if TransportStdio != "stdio" {
		t.Errorf("expected TransportStdio to be 'stdio', got %q", TransportStdio)
	}

	if TransportStreamableHTTP != "streamable-http" {
		t.Errorf("expected TransportStreamableHTTP to be 'streamable-http', got %q", TransportStreamableHTTP)
	}
}
