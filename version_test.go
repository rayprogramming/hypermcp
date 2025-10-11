package hypermcp

import "testing"

func TestServerInfo_String(t *testing.T) {
	tests := []struct {
		name     string
		info     ServerInfo
		expected string
	}{
		{
			name: "full info",
			info: ServerInfo{
				Name:      "test-server",
				Version:   "1.0.0",
				Commit:    "abc123",
				BuildDate: "2025-01-15",
			},
			expected: "1.0.0 (commit: abc123, built: 2025-01-15)",
		},
		{
			name: "empty commit",
			info: ServerInfo{
				Name:      "test-server",
				Version:   "2.0.0",
				Commit:    "",
				BuildDate: "2025-01-15",
			},
			expected: "2.0.0 (commit: , built: 2025-01-15)",
		},
		{
			name: "empty build date",
			info: ServerInfo{
				Name:      "test-server",
				Version:   "1.5.0",
				Commit:    "def456",
				BuildDate: "",
			},
			expected: "1.5.0 (commit: def456, built: )",
		},
		{
			name: "all empty except version",
			info: ServerInfo{
				Version: "3.0.0",
			},
			expected: "3.0.0 (commit: , built: )",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.info.String()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestServerInfo_Fields(t *testing.T) {
	info := ServerInfo{
		Name:      "my-server",
		Version:   "1.2.3",
		Commit:    "abc123def",
		BuildDate: "2025-10-11",
	}

	if info.Name != "my-server" {
		t.Errorf("expected Name to be 'my-server', got %q", info.Name)
	}

	if info.Version != "1.2.3" {
		t.Errorf("expected Version to be '1.2.3', got %q", info.Version)
	}

	if info.Commit != "abc123def" {
		t.Errorf("expected Commit to be 'abc123def', got %q", info.Commit)
	}

	if info.BuildDate != "2025-10-11" {
		t.Errorf("expected BuildDate to be '2025-10-11', got %q", info.BuildDate)
	}
}
