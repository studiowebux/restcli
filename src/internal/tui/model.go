package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/studiowebux/restcli/internal/analytics"
	"github.com/studiowebux/restcli/internal/history"
	"github.com/studiowebux/restcli/internal/jsonpath"
	"github.com/studiowebux/restcli/internal/keybinds"
	"github.com/studiowebux/restcli/internal/session"
	"github.com/studiowebux/restcli/internal/types"
)

// Mode represents the current TUI mode
type Mode int

const (
	ModeNormal Mode = iota
	ModeSearch
	ModeGoto
	ModeVariableList
	ModeVariableAdd
	ModeVariableEdit
	ModeVariableDelete
	ModeVariableOptions
	ModeVariableManage
	ModeVariablePromptInteractive
	ModeHeaderList
	ModeHeaderAdd
	ModeHeaderEdit
	ModeHeaderDelete
	ModeProfileSwitch
	ModeProfileCreate
	ModeDocumentation
	ModeHistory
	ModeHistoryClearConfirm
	ModeAnalytics
	ModeAnalyticsClearConfirm
	ModeStressTest
	ModeStressTestConfig
	ModeStressTestLoadConfig
	ModeStressTestProgress
	ModeStressTestResults
	ModeHelp
	ModeInspect
	ModeRename
	ModeEditorConfig
	ModeOAuthConfig
	ModeOAuthEdit
	ModeConfigView
	ModeDelete
	ModeProfileEdit
	ModeProfileDuplicate
	ModeProfileDeleteConfirm
	ModeVariableAlias
	ModeShellErrors
	ModeCreateFile
	ModeMRU
	ModeDiff
	ModeConfirmExecution
	ModeErrorDetail
	ModeStatusDetail
	ModeBodyOverride
	ModeJSONPathHistory
	ModeTagFilter
	ModeMockServer
	ModeProxyViewer
	ModeProxyDetail
	ModeWebSocket
)

// Model represents the TUI state
type Model struct {
	// Core state
	sessionMgr        *session.Manager
	analyticsManager  *analytics.Manager
	historyManager    *history.Manager
	bookmarkManager   *jsonpath.BookmarkManager
	keybinds          *keybinds.Registry
	mode              Mode
	version           string
	updateAvailable   bool
	latestVersion     string
	updateURL         string

	// File explorer state (encapsulates file navigation, search, and filtering)
	fileExplorer *FileExplorerState

	// Response-specific search/navigation (not part of file explorer)
	responseSearchMatches []int  // Line numbers in response matching search
	responseSearchIndex   int    // Current position in response search results
	searchInResponseCtx   bool   // True if current search is in response context
	gotoInput             string // Goto line input
	gotoCursor            int    // Cursor position in goto input
	searchInput           string // Temporary search input buffer while in ModeSearch
	searchCursor          int    // Cursor position in search input

	// Request/Response
	currentRequests       []types.HttpRequest
	currentRequest        *types.HttpRequest
	currentResponse       *types.RequestResult
	responseView          viewport.Model
	responseContent       string // Full formatted response content for searching

	// Response content cache tracking
	cachedResponsePtr    *types.RequestResult // Pointer to response that was cached
	cachedViewWidth      int                  // Viewport width used for cached content
	cachedFilterActive   bool                 // Filter state when cached
	cachedSearchActive   bool                 // Search highlight state when cached

	helpView              viewport.Model
	modalView             viewport.Model // For scrollable modal content

	// Streaming state
	streamState           *StreamState          // Thread-safe streaming state management
	streamedBody          string                // Accumulated streamed response body
	streamChannel         chan streamChunkMsg   // Channel for receiving stream chunks

	// Request cancellation (for regular non-streaming requests)
	requestState          *RequestState         // Thread-safe request cancellation

	// UI state
	width        int
	height       int
	statusMsg    string
	errorMsg     string      // Truncated error for footer
	fullErrorMsg   string // Full error message for detail modal
	fullStatusMsg  string // Full status message for detail modal
	focusedPanel   string // "sidebar" or "response"

	// Variable editor state
	varEditIndex     int
	varEditName      string
	varEditValue     string
	varEditCursor    int // Which field (0=name, 1=value)
	varEditNamePos   int // Cursor position in name field
	varEditValuePos  int // Cursor position in value field
	varOptionIndex   int

	// Alias editor state
	varAliasInput      string // Alias name being typed
	varAliasTargetIdx  int    // Option index being aliased

	// Header editor state
	headerEditIndex    int
	headerEditName     string
	headerEditValue    string
	headerEditCursor   int // Which field (0=name, 1=value)
	headerEditNamePos  int // Cursor position in name field
	headerEditValuePos int // Cursor position in value field

	// Profile state
	profileIndex  int
	profileName   string
	profileCursor int
	profileNamePos int // Cursor position in profile name

	// Profile edit state (encapsulates all profile editing UI state)
	profileEditState *ProfileEditState

	// Documentation viewer state (encapsulates all documentation UI state)
	docState *DocumentationState

	// History state (encapsulates all history UI state)
	historyState *HistoryState

	// Analytics state (encapsulates all analytics UI state)
	analyticsState *AnalyticsState

	// Stress test state (encapsulates all stress test UI state)
	stressTestState *StressTestState

	// Mock server state (encapsulates all mock server state)
	mockServerState *MockServerState

	// Proxy server state (encapsulates all proxy server state)
	proxyServerState *ProxyServerState

	// Rename state (encapsulates file rename input state)
	renameState *RenameState

	// OAuth config state
	oauthField  int
	oauthCursor int

	// Input states
	inputValue  string
	inputCursor int

	// Flags
	showHeaders       bool
	showBody          bool
	fullscreen        bool
	loading           bool
	gPressed          bool // Track if 'g' was pressed for 'gg' vim motion
	confirmationGiven bool // Track if user confirmed critical operation

	// Help search state
	helpSearchQuery  string
	helpSearchCursor int  // Cursor position in help search query
	helpSearchActive bool

	// Shell errors state
	shellErrors      []string
	shellErrorScroll int

	// Create file state
	createFileInput  string // Filename/path input
	createFileType   int    // Selected file type (0=http, 1=json, 2=yaml, 3=jsonc)
	createFileCursor int    // Cursor position in input

	// MRU state
	mruIndex int // Selected index in MRU list

	// Diff state
	pinnedResponse *types.RequestResult // Response pinned for comparison
	pinnedRequest  *types.HttpRequest   // Request info for pinned response
	diffView       viewport.Model       // Viewport for diff display (unified mode)
	diffViewMode   string               // "unified" or "split"
	diffLeftView   viewport.Model       // Left pane viewport (pinned) for split mode
	diffRightView  viewport.Model       // Right pane viewport (current) for split mode

	// Interactive variable prompt state
	interactiveVarNames      []string          // Queue of variables to prompt for
	interactiveVarValues     map[string]string // Collected values
	interactiveVarInput      string            // Current input value
	interactiveVarCursor     int               // Cursor position in input
	interactiveVarMode       string            // "select" or "input" - selection list or text input
	interactiveVarOptions    []string          // Available options for selection
	interactiveVarAliases    map[int][]string  // Aliases for each option (index -> alias names)
	interactiveVarActiveIdx  int               // Currently active option index
	interactiveVarSelectIdx  int               // Selected option in list

	// Streaming state
	isStreaming  bool   // True when actively streaming response
	streamCancel func() // Function to cancel ongoing stream

	// Body override state
	bodyOverrideInput  string // Edited body content
	bodyOverrideCursor int    // Cursor position (linear, not line-based)
	bodyOverride       string // Applied body override (cleared after send)

	// Filter state
	filterInput         string // JMESPath filter/query expression
	filterCursor        int    // Cursor position in filter input
	filteredResponse    string // Cached filtered response
	filterError         string // Filter error message
	filterActive        bool   // True when viewing filtered result
	filterEditing       bool   // True when actively editing filter in footer

	// JSONPath history state
	jsonpathBookmarks       []jsonpath.Bookmark // Loaded bookmarks
	jsonpathHistoryCursor   int                 // Selected bookmark index
	jsonpathHistorySearch   string              // Search filter for bookmarks
	jsonpathHistoryMatches  []jsonpath.Bookmark // Filtered bookmarks
	jsonpathHistorySearching bool                // True when in search mode

	// WebSocket state (Phase 2: Split-pane modal)
	wsState                 *WebSocketState            // Thread-safe WebSocket state management
	wsMessages              []types.ReceivedMessage    // Message history (left pane)
	wsConnectionStatus      string                     // Connection status: "connecting", "connected", "disconnected", "error"
	wsHistoryView           viewport.Model             // Left pane: message history viewport
	wsMessageMenuView       viewport.Model             // Right pane: predefined message menu viewport
	wsMessageChannel        chan types.ReceivedMessage // Channel for receiving messages from executor
	wsSendChannel           chan string                // Channel for sending user messages
	wsURL                   string                     // WebSocket URL being connected to
	wsError                 string                     // WebSocket error message if any
	wsPredefinedMessages    []types.WebSocketMessage   // All predefined messages from .ws file
	wsSendableMessages      []types.WebSocketMessage   // Filtered "send" messages for menu
	wsSelectedMessageIndex  int                        // Selected message in right pane menu
	wsFocusedPane           string                     // "history" or "menu" - which pane has focus
	wsConn                  interface{}                // Active WebSocket connection (for persistent mode)
	wsPendingMessageIndex   int                        // Message to send after connection completes (-1 = none)
	wsLastKey               string                     // Last key pressed (for detecting gg)
	wsShowClearConfirm      bool                       // True when showing clear history confirmation dialog
	wsSearchMode            bool                       // True when in search mode
	wsSearchQuery           string                     // Current search query
	wsStatusMsg             string                     // WebSocket-specific status message for footer
	wsComposerMode          bool                       // True when in custom message composer mode
	wsComposerMessage       string                     // Custom message being composed
}

// Init initializes the TUI
func (m *Model) Init() tea.Cmd {
	return nil
}

// Cleanup closes database connections and cleans up resources
func (m *Model) Cleanup() {
	if m.analyticsManager != nil {
		if err := m.analyticsManager.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "error closing analytics database: %v\n", err)
		}
	}
	if m.historyManager != nil {
		if err := m.historyManager.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "error closing history database: %v\n", err)
		}
	}
	if m.stressTestState != nil && m.stressTestState.GetManager() != nil {
		if err := m.stressTestState.GetManager().Close(); err != nil {
			fmt.Fprintf(os.Stderr, "error closing stress test database: %v\n", err)
		}
	}
	if m.bookmarkManager != nil {
		if err := m.bookmarkManager.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "error closing bookmark database: %v\n", err)
		}
	}
}

// Update handles messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		cmd = m.handleKeyPress(msg)

	// Mouse events - capture to prevent terminal scrolling, but don't use them for navigation
	case tea.MouseMsg:
		// Explicitly handle and discard mouse scroll to prevent terminal buffer scrolling
		// This keeps the app "on top" when scrolling with the mouse
		// All navigation remains keyboard-only as intended

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateViewport()

	case fileListLoadedMsg:
		m.fileExplorer.SetFiles(msg.files, msg.files)
		m.statusMsg = "Files loaded"
		// Reload current file's requests to reflect any changes
		m.loadRequestsFromCurrentFile()

	case chainCompleteMsg:
		m.loading = false
		m.requestState.Clear() // Clear cancel function
		if msg.success {
			if msg.response != nil {
				m.currentResponse = msg.response
			}
			m.errorMsg = ""
			m.fullErrorMsg = ""
			m.statusMsg = msg.message
			m.fullStatusMsg = msg.message
			m.updateResponseView()
			m.focusedPanel = "response"
		} else {
			m.errorMsg = msg.message
			m.fullErrorMsg = msg.message
			if len(msg.message) > 100 {
				m.statusMsg = msg.message[:97] + "..."
			} else {
				m.statusMsg = msg.message
			}
		}

	case requestExecutedMsg:
		m.loading = false // Clear loading flag
		m.requestState.Clear() // Clear cancel function
		m.currentResponse = msg.result
		// Clear any previous errors since request completed successfully
		m.errorMsg = ""
		m.fullErrorMsg = ""
		if len(msg.warnings) > 0 {
			statusText := fmt.Sprintf("Request completed (unresolved: %s)", strings.Join(msg.warnings, ", "))
			m.fullStatusMsg = statusText
			if len(statusText) > 100 {
				m.statusMsg = statusText[:97] + "..."
			} else {
				m.statusMsg = statusText
			}
		} else {
			m.statusMsg = "Request completed"
			m.fullStatusMsg = "Request completed"
		}
		m.updateResponseView()
		// Auto-switch focus to response panel so user can immediately scroll
		m.focusedPanel = "response"
		// Clear interactive variable values for next execution
		m.interactiveVarValues = nil
		// Show shell errors modal if any
		if len(msg.shellErrors) > 0 {
			m.shellErrors = msg.shellErrors
			m.shellErrorScroll = 0
			m.mode = ModeShellErrors
			m.updateShellErrorsView()
		}

	case streamChunkMsg:
		// Accumulate streaming chunks
		m.streamedBody += string(msg.chunk)

		// Update the display with current streamed content
		if m.currentResponse == nil {
			m.currentResponse = &types.RequestResult{}
		}
		m.currentResponse.Body = m.streamedBody
		m.cachedResponsePtr = nil // Invalidate cache since body changed in place
		m.statusMsg = "Streaming... (press 'q' to stop)"
		m.updateResponseView()
		m.focusedPanel = "response"

		// Auto-scroll to bottom to show latest data
		m.responseView.GotoBottom()

		// If not done, wait for next chunk
		if !msg.done {
			return m, m.waitForStreamChunk()
		}

		// Stream complete
		m.loading = false // Clear loading flag
		m.streamState.Stop()
		m.streamChannel = nil
		// Clear any previous errors since stream completed successfully
		m.errorMsg = ""
		m.fullErrorMsg = ""
		m.statusMsg = "Stream completed"

	case wsMessageReceivedMsg:
		// Add message to list
		if msg.message != nil {
			m.wsMessages = append(m.wsMessages, *msg.message)

			// Update history viewport with new message
			modalWidth := m.width - ModalWidthMargin
			modalHeight := m.height - ModalHeightMargin
			paneHeight := modalHeight - ViewportPaddingHorizontal
			historyWidth := (modalWidth * 6) / 10
			m.updateWebSocketHistoryView(historyWidth-4, paneHeight-2)

			// Auto-scroll to bottom to show latest message
			m.wsHistoryView.GotoBottom()

			// Update connection status based on message type
			if msg.message.Type == "system" {
				if strings.Contains(msg.message.Content, "Connected to") {
					m.wsConnectionStatus = "connected"

					// Send pending message if any
					if m.wsPendingMessageIndex >= 0 && m.wsPendingMessageIndex < len(m.wsSendableMessages) {
						pendingMsg := m.wsSendableMessages[m.wsPendingMessageIndex]
						if m.wsSendChannel != nil {
							go func() {
								select {
								case m.wsSendChannel <- pendingMsg.Content:
								default:
								}
							}()
						}
						m.wsPendingMessageIndex = -1 // Clear pending message
					}
				} else if strings.Contains(msg.message.Content, "Disconnected") || strings.Contains(msg.message.Content, "Error") {
					m.wsState.Stop()
					m.wsConnectionStatus = "disconnected"
				}
			}
		}

		// Wait for next message if connection still active
		if m.wsState.IsActive() {
			return m, m.waitForWsMessage()
		}

	case wsConnectionStatusMsg:
		m.wsConnectionStatus = msg.status
		m.statusMsg = fmt.Sprintf("WebSocket: %s", msg.status)

	case wsConnectionCompleteMsg:
		m.wsState.Stop()
		m.wsConnectionStatus = "disconnected"
		m.wsMessageChannel = nil
		if msg.err != nil {
			m.wsError = msg.err.Error()
			m.errorMsg = fmt.Sprintf("WebSocket error: %v", msg.err)
		} else {
			m.statusMsg = "WebSocket connection closed"
		}

	case oauthSuccessMsg:
		m.statusMsg = fmt.Sprintf("OAuth successful! Token stored (expires in %d seconds)", msg.expiresIn)

	case historyLoadedMsg:
		m.historyState.SetEntries(msg.entries)
		m.historyState.SetAllEntries(msg.entries) // Store unfiltered for search
		m.historyState.SetIndex(0)
		m.historyState.SetSearchQuery("") // Reset search on load
		m.historyState.SetSearchActive(false)
		if len(msg.entries) > 0 {
			m.statusMsg = fmt.Sprintf("Loaded %d history entries", len(msg.entries))
		}
		m.updateHistoryView() // Update viewport content with loaded history

	case analyticsLoadedMsg:
		m.analyticsState.SetStats(msg.stats)
		m.analyticsState.SetIndex(0)
		if len(msg.stats) > 0 {
			m.statusMsg = fmt.Sprintf("Loaded %d analytics entries", len(msg.stats))
		}
		m.updateAnalyticsView() // Update viewport content with loaded analytics

	case versionCheckMsg:
		if msg.err == nil && msg.available {
			m.updateAvailable = true
			m.latestVersion = msg.latestVersion
			m.updateURL = msg.url
			m.updateHelpView() // Refresh help view to show update notice
		}

	case stressTestRunsLoadedMsg:
		m.stressTestState.SetRuns(msg.runs)
		m.stressTestState.SetRunIndex(0)
		if len(msg.runs) > 0 {
			m.statusMsg = fmt.Sprintf("Loaded %d stress test runs", len(msg.runs))
		}
		m.updateStressTestListView()
		detailView := m.stressTestState.GetDetailView()
		detailView.GotoTop() // Reset scroll position for new load
		m.stressTestState.SetDetailView(detailView)

	case stressTestConfigsLoadedMsg:
		m.stressTestState.SetConfigs(msg.configs)
		m.stressTestState.SetConfigIndex(0)
		m.mode = ModeStressTestLoadConfig
		if len(msg.configs) > 0 {
			m.statusMsg = fmt.Sprintf("Loaded %d stress test configurations", len(msg.configs))
		} else {
			m.statusMsg = "No saved configurations found"
		}

	case stressTestProgressMsg:
		// Continue polling for progress
		return m, m.pollStressTestProgress()

	case stressTestCompletedMsg:
		// Stress test completed
		if m.stressTestState.GetExecutor() != nil {
			m.stressTestState.GetExecutor().Wait()
			m.stressTestState.SetExecutor(nil)
		}
		m.stressTestState.SetActiveRequest(nil)
		m.stressTestState.SetStopping(false)
		m.statusMsg = "Stress test completed"
		m.mode = ModeStressTestResults
		return m, m.loadStressTestRuns()

	case stressTestStoppedMsg:
		// Stress test stopped by user
		m.stressTestState.SetExecutor(nil)
		m.stressTestState.SetActiveRequest(nil)
		m.stressTestState.SetStopping(false)

		// Check if stop timed out
		if msg.err != nil {
			m.statusMsg = "Stress test cancelled (cleanup timeout - some resources may not have been released)"
			m.errorMsg = fmt.Sprintf("Stop timeout: %v", msg.err)
		} else {
			m.statusMsg = "Stress test cancelled"
		}

		m.mode = ModeStressTestResults
		return m, m.loadStressTestRuns()

	case promptInteractiveVarsMsg:
		// Initialize interactive variable prompting
		m.interactiveVarNames = msg.varNames
		m.interactiveVarValues = make(map[string]string)
		m.interactiveVarInput = ""
		m.interactiveVarCursor = 0
		m.mode = ModeVariablePromptInteractive
		m.initInteractiveVarPrompt()

	case mockServerStartedMsg:
		m.mockServerState.Start(msg.server, msg.configPath)
		m.statusMsg = fmt.Sprintf("Mock server started at %s", msg.address)
		// Start ticker for refreshing logs
		cmd = m.tickMockServer()

	case mockServerStoppedMsg:
		m.mockServerState.Stop()
		m.statusMsg = "Mock server stopped"

	case mockServerTickMsg:
		// Refresh mock server view if in that mode and server is running
		if m.mode == ModeMockServer && m.mockServerState.IsRunning() {
			// Return another tick to keep refreshing
			cmd = m.tickMockServer()
		}

	case proxyViewerTickMsg:
		// Refresh proxy viewer if in that mode and proxy is running
		if m.mode == ModeProxyViewer && m.proxyServerState.IsRunning() {
			m.updateProxyView()
			// Return another tick to keep refreshing
			cmd = m.tickProxyViewer()
		}

	case proxyLogReceivedMsg:
		// New proxy log received - update view
		if m.mode == ModeProxyViewer && m.proxyServerState.IsRunning() {
			m.updateProxyView()
		}
		// Continue listening for more logs
		if m.proxyServerState.IsRunning() && m.proxyServerState.GetServer() != nil {
			cmd = m.listenForProxyLogs()
		}

	case clearStatusMsg:
		m.statusMsg = ""

	case clearErrorMsg:
		m.errorMsg = ""
		m.fullErrorMsg = ""

	case clearWSStatusMsg:
		m.wsStatusMsg = ""

	case setWSStatusMsg:
		m.wsStatusMsg = msg.message
		// Auto-clear after 2 seconds
		cmd = tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
			return clearWSStatusMsg{}
		})

	case errorMsg:
		m.loading = false // Clear loading flag on error
		fullMsg := string(msg)
		m.fullErrorMsg = fullMsg
		// Truncate for footer display (max 100 chars)
		if len(fullMsg) > 100 {
			m.errorMsg = fullMsg[:97] + "..."
		} else {
			m.errorMsg = fullMsg
		}
		// Update response view to remove loading indicator
		m.updateResponseView()
		// Schedule auto-clear if configured
		profile := m.sessionMgr.GetActiveProfile()
		if profile != nil && profile.MessageTimeout != nil && *profile.MessageTimeout > 0 {
			timeout := time.Duration(*profile.MessageTimeout) * time.Second
			cmd = tea.Tick(timeout, func(time.Time) tea.Msg {
				return clearErrorMsg{}
			})
		}
	}

	// Update viewports based on current mode (only if no command was set)
	// Don't overwrite important commands like tea.Quit
	// IMPORTANT: Don't update viewports - we handle scrolling manually in key handlers
	// This prevents double-processing of keys and mouse events
	if cmd == nil {
		// Only update help viewport since it doesn't have manual key handling
		if m.mode == ModeHelp {
			// Filter out mouse messages
			if _, isMouseMsg := msg.(tea.MouseMsg); !isMouseMsg {
				m.helpView, cmd = m.helpView.Update(msg)
			}
		}
		// For all other viewports (response, modal), we handle keys manually in key handlers
		// so we don't update them here to avoid double-processing
	}

	return m, cmd
}

// View renders the TUI
func (m Model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	switch m.mode {
	case ModeHelp:
		return m.renderHelp()
	case ModeVariableList, ModeVariableAdd, ModeVariableEdit, ModeVariableDelete, ModeVariableOptions, ModeVariableManage, ModeVariableAlias:
		return m.renderVariableEditor()
	case ModeVariablePromptInteractive:
		return m.renderInteractiveVariablePrompt()
	case ModeHeaderList, ModeHeaderAdd, ModeHeaderEdit, ModeHeaderDelete:
		return m.renderHeaderEditor()
	case ModeProfileSwitch, ModeProfileCreate, ModeProfileEdit, ModeProfileDuplicate, ModeProfileDeleteConfirm:
		return m.renderProfileModal()
	case ModeDocumentation:
		return m.renderDocumentation()
	case ModeHistory:
		return m.renderHistory()
	case ModeHistoryClearConfirm:
		return m.renderHistoryClearConfirmation()
	case ModeAnalytics:
		return m.renderAnalytics()
	case ModeAnalyticsClearConfirm:
		return m.renderAnalyticsClearConfirmation()
	case ModeStressTestConfig:
		return m.renderStressTestConfig()
	case ModeStressTestLoadConfig:
		return m.renderStressTestLoadConfig()
	case ModeStressTestProgress:
		return m.renderStressTestProgress()
	case ModeStressTestResults:
		return m.renderStressTestResults()
	case ModeInspect:
		return m.renderInspect()
	case ModeOAuthConfig:
		return m.renderOAuthConfig()
	case ModeOAuthEdit:
		return m.renderOAuthEdit()
	case ModeRename:
		return m.renderRenameModal()
	case ModeEditorConfig:
		return m.renderEditorConfigModal()
	case ModeConfigView:
		return m.renderConfigView()
	case ModeDelete:
		return m.renderDeleteModal()
	case ModeConfirmExecution:
		return m.renderConfirmExecutionModal()
	case ModeErrorDetail:
		return m.renderErrorDetailModal()
	case ModeStatusDetail:
		return m.renderStatusDetailModal()
	case ModeShellErrors:
		return m.renderShellErrorsModal()
	case ModeCreateFile:
		return m.renderCreateFileModal()
	case ModeMockServer:
		return m.renderMockServer()
	case ModeProxyViewer:
		return m.renderProxyModal()
	case ModeProxyDetail:
		return m.renderProxyDetailModal()
	case ModeMRU:
		return m.renderMRUModal()
	case ModeDiff:
		return m.renderDiffModal()
	case ModeBodyOverride:
		return m.renderBodyOverrideModal()
	case ModeJSONPathHistory:
		return m.renderJSONPathHistoryModal()
	case ModeWebSocket:
		return m.renderWebSocketModal()
	default:
		return m.renderMain()
	}
}

// Custom message types
type fileListLoadedMsg struct {
	files []types.FileInfo
}

type requestExecutedMsg struct {
	result      *types.RequestResult
	warnings    []string // Unresolved variables
	shellErrors []string // Shell command errors
}

type oauthSuccessMsg struct {
	accessToken  string
	refreshToken string
	expiresIn    int
}

type historyLoadedMsg struct {
	entries []types.HistoryEntry
}

type analyticsLoadedMsg struct {
	stats []analytics.Stats
}

type promptInteractiveVarsMsg struct {
	varNames []string
}

type streamChunkMsg struct {
	chunk []byte
	done  bool
}

type versionCheckMsg struct {
	available      bool
	latestVersion  string
	url            string
	err            error
}

type clearStatusMsg struct{}
type clearErrorMsg struct{}
type clearWSStatusMsg struct{}
type setWSStatusMsg struct {
	message string
}

type chainCompleteMsg struct {
	success  bool
	message  string
	response *types.RequestResult
}

type mockServerTickMsg struct{}
type mockLogReceivedMsg struct{}
type proxyViewerTickMsg struct{}
type proxyLogReceivedMsg struct{}

// WebSocket message types
type wsMessageReceivedMsg struct {
	message *types.ReceivedMessage
}

type wsConnectionStatusMsg struct {
	status string // "connecting", "connected", "disconnected", "error"
}

type wsConnectionCompleteMsg struct {
	result *types.WebSocketResult
	err    error
}

type errorMsg string

// Helper methods for setting messages with optional timeout
func (m *Model) setStatusMessage(msg string) tea.Cmd {
	fullMsg := msg
	m.fullStatusMsg = fullMsg
	// Truncate for footer display (max 100 chars)
	if len(fullMsg) > 100 {
		m.statusMsg = fullMsg[:97] + "..."
	} else {
		m.statusMsg = fullMsg
	}

	// Check if profile has message timeout configured
	profile := m.sessionMgr.GetActiveProfile()
	if profile != nil && profile.MessageTimeout != nil && *profile.MessageTimeout > 0 {
		timeout := time.Duration(*profile.MessageTimeout) * time.Second
		return tea.Tick(timeout, func(time.Time) tea.Msg {
			return clearStatusMsg{}
		})
	}
	return nil
}

func (m *Model) setErrorMessage(msg string) tea.Cmd {
	fullMsg := msg
	m.fullErrorMsg = fullMsg
	// Truncate for footer display (max 100 chars)
	if len(fullMsg) > 100 {
		m.errorMsg = fullMsg[:97] + "..."
	} else {
		m.errorMsg = fullMsg
	}

	// Check if profile has message timeout configured
	profile := m.sessionMgr.GetActiveProfile()
	if profile != nil && profile.MessageTimeout != nil && *profile.MessageTimeout > 0 {
		timeout := time.Duration(*profile.MessageTimeout) * time.Second
		return tea.Tick(timeout, func(time.Time) tea.Msg {
			return clearErrorMsg{}
		})
	}
	return nil
}

// tickMockServer returns a command that will send mockServerTickMsg after a short delay
func (m *Model) tickMockServer() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
		return mockServerTickMsg{}
	})
}

// tickProxyViewer returns a command that will send proxyViewerTickMsg after a short delay
func (m *Model) tickProxyViewer() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
		return proxyViewerTickMsg{}
	})
}

// listenForProxyLogs waits for new proxy log notifications
func (m *Model) listenForProxyLogs() tea.Cmd {
	server := m.proxyServerState.GetServer()
	if server == nil {
		return nil
	}
	return func() tea.Msg {
		<-server.NotifyChannel()
		return proxyLogReceivedMsg{}
	}
}

// listenForMockLogs waits for new mock server log notifications
func (m *Model) listenForMockLogs() tea.Cmd {
	server := m.mockServerState.GetServer()
	if server == nil {
		return nil
	}
	return func() tea.Msg {
		<-server.NotifyChannel()
		return mockLogReceivedMsg{}
	}
}
