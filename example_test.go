package hypermcp_test

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rayprogramming/hypermcp"
	"github.com/rayprogramming/hypermcp/cache"
	"go.uber.org/zap"
)

// ExampleNew demonstrates basic server creation.
func ExampleNew() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg := hypermcp.Config{
		Name:         "example-server",
		Version:      "1.0.0",
		CacheEnabled: false, // Disable cache for simplicity
	}

	srv, err := hypermcp.New(cfg, logger)
	if err != nil {
		logger.Fatal("failed to create server", zap.Error(err))
	}

	fmt.Printf("Server created: %s v%s\n", cfg.Name, cfg.Version)
	_ = srv // Use the server
	// Output: Server created: example-server v1.0.0
}

// ExampleServer_Cache demonstrates cache usage.
func ExampleServer_Cache() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	cfg := hypermcp.Config{
		Name:         "cache-example",
		Version:      "1.0.0",
		CacheEnabled: true,
		CacheConfig: cache.Config{
			MaxCost:     10 * 1024 * 1024, // 10MB
			NumCounters: 10000,
			BufferItems: 64,
		},
	}

	srv, _ := hypermcp.New(cfg, logger)

	// Set a value in cache
	cacheKey := "user:123"
	userData := map[string]string{"name": "John", "email": "john@example.com"}
	srv.Cache().Set(cacheKey, userData, 5*time.Minute)

	// Wait for cache to process (Ristretto is async)
	time.Sleep(10 * time.Millisecond)

	// Retrieve from cache
	if cached, ok := srv.Cache().Get(cacheKey); ok {
		fmt.Println("Cache hit!")
		_ = cached
	}
	// Output: Cache hit!
}

// ExampleServer_GetMetrics demonstrates metrics tracking.
func ExampleServer_GetMetrics() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	cfg := hypermcp.Config{
		Name:         "metrics-example",
		Version:      "1.0.0",
		CacheEnabled: false,
	}

	srv, _ := hypermcp.New(cfg, logger)

	// Simulate some operations
	srv.Metrics().IncrementToolInvocations()
	srv.Metrics().IncrementToolInvocations()
	srv.Metrics().IncrementResourceReads()

	// Get metrics snapshot
	metrics := srv.GetMetrics()
	fmt.Printf("Tool invocations: %d\n", metrics.ToolInvocations)
	fmt.Printf("Resource reads: %d\n", metrics.ResourceReads)
	// Output: Tool invocations: 2
	// Resource reads: 1
}

// ExampleAddTool demonstrates tool registration.
func ExampleAddTool() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	cfg := hypermcp.Config{
		Name:         "tool-example",
		Version:      "1.0.0",
		CacheEnabled: false,
	}

	srv, _ := hypermcp.New(cfg, logger)

	type EchoInput struct {
		Message string `json:"message"`
	}
	type EchoOutput struct {
		Echo string `json:"echo"`
	}

	tool := &mcp.Tool{
		Name:        "echo",
		Description: "Echoes back the input message",
	}

	handler := func(ctx context.Context, req *mcp.CallToolRequest, input EchoInput) (*mcp.CallToolResult, EchoOutput, error) {
		return nil, EchoOutput{Echo: input.Message}, nil
	}

	hypermcp.AddTool(srv, tool, handler)

	fmt.Println("Tool registered successfully")
	// Output: Tool registered successfully
}

// ExampleServer_AddResource demonstrates resource registration.
func ExampleServer_AddResource() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	cfg := hypermcp.Config{
		Name:         "resource-example",
		Version:      "1.0.0",
		CacheEnabled: false,
	}

	srv, _ := hypermcp.New(cfg, logger)

	resource := &mcp.Resource{
		URI:         "example://data",
		Name:        "Example Data",
		Description: "Example data resource",
		MIMEType:    "text/plain",
	}

	handler := func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      req.Params.URI,
					MIMEType: "text/plain",
					Text:     "Example data content",
				},
			},
		}, nil
	}

	srv.AddResource(resource, handler)

	fmt.Println("Resource registered successfully")
	// Output: Resource registered successfully
}

// ExampleConfig_Validate demonstrates configuration validation.
func ExampleConfig_Validate() {
	// Valid configuration
	validCfg := hypermcp.Config{
		Name:    "my-server",
		Version: "1.0.0",
	}

	if err := validCfg.Validate(); err != nil {
		fmt.Println("Validation failed")
	} else {
		fmt.Println("Valid configuration")
	}

	// Invalid configuration (missing name)
	invalidCfg := hypermcp.Config{
		Version: "1.0.0",
	}

	if err := invalidCfg.Validate(); err != nil {
		fmt.Println("Invalid configuration")
	}

	// Output:
	// Valid configuration
	// Invalid configuration
}
