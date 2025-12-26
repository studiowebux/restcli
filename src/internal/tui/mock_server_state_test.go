package tui

import (
	"sync"
	"testing"

	"github.com/studiowebux/restcli/internal/mock"
)

func TestNewMockServerState(t *testing.T) {
	state := NewMockServerState()

	if state == nil {
		t.Fatal("NewMockServerState returned nil")
	}

	if state.GetServer() != nil {
		t.Error("Expected nil server initially")
	}

	if state.IsRunning() {
		t.Error("Expected running to be false initially")
	}

	if state.GetConfigPath() != "" {
		t.Errorf("Expected empty config path, got %s", state.GetConfigPath())
	}

	if len(state.GetLogs()) != 0 {
		t.Errorf("Expected empty logs, got %d entries", len(state.GetLogs()))
	}
}

func TestMockServerState_ServerOperations(t *testing.T) {
	state := NewMockServerState()

	// Test initial state
	if state.GetServer() != nil {
		t.Error("Expected nil server initially")
	}

	// Create a mock server (we'll just use a pointer, no need to actually start it)
	server := &mock.Server{}
	state.SetServer(server)

	if state.GetServer() == nil {
		t.Error("Expected non-nil server after SetServer")
	}

	// Set back to nil
	state.SetServer(nil)
	if state.GetServer() != nil {
		t.Error("Expected nil server after setting to nil")
	}
}

func TestMockServerState_RunningOperations(t *testing.T) {
	state := NewMockServerState()

	// Test initial state
	if state.IsRunning() {
		t.Error("Expected running to be false initially")
	}

	// Set running to true
	state.SetRunning(true)
	if !state.IsRunning() {
		t.Error("Expected running to be true after SetRunning(true)")
	}

	// Set running to false
	state.SetRunning(false)
	if state.IsRunning() {
		t.Error("Expected running to be false after SetRunning(false)")
	}
}

func TestMockServerState_ConfigPathOperations(t *testing.T) {
	state := NewMockServerState()

	// Test initial state
	if state.GetConfigPath() != "" {
		t.Errorf("Expected empty config path, got %s", state.GetConfigPath())
	}

	// Set config path
	state.SetConfigPath("/path/to/config.json")
	if state.GetConfigPath() != "/path/to/config.json" {
		t.Errorf("Expected '/path/to/config.json', got %s", state.GetConfigPath())
	}

	// Update config path
	state.SetConfigPath("/new/path.json")
	if state.GetConfigPath() != "/new/path.json" {
		t.Errorf("Expected '/new/path.json', got %s", state.GetConfigPath())
	}
}

func TestMockServerState_LogsOperations(t *testing.T) {
	state := NewMockServerState()

	// Test initial state
	if len(state.GetLogs()) != 0 {
		t.Errorf("Expected empty logs, got %d entries", len(state.GetLogs()))
	}

	// Create test logs
	log1 := mock.RequestLog{Method: "GET", Path: "/api/users"}
	log2 := mock.RequestLog{Method: "POST", Path: "/api/users"}

	// Set logs
	state.SetLogs([]mock.RequestLog{log1, log2})
	logs := state.GetLogs()
	if len(logs) != 2 {
		t.Errorf("Expected 2 logs, got %d", len(logs))
	}

	// Verify immutability (modifying returned slice shouldn't affect state)
	logs[0].Method = "MODIFIED"
	logs2 := state.GetLogs()
	if logs2[0].Method == "MODIFIED" {
		t.Error("Logs were not properly copied - modification affected internal state")
	}

	// Append log
	log3 := mock.RequestLog{Method: "DELETE", Path: "/api/users/1"}
	state.AppendLog(log3)
	logs = state.GetLogs()
	if len(logs) != 3 {
		t.Errorf("Expected 3 logs after append, got %d", len(logs))
	}

	// Clear logs
	state.ClearLogs()
	if len(state.GetLogs()) != 0 {
		t.Errorf("Expected empty logs after clear, got %d entries", len(state.GetLogs()))
	}
}

func TestMockServerState_StartStop(t *testing.T) {
	state := NewMockServerState()

	// Create a mock server
	server := &mock.Server{}
	configPath := "/config/mock.json"

	// Start server
	state.Start(server, configPath)

	if state.GetServer() == nil {
		t.Error("Expected non-nil server after Start")
	}
	if !state.IsRunning() {
		t.Error("Expected running to be true after Start")
	}
	if state.GetConfigPath() != configPath {
		t.Errorf("Expected config path %s, got %s", configPath, state.GetConfigPath())
	}

	// Stop server
	state.Stop()

	if state.GetServer() != nil {
		t.Error("Expected nil server after Stop")
	}
	if state.IsRunning() {
		t.Error("Expected running to be false after Stop")
	}
}

func TestMockServerState_Reset(t *testing.T) {
	state := NewMockServerState()

	// Set various state
	server := &mock.Server{}
	state.SetServer(server)
	state.SetRunning(true)
	state.SetConfigPath("/config.json")
	state.SetLogs([]mock.RequestLog{
		{Method: "GET", Path: "/test"},
	})

	// Reset
	state.Reset()

	// Verify everything is reset
	if state.GetServer() != nil {
		t.Error("Expected nil server after reset")
	}
	if state.IsRunning() {
		t.Error("Expected running to be false after reset")
	}
	if state.GetConfigPath() != "" {
		t.Errorf("Expected empty config path after reset, got %s", state.GetConfigPath())
	}
	if len(state.GetLogs()) != 0 {
		t.Errorf("Expected empty logs after reset, got %d entries", len(state.GetLogs()))
	}
}

func TestMockServerState_ConcurrentServerAccess(t *testing.T) {
	state := NewMockServerState()

	var wg sync.WaitGroup
	iterations := 50

	for i := 0; i < iterations; i++ {
		wg.Add(2)

		go func() {
			defer wg.Done()
			server := &mock.Server{}
			state.SetServer(server)
		}()

		go func() {
			defer wg.Done()
			_ = state.GetServer()
		}()
	}

	wg.Wait()
	// If we get here without panic or data race, success
}

func TestMockServerState_ConcurrentRunningAccess(t *testing.T) {
	state := NewMockServerState()

	var wg sync.WaitGroup
	iterations := 50

	for i := 0; i < iterations; i++ {
		wg.Add(2)

		go func(running bool) {
			defer wg.Done()
			state.SetRunning(running)
		}(i%2 == 0)

		go func() {
			defer wg.Done()
			_ = state.IsRunning()
		}()
	}

	wg.Wait()
	// If we get here without panic or data race, success
}

func TestMockServerState_ConcurrentLogsAccess(t *testing.T) {
	state := NewMockServerState()

	var wg sync.WaitGroup
	iterations := 50

	for i := 0; i < iterations; i++ {
		wg.Add(4)

		go func() {
			defer wg.Done()
			state.AppendLog(mock.RequestLog{Method: "GET", Path: "/test"})
		}()

		go func() {
			defer wg.Done()
			_ = state.GetLogs()
		}()

		go func() {
			defer wg.Done()
			state.SetLogs([]mock.RequestLog{{Method: "POST", Path: "/api"}})
		}()

		go func() {
			defer wg.Done()
			state.ClearLogs()
		}()
	}

	wg.Wait()
	// If we get here without panic or data race, success
}

func TestMockServerState_ConcurrentStartStop(t *testing.T) {
	state := NewMockServerState()

	var wg sync.WaitGroup
	iterations := 50

	for i := 0; i < iterations; i++ {
		wg.Add(3)

		go func() {
			defer wg.Done()
			server := &mock.Server{}
			state.Start(server, "/config.json")
		}()

		go func() {
			defer wg.Done()
			state.Stop()
		}()

		go func() {
			defer wg.Done()
			_ = state.IsRunning()
			_ = state.GetServer()
		}()
	}

	wg.Wait()
	// If we get here without panic or data race, success
}

func TestMockServerState_ConcurrentReset(t *testing.T) {
	state := NewMockServerState()

	var wg sync.WaitGroup
	iterations := 50

	for i := 0; i < iterations; i++ {
		wg.Add(3)

		go func() {
			defer wg.Done()
			server := &mock.Server{}
			state.Start(server, "/config.json")
		}()

		go func() {
			defer wg.Done()
			state.Reset()
		}()

		go func() {
			defer wg.Done()
			_ = state.IsRunning()
			_ = state.GetLogs()
		}()
	}

	wg.Wait()
	// If we get here without panic or data race, success
}
