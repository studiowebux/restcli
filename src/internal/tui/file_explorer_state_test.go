package tui

import (
	"sync"
	"testing"

	"github.com/studiowebux/restcli/internal/types"
)

func TestFileExplorerState_NewInitialization(t *testing.T) {
	state := NewFileExplorerState()

	if state == nil {
		t.Fatal("NewFileExplorerState returned nil")
	}

	AssertModelField(t, "fileIndex", state.GetCurrentIndex(), 0)
	AssertModelField(t, "fileOffset", state.GetScrollOffset(), 0)

	files := state.GetFiles()
	if len(files) != 0 {
		t.Errorf("Expected empty files, got %d files", len(files))
	}

	query, current, total := state.GetSearchInfo()
	AssertModelField(t, "searchQuery", query, "")
	AssertModelField(t, "currentMatch", current, 1)
	AssertModelField(t, "totalMatches", total, 0)
}

func TestFileExplorerState_SetFiles(t *testing.T) {
	state := NewFileExplorerState()

	files := []types.FileInfo{
		{Name: "test1.http", Path: "/path/test1.http"},
		{Name: "test2.http", Path: "/path/test2.http"},
		{Name: "test3.http", Path: "/path/test3.http"},
	}

	state.SetFiles(files, files)

	gotFiles := state.GetFiles()
	if len(gotFiles) != 3 {
		t.Errorf("Expected 3 files, got %d", len(gotFiles))
	}

	AssertModelField(t, "first file name", gotFiles[0].Name, "test1.http")
}

func TestFileExplorerState_Navigate(t *testing.T) {
	state := NewFileExplorerState()

	files := []types.FileInfo{
		{Name: "a.http"},
		{Name: "b.http"},
		{Name: "c.http"},
	}
	state.SetFiles(files, files)

	// Navigate down
	state.Navigate(1, 10)
	AssertModelField(t, "after navigate +1", state.GetCurrentIndex(), 1)

	state.Navigate(1, 10)
	AssertModelField(t, "after navigate +1", state.GetCurrentIndex(), 2)

	// Wrap around to beginning
	state.Navigate(1, 10)
	AssertModelField(t, "after wrap around", state.GetCurrentIndex(), 0)

	// Navigate up (wraps to end)
	state.Navigate(-1, 10)
	AssertModelField(t, "after navigate -1 wrap", state.GetCurrentIndex(), 2)
}

func TestFileExplorerState_GetCurrentFile(t *testing.T) {
	state := NewFileExplorerState()

	// No files
	file := state.GetCurrentFile()
	if file != nil {
		t.Error("Expected nil file when no files loaded")
	}

	// With files
	files := []types.FileInfo{
		{Name: "a.http", Path: "/a.http"},
		{Name: "b.http", Path: "/b.http"},
	}
	state.SetFiles(files, files)

	file = state.GetCurrentFile()
	if file == nil {
		t.Fatal("Expected non-nil file")
	}
	AssertModelField(t, "current file name", file.Name, "a.http")

	// After navigation
	state.Navigate(1, 10)
	file = state.GetCurrentFile()
	if file == nil {
		t.Fatal("Expected non-nil file after navigation")
	}
	AssertModelField(t, "current file name after nav", file.Name, "b.http")
}

func TestFileExplorerState_Search(t *testing.T) {
	state := NewFileExplorerState()

	files := []types.FileInfo{
		{Name: "users.http"},
		{Name: "auth.http"},
		{Name: "posts.http"},
		{Name: "user-profile.http"},
	}
	state.SetFiles(files, files)

	// Search for "user"
	count, err := state.Search("user", 10)
	if err != "" {
		t.Errorf("Search returned error: %s", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 matches, got %d", count)
	}

	// Should jump to first match
	AssertModelField(t, "fileIndex after search", state.GetCurrentIndex(), 0) // "users.http"

	query, current, total := state.GetSearchInfo()
	AssertModelField(t, "search query", query, "user")
	AssertModelField(t, "current match", current, 1)
	AssertModelField(t, "total matches", total, 2)
}

func TestFileExplorerState_SearchRegex(t *testing.T) {
	state := NewFileExplorerState()

	files := []types.FileInfo{
		{Name: "test1.http"},
		{Name: "test2.http"},
		{Name: "prod.http"},
	}
	state.SetFiles(files, files)

	// Regex pattern
	count, err := state.Search("^test", 10)
	if err != "" {
		t.Errorf("Regex search returned error: %s", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 matches, got %d", count)
	}
}

func TestFileExplorerState_SearchNoMatches(t *testing.T) {
	state := NewFileExplorerState()

	files := []types.FileInfo{
		{Name: "a.http"},
		{Name: "b.http"},
	}
	state.SetFiles(files, files)

	count, err := state.Search("xyz", 10)
	if err == "" {
		t.Error("Expected error message for no matches")
	}
	if count != 0 {
		t.Errorf("Expected 0 matches, got %d", count)
	}
}

func TestFileExplorerState_NextPrevSearchMatch(t *testing.T) {
	state := NewFileExplorerState()

	files := []types.FileInfo{
		{Name: "a1.http"}, // index 0
		{Name: "b.http"},  // index 1
		{Name: "a2.http"}, // index 2
		{Name: "c.http"},  // index 3
		{Name: "a3.http"}, // index 4
	}
	state.SetFiles(files, files)

	// Search for "a"
	state.Search("a", 10)

	// Should be at first match (index 0)
	AssertModelField(t, "first match", state.GetCurrentIndex(), 0)

	// Next match
	state.NextSearchMatch(10)
	AssertModelField(t, "second match", state.GetCurrentIndex(), 2)

	state.NextSearchMatch(10)
	AssertModelField(t, "third match", state.GetCurrentIndex(), 4)

	// Wrap around
	state.NextSearchMatch(10)
	AssertModelField(t, "wrap to first", state.GetCurrentIndex(), 0)

	// Previous match
	state.PrevSearchMatch(10)
	AssertModelField(t, "prev to last", state.GetCurrentIndex(), 4)
}

func TestFileExplorerState_ClearSearch(t *testing.T) {
	state := NewFileExplorerState()

	files := []types.FileInfo{
		{Name: "test.http"},
	}
	state.SetFiles(files, files)

	state.Search("test", 10)

	query, _, total := state.GetSearchInfo()
	if query == "" || total == 0 {
		t.Error("Expected search to be active")
	}

	state.ClearSearch()

	query, _, total = state.GetSearchInfo()
	AssertModelField(t, "query after clear", query, "")
	AssertModelField(t, "total after clear", total, 0)
}

func TestFileExplorerState_SetTagFilter(t *testing.T) {
	state := NewFileExplorerState()

	allFiles := []types.FileInfo{
		{Name: "a.http", Tags: []string{"auth", "v1"}},
		{Name: "b.http", Tags: []string{"users", "v1"}},
		{Name: "c.http", Tags: []string{"auth", "v2"}},
		{Name: "d.http", Tags: []string{"posts"}},
	}
	state.SetFiles(allFiles, allFiles)

	// Filter by "auth" tag
	state.SetTagFilter([]string{"auth"})

	files := state.GetFiles()
	if len(files) != 2 {
		t.Errorf("Expected 2 files with 'auth' tag, got %d", len(files))
	}

	// Clear filter
	state.SetTagFilter([]string{})
	files = state.GetFiles()
	if len(files) != 4 {
		t.Errorf("Expected all 4 files after clearing filter, got %d", len(files))
	}

	// Filter by multiple tags (OR logic)
	state.SetTagFilter([]string{"auth", "users"})
	files = state.GetFiles()
	if len(files) != 3 {
		t.Errorf("Expected 3 files with 'auth' or 'users' tag, got %d", len(files))
	}
}

func TestFileExplorerState_CollectionManagement(t *testing.T) {
	state := NewFileExplorerState()

	// Initial state
	if state.GetActiveCollection() != nil {
		t.Error("Expected nil active collection initially")
	}
	AssertModelField(t, "initial collection index", state.GetCollectionIndex(), 0)

	// Set collection
	collection := &types.Collection{Name: "Test Collection"}
	state.SetActiveCollection(collection)

	got := state.GetActiveCollection()
	if got == nil || got.Name != "Test Collection" {
		t.Error("Failed to set active collection")
	}

	// Set collection index
	state.SetCollectionIndex(5)
	AssertModelField(t, "collection index", state.GetCollectionIndex(), 5)
}

func TestFileExplorerState_ScrollOffset(t *testing.T) {
	state := NewFileExplorerState()

	files := make([]types.FileInfo, 20)
	for i := range files {
		files[i] = types.FileInfo{Name: string(rune('a' + i)) + ".http"}
	}
	state.SetFiles(files, files)

	pageSize := 10

	// Navigate to item 15 (should adjust scroll)
	for i := 0; i < 15; i++ {
		state.Navigate(1, pageSize)
	}

	offset := state.GetScrollOffset()
	if offset < 6 { // Should have scrolled down
		t.Errorf("Expected scroll offset >= 6, got %d", offset)
	}

	// Navigate back to top (navigate explicitly to index 0)
	currentIdx := state.GetCurrentIndex()
	state.Navigate(-currentIdx, pageSize) // Navigate to beginning
	AssertModelField(t, "back to top index", state.GetCurrentIndex(), 0)
	AssertModelField(t, "back to top offset", state.GetScrollOffset(), 0)
}

func TestFileExplorerState_ConcurrentAccess(t *testing.T) {
	state := NewFileExplorerState()

	files := make([]types.FileInfo, 100)
	for i := range files {
		files[i] = types.FileInfo{Name: "file" + string(rune('0'+i%10)) + ".http"}
	}
	state.SetFiles(files, files)

	var wg sync.WaitGroup
	iterations := 50

	// Concurrent navigation
	for i := 0; i < iterations; i++ {
		wg.Add(3)

		go func() {
			defer wg.Done()
			state.Navigate(1, 10)
		}()

		go func() {
			defer wg.Done()
			state.Navigate(-1, 10)
		}()

		go func() {
			defer wg.Done()
			_ = state.GetCurrentFile()
		}()
	}

	// Concurrent search
	for i := 0; i < iterations; i++ {
		wg.Add(2)

		go func() {
			defer wg.Done()
			state.Search("file", 10)
		}()

		go func() {
			defer wg.Done()
			state.NextSearchMatch(10)
		}()
	}

	wg.Wait()
	// If we get here without panic or data race, success
}

func BenchmarkFileExplorerState_Navigate(b *testing.B) {
	state := NewFileExplorerState()

	files := make([]types.FileInfo, 1000)
	for i := range files {
		files[i] = types.FileInfo{Name: "file" + string(rune(i)) + ".http"}
	}
	state.SetFiles(files, files)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		state.Navigate(1, 10)
	}
}

func BenchmarkFileExplorerState_Search(b *testing.B) {
	state := NewFileExplorerState()

	files := make([]types.FileInfo, 1000)
	for i := range files {
		files[i] = types.FileInfo{Name: "file" + string(rune(i%100)) + ".http"}
	}
	state.SetFiles(files, files)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		state.Search("file", 10)
	}
}

func TestFileExplorerState_GoToTop(t *testing.T) {
	state := NewFileExplorerState()

	files := []types.FileInfo{
		{Name: "a.http", Path: "/test/a.http"},
		{Name: "b.http", Path: "/test/b.http"},
		{Name: "c.http", Path: "/test/c.http"},
	}
	state.SetFiles(files, files)

	// Navigate to middle
	state.Navigate(1, 10)
	AssertModelField(t, "navigate to middle", state.GetCurrentIndex(), 1)

	// GoToTop should jump to first file
	state.GoToTop(10)
	AssertModelField(t, "after GoToTop", state.GetCurrentIndex(), 0)

	// Empty file list
	state.SetFiles([]types.FileInfo{}, []types.FileInfo{})
	state.GoToTop(10) // Should not panic
	AssertModelField(t, "empty list", state.GetCurrentIndex(), 0)
}

func TestFileExplorerState_GoToBottom(t *testing.T) {
	state := NewFileExplorerState()

	files := []types.FileInfo{
		{Name: "a.http", Path: "/test/a.http"},
		{Name: "b.http", Path: "/test/b.http"},
		{Name: "c.http", Path: "/test/c.http"},
	}
	state.SetFiles(files, files)

	// Start at top
	AssertModelField(t, "initial index", state.GetCurrentIndex(), 0)

	// GoToBottom should jump to last file
	state.GoToBottom(10)
	AssertModelField(t, "after GoToBottom", state.GetCurrentIndex(), 2)

	// Empty file list
	state.SetFiles([]types.FileInfo{}, []types.FileInfo{})
	state.GoToBottom(10) // Should not panic
	AssertModelField(t, "empty list", state.GetCurrentIndex(), 0)
}

func TestFileExplorerState_NavigateToFile(t *testing.T) {
	state := NewFileExplorerState()

	files := []types.FileInfo{
		{Name: "a.http", Path: "/test/a.http"},
		{Name: "b.http", Path: "/test/b.http"},
		{Name: "c.http", Path: "/test/c.http"},
	}
	state.SetFiles(files, files)

	// Navigate to b.http
	found := state.NavigateToFile("/test/b.http", 10)
	AssertModelField(t, "found b.http", found, true)
	AssertModelField(t, "navigated to b.http", state.GetCurrentIndex(), 1)

	// Navigate to c.http
	found = state.NavigateToFile("/test/c.http", 10)
	AssertModelField(t, "found c.http", found, true)
	AssertModelField(t, "navigated to c.http", state.GetCurrentIndex(), 2)

	// Navigate to non-existent file
	found = state.NavigateToFile("/test/d.http", 10)
	AssertModelField(t, "not found d.http", found, false)
	AssertModelField(t, "index unchanged", state.GetCurrentIndex(), 2)

	// Empty file list
	state.SetFiles([]types.FileInfo{}, []types.FileInfo{})
	found = state.NavigateToFile("/test/a.http", 10)
	AssertModelField(t, "empty list not found", found, false)
}
