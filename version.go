// Package hypermcp provides reusable MCP server infrastructure
package hypermcp

// ServerInfo holds version and build information for the MCP server
type ServerInfo struct {
	Name      string
	Version   string
	Commit    string
	BuildDate string
}

// String returns a formatted version string
func (si ServerInfo) String() string {
	return si.Version + " (commit: " + si.Commit + ", built: " + si.BuildDate + ")"
}
