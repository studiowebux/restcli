package tui

import (
	"sync"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/studiowebux/restcli/internal/types"
)

// HistoryState encapsulates all history-related UI state
type HistoryState struct {
	mu sync.RWMutex

	// History data and navigation
	entries    []types.HistoryEntry
	allEntries []types.HistoryEntry // Unfiltered entries for search
	index      int

	// Viewport for preview pane
	previewView viewport.Model

	// UI state
	previewVisible bool   // Toggle for showing/hiding response preview pane
	searchActive   bool   // True when search input is active
	searchQuery    string // Search query for filtering history
}

// NewHistoryState creates a new history state
func NewHistoryState() *HistoryState {
	return &HistoryState{
		entries:        []types.HistoryEntry{},
		allEntries:     []types.HistoryEntry{},
		index:          0,
		previewView:    viewport.New(80, 20),
		previewVisible: true,
		searchActive:   false,
		searchQuery:    "",
	}
}

// GetEntries returns a copy of the entries slice
func (s *HistoryState) GetEntries() []types.HistoryEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]types.HistoryEntry, len(s.entries))
	copy(result, s.entries)
	return result
}

// SetEntries sets the entries slice
func (s *HistoryState) SetEntries(entries []types.HistoryEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = entries
}

// GetAllEntries returns a copy of the allEntries slice
func (s *HistoryState) GetAllEntries() []types.HistoryEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]types.HistoryEntry, len(s.allEntries))
	copy(result, s.allEntries)
	return result
}

// SetAllEntries sets the allEntries slice
func (s *HistoryState) SetAllEntries(entries []types.HistoryEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.allEntries = entries
}

// GetIndex returns the current index
func (s *HistoryState) GetIndex() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.index
}

// SetIndex sets the current index
func (s *HistoryState) SetIndex(index int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.index = index
}

// Navigate moves the selection by delta
func (s *HistoryState) Navigate(delta int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.entries) == 0 {
		return
	}

	s.index += delta

	// Wrap around
	if s.index < 0 {
		s.index = len(s.entries) - 1
	} else if s.index >= len(s.entries) {
		s.index = 0
	}
}

// GetCurrentEntry returns the currently selected history entry
func (s *HistoryState) GetCurrentEntry() *types.HistoryEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.entries) == 0 || s.index < 0 || s.index >= len(s.entries) {
		return nil
	}

	return &s.entries[s.index]
}

// GetPreviewView returns a copy of the preview viewport
func (s *HistoryState) GetPreviewView() viewport.Model {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.previewView
}

// SetPreviewView sets the preview viewport
func (s *HistoryState) SetPreviewView(v viewport.Model) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.previewView = v
}

// GetPreviewVisible returns the preview visibility state
func (s *HistoryState) GetPreviewVisible() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.previewVisible
}

// SetPreviewVisible sets the preview visibility state
func (s *HistoryState) SetPreviewVisible(visible bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.previewVisible = visible
}

// TogglePreview toggles the preview visibility
func (s *HistoryState) TogglePreview() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.previewVisible = !s.previewVisible
}

// GetSearchActive returns the search active state
func (s *HistoryState) GetSearchActive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.searchActive
}

// SetSearchActive sets the search active state
func (s *HistoryState) SetSearchActive(active bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.searchActive = active
}

// ActivateSearch activates the search mode
func (s *HistoryState) ActivateSearch() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.searchActive = true
}

// DeactivateSearch deactivates the search mode
func (s *HistoryState) DeactivateSearch() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.searchActive = false
}

// ToggleSearch toggles the search active state
func (s *HistoryState) ToggleSearch() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.searchActive = !s.searchActive
}

// GetSearchQuery returns the search query
func (s *HistoryState) GetSearchQuery() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.searchQuery
}

// SetSearchQuery sets the search query
func (s *HistoryState) SetSearchQuery(query string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.searchQuery = query
}

// ClearSearch clears the search query and deactivates search
func (s *HistoryState) ClearSearch() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.searchQuery = ""
	s.searchActive = false
}
