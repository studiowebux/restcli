package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/studiowebux/restcli/internal/keybinds"
)

// initInteractiveVarPrompt initializes the state for the current variable prompt
func (m *Model) initInteractiveVarPrompt() {
	if len(m.interactiveVarNames) == 0 {
		return
	}

	currentVar := m.interactiveVarNames[0]
	profile := m.sessionMgr.GetActiveProfile()

	// Check if this is a multi-value variable
	if profile != nil {
		if varValue, exists := profile.Variables[currentVar]; exists && varValue.IsMultiValue() {
			// Set up selection mode
			m.interactiveVarMode = "select"
			mv := varValue.MultiValue

			// Extract options
			m.interactiveVarOptions = make([]string, len(mv.Options))
			copy(m.interactiveVarOptions, mv.Options)

			// Extract aliases (index -> alias names)
			m.interactiveVarAliases = make(map[int][]string)
			if mv.Aliases != nil {
				for aliasName, idx := range mv.Aliases {
					m.interactiveVarAliases[idx] = append(m.interactiveVarAliases[idx], aliasName)
				}
			}

			// Get active index
			m.interactiveVarActiveIdx = mv.Active

			// Start selection at active index
			m.interactiveVarSelectIdx = m.interactiveVarActiveIdx
			if m.interactiveVarSelectIdx >= len(m.interactiveVarOptions) {
				m.interactiveVarSelectIdx = 0
			}
		} else {
			// Regular text input mode
			m.interactiveVarMode = "input"
			m.interactiveVarOptions = nil
			m.interactiveVarAliases = nil
		}
	} else {
		// No profile, default to input mode
		m.interactiveVarMode = "input"
		m.interactiveVarOptions = nil
		m.interactiveVarAliases = nil
	}
}

// handleInteractiveVariablePromptKeys handles key input for interactive variable prompt modal
func (m *Model) handleInteractiveVariablePromptKeys(msg tea.KeyMsg) tea.Cmd {
	// Handle select mode keys first (special - dual key bindings)
	if m.interactiveVarMode == "select" {
		switch msg.String() {
		case "up", "k":
			if m.interactiveVarSelectIdx > 0 {
				m.interactiveVarSelectIdx--
			}
			return nil
		case "down", "j":
			if m.interactiveVarSelectIdx < len(m.interactiveVarOptions)-1 {
				m.interactiveVarSelectIdx++
			}
			return nil
		case "home", "g":
			m.interactiveVarSelectIdx = 0
			return nil
		case "end", "G":
			m.interactiveVarSelectIdx = len(m.interactiveVarOptions) - 1
			return nil
		case "c", "C":
			// Switch to custom input mode
			m.interactiveVarMode = "input"
			m.interactiveVarInput = ""
			m.interactiveVarCursor = 0
			return nil
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			// Quick select by number
			num := int(msg.String()[0] - '0')
			index := num - 1
			if index >= 0 && index < len(m.interactiveVarOptions) {
				m.interactiveVarSelectIdx = index
				// Auto-submit on quick select
				return m.submitInteractiveVar()
			}
			return nil
		}
	}

	// Handle common keys (esc, enter)
	switch msg.String() {
	case "esc":
		// Cancel - clear the queue and go back to normal mode
		m.interactiveVarNames = nil
		m.interactiveVarValues = nil
		m.mode = ModeNormal
		m.statusMsg = "Interactive variable prompt cancelled"
		return nil

	case "enter":
		return m.submitInteractiveVar()
	}

	// Text input keys only apply in input mode
	if m.interactiveVarMode == "input" {
		// Handle ctrl+w specially (delete word - not in registry)
		if msg.String() == "ctrl+w" {
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
		}

		// Use registry for text input actions
		action, ok := m.keybinds.Match(keybinds.ContextTextInput, msg.String())
		if !ok {
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

		switch action {
		case keybinds.ActionTextMoveLeft:
			if m.interactiveVarCursor > 0 {
				m.interactiveVarCursor--
			}

		case keybinds.ActionTextMoveRight:
			if m.interactiveVarCursor < len(m.interactiveVarInput) {
				m.interactiveVarCursor++
			}

		case keybinds.ActionTextMoveHome:
			m.interactiveVarCursor = 0

		case keybinds.ActionTextMoveEnd:
			m.interactiveVarCursor = len(m.interactiveVarInput)

		case keybinds.ActionTextBackspace:
			if m.interactiveVarCursor > 0 {
				m.interactiveVarInput = m.interactiveVarInput[:m.interactiveVarCursor-1] + m.interactiveVarInput[m.interactiveVarCursor:]
				m.interactiveVarCursor--
			}

		case keybinds.ActionTextDelete:
			if m.interactiveVarCursor < len(m.interactiveVarInput) {
				m.interactiveVarInput = m.interactiveVarInput[:m.interactiveVarCursor] + m.interactiveVarInput[m.interactiveVarCursor+1:]
			}

		case keybinds.ActionTextClearBefore:
			// Clear from cursor to beginning (ctrl+u)
			m.interactiveVarInput = m.interactiveVarInput[m.interactiveVarCursor:]
			m.interactiveVarCursor = 0

		case keybinds.ActionTextClearAfter:
			// Clear from cursor to end (ctrl+k)
			m.interactiveVarInput = m.interactiveVarInput[:m.interactiveVarCursor]
		}
	}

	return nil
}

// submitInteractiveVar saves the current variable value and moves to the next one
func (m *Model) submitInteractiveVar() tea.Cmd {
	if len(m.interactiveVarNames) == 0 {
		return nil
	}

	// Save the current value
	currentVar := m.interactiveVarNames[0]
	if m.interactiveVarValues == nil {
		m.interactiveVarValues = make(map[string]string)
	}

	// Get value based on mode
	var value string
	if m.interactiveVarMode == "select" && m.interactiveVarSelectIdx >= 0 && m.interactiveVarSelectIdx < len(m.interactiveVarOptions) {
		value = m.interactiveVarOptions[m.interactiveVarSelectIdx]
	} else {
		value = m.interactiveVarInput
	}
	m.interactiveVarValues[currentVar] = value

	// Move to next variable or finish
	m.interactiveVarNames = m.interactiveVarNames[1:]
	m.interactiveVarInput = ""
	m.interactiveVarCursor = 0

	if len(m.interactiveVarNames) == 0 {
		// Done with all prompts - proceed with request execution
		m.mode = ModeNormal
		return m.executeRequestWithInteractiveVars()
	}

	// More variables to prompt for - initialize next prompt
	m.initInteractiveVarPrompt()
	return nil
}

// renderInteractiveVariablePrompt renders the interactive variable prompt modal
func (m *Model) renderInteractiveVariablePrompt() string {
	if len(m.interactiveVarNames) == 0 {
		m.mode = ModeNormal
		return m.View()
	}

	currentVar := m.interactiveVarNames[0]
	profile := m.sessionMgr.GetActiveProfile()

	// Get variable info for title
	var defaultValue string
	if profile != nil {
		if varValue, exists := profile.Variables[currentVar]; exists {
			defaultValue = varValue.GetValue()
		}
	}

	progress := fmt.Sprintf("(%d/%d)",
		len(m.interactiveVarValues)+1,
		len(m.interactiveVarValues)+len(m.interactiveVarNames))

	// Build title with progress
	title := fmt.Sprintf("Interactive Variable %s", progress)

	// Render based on mode
	if m.interactiveVarMode == "select" {
		// Header content (non-scrollable)
		var content strings.Builder
		content.WriteString(fmt.Sprintf("Variable: %s\n", currentVar))
		if defaultValue != "" {
			content.WriteString(fmt.Sprintf("Default: %s\n", truncate(defaultValue, 50)))
		}
		content.WriteString("\n")

		// Only options scroll
		selectedLine := 0 // First option is line 0 in scrollable area
		for i := 0; i < len(m.interactiveVarOptions); i++ {
			if i == m.interactiveVarSelectIdx {
				selectedLine = i
			}

			opt := m.interactiveVarOptions[i]
			prefix := "  "
			suffix := ""

			// Show selection indicator
			if i == m.interactiveVarSelectIdx {
				prefix = "> "
			}

			// Show active indicator
			if i == m.interactiveVarActiveIdx {
				suffix = " [active]"
			}

			// Show alias if available
			if aliases, ok := m.interactiveVarAliases[i]; ok && len(aliases) > 0 {
				suffix += fmt.Sprintf(" (alias: %s)", strings.Join(aliases, ", "))
			}

			// Show number for quick select (1-9)
			displayNum := i + 1
			if displayNum <= 9 {
				content.WriteString(fmt.Sprintf("%s%d. %s%s\n", prefix, displayNum, truncate(opt, 45), suffix))
			} else {
				content.WriteString(fmt.Sprintf("%s%s%s\n", prefix, truncate(opt, 50), suffix))
			}
		}

		// Add counter at end if many options
		if len(m.interactiveVarOptions) > 8 {
			content.WriteString(fmt.Sprintf("\n[%d/%d options]", m.interactiveVarSelectIdx+1, len(m.interactiveVarOptions)))
		}

		footer := "↑/↓/j/k: navigate • enter: select • c: custom value • esc: cancel"
		// Add header lines to selectedLine for proper scrolling (Variable, Default if present, blank line)
		headerLines := 2
		if defaultValue != "" {
			headerLines = 3
		}
		return m.renderModalWithFooterAndScroll(title, content.String(), footer, 80, 20, selectedLine+headerLines)
	}

	// Input mode - show text input
	var inputContent strings.Builder
	inputContent.WriteString(fmt.Sprintf("Variable: %s\n", currentVar))
	if defaultValue != "" {
		inputContent.WriteString(fmt.Sprintf("Default: %s\n", truncate(defaultValue, 50)))
	}
	inputContent.WriteString("\n")

	inputWithCursor := addCursorAt(m.interactiveVarInput, m.interactiveVarCursor)
	wrappedValue := wrapText(inputWithCursor, 55)
	inputContent.WriteString(fmt.Sprintf("Value: %s", wrappedValue))

	if m.interactiveVarOptions != nil && len(m.interactiveVarOptions) > 0 {
		inputContent.WriteString("\n\nEnter custom value or ESC to cancel")
	} else {
		inputContent.WriteString("\n\nEnter value and press Enter to continue, ESC to cancel")
	}

	return m.renderModal(title, inputContent.String(), 70, 12)
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
