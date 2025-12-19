package tui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/studiowebux/restcli/internal/analytics"
	"github.com/studiowebux/restcli/internal/history"
	"github.com/studiowebux/restcli/internal/jsonpath"
	"github.com/studiowebux/restcli/internal/session"
	"github.com/studiowebux/restcli/internal/stresstest"
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
)

// Model represents the TUI state
type Model struct {
	// Core state
	sessionMgr        *session.Manager
	analyticsManager  *analytics.Manager
	historyManager    *history.Manager
	bookmarkManager   *jsonpath.BookmarkManager
	mode              Mode
	version           string
	updateAvailable   bool
	latestVersion     string
	updateURL         string

	// File list
	files         []types.FileInfo
	fileIndex     int    // Current selected file
	fileOffset          int    // Scroll offset for file list
	searchQuery         string // Search query (files or response based on focus)
	searchMatches       []int  // Indices of files matching search OR line numbers in response
	searchIndex         int    // Current position in searchMatches
	searchInResponseCtx bool   // True if current search is in response context
	gotoInput           string // Goto line input

	// Collections and tag filtering
	allFiles          []types.FileInfo    // Unfiltered file list
	activeCollection  *types.Collection   // Currently active collection filter
	tagFilter         []string            // Active tag filters
	collectionIndex   int                 // Selected collection in browser

	// Request/Response
	currentRequests       []types.HttpRequest
	currentRequest        *types.HttpRequest
	currentResponse       *types.RequestResult
	responseView          viewport.Model
	responseContent       string // Full formatted response content for searching
	helpView              viewport.Model
	modalView             viewport.Model // For scrollable modal content
	historyPreviewView    viewport.Model // For history response preview in split view
	analyticsListView     viewport.Model // For analytics list in split view
	analyticsDetailView   viewport.Model // For analytics detail in split view

	// Streaming state
	streamingActive       bool                  // True when streaming is in progress
	streamedBody          string                // Accumulated streamed response body
	streamChannel         chan streamChunkMsg   // Channel for receiving stream chunks
	streamCancelFunc      context.CancelFunc    // Function to cancel the streaming request

	// Request cancellation (for regular non-streaming requests)
	requestCancelFunc     context.CancelFunc    // Function to cancel the current request

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

	// Profile edit state
	profileEditField            int    // 0=name, 1=workdir, 2=editor, 3=output, 4=history, 5=analytics
	profileEditName             string
	profileEditWorkdir          string
	profileEditEditor           string
	profileEditOutput           string
	profileEditHistoryEnabled   *bool  // nil=default, true/false=override
	profileEditAnalyticsEnabled *bool  // nil=default (false), true/false=override
	profileEditNamePos          int
	profileEditWorkdirPos       int
	profileEditEditorPos        int
	profileEditOutputPos        int

	// Documentation viewer state
	docCollapsed      map[int]bool
	docSelectedIdx    int                     // Currently selected item in doc viewer
	docItemCount      int                     // Cached total navigable items count
	docFieldTreeCache map[int][]DocField      // Cached field trees per response index
	docChildrenCache  map[int]map[string]bool // Cached hasChildren results per response index

	// History state
	historyEntries        []types.HistoryEntry
	historyIndex          int
	historyPreviewVisible bool // Toggle for showing/hiding response preview pane

	// Analytics state
	analyticsStats        []analytics.Stats
	analyticsIndex        int
	analyticsPreviewVisible bool // Toggle for showing/hiding stats detail pane
	analyticsGroupByPath  bool // Toggle between per-file and normalized-path grouping
	analyticsFocusedPane  string // "list" or "details" - which pane has focus in split view

	// Stress test state
	stressTestManager       *stresstest.Manager
	stressTestExecutor      *stresstest.Executor
	stressTestActiveRequest *types.HttpRequest // The request currently being tested
	stressTestConfigs       []*stresstest.Config
	stressTestConfigIndex   int
	stressTestRuns          []*stresstest.Run
	stressTestRunIndex      int
	stressTestListView      viewport.Model
	stressTestDetailView    viewport.Model
	stressTestFocusedPane   string // "list" or "details"
	stressTestConfigEdit          *stresstest.Config // Config being edited
	stressTestConfigField         int // Which field is being edited
	stressTestConfigInput         string // Input buffer for config fields
	stressTestConfigCursor        int // Cursor position in input
	stressTestConfigInsertMode    bool // True when actively typing (disables vim navigation)
	stressTestFilePickerActive    bool // True when file picker dropdown is shown
	stressTestFilePickerFiles     []types.FileInfo // Available files for selection
	stressTestFilePickerIndex     int // Selected index in file picker
	stressTestStopping            bool // True when test is being canceled/stopped

	// Rename state
	renameInput  string
	renameCursor int

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
	if m.stressTestManager != nil {
		if err := m.stressTestManager.Close(); err != nil {
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
		m.files = msg.files
		m.statusMsg = "Files loaded"
		// Reload current file's requests to reflect any changes
		m.loadRequestsFromCurrentFile()

	case chainCompleteMsg:
		m.loading = false
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
		m.requestCancelFunc = nil // Clear cancel function
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
		m.streamingActive = false
		m.streamChannel = nil
		m.streamCancelFunc = nil
		// Clear any previous errors since stream completed successfully
		m.errorMsg = ""
		m.fullErrorMsg = ""
		m.statusMsg = "Stream completed"

	case oauthSuccessMsg:
		m.statusMsg = fmt.Sprintf("OAuth successful! Token stored (expires in %d seconds)", msg.expiresIn)

	case historyLoadedMsg:
		m.historyEntries = msg.entries
		m.historyIndex = 0
		if len(msg.entries) > 0 {
			m.statusMsg = fmt.Sprintf("Loaded %d history entries", len(msg.entries))
		}
		m.updateHistoryView() // Update viewport content with loaded history

	case analyticsLoadedMsg:
		m.analyticsStats = msg.stats
		m.analyticsIndex = 0
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
		m.stressTestRuns = msg.runs
		m.stressTestRunIndex = 0
		if len(msg.runs) > 0 {
			m.statusMsg = fmt.Sprintf("Loaded %d stress test runs", len(msg.runs))
		}
		m.updateStressTestListView()
		m.stressTestDetailView.GotoTop() // Reset scroll position for new load

	case stressTestConfigsLoadedMsg:
		m.stressTestConfigs = msg.configs
		m.stressTestConfigIndex = 0
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
		if m.stressTestExecutor != nil {
			m.stressTestExecutor.Wait()
			m.stressTestExecutor = nil
		}
		m.stressTestActiveRequest = nil
		m.stressTestStopping = false
		m.statusMsg = "Stress test completed"
		m.mode = ModeStressTestResults
		return m, m.loadStressTestRuns()

	case stressTestStoppedMsg:
		// Stress test stopped by user
		m.stressTestExecutor = nil
		m.stressTestActiveRequest = nil
		m.stressTestStopping = false
		m.statusMsg = "Stress test cancelled"
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

	case clearStatusMsg:
		m.statusMsg = ""

	case clearErrorMsg:
		m.errorMsg = ""
		m.fullErrorMsg = ""

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
	case ModeMRU:
		return m.renderMRUModal()
	case ModeDiff:
		return m.renderDiffModal()
	case ModeBodyOverride:
		return m.renderBodyOverrideModal()
	case ModeJSONPathHistory:
		return m.renderJSONPathHistoryModal()
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

type chainCompleteMsg struct {
	success  bool
	message  string
	response *types.RequestResult
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
