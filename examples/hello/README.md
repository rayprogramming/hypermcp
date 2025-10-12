# Hello World Example

A minimal MCP server that demonstrates basic server setup with a single "hello" tool.

## What This Example Demonstrates

- Basic server configuration
- Tool registration with input schema
- Stdio transport setup
- Graceful shutdown handling
- Logging

## Running the Example

```bash
go run main.go
```

## Testing with MCP Inspector

```bash
npm install -g @modelcontextprotocol/inspector
mcp-inspector go run main.go
```

Then in the inspector, call the `hello` tool with:

```json
{
  "name": "World"
}
```

## Testing with Claude Desktop

Add to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "hello": {
      "command": "go",
      "args": ["run", "/absolute/path/to/examples/hello/main.go"]
    }
  }
}
```

Restart Claude Desktop and try:
> "Use the hello tool to greet Alice"

## Code Structure

- **Configuration**: Sets up server name and version
- **Tool Registration**: Registers a single "hello" tool
- **Handler Function**: Simple greeting logic
- **Shutdown Handling**: Graceful shutdown on SIGTERM/SIGINT
- **Transport**: Uses stdio for communication

## Next Steps

- Add more tools with different input schemas
- Use the HTTP client for external API calls
- Add caching for expensive operations
- Register resources alongside tools

See the [weather example](../weather/) for a more complex server.
