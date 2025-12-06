package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/lipgloss"
	"github.com/studiowebux/restcli/internal/config"
	"github.com/studiowebux/restcli/internal/executor"
	"github.com/studiowebux/restcli/internal/parser"
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

	styleSubtle = lipgloss.NewStyle().
			Foreground(colorGray)

	styleSearchHighlight = lipgloss.NewStyle().
				Background(colorYellow).                                  // Adaptive yellow background
				Foreground(lipgloss.AdaptiveColor{Light: "0", Dark: "0"}) // Black text for both modes

	// Diff background styles for split view highlighting
	styleDiffRemoved = lipgloss.NewStyle().
				Background(lipgloss.AdaptiveColor{Light: "1", Dark: "9"}). // Red / Bright red
				Foreground(lipgloss.AdaptiveColor{Light: "15", Dark: "0"}) // White / Black

	styleDiffAdded = lipgloss.NewStyle().
			Background(lipgloss.AdaptiveColor{Light: "2", Dark: "10"}). // Green / Bright green
			Foreground(lipgloss.AdaptiveColor{Light: "15", Dark: "0"})  // White / Black

	styleDiffNeutral = lipgloss.NewStyle().
				Foreground(colorGray)
)

// highlightJSON applies syntax highlighting to JSON content
func highlightJSON(jsonStr string) string {
	lexer := lexers.Get("json")
	if lexer == nil {
		return jsonStr
	}
	lexer = chroma.Coalesce(lexer)

	// Use monokai style for good contrast in terminals
	style := styles.Get("monokai")
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
		response := m.renderResponse(m.width-4, m.height-5)
		responseBox := lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(colorCyan). // Cyan for focused (fullscreen) panel
			Width(m.width - 2).
			Height(m.height - 3). // Leave room for status bar + top border visibility
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
	responseWidth := m.width - sidebarWidth - 4 // Account for borders

	// Render components with borders
	sidebar := m.renderSidebar(sidebarWidth-2, m.height-5) // -5 = -1 (status) -2 (borders) -2 (top visibility)
	response := m.renderResponse(responseWidth-2, m.height-5)

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
		Height(m.height - 3). // Leave room for status bar + top border visibility
		Padding(0).           // No padding inside box border
		AlignVertical(lipgloss.Top).
		Render(sidebar)

	responseBox := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(responseBorderColor).
		Width(responseWidth).
		Height(m.height - 3). // Leave room for status bar + top border visibility
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

// renderSidebar renders the file list sidebar
func (m Model) renderSidebar(width, height int) string {
	var lines []string

	// Title - bold if focused
	titleStyle := styleTitleUnfocused
	if m.focusedPanel == "sidebar" {
		titleStyle = styleTitleFocused
	}
	title := titleStyle.Render("Files")
	lines = append(lines, title)
	lines = append(lines, "")

	// File list
	pageSize := height - 4 // Reserve space for title, blank lines, footer, and padding
	if pageSize < 1 {
		pageSize = 1
	}

	endIdx := m.fileOffset + pageSize
	if endIdx > len(m.files) {
		endIdx = len(m.files)
	}

	for i := m.fileOffset; i < endIdx; i++ {
		file := m.files[i]

		// Hex number
		hexNum := fmt.Sprintf("%x", i)

		// File name (truncate if too long)
		maxNameLen := width - len(hexNum) - 4
		if maxNameLen < 10 {
			maxNameLen = 10
		}
		name := file.Name
		if len(name) > maxNameLen {
			name = name[:maxNameLen-3] + "..."
		}

		line := fmt.Sprintf("%s %s", hexNum, name)

		// Check if this file is a search match (only when searching files, not response)
		isSearchMatch := false
		if !m.searchInResponseCtx {
			for _, matchIdx := range m.searchMatches {
				if i == matchIdx {
					isSearchMatch = true
					break
				}
			}
		}

		// Apply styling - selected gets green, search matches get yellow
		if i == m.fileIndex {
			line = styleSelected.Render(line)
		} else if isSearchMatch {
			line = styleWarning.Render(line) // Yellow for search matches
		}

		lines = append(lines, line)
	}

	// Footer - show position
	if len(m.files) > 0 {
		lines = append(lines, "")
		footer := fmt.Sprintf("[%d/%d]", m.fileIndex+1, len(m.files))
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
	timing := fmt.Sprintf("Duration: %s | Size: %s",
		executor.FormatDuration(m.currentResponse.Duration),
		executor.FormatSize(m.currentResponse.ResponseSize))
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

	// Show input for special modes
	switch m.mode {
	case ModeGoto:
		right = fmt.Sprintf("Goto: :%s", addCursor(m.gotoInput))
	case ModeSearch:
		right = fmt.Sprintf("Search: %s", addCursor(m.searchQuery))
	default:
		// Show search results if active
		if len(m.searchMatches) > 0 {
			right = styleWarning.Render(fmt.Sprintf("Search: %d of %d | ", m.searchIndex+1, len(m.searchMatches)))
		}

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
		} else if len(m.searchMatches) == 0 {
			right += styleSubtle.Render("Press / to search | ? for help | q to quit")
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
	return m.height - 9
}

// updateViewport updates the response viewport
func (m *Model) updateViewport() {
	// Initialize or update viewport for response scrolling
	// MUST match width calculations in renderMain!
	var responseWidth int
	if m.fullscreen {
		// In fullscreen, use full width
		responseWidth = m.width - 2 // Just account for borders
	} else {
		// In split view, account for sidebar
		sidebarWidth := max(40, m.width*40/100)
		if m.width < 100 {
			sidebarWidth = m.width / 2
		}
		responseWidth = m.width - sidebarWidth - 4 // Account for borders
	}

	// Viewport width = renderResponse width - content padding
	m.responseView.Width = responseWidth - 4 // -2 for renderResponse param, -2 for " " prefix padding in content
	// Viewport height calculation:
	// renderResponse gets: m.height - 5
	// Applies Height(height): m.height - 5
	// Content has: title (1) + blank (1) + viewport
	// So viewport = m.height - 5 - 2 = m.height - 7
	m.responseView.Height = m.height - 7

	// Initialize help viewport
	// Modal width: m.width - 10, Padding: 2 horizontal each side = 4, Border: included
	// So viewport width = (m.width - 10) - 4 (horizontal padding) = m.width - 14
	m.helpView.Width = m.width - 14
	if m.helpView.Width < 10 {
		m.helpView.Width = 10
	}
	// Modal height: m.height - 4, Padding: 1 vertical each side = 2
	// Content: title (1) + blank (1) + viewport + blank (1) + footer (1) = viewport + 4
	// So viewport height = (m.height - 4) - 2 (padding) - 4 (non-viewport lines) = m.height - 10
	m.helpView.Height = m.height - 10
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

		// Resolve variables for display
		resolver := parser.NewVariableResolver(profile.Variables, session.Variables, nil, parser.LoadSystemEnv())
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
	content.WriteString(styleSubtle.Render(fmt.Sprintf("Duration: %s | Size: %s",
		executor.FormatDuration(m.currentResponse.Duration),
		executor.FormatSize(m.currentResponse.ResponseSize))))
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
		content.WriteString(styleTitle.Render("Body") + "\n")

		// Try to pretty-print and highlight JSON
		var bodyText string
		var isJSON bool
		var jsonData interface{}
		if err := json.Unmarshal([]byte(m.currentResponse.Body), &jsonData); err == nil {
			if prettyJSON, err := json.MarshalIndent(jsonData, "", "  "); err == nil {
				bodyText = string(prettyJSON)
				isJSON = true
			} else {
				// Fallback to raw body if formatting fails
				bodyText = m.currentResponse.Body
			}
		} else {
			// Not JSON, show raw body
			bodyText = m.currentResponse.Body
		}

		// Wrap text to viewport width to prevent truncation
		wrapWidth := m.responseView.Width
		if wrapWidth < 40 {
			wrapWidth = 40 // Minimum reasonable width
		}
		wrappedBody := wrapText(bodyText, wrapWidth)

		// Apply syntax highlighting after wrapping (for JSON only)
		if isJSON {
			wrappedBody = highlightJSON(wrappedBody)
		}

		content.WriteString(wrappedBody)
		content.WriteString("\n")
	}

	// Error
	if m.currentResponse.Error != "" {
		content.WriteString("\n")
		// Wrap error text to viewport width to prevent cropping
		wrapWidth := m.responseView.Width
		if wrapWidth < 40 {
			wrapWidth = 40
		}
		errorText := "Error: " + m.currentResponse.Error
		wrappedError := wrapText(errorText, wrapWidth)
		content.WriteString(styleError.Render(wrappedError))
	}

	contentStr := content.String()
	m.responseContent = contentStr

	// Apply search highlighting if we're searching in response
	if m.searchInResponseCtx && len(m.searchMatches) > 0 {
		contentStr = m.highlightSearchMatches(contentStr)
	}

	m.responseView.SetContent(contentStr)
}

// highlightSearchMatches highlights lines that match the current search
func (m *Model) highlightSearchMatches(content string) string {
	lines := strings.Split(content, "\n")

	// Create a map of line numbers to highlight for faster lookup
	matchLines := make(map[int]bool)
	for _, lineNum := range m.searchMatches {
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
	helpText := `REST CLI - Keyboard Shortcuts

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

RESPONSE
  s            Save response to file
  c            Copy full response to clipboard
  b            Toggle body visibility
  B            Toggle headers visibility (request + response)
  f            Toggle fullscreen (ESC to exit)
  w            Pin response for comparison
  W            Show diff (compare pinned vs current)
  ↑/↓, j/k     Scroll response (when body shown)

CONFIGURATION
  v            Variable editor
  h            Header editor
  p            Switch profile (press [e] to edit, [d] duplicate, [D] delete)
  n            Create new profile (when no search active)
  C            View current configuration
  P            Edit .profiles.json
  S            Edit .session.json

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
  p            Toggle preview pane (when in history/analytics)
  t            Toggle grouping (when in analytics: per-file ↔ by path)
  r            Replay request (when in history modal)
  C            Clear all (when in history/analytics modal - with confirmation)

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

TEXT INPUT (in modals)
  Ctrl+V       Paste from clipboard (recommended)
  Cmd+V        Paste from clipboard (macOS, may not work in Terminal.app)
  Shift+Insert Paste (alternative)
  Ctrl+Y       Paste (alternative)
  Ctrl+K       Clear input
  Backspace    Delete character

OTHER
  ?            Show this help
  e            View full error details (when error in footer)
  q            Quit

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
	m.modalView.Width = m.width - 10  // Modal content width minus padding
	m.modalView.Height = m.height - 9 // Modal content height with footer

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
		collapsed := m.docCollapsed[index]
		marker := "▼" // Expanded
		if collapsed {
			marker = "▶" // Collapsed
		}

		// Highlight if selected
		line := fmt.Sprintf("%s %s", marker, title)
		if currentIdx == m.docSelectedIdx {
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
				if currentIdx == m.docSelectedIdx {
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
				if currentIdx == m.docSelectedIdx {
					line = styleSelected.Render(line)
					selectedLineNum = strings.Count(content.String(), "\n")
				}
				content.WriteString(line + "\n")
				currentIdx++

				// Response fields - lazy rendering (only build tree when response is expanded)
				if len(resp.Fields) > 0 {
					responseKey := 100 + respIdx
					responseFieldsCollapsed := m.docCollapsed[responseKey]

					if !responseFieldsCollapsed {
						// Response fields are expanded - use cached tree
						allFields, ok := m.docFieldTreeCache[respIdx]
						if !ok {
							// Build and cache the tree
							allFields = buildVirtualFieldTree(resp.Fields)
							m.docFieldTreeCache[respIdx] = allFields
							m.docChildrenCache[respIdx] = buildHasChildrenCache(allFields)
						}
						hasChildrenCache := m.docChildrenCache[respIdx]
						m.renderResponseFieldsTree(respIdx, "", allFields, hasChildrenCache, &currentIdx, &content, 0, &selectedLineNum)
					} else {
						// Response fields are collapsed - show indicator only (no tree building!)
						line := fmt.Sprintf("    ▶ %d fields", len(resp.Fields))
						if currentIdx == m.docSelectedIdx {
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
	m.modalView.Width = m.width - 10  // Modal content width minus padding
	m.modalView.Height = m.height - 9 // Modal content height minus padding, title lines, and footer

	if m.currentRequest == nil {
		m.modalView.SetContent("No request selected\n\nPress ESC to close")
		return
	}

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
	modalWidth := m.width - 6
	modalHeight := m.height - 3
	paneHeight := modalHeight - 4 // Account for borders and padding

	// Adjust viewport widths based on preview visibility
	if m.historyPreviewVisible {
		// Split view mode: calculate widths for both panes
		listWidth := (modalWidth - 3) / 2          // Left pane width
		previewWidth := modalWidth - listWidth - 3 // Right pane width

		// Set viewport dimensions for left pane (history list)
		m.modalView.Width = listWidth - 4   // Account for padding and borders
		m.modalView.Height = paneHeight - 2 // Account for title

		// Set viewport dimensions for right pane (response preview)
		m.historyPreviewView.Width = previewWidth - 4
		m.historyPreviewView.Height = paneHeight - 2
	} else {
		// Preview hidden: expand list to full width
		m.modalView.Width = modalWidth - 4  // Account for padding and borders
		m.modalView.Height = paneHeight - 2 // Account for title
	}

	// Build content for left pane (history list)
	var listContent strings.Builder
	if len(m.historyEntries) == 0 {
		listContent.WriteString("No history entries")
	} else {
		// Show ALL entries (not just first 10) - viewport handles scrolling
		for i, entry := range m.historyEntries {
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
			if i == m.historyIndex {
				line = styleSelected.Render(line)
			}

			listContent.WriteString(line + "\n")
		}
	}

	// Build content for right pane (response preview) - ONLY if preview is visible
	if m.historyPreviewVisible {
		var previewContent strings.Builder
		if len(m.historyEntries) > 0 && m.historyIndex >= 0 && m.historyIndex < len(m.historyEntries) {
			entry := m.historyEntries[m.historyIndex]

			// Show response metadata
			previewContent.WriteString(fmt.Sprintf("%s %s\n", entry.Method, entry.URL))
			previewContent.WriteString(fmt.Sprintf("Status: %d %s\n", entry.ResponseStatus, entry.ResponseStatusText))
			previewContent.WriteString(fmt.Sprintf("Size: %d bytes\n", entry.ResponseSize))
			previewContent.WriteString(fmt.Sprintf("Time: %s\n\n", entry.Timestamp[:19]))

			// Show response body
			previewContent.WriteString(entry.ResponseBody)
		} else {
			previewContent.WriteString("No history entry selected")
		}

		// Set preview content (always start from top when selection changes)
		m.historyPreviewView.SetContent(previewContent.String())
		m.historyPreviewView.GotoTop()
	} else {
		// Clear preview content when hidden (security/privacy)
		m.historyPreviewView.SetContent("")
	}

	// Save current scroll positions before updating content
	listYOffset := m.modalView.YOffset
	m.modalView.SetContent(listContent.String())

	// Auto-scroll list to keep selected item visible
	if len(m.historyEntries) > 0 && m.historyIndex >= 0 {
		selectedLine := m.historyIndex

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
