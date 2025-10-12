package hypermcp

import (
	"sync/atomic"
	"time"
)

// Metrics tracks server performance and usage statistics.
//
// All counters are thread-safe using atomic operations and can be safely
// incremented from multiple goroutines.
type Metrics struct {
	// Server lifecycle
	startTime time.Time

	// Tool and resource usage
	toolInvocations atomic.Int64
	resourceReads   atomic.Int64

	// Cache statistics
	cacheHits   atomic.Int64
	cacheMisses atomic.Int64

	// Error tracking
	errors atomic.Int64
}

// MetricsSnapshot provides a point-in-time view of server metrics.
//
// This struct is returned by Server.GetMetrics() and contains copied values
// that won't change, making it safe to use without synchronization.
type MetricsSnapshot struct {
	// Server uptime
	Uptime time.Duration

	// Tool and resource usage
	ToolInvocations int64
	ResourceReads   int64

	// Cache statistics
	CacheHits    int64
	CacheMisses  int64
	CacheHitRate float64 // Calculated as hits / (hits + misses)

	// Error tracking
	Errors int64
}

// newMetrics creates a new Metrics instance with the current time as start time.
func newMetrics() *Metrics {
	return &Metrics{
		startTime: time.Now(),
	}
}

// IncrementToolInvocations increments the tool invocation counter.
func (m *Metrics) IncrementToolInvocations() {
	m.toolInvocations.Add(1)
}

// IncrementResourceReads increments the resource read counter.
func (m *Metrics) IncrementResourceReads() {
	m.resourceReads.Add(1)
}

// IncrementCacheHits increments the cache hit counter.
func (m *Metrics) IncrementCacheHits() {
	m.cacheHits.Add(1)
}

// IncrementCacheMisses increments the cache miss counter.
func (m *Metrics) IncrementCacheMisses() {
	m.cacheMisses.Add(1)
}

// IncrementErrors increments the error counter.
func (m *Metrics) IncrementErrors() {
	m.errors.Add(1)
}

// Snapshot creates a point-in-time snapshot of current metrics.
func (m *Metrics) Snapshot() MetricsSnapshot {
	hits := m.cacheHits.Load()
	misses := m.cacheMisses.Load()

	var hitRate float64
	totalCacheAccess := hits + misses
	if totalCacheAccess > 0 {
		hitRate = float64(hits) / float64(totalCacheAccess)
	}

	return MetricsSnapshot{
		Uptime:          time.Since(m.startTime),
		ToolInvocations: m.toolInvocations.Load(),
		ResourceReads:   m.resourceReads.Load(),
		CacheHits:       hits,
		CacheMisses:     misses,
		CacheHitRate:    hitRate,
		Errors:          m.errors.Load(),
	}
}

// GetMetrics returns a snapshot of current server metrics.
//
// The returned snapshot is a copy of the current metrics and can be safely
// used without worrying about concurrent modifications.
//
// Example:
//
//	metrics := srv.GetMetrics()
//	fmt.Printf("Uptime: %v\n", metrics.Uptime)
//	fmt.Printf("Tool invocations: %d\n", metrics.ToolInvocations)
//	fmt.Printf("Cache hit rate: %.2f%%\n", metrics.CacheHitRate*100)
func (s *Server) GetMetrics() MetricsSnapshot {
	return s.metrics.Snapshot()
}

// Metrics returns the raw Metrics instance for direct access.
//
// This is useful for custom metric tracking or integration with monitoring systems.
// Most users should use GetMetrics() instead, which returns a safe snapshot.
func (s *Server) Metrics() *Metrics {
	return s.metrics
}
