package tui

import (
	"sync"
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/studiowebux/restcli/internal/analytics"
)

func TestNewAnalyticsState(t *testing.T) {
	state := NewAnalyticsState(nil)

	if state == nil {
		t.Fatal("NewAnalyticsState returned nil")
	}

	if state.GetIndex() != 0 {
		t.Errorf("Expected index 0, got %d", state.GetIndex())
	}

	if !state.GetPreviewVisible() {
		t.Error("Expected preview visible by default")
	}

	if state.GetGroupByPath() {
		t.Error("Expected groupByPath false by default")
	}

	if state.GetFocusedPane() != "list" {
		t.Errorf("Expected focused pane 'list', got '%s'", state.GetFocusedPane())
	}

	if len(state.GetStats()) != 0 {
		t.Errorf("Expected 0 stats, got %d", len(state.GetStats()))
	}
}

func TestAnalyticsState_StatsOperations(t *testing.T) {
	state := NewAnalyticsState(nil)

	// Test initial state
	if state.GetCurrentStats() != nil {
		t.Error("Expected nil stats for empty state")
	}

	// Set stats
	stats := []analytics.Stats{
		{FilePath: "test1.http", TotalCalls: 10},
		{FilePath: "test2.http", TotalCalls: 20},
		{FilePath: "test3.http", TotalCalls: 30},
	}
	state.SetStats(stats)

	if len(state.GetStats()) != 3 {
		t.Errorf("Expected 3 stats, got %d", len(state.GetStats()))
	}

	// Test current stats
	current := state.GetCurrentStats()
	if current == nil {
		t.Fatal("Expected non-nil current stats")
	}
	if current.FilePath != "test1.http" {
		t.Errorf("Expected test1.http, got %s", current.FilePath)
	}

	// Test navigation
	state.Navigate(1)
	if state.GetIndex() != 1 {
		t.Errorf("Expected index 1, got %d", state.GetIndex())
	}

	current = state.GetCurrentStats()
	if current == nil {
		t.Fatal("Expected non-nil current stats")
	}
	if current.FilePath != "test2.http" {
		t.Errorf("Expected test2.http, got %s", current.FilePath)
	}

	// Test wrap around forward
	state.SetIndex(2)
	state.Navigate(1)
	if state.GetIndex() != 0 {
		t.Errorf("Expected index 0 (wrap), got %d", state.GetIndex())
	}

	// Test wrap around backward
	state.Navigate(-1)
	if state.GetIndex() != 2 {
		t.Errorf("Expected index 2 (wrap), got %d", state.GetIndex())
	}
}

func TestAnalyticsState_ViewportOperations(t *testing.T) {
	state := NewAnalyticsState(nil)

	// Test list view
	listView := state.GetListView()
	listView.Width = 100
	listView.Height = 50
	state.SetListView(listView)

	retrieved := state.GetListView()
	if retrieved.Width != 100 {
		t.Errorf("Expected width 100, got %d", retrieved.Width)
	}
	if retrieved.Height != 50 {
		t.Errorf("Expected height 50, got %d", retrieved.Height)
	}

	// Test detail view
	detailView := state.GetDetailView()
	detailView.Width = 200
	detailView.Height = 100
	state.SetDetailView(detailView)

	retrieved = state.GetDetailView()
	if retrieved.Width != 200 {
		t.Errorf("Expected width 200, got %d", retrieved.Width)
	}
	if retrieved.Height != 100 {
		t.Errorf("Expected height 100, got %d", retrieved.Height)
	}
}

func TestAnalyticsState_PreviewToggle(t *testing.T) {
	state := NewAnalyticsState(nil)

	// Initial state
	if !state.GetPreviewVisible() {
		t.Error("Expected preview visible initially")
	}

	// Toggle off
	state.TogglePreview()
	if state.GetPreviewVisible() {
		t.Error("Expected preview hidden after toggle")
	}

	// Toggle on
	state.TogglePreview()
	if !state.GetPreviewVisible() {
		t.Error("Expected preview visible after second toggle")
	}

	// Test direct set
	state.SetPreviewVisible(false)
	if state.GetPreviewVisible() {
		t.Error("Expected preview hidden after SetPreviewVisible(false)")
	}
}

func TestAnalyticsState_GroupByPathToggle(t *testing.T) {
	state := NewAnalyticsState(nil)

	// Initial state
	if state.GetGroupByPath() {
		t.Error("Expected groupByPath false initially")
	}

	// Toggle on
	state.ToggleGroupByPath()
	if !state.GetGroupByPath() {
		t.Error("Expected groupByPath true after toggle")
	}

	// Toggle off
	state.ToggleGroupByPath()
	if state.GetGroupByPath() {
		t.Error("Expected groupByPath false after second toggle")
	}

	// Test direct set
	state.SetGroupByPath(true)
	if !state.GetGroupByPath() {
		t.Error("Expected groupByPath true after SetGroupByPath(true)")
	}
}

func TestAnalyticsState_FocusToggle(t *testing.T) {
	state := NewAnalyticsState(nil)

	// Initial state
	if state.GetFocusedPane() != "list" {
		t.Errorf("Expected focused pane 'list', got '%s'", state.GetFocusedPane())
	}

	// Toggle to details
	state.ToggleFocus()
	if state.GetFocusedPane() != "details" {
		t.Errorf("Expected focused pane 'details', got '%s'", state.GetFocusedPane())
	}

	// Toggle back to list
	state.ToggleFocus()
	if state.GetFocusedPane() != "list" {
		t.Errorf("Expected focused pane 'list', got '%s'", state.GetFocusedPane())
	}

	// Test direct set
	state.SetFocusedPane("details")
	if state.GetFocusedPane() != "details" {
		t.Errorf("Expected focused pane 'details', got '%s'", state.GetFocusedPane())
	}
}

func TestAnalyticsState_ConcurrentAccess(t *testing.T) {
	state := NewAnalyticsState(nil)

	// Create test data
	stats := make([]analytics.Stats, 20)
	for i := range stats {
		stats[i] = analytics.Stats{FilePath: "test.http", TotalCalls: i}
	}
	state.SetStats(stats)

	var wg sync.WaitGroup
	iterations := 50

	// Concurrent navigation
	for i := 0; i < iterations; i++ {
		wg.Add(3)

		go func() {
			defer wg.Done()
			state.Navigate(1)
		}()

		go func() {
			defer wg.Done()
			state.Navigate(-1)
		}()

		go func() {
			defer wg.Done()
			_ = state.GetCurrentStats()
		}()
	}

	wg.Wait()
	// If we get here without panic or data race, success
}

func TestAnalyticsState_ConcurrentViewportAccess(t *testing.T) {
	state := NewAnalyticsState(nil)

	var wg sync.WaitGroup
	iterations := 50

	for i := 0; i < iterations; i++ {
		wg.Add(4)

		go func() {
			defer wg.Done()
			v := state.GetListView()
			v.Width = 100
			state.SetListView(v)
		}()

		go func() {
			defer wg.Done()
			_ = state.GetListView()
		}()

		go func() {
			defer wg.Done()
			v := state.GetDetailView()
			v.Height = 50
			state.SetDetailView(v)
		}()

		go func() {
			defer wg.Done()
			_ = state.GetDetailView()
		}()
	}

	wg.Wait()
	// If we get here without panic or data race, success
}

func TestAnalyticsState_ConcurrentToggles(t *testing.T) {
	state := NewAnalyticsState(nil)

	var wg sync.WaitGroup
	iterations := 50

	for i := 0; i < iterations; i++ {
		wg.Add(6)

		go func() {
			defer wg.Done()
			state.TogglePreview()
		}()

		go func() {
			defer wg.Done()
			_ = state.GetPreviewVisible()
		}()

		go func() {
			defer wg.Done()
			state.ToggleGroupByPath()
		}()

		go func() {
			defer wg.Done()
			_ = state.GetGroupByPath()
		}()

		go func() {
			defer wg.Done()
			state.ToggleFocus()
		}()

		go func() {
			defer wg.Done()
			_ = state.GetFocusedPane()
		}()
	}

	wg.Wait()
	// If we get here without panic or data race, success
}

func TestAnalyticsState_NavigateEmptyStats(t *testing.T) {
	state := NewAnalyticsState(nil)

	// Navigate on empty stats should not panic
	state.Navigate(1)
	state.Navigate(-1)

	if state.GetIndex() != 0 {
		t.Errorf("Expected index 0 after navigating empty stats, got %d", state.GetIndex())
	}

	if state.GetCurrentStats() != nil {
		t.Error("Expected nil current stats for empty state")
	}
}

func TestAnalyticsState_GetCurrentStatsOutOfBounds(t *testing.T) {
	state := NewAnalyticsState(nil)

	stats := []analytics.Stats{
		{FilePath: "test.http", TotalCalls: 10},
	}
	state.SetStats(stats)

	// Test negative index
	state.SetIndex(-1)
	if state.GetCurrentStats() != nil {
		t.Error("Expected nil for negative index")
	}

	// Test out of bounds index
	state.SetIndex(100)
	if state.GetCurrentStats() != nil {
		t.Error("Expected nil for out of bounds index")
	}
}

func TestAnalyticsState_StatsImmutability(t *testing.T) {
	state := NewAnalyticsState(nil)

	original := []analytics.Stats{
		{FilePath: "test1.http", TotalCalls: 10},
		{FilePath: "test2.http", TotalCalls: 20},
	}
	state.SetStats(original)

	// Get stats and modify
	retrieved := state.GetStats()
	retrieved[0].TotalCalls = 999

	// Original should be unchanged
	current := state.GetStats()
	if current[0].TotalCalls == 999 {
		t.Error("Stats were not properly copied - modification affected internal state")
	}
	if current[0].TotalCalls != 10 {
		t.Errorf("Expected TotalCalls 10, got %d", current[0].TotalCalls)
	}
}

func TestAnalyticsState_ViewportImmutability(t *testing.T) {
	state := NewAnalyticsState(nil)

	// Set initial viewport
	v := viewport.New(100, 50)
	state.SetListView(v)

	// Get and modify
	retrieved := state.GetListView()
	retrieved.Width = 999

	// Check that internal state wasn't affected
	current := state.GetListView()
	if current.Width == 999 {
		t.Error("Viewport was not properly copied - modification affected internal state")
	}
	if current.Width != 100 {
		t.Errorf("Expected Width 100, got %d", current.Width)
	}
}
