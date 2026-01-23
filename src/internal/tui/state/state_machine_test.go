package state

import (
	"errors"
	"sync"
	"testing"
)

// Mock states for testing
type mockState struct {
	name           string
	allowedStates  []string
	onEnterError   error
	onExitError    error
	enterCallCount int
	exitCallCount  int
	mu             sync.Mutex
}

func (m *mockState) Name() string {
	return m.name
}

func (m *mockState) CanTransitionTo(state State) bool {
	for _, allowed := range m.allowedStates {
		if state.Name() == allowed {
			return true
		}
	}
	return false
}

func (m *mockState) OnEnter() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enterCallCount++
	return m.onEnterError
}

func (m *mockState) OnExit() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.exitCallCount++
	return m.onExitError
}

func (m *mockState) getEnterCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.enterCallCount
}

func (m *mockState) getExitCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.exitCallCount
}

func TestNewStateMachine(t *testing.T) {
	initial := &mockState{name: "initial"}
	sm := NewStateMachine(initial)

	if sm == nil {
		t.Fatal("NewStateMachine returned nil")
	}

	if sm.Current() != initial {
		t.Errorf("Expected current state to be %s, got %s", initial.Name(), sm.Current().Name())
	}

	if sm.Previous() != nil {
		t.Error("Expected previous state to be nil")
	}
}

func TestStateMachine_Current(t *testing.T) {
	state1 := &mockState{name: "state1"}
	sm := NewStateMachine(state1)

	current := sm.Current()
	if current == nil {
		t.Fatal("Current() returned nil")
	}

	if current.Name() != "state1" {
		t.Errorf("Expected current state 'state1', got '%s'", current.Name())
	}
}

func TestStateMachine_CurrentName(t *testing.T) {
	state1 := &mockState{name: "test_state"}
	sm := NewStateMachine(state1)

	if sm.CurrentName() != "test_state" {
		t.Errorf("Expected CurrentName 'test_state', got '%s'", sm.CurrentName())
	}
}

func TestStateMachine_Transition_Success(t *testing.T) {
	state1 := &mockState{
		name:          "state1",
		allowedStates: []string{"state2"},
	}
	state2 := &mockState{
		name:          "state2",
		allowedStates: []string{"state1"},
	}

	sm := NewStateMachine(state1)

	// Transition from state1 to state2
	err := sm.Transition(state2)
	if err != nil {
		t.Fatalf("Transition failed: %v", err)
	}

	// Verify current state
	if sm.Current().Name() != "state2" {
		t.Errorf("Expected current state 'state2', got '%s'", sm.Current().Name())
	}

	// Verify previous state
	if sm.Previous().Name() != "state1" {
		t.Errorf("Expected previous state 'state1', got '%s'", sm.Previous().Name())
	}

	// Verify lifecycle methods were called
	if state1.getExitCallCount() != 1 {
		t.Errorf("Expected OnExit to be called once on state1, got %d", state1.getExitCallCount())
	}

	if state2.getEnterCallCount() != 1 {
		t.Errorf("Expected OnEnter to be called once on state2, got %d", state2.getEnterCallCount())
	}
}

func TestStateMachine_Transition_InvalidTransition(t *testing.T) {
	state1 := &mockState{
		name:          "state1",
		allowedStates: []string{}, // No allowed transitions
	}
	state2 := &mockState{
		name:          "state2",
		allowedStates: []string{},
	}

	sm := NewStateMachine(state1)

	// Try invalid transition
	err := sm.Transition(state2)
	if err == nil {
		t.Fatal("Expected error for invalid transition, got nil")
	}

	// Verify state didn't change
	if sm.Current().Name() != "state1" {
		t.Errorf("State should not have changed, got '%s'", sm.Current().Name())
	}

	// Verify lifecycle methods were not called
	if state1.getExitCallCount() != 0 {
		t.Error("OnExit should not have been called on failed transition")
	}

	if state2.getEnterCallCount() != 0 {
		t.Error("OnEnter should not have been called on failed transition")
	}
}

func TestStateMachine_Transition_NilState(t *testing.T) {
	state1 := &mockState{name: "state1"}
	sm := NewStateMachine(state1)

	err := sm.Transition(nil)
	if err == nil {
		t.Fatal("Expected error when transitioning to nil state")
	}

	if sm.Current().Name() != "state1" {
		t.Error("State should not have changed")
	}
}

func TestStateMachine_Transition_OnExitError(t *testing.T) {
	exitError := errors.New("exit error")
	state1 := &mockState{
		name:          "state1",
		allowedStates: []string{"state2"},
		onExitError:   exitError,
	}
	state2 := &mockState{
		name:          "state2",
		allowedStates: []string{},
	}

	sm := NewStateMachine(state1)

	err := sm.Transition(state2)
	if err == nil {
		t.Fatal("Expected error from OnExit")
	}

	// State should not have changed
	if sm.Current().Name() != "state1" {
		t.Error("State should not have changed when OnExit fails")
	}

	// OnEnter should not have been called
	if state2.getEnterCallCount() != 0 {
		t.Error("OnEnter should not be called when OnExit fails")
	}
}

func TestStateMachine_Transition_OnEnterError(t *testing.T) {
	enterError := errors.New("enter error")
	state1 := &mockState{
		name:          "state1",
		allowedStates: []string{"state2"},
	}
	state2 := &mockState{
		name:          "state2",
		allowedStates: []string{},
		onEnterError:  enterError,
	}

	sm := NewStateMachine(state1)

	err := sm.Transition(state2)
	if err == nil {
		t.Fatal("Expected error from OnEnter")
	}

	// State should not have changed (rollback)
	if sm.Current().Name() != "state1" {
		t.Error("State should have rolled back when OnEnter fails")
	}

	// OnEnter should have been called twice: once on state2 (failed), once on state1 (rollback)
	if state1.getEnterCallCount() != 1 {
		t.Errorf("Expected OnEnter to be called once for rollback, got %d", state1.getEnterCallCount())
	}
}

func TestStateMachine_CanGoBack(t *testing.T) {
	state1 := &mockState{
		name:          "state1",
		allowedStates: []string{"state2"},
	}
	state2 := &mockState{
		name:          "state2",
		allowedStates: []string{"state1"},
	}

	sm := NewStateMachine(state1)

	// Initially no previous state
	if sm.CanGoBack() {
		t.Error("Should not be able to go back initially")
	}

	// Transition to state2
	sm.Transition(state2)

	// Now should be able to go back
	if !sm.CanGoBack() {
		t.Error("Should be able to go back after transition")
	}
}

func TestStateMachine_GoBack_Success(t *testing.T) {
	state1 := &mockState{
		name:          "state1",
		allowedStates: []string{"state2"},
	}
	state2 := &mockState{
		name:          "state2",
		allowedStates: []string{"state1"},
	}

	sm := NewStateMachine(state1)

	// Transition forward
	sm.Transition(state2)

	// Go back
	err := sm.GoBack()
	if err != nil {
		t.Fatalf("GoBack failed: %v", err)
	}

	// Should be back to state1
	if sm.Current().Name() != "state1" {
		t.Errorf("Expected to be back at state1, got %s", sm.Current().Name())
	}
}

func TestStateMachine_GoBack_NoPrevious(t *testing.T) {
	state1 := &mockState{name: "state1"}
	sm := NewStateMachine(state1)

	err := sm.GoBack()
	if err == nil {
		t.Fatal("Expected error when going back with no previous state")
	}
}

func TestStateMachine_ConcurrentAccess(t *testing.T) {
	state1 := &mockState{
		name:          "state1",
		allowedStates: []string{"state2", "state3"},
	}
	state2 := &mockState{
		name:          "state2",
		allowedStates: []string{"state1", "state3"},
	}
	state3 := &mockState{
		name:          "state3",
		allowedStates: []string{"state1", "state2"},
	}

	sm := NewStateMachine(state1)

	// Run concurrent transitions
	var wg sync.WaitGroup
	iterations := 100

	wg.Add(3)

	// Goroutine 1: state1 <-> state2
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			current := sm.Current()
			if current.Name() == "state1" {
				sm.Transition(state2)
			}
		}
	}()

	// Goroutine 2: state1 <-> state3
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			current := sm.Current()
			if current.Name() == "state1" {
				sm.Transition(state3)
			}
		}
	}()

	// Goroutine 3: Read current state
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_ = sm.Current()
			_ = sm.CurrentName()
			_ = sm.CanGoBack()
		}
	}()

	wg.Wait()

	// Test should not panic or deadlock
}

func TestStateMachine_MultipleTransitions(t *testing.T) {
	state1 := &mockState{
		name:          "state1",
		allowedStates: []string{"state2"},
	}
	state2 := &mockState{
		name:          "state2",
		allowedStates: []string{"state3", "state1"},
	}
	state3 := &mockState{
		name:          "state3",
		allowedStates: []string{"state1", "state2"},
	}

	sm := NewStateMachine(state1)

	// state1 -> state2
	if err := sm.Transition(state2); err != nil {
		t.Fatalf("Transition 1 failed: %v", err)
	}

	// state2 -> state3
	if err := sm.Transition(state3); err != nil {
		t.Fatalf("Transition 2 failed: %v", err)
	}

	// Verify current is state3
	if sm.Current().Name() != "state3" {
		t.Errorf("Expected state3, got %s", sm.Current().Name())
	}

	// Verify previous is state2
	if sm.Previous().Name() != "state2" {
		t.Errorf("Expected previous state2, got %s", sm.Previous().Name())
	}

	// Go back to state2
	if err := sm.GoBack(); err != nil {
		t.Fatalf("GoBack failed: %v", err)
	}

	if sm.Current().Name() != "state2" {
		t.Errorf("Expected state2 after GoBack, got %s", sm.Current().Name())
	}
}
