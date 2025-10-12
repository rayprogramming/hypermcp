package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rayprogramming/hypermcp"
	"go.uber.org/zap"
)

func main() {
	// Create logger
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Create server
	cfg := hypermcp.Config{
		Name:         "fileserver",
		Version:      "1.0.0",
		CacheEnabled: false,
	}

	srv, err := hypermcp.New(cfg, logger)
	if err != nil {
		logger.Fatal("failed to create server", zap.Error(err))
	}

	// Get base directory (examples/fileserver by default)
	baseDir, _ := os.Getwd()

	// Register a resource for the examples directory
	srv.AddResource(&mcp.Resource{
		URI:         "file:///examples",
		Name:        "Examples Directory",
		Description: "Contents of the examples directory",
		MIMEType:    "text/plain",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		entries, err := os.ReadDir(filepath.Join(baseDir, ".."))
		if err != nil {
			return nil, fmt.Errorf("failed to read directory: %w", err)
		}

		var content string
		content += "Examples Directory:\n\n"
		for _, entry := range entries {
			if entry.IsDir() {
				content += fmt.Sprintf("📁 %s/\n", entry.Name())
			} else {
				info, _ := entry.Info()
				content += fmt.Sprintf("📄 %s (%d bytes)\n", entry.Name(), info.Size())
			}
		}

		srv.Metrics().IncrementResourceReads()

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      req.Params.URI,
					MIMEType: "text/plain",
					Text:     content,
				},
			},
		}, nil
	})

	// Register a resource template for reading specific files
	srv.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "file:///examples/{filename}",
		Name:        "Example File",
		Description: "Read a specific file from the examples directory",
		MIMEType:    "text/plain",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		// Extract filename from URI
		// In a real implementation, you'd properly parse the URI template
		// For this demo, we'll use a simplified approach
		uri := req.Params.URI
		filename := filepath.Base(uri)

		// Security: prevent directory traversal
		if filepath.IsAbs(filename) || filepath.Dir(filename) != "." {
			return nil, fmt.Errorf("invalid filename")
		}

		fullPath := filepath.Join(baseDir, "..", filename, "README.md")
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}

		srv.Metrics().IncrementResourceReads()

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      req.Params.URI,
					MIMEType: "text/plain",
					Text:     string(content),
				},
			},
		}, nil
	})

	// Log registration stats
	srv.LogRegistrationStats()

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("shutting down...")
		
		// Log metrics
		metrics := srv.GetMetrics()
		logger.Info("final metrics",
			zap.Duration("uptime", metrics.Uptime),
			zap.Int64("resource_reads", metrics.ResourceReads),
		)
		
		cancel()
	}()

	// Run server
	if err := hypermcp.RunWithTransport(ctx, srv, hypermcp.TransportStdio, logger); err != nil {
		logger.Error("server error", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("server stopped")
}
