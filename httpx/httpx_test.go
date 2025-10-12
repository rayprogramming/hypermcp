package httpx

import (
"context"
"encoding/json"
"errors"
"net/http"
"net/http/httptest"
"testing"
"time"

"go.uber.org/zap/zaptest"
)

func TestClient_DoJSON_Success(t *testing.T) {
// Create test server
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
response := map[string]string{"message": "success"}
w.Header().Set("Content-Type", "application/json")
if err := json.NewEncoder(w).Encode(response); err != nil {
http.Error(w, err.Error(), http.StatusInternalServerError)
}
}))
defer server.Close()

logger := zaptest.NewLogger(t)
client, err := New(logger)
if err != nil {
t.Fatalf("failed to create client: %v", err)
}

var result map[string]string
err = client.Get(context.Background(), server.URL, &result)

if err != nil {
t.Fatalf("expected no error, got %v", err)
}

if result["message"] != "success" {
t.Errorf("expected message=success, got %s", result["message"])
}
}

func TestClient_DoJSON_Retry(t *testing.T) {
attempts := 0
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
attempts++
if attempts < 3 {
w.WriteHeader(http.StatusInternalServerError)
return
}
response := map[string]string{"message": "success"}
if err := json.NewEncoder(w).Encode(response); err != nil {
http.Error(w, err.Error(), http.StatusInternalServerError)
}
}))
defer server.Close()

logger := zaptest.NewLogger(t)
client, err := New(logger)
if err != nil {
t.Fatalf("failed to create client: %v", err)
}

var result map[string]string
err = client.Get(context.Background(), server.URL, &result)

if err != nil {
t.Fatalf("expected no error after retries, got %v", err)
}

if attempts < 3 {
t.Errorf("expected at least 3 attempts, got %d", attempts)
}
}

func TestClient_DoJSON_ContextCancellation(t *testing.T) {
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
time.Sleep(200 * time.Millisecond)
if err := json.NewEncoder(w).Encode(map[string]string{"message": "success"}); err != nil {
http.Error(w, err.Error(), http.StatusInternalServerError)
}
}))
defer server.Close()

logger := zaptest.NewLogger(t)
client, err := New(logger)
if err != nil {
t.Fatalf("failed to create client: %v", err)
}

ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
defer cancel()

var result map[string]string
err = client.Get(ctx, server.URL, &result)

if err == nil {
t.Fatal("expected context cancellation error")
}
}

func BenchmarkClient_DoJSON(b *testing.B) {
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
response := map[string]string{
"title":   "Test Article",
"extract": "This is a test extract with some content.",
"url":     "https://example.com",
}
if err := json.NewEncoder(w).Encode(response); err != nil {
http.Error(w, err.Error(), http.StatusInternalServerError)
}
}))
defer server.Close()

logger := zaptest.NewLogger(b)
client, err := New(logger)
if err != nil {
b.Fatalf("failed to create client: %v", err)
}
ctx := context.Background()

b.ResetTimer()
b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var result map[string]string
		err2 := client.Get(ctx, server.URL, &result)
		if err2 != nil {
			b.Fatal(err2)
		}
	}
}

func TestClient_UserAgent(t *testing.T) {
tests := []struct {
name            string
userAgent       string
expectedUA      string
}{
{
name:            "default user agent",
userAgent:       "",
expectedUA:      "hypermcp",
},
{
name:            "custom user agent",
userAgent:       "my-mcp-server/1.0.0",
expectedUA:      "my-mcp-server/1.0.0",
},
{
name:            "user agent with version",
userAgent:       "hypermcp/0.3.0",
expectedUA:      "hypermcp/0.3.0",
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
// Create test server that captures User-Agent
var capturedUA string
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
capturedUA = r.Header.Get("User-Agent")
w.Header().Set("Content-Type", "application/json")
if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
http.Error(w, err.Error(), http.StatusInternalServerError)
}
}))
defer server.Close()

logger := zaptest.NewLogger(t)
cfg := DefaultConfig()
cfg.UserAgent = tt.userAgent
client, err := NewWithConfig(cfg, logger)
if err != nil {
t.Fatalf("failed to create client: %v", err)
}

var result map[string]string
err = client.Get(context.Background(), server.URL, &result)

if err != nil {
t.Fatalf("expected no error, got %v", err)
}

if capturedUA != tt.expectedUA {
t.Errorf("expected User-Agent %q, got %q", tt.expectedUA, capturedUA)
}
})
}
}

func TestConfig_Validate(t *testing.T) {
logger := zaptest.NewLogger(t)

tests := []struct {
name          string
cfg           Config
wantError     bool
expectedError error
}{
{
name:      "valid config",
cfg:       DefaultConfig(),
wantError: false,
},
{
name: "invalid DialTimeout",
cfg: Config{
DialTimeout:           0,
TLSHandshakeTimeout:   2 * time.Second,
ResponseHeaderTimeout: 4 * time.Second,
RequestTimeout:        6 * time.Second,
MaxRetries:            3,
InitialInterval:       100 * time.Millisecond,
MaxInterval:           2 * time.Second,
MaxResponseSize:       10 * 1024 * 1024,
MaxIdleConns:          100,
MaxIdleConnsPerHost:   10,
IdleConnTimeout:       90 * time.Second,
},
wantError:     true,
expectedError: ErrInvalidTimeout,
},
{
name: "negative MaxRetries",
cfg: Config{
DialTimeout:           2 * time.Second,
TLSHandshakeTimeout:   2 * time.Second,
ResponseHeaderTimeout: 4 * time.Second,
RequestTimeout:        6 * time.Second,
MaxRetries:            -1,
InitialInterval:       100 * time.Millisecond,
MaxInterval:           2 * time.Second,
MaxResponseSize:       10 * 1024 * 1024,
MaxIdleConns:          100,
MaxIdleConnsPerHost:   10,
IdleConnTimeout:       90 * time.Second,
},
wantError:     true,
expectedError: ErrInvalidMaxRetries,
},
{
name: "invalid MaxResponseSize",
cfg: Config{
DialTimeout:           2 * time.Second,
TLSHandshakeTimeout:   2 * time.Second,
ResponseHeaderTimeout: 4 * time.Second,
RequestTimeout:        6 * time.Second,
MaxRetries:            3,
InitialInterval:       100 * time.Millisecond,
MaxInterval:           2 * time.Second,
MaxResponseSize:       0,
MaxIdleConns:          100,
MaxIdleConnsPerHost:   10,
IdleConnTimeout:       90 * time.Second,
},
wantError:     true,
expectedError: ErrInvalidMaxResponseSize,
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
_, err := NewWithConfig(tt.cfg, logger)

if !tt.wantError {
if err != nil {
t.Errorf("unexpected error: %v", err)
}
return
}

// Error expected
if err == nil {
t.Error("expected error but got none")
return
}

// Check that the error is the expected sentinel error
if tt.expectedError != nil && !errors.Is(err, tt.expectedError) {
t.Errorf("expected error %v, got %v", tt.expectedError, err)
}

// Check that it's a ConfigError
var cfgErr *ConfigError
if !errors.As(err, &cfgErr) {
t.Errorf("expected ConfigError type, got %T", err)
}
})
}
}
