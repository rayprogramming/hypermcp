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
	"github.com/rayprogramming/hypermcp/cache"
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
		Name:         "metrics-demo-server",
		Version:      "1.0.0",
		CacheEnabled: true,
		CacheConfig: cache.Config{
			MaxCost:     50 * 1024 * 1024, // 50MB
			NumCounters: 10_000,
			BufferItems: 64,
		},
	}

	srv, err := hypermcp.New(cfg, logger)
	if err != nil {
		logger.Fatal("failed to create server", zap.Error(err))
	}

	// Register a tool that demonstrates metrics tracking
	hypermcp.AddTool(srv, &mcp.Tool{
		Name:        "process_data",
		Description: "Process data with caching and metrics tracking",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"key": map[string]any{
					"type":        "string",
					"description": "Data key to process",
				},
			},
			"required": []string{"key"},
		},
	}, func(ctx context.Context, req *mcp.CallToolRequest, input struct {
		Key string `json:"key"`
	}) (*mcp.CallToolResult, any, error) {
		cacheKey := fmt.Sprintf("data:%s", input.Key)

		// Track tool invocation
		srv.Metrics().IncrementToolInvocations()

		// Check cache
		if cached, ok := srv.Cache().Get(cacheKey); ok {
			srv.Metrics().IncrementCacheHits()
			logger.Debug("cache hit", zap.String("key", cacheKey))

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Cached result: %v", cached),
					},
				},
			}, nil, nil
		}

		srv.Metrics().IncrementCacheMisses()

		// Simulate data processing
		result := fmt.Sprintf("Processed data for key: %s at %s", input.Key, time.Now().Format(time.RFC3339))

		// Cache result for 2 minutes
		srv.Cache().Set(cacheKey, result, 2*time.Minute)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: result,
				},
			},
		}, nil, nil
	})

	// Register a tool to retrieve current metrics
	hypermcp.AddTool(srv, &mcp.Tool{
		Name:        "get_metrics",
		Description: "Retrieve current server metrics",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, any, error) {
		metrics := srv.GetMetrics()

		metricsText := fmt.Sprintf(`Server Metrics:
- Uptime: %v
- Tool Invocations: %d
- Resource Reads: %d
- Cache Hits: %d
- Cache Misses: %d
- Cache Hit Rate: %.2f%%
- Errors: %d`,
			metrics.Uptime,
			metrics.ToolInvocations,
			metrics.ResourceReads,
			metrics.CacheHits,
			metrics.CacheMisses,
			metrics.CacheHitRate*100,
			metrics.Errors,
		)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: metricsText,
				},
			},
		}, nil, nil
	})

	// Log registration stats
	srv.LogRegistrationStats()

	// Setup graceful shutdown with metrics logging
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start periodic metrics logging
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				metrics := srv.GetMetrics()
				logger.Info("periodic metrics",
					zap.Duration("uptime", metrics.Uptime),
					zap.Int64("tool_invocations", metrics.ToolInvocations),
					zap.Int64("cache_hits", metrics.CacheHits),
					zap.Int64("cache_misses", metrics.CacheMisses),
					zap.Float64("cache_hit_rate", metrics.CacheHitRate),
					zap.Int64("errors", metrics.Errors),
				)
			}
		}
	}()

	go func() {
		<-sigChan
		logger.Info("shutting down...")

		// Log final metrics
		metrics := srv.GetMetrics()
		logger.Info("final metrics",
			zap.Duration("uptime", metrics.Uptime),
			zap.Int64("tool_invocations", metrics.ToolInvocations),
			zap.Int64("resource_reads", metrics.ResourceReads),
			zap.Int64("cache_hits", metrics.CacheHits),
			zap.Int64("cache_misses", metrics.CacheMisses),
			zap.Float64("cache_hit_rate", metrics.CacheHitRate),
			zap.Int64("errors", metrics.Errors),
		)

		// Get cache metrics
		if srv.Cache() != nil {
			cacheMetrics := srv.Cache().Metrics()
			logger.Info("cache metrics",
				zap.Uint64("hits", cacheMetrics.Hits()),
				zap.Uint64("misses", cacheMetrics.Misses()),
				zap.Float64("ratio", cacheMetrics.Ratio()),
			)
		}

		cancel()
	}()

	// Run server
	if err := hypermcp.RunWithTransport(ctx, srv, hypermcp.TransportStdio, logger); err != nil {
		logger.Error("server error", zap.Error(err))
		os.Exit(1)
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", zap.Error(err))
	}

	logger.Info("server stopped")
}
