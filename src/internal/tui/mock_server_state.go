package tui

import (
	"sync"

	"github.com/studiowebux/restcli/internal/mock"
)

// MockServerState encapsulates all mock server state
type MockServerState struct {
	mu sync.RWMutex

	server     *mock.Server
	running    bool
	configPath string
	logs       []mock.RequestLog
}

// NewMockServerState creates a new mock server state
func NewMockServerState() *MockServerState {
	return &MockServerState{
		server:     nil,
		running:    false,
		configPath: "",
		logs:       []mock.RequestLog{},
	}
}

// GetServer returns the mock server instance
func (s *MockServerState) GetServer() *mock.Server {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.server
}

// SetServer sets the mock server instance
func (s *MockServerState) SetServer(server *mock.Server) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.server = server
}

// IsRunning returns whether the mock server is running
func (s *MockServerState) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// SetRunning sets the running state
func (s *MockServerState) SetRunning(running bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = running
}

// GetConfigPath returns the mock config path
func (s *MockServerState) GetConfigPath() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.configPath
}

// SetConfigPath sets the mock config path
func (s *MockServerState) SetConfigPath(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.configPath = path
}

// GetLogs returns a copy of the mock server logs
func (s *MockServerState) GetLogs() []mock.RequestLog {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return copy to maintain immutability
	result := make([]mock.RequestLog, len(s.logs))
	copy(result, s.logs)
	return result
}

// SetLogs sets the mock server logs
func (s *MockServerState) SetLogs(logs []mock.RequestLog) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logs = logs
}

// AppendLog appends a log entry to the logs
func (s *MockServerState) AppendLog(log mock.RequestLog) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logs = append(s.logs, log)
}

// ClearLogs clears all logs
func (s *MockServerState) ClearLogs() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logs = []mock.RequestLog{}
}

// Start sets the server and marks it as running
func (s *MockServerState) Start(server *mock.Server, configPath string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.server = server
	s.running = true
	s.configPath = configPath
}

// Stop clears the server and marks it as stopped
func (s *MockServerState) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.server = nil
	s.running = false
}

// Reset resets all mock server state
func (s *MockServerState) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.server = nil
	s.running = false
	s.configPath = ""
	s.logs = []mock.RequestLog{}
}
