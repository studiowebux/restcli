package tui

import (
	"sync"
	"testing"

	"github.com/studiowebux/restcli/internal/proxy"
)

func TestNewProxyServerState(t *testing.T) {
	state := NewProxyServerState()

	if state == nil {
		t.Fatal("NewProxyServerState returned nil")
	}

	if state.GetServer() != nil {
		t.Error("Expected nil server initially")
	}

	if state.IsRunning() {
		t.Error("Expected running to be false initially")
	}

	if len(state.GetLogs()) != 0 {
		t.Errorf("Expected empty logs, got %d entries", len(state.GetLogs()))
	}

	if state.GetSelectedIndex() != 0 {
		t.Errorf("Expected selectedIndex 0, got %d", state.GetSelectedIndex())
	}
}

func TestProxyServerState_ServerOperations(t *testing.T) {
	state := NewProxyServerState()

	// Test initial state
	if state.GetServer() != nil {
		t.Error("Expected nil server initially")
	}

	// Create a proxy server (we'll just use a pointer, no need to actually start it)
	server := &proxy.Proxy{}
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

func TestProxyServerState_RunningOperations(t *testing.T) {
	state := NewProxyServerState()

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

func TestProxyServerState_LogsOperations(t *testing.T) {
	state := NewProxyServerState()

	// Test initial state
	if len(state.GetLogs()) != 0 {
		t.Errorf("Expected empty logs, got %d entries", len(state.GetLogs()))
	}

	// Create test logs
	log1 := &proxy.ProxyLog{Method: "GET", URL: "/api/users"}
	log2 := &proxy.ProxyLog{Method: "POST", URL: "/api/users"}

	// Set logs
	state.SetLogs([]*proxy.ProxyLog{log1, log2})
	logs := state.GetLogs()
	if len(logs) != 2 {
		t.Errorf("Expected 2 logs, got %d", len(logs))
	}

	// Verify slice is copied (modifying returned slice length shouldn't affect state)
	originalLen := len(logs)
	logs = append(logs, &proxy.ProxyLog{Method: "EXTRA", URL: "/extra"})
	logs2 := state.GetLogs()
	if len(logs2) != originalLen {
		t.Error("Slice was not properly copied - append affected internal state")
	}

	// Append log
	log3 := &proxy.ProxyLog{Method: "DELETE", URL: "/api/users/1"}
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

func TestProxyServerState_SelectedIndexOperations(t *testing.T) {
	state := NewProxyServerState()

	// Test initial state
	if state.GetSelectedIndex() != 0 {
		t.Errorf("Expected selectedIndex 0, got %d", state.GetSelectedIndex())
	}

	// Set index
	state.SetSelectedIndex(5)
	if state.GetSelectedIndex() != 5 {
		t.Errorf("Expected selectedIndex 5, got %d", state.GetSelectedIndex())
	}

	// Navigate forward
	state.SetLogs([]*proxy.ProxyLog{
		{Method: "GET", URL: "/1"},
		{Method: "GET", URL: "/2"},
		{Method: "GET", URL: "/3"},
		{Method: "GET", URL: "/4"},
		{Method: "GET", URL: "/5"},
	})
	state.SetSelectedIndex(2)
	state.Navigate(1)
	if state.GetSelectedIndex() != 3 {
		t.Errorf("Expected selectedIndex 3, got %d", state.GetSelectedIndex())
	}

	// Navigate backward
	state.Navigate(-2)
	if state.GetSelectedIndex() != 1 {
		t.Errorf("Expected selectedIndex 1, got %d", state.GetSelectedIndex())
	}

	// Navigate beyond bounds (should clamp)
	state.Navigate(10)
	if state.GetSelectedIndex() != 4 {
		t.Errorf("Expected selectedIndex 4 (clamped to max), got %d", state.GetSelectedIndex())
	}

	// Navigate below zero (should clamp)
	state.Navigate(-10)
	if state.GetSelectedIndex() != 0 {
		t.Errorf("Expected selectedIndex 0 (clamped to min), got %d", state.GetSelectedIndex())
	}
}

func TestProxyServerState_StartStop(t *testing.T) {
	state := NewProxyServerState()

	// Create a proxy server
	server := &proxy.Proxy{}

	// Start server
	state.Start(server)

	if state.GetServer() == nil {
		t.Error("Expected non-nil server after Start")
	}
	if !state.IsRunning() {
		t.Error("Expected running to be true after Start")
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

func TestProxyServerState_Reset(t *testing.T) {
	state := NewProxyServerState()

	// Set various state
	server := &proxy.Proxy{}
	state.SetServer(server)
	state.SetRunning(true)
	state.SetLogs([]*proxy.ProxyLog{
		{Method: "GET", URL: "/test"},
	})
	state.SetSelectedIndex(5)

	// Reset
	state.Reset()

	// Verify everything is reset
	if state.GetServer() != nil {
		t.Error("Expected nil server after reset")
	}
	if state.IsRunning() {
		t.Error("Expected running to be false after reset")
	}
	if len(state.GetLogs()) != 0 {
		t.Errorf("Expected empty logs after reset, got %d entries", len(state.GetLogs()))
	}
	if state.GetSelectedIndex() != 0 {
		t.Errorf("Expected selectedIndex 0 after reset, got %d", state.GetSelectedIndex())
	}
}

func TestProxyServerState_ConcurrentServerAccess(t *testing.T) {
	state := NewProxyServerState()

	var wg sync.WaitGroup
	iterations := 50

	for i := 0; i < iterations; i++ {
		wg.Add(2)

		go func() {
			defer wg.Done()
			server := &proxy.Proxy{}
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

func TestProxyServerState_ConcurrentRunningAccess(t *testing.T) {
	state := NewProxyServerState()

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

func TestProxyServerState_ConcurrentLogsAccess(t *testing.T) {
	state := NewProxyServerState()

	var wg sync.WaitGroup
	iterations := 50

	for i := 0; i < iterations; i++ {
		wg.Add(4)

		go func() {
			defer wg.Done()
			state.AppendLog(&proxy.ProxyLog{Method: "GET", URL: "/test"})
		}()

		go func() {
			defer wg.Done()
			_ = state.GetLogs()
		}()

		go func() {
			defer wg.Done()
			state.SetLogs([]*proxy.ProxyLog{{Method: "POST", URL: "/api"}})
		}()

		go func() {
			defer wg.Done()
			state.ClearLogs()
		}()
	}

	wg.Wait()
	// If we get here without panic or data race, success
}

func TestProxyServerState_ConcurrentNavigationAccess(t *testing.T) {
	state := NewProxyServerState()

	// Set up some logs for navigation
	state.SetLogs([]*proxy.ProxyLog{
		{Method: "GET", URL: "/1"},
		{Method: "GET", URL: "/2"},
		{Method: "GET", URL: "/3"},
	})

	var wg sync.WaitGroup
	iterations := 50

	for i := 0; i < iterations; i++ {
		wg.Add(3)

		go func(delta int) {
			defer wg.Done()
			state.Navigate(delta)
		}(i % 3)

		go func(index int) {
			defer wg.Done()
			state.SetSelectedIndex(index)
		}(i % 3)

		go func() {
			defer wg.Done()
			_ = state.GetSelectedIndex()
		}()
	}

	wg.Wait()
	// If we get here without panic or data race, success
}

func TestProxyServerState_ConcurrentStartStop(t *testing.T) {
	state := NewProxyServerState()

	var wg sync.WaitGroup
	iterations := 50

	for i := 0; i < iterations; i++ {
		wg.Add(3)

		go func() {
			defer wg.Done()
			server := &proxy.Proxy{}
			state.Start(server)
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

func TestProxyServerState_ConcurrentReset(t *testing.T) {
	state := NewProxyServerState()

	var wg sync.WaitGroup
	iterations := 50

	for i := 0; i < iterations; i++ {
		wg.Add(3)

		go func() {
			defer wg.Done()
			server := &proxy.Proxy{}
			state.Start(server)
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
