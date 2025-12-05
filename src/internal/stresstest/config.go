package stresstest

import (
	"fmt"
	"time"

	"github.com/studiowebux/restcli/internal/types"
)

// Config represents a stress test configuration
type Config struct {
	ID                   int64
	Name                 string
	RequestFile          string
	ProfileName          string
	ConcurrentConns      int
	TotalRequests        int
	RampUpDurationSec    int
	TestDurationSec      int
	RequestTimeoutSec    int // Timeout for individual requests (default: 10s)
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// Run represents a stress test run record
type Run struct {
	ID                     int64
	ConfigID               *int64
	ConfigName             string
	RequestFile            string
	ProfileName            string
	StartedAt              time.Time
	CompletedAt            *time.Time
	Status                 string // "running", "completed", "cancelled", "failed"
	TotalRequestsSent      int
	TotalRequestsCompleted int
	TotalErrors            int
	AvgDurationMs          float64
	MinDurationMs          int64
	MaxDurationMs          int64
	P50DurationMs          int64
	P95DurationMs          int64
	P99DurationMs          int64
}

// Metric represents a single request metric in a stress test
type Metric struct {
	ID           int64
	RunID        int64
	Timestamp    time.Time
	ElapsedMs    int64
	StatusCode   int
	DurationMs   int64
	RequestSize  int64
	ResponseSize int64
	ErrorMessage string
}

// ExecutionConfig contains the runtime configuration for executing a stress test
type ExecutionConfig struct {
	Request         *types.HttpRequest
	TLSConfig       *types.TLSConfig
	Config          *Config
}

// Validate validates the stress test configuration
func (c *Config) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("config name is required")
	}
	if c.RequestFile == "" {
		return fmt.Errorf("request file is required")
	}
	if c.ConcurrentConns <= 0 {
		return fmt.Errorf("concurrent connections must be greater than 0")
	}
	if c.ConcurrentConns > 1000 {
		return fmt.Errorf("concurrent connections cannot exceed 1000")
	}
	if c.TotalRequests <= 0 {
		return fmt.Errorf("total requests must be greater than 0")
	}
	if c.TotalRequests > 1000000 {
		return fmt.Errorf("total requests cannot exceed 1,000,000")
	}
	if c.RampUpDurationSec < 0 {
		return fmt.Errorf("ramp-up duration cannot be negative")
	}
	if c.TestDurationSec < 0 {
		return fmt.Errorf("test duration cannot be negative")
	}
	return nil
}

// GetRampUpDuration returns the ramp-up duration as time.Duration
func (c *Config) GetRampUpDuration() time.Duration {
	return time.Duration(c.RampUpDurationSec) * time.Second
}

// GetTestDuration returns the test duration as time.Duration
func (c *Config) GetTestDuration() time.Duration {
	if c.TestDurationSec == 0 {
		return 0 // Unlimited
	}
	return time.Duration(c.TestDurationSec) * time.Second
}

// GetRequestTimeout returns the request timeout as time.Duration
func (c *Config) GetRequestTimeout() time.Duration {
	if c.RequestTimeoutSec == 0 {
		return 10 * time.Second // Default 10 seconds
	}
	return time.Duration(c.RequestTimeoutSec) * time.Second
}

// IsRunning returns true if the run is currently in progress
func (r *Run) IsRunning() bool {
	return r.Status == "running"
}

// IsCompleted returns true if the run has finished
func (r *Run) IsCompleted() bool {
	return r.Status == "completed" || r.Status == "cancelled" || r.Status == "failed"
}
