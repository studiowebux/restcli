package keybinds

import (
	"fmt"
	"strings"
)

// Binding represents a keybinding mapping
type Binding struct {
	Key     string
	Action  Action
	Context Context
}

// Registry manages keybinding mappings and matching
type Registry struct {
	// bindings maps context -> key -> action
	bindings map[Context]map[string]Action

	// multiKeyState tracks multi-key sequences (like 'gg' in vim)
	multiKeyState map[Context]string
}

// NewRegistry creates a new keybinding registry
func NewRegistry() *Registry {
	return &Registry{
		bindings:      make(map[Context]map[string]Action),
		multiKeyState: make(map[Context]string),
	}
}

// Register adds a keybinding to the registry
func (r *Registry) Register(context Context, key string, action Action) {
	if r.bindings[context] == nil {
		r.bindings[context] = make(map[string]Action)
	}
	r.bindings[context][key] = action
}

// RegisterMultiple registers multiple keybindings for the same action
func (r *Registry) RegisterMultiple(context Context, keys []string, action Action) {
	for _, key := range keys {
		r.Register(context, key, action)
	}
}

// Match attempts to match a key to an action in the given context
// Returns the action and whether a match was found
// Contexts are checked in priority order: specific context -> global
func (r *Registry) Match(context Context, key string) (Action, bool) {
	// First check for exact match in specific context
	if contextBindings, ok := r.bindings[context]; ok {
		if action, ok := contextBindings[key]; ok {
			return action, true
		}
	}

	// Then check global context
	if globalBindings, ok := r.bindings[ContextGlobal]; ok {
		if action, ok := globalBindings[key]; ok {
			return action, true
		}
	}

	return "", false
}

// MatchMultiKey handles multi-key sequences like 'gg' for go-to-top
// Returns the action, whether it's a complete match, and whether it's a partial match
func (r *Registry) MatchMultiKey(context Context, key string) (Action, bool, bool) {
	// Check if we have a pending multi-key state
	if prevKey, hasPending := r.multiKeyState[context]; hasPending {
		// Try to match the sequence
		sequence := prevKey + key

		// Clear state first
		delete(r.multiKeyState, context)

		// Check for match
		if action, ok := r.Match(context, sequence); ok {
			return action, true, false
		}

		// No match for sequence, return no match
		return "", false, false
	}

	// Check if this key could start a sequence (currently only 'g' for 'gg')
	if key == "g" {
		// Mark this as a potential multi-key start
		r.multiKeyState[context] = key
		return "", false, true // Partial match
	}

	// Regular single-key match
	action, ok := r.Match(context, key)
	return action, ok, false
}

// ClearMultiKeyState clears any pending multi-key state for a context
func (r *Registry) ClearMultiKeyState(context Context) {
	delete(r.multiKeyState, context)
}

// GetBinding returns the key(s) bound to an action in a context
func (r *Registry) GetBinding(context Context, action Action) []string {
	var keys []string

	// Check specific context
	if contextBindings, ok := r.bindings[context]; ok {
		for key, act := range contextBindings {
			if act == action {
				keys = append(keys, key)
			}
		}
	}

	// If not found, check global
	if len(keys) == 0 {
		if globalBindings, ok := r.bindings[ContextGlobal]; ok {
			for key, act := range globalBindings {
				if act == action {
					keys = append(keys, key)
				}
			}
		}
	}

	return keys
}

// GetBindingString returns a human-readable string of keys bound to an action
func (r *Registry) GetBindingString(context Context, action Action) string {
	keys := r.GetBinding(context, action)
	if len(keys) == 0 {
		return "unbound"
	}
	return strings.Join(keys, ", ")
}

// ListBindings returns all bindings for a context
func (r *Registry) ListBindings(context Context) []Binding {
	var bindings []Binding

	// Add context-specific bindings
	if contextBindings, ok := r.bindings[context]; ok {
		for key, action := range contextBindings {
			bindings = append(bindings, Binding{
				Key:     key,
				Action:  action,
				Context: context,
			})
		}
	}

	// Add global bindings
	if globalBindings, ok := r.bindings[ContextGlobal]; ok {
		for key, action := range globalBindings {
			bindings = append(bindings, Binding{
				Key:     key,
				Action:  action,
				Context: ContextGlobal,
			})
		}
	}

	return bindings
}

// Validate checks for conflicts and invalid bindings
func (r *Registry) Validate() error {
	// Check for duplicate bindings within the same context
	for context, contextBindings := range r.bindings {
		keyCount := make(map[string]int)
		for key := range contextBindings {
			keyCount[key]++
			if keyCount[key] > 1 {
				return fmt.Errorf("duplicate binding for key '%s' in context '%s'", key, context)
			}
		}
	}

	return nil
}

// HasBinding checks if a key is bound in a context
func (r *Registry) HasBinding(context Context, key string) bool {
	if contextBindings, ok := r.bindings[context]; ok {
		if _, ok := contextBindings[key]; ok {
			return true
		}
	}

	// Check global
	if globalBindings, ok := r.bindings[ContextGlobal]; ok {
		if _, ok := globalBindings[key]; ok {
			return true
		}
	}

	return false
}

// Clone creates a deep copy of the registry
func (r *Registry) Clone() *Registry {
	clone := NewRegistry()

	for context, contextBindings := range r.bindings {
		for key, action := range contextBindings {
			clone.Register(context, key, action)
		}
	}

	return clone
}

// Merge combines bindings from another registry, with other taking precedence
func (r *Registry) Merge(other *Registry) {
	for context, contextBindings := range other.bindings {
		for key, action := range contextBindings {
			r.Register(context, key, action)
		}
	}
}
