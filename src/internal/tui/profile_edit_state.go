package tui

import (
	"sync"
)

// ProfileEditState encapsulates all profile editing modal state
type ProfileEditState struct {
	mu sync.RWMutex

	// Field selection and input values
	field            int // 0=name, 1=workdir, 2=editor, 3=output, 4=history, 5=analytics
	name             string
	workdir          string
	editor           string
	output           string
	historyEnabled   *bool // nil=default, true/false=override
	analyticsEnabled *bool // nil=default (false), true/false=override

	// Cursor positions for each text field
	namePos    int
	workdirPos int
	editorPos  int
	outputPos  int
}

// NewProfileEditState creates a new profile edit state
func NewProfileEditState() *ProfileEditState {
	return &ProfileEditState{
		field:            0,
		name:             "",
		workdir:          "",
		editor:           "",
		output:           "",
		historyEnabled:   nil,
		analyticsEnabled: nil,
		namePos:          0,
		workdirPos:       0,
		editorPos:        0,
		outputPos:        0,
	}
}

// GetField returns the currently selected field index
func (s *ProfileEditState) GetField() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.field
}

// SetField sets the currently selected field index
func (s *ProfileEditState) SetField(field int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.field = field
}

// Navigate moves the field selection by delta
func (s *ProfileEditState) Navigate(delta int, maxFields int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.field += delta

	// Wrap around
	if s.field < 0 {
		s.field = maxFields - 1
	} else if s.field >= maxFields {
		s.field = 0
	}
}

// GetName returns the name field value
func (s *ProfileEditState) GetName() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.name
}

// SetName sets the name field value
func (s *ProfileEditState) SetName(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.name = name
}

// GetWorkdir returns the workdir field value
func (s *ProfileEditState) GetWorkdir() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.workdir
}

// SetWorkdir sets the workdir field value
func (s *ProfileEditState) SetWorkdir(workdir string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.workdir = workdir
}

// GetEditor returns the editor field value
func (s *ProfileEditState) GetEditor() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.editor
}

// SetEditor sets the editor field value
func (s *ProfileEditState) SetEditor(editor string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.editor = editor
}

// GetOutput returns the output field value
func (s *ProfileEditState) GetOutput() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.output
}

// SetOutput sets the output field value
func (s *ProfileEditState) SetOutput(output string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.output = output
}

// GetHistoryEnabled returns the history enabled field value
func (s *ProfileEditState) GetHistoryEnabled() *bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.historyEnabled
}

// SetHistoryEnabled sets the history enabled field value
func (s *ProfileEditState) SetHistoryEnabled(enabled *bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.historyEnabled = enabled
}

// GetAnalyticsEnabled returns the analytics enabled field value
func (s *ProfileEditState) GetAnalyticsEnabled() *bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.analyticsEnabled
}

// SetAnalyticsEnabled sets the analytics enabled field value
func (s *ProfileEditState) SetAnalyticsEnabled(enabled *bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.analyticsEnabled = enabled
}

// GetNamePos returns the name cursor position
func (s *ProfileEditState) GetNamePos() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.namePos
}

// SetNamePos sets the name cursor position
func (s *ProfileEditState) SetNamePos(pos int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.namePos = pos
}

// GetWorkdirPos returns the workdir cursor position
func (s *ProfileEditState) GetWorkdirPos() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.workdirPos
}

// SetWorkdirPos sets the workdir cursor position
func (s *ProfileEditState) SetWorkdirPos(pos int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.workdirPos = pos
}

// GetEditorPos returns the editor cursor position
func (s *ProfileEditState) GetEditorPos() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.editorPos
}

// SetEditorPos sets the editor cursor position
func (s *ProfileEditState) SetEditorPos(pos int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.editorPos = pos
}

// GetOutputPos returns the output cursor position
func (s *ProfileEditState) GetOutputPos() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.outputPos
}

// SetOutputPos sets the output cursor position
func (s *ProfileEditState) SetOutputPos(pos int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.outputPos = pos
}

// LoadFromProfile initializes the edit state from a profile
func (s *ProfileEditState) LoadFromProfile(name, workdir, editor, output string, historyEnabled, analyticsEnabled *bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.name = name
	s.workdir = workdir
	s.editor = editor
	s.output = output
	s.historyEnabled = historyEnabled
	s.analyticsEnabled = analyticsEnabled

	// Reset cursor positions
	s.namePos = len(name)
	s.workdirPos = len(workdir)
	s.editorPos = len(editor)
	s.outputPos = len(output)
}

// Reset resets all profile edit state
func (s *ProfileEditState) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.field = 0
	s.name = ""
	s.workdir = ""
	s.editor = ""
	s.output = ""
	s.historyEnabled = nil
	s.analyticsEnabled = nil
	s.namePos = 0
	s.workdirPos = 0
	s.editorPos = 0
	s.outputPos = 0
}
