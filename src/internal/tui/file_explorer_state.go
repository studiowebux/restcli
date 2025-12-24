package tui

import (
	"regexp"
	"strings"
	"sync"

	"github.com/studiowebux/restcli/internal/types"
)

// FileExplorerState manages file navigation and filtering with thread safety
type FileExplorerState struct {
	mu sync.RWMutex

	// File lists
	files    []types.FileInfo // Filtered/displayed file list
	allFiles []types.FileInfo // Unfiltered file list for tag filtering

	// Navigation
	fileIndex  int // Current selected file index
	fileOffset int // Scroll offset for file list

	// Search
	searchQuery   string // Current search query
	searchMatches []int  // Indices of files matching search
	searchIndex   int    // Current position in search results

	// Filtering
	activeCollection *types.Collection // Currently active collection filter
	tagFilter        []string          // Active tag filters
	collectionIndex  int               // Selected collection in browser
}

// NewFileExplorerState creates a new file explorer state
func NewFileExplorerState() *FileExplorerState {
	return &FileExplorerState{
		files:         []types.FileInfo{},
		allFiles:      []types.FileInfo{},
		fileIndex:     0,
		fileOffset:    0,
		searchMatches: []int{},
		searchIndex:   0,
		tagFilter:     []string{},
	}
}

// GetFiles returns a copy of the current file list
func (f *FileExplorerState) GetFiles() []types.FileInfo {
	f.mu.RLock()
	defer f.mu.RUnlock()
	files := make([]types.FileInfo, len(f.files))
	copy(files, f.files)
	return files
}

// SetFiles updates the file list
func (f *FileExplorerState) SetFiles(files, allFiles []types.FileInfo) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.files = files
	f.allFiles = allFiles

	// Reset navigation if current index is out of bounds
	if f.fileIndex >= len(f.files) {
		f.fileIndex = 0
		f.fileOffset = 0
	}
}

// GetCurrentIndex returns the current file index
func (f *FileExplorerState) GetCurrentIndex() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.fileIndex
}

// GetCurrentFile returns the currently selected file (or nil if none)
func (f *FileExplorerState) GetCurrentFile() *types.FileInfo {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if len(f.files) == 0 || f.fileIndex < 0 || f.fileIndex >= len(f.files) {
		return nil
	}

	file := f.files[f.fileIndex]
	return &file
}

// Navigate moves the selection by delta positions (supports wrapping)
func (f *FileExplorerState) Navigate(delta int, pageSize int) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if len(f.files) == 0 {
		return
	}

	f.fileIndex += delta

	// Wrap around (circular navigation)
	if f.fileIndex < 0 {
		f.fileIndex = len(f.files) - 1
	} else if f.fileIndex >= len(f.files) {
		f.fileIndex = 0
	}

	// Adjust scroll offset
	f.adjustScrollOffsetLocked(pageSize)
}

// adjustScrollOffsetLocked adjusts scroll offset (must be called with lock held)
func (f *FileExplorerState) adjustScrollOffsetLocked(pageSize int) {
	if f.fileIndex < f.fileOffset {
		f.fileOffset = f.fileIndex
	} else if f.fileIndex >= f.fileOffset+pageSize {
		f.fileOffset = f.fileIndex - pageSize + 1
	}
}

// AdjustScrollOffset adjusts the scroll offset based on current index and page size
func (f *FileExplorerState) AdjustScrollOffset(pageSize int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.adjustScrollOffsetLocked(pageSize)
}

// GetScrollOffset returns the current scroll offset
func (f *FileExplorerState) GetScrollOffset() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.fileOffset
}

// Search performs a search on file names (regex or substring)
func (f *FileExplorerState) Search(query string, pageSize int) (matchCount int, errorMsg string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.searchQuery = query
	f.searchMatches = nil

	if query == "" {
		return 0, ""
	}

	// Auto-detect regex
	useRegex := isRegexPattern(query)

	if useRegex {
		pattern, err := regexp.Compile(query)
		if err != nil {
			// Fall back to substring search
			f.searchSubstringLocked(query)
		} else {
			for i, file := range f.files {
				if pattern.MatchString(file.Name) {
					f.searchMatches = append(f.searchMatches, i)
				}
			}
		}
	} else {
		f.searchSubstringLocked(query)
	}

	if len(f.searchMatches) == 0 {
		return 0, "No matching files found"
	}

	// Jump to first match
	f.searchIndex = 0
	f.fileIndex = f.searchMatches[0]
	f.adjustScrollOffsetLocked(pageSize)

	return len(f.searchMatches), ""
}

// searchSubstringLocked performs case-insensitive substring search (must be called with lock held)
func (f *FileExplorerState) searchSubstringLocked(query string) {
	queryLower := strings.ToLower(query)
	for i, file := range f.files {
		if strings.Contains(strings.ToLower(file.Name), queryLower) {
			f.searchMatches = append(f.searchMatches, i)
		}
	}
}

// NextSearchMatch navigates to the next search match
func (f *FileExplorerState) NextSearchMatch(pageSize int) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	if len(f.searchMatches) == 0 {
		return false
	}

	f.searchIndex = (f.searchIndex + 1) % len(f.searchMatches)
	f.fileIndex = f.searchMatches[f.searchIndex]
	f.adjustScrollOffsetLocked(pageSize)

	return true
}

// PrevSearchMatch navigates to the previous search match
func (f *FileExplorerState) PrevSearchMatch(pageSize int) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	if len(f.searchMatches) == 0 {
		return false
	}

	f.searchIndex--
	if f.searchIndex < 0 {
		f.searchIndex = len(f.searchMatches) - 1
	}
	f.fileIndex = f.searchMatches[f.searchIndex]
	f.adjustScrollOffsetLocked(pageSize)

	return true
}

// GetSearchInfo returns current search state
func (f *FileExplorerState) GetSearchInfo() (query string, currentMatch, totalMatches int) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.searchQuery, f.searchIndex + 1, len(f.searchMatches)
}

// ClearSearch clears the current search
func (f *FileExplorerState) ClearSearch() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.searchQuery = ""
	f.searchMatches = nil
	f.searchIndex = 0
}

// SetTagFilter applies tag filtering to the file list
func (f *FileExplorerState) SetTagFilter(tags []string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.tagFilter = tags

	if len(tags) == 0 {
		// No filter - show all files
		f.files = f.allFiles
	} else {
		// Filter files by tags
		f.files = nil
		for _, file := range f.allFiles {
			if hasAnyTag(file.Tags, tags) {
				f.files = append(f.files, file)
			}
		}
	}

	// Reset index if out of bounds
	if f.fileIndex >= len(f.files) {
		f.fileIndex = 0
		f.fileOffset = 0
	}
}

// GetTagFilter returns the current tag filter
func (f *FileExplorerState) GetTagFilter() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	tags := make([]string, len(f.tagFilter))
	copy(tags, f.tagFilter)
	return tags
}

// SetActiveCollection sets the active collection
func (f *FileExplorerState) SetActiveCollection(collection *types.Collection) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.activeCollection = collection
}

// GetActiveCollection returns the active collection
func (f *FileExplorerState) GetActiveCollection() *types.Collection {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.activeCollection
}

// SetCollectionIndex sets the collection browser index
func (f *FileExplorerState) SetCollectionIndex(index int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.collectionIndex = index
}

// GetCollectionIndex returns the collection browser index
func (f *FileExplorerState) GetCollectionIndex() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.collectionIndex
}

// hasAnyTag checks if a file has any of the specified tags
func hasAnyTag(fileTags, filterTags []string) bool {
	for _, ft := range filterTags {
		for _, tag := range fileTags {
			if tag == ft {
				return true
			}
		}
	}
	return false
}
