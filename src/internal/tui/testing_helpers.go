package tui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/studiowebux/restcli/internal/config"
	"github.com/studiowebux/restcli/internal/session"
)

// CreateTestModel creates a Model instance for testing with minimal dependencies
func CreateTestModel(t *testing.T) *Model {
	t.Helper()

	// Create temporary directory for test database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	// Set test database path
	originalDBPath := config.DatabasePath
	config.DatabasePath = dbPath
	t.Cleanup(func() {
		config.DatabasePath = originalDBPath
	})

	// Create minimal session manager
	mgr := session.NewManager()

	// Create model with error handling
	m, err := New(mgr, "test-version")
	if err != nil {
		t.Fatalf("Failed to create test model: %v", err)
	}

	return &m
}

// CreateTestModelWithFiles creates a Model with test HTTP request files
func CreateTestModelWithFiles(t *testing.T, fileContents map[string]string) *Model {
	t.Helper()

	// Create temporary directory structure
	tempDir := t.TempDir()

	// Save original working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Change to temp directory for the test
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(originalWd)
	})

	// Write test files
	for filename, content := range fileContents {
		filePath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write test file %s: %v", filename, err)
		}
	}

	// Set test database path
	dbPath := filepath.Join(tempDir, "test.db")
	originalDBPath := config.DatabasePath
	config.DatabasePath = dbPath
	t.Cleanup(func() {
		config.DatabasePath = originalDBPath
	})

	// Create session manager (will use temp dir as working directory)
	mgr := session.NewManager()

	// Create model
	m, err := New(mgr, "test-version")
	if err != nil {
		t.Fatalf("Failed to create test model with files: %v", err)
	}

	return &m
}

// AssertModelField is a generic helper for checking model field values
func AssertModelField[T comparable](t *testing.T, fieldName string, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %v, want %v", fieldName, got, want)
	}
}

// AssertNoError verifies that an error is nil
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// AssertError verifies that an error occurred
func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Error("Expected error but got nil")
	}
}
