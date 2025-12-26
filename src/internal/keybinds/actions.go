package keybinds

// Action represents a user action that can be triggered by a keybinding
type Action string

// Context represents the context in which keybindings are active
type Context string

const (
	// Contexts define where keybindings are active
	ContextGlobal       Context = "global"        // Available everywhere
	ContextNormal       Context = "normal"        // Normal mode
	ContextSearch       Context = "search"        // Search input mode
	ContextGoto         Context = "goto"          // Goto line mode
	ContextVariableList Context = "variable_list" // Variable list view
	ContextVariableEdit Context = "variable_edit" // Variable editing
	ContextHeaderList   Context = "header_list"   // Header list view
	ContextHeaderEdit   Context = "header_edit"   // Header editing
	ContextProfileList  Context = "profile_list"  // Profile list view
	ContextProfileEdit  Context = "profile_edit"  // Profile editing
	ContextDocumentation Context = "documentation" // Documentation viewer
	ContextHistory      Context = "history"       // History browser
	ContextAnalytics    Context = "analytics"     // Analytics viewer
	ContextStressTest   Context = "stress_test"   // Stress test modes
	ContextHelp         Context = "help"          // Help viewer
	ContextInspect      Context = "inspect"       // Request inspector
	ContextWebSocket    Context = "websocket"     // WebSocket interface
	ContextModal        Context = "modal"         // Generic modal (applies to all modals)
	ContextTextInput    Context = "text_input"    // Text input (applies to all text inputs)
	ContextConfirm      Context = "confirm"       // Confirmation dialogs
	ContextViewer       Context = "viewer"        // Generic viewer (scrollable content)
)

const (
	// Global actions
	ActionQuit        Action = "quit"         // Quit application
	ActionQuitForce   Action = "quit_force"   // Force quit (ctrl+c)
	ActionStopStream  Action = "stop_stream"  // Stop active stream

	// Navigation actions
	ActionNavigateUp       Action = "navigate_up"        // Move up one item
	ActionNavigateDown     Action = "navigate_down"      // Move down one item
	ActionPageUp           Action = "page_up"            // Move up one page
	ActionPageDown         Action = "page_down"          // Move down one page
	ActionHalfPageUp       Action = "half_page_up"       // Move up half page (ctrl+u)
	ActionHalfPageDown     Action = "half_page_down"     // Move down half page (ctrl+d)
	ActionGoToTop          Action = "go_to_top"          // Go to top
	ActionGoToBottom       Action = "go_to_bottom"       // Go to bottom
	ActionGoToTopPrepare   Action = "go_to_top_prepare"  // First 'g' in 'gg' sequence
	ActionScrollUp         Action = "scroll_up"          // Scroll viewport up
	ActionScrollDown       Action = "scroll_down"        // Scroll viewport down

	// Focus and panel switching
	ActionSwitchFocus      Action = "switch_focus"       // Switch focus between panels
	ActionFocusSidebar     Action = "focus_sidebar"      // Focus sidebar
	ActionFocusResponse    Action = "focus_response"     // Focus response panel
	ActionSwitchPane       Action = "switch_pane"        // Switch between panes in multi-pane modals

	// Text input actions
	ActionTextInsertChar   Action = "text_insert_char"   // Insert character
	ActionTextBackspace    Action = "text_backspace"     // Delete char before cursor
	ActionTextDelete       Action = "text_delete"        // Delete char at cursor
	ActionTextMoveLeft     Action = "text_move_left"     // Move cursor left
	ActionTextMoveRight    Action = "text_move_right"    // Move cursor right
	ActionTextMoveHome     Action = "text_move_home"     // Move cursor to start
	ActionTextMoveEnd      Action = "text_move_end"      // Move cursor to end
	ActionTextPaste        Action = "text_paste"         // Paste from clipboard
	ActionTextDeleteWord   Action = "text_delete_word"   // Delete word
	ActionTextClearBefore  Action = "text_clear_before"  // Clear before cursor
	ActionTextClearAfter   Action = "text_clear_after"   // Clear after cursor
	ActionTextSubmit       Action = "text_submit"        // Submit text input
	ActionTextCancel       Action = "text_cancel"        // Cancel text input

	// Modal actions
	ActionCloseModal       Action = "close_modal"        // Close current modal
	ActionCloseModalAlt    Action = "close_modal_alt"    // Alternative close (usually 'q')
	ActionConfirm          Action = "confirm"            // Confirm action (y/Y)
	ActionCancel           Action = "cancel"             // Cancel action (n/N)

	// File operations (Normal mode, sidebar focused)
	ActionExecute          Action = "execute"            // Execute request
	ActionOpenEditor       Action = "open_editor"        // Open in external editor
	ActionConfigureEditor  Action = "configure_editor"   // Configure editor
	ActionDuplicateFile    Action = "duplicate_file"     // Duplicate file
	ActionDeleteFile       Action = "delete_file"        // Delete file (with confirm)
	ActionRenameFile       Action = "rename_file"        // Rename file
	ActionCreateFile       Action = "create_file"        // Create new file
	ActionRefreshFiles     Action = "refresh_files"      // Refresh file list

	// Response operations (Normal mode)
	ActionSaveResponse     Action = "save_response"      // Save response to file
	ActionCopyToClipboard  Action = "copy_to_clipboard"  // Copy response to clipboard
	ActionToggleBody       Action = "toggle_body"        // Toggle body visibility
	ActionToggleHeaders    Action = "toggle_headers"     // Toggle headers visibility
	ActionToggleFullscreen Action = "toggle_fullscreen"  // Toggle fullscreen mode
	ActionPinResponse      Action = "pin_response"       // Pin response for comparison
	ActionShowDiff         Action = "show_diff"          // Show diff with pinned response
	ActionFilterResponse   Action = "filter_response"    // Filter response with JMESPath
	ActionOpenErrorDetail  Action = "open_error_detail"  // Open error detail modal
	ActionOpenBodyOverride Action = "open_body_override" // Open body override editor

	// Modal launchers (Normal mode)
	ActionOpenInspect       Action = "open_inspect"        // Open request inspector
	ActionOpenVariables     Action = "open_variables"      // Open variable editor
	ActionOpenHeaders       Action = "open_headers"        // Open header editor
	ActionOpenInteractive   Action = "open_interactive"    // Open interactive variables
	ActionOpenProfiles      Action = "open_profiles"       // Open profile switcher
	ActionOpenRecentFiles   Action = "open_recent_files"   // Open MRU list
	ActionOpenHistory       Action = "open_history"        // Open history browser
	ActionOpenAnalytics     Action = "open_analytics"      // Open analytics viewer
	ActionOpenStressTest    Action = "open_stress_test"    // Open stress test config
	ActionOpenMockServer    Action = "open_mock_server"    // Open mock server
	ActionOpenProxy         Action = "open_proxy"          // Open proxy viewer
	ActionOpenTagFilter     Action = "open_tag_filter"     // Open tag filter
	ActionClearTagFilter    Action = "clear_tag_filter"    // Clear tag filter
	ActionOpenHelp          Action = "open_help"           // Open help viewer
	ActionOpenOAuth         Action = "open_oauth"          // Open OAuth config
	ActionOpenOAuthDetail   Action = "open_oauth_detail"   // Open OAuth detail
	ActionOpenConfigView    Action = "open_config_view"    // Open config viewer
	ActionOpenDocumentation Action = "open_documentation"  // Open documentation
	ActionOpenGoto          Action = "open_goto"           // Open goto line input
	ActionOpenSearch        Action = "open_search"         // Open search input

	// Variable editor actions
	ActionVarAdd      Action = "var_add"      // Add variable
	ActionVarEdit     Action = "var_edit"     // Edit variable
	ActionVarDelete   Action = "var_delete"   // Delete variable
	ActionVarManage   Action = "var_manage"   // Manage variable (options)
	ActionVarToggle   Action = "var_toggle"   // Toggle variable selection

	// Header editor actions
	ActionHeaderAdd    Action = "header_add"    // Add header
	ActionHeaderEdit   Action = "header_edit"   // Edit header
	ActionHeaderDelete Action = "header_delete" // Delete header

	// Profile actions
	ActionProfileSwitch    Action = "profile_switch"    // Switch to profile
	ActionProfileCreate    Action = "profile_create"    // Create new profile
	ActionProfileDuplicate Action = "profile_duplicate" // Duplicate profile
	ActionProfileDelete    Action = "profile_delete"    // Delete profile

	// History actions
	ActionHistoryExecute   Action = "history_execute"   // Execute from history
	ActionHistoryRollback  Action = "history_rollback"  // Rollback history
	ActionHistoryPaginate  Action = "history_paginate"  // Paginate history
	ActionHistoryClear     Action = "history_clear"     // Clear history

	// Analytics actions
	ActionAnalyticsPaginate Action = "analytics_paginate" // Paginate analytics
	ActionAnalyticsClear    Action = "analytics_clear"    // Clear analytics

	// Stress test actions
	ActionStressTestStart  Action = "stress_test_start"  // Start stress test
	ActionStressTestStop   Action = "stress_test_stop"   // Stop stress test
	ActionStressTestSave   Action = "stress_test_save"   // Save stress test config
	ActionStressTestLoad   Action = "stress_test_load"   // Load stress test config
	ActionStressTestDelete Action = "stress_test_delete" // Delete stress test result
	ActionStressTestExport Action = "stress_test_export" // Export stress test result

	// WebSocket actions
	ActionWSConnect      Action = "ws_connect"       // Connect to WebSocket
	ActionWSDisconnect   Action = "ws_disconnect"    // Disconnect WebSocket
	ActionWSSend         Action = "ws_send"          // Send WebSocket message
	ActionWSClear        Action = "ws_clear"         // Clear WebSocket messages
	ActionWSSelectChannel Action = "ws_select_channel" // Select WebSocket channel

	// Search actions
	ActionSearchNext     Action = "search_next"      // Go to next search result
	ActionSearchPrevious Action = "search_previous"  // Go to previous search result
	ActionSearchClear    Action = "search_clear"     // Clear search

	// Mock server actions
	ActionMockToggle   Action = "mock_toggle"    // Toggle mock server

	// Diff viewer actions
	ActionDiffClose    Action = "diff_close"     // Close diff viewer

	// JSONPath actions
	ActionJSONPathSave Action = "jsonpath_save"  // Save JSONPath bookmark

	// Other actions
	ActionShowStatusDetail Action = "show_status_detail" // Show status detail
	ActionRefresh          Action = "refresh"            // Refresh/reload current view
	ActionNoOp             Action = "noop"               // No operation (ignore key)
)

// ActionInfo contains metadata about an action
type ActionInfo struct {
	Action      Action
	Description string
	Category    string
}

// GetActionInfo returns human-readable information about an action
func GetActionInfo(action Action) ActionInfo {
	infos := map[Action]ActionInfo{
		ActionQuit:             {ActionQuit, "Quit application", "Global"},
		ActionQuitForce:        {ActionQuitForce, "Force quit", "Global"},
		ActionStopStream:       {ActionStopStream, "Stop active stream", "Global"},
		ActionNavigateUp:       {ActionNavigateUp, "Move up", "Navigation"},
		ActionNavigateDown:     {ActionNavigateDown, "Move down", "Navigation"},
		ActionPageUp:           {ActionPageUp, "Page up", "Navigation"},
		ActionPageDown:         {ActionPageDown, "Page down", "Navigation"},
		ActionHalfPageUp:       {ActionHalfPageUp, "Half page up", "Navigation"},
		ActionHalfPageDown:     {ActionHalfPageDown, "Half page down", "Navigation"},
		ActionGoToTop:          {ActionGoToTop, "Go to top", "Navigation"},
		ActionGoToBottom:       {ActionGoToBottom, "Go to bottom", "Navigation"},
		ActionExecute:          {ActionExecute, "Execute request", "File Operations"},
		ActionOpenEditor:       {ActionOpenEditor, "Open in editor", "File Operations"},
		ActionSaveResponse:     {ActionSaveResponse, "Save response", "Response"},
		ActionCopyToClipboard:  {ActionCopyToClipboard, "Copy to clipboard", "Response"},
		ActionToggleBody:       {ActionToggleBody, "Toggle body", "Response"},
		ActionToggleHeaders:    {ActionToggleHeaders, "Toggle headers", "Response"},
		ActionToggleFullscreen: {ActionToggleFullscreen, "Toggle fullscreen", "View"},
		ActionOpenVariables:    {ActionOpenVariables, "Open variables", "Editors"},
		ActionOpenHeaders:      {ActionOpenHeaders, "Open headers", "Editors"},
		ActionOpenHelp:         {ActionOpenHelp, "Open help", "Information"},
		// ... add more as needed
	}

	if info, ok := infos[action]; ok {
		return info
	}

	return ActionInfo{action, string(action), "Unknown"}
}

// IsGlobalAction returns true if the action is available in all contexts
func IsGlobalAction(action Action) bool {
	globalActions := map[Action]bool{
		ActionQuit:      true,
		ActionQuitForce: true,
		ActionStopStream: true,
	}
	return globalActions[action]
}
