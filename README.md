# hypermcp - Reusable MCP Server Infrastructure

`hypermcp` is a reusable package that provides common infrastructure for building Model Context Protocol (MCP) servers. It handles all the boilerplate so you can focus on implementing your custom tools and resources.

## Features

- ✅ MCP server setup and lifecycle management
- ✅ HTTP client with logging
- ✅ Caching layer (with optional disable)
- ✅ Transport abstraction (stdio, future: Streamable HTTP)
- ✅ Structured logging with zap
- ✅ Helper methods for registering tools and resources
- ✅ Automatic stats tracking

## Quick Start

### 1. Create a New MCP Server

```go
package main

import (
    "context"
    "github.com/rayprogramming/hypermcp"
    "github.com/rayprogramming/hypermcp/cache"
    "go.uber.org/zap"
)

func main() {
    // Setup logger
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    // Configure server
    cfg := hypermcp.Config{
        Name:         "my-mcp-server",
        Version:      "1.0.0",
        CacheEnabled: true,
        CacheConfig: cache.Config{
            MaxCost:     100 * 1024 * 1024, // 100MB
            NumCounters: 10_000,
            BufferItems: 64,
        },
    }

    // Create base server
    srv, err := hypermcp.New(cfg, logger)
    if err != nil {
        logger.Fatal("failed to create server", zap.Error(err))
    }

    // Register your tools and resources
    registerFeatures(srv)

    // Log registration stats
    srv.LogRegistrationStats()

    // Run with stdio transport
    ctx := context.Background()
    if err := hypermcp.RunWithTransport(ctx, srv, hypermcp.TransportStdio, logger); err != nil {
        logger.Fatal("server failed", zap.Error(err))
    }
}
```

### 2. Implement Your Providers

Create providers that use the shared infrastructure:

```go
package providers

import (
    "context"
    "github.com/modelcontextprotocol/go-sdk/mcp"
    "github.com/rayprogramming/hypermcp"
    "github.com/rayprogramming/hypermcp/cache"
    "github.com/rayprogramming/hypermcp/httpx"
    "go.uber.org/zap"
)

type MyProvider struct {
    httpClient *httpx.Client
    cache      *cache.Cache
    logger     *zap.Logger
}

func NewMyProvider(srv *hypermcp.Server) *MyProvider {
    return &MyProvider{
        httpClient: srv.HTTPClient(),
        cache:      srv.Cache(),
        logger:     srv.Logger(),
    }
}

func (p *MyProvider) MyTool(
    ctx context.Context,
    req *mcp.CallToolRequest,
    input MyToolInput,
) (*mcp.CallToolResult, MyToolOutput, error) {
    // Your implementation here
    // Use p.httpClient, p.cache, p.logger as needed
}
```

### 3. Register Your Features

```go
func registerFeatures(srv *hypermcp.Server) {
    // Create providers
    myProvider := providers.NewMyProvider(srv)

    // Register tools using the helper function
    hypermcp.AddTool(
        srv,
        &mcp.Tool{
            Name:        "my_tool",
            Description: "Does something cool",
        },
        myProvider.MyTool,
    )

    // Register resources using the helper method
    srv.AddResource(
        &mcp.Resource{
            URI:         "myresource://data",
            Name:        "My Resource",
            Description: "Provides some data",
            MIMEType:    "application/json",
        },
        func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
            // Your resource implementation
        },
    )
}
```

## API Reference

### Server Creation

```go
func New(cfg Config, logger *zap.Logger) (*Server, error)
```

Creates a new MCP server with common infrastructure.

### Configuration

```go
type Config struct {
    Name         string        // Server name
    Version      string        // Server version
    CacheEnabled bool          // Enable caching
    CacheConfig  cache.Config  // Cache configuration
}
```

### Server Methods

- `HTTPClient() *httpx.Client` - Get the shared HTTP client
- `Cache() *cache.Cache` - Get the cache instance
- `Logger() *zap.Logger` - Get the logger
- `Metrics() *Metrics` - Get metrics instance for tracking
- `GetMetrics() MetricsSnapshot` - Get snapshot of current metrics
- `MCP() *mcp.Server` - Get the underlying MCP server
- `AddResource(resource, handler)` - Register a resource (auto-increments counter)
- `AddResourceTemplate(template, handler)` - Register a resource template (auto-increments counter)
- `LogRegistrationStats()` - Log tool/resource counts
- `Run(ctx, transport)` - Start the server
- `Shutdown(ctx)` - Gracefully shutdown (closes cache, logs final stats)

### Package-Level Functions

- `AddTool[In, Out](srv, tool, handler)` - Register a tool (auto-increments counter)
- `New(cfg, logger)` - Create a new server instance
- `RunWithTransport(ctx, srv, transportType, logger)` - Start server with specified transport

### Transport

```go
func RunWithTransport(ctx context.Context, srv *Server, transportType TransportType, logger *zap.Logger) error
```

Starts the server with the specified transport.

Available transports:
- `TransportStdio` - Standard input/output (recommended for most use cases)
- `TransportStreamableHTTP` - Streamable HTTP (for servers handling multiple client connections, not yet implemented)

## Transport Types

`hypermcp` supports the MCP specification's recommended transports:

### Stdio Transport (Recommended)
- **Default choice** for most MCP servers
- Client launches server as subprocess
- Communication over stdin/stdout
- Simpler setup and deployment
- Clients SHOULD support stdio whenever possible (per MCP spec)

### Streamable HTTP Transport
- For servers handling multiple concurrent clients
- HTTP-based with optional Server-Sent Events
- Replaces the deprecated HTTP+SSE transport
- More complex but supports advanced scenarios
- **Note**: Not yet implemented in this package

## Benefits

### For You
- 🚀 **Fast Setup**: Get a server running in minutes, not hours
- 🔧 **Focus on Features**: Spend time on tools, not boilerplate
- 📦 **Batteries Included**: HTTP client, caching, logging all configured
- 🎯 **Best Practices**: Follows MCP patterns and Go idioms

### For Your Users
- ⚡ **Performance**: Built-in caching and connection pooling
- 📊 **Observability**: Structured logging with zap
- 🛡️ **Reliability**: Proper error handling and graceful shutdown

## Examples

The `examples/` directory contains complete, working examples:

### [hello](examples/hello/) - Basic Server
Minimal example showing:
- Server setup with configuration
- Simple tool registration
- Graceful shutdown handling

### [weather](examples/weather/) - Caching Demo
Demonstrates:
- Enabling and using the cache
- Cache hit/miss patterns
- Multiple tools in one server
- Metrics logging on shutdown

### [metrics](examples/metrics/) - Metrics & Monitoring
Shows how to:
- Track tool invocations
- Monitor cache performance
- Expose metrics via tools
- Log periodic statistics
- Access cache-specific metrics

### [fileserver](examples/fileserver/) - Resource Provider
Example of:
- Resource registration
- Resource templates with parameters
- Working with files and data

Run any example:
```bash
cd examples/hello && go run main.go
cd examples/weather && go run main.go
cd examples/metrics && go run main.go
```

## Performance Metrics

`hypermcp` includes built-in performance tracking:

```go
// Track operations
srv.Metrics().IncrementToolInvocations()
srv.Metrics().IncrementCacheHits()
srv.Metrics().IncrementCacheMisses()
srv.Metrics().IncrementErrors()

// Get snapshot
metrics := srv.GetMetrics()
fmt.Printf("Uptime: %v\n", metrics.Uptime)
fmt.Printf("Cache hit rate: %.2f%%\n", metrics.CacheHitRate*100)
```

Metrics tracked:
- Server uptime
- Tool invocations
- Resource reads
- Cache hits/misses and hit rate
- Error counts

## Best Practices

### Graceful Shutdown

Always implement graceful shutdown to ensure resources are properly cleaned up:

```go
// Setup signal handling
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

// Graceful shutdown with timeout
shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
defer shutdownCancel()

if err := srv.Shutdown(shutdownCtx); err != nil {
    logger.Error("shutdown error", zap.Error(err))
}
```

### Cache Usage

Use caching for expensive operations:

```go
cacheKey := fmt.Sprintf("data:%s", id)

// Check cache first
if cached, ok := srv.Cache().Get(cacheKey); ok {
    srv.Metrics().IncrementCacheHits()
    return cached
}
srv.Metrics().IncrementCacheMisses()

// Fetch data...
result := fetchExpensiveData(id)

// Cache for 5 minutes
srv.Cache().Set(cacheKey, result, 5*time.Minute)
```

### HTTP Client Usage

The provided HTTP client includes retries and proper timeouts:

```go
type Response struct {
    Status string `json:"status"`
}

var resp Response
if err := srv.HTTPClient().Get(ctx, apiURL, &resp); err != nil {
    srv.Metrics().IncrementErrors()
    return err
}
```

## Examples

## Dependencies

- `github.com/modelcontextprotocol/go-sdk` - MCP SDK
- `go.uber.org/zap` - Structured logging
- `github.com/dgraph-io/ristretto` - Caching (via pkg/cache)
