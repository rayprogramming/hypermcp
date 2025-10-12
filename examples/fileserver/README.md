# File Server Example

An MCP server that demonstrates resource registration and resource templates.

## What This Example Demonstrates

- Resource registration with static URIs
- Resource templates with URI parameters
- Reading directory contents
- Reading specific files
- Security considerations (path traversal prevention)
- Resource read metrics

## Resources

### 1. Examples Directory
- **URI**: `file:///examples`
- **Description**: Lists all examples in the examples directory
- **Format**: Plain text directory listing

### 2. Example File (Template)
- **URI**: `file:///examples/{filename}`
- **Description**: Read a specific example's README
- **Format**: Plain text file content
- **Example**: `file:///examples/hello` reads `hello/README.md`

## Running the Example

```bash
cd examples/fileserver
go run main.go
```

## Testing with MCP Inspector

```bash
mcp-inspector go run main.go
```

Then use the resource browser to:
1. View `file:///examples` to see all examples
2. View `file:///examples/hello` to read the hello example's README

## Security Notes

This example includes basic security measures:
- Path traversal prevention
- Restricts access to examples directory only
- Validates filename format

In a production server, you should add:
- More robust URI parsing
- Authentication/authorization
- Rate limiting
- Input validation
- Logging of access attempts

## Implementation Details

### Resource vs Resource Template

- **Resource**: Fixed URI, like `file:///examples`
- **Resource Template**: Parameterized URI, like `file:///examples/{filename}`

### URI Template Parsing

This example uses a simplified URI parsing approach. For production use, implement proper URI template parsing according to RFC 6570.

## Next Steps

- Add write operations (as tools, not resources)
- Support more file types with appropriate MIME types
- Add resource listing capabilities
- Implement proper URI template parsing
- Add authentication

See the main [README](../../README.md) for more information about building MCP servers.
