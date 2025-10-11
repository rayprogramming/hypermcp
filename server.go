// Package hypermcp provides reusable MCP server infrastructure
package hypermcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rayprogramming/hypermcp/cache"
	"github.com/rayprogramming/hypermcp/httpx"
	"go.uber.org/zap"
)

// Server wraps the MCP server with common infrastructure
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

// Config holds server configuration
type Config struct {
	Name         string
	Version      string
	CacheEnabled bool
	CacheConfig  cache.Config
}

// Validate checks if the configuration is valid
func (c Config) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("server name cannot be empty")
	}
	if c.Version == "" {
		return fmt.Errorf("server version cannot be empty")
	}
	return nil
}

// New creates a new MCP server with common infrastructure
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

// HTTPClient returns the shared HTTP client
func (s *Server) HTTPClient() *httpx.Client {
	return s.httpClient
}

// Cache returns the cache instance
func (s *Server) Cache() *cache.Cache {
	return s.cache
}

// Logger returns the logger instance
func (s *Server) Logger() *zap.Logger {
	return s.logger
}

// MCP returns the underlying MCP server for direct access if needed
func (s *Server) MCP() *mcp.Server {
	return s.mcp
}

// IncrementToolCount increments the tool counter (call after AddTool)
func (s *Server) IncrementToolCount() {
	s.toolCount++
}

// IncrementResourceCount increments the resource counter (call after AddResource/AddResourceTemplate)
func (s *Server) IncrementResourceCount() {
	s.resourceCount++
}

// LogRegistrationStats logs the number of registered tools and resources
func (s *Server) LogRegistrationStats() {
	s.logger.Info("registered tools and resources",
		zap.Int("tools", s.toolCount),
		zap.Int("resources", s.resourceCount),
	)
}

// Run starts the server with the given transport
func (s *Server) Run(ctx context.Context, transport mcp.Transport) error {
	s.logger.Info("starting mcp server")
	return s.mcp.Run(ctx, transport)
}

// AddTool registers a tool with the MCP server and increments the tool counter
func AddTool[In, Out any](s *Server, tool *mcp.Tool, handler mcp.ToolHandlerFor[In, Out]) {
	mcp.AddTool(s.mcp, tool, handler)
	s.IncrementToolCount()
}

// AddResource registers a resource with the MCP server and increments the resource counter
func (s *Server) AddResource(resource *mcp.Resource, handler mcp.ResourceHandler) {
	s.mcp.AddResource(resource, handler)
	s.IncrementResourceCount()
}

// AddResourceTemplate registers a resource template with the MCP server and increments the resource counter
func (s *Server) AddResourceTemplate(template *mcp.ResourceTemplate, handler mcp.ResourceHandler) {
	s.mcp.AddResourceTemplate(template, handler)
	s.IncrementResourceCount()
}

// Shutdown performs cleanup
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down server")
	// Add any cleanup logic here
	return nil
}
