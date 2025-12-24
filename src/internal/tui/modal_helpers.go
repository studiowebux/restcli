package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// SplitPaneConfig defines the configuration for a generic split-pane modal
type SplitPaneConfig struct {
	// Modal dimensions
	ModalWidth  int
	ModalHeight int

	// Split view control
	IsSplitView bool // If false, shows only left pane at full width

	// Left pane
	LeftTitle      string
	LeftContent    string
	LeftBorderColor lipgloss.AdaptiveColor
	LeftIsFocused  bool

	// Right pane (only used if IsSplitView is true)
	RightTitle       string
	RightContent     string
	RightBorderColor lipgloss.AdaptiveColor
	RightIsFocused   bool

	// Footer
	Footer string

	// Width ratio for split view (0.0 to 1.0, default 0.5 for equal split)
	// Left pane gets this ratio, right pane gets the remainder
	LeftWidthRatio float64
}

// renderSplitPaneModal renders a generic split-pane modal layout
func renderSplitPaneModal(cfg SplitPaneConfig, totalWidth, totalHeight int) string {
	paneHeight := cfg.ModalHeight - 4 // Account for borders and padding

	var mainView string

	if cfg.IsSplitView {
		// Calculate split widths
		ratio := cfg.LeftWidthRatio
		if ratio <= 0 || ratio >= 1 {
			ratio = 0.5 // Default to equal split
		}

		listWidth := int(float64(cfg.ModalWidth-3) * ratio)
		previewWidth := cfg.ModalWidth - listWidth - 3

		// Determine title styles based on focus
		leftTitleStyle := styleTitleUnfocused
		rightTitleStyle := styleTitleUnfocused
		if cfg.LeftIsFocused {
			leftTitleStyle = styleTitleFocused
		}
		if cfg.RightIsFocused {
			rightTitleStyle = styleTitleFocused
		}

		// Left pane
		leftPane := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cfg.LeftBorderColor).
			Width(listWidth).
			Height(paneHeight).
			Padding(0, 1).
			Render(leftTitleStyle.Render(cfg.LeftTitle) + "\n" + cfg.LeftContent)

		// Right pane
		rightPane := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cfg.RightBorderColor).
			Width(previewWidth).
			Height(paneHeight).
			Padding(0, 1).
			Render(rightTitleStyle.Render(cfg.RightTitle) + "\n" + cfg.RightContent)

		// Join panes horizontally
		mainView = lipgloss.JoinHorizontal(
			lipgloss.Top,
			leftPane,
			rightPane,
		)
	} else {
		// Single pane mode: expand left pane to full width
		leftTitleStyle := styleTitleFocused // Always focused when it's the only pane
		if !cfg.LeftIsFocused {
			leftTitleStyle = styleTitleUnfocused
		}

		mainView = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cfg.LeftBorderColor).
			Width(cfg.ModalWidth).
			Height(paneHeight).
			Padding(0, 1).
			Render(leftTitleStyle.Render(cfg.LeftTitle) + "\n" + cfg.LeftContent)
	}

	// Add footer
	footer := styleSubtle.Render(cfg.Footer)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		mainView,
		"\n"+footer,
	)

	return lipgloss.Place(
		totalWidth,
		totalHeight,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}
