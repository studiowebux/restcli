package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/atotto/clipboard"
	"github.com/studiowebux/restcli/internal/analytics"
	"github.com/studiowebux/restcli/internal/chain"
	"github.com/studiowebux/restcli/internal/config"
	"github.com/studiowebux/restcli/internal/executor"
	"github.com/studiowebux/restcli/internal/filter"
	"github.com/studiowebux/restcli/internal/oauth"
	"github.com/studiowebux/restcli/internal/parser"
	"github.com/studiowebux/restcli/internal/types"
	"github.com/studiowebux/restcli/internal/version"
)

// navigateFiles moves the file selection up or down
func (m *Model) navigateFiles(delta int) {
	pageSize := m.getFileListHeight()
	m.fileExplorer.Navigate(delta, pageSize)

	// Load requests from selected file
	m.loadRequestsFromCurrentFile()
}

// executeRequest executes the current request
func (m *Model) executeRequest() tea.Cmd {
	// Prevent concurrent requests
	if m.loading {
		return func() tea.Msg {
			return errorMsg("Request already in progress")
		}
	}

	// Create local copy to prevent nil pointer issues if model state changes
	request := m.currentRequest
	if request == nil {
		return func() tea.Msg {
			return errorMsg("No request selected")
		}
	}

	// Check if request has dependencies - execute chain if needed
	if chain.HasDependencies(request) {
		return m.executeChain()
	}

	// IMMEDIATELY mark as loading to prevent concurrent execution
	// Clear it if we need to prompt for interactive vars or confirmation
	m.loading = true

	profile := m.sessionMgr.GetActiveProfile()

	// Check for interactive variables that need prompting (only if we haven't collected values yet)
	if m.interactiveVarValues == nil {
		interactiveVars := m.getInteractiveVariables()
		if len(interactiveVars) > 0 {
			// Clear loading flag since we're not executing yet (waiting for user input)
			m.loading = false
			// Trigger interactive prompt mode
			return func() tea.Msg {
				return promptInteractiveVarsMsg{varNames: interactiveVars}
			}
		}
	}

	// Check if request requires confirmation (and hasn't been confirmed yet)
	if request.RequiresConfirmation && !m.confirmationGiven {
		// Clear loading flag since we're not executing yet (waiting for confirmation)
		m.loading = false
		// Show confirmation modal
		m.mode = ModeConfirmExecution
		m.statusMsg = fmt.Sprintf("Confirm execution of: %s", request.Name)
		return nil
	}

	// Clear confirmation flag for next execution
	m.confirmationGiven = false

	// Clear any previous error, status messages, and response
	m.errorMsg = ""
	m.fullErrorMsg = ""
	m.statusMsg = "Executing request..."
	m.currentResponse = nil

	// Update response view to show loading indicator
	m.updateResponseView()

	// Create a copy of the request to avoid mutation
	requestCopy := *request
	requestCopy.Headers = make(map[string]string)

	// Merge headers into the copy
	for k, v := range profile.Headers {
		requestCopy.Headers[k] = v
	}
	for k, v := range request.Headers {
		requestCopy.Headers[k] = v
	}

	// Apply body override if set (ephemeral, one-time)
	if m.bodyOverride != "" {
		requestCopy.Body = m.bodyOverride
	}

	// Resolve variables (load system env vars for {{env.VAR_NAME}} support)
	// Include any interactive variable values collected
	cliVars := m.interactiveVarValues
	resolver := parser.NewVariableResolver(profile.Variables, m.sessionMgr.GetSession().Variables, cliVars, parser.LoadSystemEnv())
	resolvedRequest, err := resolver.ResolveRequest(&requestCopy)
	if err != nil {
		m.loading = false // Clear loading flag on error
		m.updateResponseView() // Update view to remove loading indicator
		return func() tea.Msg {
			return errorMsg(fmt.Sprintf("Failed to resolve variables: %v", err))
		}
	}

	// Clear body override after using it (one-time use)
	m.bodyOverride = ""

	// Get warnings for unresolved variables (short, for status bar)
	warnings := resolver.GetUnresolvedVariables()
	shellErrs := resolver.GetShellErrors()

	// Merge TLS config: request-level overrides profile-level
	// Resolve profile TLS config if present
	var tlsConfig *types.TLSConfig
	if profile.TLS != nil {
		resolvedProfileTLS := &types.TLSConfig{
			InsecureSkipVerify: profile.TLS.InsecureSkipVerify,
		}
		if profile.TLS.CertFile != "" {
			certFile, _ := resolver.Resolve(profile.TLS.CertFile)
			resolvedProfileTLS.CertFile = certFile
		}
		if profile.TLS.KeyFile != "" {
			keyFile, _ := resolver.Resolve(profile.TLS.KeyFile)
			resolvedProfileTLS.KeyFile = keyFile
		}
		if profile.TLS.CAFile != "" {
			caFile, _ := resolver.Resolve(profile.TLS.CAFile)
			resolvedProfileTLS.CAFile = caFile
		}
		tlsConfig = resolvedProfileTLS
	}
	// Request-level TLS overrides profile-level (already resolved in resolvedRequest)
	if resolvedRequest.TLS != nil {
		tlsConfig = resolvedRequest.TLS
	}

	// Check if this is a streaming request
	if resolvedRequest.Streaming {
		m.statusMsg = fmt.Sprintf("Starting streaming request: %s", resolvedRequest.Name)
		return m.executeStreamingRequest(resolvedRequest, tlsConfig, warnings, shellErrs, profile)
	}

	// Regular non-streaming execution
	m.statusMsg = fmt.Sprintf("Executing request: %s", resolvedRequest.Name)
	return m.executeRegularRequest(resolvedRequest, tlsConfig, warnings, shellErrs, profile)
}

// executeWebSocket opens WebSocket modal and loads predefined messages
func (m *Model) executeWebSocket() tea.Cmd {
	currentFile := m.fileExplorer.GetCurrentFile()
	if currentFile == nil {
		return func() tea.Msg {
			return errorMsg("No file selected")
		}
	}

	filePath := currentFile.Path

	// Parse the .ws file
	wsReq, err := parser.ParseWebSocketFile(filePath)
	if err != nil {
		return func() tea.Msg {
			return errorMsg(fmt.Sprintf("Failed to parse WebSocket file: %v", err))
		}
	}

	// Initialize WebSocket state (wsState already initialized)
	// Only clear messages if switching to a different WebSocket URL
	if m.wsURL != wsReq.URL {
		m.wsMessages = []types.ReceivedMessage{}
	}
	m.wsConnectionStatus = "not connected"
	m.wsURL = wsReq.URL
	m.wsError = ""
	m.wsPredefinedMessages = wsReq.Messages // Load all predefined messages

	// Filter to only "send" messages for the menu
	m.wsSendableMessages = []types.WebSocketMessage{}
	for _, msg := range wsReq.Messages {
		if msg.Direction == "send" {
			m.wsSendableMessages = append(m.wsSendableMessages, msg)
		}
	}

	m.wsSelectedMessageIndex = 0
	m.wsFocusedPane = "menu"       // Start with menu focused
	m.wsPendingMessageIndex = -1   // No pending message initially
	m.mode = ModeWebSocket

	// Initialize viewport content
	m.updateWebSocketViews(m.width, m.height)

	return nil
}

// sendWebSocketMessage sends a predefined message from the menu
func (m *Model) sendWebSocketMessage(msgIndex int) tea.Cmd {
	if msgIndex < 0 || msgIndex >= len(m.wsSendableMessages) {
		return func() tea.Msg {
			return errorMsg("Invalid message index")
		}
	}

	// If not connected, initiate connection first and store message to send after
	if !m.wsState.IsActive() {
		m.wsPendingMessageIndex = msgIndex
		return m.connectWebSocket()
	}

	// Get the selected send message
	msg := m.wsSendableMessages[msgIndex]

	// Send via channel
	if m.wsSendChannel != nil {
		go func() {
			select {
			case m.wsSendChannel <- msg.Content:
			default:
				// Channel full or closed
			}
		}()
	}

	return nil
}

// connectWebSocket establishes WebSocket connection
func (m *Model) connectWebSocket() tea.Cmd {
	profile := m.sessionMgr.GetActiveProfile()
	if profile == nil {
		return func() tea.Msg {
			return errorMsg("No active profile")
		}
	}

	// Parse the .ws file again to get full request
	currentFile := m.fileExplorer.GetCurrentFile()
	if currentFile == nil {
		return func() tea.Msg {
			return errorMsg("No file selected")
		}
	}

	filePath := currentFile.Path
	wsReq, err := parser.ParseWebSocketFile(filePath)
	if err != nil {
		return func() tea.Msg {
			return errorMsg(fmt.Sprintf("Failed to parse WebSocket file: %v", err))
		}
	}

	// Set connecting status
	m.wsConnectionStatus = "connecting"

	// Create channels for WebSocket communication
	m.wsMessageChannel = make(chan types.ReceivedMessage, WebSocketMessageBuffer)
	m.wsSendChannel = make(chan string, WebSocketSendBuffer)

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	m.wsState.Start(cancel)

	// Start persistent WebSocket connection in a goroutine
	go func() {
		msgChan := m.wsMessageChannel
		sendChan := m.wsSendChannel
		defer cancel()
		defer close(msgChan)

		// Create callback that sends messages to the channel
		callback := func(message *types.ReceivedMessage, done bool) {
			if done {
				return
			}
			if message != nil {
				select {
				case msgChan <- *message:
					// Message sent successfully
				default:
					// Channel full - drop message and increment counter
					m.wsState.IncrementDropped()
				}
			}
		}

		// Create variable resolver for message resolution
		session := m.sessionMgr.GetSession()
		resolver := parser.NewVariableResolver(
			profile.Variables,
			session.Variables,
			nil, // No CLI vars for WebSocket
			parser.LoadSystemEnv(),
		)

		// Merge headers: profile headers first, then .ws file headers (which override)
		mergedHeaders := make(map[string]string)
		// Copy profile headers first
		if profile.Headers != nil {
			for k, v := range profile.Headers {
				mergedHeaders[k] = v
			}
		}
		// Copy .ws file headers second (these override profile headers)
		if wsReq.Headers != nil {
			for k, v := range wsReq.Headers {
				mergedHeaders[k] = v
			}
		}

		// TODO(#TODO-001): OAuth auto-injection - fetch token if profile.OAuth.Enabled and inject as Authorization header
		// For now, users can manually add "Authorization: Bearer <token>" in profile or .ws file headers
		// See TODO.md for details and implementation plan

		// Execute PERSISTENT WebSocket connection
		err := executor.ExecuteWebSocketInteractive(
			ctx,
			wsReq.URL,
			mergedHeaders,
			wsReq.Subprotocols,
			profile.TLS,
			resolver,
			sendChan,
			callback,
		)

		// Send completion message if error
		if err != nil {
			msgChan <- types.ReceivedMessage{
				Type:      "system",
				Content:   fmt.Sprintf("Error: %v", err),
				Timestamp: time.Now().Format(time.RFC3339),
				Direction: "system",
			}
		}
	}()

	// Send initial status and wait for first message
	return tea.Batch(
		func() tea.Msg {
			return wsConnectionStatusMsg{status: "connecting"}
		},
		m.waitForWsMessage(),
	)
}

// waitForWsMessage waits for the next WebSocket message from the channel
func (m *Model) waitForWsMessage() tea.Cmd {
	return func() tea.Msg {
		if m.wsMessageChannel == nil {
			return errorMsg("No active WebSocket connection")
		}
		msg, ok := <-m.wsMessageChannel
		if !ok {
			return wsConnectionCompleteMsg{result: nil, err: nil}
		}
		return wsMessageReceivedMsg{message: &msg}
	}
}

// exportWebSocketMessages exports the message history to a JSON file
func (m *Model) exportWebSocketMessages() tea.Cmd {
	return func() tea.Msg {
		if len(m.wsMessages) == 0 {
			return setWSStatusMsg{message: "No messages to export"}
		}

		// Generate filename with timestamp
		timestamp := time.Now().Format("20060102-150405")
		filename := fmt.Sprintf("websocket-messages-%s.json", timestamp)

		// Marshal messages to JSON
		data, err := json.MarshalIndent(m.wsMessages, "", "  ")
		if err != nil {
			return setWSStatusMsg{message: fmt.Sprintf("Export failed: %v", err)}
		}

		// Write to file
		if err := os.WriteFile(filename, data, 0644); err != nil {
			return setWSStatusMsg{message: fmt.Sprintf("Export failed: %v", err)}
		}

		return setWSStatusMsg{message: fmt.Sprintf("Exported %d messages to %s", len(m.wsMessages), filename)}
	}
}

// executeRegularRequest executes a standard (non-streaming) HTTP request with cancellation support
func (m *Model) executeRegularRequest(resolvedRequest *types.HttpRequest, tlsConfig *types.TLSConfig, warnings, shellErrs []string, profile *types.Profile) tea.Cmd {
	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	m.requestState.SetCancel(cancel)

	return func() tea.Msg {
		// Create a channel for the result
		type result struct {
			data *types.RequestResult
			err  error
		}
		resultChan := make(chan result, 1)

		// Execute request in goroutine
		go func() {
			res, err := executor.ExecuteWithContext(ctx, resolvedRequest, tlsConfig, profile)
			resultChan <- result{data: res, err: err}
		}()

		// Wait for either result or cancellation
		select {
		case <-ctx.Done():
			// Request was cancelled
			return errorMsg("Request cancelled by user")
		case res := <-resultChan:
			// Request completed
			if res.err != nil {
				// Track network errors in analytics
				shouldSaveAnalytics := false
				if profile != nil && profile.AnalyticsEnabled != nil {
					shouldSaveAnalytics = *profile.AnalyticsEnabled
				}
				if shouldSaveAnalytics && m.analyticsManager != nil {
					currentFile := m.fileExplorer.GetCurrentFile()
					if currentFile != nil {
						filePath := currentFile.Path
						normalizedPath := normalizePath(resolvedRequest.URL)

						entry := analytics.Entry{
							FilePath:       filePath,
							NormalizedPath: normalizedPath,
							Method:         resolvedRequest.Method,
							StatusCode:     0, // 0 indicates network error (no HTTP response)
							RequestSize:    int64(len(resolvedRequest.Body)),
							ResponseSize:   0,
							DurationMs:     0,
							ErrorMessage:   res.err.Error(),
							Timestamp:      time.Now(),
							ProfileName:    profile.Name,
						}

						_ = m.analyticsManager.Save(entry)
					}
				}

				return errorMsg(categorizeError(res.err))
			}

			result := res.data

			// Apply filter and query
			filterExpr := resolvedRequest.Filter
			if filterExpr == "" {
				filterExpr = profile.DefaultFilter
			}
			queryExpr := resolvedRequest.Query
			if queryExpr == "" {
				queryExpr = profile.DefaultQuery
			}

			if filterExpr != "" || queryExpr != "" {
				filteredBody, err := filter.Apply(result.Body, filterExpr, queryExpr)
				if err != nil {
					_ = err
				} else {
					result.Body = filteredBody
				}
			}

			// Parse escape sequences
			if resolvedRequest.ParseEscapes {
				result.Body = executor.ParseEscapeSequences(result.Body)
			}

			// Save to history
			shouldSaveHistory := m.sessionMgr.IsHistoryEnabled()
			if profile != nil && profile.HistoryEnabled != nil {
				shouldSaveHistory = *profile.HistoryEnabled
			}
			if shouldSaveHistory && m.historyManager != nil {
				currentFile := m.fileExplorer.GetCurrentFile()
				if currentFile != nil {
					filePath := currentFile.Path
					_ = m.historyManager.Save(filePath, profile.Name, resolvedRequest, result)
				}
			}

			// Save to analytics
			shouldSaveAnalytics := false
			if profile != nil && profile.AnalyticsEnabled != nil {
				shouldSaveAnalytics = *profile.AnalyticsEnabled
			}
			if shouldSaveAnalytics && m.analyticsManager != nil {
				currentFile := m.fileExplorer.GetCurrentFile()
				if currentFile != nil {
					filePath := currentFile.Path
					normalizedPath := normalizePath(resolvedRequest.URL)

					entry := analytics.Entry{
						FilePath:       filePath,
						NormalizedPath: normalizedPath,
						Method:         resolvedRequest.Method,
						StatusCode:     result.Status,
						RequestSize:    int64(len(resolvedRequest.Body)),
						ResponseSize:   int64(len(result.Body)),
						DurationMs:     result.Duration,
						Timestamp:      time.Now(),
						ProfileName:    profile.Name,
					}

					_ = m.analyticsManager.Save(entry) // Ignore errors to not interrupt the flow
				}
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

			return requestExecutedMsg{result: result, warnings: warnings, shellErrors: shellErrs}
		}
	}
}

// executeStreamingRequest starts a streaming request in a goroutine with real-time updates
func (m *Model) executeStreamingRequest(resolvedRequest *types.HttpRequest, tlsConfig *types.TLSConfig, warnings, shellErrs []string, profile *types.Profile) tea.Cmd {
	// Create a channel for streaming chunks
	m.streamChannel = make(chan streamChunkMsg, StreamMessageBuffer)
	m.streamedBody = ""

	// Create a cancellable context for the request
	ctx, cancel := context.WithCancel(context.Background())
	m.streamState.Start(cancel)

	// Start the request in a goroutine
	go func() {
		chunkChan := m.streamChannel
		defer cancel()
		defer close(chunkChan)

		// Execute with streaming callback - sends chunks as they arrive
		_, err := executor.ExecuteWithStreaming(ctx, resolvedRequest, tlsConfig, profile, func(chunk []byte, done bool) {
			chunkChan <- streamChunkMsg{chunk: chunk, done: done}
		})

		if err != nil {
			chunkChan <- streamChunkMsg{chunk: []byte(fmt.Sprintf("Error: %v", err)), done: true}
			return
		}

		// Streaming complete - result is accumulated in streamedBody by the Update handler
	}()

	// Return a command that waits for the first chunk
	return m.waitForStreamChunk()
}

// waitForStreamChunk returns a Cmd that waits for the next chunk from the stream channel
func (m *Model) waitForStreamChunk() tea.Cmd {
	return func() tea.Msg {
		if m.streamChannel == nil {
			return errorMsg("No active stream")
		}
		msg, ok := <-m.streamChannel
		if !ok {
			return errorMsg("Stream closed unexpectedly")
		}
		return msg
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

// executeChain executes a request chain (request with dependencies)
func (m *Model) executeChain() tea.Cmd {
	file := m.fileExplorer.GetCurrentFile()
	if file == nil {
		return func() tea.Msg {
			return errorMsg("No file selected")
		}
	}

	currentFile := file.Path
	profile := m.sessionMgr.GetActiveProfile()

	// Build dependency graph
	graph := chain.NewGraph(profile.Workdir)
	if err := graph.BuildGraph(currentFile); err != nil {
		return func() tea.Msg {
			return errorMsg(fmt.Sprintf("Failed to build dependency graph: %v", err))
		}
	}

	// Get execution order
	executionOrder, err := graph.GetExecutionOrder(currentFile)
	if err != nil {
		return func() tea.Msg {
			return errorMsg(fmt.Sprintf("Failed to determine execution order: %v", err))
		}
	}

	// Mark as loading
	m.loading = true
	m.statusMsg = fmt.Sprintf("Executing chain: %d requests", len(executionOrder))
	m.updateResponseView()

	// Create a cancellable context for the entire chain
	ctx, cancel := context.WithCancel(context.Background())
	m.requestState.SetCancel(cancel)

	// Execute chain asynchronously
	return func() tea.Msg {
		// Execute each request in order
		for i, filePath := range executionOrder {
			// Check for cancellation before each request
			select {
			case <-ctx.Done():
				return chainCompleteMsg{
					success: false,
					message: fmt.Sprintf("Chain cancelled after %d/%d requests", i, len(executionOrder)),
				}
			default:
				// Continue with request execution
			}
			// Parse the file
			requests, err := parser.Parse(filePath)
			if err != nil {
				return chainCompleteMsg{
					success: false,
					message: fmt.Sprintf("Failed to parse %s: %v", filepath.Base(filePath), err),
				}
			}

			if len(requests) == 0 {
				return chainCompleteMsg{
					success: false,
					message: fmt.Sprintf("No requests found in %s", filepath.Base(filePath)),
				}
			}

			req := &requests[0]

			// Resolve variables
			resolver := parser.NewVariableResolver(profile.Variables, m.sessionMgr.GetSession().Variables, nil, parser.LoadSystemEnv())
			resolvedRequest, err := resolver.ResolveRequest(req)
			if err != nil {
				return chainCompleteMsg{
					success: false,
					message: fmt.Sprintf("Failed to resolve variables in %s: %v", filepath.Base(filePath), err),
				}
			}

			// Merge TLS config
			var tlsConfig *types.TLSConfig
			if profile.TLS != nil {
				resolvedProfileTLS := &types.TLSConfig{
					InsecureSkipVerify: profile.TLS.InsecureSkipVerify,
				}
				if profile.TLS.CertFile != "" {
					certFile, _ := resolver.Resolve(profile.TLS.CertFile)
					resolvedProfileTLS.CertFile = certFile
				}
				if profile.TLS.KeyFile != "" {
					keyFile, _ := resolver.Resolve(profile.TLS.KeyFile)
					resolvedProfileTLS.KeyFile = keyFile
				}
				if profile.TLS.CAFile != "" {
					caFile, _ := resolver.Resolve(profile.TLS.CAFile)
					resolvedProfileTLS.CAFile = caFile
				}
				tlsConfig = resolvedProfileTLS
			}
			if resolvedRequest.TLS != nil {
				tlsConfig = resolvedRequest.TLS
			}

			// Execute request with cancellation support
			result, err := executor.ExecuteWithContext(ctx, resolvedRequest, tlsConfig, profile)
			if err != nil {
				return chainCompleteMsg{
					success: false,
					message: fmt.Sprintf("Request %d/%d (%s) failed: %s", i+1, len(executionOrder), filepath.Base(filePath), categorizeError(err)),
				}
			}

			// Save to history
			shouldSaveHistory := m.sessionMgr.IsHistoryEnabled()
			if profile != nil && profile.HistoryEnabled != nil {
				shouldSaveHistory = *profile.HistoryEnabled
			}
			if shouldSaveHistory && m.historyManager != nil {
				_ = m.historyManager.Save(filePath, profile.Name, resolvedRequest, result)
			}

			// Track analytics if enabled
			if profile.AnalyticsEnabled != nil && *profile.AnalyticsEnabled && m.analyticsManager != nil {
				entry := analytics.Entry{
					FilePath:       filePath,
					NormalizedPath: resolvedRequest.URL,
					Method:         resolvedRequest.Method,
					StatusCode:     result.Status,
					RequestSize:    int64(result.RequestSize),
					ResponseSize:   int64(result.ResponseSize),
					DurationMs:     result.Duration,
					Timestamp:      time.Now(),
					ProfileName:    profile.Name,
				}
				_ = m.analyticsManager.Save(entry)
			}

			// Extract variables if specified
			if chain.HasExtractions(req) {
				extracted, err := chain.ExtractVariables(req, result.Body)
				if err != nil {
					return chainCompleteMsg{
						success: false,
						message: fmt.Sprintf("Failed to extract variables from %s: %v", filepath.Base(filePath), err),
					}
				}

				// Store extracted variables in session
				for varName, varValue := range extracted {
					m.sessionMgr.SetSessionVariable(varName, varValue)
				}
			}

			// If this is the last request, save the result
			if i == len(executionOrder)-1 {
				return chainCompleteMsg{
					success:  true,
					message:  fmt.Sprintf("Chain completed: %d requests executed", len(executionOrder)),
					response: result,
				}
			}
		}

		return chainCompleteMsg{
			success: true,
			message: fmt.Sprintf("Chain completed: %d requests executed", len(executionOrder)),
		}
	}
}

// openInEditor opens the current file in external editor
func (m *Model) openInEditor() tea.Cmd {
	currentFile := m.fileExplorer.GetCurrentFile()
	if currentFile == nil {
		return m.setErrorMessage("No file selected")
	}

	filePath := currentFile.Path
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
		currentFile := m.fileExplorer.GetCurrentFile()
		if currentFile == nil {
			return errorMsg("No file selected")
		}

		srcPath := currentFile.Path
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

		if err := os.WriteFile(dstPath, data, config.FilePermissions); err != nil {
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
		currentFile := m.fileExplorer.GetCurrentFile()
		if currentFile == nil {
			return errorMsg("No file selected")
		}

		filePath := currentFile.Path
		fileName := currentFile.Name

		// Delete the file
		if err := os.Remove(filePath); err != nil {
			return errorMsg(fmt.Sprintf("Failed to delete file: %v", err))
		}

		// File index will be adjusted when files are reloaded

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
		currentFile := m.fileExplorer.GetCurrentFile()
		if currentFile != nil {
			base := filepath.Base(currentFile.Name)
			baseName := strings.TrimSuffix(base, filepath.Ext(base))
			filename = fmt.Sprintf("%s_response_%s.json", baseName, timestamp)
		}

		// Create full response object with metadata
		// Use filtered response if active, otherwise use original
		var bodySource string
		if m.filterActive && m.filteredResponse != "" {
			bodySource = m.filteredResponse
		} else {
			bodySource = m.currentResponse.Body
		}

		// Try to parse body as JSON to avoid double-stringification
		var bodyData interface{}
		if err := json.Unmarshal([]byte(bodySource), &bodyData); err != nil {
			// Body is not JSON (HTML, CBOR, etc) - keep as string
			bodyData = bodySource
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
			resolver := parser.NewVariableResolver(profile.Variables, session.Variables, nil, parser.LoadSystemEnv())
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

		// Add filter note if a filter was applied
		if m.filterActive && m.filteredResponse != "" {
			fullResponse["filter"] = m.filterInput
		}

		data, err := json.MarshalIndent(fullResponse, "", "  ")
		if err != nil {
			return errorMsg(fmt.Sprintf("Failed to marshal response: %v", err))
		}

		if err := os.WriteFile(filename, data, config.FilePermissions); err != nil {
			return errorMsg(fmt.Sprintf("Failed to save response: %v", err))
		}

		// Return success message (will be rendered green by status bar)
		m.statusMsg = fmt.Sprintf("Response saved to %s", filename)
		m.errorMsg = ""
		return nil
	}
}

// copyToClipboard copies the FULL response body or error to clipboard
func (m *Model) copyToClipboard() tea.Cmd {
	return func() tea.Msg {
		var contentToCopy string

		if m.currentResponse == nil {
			// No response, try to copy error message if available
			if m.fullErrorMsg != "" {
				contentToCopy = m.fullErrorMsg
			} else if m.errorMsg != "" {
				contentToCopy = m.errorMsg
			} else {
				return errorMsg("No response or error to copy")
			}
		} else {
			// If response has an error field, copy that; otherwise copy the body
			if m.currentResponse.Error != "" {
				contentToCopy = m.currentResponse.Error
			} else {
				// If filter is active, copy the filtered response
				if m.filterActive && m.filteredResponse != "" {
					contentToCopy = m.filteredResponse
				} else {
					contentToCopy = m.currentResponse.Body
				}
			}
		}

		if err := clipboard.WriteAll(contentToCopy); err != nil {
			return errorMsg(fmt.Sprintf("Failed to copy to clipboard: %v", err))
		}

		// Return success message (will be rendered green by status bar)
		m.errorMsg = ""
		return m.setStatusMessage("Response copied to clipboard")
	}
}

// performSearch performs context-aware search (files or response based on focus)
func (m *Model) performSearch() {
	// Use searchInput from ModeSearch (not the currently active search)
	query := m.searchInput
	if query == "" {
		wasSearchingResponse := m.searchInResponseCtx
		m.fileExplorer.ClearSearch()
		m.responseSearchMatches = nil
		m.responseSearchIndex = 0
		m.searchInResponseCtx = false
		// Clear highlighting from response if we were searching there
		if wasSearchingResponse && m.currentResponse != nil {
			m.updateResponseView()
		}
		return
	}

	// Determine context: searching in response or files
	if m.focusedPanel == "response" && m.currentResponse != nil {
		m.searchInResponse()
	} else {
		m.searchInFiles()
	}
}

// isRegexPattern detects if a pattern looks like regex
func isRegexPattern(s string) bool {
	regexChars := ".*+?[]{}()|^$\\"
	for _, char := range regexChars {
		if strings.ContainsRune(s, char) {
			return true
		}
	}
	return false
}

// searchInFiles searches in file names with optional regex support
func (m *Model) searchInFiles() {
	m.searchInResponseCtx = false
	m.errorMsg = "" // Clear any previous errors

	// Use searchInput (from ModeSearch) for new searches
	query := m.searchInput
	pageSize := m.getFileListHeight()

	matchCount, errMsg := m.fileExplorer.Search(query, pageSize)

	if errMsg != "" {
		m.errorMsg = errMsg
		return
	}

	// Load requests from selected file
	m.loadRequestsFromCurrentFile()

	mode := "search"
	if isRegexPattern(query) {
		mode = "regex"
	}
	m.statusMsg = fmt.Sprintf("[Files] Match 1 of %d (%s)", matchCount, mode)
}

// searchInFilesSubstring performs case-insensitive substring search
// Note: This is now handled by FileExplorerState.Search internally
func (m *Model) searchInFilesSubstring() {
	m.searchInFiles()
}

// searchInResponse searches in response body
func (m *Model) searchInResponse() {
	m.responseSearchMatches = nil
	m.searchInResponseCtx = true
	m.errorMsg = "" // Clear any previous errors

	// Get full response content (not just visible viewport)
	content := m.responseContent
	if content == "" {
		m.errorMsg = "No response to search"
		return
	}
	lines := strings.Split(content, "\n")

	// Use searchInput (from ModeSearch) for new searches
	query := m.searchInput

	// Auto-detect regex
	useRegex := isRegexPattern(query)

	if useRegex {
		// Try regex search
		pattern, err := regexp.Compile(query)
		if err != nil {
			// Fall back to substring search if regex is invalid
			m.searchInResponseSubstring(lines)
			return
		}

		for lineNum, line := range lines {
			// Strip ANSI codes for searching
			cleanLine := stripANSI(line)
			if pattern.MatchString(cleanLine) {
				m.responseSearchMatches = append(m.responseSearchMatches, lineNum)
			}
		}

		if len(m.responseSearchMatches) == 0 {
			m.errorMsg = "No matches found in response"
			return
		}

		m.responseSearchIndex = 0
		m.responseView.SetYOffset(m.centerLineInViewport(m.responseSearchMatches[0]))
		m.statusMsg = fmt.Sprintf("[Response] Match 1 of %d (regex)", len(m.responseSearchMatches))
		m.updateResponseView() // Re-render with highlighting
	} else {
		m.searchInResponseSubstring(lines)
	}
}

// searchInResponseSubstring performs case-insensitive substring search in response
func (m *Model) searchInResponseSubstring(lines []string) {
	query := m.searchInput
	queryLower := strings.ToLower(query)

	for lineNum, line := range lines {
		cleanLine := stripANSI(line)
		if strings.Contains(strings.ToLower(cleanLine), queryLower) {
			m.responseSearchMatches = append(m.responseSearchMatches, lineNum)
		}
	}

	if len(m.responseSearchMatches) == 0 {
		m.errorMsg = "No matches found in response"
		return
	}

	m.responseSearchIndex = 0
	m.responseView.SetYOffset(m.centerLineInViewport(m.responseSearchMatches[0]))
	m.statusMsg = fmt.Sprintf("[Response] Match 1 of %d (text)", len(m.responseSearchMatches))
	m.updateResponseView() // Re-render with highlighting
}

// centerLineInViewport calculates the Y offset to center a line in the viewport
func (m *Model) centerLineInViewport(lineNum int) int {
	// Center the line by setting offset to lineNum - half viewport height
	offset := lineNum - (m.responseView.Height / 2)
	if offset < 0 {
		offset = 0
	}
	return offset
}

// stripANSI removes ANSI color codes from a string
func stripANSI(s string) string {
	ansiPattern := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return ansiPattern.ReplaceAllString(s, "")
}

// adjustScrollOffset adjusts scroll offset to keep selected file visible
// Note: This is now handled by FileExplorerState.Navigate and FileExplorerState.AdjustScrollOffset
func (m *Model) adjustScrollOffset() {
	pageSize := m.getFileListHeight()
	m.fileExplorer.AdjustScrollOffset(pageSize)
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

	files := m.fileExplorer.GetFiles()
	if lineNum < 0 || int(lineNum) >= len(files) {
		m.errorMsg = "Line number out of range"
		return
	}

	// Navigate to the specified file index
	currentIdx := m.fileExplorer.GetCurrentIndex()
	delta := int(lineNum) - currentIdx
	pageSize := m.getFileListHeight()
	m.fileExplorer.Navigate(delta, pageSize)
	m.loadRequestsFromCurrentFile()
}

// loadHistory loads request history
func (m *Model) loadHistory() tea.Cmd {
	return func() tea.Msg {
		if m.historyManager == nil {
			return historyLoadedMsg{entries: []types.HistoryEntry{}}
		}
		// Get active profile name
		profileName := ""
		if profile := m.sessionMgr.GetActiveProfile(); profile != nil {
			profileName = profile.Name
		}
		entries, err := m.historyManager.Load(profileName)
		if err != nil {
			return errorMsg(fmt.Sprintf("Failed to load history: %v", err))
		}
		return historyLoadedMsg{entries: entries}
	}
}

// loadHistoryEntry loads a specific history entry into the response view
func (m *Model) loadHistoryEntry(index int) tea.Cmd {
	if index < 0 || index >= len(m.historyState.GetEntries()) {
		return nil
	}

	entry := m.historyState.GetEntries()[index]

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
		Timestamp:    entry.Timestamp,
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

// filterHistoryEntries filters history entries based on search query
// Searches in: URL, method, request name, status code, timestamp
func (m *Model) filterHistoryEntries() {
	query := strings.ToLower(m.historyState.GetSearchQuery())

	if query == "" {
		m.historyState.SetEntries(m.historyState.GetAllEntries())
		m.historyState.SetIndex(0)
		m.updateHistoryView()
		return
	}

	var filtered []types.HistoryEntry
	for _, entry := range m.historyState.GetAllEntries() {
		// Search in multiple fields
		if strings.Contains(strings.ToLower(entry.URL), query) ||
			strings.Contains(strings.ToLower(entry.Method), query) ||
			strings.Contains(strings.ToLower(entry.RequestName), query) ||
			strings.Contains(entry.ResponseStatusText, query) ||
			strings.Contains(fmt.Sprintf("%d", entry.ResponseStatus), query) ||
			strings.Contains(entry.Timestamp, query) {
			filtered = append(filtered, entry)
		}
	}

	m.historyState.SetEntries(filtered)
	m.historyState.SetIndex(0)
	m.updateHistoryView()
}

// replayHistoryEntry re-executes a request from history
func (m *Model) replayHistoryEntry(index int) tea.Cmd {
	if index < 0 || index >= len(m.historyState.GetEntries()) {
		return nil
	}

	entry := m.historyState.GetEntries()[index]

	// Convert history entry to HttpRequest
	request := &types.HttpRequest{
		Name:    entry.RequestName,
		Method:  entry.Method,
		URL:     entry.URL,
		Headers: entry.Headers,
		Body:    entry.Body,
	}

	// Set as current request
	m.currentRequest = request

	// Close history modal
	m.mode = ModeNormal
	m.statusMsg = fmt.Sprintf("Replaying request from %s", entry.Timestamp[:19])

	// Execute the request
	return m.executeRequest()
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
			return errorMsg(fmt.Sprintf("OAuth flow failed: %s", categorizeError(err)))
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
	currentFile := m.fileExplorer.GetCurrentFile()
	if currentFile == nil {
		m.currentRequests = nil
		m.currentRequest = nil
		return
	}

	// Skip parsing for WebSocket files - they're executed directly
	if currentFile.HTTPMethod == "WS" {
		m.currentRequests = nil
		m.currentRequest = nil
		m.errorMsg = "" // Clear any errors
		return
	}

	filePath := currentFile.Path
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

	// Track in MRU list
	_ = m.sessionMgr.AddRecentFile(filePath)
}

// getInteractiveVariables returns a list of interactive variables that are used in the current request
func (m *Model) getInteractiveVariables() []string {
	profile := m.sessionMgr.GetActiveProfile()
	if profile == nil || m.currentRequest == nil {
		return nil
	}

	// Extract variables actually used in the current request
	requiredVars := parser.ExtractRequestVariables(m.currentRequest)
	requiredVarsMap := make(map[string]bool)
	for _, varName := range requiredVars {
		requiredVarsMap[varName] = true
	}

	// Only return interactive variables that are actually used in this request
	var interactiveVars []string
	for name, value := range profile.Variables {
		if value.Interactive && requiredVarsMap[name] {
			interactiveVars = append(interactiveVars, name)
		}
	}
	return interactiveVars
}

// executeRequestWithInteractiveVars is called after interactive variables are collected
func (m *Model) executeRequestWithInteractiveVars() tea.Cmd {
	// Simply call executeRequest again, which will now have the values
	return m.executeRequest()
}

// normalizePath extracts and normalizes the path from a URL
// Removes query parameters and domain, keeping only the path
// For analytics grouping purposes
func normalizePath(rawURL string) string {
	// Remove query parameters
	if idx := strings.Index(rawURL, "?"); idx != -1 {
		rawURL = rawURL[:idx]
	}

	// Remove fragment
	if idx := strings.Index(rawURL, "#"); idx != -1 {
		rawURL = rawURL[:idx]
	}

	// Extract path from URL
	// Remove protocol and domain
	re := regexp.MustCompile(`^https?://[^/]+`)
	path := re.ReplaceAllString(rawURL, "")

	// If no path, return root
	if path == "" {
		return "/"
	}

	return path
}

// renderHistoryClearConfirmation renders the confirmation modal for clearing all history
func (m *Model) renderHistoryClearConfirmation() string {
	count := len(m.historyState.GetEntries())
	content := "WARNING\n\n"
	content += "This will permanently delete ALL history entries.\n\n"
	content += fmt.Sprintf("Total entries to delete: %d\n\n", count)
	content += "This action cannot be undone!\n\n"
	content += "Are you sure you want to continue?"

	footer := "[y]es [n]o/ESC"
	return m.renderModalWithFooter("Clear All History", content, footer, 60, 14)
}

// checkForUpdate checks if a new version is available
func (m *Model) checkForUpdate() tea.Cmd {
	return func() tea.Msg {
		available, latestVersion, url, err := version.CheckForUpdate(m.version)
		return versionCheckMsg{
			available:     available,
			latestVersion: latestVersion,
			url:           url,
			err:           err,
		}
	}
}

// applyTagFilter applies the current tag filter to the file list
func (m *Model) applyTagFilter() {
	tags := m.fileExplorer.GetTagFilter()
	m.fileExplorer.SetTagFilter(tags)

	// Load requests from current file after filtering
	m.loadRequestsFromCurrentFile()
}
