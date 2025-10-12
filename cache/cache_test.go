package cache

import (
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestNew_InvalidConfig(t *testing.T) {
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name          string
		cfg           Config
		wantError     bool
		expectedError error
	}{
		{
			name: "valid config",
			cfg: Config{
				MaxCost:     1024,
				NumCounters: 100,
				BufferItems: 10,
			},
			wantError: false,
		},
		{
			name: "zero MaxCost",
			cfg: Config{
				MaxCost:     0,
				NumCounters: 100,
				BufferItems: 10,
			},
			wantError:     true,
			expectedError: ErrInvalidMaxCost,
		},
		{
			name: "negative MaxCost",
			cfg: Config{
				MaxCost:     -1,
				NumCounters: 100,
				BufferItems: 10,
			},
			wantError:     true,
			expectedError: ErrInvalidMaxCost,
		},
		{
			name: "zero NumCounters",
			cfg: Config{
				MaxCost:     1024,
				NumCounters: 0,
				BufferItems: 10,
			},
			wantError:     true,
			expectedError: ErrInvalidNumCounters,
		},
		{
			name: "zero BufferItems",
			cfg: Config{
				MaxCost:     1024,
				NumCounters: 100,
				BufferItems: 0,
			},
			wantError:     true,
			expectedError: ErrInvalidBufferItems,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runConfigTest(t, logger, tt.cfg, tt.wantError, tt.expectedError)
		})
	}
}

func runConfigTest(t *testing.T, logger *zap.Logger, cfg Config, wantError bool, expectedError error) {
	t.Helper()
	c, err := New(cfg, logger)

	if !wantError {
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if c != nil {
			c.Close()
		}
		return
	}

	// Error expected
	if err == nil {
		t.Error("expected error but got none")
		return
	}

	// Check that the error is the expected sentinel error
	if expectedError != nil && !errors.Is(err, expectedError) {
		t.Errorf("expected error %v, got %v", expectedError, err)
	}

	// Check that it's a ValidationError
	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Errorf("expected ValidationError type, got %T", err)
	}
}

func TestCache_GetSet(t *testing.T) {
	logger := zaptest.NewLogger(t)
	c, err := New(DefaultConfig(), logger)
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}
	defer c.Close()

	// Set a value
	key := "test-key"
	value := "test-value"
	c.Set(key, value, 5*time.Second)

	// Wait for ristretto to process
	time.Sleep(10 * time.Millisecond)

	// Get the value
	result, found := c.Get(key)
	if !found {
		t.Fatal("expected value to be found")
	}

	if result != value {
		t.Errorf("expected %v, got %v", value, result)
	}
}

func TestCache_Expiration(t *testing.T) {
	logger := zaptest.NewLogger(t)
	c, err := New(DefaultConfig(), logger)
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}
	defer c.Close()

	key := "expire-key"
	value := "expire-value"
	c.Set(key, value, 50*time.Millisecond)

	time.Sleep(10 * time.Millisecond)

	// Should still be there
	_, found := c.Get(key)
	if !found {
		t.Error("expected value to be found before expiration")
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be gone
	_, found = c.Get(key)
	if found {
		t.Error("expected value to be expired")
	}
}

func TestCache_Delete(t *testing.T) {
	logger := zaptest.NewLogger(t)
	c, err := New(DefaultConfig(), logger)
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}
	defer c.Close()

	key := "delete-key"
	value := "delete-value"
	c.Set(key, value, 5*time.Second)

	time.Sleep(10 * time.Millisecond)

	// Delete
	c.Delete(key)

	// Should be gone
	_, found := c.Get(key)
	if found {
		t.Error("expected value to be deleted")
	}
}

func BenchmarkCache_Get(b *testing.B) {
	logger := zaptest.NewLogger(b)
	c, err := New(DefaultConfig(), logger)
	if err != nil {
		b.Fatalf("failed to create cache: %v", err)
	}
	defer c.Close()

	// Pre-populate cache
	c.Set("bench-key", "bench-value", 60*time.Second)
	time.Sleep(10 * time.Millisecond)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		c.Get("bench-key")
	}
}

func BenchmarkCache_Set(b *testing.B) {
	logger := zaptest.NewLogger(b)
	c, err := New(DefaultConfig(), logger)
	if err != nil {
		b.Fatalf("failed to create cache: %v", err)
	}
	defer c.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		c.Set("bench-key", "bench-value", 60*time.Second)
	}
}

func TestCache_Clear(t *testing.T) {
	logger := zaptest.NewLogger(t)
	c, err := New(DefaultConfig(), logger)
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}
	defer c.Close()

	// Add multiple values
	c.Set("key1", "value1", 5*time.Second)
	c.Set("key2", "value2", 5*time.Second)
	c.Set("key3", "value3", 5*time.Second)

	time.Sleep(10 * time.Millisecond)

	// Verify they exist
	if _, found := c.Get("key1"); !found {
		t.Error("expected key1 to be found before clear")
	}

	// Clear the cache
	c.Clear()

	// Verify all keys are gone
	if _, found := c.Get("key1"); found {
		t.Error("expected key1 to be cleared")
	}
	if _, found := c.Get("key2"); found {
		t.Error("expected key2 to be cleared")
	}
	if _, found := c.Get("key3"); found {
		t.Error("expected key3 to be cleared")
	}
}

func TestCache_Metrics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	c, err := New(DefaultConfig(), logger)
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}
	defer c.Close()

	// Get metrics
	metrics := c.Metrics()
	if metrics == nil {
		t.Fatal("expected metrics to be non-nil")
	}

	// Set a value and access it
	c.Set("metrics-key", "metrics-value", 5*time.Second)
	time.Sleep(10 * time.Millisecond)

	// Access the key (should be a hit)
	c.Get("metrics-key")

	// Verify metrics are being tracked
	if metrics.Hits() == 0 {
		t.Error("expected at least one cache hit to be recorded")
	}

	// Access a non-existent key (should be a miss)
	c.Get("non-existent-key")

	// Verify misses are tracked
	if metrics.Misses() == 0 {
		t.Error("expected at least one cache miss to be recorded")
	}
}
