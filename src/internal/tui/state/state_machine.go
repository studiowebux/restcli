package state

import (
	"fmt"
	"sync"
)

// State represents a UI state in the TUI application.
// Each state defines its own behavior, rendering, and transition rules.
type State interface {
	// Name returns the unique identifier for this state
	Name() string

	// CanTransitionTo checks if transitioning to the given state is allowed
	CanTransitionTo(state State) bool

	// OnEnter is called when entering this state
	OnEnter() error

	// OnExit is called when leaving this state
	OnExit() error
}

// StateMachine manages state transitions with validation and history.
// It ensures only valid state transitions occur and provides thread-safe access.
type StateMachine struct {
	current  State
	previous State
	mu       sync.RWMutex
}

// NewStateMachine creates a new state machine with the given initial state
func NewStateMachine(initial State) *StateMachine {
	return &StateMachine{
		current: initial,
	}
}

// Current returns the current state (thread-safe)
func (sm *StateMachine) Current() State {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.current
}

// Previous returns the previous state (thread-safe)
func (sm *StateMachine) Previous() State {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.previous
}

// Transition attempts to transition to the new state.
// Returns error if the transition is not allowed or if OnEnter/OnExit fails.
func (sm *StateMachine) Transition(to State) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if to == nil {
		return fmt.Errorf("cannot transition to nil state")
	}

	// Check if transition is allowed
	if !sm.current.CanTransitionTo(to) {
		return fmt.Errorf("invalid transition: %s -> %s",
			sm.current.Name(), to.Name())
	}

	// Exit current state
	if err := sm.current.OnExit(); err != nil {
		return fmt.Errorf("failed to exit state %s: %w", sm.current.Name(), err)
	}

	// Save previous state
	sm.previous = sm.current

	// Enter new state
	if err := to.OnEnter(); err != nil {
		// Rollback on failure
		sm.current.OnEnter() // Re-enter previous state
		return fmt.Errorf("failed to enter state %s: %w", to.Name(), err)
	}

	// Update current state
	sm.current = to

	return nil
}

// CanGoBack returns true if there is a previous state to return to
func (sm *StateMachine) CanGoBack() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.previous != nil
}

// GoBack transitions back to the previous state if one exists
func (sm *StateMachine) GoBack() error {
	sm.mu.RLock()
	prev := sm.previous
	sm.mu.RUnlock()

	if prev == nil {
		return fmt.Errorf("no previous state to return to")
	}

	return sm.Transition(prev)
}

// CurrentName returns the name of the current state (convenience method)
func (sm *StateMachine) CurrentName() string {
	return sm.Current().Name()
}
