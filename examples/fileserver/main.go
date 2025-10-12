package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
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
	// Flush logs; ignore non-fatal sync errors
	defer func() {
		if syncErr := logger.Sync(); syncErr != nil {
			fmt.Fprintf(os.Stderr, "logger sync error: %v\n", syncErr)
		}
	}()

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
		return readExamplesDir(srv, baseDir, req)
	})

	// Register a resource template for reading specific files
	srv.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "file:///examples/{filename}",
		Name:        "Example File",
		Description: "Read a specific file from the examples directory",
		MIMEType:    "text/plain",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return readExampleFile(srv, baseDir, req)
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
	var runErr error
	if runErr = hypermcp.RunWithTransport(ctx, srv, hypermcp.TransportStdio, logger); runErr != nil {
		logger.Error("server error", zap.Error(runErr))
		os.Exit(1)
	}

	logger.Info("server stopped")
}

// readExamplesDir lists the examples directory contents
func readExamplesDir(srv *hypermcp.Server, baseDir string, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	entries, readDirErr := os.ReadDir(filepath.Join(baseDir, ".."))
	if readDirErr != nil {
		return nil, fmt.Errorf("failed to read directory: %w", readDirErr)
	}

	var content strings.Builder
	content.WriteString("Examples Directory:\n\n")
	for _, entry := range entries {
		if entry.IsDir() {
			content.WriteString(fmt.Sprintf("📁 %s/\n", entry.Name()))
		} else {
			info, _ := entry.Info()
			content.WriteString(fmt.Sprintf("📄 %s (%d bytes)\n", entry.Name(), info.Size()))
		}
	}

	srv.Metrics().IncrementResourceReads()

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      req.Params.URI,
				MIMEType: "text/plain",
				Text:     content.String(),
			},
		},
	}, nil
}

// readExampleFile reads a specific example file safely within allowed base directory
func readExampleFile(srv *hypermcp.Server, baseDir string, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	// Extract filename from URI (simplified)
	uri := req.Params.URI
	filename := filepath.Base(uri)

	// Security: prevent directory traversal
	if filepath.IsAbs(filename) || filepath.Dir(filename) != "." {
		return nil, fmt.Errorf("invalid filename")
	}

	// Build path within allowed base directory and clean it to prevent traversal
	fullPath := filepath.Clean(filepath.Join(baseDir, "..", filename, "README.md"))
	// Ensure the resolved path is still within the expected directory
	allowedDir := filepath.Clean(filepath.Join(baseDir, ".."))
	rel, relErr := filepath.Rel(allowedDir, fullPath)
	if relErr != nil || rel == ".." || rel == "." || strings.HasPrefix(rel, "..") {
		return nil, fmt.Errorf("invalid path")
	}

	content, readFileErr := os.ReadFile(fullPath)
	if readFileErr != nil {
		return nil, fmt.Errorf("failed to read file: %w", readFileErr)
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
}
