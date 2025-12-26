package tui

import (
	"sync"
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/studiowebux/restcli/internal/types"
)

func TestNewHistoryState(t *testing.T) {
	state := NewHistoryState()

	if state == nil {
		t.Fatal("NewHistoryState returned nil")
	}

	if state.GetIndex() != 0 {
		t.Errorf("Expected index 0, got %d", state.GetIndex())
	}

	if !state.GetPreviewVisible() {
		t.Error("Expected preview visible by default")
	}

	if state.GetSearchActive() {
		t.Error("Expected search inactive by default")
	}

	if state.GetSearchQuery() != "" {
		t.Errorf("Expected empty search query, got '%s'", state.GetSearchQuery())
	}

	if len(state.GetEntries()) != 0 {
		t.Errorf("Expected 0 entries, got %d", len(state.GetEntries()))
	}

	if len(state.GetAllEntries()) != 0 {
		t.Errorf("Expected 0 all entries, got %d", len(state.GetAllEntries()))
	}
}

func TestHistoryState_EntriesOperations(t *testing.T) {
	state := NewHistoryState()

	// Test initial state
	if state.GetCurrentEntry() != nil {
		t.Error("Expected nil entry for empty state")
	}

	// Set entries
	entries := []types.HistoryEntry{
		{Method: "GET", URL: "http://test1.com", Timestamp: "2025-01-01 10:00:00"},
		{Method: "POST", URL: "http://test2.com", Timestamp: "2025-01-01 11:00:00"},
		{Method: "PUT", URL: "http://test3.com", Timestamp: "2025-01-01 12:00:00"},
	}
	state.SetEntries(entries)

	if len(state.GetEntries()) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(state.GetEntries()))
	}

	// Test current entry
	current := state.GetCurrentEntry()
	if current == nil {
		t.Fatal("Expected non-nil current entry")
	}
	if current.Method != "GET" {
		t.Errorf("Expected Method GET, got %s", current.Method)
	}

	// Test navigation
	state.Navigate(1)
	if state.GetIndex() != 1 {
		t.Errorf("Expected index 1, got %d", state.GetIndex())
	}

	current = state.GetCurrentEntry()
	if current == nil {
		t.Fatal("Expected non-nil current entry")
	}
	if current.Method != "POST" {
		t.Errorf("Expected Method POST, got %s", current.Method)
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

func TestHistoryState_AllEntriesOperations(t *testing.T) {
	state := NewHistoryState()

	allEntries := []types.HistoryEntry{
		{Method: "GET", URL: "http://test1.com", Timestamp: "2025-01-01 10:00:00"},
		{Method: "POST", URL: "http://test2.com", Timestamp: "2025-01-01 11:00:00"},
		{Method: "PUT", URL: "http://test3.com", Timestamp: "2025-01-01 12:00:00"},
		{Method: "DELETE", URL: "http://test4.com", Timestamp: "2025-01-01 13:00:00"},
	}
	state.SetAllEntries(allEntries)

	if len(state.GetAllEntries()) != 4 {
		t.Errorf("Expected 4 all entries, got %d", len(state.GetAllEntries()))
	}

	// Set filtered entries
	filteredEntries := []types.HistoryEntry{
		{Method: "GET", URL: "http://test1.com", Timestamp: "2025-01-01 10:00:00"},
		{Method: "PUT", URL: "http://test3.com", Timestamp: "2025-01-01 12:00:00"},
	}
	state.SetEntries(filteredEntries)

	if len(state.GetEntries()) != 2 {
		t.Errorf("Expected 2 filtered entries, got %d", len(state.GetEntries()))
	}

	if len(state.GetAllEntries()) != 4 {
		t.Errorf("Expected 4 all entries (unchanged), got %d", len(state.GetAllEntries()))
	}
}

func TestHistoryState_ViewportOperations(t *testing.T) {
	state := NewHistoryState()

	// Test preview view
	previewView := state.GetPreviewView()
	previewView.Width = 100
	previewView.Height = 50
	state.SetPreviewView(previewView)

	retrieved := state.GetPreviewView()
	if retrieved.Width != 100 {
		t.Errorf("Expected width 100, got %d", retrieved.Width)
	}
	if retrieved.Height != 50 {
		t.Errorf("Expected height 50, got %d", retrieved.Height)
	}
}

func TestHistoryState_PreviewToggle(t *testing.T) {
	state := NewHistoryState()

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

func TestHistoryState_SearchOperations(t *testing.T) {
	state := NewHistoryState()

	// Initial state
	if state.GetSearchActive() {
		t.Error("Expected search inactive initially")
	}
	if state.GetSearchQuery() != "" {
		t.Error("Expected empty search query initially")
	}

	// Activate search
	state.ActivateSearch()
	if !state.GetSearchActive() {
		t.Error("Expected search active after ActivateSearch")
	}

	// Set search query
	state.SetSearchQuery("test query")
	if state.GetSearchQuery() != "test query" {
		t.Errorf("Expected 'test query', got '%s'", state.GetSearchQuery())
	}

	// Clear search
	state.ClearSearch()
	if state.GetSearchActive() {
		t.Error("Expected search inactive after ClearSearch")
	}
	if state.GetSearchQuery() != "" {
		t.Errorf("Expected empty query after ClearSearch, got '%s'", state.GetSearchQuery())
	}

	// Test toggle
	state.ToggleSearch()
	if !state.GetSearchActive() {
		t.Error("Expected search active after ToggleSearch")
	}

	state.ToggleSearch()
	if state.GetSearchActive() {
		t.Error("Expected search inactive after second ToggleSearch")
	}

	// Test direct set
	state.SetSearchActive(true)
	if !state.GetSearchActive() {
		t.Error("Expected search active after SetSearchActive(true)")
	}

	// Test deactivate
	state.DeactivateSearch()
	if state.GetSearchActive() {
		t.Error("Expected search inactive after DeactivateSearch")
	}
}

func TestHistoryState_ConcurrentAccess(t *testing.T) {
	state := NewHistoryState()

	// Create test data
	entries := make([]types.HistoryEntry, 20)
	for i := range entries {
		entries[i] = types.HistoryEntry{Method: "GET", URL: "http://test.com", Timestamp: "2025-01-01 10:00:00"}
	}
	state.SetEntries(entries)
	state.SetAllEntries(entries)

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
			_ = state.GetCurrentEntry()
		}()
	}

	wg.Wait()
	// If we get here without panic or data race, success
}

func TestHistoryState_ConcurrentViewportAccess(t *testing.T) {
	state := NewHistoryState()

	var wg sync.WaitGroup
	iterations := 50

	for i := 0; i < iterations; i++ {
		wg.Add(2)

		go func() {
			defer wg.Done()
			v := state.GetPreviewView()
			v.Width = 100
			state.SetPreviewView(v)
		}()

		go func() {
			defer wg.Done()
			_ = state.GetPreviewView()
		}()
	}

	wg.Wait()
	// If we get here without panic or data race, success
}

func TestHistoryState_ConcurrentSearchOperations(t *testing.T) {
	state := NewHistoryState()

	var wg sync.WaitGroup
	iterations := 50

	for i := 0; i < iterations; i++ {
		wg.Add(6)

		go func() {
			defer wg.Done()
			state.ToggleSearch()
		}()

		go func() {
			defer wg.Done()
			_ = state.GetSearchActive()
		}()

		go func() {
			defer wg.Done()
			state.SetSearchQuery("test")
		}()

		go func() {
			defer wg.Done()
			_ = state.GetSearchQuery()
		}()

		go func() {
			defer wg.Done()
			state.ClearSearch()
		}()

		go func() {
			defer wg.Done()
			state.TogglePreview()
		}()
	}

	wg.Wait()
	// If we get here without panic or data race, success
}

func TestHistoryState_NavigateEmptyEntries(t *testing.T) {
	state := NewHistoryState()

	// Navigate on empty entries should not panic
	state.Navigate(1)
	state.Navigate(-1)

	if state.GetIndex() != 0 {
		t.Errorf("Expected index 0 after navigating empty entries, got %d", state.GetIndex())
	}

	if state.GetCurrentEntry() != nil {
		t.Error("Expected nil current entry for empty state")
	}
}

func TestHistoryState_GetCurrentEntryOutOfBounds(t *testing.T) {
	state := NewHistoryState()

	entries := []types.HistoryEntry{
		{Method: "GET", URL: "http://test.com", Timestamp: "2025-01-01 10:00:00"},
	}
	state.SetEntries(entries)

	// Test negative index
	state.SetIndex(-1)
	if state.GetCurrentEntry() != nil {
		t.Error("Expected nil for negative index")
	}

	// Test out of bounds index
	state.SetIndex(100)
	if state.GetCurrentEntry() != nil {
		t.Error("Expected nil for out of bounds index")
	}
}

func TestHistoryState_EntriesImmutability(t *testing.T) {
	state := NewHistoryState()

	original := []types.HistoryEntry{
		{Method: "GET", URL: "http://test1.com", Timestamp: "2025-01-01 10:00:00"},
		{Method: "POST", URL: "http://test2.com", Timestamp: "2025-01-01 11:00:00"},
	}
	state.SetEntries(original)

	// Get entries and modify
	retrieved := state.GetEntries()
	retrieved[0].Method = "XXX"

	// Original should be unchanged
	current := state.GetEntries()
	if current[0].Method == "XXX" {
		t.Error("Entries were not properly copied - modification affected internal state")
	}
	if current[0].Method != "GET" {
		t.Errorf("Expected Method GET, got %s", current[0].Method)
	}
}

func TestHistoryState_AllEntriesImmutability(t *testing.T) {
	state := NewHistoryState()

	original := []types.HistoryEntry{
		{Method: "GET", URL: "http://test1.com", Timestamp: "2025-01-01 10:00:00"},
		{Method: "POST", URL: "http://test2.com", Timestamp: "2025-01-01 11:00:00"},
	}
	state.SetAllEntries(original)

	// Get all entries and modify
	retrieved := state.GetAllEntries()
	retrieved[0].Method = "XXX"

	// Original should be unchanged
	current := state.GetAllEntries()
	if current[0].Method == "XXX" {
		t.Error("AllEntries were not properly copied - modification affected internal state")
	}
	if current[0].Method != "GET" {
		t.Errorf("Expected Method GET, got %s", current[0].Method)
	}
}

func TestHistoryState_ViewportImmutability(t *testing.T) {
	state := NewHistoryState()

	// Set initial viewport
	v := viewport.New(100, 50)
	state.SetPreviewView(v)

	// Get and modify
	retrieved := state.GetPreviewView()
	retrieved.Width = 999

	// Check that internal state wasn't affected
	current := state.GetPreviewView()
	if current.Width == 999 {
		t.Error("Viewport was not properly copied - modification affected internal state")
	}
	if current.Width != 100 {
		t.Errorf("Expected Width 100, got %d", current.Width)
	}
}
