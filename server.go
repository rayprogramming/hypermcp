// Package hypermcp provides reusable MCP server infrastructure.
//
// This package simplifies building Model Context Protocol (MCP) servers by providing
// common infrastructure components including HTTP client with retry logic, caching,
// structured logging, and transport abstraction.
//
// Example usage:
//
//	cfg := hypermcp.Config{
//	    Name:         "my-server",
//	    Version:      "1.0.0",
//	    CacheEnabled: true,
//	}
//	srv, err := hypermcp.New(cfg, logger)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	hypermcp.AddTool(srv, tool, handler)
//	srv.AddResource(resource, handler)
//	hypermcp.RunWithTransport(ctx, srv, hypermcp.TransportStdio, logger)
package hypermcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rayprogramming/hypermcp/cache"
	"github.com/rayprogramming/hypermcp/httpx"
	"go.uber.org/zap"
)

// Server wraps the MCP server with common infrastructure.
//
// It provides access to shared resources like HTTP client, cache, and logger,
// along with helper methods for registering tools and resources with automatic
// counter tracking.
type Server struct {
	mcp        *mcp.Server
	httpClient *httpx.Client
	cache      *cache.Cache
	logger     *zap.Logger
	config     Config

	// Stats for logging
	toolCount     int
	resourceCount int
}

// Config holds server configuration.
//
// Name and Version are required fields and will be validated.
// CacheEnabled determines whether to initialize a full cache instance.
type Config struct {
	Name         string
	Version      string
	CacheEnabled bool
	CacheConfig  cache.Config
}

// Validate checks if the configuration is valid.
//
// Returns an error if Name or Version is empty.
func (c Config) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("server name cannot be empty")
	}
	if c.Version == "" {
		return fmt.Errorf("server version cannot be empty")
	}
	return nil
}

// New creates a new MCP server with common infrastructure.
//
// It initializes the HTTP client with retry logic, creates a cache instance
// (if enabled), and sets up the underlying MCP server. The configuration is
// validated before creating the server.
//
// Returns an error if the configuration is invalid or if cache creation fails.
func New(cfg Config, logger *zap.Logger) (*Server, error) {
	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	// Create shared HTTP client
	httpClient := httpx.New(logger)

	// Create cache
	var cacheInstance *cache.Cache
	var err error
	if cfg.CacheEnabled {
		cacheInstance, err = cache.New(cfg.CacheConfig, logger)
		if err != nil {
			return nil, fmt.Errorf("create cache: %w", err)
		}
	} else {
		// Create a minimal no-op cache
		cacheInstance, _ = cache.New(cache.Config{
			MaxCost:     1024,
			NumCounters: 100,
			BufferItems: 1,
		}, logger)
	}

	// Create MCP server
	impl := &mcp.Implementation{
		Name:    cfg.Name,
		Version: cfg.Version,
	}
	mcpServer := mcp.NewServer(impl, nil)

	// Create server instance
	s := &Server{
		mcp:        mcpServer,
		httpClient: httpClient,
		cache:      cacheInstance,
		logger:     logger,
		config:     cfg,
	}

	logger.Info("base server initialized",
		zap.String("name", cfg.Name),
		zap.String("version", cfg.Version),
		zap.Bool("cache_enabled", cfg.CacheEnabled),
	)

	return s, nil
}

// HTTPClient returns the shared HTTP client.
//
// The client includes retry logic, proper timeouts, and is safe for concurrent use.
func (s *Server) HTTPClient() *httpx.Client {
	return s.httpClient
}

// Cache returns the cache instance.
//
// Even when CacheEnabled is false, a minimal cache instance is returned.
func (s *Server) Cache() *cache.Cache {
	return s.cache
}

// Logger returns the logger instance.
//
// This is the same logger passed to New() during server creation.
func (s *Server) Logger() *zap.Logger {
	return s.logger
}

// MCP returns the underlying MCP server for direct access if needed.
//
// Most users should prefer using the helper methods (AddTool, AddResource, etc.)
// rather than accessing the MCP server directly.
func (s *Server) MCP() *mcp.Server {
	return s.mcp
}

// IncrementToolCount increments the tool counter.
//
// This is called automatically by AddTool, so you typically don't need to call it manually.
func (s *Server) IncrementToolCount() {
	s.toolCount++
}

// IncrementResourceCount increments the resource counter.
//
// This is called automatically by AddResource and AddResourceTemplate,
// so you typically don't need to call it manually.
func (s *Server) IncrementResourceCount() {
	s.resourceCount++
}

// LogRegistrationStats logs the number of registered tools and resources.
//
// This is useful for debugging and verifying that all expected features were registered.
func (s *Server) LogRegistrationStats() {
	s.logger.Info("registered tools and resources",
		zap.Int("tools", s.toolCount),
		zap.Int("resources", s.resourceCount),
	)
}

// Run starts the server with the given transport.
//
// This method blocks until the context is canceled or an error occurs.
// Most users should use RunWithTransport instead of calling this directly.
func (s *Server) Run(ctx context.Context, transport mcp.Transport) error {
	s.logger.Info("starting mcp server")
	return s.mcp.Run(ctx, transport)
}

// AddTool registers a tool with the MCP server and automatically increments the tool counter.
//
// This is a generic function that provides type-safe tool registration. The input and output
// types are inferred from the handler function signature. If the tool's input or output schema
// is nil, it will be automatically generated from the type parameters.
//
// Example:
//
//	type Input struct {
//	    Message string `json:"message"`
//	}
//	type Output struct {
//	    Result string `json:"result"`
//	}
//	hypermcp.AddTool(srv, &mcp.Tool{Name: "echo"}, func(ctx context.Context, req *mcp.CallToolRequest, input Input) (*mcp.CallToolResult, Output, error) {
//	    return nil, Output{Result: input.Message}, nil
//	})
func AddTool[In, Out any](s *Server, tool *mcp.Tool, handler mcp.ToolHandlerFor[In, Out]) {
	mcp.AddTool(s.mcp, tool, handler)
	s.IncrementToolCount()
}

// AddResource registers a resource with the MCP server and automatically increments the resource counter.
//
// Resources provide static or dynamic content that can be read by MCP clients.
//
// Example:
//
//	srv.AddResource(&mcp.Resource{
//	    URI: "myapp://data",
//	    Name: "Application Data",
//	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
//	    return &mcp.ReadResourceResult{...}, nil
//	})
func (s *Server) AddResource(resource *mcp.Resource, handler mcp.ResourceHandler) {
	s.mcp.AddResource(resource, handler)
	s.IncrementResourceCount()
}

// AddResourceTemplate registers a resource template with the MCP server and automatically
// increments the resource counter.
//
// Resource templates allow parameterized URIs using URI template syntax (RFC 6570).
//
// Example:
//
//	srv.AddResourceTemplate(&mcp.ResourceTemplate{
//	    URITemplate: "myapp://users/{userId}",
//	    Name: "User Data",
//	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
//	    userId := req.Params.URI // Extract from actual request
//	    return &mcp.ReadResourceResult{...}, nil
//	})
func (s *Server) AddResourceTemplate(template *mcp.ResourceTemplate, handler mcp.ResourceHandler) {
	s.mcp.AddResourceTemplate(template, handler)
	s.IncrementResourceCount()
}

// Shutdown performs cleanup and gracefully shuts down the server.
//
// This method closes the cache and logs final statistics. It respects the provided
// context's deadline or cancellation. If cleanup takes longer than the context allows,
// it returns the context error.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	if err := srv.Shutdown(ctx); err != nil {
//	    log.Printf("shutdown error: %v", err)
//	}
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down server")

	// Create a channel to signal completion
	done := make(chan struct{})

	go func() {
		defer close(done)

		// Log final statistics
		s.LogRegistrationStats()

		// Close cache
		if s.cache != nil {
			s.logger.Debug("closing cache")
			s.cache.Close()
		}

		s.logger.Info("server shutdown complete")
	}()

	// Wait for cleanup or context cancellation
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		s.logger.Warn("shutdown canceled or timed out", zap.Error(ctx.Err()))
		return ctx.Err()
	}
}
