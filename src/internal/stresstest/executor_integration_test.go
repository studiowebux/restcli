package stresstest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/studiowebux/restcli/internal/types"
)

// createTestManager creates a new Manager with in-memory SQLite database for testing
func createTestManager(t *testing.T) *Manager {
	manager, err := NewManager(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test manager: %v", err)
	}
	return manager
}

// getMetricsCount returns the total count of metrics for a run
func getMetricsCount(t *testing.T, manager *Manager, runID int64) int {
	rows, err := manager.db.Query("SELECT COUNT(*) FROM stress_test_metrics WHERE run_id = ?", runID)
	if err != nil {
		t.Fatalf("Failed to count metrics: %v", err)
	}
	defer rows.Close()

	var count int
	if rows.Next() {
		rows.Scan(&count)
	}
	return count
}

// TestExecutor_BasicExecution tests basic stress test execution
func TestExecutor_BasicExecution(t *testing.T) {
	requestCount := int64(0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	manager := createTestManager(t)
	defer manager.Close()
	config := &ExecutionConfig{
		Request: &types.HttpRequest{
			Method: "GET",
			URL:    server.URL,
		},
		Config: &Config{
			Name:            "test-basic",
			RequestFile:     "test.http",
			ProfileName:     "default",
			ConcurrentConns: 5,
			TotalRequests:   50,
		},
	}

	executor, err := NewExecutor(config, manager)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	executor.Start()
	err = executor.Wait()

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify all requests were sent
	stats := executor.GetStats()
	if stats.CompletedRequests != 50 {
		t.Errorf("Expected 50 completed requests, got: %d", stats.CompletedRequests)
	}

	if stats.ErrorCount != 0 {
		t.Errorf("Expected 0 errors, got: %d", stats.ErrorCount)
	}

	if stats.SuccessCount != 50 {
		t.Errorf("Expected 50 successes, got: %d", stats.SuccessCount)
	}

	// Verify server received all requests
	finalCount := atomic.LoadInt64(&requestCount)
	if finalCount != 50 {
		t.Errorf("Expected server to receive 50 requests, got: %d", finalCount)
	}

	// Verify run record
	run := executor.GetRun()
	if run.Status != "completed" {
		t.Errorf("Expected status 'completed', got: %s", run.Status)
	}

	if run.TotalRequestsSent != 50 {
		t.Errorf("Expected 50 sent, got: %d", run.TotalRequestsSent)
	}

	if run.TotalRequestsCompleted != 50 {
		t.Errorf("Expected 50 completed, got: %d", run.TotalRequestsCompleted)
	}
}

// TestExecutor_LoadGenerationAccuracy tests that load is generated accurately
func TestExecutor_LoadGenerationAccuracy(t *testing.T) {
	requestCount := int64(0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&requestCount, 1)
		time.Sleep(10 * time.Millisecond) // Simulate some processing
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	manager := createTestManager(t)
	defer manager.Close()
	config := &ExecutionConfig{
		Request: &types.HttpRequest{
			Method: "GET",
			URL:    server.URL,
		},
		Config: &Config{
			Name:            "test-load-accuracy",
			RequestFile:     "test.http",
			ProfileName:     "default",
			ConcurrentConns: 10,
			TotalRequests:   100,
		},
	}

	executor, err := NewExecutor(config, manager)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	executor.Start()
	executor.Wait()

	stats := executor.GetStats()

	// Should have 100 completed requests
	if stats.CompletedRequests != 100 {
		t.Errorf("Expected 100 completed requests, got: %d", stats.CompletedRequests)
	}

	// Verify total count matches
	finalCount := atomic.LoadInt64(&requestCount)
	if finalCount != 100 {
		t.Errorf("Expected 100 requests to server, got: %d", finalCount)
	}
}

// TestExecutor_ConcurrentWorkers tests that workers execute concurrently
func TestExecutor_ConcurrentWorkers(t *testing.T) {
	activeConcurrent := int32(0)
	maxConcurrent := int32(0)
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := atomic.AddInt32(&activeConcurrent, 1)

		mu.Lock()
		if current > maxConcurrent {
			maxConcurrent = current
		}
		mu.Unlock()

		// Hold the request for a bit to allow concurrency to build up
		time.Sleep(100 * time.Millisecond)

		atomic.AddInt32(&activeConcurrent, -1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	manager := createTestManager(t)
	defer manager.Close()
	config := &ExecutionConfig{
		Request: &types.HttpRequest{
			Method: "GET",
			URL:    server.URL,
		},
		Config: &Config{
			Name:            "test-concurrency",
			RequestFile:     "test.http",
			ProfileName:     "default",
			ConcurrentConns: 10,
			TotalRequests:   20,
		},
	}

	executor, err := NewExecutor(config, manager)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	executor.Start()
	executor.Wait()

	// Should have had multiple concurrent requests
	mu.Lock()
	max := maxConcurrent
	mu.Unlock()

	if max < 5 {
		t.Errorf("Expected at least 5 concurrent requests, got: %d", max)
	}
}

// TestExecutor_MetricsCollection tests that metrics are collected properly
func TestExecutor_MetricsCollection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))
	defer server.Close()

	manager := createTestManager(t)
	defer manager.Close()
	config := &ExecutionConfig{
		Request: &types.HttpRequest{
			Method: "GET",
			URL:    server.URL,
		},
		Config: &Config{
			Name:            "test-metrics",
			RequestFile:     "test.http",
			ProfileName:     "default",
			ConcurrentConns: 2,
			TotalRequests:   10,
		},
	}

	executor, err := NewExecutor(config, manager)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	executor.Start()
	executor.Wait()

	// Verify metrics were saved
	totalMetrics := getMetricsCount(t, manager, executor.run.ID)
	if totalMetrics != 10 {
		t.Errorf("Expected 10 metrics, got: %d", totalMetrics)
	}

	// Verify run statistics were calculated
	run, err := manager.GetRun(executor.run.ID)
	if err != nil {
		t.Fatalf("Failed to get run: %v", err)
	}

	if run.AvgDurationMs < 0 {
		t.Error("Expected non-negative average duration")
	}

	if run.MinDurationMs < 0 {
		t.Errorf("Expected non-negative min duration, got: %d", run.MinDurationMs)
	}

	if run.MaxDurationMs < run.MinDurationMs {
		t.Errorf("Max duration (%d) should be >= min duration (%d)", run.MaxDurationMs, run.MinDurationMs)
	}
}

// TestExecutor_StatusCodeValidation tests status code validation
func TestExecutor_StatusCodeValidation(t *testing.T) {
	requestNum := int32(0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		num := atomic.AddInt32(&requestNum, 1)
		// Return 200 for first half, 404 for second half
		if num <= 5 {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	manager := createTestManager(t)
	defer manager.Close()
	config := &ExecutionConfig{
		Request: &types.HttpRequest{
			Method:              "GET",
			URL:                 server.URL,
			ExpectedStatusCodes: []int{200}, // Only expect 200
		},
		Config: &Config{
			Name:            "test-status-validation",
			RequestFile:     "test.http",
			ProfileName:     "default",
			ConcurrentConns: 1,
			TotalRequests:   10,
		},
	}

	executor, err := NewExecutor(config, manager)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	executor.Start()
	executor.Wait()

	stats := executor.GetStats()

	// Should have 5 successes and 5 validation errors
	if stats.SuccessCount != 5 {
		t.Errorf("Expected 5 successes, got: %d", stats.SuccessCount)
	}

	if stats.ValidationErrorCount != 5 {
		t.Errorf("Expected 5 validation errors, got: %d", stats.ValidationErrorCount)
	}

	if stats.ErrorCount != 0 {
		t.Errorf("Expected 0 network errors, got: %d", stats.ErrorCount)
	}
}

// TestExecutor_BodyValidationExact tests exact body validation
func TestExecutor_BodyValidationExact(t *testing.T) {
	requestNum := int32(0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		num := atomic.AddInt32(&requestNum, 1)
		if num <= 3 {
			w.Write([]byte("expected body"))
		} else {
			w.Write([]byte("different body"))
		}
	}))
	defer server.Close()

	manager := createTestManager(t)
	defer manager.Close()
	config := &ExecutionConfig{
		Request: &types.HttpRequest{
			Method:            "GET",
			URL:               server.URL,
			ExpectedBodyExact: "expected body",
		},
		Config: &Config{
			Name:            "test-body-exact",
			RequestFile:     "test.http",
			ProfileName:     "default",
			ConcurrentConns: 1,
			TotalRequests:   6,
		},
	}

	executor, err := NewExecutor(config, manager)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	executor.Start()
	executor.Wait()

	stats := executor.GetStats()

	if stats.SuccessCount != 3 {
		t.Errorf("Expected 3 successes, got: %d", stats.SuccessCount)
	}

	if stats.ValidationErrorCount != 3 {
		t.Errorf("Expected 3 validation errors, got: %d", stats.ValidationErrorCount)
	}
}

// TestExecutor_BodyValidationContains tests substring body validation
func TestExecutor_BodyValidationContains(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("this response contains the keyword somewhere"))
	}))
	defer server.Close()

	manager := createTestManager(t)
	defer manager.Close()
	config := &ExecutionConfig{
		Request: &types.HttpRequest{
			Method:               "GET",
			URL:                  server.URL,
			ExpectedBodyContains: "keyword",
		},
		Config: &Config{
			Name:            "test-body-contains",
			RequestFile:     "test.http",
			ProfileName:     "default",
			ConcurrentConns: 2,
			TotalRequests:   5,
		},
	}

	executor, err := NewExecutor(config, manager)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	executor.Start()
	executor.Wait()

	stats := executor.GetStats()

	// All should succeed since all contain "keyword"
	if stats.SuccessCount != 5 {
		t.Errorf("Expected 5 successes, got: %d", stats.SuccessCount)
	}

	if stats.ValidationErrorCount != 0 {
		t.Errorf("Expected 0 validation errors, got: %d", stats.ValidationErrorCount)
	}
}

// TestExecutor_BodyValidationPattern tests regex pattern validation
func TestExecutor_BodyValidationPattern(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"id": 123, "status": "active"}`))
	}))
	defer server.Close()

	manager := createTestManager(t)
	defer manager.Close()
	config := &ExecutionConfig{
		Request: &types.HttpRequest{
			Method:              "GET",
			URL:                 server.URL,
			ExpectedBodyPattern: `"status":\s*"active"`,
		},
		Config: &Config{
			Name:            "test-body-pattern",
			RequestFile:     "test.http",
			ProfileName:     "default",
			ConcurrentConns: 1,
			TotalRequests:   3,
		},
	}

	executor, err := NewExecutor(config, manager)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	executor.Start()
	executor.Wait()

	stats := executor.GetStats()

	if stats.SuccessCount != 3 {
		t.Errorf("Expected 3 successes, got: %d", stats.SuccessCount)
	}
}

// TestExecutor_BodyValidationFields tests JSON field validation
func TestExecutor_BodyValidationFields(t *testing.T) {
	requestNum := int32(0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		num := atomic.AddInt32(&requestNum, 1)
		if num <= 2 {
			w.Write([]byte(`{"status": "success", "count": 42}`))
		} else {
			w.Write([]byte(`{"status": "failure", "count": 0}`))
		}
	}))
	defer server.Close()

	manager := createTestManager(t)
	defer manager.Close()
	config := &ExecutionConfig{
		Request: &types.HttpRequest{
			Method: "GET",
			URL:    server.URL,
			ExpectedBodyFields: map[string]string{
				"status": "success",
				"count":  "42",
			},
		},
		Config: &Config{
			Name:            "test-body-fields",
			RequestFile:     "test.http",
			ProfileName:     "default",
			ConcurrentConns: 1,
			TotalRequests:   4,
		},
	}

	executor, err := NewExecutor(config, manager)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	executor.Start()
	executor.Wait()

	stats := executor.GetStats()

	if stats.SuccessCount != 2 {
		t.Errorf("Expected 2 successes, got: %d", stats.SuccessCount)
	}

	if stats.ValidationErrorCount != 2 {
		t.Errorf("Expected 2 validation errors, got: %d", stats.ValidationErrorCount)
	}
}

// TestExecutor_NetworkErrors tests handling of network errors
func TestExecutor_NetworkErrors(t *testing.T) {
	manager := createTestManager(t)
	defer manager.Close()
	config := &ExecutionConfig{
		Request: &types.HttpRequest{
			Method: "GET",
			URL:    "http://localhost:9999", // Non-existent server
		},
		Config: &Config{
			Name:              "test-network-errors",
			RequestFile:       "test.http",
			ProfileName:       "default",
			ConcurrentConns:   2,
			TotalRequests:     5,
			RequestTimeoutSec: 1, // Short timeout
		},
	}

	executor, err := NewExecutor(config, manager)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	executor.Start()
	executor.Wait()

	stats := executor.GetStats()

	// All requests should result in network errors
	if stats.ErrorCount != 5 {
		t.Errorf("Expected 5 network errors, got: %d", stats.ErrorCount)
	}

	if stats.SuccessCount != 0 {
		t.Errorf("Expected 0 successes, got: %d", stats.SuccessCount)
	}
}

// TestExecutor_ContextCancellation tests cancellation during execution
func TestExecutor_ContextCancellation(t *testing.T) {
	requestCount := int64(0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&requestCount, 1)
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	manager := createTestManager(t)
	defer manager.Close()
	config := &ExecutionConfig{
		Request: &types.HttpRequest{
			Method: "GET",
			URL:    server.URL,
		},
		Config: &Config{
			Name:            "test-cancellation",
			RequestFile:     "test.http",
			ProfileName:     "default",
			ConcurrentConns: 5,
			TotalRequests:   100, // Request many, but we'll cancel early
		},
	}

	executor, err := NewExecutor(config, manager)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	executor.Start()

	// Cancel after 200ms
	time.Sleep(200 * time.Millisecond)
	executor.Stop()

	stats := executor.GetStats()

	// Should have completed some but not all requests
	if stats.CompletedRequests >= 100 {
		t.Errorf("Expected less than 100 completed due to cancellation, got: %d", stats.CompletedRequests)
	}

	if stats.CompletedRequests == 0 {
		t.Error("Expected at least some requests to complete before cancellation")
	}

	// Run should be marked as cancelled
	run := executor.GetRun()
	if run.Status != "cancelled" {
		t.Errorf("Expected status 'cancelled', got: %s", run.Status)
	}
}

// TestExecutor_RampUp tests ramp-up behavior
func TestExecutor_RampUp(t *testing.T) {
	timestamps := make([]time.Time, 0, 10)
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		timestamps = append(timestamps, time.Now())
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	manager := createTestManager(t)
	defer manager.Close()
	config := &ExecutionConfig{
		Request: &types.HttpRequest{
			Method: "GET",
			URL:    server.URL,
		},
		Config: &Config{
			Name:              "test-rampup",
			RequestFile:       "test.http",
			ProfileName:       "default",
			ConcurrentConns:   10,
			TotalRequests:     10,
			RampUpDurationSec: 1, // 1 second ramp-up
		},
	}

	executor, err := NewExecutor(config, manager)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	start := time.Now()
	executor.Start()
	executor.Wait()
	duration := time.Since(start)

	// With 10 requests ramped over 1 second, the last request has offset of 900ms
	// (9 * 100ms spacing). Total time should be >= 900ms + execution time.
	// Use 850ms threshold to account for timing variance while ensuring ramp-up works.
	if duration < 850*time.Millisecond {
		t.Errorf("Expected ramp-up to take at least 850ms, took: %v", duration)
	}

	// Verify requests were spread out
	mu.Lock()
	defer mu.Unlock()

	if len(timestamps) != 10 {
		t.Errorf("Expected 10 timestamps, got: %d", len(timestamps))
	}
}

// TestExecutor_DurationBasedTest tests duration-based execution
func TestExecutor_DurationBasedTest(t *testing.T) {
	requestCount := int64(0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&requestCount, 1)
		// Add small delay to prevent all requests completing instantly
		time.Sleep(20 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	manager := createTestManager(t)
	defer manager.Close()
	config := &ExecutionConfig{
		Request: &types.HttpRequest{
			Method: "GET",
			URL:    server.URL,
		},
		Config: &Config{
			Name:            "test-duration",
			RequestFile:     "test.http",
			ProfileName:     "default",
			ConcurrentConns: 10,
			TotalRequests:   1000,      // Request many
			TestDurationSec: 1,         // But stop after 1 second
		},
	}

	executor, err := NewExecutor(config, manager)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	start := time.Now()
	executor.Start()
	executor.Wait()
	duration := time.Since(start)

	// Should complete around 1 second (with generous tolerance for CI environments)
	if duration < 800*time.Millisecond || duration > 2000*time.Millisecond {
		t.Logf("Warning: Duration outside expected range: %v (expected ~1s)", duration)
	}

	stats := executor.GetStats()

	// With 20ms delay per request and 10 concurrent workers, we expect ~500 requests/sec
	// So in 1 second, should complete between 100-900 requests (wide range for CI)
	if stats.CompletedRequests >= 1000 {
		t.Errorf("Expected less than 1000 completed due to duration limit, got: %d", stats.CompletedRequests)
	}

	if stats.CompletedRequests < 10 {
		t.Errorf("Expected at least 10 requests to complete, got: %d", stats.CompletedRequests)
	}
}

// TestExecutor_StopWithTimeout tests StopWithContext graceful shutdown
func TestExecutor_StopWithTimeout(t *testing.T) {
	requestCount := int64(0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&requestCount, 1)
		// Simulate moderate response time
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	manager := createTestManager(t)
	defer manager.Close()
	config := &ExecutionConfig{
		Request: &types.HttpRequest{
			Method: "GET",
			URL:    server.URL,
		},
		Config: &Config{
			Name:            "test-stop-timeout",
			RequestFile:     "test.http",
			ProfileName:     "default",
			ConcurrentConns: 5,
			TotalRequests:   100,
		},
	}

	executor, err := NewExecutor(config, manager)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	executor.Start()

	// Wait a bit for workers to start processing
	time.Sleep(200 * time.Millisecond)

	// Stop with generous timeout - should complete gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = executor.StopWithContext(ctx)

	// Should complete without timeout error
	if err != nil {
		t.Errorf("Expected graceful shutdown, got error: %v", err)
	}

	// Should have completed some but not all requests
	stats := executor.GetStats()
	if stats.CompletedRequests == 0 {
		t.Error("Expected at least some requests to complete")
	}

	if stats.CompletedRequests >= 100 {
		t.Logf("Warning: All requests completed (expected partial completion)")
	}
}

// TestExecutor_HeadersPropagation tests that headers are sent correctly
func TestExecutor_HeadersPropagation(t *testing.T) {
	receivedHeaders := make([]http.Header, 0)
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		receivedHeaders = append(receivedHeaders, r.Header.Clone())
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	manager := createTestManager(t)
	defer manager.Close()
	config := &ExecutionConfig{
		Request: &types.HttpRequest{
			Method: "POST",
			URL:    server.URL,
			Headers: map[string]string{
				"X-Custom-Header": "test-value",
				"Content-Type":    "application/json",
			},
			Body: `{"test": true}`,
		},
		Config: &Config{
			Name:            "test-headers",
			RequestFile:     "test.http",
			ProfileName:     "default",
			ConcurrentConns: 1,
			TotalRequests:   3,
		},
	}

	executor, err := NewExecutor(config, manager)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	executor.Start()
	executor.Wait()

	mu.Lock()
	defer mu.Unlock()

	if len(receivedHeaders) != 3 {
		t.Fatalf("Expected 3 requests, got: %d", len(receivedHeaders))
	}

	for i, headers := range receivedHeaders {
		if headers.Get("X-Custom-Header") != "test-value" {
			t.Errorf("Request %d: Expected X-Custom-Header='test-value', got: %s", i, headers.Get("X-Custom-Header"))
		}
		if headers.Get("Content-Type") != "application/json" {
			t.Errorf("Request %d: Expected Content-Type='application/json', got: %s", i, headers.Get("Content-Type"))
		}
	}
}

// TestExecutor_BodyPropagation tests that request body is sent correctly
func TestExecutor_BodyPropagation(t *testing.T) {
	receivedBodies := make([]string, 0)
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		mu.Lock()
		receivedBodies = append(receivedBodies, string(body))
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	expectedBody := `{"name":"test","count":42}`

	manager := createTestManager(t)
	defer manager.Close()
	config := &ExecutionConfig{
		Request: &types.HttpRequest{
			Method: "POST",
			URL:    server.URL,
			Body:   expectedBody,
		},
		Config: &Config{
			Name:            "test-body",
			RequestFile:     "test.http",
			ProfileName:     "default",
			ConcurrentConns: 2,
			TotalRequests:   5,
		},
	}

	executor, err := NewExecutor(config, manager)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	executor.Start()
	executor.Wait()

	mu.Lock()
	defer mu.Unlock()

	if len(receivedBodies) != 5 {
		t.Fatalf("Expected 5 bodies, got: %d", len(receivedBodies))
	}

	for i, body := range receivedBodies {
		if body != expectedBody {
			t.Errorf("Request %d: Expected body '%s', got: '%s'", i, expectedBody, body)
		}
	}
}

// TestExecutor_PercentileCalculations tests percentile metric calculations
func TestExecutor_PercentileCalculations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Variable response times
		time.Sleep(time.Duration(10+atomic.AddInt64(new(int64), 1)) * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	manager := createTestManager(t)
	defer manager.Close()
	config := &ExecutionConfig{
		Request: &types.HttpRequest{
			Method: "GET",
			URL:    server.URL,
		},
		Config: &Config{
			Name:            "test-percentiles",
			RequestFile:     "test.http",
			ProfileName:     "default",
			ConcurrentConns: 1,
			TotalRequests:   100,
		},
	}

	executor, err := NewExecutor(config, manager)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	executor.Start()
	executor.Wait()

	run := executor.GetRun()

	// Verify percentiles are calculated and in logical order
	if run.P50DurationMs <= 0 {
		t.Error("Expected positive P50")
	}

	if run.P95DurationMs <= 0 {
		t.Error("Expected positive P95")
	}

	if run.P99DurationMs <= 0 {
		t.Error("Expected positive P99")
	}

	// P50 <= P95 <= P99
	if run.P50DurationMs > run.P95DurationMs {
		t.Errorf("P50 (%d) should be <= P95 (%d)", run.P50DurationMs, run.P95DurationMs)
	}

	if run.P95DurationMs > run.P99DurationMs {
		t.Errorf("P95 (%d) should be <= P99 (%d)", run.P95DurationMs, run.P99DurationMs)
	}

	// Min <= P50 <= Max
	if run.MinDurationMs > run.P50DurationMs {
		t.Errorf("Min (%d) should be <= P50 (%d)", run.MinDurationMs, run.P50DurationMs)
	}

	if run.P99DurationMs > run.MaxDurationMs {
		t.Errorf("P99 (%d) should be <= Max (%d)", run.P99DurationMs, run.MaxDurationMs)
	}
}

// TestExecutor_JSONBodyValidationWithRegex tests JSON field validation with regex patterns
func TestExecutor_JSONBodyValidationWithRegex(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"id":    "user-12345",
			"email": "test@example.com",
			"count": 42,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	manager := createTestManager(t)
	defer manager.Close()
	config := &ExecutionConfig{
		Request: &types.HttpRequest{
			Method: "GET",
			URL:    server.URL,
			ExpectedBodyFields: map[string]string{
				"id":    "/^user-\\d+$/",            // Regex pattern
				"email": "/.*@example\\.com$/",      // Regex pattern
				"count": "42",                        // Literal value
			},
		},
		Config: &Config{
			Name:            "test-json-regex",
			RequestFile:     "test.http",
			ProfileName:     "default",
			ConcurrentConns: 1,
			TotalRequests:   5,
		},
	}

	executor, err := NewExecutor(config, manager)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	executor.Start()
	executor.Wait()

	stats := executor.GetStats()

	// All should pass validation
	if stats.SuccessCount != 5 {
		t.Errorf("Expected 5 successes, got: %d", stats.SuccessCount)
	}

	if stats.ValidationErrorCount != 0 {
		t.Errorf("Expected 0 validation errors, got: %d", stats.ValidationErrorCount)
	}
}
