// Package httpx provides a shared HTTP client with retry logic, timeouts, and helpers
package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
	"go.uber.org/zap"
)

// Sentinel errors for httpx configuration validation.
var (
	// ErrInvalidTimeout indicates a timeout value is not positive.
	ErrInvalidTimeout = errors.New("timeout must be positive")

	// ErrInvalidMaxRetries indicates MaxRetries is negative.
	ErrInvalidMaxRetries = errors.New("MaxRetries cannot be negative")

	// ErrInvalidMaxResponseSize indicates MaxResponseSize is not positive.
	ErrInvalidMaxResponseSize = errors.New("MaxResponseSize must be positive")

	// ErrInvalidMaxIdleConns indicates MaxIdleConns is not positive.
	ErrInvalidMaxIdleConns = errors.New("MaxIdleConns must be positive")

	// ErrInvalidRetryInterval indicates retry interval is not positive.
	ErrInvalidRetryInterval = errors.New("retry interval must be positive")
)

// ConfigError wraps httpx configuration validation errors with context.
type ConfigError struct {
	Err   error
	Field string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("httpx config validation error in field %q: %v", e.Field, e.Err)
}

func (e *ConfigError) Unwrap() error {
	return e.Err
}

// Config holds HTTP client configuration options.
type Config struct {
	// Timeouts
	DialTimeout           time.Duration
	TLSHandshakeTimeout   time.Duration
	ResponseHeaderTimeout time.Duration
	RequestTimeout        time.Duration

	// Retry configuration
	MaxRetries      int
	InitialInterval time.Duration
	MaxInterval     time.Duration

	// Request limits
	MaxResponseSize int64

	// Connection pooling
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	IdleConnTimeout     time.Duration

	// UserAgent to use in HTTP requests (optional, defaults to "hypermcp")
	UserAgent string

	// DisableCompression, if true, prevents the Transport from requesting compression
	// with an "Accept-Encoding: gzip" request header when the Request contains no
	// existing Accept-Encoding value. Defaults to false (compression enabled).
	DisableCompression bool

	// ForceAttemptHTTP2 controls whether HTTP/2 is enabled when a non-zero
	// Dial, DialTLS, or DialContext func or TLSClientConfig is provided.
	// Defaults to true.
	ForceAttemptHTTP2 bool
}

// DefaultConfig returns sensible default configuration for the HTTP client.
func DefaultConfig() Config {
	return Config{
		DialTimeout:           2 * time.Second,
		TLSHandshakeTimeout:   2 * time.Second,
		ResponseHeaderTimeout: 4 * time.Second,
		RequestTimeout:        6 * time.Second,
		MaxRetries:            3,
		InitialInterval:       100 * time.Millisecond,
		MaxInterval:           2 * time.Second,
		MaxResponseSize:       10 * 1024 * 1024, // 10MB
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		UserAgent:             "hypermcp",
		DisableCompression:    false, // Enable gzip compression
		ForceAttemptHTTP2:     true,  // Enable HTTP/2
	}
}

// Validate checks if the configuration is valid.
//
// Returns a ConfigError if any field contains an invalid value.
func (c Config) Validate() error {
	if c.DialTimeout <= 0 {
		return &ConfigError{
			Err:   ErrInvalidTimeout,
			Field: "DialTimeout",
		}
	}
	if c.TLSHandshakeTimeout <= 0 {
		return &ConfigError{
			Err:   ErrInvalidTimeout,
			Field: "TLSHandshakeTimeout",
		}
	}
	if c.ResponseHeaderTimeout <= 0 {
		return &ConfigError{
			Err:   ErrInvalidTimeout,
			Field: "ResponseHeaderTimeout",
		}
	}
	if c.RequestTimeout <= 0 {
		return &ConfigError{
			Err:   ErrInvalidTimeout,
			Field: "RequestTimeout",
		}
	}
	if c.MaxRetries < 0 {
		return &ConfigError{
			Err:   ErrInvalidMaxRetries,
			Field: "MaxRetries",
		}
	}
	if c.InitialInterval <= 0 {
		return &ConfigError{
			Err:   ErrInvalidRetryInterval,
			Field: "InitialInterval",
		}
	}
	if c.MaxInterval <= 0 {
		return &ConfigError{
			Err:   ErrInvalidRetryInterval,
			Field: "MaxInterval",
		}
	}
	if c.MaxResponseSize <= 0 {
		return &ConfigError{
			Err:   ErrInvalidMaxResponseSize,
			Field: "MaxResponseSize",
		}
	}
	if c.MaxIdleConns <= 0 {
		return &ConfigError{
			Err:   ErrInvalidMaxIdleConns,
			Field: "MaxIdleConns",
		}
	}
	if c.MaxIdleConnsPerHost <= 0 {
		return &ConfigError{
			Err:   ErrInvalidMaxIdleConns,
			Field: "MaxIdleConnsPerHost",
		}
	}
	if c.IdleConnTimeout <= 0 {
		return &ConfigError{
			Err:   ErrInvalidTimeout,
			Field: "IdleConnTimeout",
		}
	}
	return nil
}

// Client wraps an HTTP client with retry logic and performance optimizations
type Client struct {
	client *http.Client
	logger *zap.Logger
	config Config
}

// New creates a new HTTP client with default configuration.
func New(logger *zap.Logger) (*Client, error) {
	return NewWithConfig(DefaultConfig(), logger)
}

// NewWithConfig creates a new HTTP client with custom configuration.
//
// Returns an error if the configuration is invalid.
func NewWithConfig(cfg Config, logger *zap.Logger) (*Client, error) {
	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   cfg.DialTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   cfg.TLSHandshakeTimeout,
		ResponseHeaderTimeout: cfg.ResponseHeaderTimeout,
		MaxIdleConns:          cfg.MaxIdleConns,
		MaxIdleConnsPerHost:   cfg.MaxIdleConnsPerHost,
		IdleConnTimeout:       cfg.IdleConnTimeout,
		DisableCompression:    cfg.DisableCompression,
		ForceAttemptHTTP2:     cfg.ForceAttemptHTTP2,
	}

	return &Client{
		client: &http.Client{
			Transport: transport,
			Timeout:   cfg.RequestTimeout,
		},
		logger: logger,
		config: cfg,
	}, nil
}

// DoJSON performs an HTTP request and unmarshals the JSON response.
// It includes retry logic with exponential backoff for transient errors.
//
// The method automatically handles:
// - Exponential backoff with jitter for retryable errors (429, 5xx)
// - Request timeouts to prevent hanging
// - Response size limits to prevent memory exhaustion (10MB default)
// - Context cancellation for early termination
//
// Retryable status codes: 429 (Too Many Requests), 500-504 (Server Errors)
// Non-retryable errors: 4xx (except 429), JSON decode errors, network errors
//
// The request context controls the overall timeout, while individual retry
// attempts have their own timeouts configured via Config.RequestTimeout.
func (c *Client) DoJSON(ctx context.Context, req *http.Request, result interface{}) error {
	reqID := fmt.Sprintf("%p", req)
	startTime := time.Now()

	operation := func() error {
		// Clone request for retry safety
		clonedReq := req.Clone(ctx)

		resp, err := c.client.Do(clonedReq)
		if err != nil {
			c.logger.Debug("http request failed",
				zap.String("url", req.URL.String()),
				zap.Error(err),
			)
			return err
		}
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				c.logger.Warn("failed to close response body", zap.Error(closeErr))
			}
		}()

		// Limit response size to prevent memory exhaustion
		limitedReader := io.LimitReader(resp.Body, c.config.MaxResponseSize)

		// Check for retryable HTTP status codes
		if shouldRetry(resp.StatusCode) {
			bodyBytes, _ := io.ReadAll(limitedReader)
			c.logger.Debug("retryable http status",
				zap.Int("status", resp.StatusCode),
				zap.String("url", req.URL.String()),
			)
			return fmt.Errorf("retryable status %d: %s", resp.StatusCode, string(bodyBytes))
		}

		// Non-2xx status that shouldn't retry
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			bodyBytes, _ := io.ReadAll(limitedReader)
			return backoff.Permanent(fmt.Errorf("http %d: %s", resp.StatusCode, string(bodyBytes)))
		}

		// Decode JSON response
		decoder := json.NewDecoder(limitedReader)
		if decodeErr := decoder.Decode(result); decodeErr != nil {
			return backoff.Permanent(fmt.Errorf("json decode error: %w", decodeErr))
		}

		return nil
	}

	// Configure exponential backoff with jitter
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = c.config.InitialInterval
	expBackoff.MaxInterval = c.config.MaxInterval
	expBackoff.MaxElapsedTime = c.config.RequestTimeout

	// Clamp MaxRetries to zero if negative before converting to uint64
	maxRetries := c.config.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}
	backoffWithRetries := backoff.WithMaxRetries(expBackoff, uint64(maxRetries)) // #nosec G115
	backoffWithContext := backoff.WithContext(backoffWithRetries, ctx)

	err := backoff.Retry(operation, backoffWithContext)

	duration := time.Since(startTime)

	if err != nil {
		c.logger.Warn("http request failed after retries",
			zap.String("req_id", reqID),
			zap.String("url", req.URL.String()),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return err
	}

	c.logger.Debug("http request completed",
		zap.String("req_id", reqID),
		zap.String("url", req.URL.String()),
		zap.Duration("duration", duration),
	)

	return nil
}

// shouldRetry determines if an HTTP status code warrants a retry
func shouldRetry(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests, // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		return true
	default:
		return false
	}
}

// Get is a convenience wrapper for GET requests
func (c *Client) Get(ctx context.Context, url string, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	// Use configured UserAgent or default
	userAgent := c.config.UserAgent
	if userAgent == "" {
		userAgent = "hypermcp"
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")
	// Don't set Accept-Encoding manually - let Go's Transport handle gzip automatically
	// when DisableCompression is false

	return c.DoJSON(ctx, req, result)
}
