package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/studiowebux/restcli/internal/proxy"
)

// updateProxyView updates the proxy viewer modal content
func (m *Model) updateProxyView() {
	// Get logs from proxy if running
	if m.proxyRunning && m.proxyServer != nil {
		m.proxyLogs = m.proxyServer.GetLogs()
	}
}

// renderProxyModal renders the proxy viewer modal
func (m Model) renderProxyModal() string {
	var content strings.Builder

	// Calculate dimensions
	modalWidth := m.width - 6
	modalHeight := m.height - 3
	contentWidth := modalWidth - 4

	if !m.proxyRunning || m.proxyServer == nil {
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
	content.WriteString(styleTitleFocused.Render(fmt.Sprintf("Debug Proxy - Port %d", m.proxyServer.Port)) + "\n")
	content.WriteString(styleSubtle.Render(fmt.Sprintf("Captured: %d requests | Updates in real-time", len(m.proxyLogs))) + "\n\n")

	if len(m.proxyLogs) == 0 {
		content.WriteString("No requests captured yet.\n\n")
		content.WriteString("Configure your application:\n")
		content.WriteString(fmt.Sprintf("  export HTTP_PROXY=http://localhost:%d\n", m.proxyServer.Port))
		content.WriteString(fmt.Sprintf("  export http_proxy=http://localhost:%d\n", m.proxyServer.Port))

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
	for i := 0; i < len(m.proxyLogs); i++ {
		log := m.proxyLogs[i]

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
		if i == m.proxySelectedIndex {
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
	switch msg.String() {
	case "esc", "q":
		m.mode = ModeNormal

	case "s":
		// Start or stop proxy
		if m.proxyRunning && m.proxyServer != nil {
			// Stop proxy
			m.proxyServer.Stop()
			m.proxyServer = nil
			m.proxyRunning = false
			m.proxyLogs = nil
			m.proxySelectedIndex = 0
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
				m.proxyServer = p
				m.proxyRunning = true
				m.statusMsg = fmt.Sprintf("Proxy started on port %d", proxyPort)
				// Start event-based listener
				return m.listenForProxyLogs()
			}
		}

	case "c":
		// Clear captured logs
		if m.proxyServer != nil {
			m.proxyServer.ClearLogs()
			m.proxyLogs = nil
			m.proxySelectedIndex = 0
		}

	case "j", "down":
		if len(m.proxyLogs) > 0 && m.proxySelectedIndex < len(m.proxyLogs)-1 {
			m.proxySelectedIndex++
		}

	case "k", "up":
		if len(m.proxyLogs) > 0 && m.proxySelectedIndex > 0 {
			m.proxySelectedIndex--
		}

	case "ctrl+d":
		// Half page down
		if len(m.proxyLogs) > 0 {
			halfPage := m.modalView.Height / 2
			m.proxySelectedIndex += halfPage
			if m.proxySelectedIndex >= len(m.proxyLogs) {
				m.proxySelectedIndex = len(m.proxyLogs) - 1
			}
		}

	case "ctrl+u":
		// Half page up
		if len(m.proxyLogs) > 0 {
			halfPage := m.modalView.Height / 2
			m.proxySelectedIndex -= halfPage
			if m.proxySelectedIndex < 0 {
				m.proxySelectedIndex = 0
			}
		}

	case "pgup":
		m.modalView.PageUp()

	case "pgdown":
		m.modalView.PageDown()

	case "g":
		if len(m.proxyLogs) > 0 {
			m.proxySelectedIndex = 0
		}
		m.modalView.GotoTop()

	case "G":
		if len(m.proxyLogs) > 0 {
			m.proxySelectedIndex = len(m.proxyLogs) - 1
		}
		m.modalView.GotoBottom()

	case "enter":
		if m.proxySelectedIndex >= 0 && m.proxySelectedIndex < len(m.proxyLogs) {
			m.mode = ModeProxyDetail
			m.updateProxyDetailView() // Set content once when entering modal
		}
	}

	return nil
}
