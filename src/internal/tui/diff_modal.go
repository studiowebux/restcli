package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// handleDiffKeys handles keyboard input in diff mode
func (m *Model) handleDiffKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "q", "W":
		m.mode = ModeNormal

	// Vim-style navigation
	case "j", "down":
		m.diffView.LineDown(1)
	case "k", "up":
		m.diffView.LineUp(1)
	case "ctrl+d":
		m.diffView.HalfViewDown()
	case "ctrl+u":
		m.diffView.HalfViewUp()
	case "ctrl+f", "pgdown":
		m.diffView.ViewDown()
	case "ctrl+b", "pgup":
		m.diffView.ViewUp()
	case "g":
		if m.gPressed {
			m.diffView.GotoTop()
			m.gPressed = false
		} else {
			m.gPressed = true
		}
	case "G":
		m.diffView.GotoBottom()
		m.gPressed = false
	case "home":
		m.diffView.GotoTop()
	case "end":
		m.diffView.GotoBottom()

	default:
		m.gPressed = false
	}

	return nil
}

// renderDiffModal renders the diff comparison modal
func (m *Model) renderDiffModal() string {
	// Use nearly full screen
	modalWidth := m.width - 6
	modalHeight := m.height - 3

	diffView := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBlue).
		Width(modalWidth).
		Height(modalHeight).
		Padding(1, 2).
		Render(styleTitle.Render("Response Comparison") + "\n\n" + m.diffView.View())

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		diffView,
	)
}

// updateDiffView generates the diff content
func (m *Model) updateDiffView() {
	// Set viewport dimensions
	m.diffView.Width = m.width - 10
	m.diffView.Height = m.height - 7

	var content strings.Builder

	// Header
	pinnedName := "Pinned"
	if m.pinnedRequest != nil && m.pinnedRequest.Name != "" {
		pinnedName = m.pinnedRequest.Name
	}
	currentName := "Current"
	if m.currentRequest != nil && m.currentRequest.Name != "" {
		currentName = m.currentRequest.Name
	}

	content.WriteString(styleSuccess.Render(fmt.Sprintf("PINNED: %s", pinnedName)))
	content.WriteString(" vs ")
	content.WriteString(styleWarning.Render(fmt.Sprintf("CURRENT: %s", currentName)))
	content.WriteString("\n\n")

	// Status comparison
	content.WriteString(styleTitle.Render("Status:") + "\n")
	pinnedStatus := fmt.Sprintf("%d %s", m.pinnedResponse.Status, m.pinnedResponse.StatusText)
	currentStatus := fmt.Sprintf("%d %s", m.currentResponse.Status, m.currentResponse.StatusText)

	if pinnedStatus == currentStatus {
		content.WriteString(fmt.Sprintf("  %s\n", pinnedStatus))
	} else {
		content.WriteString(fmt.Sprintf("  - %s\n", styleError.Render(pinnedStatus)))
		content.WriteString(fmt.Sprintf("  + %s\n", styleSuccess.Render(currentStatus)))
	}
	content.WriteString("\n")

	// Duration comparison
	content.WriteString(styleTitle.Render("Duration:") + "\n")
	if m.pinnedResponse.Duration == m.currentResponse.Duration {
		content.WriteString(fmt.Sprintf("  %dms\n", m.pinnedResponse.Duration))
	} else {
		content.WriteString(fmt.Sprintf("  - %s\n", styleError.Render(fmt.Sprintf("%dms", m.pinnedResponse.Duration))))
		content.WriteString(fmt.Sprintf("  + %s\n", styleSuccess.Render(fmt.Sprintf("%dms", m.currentResponse.Duration))))
	}
	content.WriteString("\n")

	// Headers comparison
	content.WriteString(styleTitle.Render("Headers:") + "\n")
	headerDiff := compareHeaders(m.pinnedResponse.Headers, m.currentResponse.Headers)
	if headerDiff == "" {
		content.WriteString("  (no differences)\n")
	} else {
		content.WriteString(headerDiff)
	}
	content.WriteString("\n")

	// Body comparison
	content.WriteString(styleTitle.Render("Response Body:") + "\n")
	bodyDiff := compareTextLineByLine(m.pinnedResponse.Body, m.currentResponse.Body)
	content.WriteString(bodyDiff)

	content.WriteString("\n\n" + styleSubtle.Render("↑/↓ j/k: Scroll | gg/G: Top/Bottom | ESC/W: Close"))

	m.diffView.SetContent(content.String())
	m.diffView.GotoTop()
}

// compareHeaders generates a diff view for headers
func compareHeaders(pinned, current map[string]string) string {
	var result strings.Builder
	seen := make(map[string]bool)

	// Check all pinned headers
	for key, pinnedVal := range pinned {
		seen[key] = true
		if currentVal, exists := current[key]; exists {
			if pinnedVal != currentVal {
				result.WriteString(fmt.Sprintf("  %s:\n", key))
				result.WriteString(fmt.Sprintf("    - %s\n", styleError.Render(pinnedVal)))
				result.WriteString(fmt.Sprintf("    + %s\n", styleSuccess.Render(currentVal)))
			}
		} else {
			result.WriteString(fmt.Sprintf("  - %s: %s\n", styleError.Render(key), styleError.Render(pinnedVal)))
		}
	}

	// Check for new headers in current
	for key, currentVal := range current {
		if !seen[key] {
			result.WriteString(fmt.Sprintf("  + %s: %s\n", styleSuccess.Render(key), styleSuccess.Render(currentVal)))
		}
	}

	return result.String()
}

// compareTextLineByLine generates a unified diff view for text
func compareTextLineByLine(pinned, current string) string {
	pinnedLines := strings.Split(pinned, "\n")
	currentLines := strings.Split(current, "\n")

	// Simple line-by-line comparison (not a true diff algorithm, but good enough)
	var result strings.Builder
	maxLines := len(pinnedLines)
	if len(currentLines) > maxLines {
		maxLines = len(currentLines)
	}

	// Limit to first 100 lines to avoid overwhelming display
	if maxLines > 100 {
		maxLines = 100
		result.WriteString(styleSubtle.Render("  (showing first 100 lines)\n"))
	}

	differences := 0
	for i := 0; i < maxLines; i++ {
		pinnedLine := ""
		if i < len(pinnedLines) {
			pinnedLine = pinnedLines[i]
		}

		currentLine := ""
		if i < len(currentLines) {
			currentLine = currentLines[i]
		}

		if pinnedLine == currentLine && pinnedLine != "" {
			// Same line - show it with neutral color
			result.WriteString(fmt.Sprintf("  %s\n", pinnedLine))
		} else {
			if pinnedLine != "" {
				result.WriteString(fmt.Sprintf("  - %s\n", styleError.Render(pinnedLine)))
				differences++
			}
			if currentLine != "" {
				result.WriteString(fmt.Sprintf("  + %s\n", styleSuccess.Render(currentLine)))
				differences++
			}
		}
	}

	if differences == 0 && len(pinnedLines) > 0 && len(currentLines) > 0 {
		return "  (no differences)\n"
	}

	return result.String()
}
