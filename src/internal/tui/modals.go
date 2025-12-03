package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderHelp renders the help screen
func (m Model) renderHelp() string {
	// Build title
	title := styleTitle.Render("Keyboard Shortcuts")

	// Build footer with search info if active
	var footer string
	if m.helpSearchActive {
		footer = styleWarning.Render("Search: "+m.helpSearchQuery+"█") + " | ESC: cancel"
	} else if m.helpSearchQuery != "" {
		footer = styleSubtle.Render("Search: "+m.helpSearchQuery) + " | /: search | ESC: clear"
	} else {
		footer = "↑/↓ j/k: scroll | /: search | ESC/?: close"
	}

	// Combine title, viewport, and footer (footer is OUTSIDE viewport so it stays visible)
	fullContent := title + "\n\n" + m.helpView.View() + "\n\n" + styleSubtle.Render(footer)

	// Render modal with scrolling viewport
	helpView := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBlue).
		Width(m.width - 10).
		Height(m.height - 4).
		Padding(1, 2).
		Render(fullContent)

	// Center the help box
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		helpView,
	)
}

// renderDocumentation renders the documentation viewer modal with collapsible tree structure
func (m *Model) renderDocumentation() string {
	// Use nearly full screen but leave small margin
	modalWidth := m.width - 6
	modalHeight := m.height - 3

	// Footer with keybinds
	// Note: viewport dimensions are set in updateDocumentationView()
	footer := styleSubtle.Render("↑/↓/j/k: Navigate | PgUp/PgDn: Scroll | Enter/Space: Toggle | ESC: Close")

	// Create modal content with title, viewport, and footer
	fullContent := styleTitle.Render("Documentation") + "\n\n" + m.modalView.View() + "\n\n" + footer

	docView := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBlue).
		Width(modalWidth).
		Height(modalHeight).
		Padding(1, 2).
		Render(fullContent)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		docView,
	)
}

// min returns the smaller of two ints
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// renderHistory renders the history viewer modal with split view (Telescope-style)
func (m *Model) renderHistory() string {
	// Use nearly full screen but leave small margin
	modalWidth := m.width - 6
	modalHeight := m.height - 3
	paneHeight := modalHeight - 4 // Account for borders and padding

	var mainView string

	if m.historyPreviewVisible {
		// Split view mode: show both list and preview
		listWidth := (modalWidth - 3) / 2          // Left pane: history list
		previewWidth := modalWidth - listWidth - 3 // Right pane: response preview

		// Left pane: History list
		leftPane := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBlue).
			Width(listWidth).
			Height(paneHeight).
			Padding(0, 1).
			Render(styleTitle.Render("History") + "\n" + m.modalView.View())

		// Right pane: Response preview (content populated by updateHistoryView)
		rightPane := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorGreen).
			Width(previewWidth).
			Height(paneHeight).
			Padding(0, 1).
			Render(styleTitle.Render("Response Preview") + "\n" + m.historyPreviewView.View())

		// Join panes horizontally (Telescope-style split)
		mainView = lipgloss.JoinHorizontal(
			lipgloss.Top,
			leftPane,
			rightPane,
		)
	} else {
		// Preview hidden: expand list to full width
		mainView = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBlue).
			Width(modalWidth).
			Height(paneHeight).
			Padding(0, 1).
			Render(styleTitle.Render("History") + "\n" + m.modalView.View())
	}

	// Add footer with instructions (include 'p' key)
	footer := styleSubtle.Render("↑/↓ j/k: Navigate | Enter: Load | r: Replay | p: Toggle Preview | C: Clear All | ESC/H/q: Close")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		mainView,
		"\n"+footer,
	)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

// renderModal renders a generic modal dialog with scrollable content
func (m *Model) renderModal(title, content string, width, height int) string {
	return m.renderModalWithFooterAndScroll(title, content, "", width, height, -1)
}

// renderModalWithFooter renders a modal dialog with scrollable content and a fixed footer
func (m *Model) renderModalWithFooter(title, content, footer string, width, height int) string {
	return m.renderModalWithFooterAndScroll(title, content, footer, width, height, -1)
}

// renderModalWithFooterAndScroll renders a modal with footer and auto-scrolls to keep selectedLine visible
// Pass selectedLine=-1 to preserve existing scroll position
func (m *Model) renderModalWithFooterAndScroll(title, content, footer string, width, height, selectedLine int) string {
	// For small terminals, use almost full screen
	maxWidth := m.width - 4   // Leave minimal margin
	maxHeight := m.height - 2 // Leave minimal margin

	// Adjust requested dimensions to fit screen
	if width > maxWidth {
		width = maxWidth
	}
	if height > maxHeight {
		height = maxHeight
	}

	// Ensure minimum reasonable size (but allow small for tiny terminals)
	if width < 30 && m.width >= 30 {
		width = 30
	}
	if height < 8 && m.height >= 8 {
		height = 8
	}

	// Update modal viewport size and content
	// Account for title (2 lines), padding (2), and border (2) = 6 lines total overhead
	// Also account for footer if present (2 lines: blank + footer)
	footerLines := 0
	if footer != "" {
		footerLines = 2
	}
	contentHeight := height - 6 - footerLines
	if contentHeight < 1 {
		// For very small terminals, reduce overhead
		contentHeight = height - 4 - footerLines // Just border and title
		if contentHeight < 1 {
			contentHeight = 1
		}
	}

	m.modalView.Width = width - 4 // Account for horizontal padding (1 left + 1 right) * 2 for border
	if m.modalView.Width < 10 {
		m.modalView.Width = 10
	}
	m.modalView.Height = contentHeight

	// Save scroll before SetContent resets it
	savedOffset := m.modalView.YOffset

	// Always update content for dynamic modals
	m.modalView.SetContent(content)

	// Auto-scroll only when selected item would be out of view
	if selectedLine >= 0 && m.modalView.Height > 0 {
		// Check if selected line is visible in current scroll position
		topVisible := savedOffset
		bottomVisible := savedOffset + m.modalView.Height - 1

		if selectedLine < topVisible {
			// Selected is above viewport - scroll up
			m.modalView.SetYOffset(selectedLine)
		} else if selectedLine > bottomVisible {
			// Selected is below viewport - scroll down just enough
			m.modalView.SetYOffset(selectedLine - m.modalView.Height + 1)
		} else {
			// Selected is visible - keep current scroll
			m.modalView.SetYOffset(savedOffset)
		}
	} else {
		// Keep scroll position for other content
		m.modalView.SetYOffset(savedOffset)
	}

	// Create modal content with title, viewport, and optional footer
	fullContent := styleTitle.Render(title) + "\n\n" + m.modalView.View()
	if footer != "" {
		fullContent += "\n\n" + styleSubtle.Render(footer)
	}

	// Create modal box
	modalBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBlue).
		Width(width).
		Height(height).
		Padding(1, 2).
		Render(fullContent)

	// For small terminals or large modals, don't center - just render
	if width >= m.width-2 || height >= m.height-1 {
		// Modal is full screen or nearly full screen
		return modalBox
	}

	// Create centered modal
	centeredModal := lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		modalBox,
	)

	return centeredModal
}

// renderShellErrorsModal renders the shell errors modal
func (m *Model) renderShellErrorsModal() string {
	var content strings.Builder
	for i, err := range m.shellErrors {
		if i > 0 {
			content.WriteString("\n\n")
		}
		content.WriteString(styleError.Render(err))
	}

	// Use existing modal helper with footer
	width := m.width - 6
	height := m.height - 4
	if width < 40 {
		width = 40
	}
	if height < 10 {
		height = 10
	}

	return m.renderModalWithFooter("Shell Errors", content.String(), "j/k: scroll | g/G: top/bottom | ESC: close", width, height)
}

// updateShellErrorsView is a no-op since renderModalWithFooter handles content
func (m *Model) updateShellErrorsView() {
	// Content is built in renderShellErrorsModal, no need to update modalView separately
}

func (m *Model) renderErrorDetailModal() string {
	// Wrap error message for better readability
	width := m.width - 6
	height := m.height - 4
	if width < 50 {
		width = 50
	}
	if height < 10 {
		height = 10
	}

	// Wrap the error text to fit the modal width
	contentWidth := width - 4 // Account for modal padding/borders
	wrappedError := wrapText(m.fullErrorMsg, contentWidth)
	content := styleError.Render(wrappedError)

	return m.renderModalWithFooter("Error Details", content, "j/k: scroll | g/G: top/bottom | ESC: close", width, height)
}

func (m *Model) renderStatusDetailModal() string {
	// Wrap status message for better readability
	width := m.width - 6
	height := m.height - 2
	if width < 50 {
		width = 50
	}
	if height < 10 {
		height = 10
	}

	// Wrap the status text to fit the modal width
	contentWidth := width - 4 // Account for modal padding/borders
	wrappedStatus := wrapText(m.fullStatusMsg, contentWidth)
	content := wrappedStatus

	return m.renderModalWithFooter("Status Message", content, "ESC: close", width, height)
}
