package tui

import (
	"sync"
)

// DocumentationState encapsulates all documentation viewer-related UI state
type DocumentationState struct {
	mu sync.RWMutex

	// Navigation and selection
	collapsed   map[int]bool // Collapsed state per response index
	selectedIdx int          // Currently selected item in doc viewer
	itemCount   int          // Cached total navigable items count

	// Caches for performance
	fieldTreeCache map[int][]DocField      // Cached field trees per response index
	childrenCache  map[int]map[string]bool // Cached hasChildren results per response index
}

// NewDocumentationState creates a new documentation state
func NewDocumentationState() *DocumentationState {
	return &DocumentationState{
		collapsed:      make(map[int]bool),
		selectedIdx:    0,
		itemCount:      0,
		fieldTreeCache: make(map[int][]DocField),
		childrenCache:  make(map[int]map[string]bool),
	}
}

// GetCollapsed returns whether an item is collapsed
func (s *DocumentationState) GetCollapsed(key int) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.collapsed[key]
}

// SetCollapsed sets the collapsed state for an item
func (s *DocumentationState) SetCollapsed(key int, collapsed bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.collapsed[key] = collapsed
}

// ToggleCollapsed toggles the collapsed state for an item
func (s *DocumentationState) ToggleCollapsed(key int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.collapsed[key] = !s.collapsed[key]
}

// ClearCollapsed clears all collapsed states
func (s *DocumentationState) ClearCollapsed() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.collapsed = make(map[int]bool)
}

// GetSelectedIdx returns the currently selected index
func (s *DocumentationState) GetSelectedIdx() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.selectedIdx
}

// SetSelectedIdx sets the currently selected index
func (s *DocumentationState) SetSelectedIdx(idx int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.selectedIdx = idx
}

// Navigate moves the selection by delta
func (s *DocumentationState) Navigate(delta int, maxCount int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.selectedIdx += delta

	// Clamp to valid range
	if s.selectedIdx < 0 {
		s.selectedIdx = 0
	} else if s.selectedIdx >= maxCount {
		s.selectedIdx = maxCount - 1
	}

	if s.selectedIdx < 0 {
		s.selectedIdx = 0
	}
}

// GetItemCount returns the cached item count
func (s *DocumentationState) GetItemCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.itemCount
}

// SetItemCount sets the cached item count
func (s *DocumentationState) SetItemCount(count int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.itemCount = count
}

// GetFieldTreeCache returns a copy of the field tree for a response index
func (s *DocumentationState) GetFieldTreeCache(respIdx int) []DocField {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tree, exists := s.fieldTreeCache[respIdx]
	if !exists {
		return nil
	}

	// Return copy to maintain immutability
	result := make([]DocField, len(tree))
	copy(result, tree)
	return result
}

// SetFieldTreeCache sets the field tree cache for a response index
func (s *DocumentationState) SetFieldTreeCache(respIdx int, tree []DocField) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fieldTreeCache[respIdx] = tree
}

// ClearFieldTreeCache clears the field tree cache
func (s *DocumentationState) ClearFieldTreeCache() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fieldTreeCache = make(map[int][]DocField)
}

// GetChildrenCache returns a copy of the children cache for a response index
func (s *DocumentationState) GetChildrenCache(respIdx int) map[string]bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cache, exists := s.childrenCache[respIdx]
	if !exists {
		return nil
	}

	// Return copy to maintain immutability
	result := make(map[string]bool)
	for k, v := range cache {
		result[k] = v
	}
	return result
}

// SetChildrenCache sets the children cache for a response index
func (s *DocumentationState) SetChildrenCache(respIdx int, cache map[string]bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.childrenCache[respIdx] = cache
}

// ClearChildrenCache clears the children cache
func (s *DocumentationState) ClearChildrenCache() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.childrenCache = make(map[int]map[string]bool)
}

// ClearAllCaches clears all caches (field tree and children)
func (s *DocumentationState) ClearAllCaches() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fieldTreeCache = make(map[int][]DocField)
	s.childrenCache = make(map[int]map[string]bool)
}

// Reset resets all documentation state
func (s *DocumentationState) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.collapsed = make(map[int]bool)
	s.selectedIdx = 0
	s.itemCount = 0
	s.fieldTreeCache = make(map[int][]DocField)
	s.childrenCache = make(map[int]map[string]bool)
}
