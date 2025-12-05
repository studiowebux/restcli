package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/studiowebux/restcli/internal/parser"
	"github.com/studiowebux/restcli/internal/stresstest"
	"github.com/studiowebux/restcli/internal/types"
)

// startStressTest starts a stress test execution
func (m *Model) startStressTest() tea.Cmd {
	if m.stressTestConfigEdit == nil {
		return func() tea.Msg {
			return errorMsg("No stress test configuration")
		}
	}

	// Set profile name from active profile
	profile := m.sessionMgr.GetActiveProfile()
	if profile != nil {
		m.stressTestConfigEdit.ProfileName = profile.Name
	}

	// Load requests from the configured file
	requests, err := parser.Parse(m.stressTestConfigEdit.RequestFile)
	if err != nil {
		return func() tea.Msg {
			return errorMsg(fmt.Sprintf("Failed to parse request file: %v", err))
		}
	}

	if len(requests) == 0 {
		return func() tea.Msg {
			return errorMsg("No requests found in file")
		}
	}

	// Always use the first request in the file
	selectedRequest := &requests[0]

	// Make a copy of the request and resolve variables
	requestCopy := *selectedRequest

	// Merge profile headers into request
	if profile != nil && profile.Headers != nil {
		if requestCopy.Headers == nil {
			requestCopy.Headers = make(map[string]string)
		}
		for key, value := range profile.Headers {
			if _, exists := requestCopy.Headers[key]; !exists {
				requestCopy.Headers[key] = value
			}
		}
	}

	// Resolve variables in the request
	if profile != nil {
		resolver := parser.NewVariableResolver(
			profile.Variables,
			m.sessionMgr.GetSession().Variables,
			nil, // No CLI vars for stress test
			parser.LoadSystemEnv(),
		)
		resolvedRequest, err := resolver.ResolveRequest(&requestCopy)
		if err != nil {
			return func() tea.Msg {
				return errorMsg(fmt.Sprintf("Failed to resolve variables: %v", err))
			}
		}
		requestCopy = *resolvedRequest
	}

	// Get TLS config from profile
	var tlsConfig *types.TLSConfig
	if profile != nil && profile.TLS != nil {
		tlsConfig = profile.TLS
	}

	// Merge request-level TLS config
	if requestCopy.TLS != nil {
		if tlsConfig == nil {
			tlsConfig = requestCopy.TLS
		} else {
			// Merge: request-level overrides profile-level
			merged := *tlsConfig
			if requestCopy.TLS.CAFile != "" {
				merged.CAFile = requestCopy.TLS.CAFile
			}
			if requestCopy.TLS.CertFile != "" {
				merged.CertFile = requestCopy.TLS.CertFile
			}
			if requestCopy.TLS.KeyFile != "" {
				merged.KeyFile = requestCopy.TLS.KeyFile
			}
			if requestCopy.TLS.InsecureSkipVerify {
				merged.InsecureSkipVerify = requestCopy.TLS.InsecureSkipVerify
			}
			tlsConfig = &merged
		}
	}

	// Create execution config
	execConfig := &stresstest.ExecutionConfig{
		Request:   &requestCopy,
		TLSConfig: tlsConfig,
		Config:    m.stressTestConfigEdit,
	}

	// Create executor
	executor, err := stresstest.NewExecutor(execConfig, m.stressTestManager)
	if err != nil {
		return func() tea.Msg {
			return errorMsg(fmt.Sprintf("Failed to create stress test executor: %v", err))
		}
	}

	// Store executor and request info for display
	m.stressTestExecutor = executor
	m.stressTestActiveRequest = &requestCopy
	m.stressTestExecutor.Start()

	// Switch to progress mode
	m.mode = ModeStressTestProgress
	m.statusMsg = "Stress test started"

	// Start polling for updates
	return m.pollStressTestProgress()
}

// pollStressTestProgress polls the stress test executor for progress updates
func (m *Model) pollStressTestProgress() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		if m.stressTestExecutor == nil {
			return nil
		}

		if m.stressTestExecutor.IsExecutionComplete() {
			// Test execution completed, trigger finalization
			return stressTestCompletedMsg{}
		}

		// Still running, schedule next poll
		return stressTestProgressMsg{}
	})
}

// loadStressTestFilePicker loads files from active profile's workdir for file picker
func (m *Model) loadStressTestFilePicker() {
	profile := m.sessionMgr.GetActiveProfile()
	if profile == nil {
		m.stressTestFilePickerFiles = nil
		m.stressTestFilePickerIndex = 0
		return
	}

	// Get files from workdir
	files := m.files // Use already loaded files from main view

	// Filter by supported extensions
	supportedExts := map[string]bool{
		".http":  true,
		".yaml":  true,
		".yml":   true,
		".json":  true,
		".jsonc": true,
	}

	filtered := []types.FileInfo{}
	for _, file := range files {
		// Check if extension is supported
		for ext := range supportedExts {
			if len(file.Name) >= len(ext) && file.Name[len(file.Name)-len(ext):] == ext {
				filtered = append(filtered, file)
				break
			}
		}
	}

	m.stressTestFilePickerFiles = filtered
	m.stressTestFilePickerIndex = 0
}

// Message types
type stressTestProgressMsg struct{}
type stressTestCompletedMsg struct{}
