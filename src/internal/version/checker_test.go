package version

import "testing"

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		name     string
		latest   string
		current  string
		expected bool
	}{
		{"same version", "0.0.28", "0.0.28", false},
		{"patch upgrade", "0.0.29", "0.0.28", true},
		{"patch downgrade", "0.0.27", "0.0.28", false},
		{"minor upgrade", "0.1.0", "0.0.28", true},
		{"minor downgrade", "0.0.1", "0.1.0", false},
		{"major upgrade", "1.0.0", "0.0.28", true},
		{"major downgrade", "0.0.28", "1.0.0", false},
		{"multi-digit patch", "0.0.100", "0.0.99", true},
		{"multi-digit minor", "0.100.0", "0.99.0", true},
		{"different lengths v1", "1.0", "0.0.28", true},
		{"different lengths v2", "0.0.28", "1.0", false},
		{"dev version ahead", "0.0.29-dev", "0.0.28", true},      // 0.0.29 > 0.0.28
		{"latest ahead", "0.0.30", "0.0.29", true},               // This IS newer
		{"current ahead", "0.0.28", "0.0.29", false},             // Latest is older
		{"way ahead", "1.2.3", "0.0.28", true},
		{"pre-release same base", "0.0.28-alpha", "0.0.28", false}, // Same numeric version
		{"build metadata", "0.0.29+build123", "0.0.28", true},      // 0.0.29 > 0.0.28
		{"both pre-release", "0.0.29-beta", "0.0.29-alpha", false}, // Same numeric version
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNewerVersion(tt.latest, tt.current)
			if result != tt.expected {
				t.Errorf("isNewerVersion(%q, %q) = %v, want %v", tt.latest, tt.current, result, tt.expected)
			}
		})
	}
}
