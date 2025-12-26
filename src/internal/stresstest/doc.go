/*
Package stresstest provides load testing functionality for HTTP endpoints.

# Overview

The stresstest package implements a concurrent HTTP load testing system with:
  - Configurable worker pools
  - Request rate limiting and ramp-up
  - Real-time metrics collection
  - Response validation (status codes, body patterns)
  - Database persistence of results

# Architecture

The package consists of three main components:

1. Manager (manager.go): Database operations for configs, runs, and metrics
2. Executor (executor.go): Concurrent test execution engine
3. Config (config.go): Configuration and validation

# Executor Design

The Executor uses a worker pool pattern:
  - Fixed number of concurrent workers
  - Shared HTTP client with connection pooling
  - Request channel for work distribution
  - Result channel for metrics collection
  - Context-based cancellation

Worker lifecycle:
  1. Workers signal ready via WaitGroup
  2. Scheduler queues requests with optional ramp-up delays
  3. Workers execute requests and send results
  4. Result collector aggregates metrics
  5. Cleanup on completion or cancellation

# Metrics Collection

Metrics are collected for each request:
  - Duration (ms)
  - Status code
  - Request/response sizes
  - Network errors
  - Validation errors

Statistics calculated:
  - Min/max/average duration
  - Percentiles (P50, P95, P99)
  - Success/error/validation error counts

# Validation

Requests can be validated against:
  - Expected status codes
  - Exact body match
  - Substring contains
  - Regex patterns
  - JSON field values (with regex support)

# Database Schema

SQLite database stores:
  - stress_test_configs: Test configurations
  - stress_test_runs: Test execution records
  - stress_test_metrics: Individual request metrics

# Example Usage

	// Create manager
	manager, err := NewManager("restcli.db")
	if err != nil {
		return err
	}
	defer manager.Close()

	// Create configuration
	config := &Config{
		Name:            "API Load Test",
		RequestFile:     "/path/to/request.http",
		ProfileName:     "production",
		ConcurrentConns: 10,
		TotalRequests:   1000,
		RampUpDurationSec: 5,
	}

	// Execute stress test
	execConfig := &ExecutionConfig{
		Request: parsedRequest,
		Config:  config,
	}

	executor, err := NewExecutor(execConfig, manager)
	if err != nil {
		return err
	}

	executor.Start()
	err = executor.Wait()

	// Get results
	run := executor.GetRun()
	fmt.Printf("Completed %d requests\n", run.TotalRequestsCompleted)
	fmt.Printf("Average duration: %.2fms\n", run.AvgDurationMs)
	fmt.Printf("P95 latency: %dms\n", run.P95DurationMs)

# Thread Safety

All public methods of Manager and Executor are thread-safe.
The Executor uses atomic counters and mutexes for concurrent access.

# Cancellation

Tests can be cancelled via:
  - Duration timeout (TestDurationSec)
  - Context cancellation (Stop/StopWithContext)
  - User interruption

Graceful shutdown ensures:
  - Active workers complete current requests
  - Metrics are flushed to database
  - Resources are cleaned up
*/
package stresstest
