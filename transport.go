// Package hypermcp provides reusable MCP server infrastructure
package hypermcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"
)

// Transport Notes (MCP Specification 2025-06-18):
//
// The MCP protocol defines two standard transports:
//
// 1. stdio - Communication over standard input/output
//    - Client launches MCP server as a subprocess
//    - Server reads JSON-RPC messages from stdin, writes to stdout
//    - Messages are delimited by newlines
//    - Recommended for most use cases
//    - Clients SHOULD support stdio whenever possible
//
// 2. Streamable HTTP - HTTP-based transport for multiple client connections
//    - Replaces the deprecated HTTP+SSE transport (from protocol version 2024-11-05)
//    - Server operates as independent process handling multiple clients
//    - Uses HTTP POST/GET requests with optional Server-Sent Events
//    - Suitable for servers that need to handle multiple concurrent client connections
//
// Note: The old HTTP+SSE transport is deprecated but servers can maintain backwards
// compatibility by supporting both old and new transports.

// TransportType defines the type of transport to use.
//
// The MCP specification defines two standard transports: stdio and Streamable HTTP.
type TransportType string

const (
	// TransportStdio uses standard input/output for communication.
	// This is the recommended transport for most MCP servers where the client
	// launches the server as a subprocess.
	TransportStdio TransportType = "stdio"

	// TransportStreamableHTTP uses HTTP-based transport for multiple client connections.
	// This replaces the deprecated HTTP+SSE transport and is suitable for servers
	// that need to handle multiple concurrent clients.
	// Note: Not yet implemented in this library.
	TransportStreamableHTTP TransportType = "streamable-http"
)

// RunWithTransport starts the MCP server with the specified transport.
//
// The function logs the selected transport and blocks until the context is canceled
// or an error occurs. Currently only stdio transport is implemented.
func RunWithTransport(ctx context.Context, srv *Server, transportType TransportType, logger *zap.Logger) error {
	var transport mcp.Transport

	switch transportType {
	case TransportStdio:
		logger.Info("using stdio transport (recommended)")
		transport = &mcp.StdioTransport{}
	case TransportStreamableHTTP:
		return NewTransportError(transportType, ErrTransportNotSupported)
	default:
		return NewTransportError(transportType, fmt.Errorf("unknown transport type"))
	}

	logger.Info("server ready")

	if err := srv.Run(ctx, transport); err != nil {
		return fmt.Errorf("server run failed: %w", err)
	}

	return nil
}
