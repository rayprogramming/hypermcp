package hypermcp

import (
	"testing"
	"time"

	"go.uber.org/zap/zaptest"
)

func TestMetrics_IncrementCounters(t *testing.T) {
	m := newMetrics()

	// Test initial state
	snapshot := m.Snapshot()
	if snapshot.ToolInvocations != 0 {
		t.Errorf("expected 0 tool invocations, got %d", snapshot.ToolInvocations)
	}

	// Increment various counters
	m.IncrementToolInvocations()
	m.IncrementToolInvocations()
	m.IncrementResourceReads()
	m.IncrementCacheHits()
	m.IncrementCacheMisses()
	m.IncrementErrors()

	// Verify counts
	snapshot = m.Snapshot()
	if snapshot.ToolInvocations != 2 {
		t.Errorf("expected 2 tool invocations, got %d", snapshot.ToolInvocations)
	}
	if snapshot.ResourceReads != 1 {
		t.Errorf("expected 1 resource read, got %d", snapshot.ResourceReads)
	}
	if snapshot.CacheHits != 1 {
		t.Errorf("expected 1 cache hit, got %d", snapshot.CacheHits)
	}
	if snapshot.CacheMisses != 1 {
		t.Errorf("expected 1 cache miss, got %d", snapshot.CacheMisses)
	}
	if snapshot.Errors != 1 {
		t.Errorf("expected 1 error, got %d", snapshot.Errors)
	}
}

func TestMetrics_CacheHitRate(t *testing.T) {
	tests := []struct {
		name     string
		hits     int
		misses   int
		expected float64
	}{
		{
			name:     "no access",
			hits:     0,
			misses:   0,
			expected: 0,
		},
		{
			name:     "all hits",
			hits:     10,
			misses:   0,
			expected: 1.0,
		},
		{
			name:     "all misses",
			hits:     0,
			misses:   10,
			expected: 0.0,
		},
		{
			name:     "50% hit rate",
			hits:     5,
			misses:   5,
			expected: 0.5,
		},
		{
			name:     "75% hit rate",
			hits:     75,
			misses:   25,
			expected: 0.75,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMetrics()

			for i := 0; i < tt.hits; i++ {
				m.IncrementCacheHits()
			}
			for i := 0; i < tt.misses; i++ {
				m.IncrementCacheMisses()
			}

			snapshot := m.Snapshot()
			if snapshot.CacheHitRate != tt.expected {
				t.Errorf("expected hit rate %.2f, got %.2f", tt.expected, snapshot.CacheHitRate)
			}
		})
	}
}

func TestMetrics_Uptime(t *testing.T) {
	m := newMetrics()

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	snapshot := m.Snapshot()
	if snapshot.Uptime < 10*time.Millisecond {
		t.Errorf("expected uptime >= 10ms, got %v", snapshot.Uptime)
	}
	if snapshot.Uptime > 1*time.Second {
		t.Errorf("expected uptime < 1s, got %v", snapshot.Uptime)
	}
}

func TestServer_GetMetrics(t *testing.T) {
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

	// Get initial metrics
	metrics := srv.GetMetrics()
	if metrics.ToolInvocations != 0 {
		t.Errorf("expected 0 tool invocations, got %d", metrics.ToolInvocations)
	}

	// Increment some metrics
	srv.Metrics().IncrementToolInvocations()
	srv.Metrics().IncrementResourceReads()

	// Get updated metrics
	metrics = srv.GetMetrics()
	if metrics.ToolInvocations != 1 {
		t.Errorf("expected 1 tool invocation, got %d", metrics.ToolInvocations)
	}
	if metrics.ResourceReads != 1 {
		t.Errorf("expected 1 resource read, got %d", metrics.ResourceReads)
	}
}

func TestMetrics_Concurrent(t *testing.T) {
	m := newMetrics()

	// Run concurrent increments
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				m.IncrementToolInvocations()
				m.IncrementResourceReads()
				m.IncrementCacheHits()
				m.IncrementCacheMisses()
				m.IncrementErrors()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify counts
	snapshot := m.Snapshot()
	expected := int64(1000) // 10 goroutines * 100 iterations

	if snapshot.ToolInvocations != expected {
		t.Errorf("expected %d tool invocations, got %d", expected, snapshot.ToolInvocations)
	}
	if snapshot.ResourceReads != expected {
		t.Errorf("expected %d resource reads, got %d", expected, snapshot.ResourceReads)
	}
	if snapshot.CacheHits != expected {
		t.Errorf("expected %d cache hits, got %d", expected, snapshot.CacheHits)
	}
	if snapshot.CacheMisses != expected {
		t.Errorf("expected %d cache misses, got %d", expected, snapshot.CacheMisses)
	}
	if snapshot.Errors != expected {
		t.Errorf("expected %d errors, got %d", expected, snapshot.Errors)
	}
}
