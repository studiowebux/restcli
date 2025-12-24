package tui

import (
	"sync"

	"github.com/studiowebux/restcli/internal/proxy"
)

// ProxyServerState encapsulates all proxy server state
type ProxyServerState struct {
	mu sync.RWMutex

	server        *proxy.Proxy
	running       bool
	logs          []*proxy.ProxyLog
	selectedIndex int
}

// NewProxyServerState creates a new proxy server state
func NewProxyServerState() *ProxyServerState {
	return &ProxyServerState{
		server:        nil,
		running:       false,
		logs:          []*proxy.ProxyLog{},
		selectedIndex: 0,
	}
}

// GetServer returns the proxy server instance
func (s *ProxyServerState) GetServer() *proxy.Proxy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.server
}

// SetServer sets the proxy server instance
func (s *ProxyServerState) SetServer(server *proxy.Proxy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.server = server
}

// IsRunning returns whether the proxy server is running
func (s *ProxyServerState) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// SetRunning sets the running state
func (s *ProxyServerState) SetRunning(running bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = running
}

// GetLogs returns a copy of the proxy server logs
func (s *ProxyServerState) GetLogs() []*proxy.ProxyLog {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return copy to maintain immutability
	result := make([]*proxy.ProxyLog, len(s.logs))
	copy(result, s.logs)
	return result
}

// SetLogs sets the proxy server logs
func (s *ProxyServerState) SetLogs(logs []*proxy.ProxyLog) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logs = logs
}

// AppendLog appends a log entry to the logs
func (s *ProxyServerState) AppendLog(log *proxy.ProxyLog) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logs = append(s.logs, log)
}

// ClearLogs clears all logs
func (s *ProxyServerState) ClearLogs() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logs = []*proxy.ProxyLog{}
}

// GetSelectedIndex returns the selected log index
func (s *ProxyServerState) GetSelectedIndex() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.selectedIndex
}

// SetSelectedIndex sets the selected log index
func (s *ProxyServerState) SetSelectedIndex(index int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.selectedIndex = index
}

// Navigate moves the selected index by delta with bounds checking
func (s *ProxyServerState) Navigate(delta int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.selectedIndex += delta

	// Clamp to valid range
	if s.selectedIndex < 0 {
		s.selectedIndex = 0
	} else if s.selectedIndex >= len(s.logs) {
		s.selectedIndex = len(s.logs) - 1
	}

	// Ensure non-negative
	if s.selectedIndex < 0 {
		s.selectedIndex = 0
	}
}

// Start sets the server and marks it as running
func (s *ProxyServerState) Start(server *proxy.Proxy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.server = server
	s.running = true
}

// Stop clears the server and marks it as stopped
func (s *ProxyServerState) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.server = nil
	s.running = false
}

// Reset resets all proxy server state
func (s *ProxyServerState) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.server = nil
	s.running = false
	s.logs = []*proxy.ProxyLog{}
	s.selectedIndex = 0
}
