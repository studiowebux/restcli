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
			"ctrl+c": "quit_force",
		},
		Normal: map[string]string{
			// Navigation (key -> action)
			"up":      "navigate_up",
			"k":       "navigate_up",
			"down":    "navigate_down",
			"j":       "navigate_down",
			"pgup":    "page_up",
			"pgdown":  "page_down",
			"ctrl+u":  "half_page_up",
			"ctrl+d":  "half_page_down",
			"gg":      "go_to_top",
			"home":    "go_to_top",
			"G":       "go_to_bottom",
			"end":     "go_to_bottom",

			// Focus
			"tab": "switch_focus",

			// File operations
			"q":     "quit",
			"enter": "execute",
			"x":     "open_editor",
			"X":     "configure_editor",
			"d":     "duplicate_file",
			"D":     "delete_file",
			"R":     "rename_file",
			"F":     "create_file",
			"r":     "refresh_files",

			// Response operations
			"s": "save_response",
			"c": "copy_to_clipboard",
			"b": "toggle_body",
			"B": "toggle_headers",
			"f": "toggle_fullscreen",
			"w": "pin_response",
			"W": "show_diff",
			"J": "filter_response",

			// Modal launchers
			"i": "open_inspect",
			"v": "open_variables",
			"h": "open_headers",
			"?": "open_help",
			"H": "open_history",
			"A": "open_analytics",
			"S": "open_stress_test",
			"p": "open_profiles",
			"m": "open_documentation",
			":": "open_goto",
			"/": "open_search",
		},
		Variables: map[string]string{
			"esc": "close_modal",
			"v":   "close_modal",
			"q":   "close_modal",
			"up":  "navigate_up",
			"k":   "navigate_up",
			"down": "navigate_down",
			"j":   "navigate_down",
			"a":   "var_add",
			"e":   "var_edit",
			"d":   "var_delete",
			"m":   "var_manage",
		},
		Headers: map[string]string{
			"esc":   "close_modal",
			"h":     "close_modal",
			"q":     "close_modal",
			"up":    "navigate_up",
			"k":     "navigate_up",
			"down":  "navigate_down",
			"j":     "navigate_down",
			"C":     "header_add",
			"enter": "header_edit",
			"r":     "header_delete",
		},
		Modal: map[string]string{
			"esc":  "close_modal",
			"q":    "close_modal",
			"up":   "navigate_up",
			"k":    "navigate_up",
			"down": "navigate_down",
			"j":    "navigate_down",
		},
		TextInput: map[string]string{
			"enter":     "text_submit",
			"esc":       "text_cancel",
			"ctrl+v":    "text_paste",
			"backspace": "text_backspace",
			"delete":    "text_delete",
			"left":      "text_move_left",
			"right":     "text_move_right",
			"home":      "text_move_home",
			"ctrl+a":    "text_move_home",
			"end":       "text_move_end",
			"ctrl+e":    "text_move_end",
			"ctrl+k":    "text_clear_after",
		},
		Help: map[string]string{
			"esc": "close_modal",
			"?":   "close_modal",
			"q":   "close_modal",
		},
		History: map[string]string{
			"esc":   "close_modal",
			"H":     "close_modal",
			"q":     "close_modal",
			"enter": "history_execute",
			"r":     "history_rollback",
			"C":     "history_clear",
		},
		WebSocket: map[string]string{
			"esc":   "close_modal",
			"q":     "close_modal",
			"tab":   "switch_pane",
			"enter": "ws_send",
			"d":     "ws_disconnect",
			"C":     "ws_clear",
		},
	}

	return SaveConfig(config, path)
}
