# hypermcp Examples

This directory contains complete, working examples demonstrating different features of hypermcp.

## Examples Overview

### [hello/](hello/) - Basic Server
**Best for**: Getting started, understanding basic concepts

A minimal MCP server example that demonstrates:
- Server configuration and initialization
- Simple tool registration with the `AddTool` helper
- Proper signal handling for graceful shutdown
- Basic logging patterns

Perfect starting point for new users.

### [weather/](weather/) - Caching Demo
**Best for**: Learning caching patterns, multiple tools

A weather information server that shows:
- Enabling and configuring the cache
- Cache hit/miss tracking with metrics
- Multiple tool registration in one server
- Mock data patterns (easily adaptable to real APIs)
- Metrics logging on shutdown

Demonstrates the performance benefits of caching.

### [metrics/](metrics/) - Metrics & Monitoring
**Best for**: Understanding performance tracking, production deployments

A comprehensive example showing:
- Real-time metrics tracking (tool invocations, cache performance)
- Exposing metrics via MCP tools
- Periodic metrics logging (every 30 seconds)
- Accessing Ristretto cache-specific metrics
- Best practices for production monitoring

Essential for production deployments.

### [fileserver/](fileserver/) - Resource Provider
**Best for**: Working with resources, file operations

Demonstrates:
- Resource registration
- Resource templates with URI parameters
- File reading and serving
- Error handling for missing resources

## Running the Examples

Each example is a standalone Go program:

```bash
# Run hello example
cd hello && go run main.go

# Run weather example
cd weather && go run main.go

# Run metrics example
cd metrics && go run main.go

# Run fileserver example
cd fileserver && go run main.go
```

## Testing with MCP Inspector

The [MCP Inspector](https://github.com/modelcontextprotocol/inspector) is a useful tool for testing MCP servers:

1. Install the MCP Inspector
2. Configure it to connect to your server via stdio
3. Explore the available tools and resources
4. Test tool invocations and resource reads

## Example-Specific READMEs

Each example has its own README with:
- Detailed feature explanations
- Example tool invocations
- Expected outputs
- Learning objectives

Check each example's directory for more information.

## Building Examples

Build all examples to verify they compile:

```bash
# From examples directory
for dir in hello weather metrics fileserver; do
    echo "Building $dir..."
    (cd $dir && go build)
done
```

Or use the project Makefile from the root:

```bash
cd .. && make build
```

## Common Patterns

### Tool Registration

```go
hypermcp.AddTool(srv, &mcp.Tool{
    Name:        "my_tool",
    Description: "Description of what this tool does",
    InputSchema: map[string]any{
        "type": "object",
        "properties": map[string]any{
            "param": map[string]any{
                "type": "string",
                "description": "Parameter description",
            },
        },
        "required": []string{"param"},
    },
}, func(ctx context.Context, req *mcp.CallToolRequest, input struct {
    Param string `json:"param"`
}) (*mcp.CallToolResult, any, error) {
    // Tool implementation
})
```

### Resource Registration

```go
srv.AddResource(&mcp.Resource{
    URI:  "myapp://data",
    Name: "Application Data",
    Description: "Description of the resource",
}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
    // Resource handler
})
```

### Graceful Shutdown Pattern

All examples demonstrate this pattern:

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

go func() {
    <-sigChan
    logger.Info("shutting down...")
    cancel()
}()

if err := hypermcp.RunWithTransport(ctx, srv, hypermcp.TransportStdio, logger); err != nil {
    logger.Error("server error", zap.Error(err))
    os.Exit(1)
}

shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
defer shutdownCancel()

if err := srv.Shutdown(shutdownCtx); err != nil {
    logger.Error("shutdown error", zap.Error(err))
}
```

## Learning Path

Recommended order for learning:

1. **hello/** - Understand the basics
2. **weather/** - Learn about caching
3. **fileserver/** - Work with resources
4. **metrics/** - Production-ready patterns

## Contributing Examples

Have an interesting use case? Consider contributing an example! See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines.

## Testing Examples

You can test any example server using the MCP Inspector or by configuring it in Claude Desktop.

### Using MCP Inspector

1. Install the MCP Inspector:
   ```bash
   npm install -g @modelcontextprotocol/inspector
   ```

2. Run the inspector with your example:
   ```bash
   cd examples/hello
   mcp-inspector go run main.go
   ```

### Using with Claude Desktop

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "hello-example": {
      "command": "go",
      "args": ["run", "/path/to/hypermcp/examples/hello/main.go"]
    }
  }
}
```

## Creating Your Own

Use these examples as templates for your own MCP servers. Each example is self-contained and can be copied as a starting point.

See the [main documentation](../../EXAMPLE.md) for more details on building MCP servers with hypermcp.
