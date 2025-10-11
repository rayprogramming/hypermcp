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

## Next Steps

1. **Add More Tools**: Create more tool functions following the echo example
2. **Add Providers**: Organize related tools into provider structs
3. **Use HTTP Client**: Use `srv.HTTPClient()` for external API calls
4. **Use Cache**: Use `srv.Cache()` to cache expensive operations
5. **Add Resources**: Register resources alongside tools

See `hypermcp` for a complete example with multiple providers, resources, and proper error handling!
