package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// renderInspect renders the request inspection modal
func (m *Model) renderInspect() string {
	// Render viewport with scrolling support (like help modal)
	// Use nearly full screen but leave small margin
	modalWidth := m.width - 6
	modalHeight := m.height - 3

	// Fixed footer for keybinds
	footer := styleSubtle.Render("↑/↓ scroll [Enter] execute [ESC] close")

	inspectView := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBlue).
		Width(modalWidth).
		Height(modalHeight).
		Padding(1, 2).
		Render(styleTitle.Render("Inspect Request") + "\n\n" + m.modalView.View() + "\n\n" + footer)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		inspectView,
	)
}

// handleInspectKeys handles keyboard input in inspect mode
func (m *Model) handleInspectKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal

	case "enter":
		m.mode = ModeNormal
		return m.executeRequest()

	// Viewport scrolling like help modal
	case "up", "k":
		m.modalView.ScrollUp(1)

	case "down", "j":
		m.modalView.ScrollDown(1)

	case "pgup":
		m.modalView.PageUp()

	case "pgdown":
		m.modalView.PageDown()

	case "home":
		m.modalView.GotoTop()

	case "end":
		m.modalView.GotoBottom()
	}

	return nil
}
