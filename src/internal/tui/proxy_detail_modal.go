package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/studiowebux/restcli/internal/proxy"
)

// updateProxyDetailView updates the proxy detail modal viewport content
func (m *Model) updateProxyDetailView() {
	// Set viewport dimensions for the modal
	m.modalView.Width = m.width - 10  // Modal content width minus padding
	m.modalView.Height = m.height - 9 // Modal content height minus padding, title lines, and footer

	// Build and set content
	m.modalView.SetContent(m.buildProxyDetailContent())
}

// renderProxyDetailModal renders the detailed view of a single proxy request
func (m Model) renderProxyDetailModal() string {
	if m.proxySelectedIndex < 0 || m.proxySelectedIndex >= len(m.proxyLogs) {
		return "No request selected"
	}

	log := m.proxyLogs[m.proxySelectedIndex]

	// Fixed footer for keybinds
	footer := styleSubtle.Render("↑/↓ scroll | Ctrl+d/u half page | g/G top/bottom | ESC close")

	modalWidth := m.width - 6
	modalHeight := m.height - 3

	detailView := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBlue).
		Width(modalWidth).
		Height(modalHeight).
		Padding(1, 2).
		Render(styleTitle.Render(fmt.Sprintf("Request #%d Details", log.ID)) + "\n\n" + m.modalView.View() + "\n\n" + footer)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		detailView,
	)
}

// buildProxyDetailContent builds the content for the proxy detail modal
func (m *Model) buildProxyDetailContent() string {
	if m.proxySelectedIndex < 0 || m.proxySelectedIndex >= len(m.proxyLogs) {
		return "No request selected"
	}

	log := m.proxyLogs[m.proxySelectedIndex]
	var content strings.Builder

	// Request line
	content.WriteString(styleTitle.Render("Request") + "\n")
	content.WriteString(fmt.Sprintf("%s %s\n", log.Method, log.URL))
	content.WriteString(fmt.Sprintf("Duration: %s\n\n", proxy.FormatDuration(log.Duration)))

	// Request headers
	if len(log.ReqHeaders) > 0 {
		content.WriteString(styleSubtle.Render("Request Headers:") + "\n")
		for name, values := range log.ReqHeaders {
			content.WriteString(fmt.Sprintf("  %s: %s\n", name, strings.Join(values, ", ")))
		}
		content.WriteString("\n")
	}

	// Request body
	if len(log.ReqBody) > 0 {
		content.WriteString(styleSubtle.Render("Request Body:") + fmt.Sprintf(" (%d bytes)\n", len(log.ReqBody)))
		bodyText := string(log.ReqBody)

		// Check if content is binary
		if isBinaryContent(bodyText) {
			contentType := "unknown"
			if ct := log.ReqHeaders.Get("Content-Type"); ct != "" {
				contentType = ct
			}
			content.WriteString(styleSubtle.Render(fmt.Sprintf(
				"[Binary content - %s - %d bytes]\n\nRequest contains binary data that cannot be displayed as text.\nContent-Type: %s",
				proxy.FormatSize(len(log.ReqBody)),
				len(log.ReqBody),
				contentType,
			)) + "\n\n")
		} else {
			// Try to pretty-print JSON
			var jsonData interface{}
			if err := json.Unmarshal(log.ReqBody, &jsonData); err == nil {
				if prettyJSON, err := json.MarshalIndent(jsonData, "", "  "); err == nil {
					bodyText = string(prettyJSON)
				}
			}
			content.WriteString(bodyText + "\n\n")
		}
	}

	// Response line
	content.WriteString(styleTitle.Render("Response") + "\n")
	var statusStyle lipgloss.Style
	if log.Status >= 200 && log.Status < 300 {
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	} else if log.Status >= 300 && log.Status < 400 {
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	} else if log.Status >= 400 {
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	}
	content.WriteString(statusStyle.Render(fmt.Sprintf("%d %s", log.Status, log.StatusText)) + "\n\n")

	// Response headers
	if len(log.RespHeaders) > 0 {
		content.WriteString(styleSubtle.Render("Response Headers:") + "\n")
		for name, values := range log.RespHeaders {
			content.WriteString(fmt.Sprintf("  %s: %s\n", name, strings.Join(values, ", ")))
		}
		content.WriteString("\n")
	}

	// Response body
	if len(log.RespBody) > 0 {
		content.WriteString(styleSubtle.Render("Response Body:") + fmt.Sprintf(" (%d bytes)\n", len(log.RespBody)))
		bodyText := string(log.RespBody)

		// Check if content is binary
		if isBinaryContent(bodyText) {
			contentType := "unknown"
			if ct := log.RespHeaders.Get("Content-Type"); ct != "" {
				contentType = ct
			}
			content.WriteString(styleSubtle.Render(fmt.Sprintf(
				"[Binary content - %s - %d bytes]\n\nResponse contains binary data that cannot be displayed as text.\nContent-Type: %s",
				proxy.FormatSize(len(log.RespBody)),
				len(log.RespBody),
				contentType,
			)) + "\n\n")
		} else {
			// Try to pretty-print JSON
			var jsonData interface{}
			if err := json.Unmarshal(log.RespBody, &jsonData); err == nil {
				if prettyJSON, err := json.MarshalIndent(jsonData, "", "  "); err == nil {
					bodyText = string(prettyJSON)
				}
			}
			content.WriteString(bodyText + "\n\n")
		}
	}

	return content.String()
}

// handleProxyDetailKeys handles key events in proxy detail mode
func (m *Model) handleProxyDetailKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "q":
		m.mode = ModeProxyViewer
		return nil

	case "up", "k":
		m.modalView.ScrollUp(1)

	case "down", "j":
		m.modalView.ScrollDown(1)

	case "ctrl+u":
		m.modalView.HalfViewUp()

	case "ctrl+d":
		m.modalView.HalfViewDown()

	case "pgup":
		m.modalView.PageUp()

	case "pgdown":
		m.modalView.PageDown()

	case "g", "home":
		m.modalView.GotoTop()

	case "G", "end":
		m.modalView.GotoBottom()
	}

	return nil
}
