package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/studiowebux/restcli/internal/keybinds"
	"github.com/studiowebux/restcli/internal/proxy"
)

// updateProxyView updates the proxy viewer modal content
func (m *Model) updateProxyView() {
	// Get logs from proxy if running
	if m.proxyServerState.IsRunning() && m.proxyServerState.GetServer() != nil {
		m.proxyServerState.SetLogs(m.proxyServerState.GetServer().GetLogs())
	}
}

// renderProxyModal renders the proxy viewer modal
func (m Model) renderProxyModal() string {
	var content strings.Builder

	// Calculate dimensions
	modalWidth := m.width - 6
	modalHeight := m.height - 3
	contentWidth := modalWidth - 4

	if !m.proxyServerState.IsRunning() || m.proxyServerState.GetServer() == nil {
		content.WriteString(styleTitleFocused.Render("Debug Proxy") + "\n\n")
		content.WriteString(styleSubtle.Render("○ Proxy Stopped") + "\n\n")

		// Get configured port from profile
		profile := m.sessionMgr.GetActiveProfile()
		proxyPort := profile.GetProxyPort()

		// Show configuration
		content.WriteString(fmt.Sprintf("Port: %d", proxyPort))
		if profile.ProxyPort != nil {
			content.WriteString(styleSubtle.Render(" (from profile)"))
		} else {
			content.WriteString(styleSubtle.Render(" (default)"))
		}
		content.WriteString("\n\n")
		content.WriteString("Configure your application:\n")
		content.WriteString(fmt.Sprintf("  export HTTP_PROXY=http://localhost:%d\n", proxyPort))
		content.WriteString(fmt.Sprintf("  export http_proxy=http://localhost:%d\n", proxyPort))

		// Set viewport dimensions and content
		m.modalView.Width = contentWidth
		m.modalView.Height = modalHeight - 6
		m.modalView.SetContent(content.String())

		footer := styleSubtle.Render("s start | ESC close")
		modalBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBlue).
			Width(modalWidth).
			Height(modalHeight).
			Padding(1, 2).
			Render(m.modalView.View() + "\n\n" + footer)

		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modalBox)
	}

	// Header
	content.WriteString(styleTitleFocused.Render(fmt.Sprintf("Debug Proxy - Port %d", m.proxyServerState.GetServer().Port)) + "\n")
	content.WriteString(styleSubtle.Render(fmt.Sprintf("Captured: %d requests | Updates in real-time", len(m.proxyServerState.GetLogs()))) + "\n\n")

	if len(m.proxyServerState.GetLogs()) == 0 {
		content.WriteString("No requests captured yet.\n\n")
		content.WriteString("Configure your application:\n")
		content.WriteString(fmt.Sprintf("  export HTTP_PROXY=http://localhost:%d\n", m.proxyServerState.GetServer().Port))
		content.WriteString(fmt.Sprintf("  export http_proxy=http://localhost:%d\n", m.proxyServerState.GetServer().Port))

		m.modalView.Width = contentWidth
		m.modalView.Height = modalHeight - 6
		m.modalView.SetContent(content.String())

		footer := styleSubtle.Render("s stop | ESC close")
		modalBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBlue).
			Width(modalWidth).
			Height(modalHeight).
			Padding(1, 2).
			Render(m.modalView.View() + "\n\n" + footer)

		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modalBox)
	}

	// Traffic log
	content.WriteString(styleTitle.Render("Traffic Log") + "\n")
	content.WriteString(strings.Repeat("─", contentWidth) + "\n")

	// Show all requests (chronological order - oldest first)
	for i := 0; i < len(m.proxyServerState.GetLogs()); i++ {
		log := m.proxyServerState.GetLogs()[i]

		// Format method with color
		var methodStyle lipgloss.Style
		switch log.Method {
		case "GET":
			methodStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
		case "POST":
			methodStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
		case "PUT":
			methodStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
		case "DELETE":
			methodStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
		default:
			methodStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		}

		// Format status with color
		var statusStyle lipgloss.Style
		if log.Status >= 200 && log.Status < 300 {
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
		} else if log.Status >= 300 && log.Status < 400 {
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
		} else if log.Status >= 400 {
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
		}

		// Highlight selected
		lineStyle := lipgloss.NewStyle()
		if i == m.proxyServerState.GetSelectedIndex() {
			lineStyle = lineStyle.Background(lipgloss.Color("237"))
		}

		// Format: #ID METHOD URL → STATUS SIZE DURATION
		line := fmt.Sprintf("#%-4d %s %-50s → %s %8s %8s",
			log.ID,
			methodStyle.Render(fmt.Sprintf("%-6s", log.Method)),
			truncate(log.URL, 50),
			statusStyle.Render(fmt.Sprintf("%-3d", log.Status)),
			proxy.FormatSize(len(log.RespBody)),
			proxy.FormatDuration(log.Duration),
		)

		content.WriteString(lineStyle.Render(line) + "\n")
	}

	// Set viewport dimensions and content
	m.modalView.Width = contentWidth
	m.modalView.Height = modalHeight - 6
	m.modalView.SetContent(content.String())

	footer := styleSubtle.Render("↑/↓ select | Ctrl+d/u half page | Enter details | s stop | c clear | ESC close")
	modalBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBlue).
		Width(modalWidth).
		Height(modalHeight).
		Padding(1, 2).
		Render(m.modalView.View() + "\n\n" + footer)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modalBox)
}

// handleProxyViewerKeys handles key events in proxy viewer mode
func (m *Model) handleProxyViewerKeys(msg tea.KeyMsg) tea.Cmd {
	// Handle special keys not in registry
	switch msg.String() {
	case "s":
		// Start or stop proxy
		if m.proxyServerState.IsRunning() && m.proxyServerState.GetServer() != nil {
			// Stop proxy
			m.proxyServerState.GetServer().Stop()
			m.proxyServerState.Stop()
			m.statusMsg = "Proxy stopped"
			return nil
		} else {
			// Start proxy
			profile := m.sessionMgr.GetActiveProfile()
			proxyPort := profile.GetProxyPort()
			p := proxy.NewProxy(proxyPort)
			if err := p.Start(); err != nil {
				m.errorMsg = fmt.Sprintf("Failed to start proxy: %v", err)
				return nil
			} else {
				m.proxyServerState.Start(p)
				m.statusMsg = fmt.Sprintf("Proxy started on port %d", proxyPort)
				// Start event-based listener
				return m.listenForProxyLogs()
			}
		}

	case "c":
		// Clear captured logs
		if m.proxyServerState.GetServer() != nil {
			m.proxyServerState.GetServer().ClearLogs()
			m.proxyServerState.ClearLogs()
			m.proxyServerState.SetSelectedIndex(0)
		}
		return nil

	case "enter":
		if m.proxyServerState.GetSelectedIndex() >= 0 && m.proxyServerState.GetSelectedIndex() < len(m.proxyServerState.GetLogs()) {
			m.mode = ModeProxyDetail
			m.updateProxyDetailView() // Set content once when entering modal
		}
		return nil
	}

	// Use registry for navigation
	action, ok := m.keybinds.Match(keybinds.ContextModal, msg.String())
	if !ok {
		return nil
	}

	switch action {
	case keybinds.ActionCloseModal:
		m.mode = ModeNormal

	case keybinds.ActionNavigateDown:
		if len(m.proxyServerState.GetLogs()) > 0 && m.proxyServerState.GetSelectedIndex() < len(m.proxyServerState.GetLogs())-1 {
			m.proxyServerState.Navigate(1)
		}

	case keybinds.ActionNavigateUp:
		if len(m.proxyServerState.GetLogs()) > 0 && m.proxyServerState.GetSelectedIndex() > 0 {
			m.proxyServerState.Navigate(-1)
		}

	case keybinds.ActionHalfPageDown:
		// Half page down
		if len(m.proxyServerState.GetLogs()) > 0 {
			halfPage := m.modalView.Height / 2
			m.proxyServerState.Navigate(halfPage)
		}

	case keybinds.ActionHalfPageUp:
		// Half page up
		if len(m.proxyServerState.GetLogs()) > 0 {
			halfPage := m.modalView.Height / 2
			m.proxyServerState.Navigate(-halfPage)
		}

	case keybinds.ActionPageUp:
		m.modalView.PageUp()

	case keybinds.ActionPageDown:
		m.modalView.PageDown()

	case keybinds.ActionGoToTop:
		if len(m.proxyServerState.GetLogs()) > 0 {
			m.proxyServerState.SetSelectedIndex(0)
		}
		m.modalView.GotoTop()

	case keybinds.ActionGoToBottom:
		if len(m.proxyServerState.GetLogs()) > 0 {
			m.proxyServerState.SetSelectedIndex(len(m.proxyServerState.GetLogs()) - 1)
		}
		m.modalView.GotoBottom()
	}

	return nil
}
