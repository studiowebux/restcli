package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/atotto/clipboard"
	"github.com/studiowebux/restcli/internal/config"
	"github.com/studiowebux/restcli/internal/executor"
	"github.com/studiowebux/restcli/internal/history"
	"github.com/studiowebux/restcli/internal/oauth"
	"github.com/studiowebux/restcli/internal/parser"
	"github.com/studiowebux/restcli/internal/types"
)

// navigateFiles moves the file selection up or down
func (m *Model) navigateFiles(delta int) {
	if len(m.files) == 0 {
		return
	}

	m.fileIndex += delta

	// Wrap around (circular navigation as per TypeScript version)
	if m.fileIndex < 0 {
		m.fileIndex = len(m.files) - 1
	} else if m.fileIndex >= len(m.files) {
		m.fileIndex = 0
	}

	// Adjust scroll offset
	pageSize := m.getFileListHeight()
	if m.fileIndex < m.fileOffset {
		m.fileOffset = m.fileIndex
	} else if m.fileIndex >= m.fileOffset+pageSize {
		m.fileOffset = m.fileIndex - pageSize + 1
	}

	// Load requests from selected file
	m.loadRequestsFromCurrentFile()
}

// executeRequest executes the current request
func (m *Model) executeRequest() tea.Cmd {
	return func() tea.Msg {
		if m.currentRequest == nil {
			return errorMsg("No request selected")
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

		// Resolve variables
		resolver := parser.NewVariableResolver(profile.Variables, m.sessionMgr.GetSession().Variables, nil)
		resolvedRequest, err := resolver.ResolveRequest(&requestCopy)
		if err != nil {
			return errorMsg(fmt.Sprintf("Failed to resolve variables: %v", err))
		}

		// Execute request
		result, err := executor.Execute(resolvedRequest)
		if err != nil {
			return errorMsg(fmt.Sprintf("Failed to execute request: %v", err))
		}

		// Save to history
		if m.sessionMgr.IsHistoryEnabled() && len(m.files) > 0 {
			filePath := m.files[m.fileIndex].Path
			history.Save(filePath, resolvedRequest, result)
		}

		// Auto-extract tokens
		if result.Status >= 200 && result.Status < 300 {
			if token, err := parser.ExtractJSONToken(result.Body, "access_token"); err == nil {
				m.sessionMgr.SetSessionVariable("token", token)
			}
			if token, err := parser.ExtractJSONToken(result.Body, "token"); err == nil {
				m.sessionMgr.SetSessionVariable("token", token)
			}
		}

		return requestExecutedMsg{result: result}
	}
}

// refreshFiles reloads the file list AND profiles/session
func (m *Model) refreshFiles() tea.Cmd {
	return func() tea.Msg {
		// Reload session and profiles
		if err := m.sessionMgr.Load(); err != nil {
			return errorMsg(fmt.Sprintf("Failed to reload session: %v", err))
		}

		// Reload files
		files, err := loadFiles(m.sessionMgr)
		if err != nil {
			return errorMsg(fmt.Sprintf("Failed to load files: %v", err))
		}

		m.statusMsg = "Files and profiles reloaded"
		return fileListLoadedMsg{files: files}
	}
}

// openInEditor opens the current file in external editor
func (m *Model) openInEditor() tea.Cmd {
	if len(m.files) == 0 {
		m.errorMsg = "No file selected"
		return nil
	}

	filePath := m.files[m.fileIndex].Path
	profile := m.sessionMgr.GetActiveProfile()
	editor := profile.Editor
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = "vi"
	}

	// Use tea.ExecProcess to properly suspend/resume TUI
	c := exec.Command(editor, filePath)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			return errorMsg(fmt.Sprintf("Editor error: %v", err))
		}
		// Reload files after editing
		files, _ := loadFiles(m.sessionMgr)
		return fileListLoadedMsg{files: files}
	})
}

// duplicateFile duplicates the current file
func (m *Model) duplicateFile() tea.Cmd {
	return func() tea.Msg {
		if len(m.files) == 0 {
			return errorMsg("No file selected")
		}

		srcPath := m.files[m.fileIndex].Path
		dir := filepath.Dir(srcPath)
		base := filepath.Base(srcPath)
		ext := filepath.Ext(base)
		name := strings.TrimSuffix(base, ext)

		// Find a unique name
		dstPath := filepath.Join(dir, name+"_copy"+ext)
		counter := 2
		for {
			if _, err := os.Stat(dstPath); os.IsNotExist(err) {
				break
			}
			dstPath = filepath.Join(dir, fmt.Sprintf("%s_copy%d%s", name, counter, ext))
			counter++
		}

		// Copy file
		data, err := os.ReadFile(srcPath)
		if err != nil {
			return errorMsg(fmt.Sprintf("Failed to read file: %v", err))
		}

		if err := os.WriteFile(dstPath, data, 0644); err != nil {
			return errorMsg(fmt.Sprintf("Failed to write file: %v", err))
		}

		// Refresh file list
		files, _ := loadFiles(m.sessionMgr)
		return fileListLoadedMsg{files: files}
	}
}

// deleteFile deletes the current file
func (m *Model) deleteFile() tea.Cmd {
	return func() tea.Msg {
		if len(m.files) == 0 {
			return errorMsg("No file selected")
		}

		filePath := m.files[m.fileIndex].Path
		fileName := m.files[m.fileIndex].Name

		// Delete the file
		if err := os.Remove(filePath); err != nil {
			return errorMsg(fmt.Sprintf("Failed to delete file: %v", err))
		}

		// Adjust file index if we deleted the last file
		if m.fileIndex >= len(m.files)-1 && m.fileIndex > 0 {
			m.fileIndex--
		}

		m.statusMsg = fmt.Sprintf("Deleted: %s", fileName)
		m.mode = ModeNormal

		// Refresh file list
		files, _ := loadFiles(m.sessionMgr)
		return fileListLoadedMsg{files: files}
	}
}

// saveResponse saves the current response to a file with full metadata
func (m *Model) saveResponse() tea.Cmd {
	return func() tea.Msg {
		if m.currentResponse == nil {
			return errorMsg("No response to save")
		}

		// Generate filename with timestamp
		timestamp := time.Now().Format("20060102_150405")
		filename := fmt.Sprintf("response_%s.json", timestamp)
		if len(m.files) > 0 {
			base := filepath.Base(m.files[m.fileIndex].Name)
			baseName := strings.TrimSuffix(base, filepath.Ext(base))
			filename = fmt.Sprintf("%s_response_%s.json", baseName, timestamp)
		}

		// Create full response object with metadata
		// Try to parse body as JSON to avoid double-stringification
		var bodyData interface{}
		if err := json.Unmarshal([]byte(m.currentResponse.Body), &bodyData); err != nil {
			// Body is not JSON (HTML, CBOR, etc) - keep as string
			bodyData = m.currentResponse.Body
		}
		// else: Body is valid JSON - keep as parsed object

		// Prepare request details with resolved variables
		requestDetails := map[string]interface{}{}
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

			// Resolve variables
			resolver := parser.NewVariableResolver(profile.Variables, session.Variables, nil)
			resolvedRequest, err := resolver.ResolveRequest(&requestCopy)
			if err == nil && resolvedRequest != nil {
				// Use resolved values
				requestDetails["method"] = resolvedRequest.Method
				requestDetails["url"] = resolvedRequest.URL
				requestDetails["headers"] = resolvedRequest.Headers
				requestDetails["body"] = resolvedRequest.Body
			} else {
				// Fallback to unresolved values if resolution fails
				requestDetails["method"] = requestCopy.Method
				requestDetails["url"] = requestCopy.URL
				requestDetails["headers"] = requestCopy.Headers
				requestDetails["body"] = requestCopy.Body
			}
			// Note: profileVariables and sessionVariables removed - internal configs only
		}

		fullResponse := map[string]interface{}{
			"request":      requestDetails,
			"response": map[string]interface{}{
				"status":       m.currentResponse.Status,
				"statusText":   m.currentResponse.StatusText,
				"headers":      m.currentResponse.Headers,
				"body":         bodyData, // Already parsed if JSON, string if not
			},
			"duration":     m.currentResponse.Duration,
			"requestSize":  m.currentResponse.RequestSize,
			"responseSize": m.currentResponse.ResponseSize,
		}

		data, err := json.MarshalIndent(fullResponse, "", "  ")
		if err != nil {
			return errorMsg(fmt.Sprintf("Failed to marshal response: %v", err))
		}

		if err := os.WriteFile(filename, data, 0644); err != nil {
			return errorMsg(fmt.Sprintf("Failed to save response: %v", err))
		}

		// Return success message (will be rendered green by status bar)
		m.statusMsg = fmt.Sprintf("Response saved to %s", filename)
		m.errorMsg = ""
		return nil
	}
}

// copyToClipboard copies the FULL response body to clipboard
func (m *Model) copyToClipboard() tea.Cmd {
	return func() tea.Msg {
		if m.currentResponse == nil {
			return errorMsg("No response to copy")
		}

		// Copy the complete body, not truncated
		fullBody := m.currentResponse.Body

		if err := clipboard.WriteAll(fullBody); err != nil {
			return errorMsg(fmt.Sprintf("Failed to copy to clipboard: %v", err))
		}

		// Return success message (will be rendered green by status bar)
		m.statusMsg = "Response copied to clipboard"
		m.errorMsg = ""
		return nil
	}
}

// performSearch performs file search - finds all matches
func (m *Model) performSearch() {
	if m.searchQuery == "" {
		m.searchMatches = nil
		m.searchIndex = 0
		return
	}

	query := strings.ToLower(m.searchQuery)
	m.searchMatches = nil

	// Find all matching files
	for i, file := range m.files {
		if strings.Contains(strings.ToLower(file.Name), query) {
			m.searchMatches = append(m.searchMatches, i)
		}
	}

	if len(m.searchMatches) == 0 {
		m.errorMsg = "No matching files found"
		return
	}

	// Jump to first match
	m.searchIndex = 0
	m.fileIndex = m.searchMatches[0]
	m.adjustScrollOffset()
	m.loadRequestsFromCurrentFile()
	m.statusMsg = fmt.Sprintf("Match 1 of %d", len(m.searchMatches))
}

// adjustScrollOffset adjusts scroll offset to keep selected file visible
func (m *Model) adjustScrollOffset() {
	pageSize := m.getFileListHeight()
	if m.fileIndex < m.fileOffset {
		m.fileOffset = m.fileIndex
	} else if m.fileIndex >= m.fileOffset+pageSize {
		m.fileOffset = m.fileIndex - pageSize + 1
	}
}

// performGoto jumps to a hex line number
func (m *Model) performGoto() {
	if m.gotoInput == "" {
		return
	}

	// Parse hex input
	lineNum, err := strconv.ParseInt(m.gotoInput, 16, 64)
	if err != nil {
		m.errorMsg = "Invalid hex number"
		return
	}

	if lineNum < 0 || int(lineNum) >= len(m.files) {
		m.errorMsg = "Line number out of range"
		return
	}

	m.fileIndex = int(lineNum)
	m.fileOffset = int(lineNum)
	m.loadRequestsFromCurrentFile()
}

// loadHistory loads request history
func (m *Model) loadHistory() tea.Cmd {
	return func() tea.Msg {
		entries, err := history.Load()
		if err != nil {
			return errorMsg(fmt.Sprintf("Failed to load history: %v", err))
		}
		return historyLoadedMsg{entries: entries}
	}
}

// loadHistoryEntry loads a specific history entry into the response view
func (m *Model) loadHistoryEntry(index int) tea.Cmd {
	if index < 0 || index >= len(m.historyEntries) {
		return nil
	}

	entry := m.historyEntries[index]

	// Convert to RequestResult
	m.currentResponse = &types.RequestResult{
		Status:       entry.ResponseStatus,
		StatusText:   entry.ResponseStatusText,
		Headers:      entry.ResponseHeaders,
		Body:         entry.ResponseBody,
		Duration:     entry.Duration,
		RequestSize:  entry.RequestSize,
		ResponseSize: entry.ResponseSize,
		Error:        entry.Error,
	}

	// Convert to HttpRequest
	m.currentRequest = &types.HttpRequest{
		Name:    entry.RequestName,
		Method:  entry.Method,
		URL:     entry.URL,
		Headers: entry.Headers,
		Body:    entry.Body,
	}

	// Update response view
	m.updateResponseView()

	// Switch to normal mode and focus response panel
	m.mode = ModeNormal
	m.focusedPanel = "response"
	m.statusMsg = fmt.Sprintf("Loaded history entry from %s", entry.Timestamp[:19])

	return nil
}

// startOAuthFlow starts the OAuth PKCE flow
func (m *Model) startOAuthFlow() tea.Cmd {
	return func() tea.Msg {
		profile := m.sessionMgr.GetActiveProfile()

		// Check if OAuth is configured
		if profile.OAuth == nil || !profile.OAuth.Enabled {
			return errorMsg("OAuth is not configured. Press 'O' to configure.")
		}

		// Validate required fields - support both manual (authEndpoint) and auto-build (authUrl) modes
		hasManualMode := profile.OAuth.AuthEndpoint != ""
		hasAutoMode := profile.OAuth.AuthURL != ""

		if !hasManualMode && !hasAutoMode {
			return errorMsg("OAuth configuration incomplete. Either authEndpoint (complete URL) or authUrl (base URL) is required.")
		}
		if profile.OAuth.TokenURL == "" {
			return errorMsg("OAuth configuration incomplete. Token URL is required.")
		}
		if profile.OAuth.ClientID == "" {
			return errorMsg("OAuth configuration incomplete. Client ID is required.")
		}

		// Prepare OAuth config
		config := &oauth.Config{
			AuthURL:      profile.OAuth.AuthURL,      // For auto-build mode
			AuthEndpoint: profile.OAuth.AuthEndpoint, // For manual mode (complete URL)
			TokenURL:     profile.OAuth.TokenURL,
			ClientID:     profile.OAuth.ClientID,
			ClientSecret: profile.OAuth.ClientSecret,
			RedirectURL:  profile.OAuth.RedirectURI,
			Scope:        profile.OAuth.Scope,
			CallbackPort: profile.OAuth.WebhookPort,
		}

		// Start OAuth flow
		token, err := oauth.StartFlow(config)
		if err != nil {
			return errorMsg(fmt.Sprintf("OAuth flow failed: %v", err))
		}

		// Store token in session variable
		tokenKey := profile.OAuth.TokenStorageKey
		if tokenKey == "" {
			tokenKey = "token"
		}
		m.sessionMgr.SetSessionVariable(tokenKey, token.AccessToken)

		// Also store refresh token if available
		if token.RefreshToken != "" {
			m.sessionMgr.SetSessionVariable(tokenKey+"_refresh", token.RefreshToken)
		}

		return oauthSuccessMsg{
			accessToken:  token.AccessToken,
			refreshToken: token.RefreshToken,
			expiresIn:    token.ExpiresIn,
		}
	}
}

// openProfilesInEditor opens .profiles.json in external editor
func (m *Model) openProfilesInEditor() tea.Cmd {
	return m.openConfigFile(config.GetProfilesFilePath())
}

// openSessionInEditor opens .session.json in external editor
func (m *Model) openSessionInEditor() tea.Cmd {
	return m.openConfigFile(config.GetSessionFilePath())
}

// openConfigFile opens a config file in external editor (returns Cmd, not Msg)
func (m *Model) openConfigFile(filePath string) tea.Cmd {
	profile := m.sessionMgr.GetActiveProfile()
	editor := profile.Editor
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = "vi"
	}

	// Use tea.ExecProcess to properly suspend/resume TUI
	c := exec.Command(editor, filePath)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			return errorMsg(fmt.Sprintf("Editor error: %v", err))
		}
		// Reload session/profiles
		m.sessionMgr.Load()
		return errorMsg("Config reloaded")
	})
}

// loadRequestsFromCurrentFile loads requests from the currently selected file
func (m *Model) loadRequestsFromCurrentFile() {
	if len(m.files) == 0 || m.fileIndex >= len(m.files) {
		m.currentRequests = nil
		m.currentRequest = nil
		return
	}

	filePath := m.files[m.fileIndex].Path
	requests, err := parser.Parse(filePath)
	if err != nil {
		m.errorMsg = fmt.Sprintf("Failed to parse file: %v", err)
		return
	}

	m.currentRequests = requests
	if len(requests) > 0 {
		m.currentRequest = &requests[0]
	} else {
		m.currentRequest = nil
	}
}
