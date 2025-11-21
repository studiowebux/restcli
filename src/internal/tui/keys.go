package tui

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// handleKeyPress routes key presses based on current mode
func (m *Model) handleKeyPress(msg tea.KeyMsg) tea.Cmd {
	// Global keys (work in all modes)
	switch msg.String() {
	case "ctrl+c":
		return tea.Quit
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
	case ModeHeaderList, ModeHeaderAdd, ModeHeaderEdit, ModeHeaderDelete:
		return m.handleHeaderEditorKeys(msg)
	case ModeProfileSwitch, ModeProfileCreate, ModeProfileEdit:
		return m.handleProfileKeys(msg)
	case ModeDocumentation:
		return m.handleDocumentationKeys(msg)
	case ModeHistory:
		return m.handleHistoryKeys(msg)
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
	case ModeShellErrors:
		return m.handleShellErrorsKeys(msg)
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
	case "F":
		// Create new file
		m.mode = ModeCreateFile
		m.createFileInput = ""
		m.createFileType = 0 // Default to .http
		m.createFileCursor = 0
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
	case "S":
		return m.openSessionInEditor()

	// Escape - exit fullscreen, clear search, or clear errors/messages
	case "esc":
		if m.fullscreen {
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

	case "home":
		m.helpView.GotoTop()

	case "end":
		m.helpView.GotoBottom()
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

	// Item selection
	case "up", "k":
		if m.historyIndex > 0 {
			m.historyIndex--
			m.updateHistoryView() // Regenerate to update highlight
		}

	case "down", "j":
		if m.historyIndex < len(m.historyEntries)-1 {
			m.historyIndex++
			m.updateHistoryView() // Regenerate to update highlight
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
