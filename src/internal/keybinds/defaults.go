package keybinds

// NewDefaultRegistry creates a registry with all default keybindings
func NewDefaultRegistry() *Registry {
	r := NewRegistry()

	// Register all default keybindings
	registerGlobalBindings(r)
	registerNavigationBindings(r)
	registerTextInputBindings(r)
	registerNormalModeBindings(r)
	registerSearchBindings(r)
	registerGotoBindings(r)
	registerVariableBindings(r)
	registerHeaderBindings(r)
	registerProfileBindings(r)
	registerDocumentationBindings(r)
	registerHistoryBindings(r)
	registerAnalyticsBindings(r)
	registerStressTestBindings(r)
	registerHelpBindings(r)
	registerInspectBindings(r)
	registerWebSocketBindings(r)
	registerModalBindings(r)
	registerConfirmBindings(r)
	registerViewerBindings(r)

	return r
}

// registerGlobalBindings sets up bindings available in all modes
func registerGlobalBindings(r *Registry) {
	r.Register(ContextGlobal, "ctrl+c", ActionQuitForce)
}

// registerNavigationBindings sets up common navigation bindings for viewers
func registerNavigationBindings(r *Registry) {
	// These are registered in ContextViewer and will be inherited by specific viewer contexts
	r.RegisterMultiple(ContextViewer, []string{"up", "k"}, ActionNavigateUp)
	r.RegisterMultiple(ContextViewer, []string{"down", "j"}, ActionNavigateDown)
	r.Register(ContextViewer, "pgup", ActionPageUp)
	r.Register(ContextViewer, "pgdown", ActionPageDown)
	r.Register(ContextViewer, "ctrl+u", ActionHalfPageUp)
	r.Register(ContextViewer, "ctrl+d", ActionHalfPageDown)
	r.Register(ContextViewer, "g", ActionGoToTopPrepare)
	r.Register(ContextViewer, "gg", ActionGoToTop)
	r.Register(ContextViewer, "G", ActionGoToBottom)
	r.Register(ContextViewer, "home", ActionGoToTop)
	r.Register(ContextViewer, "end", ActionGoToBottom)
}

// registerTextInputBindings sets up common text input bindings
func registerTextInputBindings(r *Registry) {
	r.Register(ContextTextInput, "backspace", ActionTextBackspace)
	r.Register(ContextTextInput, "delete", ActionTextDelete)
	r.Register(ContextTextInput, "left", ActionTextMoveLeft)
	r.Register(ContextTextInput, "right", ActionTextMoveRight)
	r.RegisterMultiple(ContextTextInput, []string{"home", "ctrl+a"}, ActionTextMoveHome)
	r.RegisterMultiple(ContextTextInput, []string{"end", "ctrl+e"}, ActionTextMoveEnd)
	r.RegisterMultiple(ContextTextInput, []string{"ctrl+v", "shift+insert", "super+v"}, ActionTextPaste)
	r.Register(ContextTextInput, "ctrl+y", ActionTextDeleteWord)
	r.Register(ContextTextInput, "ctrl+u", ActionTextClearBefore)
	r.Register(ContextTextInput, "ctrl+k", ActionTextClearAfter)
	r.Register(ContextTextInput, "enter", ActionTextSubmit)
	r.Register(ContextTextInput, "esc", ActionTextCancel)
}

// registerNormalModeBindings sets up keybindings for normal mode
func registerNormalModeBindings(r *Registry) {
	// Quit
	r.Register(ContextNormal, "q", ActionQuit)

	// Focus switching
	r.Register(ContextNormal, "tab", ActionSwitchFocus)

	// Navigation (sidebar/response)
	r.RegisterMultiple(ContextNormal, []string{"up", "k"}, ActionNavigateUp)
	r.RegisterMultiple(ContextNormal, []string{"down", "j"}, ActionNavigateDown)
	r.Register(ContextNormal, "pgup", ActionPageUp)
	r.Register(ContextNormal, "pgdown", ActionPageDown)
	r.Register(ContextNormal, "ctrl+u", ActionHalfPageUp)
	r.Register(ContextNormal, "ctrl+d", ActionHalfPageDown)
	r.Register(ContextNormal, "home", ActionGoToTop)
	r.Register(ContextNormal, "end", ActionGoToBottom)
	r.Register(ContextNormal, "g", ActionGoToTopPrepare)
	r.Register(ContextNormal, "gg", ActionGoToTop)
	r.Register(ContextNormal, "G", ActionGoToBottom)

	// Goto and search
	r.Register(ContextNormal, ":", ActionOpenGoto)
	r.Register(ContextNormal, "/", ActionOpenSearch)

	// File operations
	r.Register(ContextNormal, "enter", ActionExecute)
	r.Register(ContextNormal, "i", ActionOpenInspect)
	r.Register(ContextNormal, "x", ActionOpenEditor)
	r.Register(ContextNormal, "X", ActionConfigureEditor)
	r.Register(ContextNormal, "d", ActionDuplicateFile)
	r.Register(ContextNormal, "D", ActionDeleteFile)
	r.Register(ContextNormal, "R", ActionRenameFile)
	r.Register(ContextNormal, "F", ActionCreateFile)
	r.Register(ContextNormal, "r", ActionRefreshFiles)

	// Response operations
	r.Register(ContextNormal, "s", ActionSaveResponse)
	r.Register(ContextNormal, "c", ActionCopyToClipboard)
	r.Register(ContextNormal, "b", ActionToggleBody)
	r.Register(ContextNormal, "B", ActionToggleHeaders)
	r.Register(ContextNormal, "f", ActionToggleFullscreen)
	r.Register(ContextNormal, "w", ActionPinResponse)
	r.Register(ContextNormal, "W", ActionShowDiff)
	r.Register(ContextNormal, "J", ActionFilterResponse)

	// Modal launchers
	r.Register(ContextNormal, "v", ActionOpenVariables)
	r.Register(ContextNormal, "h", ActionOpenHeaders)
	r.Register(ContextNormal, "e", ActionOpenErrorDetail)
	r.Register(ContextNormal, "E", ActionOpenBodyOverride)
	r.Register(ContextNormal, "I", ActionShowStatusDetail)
	r.Register(ContextNormal, "p", ActionOpenProfiles)
	r.Register(ContextNormal, "ctrl+p", ActionOpenRecentFiles)
	r.Register(ContextNormal, "H", ActionOpenHistory)
	r.Register(ContextNormal, "A", ActionOpenAnalytics)
	r.Register(ContextNormal, "S", ActionOpenStressTest)
	r.Register(ContextNormal, "M", ActionOpenMockServer)
	r.Register(ContextNormal, "y", ActionOpenProxy)
	r.Register(ContextNormal, "t", ActionOpenTagFilter)
	r.Register(ContextNormal, "T", ActionClearTagFilter)
	r.Register(ContextNormal, "?", ActionOpenHelp)
	r.Register(ContextNormal, "o", ActionOpenOAuth)
	r.Register(ContextNormal, "O", ActionOpenOAuthDetail)
	r.Register(ContextNormal, "C", ActionOpenConfigView)
	r.Register(ContextNormal, "m", ActionOpenDocumentation)
	r.Register(ContextNormal, "n", ActionSearchNext)
	r.Register(ContextNormal, "N", ActionSearchPrevious)
	r.Register(ContextNormal, "ctrl+r", ActionRefresh)
	r.Register(ContextNormal, "P", ActionNoOp) // openProfilesInEditor - handled specially
	r.Register(ContextNormal, "ctrl+x", ActionNoOp) // openSessionInEditor - handled specially
}

// registerSearchBindings sets up keybindings for search mode
func registerSearchBindings(r *Registry) {
	// Search mode uses text input bindings plus specific actions
	r.RegisterMultiple(ContextSearch, []string{"ctrl+v", "shift+insert", "super+v"}, ActionTextPaste)
	r.Register(ContextSearch, "ctrl+y", ActionTextDeleteWord)
	r.Register(ContextSearch, "ctrl+k", ActionTextClearAfter)
	r.Register(ContextSearch, "backspace", ActionTextBackspace)
	r.Register(ContextSearch, "left", ActionTextMoveLeft)
	r.Register(ContextSearch, "right", ActionTextMoveRight)
	r.RegisterMultiple(ContextSearch, []string{"home", "ctrl+a"}, ActionTextMoveHome)
	r.RegisterMultiple(ContextSearch, []string{"end", "ctrl+e"}, ActionTextMoveEnd)
	r.Register(ContextSearch, "enter", ActionTextSubmit)
	r.Register(ContextSearch, "esc", ActionTextCancel)
	r.Register(ContextSearch, "ctrl+r", ActionOpenRecentFiles)
}

// registerGotoBindings sets up keybindings for goto mode
func registerGotoBindings(r *Registry) {
	// Goto mode uses text input bindings
	r.Register(ContextGoto, "esc", ActionTextCancel)
	r.Register(ContextGoto, "enter", ActionTextSubmit)
	r.Register(ContextGoto, "backspace", ActionTextBackspace)
}

// registerVariableBindings sets up keybindings for variable editor
func registerVariableBindings(r *Registry) {
	r.RegisterMultiple(ContextVariableList, []string{"esc", "v", "q"}, ActionCloseModal)
	r.RegisterMultiple(ContextVariableList, []string{"up", "k"}, ActionNavigateUp)
	r.RegisterMultiple(ContextVariableList, []string{"down", "j"}, ActionNavigateDown)
	r.Register(ContextVariableList, "a", ActionVarAdd)
	r.Register(ContextVariableList, "e", ActionVarEdit)
	r.Register(ContextVariableList, "d", ActionVarDelete)
	r.Register(ContextVariableList, "m", ActionVarManage)
	r.Register(ContextVariableList, " ", ActionVarToggle)
	r.Register(ContextVariableList, "ctrl+s", ActionTextSubmit)

	// Variable edit/add/delete modes use text input
	r.Register(ContextVariableEdit, "esc", ActionTextCancel)
	r.Register(ContextVariableEdit, "enter", ActionTextSubmit)
	r.Register(ContextVariableEdit, "backspace", ActionTextBackspace)
	r.Register(ContextVariableEdit, "up", ActionNavigateUp)
}

// registerHeaderBindings sets up keybindings for header editor
func registerHeaderBindings(r *Registry) {
	r.RegisterMultiple(ContextHeaderList, []string{"esc", "h", "q", "tab"}, ActionCloseModal)
	r.RegisterMultiple(ContextHeaderList, []string{"up", "k"}, ActionNavigateUp)
	r.RegisterMultiple(ContextHeaderList, []string{"down", "j"}, ActionNavigateDown)
	r.RegisterMultiple(ContextHeaderList, []string{"shift+up", "K"}, ActionNavigateUp) // Move header up
	r.RegisterMultiple(ContextHeaderList, []string{"shift+down", "J"}, ActionNavigateDown) // Move header down
	r.Register(ContextHeaderList, "enter", ActionHeaderEdit)
	r.Register(ContextHeaderList, "r", ActionHeaderDelete)
	r.Register(ContextHeaderList, "p", ActionTextPaste)
	r.Register(ContextHeaderList, "C", ActionHeaderAdd)
	r.Register(ContextHeaderList, "pgup", ActionPageUp)
	r.Register(ContextHeaderList, "pgdown", ActionPageDown)
	r.Register(ContextHeaderList, "ctrl+u", ActionHalfPageUp)
	r.Register(ContextHeaderList, "ctrl+d", ActionHalfPageDown)
	r.Register(ContextHeaderList, "home", ActionGoToTop)
	r.Register(ContextHeaderList, "end", ActionGoToBottom)
	r.Register(ContextHeaderList, "g", ActionGoToTopPrepare)
	r.Register(ContextHeaderList, "gg", ActionGoToTop)
	r.Register(ContextHeaderList, "G", ActionGoToBottom)

	// Header edit mode uses text input
	r.Register(ContextHeaderEdit, "esc", ActionTextCancel)
	r.Register(ContextHeaderEdit, "enter", ActionTextSubmit)
	r.Register(ContextHeaderEdit, "backspace", ActionTextBackspace)
}

// registerProfileBindings sets up keybindings for profile management
func registerProfileBindings(r *Registry) {
	r.RegisterMultiple(ContextProfileList, []string{"esc", "p", "q", "ctrl+p"}, ActionCloseModal)
	r.RegisterMultiple(ContextProfileList, []string{"up", "k"}, ActionNavigateUp)
	r.RegisterMultiple(ContextProfileList, []string{"down", "j"}, ActionNavigateDown)
	r.Register(ContextProfileList, "enter", ActionProfileSwitch)
	r.Register(ContextProfileList, "n", ActionProfileCreate)
	r.Register(ContextProfileList, "d", ActionProfileDuplicate)
	r.Register(ContextProfileList, "D", ActionProfileDelete)

	// Profile edit mode uses text input
	r.Register(ContextProfileEdit, "esc", ActionTextCancel)
	r.Register(ContextProfileEdit, "enter", ActionTextSubmit)
	r.Register(ContextProfileEdit, "backspace", ActionTextBackspace)
}

// registerDocumentationBindings sets up keybindings for documentation viewer
func registerDocumentationBindings(r *Registry) {
	r.RegisterMultiple(ContextDocumentation, []string{"esc", "m", "q"}, ActionCloseModal)
	r.RegisterMultiple(ContextDocumentation, []string{"up", "k"}, ActionNavigateUp)
	r.RegisterMultiple(ContextDocumentation, []string{"down", "j"}, ActionNavigateDown)
	r.Register(ContextDocumentation, "home", ActionGoToTop)
	r.Register(ContextDocumentation, "end", ActionGoToBottom)
	r.Register(ContextDocumentation, "g", ActionGoToTopPrepare)
	r.Register(ContextDocumentation, "gg", ActionGoToTop)
	r.Register(ContextDocumentation, "G", ActionGoToBottom)
	r.Register(ContextDocumentation, "pgup", ActionPageUp)
	r.Register(ContextDocumentation, "pgdown", ActionPageDown)
	r.Register(ContextDocumentation, "ctrl+u", ActionHalfPageUp)
	r.Register(ContextDocumentation, "ctrl+d", ActionHalfPageDown)
	r.RegisterMultiple(ContextDocumentation, []string{"enter", " "}, ActionTextSubmit) // Navigate docs
}

// registerHistoryBindings sets up keybindings for history browser
func registerHistoryBindings(r *Registry) {
	r.RegisterMultiple(ContextHistory, []string{"esc", "H", "q", "tab"}, ActionCloseModal)
	r.RegisterMultiple(ContextHistory, []string{"up", "k"}, ActionNavigateUp)
	r.RegisterMultiple(ContextHistory, []string{"down", "j"}, ActionNavigateDown)
	r.Register(ContextHistory, "/", ActionOpenSearch)
	r.Register(ContextHistory, "enter", ActionHistoryExecute)
	r.Register(ContextHistory, "r", ActionHistoryRollback)
	r.Register(ContextHistory, "p", ActionHistoryPaginate)
	r.Register(ContextHistory, "C", ActionHistoryClear)
	r.Register(ContextHistory, "pgup", ActionPageUp)
	r.Register(ContextHistory, "pgdown", ActionPageDown)
	r.Register(ContextHistory, "ctrl+u", ActionHalfPageUp)
	r.Register(ContextHistory, "ctrl+d", ActionHalfPageDown)
	r.Register(ContextHistory, "home", ActionGoToTop)
	r.Register(ContextHistory, "end", ActionGoToBottom)
	r.Register(ContextHistory, "g", ActionGoToTopPrepare)
	r.Register(ContextHistory, "gg", ActionGoToTop)
	r.Register(ContextHistory, "G", ActionGoToBottom)
}

// registerAnalyticsBindings sets up keybindings for analytics viewer
func registerAnalyticsBindings(r *Registry) {
	r.RegisterMultiple(ContextAnalytics, []string{"esc", "A", "q"}, ActionCloseModal)
	r.Register(ContextAnalytics, "tab", ActionSwitchPane)
	r.RegisterMultiple(ContextAnalytics, []string{"up", "k"}, ActionNavigateUp)
	r.RegisterMultiple(ContextAnalytics, []string{"down", "j"}, ActionNavigateDown)
	r.Register(ContextAnalytics, "enter", ActionTextSubmit) // Select/view
	r.Register(ContextAnalytics, "p", ActionAnalyticsPaginate)
	r.Register(ContextAnalytics, "t", ActionOpenTagFilter)
	r.Register(ContextAnalytics, "C", ActionAnalyticsClear)
	r.Register(ContextAnalytics, "pgup", ActionPageUp)
	r.Register(ContextAnalytics, "pgdown", ActionPageDown)
	r.Register(ContextAnalytics, "ctrl+u", ActionHalfPageUp)
	r.Register(ContextAnalytics, "ctrl+d", ActionHalfPageDown)
	r.Register(ContextAnalytics, "g", ActionGoToTopPrepare)
	r.Register(ContextAnalytics, "gg", ActionGoToTop)
	r.Register(ContextAnalytics, "G", ActionGoToBottom)
	r.Register(ContextAnalytics, "home", ActionGoToTop)
	r.Register(ContextAnalytics, "end", ActionGoToBottom)
}

// registerStressTestBindings sets up keybindings for stress test modes
func registerStressTestBindings(r *Registry) {
	r.RegisterMultiple(ContextStressTest, []string{"esc", "n", "N"}, ActionCloseModal)
	r.RegisterMultiple(ContextStressTest, []string{"y", "Y"}, ActionConfirm)
	r.Register(ContextStressTest, "ctrl+l", ActionStressTestLoad)
	r.RegisterMultiple(ContextStressTest, []string{"up", "k"}, ActionNavigateUp)
	r.RegisterMultiple(ContextStressTest, []string{"down", "j"}, ActionNavigateDown)
	r.Register(ContextStressTest, "enter", ActionTextSubmit)
	r.Register(ContextStressTest, "ctrl+s", ActionStressTestSave)
	r.Register(ContextStressTest, "backspace", ActionTextBackspace)
	r.Register(ContextStressTest, "delete", ActionTextDelete)
	r.Register(ContextStressTest, "left", ActionTextMoveLeft)
	r.Register(ContextStressTest, "right", ActionTextMoveRight)
	r.Register(ContextStressTest, "home", ActionTextMoveHome)
	r.Register(ContextStressTest, "end", ActionTextMoveEnd)
	r.Register(ContextStressTest, "d", ActionStressTestDelete)
	r.Register(ContextStressTest, "l", ActionStressTestLoad)
	r.Register(ContextStressTest, "r", ActionRefresh)
}

// registerHelpBindings sets up keybindings for help viewer
func registerHelpBindings(r *Registry) {
	r.RegisterMultiple(ContextHelp, []string{"esc", "?", "q"}, ActionCloseModal)
	r.Register(ContextHelp, "/", ActionOpenSearch)
	r.RegisterMultiple(ContextHelp, []string{"up", "k"}, ActionNavigateUp)
	r.RegisterMultiple(ContextHelp, []string{"down", "j"}, ActionNavigateDown)
	r.Register(ContextHelp, "pgup", ActionPageUp)
	r.Register(ContextHelp, "pgdown", ActionPageDown)
	r.Register(ContextHelp, "ctrl+u", ActionHalfPageUp)
	r.Register(ContextHelp, "ctrl+d", ActionHalfPageDown)
	r.Register(ContextHelp, "g", ActionGoToTopPrepare)
	r.Register(ContextHelp, "gg", ActionGoToTop)
	r.Register(ContextHelp, "G", ActionGoToBottom)
	r.Register(ContextHelp, "home", ActionGoToTop)
	r.Register(ContextHelp, "end", ActionGoToBottom)
}

// registerInspectBindings sets up keybindings for request inspector
func registerInspectBindings(r *Registry) {
	r.RegisterMultiple(ContextInspect, []string{"esc", "i", "q"}, ActionCloseModal)
	r.Register(ContextInspect, "enter", ActionExecute)
	r.RegisterMultiple(ContextInspect, []string{"up", "k"}, ActionNavigateUp)
	r.RegisterMultiple(ContextInspect, []string{"down", "j"}, ActionNavigateDown)
	r.Register(ContextInspect, "pgup", ActionPageUp)
	r.Register(ContextInspect, "pgdown", ActionPageDown)
	r.Register(ContextInspect, "g", ActionGoToTopPrepare)
	r.Register(ContextInspect, "gg", ActionGoToTop)
	r.Register(ContextInspect, "G", ActionGoToBottom)
	r.Register(ContextInspect, "home", ActionGoToTop)
	r.Register(ContextInspect, "end", ActionGoToBottom)
}

// registerWebSocketBindings sets up keybindings for WebSocket interface
func registerWebSocketBindings(r *Registry) {
	r.RegisterMultiple(ContextWebSocket, []string{"esc", "q"}, ActionCloseModal)
	r.Register(ContextWebSocket, "tab", ActionSwitchPane)
	r.RegisterMultiple(ContextWebSocket, []string{"up", "k"}, ActionNavigateUp)
	r.RegisterMultiple(ContextWebSocket, []string{"down", "j"}, ActionNavigateDown)
	r.Register(ContextWebSocket, "enter", ActionWSSend)
	r.Register(ContextWebSocket, "d", ActionWSDisconnect)
	r.Register(ContextWebSocket, "C", ActionWSClear)
	r.Register(ContextWebSocket, "pgup", ActionPageUp)
	r.Register(ContextWebSocket, "pgdown", ActionPageDown)
	r.Register(ContextWebSocket, "g", ActionGoToTopPrepare)
	r.Register(ContextWebSocket, "gg", ActionGoToTop)
	r.Register(ContextWebSocket, "G", ActionGoToBottom)
	r.Register(ContextWebSocket, "c", ActionWSSelectChannel)
}

// registerModalBindings sets up generic modal bindings
func registerModalBindings(r *Registry) {
	// Generic close for simple modals
	r.RegisterMultiple(ContextModal, []string{"esc", "q"}, ActionCloseModal)
	r.RegisterMultiple(ContextModal, []string{"j", "down"}, ActionNavigateDown)
	r.RegisterMultiple(ContextModal, []string{"k", "up"}, ActionNavigateUp)
	r.Register(ContextModal, "g", ActionGoToTopPrepare)
	r.Register(ContextModal, "gg", ActionGoToTop)
	r.Register(ContextModal, "G", ActionGoToBottom)
}

// registerConfirmBindings sets up confirmation dialog bindings
func registerConfirmBindings(r *Registry) {
	r.RegisterMultiple(ContextConfirm, []string{"y", "Y"}, ActionConfirm)
	r.RegisterMultiple(ContextConfirm, []string{"n", "N", "esc"}, ActionCancel)
}

// registerViewerBindings sets up generic viewer bindings (error details, status, etc.)
func registerViewerBindings(r *Registry) {
	// Viewer is a base context that other viewers can inherit from
	// Already registered in registerNavigationBindings
}
