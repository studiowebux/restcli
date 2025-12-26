package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/studiowebux/restcli/internal/jsonpath"
	"github.com/studiowebux/restcli/internal/keybinds"
)

// handleJSONPathHistoryKeys handles keyboard input in JSONPath history mode
func (m *Model) handleJSONPathHistoryKeys(msg tea.KeyMsg) tea.Cmd {
	// Handle search mode input first
	if m.jsonpathHistorySearching {
		switch msg.String() {
		case "esc":
			// Exit search mode
			m.jsonpathHistorySearching = false
			m.jsonpathHistorySearch = ""
			m.loadFilteredBookmarks()
			m.jsonpathHistoryCursor = 0
			return nil

		case "backspace":
			// Delete search character
			if len(m.jsonpathHistorySearch) > 0 {
				m.jsonpathHistorySearch = m.jsonpathHistorySearch[:len(m.jsonpathHistorySearch)-1]
				m.loadFilteredBookmarks()
				if m.jsonpathHistoryCursor >= len(m.jsonpathHistoryMatches) && len(m.jsonpathHistoryMatches) > 0 {
					m.jsonpathHistoryCursor = len(m.jsonpathHistoryMatches) - 1
				}
			}
			return nil

		case "enter":
			// Exit search mode
			m.jsonpathHistorySearching = false
			return nil

		default:
			// All other characters are typed into search
			if len(msg.String()) == 1 && msg.Type == tea.KeyRunes {
				m.jsonpathHistorySearch += msg.String()
				m.loadFilteredBookmarks()
				m.jsonpathHistoryCursor = 0
			}
			return nil
		}
	}

	// Not in search mode - handle special keys
	switch msg.String() {
	case "/":
		// Enter search mode
		m.jsonpathHistorySearching = true
		return nil

	case "q":
		// Close modal
		m.mode = ModeNormal
		m.filterEditing = true
		m.jsonpathHistorySearch = ""
		m.jsonpathHistorySearching = false
		m.errorMsg = ""
		return nil

	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		// Quick select by number (1-9)
		num := int(msg.String()[0] - '0')
		index := num - 1
		if index < len(m.jsonpathHistoryMatches) {
			m.jsonpathHistoryCursor = index
			msg.Type = tea.KeyEnter
			return m.handleJSONPathHistoryKeys(tea.KeyMsg{Type: tea.KeyEnter})
		}
		return nil

	case "0":
		// Quick select 10th item
		index := 9
		if index < len(m.jsonpathHistoryMatches) {
			m.jsonpathHistoryCursor = index
			msg.Type = tea.KeyEnter
			return m.handleJSONPathHistoryKeys(tea.KeyMsg{Type: tea.KeyEnter})
		}
		return nil

	case "d", "delete":
		// Delete selected bookmark
		if len(m.jsonpathHistoryMatches) == 0 {
			m.errorMsg = "No bookmarks to delete"
			return nil
		}

		selectedBookmark := m.jsonpathHistoryMatches[m.jsonpathHistoryCursor]
		if err := m.bookmarkManager.Delete(selectedBookmark.ID); err != nil {
			m.errorMsg = fmt.Sprintf("Failed to delete bookmark: %v", err)
			return nil
		}

		// Reload bookmarks
		m.loadFilteredBookmarks()

		// Adjust cursor if needed
		if m.jsonpathHistoryCursor >= len(m.jsonpathHistoryMatches) && len(m.jsonpathHistoryMatches) > 0 {
			m.jsonpathHistoryCursor = len(m.jsonpathHistoryMatches) - 1
		}

		m.statusMsg = "Bookmark deleted"
		m.errorMsg = ""
		return nil

	case "enter":
		// Apply selected bookmark to filter
		if len(m.jsonpathHistoryMatches) == 0 {
			m.errorMsg = "No bookmarks available"
			return nil
		}

		selectedBookmark := m.jsonpathHistoryMatches[m.jsonpathHistoryCursor]
		m.filterInput = selectedBookmark.Expression
		m.filterCursor = len(m.filterInput)
		m.mode = ModeNormal
		m.filterEditing = true
		m.jsonpathHistorySearch = ""
		m.jsonpathHistorySearching = false
		m.statusMsg = "Bookmark loaded"
		m.errorMsg = ""
		return nil
	}

	// Use registry for navigation
	action, ok, partial := m.keybinds.MatchMultiKey(keybinds.ContextModal, msg.String())
	if partial {
		return nil
	}
	if !ok {
		// Reset 'g' press on any other key
		m.gPressed = false
		return nil
	}

	switch action {
	case keybinds.ActionCloseModal:
		// Close modal and return to filter editing
		m.mode = ModeNormal
		m.filterEditing = true
		m.jsonpathHistorySearch = ""
		m.jsonpathHistorySearching = false
		m.errorMsg = ""

	case keybinds.ActionNavigateDown:
		if len(m.jsonpathHistoryMatches) > 0 {
			m.jsonpathHistoryCursor = (m.jsonpathHistoryCursor + 1) % len(m.jsonpathHistoryMatches)
		}

	case keybinds.ActionNavigateUp:
		if len(m.jsonpathHistoryMatches) > 0 {
			m.jsonpathHistoryCursor = (m.jsonpathHistoryCursor - 1 + len(m.jsonpathHistoryMatches)) % len(m.jsonpathHistoryMatches)
		}

	case keybinds.ActionGoToTop:
		// Go to top (triggered by gg)
		m.jsonpathHistoryCursor = 0

	case keybinds.ActionGoToBottom:
		// Go to bottom
		if len(m.jsonpathHistoryMatches) > 0 {
			m.jsonpathHistoryCursor = len(m.jsonpathHistoryMatches) - 1
		}
	}

	m.gPressed = false
	return nil
}

// loadFilteredBookmarks loads bookmarks filtered by search query
func (m *Model) loadFilteredBookmarks() {
	if m.bookmarkManager == nil {
		m.jsonpathHistoryMatches = []jsonpath.Bookmark{}
		return
	}

	var err error
	if m.jsonpathHistorySearch == "" {
		m.jsonpathHistoryMatches, err = m.bookmarkManager.List()
	} else {
		m.jsonpathHistoryMatches, err = m.bookmarkManager.Search(m.jsonpathHistorySearch)
	}

	if err != nil {
		m.errorMsg = fmt.Sprintf("Failed to load bookmarks: %v", err)
		m.jsonpathHistoryMatches = []jsonpath.Bookmark{}
	}
}

// renderJSONPathHistoryModal renders the JSONPath bookmark history modal
func (m *Model) renderJSONPathHistoryModal() string {
	if len(m.jsonpathHistoryMatches) == 0 {
		content := "No bookmarks found\n\n"
		content += "Press Ctrl+S in filter modal to save bookmarks\nPress ESC to close"

		footer := m.buildJSONPathHistoryFooter()
		return m.renderModalWithFooter("JSONPath Bookmarks", content, footer, 70, 10)
	}

	// Build bookmark list with selection
	var content strings.Builder

	// Render all bookmarks
	for i, bookmark := range m.jsonpathHistoryMatches {
		// Add number prefix for first 10 items
		var prefix string
		if i < 9 {
			prefix = fmt.Sprintf("%d. ", i+1)
		} else if i == 9 {
			prefix = "0. "
		} else {
			prefix = "   "
		}

		// Truncate long expressions
		expr := bookmark.Expression
		maxLen := 60
		if len(expr) > maxLen {
			expr = expr[:maxLen-3] + "..."
		}

		if i == m.jsonpathHistoryCursor {
			content.WriteString(styleSelected.Render(fmt.Sprintf("%s%s", prefix, expr)) + "\n")
		} else {
			content.WriteString(fmt.Sprintf("  %s%s\n", prefix, expr))
		}
	}

	// Show error if present
	if m.errorMsg != "" {
		content.WriteString("\n\n")
		content.WriteString(styleError.Render(m.errorMsg))
	}

	footer := m.buildJSONPathHistoryFooter()

	return m.renderModalWithFooterAndScroll("JSONPath Bookmarks", content.String(), footer, 75, 20, m.jsonpathHistoryCursor)
}

// buildJSONPathHistoryFooter builds the footer with search input or help text
func (m *Model) buildJSONPathHistoryFooter() string {
	if m.jsonpathHistorySearching {
		// Show search input in footer
		return fmt.Sprintf("Search: %s█", m.jsonpathHistorySearch)
	}

	// Show help text
	if m.jsonpathHistorySearch != "" {
		// Search is active but not currently typing
		return fmt.Sprintf("[↑/↓ j/k] navigate [1-9,0] select [d] delete [enter] apply [/] search again [esc] clear • Filter: %s", m.jsonpathHistorySearch)
	}

	return "[↑/↓ j/k] navigate [1-9,0] select [d] delete [enter] apply [/] search [esc] close"
}
