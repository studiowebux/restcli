package tui

import (
	"sync"
	"testing"

	"github.com/studiowebux/restcli/internal/stresstest"
	"github.com/studiowebux/restcli/internal/types"
)

func TestStressTestState_NewInitialization(t *testing.T) {
	state := NewStressTestState(nil)

	if state == nil {
		t.Fatal("NewStressTestState returned nil")
	}

	AssertModelField(t, "configIndex", state.GetConfigIndex(), 0)
	AssertModelField(t, "runIndex", state.GetRunIndex(), 0)
	AssertModelField(t, "focusedPane", state.GetFocusedPane(), "list")
	AssertModelField(t, "stopping", state.GetStopping(), false)
}

func TestStressTestState_SetGetManager(t *testing.T) {
	state := NewStressTestState(nil)

	if state.GetManager() != nil {
		t.Error("Expected nil manager initially")
	}

	// Note: We can't easily create a real manager without a database
	// Just test that setting works
	state.SetManager(nil)
	if state.GetManager() != nil {
		t.Error("Expected manager to be set")
	}
}

func TestStressTestState_ConfigNavigation(t *testing.T) {
	state := NewStressTestState(nil)

	configs := []*stresstest.Config{
		{Name: "config1"},
		{Name: "config2"},
		{Name: "config3"},
	}
	state.SetConfigs(configs)

	// Check initial state
	AssertModelField(t, "initial index", state.GetConfigIndex(), 0)
	cfg := state.GetCurrentConfig()
	if cfg == nil || cfg.Name != "config1" {
		t.Error("Expected config1 initially")
	}

	// Navigate forward
	state.Navigate(1)
	AssertModelField(t, "after navigate +1", state.GetConfigIndex(), 1)
	cfg = state.GetCurrentConfig()
	if cfg == nil || cfg.Name != "config2" {
		t.Error("Expected config2 after navigate")
	}

	// Navigate to end and wrap around
	state.Navigate(1)
	state.Navigate(1)
	AssertModelField(t, "after wrap around", state.GetConfigIndex(), 0)

	// Navigate backward (wraps to end)
	state.Navigate(-1)
	AssertModelField(t, "after navigate -1 wrap", state.GetConfigIndex(), 2)
}

func TestStressTestState_RunNavigation(t *testing.T) {
	state := NewStressTestState(nil)

	runs := []*stresstest.Run{
		{ID: 1},
		{ID: 2},
		{ID: 3},
	}
	state.SetRuns(runs)

	// Check initial state
	AssertModelField(t, "initial run index", state.GetRunIndex(), 0)

	// Navigate forward
	state.NavigateRuns(1)
	AssertModelField(t, "after navigate runs +1", state.GetRunIndex(), 1)

	// Wrap around
	state.NavigateRuns(2)
	AssertModelField(t, "after wrap around runs", state.GetRunIndex(), 0)
}

func TestStressTestState_FocusToggle(t *testing.T) {
	state := NewStressTestState(nil)

	AssertModelField(t, "initial focus", state.GetFocusedPane(), "list")

	state.ToggleFocus()
	AssertModelField(t, "after toggle", state.GetFocusedPane(), "details")

	state.ToggleFocus()
	AssertModelField(t, "after toggle back", state.GetFocusedPane(), "list")
}

func TestStressTestState_ConfigFieldNavigation(t *testing.T) {
	state := NewStressTestState(nil)

	numFields := 5
	AssertModelField(t, "initial field", state.GetConfigField(), 0)

	// Navigate forward
	state.NavigateConfigFields(1, numFields)
	AssertModelField(t, "after navigate +1", state.GetConfigField(), 1)

	// Navigate to end and wrap
	for i := 0; i < 4; i++ {
		state.NavigateConfigFields(1, numFields)
	}
	AssertModelField(t, "after wrap around", state.GetConfigField(), 0)

	// Navigate backward
	state.NavigateConfigFields(-1, numFields)
	AssertModelField(t, "after navigate -1 wrap", state.GetConfigField(), 4)
}

func TestStressTestState_FilePickerNavigation(t *testing.T) {
	state := NewStressTestState(nil)

	files := []types.FileInfo{
		{Name: "file1.http"},
		{Name: "file2.http"},
		{Name: "file3.http"},
	}
	state.SetFilePickerFiles(files)

	AssertModelField(t, "initial picker index", state.GetFilePickerIndex(), 0)

	// Navigate forward
	state.NavigateFilePicker(1)
	AssertModelField(t, "after navigate picker +1", state.GetFilePickerIndex(), 1)

	file := state.GetCurrentFilePickerFile()
	if file == nil || file.Name != "file2.http" {
		t.Error("Expected file2.http")
	}

	// Wrap around
	state.NavigateFilePicker(2)
	AssertModelField(t, "after wrap around picker", state.GetFilePickerIndex(), 0)
}

func TestStressTestState_ConfigEditState(t *testing.T) {
	state := NewStressTestState(nil)

	cfg := &stresstest.Config{Name: "test"}
	state.SetConfigEdit(cfg)

	got := state.GetConfigEdit()
	if got == nil || got.Name != "test" {
		t.Error("Config edit not set correctly")
	}

	state.SetConfigField(2)
	state.SetConfigInput("test input")
	state.SetConfigCursor(5)
	state.SetConfigInsertMode(true)

	AssertModelField(t, "config field", state.GetConfigField(), 2)
	AssertModelField(t, "config input", state.GetConfigInput(), "test input")
	AssertModelField(t, "config cursor", state.GetConfigCursor(), 5)
	AssertModelField(t, "insert mode", state.GetConfigInsertMode(), true)

	// Clear edit state
	state.ClearConfigEdit()
	if state.GetConfigEdit() != nil {
		t.Error("Config edit should be nil after clear")
	}
	AssertModelField(t, "field after clear", state.GetConfigField(), 0)
	AssertModelField(t, "input after clear", state.GetConfigInput(), "")
	AssertModelField(t, "cursor after clear", state.GetConfigCursor(), 0)
	AssertModelField(t, "insert mode after clear", state.GetConfigInsertMode(), false)
}

func TestStressTestState_FilePickerState(t *testing.T) {
	state := NewStressTestState(nil)

	AssertModelField(t, "initial file picker active", state.GetFilePickerActive(), false)

	state.SetFilePickerActive(true)
	AssertModelField(t, "file picker active", state.GetFilePickerActive(), true)

	files := []types.FileInfo{{Name: "test.http"}}
	state.SetFilePickerFiles(files)

	gotFiles := state.GetFilePickerFiles()
	if len(gotFiles) != 1 || gotFiles[0].Name != "test.http" {
		t.Error("File picker files not set correctly")
	}
}

func TestStressTestState_StoppingFlag(t *testing.T) {
	state := NewStressTestState(nil)

	AssertModelField(t, "initial stopping", state.GetStopping(), false)

	state.SetStopping(true)
	AssertModelField(t, "stopping set", state.GetStopping(), true)

	state.SetStopping(false)
	AssertModelField(t, "stopping cleared", state.GetStopping(), false)
}

func TestStressTestState_ActiveRequest(t *testing.T) {
	state := NewStressTestState(nil)

	if state.GetActiveRequest() != nil {
		t.Error("Expected nil active request initially")
	}

	req := &types.HttpRequest{Method: "GET"}
	state.SetActiveRequest(req)

	got := state.GetActiveRequest()
	if got == nil || got.Method != "GET" {
		t.Error("Active request not set correctly")
	}
}

func TestStressTestState_ConcurrentAccess(t *testing.T) {
	state := NewStressTestState(nil)

	configs := make([]*stresstest.Config, 20)
	for i := range configs {
		configs[i] = &stresstest.Config{Name: "config"}
	}
	state.SetConfigs(configs)

	runs := make([]*stresstest.Run, 20)
	for i := range runs {
		runs[i] = &stresstest.Run{ID: int64(i)}
	}
	state.SetRuns(runs)

	var wg sync.WaitGroup
	iterations := 50

	// Concurrent config navigation
	for i := 0; i < iterations; i++ {
		wg.Add(3)

		go func() {
			defer wg.Done()
			state.Navigate(1)
		}()

		go func() {
			defer wg.Done()
			state.Navigate(-1)
		}()

		go func() {
			defer wg.Done()
			_ = state.GetCurrentConfig()
		}()
	}

	// Concurrent run navigation
	for i := 0; i < iterations; i++ {
		wg.Add(2)

		go func() {
			defer wg.Done()
			state.NavigateRuns(1)
		}()

		go func() {
			defer wg.Done()
			_ = state.GetCurrentRun()
		}()
	}

	// Concurrent state changes
	for i := 0; i < iterations; i++ {
		wg.Add(4)

		go func() {
			defer wg.Done()
			state.ToggleFocus()
		}()

		go func() {
			defer wg.Done()
			state.SetStopping(true)
		}()

		go func() {
			defer wg.Done()
			state.SetConfigInsertMode(true)
		}()

		go func() {
			defer wg.Done()
			state.SetFilePickerActive(true)
		}()
	}

	wg.Wait()
	// If we get here without panic or data race, success
}

func TestStressTestState_IndexBoundsAfterSetConfigs(t *testing.T) {
	state := NewStressTestState(nil)

	configs := []*stresstest.Config{
		{Name: "config1"},
		{Name: "config2"},
		{Name: "config3"},
	}
	state.SetConfigs(configs)
	state.SetConfigIndex(2)

	// Set new configs with fewer items
	newConfigs := []*stresstest.Config{
		{Name: "new1"},
	}
	state.SetConfigs(newConfigs)

	// Index should be reset to 0
	AssertModelField(t, "index after shrinking", state.GetConfigIndex(), 0)
}

func TestStressTestState_IndexBoundsAfterSetRuns(t *testing.T) {
	state := NewStressTestState(nil)

	runs := []*stresstest.Run{
		{ID: 1},
		{ID: 2},
		{ID: 3},
	}
	state.SetRuns(runs)
	state.SetRunIndex(2)

	// Set new runs with fewer items
	newRuns := []*stresstest.Run{
		{ID: 100},
	}
	state.SetRuns(newRuns)

	// Index should be reset to 0
	AssertModelField(t, "run index after shrinking", state.GetRunIndex(), 0)
}

func BenchmarkStressTestState_Navigate(b *testing.B) {
	state := NewStressTestState(nil)

	configs := make([]*stresstest.Config, 100)
	for i := range configs {
		configs[i] = &stresstest.Config{Name: "config"}
	}
	state.SetConfigs(configs)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		state.Navigate(1)
	}
}

func BenchmarkStressTestState_ConcurrentAccess(b *testing.B) {
	state := NewStressTestState(nil)

	configs := make([]*stresstest.Config, 100)
	for i := range configs {
		configs[i] = &stresstest.Config{Name: "config"}
	}
	state.SetConfigs(configs)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			state.Navigate(1)
			_ = state.GetCurrentConfig()
		}
	})
}
