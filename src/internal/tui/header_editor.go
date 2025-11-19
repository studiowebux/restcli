package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// getSortedHeaderNames returns sorted header names
func getSortedHeaderNames(headers map[string]string) []string {
	names := make([]string, 0, len(headers))
	for name := range headers {
		names = append(names, name)
	}
	// Simple bubble sort
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			if names[i] > names[j] {
				names[i], names[j] = names[j], names[i]
			}
		}
	}
	return names
}

// renderHeaderEditor renders the header editor in its various modes
func (m *Model) renderHeaderEditor() string {
	profile := m.sessionMgr.GetActiveProfile()

	var content strings.Builder
	var footer string

	switch m.mode {
	case ModeHeaderList:
		content.WriteString("Profile Headers:\n")

		if len(profile.Headers) == 0 {
			content.WriteString("  (none)\n")
		} else {
			sortedNames := getSortedHeaderNames(profile.Headers)
			for i, name := range sortedNames {
				line := fmt.Sprintf("  %s: %s", name, truncate(profile.Headers[name], 50))
				if i == m.headerEditIndex {
					line = styleSelected.Render(line)
				}
				content.WriteString(line + "\n")
			}
		}

		footer = "[a]dd [e]dit [d]elete [ESC]"

	case ModeHeaderAdd:
		content.WriteString("Add Header\n\n")
		nameField := m.headerEditName
		valueField := m.headerEditValue
		if m.headerEditCursor == 0 {
			nameField = addCursor(nameField)
		} else {
			valueField = addCursor(valueField)
		}
		content.WriteString("Name:  " + nameField + "\n")
		content.WriteString("Value: " + valueField + "\n")
		footer = "[TAB] switch fields [Enter] save [ESC] cancel"

	case ModeHeaderEdit:
		content.WriteString("Edit Header\n\n")
		nameField := m.headerEditName
		valueField := m.headerEditValue
		if m.headerEditCursor == 0 {
			nameField = addCursor(nameField)
		} else {
			valueField = addCursor(valueField)
		}
		content.WriteString("Name:  " + nameField + "\n")
		content.WriteString("Value: " + valueField + "\n")
		footer = "[TAB] switch fields [Enter] save [ESC] cancel"

	case ModeHeaderDelete:
		content.WriteString("Delete Header\n\n")
		content.WriteString(fmt.Sprintf("Are you sure you want to delete '%s'?", m.headerEditName))
		footer = "[y]es [n]o"
	}

	// Calculate selected line for auto-scroll (only in list mode)
	selectedLine := -1
	if m.mode == ModeHeaderList {
		// Line 0 is "Profile Headers:", selected item is at line 1 + index
		selectedLine = 1 + m.headerEditIndex
	}

	return m.renderModalWithFooterAndScroll("Headers", content.String(), footer, 70, 20, selectedLine)
}

// handleHeaderEditorKeys handles keyboard input in header editor modes
func (m *Model) handleHeaderEditorKeys(msg tea.KeyMsg) tea.Cmd {
	profile := m.sessionMgr.GetActiveProfile()

	switch m.mode {
	case ModeHeaderList:
		sortedNames := getSortedHeaderNames(profile.Headers)

		switch msg.String() {
		case "esc":
			m.mode = ModeNormal

		case "up", "k":
			if m.headerEditIndex > 0 {
				m.headerEditIndex--
			}

		case "down", "j":
			if m.headerEditIndex < len(sortedNames)-1 {
				m.headerEditIndex++
			}

		// Page up/down - move cursor by page amount
		case "pgup":
			pageSize := m.modalView.Height
			if pageSize < 1 {
				pageSize = 10
			}
			m.headerEditIndex -= pageSize
			if m.headerEditIndex < 0 {
				m.headerEditIndex = 0
			}

		case "pgdown":
			pageSize := m.modalView.Height
			if pageSize < 1 {
				pageSize = 10
			}
			m.headerEditIndex += pageSize
			if m.headerEditIndex >= len(sortedNames) {
				m.headerEditIndex = len(sortedNames) - 1
			}
			if m.headerEditIndex < 0 {
				m.headerEditIndex = 0
			}

		case "home":
			m.headerEditIndex = 0

		case "end":
			if len(sortedNames) > 0 {
				m.headerEditIndex = len(sortedNames) - 1
			}

		case "g":
			// Vim-style 'gg' to go to top
			if m.gPressed {
				m.gPressed = false
				m.headerEditIndex = 0
			} else {
				m.gPressed = true
			}
			return nil // Don't reset gPressed

		case "G":
			// Vim-style 'G' to go to bottom
			if len(sortedNames) > 0 {
				m.headerEditIndex = len(sortedNames) - 1
			}

		case "a":
			m.mode = ModeHeaderAdd
			m.headerEditName = ""
			m.headerEditValue = ""
			m.headerEditCursor = 0

		case "e":
			if len(sortedNames) > 0 && m.headerEditIndex < len(sortedNames) {
				m.mode = ModeHeaderEdit
				name := sortedNames[m.headerEditIndex]
				m.headerEditName = name
				m.headerEditValue = profile.Headers[name]
				m.headerEditCursor = 0
			}

		case "d":
			if len(sortedNames) > 0 && m.headerEditIndex < len(sortedNames) {
				m.mode = ModeHeaderDelete
				m.headerEditName = sortedNames[m.headerEditIndex]
			}
		}

	case ModeHeaderAdd, ModeHeaderEdit:
		return m.handleHeaderInputKeys(msg)

	case ModeHeaderDelete:
		switch msg.String() {
		case "y":
			// Delete the header
			delete(profile.Headers, m.headerEditName)
			m.sessionMgr.SaveProfiles()
			m.mode = ModeHeaderList
			m.statusMsg = fmt.Sprintf("Deleted header: %s", m.headerEditName)

		case "n", "esc":
			m.mode = ModeHeaderList
		}
	}

	// Reset 'g' state on any key except 'g' itself
	m.gPressed = false
	return nil
}

// handleHeaderInputKeys handles text input for add/edit header
func (m *Model) handleHeaderInputKeys(msg tea.KeyMsg) tea.Cmd {
	profile := m.sessionMgr.GetActiveProfile()

	switch msg.String() {
	case "esc":
		m.mode = ModeHeaderList

	case "tab":
		m.headerEditCursor = (m.headerEditCursor + 1) % 2

	case "enter":
		if m.headerEditName == "" {
			m.errorMsg = "Header name cannot be empty"
			return nil
		}

		// Create or update header
		if profile.Headers == nil {
			profile.Headers = make(map[string]string)
		}

		profile.Headers[m.headerEditName] = m.headerEditValue

		m.sessionMgr.SaveProfiles()
		m.mode = ModeHeaderList
		m.statusMsg = fmt.Sprintf("Saved header: %s", m.headerEditName)

	default:
		// Handle common text input operations (paste, clear, backspace)
		if m.headerEditCursor == 0 {
			if _, shouldContinue := handleTextInput(&m.headerEditName, msg); shouldContinue {
				return nil
			}
			// Append character to name
			if len(msg.String()) == 1 {
				m.headerEditName += msg.String()
			}
		} else {
			if _, shouldContinue := handleTextInput(&m.headerEditValue, msg); shouldContinue {
				return nil
			}
			// Append character to value
			if len(msg.String()) == 1 {
				m.headerEditValue += msg.String()
			}
		}
	}

	return nil
}
