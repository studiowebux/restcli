package keybinds

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the user's keybinding configuration
type Config struct {
	Version string                       `json:"version"`
	Global  map[string]string            `json:"global,omitempty"`
	Normal  map[string]string            `json:"normal,omitempty"`
	Search  map[string]string            `json:"search,omitempty"`
	Goto    map[string]string            `json:"goto,omitempty"`
	Variables map[string]string          `json:"variables,omitempty"`
	Headers map[string]string            `json:"headers,omitempty"`
	Profiles map[string]string           `json:"profiles,omitempty"`
	Documentation map[string]string       `json:"documentation,omitempty"`
	History map[string]string            `json:"history,omitempty"`
	Analytics map[string]string          `json:"analytics,omitempty"`
	StressTest map[string]string         `json:"stress_test,omitempty"`
	Help    map[string]string            `json:"help,omitempty"`
	Inspect map[string]string            `json:"inspect,omitempty"`
	WebSocket map[string]string          `json:"websocket,omitempty"`
	Modal   map[string]string            `json:"modal,omitempty"`
	TextInput map[string]string          `json:"text_input,omitempty"`
	Confirm map[string]string            `json:"confirm,omitempty"`
	Custom  map[string]map[string]string `json:"custom,omitempty"`
}

// LoadConfig loads keybinding configuration from a JSON file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("invalid keybinds.json format: %w", err)
	}

	return &config, nil
}

// LoadConfigFromDir loads keybinding configuration from a directory
// Looks for keybinds.json in the directory
func LoadConfigFromDir(dir string) (*Config, error) {
	path := filepath.Join(dir, "keybinds.json")
	return LoadConfig(path)
}

// SaveConfig saves keybinding configuration to a JSON file
func SaveConfig(config *Config, path string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// ApplyConfig applies user configuration to a registry
// User bindings override default bindings
func ApplyConfig(registry *Registry, config *Config) error {
	// Map config sections to contexts
	contextMappings := map[Context]map[string]string{
		ContextGlobal:        config.Global,
		ContextNormal:        config.Normal,
		ContextSearch:        config.Search,
		ContextGoto:          config.Goto,
		ContextVariableList:  config.Variables,
		ContextHeaderList:    config.Headers,
		ContextProfileList:   config.Profiles,
		ContextDocumentation: config.Documentation,
		ContextHistory:       config.History,
		ContextAnalytics:     config.Analytics,
		ContextStressTest:    config.StressTest,
		ContextHelp:          config.Help,
		ContextInspect:       config.Inspect,
		ContextWebSocket:     config.WebSocket,
		ContextModal:         config.Modal,
		ContextTextInput:     config.TextInput,
		ContextConfirm:       config.Confirm,
	}

	// Apply each context's bindings
	for context, bindings := range contextMappings {
		for key, actionStr := range bindings {
			action := Action(actionStr)
			// Validate action exists (optional, could skip for flexibility)
			registry.Register(context, key, action)
		}
	}

	// Apply custom contexts if any
	for contextName, bindings := range config.Custom {
		context := Context(contextName)
		for key, actionStr := range bindings {
			action := Action(actionStr)
			registry.Register(context, key, action)
		}
	}

	return nil
}

// LoadOrDefault loads user config if it exists, otherwise returns default registry
func LoadOrDefault(configPath string) (*Registry, error) {
	// Start with defaults
	registry := NewDefaultRegistry()

	// Try to load user config
	if _, err := os.Stat(configPath); err == nil {
		config, err := LoadConfig(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load keybinds.json: %w", err)
		}

		// Apply user config over defaults
		if err := ApplyConfig(registry, config); err != nil {
			return nil, fmt.Errorf("failed to apply keybinds config: %w", err)
		}
	}
	// If config doesn't exist, that's fine - use defaults

	return registry, nil
}

// ExportDefaults exports default keybindings as a config file
// Useful for users to see what can be customized
func ExportDefaults() *Config {
	config := &Config{
		Version: "1.0",
		Global:  make(map[string]string),
		Normal:  make(map[string]string),
	}

	// This is a simplified export - could be expanded to export all bindings
	// For now, just document the structure
	config.Global["quit_force"] = "ctrl+c"
	config.Normal["quit"] = "q"
	config.Normal["execute"] = "enter"
	config.Normal["open_help"] = "?"

	return config
}

// GetDefaultConfigPath returns the default path for keybinds.json
func GetDefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".restcli", "keybinds.json"), nil
}

// CreateExampleConfig creates an example keybinds.json file with comprehensive examples
func CreateExampleConfig(path string) error {
	config := &Config{
		Version: "1.0",
		Global: map[string]string{
			"quit_force": "ctrl+c",
		},
		Normal: map[string]string{
			// Navigation
			"navigate_up":       "up,k",
			"navigate_down":     "down,j",
			"page_up":           "pgup",
			"page_down":         "pgdown",
			"half_page_up":      "ctrl+u",
			"half_page_down":    "ctrl+d",
			"go_to_top":         "gg,home",
			"go_to_bottom":      "G,end",

			// Focus
			"switch_focus":      "tab",

			// File operations
			"quit":              "q",
			"execute":           "enter",
			"open_editor":       "x",
			"configure_editor":  "X",
			"duplicate_file":    "d",
			"delete_file":       "D",
			"rename_file":       "R",
			"create_file":       "F",
			"refresh_files":     "r",

			// Response operations
			"save_response":     "s",
			"copy_to_clipboard": "c",
			"toggle_body":       "b",
			"toggle_headers":    "B",
			"toggle_fullscreen": "f",
			"pin_response":      "w",
			"show_diff":         "W",
			"filter_response":   "J",

			// Modal launchers
			"open_inspect":      "i",
			"open_variables":    "v",
			"open_headers":      "h",
			"open_help":         "?",
			"open_history":      "H",
			"open_analytics":    "A",
			"open_stress_test":  "S",
			"open_profiles":     "p",
			"open_documentation": "m",
			"open_goto":         ":",
			"open_search":       "/",
		},
		Variables: map[string]string{
			"close_modal":  "esc,v,q",
			"navigate_up":  "up,k",
			"navigate_down": "down,j",
			"var_add":      "a",
			"var_edit":     "e",
			"var_delete":   "d",
			"var_manage":   "m",
		},
		Headers: map[string]string{
			"close_modal":  "esc,h,q",
			"navigate_up":  "up,k",
			"navigate_down": "down,j",
			"header_add":   "C",
			"header_edit":  "enter",
			"header_delete": "r",
		},
		Modal: map[string]string{
			"close_modal":  "esc,q",
			"navigate_up":  "up,k",
			"navigate_down": "down,j",
		},
		TextInput: map[string]string{
			"text_submit":      "enter",
			"text_cancel":      "esc",
			"text_paste":       "ctrl+v",
			"text_backspace":   "backspace",
			"text_delete":      "delete",
			"text_move_left":   "left",
			"text_move_right":  "right",
			"text_move_home":   "home,ctrl+a",
			"text_move_end":    "end,ctrl+e",
			"text_clear_after": "ctrl+k",
		},
		Help: map[string]string{
			"close_modal": "esc,?,q",
		},
		History: map[string]string{
			"close_modal":      "esc,H,q",
			"history_execute":  "enter",
			"history_rollback": "r",
			"history_clear":    "C",
		},
		WebSocket: map[string]string{
			"close_modal":    "esc,q",
			"switch_pane":    "tab",
			"ws_send":        "enter",
			"ws_disconnect":  "d",
			"ws_clear":       "C",
		},
	}

	return SaveConfig(config, path)
}
