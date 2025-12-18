package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/studiowebux/restcli/internal/jsonpath"
)

// handleJSONPathHistoryKeys handles keyboard input in JSONPath history mode
func (m *Model) handleJSONPathHistoryKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		// Exit search mode or close modal
		if m.jsonpathHistorySearching {
			m.jsonpathHistorySearching = false
			m.jsonpathHistorySearch = ""
			m.loadFilteredBookmarks()
			m.jsonpathHistoryCursor = 0
		} else {
			// Close modal and return to filter editing
			m.mode = ModeNormal
			m.filterEditing = true
			m.jsonpathHistorySearch = ""
			m.jsonpathHistorySearching = false
			m.errorMsg = ""
		}

	case "q":
		// Close modal (only when not searching, as 'q' is a valid search char)
		if !m.jsonpathHistorySearching {
			m.mode = ModeNormal
			m.filterEditing = true
			m.jsonpathHistorySearch = ""
			m.jsonpathHistorySearching = false
			m.errorMsg = ""
		} else {
			// In search mode, 'q' is typed into search
			m.jsonpathHistorySearch += "q"
			m.loadFilteredBookmarks()
			m.jsonpathHistoryCursor = 0
		}

	case "/":
		// Enter search mode
		m.jsonpathHistorySearching = true

	case "backspace":
		// Handle search input deletion
		if m.jsonpathHistorySearching && len(m.jsonpathHistorySearch) > 0 {
			m.jsonpathHistorySearch = m.jsonpathHistorySearch[:len(m.jsonpathHistorySearch)-1]
			m.loadFilteredBookmarks()
			if m.jsonpathHistoryCursor >= len(m.jsonpathHistoryMatches) && len(m.jsonpathHistoryMatches) > 0 {
				m.jsonpathHistoryCursor = len(m.jsonpathHistoryMatches) - 1
			}
		}

	case "j", "down":
		if !m.jsonpathHistorySearching && len(m.jsonpathHistoryMatches) > 0 {
			m.jsonpathHistoryCursor = (m.jsonpathHistoryCursor + 1) % len(m.jsonpathHistoryMatches)
		}

	case "k", "up":
		if !m.jsonpathHistorySearching && len(m.jsonpathHistoryMatches) > 0 {
			m.jsonpathHistoryCursor = (m.jsonpathHistoryCursor - 1 + len(m.jsonpathHistoryMatches)) % len(m.jsonpathHistoryMatches)
		}

	case "g":
		if !m.jsonpathHistorySearching {
			// Go to top
			if m.gPressed {
				m.jsonpathHistoryCursor = 0
				m.gPressed = false
			} else {
				m.gPressed = true
			}
		} else {
			// In search mode, 'g' is typed into search
			m.jsonpathHistorySearch += "g"
			m.loadFilteredBookmarks()
			m.jsonpathHistoryCursor = 0
		}

	case "G":
		if !m.jsonpathHistorySearching {
			// Go to bottom
			if len(m.jsonpathHistoryMatches) > 0 {
				m.jsonpathHistoryCursor = len(m.jsonpathHistoryMatches) - 1
			}
			m.gPressed = false
		} else {
			// In search mode, 'G' is typed into search
			m.jsonpathHistorySearch += "G"
			m.loadFilteredBookmarks()
			m.jsonpathHistoryCursor = 0
		}

	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		if !m.jsonpathHistorySearching {
			// Quick select by number (1-9)
			num := int(msg.String()[0] - '0')
			index := num - 1
			if index < len(m.jsonpathHistoryMatches) {
				m.jsonpathHistoryCursor = index
				msg.Type = tea.KeyEnter
				return m.handleJSONPathHistoryKeys(tea.KeyMsg{Type: tea.KeyEnter})
			}
		} else {
			// In search mode, numbers are typed into search
			m.jsonpathHistorySearch += msg.String()
			m.loadFilteredBookmarks()
			m.jsonpathHistoryCursor = 0
		}

	case "0":
		if !m.jsonpathHistorySearching {
			// Quick select 10th item
			index := 9
			if index < len(m.jsonpathHistoryMatches) {
				m.jsonpathHistoryCursor = index
				msg.Type = tea.KeyEnter
				return m.handleJSONPathHistoryKeys(tea.KeyMsg{Type: tea.KeyEnter})
			}
		} else {
			// In search mode, '0' is typed into search
			m.jsonpathHistorySearch += "0"
			m.loadFilteredBookmarks()
			m.jsonpathHistoryCursor = 0
		}

	case "d", "delete":
		if !m.jsonpathHistorySearching {
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
		} else {
			// In search mode, 'd' is typed into search
			m.jsonpathHistorySearch += "d"
			m.loadFilteredBookmarks()
			m.jsonpathHistoryCursor = 0
		}

	case "enter":
		// Apply selected bookmark to filter (exit search mode first if active)
		if m.jsonpathHistorySearching {
			m.jsonpathHistorySearching = false
			return nil
		}

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

	default:
		// Handle search input only when in search mode
		if m.jsonpathHistorySearching && len(msg.String()) == 1 && msg.Type == tea.KeyRunes {
			m.jsonpathHistorySearch += msg.String()
			m.loadFilteredBookmarks()
			m.jsonpathHistoryCursor = 0
		} else if !m.jsonpathHistorySearching {
			// Reset 'g' press on any other key when not searching
			m.gPressed = false
		}
	}

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
