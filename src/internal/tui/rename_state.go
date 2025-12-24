package tui

import (
	"sync"
)

// RenameState encapsulates all file rename input state
type RenameState struct {
	mu sync.RWMutex

	input  string
	cursor int
}

// NewRenameState creates a new rename state
func NewRenameState() *RenameState {
	return &RenameState{
		input:  "",
		cursor: 0,
	}
}

// GetInput returns the rename input value
func (s *RenameState) GetInput() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.input
}

// SetInput sets the rename input value
func (s *RenameState) SetInput(input string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.input = input
}

// GetCursor returns the cursor position
func (s *RenameState) GetCursor() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cursor
}

// SetCursor sets the cursor position
func (s *RenameState) SetCursor(cursor int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cursor = cursor
}

// Initialize sets the input and cursor for renaming
func (s *RenameState) Initialize(input string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.input = input
	s.cursor = len(input)
}

// Reset resets all rename state
func (s *RenameState) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.input = ""
	s.cursor = 0
}
