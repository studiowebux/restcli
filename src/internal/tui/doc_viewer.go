package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/studiowebux/restcli/internal/types"
)

// Documentation Viewer - Tree Structure and Collapse State Management
//
// This file handles the documentation viewer's field tree construction and collapse state.
// Fields can use dot notation (e.g., "user.profile.name") which creates virtual parent nodes.
//
// COLLAPSE STATE KEY SCHEME:
//   Section headers (fixed keys):
//     0 = Parameters section
//     1 = Responses section
//
//   Response field toggles (per-response keys):
//     100 + respIdx = Fields visibility for response at index respIdx
//     Example: Response 0 = key 100, Response 1 = key 101
//
//   Individual field toggles (unique keys per field per response):
//     200 + respIdx*1000 + hashString(field.Name)
//     Example: Response 0, field "user.name" = 200 + 0*1000 + hash("user.name")
//              Response 1, field "data.id" = 200 + 1*1000 + hash("data.id")
//
// LAZY LOADING STRATEGY:
//   1. Main sections start collapsed (keys 0 and 1)
//   2. Response fields start collapsed (keys 100+idx)
//   3. Field trees are only built when a response is first expanded
//   4. Field collapse states initialized only when parent response expanded
//
// VIRTUAL TREE CONSTRUCTION:
//   Input: ["user.name", "user.email", "user.profile.avatar"]
//   Tree:  user (virtual parent, type: object)
//          ├── user.name (actual field)
//          ├── user.email (actual field)
//          └── user.profile (virtual parent, type: object)
//              └── user.profile.avatar (actual field)

// ============================================================================
// COLLAPSE KEY GENERATION
// ============================================================================
// Functions for generating unique collapse state keys for documentation sections,
// responses, and fields. The key scheme ensures no collisions between different
// types of collapsible elements.

// Collapse key constants
const (
	collapseKeyParameters = 0  // Parameters section header
	collapseKeyResponses  = 1  // Responses section header
	collapseKeyResponseFieldsBase = 100 // Base for response fields toggle keys
	collapseKeyIndividualFieldBase = 200 // Base for individual field keys
	collapseKeyResponseMultiplier = 1000 // Multiplier for response index in field keys
)

// getCollapseKeyForSection returns the collapse key for a section header.
//
// Section types:
//   - "parameters": Returns key 0
//   - "responses": Returns key 1
//
// Returns the appropriate collapse key for the section.
func getCollapseKeyForSection(sectionType string) int {
	switch sectionType {
	case "parameters":
		return collapseKeyParameters
	case "responses":
		return collapseKeyResponses
	default:
		return -1 // Invalid section
	}
}

// getCollapseKeyForResponseFields returns the collapse key for a response's fields toggle.
//
// This is the key for the "▶ N fields" line that shows/hides all fields for a response.
//
// Formula: 100 + respIdx
//
// Example: Response 0 = key 100, Response 1 = key 101
func getCollapseKeyForResponseFields(respIdx int) int {
	return collapseKeyResponseFieldsBase + respIdx
}

// getCollapseKeyForField returns the collapse key for an individual field.
//
// Formula: 200 + respIdx*1000 + hashString(field.Name)
//
// The multiplier (1000) ensures fields from different responses don't collide,
// since hashString returns values in range [0, 999].
//
// Example:
//   - Response 0, field "user.name": 200 + 0*1000 + hash("user.name")
//   - Response 1, field "data.id": 200 + 1*1000 + hash("data.id")
func getCollapseKeyForField(respIdx int, fieldName string) int {
	return collapseKeyIndividualFieldBase + respIdx*collapseKeyResponseMultiplier + hashString(fieldName)
}

// hashString creates a simple hash of a string for collapse key generation.
//
// Used to create unique collapse keys for individual fields. Hash is bounded to [0, 999]
// to avoid integer overflow when combined with response index multiplier (respIdx*1000).
//
// Formula: h = ((h * 31) + charCode) for each character, then h % 1000
//
// Returns an integer in range [0, 999].
func hashString(s string) int {
	h := 0
	for _, c := range s {
		h = (h * 31) + int(c)
	}
	if h < 0 {
		h = -h
	}
	return h % 1000
}

// ============================================================================
// TREE STRUCTURE
// ============================================================================
// Functions for building and traversing the virtual field tree structure.
// Handles dot-notation fields by creating intermediate virtual parent nodes.

// DocField represents a field node in the documentation tree (can be actual or virtual parent)
type DocField struct {
	Name        string
	Type        string
	Required    bool
	Description string
	Deprecated  bool
	IsVirtual   bool // True if this is a virtual parent node created from dot notation
}

// buildVirtualFieldTree builds a complete tree of fields including virtual parent nodes.
//
// For fields using dot notation (e.g., "user.profile.name"), this function creates intermediate
// virtual parent nodes to build a proper tree structure. Virtual nodes have:
//   - IsVirtual = true
//   - Type = "object"
//   - Required = false
//
// Example:
//   Input:  ["user.name", "user.email", "data.items"]
//   Output: Virtual tree with nodes: "user" (virtual), "user.name", "user.email", "data" (virtual), "data.items"
//
// Returns a sorted array of all fields (actual + virtual) by name.
func buildVirtualFieldTree(fields []types.ResponseField) []DocField {
	allPaths := make(map[string]bool)
	fieldMap := make(map[string]DocField)

	// Add all actual fields
	for _, field := range fields {
		allPaths[field.Name] = true
		fieldMap[field.Name] = DocField{
			Name:        field.Name,
			Type:        field.Type,
			Required:    field.Required,
			Description: field.Description,
			Deprecated:  field.Deprecated,
			IsVirtual:   false,
		}
	}

	// Add all intermediate parent paths
	for _, field := range fields {
		parts := strings.Split(field.Name, ".")
		for i := 1; i < len(parts); i++ {
			parentPath := strings.Join(parts[:i], ".")
			if !allPaths[parentPath] {
				allPaths[parentPath] = true
				fieldMap[parentPath] = DocField{
					Name:      parentPath,
					Type:      "object",
					Required:  false,
					IsVirtual: true,
				}
			}
		}
	}

	// Convert to sorted array
	var result []DocField
	for _, field := range fieldMap {
		result = append(result, field)
	}

	// Sort by name for consistent display
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}

// hasChildren checks if a field has children in the tree.
//
// A field has children if any other field's parent path matches this field's name.
// For example, "user" has children if fields like "user.name" or "user.email" exist.
//
// Parameters:
//   - fieldName: The field name to check (e.g., "user" or "user.profile")
//   - allFields: Complete array of all fields in the tree
//
// Returns true if the field has at least one child, false otherwise.
func hasChildren(fieldName string, allFields []DocField) bool {
	for _, field := range allFields {
		if field.Name == fieldName {
			continue
		}
		parts := strings.Split(field.Name, ".")
		if len(parts) > 1 {
			parent := strings.Join(parts[:len(parts)-1], ".")
			if parent == fieldName {
				return true
			}
		}
	}
	return false
}

// getDirectChildren gets direct children of a parent field (non-recursive).
//
// Returns only the immediate children of the given parent path, not grandchildren.
// For example, if parentPath is "user", returns ["user.name", "user.email", "user.profile"]
// but NOT ["user.profile.avatar"] (that's a child of "user.profile", not "user").
//
// Parameters:
//   - parentPath: Parent field name (empty string for root-level fields)
//   - allFields: Complete array of all fields in the tree
//
// Returns array of direct child fields.
func getDirectChildren(parentPath string, allFields []DocField) []DocField {
	var children []DocField
	for _, field := range allFields {
		parts := strings.Split(field.Name, ".")
		var fieldParent string
		if len(parts) > 1 {
			fieldParent = strings.Join(parts[:len(parts)-1], ".")
		}
		if fieldParent == parentPath {
			children = append(children, field)
		}
	}
	return children
}

// ============================================================================
// CACHE MANAGEMENT
// ============================================================================
// Functions for building and retrieving cached field trees and children maps.
// Implements lazy loading - trees are only built when first needed.

// buildHasChildrenCache builds a cache of which fields have children for O(1) lookups.
//
// Without this cache, checking if a field has children requires O(n) scan of all fields for each field,
// resulting in O(n²) total complexity. With the cache, we scan once (O(n)) and get O(1) lookups.
//
// Algorithm:
//   1. Initialize all fields to false (no children)
//   2. For each field, mark its parent as having children
//
// Returns a map[fieldName]hasChildren for constant-time lookups during rendering.
func buildHasChildrenCache(allFields []DocField) map[string]bool {
	cache := make(map[string]bool)
	for _, field := range allFields {
		cache[field.Name] = false
	}
	for _, field := range allFields {
		parts := strings.Split(field.Name, ".")
		if len(parts) > 1 {
			parent := strings.Join(parts[:len(parts)-1], ".")
			cache[parent] = true
		}
	}
	return cache
}

// getOrBuildFieldTree retrieves or builds the field tree for a response.
//
// Implements cache-or-build pattern for lazy field tree construction.
// If the tree is already cached, returns it immediately (O(1)).
// Otherwise, builds the tree, caches it, and returns it.
//
// This ensures field trees are only built once per response, when first needed.
//
// Parameters:
//   - respIdx: Response index (for cache lookup)
//   - fields: Response fields to build tree from (only used if not cached)
//
// Returns the complete virtual field tree (actual + virtual parent nodes).
func (m *Model) getOrBuildFieldTree(respIdx int, fields []types.ResponseField) []DocField {
	allFields := m.docState.GetFieldTreeCache(respIdx)
	if allFields == nil {
		// Build and cache the tree
		allFields = buildVirtualFieldTree(fields)
		m.docState.SetFieldTreeCache(respIdx, allFields)
		m.docState.SetChildrenCache(respIdx, buildHasChildrenCache(allFields))
	}
	return allFields
}

// ============================================================================
// STATE INITIALIZATION
// ============================================================================
// Functions for initializing collapse states when documentation is first loaded
// or when a section is first expanded (lazy initialization).

// hasValidDocumentation checks if the current request has valid documentation.
//
// Returns true if both currentRequest and its Documentation field are non-nil.
// Used as a guard check before accessing documentation data.
func (m *Model) hasValidDocumentation() bool {
	return m.currentRequest != nil && m.currentRequest.Documentation != nil
}

// initializeCollapsedFields sets default collapse state when documentation viewer opens.
//
// Called when documentation is first loaded/parsed. Implements lazy loading strategy:
//   1. Collapses Parameters section (key 0)
//   2. Collapses Responses section (key 1)
//   3. Collapses each response's fields (keys 100+idx)
//   4. Field-level collapse states are initialized LATER when user first expands that response
//
// This avoids building field trees for all responses on initial load - trees are built on-demand.
func (m *Model) initializeCollapsedFields() {
	m.docState.ClearCollapsed()

	if !m.hasValidDocumentation() {
		return
	}

	doc := m.currentRequest.Documentation

	// CRITICAL: Collapse the main sections by default!
	m.docState.SetCollapsed(getCollapseKeyForSection("parameters"), true)
	m.docState.SetCollapsed(getCollapseKeyForSection("responses"), true)

	// Collapse each individual response's fields by default (lazy loading)
	if doc.Responses != nil {
		for respIdx := range doc.Responses {
			responseKey := getCollapseKeyForResponseFields(respIdx)
			m.docState.SetCollapsed(responseKey, true) // Collapse each response's fields
		}
	}

	// NOTE: Field-level collapse states are now initialized LAZILY
	// when a response is first expanded (see initializeFieldCollapseState)
}

// initializeFieldCollapseState lazily initializes collapse state for fields in a response.
//
// Called ONLY when user first expands a response's fields (toggles the "▶ N fields" line).
// This implements lazy loading - we don't build field trees or initialize field collapse states
// until the user actually wants to see them.
//
// Algorithm:
//   1. Build virtual field tree from response fields
//   2. Build hasChildren cache for O(1) lookups
//   3. For each field that has children, set collapse key to true (collapsed by default)
//
// Collapse keys: 200 + respIdx*1000 + hashString(field.Name)
//
// This ensures all parent fields start collapsed when first displayed, creating a clean
// tree view that user can progressively expand.
func (m *Model) initializeFieldCollapseState(respIdx int, fields []types.ResponseField) {
	allFields := buildVirtualFieldTree(fields)
	hasChildrenCache := buildHasChildrenCache(allFields)

	// Collapse all parents that have children
	for _, field := range allFields {
		if hasChildrenCache[field.Name] {
			key := getCollapseKeyForField(respIdx, field.Name)
			// Set collapsed state (only called during lazy initialization)
			m.docState.SetCollapsed(key, true)
		}
	}
}

// ============================================================================
// RENDERING
// ============================================================================
// Functions for rendering the documentation tree with proper formatting,
// indentation, collapse indicators, and selection highlighting.

// renderResponseFieldsTree recursively renders fields in tree structure.
//
// Builds the visual tree representation with proper indentation, collapse indicators,
// and highlighting for selected items. Respects collapse state for each field.
//
// Parameters:
//   - respIdx: Response index (for collapse key generation)
//   - parentPath: Parent field name (empty string for root level)
//   - allFields: Complete virtual field tree
//   - hasChildrenCache: Pre-built cache for O(1) hasChildren lookups
//   - currentIdx: Pointer to current navigation index (incremented as we render each line)
//   - content: String builder to append rendered output
//   - depth: Current tree depth (for indentation calculation)
//   - selectedLineNum: Pointer to track which line number is selected (for auto-scroll)
//
// Rendering algorithm:
//   1. Get direct children of parentPath
//   2. For each child:
//      a. Calculate indentation based on depth
//      b. Add collapse indicator (▶/▼) if field has children
//      c. Render field name, type, required badge, deprecated badge
//      d. Highlight if this is the selected item (currentIdx matches selectedIdx)
//      e. Increment currentIdx (this field counts as one navigable item)
//      f. If not collapsed and not virtual, show description line (doesn't increment currentIdx)
//      g. If not collapsed and has children, recursively render children at depth+1
//
// Virtual nodes show simpler format (just name), actual fields show full metadata.
func (m *Model) renderResponseFieldsTree(
	respIdx int,
	parentPath string,
	allFields []DocField,
	hasChildrenCache map[string]bool,
	currentIdx *int,
	content *strings.Builder,
	depth int,
	selectedLineNum *int,
) {
	if depth > 100 {
		return
	}

	children := getDirectChildren(parentPath, allFields)

	for _, field := range children {
		displayName := field.Name
		if strings.Contains(displayName, ".") {
			parts := strings.Split(displayName, ".")
			displayName = parts[len(parts)-1]
		}

		baseIndent := 6 + (depth * 2)
		indent := strings.Repeat(" ", baseIndent)

		// Check if this field has children using cache
		fieldHasChildren := hasChildrenCache[field.Name]
		fieldKey := getCollapseKeyForField(respIdx, field.Name)
		isCollapsed := m.docState.GetCollapsed(fieldKey)

		// Collapse indicator
		collapseIndicator := "  "
		if fieldHasChildren {
			if isCollapsed {
				collapseIndicator = "▶ "
			} else {
				collapseIndicator = "▼ "
			}
		}

		// Build field text
		var fieldText string
		if field.IsVirtual {
			// Virtual parent nodes - simpler format
			fieldText = fmt.Sprintf("%s%s%s", indent, collapseIndicator, styleTitle.Render(displayName))
		} else {
			// Actual fields - full details
			requiredBadge := styleWarning.Render("[optional]")
			if field.Required {
				requiredBadge = styleError.Render("[required]")
			}

			deprecatedBadge := ""
			if field.Deprecated {
				deprecatedBadge = " " + styleWarning.Render("[deprecated]")
			}

			fieldText = fmt.Sprintf("%s%s%s %s %s%s",
				indent,
				collapseIndicator,
				styleTitle.Render(displayName),
				styleSubtle.Render(fmt.Sprintf("{%s}", field.Type)),
				requiredBadge,
				deprecatedBadge,
			)
		}

		// Check if this is the selected item
		if *currentIdx == m.docState.GetSelectedIdx() {
			fieldText = styleSelected.Render(fieldText)
			*selectedLineNum = strings.Count(content.String(), "\n")
		}
		content.WriteString(fieldText + "\n")
		*currentIdx++

		// Show description if not collapsed and not virtual
		if !isCollapsed && !field.IsVirtual && field.Description != "" {
			textIndent := indent + "    "
			descText := textIndent + styleSubtle.Render(field.Description)
			content.WriteString(descText + "\n")
		}

		// Recursively add children if not collapsed
		if !isCollapsed && fieldHasChildren {
			m.renderResponseFieldsTree(respIdx, field.Name, allFields, hasChildrenCache, currentIdx, content, depth+1, selectedLineNum)
		}
	}
}
