package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// renderStressTestProgress renders the stress test progress modal
func (m *Model) renderStressTestProgress() string {
	modalWidth := m.width - 10
	if modalWidth > 90 {
		modalWidth = 90
	}

	var content strings.Builder

	// Title - show stopping state
	title := "Stress Test - Running"
	if m.stressTestState.GetStopping() {
		title = "Stress Test - Stopping"
	}
	content.WriteString(styleTitle.Render(title) + "\n\n")

	// Request info
	if m.stressTestState.GetActiveRequest() != nil {
		content.WriteString(styleTitleFocused.Render("Request") + "\n")
		content.WriteString(fmt.Sprintf("%s %s\n",
			styleTitleFocused.Render(m.stressTestState.GetActiveRequest().Method),
			m.stressTestState.GetActiveRequest().URL))

		if m.stressTestState.GetActiveRequest().Body != "" {
			bodyPreview := m.stressTestState.GetActiveRequest().Body
			if len(bodyPreview) > 100 {
				bodyPreview = bodyPreview[:97] + "..."
			}
			content.WriteString(styleSubtle.Render("Body: ") + bodyPreview + "\n")
		}
		content.WriteString("\n")
	}

	// Get current stats
	var stats *StressTestStats
	if m.stressTestState.GetExecutor() != nil {
		execStats := m.stressTestState.GetExecutor().GetStats()
		run := m.stressTestState.GetExecutor().GetRun()

		elapsed := time.Since(run.StartedAt)

		stats = &StressTestStats{
			TotalRequests:        execStats.TotalRequests,
			CompletedRequests:    execStats.CompletedRequests,
			SuccessCount:         execStats.SuccessCount,
			ErrorCount:           execStats.ErrorCount,
			ValidationErrorCount: execStats.ValidationErrorCount,
			ActiveWorkers:        execStats.ActiveWorkers,
			AvgDurationMs:        execStats.AvgDurationMs(),
			MinDurationMs:        execStats.Min(),
			MaxDurationMs:        execStats.Max(),
			P50DurationMs:        execStats.P50(),
			P95DurationMs:        execStats.P95(),
			P99DurationMs:        execStats.P99(),
			Elapsed:              elapsed,
		}
	} else {
		// No executor, show empty state
		stats = &StressTestStats{}
	}

	// Progress section
	progress := stats.Progress()
	content.WriteString(styleTitleFocused.Render("Progress") + "\n")
	content.WriteString(fmt.Sprintf("%d/%d requests (%.1f%%)\n", stats.CompletedRequests, stats.TotalRequests, progress))

	// Progress bar
	barWidth := 40
	filled := int(progress / 100.0 * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	content.WriteString(bar + "\n")

	// Elapsed time and active workers
	content.WriteString(fmt.Sprintf("Elapsed: %s\n", formatDuration(stats.Elapsed)))
	content.WriteString(fmt.Sprintf("Active Workers: %d\n", stats.ActiveWorkers))

	// Show stopping message
	if m.stressTestState.GetStopping() {
		stoppingMsg := fmt.Sprintf("Waiting for %d active workers to finish...", stats.ActiveWorkers)
		content.WriteString(styleTitleFocused.Render(stoppingMsg) + "\n")
	}
	content.WriteString("\n")

	// Statistics section
	content.WriteString(styleTitleFocused.Render("Statistics") + "\n")

	// Two-column layout
	leftCol := []string{
		fmt.Sprintf("Success:    %d", stats.SuccessCount),
		fmt.Sprintf("Net Errors: %d", stats.ErrorCount),
		fmt.Sprintf("Val Errors: %d", stats.ValidationErrorCount),
		fmt.Sprintf("Avg:        %.0fms", stats.AvgDurationMs),
		fmt.Sprintf("Min:        %dms", stats.MinDurationMs),
	}

	rightCol := []string{
		fmt.Sprintf("Max:        %dms", stats.MaxDurationMs),
		fmt.Sprintf("P50:        %dms", stats.P50DurationMs),
		fmt.Sprintf("P95:        %dms", stats.P95DurationMs),
		fmt.Sprintf("P99:        %dms", stats.P99DurationMs),
		"",
	}

	// Calculate RPS
	rps := 0.0
	if stats.Elapsed.Seconds() > 0 {
		rps = float64(stats.CompletedRequests) / stats.Elapsed.Seconds()
	}

	for i := 0; i < len(leftCol); i++ {
		line := fmt.Sprintf("%-25s", leftCol[i])
		if i < len(rightCol) {
			line += rightCol[i]
		}
		content.WriteString(line + "\n")
	}

	content.WriteString(fmt.Sprintf("\nRequests/sec: %.2f\n", rps))

	// Instructions
	content.WriteString("\n")
	footer := "ESC/q: Cancel test"
	if m.stressTestState.GetStopping() {
		footer = "Stopping test gracefully... please wait"
	}
	content.WriteString(styleSubtle.Render(footer))

	// Style the modal
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorCyan).
		Padding(1, 2).
		Width(modalWidth)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		modalStyle.Render(content.String()),
	)
}

// StressTestStats holds formatted stats for display
type StressTestStats struct {
	TotalRequests        int
	CompletedRequests    int
	SuccessCount         int
	ErrorCount           int // Network errors
	ValidationErrorCount int // Validation errors
	ActiveWorkers        int
	AvgDurationMs        float64
	MinDurationMs        int64
	MaxDurationMs        int64
	P50DurationMs        int64
	P95DurationMs        int64
	P99DurationMs        int64
	Elapsed              time.Duration
}

// Progress returns the completion percentage
func (s *StressTestStats) Progress() float64 {
	if s.TotalRequests == 0 {
		return 0
	}
	return float64(s.CompletedRequests) / float64(s.TotalRequests) * 100
}

// formatDuration formats a duration for display
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm %ds", minutes, seconds)
}
