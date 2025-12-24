package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/studiowebux/restcli/internal/keybinds"
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
	// Match key to action using keybinds registry
	action, ok := m.keybinds.Match(keybinds.ContextInspect, msg.String())
	if !ok {
		return nil
	}

	switch action {
	case keybinds.ActionCloseModal, keybinds.ActionCloseModalAlt:
		m.mode = ModeNormal

	case keybinds.ActionExecute:
		m.mode = ModeNormal
		return m.executeRequest()

	case keybinds.ActionNavigateUp:
		m.modalView.ScrollUp(1)

	case keybinds.ActionNavigateDown:
		m.modalView.ScrollDown(1)

	case keybinds.ActionPageUp:
		m.modalView.PageUp()

	case keybinds.ActionPageDown:
		m.modalView.PageDown()

	case keybinds.ActionGoToTop:
		m.modalView.GotoTop()

	case keybinds.ActionGoToBottom:
		m.modalView.GotoBottom()
	}

	return nil
}
