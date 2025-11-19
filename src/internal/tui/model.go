package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
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
	ModeHeaderList
	ModeHeaderAdd
	ModeHeaderEdit
	ModeHeaderDelete
	ModeProfileSwitch
	ModeProfileCreate
	ModeDocumentation
	ModeHistory
	ModeHelp
	ModeInspect
	ModeRename
	ModeEditorConfig
	ModeOAuthConfig
	ModeOAuthEdit
	ModeConfigView
	ModeDelete
	ModeProfileEdit
	ModeVariableAlias
	ModeShellErrors
)

// Model represents the TUI state
type Model struct {
	// Core state
	sessionMgr *session.Manager
	mode       Mode

	// File list
	files         []types.FileInfo
	fileIndex     int    // Current selected file
	fileOffset    int    // Scroll offset for file list
	searchQuery   string // File search query
	searchMatches []int  // Indices of files matching search
	searchIndex   int    // Current position in searchMatches
	gotoInput     string // Goto line input

	// Request/Response
	currentRequests []types.HttpRequest
	currentRequest  *types.HttpRequest
	currentResponse *types.RequestResult
	responseView    viewport.Model
	helpView        viewport.Model
	modalView       viewport.Model // For scrollable modal content

	// UI state
	width        int
	height       int
	statusMsg    string
	errorMsg     string
	focusedPanel string // "sidebar" or "response"

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
	profileEditField     int    // 0=name, 1=workdir, 2=editor, 3=output
	profileEditName      string
	profileEditWorkdir   string
	profileEditEditor    string
	profileEditOutput    string
	profileEditNamePos   int
	profileEditWorkdirPos int
	profileEditEditorPos  int
	profileEditOutputPos  int

	// Documentation viewer state
	docCollapsed      map[int]bool
	docSelectedIdx    int                     // Currently selected item in doc viewer
	docItemCount      int                     // Cached total navigable items count
	docFieldTreeCache map[int][]DocField      // Cached field trees per response index
	docChildrenCache  map[int]map[string]bool // Cached hasChildren results per response index

	// History state
	historyEntries []types.HistoryEntry
	historyIndex   int

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
	showHeaders bool
	showBody    bool
	fullscreen  bool
	loading     bool
	gPressed    bool // Track if 'g' was pressed for 'gg' vim motion

	// Help search state
	helpSearchQuery  string
	helpSearchActive bool

	// Shell errors state
	shellErrors      []string
	shellErrorScroll int
}

// Init initializes the TUI
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		cmd = m.handleKeyPress(msg)

	// Mouse events disabled - keyboard only interface
	case tea.MouseMsg:
		// No-op - ignore all mouse events

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateViewport()

	case fileListLoadedMsg:
		m.files = msg.files
		m.statusMsg = "Files loaded"
		// Reload current file's requests to reflect any changes
		m.loadRequestsFromCurrentFile()

	case requestExecutedMsg:
		m.currentResponse = msg.result
		if len(msg.warnings) > 0 {
			m.statusMsg = fmt.Sprintf("Request completed (unresolved: %s)", strings.Join(msg.warnings, ", "))
		} else {
			m.statusMsg = "Request completed"
		}
		m.updateResponseView()
		// Auto-switch focus to response panel so user can immediately scroll
		m.focusedPanel = "response"
		// Show shell errors modal if any
		if len(msg.shellErrors) > 0 {
			m.shellErrors = msg.shellErrors
			m.shellErrorScroll = 0
			m.mode = ModeShellErrors
			m.updateShellErrorsView()
		}

	case oauthSuccessMsg:
		m.statusMsg = fmt.Sprintf("OAuth successful! Token stored (expires in %d seconds)", msg.expiresIn)

	case historyLoadedMsg:
		m.historyEntries = msg.entries
		m.historyIndex = 0
		if len(msg.entries) > 0 {
			m.statusMsg = fmt.Sprintf("Loaded %d history entries", len(msg.entries))
		}
		m.updateHistoryView() // Update viewport content with loaded history

	case errorMsg:
		m.errorMsg = string(msg)
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
	case ModeHeaderList, ModeHeaderAdd, ModeHeaderEdit, ModeHeaderDelete:
		return m.renderHeaderEditor()
	case ModeProfileSwitch, ModeProfileCreate, ModeProfileEdit:
		return m.renderProfileModal()
	case ModeDocumentation:
		return m.renderDocumentation()
	case ModeHistory:
		return m.renderHistory()
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
	case ModeShellErrors:
		return m.renderShellErrorsModal()
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

type errorMsg string
