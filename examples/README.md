# Examples

This directory contains example MCP servers built with hypermcp.

## Available Examples

### 1. [Hello World](./hello/) - Basic Server
A minimal MCP server that demonstrates:
- Basic server setup
- Single tool registration
- Stdio transport

**Run it:**
```bash
cd examples/hello
go run main.go
```

### 2. [Weather Server](./weather/) - HTTP Client Usage
A server that demonstrates:
- External API calls with HTTP client
- Caching API responses
- Error handling
- Multiple tools

**Run it:**
```bash
cd examples/weather
go run main.go
```

### 3. [File Server](./fileserver/) - Resources
A server that demonstrates:
- Resource registration
- Resource templates with parameters
- Reading and listing files
- Proper error handling

**Run it:**
```bash
cd examples/fileserver
go run main.go
```

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
