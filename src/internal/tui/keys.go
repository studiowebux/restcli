package tui

import (
	"fmt"
	"path/filepath"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/studiowebux/restcli/internal/filter"
	"github.com/studiowebux/restcli/internal/keybinds"
	"github.com/studiowebux/restcli/internal/parser"
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
	case ModeBodyOverride:
		return m.handleBodyOverrideKeys(msg)
	case ModeJSONPathHistory:
		return m.handleJSONPathHistoryKeys(msg)
	case ModeTagFilter:
		return m.handleTagFilterKeys(msg)
	case ModeMockServer:
		return m.handleMockServerKeys(msg)
	case ModeProxyViewer:
		return m.handleProxyViewerKeys(msg)
	case ModeProxyDetail:
		return m.handleProxyDetailKeys(msg)
	case ModeWebSocket:
		return m.handleWebSocketKeys(msg)
	}

	return nil
}

// handleShellErrorsKeys handles keyboard input in shell errors modal
func (m *Model) handleShellErrorsKeys(msg tea.KeyMsg) tea.Cmd {
	// Handle enter specially (closes modal)
	if msg.String() == "enter" {
		m.mode = ModeNormal
		m.shellErrors = nil
		return nil
	}

	action, ok := m.keybinds.Match(keybinds.ContextModal, msg.String())
	if !ok {
		return nil
	}

	switch action {
	case keybinds.ActionCloseModal:
		m.mode = ModeNormal
		m.shellErrors = nil

	case keybinds.ActionNavigateDown:
		m.modalView.LineDown(1)

	case keybinds.ActionNavigateUp:
		m.modalView.LineUp(1)

	case keybinds.ActionGoToTop:
		m.modalView.GotoTop()

	case keybinds.ActionGoToBottom:
		m.modalView.GotoBottom()
	}

	return nil
}

// handleErrorDetailKeys handles keyboard input in error detail modal
func (m *Model) handleErrorDetailKeys(msg tea.KeyMsg) tea.Cmd {
	// Handle enter specially (closes modal)
	if msg.String() == "enter" {
		m.mode = ModeNormal
		return nil
	}

	action, ok := m.keybinds.Match(keybinds.ContextModal, msg.String())
	if !ok {
		return nil
	}

	switch action {
	case keybinds.ActionCloseModal:
		m.mode = ModeNormal

	case keybinds.ActionNavigateDown:
		m.modalView.LineDown(1)

	case keybinds.ActionNavigateUp:
		m.modalView.LineUp(1)

	case keybinds.ActionGoToTop:
		m.modalView.GotoTop()

	case keybinds.ActionGoToBottom:
		m.modalView.GotoBottom()
	}

	return nil
}

// handleStatusDetailKeys handles keyboard input in status detail modal
func (m *Model) handleStatusDetailKeys(msg tea.KeyMsg) tea.Cmd {
	// Handle enter specially (closes modal)
	if msg.String() == "enter" {
		m.mode = ModeNormal
		return nil
	}

	action, ok := m.keybinds.Match(keybinds.ContextModal, msg.String())
	if !ok {
		return nil
	}

	switch action {
	case keybinds.ActionCloseModal:
		m.mode = ModeNormal
	}

	return nil
}

// handleEditorConfigKeys handles keyboard input in editor config mode
func (m *Model) handleEditorConfigKeys(msg tea.KeyMsg) tea.Cmd {
	// Check for text input actions
	action, ok := m.keybinds.Match(keybinds.ContextTextInput, msg.String())
	if ok {
		switch action {
		case keybinds.ActionTextCancel:
			m.mode = ModeNormal
			m.inputValue = ""
			m.inputCursor = 0
			return nil

		case keybinds.ActionTextSubmit:
			profile := m.sessionMgr.GetActiveProfile()
			profile.Editor = m.inputValue
			m.sessionMgr.SaveProfiles()
			m.statusMsg = "Editor saved: " + m.inputValue
			m.mode = ModeNormal
			m.inputValue = ""
			m.inputCursor = 0
			return nil
		}
	}

	// Handle text input with cursor support
	if _, shouldContinue := handleTextInputWithCursor(&m.inputValue, &m.inputCursor, msg); shouldContinue {
		return nil
	}

	// Insert character at cursor position
	if len(msg.String()) == 1 {
		m.inputValue = m.inputValue[:m.inputCursor] + msg.String() + m.inputValue[m.inputCursor:]
		m.inputCursor++
	}

	return nil
}

// handleNormalKeys handles keys in normal mode
func (m *Model) handleNormalKeys(msg tea.KeyMsg) tea.Cmd {
	// If filter editing is active, handle filter keys first
	if m.filterEditing {
		return m.handleFilterInlineKeys(msg)
	}

	// Match key to action using keybinds registry
	action, ok, partial := m.keybinds.MatchMultiKey(keybinds.ContextNormal, msg.String())
	if partial {
		// This is a partial match (e.g., first 'g' in 'gg' sequence)
		return nil
	}

	if !ok {
		// No action bound, clear gPressed state and return
		m.gPressed = false
		return nil
	}

	// Handle actions
	switch action {
	case keybinds.ActionQuit:
		m.Cleanup()
		return tea.Quit

	case keybinds.ActionSwitchFocus:
		// Toggle focus between sidebar and response
		if m.focusedPanel == "sidebar" {
			m.focusedPanel = "response"
			m.statusMsg = "Focus: Response panel (use TAB to switch back)"
		} else {
			m.focusedPanel = "sidebar"
			m.statusMsg = "Focus: File sidebar (use TAB to switch)"
			// Reload request from file list when switching to sidebar so Enter executes
			// the file list selection instead of re-executing the loaded history entry
			m.loadRequestsFromCurrentFile()
		}

	case keybinds.ActionNavigateUp:
		if m.focusedPanel == "response" {
			if m.showBody && m.currentResponse != nil {
				m.responseView.ScrollUp(1)
			}
		} else {
			m.navigateFiles(-1)
		}

	case keybinds.ActionNavigateDown:
		if m.focusedPanel == "response" {
			if m.showBody && m.currentResponse != nil {
				m.responseView.ScrollDown(1)
			}
		} else {
			m.navigateFiles(1)
		}

	case keybinds.ActionPageUp:
		if m.focusedPanel == "response" {
			if m.showBody && m.currentResponse != nil {
				m.responseView.PageUp()
			}
		} else {
			m.navigateFiles(-10)
		}

	case keybinds.ActionPageDown:
		if m.focusedPanel == "response" {
			if m.showBody && m.currentResponse != nil {
				m.responseView.PageDown()
			}
		} else {
			m.navigateFiles(10)
		}

	case keybinds.ActionHalfPageUp:
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

	case keybinds.ActionHalfPageDown:
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

	case keybinds.ActionGoToTop:
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

	case keybinds.ActionGoToBottom:
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

	case keybinds.ActionOpenGoto:
		m.mode = ModeGoto
		m.gotoInput = ""

	case keybinds.ActionExecute:
		// Block if request already in progress
		if m.loading {
			m.statusMsg = "Request already in progress"
			return nil
		}

		// Check if current file is a WebSocket file
		if len(m.files) > 0 && m.fileIndex < len(m.files) {
			if m.files[m.fileIndex].HTTPMethod == "WS" {
				m.statusMsg = "Connecting to WebSocket..."
				return m.executeWebSocket()
			}
		}

		m.statusMsg = "Executing request..."
		return m.executeRequest()

	case keybinds.ActionOpenInspect:
		if m.currentRequest == nil {
			m.errorMsg = "No request loaded (select a file first)"
			return nil
		}
		m.mode = ModeInspect
		m.updateInspectView() // Set content once when entering modal

	case keybinds.ActionOpenEditor:
		return m.openInEditor()

	case keybinds.ActionConfigureEditor:
		// Configure editor
		m.mode = ModeEditorConfig
		profile := m.sessionMgr.GetActiveProfile()
		m.inputValue = profile.Editor
		m.inputCursor = len(m.inputValue)

	case keybinds.ActionDuplicateFile:
		if m.focusedPanel == "sidebar" {
			return m.duplicateFile()
		}

	case keybinds.ActionDeleteFile:
		if m.focusedPanel == "sidebar" {
			// Delete file with confirmation
			if len(m.files) > 0 {
				m.mode = ModeDelete
			}
		}

	case keybinds.ActionRenameFile:
		if m.focusedPanel == "sidebar" {
			m.mode = ModeRename
			m.renameInput = ""
			m.renameCursor = 0
		}

	case keybinds.ActionCreateFile:
		if m.focusedPanel == "sidebar" {
			// Create new file
			m.mode = ModeCreateFile
			m.createFileInput = ""
			m.createFileCursor = 0
			m.createFileType = 0 // Default to .http
			m.errorMsg = ""
		}

	case keybinds.ActionRefreshFiles:
		if m.focusedPanel == "sidebar" {
			return m.refreshFiles()
		}

	case keybinds.ActionSaveResponse:
		return m.saveResponse()

	case keybinds.ActionCopyToClipboard:
		return m.copyToClipboard()

	case keybinds.ActionToggleBody:
		m.showBody = !m.showBody

	case keybinds.ActionToggleHeaders:
		m.showHeaders = !m.showHeaders
		m.updateResponseView() // Regenerate response content

	case keybinds.ActionToggleFullscreen:
		m.fullscreen = !m.fullscreen
		m.updateViewport()       // Recalculate viewport width for fullscreen
		m.updateResponseView()   // Regenerate content (wrapping changes based on fullscreen)

	case keybinds.ActionPinResponse:
		// Pin current response for comparison
		if m.currentResponse != nil {
			m.pinnedResponse = m.currentResponse
			m.pinnedRequest = m.currentRequest
			m.statusMsg = "Response pinned for comparison (press W to view diff)"
		} else {
			m.errorMsg = "No response to pin"
		}

	case keybinds.ActionShowDiff:
		// Show diff between pinned and current response
		if m.pinnedResponse == nil {
			m.errorMsg = "No pinned response (press 'w' to pin current response first)"
		} else if m.currentResponse == nil {
			m.errorMsg = "No current response to compare"
		} else {
			m.mode = ModeDiff
			m.updateDiffView()
		}

	case keybinds.ActionFilterResponse:
		// Filter response with JMESPath
		if m.currentResponse != nil && m.currentResponse.Body != "" {
			if m.filterActive {
				// Clear filter if already active
				m.filterActive = false
				m.filteredResponse = ""
				m.filterInput = ""
				m.filterError = ""
				m.updateResponseView()
				m.statusMsg = "Filter cleared"
			} else {
				// Start inline filter editing
				m.filterEditing = true
				m.filterInput = ""
				m.filterCursor = 0
				m.filterError = ""
				m.statusMsg = ""
				m.errorMsg = ""
			}
		} else {
			m.statusMsg = "No response to filter"
		}

	case keybinds.ActionOpenVariables:
		m.mode = ModeVariableList
		m.varEditIndex = 0
		m.modalView.SetYOffset(0)

	case keybinds.ActionOpenHeaders:
		m.mode = ModeHeaderList
		m.headerEditIndex = 0
		m.modalView.SetYOffset(0)

	case keybinds.ActionOpenErrorDetail:
		// Open error detail modal if there's an error
		if m.fullErrorMsg != "" {
			m.mode = ModeErrorDetail
		}

	case keybinds.ActionOpenBodyOverride:
		// Open body override editor
		if m.currentRequest != nil {
			// Initialize with current body resolved
			profile := m.sessionMgr.GetActiveProfile()
			requestCopy := *m.currentRequest
			resolver := parser.NewVariableResolver(profile.Variables, m.sessionMgr.GetSession().Variables, m.interactiveVarValues, parser.LoadSystemEnv())
			resolvedRequest, err := resolver.ResolveRequest(&requestCopy)
			if err == nil && resolvedRequest != nil {
				m.bodyOverrideInput = resolvedRequest.Body
			} else {
				m.bodyOverrideInput = m.currentRequest.Body
			}
			m.bodyOverrideCursor = 0
			m.mode = ModeBodyOverride
			m.statusMsg = "Editing request body (one-time override)"
		} else {
			m.statusMsg = "No request selected"
		}

	case keybinds.ActionShowStatusDetail:
		// Open status detail modal if there's a status message
		if m.fullStatusMsg != "" {
			m.mode = ModeStatusDetail
		}

	case keybinds.ActionOpenProfiles:
		m.mode = ModeProfileSwitch
		m.profileIndex = 0

	case keybinds.ActionSearchNext:
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

	case keybinds.ActionSearchPrevious:
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

	case keybinds.ActionOpenDocumentation:
		m.mode = ModeDocumentation
		// Initialize caches for field trees (prevents rebuilding on every navigation)
		m.docFieldTreeCache = make(map[int][]DocField)
		m.docChildrenCache = make(map[int]map[string]bool)
		m.updateDocumentationView() // Set content and initialize collapse state
		m.docItemCount = m.countDocItems() // Cache item count

	case keybinds.ActionOpenHistory:
		m.mode = ModeHistory
		return m.loadHistory()

	case keybinds.ActionOpenAnalytics:
		m.mode = ModeAnalytics
		m.analyticsPreviewVisible = true
		m.analyticsGroupByPath = false
		return m.loadAnalytics()

	case keybinds.ActionOpenStressTest:
		m.mode = ModeStressTestResults
		m.stressTestFocusedPane = "list"
		return m.loadStressTestRuns()

	case keybinds.ActionOpenMockServer:
		m.mode = ModeMockServer

	case keybinds.ActionOpenProxy:
		// Open proxy viewer (debug proxy)
		m.mode = ModeProxyViewer
		m.updateProxyView()
		// Start event listener if proxy is running
		if m.proxyRunning && m.proxyServer != nil {
			return m.listenForProxyLogs()
		}
		return nil

	case keybinds.ActionOpenTagFilter:
		// Category filter mode
		m.mode = ModeTagFilter
		m.inputValue = ""
		m.inputCursor = 0
		m.statusMsg = "Enter category to filter (press T to clear)"

	case keybinds.ActionClearTagFilter:
		// Clear category filter (Shift+t)
		if len(m.tagFilter) > 0 {
			m.tagFilter = nil
			m.files = m.allFiles
			m.fileIndex = 0
			m.fileOffset = 0
			m.statusMsg = "Category filter cleared"
		}

	case keybinds.ActionOpenHelp:
		m.mode = ModeHelp
		m.updateHelpView()
		return m.checkForUpdate()

	case keybinds.ActionOpenOAuth:
		return m.startOAuthFlow()

	case keybinds.ActionOpenOAuthDetail:
		m.mode = ModeOAuthConfig

	case keybinds.ActionOpenSearch:
		m.mode = ModeSearch
		m.searchQuery = ""
		m.searchMatches = nil
		m.searchIndex = 0

	case keybinds.ActionRefresh:
		// Next search result (ctrl+r alternative)
		if len(m.searchMatches) > 0 {
			m.searchIndex = (m.searchIndex + 1) % len(m.searchMatches)
			m.fileIndex = m.searchMatches[m.searchIndex]
			m.adjustScrollOffset()
			m.loadRequestsFromCurrentFile()
			m.statusMsg = fmt.Sprintf("Match %d of %d", m.searchIndex+1, len(m.searchMatches))
		} else {
			m.statusMsg = "No active search - press / to search"
		}

	case keybinds.ActionOpenRecentFiles:
		// Open MRU (Most Recently Used) files list
		m.mode = ModeMRU
		m.mruIndex = 0
		m.errorMsg = ""

	case keybinds.ActionOpenConfigView:
		m.mode = ModeConfigView

	case keybinds.ActionNoOp:
		// External config editors (mapped to different keys)
		if msg.String() == "P" {
			return m.openProfilesInEditor()
		} else if msg.String() == "ctrl+x" {
			return m.openSessionInEditor()
		}
	}

	// Handle escape separately since it has complex logic
	if msg.String() == "esc" {
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
			wasSearchingResponse := m.searchInResponseCtx
			m.searchMatches = nil
			m.searchQuery = ""
			m.searchIndex = 0
			m.searchInResponseCtx = false
			m.statusMsg = "Search cleared"
			// Clear highlighting from response if we were searching there
			if wasSearchingResponse && m.currentResponse != nil {
				m.updateResponseView()
			}
		} else {
			m.errorMsg = ""
			m.statusMsg = ""
		}
	}

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
	// Check for text input actions
	action, ok := m.keybinds.Match(keybinds.ContextTextInput, msg.String())
	if ok {
		switch action {
		case keybinds.ActionTextCancel:
			wasSearchingResponse := m.searchInResponseCtx
			m.mode = ModeNormal
			m.searchQuery = ""
			m.searchMatches = nil
			m.searchIndex = 0
			m.searchInResponseCtx = false
			if wasSearchingResponse && m.currentResponse != nil {
				m.updateResponseView()
			}
			return nil

		case keybinds.ActionTextSubmit:
			m.mode = ModeNormal
			m.performSearch()
			return nil
		}
	}

	// Ignore ctrl+r (don't append to search)
	if msg.String() == "ctrl+r" {
		return nil
	}

	// Handle common text input operations
	if _, shouldContinue := handleTextInput(&m.searchQuery, msg); shouldContinue {
		return nil
	}

	// Only append single printable characters
	if len(msg.String()) == 1 {
		m.searchQuery += msg.String()
	}

	return nil
}

// handleFilterInlineKeys handles keys when filter editing is active in footer
func (m *Model) handleFilterInlineKeys(msg tea.KeyMsg) tea.Cmd {
	// Handle special keys not in registry
	switch msg.String() {
	case "ctrl+s":
		// Save current expression as bookmark
		if m.filterInput == "" {
			m.filterError = "Cannot save empty expression"
			return nil
		}

		if m.bookmarkManager == nil {
			m.filterError = "Bookmark manager not available"
			return nil
		}

		saved, err := m.bookmarkManager.Save(m.filterInput)
		if err != nil {
			m.filterError = fmt.Sprintf("Failed to save bookmark: %v", err)
			return nil
		}

		// Clear any previous error and show success/duplicate message
		m.filterError = ""
		if saved {
			m.statusMsg = "✓ Bookmark saved successfully"
		} else {
			m.statusMsg = "Bookmark already exists"
		}
		return nil

	case "up":
		// Open history modal if input is empty
		if m.filterInput == "" {
			m.mode = ModeJSONPathHistory
			m.jsonpathHistorySearch = ""
			m.jsonpathHistorySearching = false
			m.jsonpathHistoryCursor = 0
			m.filterEditing = false
			m.loadFilteredBookmarks()
			return nil
		}
		// Otherwise, do nothing (no history navigation in input)
		return nil
	}

	action, ok := m.keybinds.Match(keybinds.ContextTextInput, msg.String())
	if ok {
		switch action {
		case keybinds.ActionTextCancel:
			// Cancel filter editing
			m.filterEditing = false
			m.filterInput = ""
			m.filterCursor = 0
			m.filterError = ""
			m.statusMsg = "Filter cancelled"
			return nil

		case keybinds.ActionTextSubmit:
			// Apply filter
			if m.filterInput == "" {
				m.filterError = "Filter expression cannot be empty"
				return nil
			}

			if m.currentResponse == nil || m.currentResponse.Body == "" {
				m.filterError = "No response to filter"
				m.filterEditing = false
				return nil
			}

			// Apply the filter/query
			result, err := filter.Apply(m.currentResponse.Body, "", m.filterInput)
			if err != nil {
				m.filterError = fmt.Sprintf("Failed to apply filter: %s", err.Error())
				return nil
			}

			// Store filtered result and show it
			m.filteredResponse = result
			m.filterActive = true
			m.filterError = ""
			m.filterEditing = false
			m.statusMsg = fmt.Sprintf("Filter applied: %s", m.filterInput)

			// Update response view to show filtered content
			m.updateResponseView()
			return nil
		}
	}

	// Handle text input with cursor support
	if _, shouldContinue := handleTextInputWithCursor(&m.filterInput, &m.filterCursor, msg); shouldContinue {
		return nil
	}
	// Only append single printable characters
	if len(msg.String()) == 1 {
		m.filterInput = m.filterInput[:m.filterCursor] + msg.String() + m.filterInput[m.filterCursor:]
		m.filterCursor++
	}

	return nil
}

// handleGotoKeys handles keys in goto mode
func (m *Model) handleGotoKeys(msg tea.KeyMsg) tea.Cmd {
	// Check for text input actions
	action, ok := m.keybinds.Match(keybinds.ContextTextInput, msg.String())
	if ok {
		switch action {
		case keybinds.ActionTextCancel:
			m.mode = ModeNormal
			m.gotoInput = ""
			return nil

		case keybinds.ActionTextSubmit:
			m.mode = ModeNormal
			m.performGoto()
			return nil
		}
	}

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

	// Match key to action using keybinds registry
	action, ok, partial := m.keybinds.MatchMultiKey(keybinds.ContextHelp, msg.String())
	if partial {
		// This is a partial match (e.g., first 'g' in 'gg' sequence)
		return nil
	}

	if ok {
		switch action {
		case keybinds.ActionCloseModal, keybinds.ActionCloseModalAlt:
			// If there's an active search filter, clear it first
			if m.helpSearchQuery != "" {
				m.helpSearchQuery = ""
				m.updateHelpView() // Reset to full content
			} else {
				m.mode = ModeNormal
				m.helpSearchQuery = ""
				m.helpSearchActive = false
			}

		case keybinds.ActionOpenSearch:
			m.helpSearchActive = true
			m.helpSearchQuery = ""

		case keybinds.ActionNavigateUp:
			m.helpView.ScrollUp(1)

		case keybinds.ActionNavigateDown:
			m.helpView.ScrollDown(1)

		case keybinds.ActionPageUp:
			m.helpView.PageUp()

		case keybinds.ActionPageDown:
			m.helpView.PageDown()

		case keybinds.ActionHalfPageUp:
			// Vim-style half-page up
			halfPage := m.helpView.Height / 2
			if halfPage < 1 {
				halfPage = 5
			}
			m.helpView.ScrollUp(halfPage)

		case keybinds.ActionHalfPageDown:
			// Vim-style half-page down
			halfPage := m.helpView.Height / 2
			if halfPage < 1 {
				halfPage = 5
			}
			m.helpView.ScrollDown(halfPage)

		case keybinds.ActionGoToTop:
			m.helpView.GotoTop()

		case keybinds.ActionGoToBottom:
			m.helpView.GotoBottom()
		}
	}

	return nil
}

// handleDocumentationKeys handles keys in documentation viewer mode
func (m *Model) handleDocumentationKeys(msg tea.KeyMsg) tea.Cmd {
	action, ok, partial := m.keybinds.MatchMultiKey(keybinds.ContextDocumentation, msg.String())
	if partial {
		return nil
	}

	if !ok {
		m.gPressed = false
		return nil
	}

	switch action {
	case keybinds.ActionCloseModal:
		m.mode = ModeNormal
		m.docSelectedIdx = 0

	case keybinds.ActionNavigateUp:
		if m.docSelectedIdx > 0 {
			m.docSelectedIdx--
			m.updateDocumentationView()
		}

	case keybinds.ActionNavigateDown:
		if m.docSelectedIdx < m.docItemCount-1 {
			m.docSelectedIdx++
			m.updateDocumentationView()
		}

	case keybinds.ActionGoToTop:
		m.docSelectedIdx = 0
		m.updateDocumentationView()

	case keybinds.ActionGoToBottom:
		if m.docItemCount > 0 {
			m.docSelectedIdx = m.docItemCount - 1
		}
		m.updateDocumentationView()

	case keybinds.ActionPageUp:
		pageSize := m.modalView.Height
		if pageSize < 1 {
			pageSize = 10
		}
		m.docSelectedIdx -= pageSize
		if m.docSelectedIdx < 0 {
			m.docSelectedIdx = 0
		}
		m.updateDocumentationView()

	case keybinds.ActionPageDown:
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

	case keybinds.ActionHalfPageUp:
		halfPage := m.modalView.Height / 2
		if halfPage < 1 {
			halfPage = 5
		}
		m.docSelectedIdx -= halfPage
		if m.docSelectedIdx < 0 {
			m.docSelectedIdx = 0
		}
		m.updateDocumentationView()

	case keybinds.ActionHalfPageDown:
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

	case keybinds.ActionTextSubmit:
		m.toggleDocSection()
	}

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
	// If search is active, handle search input first
	if m.historySearchActive {
		switch msg.String() {
		case "esc":
			m.historySearchActive = false
			m.historySearchQuery = ""
			m.historyEntries = m.historyAllEntries
			m.historyIndex = 0
			m.updateHistoryView()
			m.statusMsg = "Search cleared"
			return nil
		case "enter":
			m.historySearchActive = false
			m.statusMsg = fmt.Sprintf("Filtered to %d entries", len(m.historyEntries))
			return nil
		case "backspace":
			if len(m.historySearchQuery) > 0 {
				m.historySearchQuery = m.historySearchQuery[:len(m.historySearchQuery)-1]
				m.filterHistoryEntries()
			}
			return nil
		default:
			if len(msg.String()) == 1 {
				m.historySearchQuery += msg.String()
				m.filterHistoryEntries()
			}
			return nil
		}
	}

	// Handle preview pane scrolling (not in keybinds registry)
	switch msg.String() {
	case "shift+up", "K":
		if m.historyPreviewVisible {
			m.historyPreviewView.LineUp(1)
		}
		return nil
	case "shift+down", "J":
		if m.historyPreviewVisible {
			m.historyPreviewView.LineDown(1)
		}
		return nil
	}

	action, ok, partial := m.keybinds.MatchMultiKey(keybinds.ContextHistory, msg.String())
	if partial {
		return nil
	}

	if !ok {
		m.gPressed = false
		return nil
	}

	switch action {
	case keybinds.ActionCloseModal:
		m.mode = ModeNormal

	case keybinds.ActionOpenSearch:
		m.historySearchActive = true
		m.historySearchQuery = ""
		m.statusMsg = "Search history (ESC to cancel, Enter to apply)"
		return nil

	case keybinds.ActionNavigateUp:
		if m.historyIndex > 0 {
			m.historyIndex--
			m.updateHistoryView()
		}

	case keybinds.ActionNavigateDown:
		if m.historyIndex < len(m.historyEntries)-1 {
			m.historyIndex++
			m.updateHistoryView()
		}

	case keybinds.ActionHistoryExecute:
		if len(m.historyEntries) > 0 && m.historyIndex < len(m.historyEntries) {
			return m.loadHistoryEntry(m.historyIndex)
		}

	case keybinds.ActionHistoryRollback:
		if len(m.historyEntries) > 0 && m.historyIndex < len(m.historyEntries) {
			return m.replayHistoryEntry(m.historyIndex)
		}

	case keybinds.ActionHistoryPaginate:
		m.historyPreviewVisible = !m.historyPreviewVisible
		if m.historyPreviewVisible {
			m.statusMsg = "Preview pane shown"
		} else {
			m.statusMsg = "Preview pane hidden"
		}
		m.updateHistoryView()

	case keybinds.ActionHistoryClear:
		m.mode = ModeHistoryClearConfirm

	case keybinds.ActionPageUp:
		pageSize := m.modalView.Height
		if pageSize < 1 {
			pageSize = 10
		}
		m.historyIndex -= pageSize
		if m.historyIndex < 0 {
			m.historyIndex = 0
		}
		m.updateHistoryView()

	case keybinds.ActionPageDown:
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

	case keybinds.ActionHalfPageUp:
		halfPage := m.modalView.Height / 2
		if halfPage < 1 {
			halfPage = 5
		}
		m.historyIndex -= halfPage
		if m.historyIndex < 0 {
			m.historyIndex = 0
		}
		m.updateHistoryView()

	case keybinds.ActionHalfPageDown:
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

	case keybinds.ActionGoToTop:
		if len(m.historyEntries) > 0 {
			m.historyIndex = 0
			m.updateHistoryView()
		}

	case keybinds.ActionGoToBottom:
		if len(m.historyEntries) > 0 {
			m.historyIndex = len(m.historyEntries) - 1
			m.updateHistoryView()
		}
	}

	m.gPressed = false
	return nil
}

// handleConfigViewKeys handles keys in config view mode
func (m *Model) handleConfigViewKeys(msg tea.KeyMsg) tea.Cmd {
	// Handle 'C' specially (closes modal - same key that opens it)
	if msg.String() == "C" {
		m.mode = ModeNormal
		return nil
	}

	action, ok := m.keybinds.Match(keybinds.ContextModal, msg.String())
	if !ok {
		return nil
	}

	switch action {
	case keybinds.ActionCloseModal:
		m.mode = ModeNormal
	}

	return nil
}

// handleDeleteKeys handles keys in delete confirmation mode
func (m *Model) handleDeleteKeys(msg tea.KeyMsg) tea.Cmd {
	action, ok := m.keybinds.Match(keybinds.ContextConfirm, msg.String())
	if !ok {
		return nil
	}

	switch action {
	case keybinds.ActionCancel:
		m.mode = ModeNormal
		m.statusMsg = "Delete cancelled"

	case keybinds.ActionConfirm:
		return m.deleteFile()
	}

	return nil
}

// handleConfirmExecutionKeys handles keys in execution confirmation mode
func (m *Model) handleConfirmExecutionKeys(msg tea.KeyMsg) tea.Cmd {
	action, ok := m.keybinds.Match(keybinds.ContextConfirm, msg.String())
	if !ok {
		return nil
	}

	switch action {
	case keybinds.ActionCancel:
		m.mode = ModeNormal
		m.statusMsg = "Execution cancelled"
		m.confirmationGiven = false

	case keybinds.ActionConfirm:
		m.confirmationGiven = true
		m.mode = ModeNormal
		return m.executeRequest()
	}

	return nil
}

// handleHistoryClearConfirmKeys handles keys in history clear confirmation mode
func (m *Model) handleHistoryClearConfirmKeys(msg tea.KeyMsg) tea.Cmd {
	action, ok := m.keybinds.Match(keybinds.ContextConfirm, msg.String())
	if !ok {
		return nil
	}

	switch action {
	case keybinds.ActionCancel:
		m.mode = ModeHistory
		m.statusMsg = "Clear history cancelled"

	case keybinds.ActionConfirm:
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
	action, ok, partial := m.keybinds.MatchMultiKey(keybinds.ContextAnalytics, msg.String())
	if partial {
		return nil
	}

	if !ok {
		m.gPressed = false
		return nil
	}

	switch action {
	case keybinds.ActionCloseModal:
		m.mode = ModeNormal

	case keybinds.ActionSwitchPane:
		if m.analyticsFocusedPane == "list" {
			m.analyticsFocusedPane = "details"
			m.statusMsg = "Focus: Details panel (use TAB to switch back)"
		} else {
			m.analyticsFocusedPane = "list"
			m.statusMsg = "Focus: List panel (use TAB to switch)"
		}

	case keybinds.ActionNavigateUp:
		if m.analyticsFocusedPane == "details" {
			if m.analyticsPreviewVisible {
				m.analyticsDetailView.LineUp(1)
			}
		} else {
			if m.analyticsIndex > 0 {
				m.analyticsIndex--
				m.updateAnalyticsView()
			}
		}

	case keybinds.ActionNavigateDown:
		if m.analyticsFocusedPane == "details" {
			if m.analyticsPreviewVisible {
				m.analyticsDetailView.LineDown(1)
			}
		} else {
			if m.analyticsIndex < len(m.analyticsStats)-1 {
				m.analyticsIndex++
				m.updateAnalyticsView()
			}
		}

	case keybinds.ActionTextSubmit:
		if len(m.analyticsStats) > 0 && m.analyticsIndex < len(m.analyticsStats) {
			if m.analyticsGroupByPath {
				m.statusMsg = "Switch to per-file mode (press 't') to load a specific file"
				return nil
			}

			stat := m.analyticsStats[m.analyticsIndex]
			fileFound := false
			for i, file := range m.files {
				if file.Path == stat.FilePath {
					m.fileIndex = i
					m.adjustScrollOffset()
					m.focusedPanel = "sidebar"
					m.mode = ModeNormal
					m.loadRequestsFromCurrentFile()
					m.statusMsg = fmt.Sprintf("Loaded %s from analytics", filepath.Base(file.Path))
					fileFound = true
					break
				}
			}

			if !fileFound {
				m.statusMsg = "File not found in current directory"
			}
		}

	case keybinds.ActionAnalyticsPaginate:
		m.analyticsPreviewVisible = !m.analyticsPreviewVisible
		if m.analyticsPreviewVisible {
			m.statusMsg = "Preview pane shown"
		} else {
			m.statusMsg = "Preview pane hidden"
		}
		m.updateAnalyticsView()

	case keybinds.ActionOpenTagFilter:
		m.analyticsGroupByPath = !m.analyticsGroupByPath
		m.analyticsIndex = 0
		if m.analyticsGroupByPath {
			m.statusMsg = "Grouping by normalized path"
		} else {
			m.statusMsg = "Grouping by file"
		}
		return m.loadAnalytics()

	case keybinds.ActionAnalyticsClear:
		m.mode = ModeAnalyticsClearConfirm
		m.statusMsg = "Confirm clear all analytics"

	case keybinds.ActionPageUp:
		if m.analyticsFocusedPane == "details" {
			if m.analyticsPreviewVisible {
				m.analyticsDetailView.HalfViewUp()
			}
		} else {
			pageSize := 10
			m.analyticsIndex -= pageSize
			if m.analyticsIndex < 0 {
				m.analyticsIndex = 0
			}
			m.updateAnalyticsView()
		}

	case keybinds.ActionPageDown:
		if m.analyticsFocusedPane == "details" {
			if m.analyticsPreviewVisible {
				m.analyticsDetailView.HalfViewDown()
			}
		} else {
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

	case keybinds.ActionHalfPageUp:
		if m.analyticsFocusedPane == "details" {
			if m.analyticsPreviewVisible {
				m.analyticsDetailView.HalfViewUp()
			}
		} else {
			halfPage := 5
			m.analyticsIndex -= halfPage
			if m.analyticsIndex < 0 {
				m.analyticsIndex = 0
			}
			m.updateAnalyticsView()
		}

	case keybinds.ActionHalfPageDown:
		if m.analyticsFocusedPane == "details" {
			if m.analyticsPreviewVisible {
				m.analyticsDetailView.HalfViewDown()
			}
		} else {
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

	case keybinds.ActionGoToTop:
		if m.analyticsFocusedPane == "details" {
			if m.analyticsPreviewVisible {
				m.analyticsDetailView.GotoTop()
			}
		} else {
			if len(m.analyticsStats) > 0 {
				m.analyticsIndex = 0
				m.updateAnalyticsView()
			}
		}

	case keybinds.ActionGoToBottom:
		if m.analyticsFocusedPane == "details" {
			if m.analyticsPreviewVisible {
				m.analyticsDetailView.GotoBottom()
			}
		} else {
			if len(m.analyticsStats) > 0 {
				m.analyticsIndex = len(m.analyticsStats) - 1
				m.updateAnalyticsView()
			}
		}
	}

	m.gPressed = false
	return nil
}

// handleAnalyticsClearConfirmKeys handles keys in analytics clear confirmation mode
func (m *Model) handleAnalyticsClearConfirmKeys(msg tea.KeyMsg) tea.Cmd {
	action, ok := m.keybinds.Match(keybinds.ContextConfirm, msg.String())
	if !ok {
		return nil
	}

	switch action {
	case keybinds.ActionCancel:
		m.mode = ModeAnalytics
		m.statusMsg = "Clear analytics cancelled"

	case keybinds.ActionConfirm:
		if m.analyticsManager != nil {
			if err := m.analyticsManager.Clear(); err != nil {
				m.errorMsg = fmt.Sprintf("Failed to clear analytics: %v", err)
				m.mode = ModeAnalytics
			} else {
				m.analyticsStats = nil
				m.analyticsIndex = 0
				m.analyticsFocusedPane = "list"
				m.analyticsDetailView.SetContent("")
				m.mode = ModeAnalytics
				m.statusMsg = "All analytics cleared"
				m.updateAnalyticsView()
			}
		}
	}

	return nil
}

// handleStressTestConfigKeys handles key events in stress test config mode
func (m *Model) handleStressTestConfigKeys(msg tea.KeyMsg) tea.Cmd {
	// Determine if we're in a text input field (not file picker)
	inTextInput := m.stressTestConfigField != 1

	// Handle special keys first (always active regardless of mode)
	switch msg.String() {
	case "up":
		// Field navigation up
		if m.stressTestConfigField == 1 && m.stressTestFilePickerActive {
			if m.stressTestFilePickerIndex > 0 {
				m.stressTestFilePickerIndex--
			}
			return nil
		}
		if m.stressTestConfigField == 1 && m.stressTestConfigInput == "" {
			m.errorMsg = "Please select a file first (press Enter to confirm)"
			return nil
		}
		if err := m.applyStressTestConfigInput(); err != nil {
			m.errorMsg = err.Error()
			return nil
		}
		if m.stressTestConfigField > 0 {
			m.stressTestConfigField--
			m.updateStressTestConfigInput()
			if m.stressTestConfigField == 1 {
				m.loadStressTestFilePicker()
				m.stressTestFilePickerActive = true
			}
		}
		return nil

	case "down":
		// Field navigation down
		if m.stressTestConfigField == 1 && m.stressTestFilePickerActive {
			if m.stressTestFilePickerIndex < len(m.stressTestFilePickerFiles)-1 {
				m.stressTestFilePickerIndex++
			}
			return nil
		}
		if m.stressTestConfigField == 1 && m.stressTestConfigInput == "" {
			m.errorMsg = "Please select a file first (press Enter to confirm)"
			return nil
		}
		if err := m.applyStressTestConfigInput(); err != nil {
			m.errorMsg = err.Error()
			return nil
		}
		if m.stressTestConfigField < 5 {
			m.stressTestConfigField++
			m.updateStressTestConfigInput()
			if m.stressTestConfigField == 1 {
				m.loadStressTestFilePicker()
				m.stressTestFilePickerActive = true
			}
		}
		return nil

	case "esc", "n", "N":
		// Close modal
		m.mode = ModeNormal
		m.stressTestConfigEdit = nil
		m.stressTestConfigInput = ""
		m.stressTestFilePickerActive = false
		m.statusMsg = "Stress test configuration cancelled"
		return nil

	case "ctrl+l":
		// Load configs
		return m.loadStressTestConfigs()

	case "ctrl+s":
		// Save config and start test
		if err := m.applyStressTestConfigInput(); err != nil {
			m.errorMsg = err.Error()
			return nil
		}
		if err := m.stressTestConfigEdit.Validate(); err != nil {
			m.errorMsg = fmt.Sprintf("Invalid configuration: %v", err)
			return nil
		}
		if m.stressTestConfigEdit.Name != "" {
			if err := m.stressTestManager.SaveConfig(m.stressTestConfigEdit); err != nil {
				m.errorMsg = fmt.Sprintf("Failed to save config: %v", err)
				return nil
			}
		}
		return m.startStressTest()

	case "enter":
		// Submit/select
		if m.stressTestConfigField == 1 && m.stressTestFilePickerActive {
			if len(m.stressTestFilePickerFiles) > 0 && m.stressTestFilePickerIndex < len(m.stressTestFilePickerFiles) {
				selectedFile := m.stressTestFilePickerFiles[m.stressTestFilePickerIndex]
				m.stressTestConfigInput = selectedFile.Path
				m.stressTestConfigCursor = len(m.stressTestConfigInput)
				m.stressTestFilePickerActive = false
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
		if err := m.applyStressTestConfigInput(); err != nil {
			m.errorMsg = err.Error()
		}
		return nil
	}

	// If in text input field, use ContextTextInput for text editing (no ContextStressTest keybind conflicts)
	if inTextInput {
		action, ok := m.keybinds.Match(keybinds.ContextTextInput, msg.String())
		if ok {
			switch action {
			case keybinds.ActionTextBackspace:
				if m.stressTestConfigCursor > 0 {
					input := m.stressTestConfigInput
					m.stressTestConfigInput = input[:m.stressTestConfigCursor-1] + input[m.stressTestConfigCursor:]
					m.stressTestConfigCursor--
				}

			case keybinds.ActionTextDelete:
				input := m.stressTestConfigInput
				if m.stressTestConfigCursor < len(input) {
					m.stressTestConfigInput = input[:m.stressTestConfigCursor] + input[m.stressTestConfigCursor+1:]
				}

			case keybinds.ActionTextMoveLeft:
				if m.stressTestConfigCursor > 0 {
					m.stressTestConfigCursor--
				}

			case keybinds.ActionTextMoveRight:
				if m.stressTestConfigCursor < len(m.stressTestConfigInput) {
					m.stressTestConfigCursor++
				}

			case keybinds.ActionTextMoveHome:
				if m.stressTestConfigCursor > 0 {
					m.stressTestConfigCursor = 0
				}

			case keybinds.ActionTextMoveEnd:
				m.stressTestConfigCursor = len(m.stressTestConfigInput)
			}
			return nil
		}

		// Handle character input
		if len(msg.String()) == 1 {
			input := m.stressTestConfigInput
			m.stressTestConfigInput = input[:m.stressTestConfigCursor] + msg.String() + input[m.stressTestConfigCursor:]
			m.stressTestConfigCursor++
		}
		return nil
	}

	// If not in text input (file picker), handle other ContextStressTest actions
	action, ok := m.keybinds.Match(keybinds.ContextStressTest, msg.String())
	if ok {
		switch action {
		case keybinds.ActionStressTestLoad:
			return m.loadStressTestConfigs()
		}
	}

	return nil
}

// handleStressTestProgressKeys handles key events in stress test progress mode
func (m *Model) handleStressTestProgressKeys(msg tea.KeyMsg) tea.Cmd {
	action, ok := m.keybinds.Match(keybinds.ContextStressTest, msg.String())
	if !ok {
		return nil
	}

	switch action {
	case keybinds.ActionCloseModal:
		if !m.stressTestStopping && m.stressTestExecutor != nil {
			m.stressTestStopping = true
			m.statusMsg = "Stopping stress test..."
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
	action, ok := m.keybinds.Match(keybinds.ContextStressTest, msg.String())
	if !ok {
		return nil
	}

	switch action {
	case keybinds.ActionCloseModal:
		m.mode = ModeStressTestConfig
		m.stressTestFilePickerActive = false
		m.stressTestFilePickerFiles = nil
		m.stressTestFilePickerIndex = 0
		m.statusMsg = "Load cancelled"

	case keybinds.ActionNavigateUp:
		if m.stressTestConfigIndex > 0 {
			m.stressTestConfigIndex--
		}

	case keybinds.ActionNavigateDown:
		if m.stressTestConfigIndex < len(m.stressTestConfigs)-1 {
			m.stressTestConfigIndex++
		}

	case keybinds.ActionTextSubmit:
		if len(m.stressTestConfigs) > 0 && m.stressTestConfigIndex < len(m.stressTestConfigs) {
			config := m.stressTestConfigs[m.stressTestConfigIndex]
			m.stressTestConfigEdit = config
			m.mode = ModeStressTestConfig
			m.stressTestFilePickerActive = false
			m.stressTestFilePickerFiles = nil
			m.stressTestFilePickerIndex = 0
			m.stressTestConfigField = 0
			m.updateStressTestConfigInput()
			m.statusMsg = fmt.Sprintf("Loaded config: %s", config.Name)
		}

	case keybinds.ActionStressTestDelete:
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
	// Handle special keys not in registry
	switch msg.String() {
	case "S":
		m.mode = ModeNormal
		return nil
	case "n":
		m.mode = ModeStressTestConfig
		m.stressTestConfigEdit = &stresstest.Config{
			Name:              "",
			RequestFile:       "",
			ConcurrentConns:   10,
			TotalRequests:     100,
			RampUpDurationSec: 0,
			TestDurationSec:   0,
		}
		if len(m.files) > 0 && m.fileIndex < len(m.files) {
			m.stressTestConfigEdit.RequestFile = m.files[m.fileIndex].Path
		}
		m.stressTestFilePickerActive = false
		m.stressTestFilePickerFiles = nil
		m.stressTestFilePickerIndex = 0
		m.stressTestConfigField = 0
		m.updateStressTestConfigInput()
		return nil
	}

	action, ok := m.keybinds.Match(keybinds.ContextStressTest, msg.String())
	if !ok {
		return nil
	}

	switch action {
	case keybinds.ActionCloseModal:
		m.mode = ModeNormal

	case keybinds.ActionSwitchPane:
		if m.stressTestFocusedPane == "list" {
			m.stressTestFocusedPane = "details"
			m.statusMsg = "Focus: Details panel (use TAB to switch back)"
		} else {
			m.stressTestFocusedPane = "list"
			m.statusMsg = "Focus: List panel (use TAB to switch)"
		}

	case keybinds.ActionNavigateUp:
		if m.stressTestFocusedPane == "details" {
			m.stressTestDetailView.LineUp(1)
		} else if m.stressTestFocusedPane == "list" {
			if m.stressTestRunIndex > 0 {
				m.stressTestRunIndex--
				m.updateStressTestListView()
				m.stressTestDetailView.GotoTop()
			}
		}

	case keybinds.ActionNavigateDown:
		if m.stressTestFocusedPane == "details" {
			m.stressTestDetailView.LineDown(1)
		} else if m.stressTestFocusedPane == "list" {
			if m.stressTestRunIndex < len(m.stressTestRuns)-1 {
				m.stressTestRunIndex++
				m.updateStressTestListView()
				m.stressTestDetailView.GotoTop()
			}
		}

	case keybinds.ActionPageUp:
		if m.stressTestFocusedPane == "details" {
			m.stressTestDetailView.ViewUp()
		}

	case keybinds.ActionPageDown:
		if m.stressTestFocusedPane == "details" {
			m.stressTestDetailView.ViewDown()
		}

	case keybinds.ActionGoToTopPrepare:
		if m.stressTestFocusedPane == "details" {
			m.stressTestDetailView.GotoTop()
		}

	case keybinds.ActionGoToBottom:
		if m.stressTestFocusedPane == "details" {
			m.stressTestDetailView.GotoBottom()
		}

	case keybinds.ActionStressTestDelete:
		if len(m.stressTestRuns) > 0 && m.stressTestRunIndex < len(m.stressTestRuns) {
			run := m.stressTestRuns[m.stressTestRunIndex]
			if err := m.stressTestManager.DeleteRun(run.ID); err != nil {
				m.errorMsg = fmt.Sprintf("Failed to delete run: %v", err)
			} else {
				m.statusMsg = "Stress test run deleted"
				return m.loadStressTestRuns()
			}
		}

	case keybinds.ActionStressTestLoad:
		return m.loadStressTestConfigs()

	case keybinds.ActionRefresh:
		if len(m.stressTestRuns) > 0 && m.stressTestRunIndex < len(m.stressTestRuns) {
			run := m.stressTestRuns[m.stressTestRunIndex]
			if run.ConfigID == nil {
				m.errorMsg = "Cannot re-run: this test was not saved with a configuration"
				return nil
			}
			config, err := m.stressTestManager.GetConfig(*run.ConfigID)
			if err != nil {
				m.errorMsg = fmt.Sprintf("Failed to load config: %v", err)
				return nil
			}
			m.stressTestConfigEdit = config
			m.statusMsg = fmt.Sprintf("Re-running test: %s", config.Name)
			return m.startStressTest()
		}
	}

	return nil
}

// handleTagFilterKeys handles keys in tag filter input mode
func (m *Model) handleTagFilterKeys(msg tea.KeyMsg) tea.Cmd {
	// Check for text input actions
	action, ok := m.keybinds.Match(keybinds.ContextTextInput, msg.String())
	if ok {
		switch action {
		case keybinds.ActionTextCancel:
			m.mode = ModeNormal
			m.tagFilter = nil
			m.files = m.allFiles
			m.fileIndex = 0
			m.fileOffset = 0
			m.statusMsg = "Tag filter cleared"
			m.inputValue = ""
			m.inputCursor = 0
			return nil

		case keybinds.ActionTextSubmit:
			if m.inputValue == "" {
				m.mode = ModeNormal
				m.tagFilter = nil
				m.files = m.allFiles
				m.fileIndex = 0
				m.fileOffset = 0
				m.statusMsg = "Tag filter cleared"
			} else {
				m.tagFilter = []string{m.inputValue}
				m.applyTagFilter()
				m.mode = ModeNormal
				m.statusMsg = fmt.Sprintf("Filtered by category: %s (%d files)", m.inputValue, len(m.files))
			}
			m.inputValue = ""
			m.inputCursor = 0
			return nil
		}
	}

	// Handle text input with cursor support
	if _, shouldContinue := handleTextInputWithCursor(&m.inputValue, &m.inputCursor, msg); shouldContinue {
		return nil
	}

	// Insert character at cursor position
	if len(msg.String()) == 1 {
		m.inputValue = m.inputValue[:m.inputCursor] + msg.String() + m.inputValue[m.inputCursor:]
		m.inputCursor++
	}

	return nil
}
