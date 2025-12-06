package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// handleInteractiveVariablePromptKeys handles key input for interactive variable prompt modal
func (m *Model) handleInteractiveVariablePromptKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		// Cancel - clear the queue and go back to normal mode
		m.interactiveVarNames = nil
		m.interactiveVarValues = nil
		m.mode = ModeNormal
		m.statusMsg = "Interactive variable prompt cancelled"
		return nil

	case "enter":
		if len(m.interactiveVarNames) == 0 {
			return nil
		}

		// Save the current value
		currentVar := m.interactiveVarNames[0]
		if m.interactiveVarValues == nil {
			m.interactiveVarValues = make(map[string]string)
		}
		m.interactiveVarValues[currentVar] = m.interactiveVarInput

		// Move to next variable or finish
		m.interactiveVarNames = m.interactiveVarNames[1:]
		m.interactiveVarInput = ""
		m.interactiveVarCursor = 0

		if len(m.interactiveVarNames) == 0 {
			// Done with all prompts - proceed with request execution
			m.mode = ModeNormal
			return m.executeRequestWithInteractiveVars()
		}

		// More variables to prompt for
		return nil

	case "left":
		if m.interactiveVarCursor > 0 {
			m.interactiveVarCursor--
		}
		return nil

	case "right":
		if m.interactiveVarCursor < len(m.interactiveVarInput) {
			m.interactiveVarCursor++
		}
		return nil

	case "home", "ctrl+a":
		m.interactiveVarCursor = 0
		return nil

	case "end", "ctrl+e":
		m.interactiveVarCursor = len(m.interactiveVarInput)
		return nil

	case "backspace":
		if m.interactiveVarCursor > 0 {
			m.interactiveVarInput = m.interactiveVarInput[:m.interactiveVarCursor-1] + m.interactiveVarInput[m.interactiveVarCursor:]
			m.interactiveVarCursor--
		}
		return nil

	case "delete":
		if m.interactiveVarCursor < len(m.interactiveVarInput) {
			m.interactiveVarInput = m.interactiveVarInput[:m.interactiveVarCursor] + m.interactiveVarInput[m.interactiveVarCursor+1:]
		}
		return nil

	case "ctrl+u":
		// Clear from cursor to beginning
		m.interactiveVarInput = m.interactiveVarInput[m.interactiveVarCursor:]
		m.interactiveVarCursor = 0
		return nil

	case "ctrl+k":
		// Clear from cursor to end
		m.interactiveVarInput = m.interactiveVarInput[:m.interactiveVarCursor]
		return nil

	case "ctrl+w":
		// Delete word before cursor
		if m.interactiveVarCursor == 0 {
			return nil
		}
		// Find start of current word
		newCursor := m.interactiveVarCursor - 1
		for newCursor > 0 && m.interactiveVarInput[newCursor] == ' ' {
			newCursor--
		}
		for newCursor > 0 && m.interactiveVarInput[newCursor-1] != ' ' {
			newCursor--
		}
		m.interactiveVarInput = m.interactiveVarInput[:newCursor] + m.interactiveVarInput[m.interactiveVarCursor:]
		m.interactiveVarCursor = newCursor
		return nil

	default:
		// Handle character input
		if len(msg.String()) == 1 {
			// Insert character at cursor position
			m.interactiveVarInput = m.interactiveVarInput[:m.interactiveVarCursor] +
				msg.String() +
				m.interactiveVarInput[m.interactiveVarCursor:]
			m.interactiveVarCursor++
		}
		return nil
	}
}

// renderInteractiveVariablePrompt renders the interactive variable prompt modal
func (m *Model) renderInteractiveVariablePrompt() string {
	if len(m.interactiveVarNames) == 0 {
		m.mode = ModeNormal
		return m.View()
	}

	currentVar := m.interactiveVarNames[0]
	profile := m.sessionMgr.GetActiveProfile()

	// Get variable info
	varInfo := ""
	if profile != nil {
		if varValue, exists := profile.Variables[currentVar]; exists {
			defaultValue := varValue.GetValue()
			if defaultValue != "" {
				varInfo = fmt.Sprintf("\nDefault value: %s", truncate(defaultValue, 50))
			}
			if varValue.IsMultiValue() {
				varInfo += " (multi-value)"
			}
		}
	}

	progress := fmt.Sprintf("(%d/%d)",
		len(m.interactiveVarValues)+1,
		len(m.interactiveVarValues)+len(m.interactiveVarNames))

	content := fmt.Sprintf("Interactive Variable %s\n\n", progress)
	content += fmt.Sprintf("Variable: %s%s\n\n", currentVar, varInfo)

	// Add cursor and wrap the input value to prevent overflow
	// Modal width is 70, minus padding/border (4) = 66, minus content padding (4) = 62
	// "Value: " takes 7 chars, leaving ~55 chars for wrapped input
	inputWithCursor := addCursorAt(m.interactiveVarInput, m.interactiveVarCursor)
	wrappedValue := wrapText(inputWithCursor, 55)
	content += fmt.Sprintf("Value: %s", wrappedValue)
	content += "\n\nEnter value and press Enter to continue, ESC to cancel"

	return m.renderModal("Interactive Variable Prompt", content, 70, 12)
}

// addCursorAt adds a cursor indicator at the specified position
func addCursorAt(s string, pos int) string {
	if pos < 0 {
		pos = 0
	}
	if pos > len(s) {
		pos = len(s)
	}
	if pos == len(s) {
		return s + "█"
	}
	return s[:pos] + "█" + s[pos:]
}
