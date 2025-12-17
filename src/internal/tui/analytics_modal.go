package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/studiowebux/restcli/internal/analytics"
)

// renderAnalytics renders the analytics modal with telescope-style split view
func (m *Model) renderAnalytics() string {
	modalWidth := m.width - 6
	modalHeight := m.height - 3
	paneHeight := modalHeight - 4

	var mainView string

	if m.analyticsPreviewVisible {
		// Split view mode: show both list and detail
		listWidth := (modalWidth - 3) / 2
		previewWidth := modalWidth - listWidth - 3

		// Determine border colors and title styles based on focus
		listBorderColor := colorGray
		detailBorderColor := colorGray
		listTitleStyle := styleTitleUnfocused
		detailTitleStyle := styleTitleUnfocused
		if m.analyticsFocusedPane == "list" {
			listBorderColor = colorCyan
			listTitleStyle = styleTitleFocused
		} else if m.analyticsFocusedPane == "details" {
			detailBorderColor = colorCyan
			detailTitleStyle = styleTitleFocused
		}

		// Left pane: Analytics list
		leftPane := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(listBorderColor).
			Width(listWidth).
			Height(paneHeight).
			Padding(0, 1).
			Render(listTitleStyle.Render("Analytics") + "\n" + m.analyticsListView.View())

		// Right pane: Stats detail with scroll indicator
		detailTitle := detailTitleStyle.Render("Details")
		scrollIndicator := m.getAnalyticsDetailScrollIndicator()
		if scrollIndicator != "" {
			detailTitle = lipgloss.JoinHorizontal(lipgloss.Top, detailTitle, "  ", styleSubtle.Render(scrollIndicator))
		}

		rightPane := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(detailBorderColor).
			Width(previewWidth).
			Height(paneHeight).
			Padding(0, 1).
			Render(detailTitle + "\n" + m.analyticsDetailView.View())

		mainView = lipgloss.JoinHorizontal(
			lipgloss.Top,
			leftPane,
			rightPane,
		)
	} else {
		// Preview hidden: expand list to full width
		listBorderColor := colorCyan // Always cyan when list is the only visible pane
		listTitleStyle := styleTitleFocused // Always focused when it's the only pane
		mainView = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(listBorderColor).
			Width(modalWidth).
			Height(paneHeight).
			Padding(0, 1).
			Render(listTitleStyle.Render("Analytics") + "\n" + m.analyticsListView.View())
	}

	// Build footer with instructions and scroll position
	groupMode := "Per File"
	if m.analyticsGroupByPath {
		groupMode = "By Path"
	}
	footerText := fmt.Sprintf("TAB: Switch Focus | ↑/↓ j/k: Nav | Enter: Load | p: Preview | t: Toggle Group (%s) | C: Clear | ESC/q: Close", groupMode)

	// Add scroll indicator if there are stats
	if len(m.analyticsStats) > 0 {
		current := m.analyticsIndex + 1
		total := len(m.analyticsStats)
		percentage := int(float64(current) / float64(total) * 100)
		scrollInfo := fmt.Sprintf(" [%d/%d] (%d%%)", current, total, percentage)
		footerText += scrollInfo
	}

	footer := styleSubtle.Render(footerText)

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

// updateAnalyticsView updates the analytics viewport content for split view
func (m *Model) updateAnalyticsView() {
	// Calculate dimensions
	modalWidth := m.width - 6
	modalHeight := m.height - 3
	paneHeight := modalHeight - 4

	// Adjust viewport widths based on preview visibility
	if m.analyticsPreviewVisible {
		// Split view mode: calculate widths for both panes
		listWidth := (modalWidth - 3) / 2
		previewWidth := modalWidth - listWidth - 3

		// Set viewport dimensions for left pane (analytics list)
		m.analyticsListView.Width = listWidth - 4
		m.analyticsListView.Height = paneHeight - 2

		// Set viewport dimensions for right pane (analytics detail)
		m.analyticsDetailView.Width = previewWidth - 4
		m.analyticsDetailView.Height = paneHeight - 2
	} else {
		// Preview hidden: expand list to full width
		m.analyticsListView.Width = modalWidth - 4
		m.analyticsListView.Height = paneHeight - 2
	}

	// Build content for left pane (analytics list)
	var listContent strings.Builder
	if len(m.analyticsStats) == 0 {
		listContent.WriteString("No analytics data available.\n\nEnable analytics in your profile to start tracking:\n\"analyticsEnabled\": true")
	} else {
		// Show ALL entries - viewport handles scrolling
		for i, stat := range m.analyticsStats {
			// Format display name
			displayName := filepath.Base(stat.FilePath)
			if m.analyticsGroupByPath {
				displayName = stat.NormalizedPath
			}

			// Calculate success rate
			successRate := 0.0
			if stat.TotalCalls > 0 {
				successRate = float64(stat.SuccessCount) / float64(stat.TotalCalls) * 100
			}

			// Format line
			line := fmt.Sprintf("%s %s | Calls: %d | Success: %.1f%% | Avg: %.0fms",
				stat.Method,
				displayName,
				stat.TotalCalls,
				successRate,
				stat.AvgDurationMs,
			)

			// Highlight selected
			if i == m.analyticsIndex {
				line = styleSelected.Render("> " + line)
			} else {
				line = "  " + line
			}

			listContent.WriteString(line + "\n")
		}
	}

	m.analyticsListView.SetContent(listContent.String())

	// Auto-scroll to keep selected item visible
	if len(m.analyticsStats) > 0 && m.analyticsIndex >= 0 && m.analyticsIndex < len(m.analyticsStats) {
		// Calculate the line position (0-indexed)
		linePos := m.analyticsIndex

		// Get viewport height
		viewportHeight := m.analyticsListView.Height

		// Calculate desired Y offset to keep selected item centered
		desiredOffset := linePos - (viewportHeight / 2)
		if desiredOffset < 0 {
			desiredOffset = 0
		}

		m.analyticsListView.SetYOffset(desiredOffset)
	}

	// Build content for right pane (analytics detail)
	var detailContent strings.Builder
	if len(m.analyticsStats) == 0 || m.analyticsIndex >= len(m.analyticsStats) {
		detailContent.WriteString("No analytics selected")
	} else {
		stat := m.analyticsStats[m.analyticsIndex]

		// Header
		detailContent.WriteString(styleTitle.Render(fmt.Sprintf("%s %s", stat.Method, stat.NormalizedPath)) + "\n\n")

		// File path (if grouping by path)
		if m.analyticsGroupByPath && stat.FilePath != "" {
			detailContent.WriteString(styleSubtle.Render("File: ") + filepath.Base(stat.FilePath) + "\n\n")
		}

		// Summary stats
		detailContent.WriteString(styleTitle.Render("Summary") + "\n")
		detailContent.WriteString(fmt.Sprintf("Total Calls:    %d\n", stat.TotalCalls))
		detailContent.WriteString(fmt.Sprintf("Success:        %d (%.1f%%)\n",
			stat.SuccessCount,
			float64(stat.SuccessCount)/float64(stat.TotalCalls)*100))
		detailContent.WriteString(fmt.Sprintf("Errors:         %d (%.1f%%)\n",
			stat.ErrorCount,
			float64(stat.ErrorCount)/float64(stat.TotalCalls)*100))
		detailContent.WriteString(fmt.Sprintf("Network Errors: %d (%.1f%%)\n\n",
			stat.NetworkErrors,
			float64(stat.NetworkErrors)/float64(stat.TotalCalls)*100))

		// Timing stats
		detailContent.WriteString(styleTitle.Render("Timing") + "\n")
		detailContent.WriteString(fmt.Sprintf("Average:        %.0fms\n", stat.AvgDurationMs))
		detailContent.WriteString(fmt.Sprintf("Min:            %dms\n", stat.MinDurationMs))
		detailContent.WriteString(fmt.Sprintf("Max:            %dms\n\n", stat.MaxDurationMs))

		// Data transfer
		detailContent.WriteString(styleTitle.Render("Data Transfer") + "\n")
		detailContent.WriteString(fmt.Sprintf("Total Req:      %s\n", formatBytes(stat.TotalReqSize)))
		detailContent.WriteString(fmt.Sprintf("Total Resp:     %s\n", formatBytes(stat.TotalRespSize)))
		detailContent.WriteString(fmt.Sprintf("Avg Req:        %s\n", formatBytes(stat.TotalReqSize/int64(stat.TotalCalls))))
		detailContent.WriteString(fmt.Sprintf("Avg Resp:       %s\n\n", formatBytes(stat.TotalRespSize/int64(stat.TotalCalls))))

		// Status codes
		detailContent.WriteString(styleTitle.Render("Status Codes") + "\n")
		for code, count := range stat.StatusCodes {
			percentage := float64(count) / float64(stat.TotalCalls) * 100
			detailContent.WriteString(fmt.Sprintf("%d:             %d (%.1f%%)\n", code, count, percentage))
		}
		detailContent.WriteString("\n")

		// Last called
		detailContent.WriteString(styleTitle.Render("Last Called") + "\n")
		detailContent.WriteString(formatRelativeTime(stat.LastCalled) + "\n")
	}

	m.analyticsDetailView.SetContent(detailContent.String())
	m.analyticsDetailView.GotoTop()
}

// formatBytes formats bytes into human-readable format
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// formatRelativeTime formats a time relative to now
func formatRelativeTime(t time.Time) string {
	duration := time.Since(t)

	switch {
	case duration < time.Minute:
		return "just now"
	case duration < time.Hour:
		mins := int(duration.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case duration < 7*24*time.Hour:
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		return t.Format("2006-01-02 15:04")
	}
}

// getAnalyticsDetailScrollIndicator returns a scroll position indicator for the analytics detail viewport
func (m Model) getAnalyticsDetailScrollIndicator() string {
	// Only show indicator when preview is visible and content is scrollable
	if !m.analyticsPreviewVisible {
		return ""
	}

	// Count total lines in viewport content
	totalLines := strings.Count(m.analyticsDetailView.View(), "\n") + 1
	visibleLines := m.analyticsDetailView.Height

	// No scroll indicator if all content fits in viewport
	if totalLines <= visibleLines {
		return ""
	}

	// Calculate scroll percentage
	currentLine := m.analyticsDetailView.YOffset
	scrollableLines := totalLines - visibleLines

	var percentage int
	if scrollableLines > 0 {
		percentage = (currentLine * 100) / scrollableLines
	}

	// Clamp percentage to 0-100 range
	if percentage < 0 {
		percentage = 0
	} else if percentage > 100 {
		percentage = 100
	}

	// Format indicator with position info
	return fmt.Sprintf("[%d%%] %d/%d", percentage, currentLine+visibleLines, totalLines)
}

// loadAnalytics loads analytics stats from the database
func (m *Model) loadAnalytics() tea.Cmd {
	return func() tea.Msg {
		if m.analyticsManager == nil {
			return analyticsLoadedMsg{stats: []analytics.Stats{}}
		}

		// Get active profile name
		profileName := ""
		if profile := m.sessionMgr.GetActiveProfile(); profile != nil {
			profileName = profile.Name
		}

		var stats []analytics.Stats
		var err error

		if m.analyticsGroupByPath {
			stats, err = m.analyticsManager.GetStatsPerNormalizedPath(profileName)
		} else {
			stats, err = m.analyticsManager.GetStatsPerFile(profileName)
		}

		if err != nil {
			return errorMsg(fmt.Sprintf("Failed to load analytics: %v", err))
		}

		return analyticsLoadedMsg{stats: stats}
	}
}

// renderAnalyticsClearConfirmation renders the confirmation modal for clearing all analytics
func (m *Model) renderAnalyticsClearConfirmation() string {
	count := len(m.analyticsStats)
	content := "WARNING\n\n"
	content += "This will permanently delete ALL analytics data.\n\n"
	content += fmt.Sprintf("Total entries to delete: %d\n\n", count)
	content += "This action cannot be undone!\n\n"
	content += "Are you sure you want to continue?"

	footer := "[y]es [n]o/ESC"
	return m.renderModalWithFooter("Clear All Analytics", content, footer, 60, 14)
}
