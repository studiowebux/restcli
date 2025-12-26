package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/studiowebux/restcli/internal/stresstest"
)

// renderStressTestResults renders the stress test results modal with split view
func (m *Model) renderStressTestResults() string {
	modalWidth := m.width - ModalWidthMargin
	modalHeight := m.height - ModalHeightMargin
	paneHeight := modalHeight - 4

	// Split view: list on left, details on right
	listWidth := (modalWidth - 3) / 2
	detailWidth := modalWidth - listWidth - 3

	// Determine border colors and focus based on focus state
	listBorderColor := colorGray
	detailBorderColor := colorGray
	leftIsFocused := false
	rightIsFocused := false

	if m.stressTestState.GetFocusedPane() == "list" {
		listBorderColor = colorCyan
		leftIsFocused = true
	} else if m.stressTestState.GetFocusedPane() == "details" {
		detailBorderColor = colorCyan
		rightIsFocused = true
	}

	// Set viewport dimensions
	listView := m.stressTestState.GetListView()
	listView.Width = listWidth - 4
	listView.Height = paneHeight - 2
	m.stressTestState.SetListView(listView)

	detailView := m.stressTestState.GetDetailView()
	detailView.Width = detailWidth - 4
	detailView.Height = paneHeight - 2
	m.stressTestState.SetDetailView(detailView)

	// Update list content
	m.updateStressTestListView()

	// Configure split-pane modal
	cfg := SplitPaneConfig{
		ModalWidth:       modalWidth,
		ModalHeight:      modalHeight,
		IsSplitView:      true, // Stress test results always shows split view
		LeftTitle:        "Test Runs",
		LeftContent:      m.stressTestState.GetListView().View(),
		LeftBorderColor:  listBorderColor,
		LeftIsFocused:    leftIsFocused,
		RightTitle:       "Details",
		RightContent:     m.stressTestState.GetDetailView().View(),
		RightBorderColor: detailBorderColor,
		RightIsFocused:   rightIsFocused,
		Footer:           "n: New | r: Re-run | l: Load Config | TAB: Switch Focus | ↑/↓ j/k: Navigate | g/G: Top/Bottom | d: Delete | ESC/q: Close",
		LeftWidthRatio:   SplitViewEqual,
	}

	return renderSplitPaneModal(cfg, m.width, m.height)
}

// updateStressTestListView updates the stress test runs list view
func (m *Model) updateStressTestListView() {
	var listContent strings.Builder

	if len(m.stressTestState.GetRuns()) == 0 {
		listContent.WriteString("No stress test runs found.\n\nPress 'n' to create and run a stress test.")
	} else {
		for i, run := range m.stressTestState.GetRuns() {
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
			if i == m.stressTestState.GetRunIndex() {
				line = styleSelected.Render("> " + line)
			} else {
				line = "  " + line
			}

			listContent.WriteString(line + "\n")
		}
	}

	listView := m.stressTestState.GetListView()
	listView.SetContent(listContent.String())
	m.stressTestState.SetListView(listView)

	// Auto-scroll to keep selected item visible
	if len(m.stressTestState.GetRuns()) > 0 && m.stressTestState.GetRunIndex() >= 0 && m.stressTestState.GetRunIndex() < len(m.stressTestState.GetRuns()) {
		linePos := m.stressTestState.GetRunIndex() * 2 // 2 lines per item
		listView := m.stressTestState.GetListView()
		viewportHeight := listView.Height

		desiredOffset := linePos - (viewportHeight / 2)
		if desiredOffset < 0 {
			desiredOffset = 0
		}

		listView.SetYOffset(desiredOffset)
		m.stressTestState.SetListView(listView)
	}

	// Update detail view
	m.updateStressTestDetailView()
}

// updateStressTestDetailView updates the stress test run details view
func (m *Model) updateStressTestDetailView() {
	var detailContent strings.Builder

	if len(m.stressTestState.GetRuns()) == 0 || m.stressTestState.GetRunIndex() >= len(m.stressTestState.GetRuns()) {
		detailContent.WriteString("No test run selected")
	} else {
		run := m.stressTestState.GetRuns()[m.stressTestState.GetRunIndex()]

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

	detailView := m.stressTestState.GetDetailView()
	detailView.SetContent(detailContent.String())
	m.stressTestState.SetDetailView(detailView)
}

// loadStressTestRuns loads stress test runs from the database
func (m *Model) loadStressTestRuns() tea.Cmd {
	return func() tea.Msg {
		if m.stressTestState.GetManager() == nil {
			return stressTestRunsLoadedMsg{runs: []*stresstest.Run{}}
		}

		// Get active profile name
		profileName := ""
		if profile := m.sessionMgr.GetActiveProfile(); profile != nil {
			profileName = profile.Name
		}

		runs, err := m.stressTestState.GetManager().ListRuns(profileName, 100) // Limit to 100 most recent
		if err != nil {
			return errorMsg(fmt.Sprintf("Failed to load stress test runs: %v", err))
		}

		return stressTestRunsLoadedMsg{runs: runs}
	}
}

// loadStressTestConfigs loads stress test configs from the database
func (m *Model) loadStressTestConfigs() tea.Cmd {
	return func() tea.Msg {
		if m.stressTestState.GetManager() == nil {
			return stressTestConfigsLoadedMsg{configs: []*stresstest.Config{}}
		}

		// Get active profile name
		profileName := ""
		if profile := m.sessionMgr.GetActiveProfile(); profile != nil {
			profileName = profile.Name
		}

		configs, err := m.stressTestState.GetManager().ListConfigs(profileName)
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
