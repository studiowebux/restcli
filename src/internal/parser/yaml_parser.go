package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/studiowebux/restcli/internal/converter"
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

// IsOpenAPISpec checks if the data is an OpenAPI specification
func IsOpenAPISpec(data []byte) bool {
	var probe struct {
		OpenAPI string `json:"openapi" yaml:"openapi"`
		Swagger string `json:"swagger" yaml:"swagger"`
	}

	// Try YAML first (also handles JSON)
	if err := yaml.Unmarshal(data, &probe); err == nil {
		return probe.OpenAPI != "" || probe.Swagger != ""
	}
	return false
}

// parseOpenAPISnippet converts an OpenAPI spec to HttpRequest(s)
func parseOpenAPISnippet(data []byte) ([]types.HttpRequest, error) {
	// Parse the OpenAPI spec
	var spec converter.OpenAPISpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		// Try JSON
		if err := json.Unmarshal(data, &spec); err != nil {
			return nil, fmt.Errorf("failed to parse OpenAPI spec: %w", err)
		}
	}

	// Get base URL from servers
	baseURL := "{{baseUrl}}"
	if len(spec.Servers) > 0 && spec.Servers[0].URL != "" {
		baseURL = spec.Servers[0].URL
	}

	// Convert each operation to HttpRequest
	var requests []types.HttpRequest

	for path, pathItem := range spec.Paths {
		operations := map[string]*converter.OpenAPIOperation{
			"GET":    pathItem.Get,
			"POST":   pathItem.Post,
			"PUT":    pathItem.Put,
			"DELETE": pathItem.Delete,
			"PATCH":  pathItem.Patch,
		}

		for method, operation := range operations {
			if operation == nil {
				continue
			}

			req := converter.OperationToHttpRequest(method, path, operation, baseURL, &spec)
			requests = append(requests, req)
		}
	}

	if len(requests) == 0 {
		return nil, fmt.Errorf("no operations found in OpenAPI spec")
	}

	return requests, nil
}

// DetectFormat detects whether a file is traditional .http format, YAML/JSON, or OpenAPI
func DetectFormat(filePath string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	// Extension-based detection
	switch ext {
	case ".yaml", ".yml", ".json":
		// For YAML/JSON files, check if it's OpenAPI
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", err
		}

		if IsOpenAPISpec(data) {
			return "openapi", nil
		}

		if ext == ".json" {
			return "json", nil
		}
		return "yaml", nil

	case ".http":
		// For .http files, we need to peek at the content
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", err
		}

		// If it starts with YAML document separator or looks like JSON, treat as structured
		content := strings.TrimSpace(string(data))
		if strings.HasPrefix(content, "---") || strings.HasPrefix(content, "{") || strings.HasPrefix(content, "[") {
			// Check if it's OpenAPI
			if IsOpenAPISpec(data) {
				return "openapi", nil
			}

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
	case "openapi":
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		return parseOpenAPISnippet(data)
	case "yaml", "json":
		return ParseYAMLFile(filePath)
	case "http":
		return ParseHTTPFile(filePath)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", format)
	}
}
