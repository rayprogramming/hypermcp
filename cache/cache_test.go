package cache

import (
	"testing"
	"time"

	"go.uber.org/zap/zaptest"
)

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
