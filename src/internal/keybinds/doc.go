/*
Package keybinds provides customizable keyboard binding management.

# Overview

The keybinds package implements a hierarchical, context-aware keyboard
binding system that allows users to customize all keybindings through
configuration files.

# Key Concepts

Context Hierarchy:
  - Global: Bindings available everywhere
  - Normal: Main application view
  - Modal: Generic modal context
  - Specific: Modal-specific contexts (history, analytics, etc.)

Keys shadow from specific → modal → global. If a key is bound in a
specific context, it overrides the global binding.

Action System:
  - Actions are constants (ActionQuit, ActionExecute, etc.)
  - Keys map to actions within contexts
  - Same action can have different keys in different contexts

# Components

Registry (registry.go):
  - Central storage for keybindings
  - Context-aware key matching
  - Multi-key sequence support (e.g., "gg" for go-to-top)
  - Thread-safe for concurrent access

Validator (validator.go):
  - Validates keybinding configurations
  - Detects conflicts and duplicates
  - Warns about shadowing
  - Protects reserved keys

Defaults (defaults.go):
  - Default keybinding configuration
  - Covers all contexts and actions
  - Used when no custom config exists

# Configuration File Format

Keybindings are stored in JSON format:

	{
	  "global": {
	    "q": "quit",
	    "ctrl+c": "quit",
	    "?": "help"
	  },
	  "normal": {
	    "enter": "execute",
	    "e": "edit",
	    "d": "delete"
	  },
	  "history": {
	    "enter": "load_from_history",
	    "d": "delete_history_entry"
	  }
	}

# Reserved Keys

Certain keys are reserved for core functionality:
  - q: Quit/close (global)
  - ctrl+c: Interrupt/quit (global)
  - esc: Close modal/cancel (global)
  - enter: Confirm/execute (context-dependent)

Rebinding reserved keys generates warnings.

# Multi-Key Sequences

The registry supports multi-key sequences:
  - "gg": Go to top
  - "G": Go to bottom (capital G)
  - Timeout-based sequence detection

# Validation

The validator checks for:
  - Duplicate bindings in same context
  - Invalid key formats
  - Invalid action names
  - Shadowing (warnings, not errors)
  - Reserved key rebindings (warnings)

# Example Usage

	// Create registry with defaults
	registry := NewRegistry()
	LoadDefaults(registry)

	// Override from user config
	config, err := LoadConfig("~/.config/restcli/keybinds.json")
	if err == nil {
		ApplyConfig(registry, config)
	}

	// Validate configuration
	validator := NewValidator(registry)
	result := validator.ValidateRegistry()
	if result.HasErrors() {
		fmt.Println("Configuration errors:")
		fmt.Println(result.String())
	}

	// Match keys during runtime
	if action := registry.Match(ContextNormal, "enter"); action != "" {
		// Handle action
	}

# Thread Safety

The Registry is thread-safe for concurrent reads.
Writes (Register, RegisterMultiple) should be done during initialization.

# Extension

To add new actions:
  1. Define action constant (e.g., ActionNewFeature)
  2. Add to defaults.go
  3. Handle in TUI key handlers
  4. Document in user documentation

# Best Practices

  - Use descriptive action names
  - Maintain consistency across contexts
  - Avoid rebinding core navigation keys (hjkl, arrows)
  - Test configuration with validator before deployment
  - Provide fallbacks for missing bindings
*/
package keybinds
