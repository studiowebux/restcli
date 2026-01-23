package state

import (
	"testing"
)

func TestNormalState_Name(t *testing.T) {
	state := NewNormalState()
	if state.Name() != "normal" {
		t.Errorf("Expected name 'normal', got '%s'", state.Name())
	}
}

func TestNormalState_CanTransitionTo(t *testing.T) {
	normal := NewNormalState()

	tests := []struct {
		name      string
		target    State
		shouldAllow bool
	}{
		{"to search", NewSearchState(), true},
		{"to goto", NewGotoState(), true},
		{"to variable modal", NewVariableModalState("list"), true},
		{"to header modal", NewHeaderModalState("list"), true},
		{"to profile modal", NewProfileModalState("switch"), true},
		{"to documentation modal", NewDocumentationModalState(), true},
		{"to history modal", NewHistoryModalState(), true},
		{"to analytics modal", NewAnalyticsModalState(), true},
		{"to stress test modal", NewStressTestModalState("list"), true},
		{"to help modal", NewHelpModalState(), true},
		{"to inspect modal", NewInspectModalState(), true},
		{"to websocket modal", NewWebSocketModalState(), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canTransition := normal.CanTransitionTo(tt.target)
			if canTransition != tt.shouldAllow {
				t.Errorf("CanTransitionTo %s: expected %v, got %v",
					tt.name, tt.shouldAllow, canTransition)
			}
		})
	}
}

func TestSearchState_CanTransitionTo(t *testing.T) {
	search := NewSearchState()

	// Can only transition back to normal
	if !search.CanTransitionTo(NewNormalState()) {
		t.Error("SearchState should allow transition to NormalState")
	}

	// Cannot transition to modals
	if search.CanTransitionTo(NewHelpModalState()) {
		t.Error("SearchState should not allow transition to modal")
	}
}

func TestGotoState_CanTransitionTo(t *testing.T) {
	goto_ := NewGotoState()

	// Can only transition back to normal
	if !goto_.CanTransitionTo(NewNormalState()) {
		t.Error("GotoState should allow transition to NormalState")
	}

	// Cannot transition to modals
	if goto_.CanTransitionTo(NewHelpModalState()) {
		t.Error("GotoState should not allow transition to modal")
	}
}

func TestVariableModalState_SubStates(t *testing.T) {
	subStates := []string{"list", "add", "edit", "delete", "options", "manage", "prompt", "alias"}

	for _, sub := range subStates {
		state := NewVariableModalState(sub)
		expectedName := "variable_modal:" + sub
		if state.Name() != expectedName {
			t.Errorf("Expected name '%s', got '%s'", expectedName, state.Name())
		}
		if state.SubState != sub {
			t.Errorf("Expected SubState '%s', got '%s'", sub, state.SubState)
		}
	}
}

func TestVariableModalState_CanTransitionTo(t *testing.T) {
	varModal := NewVariableModalState("list")

	tests := []struct {
		name        string
		target      State
		shouldAllow bool
	}{
		{"to normal", NewNormalState(), true},
		{"to other variable substate", NewVariableModalState("edit"), true},
		{"to jsonpath history", NewJSONPathHistoryModalState(), true},
		{"to unrelated modal", NewHelpModalState(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canTransition := varModal.CanTransitionTo(tt.target)
			if canTransition != tt.shouldAllow {
				t.Errorf("CanTransitionTo %s: expected %v, got %v",
					tt.name, tt.shouldAllow, canTransition)
			}
		})
	}
}

func TestHeaderModalState_CanTransitionTo(t *testing.T) {
	headerModal := NewHeaderModalState("list")

	tests := []struct {
		name        string
		target      State
		shouldAllow bool
	}{
		{"to normal", NewNormalState(), true},
		{"to other header substate", NewHeaderModalState("edit"), true},
		{"to unrelated modal", NewVariableModalState("list"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canTransition := headerModal.CanTransitionTo(tt.target)
			if canTransition != tt.shouldAllow {
				t.Errorf("CanTransitionTo %s: expected %v, got %v",
					tt.name, tt.shouldAllow, canTransition)
			}
		})
	}
}

func TestProfileModalState_CanTransitionTo(t *testing.T) {
	profileModal := NewProfileModalState("switch")

	tests := []struct {
		name        string
		target      State
		shouldAllow bool
	}{
		{"to normal", NewNormalState(), true},
		{"to other profile substate", NewProfileModalState("edit"), true},
		{"to unrelated modal", NewVariableModalState("list"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canTransition := profileModal.CanTransitionTo(tt.target)
			if canTransition != tt.shouldAllow {
				t.Errorf("CanTransitionTo %s: expected %v, got %v",
					tt.name, tt.shouldAllow, canTransition)
			}
		})
	}
}

func TestHistoryModalState_CanTransitionTo(t *testing.T) {
	historyModal := NewHistoryModalState()

	tests := []struct {
		name        string
		target      State
		shouldAllow bool
	}{
		{"to normal", NewNormalState(), true},
		{"to clear confirm", NewHistoryClearConfirmModalState(), true},
		{"to unrelated modal", NewAnalyticsModalState(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canTransition := historyModal.CanTransitionTo(tt.target)
			if canTransition != tt.shouldAllow {
				t.Errorf("CanTransitionTo %s: expected %v, got %v",
					tt.name, tt.shouldAllow, canTransition)
			}
		})
	}
}

func TestHistoryClearConfirmModalState_CanTransitionTo(t *testing.T) {
	confirmModal := NewHistoryClearConfirmModalState()

	tests := []struct {
		name        string
		target      State
		shouldAllow bool
	}{
		{"to normal", NewNormalState(), true},
		{"to history", NewHistoryModalState(), true},
		{"to unrelated modal", NewAnalyticsModalState(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canTransition := confirmModal.CanTransitionTo(tt.target)
			if canTransition != tt.shouldAllow {
				t.Errorf("CanTransitionTo %s: expected %v, got %v",
					tt.name, tt.shouldAllow, canTransition)
			}
		})
	}
}

func TestAnalyticsModalState_CanTransitionTo(t *testing.T) {
	analyticsModal := NewAnalyticsModalState()

	tests := []struct {
		name        string
		target      State
		shouldAllow bool
	}{
		{"to normal", NewNormalState(), true},
		{"to clear confirm", NewAnalyticsClearConfirmModalState(), true},
		{"to unrelated modal", NewHistoryModalState(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canTransition := analyticsModal.CanTransitionTo(tt.target)
			if canTransition != tt.shouldAllow {
				t.Errorf("CanTransitionTo %s: expected %v, got %v",
					tt.name, tt.shouldAllow, canTransition)
			}
		})
	}
}

func TestStressTestModalState_CanTransitionTo(t *testing.T) {
	stressModal := NewStressTestModalState("list")

	tests := []struct {
		name        string
		target      State
		shouldAllow bool
	}{
		{"to normal", NewNormalState(), true},
		{"to other stress test substate", NewStressTestModalState("config"), true},
		{"to unrelated modal", NewHelpModalState(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canTransition := stressModal.CanTransitionTo(tt.target)
			if canTransition != tt.shouldAllow {
				t.Errorf("CanTransitionTo %s: expected %v, got %v",
					tt.name, tt.shouldAllow, canTransition)
			}
		})
	}
}

func TestJSONPathHistoryModalState_CanTransitionTo(t *testing.T) {
	jsonPathModal := NewJSONPathHistoryModalState()

	tests := []struct {
		name        string
		target      State
		shouldAllow bool
	}{
		{"to normal", NewNormalState(), true},
		{"to variable modal", NewVariableModalState("list"), true},
		{"to unrelated modal", NewHelpModalState(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canTransition := jsonPathModal.CanTransitionTo(tt.target)
			if canTransition != tt.shouldAllow {
				t.Errorf("CanTransitionTo %s: expected %v, got %v",
					tt.name, tt.shouldAllow, canTransition)
			}
		})
	}
}

func TestProxyViewerModalState_CanTransitionTo(t *testing.T) {
	proxyModal := NewProxyViewerModalState()

	tests := []struct {
		name        string
		target      State
		shouldAllow bool
	}{
		{"to normal", NewNormalState(), true},
		{"to proxy detail", NewProxyDetailModalState(), true},
		{"to unrelated modal", NewHelpModalState(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canTransition := proxyModal.CanTransitionTo(tt.target)
			if canTransition != tt.shouldAllow {
				t.Errorf("CanTransitionTo %s: expected %v, got %v",
					tt.name, tt.shouldAllow, canTransition)
			}
		})
	}
}

func TestProxyDetailModalState_CanTransitionTo(t *testing.T) {
	detailModal := NewProxyDetailModalState()

	tests := []struct {
		name        string
		target      State
		shouldAllow bool
	}{
		{"to normal", NewNormalState(), true},
		{"to proxy viewer", NewProxyViewerModalState(), true},
		{"to unrelated modal", NewHelpModalState(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canTransition := detailModal.CanTransitionTo(tt.target)
			if canTransition != tt.shouldAllow {
				t.Errorf("CanTransitionTo %s: expected %v, got %v",
					tt.name, tt.shouldAllow, canTransition)
			}
		})
	}
}

func TestBaseModalState_IsModal(t *testing.T) {
	modals := []ModalState{
		NewVariableModalState("list"),
		NewHeaderModalState("list"),
		NewProfileModalState("switch"),
		NewDocumentationModalState(),
		NewHistoryModalState(),
		NewAnalyticsModalState(),
		NewStressTestModalState("list"),
		NewHelpModalState(),
		NewInspectModalState(),
		NewRenameModalState(),
		NewEditorConfigModalState(),
		NewOAuthConfigModalState(),
		NewOAuthEditModalState(),
		NewConfigViewModalState(),
		NewDeleteConfirmModalState(),
		NewShellErrorsModalState(),
		NewCreateFileModalState(),
		NewMRUModalState(),
		NewDiffModalState(),
		NewConfirmExecutionModalState(),
		NewErrorDetailModalState(),
		NewStatusDetailModalState(),
		NewBodyOverrideModalState(),
		NewJSONPathHistoryModalState(),
		NewTagFilterModalState(),
		NewMockServerModalState(),
		NewProxyViewerModalState(),
		NewProxyDetailModalState(),
		NewWebSocketModalState(),
	}

	for _, modal := range modals {
		if !modal.IsModal() {
			t.Errorf("Modal %s should return true for IsModal()", modal.Name())
		}
	}
}

func TestAllModalStates_DefaultCanTransitionToNormal(t *testing.T) {
	// Modals with custom CanTransitionTo
	customTransitions := map[string]bool{
		"variable_modal:list":           true,
		"header_modal:list":             true,
		"profile_modal:switch":          true,
		"history_modal":                 true,
		"history_clear_confirm_modal":   true,
		"analytics_modal":               true,
		"analytics_clear_confirm_modal": true,
		"stress_test_modal:list":        true,
		"jsonpath_history_modal":        true,
		"proxy_viewer_modal":            true,
		"proxy_detail_modal":            true,
	}

	modals := []ModalState{
		NewRenameModalState(),
		NewEditorConfigModalState(),
		NewOAuthConfigModalState(),
		NewOAuthEditModalState(),
		NewConfigViewModalState(),
		NewDeleteConfirmModalState(),
		NewShellErrorsModalState(),
		NewCreateFileModalState(),
		NewMRUModalState(),
		NewDiffModalState(),
		NewConfirmExecutionModalState(),
		NewErrorDetailModalState(),
		NewStatusDetailModalState(),
		NewBodyOverrideModalState(),
		NewTagFilterModalState(),
		NewMockServerModalState(),
		NewWebSocketModalState(),
		NewDocumentationModalState(),
		NewHelpModalState(),
		NewInspectModalState(),
	}

	normalState := NewNormalState()

	for _, modal := range modals {
		if customTransitions[modal.Name()] {
			continue // Skip modals with custom transition logic
		}

		if !modal.CanTransitionTo(normalState) {
			t.Errorf("Modal %s should allow transition to normal state", modal.Name())
		}
	}
}

func TestOnEnterOnExit_DoNotFail(t *testing.T) {
	states := []State{
		NewNormalState(),
		NewSearchState(),
		NewGotoState(),
		NewVariableModalState("list"),
		NewHeaderModalState("list"),
		NewProfileModalState("switch"),
		NewDocumentationModalState(),
		NewHistoryModalState(),
		NewAnalyticsModalState(),
		NewStressTestModalState("list"),
		NewHelpModalState(),
	}

	for _, state := range states {
		if err := state.OnEnter(); err != nil {
			t.Errorf("State %s OnEnter() failed: %v", state.Name(), err)
		}

		if err := state.OnExit(); err != nil {
			t.Errorf("State %s OnExit() failed: %v", state.Name(), err)
		}
	}
}
