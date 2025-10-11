package httpx

import (
	"context"
	"encoding/json"
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
	client := New(logger)

	var result map[string]string
	err := client.Get(context.Background(), server.URL, &result)

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
	client := New(logger)

	var result map[string]string
	err := client.Get(context.Background(), server.URL, &result)

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
	client := New(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	var result map[string]string
	err := client.Get(ctx, server.URL, &result)

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
	client := New(logger)
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var result map[string]string
		if err := client.Get(ctx, server.URL, &result); err != nil {
			b.Fatal(err)
		}
	}
}
