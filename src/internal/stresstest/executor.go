package stresstest

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/studiowebux/restcli/internal/types"
)

const (
	// HTTP client configuration timeouts
	TCPDialTimeout         = 5 * time.Second
	TCPKeepAliveInterval   = 30 * time.Second
	TLSHandshakeTimeout    = 5 * time.Second
	IdleConnTimeout        = 90 * time.Second
	ExpectContinueTimeout  = 1 * time.Second
	ShutdownGracePeriod    = 100 * time.Millisecond
)

// RequestTask represents a single request to be executed
type RequestTask struct {
	SequenceNum int
	StartOffset time.Duration
}

// RequestResult represents the result of a single request execution
type RequestResult struct {
	SequenceNum  int
	StatusCode   int
	DurationMs   int64
	ElapsedMs    int64
	RequestSize  int64
	ResponseSize int64
	Body         string // Response body for validation
	Error        error
	Timestamp    time.Time
}

// Executor handles concurrent stress test execution
type Executor struct {
	config        *ExecutionConfig
	manager       *Manager
	run           *Run
	stats         *Stats
	ctx           context.Context
	cancelFunc    context.CancelFunc
	wg            sync.WaitGroup
	workersReady  sync.WaitGroup
	requestChan   chan *RequestTask
	resultChan    chan *RequestResult
	closeOnce     sync.Once // Ensures resultChan is only closed once
	testStart     time.Time
	statsMu       sync.Mutex
	requestsSent  int // Actual number of requests queued/sent
	activeWorkers int32 // Atomic counter for active workers
	metricsBuf    []*Metric
	bufferSize    int
	httpClient    *http.Client // Shared HTTP client with connection pooling
}

// NewExecutor creates a new stress test executor
func NewExecutor(config *ExecutionConfig, manager *Manager) (*Executor, error) {
	if err := config.Config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Create run record
	run := &Run{
		ConfigName:  config.Config.Name,
		RequestFile: config.Config.RequestFile,
		ProfileName: config.Config.ProfileName,
		StartedAt:   time.Now(),
		Status:      "running",
	}
	if config.Config.ID > 0 {
		run.ConfigID = &config.Config.ID
	}

	if err := manager.CreateRun(run); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create run record: %w", err)
	}

	stats := NewStats()
	stats.TotalRequests = config.Config.TotalRequests

	// Create HTTP client with connection pooling and timeout
	httpClient, err := buildStressTestHTTPClient(config.Config, config.TLSConfig)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to build HTTP client: %w", err)
	}

	return &Executor{
		config:      config,
		manager:     manager,
		run:         run,
		stats:       stats,
		ctx:         ctx,
		cancelFunc:  cancel,
		requestChan: make(chan *RequestTask, config.Config.ConcurrentConns*2),
		resultChan:  make(chan *RequestResult, config.Config.ConcurrentConns*2),
		metricsBuf:  make([]*Metric, 0, 100),
		bufferSize:  100,
		httpClient:  httpClient,
	}, nil
}

// Start begins the stress test execution
func (e *Executor) Start() {
	e.testStart = time.Now()

	// Signal we need N workers to be ready before scheduling
	e.workersReady.Add(e.config.Config.ConcurrentConns)

	// Start worker goroutines
	for i := 0; i < e.config.Config.ConcurrentConns; i++ {
		e.wg.Add(1)
		go e.worker()
	}

	// Start result collector
	go e.collectResults()

	// Wait for all workers to be ready, then schedule requests
	// This prevents race condition where channel closes before workers start
	go func() {
		// Use a channel to make Wait() cancellable
		done := make(chan struct{})
		go func() {
			e.workersReady.Wait()
			close(done)
		}()

		select {
		case <-done:
			e.scheduleRequests()
		case <-e.ctx.Done():
			// Context cancelled before workers ready, exit scheduler
			return
		}
	}()

	// Start duration timer if test duration is set
	testDuration := e.config.Config.GetTestDuration()
	if testDuration > 0 {
		go e.durationTimer(testDuration)
	}
}

// durationTimer cancels the test after the specified duration
func (e *Executor) durationTimer(duration time.Duration) {
	select {
	case <-time.After(duration):
		// Duration elapsed, cancel the test
		e.cancelFunc()
	case <-e.ctx.Done():
		// Test already cancelled/completed
		return
	}
}

// Stop cancels the stress test execution
func (e *Executor) Stop() {
	e.cancelFunc()
	e.wg.Wait()
	e.closeResultChan()
	e.finalize("cancelled")
}

// StopWithContext cancels the stress test execution with a timeout
// Returns an error if cleanup doesn't complete within the context deadline
func (e *Executor) StopWithContext(ctx context.Context) error {
	// Signal cancellation
	e.cancelFunc()

	// Wait for workers with timeout
	done := make(chan struct{})
	go func() {
		e.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Workers completed successfully
		e.closeResultChan()
		e.finalize("cancelled")
		return nil
	case <-ctx.Done():
		// Timeout - workers didn't finish
		// Still close channels to prevent leaks, but don't wait further
		e.closeResultChan()
		e.finalize("cancelled (timeout)")
		return ctx.Err()
	}
}

// closeResultChan safely closes the result channel (only once)
func (e *Executor) closeResultChan() {
	e.closeOnce.Do(func() {
		close(e.resultChan)
	})
}

// Wait waits for the stress test to complete
func (e *Executor) Wait() error {
	// Wait for workers to finish
	e.wg.Wait()
	e.closeResultChan()

	// Wait for result collector to finish
	// (it closes when resultChan is closed and drained)
	time.Sleep(ShutdownGracePeriod)

	// Determine status based on completion
	status := "completed"
	e.statsMu.Lock()
	completedRequests := e.stats.CompletedRequests
	totalRequests := e.stats.TotalRequests
	e.statsMu.Unlock()

	if completedRequests < totalRequests {
		// Not all requests were completed, likely stopped by duration or manual cancel
		testDuration := e.config.Config.GetTestDuration()
		elapsed := time.Since(e.testStart)

		if testDuration > 0 && elapsed >= testDuration {
			status = "completed" // Duration reached is still "completed"
		} else {
			status = "cancelled" // Manually cancelled or stopped early
		}
	}

	// Finalize the run
	e.finalize(status)

	return nil
}

// GetStats returns the current statistics (thread-safe)
func (e *Executor) GetStats() *Stats {
	e.statsMu.Lock()
	defer e.statsMu.Unlock()

	// Return a copy to avoid race conditions
	// Use configured total for accurate progress calculation
	statsCopy := &Stats{
		TotalRequests:        e.config.Config.TotalRequests,
		CompletedRequests:    e.stats.CompletedRequests,
		ErrorCount:           e.stats.ErrorCount,
		ValidationErrorCount: e.stats.ValidationErrorCount,
		SuccessCount:         e.stats.SuccessCount,
		ActiveWorkers:        int(atomic.LoadInt32(&e.activeWorkers)),
		TotalDurationMs:      e.stats.TotalDurationMs,
		MinDurationMs:        e.stats.MinDurationMs,
		MaxDurationMs:        e.stats.MaxDurationMs,
		Durations:            make([]int64, len(e.stats.Durations)),
	}
	copy(statsCopy.Durations, e.stats.Durations)

	return statsCopy
}

// GetRun returns the current run record
func (e *Executor) GetRun() *Run {
	return e.run
}

// IsExecutionComplete returns true if all requests have been processed or context cancelled
func (e *Executor) IsExecutionComplete() bool {
	e.statsMu.Lock()
	defer e.statsMu.Unlock()

	// Check if context is done (cancelled by duration or manual stop)
	select {
	case <-e.ctx.Done():
		return true
	default:
	}

	// Check if all sent requests have been completed
	// Use requestsSent (actual queued) not the configured total
	return e.requestsSent > 0 && e.stats.CompletedRequests >= e.requestsSent
}

// worker executes requests from the request channel
func (e *Executor) worker() {
	defer e.wg.Done()

	// Signal this worker is ready to receive tasks
	e.workersReady.Done()

	for {
		select {
		case <-e.ctx.Done():
			return
		case task, ok := <-e.requestChan:
			if !ok {
				return
			}

			// Wait for ramp-up offset if needed
			if task.StartOffset > 0 {
				select {
				case <-e.ctx.Done():
					return
				case <-time.After(task.StartOffset):
				}
			}

			// Execute request using shared HTTP client
			// Track active worker
			atomic.AddInt32(&e.activeWorkers, 1)
			start := time.Now()
			result, err := e.executeRequest(e.config.Request)
			duration := time.Since(start)
			elapsed := time.Since(e.testStart)
			atomic.AddInt32(&e.activeWorkers, -1)

			// Prepare result
			requestResult := &RequestResult{
				SequenceNum: task.SequenceNum,
				DurationMs:  duration.Milliseconds(),
				ElapsedMs:   elapsed.Milliseconds(),
				Error:       err,
				Timestamp:   time.Now(),
			}

			if result != nil {
				requestResult.StatusCode = result.Status
				requestResult.RequestSize = int64(result.RequestSize)
				requestResult.ResponseSize = int64(result.ResponseSize)
				requestResult.Body = result.Body

				// CRITICAL: Check the Error field in the result (string)
				// HTTP execution returns errors in result.Error field
				if result.Error != "" {
					requestResult.Error = fmt.Errorf("%s", result.Error)
				}
			}

			// Send result
			select {
			case <-e.ctx.Done():
				return
			case e.resultChan <- requestResult:
			}
		}
	}
}

// scheduleRequests schedules requests with optional ramp-up
func (e *Executor) scheduleRequests() {
	rampUpPerRequest := time.Duration(0)
	totalRequests := e.config.Config.TotalRequests
	rampUpDuration := e.config.Config.GetRampUpDuration()

	if rampUpDuration > 0 && totalRequests > 0 {
		rampUpPerRequest = rampUpDuration / time.Duration(totalRequests)
	}

	for i := 0; i < totalRequests; i++ {
		select {
		case <-e.ctx.Done():
			close(e.requestChan)
			return
		case e.requestChan <- &RequestTask{
			SequenceNum: i,
			StartOffset: time.Duration(i) * rampUpPerRequest,
		}:
			// Track that we actually sent/queued this request
			e.statsMu.Lock()
			e.requestsSent++
			e.statsMu.Unlock()
		}
	}

	close(e.requestChan)
}

// validateBody validates the response body against expected patterns
// Returns empty string if validation passes, or error message if validation fails
func (e *Executor) validateBody(body string) string {
	req := e.config.Request

	// Check ExpectedBodyExact (exact string match)
	if req.ExpectedBodyExact != "" {
		if body != req.ExpectedBodyExact {
			return fmt.Sprintf("body does not match expected exact value (expected: %q, got: %q)", req.ExpectedBodyExact, body)
		}
	}

	// Check ExpectedBodyContains (substring match)
	if req.ExpectedBodyContains != "" {
		if !strings.Contains(body, req.ExpectedBodyContains) {
			return fmt.Sprintf("body does not contain expected substring: %s", req.ExpectedBodyContains)
		}
	}

	// Check ExpectedBodyPattern (regex match)
	if req.ExpectedBodyPattern != "" {
		matched, err := regexp.MatchString(req.ExpectedBodyPattern, body)
		if err != nil {
			return fmt.Sprintf("invalid body pattern regex: %v", err)
		}
		if !matched {
			return fmt.Sprintf("body does not match expected pattern: %s", req.ExpectedBodyPattern)
		}
	}

	// Check ExpectedBodyFields (partial JSON field matching)
	if len(req.ExpectedBodyFields) > 0 {
		var bodyJSON map[string]interface{}
		if err := json.Unmarshal([]byte(body), &bodyJSON); err != nil {
			return fmt.Sprintf("failed to parse JSON body for field validation: %v", err)
		}

		for fieldName, expectedValue := range req.ExpectedBodyFields {
			actualValue, exists := bodyJSON[fieldName]
			if !exists {
				return fmt.Sprintf("expected field '%s' not found in response", fieldName)
			}

			// Convert actual value to string for comparison
			actualStr := fmt.Sprintf("%v", actualValue)

			// Check if expected value is a regex pattern (starts and ends with /)
			if strings.HasPrefix(expectedValue, "/") && strings.HasSuffix(expectedValue, "/") {
				pattern := expectedValue[1 : len(expectedValue)-1]
				matched, err := regexp.MatchString(pattern, actualStr)
				if err != nil {
					return fmt.Sprintf("invalid regex pattern for field '%s': %v", fieldName, err)
				}
				if !matched {
					return fmt.Sprintf("field '%s' value '%s' does not match pattern '%s'", fieldName, actualStr, pattern)
				}
			} else {
				// Literal value comparison
				if actualStr != expectedValue {
					return fmt.Sprintf("field '%s' expected '%s' but got '%s'", fieldName, expectedValue, actualStr)
				}
			}
		}
	}

	return "" // All validations passed
}

// collectResults collects and processes request results
func (e *Executor) collectResults() {
	for result := range e.resultChan {
		// Determine error types
		isNetworkError := result.Error != nil || result.StatusCode == 0
		isValidationError := false
		validationErrorMsg := ""

		// Only validate if no network error occurred
		if !isNetworkError {
			// Check status code validation
			if !e.config.Request.IsExpectedStatus(result.StatusCode) {
				isValidationError = true
				validationErrorMsg = fmt.Sprintf("unexpected status %d", result.StatusCode)
			} else {
				// Status code is valid, check body validation
				bodyValidationErr := e.validateBody(result.Body)
				if bodyValidationErr != "" {
					isValidationError = true
					validationErrorMsg = bodyValidationErr
				}
			}
		}

		// Update statistics
		e.statsMu.Lock()
		e.stats.AddResult(result.DurationMs, isNetworkError, isValidationError)
		e.statsMu.Unlock()

		// Buffer metric for batch insert
		metric := &Metric{
			RunID:        e.run.ID,
			Timestamp:    result.Timestamp,
			ElapsedMs:    result.ElapsedMs,
			StatusCode:   result.StatusCode,
			DurationMs:   result.DurationMs,
			RequestSize:  result.RequestSize,
			ResponseSize: result.ResponseSize,
		}
		if result.Error != nil {
			metric.ErrorMessage = result.Error.Error()
		} else if isValidationError {
			metric.ValidationError = validationErrorMsg
		}

		e.metricsBuf = append(e.metricsBuf, metric)

		// Flush buffer if full
		if len(e.metricsBuf) >= e.bufferSize {
			e.flushMetrics()
		}
	}

	// Flush any remaining metrics
	e.flushMetrics()
}

// flushMetrics writes buffered metrics to database
func (e *Executor) flushMetrics() {
	if len(e.metricsBuf) == 0 {
		return
	}

	err := e.manager.SaveMetricsBatch(e.metricsBuf)
	if err != nil {
		// Log error but don't stop execution
		fmt.Printf("Failed to save metrics: %v\n", err)
	}

	e.metricsBuf = e.metricsBuf[:0]
}

// finalize completes the run record with final statistics
func (e *Executor) finalize(status string) {
	e.statsMu.Lock()
	defer e.statsMu.Unlock()

	now := time.Now()
	e.run.CompletedAt = &now
	e.run.Status = status
	e.run.TotalRequestsSent = e.requestsSent // Use actual sent count, not configured total
	e.run.TotalRequestsCompleted = e.stats.CompletedRequests
	e.run.TotalErrors = e.stats.ErrorCount
	e.run.TotalValidationErrors = e.stats.ValidationErrorCount
	e.run.AvgDurationMs = e.stats.AvgDurationMs()
	e.run.MinDurationMs = e.stats.Min()
	e.run.MaxDurationMs = e.stats.Max()
	e.run.P50DurationMs = e.stats.P50()
	e.run.P95DurationMs = e.stats.P95()
	e.run.P99DurationMs = e.stats.P99()

	err := e.manager.UpdateRun(e.run)
	if err != nil {
		fmt.Printf("Failed to update run record: %v\n", err)
	}
}

// buildStressTestHTTPClient creates an HTTP client optimized for stress testing
// with connection pooling, timeouts, and resource limits
func buildStressTestHTTPClient(config *Config, tlsConfig *types.TLSConfig) (*http.Client, error) {
	transport := &http.Transport{
		// Connection pool settings to prevent resource exhaustion
		MaxIdleConns:        config.ConcurrentConns,           // Total idle connections across all hosts
		MaxIdleConnsPerHost: config.ConcurrentConns,           // Idle connections per host
		MaxConnsPerHost:     config.ConcurrentConns * 2, // Max connections per host (active + idle)
		IdleConnTimeout:     IdleConnTimeout,            // How long idle connections stay open
		DisableKeepAlives:   false,                      // Enable keep-alive for connection reuse
		DisableCompression:  false,                            // Enable compression
		ForceAttemptHTTP2:   true,                             // Try HTTP/2 when possible

		// Timeouts for connection establishment
		DialContext: (&net.Dialer{
			Timeout:   TCPDialTimeout,        // Max time to establish TCP connection
			KeepAlive: TCPKeepAliveInterval, // Keep-alive probe interval
		}).DialContext,

		// TLS handshake timeout
		TLSHandshakeTimeout: TLSHandshakeTimeout,

		// Timeout for reading response headers
		ResponseHeaderTimeout: config.GetRequestTimeout(),

		// Expect Continue timeout
		ExpectContinueTimeout: ExpectContinueTimeout,
	}

	// Configure TLS if provided
	if tlsConfig != nil {
		tlsCfg := &tls.Config{
			InsecureSkipVerify: tlsConfig.InsecureSkipVerify,
		}

		// Load client certificate if provided (for mTLS)
		if tlsConfig.CertFile != "" && tlsConfig.KeyFile != "" {
			cert, err := tls.LoadX509KeyPair(tlsConfig.CertFile, tlsConfig.KeyFile)
			if err != nil {
				return nil, fmt.Errorf("failed to load client certificate: %w", err)
			}
			tlsCfg.Certificates = []tls.Certificate{cert}
		}

		// Load CA certificate if provided (for server verification)
		if tlsConfig.CAFile != "" {
			caCert, err := os.ReadFile(tlsConfig.CAFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read CA certificate: %w", err)
			}
			caCertPool := x509.NewCertPool()
			if !caCertPool.AppendCertsFromPEM(caCert) {
				return nil, fmt.Errorf("failed to parse CA certificate")
			}
			tlsCfg.RootCAs = caCertPool
		}

		transport.TLSClientConfig = tlsCfg
	}

	return &http.Client{
		Timeout:   config.GetRequestTimeout(),
		Transport: transport,
	}, nil
}

// executeRequest executes a single HTTP request using the shared client
func (e *Executor) executeRequest(req *types.HttpRequest) (*types.RequestResult, error) {
	startTime := time.Now()

	// Create HTTP request
	var bodyReader io.Reader
	requestSize := 0
	if req.Body != "" {
		bodyReader = bytes.NewBufferString(req.Body)
		requestSize = len(req.Body)
	}

	// Use context to allow cancellation of in-flight requests
	httpReq, err := http.NewRequestWithContext(e.ctx, req.Method, req.URL, bodyReader)
	if err != nil {
		return &types.RequestResult{
			Error:       fmt.Sprintf("failed to create request: %v", err),
			Duration:    0,
			RequestSize: requestSize,
		}, nil
	}

	// Set headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Execute request with shared client
	resp, err := e.httpClient.Do(httpReq)
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		// Connection failed, timeout, or other network error
		return &types.RequestResult{
			Error:       err.Error(),
			Duration:    duration,
			RequestSize: requestSize,
			Status:      0, // Status 0 indicates connection failure
		}, nil
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return &types.RequestResult{
			Status:      resp.StatusCode,
			StatusText:  resp.Status,
			Error:       fmt.Sprintf("failed to read response body: %v", err),
			Duration:    duration,
			RequestSize: requestSize,
		}, nil
	}

	// Build response headers map
	headers := make(map[string]string)
	for key, values := range resp.Header {
		headers[key] = strings.Join(values, ", ")
	}

	result := &types.RequestResult{
		Status:       resp.StatusCode,
		StatusText:   resp.Status,
		Headers:      headers,
		Body:         string(bodyBytes),
		Duration:     duration,
		RequestSize:  requestSize,
		ResponseSize: len(bodyBytes),
	}

	return result, nil
}
