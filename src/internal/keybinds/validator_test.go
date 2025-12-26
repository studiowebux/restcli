package keybinds

import (
	"strings"
	"testing"
)

func TestNewValidator(t *testing.T) {
	v := NewValidator()

	if v == nil {
		t.Fatal("NewValidator returned nil")
	}

	if len(v.reservedKeys) == 0 {
		t.Error("Expected reserved keys to be initialized")
	}

	if !v.reservedKeys["ctrl+c"] {
		t.Error("Expected ctrl+c to be a reserved key")
	}

	if len(v.contextHierarchy) == 0 {
		t.Error("Expected context hierarchy to be initialized")
	}
}

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      ValidationError
		expected string
	}{
		{
			name: "conflict error",
			err: ValidationError{
				Type:    "conflict",
				Context: ContextNormal,
				Key:     "q",
				Message: "key bound 2 times",
			},
			expected: "[conflict] q in context 'normal': key bound 2 times",
		},
		{
			name: "invalid error",
			err: ValidationError{
				Type:    "invalid",
				Context: ContextGlobal,
				Key:     "",
				Message: "empty key",
			},
			expected: "[invalid]  in context 'global': empty key",
		},
		{
			name: "warning",
			err: ValidationError{
				Type:    "warning",
				Context: ContextWebSocket,
				Key:     "tab",
				Message: "shadows global binding",
			},
			expected: "[warning] tab in context 'websocket': shadows global binding",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestValidationResult_HasErrors(t *testing.T) {
	tests := []struct {
		name     string
		result   *ValidationResult
		expected bool
	}{
		{
			name:     "no errors",
			result:   &ValidationResult{Errors: []ValidationError{}},
			expected: false,
		},
		{
			name: "has errors",
			result: &ValidationResult{
				Errors: []ValidationError{
					{Type: "conflict", Message: "duplicate"},
				},
			},
			expected: true,
		},
		{
			name: "multiple errors",
			result: &ValidationResult{
				Errors: []ValidationError{
					{Type: "conflict", Message: "duplicate"},
					{Type: "invalid", Message: "bad key"},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result.HasErrors()
			if got != tt.expected {
				t.Errorf("HasErrors() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestValidationResult_HasWarnings(t *testing.T) {
	tests := []struct {
		name     string
		result   *ValidationResult
		expected bool
	}{
		{
			name:     "no warnings",
			result:   &ValidationResult{Warnings: []ValidationError{}},
			expected: false,
		},
		{
			name: "has warnings",
			result: &ValidationResult{
				Warnings: []ValidationError{
					{Type: "warning", Message: "shadowing"},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result.HasWarnings()
			if got != tt.expected {
				t.Errorf("HasWarnings() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestValidationResult_String(t *testing.T) {
	tests := []struct {
		name     string
		result   *ValidationResult
		contains []string
	}{
		{
			name:     "no issues",
			result:   &ValidationResult{},
			contains: []string{"No issues found"},
		},
		{
			name: "only errors",
			result: &ValidationResult{
				Errors: []ValidationError{
					{Type: "conflict", Context: ContextNormal, Key: "q", Message: "duplicate"},
				},
			},
			contains: []string{"Errors (1)", "conflict", "normal", "q"},
		},
		{
			name: "only warnings",
			result: &ValidationResult{
				Warnings: []ValidationError{
					{Type: "warning", Context: ContextWebSocket, Key: "tab", Message: "shadows"},
				},
			},
			contains: []string{"Warnings (1)", "warning", "websocket", "tab"},
		},
		{
			name: "both errors and warnings",
			result: &ValidationResult{
				Errors: []ValidationError{
					{Type: "conflict", Context: ContextNormal, Key: "q", Message: "duplicate"},
				},
				Warnings: []ValidationError{
					{Type: "warning", Context: ContextWebSocket, Key: "tab", Message: "shadows"},
				},
			},
			contains: []string{"Errors (1)", "Warnings (1)", "conflict", "warning"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result.String()
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("String() output missing %q, got:\n%s", want, got)
				}
			}
		})
	}
}

func TestCheckDuplicateBindings(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name          string
		setupRegistry func() *Registry
		expectErrors  int
	}{
		{
			name: "no duplicates",
			setupRegistry: func() *Registry {
				r := NewRegistry()
				r.Register(ContextNormal, "q", ActionQuit)
				r.Register(ContextNormal, "h", ActionOpenHelp)
				return r
			},
			expectErrors: 0,
		},
		{
			name: "same key in same context gets overwritten (no error)",
			setupRegistry: func() *Registry {
				r := NewRegistry()
				r.Register(ContextNormal, "q", ActionQuit)
				r.Register(ContextNormal, "q", ActionQuit) // Overwrites, not duplicate
				return r
			},
			expectErrors: 0,
		},
		{
			name: "same key in different contexts (OK)",
			setupRegistry: func() *Registry {
				r := NewRegistry()
				r.Register(ContextNormal, "q", ActionQuit)
				r.Register(ContextModal, "q", ActionCloseModal)
				return r
			},
			expectErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ValidationResult{
				Errors:   []ValidationError{},
				Warnings: []ValidationError{},
			}

			registry := tt.setupRegistry()
			v.checkDuplicateBindings(registry, result)

			if len(result.Errors) != tt.expectErrors {
				t.Errorf("Expected %d errors, got %d", tt.expectErrors, len(result.Errors))
			}

			for _, err := range result.Errors {
				if err.Type != "conflict" {
					t.Errorf("Expected error type 'conflict', got %q", err.Type)
				}
			}
		})
	}
}

func TestCheckReservedKeys(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name           string
		setupRegistry  func() *Registry
		expectWarnings int
	}{
		{
			name: "no reserved keys",
			setupRegistry: func() *Registry {
				r := NewRegistry()
				r.Register(ContextNormal, "q", ActionQuit)
				return r
			},
			expectWarnings: 0,
		},
		{
			name: "reserved key with correct action",
			setupRegistry: func() *Registry {
				r := NewRegistry()
				r.Register(ContextGlobal, "ctrl+c", ActionQuitForce)
				return r
			},
			expectWarnings: 0,
		},
		{
			name: "reserved key rebound",
			setupRegistry: func() *Registry {
				r := NewRegistry()
				r.Register(ContextGlobal, "ctrl+c", ActionQuit) // Wrong action
				return r
			},
			expectWarnings: 1,
		},
		{
			name: "reserved key in non-global context (OK)",
			setupRegistry: func() *Registry {
				r := NewRegistry()
				r.Register(ContextNormal, "ctrl+c", ActionStopStream)
				return r
			},
			expectWarnings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ValidationResult{
				Errors:   []ValidationError{},
				Warnings: []ValidationError{},
			}

			registry := tt.setupRegistry()
			v.checkReservedKeys(registry, result)

			if len(result.Warnings) != tt.expectWarnings {
				t.Errorf("Expected %d warnings, got %d", tt.expectWarnings, len(result.Warnings))
			}
		})
	}
}

func TestCheckShadowing(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name           string
		setupRegistry  func() *Registry
		expectWarnings int
	}{
		{
			name: "no shadowing",
			setupRegistry: func() *Registry {
				r := NewRegistry()
				r.Register(ContextGlobal, "q", ActionQuit)
				r.Register(ContextNormal, "h", ActionOpenHelp)
				return r
			},
			expectWarnings: 0,
		},
		{
			name: "context shadows global with different action",
			setupRegistry: func() *Registry {
				r := NewRegistry()
				r.Register(ContextGlobal, "q", ActionQuit)
				r.Register(ContextModal, "q", ActionCloseModal)
				return r
			},
			expectWarnings: 1,
		},
		{
			name: "context uses same action as global (no warning)",
			setupRegistry: func() *Registry {
				r := NewRegistry()
				r.Register(ContextGlobal, "q", ActionQuit)
				r.Register(ContextModal, "q", ActionQuit)
				return r
			},
			expectWarnings: 0,
		},
		{
			name: "multiple shadowing",
			setupRegistry: func() *Registry {
				r := NewRegistry()
				r.Register(ContextGlobal, "q", ActionQuit)
				r.Register(ContextGlobal, "h", ActionOpenHelp)
				r.Register(ContextModal, "q", ActionCloseModal)
				r.Register(ContextModal, "h", ActionNavigateDown)
				return r
			},
			expectWarnings: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ValidationResult{
				Errors:   []ValidationError{},
				Warnings: []ValidationError{},
			}

			registry := tt.setupRegistry()
			v.checkShadowing(registry, result)

			if len(result.Warnings) != tt.expectWarnings {
				t.Errorf("Expected %d warnings, got %d", tt.expectWarnings, len(result.Warnings))
				for _, w := range result.Warnings {
					t.Logf("  Warning: %s", w.Error())
				}
			}
		})
	}
}

func TestValidateRegistry(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name           string
		setupRegistry  func() *Registry
		expectErrors   int
		expectWarnings int
	}{
		{
			name: "valid registry",
			setupRegistry: func() *Registry {
				r := NewRegistry()
				r.Register(ContextNormal, "q", ActionQuit)
				r.Register(ContextNormal, "h", ActionOpenHelp)
				r.Register(ContextModal, "esc", ActionCloseModal)
				return r
			},
			expectErrors:   0,
			expectWarnings: 0,
		},
		{
			name: "no duplicates possible in registry (map overwrites)",
			setupRegistry: func() *Registry {
				r := NewRegistry()
				r.Register(ContextNormal, "q", ActionQuit)
				r.Register(ContextNormal, "q", ActionQuit) // Just overwrites
				return r
			},
			expectErrors:   0,
			expectWarnings: 0,
		},
		{
			name: "shadowing warnings",
			setupRegistry: func() *Registry {
				r := NewRegistry()
				r.Register(ContextGlobal, "q", ActionQuit)
				r.Register(ContextModal, "q", ActionCloseModal)
				return r
			},
			expectErrors:   0,
			expectWarnings: 1,
		},
		{
			name: "reserved key rebound",
			setupRegistry: func() *Registry {
				r := NewRegistry()
				r.Register(ContextGlobal, "ctrl+c", ActionQuit)
				return r
			},
			expectErrors:   0,
			expectWarnings: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := tt.setupRegistry()
			result := v.ValidateRegistry(registry)

			if len(result.Errors) != tt.expectErrors {
				t.Errorf("Expected %d errors, got %d", tt.expectErrors, len(result.Errors))
				for _, err := range result.Errors {
					t.Logf("  Error: %s", err.Error())
				}
			}

			if len(result.Warnings) != tt.expectWarnings {
				t.Errorf("Expected %d warnings, got %d", tt.expectWarnings, len(result.Warnings))
				for _, warn := range result.Warnings {
					t.Logf("  Warning: %s", warn.Error())
				}
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name         string
		config       *Config
		expectErrors bool
		skipTest     bool
	}{
		{
			name: "valid config",
			config: &Config{
				Global: map[string]string{
					"q": "quit",
					"h": "open_help",
				},
				Normal: map[string]string{
					"j": "navigate_down",
					"k": "navigate_up",
				},
			},
			expectErrors: false,
		},
		{
			name:         "empty config",
			config:       &Config{},
			expectErrors: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipTest {
				t.Skip("Skipping test case")
			}

			result := v.ValidateConfig(tt.config)

			if tt.expectErrors && !result.HasErrors() {
				t.Error("Expected errors but got none")
			}

			if !tt.expectErrors && result.HasErrors() {
				t.Errorf("Expected no errors but got: %v", result.Errors)
			}
		})
	}
}

func TestFindConflicts(t *testing.T) {
	tests := []struct {
		name           string
		config         *Config
		expectConflict bool
	}{
		{
			name: "no conflicts",
			config: &Config{
				Normal: map[string]string{
					"q": "quit",
					"h": "open_help",
				},
			},
			expectConflict: false,
		},
		{
			name:           "empty config",
			config:         &Config{},
			expectConflict: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conflicts := FindConflicts(tt.config)

			if tt.expectConflict && len(conflicts) == 0 {
				t.Error("Expected conflicts but got none")
			}

			if !tt.expectConflict && len(conflicts) > 0 {
				t.Errorf("Expected no conflicts but got: %v", conflicts)
			}
		})
	}
}

func TestValidateKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "empty key",
			key:     "",
			wantErr: true,
		},
		{
			name:    "simple key",
			key:     "q",
			wantErr: false,
		},
		{
			name:    "multi-char key",
			key:     "esc",
			wantErr: false,
		},
		{
			name:    "ctrl modifier",
			key:     "ctrl+c",
			wantErr: false,
		},
		{
			name:    "alt modifier",
			key:     "alt+f",
			wantErr: false,
		},
		{
			name:    "shift modifier",
			key:     "shift+tab",
			wantErr: false,
		},
		{
			name:    "super modifier",
			key:     "super+k",
			wantErr: false,
		},
		{
			name:    "modifier only",
			key:     "ctrl+",
			wantErr: true,
		},
		{
			name:    "multi-key sequence",
			key:     "gg",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateKey(tt.key)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAction(t *testing.T) {
	tests := []struct {
		name      string
		actionStr string
		wantErr   bool
	}{
		{
			name:      "empty action",
			actionStr: "",
			wantErr:   true,
		},
		{
			name:      "valid action",
			actionStr: "quit",
			wantErr:   false,
		},
		{
			name:      "action with underscores",
			actionStr: "open_help",
			wantErr:   false,
		},
		{
			name:      "any non-empty string",
			actionStr: "custom_action",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAction(tt.actionStr)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestContextHierarchy(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		context      Context
		expectParent Context
	}{
		{ContextNormal, ContextGlobal},
		{ContextSearch, ContextGlobal},
		{ContextModal, ContextGlobal},
		{ContextWebSocket, ContextGlobal},
		{ContextAnalytics, ContextGlobal},
	}

	for _, tt := range tests {
		t.Run(string(tt.context), func(t *testing.T) {
			parent, exists := v.contextHierarchy[tt.context]
			if !exists {
				t.Errorf("Expected %s to have a parent in hierarchy", tt.context)
			}
			if parent != tt.expectParent {
				t.Errorf("Expected parent %s, got %s", tt.expectParent, parent)
			}
		})
	}
}

func TestMultipleValidationIssues(t *testing.T) {
	v := NewValidator()

	// Create a registry with multiple types of warnings
	r := NewRegistry()
	r.Register(ContextGlobal, "q", ActionQuit)        // Will be shadowed
	r.Register(ContextModal, "q", ActionCloseModal)   // Shadows global
	r.Register(ContextGlobal, "ctrl+c", ActionQuit)   // Reserved key rebound

	result := v.ValidateRegistry(r)

	// Should have at least one warning (shadowing or reserved key)
	if !result.HasWarnings() {
		t.Error("Expected at least one warning for shadowing or reserved key")
	}

	// Check that String() output is informative
	output := result.String()
	if !strings.Contains(output, "Warnings") {
		t.Error("Expected String() output to mention warnings")
	}

	// Should have at least 2 warnings (one for shadowing, one for reserved key)
	if len(result.Warnings) < 2 {
		t.Errorf("Expected at least 2 warnings, got %d", len(result.Warnings))
		for _, w := range result.Warnings {
			t.Logf("  Warning: %s", w.Error())
		}
	}
}
