package hypermcp

import (
	"context"
	"testing"

	"go.uber.org/zap/zaptest"
)

func TestRunWithTransport_Stdio(t *testing.T) {
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

	// Skip actual transport test when running with coverage
	// because stdio transport captures stdout, breaking coverage output
	if testing.CoverMode() != "" {
		t.Skip("Skipping stdio transport test when coverage is enabled")
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

	expectedMsg := "Streamable HTTP transport not yet implemented"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
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
