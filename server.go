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
	metrics    *Metrics
	config     Config

	// Stats for logging
	toolCount     int
	resourceCount int
}

// Config holds server configuration.
//
// Name and Version are required fields and will be validated.
// CacheEnabled determines whether to initialize a full cache instance.
// HTTPConfig allows customization of HTTP client behavior (optional, uses defaults if not set).
type Config struct {
	HTTPConfig   *httpx.Config // Optional: uses defaults if nil
	CacheConfig  cache.Config
	Name         string
	Version      string
	CacheEnabled bool
}

// Validate checks if the configuration is valid.
//
// Returns an error if Name or Version is empty.
func (c Config) Validate() error {
	if c.Name == "" {
		return NewConfigError("Name", fmt.Errorf("cannot be empty"))
	}
	if c.Version == "" {
		return NewConfigError("Version", fmt.Errorf("cannot be empty"))
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
		return nil, fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}

	// Create shared HTTP client with optional custom config
	var httpClient *httpx.Client
	var err error
	if cfg.HTTPConfig != nil {
		httpClient, err = httpx.NewWithConfig(*cfg.HTTPConfig, logger)
		if err != nil {
			return nil, fmt.Errorf("create http client: %w", err)
		}
	} else {
		httpClient, err = httpx.New(logger)
		if err != nil {
			return nil, fmt.Errorf("create http client: %w", err)
		}
	}

	// Create cache
	var cacheInstance *cache.Cache
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
		metrics:    newMetrics(),
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
// Also includes cache configuration information if caching is enabled.
func (s *Server) LogRegistrationStats() {
	fields := []zap.Field{
		zap.Int("tools", s.toolCount),
		zap.Int("resources", s.resourceCount),
	}

	// Add cache info if enabled
	if s.config.CacheEnabled && s.cache != nil {
		metrics := s.cache.Metrics()
		fields = append(fields,
			zap.Bool("cache_enabled", true),
			zap.Uint64("cache_hits", metrics.Hits()),
			zap.Uint64("cache_misses", metrics.Misses()),
			zap.Float64("cache_ratio", metrics.Ratio()),
		)
	} else {
		fields = append(fields, zap.Bool("cache_enabled", false))
	}

	s.logger.Info("registered tools and resources", fields...)
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
// This method performs the following cleanup operations in order:
// 1. Logs final registration statistics (tools and resources)
// 2. Closes the cache instance (stops background goroutines)
// 3. Checks for context cancellation or timeout
//
// It's safe to call Shutdown multiple times, though subsequent calls
// will have no effect (except checking context status).
//
// Example with timeout:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	if err := srv.Shutdown(ctx); err != nil {
//	    log.Printf("shutdown error: %v", err)
//	}
//
// Returns an error if the context was canceled or timed out during cleanup.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down server")

	// Log final statistics
	s.LogRegistrationStats()

	// Close cache
	if s.cache != nil {
		s.logger.Debug("closing cache")
		s.cache.Close()
	}

	s.logger.Info("server shutdown complete")

	// Check if context was canceled during cleanup
	if ctx.Err() != nil {
		s.logger.Warn("shutdown canceled or timed out", zap.Error(ctx.Err()))
		return ctx.Err()
	}

	return nil
}
