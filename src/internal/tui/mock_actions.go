package tui

import (
	"fmt"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/studiowebux/restcli/internal/mock"
)

type mockServerStartedMsg struct {
	server     *mock.Server
	configPath string
	address    string
}

type mockServerStoppedMsg struct{}

// startMockServer starts the mock server
func (m *Model) startMockServer() tea.Cmd {
	// Find first available mock config
	mockFiles := m.findMockConfigs()
	if len(mockFiles) == 0 {
		m.errorMsg = "No mock config files found"
		return nil
	}

	configPath := mockFiles[0]

	return func() tea.Msg {
		// Load config
		config, err := mock.LoadConfig(configPath)
		if err != nil {
			return errorMsg(fmt.Sprintf("Failed to load mock config: %v", err))
		}

		// Get workdir for resolving relative paths in config
		profile := m.sessionMgr.GetActiveProfile()
		workdir := profile.Workdir
		if workdir == "" {
			workdir = filepath.Dir(configPath)
		}

		// Create server
		server := mock.NewServer(config, workdir)

		// Start server
		if err := server.Start(); err != nil {
			return errorMsg(fmt.Sprintf("Failed to start mock server: %v", err))
		}

		return mockServerStartedMsg{
			server:     server,
			configPath: configPath,
			address:    server.GetAddress(),
		}
	}
}

// stopMockServer stops the mock server
func (m *Model) stopMockServer() tea.Cmd {
	return func() tea.Msg {
		if m.mockServer != nil {
			if err := m.mockServer.Stop(); err != nil {
				return errorMsg(fmt.Sprintf("Failed to stop mock server: %v", err))
			}
		}
		return mockServerStoppedMsg{}
	}
}
