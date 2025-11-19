package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/studiowebux/restcli/internal/config"
	"github.com/studiowebux/restcli/internal/executor"
	"github.com/studiowebux/restcli/internal/parser"
)

// Adaptive color definitions for light/dark terminal support
var (
	colorGreen = lipgloss.AdaptiveColor{Light: "#006400", Dark: "#00ff00"} // Dark green / Bright green
	colorRed   = lipgloss.AdaptiveColor{Light: "#8b0000", Dark: "#ff0000"} // Dark red / Bright red
	colorYellow = lipgloss.AdaptiveColor{Light: "#b8860b", Dark: "#ffff00"} // Dark goldenrod / Yellow
	colorBlue  = lipgloss.AdaptiveColor{Light: "#00008b", Dark: "#0000ff"} // Dark blue / Blue
	colorGray  = lipgloss.AdaptiveColor{Light: "#555555", Dark: "#888888"} // Dark gray / Light gray
	colorCyan  = lipgloss.AdaptiveColor{Light: "#008b8b", Dark: "#00ffff"} // Dark cyan / Cyan
)

// Style definitions
var (
	styleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorCyan)

	styleSelected = lipgloss.NewStyle().
			Background(lipgloss.AdaptiveColor{Light: "#d3d3d3", Dark: "#3a3a3a"}).
			Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#ffffff"})

	styleSuccess = lipgloss.NewStyle().
			Foreground(colorGreen)

	styleError = lipgloss.NewStyle().
			Foreground(colorRed)

	styleWarning = lipgloss.NewStyle().
			Foreground(colorYellow)

	styleSubtle = lipgloss.NewStyle().
			Foreground(colorGray)
)

// renderMain renders the main TUI view (file sidebar + response panel)
func (m Model) renderMain() string {
	if m.width == 0 {
		return ""
	}

	// Fullscreen mode - show only response
	if m.fullscreen {
		response := m.renderResponse(m.width-4, m.height-3)
		responseBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorGreen). // Always green in fullscreen
			Width(m.width - 2).
			Height(m.height - 1). // Leave 1 line for status bar
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
	sidebar := m.renderSidebar(sidebarWidth-2, m.height-3) // -3 = -1 (status) -2 (borders)
	response := m.renderResponse(responseWidth-2, m.height-3)

	// Add borders - highlight the focused panel
	sidebarBorderColor := colorGray
	responseBorderColor := colorGray
	if m.focusedPanel == "sidebar" {
		sidebarBorderColor = colorGreen
	} else if m.focusedPanel == "response" {
		responseBorderColor = colorGreen
	}

	sidebarBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(sidebarBorderColor).
		Width(sidebarWidth).
		Height(m.height - 1). // Leave 1 line for status bar
		Render(sidebar)

	responseBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(responseBorderColor).
		Width(responseWidth).
		Height(m.height - 1). // Leave 1 line for status bar
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

	// Title
	title := styleTitle.Render("Files")
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

		// Check if this file is a search match
		isSearchMatch := false
		for _, matchIdx := range m.searchMatches {
			if i == matchIdx {
				isSearchMatch = true
				break
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
		Height(height - 5).
		Padding(1)

	return style.Render(content)
}

// renderResponse renders the response panel
func (m Model) renderResponse(width, height int) string {
	if m.currentResponse == nil {
		noResponse := styleSubtle.Render("No response yet\n\nPress Enter to execute request\n\nPress 'b' to toggle body visibility")
		return lipgloss.NewStyle().
			MaxWidth(width).
			Height(height).
			Padding(1).
			AlignHorizontal(lipgloss.Left).
			Render(noResponse)
	}

	// If body is shown, use viewport for scrolling
	if m.showBody {
		// Return viewport directly with padding applied to container
		viewportContent := m.responseView.View()
		// Add padding manually
		paddedLines := []string{""}
		for _, line := range strings.Split(viewportContent, "\n") {
			paddedLines = append(paddedLines, " "+line)
		}
		return strings.Join(paddedLines, "\n")
	}

	// Otherwise show just status and headers (no scrolling needed)
	var lines []string

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
		Padding(1).
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
	return m.height - 7 // Account for title, footer, status bar
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

	// Viewport width = renderResponse width - 2 (padding)
	m.responseView.Width = responseWidth - 4 // -2 for renderResponse param, -2 for padding
	m.responseView.Height = m.height - 5     // -5 = -1 (status) -2 (borders) -2 (padding)

	// Initialize help viewport
	m.helpView.Width = m.width - 4
	m.helpView.Height = m.height - 4
}

// updateResponseView updates the response viewport content
func (m *Model) updateResponseView() {
	if m.currentResponse == nil {
		m.responseView.SetContent("")
		return
	}

	var content strings.Builder

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

		// Try to pretty-print JSON
		var bodyText string
		var jsonData interface{}
		if err := json.Unmarshal([]byte(m.currentResponse.Body), &jsonData); err == nil {
			if prettyJSON, err := json.MarshalIndent(jsonData, "", "  "); err == nil {
				bodyText = string(prettyJSON)
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
		content.WriteString(wrappedBody)
		content.WriteString("\n")
	}

	// Error
	if m.currentResponse.Error != "" {
		content.WriteString("\n")
		content.WriteString(styleError.Render("Error: " + m.currentResponse.Error))
	}

	m.responseView.SetContent(content.String())
}

// updateHelpView updates the help viewport content
func (m *Model) updateHelpView() {
	helpText := `REST CLI - Keyboard Shortcuts

NAVIGATION (Keyboard Only)
  TAB            Switch focus (sidebar ↔ response)
  ↑/↓, j/k       Navigate files OR scroll response
  gg             Go to first item (vim-style)
  G              Go to last item (vim-style)
  Ctrl+U/D       Half-page up/down (vim-style)
  PageUp/Down    Full page scroll (focused panel)
  Home/End       Jump to first/last
  :              Goto hex line

SEARCH
  /              Search files
  n              Next search result (vim-style)
  N              Previous search result (vim-style)
  Ctrl+R         Next search result (alternative)
  ESC            Clear search / Cancel

FOCUS
  Green border   Shows which panel is focused
  Arrow keys     Control the focused panel only

FILE OPERATIONS
  Enter        Execute request
  i            Inspect request
  x            Open in editor
  X            Configure editor
  d            Duplicate file
  D            Delete file (with confirmation)
  R            Rename file
  r            Refresh file list

RESPONSE
  s            Save response to file
  c            Copy full response to clipboard
  b            Toggle body visibility
  B            Toggle headers visibility (request + response)
  f            Toggle fullscreen (ESC to exit)
  ↑/↓, j/k     Scroll response (when body shown)

CONFIGURATION
  v            Variable editor
  h            Header editor
  p            Switch profile (press [e] to edit profile)
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

DOCUMENTATION & HISTORY
  m            View documentation
  H            View history

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
  Cmd+V        Paste from clipboard (macOS)
  Ctrl+V       Paste from clipboard
  Shift+Insert Paste (alternative)
  Ctrl+Y       Paste (alternative)
  Ctrl+K       Clear input
  Backspace    Delete character

OTHER
  ?            Show this help
  q            Quit

Use ↑/↓ or j/k to scroll, / to search, ESC or ? to close`

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
		}
	}

	m.modalView.SetContent(content.String())
	m.modalView.GotoTop()
}

// updateHistoryView updates the history viewport content
func (m *Model) updateHistoryView() {
	// Set viewport dimensions for the modal (nearly full screen: m.width-6, m.height-3)
	m.modalView.Width = m.width - 10  // Modal content width minus padding
	m.modalView.Height = m.height - 7 // Modal content height minus padding and title lines

	var content strings.Builder

	if len(m.historyEntries) == 0 {
		content.WriteString("No history entries\n\nPress ESC to close")
	} else {
		content.WriteString(fmt.Sprintf("%d history entries\n\n", len(m.historyEntries)))
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

			content.WriteString(line + "\n")
		}
		content.WriteString("\n" + styleSubtle.Render("↑/↓: Navigate | Enter: Load | ESC: Close"))
	}

	// Save current scroll position before updating content
	yOffset := m.modalView.YOffset
	m.modalView.SetContent(content.String())

	// If we have a selected item, scroll to make it visible
	if len(m.historyEntries) > 0 && m.historyIndex >= 0 {
		// Each entry is 1 line, plus 2 lines for header
		selectedLine := m.historyIndex + 2

		// Ensure selected item is visible in viewport
		if selectedLine < m.modalView.YOffset {
			// Selected item is above viewport, scroll up to it
			m.modalView.YOffset = selectedLine
		} else if selectedLine >= m.modalView.YOffset+m.modalView.Height {
			// Selected item is below viewport, scroll down to it
			m.modalView.YOffset = selectedLine - m.modalView.Height + 1
		} else {
			// Selected item is already visible, keep current scroll position
			m.modalView.YOffset = yOffset
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
