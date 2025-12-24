package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/studiowebux/restcli/internal/filter"
	"github.com/studiowebux/restcli/internal/keybinds"
)

// handleFilterKeys handles key input for filter modal
func (m *Model) handleFilterKeys(msg tea.KeyMsg) tea.Cmd {
	// Handle special keys not in registry
	switch msg.String() {
	case "ctrl+s":
		// Save current expression as bookmark
		if m.filterInput == "" {
			m.filterError = "Cannot save empty expression"
			return nil
		}

		if m.bookmarkManager == nil {
			m.filterError = "Bookmark manager not available"
			return nil
		}

		saved, err := m.bookmarkManager.Save(m.filterInput)
		if err != nil {
			m.filterError = fmt.Sprintf("Failed to save bookmark: %v", err)
			return nil
		}

		if saved {
			m.filterError = styleSuccess.Render("✓ Bookmark saved successfully")
		} else {
			m.filterError = styleWarning.Render("Bookmark already exists")
		}
		return nil

	case "up":
		// Open history modal if input is empty
		if m.filterInput == "" {
			m.mode = ModeJSONPathHistory
			m.jsonpathHistorySearch = ""
			m.jsonpathHistorySearching = false
			m.jsonpathHistoryCursor = 0
			m.loadFilteredBookmarks()
			return nil
		}
		// Otherwise, do nothing (no history navigation in input)
		return nil

	case "ctrl+w":
		// Delete word before cursor
		if m.filterCursor == 0 {
			return nil
		}
		// Find start of current word
		newCursor := m.filterCursor - 1
		for newCursor > 0 && m.filterInput[newCursor] == ' ' {
			newCursor--
		}
		for newCursor > 0 && m.filterInput[newCursor-1] != ' ' {
			newCursor--
		}
		m.filterInput = m.filterInput[:newCursor] + m.filterInput[m.filterCursor:]
		m.filterCursor = newCursor
		return nil
	}

	// Use registry for text input actions
	action, ok := m.keybinds.Match(keybinds.ContextTextInput, msg.String())
	if !ok {
		// Handle character input
		if len(msg.String()) == 1 {
			m.filterInput = m.filterInput[:m.filterCursor] + msg.String() + m.filterInput[m.filterCursor:]
			m.filterCursor++
		}
		return nil
	}

	switch action {
	case keybinds.ActionTextCancel:
		// Clear filter and return to normal mode
		m.mode = ModeNormal
		m.filterInput = ""
		m.filterCursor = 0
		m.filterActive = false
		m.filteredResponse = ""
		m.filterError = ""
		m.statusMsg = "Filter cancelled"

	case keybinds.ActionTextSubmit:
		// Apply filter
		if m.filterInput == "" {
			m.filterError = "Filter expression cannot be empty"
			return nil
		}

		if m.currentResponse == nil || m.currentResponse.Body == "" {
			m.filterError = "No response to filter"
			return nil
		}

		// Clear any previous error before applying
		m.filterError = ""

		// Apply the filter/query
		result, err := filter.Apply(m.currentResponse.Body, "", m.filterInput)
		if err != nil {
			// Keep modal open and show error
			m.filterError = fmt.Sprintf("Failed to apply filter: %s", err.Error())
			// Stay in ModeFilter to show the error
			return nil
		}

		// Store filtered result and show it
		m.filteredResponse = result
		m.filterActive = true
		m.filterError = ""
		m.mode = ModeNormal
		m.statusMsg = fmt.Sprintf("Filter applied: %s", m.filterInput)

		// Update response view to show filtered content
		m.updateResponseView()

	case keybinds.ActionTextMoveLeft:
		if m.filterCursor > 0 {
			m.filterCursor--
		}

	case keybinds.ActionTextMoveRight:
		if m.filterCursor < len(m.filterInput) {
			m.filterCursor++
		}

	case keybinds.ActionTextMoveHome:
		m.filterCursor = 0

	case keybinds.ActionTextMoveEnd:
		m.filterCursor = len(m.filterInput)

	case keybinds.ActionTextBackspace:
		if m.filterCursor > 0 {
			m.filterInput = m.filterInput[:m.filterCursor-1] + m.filterInput[m.filterCursor:]
			m.filterCursor--
		}

	case keybinds.ActionTextDelete:
		if m.filterCursor < len(m.filterInput) {
			m.filterInput = m.filterInput[:m.filterCursor] + m.filterInput[m.filterCursor+1:]
		}

	case keybinds.ActionTextClearBefore:
		// Clear from cursor to beginning (ctrl+u)
		m.filterInput = m.filterInput[m.filterCursor:]
		m.filterCursor = 0

	case keybinds.ActionTextClearAfter:
		// Clear from cursor to end (ctrl+k)
		m.filterInput = m.filterInput[:m.filterCursor]
	}

	return nil
}

// renderFilterModal renders the filter input modal
func (m *Model) renderFilterModal() string {
	var content strings.Builder

	content.WriteString("Filter Response with JMESPath\n\n")
	content.WriteString("Enter a JMESPath expression to filter/query the response.\n")
	content.WriteString("Examples:\n")
	content.WriteString("  items[?price > `100`]     - Filter items by condition\n")
	content.WriteString("  [].name                    - Extract all names\n")
	content.WriteString("  length(items)              - Count items\n")
	content.WriteString("  $(jq '.items[]')           - Use shell command (jq)\n\n")

	// Show input with cursor
	inputWithCursor := m.filterInput[:m.filterCursor] + "█" + m.filterInput[m.filterCursor:]
	wrappedInput := wrapText(inputWithCursor, 65)
	content.WriteString("Expression: " + wrappedInput + "\n")

	// Show error if any
	if m.filterError != "" {
		content.WriteString("\nError: ")
		errorWrapped := wrapText(m.filterError, 65)
		content.WriteString(errorWrapped)
	}

	// Validation feedback
	if m.filterInput != "" && m.filterError == "" {
		if filter.IsShellCommand(m.filterInput) {
			content.WriteString("\nShell command detected")
		} else if filter.IsValidJMESPath(m.filterInput) {
			content.WriteString("\nValid JMESPath expression")
		} else {
			content.WriteString("\nInvalid JMESPath (will show error on apply)")
		}
	}

	footer := "[Enter] apply • [Ctrl+S] save • [↑] history • [ESC] cancel"
	return m.renderModalWithFooter("Response Filter", content.String(), footer, 75, 18)
}
