# Example: Creating a Simple MCP Server with hypermcp

This is a minimal example showing how to create an MCP server using the `hypermcp` package.

## Setup

```bash
mkdir my-mcp-server
cd my-mcp-server
go mod init github.com/myuser/my-mcp-server
```

## Install Dependencies

```bash
go get github.com/rayprogramming/hypermcp
go get github.com/rayprogramming/hypermcp/cache
go get github.com/rayprogramming/hypermcp/httpx
go get github.com/modelcontextprotocol/go-sdk
go get go.uber.org/zap
```

## Create main.go

\`\`\`go
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rayprogramming/hypermcp/cache"
	"github.com/rayprogramming/hypermcp"
	"go.uber.org/zap"
)

func main() {
	// Parse flags
	showVersion := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println("my-mcp-server v1.0.0")
		os.Exit(0)
	}

	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Configure server
	cfg := hypermcp.Config{
		Name:         "my-mcp-server",
		Version:      "1.0.0",
		CacheEnabled: true,
		CacheConfig:  cache.DefaultConfig(),
	}

	// Create base server
	srv, err := hypermcp.New(cfg, logger)
	if err != nil {
		logger.Fatal("failed to create server", zap.Error(err))
	}

	// Register a simple tool
	registerTools(srv, logger)

	// Log registration stats
	srv.LogRegistrationStats()

	// Run with stdio transport
	ctx := context.Background()
	if err := hypermcp.RunWithTransport(ctx, srv, hypermcp.TransportStdio, logger); err != nil {
		logger.Fatal("server failed", zap.Error(err))
	}
}

// Tool input/output schemas
type EchoInput struct {
	Message string \`json:"message" jsonschema:"the message to echo back"\`
}

type EchoOutput struct {
	Echo string \`json:"echo" jsonschema:"the echoed message"\`
}

// registerTools registers all tools with the MCP server
func registerTools(srv *hypermcp.Server, logger *zap.Logger) {
	// Register a simple echo tool using the helper function
	hypermcp.AddTool(
		srv,
		&mcp.Tool{
			Name:        "echo",
			Description: "Echo back a message",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, input EchoInput) (*mcp.CallToolResult, EchoOutput, error) {
			logger.Info("echo tool called", zap.String("message", input.Message))
			return nil, EchoOutput{Echo: input.Message}, nil
		},
	)

	// Add more tools here...
}
```

## Build and Run

```bash
go build -o my-mcp-server
./my-mcp-server
```

## Test It

You can test your MCP server by sending JSON-RPC messages to it:

```bash
# List available tools
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | ./my-mcp-server

# Call the echo tool
echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"echo","arguments":{"message":"Hello, World!"}}}' | ./my-mcp-server
```

## Configure for VS Code

Add to your MCP client configuration (e.g., `~/.config/Code/User/globalStorage/github.copilot-chat/mcpServers.json`):

```json
{
  "mcpServers": {
    "my-mcp-server": {
      "command": "/path/to/my-mcp-server",
      "args": [],
      "env": {}
    }
  }
}
```

Or if using Claude Desktop, add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "my-mcp-server": {
      "command": "/path/to/my-mcp-server",
      "args": []
    }
  }
}
```

## Advanced Example: Using HTTP Client and Cache

Here's a more advanced example showing how to use the shared infrastructure:

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rayprogramming/hypermcp"
	"github.com/rayprogramming/hypermcp/cache"
	"go.uber.org/zap"
)

type WeatherInput struct {
	City string `json:"city" jsonschema:"the city to get weather for"`
}

type WeatherOutput struct {
	City        string  `json:"city"`
	Temperature float64 `json:"temperature"`
	Conditions  string  `json:"conditions"`
}

type WeatherAPI struct {
	Temp       float64 `json:"temp"`
	Conditions string  `json:"conditions"`
}

func registerAdvancedTools(srv *hypermcp.Server, logger *zap.Logger) {
	// Weather tool with caching and HTTP
	hypermcp.AddTool(
		srv,
		&mcp.Tool{
			Name:        "get_weather",
			Description: "Get current weather for a city",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, input WeatherInput) (*mcp.CallToolResult, WeatherOutput, error) {
			cacheKey := fmt.Sprintf("weather:%s", input.City)

			// Check cache first
			if cached, ok := srv.Cache().Get(cacheKey); ok {
				logger.Info("cache hit for weather", zap.String("city", input.City))
				if weather, ok := cached.(WeatherOutput); ok {
					return nil, weather, nil
				}
			}

			// Fetch from API
			logger.Info("fetching weather from API", zap.String("city", input.City))
			var apiResp WeatherAPI
			url := fmt.Sprintf("https://api.weather.example.com/current?city=%s", input.City)
			
			if err := srv.HTTPClient().Get(ctx, url, &apiResp); err != nil {
				return nil, WeatherOutput{}, fmt.Errorf("failed to fetch weather: %w", err)
			}

			weather := WeatherOutput{
				City:        input.City,
				Temperature: apiResp.Temp,
				Conditions:  apiResp.Conditions,
			}

			// Cache for 5 minutes
			srv.Cache().Set(cacheKey, weather, 5*time.Minute)

			return nil, weather, nil
		},
	)
}
```

## Next Steps

1. **Add More Tools**: Create more tool functions following the examples above
2. **Add Resources**: Register resources to expose data to clients
3. **Error Handling**: Add comprehensive error handling and validation
4. **Logging**: Use structured logging throughout your handlers
5. **Testing**: Write tests for your tools and providers
6. **Documentation**: Document your tools and their parameters

## Common Patterns

### Provider Pattern

Organize related tools into provider structs:

```go
type WeatherProvider struct {
	httpClient *httpx.Client
	cache      *cache.Cache
	logger     *zap.Logger
	apiKey     string
}

func NewWeatherProvider(srv *hypermcp.Server, apiKey string) *WeatherProvider {
	return &WeatherProvider{
		httpClient: srv.HTTPClient(),
		cache:      srv.Cache(),
		logger:     srv.Logger(),
		apiKey:     apiKey,
	}
}

func (p *WeatherProvider) GetWeather(ctx context.Context, req *mcp.CallToolRequest, input WeatherInput) (*mcp.CallToolResult, WeatherOutput, error) {
	// Implementation here
	return nil, WeatherOutput{}, nil
}
```

### Resource Example

```go
srv.AddResource(
	&mcp.Resource{
		URI:         "myapp://status",
		Name:        "Server Status",
		Description: "Current server status and metrics",
		MIMEType:    "application/json",
	},
	func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		status := map[string]interface{}{
			"status":  "healthy",
			"uptime":  time.Since(startTime).String(),
			"version": "1.0.0",
		}
		
		data, _ := json.Marshal(status)
		
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      "myapp://status",
					MIMEType: "application/json",
					Text:     string(data),
				},
			},
		}, nil
	},
)
```

## Troubleshooting

### Server Not Starting
- Check that the binary has execute permissions
- Verify the MCP client configuration path is correct
- Check server logs for initialization errors

### Tools Not Appearing
- Ensure `srv.LogRegistrationStats()` shows the correct count
- Verify tool registration happens before `RunWithTransport()`
- Check for panic or errors during registration

### Cache Not Working
- Verify `CacheEnabled: true` in config
- Check cache TTL is not too short
- Use `srv.Cache().Metrics()` to inspect cache performance

## Full Working Example

See the `examples/` directory in the hypermcp repository for complete, runnable examples including:
- Simple echo server
- HTTP API integration
- Database integration
- Multi-provider architecture

Happy coding! 🚀
````
```

## Build and Run

```bash
go build -o my-mcp-server
./my-mcp-server
```

## Test It

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | ./my-mcp-server
```

## Configure for VS Code

Add to `~/.config/Code/User/globalStorage/github.copilot-chat/mcpServers.json`:

```json
{
  "mcpServers": {
    "my-mcp-server": {
      "command": "/path/to/my-mcp-server",
      "args": [],
      "env": {}
    }
  }
}
```

## Metrics and Observability

Track server performance with built-in metrics:

```go
// Get current metrics
metrics := srv.GetMetrics()
fmt.Printf("Uptime: %v\n", metrics.Uptime)
fmt.Printf("Tool invocations: %d\n", metrics.ToolInvocations)
fmt.Printf("Resource reads: %d\n", metrics.ResourceReads)
fmt.Printf("Cache hit rate: %.2f%%\n", metrics.CacheHitRate*100)
fmt.Printf("Errors: %d\n", metrics.Errors)

// Track custom metrics
srv.Metrics().IncrementToolInvocations()
srv.Metrics().IncrementCacheHits()
srv.Metrics().IncrementErrors()
```

Available metrics:
- **Uptime**: How long the server has been running
- **ToolInvocations**: Number of tool calls
- **ResourceReads**: Number of resource reads
- **CacheHits/Misses**: Cache performance
- **CacheHitRate**: Percentage of cache hits
- **Errors**: Total error count

## Next Steps

1. **Add More Tools**: Create more tool functions following the echo example
2. **Add Providers**: Organize related tools into provider structs
3. **Use HTTP Client**: Use `srv.HTTPClient()` for external API calls
4. **Use Cache**: Use `srv.Cache()` to cache expensive operations
5. **Add Resources**: Register resources alongside tools
6. **Monitor Metrics**: Track performance with `srv.GetMetrics()`

See `hypermcp` for a complete example with multiple providers, resources, and proper error handling!
