package mock

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadConfig loads a mock configuration from a file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config file format: %s (use .yaml, .yml, or .json)", ext)
	}

	// Validate config
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

// validateConfig validates the mock configuration
func validateConfig(config *Config) error {
	if len(config.Routes) == 0 {
		return fmt.Errorf("no routes defined")
	}

	for i, route := range config.Routes {
		if route.Method == "" {
			return fmt.Errorf("route %d: method is required", i)
		}
		if route.Path == "" {
			return fmt.Errorf("route %d: path is required", i)
		}
		if route.PathType != "" && route.PathType != "exact" && route.PathType != "prefix" && route.PathType != "regex" {
			return fmt.Errorf("route %d: pathType must be 'exact', 'prefix', or 'regex'", i)
		}
	}

	return nil
}

// SaveConfig saves a mock configuration to a file
func SaveConfig(config *Config, path string) error {
	var data []byte
	var err error

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		data, err = yaml.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to marshal YAML: %w", err)
		}
	case ".json":
		data, err = json.MarshalIndent(config, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
	default:
		return fmt.Errorf("unsupported config file format: %s (use .yaml, .yml, or .json)", ext)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
