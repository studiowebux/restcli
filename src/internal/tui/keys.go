package tui

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

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
		if m.streamState.IsActive() {
			m.streamState.Cancel()
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

// handleModalOpenAction handles modal opening actions
// Returns tea.Cmd if an async operation is needed, nil otherwise
func (m *Model) handleModalOpenAction(action keybinds.Action) tea.Cmd {
	switch action {
	case keybinds.ActionOpenVariables:
		m.mode = ModeVariableList
		m.varEditIndex = 0
		m.modalView.SetYOffset(0)
		return nil

	case keybinds.ActionOpenHeaders:
		m.mode = ModeHeaderList
		m.headerEditIndex = 0
		m.modalView.SetYOffset(0)
		return nil

	case keybinds.ActionOpenErrorDetail:
		if m.fullErrorMsg != "" {
			m.mode = ModeErrorDetail
		}
		return nil

	case keybinds.ActionShowStatusDetail:
		if m.fullStatusMsg != "" {
			m.mode = ModeStatusDetail
		}
		return nil

	case keybinds.ActionOpenProfiles:
		m.mode = ModeProfileSwitch
		m.profileIndex = 0
		return nil

	case keybinds.ActionOpenMockServer:
		m.mode = ModeMockServer
		return nil

	case keybinds.ActionOpenOAuthDetail:
		m.mode = ModeOAuthConfig
		return nil

	case keybinds.ActionOpenRecentFiles:
		m.mode = ModeMRU
		m.mruIndex = 0
		m.errorMsg = ""
		return nil

	case keybinds.ActionOpenConfigView:
		m.mode = ModeConfigView
		return nil

	default:
		return nil
	}
}

// handleComplexModalAction handles complex modal opening actions with initialization/async operations
// Returns tea.Cmd if an async operation is needed, nil otherwise
func (m *Model) handleComplexModalAction(action keybinds.Action) tea.Cmd {
	switch action {
	case keybinds.ActionOpenBodyOverride:
		// Open body override editor with variable resolution
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
		return nil

	case keybinds.ActionOpenDocumentation:
		m.mode = ModeDocumentation
		// Initialize caches for field trees (prevents rebuilding on every navigation)
		m.docState.ClearFieldTreeCache()
		m.docState.ClearChildrenCache()
		m.updateDocumentationView()                // Set content and initialize collapse state
		m.docState.SetItemCount(m.countDocItems()) // Cache item count
		return nil

	case keybinds.ActionOpenHistory:
		m.mode = ModeHistory
		m.statusMsg = "Loading history..."
		return m.loadHistory()

	case keybinds.ActionOpenAnalytics:
		m.mode = ModeAnalytics
		m.analyticsState.SetPreviewVisible(true)
		m.analyticsState.SetGroupByPath(false)
		m.statusMsg = "Loading analytics..."
		return m.loadAnalytics()

	case keybinds.ActionOpenStressTest:
		m.mode = ModeStressTestResults
		m.stressTestState.SetFocusedPane("list")
		m.statusMsg = "Loading stress tests..."
		return m.loadStressTestRuns()

	case keybinds.ActionOpenProxy:
		// Open proxy viewer (debug proxy)
		m.mode = ModeProxyViewer
		m.updateProxyView()
		// Start event listener if proxy is running
		if m.proxyServerState.IsRunning() && m.proxyServerState.GetServer() != nil {
			return m.listenForProxyLogs()
		}
		return nil

	case keybinds.ActionOpenHelp:
		m.mode = ModeHelp
		m.updateHelpView()
		return m.checkForUpdate()

	case keybinds.ActionOpenOAuth:
		return m.startOAuthFlow()

	case keybinds.ActionOpenSearch:
		m.mode = ModeSearch
		// Clear search input text for new search
		m.searchInput = ""
		m.searchCursor = 0
		// Clear both file and response search state
		m.fileExplorer.ClearSearch()
		m.responseSearchMatches = nil
		m.responseSearchIndex = 0
		// Clear cached highlighting
		m.cachedHighlightedBody = ""
		m.cachedSearchMatchCount = 0
		// Clear search context flag and update response view to remove highlighting
		if m.searchInResponseCtx {
			m.searchInResponseCtx = false
			if m.currentResponse != nil {
				m.updateResponseView()
			}
		}
		return nil

	default:
		return nil
	}
}

// handleExecuteAction handles request execution (HTTP or WebSocket)
// Returns tea.Cmd for async execution, nil if blocked
func (m *Model) handleExecuteAction() tea.Cmd {
	// Block if request already in progress
	if m.loading {
		m.statusMsg = "Request already in progress"
		return nil
	}

	// Check if current file is a WebSocket file
	currentFile := m.fileExplorer.GetCurrentFile()
	if currentFile != nil && currentFile.HTTPMethod == "WS" {
		m.statusMsg = "Connecting to WebSocket..."
		return m.executeWebSocket()
	}

	m.statusMsg = "Executing request..."
	return m.executeRequest()
}

// handleFocusSwitchAction handles switching focus between sidebar and response panel
func (m *Model) handleFocusSwitchAction() {
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
}

// handleEditorAction handles editor-related actions
// Returns tea.Cmd if an async operation is needed (opening external editor), nil otherwise
func (m *Model) handleEditorAction(action keybinds.Action, msg tea.KeyMsg) tea.Cmd {
	switch action {
	case keybinds.ActionOpenEditor:
		return m.openInEditor()

	case keybinds.ActionConfigureEditor:
		// Configure editor
		m.mode = ModeEditorConfig
		profile := m.sessionMgr.GetActiveProfile()
		m.inputValue = profile.Editor
		m.inputCursor = len(m.inputValue)
		return nil

	case keybinds.ActionNoOp:
		// External config editors (mapped to different keys)
		if msg.String() == "P" {
			return m.openProfilesInEditor()
		} else if msg.String() == "ctrl+x" {
			return m.openSessionInEditor()
		}
		return nil

	default:
		return nil
	}
}

// handleTagFilterAction handles tag/category filtering actions
func (m *Model) handleTagFilterAction(action keybinds.Action) {
	switch action {
	case keybinds.ActionOpenTagFilter:
		// Category filter mode
		m.mode = ModeTagFilter
		m.inputValue = ""
		m.inputCursor = 0
		m.statusMsg = "Enter category to filter (press T to clear)"

	case keybinds.ActionClearTagFilter:
		// Clear category filter (Shift+t)
		if len(m.fileExplorer.GetTagFilter()) > 0 {
			m.fileExplorer.SetTagFilter(nil)
			m.statusMsg = "Category filter cleared"
		}
	}
}

// handleResponseAction handles response-related actions (save, copy, pin, diff, filter)
// Returns tea.Cmd if an async operation is needed, nil otherwise
func (m *Model) handleResponseAction(action keybinds.Action) tea.Cmd {
	switch action {
	case keybinds.ActionSaveResponse:
		return m.saveResponse()

	case keybinds.ActionCopyToClipboard:
		return m.copyToClipboard()

	case keybinds.ActionPinResponse:
		// Pin current response for comparison
		if m.currentResponse == nil {
			return m.setErrorMessage("No response to pin")
		}
		m.pinnedResponse = m.currentResponse
		m.pinnedRequest = m.currentRequest
		m.statusMsg = "Response pinned for comparison (press W to view diff)"
		return nil

	case keybinds.ActionShowDiff:
		// Show diff between pinned and current response
		if m.pinnedResponse == nil {
			return m.setErrorMessage("No pinned response (press 'w' to pin current response first)")
		}
		if m.currentResponse == nil {
			return m.setErrorMessage("No current response to compare")
		}
		m.mode = ModeDiff
		m.updateDiffView()
		return nil

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
		return nil

	default:
		return nil
	}
}

// handleToggleAction handles view toggle actions (body, headers, fullscreen)
func (m *Model) handleToggleAction(action keybinds.Action) {
	switch action {
	case keybinds.ActionToggleBody:
		m.showBody = !m.showBody
		if m.showBody {
			m.statusMsg = "Body shown"
		} else {
			m.statusMsg = "Body hidden"
		}

	case keybinds.ActionToggleHeaders:
		m.showHeaders = !m.showHeaders
		if m.showHeaders {
			m.statusMsg = "Headers shown"
		} else {
			m.statusMsg = "Headers hidden"
		}
		m.updateResponseView() // Regenerate response content

	case keybinds.ActionToggleFullscreen:
		m.fullscreen = !m.fullscreen
		if m.fullscreen {
			m.statusMsg = "Fullscreen enabled"
		} else {
			m.statusMsg = "Fullscreen disabled"
		}
		m.updateViewport()     // Recalculate viewport width for fullscreen
		m.updateResponseView() // Regenerate content (wrapping changes based on fullscreen)
	}
}

// handleFileOperationAction handles file management actions (create, delete, rename, etc.)
// Returns tea.Cmd if an async operation is needed, nil otherwise
func (m *Model) handleFileOperationAction(action keybinds.Action) tea.Cmd {
	// All file operations require sidebar focus
	if m.focusedPanel != "sidebar" {
		return nil
	}

	switch action {
	case keybinds.ActionDuplicateFile:
		return m.duplicateFile()

	case keybinds.ActionDeleteFile:
		if len(m.fileExplorer.GetFiles()) > 0 {
			m.mode = ModeDelete
		}
		return nil

	case keybinds.ActionRenameFile:
		m.mode = ModeRename
		m.renameState.Reset()
		return nil

	case keybinds.ActionCreateFile:
		m.mode = ModeCreateFile
		m.createFileInput = ""
		m.createFileCursor = 0
		m.createFileType = 0 // Default to .http
		m.errorMsg = ""
		return nil

	case keybinds.ActionRefreshFiles:
		m.statusMsg = "Loading files..."
		return m.refreshFiles()

	default:
		return nil
	}
}

// handleSearchNavigationAction handles search result navigation (next/previous match, refresh)
func (m *Model) handleSearchNavigationAction(action keybinds.Action) {
	// ActionRefresh is an alias for next match (ctrl+r shortcut)
	isNext := action == keybinds.ActionSearchNext || action == keybinds.ActionRefresh

	if m.searchInResponseCtx {
		// Navigate in response search results
		if len(m.responseSearchMatches) == 0 {
			return
		}

		if isNext {
			m.responseSearchIndex = (m.responseSearchIndex + 1) % len(m.responseSearchMatches)
		} else {
			m.responseSearchIndex--
			if m.responseSearchIndex < 0 {
				m.responseSearchIndex = len(m.responseSearchMatches) - 1
			}
		}

		m.responseView.SetYOffset(m.centerLineInViewport(m.responseSearchMatches[m.responseSearchIndex]))
		query, _, _ := m.fileExplorer.GetSearchInfo()
		context := "text"
		if isRegexPattern(query) {
			context = "regex"
		}
		m.statusMsg = fmt.Sprintf("[Response] Match %d of %d (%s)", m.responseSearchIndex+1, len(m.responseSearchMatches), context)
	} else {
		// Navigate in file search results
		_, _, totalMatches := m.fileExplorer.GetSearchInfo()
		if totalMatches > 0 {
			pageSize := m.getFileListHeight()
			if isNext {
				m.fileExplorer.NextSearchMatch(pageSize)
			} else {
				m.fileExplorer.PrevSearchMatch(pageSize)
			}
			m.loadRequestsFromCurrentFile()
			query, currentMatch, total := m.fileExplorer.GetSearchInfo()
			context := "text"
			if isRegexPattern(query) {
				context = "regex"
			}
			m.statusMsg = fmt.Sprintf("[Files] Match %d of %d (%s)", currentMatch, total, context)
		} else if action == keybinds.ActionRefresh {
			// ActionRefresh with no active search: show help message
			m.statusMsg = "No active search - press / to search"
		} else if isNext {
			// ActionSearchNext with no active search: create new profile (legacy 'n' behavior)
			m.mode = ModeProfileCreate
			m.profileName = ""
		}
	}
}

// handleNavigationAction handles navigation key actions (up, down, page, goto, etc.)
// Returns true if the action was handled, false otherwise
func (m *Model) handleNavigationAction(action keybinds.Action) bool {
	switch action {
	case keybinds.ActionNavigateUp:
		if m.focusedPanel == "response" {
			if m.showBody && m.currentResponse != nil {
				m.responseView.ScrollUp(1)
			}
		} else {
			m.navigateFiles(-1)
		}
		return true

	case keybinds.ActionNavigateDown:
		if m.focusedPanel == "response" {
			if m.showBody && m.currentResponse != nil {
				m.responseView.ScrollDown(1)
			}
		} else {
			m.navigateFiles(1)
		}
		return true

	case keybinds.ActionPageUp:
		if m.focusedPanel == "response" {
			if m.showBody && m.currentResponse != nil {
				m.responseView.PageUp()
			}
		} else {
			m.navigateFiles(-10)
		}
		return true

	case keybinds.ActionPageDown:
		if m.focusedPanel == "response" {
			if m.showBody && m.currentResponse != nil {
				m.responseView.PageDown()
			}
		} else {
			m.navigateFiles(10)
		}
		return true

	case keybinds.ActionHalfPageUp:
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
		return true

	case keybinds.ActionHalfPageDown:
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
		return true

	case keybinds.ActionGoToTop:
		if m.focusedPanel == "response" {
			if m.showBody && m.currentResponse != nil {
				m.responseView.GotoTop()
			}
		} else {
			pageSize := m.getFileListHeight()
			m.fileExplorer.GoToTop(pageSize)
			m.loadRequestsFromCurrentFile()
		}
		return true

	case keybinds.ActionGoToBottom:
		if m.focusedPanel == "response" {
			if m.showBody && m.currentResponse != nil {
				m.responseView.GotoBottom()
			}
		} else {
			pageSize := m.getFileListHeight()
			m.fileExplorer.GoToBottom(pageSize)
			m.loadRequestsFromCurrentFile()
		}
		return true

	default:
		return false
	}
}

// handleEscapeKey handles ESC key press in normal mode with priority-based cancellation
func (m *Model) handleEscapeKey() tea.Cmd {
	// First priority: Cancel running request
	if m.loading {
		if m.streamState.IsActive() {
			// Cancel streaming request
			m.streamState.Cancel()
		} else {
			// Cancel regular request or chain
			m.requestState.Cancel()
		}
		m.loading = false
		m.statusMsg = "Request cancelled by user"
		m.updateResponseView() // Remove loading indicator
		return nil
	}

	// Second priority: Exit fullscreen mode
	if m.fullscreen {
		m.fullscreen = false
		m.updateViewport()
		m.updateResponseView()
		return nil
	}

	// Third priority: Clear active search
	_, _, fileMatches := m.fileExplorer.GetSearchInfo()
	hasResponseSearch := len(m.responseSearchMatches) > 0

	if fileMatches > 0 || hasResponseSearch {
		// Clear active search
		wasSearchingResponse := m.searchInResponseCtx
		m.fileExplorer.ClearSearch()
		m.responseSearchMatches = nil
		m.responseSearchIndex = 0
		m.searchInResponseCtx = false
		m.statusMsg = "Search cleared"
		// Clear highlighting from response if we were searching there
		if wasSearchingResponse && m.currentResponse != nil {
			m.updateResponseView()
		}
		return nil
	}

	// Default: Clear error and status messages
	m.errorMsg = ""
	m.statusMsg = ""
	return nil
}

// handleNormalKeys handles keys in normal mode
func (m *Model) handleNormalKeys(msg tea.KeyMsg) tea.Cmd {
	// If filter editing is active, handle filter keys first
	if m.filterEditing {
		return m.handleFilterInlineKeys(msg)
	}

	// Handle ESC for cancellation BEFORE keybinds matching
	// This allows ESC to cancel requests even though it's not bound in ContextNormal
	if msg.String() == "esc" {
		return m.handleEscapeKey()
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
		m.handleFocusSwitchAction()

	case keybinds.ActionNavigateUp, keybinds.ActionNavigateDown,
		keybinds.ActionPageUp, keybinds.ActionPageDown,
		keybinds.ActionHalfPageUp, keybinds.ActionHalfPageDown,
		keybinds.ActionGoToTop, keybinds.ActionGoToBottom:
		m.handleNavigationAction(action)

	case keybinds.ActionOpenGoto:
		m.mode = ModeGoto
		m.gotoInput = ""

	case keybinds.ActionExecute:
		return m.handleExecuteAction()

	case keybinds.ActionOpenInspect:
		if m.currentRequest == nil {
			return m.setErrorMessage("No request loaded (select a file first)")
		}
		m.mode = ModeInspect
		m.updateInspectView() // Set content once when entering modal

	case keybinds.ActionOpenEditor, keybinds.ActionConfigureEditor, keybinds.ActionNoOp:
		return m.handleEditorAction(action, msg)

	case keybinds.ActionDuplicateFile, keybinds.ActionDeleteFile,
		keybinds.ActionRenameFile, keybinds.ActionCreateFile,
		keybinds.ActionRefreshFiles:
		return m.handleFileOperationAction(action)

	case keybinds.ActionSaveResponse, keybinds.ActionCopyToClipboard,
		keybinds.ActionPinResponse, keybinds.ActionShowDiff,
		keybinds.ActionFilterResponse:
		return m.handleResponseAction(action)

	case keybinds.ActionToggleBody, keybinds.ActionToggleHeaders, keybinds.ActionToggleFullscreen:
		m.handleToggleAction(action)

	case keybinds.ActionOpenVariables, keybinds.ActionOpenHeaders,
		keybinds.ActionOpenErrorDetail, keybinds.ActionShowStatusDetail,
		keybinds.ActionOpenProfiles, keybinds.ActionOpenMockServer,
		keybinds.ActionOpenOAuthDetail, keybinds.ActionOpenRecentFiles,
		keybinds.ActionOpenConfigView:
		return m.handleModalOpenAction(action)

	case keybinds.ActionOpenBodyOverride, keybinds.ActionOpenDocumentation,
		keybinds.ActionOpenHistory, keybinds.ActionOpenAnalytics,
		keybinds.ActionOpenStressTest, keybinds.ActionOpenProxy,
		keybinds.ActionOpenHelp, keybinds.ActionOpenOAuth,
		keybinds.ActionOpenSearch:
		return m.handleComplexModalAction(action)

	case keybinds.ActionSearchNext, keybinds.ActionSearchPrevious, keybinds.ActionRefresh:
		m.handleSearchNavigationAction(action)

	case keybinds.ActionOpenTagFilter, keybinds.ActionClearTagFilter:
		m.handleTagFilterAction(action)
	}

	return nil
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
			m.searchInput = ""
			m.searchCursor = 0
			m.searchInResponseCtx = false
			// Clear cached highlighting
			m.cachedHighlightedBody = ""
			m.cachedSearchMatchCount = 0
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

	// Handle text input with cursor support
	if _, shouldContinue := handleTextInputWithCursor(&m.searchInput, &m.searchCursor, msg); shouldContinue {
		return nil
	}

	// Insert character at cursor position
	if len(msg.String()) == 1 {
		m.searchInput = m.searchInput[:m.searchCursor] + msg.String() + m.searchInput[m.searchCursor:]
		m.searchCursor++
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

	// Handle text input with cursor support
	if _, shouldContinue := handleTextInputWithCursor(&m.gotoInput, &m.gotoCursor, msg); shouldContinue {
		// For goto, filter to hex characters only after paste/delete
		filtered := ""
		for _, ch := range m.gotoInput {
			if (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F') {
				filtered += string(ch)
			}
		}
		m.gotoInput = filtered
		// Ensure cursor is still valid after filtering
		if m.gotoCursor > len(m.gotoInput) {
			m.gotoCursor = len(m.gotoInput)
		}
		return nil
	}

	// Insert hex character at cursor position (0-9, a-f, A-F)
	if len(msg.String()) == 1 {
		ch := msg.String()[0]
		if (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F') {
			m.gotoInput = m.gotoInput[:m.gotoCursor] + msg.String() + m.gotoInput[m.gotoCursor:]
			m.gotoCursor++
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
		default:
			// Handle text input with cursor support
			if _, shouldContinue := handleTextInputWithCursor(&m.helpSearchQuery, &m.helpSearchCursor, msg); shouldContinue {
				m.updateHelpView()
				return nil
			}
			// Insert character at cursor position
			if len(msg.String()) == 1 {
				m.helpSearchQuery = m.helpSearchQuery[:m.helpSearchCursor] + msg.String() + m.helpSearchQuery[m.helpSearchCursor:]
				m.helpSearchCursor++
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
		m.docState.SetSelectedIdx(0)

	case keybinds.ActionNavigateUp:
		if m.docState.GetSelectedIdx() > 0 {
			m.docState.Navigate(-1, m.docState.GetItemCount())
			m.updateDocumentationView()
		}

	case keybinds.ActionNavigateDown:
		if m.docState.GetSelectedIdx() < m.docState.GetItemCount()-1 {
			m.docState.Navigate(1, m.docState.GetItemCount())
			m.updateDocumentationView()
		}

	case keybinds.ActionGoToTop:
		m.docState.SetSelectedIdx(0)
		m.updateDocumentationView()

	case keybinds.ActionGoToBottom:
		if m.docState.GetItemCount() > 0 {
			m.docState.SetSelectedIdx(m.docState.GetItemCount() - 1)
		}
		m.updateDocumentationView()

	case keybinds.ActionPageUp:
		pageSize := m.modalView.Height
		if pageSize < 1 {
			pageSize = 10
		}
		idx := m.docState.GetSelectedIdx() - pageSize
		m.docState.SetSelectedIdx(idx)
		if m.docState.GetSelectedIdx() < 0 {
			m.docState.SetSelectedIdx(0)
		}
		m.updateDocumentationView()

	case keybinds.ActionPageDown:
		pageSize := m.modalView.Height
		if pageSize < 1 {
			pageSize = 10
		}
		idx := m.docState.GetSelectedIdx() + pageSize
		m.docState.SetSelectedIdx(idx)
		if m.docState.GetSelectedIdx() >= m.docState.GetItemCount() {
			m.docState.SetSelectedIdx(m.docState.GetItemCount() - 1)
		}
		if m.docState.GetSelectedIdx() < 0 {
			m.docState.SetSelectedIdx(0)
		}
		m.updateDocumentationView()

	case keybinds.ActionHalfPageUp:
		halfPage := m.modalView.Height / 2
		if halfPage < 1 {
			halfPage = 5
		}
		idx := m.docState.GetSelectedIdx() - halfPage
		m.docState.SetSelectedIdx(idx)
		if m.docState.GetSelectedIdx() < 0 {
			m.docState.SetSelectedIdx(0)
		}
		m.updateDocumentationView()

	case keybinds.ActionHalfPageDown:
		halfPage := m.modalView.Height / 2
		if halfPage < 1 {
			halfPage = 5
		}
		idx := m.docState.GetSelectedIdx() + halfPage
		m.docState.SetSelectedIdx(idx)
		if m.docState.GetSelectedIdx() >= m.docState.GetItemCount() {
			m.docState.SetSelectedIdx(m.docState.GetItemCount() - 1)
		}
		if m.docState.GetSelectedIdx() < 0 {
			m.docState.SetSelectedIdx(0)
		}
		m.updateDocumentationView()

	case keybinds.ActionTextSubmit:
		m.toggleDocSection()
	}

	m.gPressed = false
	return nil
}

// countDocItems returns the total number of navigable items in the documentation.
//
// Calculates the count based on current collapse state. This is used to:
//   1. Set bounds for navigation (can't navigate beyond itemCount-1)
//   2. Cache the count to avoid recalculating on every navigation
//
// Counting algorithm mirrors rendering order exactly:
//   - If Parameters section exists: count header (1)
//   - If Parameters expanded (key 0): count each parameter
//   - If Responses section exists: count header (1)
//   - If Responses expanded (key 1): for each response:
//     - Count response line (1)
//     - If response fields expanded (key 100+idx): count field tree items recursively
//     - Else if response has fields: count "▶ N fields" indicator line (1)
//
// Uses lazy tree building: only builds field tree if not already cached.
//
// Returns total count of navigable items (lines user can select with ↑↓).
func (m *Model) countDocItems() int {
	if !m.hasValidDocumentation() {
		return 0
	}

	doc := m.currentRequest.Documentation
	count := 0

	// Parameters section header
	if len(doc.Parameters) > 0 {
		count++ // Section header
		if !m.docState.GetCollapsed(getCollapseKeyForSection("parameters")) {
			count += len(doc.Parameters) // Each parameter
		}
	}

	// Responses section header
	if len(doc.Responses) > 0 {
		count++ // Section header
		if !m.docState.GetCollapsed(getCollapseKeyForSection("responses")) {
			for respIdx, resp := range doc.Responses {
				count++ // Response line

				// Check if THIS response's fields are expanded
				responseKey := getCollapseKeyForResponseFields(respIdx)
				if !m.docState.GetCollapsed(responseKey) && len(resp.Fields) > 0 {
					// Fields are expanded - get or build cached tree
					allFields := m.getOrBuildFieldTree(respIdx, resp.Fields)
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

// countFieldsInTree recursively counts visible fields in the tree based on collapse state.
//
// This mirrors the rendering logic - only counts fields that would actually be displayed.
// A field is counted if:
//   1. It is a direct child of parentPath
//   2. Its parent is not collapsed (so it's visible)
//
// For each field:
//   - Count the field itself (1)
//   - If field has children AND is not collapsed: recursively count children
//
// Parameters:
//   - respIdx: Response index (for collapse key generation)
//   - parentPath: Parent field name (empty string for root level)
//   - allFields: Complete virtual field tree
//
// Returns total count of visible navigable fields under this parent.
func (m *Model) countFieldsInTree(respIdx int, parentPath string, allFields []DocField) int {
	count := 0
	children := getDirectChildren(parentPath, allFields)

	for _, field := range children {
		count++ // This field

		// Recurse into children if not collapsed
		fieldKey := getCollapseKeyForField(respIdx, field.Name)
		isCollapsed := m.docState.GetCollapsed(fieldKey)
		fieldHasChildren := hasChildren(field.Name, allFields)
		if !isCollapsed && fieldHasChildren {
			count += m.countFieldsInTree(respIdx, field.Name, allFields)
		}
	}

	return count
}

// toggleDocSection toggles the collapsed state of the currently selected documentation section.
//
// This function walks through the documentation structure in display order, tracking a currentIdx
// counter. When currentIdx matches the selected index, it toggles that item's collapse state.
//
// Algorithm:
//   1. Walk through documentation in rendering order:
//      - Parameters header (idx 0)
//      - Each parameter (if Parameters expanded)
//      - Responses header (idx 1)
//      - Each response and its fields (if Responses expanded)
//   2. When currentIdx matches selectedIdx:
//      - Toggle the appropriate collapse key
//      - If expanding response fields for first time: initializeFieldCollapseState()
//      - Update view and recalculate item count
//      - Return early
//
// Collapse keys used:
//   - 0: Parameters section
//   - 1: Responses section
//   - 100+respIdx: Response fields toggle
//   - 200+respIdx*1000+hash(name): Individual field toggle
//
// Note: currentIdx must be tracked exactly as in rendering/counting to find the right item.
func (m *Model) toggleDocSection() {
	if !m.hasValidDocumentation() {
		return
	}

	doc := m.currentRequest.Documentation
	currentIdx := 0

	// Check if we're on the Parameters section header
	if len(doc.Parameters) > 0 {
		if currentIdx == m.docState.GetSelectedIdx() {
			m.docState.SetCollapsed(getCollapseKeyForSection("parameters"), !m.docState.GetCollapsed(getCollapseKeyForSection("parameters")))
			m.updateDocumentationView()
			m.docState.SetItemCount(m.countDocItems()) // Recalculate after toggle
			return
		}
		currentIdx++
		if !m.docState.GetCollapsed(getCollapseKeyForSection("parameters")) {
			currentIdx += len(doc.Parameters)
		}
	}

	// Check if we're on the Responses section header
	if len(doc.Responses) > 0 {
		if currentIdx == m.docState.GetSelectedIdx() {
			m.docState.SetCollapsed(getCollapseKeyForSection("responses"), !m.docState.GetCollapsed(getCollapseKeyForSection("responses")))
			m.updateDocumentationView()
			m.docState.SetItemCount(m.countDocItems()) // Recalculate after toggle
			return
		}
		currentIdx++

		// Check if we're on a response or nested field
		if !m.docState.GetCollapsed(getCollapseKeyForSection("responses")) {
			for respIdx, resp := range doc.Responses {
				// Response line (200:) is NOT toggleable - only the "▶ N fields" line below can toggle
				currentIdx++

				// Only process field toggles if this response's fields are visible
				responseKey := getCollapseKeyForResponseFields(respIdx)
				if !m.docState.GetCollapsed(responseKey) && len(resp.Fields) > 0 {
					// Fields are expanded - get or build cached tree
					allFields := m.getOrBuildFieldTree(respIdx, resp.Fields)
					m.toggleFieldInTree(respIdx, "", allFields, &currentIdx)
				} else if len(resp.Fields) > 0 {
					// Fields are collapsed - check if user is toggling the "▶ N fields" line
					if currentIdx == m.docState.GetSelectedIdx() {
						// Toggle fields visibility
						m.docState.SetCollapsed(responseKey, !m.docState.GetCollapsed(responseKey))

						// Lazy initialization: if expanding for first time, initialize field collapse states
						if !m.docState.GetCollapsed(responseKey) {
							m.initializeFieldCollapseState(respIdx, resp.Fields)
						}

						m.updateDocumentationView()
						m.docState.SetItemCount(m.countDocItems()) // Recalculate after toggle
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
		if *currentIdx == m.docState.GetSelectedIdx() {
			// This is the selected field - toggle it
			fieldKey := getCollapseKeyForField(respIdx, field.Name)
			m.docState.SetCollapsed(fieldKey, !m.docState.GetCollapsed(fieldKey))
			m.updateDocumentationView()
			m.docState.SetItemCount(m.countDocItems()) // Recalculate after toggle
			return
		}
		*currentIdx++

		// Show description if not collapsed and not virtual
		fieldKey := getCollapseKeyForField(respIdx, field.Name)
		isCollapsed := m.docState.GetCollapsed(fieldKey)
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
	if m.historyState.GetSearchActive() {
		switch msg.String() {
		case "esc":
			m.historyState.SetSearchActive(false)
			m.historyState.SetSearchQuery("")
			m.historyState.SetEntries(m.historyState.GetAllEntries())
			m.historyState.SetIndex(0)
			m.updateHistoryView()
			m.statusMsg = "Search cleared"
			return nil
		case "enter":
			m.historyState.SetSearchActive(false)
			m.statusMsg = fmt.Sprintf("Filtered to %d entries", len(m.historyState.GetEntries()))
			return nil
		case "backspace":
			if len(m.historyState.GetSearchQuery()) > 0 {
				m.historyState.SetSearchQuery(m.historyState.GetSearchQuery()[:len(m.historyState.GetSearchQuery())-1])
				m.filterHistoryEntries()
			}
			return nil
		default:
			if len(msg.String()) == 1 {
				query := m.historyState.GetSearchQuery()
				m.historyState.SetSearchQuery(query + msg.String())
				m.filterHistoryEntries()
			}
			return nil
		}
	}

	// Handle preview pane scrolling (not in keybinds registry)
	switch msg.String() {
	case "shift+up", "K":
		if m.historyState.GetPreviewVisible() {
			previewView := m.historyState.GetPreviewView()
			previewView.LineUp(1)
			m.historyState.SetPreviewView(previewView)
		}
		return nil
	case "shift+down", "J":
		if m.historyState.GetPreviewVisible() {
			previewView := m.historyState.GetPreviewView()
			previewView.LineDown(1)
			m.historyState.SetPreviewView(previewView)
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

	case keybinds.ActionSwitchPane:
		m.historyState.ToggleFocus()
		if m.historyState.GetFocusedPane() == "list" {
			m.statusMsg = "Focus: History list"
		} else {
			m.statusMsg = "Focus: Response preview"
		}
		return nil

	case keybinds.ActionOpenSearch:
		m.historyState.SetSearchActive(true)
		m.historyState.SetSearchQuery("")
		m.statusMsg = "Search history (ESC to cancel, Enter to apply)"
		return nil

	case keybinds.ActionNavigateUp:
		if m.historyState.GetFocusedPane() == "preview" && m.historyState.GetPreviewVisible() {
			previewView := m.historyState.GetPreviewView()
			previewView.LineUp(1)
			m.historyState.SetPreviewView(previewView)
		} else if m.historyState.GetIndex() > 0 {
			m.historyState.Navigate(-1)
			m.updateHistoryView()
		}

	case keybinds.ActionNavigateDown:
		if m.historyState.GetFocusedPane() == "preview" && m.historyState.GetPreviewVisible() {
			previewView := m.historyState.GetPreviewView()
			previewView.LineDown(1)
			m.historyState.SetPreviewView(previewView)
		} else if m.historyState.GetIndex() < len(m.historyState.GetEntries())-1 {
			m.historyState.Navigate(1)
			m.updateHistoryView()
		}

	case keybinds.ActionHistoryExecute:
		if len(m.historyState.GetEntries()) > 0 && m.historyState.GetIndex() < len(m.historyState.GetEntries()) {
			return m.loadHistoryEntry(m.historyState.GetIndex())
		}

	case keybinds.ActionHistoryRollback:
		if len(m.historyState.GetEntries()) > 0 && m.historyState.GetIndex() < len(m.historyState.GetEntries()) {
			return m.replayHistoryEntry(m.historyState.GetIndex())
		}

	case keybinds.ActionHistoryPaginate:
		m.historyState.TogglePreview() // m.historyState.GetPreviewVisible()
		if m.historyState.GetPreviewVisible() {
			m.statusMsg = "Preview pane shown"
		} else {
			m.statusMsg = "Preview pane hidden"
		}
		m.updateHistoryView()

	case keybinds.ActionHistoryClear:
		m.mode = ModeHistoryClearConfirm

	case keybinds.ActionPageUp:
		if m.historyState.GetFocusedPane() == "preview" && m.historyState.GetPreviewVisible() {
			previewView := m.historyState.GetPreviewView()
			previewView.ViewUp()
			m.historyState.SetPreviewView(previewView)
		} else {
			pageSize := m.modalView.Height
			if pageSize < 1 {
				pageSize = 10
			}
			newIndex := m.historyState.GetIndex() - pageSize
			if newIndex < 0 {
				newIndex = 0
			}
			m.historyState.SetIndex(newIndex)
			m.updateHistoryView()
		}

	case keybinds.ActionPageDown:
		if m.historyState.GetFocusedPane() == "preview" && m.historyState.GetPreviewVisible() {
			previewView := m.historyState.GetPreviewView()
			previewView.ViewDown()
			m.historyState.SetPreviewView(previewView)
		} else {
			pageSize := m.modalView.Height
			if pageSize < 1 {
				pageSize = 10
			}
			newIndex := m.historyState.GetIndex() + pageSize
			if newIndex >= len(m.historyState.GetEntries()) {
				newIndex = len(m.historyState.GetEntries()) - 1
			}
			if newIndex < 0 {
				newIndex = 0
			}
			m.historyState.SetIndex(newIndex)
			m.updateHistoryView()
		}

	case keybinds.ActionHalfPageUp:
		if m.historyState.GetFocusedPane() == "preview" && m.historyState.GetPreviewVisible() {
			previewView := m.historyState.GetPreviewView()
			previewView.HalfViewUp()
			m.historyState.SetPreviewView(previewView)
		} else {
			halfPage := m.modalView.Height / 2
			if halfPage < 1 {
				halfPage = 5
			}
			newIndex := m.historyState.GetIndex() - halfPage
			if newIndex < 0 {
				newIndex = 0
			}
			m.historyState.SetIndex(newIndex)
			m.updateHistoryView()
		}

	case keybinds.ActionHalfPageDown:
		if m.historyState.GetFocusedPane() == "preview" && m.historyState.GetPreviewVisible() {
			previewView := m.historyState.GetPreviewView()
			previewView.HalfViewDown()
			m.historyState.SetPreviewView(previewView)
		} else {
			halfPage := m.modalView.Height / 2
			if halfPage < 1 {
				halfPage = 5
			}
			newIndex := m.historyState.GetIndex() + halfPage
			if newIndex >= len(m.historyState.GetEntries()) {
				newIndex = len(m.historyState.GetEntries()) - 1
			}
			if newIndex < 0 {
				newIndex = 0
			}
			m.historyState.SetIndex(newIndex)
			m.updateHistoryView()
		}

	case keybinds.ActionGoToTop:
		if m.historyState.GetFocusedPane() == "preview" && m.historyState.GetPreviewVisible() {
			previewView := m.historyState.GetPreviewView()
			previewView.GotoTop()
			m.historyState.SetPreviewView(previewView)
		} else if len(m.historyState.GetEntries()) > 0 {
			m.historyState.SetIndex(0)
			m.updateHistoryView()
		}

	case keybinds.ActionGoToBottom:
		if m.historyState.GetFocusedPane() == "preview" && m.historyState.GetPreviewVisible() {
			previewView := m.historyState.GetPreviewView()
			previewView.GotoBottom()
			m.historyState.SetPreviewView(previewView)
		} else if len(m.historyState.GetEntries()) > 0 {
			m.historyState.SetIndex(len(m.historyState.GetEntries()) - 1)
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
				m.mode = ModeHistory
				return m.setErrorMessage(fmt.Sprintf("Failed to clear history: %v", err))
			}
			m.historyState.SetEntries(nil)
			m.historyState.SetIndex(0)
			m.mode = ModeHistory
			m.statusMsg = "All history cleared"
			m.updateHistoryView()
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
		if m.analyticsState.GetFocusedPane() == "list" {
			m.analyticsState.SetFocusedPane("details")
			m.statusMsg = "Focus: Details panel (use TAB to switch back)"
		} else {
			m.analyticsState.SetFocusedPane("list")
			m.statusMsg = "Focus: List panel (use TAB to switch)"
		}

	case keybinds.ActionNavigateUp:
		if m.analyticsState.GetFocusedPane() == "details" {
			if m.analyticsState.GetPreviewVisible() {
				detailView := m.analyticsState.GetDetailView()
				detailView.LineUp(1)
				m.analyticsState.SetDetailView(detailView)
			}
		} else {
			if m.analyticsState.GetIndex() > 0 {
				m.analyticsState.Navigate(-1)
				m.updateAnalyticsView()
			}
		}

	case keybinds.ActionNavigateDown:
		if m.analyticsState.GetFocusedPane() == "details" {
			if m.analyticsState.GetPreviewVisible() {
				detailView := m.analyticsState.GetDetailView()
				detailView.LineDown(1)
				m.analyticsState.SetDetailView(detailView)
			}
		} else {
			if m.analyticsState.GetIndex() < len(m.analyticsState.GetStats())-1 {
				m.analyticsState.Navigate(1)
				m.updateAnalyticsView()
			}
		}

	case keybinds.ActionTextSubmit:
		if len(m.analyticsState.GetStats()) > 0 && m.analyticsState.GetIndex() < len(m.analyticsState.GetStats()) {
			if m.analyticsState.GetGroupByPath() {
				m.statusMsg = "Switch to per-file mode (press 't') to load a specific file"
				return nil
			}

			stat := m.analyticsState.GetStats()[m.analyticsState.GetIndex()]

			// Navigate to the file atomically
			pageSize := m.getFileListHeight()
			fileFound := m.fileExplorer.NavigateToFile(stat.FilePath, pageSize)
			if fileFound {
				m.focusedPanel = "sidebar"
				m.mode = ModeNormal
				m.loadRequestsFromCurrentFile()
				m.statusMsg = fmt.Sprintf("Loaded %s from analytics", filepath.Base(stat.FilePath))
			} else {
				m.statusMsg = "File not found in current directory"
			}
		}

	case keybinds.ActionAnalyticsPaginate:
		m.analyticsState.TogglePreview()
		if m.analyticsState.GetPreviewVisible() {
			m.statusMsg = "Preview pane shown"
		} else {
			m.statusMsg = "Preview pane hidden"
		}
		m.updateAnalyticsView()

	case keybinds.ActionOpenTagFilter:
		m.analyticsState.ToggleGroupByPath()
		m.analyticsState.SetIndex(0)
		if m.analyticsState.GetGroupByPath() {
			m.statusMsg = "Grouping by normalized path"
		} else {
			m.statusMsg = "Grouping by file"
		}
		return m.loadAnalytics()

	case keybinds.ActionAnalyticsClear:
		m.mode = ModeAnalyticsClearConfirm
		m.statusMsg = "Confirm clear all analytics"

	case keybinds.ActionPageUp:
		if m.analyticsState.GetFocusedPane() == "details" {
			if m.analyticsState.GetPreviewVisible() {
				detailView := m.analyticsState.GetDetailView()
				detailView.HalfViewUp()
				m.analyticsState.SetDetailView(detailView)
			}
		} else {
			pageSize := 10
			newIndex := m.analyticsState.GetIndex() - pageSize
			if newIndex < 0 {
				newIndex = 0
			}
			m.analyticsState.SetIndex(newIndex)
			m.updateAnalyticsView()
		}

	case keybinds.ActionPageDown:
		if m.analyticsState.GetFocusedPane() == "details" {
			if m.analyticsState.GetPreviewVisible() {
				detailView := m.analyticsState.GetDetailView()
				detailView.HalfViewDown()
				m.analyticsState.SetDetailView(detailView)
			}
		} else {
			pageSize := 10
			newIndex := m.analyticsState.GetIndex() + pageSize
			if newIndex >= len(m.analyticsState.GetStats()) {
				newIndex = len(m.analyticsState.GetStats()) - 1
			}
			if newIndex < 0 {
				newIndex = 0
			}
			m.analyticsState.SetIndex(newIndex)
			m.updateAnalyticsView()
		}

	case keybinds.ActionHalfPageUp:
		if m.analyticsState.GetFocusedPane() == "details" {
			if m.analyticsState.GetPreviewVisible() {
				detailView := m.analyticsState.GetDetailView()
				detailView.HalfViewUp()
				m.analyticsState.SetDetailView(detailView)
			}
		} else {
			halfPage := 5
			newIndex := m.analyticsState.GetIndex() - halfPage
			if newIndex < 0 {
				newIndex = 0
			}
			m.analyticsState.SetIndex(newIndex)
			m.updateAnalyticsView()
		}

	case keybinds.ActionHalfPageDown:
		if m.analyticsState.GetFocusedPane() == "details" {
			if m.analyticsState.GetPreviewVisible() {
				detailView := m.analyticsState.GetDetailView()
				detailView.HalfViewDown()
				m.analyticsState.SetDetailView(detailView)
			}
		} else {
			halfPage := 5
			newIndex := m.analyticsState.GetIndex() + halfPage
			if newIndex >= len(m.analyticsState.GetStats()) {
				newIndex = len(m.analyticsState.GetStats()) - 1
			}
			if newIndex < 0 {
				newIndex = 0
			}
			m.analyticsState.SetIndex(newIndex)
			m.updateAnalyticsView()
		}

	case keybinds.ActionGoToTop:
		if m.analyticsState.GetFocusedPane() == "details" {
			if m.analyticsState.GetPreviewVisible() {
				detailView := m.analyticsState.GetDetailView()
				detailView.GotoTop()
				m.analyticsState.SetDetailView(detailView)
			}
		} else {
			if len(m.analyticsState.GetStats()) > 0 {
				m.analyticsState.SetIndex(0)
				m.updateAnalyticsView()
			}
		}

	case keybinds.ActionGoToBottom:
		if m.analyticsState.GetFocusedPane() == "details" {
			if m.analyticsState.GetPreviewVisible() {
				detailView := m.analyticsState.GetDetailView()
				detailView.GotoBottom()
				m.analyticsState.SetDetailView(detailView)
			}
		} else {
			if len(m.analyticsState.GetStats()) > 0 {
				m.analyticsState.SetIndex(len(m.analyticsState.GetStats()) - 1)
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
				m.mode = ModeAnalytics
				return m.setErrorMessage(fmt.Sprintf("Failed to clear analytics: %v", err))
			}
			m.analyticsState.SetStats(nil)
			m.analyticsState.SetIndex(0)
			m.analyticsState.SetFocusedPane("list")
			detailView := m.analyticsState.GetDetailView()
			detailView.SetContent("")
			m.analyticsState.SetDetailView(detailView)
			m.mode = ModeAnalytics
			m.statusMsg = "All analytics cleared"
			m.updateAnalyticsView()
		}
	}

	return nil
}

// handleStressTestConfigKeys handles key events in stress test config mode
func (m *Model) handleStressTestConfigKeys(msg tea.KeyMsg) tea.Cmd {
	// Determine if we're in a text input field (not file picker)
	inTextInput := m.stressTestState.GetConfigField() != 1

	// CRITICAL: If in text input mode, handle text input FIRST to prevent keybind conflicts
	if inTextInput {
		// Check ContextTextInput for cursor movement/editing
		action, ok := m.keybinds.Match(keybinds.ContextTextInput, msg.String())
		if ok {
			switch action {
			case keybinds.ActionTextBackspace:
				cursor := m.stressTestState.GetConfigCursor()
				if cursor > 0 {
					input := m.stressTestState.GetConfigInput()
					m.stressTestState.SetConfigInput(input[:cursor-1] + input[cursor:])
					m.stressTestState.SetConfigCursor(cursor - 1)
				}
				return nil

			case keybinds.ActionTextDelete:
				input := m.stressTestState.GetConfigInput()
				cursor := m.stressTestState.GetConfigCursor()
				if cursor < len(input) {
					m.stressTestState.SetConfigInput(input[:cursor] + input[cursor+1:])
				}
				return nil

			case keybinds.ActionTextMoveLeft:
				cursor := m.stressTestState.GetConfigCursor()
				if cursor > 0 {
					m.stressTestState.SetConfigCursor(cursor - 1)
				}
				return nil

			case keybinds.ActionTextMoveRight:
				cursor := m.stressTestState.GetConfigCursor()
				if cursor < len(m.stressTestState.GetConfigInput()) {
					m.stressTestState.SetConfigCursor(cursor + 1)
				}
				return nil

			case keybinds.ActionTextMoveHome:
				if m.stressTestState.GetConfigCursor() > 0 {
					m.stressTestState.SetConfigCursor(0)
				}
				return nil

			case keybinds.ActionTextMoveEnd:
				m.stressTestState.SetConfigCursor(len(m.stressTestState.GetConfigInput()))
				return nil
			}
		}

		// Handle character input (printable characters only)
		if len(msg.String()) == 1 {
			input := m.stressTestState.GetConfigInput()
			cursor := m.stressTestState.GetConfigCursor()
			newInput := input[:cursor] + msg.String() + input[cursor:]
			m.stressTestState.SetConfigInput(newInput)
			m.stressTestState.SetConfigCursor(cursor + 1)
			return nil
		}
	}

	// Check keybinds registry (only if NOT in text input mode, or for special actions)
	action, ok := m.keybinds.Match(keybinds.ContextStressTest, msg.String())
	if ok {
		switch action {
		case keybinds.ActionCloseModal:
			// Close modal (works in all modes)
			m.mode = ModeNormal
			m.stressTestState.SetConfigEdit(nil)
			m.stressTestState.SetConfigInput("")
			m.stressTestState.SetFilePickerActive(false)
			m.statusMsg = "Stress test configuration cancelled"
			return nil

		case keybinds.ActionStressTestLoad:
			// Load configs (only if not typing)
			if !inTextInput {
				return m.loadStressTestConfigs()
			}
			return nil

		case keybinds.ActionStressTestSave:
			// Save config and start test (works in all modes)
			if err := m.applyStressTestConfigInput(); err != nil {
				return m.setErrorMessage(err.Error())
			}
			if err := m.stressTestState.GetConfigEdit().Validate(); err != nil {
				return m.setErrorMessage(fmt.Sprintf("Invalid configuration: %v", err))
			}
			if m.stressTestState.GetConfigEdit().Name != "" {
				if err := m.stressTestState.GetManager().SaveConfig(m.stressTestState.GetConfigEdit()); err != nil {
					return m.setErrorMessage(fmt.Sprintf("Failed to save config: %v", err))
				}
			}
			return m.startStressTest()

		case keybinds.ActionTextSubmit:
			// Submit/select (enter key)
			if m.stressTestState.GetConfigField() == 1 && m.stressTestState.GetFilePickerActive() {
				if len(m.stressTestState.GetFilePickerFiles()) > 0 && m.stressTestState.GetFilePickerIndex() < len(m.stressTestState.GetFilePickerFiles()) {
					selectedFile := m.stressTestState.GetFilePickerFiles()[m.stressTestState.GetFilePickerIndex()]
					m.stressTestState.SetConfigInput(selectedFile.Path)
					m.stressTestState.SetConfigCursor(len(selectedFile.Path))
					m.stressTestState.SetFilePickerActive(false)
					if err := m.applyStressTestConfigInput(); err != nil {
						return m.setErrorMessage(err.Error())
					}
					m.statusMsg = "File selected - use arrows to navigate to next field"
				} else {
					return m.setErrorMessage("No files available to select")
				}
			}
			if err := m.applyStressTestConfigInput(); err != nil {
				return m.setErrorMessage(err.Error())
			}
			return nil

		case keybinds.ActionNavigateUp:
			// Field navigation up
			if m.stressTestState.GetConfigField() == 1 && m.stressTestState.GetFilePickerActive() {
				if m.stressTestState.GetFilePickerIndex() > 0 {
					m.stressTestState.NavigateFilePicker(-1)
				}
				return nil
			}
			if m.stressTestState.GetConfigField() == 1 && m.stressTestState.GetConfigInput() == "" {
				return m.setErrorMessage("Please select a file first (press Enter to confirm)")
			}
			if err := m.applyStressTestConfigInput(); err != nil {
				return m.setErrorMessage(err.Error())
			}
			if m.stressTestState.GetConfigField() > 0 {
				m.stressTestState.NavigateConfigFields(-1, 6) // 6 fields total
				m.updateStressTestConfigInput()
				if m.stressTestState.GetConfigField() == 1 {
					m.loadStressTestFilePicker()
					m.stressTestState.SetFilePickerActive(true)
				}
			}
			return nil

		case keybinds.ActionNavigateDown:
			// Field navigation down
			if m.stressTestState.GetConfigField() == 1 && m.stressTestState.GetFilePickerActive() {
				if m.stressTestState.GetFilePickerIndex() < len(m.stressTestState.GetFilePickerFiles())-1 {
					m.stressTestState.NavigateFilePicker(1)
				}
				return nil
			}
			if m.stressTestState.GetConfigField() == 1 && m.stressTestState.GetConfigInput() == "" {
				return m.setErrorMessage("Please select a file first (press Enter to confirm)")
			}
			if err := m.applyStressTestConfigInput(); err != nil {
				return m.setErrorMessage(err.Error())
			}
			if m.stressTestState.GetConfigField() < 5 {
				m.stressTestState.NavigateConfigFields(1, 6) // 6 fields total
				m.updateStressTestConfigInput()
				if m.stressTestState.GetConfigField() == 1 {
					m.loadStressTestFilePicker()
					m.stressTestState.SetFilePickerActive(true)
				}
			}
			return nil
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
		if !m.stressTestState.GetStopping() && m.stressTestState.GetExecutor() != nil {
			// Stress test is running - stop it first
			m.stressTestState.SetStopping(true)
			m.statusMsg = "Stopping stress test..."
			return func() tea.Msg {
				// Use context with 5-second timeout for proper resource cleanup
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				err := m.stressTestState.GetExecutor().StopWithContext(ctx)
				return stressTestStoppedMsg{err: err}
			}
		} else {
			// No active stress test or already stopping - just close modal
			m.mode = ModeNormal
			m.stressTestState.SetStopping(false)
		}
	}

	return nil
}

// stressTestStoppedMsg indicates the stress test has finished stopping
type stressTestStoppedMsg struct {
	err error // Optional error if stop timed out
}

// handleStressTestLoadConfigKeys handles key events in load config mode
func (m *Model) handleStressTestLoadConfigKeys(msg tea.KeyMsg) tea.Cmd {
	action, ok := m.keybinds.Match(keybinds.ContextStressTest, msg.String())
	if !ok {
		return nil
	}

	switch action {
	case keybinds.ActionCloseModal:
		m.mode = ModeStressTestConfig
		m.stressTestState.SetFilePickerActive(false)
		m.stressTestState.SetFilePickerFiles(nil)
		m.stressTestState.SetFilePickerIndex(0)
		m.statusMsg = "Load cancelled"

	case keybinds.ActionNavigateUp:
		if m.stressTestState.GetConfigIndex() > 0 {
			m.stressTestState.Navigate(-1)
		}

	case keybinds.ActionNavigateDown:
		if m.stressTestState.GetConfigIndex() < len(m.stressTestState.GetConfigs())-1 {
			m.stressTestState.Navigate(1)
		}

	case keybinds.ActionTextSubmit:
		if len(m.stressTestState.GetConfigs()) > 0 && m.stressTestState.GetConfigIndex() < len(m.stressTestState.GetConfigs()) {
			config := m.stressTestState.GetConfigs()[m.stressTestState.GetConfigIndex()]
			m.stressTestState.SetConfigEdit(config)
			m.mode = ModeStressTestConfig
			m.stressTestState.SetFilePickerActive(false)
			m.stressTestState.SetFilePickerFiles(nil)
			m.stressTestState.SetFilePickerIndex(0)
			m.stressTestState.SetConfigField(0)
			m.updateStressTestConfigInput()
			m.statusMsg = fmt.Sprintf("Loaded config: %s", config.Name)
		}

	case keybinds.ActionStressTestDelete:
		if len(m.stressTestState.GetConfigs()) > 0 && m.stressTestState.GetConfigIndex() < len(m.stressTestState.GetConfigs()) {
			config := m.stressTestState.GetConfigs()[m.stressTestState.GetConfigIndex()]
			if err := m.stressTestState.GetManager().DeleteConfig(config.ID); err != nil {
				return m.setErrorMessage(fmt.Sprintf("Failed to delete config: %v", err))
			}
			m.statusMsg = "Configuration deleted"
			return m.loadStressTestConfigs()
		}
	}

	return nil
}

// handleStressTestResultsKeys handles key events in stress test results mode
func (m *Model) handleStressTestResultsKeys(msg tea.KeyMsg) tea.Cmd {
	// Handle "n" key for new stress test config (special action)
	if msg.String() == "n" {
		m.mode = ModeStressTestConfig
		m.stressTestState.SetConfigEdit(&stresstest.Config{
			Name:              "",
			RequestFile:       "",
			ConcurrentConns:   10,
			TotalRequests:     100,
			RampUpDurationSec: 0,
			TestDurationSec:   0,
		})
		currentFile := m.fileExplorer.GetCurrentFile()
		if currentFile != nil {
			m.stressTestState.GetConfigEdit().RequestFile = currentFile.Path
		}
		m.stressTestState.SetFilePickerActive(false)
		m.stressTestState.SetFilePickerFiles(nil)
		m.stressTestState.SetFilePickerIndex(0)
		m.stressTestState.SetConfigField(0)
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
		if m.stressTestState.GetFocusedPane() == "list" {
			m.stressTestState.SetFocusedPane("details")
			m.statusMsg = "Focus: Details panel (use TAB to switch back)"
		} else {
			m.stressTestState.SetFocusedPane("list")
			m.statusMsg = "Focus: List panel (use TAB to switch)"
		}

	case keybinds.ActionNavigateUp:
		if m.stressTestState.GetFocusedPane() == "details" {
			detailView := m.stressTestState.GetDetailView()
			detailView.LineUp(1)
			m.stressTestState.SetDetailView(detailView)
		} else if m.stressTestState.GetFocusedPane() == "list" {
			if m.stressTestState.GetRunIndex() > 0 {
				m.stressTestState.NavigateRuns(-1)
				m.updateStressTestListView()
				detailView := m.stressTestState.GetDetailView()
				detailView.GotoTop()
				m.stressTestState.SetDetailView(detailView)
			}
		}

	case keybinds.ActionNavigateDown:
		if m.stressTestState.GetFocusedPane() == "details" {
			detailView := m.stressTestState.GetDetailView()
			detailView.LineDown(1)
			m.stressTestState.SetDetailView(detailView)
		} else if m.stressTestState.GetFocusedPane() == "list" {
			if m.stressTestState.GetRunIndex() < len(m.stressTestState.GetRuns())-1 {
				m.stressTestState.NavigateRuns(1)
				m.updateStressTestListView()
				detailView := m.stressTestState.GetDetailView()
				detailView.GotoTop()
				m.stressTestState.SetDetailView(detailView)
			}
		}

	case keybinds.ActionPageUp:
		if m.stressTestState.GetFocusedPane() == "details" {
			detailView := m.stressTestState.GetDetailView()
			detailView.ViewUp()
			m.stressTestState.SetDetailView(detailView)
		}

	case keybinds.ActionPageDown:
		if m.stressTestState.GetFocusedPane() == "details" {
			detailView := m.stressTestState.GetDetailView()
			detailView.ViewDown()
			m.stressTestState.SetDetailView(detailView)
		}

	case keybinds.ActionGoToTopPrepare:
		if m.stressTestState.GetFocusedPane() == "details" {
			detailView := m.stressTestState.GetDetailView()
			detailView.GotoTop()
			m.stressTestState.SetDetailView(detailView)
		}

	case keybinds.ActionGoToBottom:
		if m.stressTestState.GetFocusedPane() == "details" {
			detailView := m.stressTestState.GetDetailView()
			detailView.GotoBottom()
			m.stressTestState.SetDetailView(detailView)
		}

	case keybinds.ActionStressTestDelete:
		if len(m.stressTestState.GetRuns()) > 0 && m.stressTestState.GetRunIndex() < len(m.stressTestState.GetRuns()) {
			run := m.stressTestState.GetRuns()[m.stressTestState.GetRunIndex()]
			if err := m.stressTestState.GetManager().DeleteRun(run.ID); err != nil {
				return m.setErrorMessage(fmt.Sprintf("Failed to delete run: %v", err))
			}
			m.statusMsg = "Stress test run deleted"
			return m.loadStressTestRuns()
		}

	case keybinds.ActionStressTestLoad:
		return m.loadStressTestConfigs()

	case keybinds.ActionRefresh:
		if len(m.stressTestState.GetRuns()) > 0 && m.stressTestState.GetRunIndex() < len(m.stressTestState.GetRuns()) {
			run := m.stressTestState.GetRuns()[m.stressTestState.GetRunIndex()]
			if run.ConfigID == nil {
				return m.setErrorMessage("Cannot re-run: this test was not saved with a configuration")
			}
			config, err := m.stressTestState.GetManager().GetConfig(*run.ConfigID)
			if err != nil {
				return m.setErrorMessage(fmt.Sprintf("Failed to load config: %v", err))
			}
			m.stressTestState.SetConfigEdit(config)
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
			m.fileExplorer.SetTagFilter(nil)
			m.statusMsg = "Tag filter cleared"
			m.inputValue = ""
			m.inputCursor = 0
			return nil

		case keybinds.ActionTextSubmit:
			if m.inputValue == "" {
				m.mode = ModeNormal
				m.fileExplorer.SetTagFilter(nil)
				m.statusMsg = "Tag filter cleared"
			} else {
				m.fileExplorer.SetTagFilter([]string{m.inputValue})
				m.loadRequestsFromCurrentFile()
				m.mode = ModeNormal
				files := m.fileExplorer.GetFiles()
				m.statusMsg = fmt.Sprintf("Filtered by category: %s (%d files)", m.inputValue, len(files))
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
