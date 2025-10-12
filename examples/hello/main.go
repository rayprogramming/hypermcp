package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rayprogramming/hypermcp"
	"go.uber.org/zap"
)

func main() {
	// Create logger
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Create server configuration
	cfg := hypermcp.Config{
		Name:         "hello-server",
		Version:      "1.0.0",
		CacheEnabled: false,
	}

	// Create server
	srv, err := hypermcp.New(cfg, logger)
	if err != nil {
		logger.Fatal("failed to create server", zap.Error(err))
	}

	// Register a simple hello tool
	hypermcp.AddTool(srv, &mcp.Tool{
		Name:        "hello",
		Description: "Say hello to someone",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "Name of the person to greet",
				},
			},
			"required": []string{"name"},
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, input struct {
		Name string `json:"name"`
	}) (*mcp.CallToolResult, any, error) {
		greeting := fmt.Sprintf("Hello, %s! 👋", input.Name)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: greeting,
				},
			},
		}, nil, nil
	})

	// Log registration stats
	srv.LogRegistrationStats()

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("shutting down...")
		cancel()
	}()

	// Run server
	if err := hypermcp.RunWithTransport(ctx, srv, hypermcp.TransportStdio, logger); err != nil {
		logger.Error("server error", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("server stopped")
}
