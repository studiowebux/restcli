package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// renderMockServer renders the mock server management modal
func (m *Model) renderMockServer() string {
	var content strings.Builder

	// Title
	content.WriteString(styleTitle.Render("Mock Server Management") + "\n\n")

	// Server status
	if m.mockServerRunning {
		address := m.mockServer.GetAddress()
		content.WriteString(styleSuccess.Render("● Server Running") + "\n")
		content.WriteString(fmt.Sprintf("  Address: %s\n", address))
		if m.mockConfigPath != "" {
			content.WriteString(fmt.Sprintf("  Config: %s\n", filepath.Base(m.mockConfigPath)))
		}
		content.WriteString("\n")

		// Recent logs
		logs := m.mockServer.GetLogs()
		if len(logs) > 0 {
			content.WriteString(styleTitle.Render("Recent Requests") + " (last 10)\n\n")

			// Show last 10 logs
			start := 0
			if len(logs) > 10 {
				start = len(logs) - 10
			}

			for _, log := range logs[start:] {
				statusStyle := styleSuccess
				if log.Status >= 400 {
					statusStyle = styleError
				} else if log.Status >= 300 {
					statusStyle = styleWarning
				}

				timestamp := log.Timestamp.Format("15:04:05")
				// Format method with padding for alignment
				method := fmt.Sprintf("%-6s", log.Method)
				content.WriteString(fmt.Sprintf("[%s] %s %s - %s - %dms\n",
					timestamp,
					method,
					log.Path,
					statusStyle.Render(fmt.Sprintf("%d", log.Status)),
					log.Duration.Milliseconds()))

				if log.MatchedRule != "none" && log.MatchedRule != "" {
					content.WriteString(fmt.Sprintf("  → %s\n", styleSubtle.Render(log.MatchedRule)))
				}
			}
		} else {
			content.WriteString(styleSubtle.Render("No requests received yet\n"))
		}

		content.WriteString("\n")
		content.WriteString(styleSubtle.Render("Press 's' to stop server, 'c' to clear logs, ESC to close"))
	} else {
		content.WriteString(styleSubtle.Render("○ Server Stopped") + "\n\n")

		// Available mock configs
		content.WriteString(styleTitle.Render("Available Mock Configs") + "\n\n")

		// Find .mock.yaml and .mock.json files
		mockFiles := m.findMockConfigs()

		if len(mockFiles) > 0 {
			for i, file := range mockFiles {
				prefix := "  "
				if i == 0 {
					prefix = styleSuccess.Render("→ ")
				}
				content.WriteString(fmt.Sprintf("%s%s\n", prefix, file))
			}
			content.WriteString("\n")
			content.WriteString(styleSubtle.Render("Press 's' to start server, ESC to close"))
		} else {
			content.WriteString(styleSubtle.Render("No .mock.yaml or .mock.json files found\n"))
			content.WriteString(styleSubtle.Render("Create a mock config file to get started\n\n"))
			content.WriteString(styleSubtle.Render("Press ESC to close"))
		}
	}

	// Calculate dimensions
	modalWidth := m.width - 6
	modalHeight := m.height - 3

	// Set viewport dimensions BEFORE setting content
	// Account for padding (1 top, 1 bottom) + title lines + footer
	m.modalView.Width = modalWidth - 4  // Subtract horizontal padding
	m.modalView.Height = modalHeight - 6 // Subtract vertical padding, title, footer

	// Set modal content
	m.modalView.SetContent(content.String())

	// Render modal
	footer := styleSubtle.Render("↑/↓ scroll")

	modalBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBlue).
		Width(modalWidth).
		Height(modalHeight).
		Padding(1, 2).
		Render(m.modalView.View() + "\n\n" + footer)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		modalBox,
	)
}

// handleMockServerKeys handles keyboard input in mock server modal
func (m *Model) handleMockServerKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "q":
		m.mode = ModeNormal

	case "s":
		if m.mockServerRunning {
			// Stop server
			return m.stopMockServer()
		} else {
			// Start server
			return m.startMockServer()
		}

	case "c":
		if m.mockServerRunning && m.mockServer != nil {
			m.mockServer.ClearLogs()
			m.statusMsg = "Mock server logs cleared"
		}

	// Viewport scrolling
	case "up", "k":
		m.modalView.LineUp(1)
	case "down", "j":
		m.modalView.LineDown(1)
	case "pgup":
		m.modalView.PageUp()
	case "pgdown":
		m.modalView.PageDown()
	case "g":
		m.modalView.GotoTop()
	case "G":
		m.modalView.GotoBottom()
	}

	return nil
}

// findMockConfigs finds all .mock.yaml and .mock.json files in workdir
func (m *Model) findMockConfigs() []string {
	profile := m.sessionMgr.GetActiveProfile()
	if profile == nil {
		return nil
	}

	var mockFiles []string
	seen := make(map[string]bool)

	// Check for mock configs in workdir
	workdir := profile.Workdir
	if workdir == "" {
		workdir = "."
	}

	// Common locations - check both workdir and current directory
	// This allows configs to be in project root even if workdir is a subdirectory
	paths := []string{
		filepath.Join(workdir, "mocks"),
		workdir,
		"mocks",     // Current dir mocks/
		".",         // Current directory
		"../mocks",  // Parent dir mocks/ (common when running from src/)
		"..",        // Parent directory
	}

	for _, dir := range paths {
		matches, _ := filepath.Glob(filepath.Join(dir, "*.mock.yaml"))
		for _, match := range matches {
			absPath, _ := filepath.Abs(match)
			if !seen[absPath] {
				mockFiles = append(mockFiles, match)
				seen[absPath] = true
			}
		}

		matches, _ = filepath.Glob(filepath.Join(dir, "*.mock.yml"))
		for _, match := range matches {
			absPath, _ := filepath.Abs(match)
			if !seen[absPath] {
				mockFiles = append(mockFiles, match)
				seen[absPath] = true
			}
		}

		matches, _ = filepath.Glob(filepath.Join(dir, "*.mock.json"))
		for _, match := range matches {
			absPath, _ := filepath.Abs(match)
			if !seen[absPath] {
				mockFiles = append(mockFiles, match)
				seen[absPath] = true
			}
		}
	}

	return mockFiles
}
