package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/studiowebux/restcli/internal/keybinds"
)

// handleBodyOverrideKeys handles key input for body override modal
func (m *Model) handleBodyOverrideKeys(msg tea.KeyMsg) tea.Cmd {
	// Handle special keys not in registry (multiline editor with custom behavior)
	switch msg.String() {
	case "ctrl+s", "ctrl+enter":
		// Save and return to normal mode
		m.bodyOverride = m.bodyOverrideInput
		m.mode = ModeNormal
		m.statusMsg = "Body override applied (will be used for next request)"
		return nil

	case "up":
		// Move cursor to previous line
		lines := strings.Split(m.bodyOverrideInput[:m.bodyOverrideCursor], "\n")
		if len(lines) > 1 {
			currentLinePos := len(lines[len(lines)-1])
			prevLineStart := m.bodyOverrideCursor - currentLinePos - 1
			if prevLineStart >= 0 {
				prevLines := strings.Split(m.bodyOverrideInput[:prevLineStart], "\n")
				if len(prevLines) > 0 {
					prevLine := prevLines[len(prevLines)-1]
					newPos := prevLineStart - len(prevLine)
					if currentLinePos <= len(prevLine) {
						newPos += currentLinePos
					} else {
						newPos += len(prevLine)
					}
					if newPos >= 0 && newPos <= len(m.bodyOverrideInput) {
						m.bodyOverrideCursor = newPos
					}
				}
			}
		}
		return nil

	case "down":
		// Move cursor to next line
		remaining := m.bodyOverrideInput[m.bodyOverrideCursor:]
		if idx := strings.Index(remaining, "\n"); idx != -1 {
			lines := strings.Split(m.bodyOverrideInput[:m.bodyOverrideCursor], "\n")
			currentLinePos := 0
			if len(lines) > 0 {
				currentLinePos = len(lines[len(lines)-1])
			}

			nextLineStart := m.bodyOverrideCursor + idx + 1
			nextLineEnd := len(m.bodyOverrideInput)
			if nextIdx := strings.Index(m.bodyOverrideInput[nextLineStart:], "\n"); nextIdx != -1 {
				nextLineEnd = nextLineStart + nextIdx
			}

			nextLineLen := nextLineEnd - nextLineStart
			newPos := nextLineStart
			if currentLinePos <= nextLineLen {
				newPos += currentLinePos
			} else {
				newPos += nextLineLen
			}

			if newPos <= len(m.bodyOverrideInput) {
				m.bodyOverrideCursor = newPos
			}
		}
		return nil

	case "enter":
		// Insert newline
		m.bodyOverrideInput = m.bodyOverrideInput[:m.bodyOverrideCursor] + "\n" + m.bodyOverrideInput[m.bodyOverrideCursor:]
		m.bodyOverrideCursor++
		return nil

	case "tab":
		// Insert 2 spaces
		m.bodyOverrideInput = m.bodyOverrideInput[:m.bodyOverrideCursor] + "  " + m.bodyOverrideInput[m.bodyOverrideCursor:]
		m.bodyOverrideCursor += 2
		return nil
	}

	// Use registry for text input actions
	action, ok := m.keybinds.Match(keybinds.ContextTextInput, msg.String())
	if !ok {
		// Handle regular character input
		if len(msg.String()) == 1 {
			m.bodyOverrideInput = m.bodyOverrideInput[:m.bodyOverrideCursor] + msg.String() + m.bodyOverrideInput[m.bodyOverrideCursor:]
			m.bodyOverrideCursor++
		}
		return nil
	}

	switch action {
	case keybinds.ActionTextCancel:
		// Cancel - discard changes
		m.mode = ModeNormal
		m.bodyOverrideInput = ""
		m.bodyOverrideCursor = 0
		m.statusMsg = "Body override cancelled"

	case keybinds.ActionTextMoveLeft:
		if m.bodyOverrideCursor > 0 {
			m.bodyOverrideCursor--
		}

	case keybinds.ActionTextMoveRight:
		if m.bodyOverrideCursor < len(m.bodyOverrideInput) {
			m.bodyOverrideCursor++
		}

	case keybinds.ActionTextMoveHome:
		// Move to start of current line
		lines := strings.Split(m.bodyOverrideInput[:m.bodyOverrideCursor], "\n")
		if len(lines) > 0 {
			currentLinePos := len(lines[len(lines)-1])
			m.bodyOverrideCursor -= currentLinePos
		}

	case keybinds.ActionTextMoveEnd:
		// Move to end of current line
		remaining := m.bodyOverrideInput[m.bodyOverrideCursor:]
		if idx := strings.Index(remaining, "\n"); idx != -1 {
			m.bodyOverrideCursor += idx
		} else {
			m.bodyOverrideCursor = len(m.bodyOverrideInput)
		}

	case keybinds.ActionTextBackspace:
		if m.bodyOverrideCursor > 0 {
			m.bodyOverrideInput = m.bodyOverrideInput[:m.bodyOverrideCursor-1] + m.bodyOverrideInput[m.bodyOverrideCursor:]
			m.bodyOverrideCursor--
		}

	case keybinds.ActionTextDelete:
		if m.bodyOverrideCursor < len(m.bodyOverrideInput) {
			m.bodyOverrideInput = m.bodyOverrideInput[:m.bodyOverrideCursor] + m.bodyOverrideInput[m.bodyOverrideCursor+1:]
		}

	case keybinds.ActionTextClearBefore:
		// Clear from cursor to start of line (ctrl+u)
		lines := strings.Split(m.bodyOverrideInput[:m.bodyOverrideCursor], "\n")
		if len(lines) > 0 {
			currentLinePos := len(lines[len(lines)-1])
			m.bodyOverrideInput = m.bodyOverrideInput[:m.bodyOverrideCursor-currentLinePos] + m.bodyOverrideInput[m.bodyOverrideCursor:]
			m.bodyOverrideCursor -= currentLinePos
		}

	case keybinds.ActionTextClearAfter:
		// Clear from cursor to end of line (ctrl+k)
		remaining := m.bodyOverrideInput[m.bodyOverrideCursor:]
		if idx := strings.Index(remaining, "\n"); idx != -1 {
			m.bodyOverrideInput = m.bodyOverrideInput[:m.bodyOverrideCursor] + remaining[idx:]
		} else {
			m.bodyOverrideInput = m.bodyOverrideInput[:m.bodyOverrideCursor]
		}
	}

	return nil
}

// renderBodyOverrideModal renders the body override editor modal
func (m *Model) renderBodyOverrideModal() string {
	var content strings.Builder

	content.WriteString("Edit Request Body (one-time override)\n\n")

	// Show validation status for JSON
	var validationMsg string
	if m.bodyOverrideInput != "" {
		var jsonData interface{}
		if err := json.Unmarshal([]byte(m.bodyOverrideInput), &jsonData); err == nil {
			validationMsg = "Valid JSON"
		} else {
			validationMsg = fmt.Sprintf("JSON Error: %s", err.Error())
		}
	}

	// Display editable content with cursor
	// For multi-line editor, show a portion around cursor
	const displayLines = 15
	const displayWidth = 70

	lines := strings.Split(m.bodyOverrideInput, "\n")
	cursorLine := 0
	cursorCol := 0
	charCount := 0

	// Find cursor position in terms of line and column
	for i, line := range lines {
		lineLen := len(line) + 1 // +1 for newline
		if charCount+lineLen > m.bodyOverrideCursor {
			cursorLine = i
			cursorCol = m.bodyOverrideCursor - charCount
			break
		}
		charCount += lineLen
	}

	// Calculate visible range centered on cursor
	startLine := cursorLine - displayLines/2
	if startLine < 0 {
		startLine = 0
	}
	endLine := startLine + displayLines
	if endLine > len(lines) {
		endLine = len(lines)
		startLine = endLine - displayLines
		if startLine < 0 {
			startLine = 0
		}
	}

	// Render visible lines
	for i := startLine; i < endLine; i++ {
		line := lines[i]
		if i == cursorLine {
			// Insert cursor on current line
			if cursorCol <= len(line) {
				line = line[:cursorCol] + "█" + line[cursorCol:]
			} else {
				line += "█"
			}
		}

		// Truncate long lines
		if len(line) > displayWidth {
			line = line[:displayWidth-3] + "..."
		}

		content.WriteString(fmt.Sprintf("%3d │ %s\n", i+1, line))
	}

	if len(lines) > displayLines {
		content.WriteString(fmt.Sprintf("\n[Showing lines %d-%d of %d]", startLine+1, endLine, len(lines)))
	}

	if validationMsg != "" {
		content.WriteString("\n\n" + validationMsg)
	}

	footer := "[Ctrl+S/Ctrl+Enter] save • [ESC] cancel"
	return m.renderModalWithFooter("Body Override", content.String(), footer, 80, 25)
}
