// Package httpx provides a shared HTTP client with retry logic, timeouts, and helpers
package httpx

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
	"go.uber.org/zap"
)

const (
	// HTTP timeouts
	dialTimeout           = 2 * time.Second
	tlsHandshakeTimeout   = 2 * time.Second
	responseHeaderTimeout = 4 * time.Second
	requestTimeout        = 6 * time.Second

	// Retry configuration
	maxRetries      = 3
	initialInterval = 100 * time.Millisecond
	maxInterval     = 2 * time.Second

	// Request limits
	maxResponseSize = 10 * 1024 * 1024 // 10MB
)

// Client wraps an HTTP client with retry logic and performance optimizations
type Client struct {
	client *http.Client
	logger *zap.Logger
}

// New creates a new HTTP client with optimal settings for performance
func New(logger *zap.Logger) *Client {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   dialTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   tlsHandshakeTimeout,
		ResponseHeaderTimeout: responseHeaderTimeout,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		DisableCompression:    false, // Enable gzip
		ForceAttemptHTTP2:     true,
	}

	return &Client{
		client: &http.Client{
			Transport: transport,
			Timeout:   requestTimeout,
		},
		logger: logger,
	}
}

// DoJSON performs an HTTP request and unmarshals the JSON response
// It includes retry logic with exponential backoff for transient errors
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
		limitedReader := io.LimitReader(resp.Body, maxResponseSize)

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
	expBackoff.InitialInterval = initialInterval
	expBackoff.MaxInterval = maxInterval
	expBackoff.MaxElapsedTime = requestTimeout

	backoffWithRetries := backoff.WithMaxRetries(expBackoff, maxRetries)
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

	req.Header.Set("User-Agent", "hypermcp/0.1.0")
	req.Header.Set("Accept", "application/json")
	// Don't set Accept-Encoding manually - let Go's Transport handle gzip automatically
	// when DisableCompression is false

	return c.DoJSON(ctx, req, result)
}
