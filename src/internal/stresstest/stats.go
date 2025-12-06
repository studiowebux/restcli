package stresstest

import (
	"sort"
)

// Stats holds runtime statistics for a stress test
type Stats struct {
	TotalRequests        int
	CompletedRequests    int
	ErrorCount           int // Network errors (timeouts, connection failures)
	ValidationErrorCount int // Validation errors (unexpected status, body mismatch)
	SuccessCount         int
	ActiveWorkers        int     // Current number of workers actively executing requests
	Durations            []int64 // For percentile calculation
	TotalDurationMs      int64
	MinDurationMs        int64
	MaxDurationMs        int64
}

// NewStats creates a new Stats instance
func NewStats() *Stats {
	return &Stats{
		Durations:     make([]int64, 0, 1000),
		MinDurationMs: -1,
		MaxDurationMs: -1,
	}
}

// AddResult adds a request result to the statistics
// isNetworkError: true for connection failures, timeouts, etc.
// isValidationError: true for unexpected status codes, body mismatches, etc.
func (s *Stats) AddResult(durationMs int64, isNetworkError bool, isValidationError bool) {
	s.CompletedRequests++
	s.TotalDurationMs += durationMs
	s.Durations = append(s.Durations, durationMs)

	if isNetworkError {
		s.ErrorCount++
	} else if isValidationError {
		s.ValidationErrorCount++
	} else {
		s.SuccessCount++
	}

	// Update min/max
	if s.MinDurationMs == -1 || durationMs < s.MinDurationMs {
		s.MinDurationMs = durationMs
	}
	if s.MaxDurationMs == -1 || durationMs > s.MaxDurationMs {
		s.MaxDurationMs = durationMs
	}
}

// AvgDurationMs returns the average duration in milliseconds
func (s *Stats) AvgDurationMs() float64 {
	if s.CompletedRequests == 0 {
		return 0
	}
	return float64(s.TotalDurationMs) / float64(s.CompletedRequests)
}

// Min returns the minimum duration, or 0 if no results
func (s *Stats) Min() int64 {
	if s.MinDurationMs == -1 {
		return 0
	}
	return s.MinDurationMs
}

// Max returns the maximum duration, or 0 if no results
func (s *Stats) Max() int64 {
	if s.MaxDurationMs == -1 {
		return 0
	}
	return s.MaxDurationMs
}

// Percentile calculates the percentile value (p should be between 0 and 100)
func (s *Stats) Percentile(p float64) int64 {
	if len(s.Durations) == 0 {
		return 0
	}

	// Make a copy and sort
	sorted := make([]int64, len(s.Durations))
	copy(sorted, s.Durations)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	// Calculate index
	index := (p / 100.0) * float64(len(sorted)-1)
	lower := int(index)
	upper := lower + 1

	if upper >= len(sorted) {
		return sorted[len(sorted)-1]
	}

	// Linear interpolation between lower and upper
	weight := index - float64(lower)
	return int64(float64(sorted[lower])*(1-weight) + float64(sorted[upper])*weight)
}

// P50 returns the 50th percentile (median)
func (s *Stats) P50() int64 {
	return s.Percentile(50)
}

// P95 returns the 95th percentile
func (s *Stats) P95() int64 {
	return s.Percentile(95)
}

// P99 returns the 99th percentile
func (s *Stats) P99() int64 {
	return s.Percentile(99)
}

// SuccessRate returns the success rate as a percentage
func (s *Stats) SuccessRate() float64 {
	if s.CompletedRequests == 0 {
		return 0
	}
	return float64(s.SuccessCount) / float64(s.CompletedRequests) * 100
}

// ErrorRate returns the network error rate as a percentage
func (s *Stats) ErrorRate() float64 {
	if s.CompletedRequests == 0 {
		return 0
	}
	return float64(s.ErrorCount) / float64(s.CompletedRequests) * 100
}

// ValidationErrorRate returns the validation error rate as a percentage
func (s *Stats) ValidationErrorRate() float64 {
	if s.CompletedRequests == 0 {
		return 0
	}
	return float64(s.ValidationErrorCount) / float64(s.CompletedRequests) * 100
}

// Progress returns the completion progress as a percentage
func (s *Stats) Progress() float64 {
	if s.TotalRequests == 0 {
		return 0
	}
	return float64(s.CompletedRequests) / float64(s.TotalRequests) * 100
}
