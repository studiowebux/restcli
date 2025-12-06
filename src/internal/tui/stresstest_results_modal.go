package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/studiowebux/restcli/internal/stresstest"
)

// renderStressTestResults renders the stress test results modal with split view
func (m *Model) renderStressTestResults() string {
	modalWidth := m.width - 6
	modalHeight := m.height - 3
	paneHeight := modalHeight - 4

	// Split view: list on left, details on right
	listWidth := (modalWidth - 3) / 2
	detailWidth := modalWidth - listWidth - 3

	// Determine border colors based on focus
	listBorderColor := colorGray
	detailBorderColor := colorGray
	listTitleStyle := styleTitleUnfocused
	detailTitleStyle := styleTitleUnfocused

	if m.stressTestFocusedPane == "list" {
		listBorderColor = colorCyan
		listTitleStyle = styleTitleFocused
	} else if m.stressTestFocusedPane == "details" {
		detailBorderColor = colorCyan
		detailTitleStyle = styleTitleFocused
	}

	// Set viewport dimensions
	m.stressTestListView.Width = listWidth - 4
	m.stressTestListView.Height = paneHeight - 2

	m.stressTestDetailView.Width = detailWidth - 4
	m.stressTestDetailView.Height = paneHeight - 2

	// Update list content
	m.updateStressTestListView()

	// Left pane: Test runs list
	leftPane := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(listBorderColor).
		Width(listWidth).
		Height(paneHeight).
		Padding(0, 1).
		Render(listTitleStyle.Render("Test Runs") + "\n" + m.stressTestListView.View())

	// Right pane: Test run details
	rightPane := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(detailBorderColor).
		Width(detailWidth).
		Height(paneHeight).
		Padding(0, 1).
		Render(detailTitleStyle.Render("Details") + "\n" + m.stressTestDetailView.View())

	mainView := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPane,
		rightPane,
	)

	// Footer
	footer := "n: New | r: Re-run | l: Load Config | TAB: Switch Focus | ↑/↓ j/k: Navigate | g/G: Top/Bottom | d: Delete | ESC/q: Close"
	footerContent := styleSubtle.Render(footer)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		mainView,
		"\n"+footerContent,
	)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

// updateStressTestListView updates the stress test runs list view
func (m *Model) updateStressTestListView() {
	var listContent strings.Builder

	if len(m.stressTestRuns) == 0 {
		listContent.WriteString("No stress test runs found.\n\nPress 'n' to create and run a stress test.")
	} else {
		for i, run := range m.stressTestRuns {
			// Format display
			displayName := run.ConfigName

			// Status icon
			statusIcon := "OK"
			if run.Status == "failed" {
				statusIcon = "ERR"
			} else if run.Status == "cancelled" {
				statusIcon = "STOP"
			} else if run.Status == "running" {
				statusIcon = "RUN"
			}

			// Format line
			line := fmt.Sprintf("%s %s\n  %s | %d reqs",
				statusIcon,
				displayName,
				run.StartedAt.Format("2006-01-02 15:04"),
				run.TotalRequestsCompleted)

			if run.AvgDurationMs > 0 {
				line += fmt.Sprintf(" | %.0fms avg", run.AvgDurationMs)
			}

			// Highlight selected
			if i == m.stressTestRunIndex {
				line = styleSelected.Render("> " + line)
			} else {
				line = "  " + line
			}

			listContent.WriteString(line + "\n")
		}
	}

	m.stressTestListView.SetContent(listContent.String())

	// Auto-scroll to keep selected item visible
	if len(m.stressTestRuns) > 0 && m.stressTestRunIndex >= 0 && m.stressTestRunIndex < len(m.stressTestRuns) {
		linePos := m.stressTestRunIndex * 2 // 2 lines per item
		viewportHeight := m.stressTestListView.Height

		desiredOffset := linePos - (viewportHeight / 2)
		if desiredOffset < 0 {
			desiredOffset = 0
		}

		m.stressTestListView.SetYOffset(desiredOffset)
	}

	// Update detail view
	m.updateStressTestDetailView()
}

// updateStressTestDetailView updates the stress test run details view
func (m *Model) updateStressTestDetailView() {
	var detailContent strings.Builder

	if len(m.stressTestRuns) == 0 || m.stressTestRunIndex >= len(m.stressTestRuns) {
		detailContent.WriteString("No test run selected")
	} else {
		run := m.stressTestRuns[m.stressTestRunIndex]

		// Header
		detailContent.WriteString(styleTitle.Render(run.ConfigName) + "\n\n")

		// File info
		detailContent.WriteString(styleSubtle.Render("File: ") + filepath.Base(run.RequestFile) + "\n")
		detailContent.WriteString("\n")

		// Status and timing
		detailContent.WriteString(styleTitle.Render("Status") + "\n")
		detailContent.WriteString(fmt.Sprintf("Status:     %s\n", run.Status))
		detailContent.WriteString(fmt.Sprintf("Started:    %s\n", run.StartedAt.Format("2006-01-02 15:04:05")))
		if run.CompletedAt != nil {
			detailContent.WriteString(fmt.Sprintf("Completed:  %s\n", run.CompletedAt.Format("2006-01-02 15:04:05")))
			duration := run.CompletedAt.Sub(run.StartedAt)
			detailContent.WriteString(fmt.Sprintf("Duration:   %s\n", formatDuration(duration)))
		}
		detailContent.WriteString("\n")

		// Request stats
		detailContent.WriteString(styleTitle.Render("Requests") + "\n")
		detailContent.WriteString(fmt.Sprintf("Sent:         %d\n", run.TotalRequestsSent))
		detailContent.WriteString(fmt.Sprintf("Completed:    %d\n", run.TotalRequestsCompleted))
		successCount := run.TotalRequestsCompleted - run.TotalErrors - run.TotalValidationErrors
		detailContent.WriteString(fmt.Sprintf("Success:      %d\n", successCount))
		detailContent.WriteString(fmt.Sprintf("Net Errors:   %d\n", run.TotalErrors))
		detailContent.WriteString(fmt.Sprintf("Val Errors:   %d\n", run.TotalValidationErrors))
		if run.TotalRequestsCompleted > 0 {
			successRate := float64(successCount) / float64(run.TotalRequestsCompleted) * 100
			detailContent.WriteString(fmt.Sprintf("Success Rate: %.1f%%\n", successRate))
		}
		detailContent.WriteString("\n")

		// Latency stats
		detailContent.WriteString(styleTitle.Render("Latency") + "\n")
		detailContent.WriteString(fmt.Sprintf("Average:    %.0fms\n", run.AvgDurationMs))
		detailContent.WriteString(fmt.Sprintf("Min:        %dms\n", run.MinDurationMs))
		detailContent.WriteString(fmt.Sprintf("Max:        %dms\n", run.MaxDurationMs))
		detailContent.WriteString(fmt.Sprintf("P50:        %dms\n", run.P50DurationMs))
		detailContent.WriteString(fmt.Sprintf("P95:        %dms\n", run.P95DurationMs))
		detailContent.WriteString(fmt.Sprintf("P99:        %dms\n", run.P99DurationMs))
	}

	m.stressTestDetailView.SetContent(detailContent.String())
}

// loadStressTestRuns loads stress test runs from the database
func (m *Model) loadStressTestRuns() tea.Cmd {
	return func() tea.Msg {
		if m.stressTestManager == nil {
			return stressTestRunsLoadedMsg{runs: []*stresstest.Run{}}
		}

		// Get active profile name
		profileName := ""
		if profile := m.sessionMgr.GetActiveProfile(); profile != nil {
			profileName = profile.Name
		}

		runs, err := m.stressTestManager.ListRuns(profileName, 100) // Limit to 100 most recent
		if err != nil {
			return errorMsg(fmt.Sprintf("Failed to load stress test runs: %v", err))
		}

		return stressTestRunsLoadedMsg{runs: runs}
	}
}

// loadStressTestConfigs loads stress test configs from the database
func (m *Model) loadStressTestConfigs() tea.Cmd {
	return func() tea.Msg {
		if m.stressTestManager == nil {
			return stressTestConfigsLoadedMsg{configs: []*stresstest.Config{}}
		}

		// Get active profile name
		profileName := ""
		if profile := m.sessionMgr.GetActiveProfile(); profile != nil {
			profileName = profile.Name
		}

		configs, err := m.stressTestManager.ListConfigs(profileName)
		if err != nil {
			return errorMsg(fmt.Sprintf("Failed to load stress test configs: %v", err))
		}

		return stressTestConfigsLoadedMsg{configs: configs}
	}
}

// Message types for stress test
type stressTestRunsLoadedMsg struct {
	runs []*stresstest.Run
}

type stressTestConfigsLoadedMsg struct {
	configs []*stresstest.Config
}
