package hypermcp

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rayprogramming/hypermcp/cache"
	"go.uber.org/zap/zaptest"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		errMsg  string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Name:         "test-server",
				Version:      "1.0.0",
				CacheEnabled: true,
			},
			wantErr: false,
		},
		{
			name: "empty name",
			config: Config{
				Name:    "",
				Version: "1.0.0",
			},
			wantErr: true,
			errMsg:  "server name cannot be empty",
		},
		{
			name: "empty version",
			config: Config{
				Name:    "test-server",
				Version: "",
			},
			wantErr: true,
			errMsg:  "server version cannot be empty",
		},
		{
			name: "both empty",
			config: Config{
				Name:    "",
				Version: "",
			},
			wantErr: true,
			errMsg:  "server name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				} else if err.Error() != tt.errMsg {
					t.Errorf("expected error %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestNew_ValidationError(t *testing.T) {
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name    string
		wantErr string
		config  Config
	}{
		{
			name: "empty name",
			config: Config{
				Name:    "",
				Version: "1.0.0",
			},
			wantErr: "invalid config: server name cannot be empty",
		},
		{
			name: "empty version",
			config: Config{
				Name:    "test-server",
				Version: "",
			},
			wantErr: "invalid config: server version cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.config, logger)
			if err == nil {
				t.Fatal("expected error but got nil")
			}
			if err.Error() != tt.wantErr {
				t.Errorf("expected error %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestNew(t *testing.T) {
	logger := zaptest.NewLogger(t)

	cfg := Config{
		Name:         "test-server",
		Version:      "1.0.0",
		CacheEnabled: true,
		CacheConfig:  cache.DefaultConfig(),
	}

	srv, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	if srv == nil {
		t.Fatal("server is nil")
	}

	if srv.HTTPClient() == nil {
		t.Error("HTTPClient is nil")
	}

	if srv.Cache() == nil {
		t.Error("Cache is nil")
	}

	if srv.Logger() == nil {
		t.Error("Logger is nil")
	}

	if srv.MCP() == nil {
		t.Error("MCP server is nil")
	}
}

func TestNew_CacheDisabled(t *testing.T) {
	logger := zaptest.NewLogger(t)

	cfg := Config{
		Name:         "test-server",
		Version:      "1.0.0",
		CacheEnabled: false,
	}

	srv, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Cache should still exist but be minimal
	if srv.Cache() == nil {
		t.Error("Cache is nil even when disabled")
	}
}

func TestServer_AddTool(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := Config{
		Name:         "test-server",
		Version:      "1.0.0",
		CacheEnabled: false,
	}

	srv, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	initialCount := srv.toolCount

	// Add a test tool
	type TestInput struct {
		Message string `json:"message"`
	}
	type TestOutput struct {
		Result string `json:"result"`
	}

	AddTool(srv, &mcp.Tool{
		Name:        "test_tool",
		Description: "A test tool",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input TestInput) (*mcp.CallToolResult, TestOutput, error) {
		return nil, TestOutput{Result: "ok"}, nil
	})

	if srv.toolCount != initialCount+1 {
		t.Errorf("expected tool count to be %d, got %d", initialCount+1, srv.toolCount)
	}
}

func TestServer_AddResource(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := Config{
		Name:         "test-server",
		Version:      "1.0.0",
		CacheEnabled: false,
	}

	srv, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	initialCount := srv.resourceCount

	srv.AddResource(&mcp.Resource{
		URI:         "test://resource",
		Name:        "Test Resource",
		Description: "A test resource",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      "test://resource",
					MIMEType: "application/json",
					Text:     "test content",
				},
			},
		}, nil
	})

	if srv.resourceCount != initialCount+1 {
		t.Errorf("expected resource count to be %d, got %d", initialCount+1, srv.resourceCount)
	}
}

func TestServer_AddResourceTemplate(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := Config{
		Name:         "test-server",
		Version:      "1.0.0",
		CacheEnabled: false,
	}

	srv, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	initialCount := srv.resourceCount

	srv.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "test://resource/{id}",
		Name:        "Test Resource Template",
		Description: "A test resource template",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      req.Params.URI,
					MIMEType: "application/json",
					Text:     "test content",
				},
			},
		}, nil
	})

	if srv.resourceCount != initialCount+1 {
		t.Errorf("expected resource count to be %d, got %d", initialCount+1, srv.resourceCount)
	}
}

func TestServer_LogRegistrationStats(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := Config{
		Name:         "test-server",
		Version:      "1.0.0",
		CacheEnabled: false,
	}

	srv, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Should not panic
	srv.LogRegistrationStats()
}

func TestServer_Shutdown(t *testing.T) {
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name         string
		cacheEnabled bool
		withTimeout  bool
		wantErr      bool
	}{
		{
			name:         "successful shutdown without cache",
			cacheEnabled: false,
			withTimeout:  false,
			wantErr:      false,
		},
		{
			name:         "successful shutdown with cache",
			cacheEnabled: true,
			withTimeout:  false,
			wantErr:      false,
		},
		{
			name:         "shutdown with timeout",
			cacheEnabled: true,
			withTimeout:  true,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				Name:         "test-server",
				Version:      "1.0.0",
				CacheEnabled: tt.cacheEnabled,
			}

			if tt.cacheEnabled {
				cfg.CacheConfig = cache.Config{
					MaxCost:     1024 * 1024, // 1MB
					NumCounters: 1000,
					BufferItems: 64,
				}
			}

		srv, err := New(cfg, logger)
		if err != nil {
			t.Fatalf("failed to create server: %v", err)
		}

		var ctx context.Context
		var cancel context.CancelFunc

		if tt.withTimeout {
			// Create a context that's already canceled to simulate timeout
			ctx, cancel = context.WithCancel(context.Background())
			cancel()
		} else {
			ctx = context.Background()
		}

		err = srv.Shutdown(ctx)
		if (err != nil) != tt.wantErr {
			t.Errorf("Shutdown() error = %v, wantErr %v", err, tt.wantErr)
		}
		})
	}
}

func TestServer_IncrementCounters(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := Config{
		Name:         "test-server",
		Version:      "1.0.0",
		CacheEnabled: false,
	}

	srv, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	initialToolCount := srv.toolCount
	initialResourceCount := srv.resourceCount

	srv.IncrementToolCount()
	if srv.toolCount != initialToolCount+1 {
		t.Errorf("expected tool count %d, got %d", initialToolCount+1, srv.toolCount)
	}

	srv.IncrementResourceCount()
	if srv.resourceCount != initialResourceCount+1 {
		t.Errorf("expected resource count %d, got %d", initialResourceCount+1, srv.resourceCount)
	}
}
