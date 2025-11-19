package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/studiowebux/restcli/internal/types"
	"gopkg.in/yaml.v3"
)

// ParseYAMLFile parses a YAML or JSON file containing HTTP requests
func ParseYAMLFile(filePath string) ([]types.HttpRequest, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(filePath))

	// Try to parse as JSON first if extension is .json
	if ext == ".json" {
		return parseJSON(data)
	}

	// Otherwise parse as YAML (which also handles JSON)
	return parseYAML(data)
}

// parseJSON parses JSON format
func parseJSON(data []byte) ([]types.HttpRequest, error) {
	// Try to unmarshal as array first
	var requests []types.HttpRequest
	if err := json.Unmarshal(data, &requests); err == nil {
		return requests, nil
	}

	// Try to unmarshal as single request
	var request types.HttpRequest
	if err := json.Unmarshal(data, &request); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return []types.HttpRequest{request}, nil
}

// parseYAML parses YAML format
func parseYAML(data []byte) ([]types.HttpRequest, error) {
	// Try to unmarshal as array first
	var requests []types.HttpRequest
	if err := yaml.Unmarshal(data, &requests); err == nil {
		// Validate that we actually got an array
		if len(requests) > 0 || string(data) == "[]" {
			return requests, nil
		}
	}

	// Try to unmarshal as single request
	var request types.HttpRequest
	if err := yaml.Unmarshal(data, &request); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return []types.HttpRequest{request}, nil
}

// DetectFormat detects whether a file is traditional .http format or YAML/JSON
func DetectFormat(filePath string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	// Extension-based detection
	switch ext {
	case ".yaml", ".yml":
		return "yaml", nil
	case ".json":
		return "json", nil
	case ".http":
		// For .http files, we need to peek at the content
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", err
		}

		// If it starts with YAML document separator or looks like JSON, treat as structured
		content := strings.TrimSpace(string(data))
		if strings.HasPrefix(content, "---") || strings.HasPrefix(content, "{") || strings.HasPrefix(content, "[") {
			// Try to parse as YAML/JSON
			if strings.HasPrefix(content, "{") || strings.HasPrefix(content, "[") {
				return "json", nil
			}
			return "yaml", nil
		}

		return "http", nil
	default:
		return "http", nil
	}
}

// Parse is the main entry point for parsing any supported file format
func Parse(filePath string) ([]types.HttpRequest, error) {
	format, err := DetectFormat(filePath)
	if err != nil {
		return nil, err
	}

	switch format {
	case "yaml", "json":
		return ParseYAMLFile(filePath)
	case "http":
		return ParseHTTPFile(filePath)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", format)
	}
}
