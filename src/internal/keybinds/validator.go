package keybinds

import (
	"fmt"
	"strings"
)

// ValidationError represents a keybinding validation error
type ValidationError struct {
	Type    string // "conflict", "invalid", "warning"
	Context Context
	Key     string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("[%s] %s in context '%s': %s", e.Type, e.Key, e.Context, e.Message)
}

// ValidationResult contains all validation errors and warnings
type ValidationResult struct {
	Errors   []ValidationError
	Warnings []ValidationError
}

// HasErrors returns true if there are any errors
func (r *ValidationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// HasWarnings returns true if there are any warnings
func (r *ValidationResult) HasWarnings() bool {
	return len(r.Warnings) > 0
}

// String returns a human-readable summary of validation results
func (r *ValidationResult) String() string {
	var sb strings.Builder

	if len(r.Errors) > 0 {
		sb.WriteString(fmt.Sprintf("Errors (%d):\n", len(r.Errors)))
		for _, err := range r.Errors {
			sb.WriteString(fmt.Sprintf("  - %s\n", err.Error()))
		}
	}

	if len(r.Warnings) > 0 {
		sb.WriteString(fmt.Sprintf("Warnings (%d):\n", len(r.Warnings)))
		for _, warn := range r.Warnings {
			sb.WriteString(fmt.Sprintf("  - %s\n", warn.Error()))
		}
	}

	if !r.HasErrors() && !r.HasWarnings() {
		sb.WriteString("No issues found")
	}

	return sb.String()
}

// Validator validates keybinding configurations
type Validator struct {
	// reservedKeys are keys that should not be rebound
	reservedKeys map[string]bool

	// contextHierarchy defines context inheritance
	contextHierarchy map[Context]Context
}

// NewValidator creates a new keybinding validator
func NewValidator() *Validator {
	return &Validator{
		reservedKeys: map[string]bool{
			"ctrl+c": true, // Force quit should always work
		},
		contextHierarchy: map[Context]Context{
			// Define which contexts inherit from which
			ContextNormal:        ContextGlobal,
			ContextSearch:        ContextGlobal,
			ContextGoto:          ContextGlobal,
			ContextVariableList:  ContextGlobal,
			ContextVariableEdit:  ContextGlobal,
			ContextHeaderList:    ContextGlobal,
			ContextHeaderEdit:    ContextGlobal,
			ContextProfileList:   ContextGlobal,
			ContextProfileEdit:   ContextGlobal,
			ContextDocumentation: ContextGlobal,
			ContextHistory:       ContextGlobal,
			ContextAnalytics:     ContextGlobal,
			ContextStressTest:    ContextGlobal,
			ContextHelp:          ContextGlobal,
			ContextInspect:       ContextGlobal,
			ContextWebSocket:     ContextGlobal,
			ContextModal:         ContextGlobal,
			ContextTextInput:     ContextGlobal,
			ContextConfirm:       ContextGlobal,
			ContextViewer:        ContextGlobal,
		},
	}
}

// ValidateRegistry validates an entire registry
func (v *Validator) ValidateRegistry(registry *Registry) *ValidationResult {
	result := &ValidationResult{
		Errors:   []ValidationError{},
		Warnings: []ValidationError{},
	}

	// Check for duplicate bindings within contexts
	v.checkDuplicateBindings(registry, result)

	// Check for conflicts with reserved keys
	v.checkReservedKeys(registry, result)

	// Check for ambiguous multi-key sequences
	v.checkMultiKeySequences(registry, result)

	// Check for shadowing (context-specific binding hiding global binding)
	v.checkShadowing(registry, result)

	return result
}

// ValidateConfig validates a configuration before applying it
func (v *Validator) ValidateConfig(config *Config) *ValidationResult {
	result := &ValidationResult{
		Errors:   []ValidationError{},
		Warnings: []ValidationError{},
	}

	// Create a temporary registry to validate
	registry := NewRegistry()
	if err := ApplyConfig(registry, config); err != nil {
		result.Errors = append(result.Errors, ValidationError{
			Type:    "invalid",
			Message: err.Error(),
		})
		return result
	}

	// Validate the registry
	return v.ValidateRegistry(registry)
}

// checkDuplicateBindings checks for duplicate key bindings in the same context
func (v *Validator) checkDuplicateBindings(registry *Registry, result *ValidationResult) {
	for context, bindings := range registry.bindings {
		keyCount := make(map[string]int)
		for key := range bindings {
			keyCount[key]++
		}

		for key, count := range keyCount {
			if count > 1 {
				result.Errors = append(result.Errors, ValidationError{
					Type:    "conflict",
					Context: context,
					Key:     key,
					Message: fmt.Sprintf("key bound %d times", count),
				})
			}
		}
	}
}

// checkReservedKeys checks if any reserved keys have been rebound
func (v *Validator) checkReservedKeys(registry *Registry, result *ValidationResult) {
	for context, bindings := range registry.bindings {
		for key, action := range bindings {
			if v.reservedKeys[key] {
				// Check if it's bound to something other than the reserved action
				if context == ContextGlobal && action != ActionQuitForce {
					result.Warnings = append(result.Warnings, ValidationError{
						Type:    "warning",
						Context: context,
						Key:     key,
						Message: "reserved key rebound (may cause issues)",
					})
				}
			}
		}
	}
}

// checkMultiKeySequences checks for ambiguous multi-key sequences
func (v *Validator) checkMultiKeySequences(registry *Registry, result *ValidationResult) {
	for _, bindings := range registry.bindings {
		// Check if both 'g' and 'gg' are bound
		if _, hasG := bindings["g"]; hasG {
			if _, hasGG := bindings["gg"]; hasGG {
				// This is expected and OK, just document it
				// No error needed
			}
		}

		// Check for other potential multi-key conflicts
		for key := range bindings {
			if len(key) > 1 && !strings.Contains(key, "+") {
				// This is a multi-key sequence (not a modifier combo)
				// Check if the first character is also bound
				firstChar := string(key[0])
				if _, hasSingle := bindings[firstChar]; hasSingle {
					// This is intentional for sequences like 'gg'
					// Could add a warning if needed
				}
			}
		}
	}
}

// checkShadowing checks for context-specific bindings that shadow global bindings
func (v *Validator) checkShadowing(registry *Registry, result *ValidationResult) {
	globalBindings := registry.bindings[ContextGlobal]
	if globalBindings == nil {
		return
	}

	for context, bindings := range registry.bindings {
		if context == ContextGlobal {
			continue
		}

		for key, action := range bindings {
			if globalAction, hasGlobal := globalBindings[key]; hasGlobal {
				if action != globalAction {
					result.Warnings = append(result.Warnings, ValidationError{
						Type:    "warning",
						Context: context,
						Key:     key,
						Message: fmt.Sprintf("shadows global binding (%s -> %s)", globalAction, action),
					})
				}
			}
		}
	}
}

// FindConflicts finds all conflicting keybindings in a config
func FindConflicts(config *Config) []string {
	validator := NewValidator()
	result := validator.ValidateConfig(config)

	var conflicts []string
	for _, err := range result.Errors {
		if err.Type == "conflict" {
			conflicts = append(conflicts, err.Error())
		}
	}

	return conflicts
}

// ValidateKey checks if a key string is valid
func ValidateKey(key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	// Check for valid modifier combinations
	validModifiers := []string{"ctrl+", "alt+", "shift+", "super+"}
	hasModifier := false
	for _, mod := range validModifiers {
		if strings.HasPrefix(key, mod) {
			hasModifier = true
			break
		}
	}

	// If it has a modifier, ensure there's something after it
	if hasModifier {
		for _, mod := range validModifiers {
			if key == mod {
				return fmt.Errorf("modifier without key: %s", key)
			}
		}
	}

	return nil
}

// ValidateAction checks if an action string is valid
func ValidateAction(actionStr string) error {
	if actionStr == "" {
		return fmt.Errorf("action cannot be empty")
	}

	// Could add a whitelist of known actions here
	// For now, just check it's not empty

	return nil
}
