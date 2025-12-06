package tui

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/studiowebux/restcli/internal/stresstest"
)

// handleKeyPress routes key presses based on current mode
func (m *Model) handleKeyPress(msg tea.KeyMsg) tea.Cmd {
	// Global keys (work in all modes)
	switch msg.String() {
	case "ctrl+c":
		m.Cleanup()
		return tea.Quit
	case "q":
		// If streaming is active, stop it
		if m.streamingActive && m.streamCancelFunc != nil {
			m.streamCancelFunc()
			m.streamingActive = false
			m.loading = false // Clear loading flag when stopping stream
			m.statusMsg = "Stream stopped by user"
			return nil
		}
	}

	// Mode-specific handling
	switch m.mode {
	case ModeNormal:
		return m.handleNormalKeys(msg)
	case ModeSearch:
		return m.handleSearchKeys(msg)
	case ModeGoto:
		return m.handleGotoKeys(msg)
	case ModeVariableList, ModeVariableAdd, ModeVariableEdit, ModeVariableDelete, ModeVariableOptions, ModeVariableManage, ModeVariableAlias:
		return m.handleVariableEditorKeys(msg)
	case ModeVariablePromptInteractive:
		return m.handleInteractiveVariablePromptKeys(msg)
	case ModeHeaderList, ModeHeaderAdd, ModeHeaderEdit, ModeHeaderDelete:
		return m.handleHeaderEditorKeys(msg)
	case ModeProfileSwitch, ModeProfileCreate, ModeProfileEdit, ModeProfileDuplicate, ModeProfileDeleteConfirm:
		return m.handleProfileKeys(msg)
	case ModeDocumentation:
		return m.handleDocumentationKeys(msg)
	case ModeHistory:
		return m.handleHistoryKeys(msg)
	case ModeHistoryClearConfirm:
		return m.handleHistoryClearConfirmKeys(msg)
	case ModeAnalytics:
		return m.handleAnalyticsKeys(msg)
	case ModeAnalyticsClearConfirm:
		return m.handleAnalyticsClearConfirmKeys(msg)
	case ModeStressTestConfig:
		return m.handleStressTestConfigKeys(msg)
	case ModeStressTestLoadConfig:
		return m.handleStressTestLoadConfigKeys(msg)
	case ModeStressTestProgress:
		return m.handleStressTestProgressKeys(msg)
	case ModeStressTestResults:
		return m.handleStressTestResultsKeys(msg)
	case ModeHelp:
		return m.handleHelpKeys(msg)
	case ModeInspect:
		return m.handleInspectKeys(msg)
	case ModeRename:
		return m.handleRenameKeys(msg)
	case ModeOAuthConfig:
		return m.handleOAuthKeys(msg)
	case ModeOAuthEdit:
		return m.handleOAuthEditKeys(msg)
	case ModeEditorConfig:
		return m.handleEditorConfigKeys(msg)
	case ModeConfigView:
		return m.handleConfigViewKeys(msg)
	case ModeDelete:
		return m.handleDeleteKeys(msg)
	case ModeConfirmExecution:
		return m.handleConfirmExecutionKeys(msg)
	case ModeShellErrors:
		return m.handleShellErrorsKeys(msg)
	case ModeErrorDetail:
		return m.handleErrorDetailKeys(msg)
	case ModeStatusDetail:
		return m.handleStatusDetailKeys(msg)
	case ModeCreateFile:
		return m.handleCreateFileKeys(msg)
	case ModeMRU:
		return m.handleMRUKeys(msg)
	case ModeDiff:
		return m.handleDiffKeys(msg)
	}

	return nil
}

// handleShellErrorsKeys handles keyboard input in shell errors modal
func (m *Model) handleShellErrorsKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "q", "enter":
		m.mode = ModeNormal
		m.shellErrors = nil
	case "j", "down":
		m.modalView.LineDown(1)
	case "k", "up":
		m.modalView.LineUp(1)
	case "g":
		m.modalView.GotoTop()
	case "G":
		m.modalView.GotoBottom()
	}
	return nil
}

// handleErrorDetailKeys handles keyboard input in error detail modal
func (m *Model) handleErrorDetailKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "q", "enter":
		m.mode = ModeNormal
	case "j", "down":
		m.modalView.LineDown(1)
	case "k", "up":
		m.modalView.LineUp(1)
	case "g":
		m.modalView.GotoTop()
	case "G":
		m.modalView.GotoBottom()
	}
	return nil
}

// handleStatusDetailKeys handles keyboard input in status detail modal
func (m *Model) handleStatusDetailKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "q", "enter":
		m.mode = ModeNormal
	}
	return nil
}

// handleEditorConfigKeys handles keyboard input in editor config mode
func (m *Model) handleEditorConfigKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		m.inputValue = ""
		m.inputCursor = 0

	case "enter":
		// Save editor to profile
		profile := m.sessionMgr.GetActiveProfile()
		profile.Editor = m.inputValue
		m.sessionMgr.SaveProfiles()
		m.statusMsg = "Editor saved: " + m.inputValue
		m.mode = ModeNormal
		m.inputValue = ""
		m.inputCursor = 0

	default:
		// Handle text input with cursor support
		if _, shouldContinue := handleTextInputWithCursor(&m.inputValue, &m.inputCursor, msg); shouldContinue {
			return nil
		}
		// Insert character at cursor position
		if len(msg.String()) == 1 {
			m.inputValue = m.inputValue[:m.inputCursor] + msg.String() + m.inputValue[m.inputCursor:]
			m.inputCursor++
		}
	}

	return nil
}

// handleNormalKeys handles keys in normal mode
func (m *Model) handleNormalKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "q":
		m.Cleanup()
		return tea.Quit

	// Focus switching
	case "tab":
		// Toggle focus between sidebar and response
		if m.focusedPanel == "sidebar" {
			m.focusedPanel = "response"
			m.statusMsg = "Focus: Response panel (use TAB to switch back)"
		} else {
			m.focusedPanel = "sidebar"
			m.statusMsg = "Focus: File sidebar (use TAB to switch)"
		}

	// Navigation - based on focused panel (EXCLUSIVE control)
	case "up", "k":
		if m.focusedPanel == "response" {
			// Only scroll response if focused on response panel
			if m.showBody && m.currentResponse != nil {
				m.responseView.ScrollUp(1)
			}
		} else {
			// Only navigate files if focused on sidebar
			m.navigateFiles(-1)
		}
	case "down", "j":
		if m.focusedPanel == "response" {
			// Only scroll response if focused on response panel
			if m.showBody && m.currentResponse != nil {
				m.responseView.ScrollDown(1)
			}
		} else {
			// Only navigate files if focused on sidebar
			m.navigateFiles(1)
		}
	case "pgup":
		if m.focusedPanel == "response" {
			// Only scroll response if focused on response panel
			if m.showBody && m.currentResponse != nil {
				m.responseView.PageUp()
			}
		} else {
			// Only navigate files if focused on sidebar
			m.navigateFiles(-10)
		}
	case "pgdown":
		if m.focusedPanel == "response" {
			// Only scroll response if focused on response panel
			if m.showBody && m.currentResponse != nil {
				m.responseView.PageDown()
			}
		} else {
			// Only navigate files if focused on sidebar
			m.navigateFiles(10)
		}
	case "ctrl+u":
		// Vim-style half-page up
		halfPage := m.getFileListHeight() / 2
		if halfPage < 1 {
			halfPage = 5
		}
		if m.focusedPanel == "response" {
			if m.showBody && m.currentResponse != nil {
				m.responseView.ScrollUp(halfPage)
			}
		} else {
			m.navigateFiles(-halfPage)
		}
	case "ctrl+d":
		// Vim-style half-page down
		halfPage := m.getFileListHeight() / 2
		if halfPage < 1 {
			halfPage = 5
		}
		if m.focusedPanel == "response" {
			if m.showBody && m.currentResponse != nil {
				m.responseView.ScrollDown(halfPage)
			}
		} else {
			m.navigateFiles(halfPage)
		}
	case "home":
		if m.focusedPanel == "response" {
			// Scroll to top of response
			if m.showBody && m.currentResponse != nil {
				m.responseView.GotoTop()
			}
		} else {
			// Go to first file
			if len(m.files) > 0 {
				m.fileIndex = 0
				m.fileOffset = 0
				m.loadRequestsFromCurrentFile()
			}
		}
	case "end":
		if m.focusedPanel == "response" {
			// Scroll to bottom of response
			if m.showBody && m.currentResponse != nil {
				m.responseView.GotoBottom()
			}
		} else {
			// Go to last file
			if len(m.files) > 0 {
				m.fileIndex = len(m.files) - 1
				pageSize := m.getFileListHeight()
				m.fileOffset = max(0, m.fileIndex-pageSize+1)
				m.loadRequestsFromCurrentFile()
			}
		}
	case "g":
		// Vim-style 'gg' to go to top
		if m.gPressed {
			m.gPressed = false
			if m.focusedPanel == "response" {
				if m.showBody && m.currentResponse != nil {
					m.responseView.GotoTop()
				}
			} else {
				if len(m.files) > 0 {
					m.fileIndex = 0
					m.fileOffset = 0
					m.loadRequestsFromCurrentFile()
				}
			}
		} else {
			m.gPressed = true
		}
		return nil // Don't reset gPressed
	case "G":
		// Vim-style 'G' to go to bottom
		if m.focusedPanel == "response" {
			if m.showBody && m.currentResponse != nil {
				m.responseView.GotoBottom()
			}
		} else {
			if len(m.files) > 0 {
				m.fileIndex = len(m.files) - 1
				pageSize := m.getFileListHeight()
				m.fileOffset = max(0, m.fileIndex-pageSize+1)
				m.loadRequestsFromCurrentFile()
			}
		}
	case ":":
		m.mode = ModeGoto
		m.gotoInput = ""

	// File operations
	case "enter":
		// Block if request already in progress
		if m.loading {
			m.statusMsg = "Request already in progress"
			return nil
		}
		m.statusMsg = "Executing request..."
		return m.executeRequest()
	case "i":
		m.mode = ModeInspect
		m.updateInspectView() // Set content once when entering modal
	case "x":
		return m.openInEditor()
	case "X":
		// Configure editor
		m.mode = ModeEditorConfig
		profile := m.sessionMgr.GetActiveProfile()
		m.inputValue = profile.Editor
		m.inputCursor = len(m.inputValue)
	case "d":
		return m.duplicateFile()
	case "D":
		// Delete file with confirmation
		if len(m.files) > 0 {
			m.mode = ModeDelete
		}
	case "R":
		m.mode = ModeRename
		m.renameInput = ""
		m.renameCursor = 0
	case "F":
		// Create new file
		m.mode = ModeCreateFile
		m.createFileInput = ""
		m.createFileCursor = 0
		m.createFileType = 0 // Default to .http
		m.errorMsg = ""
	case "r":
		return m.refreshFiles()

	// Response operations
	case "s":
		return m.saveResponse()
	case "c":
		return m.copyToClipboard()
	case "b":
		m.showBody = !m.showBody
	case "B":
		m.showHeaders = !m.showHeaders
		m.updateResponseView() // Regenerate response content
	case "f":
		m.fullscreen = !m.fullscreen
		m.updateViewport()       // Recalculate viewport width for fullscreen
		m.updateResponseView()   // Regenerate content (wrapping changes based on fullscreen)
	case "w":
		// Pin current response for comparison
		if m.currentResponse != nil {
			m.pinnedResponse = m.currentResponse
			m.pinnedRequest = m.currentRequest
			m.statusMsg = "Response pinned for comparison (press W to view diff)"
		} else {
			m.errorMsg = "No response to pin"
		}
	case "W":
		// Show diff between pinned and current response
		if m.pinnedResponse == nil {
			m.errorMsg = "No pinned response (press 'w' to pin current response first)"
		} else if m.currentResponse == nil {
			m.errorMsg = "No current response to compare"
		} else {
			m.mode = ModeDiff
			m.updateDiffView()
		}

	// Editors and modals
	case "v":
		m.mode = ModeVariableList
		m.varEditIndex = 0
		m.modalView.SetYOffset(0)
	case "h":
		m.mode = ModeHeaderList
		m.headerEditIndex = 0
		m.modalView.SetYOffset(0)
	case "e":
		// Open error detail modal if there's an error
		if m.fullErrorMsg != "" {
			m.mode = ModeErrorDetail
		}
	case "I":
		// Open status detail modal if there's a status message
		if m.fullStatusMsg != "" {
			m.mode = ModeStatusDetail
		}
	case "p":
		m.mode = ModeProfileSwitch
		m.profileIndex = 0
	case "n":
		// If search is active, go to next match (vim-style)
		if len(m.searchMatches) > 0 {
			m.searchIndex = (m.searchIndex + 1) % len(m.searchMatches)

			if m.searchInResponseCtx {
				// Navigate in response
				m.responseView.SetYOffset(m.centerLineInViewport(m.searchMatches[m.searchIndex]))
				context := "text"
				if isRegexPattern(m.searchQuery) {
					context = "regex"
				}
				m.statusMsg = fmt.Sprintf("[Response] Match %d of %d (%s)", m.searchIndex+1, len(m.searchMatches), context)
			} else {
				// Navigate in files
				m.fileIndex = m.searchMatches[m.searchIndex]
				m.adjustScrollOffset()
				m.loadRequestsFromCurrentFile()
				context := "text"
				if isRegexPattern(m.searchQuery) {
					context = "regex"
				}
				m.statusMsg = fmt.Sprintf("[Files] Match %d of %d (%s)", m.searchIndex+1, len(m.searchMatches), context)
			}
		} else {
			// No active search - create new profile
			m.mode = ModeProfileCreate
			m.profileName = ""
		}
	case "N":
		// Previous search result (vim-style)
		if len(m.searchMatches) > 0 {
			m.searchIndex--
			if m.searchIndex < 0 {
				m.searchIndex = len(m.searchMatches) - 1
			}

			if m.searchInResponseCtx {
				// Navigate in response
				m.responseView.SetYOffset(m.centerLineInViewport(m.searchMatches[m.searchIndex]))
				context := "text"
				if isRegexPattern(m.searchQuery) {
					context = "regex"
				}
				m.statusMsg = fmt.Sprintf("[Response] Match %d of %d (%s)", m.searchIndex+1, len(m.searchMatches), context)
			} else {
				// Navigate in files
				m.fileIndex = m.searchMatches[m.searchIndex]
				m.adjustScrollOffset()
				m.loadRequestsFromCurrentFile()
				context := "text"
				if isRegexPattern(m.searchQuery) {
					context = "regex"
				}
				m.statusMsg = fmt.Sprintf("[Files] Match %d of %d (%s)", m.searchIndex+1, len(m.searchMatches), context)
			}
		}
	case "m":
		m.mode = ModeDocumentation
		// Initialize caches for field trees (prevents rebuilding on every navigation)
		m.docFieldTreeCache = make(map[int][]DocField)
		m.docChildrenCache = make(map[int]map[string]bool)
		m.updateDocumentationView() // Set content and initialize collapse state
		m.docItemCount = m.countDocItems() // Cache item count
	case "H":
		m.mode = ModeHistory
		return m.loadHistory()
	case "A":
		m.mode = ModeAnalytics
		m.analyticsPreviewVisible = true
		m.analyticsGroupByPath = false
		return m.loadAnalytics()
	case "S":
		m.mode = ModeStressTestResults
		m.stressTestFocusedPane = "list"
		return m.loadStressTestRuns()
	case "?":
		m.mode = ModeHelp
		m.updateHelpView()

	// OAuth
	case "o":
		return m.startOAuthFlow()
	case "O":
		m.mode = ModeOAuthConfig

	// Search
	case "/":
		m.mode = ModeSearch
		m.searchQuery = ""
		m.searchMatches = nil
		m.searchIndex = 0

	case "ctrl+r":
		// Next search result
		if len(m.searchMatches) > 0 {
			m.searchIndex = (m.searchIndex + 1) % len(m.searchMatches)
			m.fileIndex = m.searchMatches[m.searchIndex]
			m.adjustScrollOffset()
			m.loadRequestsFromCurrentFile()
			m.statusMsg = fmt.Sprintf("Match %d of %d", m.searchIndex+1, len(m.searchMatches))
		} else {
			m.statusMsg = "No active search - press / to search"
		}

	case "ctrl+p":
		// Open MRU (Most Recently Used) files list
		m.mode = ModeMRU
		m.mruIndex = 0
		m.errorMsg = ""

	// View configuration
	case "C":
		m.mode = ModeConfigView

	// External config editors
	case "P":
		return m.openProfilesInEditor()
	case "ctrl+x":
		return m.openSessionInEditor()

	// Escape - cancel request, exit fullscreen, clear search, or clear errors/messages
	case "esc":
		// First priority: Cancel running request
		if m.loading {
			if m.streamingActive && m.streamCancelFunc != nil {
				// Cancel streaming request
				m.streamCancelFunc()
				m.streamingActive = false
				m.streamCancelFunc = nil
			} else if m.requestCancelFunc != nil {
				// Cancel regular request
				m.requestCancelFunc()
				m.requestCancelFunc = nil
			}
			m.loading = false
			m.statusMsg = "Request cancelled by user"
			m.updateResponseView() // Remove loading indicator
		} else if m.fullscreen {
			m.fullscreen = false
			m.updateViewport()
			m.updateResponseView()
		} else if len(m.searchMatches) > 0 {
			// Clear active search
			m.searchMatches = nil
			m.searchQuery = ""
			m.searchIndex = 0
			m.statusMsg = "Search cleared"
		} else {
			m.errorMsg = ""
			m.statusMsg = ""
		}
	}

	// Reset 'g' state on any key except 'g' itself (handled above with return)
	m.gPressed = false
	return nil
}

// handleTextInput handles common text input operations (paste, clear, backspace)
// Returns: modified (bool), shouldContinue (bool)
// Note: This is the old version that only appends. Use handleTextInputWithCursor for proper cursor support.
func handleTextInput(input *string, msg tea.KeyMsg) (modified bool, shouldContinue bool) {
	switch msg.String() {
	case "ctrl+v", "shift+insert", "super+v":
		// Paste from clipboard (Ctrl+V, Shift+Insert, or Cmd+V on macOS)
		if text, err := clipboard.ReadAll(); err == nil {
			*input += text
			return true, true
		}
		// If clipboard read fails, don't block - just return
		return false, true
	case "ctrl+y":
		// Alternative paste (common in some terminals)
		if text, err := clipboard.ReadAll(); err == nil {
			*input += text
			return true, true
		}
		return false, true
	case "ctrl+k":
		// Clear input
		if *input != "" {
			*input = ""
			return true, true
		}
		return false, true
	case "backspace":
		// Delete last character
		if len(*input) > 0 {
			*input = (*input)[:len(*input)-1]
			return true, true
		}
		return false, true
	}
	return false, false
}

// handleTextInputWithCursor handles text input with cursor position support
// Returns: modified (bool), shouldContinue (bool)
func handleTextInputWithCursor(input *string, cursorPos *int, msg tea.KeyMsg) (modified bool, shouldContinue bool) {
	// Ensure cursor position is valid
	if *cursorPos < 0 {
		*cursorPos = 0
	}
	if *cursorPos > len(*input) {
		*cursorPos = len(*input)
	}

	switch msg.String() {
	case "left":
		// Move cursor left
		if *cursorPos > 0 {
			*cursorPos--
		}
		return true, true

	case "right":
		// Move cursor right
		if *cursorPos < len(*input) {
			*cursorPos++
		}
		return true, true

	case "home", "ctrl+a":
		// Move to start
		*cursorPos = 0
		return true, true

	case "end", "ctrl+e":
		// Move to end
		*cursorPos = len(*input)
		return true, true

	case "ctrl+v", "shift+insert", "super+v":
		// Paste from clipboard at cursor position (Ctrl+V, Shift+Insert, or Cmd+V on macOS)
		if text, err := clipboard.ReadAll(); err == nil {
			*input = (*input)[:*cursorPos] + text + (*input)[*cursorPos:]
			*cursorPos += len(text)
			return true, true
		}
		return false, true

	case "ctrl+y":
		// Alternative paste
		if text, err := clipboard.ReadAll(); err == nil {
			*input = (*input)[:*cursorPos] + text + (*input)[*cursorPos:]
			*cursorPos += len(text)
			return true, true
		}
		return false, true

	case "ctrl+k":
		// Clear input
		if *input != "" {
			*input = ""
			*cursorPos = 0
			return true, true
		}
		return false, true

	case "backspace":
		// Delete character before cursor
		if *cursorPos > 0 {
			*input = (*input)[:*cursorPos-1] + (*input)[*cursorPos:]
			*cursorPos--
			return true, true
		}
		return false, true

	case "delete":
		// Delete character at cursor
		if *cursorPos < len(*input) {
			*input = (*input)[:*cursorPos] + (*input)[*cursorPos+1:]
			return true, true
		}
		return false, true
	}

	return false, false
}

// handleSearchKeys handles keys in search mode
func (m *Model) handleSearchKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		// Cancel search - clear query and matches
		m.mode = ModeNormal
		m.searchQuery = ""
		m.searchMatches = nil
		m.searchIndex = 0

	case "enter":
		m.mode = ModeNormal
		m.performSearch()

	case "ctrl+r":
		// Don't append ctrl+r to search, ignore it
		return nil

	default:
		// Handle common text input operations
		if _, shouldContinue := handleTextInput(&m.searchQuery, msg); shouldContinue {
			return nil
		}
		// Only append single printable characters
		if len(msg.String()) == 1 {
			m.searchQuery += msg.String()
		}
	}
	return nil
}

// handleGotoKeys handles keys in goto mode
func (m *Model) handleGotoKeys(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
		m.mode = ModeNormal
		m.gotoInput = ""
	case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
		m.mode = ModeNormal
		m.performGoto()
	default:
		// Handle common text input operations (paste, clear, backspace)
		if _, shouldContinue := handleTextInput(&m.gotoInput, msg); shouldContinue {
			// For goto, filter to hex characters only after paste
			filtered := ""
			for _, ch := range m.gotoInput {
				if (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F') {
					filtered += string(ch)
				}
			}
			m.gotoInput = filtered
			return nil
		}
		// Append hex character to goto input (0-9, a-f)
		if len(msg.String()) == 1 {
			ch := msg.String()[0]
			if (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F') {
				m.gotoInput += msg.String()
			}
		}
	}
	return nil
}

// handleHelpKeys handles keys in help mode
func (m *Model) handleHelpKeys(msg tea.KeyMsg) tea.Cmd {
	// Handle search mode
	if m.helpSearchActive {
		switch msg.String() {
		case "esc":
			m.helpSearchActive = false
			m.helpSearchQuery = ""
			m.updateHelpView() // Reset to full content
		case "enter":
			m.helpSearchActive = false
			// Keep search query and filtered results
		case "backspace":
			if len(m.helpSearchQuery) > 0 {
				m.helpSearchQuery = m.helpSearchQuery[:len(m.helpSearchQuery)-1]
				m.updateHelpView()
			}
		default:
			// Handle common text input operations
			if _, shouldContinue := handleTextInput(&m.helpSearchQuery, msg); shouldContinue {
				m.updateHelpView()
				return nil
			}
			// Append character
			if len(msg.String()) == 1 {
				m.helpSearchQuery += msg.String()
				m.updateHelpView()
			}
		}
		return nil
	}

	// Normal help mode keys
	switch msg.String() {
	case "esc":
		// If there's an active search filter, clear it first
		if m.helpSearchQuery != "" {
			m.helpSearchQuery = ""
			m.updateHelpView() // Reset to full content
		} else {
			m.mode = ModeNormal
		}

	case "?", "q":
		m.mode = ModeNormal
		m.helpSearchQuery = ""
		m.helpSearchActive = false

	case "/":
		m.helpSearchActive = true
		m.helpSearchQuery = ""

	case "up", "k":
		m.helpView.ScrollUp(1)

	case "down", "j":
		m.helpView.ScrollDown(1)

	case "pgup":
		m.helpView.PageUp()

	case "pgdown":
		m.helpView.PageDown()

	case "ctrl+u":
		// Vim-style half-page up
		halfPage := m.helpView.Height / 2
		if halfPage < 1 {
			halfPage = 5
		}
		m.helpView.ScrollUp(halfPage)

	case "ctrl+d":
		// Vim-style half-page down
		halfPage := m.helpView.Height / 2
		if halfPage < 1 {
			halfPage = 5
		}
		m.helpView.ScrollDown(halfPage)

	case "g":
		// Vim-style 'gg' to go to top
		if m.gPressed {
			m.gPressed = false
			m.helpView.GotoTop()
		} else {
			m.gPressed = true
		}
		return nil // Don't reset gPressed

	case "G":
		// Vim-style 'G' to go to bottom
		m.helpView.GotoBottom()

	case "home":
		m.helpView.GotoTop()

	case "end":
		m.helpView.GotoBottom()
	}

	// Reset gPressed on any key except 'g'
	if msg.String() != "g" {
		m.gPressed = false
	}

	return nil
}

// handleDocumentationKeys handles keys in documentation viewer mode
func (m *Model) handleDocumentationKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "m", "q":
		m.mode = ModeNormal
		m.docSelectedIdx = 0 // Reset selection

	// Navigation - update selection and regenerate view (fast when collapsed)
	case "up", "k":
		if m.docSelectedIdx > 0 {
			m.docSelectedIdx--
			m.updateDocumentationView()
		}

	case "down", "j":
		if m.docSelectedIdx < m.docItemCount-1 {
			m.docSelectedIdx++
			m.updateDocumentationView()
		}

	case "home":
		m.docSelectedIdx = 0
		m.updateDocumentationView()

	case "end":
		if m.docItemCount > 0 {
			m.docSelectedIdx = m.docItemCount - 1
		}
		m.updateDocumentationView()

	case "g":
		// Vim-style 'gg' to go to top
		if m.gPressed {
			m.gPressed = false
			m.docSelectedIdx = 0
			m.updateDocumentationView()
		} else {
			m.gPressed = true
		}
		return nil // Don't reset gPressed

	case "G":
		// Vim-style 'G' to go to bottom
		if m.docItemCount > 0 {
			m.docSelectedIdx = m.docItemCount - 1
		}
		m.updateDocumentationView()

	// Page up/down - move cursor by page amount
	case "pgup":
		pageSize := m.modalView.Height
		if pageSize < 1 {
			pageSize = 10
		}
		m.docSelectedIdx -= pageSize
		if m.docSelectedIdx < 0 {
			m.docSelectedIdx = 0
		}
		m.updateDocumentationView()

	case "pgdown":
		pageSize := m.modalView.Height
		if pageSize < 1 {
			pageSize = 10
		}
		m.docSelectedIdx += pageSize
		if m.docSelectedIdx >= m.docItemCount {
			m.docSelectedIdx = m.docItemCount - 1
		}
		if m.docSelectedIdx < 0 {
			m.docSelectedIdx = 0
		}
		m.updateDocumentationView()

	case "ctrl+u":
		// Vim-style half-page up
		halfPage := m.modalView.Height / 2
		if halfPage < 1 {
			halfPage = 5
		}
		m.docSelectedIdx -= halfPage
		if m.docSelectedIdx < 0 {
			m.docSelectedIdx = 0
		}
		m.updateDocumentationView()

	case "ctrl+d":
		// Vim-style half-page down
		halfPage := m.modalView.Height / 2
		if halfPage < 1 {
			halfPage = 5
		}
		m.docSelectedIdx += halfPage
		if m.docSelectedIdx >= m.docItemCount {
			m.docSelectedIdx = m.docItemCount - 1
		}
		if m.docSelectedIdx < 0 {
			m.docSelectedIdx = 0
		}
		m.updateDocumentationView()

	// Toggle collapse/expand - this WILL regenerate (only on toggle, not on scroll)
	case "enter", " ":
		m.toggleDocSection()
	}

	// Reset 'g' state on any key except 'g' itself
	m.gPressed = false
	return nil
}

// countDocItems returns the total number of navigable items in the documentation
func (m *Model) countDocItems() int {
	if m.currentRequest == nil || m.currentRequest.Documentation == nil {
		return 0
	}

	doc := m.currentRequest.Documentation
	count := 0

	// Parameters section header
	if len(doc.Parameters) > 0 {
		count++ // Section header
		if !m.docCollapsed[0] {
			count += len(doc.Parameters) // Each parameter
		}
	}

	// Responses section header
	if len(doc.Responses) > 0 {
		count++ // Section header
		if !m.docCollapsed[1] {
			for respIdx, resp := range doc.Responses {
				count++ // Response line

				// Check if THIS response's fields are expanded
				responseKey := 100 + respIdx
				if !m.docCollapsed[responseKey] && len(resp.Fields) > 0 {
					// Fields are expanded - use cached tree
					allFields, ok := m.docFieldTreeCache[respIdx]
					if !ok {
						// Build and cache the tree
						allFields = buildVirtualFieldTree(resp.Fields)
						m.docFieldTreeCache[respIdx] = allFields
						m.docChildrenCache[respIdx] = buildHasChildrenCache(allFields)
					}
					count += m.countFieldsInTree(respIdx, "", allFields)
				} else if len(resp.Fields) > 0 {
					// Fields are collapsed - count the "N fields" indicator line
					count++
				}
			}
		}
	}

	return count
}

// countFieldsInTree recursively counts fields in the tree
func (m *Model) countFieldsInTree(respIdx int, parentPath string, allFields []DocField) int {
	count := 0
	children := getDirectChildren(parentPath, allFields)

	for _, field := range children {
		count++ // This field

		// Recurse into children if not collapsed
		fieldKey := 200 + respIdx*1000 + hashString(field.Name)
		isCollapsed := m.docCollapsed[fieldKey]
		fieldHasChildren := hasChildren(field.Name, allFields)
		if !isCollapsed && fieldHasChildren {
			count += m.countFieldsInTree(respIdx, field.Name, allFields)
		}
	}

	return count
}

// toggleDocSection toggles the collapsed state of the currently selected documentation section
func (m *Model) toggleDocSection() {
	if m.currentRequest == nil || m.currentRequest.Documentation == nil {
		return
	}

	doc := m.currentRequest.Documentation
	currentIdx := 0

	// Check if we're on the Parameters section header
	if len(doc.Parameters) > 0 {
		if currentIdx == m.docSelectedIdx {
			m.docCollapsed[0] = !m.docCollapsed[0]
			m.updateDocumentationView()
			m.docItemCount = m.countDocItems() // Recalculate after toggle
			return
		}
		currentIdx++
		if !m.docCollapsed[0] {
			currentIdx += len(doc.Parameters)
		}
	}

	// Check if we're on the Responses section header
	if len(doc.Responses) > 0 {
		if currentIdx == m.docSelectedIdx {
			m.docCollapsed[1] = !m.docCollapsed[1]
			m.updateDocumentationView()
			m.docItemCount = m.countDocItems() // Recalculate after toggle
			return
		}
		currentIdx++

		// Check if we're on a response or nested field
		if !m.docCollapsed[1] {
			for respIdx, resp := range doc.Responses {
				// Response line (200:) is NOT toggleable - only the "▶ N fields" line below can toggle
				currentIdx++

				// Only process field toggles if this response's fields are visible
				responseKey := 100 + respIdx
				if !m.docCollapsed[responseKey] && len(resp.Fields) > 0 {
					// Fields are expanded - use cached tree
					allFields, ok := m.docFieldTreeCache[respIdx]
					if !ok {
						// Build and cache the tree
						allFields = buildVirtualFieldTree(resp.Fields)
						m.docFieldTreeCache[respIdx] = allFields
						m.docChildrenCache[respIdx] = buildHasChildrenCache(allFields)
					}
					m.toggleFieldInTree(respIdx, "", allFields, &currentIdx)
				} else if len(resp.Fields) > 0 {
					// Fields are collapsed - check if user is toggling the "▶ N fields" line
					if currentIdx == m.docSelectedIdx {
						// Toggle fields visibility
						m.docCollapsed[responseKey] = !m.docCollapsed[responseKey]

						// Lazy initialization: if expanding for first time, initialize field collapse states
						if !m.docCollapsed[responseKey] {
							m.initializeFieldCollapseState(respIdx, resp.Fields)
						}

						m.updateDocumentationView()
						m.docItemCount = m.countDocItems() // Recalculate after toggle
						return
					}
					currentIdx++
				}
			}
		}
	}
}

// toggleFieldInTree recursively finds and toggles the selected field in the tree
func (m *Model) toggleFieldInTree(respIdx int, parentPath string, allFields []DocField, currentIdx *int) {
	children := getDirectChildren(parentPath, allFields)

	for _, field := range children {
		if *currentIdx == m.docSelectedIdx {
			// This is the selected field - toggle it
			fieldKey := 200 + respIdx*1000 + hashString(field.Name)
			m.docCollapsed[fieldKey] = !m.docCollapsed[fieldKey]
			m.updateDocumentationView()
			m.docItemCount = m.countDocItems() // Recalculate after toggle
			return
		}
		*currentIdx++

		// Show description if not collapsed and not virtual
		fieldKey := 200 + respIdx*1000 + hashString(field.Name)
		isCollapsed := m.docCollapsed[fieldKey]
		if !isCollapsed && !field.IsVirtual && field.Description != "" {
			// Description line doesn't increment selection
		}

		// Recurse into children if not collapsed
		fieldHasChildren := hasChildren(field.Name, allFields)
		if !isCollapsed && fieldHasChildren {
			m.toggleFieldInTree(respIdx, field.Name, allFields, currentIdx)
		}
	}
}

// handleHistoryKeys handles keys in history viewer mode
func (m *Model) handleHistoryKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "H", "q":
		m.mode = ModeNormal

	// Item selection (left pane)
	case "up", "k":
		if m.historyIndex > 0 {
			m.historyIndex--
			m.updateHistoryView() // Regenerate to update highlight and preview
		}

	case "down", "j":
		if m.historyIndex < len(m.historyEntries)-1 {
			m.historyIndex++
			m.updateHistoryView() // Regenerate to update highlight and preview
		}

	// Scroll preview pane (right pane) when preview is visible
	case "shift+up", "K":
		if m.historyPreviewVisible {
			m.historyPreviewView.LineUp(1)
		}

	case "shift+down", "J":
		if m.historyPreviewVisible {
			m.historyPreviewView.LineDown(1)
		}

	// Load selected history entry
	case "enter":
		if len(m.historyEntries) > 0 && m.historyIndex < len(m.historyEntries) {
			return m.loadHistoryEntry(m.historyIndex)
		}

	// Replay selected history entry (re-execute the request)
	case "r":
		if len(m.historyEntries) > 0 && m.historyIndex < len(m.historyEntries) {
			return m.replayHistoryEntry(m.historyIndex)
		}

	// Toggle response preview pane visibility
	case "p":
		m.historyPreviewVisible = !m.historyPreviewVisible
		if m.historyPreviewVisible {
			m.statusMsg = "Preview pane shown"
		} else {
			m.statusMsg = "Preview pane hidden"
		}
		m.updateHistoryView() // Refresh to apply toggle

	// Clear all history with confirmation
	case "C":
		m.mode = ModeHistoryClearConfirm

	// Page up/down - move cursor by page amount
	case "pgup":
		pageSize := m.modalView.Height
		if pageSize < 1 {
			pageSize = 10
		}
		m.historyIndex -= pageSize
		if m.historyIndex < 0 {
			m.historyIndex = 0
		}
		m.updateHistoryView()

	case "pgdown":
		pageSize := m.modalView.Height
		if pageSize < 1 {
			pageSize = 10
		}
		m.historyIndex += pageSize
		if m.historyIndex >= len(m.historyEntries) {
			m.historyIndex = len(m.historyEntries) - 1
		}
		if m.historyIndex < 0 {
			m.historyIndex = 0
		}
		m.updateHistoryView()

	case "ctrl+u":
		// Vim-style half-page up
		halfPage := m.modalView.Height / 2
		if halfPage < 1 {
			halfPage = 5
		}
		m.historyIndex -= halfPage
		if m.historyIndex < 0 {
			m.historyIndex = 0
		}
		m.updateHistoryView()

	case "ctrl+d":
		// Vim-style half-page down
		halfPage := m.modalView.Height / 2
		if halfPage < 1 {
			halfPage = 5
		}
		m.historyIndex += halfPage
		if m.historyIndex >= len(m.historyEntries) {
			m.historyIndex = len(m.historyEntries) - 1
		}
		if m.historyIndex < 0 {
			m.historyIndex = 0
		}
		m.updateHistoryView()

	case "home":
		if len(m.historyEntries) > 0 {
			m.historyIndex = 0
			m.updateHistoryView()
		}

	case "end":
		if len(m.historyEntries) > 0 {
			m.historyIndex = len(m.historyEntries) - 1
			m.updateHistoryView()
		}

	case "g":
		// Vim-style 'gg' to go to top
		if m.gPressed {
			m.gPressed = false
			if len(m.historyEntries) > 0 {
				m.historyIndex = 0
				m.updateHistoryView()
			}
		} else {
			m.gPressed = true
		}
		return nil // Don't reset gPressed

	case "G":
		// Vim-style 'G' to go to bottom
		if len(m.historyEntries) > 0 {
			m.historyIndex = len(m.historyEntries) - 1
			m.updateHistoryView()
		}
	}

	// Reset 'g' state on any key except 'g' itself
	m.gPressed = false
	return nil
}

// handleConfigViewKeys handles keys in config view mode
func (m *Model) handleConfigViewKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "C", "q":
		m.mode = ModeNormal
	}
	return nil
}

// handleDeleteKeys handles keys in delete confirmation mode
func (m *Model) handleDeleteKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "n", "N":
		m.mode = ModeNormal
		m.statusMsg = "Delete cancelled"

	case "y", "Y":
		// Perform delete
		return m.deleteFile()
	}
	return nil
}

// handleConfirmExecutionKeys handles keys in execution confirmation mode
func (m *Model) handleConfirmExecutionKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "n", "N":
		m.mode = ModeNormal
		m.statusMsg = "Execution cancelled"
		m.confirmationGiven = false

	case "y", "Y":
		// Set confirmation flag and execute request
		m.confirmationGiven = true
		m.mode = ModeNormal
		return m.executeRequest()
	}
	return nil
}

// handleHistoryClearConfirmKeys handles keys in history clear confirmation mode
func (m *Model) handleHistoryClearConfirmKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "n", "N":
		m.mode = ModeHistory
		m.statusMsg = "Clear history cancelled"

	case "y", "Y":
		// Clear all history
		if m.historyManager != nil {
			if err := m.historyManager.Clear(); err != nil {
				m.errorMsg = fmt.Sprintf("Failed to clear history: %v", err)
				m.mode = ModeHistory
			} else {
				m.historyEntries = nil
				m.historyIndex = 0
				m.mode = ModeHistory
				m.statusMsg = "All history cleared"
				m.updateHistoryView()
			}
		}
	}
	return nil
}

// handleAnalyticsKeys handles key events in analytics mode
func (m *Model) handleAnalyticsKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "A", "q":
		m.mode = ModeNormal

	// Focus switching
	case "tab":
		// Toggle focus between list and details
		if m.analyticsFocusedPane == "list" {
			m.analyticsFocusedPane = "details"
			m.statusMsg = "Focus: Details panel (use TAB to switch back)"
		} else {
			m.analyticsFocusedPane = "list"
			m.statusMsg = "Focus: List panel (use TAB to switch)"
		}

	// Navigation - based on focused panel
	case "up", "k":
		if m.analyticsFocusedPane == "details" {
			// Scroll details pane if focused
			if m.analyticsPreviewVisible {
				m.analyticsDetailView.LineUp(1)
			}
		} else {
			// Navigate list if focused on list
			if m.analyticsIndex > 0 {
				m.analyticsIndex--
				m.updateAnalyticsView()
			}
		}

	case "down", "j":
		if m.analyticsFocusedPane == "details" {
			// Scroll details pane if focused
			if m.analyticsPreviewVisible {
				m.analyticsDetailView.LineDown(1)
			}
		} else {
			// Navigate list if focused on list
			if m.analyticsIndex < len(m.analyticsStats)-1 {
				m.analyticsIndex++
				m.updateAnalyticsView()
			}
		}

	// Load selected request file
	case "enter":
		if len(m.analyticsStats) > 0 && m.analyticsIndex < len(m.analyticsStats) {
			// Only works in per-file mode (not when grouping by normalized path)
			if m.analyticsGroupByPath {
				m.statusMsg = "Switch to per-file mode (press 't') to load a specific file"
				return nil
			}

			stat := m.analyticsStats[m.analyticsIndex]
			// Load the file associated with this stat
			fileFound := false
			for i, file := range m.files {
				if file.Path == stat.FilePath {
					m.fileIndex = i
					m.mode = ModeNormal
					m.loadRequestsFromCurrentFile()
					fileFound = true
					break
				}
			}

			if !fileFound {
				m.statusMsg = "File not found in current directory"
			}
		}

	// Toggle preview pane visibility
	case "p":
		m.analyticsPreviewVisible = !m.analyticsPreviewVisible
		if m.analyticsPreviewVisible {
			m.statusMsg = "Preview pane shown"
		} else {
			m.statusMsg = "Preview pane hidden"
		}
		m.updateAnalyticsView()

	// Toggle grouping mode
	case "t":
		m.analyticsGroupByPath = !m.analyticsGroupByPath
		m.analyticsIndex = 0
		if m.analyticsGroupByPath {
			m.statusMsg = "Grouping by normalized path"
		} else {
			m.statusMsg = "Grouping by file"
		}
		return m.loadAnalytics()

	// Clear all analytics with confirmation
	case "C":
		m.mode = ModeAnalyticsClearConfirm
		m.statusMsg = "Confirm clear all analytics"

	// Page navigation - based on focused panel
	case "pgup":
		if m.analyticsFocusedPane == "details" {
			// Page up in details pane if focused
			if m.analyticsPreviewVisible {
				m.analyticsDetailView.HalfViewUp()
			}
		} else {
			// Page up in list if focused on list
			pageSize := 10
			m.analyticsIndex -= pageSize
			if m.analyticsIndex < 0 {
				m.analyticsIndex = 0
			}
			m.updateAnalyticsView()
		}

	case "pgdown":
		if m.analyticsFocusedPane == "details" {
			// Page down in details pane if focused
			if m.analyticsPreviewVisible {
				m.analyticsDetailView.HalfViewDown()
			}
		} else {
			// Page down in list if focused on list
			pageSize := 10
			m.analyticsIndex += pageSize
			if m.analyticsIndex >= len(m.analyticsStats) {
				m.analyticsIndex = len(m.analyticsStats) - 1
			}
			if m.analyticsIndex < 0 {
				m.analyticsIndex = 0
			}
			m.updateAnalyticsView()
		}

	case "ctrl+u":
		if m.analyticsFocusedPane == "details" {
			// Half page up in details pane if focused
			if m.analyticsPreviewVisible {
				m.analyticsDetailView.HalfViewUp()
			}
		} else {
			// Half page up in list if focused on list
			halfPage := 5
			m.analyticsIndex -= halfPage
			if m.analyticsIndex < 0 {
				m.analyticsIndex = 0
			}
			m.updateAnalyticsView()
		}

	case "ctrl+d":
		if m.analyticsFocusedPane == "details" {
			// Half page down in details pane if focused
			if m.analyticsPreviewVisible {
				m.analyticsDetailView.HalfViewDown()
			}
		} else {
			// Half page down in list if focused on list
			halfPage := 5
			m.analyticsIndex += halfPage
			if m.analyticsIndex >= len(m.analyticsStats) {
				m.analyticsIndex = len(m.analyticsStats) - 1
			}
			if m.analyticsIndex < 0 {
				m.analyticsIndex = 0
			}
			m.updateAnalyticsView()
		}

	// Vim-style navigation - based on focused panel
	case "g":
		// Vim-style 'gg' to go to top
		if m.gPressed {
			m.gPressed = false
			if m.analyticsFocusedPane == "details" {
				// Go to top of details pane if focused
				if m.analyticsPreviewVisible {
					m.analyticsDetailView.GotoTop()
				}
			} else {
				// Go to top of list if focused on list
				if len(m.analyticsStats) > 0 {
					m.analyticsIndex = 0
					m.updateAnalyticsView()
				}
			}
		} else {
			m.gPressed = true
		}
		return nil // Don't reset gPressed

	case "G":
		// Vim-style 'G' to go to bottom
		if m.analyticsFocusedPane == "details" {
			// Go to bottom of details pane if focused
			if m.analyticsPreviewVisible {
				m.analyticsDetailView.GotoBottom()
			}
		} else {
			// Go to bottom of list if focused on list
			if len(m.analyticsStats) > 0 {
				m.analyticsIndex = len(m.analyticsStats) - 1
				m.updateAnalyticsView()
			}
		}

	case "home":
		if m.analyticsFocusedPane == "details" {
			// Go to top of details pane if focused
			if m.analyticsPreviewVisible {
				m.analyticsDetailView.GotoTop()
			}
		} else {
			// Go to top of list if focused on list
			m.analyticsIndex = 0
			m.updateAnalyticsView()
		}

	case "end":
		if m.analyticsFocusedPane == "details" {
			// Go to bottom of details pane if focused
			if m.analyticsPreviewVisible {
				m.analyticsDetailView.GotoBottom()
			}
		} else {
			// Go to bottom of list if focused on list
			if len(m.analyticsStats) > 0 {
				m.analyticsIndex = len(m.analyticsStats) - 1
				m.updateAnalyticsView()
			}
		}
	}

	// Reset 'g' state on any key except 'g' itself
	m.gPressed = false
	return nil
}

// handleAnalyticsClearConfirmKeys handles keys in analytics clear confirmation mode
func (m *Model) handleAnalyticsClearConfirmKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "n", "N":
		m.mode = ModeAnalytics
		m.statusMsg = "Clear analytics cancelled"

	case "y", "Y":
		// Clear all analytics
		if m.analyticsManager != nil {
			if err := m.analyticsManager.Clear(); err != nil {
				m.errorMsg = fmt.Sprintf("Failed to clear analytics: %v", err)
				m.mode = ModeAnalytics
			} else {
				m.analyticsStats = nil
				m.analyticsIndex = 0
				m.analyticsFocusedPane = "list" // Reset focus to list
				m.analyticsDetailView.SetContent("") // Clear details viewport
				m.mode = ModeAnalytics
				m.statusMsg = "All analytics cleared"
				m.updateAnalyticsView() // Refresh the viewport content
			}
		}
	}
	return nil
}

// handleStressTestConfigKeys handles key events in stress test config mode
func (m *Model) handleStressTestConfigKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		// Cancel modal
		m.mode = ModeNormal
		m.stressTestConfigEdit = nil
		m.stressTestConfigInput = ""
		m.stressTestFilePickerActive = false
		m.statusMsg = "Stress test configuration cancelled"

	case "ctrl+l":
		// Load saved config
		return m.loadStressTestConfigs()

	case "up":
		// If on file field with picker active, ONLY navigate picker (locked)
		if m.stressTestConfigField == 1 && m.stressTestFilePickerActive {
			if m.stressTestFilePickerIndex > 0 {
				m.stressTestFilePickerIndex--
			}
			return nil
		}
		// If on file field but no file selected yet, block navigation
		if m.stressTestConfigField == 1 && m.stressTestConfigInput == "" {
			m.errorMsg = "Please select a file first (press Enter to confirm)"
			return nil
		}
		// Otherwise navigate to previous field
		if err := m.applyStressTestConfigInput(); err != nil {
			m.errorMsg = err.Error()
			return nil
		}
		if m.stressTestConfigField > 0 {
			m.stressTestConfigField--
			m.updateStressTestConfigInput()
			// Activate picker if moving TO field 1
			if m.stressTestConfigField == 1 {
				m.loadStressTestFilePicker()
				m.stressTestFilePickerActive = true
			}
		}

	case "down":
		// If on file field with picker active, ONLY navigate picker (locked)
		if m.stressTestConfigField == 1 && m.stressTestFilePickerActive {
			if m.stressTestFilePickerIndex < len(m.stressTestFilePickerFiles)-1 {
				m.stressTestFilePickerIndex++
			}
			return nil
		}
		// If on file field but no file selected yet, block navigation
		if m.stressTestConfigField == 1 && m.stressTestConfigInput == "" {
			m.errorMsg = "Please select a file first (press Enter to confirm)"
			return nil
		}
		// Otherwise navigate to next field
		if err := m.applyStressTestConfigInput(); err != nil {
			m.errorMsg = err.Error()
			return nil
		}
		if m.stressTestConfigField < 5 {
			m.stressTestConfigField++
			m.updateStressTestConfigInput()
			// Activate picker if moving TO field 1
			if m.stressTestConfigField == 1 {
				m.loadStressTestFilePicker()
				m.stressTestFilePickerActive = true
			}
		}

	case "enter":
		// If on file field with picker, select file and close picker
		if m.stressTestConfigField == 1 && m.stressTestFilePickerActive {
			if len(m.stressTestFilePickerFiles) > 0 && m.stressTestFilePickerIndex < len(m.stressTestFilePickerFiles) {
				selectedFile := m.stressTestFilePickerFiles[m.stressTestFilePickerIndex]
				m.stressTestConfigInput = selectedFile.Path
				m.stressTestConfigCursor = len(m.stressTestConfigInput)
				m.stressTestFilePickerActive = false // Close picker after selection
				if err := m.applyStressTestConfigInput(); err != nil {
					m.errorMsg = err.Error()
				} else {
					m.statusMsg = "File selected - use arrows to navigate to next field"
				}
			} else {
				m.errorMsg = "No files available to select"
			}
			return nil
		}
		// For other fields, just apply
		if err := m.applyStressTestConfigInput(); err != nil {
			m.errorMsg = err.Error()
		}

	case "ctrl+s":
		// Save and start test
		if err := m.applyStressTestConfigInput(); err != nil {
			m.errorMsg = err.Error()
			return nil
		}

		// Validate and save config
		if err := m.stressTestConfigEdit.Validate(); err != nil {
			m.errorMsg = fmt.Sprintf("Invalid configuration: %v", err)
			return nil
		}

		// Save config if it has a name
		if m.stressTestConfigEdit.Name != "" {
			if err := m.stressTestManager.SaveConfig(m.stressTestConfigEdit); err != nil {
				m.errorMsg = fmt.Sprintf("Failed to save config: %v", err)
				return nil
			}
		}

		// Start the stress test
		return m.startStressTest()

	case "backspace":
		// Don't allow editing file field via backspace - use picker
		if m.stressTestConfigField == 1 {
			return nil
		}
		if m.stressTestConfigCursor > 0 {
			input := m.stressTestConfigInput
			m.stressTestConfigInput = input[:m.stressTestConfigCursor-1] + input[m.stressTestConfigCursor:]
			m.stressTestConfigCursor--
		}

	case "delete":
		// Don't allow editing file field via delete - use picker
		if m.stressTestConfigField == 1 {
			return nil
		}
		input := m.stressTestConfigInput
		if m.stressTestConfigCursor < len(input) {
			m.stressTestConfigInput = input[:m.stressTestConfigCursor] + input[m.stressTestConfigCursor+1:]
		}

	case "left":
		// Don't allow cursor movement in file field - use picker
		if m.stressTestConfigField == 1 {
			return nil
		}
		if m.stressTestConfigCursor > 0 {
			m.stressTestConfigCursor--
		}

	case "right":
		// Don't allow cursor movement in file field - use picker
		if m.stressTestConfigField == 1 {
			return nil
		}
		if m.stressTestConfigCursor < len(m.stressTestConfigInput) {
			m.stressTestConfigCursor++
		}

	case "home":
		if m.stressTestConfigField != 1 && m.stressTestConfigCursor > 0 {
			m.stressTestConfigCursor = 0
		}

	case "end":
		if m.stressTestConfigField != 1 {
			m.stressTestConfigCursor = len(m.stressTestConfigInput)
		}

	// Character input for the current field (except file field - use picker)
	default:
		if len(msg.String()) == 1 && m.stressTestConfigField != 1 {
			// Insert character at cursor position
			input := m.stressTestConfigInput
			m.stressTestConfigInput = input[:m.stressTestConfigCursor] + msg.String() + input[m.stressTestConfigCursor:]
			m.stressTestConfigCursor++
		}
	}

	return nil
}

// handleStressTestProgressKeys handles key events in stress test progress mode
func (m *Model) handleStressTestProgressKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "q":
		// Set stopping flag to show feedback
		if !m.stressTestStopping && m.stressTestExecutor != nil {
			m.stressTestStopping = true
			m.statusMsg = "Stopping stress test..."

			// Stop in goroutine to keep UI responsive
			return func() tea.Msg {
				m.stressTestExecutor.Stop()
				return stressTestStoppedMsg{}
			}
		}
	}

	return nil
}

// stressTestStoppedMsg indicates the stress test has finished stopping
type stressTestStoppedMsg struct{}

// handleStressTestLoadConfigKeys handles key events in load config mode
func (m *Model) handleStressTestLoadConfigKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "q":
		m.mode = ModeStressTestConfig
		// Reset picker state when returning to config
		m.stressTestFilePickerActive = false
		m.stressTestFilePickerFiles = nil
		m.stressTestFilePickerIndex = 0
		m.statusMsg = "Load cancelled"

	case "up", "k":
		if m.stressTestConfigIndex > 0 {
			m.stressTestConfigIndex--
		}

	case "down", "j":
		if m.stressTestConfigIndex < len(m.stressTestConfigs)-1 {
			m.stressTestConfigIndex++
		}

	case "enter":
		// Load selected config
		if len(m.stressTestConfigs) > 0 && m.stressTestConfigIndex < len(m.stressTestConfigs) {
			config := m.stressTestConfigs[m.stressTestConfigIndex]
			m.stressTestConfigEdit = config
			m.mode = ModeStressTestConfig
			// Reset picker state when loading config
			m.stressTestFilePickerActive = false
			m.stressTestFilePickerFiles = nil
			m.stressTestFilePickerIndex = 0
			m.stressTestConfigField = 0
			m.updateStressTestConfigInput()
			m.statusMsg = fmt.Sprintf("Loaded config: %s", config.Name)
		}

	case "d":
		// Delete selected config
		if len(m.stressTestConfigs) > 0 && m.stressTestConfigIndex < len(m.stressTestConfigs) {
			config := m.stressTestConfigs[m.stressTestConfigIndex]
			if err := m.stressTestManager.DeleteConfig(config.ID); err != nil {
				m.errorMsg = fmt.Sprintf("Failed to delete config: %v", err)
			} else {
				m.statusMsg = "Configuration deleted"
				return m.loadStressTestConfigs()
			}
		}
	}

	return nil
}

// handleStressTestResultsKeys handles key events in stress test results mode
func (m *Model) handleStressTestResultsKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "S", "q":
		m.mode = ModeNormal

	case "tab":
		// Toggle focus between list and details
		if m.stressTestFocusedPane == "list" {
			m.stressTestFocusedPane = "details"
			m.statusMsg = "Focus: Details panel (use TAB to switch back)"
		} else {
			m.stressTestFocusedPane = "list"
			m.statusMsg = "Focus: List panel (use TAB to switch)"
		}

	case "up", "k":
		if m.stressTestFocusedPane == "details" {
			m.stressTestDetailView.LineUp(1)
		} else if m.stressTestFocusedPane == "list" {
			if m.stressTestRunIndex > 0 {
				m.stressTestRunIndex--
				m.updateStressTestListView()
				m.stressTestDetailView.GotoTop() // Reset scroll when switching runs
			}
		}

	case "down", "j":
		if m.stressTestFocusedPane == "details" {
			m.stressTestDetailView.LineDown(1)
		} else if m.stressTestFocusedPane == "list" {
			if m.stressTestRunIndex < len(m.stressTestRuns)-1 {
				m.stressTestRunIndex++
				m.updateStressTestListView()
				m.stressTestDetailView.GotoTop() // Reset scroll when switching runs
			}
		}

	case "pgup":
		if m.stressTestFocusedPane == "details" {
			m.stressTestDetailView.ViewUp()
		}

	case "pgdown":
		if m.stressTestFocusedPane == "details" {
			m.stressTestDetailView.ViewDown()
		}

	case "g":
		if m.stressTestFocusedPane == "details" {
			m.stressTestDetailView.GotoTop()
		}

	case "G":
		if m.stressTestFocusedPane == "details" {
			m.stressTestDetailView.GotoBottom()
		}

	case "d":
		// Delete selected run
		if len(m.stressTestRuns) > 0 && m.stressTestRunIndex < len(m.stressTestRuns) {
			run := m.stressTestRuns[m.stressTestRunIndex]
			if err := m.stressTestManager.DeleteRun(run.ID); err != nil {
				m.errorMsg = fmt.Sprintf("Failed to delete run: %v", err)
			} else {
				m.statusMsg = "Stress test run deleted"
				return m.loadStressTestRuns()
			}
		}

	case "l":
		// Load saved config
		return m.loadStressTestConfigs()

	case "r":
		// Re-run selected test
		if len(m.stressTestRuns) > 0 && m.stressTestRunIndex < len(m.stressTestRuns) {
			run := m.stressTestRuns[m.stressTestRunIndex]

			// Check if this run has a saved config
			if run.ConfigID == nil {
				m.errorMsg = "Cannot re-run: this test was not saved with a configuration"
				return nil
			}

			// Load the config
			config, err := m.stressTestManager.GetConfig(*run.ConfigID)
			if err != nil {
				m.errorMsg = fmt.Sprintf("Failed to load config: %v", err)
				return nil
			}

			// Load the config into edit mode and start immediately
			m.stressTestConfigEdit = config
			m.statusMsg = fmt.Sprintf("Re-running test: %s", config.Name)
			return m.startStressTest()
		}

	case "n":
		// New stress test
		m.mode = ModeStressTestConfig

		// Initialize new config
		m.stressTestConfigEdit = &stresstest.Config{
			Name:              "",
			RequestFile:       "",
			ConcurrentConns:   10,
			TotalRequests:     100,
			RampUpDurationSec: 0,
			TestDurationSec:   0,
		}

		// Pre-fill with current file if available
		if len(m.files) > 0 && m.fileIndex < len(m.files) {
			m.stressTestConfigEdit.RequestFile = m.files[m.fileIndex].Path
		}

		// Reset picker state
		m.stressTestFilePickerActive = false
		m.stressTestFilePickerFiles = nil
		m.stressTestFilePickerIndex = 0

		m.stressTestConfigField = 0
		m.updateStressTestConfigInput()
	}

	return nil
}
