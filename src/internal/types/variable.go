package types

import (
	"encoding/json"
	"errors"
	"fmt"
)

// UnmarshalJSON implements custom JSON unmarshaling for VariableValue
// It handles both string values and multi-value variable objects
func (v *VariableValue) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as a string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		v.StringValue = &str
		v.MultiValue = nil
		v.Interactive = false
		return nil
	}

	// Try to unmarshal as an object (could be MultiValueVariable or have Interactive field)
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err == nil {
		// Check for interactive flag
		if interactive, ok := obj["interactive"].(bool); ok {
			v.Interactive = interactive
		}

		// Check if it has multi-value options
		if options, ok := obj["options"].([]interface{}); ok && len(options) > 0 {
			var mv MultiValueVariable
			if err := json.Unmarshal(data, &mv); err == nil {
				v.StringValue = nil
				v.MultiValue = &mv
				return nil
			}
		} else if value, ok := obj["value"].(string); ok {
			// Object with string value and interactive flag
			v.StringValue = &value
			v.MultiValue = nil
			return nil
		}
	}

	return errors.New("variable value must be either a string or a multi-value object")
}

// MarshalJSON implements custom JSON marshaling for VariableValue
func (v VariableValue) MarshalJSON() ([]byte, error) {
	// If interactive flag is set, wrap in object
	if v.Interactive {
		obj := make(map[string]interface{})
		obj["interactive"] = true
		if v.StringValue != nil {
			obj["value"] = *v.StringValue
		} else if v.MultiValue != nil {
			// Merge multi-value fields into object
			obj["options"] = v.MultiValue.Options
			obj["active"] = v.MultiValue.Active
			if v.MultiValue.Description != "" {
				obj["description"] = v.MultiValue.Description
			}
			if v.MultiValue.Aliases != nil && len(v.MultiValue.Aliases) > 0 {
				obj["aliases"] = v.MultiValue.Aliases
			}
		} else {
			obj["value"] = ""
		}
		return json.Marshal(obj)
	}

	// Non-interactive: use simple representation
	if v.StringValue != nil {
		return json.Marshal(*v.StringValue)
	}
	if v.MultiValue != nil {
		return json.Marshal(v.MultiValue)
	}
	return json.Marshal("")
}

// UnmarshalYAML implements custom YAML unmarshaling for VariableValue
func (v *VariableValue) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Try to unmarshal as a string first
	var str string
	if err := unmarshal(&str); err == nil {
		v.StringValue = &str
		v.MultiValue = nil
		v.Interactive = false
		return nil
	}

	// Try to unmarshal as an object
	var obj map[string]interface{}
	if err := unmarshal(&obj); err == nil {
		// Check for interactive flag
		if interactive, ok := obj["interactive"].(bool); ok {
			v.Interactive = interactive
		}

		// Check if it has multi-value options
		if options, ok := obj["options"].([]interface{}); ok && len(options) > 0 {
			var mv MultiValueVariable
			if err := unmarshal(&mv); err == nil {
				v.StringValue = nil
				v.MultiValue = &mv
				return nil
			}
		} else if value, ok := obj["value"].(string); ok {
			// Object with string value and interactive flag
			v.StringValue = &value
			v.MultiValue = nil
			return nil
		}
	}

	return errors.New("variable value must be either a string or a multi-value object")
}

// MarshalYAML implements custom YAML marshaling for VariableValue
func (v VariableValue) MarshalYAML() (interface{}, error) {
	// If interactive flag is set, wrap in object
	if v.Interactive {
		obj := make(map[string]interface{})
		obj["interactive"] = true
		if v.StringValue != nil {
			obj["value"] = *v.StringValue
		} else if v.MultiValue != nil {
			// Merge multi-value fields into object
			obj["options"] = v.MultiValue.Options
			obj["active"] = v.MultiValue.Active
			if v.MultiValue.Description != "" {
				obj["description"] = v.MultiValue.Description
			}
			if v.MultiValue.Aliases != nil && len(v.MultiValue.Aliases) > 0 {
				obj["aliases"] = v.MultiValue.Aliases
			}
		} else {
			obj["value"] = ""
		}
		return obj, nil
	}

	// Non-interactive: use simple representation
	if v.StringValue != nil {
		return *v.StringValue, nil
	}
	if v.MultiValue != nil {
		return v.MultiValue, nil
	}
	return "", nil
}

// GetValue returns the string value for the variable
// For multi-value variables, it returns the active option
// Returns empty string if the active index is out of bounds (use Validate to check)
func (v *VariableValue) GetValue() string {
	if v.StringValue != nil {
		return *v.StringValue
	}
	if v.MultiValue != nil && v.MultiValue.Active >= 0 && v.MultiValue.Active < len(v.MultiValue.Options) {
		return v.MultiValue.Options[v.MultiValue.Active]
	}
	return ""
}

// Validate checks if the variable configuration is valid
// Returns an error if multi-value variable has invalid active index
func (v *VariableValue) Validate(varName string) error {
	if v.MultiValue != nil {
		if len(v.MultiValue.Options) == 0 {
			return fmt.Errorf("variable '%s': multi-value variable has no options", varName)
		}
		if v.MultiValue.Active < 0 {
			return fmt.Errorf("variable '%s': active index %d is negative", varName, v.MultiValue.Active)
		}
		if v.MultiValue.Active >= len(v.MultiValue.Options) {
			return fmt.Errorf("variable '%s': active index %d is out of bounds (have %d options)",
				varName, v.MultiValue.Active, len(v.MultiValue.Options))
		}
	}
	return nil
}

// SetValue sets the string value for the variable
func (v *VariableValue) SetValue(value string) {
	v.StringValue = &value
	v.MultiValue = nil
}

// IsMultiValue returns true if this is a multi-value variable
func (v *VariableValue) IsMultiValue() bool {
	return v.MultiValue != nil
}
