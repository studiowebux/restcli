package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/studiowebux/restcli/internal/types"
)

// DocField represents a field node in the documentation tree (can be actual or virtual parent)
type DocField struct {
	Name        string
	Type        string
	Required    bool
	Description string
	Deprecated  bool
	IsVirtual   bool // True if this is a virtual parent node created from dot notation
}

// buildVirtualFieldTree builds a complete tree of fields including virtual parent nodes
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

// hasChildren checks if a field has children in the tree
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

// getDirectChildren gets direct children of a parent field
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

// initializeCollapsedFields sets default collapse state (all fields with children)
func (m *Model) initializeCollapsedFields() {
	m.docCollapsed = make(map[int]bool)

	if m.currentRequest == nil || m.currentRequest.Documentation == nil {
		return
	}

	doc := m.currentRequest.Documentation

	// CRITICAL: Collapse the main sections by default!
	m.docCollapsed[0] = true // Parameters section
	m.docCollapsed[1] = true // Responses section

	// Collapse each individual response's fields by default (lazy loading)
	if doc.Responses != nil {
		for respIdx := range doc.Responses {
			// Use key pattern: 100 + respIdx (separate from field keys which start at 200)
			responseKey := 100 + respIdx
			m.docCollapsed[responseKey] = true // Collapse each response's fields
		}
	}

	// NOTE: Field-level collapse states are now initialized LAZILY
	// when a response is first expanded (see initializeFieldCollapseState)
}

// Simple string hash for creating unique keys
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

// buildHasChildrenCache builds a cache of which fields have children (O(n) instead of O(n²))
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

// initializeFieldCollapseState lazily initializes collapse state for fields in a response
// Only called when a response is first expanded - avoids building tree for all responses on load
func (m *Model) initializeFieldCollapseState(respIdx int, fields []types.ResponseField) {
	allFields := buildVirtualFieldTree(fields)
	hasChildrenCache := buildHasChildrenCache(allFields)

	// Collapse all parents that have children
	for _, field := range allFields {
		if hasChildrenCache[field.Name] {
			key := 200 + respIdx*1000 + hashString(field.Name)
			// Only set if not already set (preserve user's expand/collapse actions)
			if _, exists := m.docCollapsed[key]; !exists {
				m.docCollapsed[key] = true
			}
		}
	}
}

// renderResponseFieldsTree recursively renders fields in tree structure
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
		fieldKey := 200 + respIdx*1000 + hashString(field.Name)
		isCollapsed := m.docCollapsed[fieldKey]

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
		if *currentIdx == m.docSelectedIdx {
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
