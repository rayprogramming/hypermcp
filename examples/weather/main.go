package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// Create server with caching enabled
	cfg := hypermcp.Config{
		Name:         "weather-server",
		Version:      "1.0.0",
		CacheEnabled: true,
	}

	srv, err := hypermcp.New(cfg, logger)
	if err != nil {
		logger.Fatal("failed to create server", zap.Error(err))
	}

	// Register weather tool
	hypermcp.AddTool(srv, &mcp.Tool{
		Name:        "get_weather",
		Description: "Get current weather for a city (demo - returns mock data)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"city": map[string]interface{}{
					"type":        "string",
					"description": "City name",
				},
			},
			"required": []string{"city"},
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, input struct {
		City string `json:"city"`
	}) (*mcp.CallToolResult, any, error) {
		// Check cache first
		cacheKey := fmt.Sprintf("weather:%s", input.City)
		if cached, ok := srv.Cache().Get(cacheKey); ok {
			logger.Debug("returning cached weather", zap.String("city", input.City))
			srv.Metrics().IncrementCacheHits()

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: cached.(string),
					},
				},
			}, nil, nil
		}
		srv.Metrics().IncrementCacheMisses()

		// In a real implementation, you would use srv.HTTPClient() here
		// to call an actual weather API. For this demo, we return mock data.
		weather := fmt.Sprintf("Weather in %s: ☀️ Sunny, 72°F (22°C)", input.City)

		// Cache for 5 minutes
		srv.Cache().Set(cacheKey, weather, 5*time.Minute)

		// Track metric
		srv.Metrics().IncrementToolInvocations()

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: weather,
				},
			},
		}, nil, nil
	})

	// Register forecast tool
	hypermcp.AddTool(srv, &mcp.Tool{
		Name:        "get_forecast",
		Description: "Get 3-day weather forecast (demo - returns mock data)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"city": map[string]interface{}{
					"type":        "string",
					"description": "City name",
				},
			},
			"required": []string{"city"},
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, input struct {
		City string `json:"city"`
	}) (*mcp.CallToolResult, any, error) {
		forecast := fmt.Sprintf("3-day forecast for %s:\n"+
			"Day 1: ☀️ Sunny, High: 75°F\n"+
			"Day 2: ⛅ Partly Cloudy, High: 72°F\n"+
			"Day 3: 🌧️ Rain, High: 68°F", input.City)

		srv.Metrics().IncrementToolInvocations()

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: forecast,
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

		// Log metrics before shutdown
		metrics := srv.GetMetrics()
		logger.Info("final metrics",
			zap.Duration("uptime", metrics.Uptime),
			zap.Int64("tool_invocations", metrics.ToolInvocations),
			zap.Int64("cache_hits", metrics.CacheHits),
			zap.Int64("cache_misses", metrics.CacheMisses),
			zap.Float64("cache_hit_rate", metrics.CacheHitRate),
		)

		cancel()
	}()

	// Run server
	if err := hypermcp.RunWithTransport(ctx, srv, hypermcp.TransportStdio, logger); err != nil {
		logger.Error("server error", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("server stopped")
}
