package tui

import (
	"sync"
	"testing"
)

func TestNewProfileEditState(t *testing.T) {
	state := NewProfileEditState()

	if state == nil {
		t.Fatal("NewProfileEditState returned nil")
	}

	if state.GetField() != 0 {
		t.Errorf("Expected field 0, got %d", state.GetField())
	}

	if state.GetName() != "" {
		t.Errorf("Expected empty name, got %s", state.GetName())
	}

	if state.GetHistoryEnabled() != nil {
		t.Error("Expected nil historyEnabled")
	}
}

func TestProfileEditState_FieldOperations(t *testing.T) {
	state := NewProfileEditState()

	// Test initial state
	if state.GetField() != 0 {
		t.Errorf("Expected field 0, got %d", state.GetField())
	}

	// Set field
	state.SetField(3)
	if state.GetField() != 3 {
		t.Errorf("Expected field 3, got %d", state.GetField())
	}

	// Navigate forward
	state.Navigate(1, 6)
	if state.GetField() != 4 {
		t.Errorf("Expected field 4, got %d", state.GetField())
	}

	// Navigate backward
	state.Navigate(-2, 6)
	if state.GetField() != 2 {
		t.Errorf("Expected field 2, got %d", state.GetField())
	}

	// Navigate with wrap around (forward)
	state.SetField(5)
	state.Navigate(1, 6)
	if state.GetField() != 0 {
		t.Errorf("Expected field 0 (wrapped), got %d", state.GetField())
	}

	// Navigate with wrap around (backward)
	state.SetField(0)
	state.Navigate(-1, 6)
	if state.GetField() != 5 {
		t.Errorf("Expected field 5 (wrapped), got %d", state.GetField())
	}
}

func TestProfileEditState_NameOperations(t *testing.T) {
	state := NewProfileEditState()

	// Test initial state
	if state.GetName() != "" {
		t.Errorf("Expected empty name, got %s", state.GetName())
	}

	// Set name
	state.SetName("production")
	if state.GetName() != "production" {
		t.Errorf("Expected 'production', got %s", state.GetName())
	}

	// Test cursor position
	if state.GetNamePos() != 0 {
		t.Errorf("Expected namePos 0, got %d", state.GetNamePos())
	}

	state.SetNamePos(5)
	if state.GetNamePos() != 5 {
		t.Errorf("Expected namePos 5, got %d", state.GetNamePos())
	}
}

func TestProfileEditState_WorkdirOperations(t *testing.T) {
	state := NewProfileEditState()

	state.SetWorkdir("/home/user/project")
	if state.GetWorkdir() != "/home/user/project" {
		t.Errorf("Expected '/home/user/project', got %s", state.GetWorkdir())
	}

	state.SetWorkdirPos(10)
	if state.GetWorkdirPos() != 10 {
		t.Errorf("Expected workdirPos 10, got %d", state.GetWorkdirPos())
	}
}

func TestProfileEditState_EditorOperations(t *testing.T) {
	state := NewProfileEditState()

	state.SetEditor("vim")
	if state.GetEditor() != "vim" {
		t.Errorf("Expected 'vim', got %s", state.GetEditor())
	}

	state.SetEditorPos(3)
	if state.GetEditorPos() != 3 {
		t.Errorf("Expected editorPos 3, got %d", state.GetEditorPos())
	}
}

func TestProfileEditState_OutputOperations(t *testing.T) {
	state := NewProfileEditState()

	state.SetOutput("/tmp/output.json")
	if state.GetOutput() != "/tmp/output.json" {
		t.Errorf("Expected '/tmp/output.json', got %s", state.GetOutput())
	}

	state.SetOutputPos(15)
	if state.GetOutputPos() != 15 {
		t.Errorf("Expected outputPos 15, got %d", state.GetOutputPos())
	}
}

func TestProfileEditState_BooleanOperations(t *testing.T) {
	state := NewProfileEditState()

	// Test nil initial state
	if state.GetHistoryEnabled() != nil {
		t.Error("Expected nil historyEnabled initially")
	}
	if state.GetAnalyticsEnabled() != nil {
		t.Error("Expected nil analyticsEnabled initially")
	}

	// Set to true
	trueVal := true
	state.SetHistoryEnabled(&trueVal)
	if state.GetHistoryEnabled() == nil || !*state.GetHistoryEnabled() {
		t.Error("Expected historyEnabled to be true")
	}

	// Set to false
	falseVal := false
	state.SetAnalyticsEnabled(&falseVal)
	if state.GetAnalyticsEnabled() == nil || *state.GetAnalyticsEnabled() {
		t.Error("Expected analyticsEnabled to be false")
	}

	// Set back to nil
	state.SetHistoryEnabled(nil)
	if state.GetHistoryEnabled() != nil {
		t.Error("Expected historyEnabled to be nil")
	}
}

func TestProfileEditState_LoadFromProfile(t *testing.T) {
	state := NewProfileEditState()

	trueVal := true
	falseVal := false

	state.LoadFromProfile(
		"staging",
		"/var/www",
		"nano",
		"/tmp/out.txt",
		&trueVal,
		&falseVal,
	)

	if state.GetName() != "staging" {
		t.Errorf("Expected 'staging', got %s", state.GetName())
	}
	if state.GetWorkdir() != "/var/www" {
		t.Errorf("Expected '/var/www', got %s", state.GetWorkdir())
	}
	if state.GetEditor() != "nano" {
		t.Errorf("Expected 'nano', got %s", state.GetEditor())
	}
	if state.GetOutput() != "/tmp/out.txt" {
		t.Errorf("Expected '/tmp/out.txt', got %s", state.GetOutput())
	}

	if state.GetHistoryEnabled() == nil || !*state.GetHistoryEnabled() {
		t.Error("Expected historyEnabled to be true")
	}
	if state.GetAnalyticsEnabled() == nil || *state.GetAnalyticsEnabled() {
		t.Error("Expected analyticsEnabled to be false")
	}

	// Check cursor positions are set to end of strings
	if state.GetNamePos() != len("staging") {
		t.Errorf("Expected namePos %d, got %d", len("staging"), state.GetNamePos())
	}
	if state.GetWorkdirPos() != len("/var/www") {
		t.Errorf("Expected workdirPos %d, got %d", len("/var/www"), state.GetWorkdirPos())
	}
}

func TestProfileEditState_Reset(t *testing.T) {
	state := NewProfileEditState()

	// Set various state
	trueVal := true
	state.SetField(3)
	state.SetName("test")
	state.SetWorkdir("/test")
	state.SetEditor("emacs")
	state.SetOutput("/out")
	state.SetHistoryEnabled(&trueVal)
	state.SetAnalyticsEnabled(&trueVal)
	state.SetNamePos(4)
	state.SetWorkdirPos(5)
	state.SetEditorPos(5)
	state.SetOutputPos(4)

	// Reset
	state.Reset()

	// Verify everything is reset
	if state.GetField() != 0 {
		t.Errorf("Expected field 0 after reset, got %d", state.GetField())
	}
	if state.GetName() != "" {
		t.Errorf("Expected empty name after reset, got %s", state.GetName())
	}
	if state.GetWorkdir() != "" {
		t.Errorf("Expected empty workdir after reset, got %s", state.GetWorkdir())
	}
	if state.GetEditor() != "" {
		t.Errorf("Expected empty editor after reset, got %s", state.GetEditor())
	}
	if state.GetOutput() != "" {
		t.Errorf("Expected empty output after reset, got %s", state.GetOutput())
	}
	if state.GetHistoryEnabled() != nil {
		t.Error("Expected nil historyEnabled after reset")
	}
	if state.GetAnalyticsEnabled() != nil {
		t.Error("Expected nil analyticsEnabled after reset")
	}
	if state.GetNamePos() != 0 {
		t.Errorf("Expected namePos 0 after reset, got %d", state.GetNamePos())
	}
}

func TestProfileEditState_ConcurrentFieldAccess(t *testing.T) {
	state := NewProfileEditState()

	var wg sync.WaitGroup
	iterations := 50

	for i := 0; i < iterations; i++ {
		wg.Add(3)

		go func(idx int) {
			defer wg.Done()
			state.SetField(idx % 6)
		}(i)

		go func() {
			defer wg.Done()
			state.Navigate(1, 6)
		}()

		go func() {
			defer wg.Done()
			_ = state.GetField()
		}()
	}

	wg.Wait()
	// If we get here without panic or data race, success
}

func TestProfileEditState_ConcurrentStringAccess(t *testing.T) {
	state := NewProfileEditState()

	var wg sync.WaitGroup
	iterations := 50

	for i := 0; i < iterations; i++ {
		wg.Add(8)

		go func() {
			defer wg.Done()
			state.SetName("name")
		}()

		go func() {
			defer wg.Done()
			_ = state.GetName()
		}()

		go func() {
			defer wg.Done()
			state.SetWorkdir("/path")
		}()

		go func() {
			defer wg.Done()
			_ = state.GetWorkdir()
		}()

		go func() {
			defer wg.Done()
			state.SetEditor("vim")
		}()

		go func() {
			defer wg.Done()
			_ = state.GetEditor()
		}()

		go func() {
			defer wg.Done()
			state.SetOutput("/out")
		}()

		go func() {
			defer wg.Done()
			_ = state.GetOutput()
		}()
	}

	wg.Wait()
	// If we get here without panic or data race, success
}

func TestProfileEditState_ConcurrentCursorAccess(t *testing.T) {
	state := NewProfileEditState()

	var wg sync.WaitGroup
	iterations := 50

	for i := 0; i < iterations; i++ {
		wg.Add(8)

		go func(pos int) {
			defer wg.Done()
			state.SetNamePos(pos)
		}(i)

		go func() {
			defer wg.Done()
			_ = state.GetNamePos()
		}()

		go func(pos int) {
			defer wg.Done()
			state.SetWorkdirPos(pos)
		}(i)

		go func() {
			defer wg.Done()
			_ = state.GetWorkdirPos()
		}()

		go func(pos int) {
			defer wg.Done()
			state.SetEditorPos(pos)
		}(i)

		go func() {
			defer wg.Done()
			_ = state.GetEditorPos()
		}()

		go func(pos int) {
			defer wg.Done()
			state.SetOutputPos(pos)
		}(i)

		go func() {
			defer wg.Done()
			_ = state.GetOutputPos()
		}()
	}

	wg.Wait()
	// If we get here without panic or data race, success
}

func TestProfileEditState_ConcurrentLoadAndReset(t *testing.T) {
	state := NewProfileEditState()

	var wg sync.WaitGroup
	iterations := 50

	trueVal := true

	for i := 0; i < iterations; i++ {
		wg.Add(3)

		go func() {
			defer wg.Done()
			state.LoadFromProfile("test", "/path", "vim", "/out", &trueVal, &trueVal)
		}()

		go func() {
			defer wg.Done()
			state.Reset()
		}()

		go func() {
			defer wg.Done()
			_ = state.GetName()
			_ = state.GetWorkdir()
			_ = state.GetEditor()
			_ = state.GetOutput()
		}()
	}

	wg.Wait()
	// If we get here without panic or data race, success
}
