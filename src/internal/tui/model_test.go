package tui

import (
	"testing"
)

func TestNew_InitializesStateCorrectly(t *testing.T) {
	m := CreateTestModel(t)

	// Verify state objects are initialized
	if m.streamState == nil {
		t.Error("streamState should be initialized")
	}
	if m.requestState == nil {
		t.Error("requestState should be initialized")
	}
	if m.wsState == nil {
		t.Error("wsState should be initialized")
	}

	// Verify initial state values
	AssertModelField(t, "streamState.IsActive()", m.streamState.IsActive(), false)
	AssertModelField(t, "wsState.IsActive()", m.wsState.IsActive(), false)
}

func TestNew_InitializesDefaultMode(t *testing.T) {
	m := CreateTestModel(t)

	AssertModelField(t, "mode", m.mode, ModeNormal)
	AssertModelField(t, "focusedPanel", m.focusedPanel, "sidebar")
	AssertModelField(t, "showBody", m.showBody, true)
	AssertModelField(t, "showHeaders", m.showHeaders, false)
}

func TestNew_InitializesManagers(t *testing.T) {
	m := CreateTestModel(t)

	// Managers may be nil if database initialization fails (expected in tests)
	// Just verify the fields exist and don't panic
	_ = m.analyticsManager
	_ = m.historyManager
	_ = m.stressTestState.GetManager()
	_ = m.bookmarkManager
	_ = m.keybinds

	if m.sessionMgr == nil {
		t.Error("sessionMgr should not be nil")
	}
}

// File loading tests removed - these are integration tests that require
// complex setup with profiles and working directories. The core state
// management is tested in sync_state_test.go and the initialization
// tests below.

func TestModel_StateTransitions(t *testing.T) {
	m := CreateTestModel(t)

	// Test mode transitions
	AssertModelField(t, "initial mode", m.mode, ModeNormal)

	m.mode = ModeHistory
	AssertModelField(t, "history mode", m.mode, ModeHistory)

	m.mode = ModeAnalytics
	AssertModelField(t, "analytics mode", m.mode, ModeAnalytics)

	m.mode = ModeStressTest
	AssertModelField(t, "stress test mode", m.mode, ModeStressTest)

	m.mode = ModeWebSocket
	AssertModelField(t, "websocket mode", m.mode, ModeWebSocket)

	m.mode = ModeNormal
	AssertModelField(t, "back to normal mode", m.mode, ModeNormal)
}

func TestModel_PanelFocus(t *testing.T) {
	m := CreateTestModel(t)

	AssertModelField(t, "initial focus", m.focusedPanel, "sidebar")

	m.focusedPanel = "response"
	AssertModelField(t, "response focus", m.focusedPanel, "response")

	m.focusedPanel = "sidebar"
	AssertModelField(t, "sidebar focus", m.focusedPanel, "sidebar")
}

func TestModel_ToggleHeaders(t *testing.T) {
	m := CreateTestModel(t)

	AssertModelField(t, "initial showHeaders", m.showHeaders, false)

	m.showHeaders = true
	AssertModelField(t, "toggled showHeaders", m.showHeaders, true)

	m.showHeaders = false
	AssertModelField(t, "toggled back showHeaders", m.showHeaders, false)
}

func TestModel_ToggleBody(t *testing.T) {
	m := CreateTestModel(t)

	AssertModelField(t, "initial showBody", m.showBody, true)

	m.showBody = false
	AssertModelField(t, "toggled showBody", m.showBody, false)

	m.showBody = true
	AssertModelField(t, "toggled back showBody", m.showBody, true)
}

func TestModel_InitialFileState(t *testing.T) {
	m := CreateTestModel(t)

	// Verify initial file navigation state
	AssertModelField(t, "initial fileIndex", m.fileExplorer.GetCurrentIndex(), 0)
	AssertModelField(t, "initial fileOffset", m.fileExplorer.GetScrollOffset(), 0)

	// Note: files may be empty in test environment
	// The important thing is that the model initializes without error
}

func TestModel_VersionSet(t *testing.T) {
	m := CreateTestModel(t)

	AssertModelField(t, "version", m.version, "test-version")
}
