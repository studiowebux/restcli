package tui

import (
	"sync"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/studiowebux/restcli/internal/stresstest"
	"github.com/studiowebux/restcli/internal/types"
)

// StressTestState manages stress test UI state with thread safety
type StressTestState struct {
	mu sync.RWMutex

	// Manager and executor
	manager  *stresstest.Manager
	executor *stresstest.Executor

	// Active request being tested
	activeRequest *types.HttpRequest

	// Configs and runs
	configs       []*stresstest.Config
	configIndex   int
	runs          []*stresstest.Run
	runIndex      int

	// Viewports for split view
	listView   viewport.Model
	detailView viewport.Model

	// Focus and visibility
	focusedPane string // "list" or "details"

	// Config editing state
	configEdit         *stresstest.Config
	configField        int
	configInput        string
	configCursor       int
	configInsertMode   bool
	filePickerActive   bool
	filePickerFiles    []types.FileInfo
	filePickerIndex    int

	// Execution state
	stopping bool
}

// NewStressTestState creates a new stress test state
func NewStressTestState(manager *stresstest.Manager) *StressTestState {
	return &StressTestState{
		manager:     manager,
		configs:     []*stresstest.Config{},
		runs:        []*stresstest.Run{},
		focusedPane: "list",
		listView:    viewport.New(80, 20),
		detailView:  viewport.New(80, 20),
	}
}

// SetManager sets the stress test manager
func (s *StressTestState) SetManager(mgr *stresstest.Manager) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.manager = mgr
}

// GetManager returns the stress test manager
func (s *StressTestState) GetManager() *stresstest.Manager {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.manager
}

// SetExecutor sets the stress test executor
func (s *StressTestState) SetExecutor(exec *stresstest.Executor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.executor = exec
}

// GetExecutor returns the stress test executor
func (s *StressTestState) GetExecutor() *stresstest.Executor {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.executor
}

// SetActiveRequest sets the active request being tested
func (s *StressTestState) SetActiveRequest(req *types.HttpRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activeRequest = req
}

// GetActiveRequest returns the active request
func (s *StressTestState) GetActiveRequest() *types.HttpRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.activeRequest
}

// SetConfigs sets the list of stress test configs
func (s *StressTestState) SetConfigs(configs []*stresstest.Config) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.configs = configs
	if s.configIndex >= len(configs) {
		s.configIndex = 0
	}
}

// GetConfigs returns the list of configs
func (s *StressTestState) GetConfigs() []*stresstest.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.configs
}

// GetCurrentConfig returns the currently selected config
func (s *StressTestState) GetCurrentConfig() *stresstest.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.configIndex < 0 || s.configIndex >= len(s.configs) {
		return nil
	}
	return s.configs[s.configIndex]
}

// Navigate moves the config selection by delta
func (s *StressTestState) Navigate(delta int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.configs) == 0 {
		return
	}

	s.configIndex += delta

	// Wrap around
	if s.configIndex < 0 {
		s.configIndex = len(s.configs) - 1
	} else if s.configIndex >= len(s.configs) {
		s.configIndex = 0
	}
}

// GetConfigIndex returns the current config index
func (s *StressTestState) GetConfigIndex() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.configIndex
}

// SetConfigIndex sets the config index
func (s *StressTestState) SetConfigIndex(idx int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.configIndex = idx
}

// SetRuns sets the list of stress test runs
func (s *StressTestState) SetRuns(runs []*stresstest.Run) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.runs = runs
	if s.runIndex >= len(runs) {
		s.runIndex = 0
	}
}

// GetRuns returns the list of runs
func (s *StressTestState) GetRuns() []*stresstest.Run {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.runs
}

// GetCurrentRun returns the currently selected run
func (s *StressTestState) GetCurrentRun() *stresstest.Run {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.runIndex < 0 || s.runIndex >= len(s.runs) {
		return nil
	}
	return s.runs[s.runIndex]
}

// NavigateRuns moves the run selection by delta
func (s *StressTestState) NavigateRuns(delta int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.runs) == 0 {
		return
	}

	s.runIndex += delta

	// Wrap around
	if s.runIndex < 0 {
		s.runIndex = len(s.runs) - 1
	} else if s.runIndex >= len(s.runs) {
		s.runIndex = 0
	}
}

// GetRunIndex returns the current run index
func (s *StressTestState) GetRunIndex() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.runIndex
}

// SetRunIndex sets the run index
func (s *StressTestState) SetRunIndex(idx int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.runIndex = idx
}

// GetListView returns the list viewport
func (s *StressTestState) GetListView() viewport.Model {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.listView
}

// SetListView sets the list viewport
func (s *StressTestState) SetListView(vp viewport.Model) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.listView = vp
}

// GetDetailView returns the detail viewport
func (s *StressTestState) GetDetailView() viewport.Model {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.detailView
}

// SetDetailView sets the detail viewport
func (s *StressTestState) SetDetailView(vp viewport.Model) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.detailView = vp
}

// GetFocusedPane returns the currently focused pane
func (s *StressTestState) GetFocusedPane() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.focusedPane
}

// SetFocusedPane sets the focused pane
func (s *StressTestState) SetFocusedPane(pane string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.focusedPane = pane
}

// ToggleFocus toggles between list and details pane
func (s *StressTestState) ToggleFocus() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.focusedPane == "list" {
		s.focusedPane = "details"
	} else {
		s.focusedPane = "list"
	}
}

// GetConfigEdit returns the config being edited
func (s *StressTestState) GetConfigEdit() *stresstest.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.configEdit
}

// SetConfigEdit sets the config being edited
func (s *StressTestState) SetConfigEdit(cfg *stresstest.Config) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.configEdit = cfg
}

// GetConfigField returns the current field being edited
func (s *StressTestState) GetConfigField() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.configField
}

// SetConfigField sets the field being edited
func (s *StressTestState) SetConfigField(field int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.configField = field
}

// NavigateConfigFields moves the field selection by delta
func (s *StressTestState) NavigateConfigFields(delta, numFields int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.configField += delta

	// Wrap around
	if s.configField < 0 {
		s.configField = numFields - 1
	} else if s.configField >= numFields {
		s.configField = 0
	}
}

// GetConfigInput returns the config input buffer
func (s *StressTestState) GetConfigInput() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.configInput
}

// SetConfigInput sets the config input buffer
func (s *StressTestState) SetConfigInput(input string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.configInput = input
}

// GetConfigCursor returns the config input cursor position
func (s *StressTestState) GetConfigCursor() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.configCursor
}

// SetConfigCursor sets the config input cursor position
func (s *StressTestState) SetConfigCursor(pos int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.configCursor = pos
}

// GetConfigInsertMode returns whether insert mode is active
func (s *StressTestState) GetConfigInsertMode() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.configInsertMode
}

// SetConfigInsertMode sets insert mode
func (s *StressTestState) SetConfigInsertMode(mode bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.configInsertMode = mode
}

// GetFilePickerActive returns whether file picker is active
func (s *StressTestState) GetFilePickerActive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.filePickerActive
}

// SetFilePickerActive sets file picker active state
func (s *StressTestState) SetFilePickerActive(active bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.filePickerActive = active
}

// GetFilePickerFiles returns the file picker files
func (s *StressTestState) GetFilePickerFiles() []types.FileInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.filePickerFiles
}

// SetFilePickerFiles sets the file picker files
func (s *StressTestState) SetFilePickerFiles(files []types.FileInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.filePickerFiles = files
	if s.filePickerIndex >= len(files) {
		s.filePickerIndex = 0
	}
}

// GetFilePickerIndex returns the file picker index
func (s *StressTestState) GetFilePickerIndex() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.filePickerIndex
}

// SetFilePickerIndex sets the file picker index
func (s *StressTestState) SetFilePickerIndex(idx int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.filePickerIndex = idx
}

// NavigateFilePicker moves the file picker selection by delta
func (s *StressTestState) NavigateFilePicker(delta int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.filePickerFiles) == 0 {
		return
	}

	s.filePickerIndex += delta

	// Wrap around
	if s.filePickerIndex < 0 {
		s.filePickerIndex = len(s.filePickerFiles) - 1
	} else if s.filePickerIndex >= len(s.filePickerFiles) {
		s.filePickerIndex = 0
	}
}

// GetCurrentFilePickerFile returns the currently selected file in picker
func (s *StressTestState) GetCurrentFilePickerFile() *types.FileInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.filePickerIndex < 0 || s.filePickerIndex >= len(s.filePickerFiles) {
		return nil
	}
	return &s.filePickerFiles[s.filePickerIndex]
}

// GetStopping returns whether test is being stopped
func (s *StressTestState) GetStopping() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stopping
}

// SetStopping sets the stopping state
func (s *StressTestState) SetStopping(stopping bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopping = stopping
}

// ClearConfigEdit clears the config editing state
func (s *StressTestState) ClearConfigEdit() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.configEdit = nil
	s.configField = 0
	s.configInput = ""
	s.configCursor = 0
	s.configInsertMode = false
	s.filePickerActive = false
	s.filePickerFiles = nil
	s.filePickerIndex = 0
}
