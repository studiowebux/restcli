package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/lipgloss"
	"github.com/studiowebux/restcli/internal/chain"
	"github.com/studiowebux/restcli/internal/config"
	"github.com/studiowebux/restcli/internal/executor"
	"github.com/studiowebux/restcli/internal/parser"
	"github.com/studiowebux/restcli/internal/types"
)

// Color definitions using terminal's native color palette with adaptive brightness
// Light mode uses regular colors (0-7), dark mode uses bright colors (8-15) for better contrast
var (
	colorGreen  = lipgloss.AdaptiveColor{Light: "2", Dark: "10"} // Green / Bright green
	colorRed    = lipgloss.AdaptiveColor{Light: "1", Dark: "9"}  // Red / Bright red
	colorYellow = lipgloss.AdaptiveColor{Light: "3", Dark: "11"} // Yellow / Bright yellow
	colorBlue   = lipgloss.AdaptiveColor{Light: "4", Dark: "12"} // Blue / Bright blue
	colorGray   = lipgloss.AdaptiveColor{Light: "8", Dark: "8"}  // Bright black (gray) for both
	colorCyan   = lipgloss.AdaptiveColor{Light: "6", Dark: "14"} // Cyan / Bright cyan
)

// Style definitions
var (
	styleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorCyan)

	styleTitleFocused = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorCyan)

	styleTitleUnfocused = lipgloss.NewStyle().
				Foreground(colorGray)

	styleSelected = lipgloss.NewStyle().
			Background(lipgloss.AdaptiveColor{Light: "7", Dark: "8"}). // Bright white / Bright black
			Foreground(lipgloss.AdaptiveColor{Light: "0", Dark: "15"}) // Black / White

	styleSuccess = lipgloss.NewStyle().
			Foreground(colorGreen)

	styleError = lipgloss.NewStyle().
			Foreground(colorRed)

	styleWarning = lipgloss.NewStyle().
			Foreground(colorYellow)

	styleSearchMatch = lipgloss.NewStyle().
			Foreground(colorYellow)

	styleSubtle = lipgloss.NewStyle().
			Foreground(colorGray)

	styleSearchHighlight = lipgloss.NewStyle().
				Background(lipgloss.AdaptiveColor{Light: "5", Dark: "13"}). // Magenta / Bright magenta background
				Foreground(lipgloss.AdaptiveColor{Light: "0", Dark: "0"})   // Black text for both modes

	// Diff background styles for split view highlighting
	styleDiffRemoved = lipgloss.NewStyle().
				Background(lipgloss.AdaptiveColor{Light: "1", Dark: "9"}). // Red / Bright red
				Foreground(lipgloss.AdaptiveColor{Light: "15", Dark: "0"}) // White / Black

	styleDiffAdded = lipgloss.NewStyle().
			Background(lipgloss.AdaptiveColor{Light: "2", Dark: "10"}). // Green / Bright green
			Foreground(lipgloss.AdaptiveColor{Light: "15", Dark: "0"})  // White / Black

	styleDiffNeutral = lipgloss.NewStyle().
				Foreground(colorGray)

	// HTTP method styles
	styleMethodGET     = lipgloss.NewStyle().Foreground(colorBlue)
	styleMethodPOST    = lipgloss.NewStyle().Foreground(colorGreen)
	styleMethodPUT     = lipgloss.NewStyle().Foreground(colorYellow)
	styleMethodPATCH   = lipgloss.NewStyle().Foreground(colorYellow)
	styleMethodDELETE  = lipgloss.NewStyle().Foreground(colorRed)
	styleMethodHEAD    = lipgloss.NewStyle().Foreground(colorGray)
	styleMethodOPTIONS = lipgloss.NewStyle().Foreground(colorGray)
	styleMethodWS      = lipgloss.NewStyle().Foreground(colorCyan)
)

// getMethodStyle returns the appropriate style for an HTTP method or protocol
func getMethodStyle(method string) lipgloss.Style {
	switch method {
	case "GET":
		return styleMethodGET
	case "POST":
		return styleMethodPOST
	case "PUT":
		return styleMethodPUT
	case "PATCH":
		return styleMethodPATCH
	case "DELETE":
		return styleMethodDELETE
	case "HEAD":
		return styleMethodHEAD
	case "OPTIONS":
		return styleMethodOPTIONS
	case "WS":
		return styleMethodWS
	default:
		return lipgloss.NewStyle()
	}
}

// highlightJSON applies syntax highlighting to JSON content
// Uses configured syntax themes from the profile, or sensible defaults
func highlightJSON(jsonStr string, profile *types.Profile) string {
	lexer := lexers.Get("json")
	if lexer == nil {
		return jsonStr
	}
	lexer = chroma.Coalesce(lexer)

	// Determine which theme to use based on terminal background
	// Use lipgloss's built-in background detection
	isDark := lipgloss.HasDarkBackground()
	var styleName string

	if isDark {
		// Dark background - use configured dark theme or default to monokai
		styleName = "monokai"
		if profile != nil && profile.SyntaxThemeDark != "" {
			styleName = profile.SyntaxThemeDark
		}
	} else {
		// Light background - use configured light theme or default to github
		styleName = "github"
		if profile != nil && profile.SyntaxThemeLight != "" {
			styleName = profile.SyntaxThemeLight
		}
	}

	style := styles.Get(styleName)
	if style == nil {
		style = styles.Fallback
	}

	// Use terminal256 formatter for wide terminal support
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		return jsonStr
	}

	iterator, err := lexer.Tokenise(nil, jsonStr)
	if err != nil {
		return jsonStr
	}

	var buf bytes.Buffer
	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		return jsonStr
	}

	return buf.String()
}

// renderMain renders the main TUI view (file sidebar + response panel)
func (m Model) renderMain() string {
	if m.width == 0 {
		return ""
	}

	// Fullscreen mode - show only response
	if m.fullscreen {
		response := m.renderResponse(m.width-ViewportPaddingHorizontal, m.height-MainViewHeightOffset)
		responseBox := lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(colorCyan). // Cyan for focused (fullscreen) panel
			Width(m.width - MinimalBorderMargin).
			Height(m.height - ModalHeightMargin). // Leave room for status bar + top border visibility
			Padding(0).           // No padding inside box border
			AlignVertical(lipgloss.Top).
			Render(response)

		statusBar := m.renderStatusBar()

		return lipgloss.JoinVertical(
			lipgloss.Left,
			responseBox,
			statusBar,
		)
	}

	// Normal mode - show sidebar and response
	// Calculate dimensions - make sidebar wider (40% of width or min 40 chars)
	sidebarWidth := max(40, m.width*40/100)
	if m.width < 100 {
		sidebarWidth = m.width / 2
	}
	responseWidth := m.width - sidebarWidth - ViewportPaddingHorizontal // Account for borders

	// Render components with borders
	sidebar := m.renderSidebar(sidebarWidth-MinimalBorderMargin, m.height-MainViewHeightOffset) // -5 = -1 (status) -2 (borders) -2 (top visibility)
	response := m.renderResponse(responseWidth-MinimalBorderMargin, m.height-MainViewHeightOffset)

	// Add borders - highlight the focused panel with cyan, unfocused with gray
	sidebarBorderColor := colorGray
	responseBorderColor := colorGray
	if m.focusedPanel == "sidebar" {
		sidebarBorderColor = colorCyan
	} else if m.focusedPanel == "response" {
		responseBorderColor = colorCyan
	}

	sidebarBox := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(sidebarBorderColor).
		Width(sidebarWidth).
		Height(m.height - ModalHeightMargin). // Leave room for status bar + top border visibility
		Padding(0).           // No padding inside box border
		AlignVertical(lipgloss.Top).
		Render(sidebar)

	responseBox := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(responseBorderColor).
		Width(responseWidth).
		Height(m.height - ModalHeightMargin). // Leave room for status bar + top border visibility
		Padding(0).           // No padding inside box border
		AlignVertical(lipgloss.Top).
		Render(response)

	// Combine sidebar and response
	mainView := lipgloss.JoinHorizontal(
		lipgloss.Top,
		sidebarBox,
		responseBox,
	)

	// Status bar
	statusBar := m.renderStatusBar()

	// Combine main view and status bar
	return lipgloss.JoinVertical(
		lipgloss.Left,
		mainView,
		statusBar,
	)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// isBinaryContent checks if the content appears to be binary data
func isBinaryContent(s string) bool {
	// Check if content is valid UTF-8
	if !utf8.ValidString(s) {
		return true
	}

	// Check for significant amount of non-printable characters
	// Allow some control chars like \n, \r, \t
	nonPrintableCount := 0
	totalChars := 0
	for _, r := range s {
		totalChars++
		// Consider anything below space (except \n, \r, \t) or above ~ as non-printable
		if (r < ' ' && r != '\n' && r != '\r' && r != '\t') || r > '~' && r < 128 {
			nonPrintableCount++
		}
	}

	// If more than 10% non-printable, consider it binary
	if totalChars > 0 && float64(nonPrintableCount)/float64(totalChars) > 0.1 {
		return true
	}

	return false
}

// renderSidebar renders the file list sidebar
func (m Model) renderSidebar(width, height int) string {
	var lines []string

	// Title - bold if focused
	titleStyle := styleTitleUnfocused
	if m.focusedPanel == "sidebar" {
		titleStyle = styleTitleFocused
	}
	title := "Files"
	tagFilter := m.fileExplorer.GetTagFilter()
	if len(tagFilter) > 0 {
		title = fmt.Sprintf("Files (%s)", strings.Join(tagFilter, ","))
	}
	lines = append(lines, titleStyle.Render(title))
	lines = append(lines, "")

	// File list
	pageSize := height - ViewportPaddingHorizontal // Reserve space for title, blank lines, footer, and padding
	if pageSize < 1 {
		pageSize = 1
	}

	files := m.fileExplorer.GetFiles()
	fileOffset := m.fileExplorer.GetScrollOffset()
	endIdx := fileOffset + pageSize
	if endIdx > len(files) {
		endIdx = len(files)
	}

	for i := fileOffset; i < endIdx; i++ {
		file := files[i]

		// Hex number
		hexNum := fmt.Sprintf("%x", i)

		// HTTP method with color (if available)
		methodPrefix := ""
		methodLen := 0
		if file.HTTPMethod != "" {
			methodStyle := getMethodStyle(file.HTTPMethod)
			methodPrefix = methodStyle.Render(file.HTTPMethod) + " "
			methodLen = len(file.HTTPMethod) + 1 // +1 for space
		}

		// Check if this file is part of a chain
		chainIndicator := ""
		chainLen := 0
		// Parse file to check for dependencies or extractions
		if requests, err := parser.Parse(file.Path); err == nil && len(requests) > 0 {
			req := &requests[0]
			if len(req.DependsOn) > 0 || len(req.Extract) > 0 {
				chainIndicator = " [CHAIN]"
				chainLen = len(chainIndicator)
			}
		}

		// File name (truncate if too long)
		// Reserve space for tags if they exist
		tagsSuffix := ""
		tagsLen := 0
		if len(file.Tags) > 0 {
			// Show first 2 tags
			displayTags := file.Tags
			if len(displayTags) > 2 {
				displayTags = displayTags[:2]
			}
			tagsSuffix = " [" + strings.Join(displayTags, ",") + "]"
			if len(file.Tags) > 2 {
				tagsSuffix += "..."
			}
			tagsLen = len(tagsSuffix)
		}

		maxNameLen := width - len(hexNum) - methodLen - tagsLen - chainLen - 4
		if maxNameLen < 10 {
			maxNameLen = 10
		}
		name := file.Name
		if len(name) > maxNameLen {
			name = name[:maxNameLen-3] + "..."
		}

		line := fmt.Sprintf("%s %s%s%s%s", hexNum, methodPrefix, name, styleSubtle.Render(chainIndicator), styleSubtle.Render(tagsSuffix))

		// Apply styling - selected gets green, search matches get yellow
		fileIndex := m.fileExplorer.GetCurrentIndex()
		searchMatches := m.fileExplorer.GetSearchMatches()
		isMatch := false
		for _, matchIdx := range searchMatches {
			if matchIdx == i {
				isMatch = true
				break
			}
		}

		if i == fileIndex {
			line = styleSelected.Render(line)
		} else if isMatch {
			line = styleSearchMatch.Render(line)
		}

		lines = append(lines, line)
	}

	// Footer - show position
	if len(files) > 0 {
		lines = append(lines, "")
		fileIndex := m.fileExplorer.GetCurrentIndex()
		footer := fmt.Sprintf("[%d/%d]", fileIndex+1, len(files))
		lines = append(lines, styleSubtle.Render(footer))
	} else {
		lines = append(lines, "")
		lines = append(lines, styleSubtle.Render("No files found"))
	}

	// Join lines and apply width
	content := strings.Join(lines, "\n")
	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(0, 1) // No vertical padding, only horizontal

	return style.Render(content)
}

// getScrollIndicator returns a scroll position indicator for the response viewport
func (m Model) getScrollIndicator() string {
	// Only show indicator when body is visible and content is scrollable
	if !m.showBody || m.currentResponse == nil {
		return ""
	}

	// Count total lines in viewport content
	totalLines := strings.Count(m.responseContent, "\n") + 1
	visibleLines := m.responseView.Height

	// No scroll indicator if all content fits in viewport
	if totalLines <= visibleLines {
		return ""
	}

	// Calculate scroll percentage
	currentLine := m.responseView.YOffset
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

// renderResponse renders the response panel
func (m Model) renderResponse(width, height int) string {
	// Add title at the top - bold if focused
	titleStyle := styleTitleUnfocused
	if m.focusedPanel == "response" {
		titleStyle = styleTitleFocused
	}
	title := titleStyle.Render("Response")

	// Add scroll indicator if viewport has scrollable content
	scrollIndicator := m.getScrollIndicator()
	if scrollIndicator != "" {
		// Add scroll indicator on the same line as title
		titleWithScroll := lipgloss.JoinHorizontal(lipgloss.Top, title, "  ", styleSubtle.Render(scrollIndicator))
		title = titleWithScroll
	}

	if m.currentResponse == nil {
		// If loading, show viewport content (which has the loading indicator)
		if m.loading {
			var content strings.Builder
			content.WriteString(title + "\n\n")
			viewportContent := m.responseView.View()
			// Add viewport content
			content.WriteString(viewportContent)

			// Apply style with height constraint (like sidebar does)
			style := lipgloss.NewStyle().
				Width(width).
				Height(height).
				Padding(0, 1) // No vertical padding, only horizontal
			return style.Render(content.String())
		}
		// Otherwise show empty state message
		noResponse := styleSubtle.Render("No response yet\n\nPress Enter to execute request\n\nPress 'b' to toggle body visibility")
		contentStr := title + "\n\n" + noResponse

		// Apply style with height constraint
		style := lipgloss.NewStyle().
			Width(width).
			Height(height).
			Padding(0, 1) // No vertical padding, only horizontal
		return style.Render(contentStr)
	}

	// If body is shown, use viewport for scrolling
	if m.showBody {
		// Return viewport content with title
		var content strings.Builder
		content.WriteString(title + "\n\n")
		viewportContent := m.responseView.View()
		content.WriteString(viewportContent)

		// Apply style with height constraint (like sidebar does)
		style := lipgloss.NewStyle().
			Width(width).
			Height(height).
			Padding(0, 1) // No vertical padding, only horizontal
		return style.Render(content.String())
	}

	// Otherwise show just status and headers (no scrolling needed)
	var lines []string
	lines = append(lines, title)
	lines = append(lines, "")

	// Status line
	statusStyle := styleSuccess
	if m.currentResponse.Status >= 400 {
		statusStyle = styleError
	} else if m.currentResponse.Status >= 300 {
		statusStyle = styleWarning
	}

	statusLine := fmt.Sprintf("%s - %s",
		statusStyle.Render(fmt.Sprintf("%d", m.currentResponse.Status)),
		m.currentResponse.StatusText)
	lines = append(lines, statusLine)

	// Timing info
	timingParts := []string{
		fmt.Sprintf("Duration: %s", executor.FormatDuration(m.currentResponse.Duration)),
		fmt.Sprintf("Size: %s", executor.FormatSize(m.currentResponse.ResponseSize)),
	}
	if m.currentResponse.Timestamp != "" {
		timingParts = append(timingParts, fmt.Sprintf("Time: %s", m.currentResponse.Timestamp))
	}
	timing := strings.Join(timingParts, " | ")
	lines = append(lines, styleSubtle.Render(timing))
	lines = append(lines, "")

	// Headers (if enabled)
	if m.showHeaders && len(m.currentResponse.Headers) > 0 {
		lines = append(lines, styleTitle.Render("Headers:"))
		for key, value := range m.currentResponse.Headers {
			headerLine := fmt.Sprintf("%s: %s", key, value)
			lines = append(lines, headerLine)
		}
		lines = append(lines, "")
	}

	lines = append(lines, styleSubtle.Render("Press 'b' to show body"))

	content := strings.Join(lines, "\n")
	return lipgloss.NewStyle().
		MaxWidth(width).
		Height(height).
		Padding(0, 1). // No vertical padding, only horizontal
		AlignHorizontal(lipgloss.Left).
		Render(content)
}

// addCursor adds a visible cursor (█) to a text string
func addCursor(text string) string {
	return text + "█"
}

// renderStatusBar renders the status bar at the bottom
func (m Model) renderStatusBar() string {
	profile := m.sessionMgr.GetActiveProfile()

	// Left side - profile
	left := fmt.Sprintf("Profile: %s", profile.Name)

	// Right side - messages or input
	right := ""

	// Show filter input if editing (takes precedence)
	if m.filterEditing {
		right = fmt.Sprintf("Filter: %s", addCursorAt(m.filterInput, m.filterCursor))
		if m.filterError != "" {
			right += " " + styleError.Render(m.filterError)
		} else if m.statusMsg != "" {
			// Show status message (e.g., "Bookmark saved")
			if strings.Contains(m.statusMsg, "✓") || strings.Contains(m.statusMsg, "saved") {
				right += " " + styleSuccess.Render(m.statusMsg)
			} else if strings.Contains(m.statusMsg, "exists") {
				right += " " + styleWarning.Render(m.statusMsg)
			} else {
				right += " " + m.statusMsg
			}
		}
		// Center spacing
		spacing := m.width - lipgloss.Width(left) - lipgloss.Width(right)
		if spacing < 1 {
			spacing = 1
		}
		return left + strings.Repeat(" ", spacing) + right
	}

	// Show input for special modes
	switch m.mode {
	case ModeGoto:
		right = fmt.Sprintf("Goto: :%s", addCursor(m.gotoInput))
	case ModeSearch:
		right = fmt.Sprintf("Search: %s", addCursor(m.searchInput))
	case ModeTagFilter:
		// Build cursor string manually for category filter
		cursorStr := m.inputValue[:m.inputCursor] + "█" + m.inputValue[m.inputCursor:]
		right = fmt.Sprintf("Category: %s", cursorStr)
	default:
		// Show search results if active (check both file and response search)
		_, _, fileMatches := m.fileExplorer.GetSearchInfo()
		hasSearch := fileMatches > 0 || len(m.responseSearchMatches) > 0

		if m.errorMsg != "" {
			right += styleError.Render(m.errorMsg)
		} else if m.statusMsg != "" {
			// Make success messages green
			if strings.Contains(m.statusMsg, "saved") || strings.Contains(m.statusMsg, "copied") ||
				strings.Contains(m.statusMsg, "Created") || strings.Contains(m.statusMsg, "Switched") ||
				strings.Contains(m.statusMsg, "Renamed") || strings.Contains(m.statusMsg, "Match") {
				right += styleSuccess.Render(m.statusMsg)
			} else {
				right += m.statusMsg
			}
			// Add hint if status message is truncated
			if len(m.fullStatusMsg) > 100 {
				right += styleSubtle.Render(" [press 'I' for full message]")
			}
		} else if !hasSearch {
			right += styleSubtle.Render("Press / to search | J to filter | ? for help | q to quit")
		}
	}

	// Center spacing
	spacing := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if spacing < 1 {
		spacing = 1
	}

	return left + strings.Repeat(" ", spacing) + right
}

// getFileListHeight calculates the height available for the file list
func (m Model) getFileListHeight() int {
	// Must match the actual pageSize calculation in renderSidebar
	// renderSidebar receives: m.height - 5 (from renderMain line 146)
	// pageSize = height - 4 = (m.height - 5) - 4 = m.height - 9
	return m.height - ContentOffsetSidebar
}

// updateViewport updates the response viewport
func (m *Model) updateViewport() {
	// Initialize or update viewport for response scrolling
	// MUST match width calculations in renderMain!
	var responseWidth int
	if m.fullscreen {
		// In fullscreen, use full width
		responseWidth = m.width - MinimalBorderMargin // Just account for borders
	} else {
		// In split view, account for sidebar
		sidebarWidth := max(40, m.width*40/100)
		if m.width < 100 {
			sidebarWidth = m.width / 2
		}
		responseWidth = m.width - sidebarWidth - ViewportPaddingHorizontal // Account for borders
	}

	// Viewport width = renderResponse width - content padding
	m.responseView.Width = responseWidth - ViewportPaddingHorizontal // -2 for renderResponse param, -2 for " " prefix padding in content
	// Viewport height calculation:
	// renderResponse gets: m.height - 5
	// Applies Height(height): m.height - 5
	// Content has: title (1) + blank (1) + viewport
	// So viewport = m.height - 5 - 2 = m.height - 7
	m.responseView.Height = m.height - ContentOffsetStandard

	// Initialize help viewport
	// Modal width: m.width - 10, Padding: 2 horizontal each side = 4, Border: included
	// So viewport width = (m.width - 10) - 4 (horizontal padding) = m.width - 14
	m.helpView.Width = m.width - HelpViewWidthOffset
	if m.helpView.Width < 10 {
		m.helpView.Width = 10
	}
	// Modal height: m.height - 4, Padding: 1 vertical each side = 2
	// Content: title (1) + blank (1) + viewport + blank (1) + footer (1) = viewport + 4
	// So viewport height = (m.height - 4) - 2 (padding) - 4 (non-viewport lines) = m.height - 10
	m.helpView.Height = m.height - ContentOffsetHelp
	if m.helpView.Height < 5 {
		m.helpView.Height = 5
	}
}

// updateResponseView updates the response viewport content
func (m *Model) updateResponseView() {
	var content strings.Builder

	// Show loading indicator FIRST (even if no response yet)
	if m.loading {
		loadingBar := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")). // Green
			Bold(true).
			Width(m.responseView.Width).
			Align(lipgloss.Center).
			Render(">>> EXECUTING REQUEST <<<")
		content.WriteString(loadingBar + "\n\n")
	}

	// Handle case where no response exists yet
	if m.currentResponse == nil {
		// If loading, show just the loading indicator
		if m.loading {
			m.responseContent = content.String()
			m.responseView.SetContent(content.String())
			return
		}
		// Otherwise show empty state
		m.responseContent = ""
		m.responseView.SetContent("")
		return
	}

	// Check if we can use cached content (performance optimization for large responses)
	cacheValid := m.cachedResponsePtr == m.currentResponse &&
		m.cachedViewWidth == m.responseView.Width &&
		m.cachedFilterActive == m.filterActive &&
		m.cachedSearchActive == m.searchInResponseCtx

	if cacheValid && !m.loading {
		// Use cached content
		contentStr := m.responseContent

		// Check if we need to re-highlight or can use cached highlighted version
		if m.searchInResponseCtx && len(m.responseSearchMatches) > 0 {
			// Only re-highlight if search matches have changed
			if len(m.responseSearchMatches) != m.cachedSearchMatchCount || m.cachedHighlightedBody == "" {
				contentStr = m.highlightSearchMatches(contentStr)
				m.cachedHighlightedBody = contentStr
				m.cachedSearchMatchCount = len(m.responseSearchMatches)
			} else {
				// Reuse cached highlighted content
				contentStr = m.cachedHighlightedBody
			}
		}

		m.responseView.SetContent(contentStr)
		return
	}

	// Request section with resolved values
	if m.currentRequest != nil {
		profile := m.sessionMgr.GetActiveProfile()
		session := m.sessionMgr.GetSession()

		// Create a copy of the request and merge headers
		requestCopy := *m.currentRequest
		requestCopy.Headers = make(map[string]string)
		for k, v := range profile.Headers {
			requestCopy.Headers[k] = v
		}
		for k, v := range m.currentRequest.Headers {
			requestCopy.Headers[k] = v
		}

		// Resolve variables for display (include interactive variables if collected)
		resolver := parser.NewVariableResolver(profile.Variables, session.Variables, m.interactiveVarValues, parser.LoadSystemEnv())
		resolvedRequest, err := resolver.ResolveRequest(&requestCopy)

		content.WriteString(styleTitle.Render("Request") + "\n")

		// Determine wrap width based on fullscreen mode
		wrapWidth := m.responseView.Width
		if wrapWidth < 40 {
			wrapWidth = 40
		}

		if err == nil && resolvedRequest != nil {
			// Show resolved values
			content.WriteString(fmt.Sprintf("%s %s\n", resolvedRequest.Method, resolvedRequest.URL))

			// Request Headers (with wrapping, toggle with Shift+B)
			if m.showHeaders && len(resolvedRequest.Headers) > 0 {
				content.WriteString("Request Headers:\n")
				for key, value := range resolvedRequest.Headers {
					// Wrap without indentation, then add it
					unwrappedLine := fmt.Sprintf("%s: %s", key, value)
					wrappedLines := wrapText(unwrappedLine, wrapWidth-2)
					for _, line := range strings.Split(wrappedLines, "\n") {
						if line != "" {
							content.WriteString("  " + line + "\n")
						}
					}
				}
			}
		} else {
			// Fallback to unresolved values if resolution fails
			content.WriteString(fmt.Sprintf("%s %s\n", requestCopy.Method, requestCopy.URL))

			// Request Headers (with wrapping, toggle with Shift+B)
			if m.showHeaders && len(requestCopy.Headers) > 0 {
				content.WriteString("Request Headers:\n")
				for key, value := range requestCopy.Headers {
					// Wrap without indentation, then add it
					unwrappedLine := fmt.Sprintf("%s: %s", key, value)
					wrappedLines := wrapText(unwrappedLine, wrapWidth-2)
					for _, line := range strings.Split(wrappedLines, "\n") {
						if line != "" {
							content.WriteString("  " + line + "\n")
						}
					}
				}
			}
		}
		content.WriteString("\n")
	}

	// Response section
	content.WriteString(styleTitle.Render("Response") + "\n")

	// Status line
	statusStyle := styleSuccess
	if m.currentResponse.Status >= 400 {
		statusStyle = styleError
	} else if m.currentResponse.Status >= 300 {
		statusStyle = styleWarning
	}

	content.WriteString(fmt.Sprintf("%s - %s\n",
		statusStyle.Render(fmt.Sprintf("%d", m.currentResponse.Status)),
		m.currentResponse.StatusText))

	// Timing info
	timingParts := []string{
		fmt.Sprintf("Duration: %s", executor.FormatDuration(m.currentResponse.Duration)),
		fmt.Sprintf("Size: %s", executor.FormatSize(m.currentResponse.ResponseSize)),
	}
	if m.currentResponse.Timestamp != "" {
		timingParts = append(timingParts, fmt.Sprintf("Time: %s", m.currentResponse.Timestamp))
	}
	content.WriteString(styleSubtle.Render(strings.Join(timingParts, " | ")))
	content.WriteString("\n")

	// Response Headers (toggle with Shift+B, with wrapping)
	if m.showHeaders && len(m.currentResponse.Headers) > 0 {
		content.WriteString("Response Headers:\n")
		wrapWidth := m.responseView.Width
		if wrapWidth < 40 {
			wrapWidth = 40
		}
		for key, value := range m.currentResponse.Headers {
			// Wrap without indentation, then add it
			unwrappedLine := fmt.Sprintf("%s: %s", key, value)
			wrappedLines := wrapText(unwrappedLine, wrapWidth-2) // Reserve 2 chars for indent
			// Add indent to each line
			for _, line := range strings.Split(wrappedLines, "\n") {
				if line != "" {
					content.WriteString("  " + line + "\n")
				}
			}
		}
	}
	content.WriteString("\n")

	// Body
	if m.currentResponse.Body != "" {
		// Show filter indicator if active
		if m.filterActive && m.filteredResponse != "" {
			content.WriteString(styleTitle.Render(fmt.Sprintf("Body (Filtered: %s)", m.filterInput)) + "\n")
		} else {
			content.WriteString(styleTitle.Render("Body") + "\n")
		}

		// Use filtered response if active, otherwise use original
		var bodySource string
		if m.filterActive && m.filteredResponse != "" {
			bodySource = m.filteredResponse
		} else {
			bodySource = m.currentResponse.Body
		}

		// Check if content is binary
		if isBinaryContent(bodySource) {
			// Show binary content indicator instead of garbage
			content.WriteString(styleSubtle.Render(fmt.Sprintf(
				"[Binary content - %s - %d bytes]\n\nResponse contains binary data that cannot be displayed as text.\n"+
					"Content-Type: %s",
				executor.FormatSize(len(bodySource)),
				len(bodySource),
				m.currentResponse.Headers["Content-Type"],
			)))
			m.responseContent = content.String()
			m.responseView.SetContent(content.String())
			return
		}

		// Try to pretty-print and highlight JSON
		var bodyText string
		var isJSON bool
		var jsonData interface{}
		if err := json.Unmarshal([]byte(bodySource), &jsonData); err == nil {
			if prettyJSON, err := json.MarshalIndent(jsonData, "", "  "); err == nil {
				bodyText = string(prettyJSON)
				isJSON = true
			} else {
				// Fallback to raw body if formatting fails
				bodyText = bodySource
			}
		} else {
			// Not JSON, show raw body
			bodyText = bodySource
		}

		// Wrap text to viewport width to prevent truncation
		wrapWidth := m.responseView.Width
		if wrapWidth < 40 {
			wrapWidth = 40 // Minimum reasonable width
		}
		wrappedBody := wrapText(bodyText, wrapWidth)

		// Apply syntax highlighting after wrapping (for JSON only)
		if isJSON {
			profile := m.sessionMgr.GetActiveProfile()
			wrappedBody = highlightJSON(wrappedBody, profile)
		}

		content.WriteString(wrappedBody)
		content.WriteString("\n")

		// Show hint to clear filter
		if m.filterActive {
			content.WriteString("\n")
			content.WriteString(styleSubtle.Render("Press J to clear filter"))
		}
	}

	// Error
	if m.currentResponse.Error != "" {
		content.WriteString("\n")
		// Wrap error text to viewport width to prevent cropping
		wrapWidth := m.responseView.Width
		if wrapWidth < 40 {
			wrapWidth = 40
		}
		errorText := categorizeRequestError(m.currentResponse.Error)
		wrappedError := wrapText(errorText, wrapWidth)
		content.WriteString(styleError.Render(wrappedError))
	}

	contentStr := content.String()
	m.responseContent = contentStr

	// Update cache tracking
	m.cachedResponsePtr = m.currentResponse
	m.cachedViewWidth = m.responseView.Width
	m.cachedFilterActive = m.filterActive
	m.cachedSearchActive = m.searchInResponseCtx

	// Apply search highlighting if we're searching in response
	if m.searchInResponseCtx && len(m.responseSearchMatches) > 0 {
		contentStr = m.highlightSearchMatches(contentStr)
		m.cachedHighlightedBody = contentStr
		m.cachedSearchMatchCount = len(m.responseSearchMatches)
	} else {
		// Clear cached highlighting when not searching
		m.cachedHighlightedBody = ""
		m.cachedSearchMatchCount = 0
	}

	m.responseView.SetContent(contentStr)
}

// highlightSearchMatches highlights lines that match the current search
func (m *Model) highlightSearchMatches(content string) string {
	lines := strings.Split(content, "\n")

	// Create a map of line numbers to highlight for faster lookup
	matchLines := make(map[int]bool)
	for _, lineNum := range m.responseSearchMatches {
		matchLines[lineNum] = true
	}

	// Highlight matching lines with background color
	for i := range lines {
		if matchLines[i] {
			lines[i] = styleSearchHighlight.Render(lines[i])
		}
	}

	return strings.Join(lines, "\n")
}

// updateHelpView updates the help viewport content
func (m *Model) updateHelpView() {
	versionLine := fmt.Sprintf("REST CLI v%s", m.version)
	if m.updateAvailable {
		versionLine += fmt.Sprintf(" (Update available: v%s - %s)", m.latestVersion, m.updateURL)
	}
	helpText := versionLine + ` - Keyboard Shortcuts

NAVIGATION (Keyboard Only)
  TAB            Switch focus (sidebar ↔ response)
  ↑/↓, j/k       Navigate files OR scroll response
  gg             Go to first item
  G              Go to last item
  Ctrl+U/D       Half-page up/down
  PageUp/Down    Full page scroll (focused panel)
  Home/End       Jump to first/last
  :              Goto hex line
  Ctrl+P         Recent files (MRU)

SEARCH
  /              Search files or response (context-aware)
                 Supports regex patterns (e.g., .*foo.*bar)
  n              Next search result
  N              Previous search result
  Ctrl+R         Next search result
  ESC            Clear search / Cancel

FOCUS
  Green border   Shows which panel is focused
  Arrow keys     Control the focused panel only

FILE OPERATIONS
  Enter        Execute request
  ESC          Cancel running request
  i            Inspect request
  x            Open in editor
  X            Configure editor
  d            Duplicate file
  D            Delete file (with confirmation)
  F            Create new file
  R            Rename file
  r            Refresh file list
  t            Filter by category
  T            Clear category filter

RESPONSE
  s            Save response to file
  c            Copy full response to clipboard
  b            Toggle body visibility
  B            Toggle headers visibility (request + response)
  E            Edit request body (one-time override)
  f            Toggle fullscreen (ESC to exit)
  w            Pin response for comparison
  W            Show diff (compare pinned vs current)
  J            Filter response with JMESPath (toggle on/off)
  ↑/↓, j/k     Scroll response (when body shown)

INLINE FILTER EDITOR (when J pressed)
  Type         Enter JMESPath expression
  Enter        Apply filter
  Esc          Cancel filter editing
  Ctrl+S       Save expression as bookmark
  Up           Open bookmark/history selector

CONFIGURATION
  v            Variable editor
  h            Header editor
  p            Switch profile
  n            Create new profile (when no search active)
  C            View current configuration
  P            Edit .profiles.json
  Ctrl+X       View session config

PROFILE SWITCHER (when in profile modal)
  Enter        Switch to selected profile
  e            Edit selected profile
  d            Duplicate selected profile
  D            Delete selected profile
  n            Create new profile

TOOLS
  M            Mock server manager
  y            Debug proxy viewer
  A            Analytics viewer
  S            Stress test results

MOCK SERVER (when in modal)
  s            Start/stop server
  c            Clear logs
  j/k          Scroll logs
  g/G          Top/bottom
  Esc, q       Close modal

DEBUG PROXY (when in viewer)
  s            Start/stop proxy
  c            Clear logs
  j/k          Navigate requests
  Ctrl+D/U     Half page down/up
  Enter        View details
  Esc, q       Close viewer

ANALYTICS (when in viewer)
  Tab          Switch focus (list/details)
  j/k          Navigate/scroll (focused pane)
  Enter        Load request file
  p            Toggle preview pane
  t            Toggle grouping
  C            Clear all analytics
  Esc, q       Close viewer

STRESS TESTING (when in results)
  Tab          Switch focus (list/details)
  j/k          Navigate/scroll (focused pane)
  n            New stress test
  r            Re-run test
  d            Delete run
  l            Load saved config
  Ctrl+S       Save & start (in config)
  Ctrl+L       Load config (in config)
  Esc, q       Close viewer

VARIABLE EDITOR (multi-value)
  m            Manage options for multi-value variable
  s            Set active option
  a            Add option
  e            Edit option
  d            Delete option
  l            Set alias for option (e.g., 'u1', 'dev')
  L            Delete aliases from option

DOCUMENTATION, HISTORY & ANALYTICS
  m            View documentation
  H            View history
  A            View analytics

DOCUMENTATION VIEWER (when in docs modal)
  j/k          Navigate tree
  Enter        Expand/collapse section
  Space        Toggle section
  Esc          Close viewer

HISTORY VIEWER (when in history modal)
  j/k          Navigate history
  Enter        Load selected response
  r            Replay selected request
  p            Toggle preview pane
  C            Clear all history (with confirmation)
  Esc          Close viewer

ANALYTICS VIEWER (when in analytics modal)
  p            Toggle preview pane
  t            Toggle grouping (per-file ↔ by path)
  C            Clear all analytics (with confirmation)

STRESS TESTING
  S            Open stress test results modal
  Enter        Run stress test (when focused on config/run)
  e            Edit configuration field
  ↑/↓, j/k     Navigate between fields or runs
  TAB          Switch focus (config ↔ results)
  ESC          Stop running test or exit modal
  C            Clear all stress test runs (with confirmation)

MODAL NAVIGATION
  ↑/↓, j/k     Navigate items
  gg           Go to first item
  G            Go to last item
  Ctrl+U/D     Half-page up/down
  PageUp/Down  Full page up/down
  Home/End     Jump to first/last

OAUTH
  o            Start OAuth flow
  O            Configure OAuth

WEBSOCKET
  j/k, ↑/↓     Navigate messages or scroll history
  gg           Go to top
  G            Go to bottom
  Ctrl+D       Page down
  Ctrl+U       Page up
  Enter        Send selected message
  r            Connect/reconnect
  d            Disconnect
  i            Compose custom message
  c            Copy last message
  C            Clear history
  e            Export to JSON
  /            Search messages
  Tab          Switch panes
  q, Esc       Close modal

DIFF VIEWER (when comparing responses)
  j/k          Scroll diff
  Esc          Close diff

MODAL OPERATIONS
  Esc          Close modal
  Enter        Confirm/Submit
  Tab          Next field
  Shift+Tab    Previous field

TEXT INPUT (in modals)
  Ctrl+V       Paste from clipboard (recommended)
  Cmd+V        Paste from clipboard (macOS, may not work in Terminal.app)
  Shift+Insert Paste (alternative)
  Ctrl+Y       Paste (alternative)
  Ctrl+K       Clear input
  Backspace    Delete character
  Delete       Delete character forward

HELP AND INFO
  ?            Show this help
  e            View full error details (when error in footer)
  I            Show full status message
  q            Quit application

CLI MODE
  restcli <file>           Execute without profile (prompts for vars)
  restcli <file> -p <name> Execute with profile (no prompts)
  restcli <file> -e k=v    Provide variable (won't be prompted)
  --env-file <path>        Load environment variables from file`

	// Apply search filter if active
	if m.helpSearchQuery != "" {
		lines := strings.Split(helpText, "\n")
		var filteredLines []string
		query := strings.ToLower(m.helpSearchQuery)

		for _, line := range lines {
			if strings.Contains(strings.ToLower(line), query) {
				// Highlight matching text
				filteredLines = append(filteredLines, line)
			}
		}

		if len(filteredLines) == 0 {
			m.helpView.SetContent(fmt.Sprintf("No matches found for: %s\n\nPress ESC to clear search", m.helpSearchQuery))
		} else {
			m.helpView.SetContent(strings.Join(filteredLines, "\n"))
		}
	} else {
		m.helpView.SetContent(helpText)
	}
}

// renderConfigView renders the current configuration view modal
func (m Model) renderConfigView() string {
	var content strings.Builder

	profile := m.sessionMgr.GetActiveProfile()

	// Helper to wrap long values with indentation for continuation lines
	wrapValue := func(label, value string, maxWidth int) string {
		labelLen := len(label)
		valueWidth := maxWidth - labelLen - 2 // 2 for initial indent
		if valueWidth < 20 {
			valueWidth = 20
		}

		if len(value) <= valueWidth {
			return fmt.Sprintf("  %s%s\n", label, value)
		}

		// Wrap long values
		var result strings.Builder
		result.WriteString(fmt.Sprintf("  %s", label))
		indent := strings.Repeat(" ", labelLen+2)

		remaining := value
		first := true
		for len(remaining) > 0 {
			chunkLen := valueWidth
			if len(remaining) < chunkLen {
				chunkLen = len(remaining)
			}

			if first {
				result.WriteString(remaining[:chunkLen])
				first = false
			} else {
				result.WriteString("\n")
				result.WriteString(indent)
				result.WriteString(remaining[:chunkLen])
			}
			remaining = remaining[chunkLen:]
		}
		result.WriteString("\n")
		return result.String()
	}

	modalWidth := 70

	// Profile info
	content.WriteString(styleTitle.Render("ACTIVE PROFILE"))
	content.WriteString("\n\n")
	content.WriteString(wrapValue("Name:     ", profile.Name, modalWidth-4))

	// Working directory
	workdir, err := config.GetWorkingDirectory(profile.Workdir)
	if err != nil {
		workdir = profile.Workdir + " (error)"
	}
	content.WriteString(wrapValue("Workdir:  ", workdir, modalWidth-4))

	// Editor
	editor := profile.Editor
	if editor == "" {
		editor = "(default)"
	}
	content.WriteString(wrapValue("Editor:   ", editor, modalWidth-4))

	// Output format
	output := profile.Output
	if output == "" {
		output = "json"
	}
	content.WriteString(wrapValue("Output:   ", output, modalWidth-4))

	// Session info
	content.WriteString("\n")
	content.WriteString(styleTitle.Render("SESSION"))
	content.WriteString("\n\n")

	// History status
	historyStatus := "enabled"
	if m.sessionMgr.IsHistoryEnabled() == false {
		historyStatus = "disabled"
	}
	content.WriteString(wrapValue("History:  ", historyStatus, modalWidth-4))

	// Variable count
	session := m.sessionMgr.GetSession()
	content.WriteString(wrapValue("Variables:", fmt.Sprintf(" %d", len(session.Variables)), modalWidth-4))

	// Header count
	headers := profile.Headers
	content.WriteString(wrapValue("Headers:  ", fmt.Sprintf("%d", len(headers)), modalWidth-4))

	// OAuth status
	oauthStatus := "not configured"
	if profile.OAuth != nil && profile.OAuth.Enabled {
		if profile.OAuth.AuthURL != "" {
			oauthStatus = "configured (auto)"
		} else if profile.OAuth.AuthEndpoint != "" {
			oauthStatus = "configured (manual)"
		} else {
			oauthStatus = "enabled"
		}
	}
	content.WriteString(wrapValue("OAuth:    ", oauthStatus, modalWidth-4))

	// Config paths
	content.WriteString("\n")
	content.WriteString(styleTitle.Render("CONFIG PATHS"))
	content.WriteString("\n\n")
	content.WriteString(wrapValue("Config:   ", config.ConfigDir, modalWidth-4))
	content.WriteString(wrapValue("Session:  ", config.GetSessionFilePath(), modalWidth-4))
	content.WriteString(wrapValue("Profiles: ", config.GetProfilesFilePath(), modalWidth-4))

	footer := "[ESC/C/q] close"

	return m.renderModalWithFooter("Current Configuration", content.String(), footer, 70, 25)
}

// updateDocumentationView updates the documentation viewport content
func (m *Model) updateDocumentationView() {
	// Set viewport dimensions for the modal (nearly full screen: m.width-6, m.height-3)
	// Account for: border (2), padding (2), title+blank (2), footer+blank (2) = 8 total
	m.modalView.Width = m.width - ModalWidthMarginNarrow  // Modal content width minus padding
	m.modalView.Height = m.height - ContentOffsetSidebar // Modal content height with footer

	// Check if we have a request and documentation
	if m.currentRequest == nil {
		m.modalView.SetContent("No request selected\n\nPress ESC to close")
		return
	}

	// Parse documentation on demand (lazy loading)
	if m.currentRequest.Documentation == nil && len(m.currentRequest.DocumentationLines) > 0 {
		m.currentRequest.EnsureDocumentationParsed(parser.ParseDocumentationLines)
		// CRITICAL: Initialize collapse state AFTER parsing documentation
		m.initializeCollapsedFields()
	}

	if m.currentRequest.Documentation == nil {
		m.modalView.SetContent("No documentation available\n\nPress ESC to close")
		return
	}

	var content strings.Builder
	doc := m.currentRequest.Documentation
	currentIdx := 0
	selectedLineNum := -1 // Track which line number the selected item is on

	// Helper to render a collapsible section
	renderSection := func(title string, index int, renderContent func()) {
		collapsed := m.docState.GetCollapsed(index)
		marker := "▼" // Expanded
		if collapsed {
			marker = "▶" // Collapsed
		}

		// Highlight if selected
		line := fmt.Sprintf("%s %s", marker, title)
		if currentIdx == m.docState.GetSelectedIdx() {
			line = styleSelected.Render(line)
			selectedLineNum = strings.Count(content.String(), "\n")
		} else {
			line = styleTitle.Render(line)
		}
		content.WriteString(line + "\n")
		currentIdx++

		if !collapsed {
			renderContent()
		}
		content.WriteString("\n")
	}

	// Description (always visible, not collapsible)
	if doc.Description != "" {
		content.WriteString(doc.Description + "\n\n")
	}

	// Tags (always visible, not collapsible)
	if len(doc.Tags) > 0 {
		content.WriteString(styleSubtle.Render("Tags: "+strings.Join(doc.Tags, ", ")) + "\n\n")
	}

	// Parameters section (collapsible)
	if len(doc.Parameters) > 0 {
		renderSection(fmt.Sprintf("Parameters (%d)", len(doc.Parameters)), 0, func() {
			for _, param := range doc.Parameters {
				required := ""
				if param.Required {
					required = styleError.Render(" [required]")
				} else {
					required = styleWarning.Render(" [optional]")
				}

				line := fmt.Sprintf("  %s", param.Name)
				if param.Type != "" {
					line += styleSubtle.Render(fmt.Sprintf(" {%s}", param.Type))
				}
				line += required

				// Highlight if this parameter is selected
				if currentIdx == m.docState.GetSelectedIdx() {
					line = styleSelected.Render(line)
					selectedLineNum = strings.Count(content.String(), "\n")
				}
				content.WriteString(line + "\n")
				currentIdx++

				// Description and example always shown for parameters
				if param.Description != "" {
					content.WriteString(styleSubtle.Render(fmt.Sprintf("      %s\n", param.Description)))
				}
				if param.Example != "" {
					content.WriteString(fmt.Sprintf("      Example: %s\n", styleSuccess.Render(param.Example)))
				}
			}
		})
	}

	// Responses section (collapsible) - with tree rendering for fields
	if len(doc.Responses) > 0 {
		renderSection(fmt.Sprintf("Responses (%d)", len(doc.Responses)), 1, func() {
			for respIdx, resp := range doc.Responses {
				// Response code and description
				line := fmt.Sprintf("  %s: %s", styleSuccess.Render(resp.Code), resp.Description)
				if currentIdx == m.docState.GetSelectedIdx() {
					line = styleSelected.Render(line)
					selectedLineNum = strings.Count(content.String(), "\n")
				}
				content.WriteString(line + "\n")
				currentIdx++

				// Response fields - lazy rendering (only build tree when response is expanded)
				if len(resp.Fields) > 0 {
					responseKey := 100 + respIdx
					responseFieldsCollapsed := m.docState.GetCollapsed(responseKey)

					if !responseFieldsCollapsed {
						// Response fields are expanded - use cached tree
						allFields := m.docState.GetFieldTreeCache(respIdx); ok := allFields != nil
						if !ok {
							// Build and cache the tree
							allFields = buildVirtualFieldTree(resp.Fields)
							m.docState.SetFieldTreeCache(respIdx, allFields)
							m.docState.SetChildrenCache(respIdx, buildHasChildrenCache(allFields))
						}
						hasChildrenCache := m.docState.GetChildrenCache(respIdx)
						m.renderResponseFieldsTree(respIdx, "", allFields, hasChildrenCache, &currentIdx, &content, 0, &selectedLineNum)
					} else {
						// Response fields are collapsed - show indicator only (no tree building!)
						line := fmt.Sprintf("    ▶ %d fields", len(resp.Fields))
						if currentIdx == m.docState.GetSelectedIdx() {
							line = styleSelected.Render(line)
							selectedLineNum = strings.Count(content.String(), "\n")
						}
						content.WriteString(line + "\n")
						currentIdx++
					}
				}
			}
		})
	}

	// Save scroll position before updating content
	savedOffset := m.modalView.YOffset
	m.modalView.SetContent(content.String())

	// Auto-scroll to keep selected item visible
	if selectedLineNum >= 0 && m.modalView.Height > 0 {
		topVisible := savedOffset
		bottomVisible := savedOffset + m.modalView.Height - 1

		if selectedLineNum < topVisible {
			// Selected is above viewport - scroll up
			m.modalView.SetYOffset(selectedLineNum)
		} else if selectedLineNum > bottomVisible {
			// Selected is below viewport - scroll down just enough
			m.modalView.SetYOffset(selectedLineNum - m.modalView.Height + 1)
		} else {
			// Selected is visible - keep current scroll
			m.modalView.SetYOffset(savedOffset)
		}
	} else {
		// No selection or invalid - keep scroll position
		m.modalView.SetYOffset(savedOffset)
	}
}

// updateInspectView updates the inspect viewport content
func (m *Model) updateInspectView() {
	// Set viewport dimensions for the modal (nearly full screen: m.width-6, m.height-3)
	m.modalView.Width = m.width - ModalWidthMarginNarrow  // Modal content width minus padding
	m.modalView.Height = m.height - ContentOffsetSidebar // Modal content height minus padding, title lines, and footer

	if m.currentRequest == nil {
		m.modalView.SetContent("No request selected\n\nPress ESC to close")
		return
	}

	// Ensure documentation is parsed before accessing categories
	m.currentRequest.EnsureDocumentationParsed(parser.ParseDocumentationLines)

	profile := m.sessionMgr.GetActiveProfile()

	// Create a copy of the request to avoid mutation
	requestCopy := *m.currentRequest
	requestCopy.Headers = make(map[string]string)

	// Merge headers into the copy
	for k, v := range profile.Headers {
		requestCopy.Headers[k] = v
	}
	for k, v := range m.currentRequest.Headers {
		requestCopy.Headers[k] = v
	}

	// Resolve variables for preview
	resolver := parser.NewVariableResolver(profile.Variables, m.sessionMgr.GetSession().Variables, nil, parser.LoadSystemEnv())
	resolvedRequest, err := resolver.ResolveRequest(&requestCopy)

	var content strings.Builder
	content.WriteString("Request Preview\n\n")

	if err != nil {
		content.WriteString(styleError.Render(fmt.Sprintf("Error resolving variables: %v\n\n", err)))
	}

	// Show request details
	if resolvedRequest != nil {
		wrapWidth := m.modalView.Width
		if wrapWidth < 40 {
			wrapWidth = 40
		}

		// URL with wrapping
		methodLine := fmt.Sprintf("%s %s", styleTitle.Render(resolvedRequest.Method), resolvedRequest.URL)
		wrappedMethod := wrapText(methodLine, wrapWidth)
		content.WriteString(wrappedMethod + "\n\n")

		if len(resolvedRequest.Headers) > 0 {
			content.WriteString("Headers:\n")
			// Get sorted header names for consistent display
			headerNames := make([]string, 0, len(resolvedRequest.Headers))
			for key := range resolvedRequest.Headers {
				headerNames = append(headerNames, key)
			}
			// Simple sort
			for i := 0; i < len(headerNames); i++ {
				for j := i + 1; j < len(headerNames); j++ {
					if headerNames[i] > headerNames[j] {
						headerNames[i], headerNames[j] = headerNames[j], headerNames[i]
					}
				}
			}

			// Headers with wrapping
			for _, key := range headerNames {
				value := resolvedRequest.Headers[key]
				headerLine := fmt.Sprintf("%s: %s", key, value)
				wrappedHeader := wrapText(headerLine, wrapWidth-2)
				// Add indent to each wrapped line
				for _, line := range strings.Split(wrappedHeader, "\n") {
					if line != "" {
						content.WriteString("  " + line + "\n")
					}
				}
			}
			content.WriteString("\n")
		}

		if resolvedRequest.Body != "" {
			content.WriteString("Body:\n")
			// Wrap body lines
			bodyLines := strings.Split(resolvedRequest.Body, "\n")
			for _, line := range bodyLines {
				wrappedLine := wrapText(line, wrapWidth-2)
				for _, wl := range strings.Split(wrappedLine, "\n") {
					if wl != "" {
						content.WriteString("  " + wl + "\n")
					}
				}
			}
			content.WriteString("\n")
		}

		// Show filter if present
		if resolvedRequest.Filter != "" {
			content.WriteString("Filter:\n")
			wrappedFilter := wrapText(resolvedRequest.Filter, wrapWidth-2)
			for _, line := range strings.Split(wrappedFilter, "\n") {
				if line != "" {
					content.WriteString("  " + line + "\n")
				}
			}
			content.WriteString("\n")
		}

		// Show query if present
		if resolvedRequest.Query != "" {
			content.WriteString("Query:\n")
			wrappedQuery := wrapText(resolvedRequest.Query, wrapWidth-2)
			for _, line := range strings.Split(wrappedQuery, "\n") {
				if line != "" {
					content.WriteString("  " + line + "\n")
				}
			}
			content.WriteString("\n")
		}

		// Show TLS configuration if present
		if resolvedRequest.TLS != nil {
			content.WriteString("TLS Configuration:\n")
			if resolvedRequest.TLS.CertFile != "" {
				content.WriteString("  Cert: " + resolvedRequest.TLS.CertFile + "\n")
			}
			if resolvedRequest.TLS.KeyFile != "" {
				content.WriteString("  Key: " + resolvedRequest.TLS.KeyFile + "\n")
			}
			if resolvedRequest.TLS.CAFile != "" {
				content.WriteString("  CA: " + resolvedRequest.TLS.CAFile + "\n")
			}
			if resolvedRequest.TLS.InsecureSkipVerify {
				content.WriteString("  " + styleWarning.Render("Insecure Skip Verify: true") + "\n")
			}
			content.WriteString("\n")
		}

		// Show categories if present
		if resolvedRequest.Documentation != nil && len(resolvedRequest.Documentation.Tags) > 0 {
			content.WriteString("Categories:\n")
			content.WriteString("  " + strings.Join(resolvedRequest.Documentation.Tags, ", ") + "\n\n")
		}

		// Show chaining configuration if present
		if len(resolvedRequest.DependsOn) > 0 || len(resolvedRequest.Extract) > 0 {
			content.WriteString("Request Chaining:\n")

			// Show execution order if there are dependencies
			if len(resolvedRequest.DependsOn) > 0 {
				// Build dependency graph to show execution order
				profile := m.sessionMgr.GetActiveProfile()
				graph := chain.NewGraph(profile.Workdir)

				// Get current file path
				currentFileInfo := m.fileExplorer.GetCurrentFile()
				currentFile := ""
				if currentFileInfo != nil {
					currentFile = currentFileInfo.Path
				}

				if currentFile != "" {
					if err := graph.BuildGraph(currentFile); err == nil {
						if execOrder, err := graph.GetExecutionOrder(currentFile); err == nil {
							content.WriteString("  Execution order:\n")
							for i, filePath := range execOrder {
								baseName := filepath.Base(filePath)

								// Check if this file has extractions
								if requests, err := parser.Parse(filePath); err == nil && len(requests) > 0 {
									req := &requests[0]
									if len(req.Extract) > 0 {
										extractVars := make([]string, 0, len(req.Extract))
										for varName := range req.Extract {
											extractVars = append(extractVars, varName)
										}
										content.WriteString(fmt.Sprintf("    %d. %s → extracts: %s\n", i+1, baseName, strings.Join(extractVars, ", ")))
									} else {
										content.WriteString(fmt.Sprintf("    %d. %s\n", i+1, baseName))
									}
								} else {
									content.WriteString(fmt.Sprintf("    %d. %s\n", i+1, baseName))
								}
							}
						}
					}
				}

				content.WriteString("\n  Dependencies:\n")
				for _, dep := range resolvedRequest.DependsOn {
					content.WriteString("    - " + dep + "\n")
				}
			}

			// Show extractions for current file
			if len(resolvedRequest.Extract) > 0 {
				content.WriteString("  This file extracts:\n")
				for varName, jmesPath := range resolvedRequest.Extract {
					content.WriteString(fmt.Sprintf("    %s = %s\n", varName, jmesPath))
				}
			}

			content.WriteString("\n")
		}

		// Show validation configuration if present (for stress testing)
		hasValidation := len(resolvedRequest.ExpectedStatusCodes) > 0 ||
			resolvedRequest.ExpectedBodyExact != "" ||
			resolvedRequest.ExpectedBodyContains != "" ||
			resolvedRequest.ExpectedBodyPattern != "" ||
			len(resolvedRequest.ExpectedBodyFields) > 0

		if hasValidation {
			content.WriteString("Validation (Stress Testing):\n")

			// Expected status codes
			if len(resolvedRequest.ExpectedStatusCodes) > 0 {
				codes := ""
				for i, code := range resolvedRequest.ExpectedStatusCodes {
					if i > 0 {
						codes += ", "
					}
					codes += fmt.Sprintf("%d", code)
					// Limit display to first 10 codes
					if i >= 9 && len(resolvedRequest.ExpectedStatusCodes) > 10 {
						codes += fmt.Sprintf(", ... (%d more)", len(resolvedRequest.ExpectedStatusCodes)-10)
						break
					}
				}
				content.WriteString("  Expected Status: " + codes + "\n")
			}

			// Expected body exact match
			if resolvedRequest.ExpectedBodyExact != "" {
				truncated := resolvedRequest.ExpectedBodyExact
				if len(truncated) > 60 {
					truncated = truncated[:57] + "..."
				}
				content.WriteString("  Body Exact: " + truncated + "\n")
			}

			// Expected body substring
			if resolvedRequest.ExpectedBodyContains != "" {
				truncated := resolvedRequest.ExpectedBodyContains
				if len(truncated) > 60 {
					truncated = truncated[:57] + "..."
				}
				content.WriteString("  Body Contains: " + truncated + "\n")
			}

			// Expected body pattern
			if resolvedRequest.ExpectedBodyPattern != "" {
				truncated := resolvedRequest.ExpectedBodyPattern
				if len(truncated) > 60 {
					truncated = truncated[:57] + "..."
				}
				content.WriteString("  Body Pattern: " + truncated + "\n")
			}

			// Expected body fields
			if len(resolvedRequest.ExpectedBodyFields) > 0 {
				content.WriteString("  Body Fields:\n")
				count := 0
				for field, value := range resolvedRequest.ExpectedBodyFields {
					truncatedValue := value
					if len(truncatedValue) > 50 {
						truncatedValue = truncatedValue[:47] + "..."
					}
					content.WriteString(fmt.Sprintf("    %s = %s\n", field, truncatedValue))
					count++
					// Limit display to first 5 fields
					if count >= 5 && len(resolvedRequest.ExpectedBodyFields) > 5 {
						content.WriteString(fmt.Sprintf("    ... (%d more fields)\n", len(resolvedRequest.ExpectedBodyFields)-5))
						break
					}
				}
			}

			content.WriteString("\n")
		}
	}

	// Show parsing option if enabled (from original request, not resolved)
	if m.currentRequest.ParseEscapes {
		content.WriteString("Response Parsing:\n")
		content.WriteString("  " + styleSuccess.Render("Enabled") + " - Escape sequences (\\n, \\t, etc.) will be parsed\n")
		content.WriteString("\n")
	}

	m.modalView.SetContent(content.String())
	m.modalView.GotoTop()
}

// updateHistoryView updates the history viewport content for split view (Telescope-style)
func (m *Model) updateHistoryView() {
	// Calculate dimensions
	modalWidth := m.width - ModalWidthMargin
	modalHeight := m.height - ModalHeightMargin
	paneHeight := modalHeight - ViewportPaddingHorizontal // Account for borders and padding

	// Adjust viewport widths based on preview visibility
	if m.historyState.GetPreviewVisible() {
		// Split view mode: calculate widths for both panes
		listWidth := (modalWidth - SplitPaneBorderWidth) / 2          // Left pane width
		previewWidth := modalWidth - listWidth - SplitPaneBorderWidth // Right pane width

		// Set viewport dimensions for left pane (history list)
		m.modalView.Width = listWidth - ViewportPaddingHorizontal   // Account for padding and borders
		m.modalView.Height = paneHeight - ViewportPaddingVertical // Account for title

		// Set viewport dimensions for right pane (response preview)
		previewView := m.historyState.GetPreviewView()
		previewView.Width = previewWidth - ViewportPaddingHorizontal
		previewView.Height = paneHeight - ViewportPaddingVertical
		m.historyState.SetPreviewView(previewView)
	} else {
		// Preview hidden: expand list to full width
		m.modalView.Width = modalWidth - ViewportPaddingHorizontal  // Account for padding and borders
		m.modalView.Height = paneHeight - ViewportPaddingVertical // Account for title
	}

	// Build content for left pane (history list)
	var listContent strings.Builder
	if len(m.historyState.GetEntries()) == 0 {
		listContent.WriteString("No history entries")
	} else {
		// Show ALL entries (not just first 10) - viewport handles scrolling
		for i, entry := range m.historyState.GetEntries() {
			statusStyle := styleSuccess
			if entry.ResponseStatus >= 400 {
				statusStyle = styleError
			}

			line := fmt.Sprintf("%s %s %s - %s",
				entry.Timestamp[:19], // Truncate timestamp
				entry.Method,
				entry.URL,
				statusStyle.Render(fmt.Sprintf("%d", entry.ResponseStatus)))

			// Highlight selected entry
			if i == m.historyState.GetIndex() {
				line = styleSelected.Render(line)
			}

			listContent.WriteString(line + "\n")
		}
	}

	// Build content for right pane (response preview) - ONLY if preview is visible
	if m.historyState.GetPreviewVisible() {
		var previewContent strings.Builder
		if len(m.historyState.GetEntries()) > 0 && m.historyState.GetIndex() >= 0 && m.historyState.GetIndex() < len(m.historyState.GetEntries()) {
			entry := m.historyState.GetEntries()[m.historyState.GetIndex()]

			// Show response metadata
			previewContent.WriteString(fmt.Sprintf("%s %s\n", entry.Method, entry.URL))
			previewContent.WriteString(fmt.Sprintf("Status: %d %s\n", entry.ResponseStatus, entry.ResponseStatusText))
			previewContent.WriteString(fmt.Sprintf("Size: %d bytes\n", entry.ResponseSize))
			previewContent.WriteString(fmt.Sprintf("Time: %s\n\n", entry.Timestamp[:19]))

			// Show response body with JSON formatting and wrapping
			bodyText := entry.ResponseBody
			isJSON := false

			// Try to pretty-print JSON
			var jsonData interface{}
			if err := json.Unmarshal([]byte(entry.ResponseBody), &jsonData); err == nil {
				if prettyJSON, err := json.MarshalIndent(jsonData, "", "  "); err == nil {
					bodyText = string(prettyJSON)
					isJSON = true
				}
			}

			// Wrap text to viewport width
			wrapWidth := m.historyState.GetPreviewView().Width
			if wrapWidth < 40 {
				wrapWidth = 40
			}
			wrappedBody := wrapText(bodyText, wrapWidth)

			// Apply syntax highlighting for JSON
			if isJSON {
				profile := m.sessionMgr.GetActiveProfile()
				wrappedBody = highlightJSON(wrappedBody, profile)
			}

			previewContent.WriteString(wrappedBody)
		} else {
			previewContent.WriteString("No history entry selected")
		}

		// Set preview content (always start from top when selection changes)
		previewView := m.historyState.GetPreviewView()
		previewView.SetContent(previewContent.String())
		previewView.GotoTop()
		m.historyState.SetPreviewView(previewView)
	} else {
		// Clear preview content when hidden (security/privacy)
		previewView := m.historyState.GetPreviewView()
		previewView.SetContent("")
		m.historyState.SetPreviewView(previewView)
	}

	// Save current scroll positions before updating content
	listYOffset := m.modalView.YOffset
	m.modalView.SetContent(listContent.String())

	// Auto-scroll list to keep selected item visible
	if len(m.historyState.GetEntries()) > 0 && m.historyState.GetIndex() >= 0 {
		selectedLine := m.historyState.GetIndex()

		// Ensure selected item is visible in viewport
		if selectedLine < m.modalView.YOffset {
			// Selected item is above viewport, scroll up to it
			m.modalView.YOffset = selectedLine
		} else if selectedLine >= m.modalView.YOffset+m.modalView.Height {
			// Selected item is below viewport, scroll down to it
			m.modalView.YOffset = selectedLine - m.modalView.Height + 1
		} else {
			// Selected item is already visible, keep current scroll position
			m.modalView.YOffset = listYOffset
		}
	} else {
		// No selection or empty history, go to top
		m.modalView.GotoTop()
	}
}

// wrapText wraps long lines to fit within the specified width
// Preserves indentation and breaks at word boundaries when possible
func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	lines := strings.Split(text, "\n")
	var wrappedLines []string

	for _, line := range lines {
		// If line fits, keep it as-is
		if len(line) <= width {
			wrappedLines = append(wrappedLines, line)
			continue
		}

		// Extract leading whitespace (indentation)
		leadingSpaces := 0
		for i, ch := range line {
			if ch != ' ' && ch != '\t' {
				leadingSpaces = i
				break
			}
		}
		indent := line[:leadingSpaces]

		// Available width after indentation
		availableWidth := width - leadingSpaces
		if availableWidth < 20 {
			// If indentation takes up too much space, reduce it
			availableWidth = width - 2
			if availableWidth < 10 {
				availableWidth = 10
			}
			indent = "  "
		}

		// Break the line into chunks
		remaining := line
		first := true

		for len(remaining) > 0 {
			var chunk string
			var nextRemaining string

			if first {
				// First line keeps original indentation
				if len(remaining) <= width {
					chunk = remaining
					nextRemaining = ""
				} else {
					// Try to break at word boundary
					breakPoint := width
					if breakPoint > len(remaining) {
						breakPoint = len(remaining)
					}
					// Look for space before width
					for i := breakPoint - 1; i > leadingSpaces && i > 0; i-- {
						if remaining[i] == ' ' || remaining[i] == ',' || remaining[i] == ';' {
							breakPoint = i + 1
							break
						}
					}
					if breakPoint > len(remaining) {
						breakPoint = len(remaining)
					}
					chunk = remaining[:breakPoint]
					if breakPoint < len(remaining) {
						nextRemaining = strings.TrimLeft(remaining[breakPoint:], " ")
					} else {
						nextRemaining = ""
					}
				}
				first = false
			} else {
				// Continuation lines use same indentation
				maxLen := availableWidth
				if len(remaining) <= maxLen {
					chunk = indent + remaining
					nextRemaining = ""
				} else {
					// Try to break at word boundary
					breakPoint := maxLen
					if breakPoint > len(remaining) {
						breakPoint = len(remaining)
					}
					// Look backwards for a good break point
					for i := breakPoint - 1; i > 0; i-- {
						if remaining[i] == ' ' || remaining[i] == ',' || remaining[i] == ';' {
							breakPoint = i + 1
							break
						}
					}
					if breakPoint > len(remaining) {
						breakPoint = len(remaining)
					}
					chunk = indent + remaining[:breakPoint]
					if breakPoint < len(remaining) {
						nextRemaining = strings.TrimLeft(remaining[breakPoint:], " ")
					} else {
						nextRemaining = ""
					}
				}
			}

			wrappedLines = append(wrappedLines, chunk)
			remaining = nextRemaining
		}
	}

	return strings.Join(wrappedLines, "\n")
}
