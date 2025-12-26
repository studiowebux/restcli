package tui

import (
	"sync"
	"testing"
)

func TestNewDocumentationState(t *testing.T) {
	state := NewDocumentationState()

	if state == nil {
		t.Fatal("NewDocumentationState returned nil")
	}

	if state.GetSelectedIdx() != 0 {
		t.Errorf("Expected selectedIdx 0, got %d", state.GetSelectedIdx())
	}

	if state.GetItemCount() != 0 {
		t.Errorf("Expected itemCount 0, got %d", state.GetItemCount())
	}

	if state.GetCollapsed(0) {
		t.Error("Expected collapsed[0] to be false by default")
	}
}

func TestDocumentationState_CollapsedOperations(t *testing.T) {
	state := NewDocumentationState()

	// Test initial state
	if state.GetCollapsed(0) {
		t.Error("Expected collapsed[0] to be false initially")
	}

	// Set collapsed
	state.SetCollapsed(0, true)
	if !state.GetCollapsed(0) {
		t.Error("Expected collapsed[0] to be true after SetCollapsed(0, true)")
	}

	// Toggle collapsed
	state.ToggleCollapsed(0)
	if state.GetCollapsed(0) {
		t.Error("Expected collapsed[0] to be false after toggle")
	}

	state.ToggleCollapsed(0)
	if !state.GetCollapsed(0) {
		t.Error("Expected collapsed[0] to be true after second toggle")
	}

	// Set multiple items
	state.SetCollapsed(1, true)
	state.SetCollapsed(2, false)
	state.SetCollapsed(3, true)

	if !state.GetCollapsed(0) {
		t.Error("Expected collapsed[0] to still be true")
	}
	if !state.GetCollapsed(1) {
		t.Error("Expected collapsed[1] to be true")
	}
	if state.GetCollapsed(2) {
		t.Error("Expected collapsed[2] to be false")
	}
	if !state.GetCollapsed(3) {
		t.Error("Expected collapsed[3] to be true")
	}

	// Clear all
	state.ClearCollapsed()
	if state.GetCollapsed(0) {
		t.Error("Expected collapsed[0] to be false after clear")
	}
	if state.GetCollapsed(1) {
		t.Error("Expected collapsed[1] to be false after clear")
	}
}

func TestDocumentationState_NavigationOperations(t *testing.T) {
	state := NewDocumentationState()

	// Test initial state
	if state.GetSelectedIdx() != 0 {
		t.Errorf("Expected selectedIdx 0, got %d", state.GetSelectedIdx())
	}

	// Set selected index
	state.SetSelectedIdx(5)
	if state.GetSelectedIdx() != 5 {
		t.Errorf("Expected selectedIdx 5, got %d", state.GetSelectedIdx())
	}

	// Navigate forward
	state.Navigate(3, 20)
	if state.GetSelectedIdx() != 8 {
		t.Errorf("Expected selectedIdx 8, got %d", state.GetSelectedIdx())
	}

	// Navigate backward
	state.Navigate(-2, 20)
	if state.GetSelectedIdx() != 6 {
		t.Errorf("Expected selectedIdx 6, got %d", state.GetSelectedIdx())
	}

	// Navigate beyond max (should clamp)
	state.Navigate(20, 10)
	if state.GetSelectedIdx() != 9 {
		t.Errorf("Expected selectedIdx 9 (clamped to max-1), got %d", state.GetSelectedIdx())
	}

	// Navigate below zero (should clamp)
	state.Navigate(-20, 10)
	if state.GetSelectedIdx() != 0 {
		t.Errorf("Expected selectedIdx 0 (clamped to min), got %d", state.GetSelectedIdx())
	}
}

func TestDocumentationState_ItemCountOperations(t *testing.T) {
	state := NewDocumentationState()

	// Test initial state
	if state.GetItemCount() != 0 {
		t.Errorf("Expected itemCount 0, got %d", state.GetItemCount())
	}

	// Set item count
	state.SetItemCount(42)
	if state.GetItemCount() != 42 {
		t.Errorf("Expected itemCount 42, got %d", state.GetItemCount())
	}

	// Update item count
	state.SetItemCount(100)
	if state.GetItemCount() != 100 {
		t.Errorf("Expected itemCount 100, got %d", state.GetItemCount())
	}
}

func TestDocumentationState_FieldTreeCacheOperations(t *testing.T) {
	state := NewDocumentationState()

	// Test initial state
	if cache := state.GetFieldTreeCache(0); cache != nil {
		t.Error("Expected nil cache for non-existent index")
	}

	// Set field tree cache
	tree := []DocField{
		{Name: "field1", Type: "string"},
		{Name: "field2", Type: "int"},
	}
	state.SetFieldTreeCache(0, tree)

	// Get field tree cache
	cached := state.GetFieldTreeCache(0)
	if cached == nil {
		t.Fatal("Expected non-nil cache")
	}
	if len(cached) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(cached))
	}
	if cached[0].Name != "field1" {
		t.Errorf("Expected field1, got %s", cached[0].Name)
	}

	// Verify immutability (modifying returned slice shouldn't affect cache)
	cached[0].Name = "modified"
	cached2 := state.GetFieldTreeCache(0)
	if cached2[0].Name == "modified" {
		t.Error("Cache was not properly copied - modification affected internal state")
	}

	// Set multiple caches
	tree2 := []DocField{
		{Name: "field3", Type: "bool"},
	}
	state.SetFieldTreeCache(1, tree2)

	if len(state.GetFieldTreeCache(0)) != 2 {
		t.Error("Cache 0 was affected by setting cache 1")
	}
	if len(state.GetFieldTreeCache(1)) != 1 {
		t.Errorf("Expected 1 field in cache 1, got %d", len(state.GetFieldTreeCache(1)))
	}

	// Clear cache
	state.ClearFieldTreeCache()
	if state.GetFieldTreeCache(0) != nil {
		t.Error("Expected nil cache after clear")
	}
	if state.GetFieldTreeCache(1) != nil {
		t.Error("Expected nil cache after clear")
	}
}

func TestDocumentationState_ChildrenCacheOperations(t *testing.T) {
	state := NewDocumentationState()

	// Test initial state
	if cache := state.GetChildrenCache(0); cache != nil {
		t.Error("Expected nil cache for non-existent index")
	}

	// Set children cache
	children := map[string]bool{
		"field1": true,
		"field2": false,
	}
	state.SetChildrenCache(0, children)

	// Get children cache
	cached := state.GetChildrenCache(0)
	if cached == nil {
		t.Fatal("Expected non-nil cache")
	}
	if len(cached) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(cached))
	}
	if !cached["field1"] {
		t.Error("Expected field1 to be true")
	}
	if cached["field2"] {
		t.Error("Expected field2 to be false")
	}

	// Verify immutability (modifying returned map shouldn't affect cache)
	cached["field1"] = false
	cached2 := state.GetChildrenCache(0)
	if !cached2["field1"] {
		t.Error("Cache was not properly copied - modification affected internal state")
	}

	// Set multiple caches
	children2 := map[string]bool{
		"field3": true,
	}
	state.SetChildrenCache(1, children2)

	if len(state.GetChildrenCache(0)) != 2 {
		t.Error("Cache 0 was affected by setting cache 1")
	}
	if len(state.GetChildrenCache(1)) != 1 {
		t.Errorf("Expected 1 entry in cache 1, got %d", len(state.GetChildrenCache(1)))
	}

	// Clear cache
	state.ClearChildrenCache()
	if state.GetChildrenCache(0) != nil {
		t.Error("Expected nil cache after clear")
	}
	if state.GetChildrenCache(1) != nil {
		t.Error("Expected nil cache after clear")
	}
}

func TestDocumentationState_ClearAllCaches(t *testing.T) {
	state := NewDocumentationState()

	// Set both caches
	tree := []DocField{{Name: "field1"}}
	state.SetFieldTreeCache(0, tree)

	children := map[string]bool{"field1": true}
	state.SetChildrenCache(0, children)

	// Verify both caches exist
	if state.GetFieldTreeCache(0) == nil {
		t.Error("Expected field tree cache to exist")
	}
	if state.GetChildrenCache(0) == nil {
		t.Error("Expected children cache to exist")
	}

	// Clear all caches
	state.ClearAllCaches()

	// Verify both caches are cleared
	if state.GetFieldTreeCache(0) != nil {
		t.Error("Expected field tree cache to be cleared")
	}
	if state.GetChildrenCache(0) != nil {
		t.Error("Expected children cache to be cleared")
	}
}

func TestDocumentationState_Reset(t *testing.T) {
	state := NewDocumentationState()

	// Set various state
	state.SetSelectedIdx(10)
	state.SetItemCount(50)
	state.SetCollapsed(0, true)
	state.SetCollapsed(1, true)
	state.SetFieldTreeCache(0, []DocField{{Name: "field1"}})
	state.SetChildrenCache(0, map[string]bool{"field1": true})

	// Reset
	state.Reset()

	// Verify everything is reset
	if state.GetSelectedIdx() != 0 {
		t.Errorf("Expected selectedIdx 0 after reset, got %d", state.GetSelectedIdx())
	}
	if state.GetItemCount() != 0 {
		t.Errorf("Expected itemCount 0 after reset, got %d", state.GetItemCount())
	}
	if state.GetCollapsed(0) {
		t.Error("Expected collapsed[0] to be false after reset")
	}
	if state.GetCollapsed(1) {
		t.Error("Expected collapsed[1] to be false after reset")
	}
	if state.GetFieldTreeCache(0) != nil {
		t.Error("Expected field tree cache to be cleared after reset")
	}
	if state.GetChildrenCache(0) != nil {
		t.Error("Expected children cache to be cleared after reset")
	}
}

func TestDocumentationState_ConcurrentAccess(t *testing.T) {
	state := NewDocumentationState()

	var wg sync.WaitGroup
	iterations := 50

	// Concurrent collapsed operations
	for i := 0; i < iterations; i++ {
		wg.Add(3)

		go func(idx int) {
			defer wg.Done()
			state.SetCollapsed(idx, true)
		}(i % 10)

		go func(idx int) {
			defer wg.Done()
			state.ToggleCollapsed(idx)
		}(i % 10)

		go func(idx int) {
			defer wg.Done()
			_ = state.GetCollapsed(idx)
		}(i % 10)
	}

	wg.Wait()
	// If we get here without panic or data race, success
}

func TestDocumentationState_ConcurrentNavigationAndCache(t *testing.T) {
	state := NewDocumentationState()

	var wg sync.WaitGroup
	iterations := 50

	for i := 0; i < iterations; i++ {
		wg.Add(6)

		go func() {
			defer wg.Done()
			state.Navigate(1, 100)
		}()

		go func() {
			defer wg.Done()
			state.Navigate(-1, 100)
		}()

		go func() {
			defer wg.Done()
			state.SetFieldTreeCache(0, []DocField{{Name: "field"}})
		}()

		go func() {
			defer wg.Done()
			_ = state.GetFieldTreeCache(0)
		}()

		go func() {
			defer wg.Done()
			state.SetChildrenCache(0, map[string]bool{"field": true})
		}()

		go func() {
			defer wg.Done()
			_ = state.GetChildrenCache(0)
		}()
	}

	wg.Wait()
	// If we get here without panic or data race, success
}

func TestDocumentationState_ConcurrentCacheClear(t *testing.T) {
	state := NewDocumentationState()

	var wg sync.WaitGroup
	iterations := 50

	for i := 0; i < iterations; i++ {
		wg.Add(4)

		go func() {
			defer wg.Done()
			state.SetFieldTreeCache(0, []DocField{{Name: "field"}})
		}()

		go func() {
			defer wg.Done()
			state.ClearFieldTreeCache()
		}()

		go func() {
			defer wg.Done()
			state.SetChildrenCache(0, map[string]bool{"field": true})
		}()

		go func() {
			defer wg.Done()
			state.ClearChildrenCache()
		}()
	}

	wg.Wait()
	// If we get here without panic or data race, success
}
