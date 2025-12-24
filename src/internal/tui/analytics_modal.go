package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/studiowebux/restcli/internal/analytics"
)

// renderAnalytics renders the analytics modal with telescope-style split view
func (m *Model) renderAnalytics() string {
	modalWidth := m.width - 6
	modalHeight := m.height - 3

	// Determine border colors based on focus
	listBorderColor := colorGray
	detailBorderColor := colorGray
	leftIsFocused := false
	rightIsFocused := false

	if m.analyticsState.GetFocusedPane() == "list" {
		listBorderColor = colorCyan
		leftIsFocused = true
	} else if m.analyticsState.GetFocusedPane() == "details" {
		detailBorderColor = colorCyan
		rightIsFocused = true
	}

	// When in single pane mode, always use cyan
	if !m.analyticsState.GetPreviewVisible() {
		listBorderColor = colorCyan
		leftIsFocused = true
	}

	// Build detail title with scroll indicator
	detailTitle := "Details"
	scrollIndicator := m.getAnalyticsDetailScrollIndicator()
	if scrollIndicator != "" {
		// We'll render the title manually in the config to include the scroll indicator
		detailTitle = "Details  " + scrollIndicator
	}

	// Build footer with instructions and scroll position
	groupMode := "Per File"
	if m.analyticsState.GetGroupByPath() {
		groupMode = "By Path"
	}
	footerText := fmt.Sprintf("TAB: Switch Focus | ↑/↓ j/k: Nav | Enter: Load | p: Preview | t: Toggle Group (%s) | C: Clear | ESC/q: Close", groupMode)

	// Add scroll indicator if there are stats
	if len(m.analyticsState.GetStats()) > 0 {
		current := m.analyticsState.GetIndex() + 1
		total := len(m.analyticsState.GetStats())
		percentage := int(float64(current) / float64(total) * 100)
		scrollInfo := fmt.Sprintf(" [%d/%d] (%d%%)", current, total, percentage)
		footerText += scrollInfo
	}

	// Configure split-pane modal
	cfg := SplitPaneConfig{
		ModalWidth:       modalWidth,
		ModalHeight:      modalHeight,
		IsSplitView:      m.analyticsState.GetPreviewVisible(),
		LeftTitle:        "Analytics",
		LeftContent:      m.analyticsState.GetListView().View(),
		LeftBorderColor:  listBorderColor,
		LeftIsFocused:    leftIsFocused,
		RightTitle:       detailTitle,
		RightContent:     m.analyticsState.GetDetailView().View(),
		RightBorderColor: detailBorderColor,
		RightIsFocused:   rightIsFocused,
		Footer:           footerText,
		LeftWidthRatio:   0.5, // Equal split
	}

	return renderSplitPaneModal(cfg, m.width, m.height)
}

// updateAnalyticsView updates the analytics viewport content for split view
func (m *Model) updateAnalyticsView() {
	// Calculate dimensions
	modalWidth := m.width - 6
	modalHeight := m.height - 3
	paneHeight := modalHeight - 4

	// Adjust viewport widths based on preview visibility
	if m.analyticsState.GetPreviewVisible() {
		// Split view mode: calculate widths for both panes
		listWidth := (modalWidth - 3) / 2
		previewWidth := modalWidth - listWidth - 3

		// Set viewport dimensions for left pane (analytics list)
		listView := m.analyticsState.GetListView()
		listView.Width = listWidth - 4
		listView.Height = paneHeight - 2
		m.analyticsState.SetListView(listView)

		// Set viewport dimensions for right pane (analytics detail)
		detailView := m.analyticsState.GetDetailView()
		detailView.Width = previewWidth - 4
		detailView.Height = paneHeight - 2
		m.analyticsState.SetDetailView(detailView)
	} else {
		// Preview hidden: expand list to full width
		listView := m.analyticsState.GetListView()
		listView.Width = modalWidth - 4
		listView.Height = paneHeight - 2
		m.analyticsState.SetListView(listView)
	}

	// Build content for left pane (analytics list)
	var listContent strings.Builder
	if len(m.analyticsState.GetStats()) == 0 {
		listContent.WriteString("No analytics data available.\n\nEnable analytics in your profile to start tracking:\n\"analyticsEnabled\": true")
	} else {
		// Show ALL entries - viewport handles scrolling
		for i, stat := range m.analyticsState.GetStats() {
			// Format display name
			displayName := filepath.Base(stat.FilePath)
			if m.analyticsState.GetGroupByPath() {
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
			if i == m.analyticsState.GetIndex() {
				line = styleSelected.Render("> " + line)
			} else {
				line = "  " + line
			}

			listContent.WriteString(line + "\n")
		}
	}

	listView := m.analyticsState.GetListView()
	listView.SetContent(listContent.String())
	m.analyticsState.SetListView(listView)

	// Auto-scroll to keep selected item visible
	if len(m.analyticsState.GetStats()) > 0 && m.analyticsState.GetIndex() >= 0 && m.analyticsState.GetIndex() < len(m.analyticsState.GetStats()) {
		// Calculate the line position (0-indexed)
		linePos := m.analyticsState.GetIndex()

		// Get viewport height
		listView := m.analyticsState.GetListView()
		viewportHeight := listView.Height

		// Calculate desired Y offset to keep selected item centered
		desiredOffset := linePos - (viewportHeight / 2)
		if desiredOffset < 0 {
			desiredOffset = 0
		}

		listView.SetYOffset(desiredOffset)
		m.analyticsState.SetListView(listView)
	}

	// Build content for right pane (analytics detail)
	var detailContent strings.Builder
	if len(m.analyticsState.GetStats()) == 0 || m.analyticsState.GetIndex() >= len(m.analyticsState.GetStats()) {
		detailContent.WriteString("No analytics selected")
	} else {
		stat := m.analyticsState.GetStats()[m.analyticsState.GetIndex()]

		// Header
		detailContent.WriteString(styleTitle.Render(fmt.Sprintf("%s %s", stat.Method, stat.NormalizedPath)) + "\n\n")

		// File path (if grouping by path)
		if m.analyticsState.GetGroupByPath() && stat.FilePath != "" {
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

	detailView := m.analyticsState.GetDetailView()
	detailView.SetContent(detailContent.String())
	detailView.GotoTop()
	m.analyticsState.SetDetailView(detailView)
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
	if !m.analyticsState.GetPreviewVisible() {
		return ""
	}

	// Count total lines in viewport content
	totalLines := strings.Count(m.analyticsState.GetDetailView().View(), "\n") + 1
	visibleLines := m.analyticsState.GetDetailView().Height

	// No scroll indicator if all content fits in viewport
	if totalLines <= visibleLines {
		return ""
	}

	// Calculate scroll percentage
	currentLine := m.analyticsState.GetDetailView().YOffset
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

		if m.analyticsState.GetGroupByPath() {
			stats, err = m.analyticsState.GetManager().GetStatsPerNormalizedPath(profileName)
		} else {
			stats, err = m.analyticsState.GetManager().GetStatsPerFile(profileName)
		}

		if err != nil {
			return errorMsg(fmt.Sprintf("Failed to load analytics: %v", err))
		}

		return analyticsLoadedMsg{stats: stats}
	}
}

// renderAnalyticsClearConfirmation renders the confirmation modal for clearing all analytics
func (m *Model) renderAnalyticsClearConfirmation() string {
	count := len(m.analyticsState.GetStats())
	content := "WARNING\n\n"
	content += "This will permanently delete ALL analytics data.\n\n"
	content += fmt.Sprintf("Total entries to delete: %d\n\n", count)
	content += "This action cannot be undone!\n\n"
	content += "Are you sure you want to continue?"

	footer := "[y]es [n]o/ESC"
	return m.renderModalWithFooter("Clear All Analytics", content, footer, 60, 14)
}
