package cache

import (
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
	"go.uber.org/zap"
)

// Cache provides a high-performance in-memory cache
type Cache struct {
	mu     sync.RWMutex
	ttls   map[string]time.Time
	store  *ristretto.Cache[string, any]
	logger *zap.Logger
}

// Config holds cache configuration
type Config struct {
	// MaxCost is the maximum cost of cache entries (in bytes approximately)
	MaxCost int64
	// NumCounters is the number of keys to track frequency
	NumCounters int64
	// BufferItems is the size of the internal buffer
	BufferItems int64
}

// DefaultConfig returns sensible defaults for the cache
func DefaultConfig() Config {
	return Config{
		MaxCost:     100 * 1024 * 1024, // 100MB
		NumCounters: 1000000,           // 1M counters
		BufferItems: 64,
	}
}

// New creates a new cache instance
func New(cfg Config, logger *zap.Logger) (*Cache, error) {
	store, err := ristretto.NewCache(&ristretto.Config[string, any]{
		MaxCost:     cfg.MaxCost,
		NumCounters: cfg.NumCounters,
		BufferItems: cfg.BufferItems,
		Metrics:     true,
	})
	if err != nil {
		return nil, err
	}

	c := &Cache{
		store:  store,
		logger: logger,
		ttls:   make(map[string]time.Time),
	}

	// Start background TTL cleanup
	go c.cleanupExpired()

	return c, nil
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	expiry, hasExpiry := c.ttls[key]
	c.mu.RUnlock()

	if hasExpiry && time.Now().After(expiry) {
		c.Delete(key)
		return nil, false
	}

	value, found := c.store.Get(key)
	if !found {
		return nil, false
	}

	c.logger.Debug("cache hit", zap.String("key", key))
	return value, true
}

// Set stores a value in the cache with TTL
func (c *Cache) Set(key string, value any, ttl time.Duration) {
	// Calculate cost (rough estimate based on type)
	cost := int64(64) // base overhead

	// Store with cost
	c.store.Set(key, value, cost)

	// Track TTL
	if ttl > 0 {
		c.mu.Lock()
		c.ttls[key] = time.Now().Add(ttl)
		c.mu.Unlock()
	}

	c.logger.Debug("cache set",
		zap.String("key", key),
		zap.Duration("ttl", ttl),
	)
}

// Delete removes a value from the cache
func (c *Cache) Delete(key string) {
	c.store.Del(key)

	c.mu.Lock()
	delete(c.ttls, key)
	c.mu.Unlock()

	c.logger.Debug("cache delete", zap.String("key", key))
}

// Clear removes all entries from the cache
func (c *Cache) Clear() {
	c.store.Clear()

	c.mu.Lock()
	c.ttls = make(map[string]time.Time)
	c.mu.Unlock()

	c.logger.Info("cache cleared")
}

// Metrics returns cache performance metrics
func (c *Cache) Metrics() *ristretto.Metrics {
	return c.store.Metrics
}

// cleanupExpired runs a background goroutine to clean up expired entries
func (c *Cache) cleanupExpired() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		var expired []string

		c.mu.RLock()
		for key, expiry := range c.ttls {
			if now.After(expiry) {
				expired = append(expired, key)
			}
		}
		c.mu.RUnlock()

		for _, key := range expired {
			c.Delete(key)
		}

		if len(expired) > 0 {
			c.logger.Debug("cleaned expired entries", zap.Int("count", len(expired)))
		}
	}
}

// Close shuts down the cache
func (c *Cache) Close() {
	c.store.Close()
}
