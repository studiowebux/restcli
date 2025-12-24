package tui

import (
	"sync"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/studiowebux/restcli/internal/analytics"
)

// AnalyticsState encapsulates all analytics-related UI state
type AnalyticsState struct {
	mu sync.RWMutex

	// Manager for database operations
	manager *analytics.Manager

	// Analytics data and navigation
	stats []analytics.Stats
	index int

	// Viewports for split view
	listView   viewport.Model
	detailView viewport.Model

	// UI state
	previewVisible bool   // Toggle for showing/hiding stats detail pane
	groupByPath    bool   // Toggle between per-file and normalized-path grouping
	focusedPane    string // "list" or "details" - which pane has focus in split view
}

// NewAnalyticsState creates a new analytics state
func NewAnalyticsState(manager *analytics.Manager) *AnalyticsState {
	return &AnalyticsState{
		manager:        manager,
		stats:          []analytics.Stats{},
		index:          0,
		listView:       viewport.New(80, 20),
		detailView:     viewport.New(80, 20),
		previewVisible: true,
		groupByPath:    false,
		focusedPane:    "list",
	}
}

// GetManager returns the analytics manager
func (s *AnalyticsState) GetManager() *analytics.Manager {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.manager
}

// GetStats returns a copy of the stats slice
func (s *AnalyticsState) GetStats() []analytics.Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]analytics.Stats, len(s.stats))
	copy(result, s.stats)
	return result
}

// SetStats sets the stats slice
func (s *AnalyticsState) SetStats(stats []analytics.Stats) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stats = stats
}

// GetIndex returns the current index
func (s *AnalyticsState) GetIndex() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.index
}

// SetIndex sets the current index
func (s *AnalyticsState) SetIndex(index int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.index = index
}

// Navigate moves the selection by delta
func (s *AnalyticsState) Navigate(delta int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.stats) == 0 {
		return
	}

	s.index += delta

	// Wrap around
	if s.index < 0 {
		s.index = len(s.stats) - 1
	} else if s.index >= len(s.stats) {
		s.index = 0
	}
}

// GetCurrentStats returns the currently selected stats entry
func (s *AnalyticsState) GetCurrentStats() *analytics.Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.stats) == 0 || s.index < 0 || s.index >= len(s.stats) {
		return nil
	}

	return &s.stats[s.index]
}

// GetListView returns a copy of the list viewport
func (s *AnalyticsState) GetListView() viewport.Model {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.listView
}

// SetListView sets the list viewport
func (s *AnalyticsState) SetListView(v viewport.Model) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.listView = v
}

// GetDetailView returns a copy of the detail viewport
func (s *AnalyticsState) GetDetailView() viewport.Model {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.detailView
}

// SetDetailView sets the detail viewport
func (s *AnalyticsState) SetDetailView(v viewport.Model) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.detailView = v
}

// GetPreviewVisible returns the preview visibility state
func (s *AnalyticsState) GetPreviewVisible() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.previewVisible
}

// SetPreviewVisible sets the preview visibility state
func (s *AnalyticsState) SetPreviewVisible(visible bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.previewVisible = visible
}

// TogglePreview toggles the preview visibility
func (s *AnalyticsState) TogglePreview() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.previewVisible = !s.previewVisible
}

// GetGroupByPath returns the groupByPath state
func (s *AnalyticsState) GetGroupByPath() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.groupByPath
}

// SetGroupByPath sets the groupByPath state
func (s *AnalyticsState) SetGroupByPath(groupByPath bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.groupByPath = groupByPath
}

// ToggleGroupByPath toggles the groupByPath state
func (s *AnalyticsState) ToggleGroupByPath() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.groupByPath = !s.groupByPath
}

// GetFocusedPane returns the focused pane
func (s *AnalyticsState) GetFocusedPane() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.focusedPane
}

// SetFocusedPane sets the focused pane
func (s *AnalyticsState) SetFocusedPane(pane string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.focusedPane = pane
}

// ToggleFocus toggles between list and details panes
func (s *AnalyticsState) ToggleFocus() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.focusedPane == "list" {
		s.focusedPane = "details"
	} else {
		s.focusedPane = "list"
	}
}
