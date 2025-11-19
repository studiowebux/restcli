package types

import (
	"encoding/json"
	"errors"
)

// UnmarshalJSON implements custom JSON unmarshaling for VariableValue
// It handles both string values and multi-value variable objects
func (v *VariableValue) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as a string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		v.StringValue = &str
		v.MultiValue = nil
		return nil
	}

	// Try to unmarshal as a MultiValueVariable
	var mv MultiValueVariable
	if err := json.Unmarshal(data, &mv); err == nil {
		v.StringValue = nil
		v.MultiValue = &mv
		return nil
	}

	return errors.New("variable value must be either a string or a multi-value object")
}

// MarshalJSON implements custom JSON marshaling for VariableValue
func (v VariableValue) MarshalJSON() ([]byte, error) {
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
		return nil
	}

	// Try to unmarshal as a MultiValueVariable
	var mv MultiValueVariable
	if err := unmarshal(&mv); err == nil {
		v.StringValue = nil
		v.MultiValue = &mv
		return nil
	}

	return errors.New("variable value must be either a string or a multi-value object")
}

// MarshalYAML implements custom YAML marshaling for VariableValue
func (v VariableValue) MarshalYAML() (interface{}, error) {
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
func (v *VariableValue) GetValue() string {
	if v.StringValue != nil {
		return *v.StringValue
	}
	if v.MultiValue != nil && v.MultiValue.Active >= 0 && v.MultiValue.Active < len(v.MultiValue.Options) {
		return v.MultiValue.Options[v.MultiValue.Active]
	}
	return ""
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
