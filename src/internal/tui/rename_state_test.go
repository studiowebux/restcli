package tui

import (
	"sync"
	"testing"
)

func TestNewRenameState(t *testing.T) {
	state := NewRenameState()

	if state == nil {
		t.Fatal("NewRenameState returned nil")
	}

	if state.GetInput() != "" {
		t.Errorf("Expected empty input, got %s", state.GetInput())
	}

	if state.GetCursor() != 0 {
		t.Errorf("Expected cursor 0, got %d", state.GetCursor())
	}
}

func TestRenameState_InputOperations(t *testing.T) {
	state := NewRenameState()

	// Test initial state
	if state.GetInput() != "" {
		t.Errorf("Expected empty input, got %s", state.GetInput())
	}

	// Set input
	state.SetInput("test.http")
	if state.GetInput() != "test.http" {
		t.Errorf("Expected 'test.http', got %s", state.GetInput())
	}

	// Update input
	state.SetInput("newfile.http")
	if state.GetInput() != "newfile.http" {
		t.Errorf("Expected 'newfile.http', got %s", state.GetInput())
	}

	// Set empty
	state.SetInput("")
	if state.GetInput() != "" {
		t.Errorf("Expected empty input, got %s", state.GetInput())
	}
}

func TestRenameState_CursorOperations(t *testing.T) {
	state := NewRenameState()

	// Test initial state
	if state.GetCursor() != 0 {
		t.Errorf("Expected cursor 0, got %d", state.GetCursor())
	}

	// Set cursor
	state.SetCursor(5)
	if state.GetCursor() != 5 {
		t.Errorf("Expected cursor 5, got %d", state.GetCursor())
	}

	// Update cursor
	state.SetCursor(10)
	if state.GetCursor() != 10 {
		t.Errorf("Expected cursor 10, got %d", state.GetCursor())
	}

	// Set to 0
	state.SetCursor(0)
	if state.GetCursor() != 0 {
		t.Errorf("Expected cursor 0, got %d", state.GetCursor())
	}
}

func TestRenameState_Initialize(t *testing.T) {
	state := NewRenameState()

	// Set some initial state
	state.SetInput("old")
	state.SetCursor(1)

	// Initialize with new input
	state.Initialize("newfile.http")

	if state.GetInput() != "newfile.http" {
		t.Errorf("Expected 'newfile.http', got %s", state.GetInput())
	}

	// Cursor should be at end of input
	expectedCursor := len("newfile.http")
	if state.GetCursor() != expectedCursor {
		t.Errorf("Expected cursor %d, got %d", expectedCursor, state.GetCursor())
	}

	// Initialize with empty string
	state.Initialize("")
	if state.GetInput() != "" {
		t.Errorf("Expected empty input, got %s", state.GetInput())
	}
	if state.GetCursor() != 0 {
		t.Errorf("Expected cursor 0, got %d", state.GetCursor())
	}
}

func TestRenameState_Reset(t *testing.T) {
	state := NewRenameState()

	// Set various state
	state.SetInput("test.http")
	state.SetCursor(5)

	// Reset
	state.Reset()

	// Verify everything is reset
	if state.GetInput() != "" {
		t.Errorf("Expected empty input after reset, got %s", state.GetInput())
	}
	if state.GetCursor() != 0 {
		t.Errorf("Expected cursor 0 after reset, got %d", state.GetCursor())
	}
}

func TestRenameState_ConcurrentInputAccess(t *testing.T) {
	state := NewRenameState()

	var wg sync.WaitGroup
	iterations := 50

	for i := 0; i < iterations; i++ {
		wg.Add(2)

		go func(val string) {
			defer wg.Done()
			state.SetInput(val)
		}("test")

		go func() {
			defer wg.Done()
			_ = state.GetInput()
		}()
	}

	wg.Wait()
	// If we get here without panic or data race, success
}

func TestRenameState_ConcurrentCursorAccess(t *testing.T) {
	state := NewRenameState()

	var wg sync.WaitGroup
	iterations := 50

	for i := 0; i < iterations; i++ {
		wg.Add(2)

		go func(pos int) {
			defer wg.Done()
			state.SetCursor(pos)
		}(i)

		go func() {
			defer wg.Done()
			_ = state.GetCursor()
		}()
	}

	wg.Wait()
	// If we get here without panic or data race, success
}

func TestRenameState_ConcurrentInitialize(t *testing.T) {
	state := NewRenameState()

	var wg sync.WaitGroup
	iterations := 50

	for i := 0; i < iterations; i++ {
		wg.Add(2)

		go func() {
			defer wg.Done()
			state.Initialize("test.http")
		}()

		go func() {
			defer wg.Done()
			_ = state.GetInput()
			_ = state.GetCursor()
		}()
	}

	wg.Wait()
	// If we get here without panic or data race, success
}

func TestRenameState_ConcurrentReset(t *testing.T) {
	state := NewRenameState()

	var wg sync.WaitGroup
	iterations := 50

	for i := 0; i < iterations; i++ {
		wg.Add(3)

		go func() {
			defer wg.Done()
			state.SetInput("test")
			state.SetCursor(4)
		}()

		go func() {
			defer wg.Done()
			state.Reset()
		}()

		go func() {
			defer wg.Done()
			_ = state.GetInput()
			_ = state.GetCursor()
		}()
	}

	wg.Wait()
	// If we get here without panic or data race, success
}

func TestRenameState_ConcurrentMixedOperations(t *testing.T) {
	state := NewRenameState()

	var wg sync.WaitGroup
	iterations := 50

	for i := 0; i < iterations; i++ {
		wg.Add(5)

		go func() {
			defer wg.Done()
			state.SetInput("file.http")
		}()

		go func() {
			defer wg.Done()
			state.SetCursor(5)
		}()

		go func() {
			defer wg.Done()
			state.Initialize("newfile.http")
		}()

		go func() {
			defer wg.Done()
			state.Reset()
		}()

		go func() {
			defer wg.Done()
			_ = state.GetInput()
			_ = state.GetCursor()
		}()
	}

	wg.Wait()
	// If we get here without panic or data race, success
}
