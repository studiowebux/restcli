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

	case "tab":
		// Toggle between unified and split view
		if m.diffViewMode == "split" {
			m.diffViewMode = "unified"
		} else {
			m.diffViewMode = "split"
		}
		m.updateDiffView() // Regenerate content for new mode

	// Vim-style navigation
	case "j", "down":
		if m.diffViewMode == "split" {
			// Synchronized scrolling for split view
			m.diffLeftView.LineDown(1)
			m.diffRightView.LineDown(1)
		} else {
			m.diffView.LineDown(1)
		}
	case "k", "up":
		if m.diffViewMode == "split" {
			// Synchronized scrolling for split view
			m.diffLeftView.LineUp(1)
			m.diffRightView.LineUp(1)
		} else {
			m.diffView.LineUp(1)
		}
	case "ctrl+d":
		if m.diffViewMode == "split" {
			m.diffLeftView.HalfViewDown()
			m.diffRightView.HalfViewDown()
		} else {
			m.diffView.HalfViewDown()
		}
	case "ctrl+u":
		if m.diffViewMode == "split" {
			m.diffLeftView.HalfViewUp()
			m.diffRightView.HalfViewUp()
		} else {
			m.diffView.HalfViewUp()
		}
	case "ctrl+f", "pgdown":
		if m.diffViewMode == "split" {
			m.diffLeftView.ViewDown()
			m.diffRightView.ViewDown()
		} else {
			m.diffView.ViewDown()
		}
	case "ctrl+b", "pgup":
		if m.diffViewMode == "split" {
			m.diffLeftView.ViewUp()
			m.diffRightView.ViewUp()
		} else {
			m.diffView.ViewUp()
		}
	case "g":
		if m.gPressed {
			if m.diffViewMode == "split" {
				m.diffLeftView.GotoTop()
				m.diffRightView.GotoTop()
			} else {
				m.diffView.GotoTop()
			}
			m.gPressed = false
		} else {
			m.gPressed = true
		}
	case "G":
		if m.diffViewMode == "split" {
			m.diffLeftView.GotoBottom()
			m.diffRightView.GotoBottom()
		} else {
			m.diffView.GotoBottom()
		}
		m.gPressed = false
	case "home":
		if m.diffViewMode == "split" {
			m.diffLeftView.GotoTop()
			m.diffRightView.GotoTop()
		} else {
			m.diffView.GotoTop()
		}
	case "end":
		if m.diffViewMode == "split" {
			m.diffLeftView.GotoBottom()
			m.diffRightView.GotoBottom()
		} else {
			m.diffView.GotoBottom()
		}

	default:
		m.gPressed = false
	}

	return nil
}

// renderDiffSplitView renders the split pane view
func (m *Model) renderDiffSplitView(modalWidth, modalHeight int) string {
	// Calculate pane dimensions
	paneWidth := (modalWidth - 3) / 2 // 50/50 split with separator
	paneHeight := modalHeight - 6     // Account for metadata header and padding

	// Metadata header
	metadata := m.renderDiffMetadata()

	// Create left pane (pinned)
	leftPane := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorGray).
		Width(paneWidth).
		Height(paneHeight).
		Padding(0, 1).
		Render(styleTitle.Render("PINNED") + "\n" + m.diffLeftView.View())

	// Create right pane (current)
	rightPane := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorGreen).
		Width(paneWidth).
		Height(paneHeight).
		Padding(0, 1).
		Render(styleTitle.Render("CURRENT") + "\n" + m.diffRightView.View())

	// Join panes horizontally
	splitPanes := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPane,
		rightPane,
	)

	// Combine metadata and split panes
	return lipgloss.JoinVertical(
		lipgloss.Left,
		metadata,
		"\n",
		splitPanes,
		"\n"+styleSubtle.Render("Tab: Toggle View | ↑/↓ j/k: Scroll | gg/G: Top/Bottom | ESC/W: Close"),
	)
}

// renderDiffModal renders the diff comparison modal
func (m *Model) renderDiffModal() string {
	// Use nearly full screen
	modalWidth := m.width - 6
	modalHeight := m.height - 3

	var content string

	// Check view mode
	if m.diffViewMode == "split" {
		content = m.renderDiffSplitView(modalWidth, modalHeight)
	} else {
		// Unified diff view (default)
		diffView := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBlue).
			Width(modalWidth).
			Height(modalHeight).
			Padding(1, 2).
			Render(styleTitle.Render("Response Comparison") + "\n\n" + m.diffView.View())

		content = diffView
	}

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

// updateDiffView generates the diff content
func (m *Model) updateDiffView() {
	// Initialize view mode if not set
	if m.diffViewMode == "" {
		m.diffViewMode = "unified"
	}

	if m.diffViewMode == "split" {
		// Split view mode - populate left and right viewports
		modalWidth := m.width - 6
		modalHeight := m.height - 3
		paneWidth := (modalWidth - 3) / 2
		paneHeight := modalHeight - 6

		// Set viewport dimensions
		m.diffLeftView.Width = paneWidth - 4
		m.diffLeftView.Height = paneHeight - 2
		m.diffRightView.Width = paneWidth - 4
		m.diffRightView.Height = paneHeight - 2

		// Generate diff-styled content with background highlighting
		leftContent, rightContent := compareTextSplitView(
			m.pinnedResponse.Body,
			m.currentResponse.Body,
			paneWidth-6,
		)

		// Set styled content for both viewports
		m.diffLeftView.SetContent(leftContent)
		m.diffRightView.SetContent(rightContent)

		// Reset scroll position
		m.diffLeftView.GotoTop()
		m.diffRightView.GotoTop()
	} else {
		// Unified diff view mode
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

		content.WriteString("\n\n" + styleSubtle.Render("Tab: Toggle View | ↑/↓ j/k: Scroll | gg/G: Top/Bottom | ESC/W: Close"))

		m.diffView.SetContent(content.String())
		m.diffView.GotoTop()
	}
}

// renderDiffMetadata generates a summary header for split view showing status, duration, and headers
func (m *Model) renderDiffMetadata() string {
	var content strings.Builder

	// Header names
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
	pinnedStatus := fmt.Sprintf("%d %s", m.pinnedResponse.Status, m.pinnedResponse.StatusText)
	currentStatus := fmt.Sprintf("%d %s", m.currentResponse.Status, m.currentResponse.StatusText)

	statusLabel := "Status: "
	if pinnedStatus == currentStatus {
		content.WriteString(statusLabel + pinnedStatus)
	} else {
		content.WriteString(statusLabel + styleError.Render(pinnedStatus) + " → " + styleSuccess.Render(currentStatus))
	}

	// Duration comparison
	content.WriteString("  |  Duration: ")
	if m.pinnedResponse.Duration == m.currentResponse.Duration {
		content.WriteString(fmt.Sprintf("%dms", m.pinnedResponse.Duration))
	} else {
		content.WriteString(styleError.Render(fmt.Sprintf("%dms", m.pinnedResponse.Duration)) + " → " + styleSuccess.Render(fmt.Sprintf("%dms", m.currentResponse.Duration)))
	}

	// Header differences count
	headerDiff := compareHeaders(m.pinnedResponse.Headers, m.currentResponse.Headers)
	if headerDiff == "" {
		content.WriteString("  |  Headers: ✓")
	} else {
		// Count differences
		diffCount := strings.Count(headerDiff, "\n")
		content.WriteString(fmt.Sprintf("  |  Headers: %d diff(s)", diffCount))
	}

	return content.String()
}

// formatResponseBody formats a response body for display in split view
func formatResponseBody(body string, width int) string {
	if body == "" {
		return styleSubtle.Render("(empty)")
	}

	lines := strings.Split(body, "\n")
	var result strings.Builder

	for i, line := range lines {
		// Wrap long lines
		wrapped := wrapText(line, width)
		result.WriteString(wrapped)
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
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
				// Wrap long header values
				wrappedPinned := wrapText(pinnedVal, 110)
				for _, line := range strings.Split(wrappedPinned, "\n") {
					result.WriteString(fmt.Sprintf("    - %s\n", styleError.Render(line)))
				}
				wrappedCurrent := wrapText(currentVal, 110)
				for _, line := range strings.Split(wrappedCurrent, "\n") {
					result.WriteString(fmt.Sprintf("    + %s\n", styleSuccess.Render(line)))
				}
			}
		} else {
			// Wrap long header values
			wrappedVal := wrapText(pinnedVal, 110)
			if strings.Contains(wrappedVal, "\n") {
				// Multi-line wrapped value
				result.WriteString(fmt.Sprintf("  - %s:\n", styleError.Render(key)))
				for _, line := range strings.Split(wrappedVal, "\n") {
					result.WriteString(fmt.Sprintf("      %s\n", styleError.Render(line)))
				}
			} else {
				result.WriteString(fmt.Sprintf("  - %s: %s\n", styleError.Render(key), styleError.Render(wrappedVal)))
			}
		}
	}

	// Check for new headers in current
	for key, currentVal := range current {
		if !seen[key] {
			// Wrap long header values
			wrappedVal := wrapText(currentVal, 110)
			if strings.Contains(wrappedVal, "\n") {
				// Multi-line wrapped value
				result.WriteString(fmt.Sprintf("  + %s:\n", styleSuccess.Render(key)))
				for _, line := range strings.Split(wrappedVal, "\n") {
					result.WriteString(fmt.Sprintf("      %s\n", styleSuccess.Render(line)))
				}
			} else {
				result.WriteString(fmt.Sprintf("  + %s: %s\n", styleSuccess.Render(key), styleSuccess.Render(wrappedVal)))
			}
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
			// Same line - show it with neutral color, wrapped
			wrapped := wrapText(pinnedLine, 120)
			for _, wrappedLine := range strings.Split(wrapped, "\n") {
				result.WriteString(fmt.Sprintf("  %s\n", wrappedLine))
			}
		} else {
			if pinnedLine != "" {
				// Wrap removed line
				wrapped := wrapText(pinnedLine, 120)
				for _, wrappedLine := range strings.Split(wrapped, "\n") {
					result.WriteString(fmt.Sprintf("  - %s\n", styleError.Render(wrappedLine)))
				}
				differences++
			}
			if currentLine != "" {
				// Wrap added line
				wrapped := wrapText(currentLine, 120)
				for _, wrappedLine := range strings.Split(wrapped, "\n") {
					result.WriteString(fmt.Sprintf("  + %s\n", styleSuccess.Render(wrappedLine)))
				}
				differences++
			}
		}
	}

	if differences == 0 && len(pinnedLines) > 0 && len(currentLines) > 0 {
		return "  (no differences)\n"
	}

	return result.String()
}

// compareTextSplitView generates side-by-side diff with background highlighting
// Returns two styled strings - one for left pane (pinned), one for right pane (current)
func compareTextSplitView(pinned, current string, width int) (string, string) {
	pinnedLines := strings.Split(pinned, "\n")
	currentLines := strings.Split(current, "\n")

	var leftResult, rightResult strings.Builder

	// Handle empty cases
	if pinned == "" && current == "" {
		return styleSubtle.Render("(empty)"), styleSubtle.Render("(empty)")
	}
	if pinned == "" {
		pinnedLines = make([]string, len(currentLines))
	}
	if current == "" {
		currentLines = make([]string, len(pinnedLines))
	}

	maxLines := len(pinnedLines)
	if len(currentLines) > maxLines {
		maxLines = len(currentLines)
	}

	for i := 0; i < maxLines; i++ {
		pinnedLine := ""
		if i < len(pinnedLines) {
			pinnedLine = pinnedLines[i]
		}

		currentLine := ""
		if i < len(currentLines) {
			currentLine = currentLines[i]
		}

		// Wrap lines before styling to ensure proper width
		wrappedPinned := wrapText(pinnedLine, width)
		wrappedCurrent := wrapText(currentLine, width)

		// Split wrapped lines to handle multi-line wraps
		pinnedWrappedLines := strings.Split(wrappedPinned, "\n")
		currentWrappedLines := strings.Split(wrappedCurrent, "\n")

		// Ensure both sides have same number of lines for alignment
		maxWrappedLines := len(pinnedWrappedLines)
		if len(currentWrappedLines) > maxWrappedLines {
			maxWrappedLines = len(currentWrappedLines)
		}

		for j := 0; j < maxWrappedLines; j++ {
			leftLine := ""
			if j < len(pinnedWrappedLines) {
				leftLine = pinnedWrappedLines[j]
			}

			rightLine := ""
			if j < len(currentWrappedLines) {
				rightLine = currentWrappedLines[j]
			}

			// Compare original lines to determine if they differ
			isDifferent := pinnedLine != currentLine

			// Apply styling based on difference
			if isDifferent {
				// Lines are different - apply background colors
				if leftLine != "" {
					leftResult.WriteString(styleDiffRemoved.Render(leftLine) + "\n")
				} else {
					leftResult.WriteString("\n") // Empty line for alignment
				}

				if rightLine != "" {
					rightResult.WriteString(styleDiffAdded.Render(rightLine) + "\n")
				} else {
					rightResult.WriteString("\n") // Empty line for alignment
				}
			} else {
				// Lines are identical - use neutral styling
				if leftLine != "" {
					leftResult.WriteString(styleDiffNeutral.Render(leftLine) + "\n")
				} else {
					leftResult.WriteString("\n")
				}

				if rightLine != "" {
					rightResult.WriteString(styleDiffNeutral.Render(rightLine) + "\n")
				} else {
					rightResult.WriteString("\n")
				}
			}
		}
	}

	return leftResult.String(), rightResult.String()
}
