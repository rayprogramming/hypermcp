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
- `MCP() *mcp.Server` - Get the underlying MCP server
- `AddResource(resource, handler)` - Register a resource (auto-increments counter)
- `AddResourceTemplate(template, handler)` - Register a resource template (auto-increments counter)
- `LogRegistrationStats()` - Log tool/resource counts
- `Run(ctx, transport)` - Start the server
- `Shutdown(ctx)` - Gracefully shutdown

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

See the `hypermcp` implementation in this repository for a complete example with:
- WIP

## Dependencies

- `github.com/modelcontextprotocol/go-sdk` - MCP SDK
- `go.uber.org/zap` - Structured logging
- `github.com/dgraph-io/ristretto` - Caching (via pkg/cache)
