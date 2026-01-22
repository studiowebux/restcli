package state

// ModalState is a marker interface for states that represent modals.
// Modals can be opened from normal state and typically return to normal state when closed.
type ModalState interface {
	State
	IsModal() bool
}

// BaseState provides common functionality for all states
type BaseState struct {
	name string
}

func (b *BaseState) Name() string {
	return b.name
}

func (b *BaseState) OnEnter() error {
	return nil
}

func (b *BaseState) OnExit() error {
	return nil
}

// NormalState represents the default application state
type NormalState struct {
	BaseState
}

func NewNormalState() *NormalState {
	return &NormalState{
		BaseState: BaseState{name: "normal"},
	}
}

func (s *NormalState) CanTransitionTo(state State) bool {
	// Can transition to any modal state or other normal-mode states
	switch state.(type) {
	case ModalState:
		return true
	case *SearchState, *GotoState:
		return true
	}
	return false
}

// SearchState represents active search mode
type SearchState struct {
	BaseState
}

func NewSearchState() *SearchState {
	return &SearchState{
		BaseState: BaseState{name: "search"},
	}
}

func (s *SearchState) CanTransitionTo(state State) bool {
	// Can only go back to normal
	_, ok := state.(*NormalState)
	return ok
}

// GotoState represents line jump mode
type GotoState struct {
	BaseState
}

func NewGotoState() *GotoState {
	return &GotoState{
		BaseState: BaseState{name: "goto"},
	}
}

func (s *GotoState) CanTransitionTo(state State) bool {
	// Can only go back to normal
	_, ok := state.(*NormalState)
	return ok
}

// BaseModalState provides common functionality for modal states
type BaseModalState struct {
	BaseState
}

func (b *BaseModalState) IsModal() bool {
	return true
}

func (b *BaseModalState) CanTransitionTo(state State) bool {
	// Modals can typically only transition back to normal
	// (unless they have sub-states)
	_, ok := state.(*NormalState)
	return ok
}

// VariableModalState represents variable editor with sub-states
type VariableModalState struct {
	BaseModalState
	SubState string // "list", "add", "edit", "delete", "options", "manage", "prompt", "alias"
}

func NewVariableModalState(subState string) *VariableModalState {
	name := "variable_modal"
	if subState != "" {
		name = "variable_modal:" + subState
	}
	return &VariableModalState{
		BaseModalState: BaseModalState{
			BaseState: BaseState{name: name},
		},
		SubState: subState,
	}
}

func (s *VariableModalState) CanTransitionTo(state State) bool {
	// Can transition to normal or other variable sub-states
	switch st := state.(type) {
	case *NormalState:
		return true
	case *VariableModalState:
		return true
	case *JSONPathHistoryModalState:
		return true // Can open JSONPath history from variable editor
	default:
		_ = st
		return false
	}
}

// HeaderModalState represents header editor with sub-states
type HeaderModalState struct {
	BaseModalState
	SubState string // "list", "add", "edit", "delete"
}

func NewHeaderModalState(subState string) *HeaderModalState {
	name := "header_modal"
	if subState != "" {
		name = "header_modal:" + subState
	}
	return &HeaderModalState{
		BaseModalState: BaseModalState{
			BaseState: BaseState{name: name},
		},
		SubState: subState,
	}
}

func (s *HeaderModalState) CanTransitionTo(state State) bool {
	// Can transition to normal or other header sub-states
	switch state.(type) {
	case *NormalState:
		return true
	case *HeaderModalState:
		return true
	default:
		return false
	}
}

// ProfileModalState represents profile manager with sub-states
type ProfileModalState struct {
	BaseModalState
	SubState string // "switch", "create", "edit", "duplicate", "delete_confirm"
}

func NewProfileModalState(subState string) *ProfileModalState {
	name := "profile_modal"
	if subState != "" {
		name = "profile_modal:" + subState
	}
	return &ProfileModalState{
		BaseModalState: BaseModalState{
			BaseState: BaseState{name: name},
		},
		SubState: subState,
	}
}

func (s *ProfileModalState) CanTransitionTo(state State) bool {
	// Can transition to normal or other profile sub-states
	switch state.(type) {
	case *NormalState:
		return true
	case *ProfileModalState:
		return true
	default:
		return false
	}
}

// Simple modal states (no sub-states)

type DocumentationModalState struct{ BaseModalState }

func NewDocumentationModalState() *DocumentationModalState {
	return &DocumentationModalState{
		BaseModalState: BaseModalState{BaseState: BaseState{name: "documentation_modal"}},
	}
}

type HistoryModalState struct{ BaseModalState }

func NewHistoryModalState() *HistoryModalState {
	return &HistoryModalState{
		BaseModalState: BaseModalState{BaseState: BaseState{name: "history_modal"}},
	}
}

func (s *HistoryModalState) CanTransitionTo(state State) bool {
	// Can transition to normal or history clear confirm
	switch state.(type) {
	case *NormalState, *HistoryClearConfirmModalState:
		return true
	default:
		return false
	}
}

type HistoryClearConfirmModalState struct{ BaseModalState }

func NewHistoryClearConfirmModalState() *HistoryClearConfirmModalState {
	return &HistoryClearConfirmModalState{
		BaseModalState: BaseModalState{BaseState: BaseState{name: "history_clear_confirm_modal"}},
	}
}

func (s *HistoryClearConfirmModalState) CanTransitionTo(state State) bool {
	// Can only go back to history or normal
	switch state.(type) {
	case *NormalState, *HistoryModalState:
		return true
	default:
		return false
	}
}

type AnalyticsModalState struct{ BaseModalState }

func NewAnalyticsModalState() *AnalyticsModalState {
	return &AnalyticsModalState{
		BaseModalState: BaseModalState{BaseState: BaseState{name: "analytics_modal"}},
	}
}

func (s *AnalyticsModalState) CanTransitionTo(state State) bool {
	// Can transition to normal or analytics clear confirm
	switch state.(type) {
	case *NormalState, *AnalyticsClearConfirmModalState:
		return true
	default:
		return false
	}
}

type AnalyticsClearConfirmModalState struct{ BaseModalState }

func NewAnalyticsClearConfirmModalState() *AnalyticsClearConfirmModalState {
	return &AnalyticsClearConfirmModalState{
		BaseModalState: BaseModalState{BaseState: BaseState{name: "analytics_clear_confirm_modal"}},
	}
}

func (s *AnalyticsClearConfirmModalState) CanTransitionTo(state State) bool {
	// Can only go back to analytics or normal
	switch state.(type) {
	case *NormalState, *AnalyticsModalState:
		return true
	default:
		return false
	}
}

type StressTestModalState struct {
	BaseModalState
	SubState string // "list", "config", "load_config", "progress", "results"
}

func NewStressTestModalState(subState string) *StressTestModalState {
	name := "stress_test_modal"
	if subState != "" {
		name = "stress_test_modal:" + subState
	}
	return &StressTestModalState{
		BaseModalState: BaseModalState{
			BaseState: BaseState{name: name},
		},
		SubState: subState,
	}
}

func (s *StressTestModalState) CanTransitionTo(state State) bool {
	// Can transition to normal or other stress test sub-states
	switch state.(type) {
	case *NormalState:
		return true
	case *StressTestModalState:
		return true
	default:
		return false
	}
}

type HelpModalState struct{ BaseModalState }

func NewHelpModalState() *HelpModalState {
	return &HelpModalState{
		BaseModalState: BaseModalState{BaseState: BaseState{name: "help_modal"}},
	}
}

type InspectModalState struct{ BaseModalState }

func NewInspectModalState() *InspectModalState {
	return &InspectModalState{
		BaseModalState: BaseModalState{BaseState: BaseState{name: "inspect_modal"}},
	}
}

type RenameModalState struct{ BaseModalState }

func NewRenameModalState() *RenameModalState {
	return &RenameModalState{
		BaseModalState: BaseModalState{BaseState: BaseState{name: "rename_modal"}},
	}
}

type EditorConfigModalState struct{ BaseModalState }

func NewEditorConfigModalState() *EditorConfigModalState {
	return &EditorConfigModalState{
		BaseModalState: BaseModalState{BaseState: BaseState{name: "editor_config_modal"}},
	}
}

type OAuthConfigModalState struct{ BaseModalState }

func NewOAuthConfigModalState() *OAuthConfigModalState {
	return &OAuthConfigModalState{
		BaseModalState: BaseModalState{BaseState: BaseState{name: "oauth_config_modal"}},
	}
}

type OAuthEditModalState struct{ BaseModalState }

func NewOAuthEditModalState() *OAuthEditModalState {
	return &OAuthEditModalState{
		BaseModalState: BaseModalState{BaseState: BaseState{name: "oauth_edit_modal"}},
	}
}

type ConfigViewModalState struct{ BaseModalState }

func NewConfigViewModalState() *ConfigViewModalState {
	return &ConfigViewModalState{
		BaseModalState: BaseModalState{BaseState: BaseState{name: "config_view_modal"}},
	}
}

type DeleteConfirmModalState struct{ BaseModalState }

func NewDeleteConfirmModalState() *DeleteConfirmModalState {
	return &DeleteConfirmModalState{
		BaseModalState: BaseModalState{BaseState: BaseState{name: "delete_confirm_modal"}},
	}
}

type ShellErrorsModalState struct{ BaseModalState }

func NewShellErrorsModalState() *ShellErrorsModalState {
	return &ShellErrorsModalState{
		BaseModalState: BaseModalState{BaseState: BaseState{name: "shell_errors_modal"}},
	}
}

type CreateFileModalState struct{ BaseModalState }

func NewCreateFileModalState() *CreateFileModalState {
	return &CreateFileModalState{
		BaseModalState: BaseModalState{BaseState: BaseState{name: "create_file_modal"}},
	}
}

type MRUModalState struct{ BaseModalState }

func NewMRUModalState() *MRUModalState {
	return &MRUModalState{
		BaseModalState: BaseModalState{BaseState: BaseState{name: "mru_modal"}},
	}
}

type DiffModalState struct{ BaseModalState }

func NewDiffModalState() *DiffModalState {
	return &DiffModalState{
		BaseModalState: BaseModalState{BaseState: BaseState{name: "diff_modal"}},
	}
}

type ConfirmExecutionModalState struct{ BaseModalState }

func NewConfirmExecutionModalState() *ConfirmExecutionModalState {
	return &ConfirmExecutionModalState{
		BaseModalState: BaseModalState{BaseState: BaseState{name: "confirm_execution_modal"}},
	}
}

type ErrorDetailModalState struct{ BaseModalState }

func NewErrorDetailModalState() *ErrorDetailModalState {
	return &ErrorDetailModalState{
		BaseModalState: BaseModalState{BaseState: BaseState{name: "error_detail_modal"}},
	}
}

type StatusDetailModalState struct{ BaseModalState }

func NewStatusDetailModalState() *StatusDetailModalState {
	return &StatusDetailModalState{
		BaseModalState: BaseModalState{BaseState: BaseState{name: "status_detail_modal"}},
	}
}

type BodyOverrideModalState struct{ BaseModalState }

func NewBodyOverrideModalState() *BodyOverrideModalState {
	return &BodyOverrideModalState{
		BaseModalState: BaseModalState{BaseState: BaseState{name: "body_override_modal"}},
	}
}

type JSONPathHistoryModalState struct{ BaseModalState }

func NewJSONPathHistoryModalState() *JSONPathHistoryModalState {
	return &JSONPathHistoryModalState{
		BaseModalState: BaseModalState{BaseState: BaseState{name: "jsonpath_history_modal"}},
	}
}

func (s *JSONPathHistoryModalState) CanTransitionTo(state State) bool {
	// Can go back to variable modal or normal
	switch state.(type) {
	case *NormalState, *VariableModalState:
		return true
	default:
		return false
	}
}

type TagFilterModalState struct{ BaseModalState }

func NewTagFilterModalState() *TagFilterModalState {
	return &TagFilterModalState{
		BaseModalState: BaseModalState{BaseState: BaseState{name: "tag_filter_modal"}},
	}
}

type MockServerModalState struct{ BaseModalState }

func NewMockServerModalState() *MockServerModalState {
	return &MockServerModalState{
		BaseModalState: BaseModalState{BaseState: BaseState{name: "mock_server_modal"}},
	}
}

type ProxyViewerModalState struct{ BaseModalState }

func NewProxyViewerModalState() *ProxyViewerModalState {
	return &ProxyViewerModalState{
		BaseModalState: BaseModalState{BaseState: BaseState{name: "proxy_viewer_modal"}},
	}
}

func (s *ProxyViewerModalState) CanTransitionTo(state State) bool {
	// Can transition to normal or proxy detail
	switch state.(type) {
	case *NormalState, *ProxyDetailModalState:
		return true
	default:
		return false
	}
}

type ProxyDetailModalState struct{ BaseModalState }

func NewProxyDetailModalState() *ProxyDetailModalState {
	return &ProxyDetailModalState{
		BaseModalState: BaseModalState{BaseState: BaseState{name: "proxy_detail_modal"}},
	}
}

func (s *ProxyDetailModalState) CanTransitionTo(state State) bool {
	// Can go back to proxy viewer or normal
	switch state.(type) {
	case *NormalState, *ProxyViewerModalState:
		return true
	default:
		return false
	}
}

type WebSocketModalState struct{ BaseModalState }

func NewWebSocketModalState() *WebSocketModalState {
	return &WebSocketModalState{
		BaseModalState: BaseModalState{BaseState: BaseState{name: "websocket_modal"}},
	}
}

type InteractiveVarModalState struct{ BaseModalState }

func NewInteractiveVarModalState() *InteractiveVarModalState {
	return &InteractiveVarModalState{
		BaseModalState: BaseModalState{BaseState: BaseState{name: "interactive_var_modal"}},
	}
}
