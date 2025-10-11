// Package hypermcp provides reusable MCP server infrastructure
package hypermcp

// ServerInfo holds version and build information for the MCP server.
//
// This struct can be populated at build time using ldflags:
//
//	go build -ldflags="-X main.version=1.0.0 -X main.commit=abc123 -X main.buildDate=2025-01-15"
type ServerInfo struct {
	Name      string
	Version   string
	Commit    string
	BuildDate string
}

// String returns a formatted version string with commit and build date information.
//
// Format: "version (commit: hash, built: date)"
func (si ServerInfo) String() string {
	return si.Version + " (commit: " + si.Commit + ", built: " + si.BuildDate + ")"
}
