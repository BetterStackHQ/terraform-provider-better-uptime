package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestClientRateLimiting(t *testing.T) {
	var requestCount int32

	// Create a test server that returns 429 for the first 2 requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)

		if count <= 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("Rate limited"))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": "success"}`))
	}))
	defer server.Close()

	// Create client with short retry intervals for testing
	client, err := newClient(ClientConfig{
		BaseURL:      server.URL,
		Token:        "test-token",
		RetryMax:     3,
		RetryWaitMin: 100 * time.Millisecond,
		RetryWaitMax: 500 * time.Millisecond,
	})

	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Make a request - it should succeed after retries
	ctx := context.Background()
	resp, err := client.Get(ctx, "/test")

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	// Verify the request succeeded
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	// Verify that retries occurred
	finalCount := atomic.LoadInt32(&requestCount)
	if finalCount != 3 {
		t.Errorf("Expected 3 requests (2 failures + 1 success), got %d", finalCount)
	}
}

func TestClientNoRetryOnSuccess(t *testing.T) {
	var requestCount int32

	// Create a test server that always returns 200
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": "success"}`))
	}))
	defer server.Close()

	// Create client
	client, err := newClient(ClientConfig{
		BaseURL: server.URL,
		Token:   "test-token",
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Make a request
	ctx := context.Background()
	resp, err := client.Get(ctx, "/test")

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	// Verify only one request was made
	finalCount := atomic.LoadInt32(&requestCount)
	if finalCount != 1 {
		t.Errorf("Expected 1 request for successful response, got %d", finalCount)
	}
}

func TestClientRespectsRetryMax(t *testing.T) {
	var requestCount int32

	// Create a test server that always returns 429
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("Rate limited"))
	}))
	defer server.Close()

	// Create client with RetryMax = 2
	client, err := newClient(ClientConfig{
		BaseURL:      server.URL,
		Token:        "test-token",
		RetryMax:     2,
		RetryWaitMin: 50 * time.Millisecond,
		RetryWaitMax: 100 * time.Millisecond,
	})

	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Make a request - it should fail after retries
	ctx := context.Background()
	resp, err := client.Get(ctx, "/test")

	// The request should fail with an error after exhausting retries
	if err == nil {
		if resp != nil {
			resp.Body.Close()
		}
		t.Fatalf("Expected request to fail after exhausting retries, but it succeeded")
	}

	// Verify that exactly RetryMax+1 requests were made
	// (1 initial request + 2 retries = 3 total)
	finalCount := atomic.LoadInt32(&requestCount)
	if finalCount != 3 {
		t.Errorf("Expected 3 total requests (1 initial + 2 retries), got %d", finalCount)
	}
}

func TestClientTimeout(t *testing.T) {
	// Create a test server that delays response longer than timeout
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": "success"}`))
	}))
	defer server.Close()

	// Create client with very short timeout
	client, err := newClient(ClientConfig{
		BaseURL:    server.URL,
		Token:      "test-token",
		HTTPClient: &http.Client{Timeout: 100 * time.Millisecond},
	})

	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Make a request - it should timeout
	ctx := context.Background()
	_, err = client.Get(ctx, "/test")

	// The request should fail with a timeout error
	if err == nil {
		t.Fatalf("Expected request to timeout, but it succeeded")
	}

	// Verify it's a timeout error
	if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "deadline exceeded") {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}
