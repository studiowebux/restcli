package converter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/studiowebux/restcli/internal/types"
	"gopkg.in/yaml.v3"
)

// OpenAPI2HttpOptions contains options for openapi2http conversion
type OpenAPI2HttpOptions struct {
	SpecPath   string
	OutputDir  string
	OrganizeBy string // tags, paths, or flat
	Format     string // http, json, yaml (default: http)
}

// OpenAPISpec represents a simplified OpenAPI 3.0 specification
type OpenAPISpec struct {
	OpenAPI    string                       `json:"openapi" yaml:"openapi"`
	Info       OpenAPIInfo                  `json:"info" yaml:"info"`
	Servers    []OpenAPIServer              `json:"servers,omitempty" yaml:"servers,omitempty"`
	Paths      map[string]OpenAPIPathItem   `json:"paths" yaml:"paths"`
	Components *OpenAPIComponents           `json:"components,omitempty" yaml:"components,omitempty"`
}

// OpenAPIComponents represents reusable components (schemas, etc.)
type OpenAPIComponents struct {
	Schemas map[string]map[string]interface{} `json:"schemas,omitempty" yaml:"schemas,omitempty"`
}

type OpenAPIInfo struct {
	Title       string `json:"title" yaml:"title"`
	Version     string `json:"version" yaml:"version"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

type OpenAPIServer struct {
	URL         string `json:"url" yaml:"url"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

type OpenAPIPathItem struct {
	Get    *OpenAPIOperation `json:"get,omitempty" yaml:"get,omitempty"`
	Post   *OpenAPIOperation `json:"post,omitempty" yaml:"post,omitempty"`
	Put    *OpenAPIOperation `json:"put,omitempty" yaml:"put,omitempty"`
	Delete *OpenAPIOperation `json:"delete,omitempty" yaml:"delete,omitempty"`
	Patch  *OpenAPIOperation `json:"patch,omitempty" yaml:"patch,omitempty"`
}

type OpenAPIOperation struct {
	Summary     string                    `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string                    `json:"description,omitempty" yaml:"description,omitempty"`
	Tags        []string                  `json:"tags,omitempty" yaml:"tags,omitempty"`
	Parameters  []OpenAPIParameter        `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody *OpenAPIRequestBody       `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses   map[string]OpenAPIResponse `json:"responses,omitempty" yaml:"responses,omitempty"`
}

type OpenAPIParameter struct {
	Name        string                 `json:"name" yaml:"name"`
	In          string                 `json:"in" yaml:"in"` // query, path, header, cookie
	Required    bool                   `json:"required,omitempty" yaml:"required,omitempty"`
	Description string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Schema      map[string]interface{} `json:"schema,omitempty" yaml:"schema,omitempty"`
}

type OpenAPIRequestBody struct {
	Description string                           `json:"description,omitempty" yaml:"description,omitempty"`
	Required    bool                             `json:"required,omitempty" yaml:"required,omitempty"`
	Content     map[string]OpenAPIMediaType      `json:"content,omitempty" yaml:"content,omitempty"`
}

type OpenAPIMediaType struct {
	Schema  map[string]interface{} `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example interface{}            `json:"example,omitempty" yaml:"example,omitempty"`
}

type OpenAPIResponse struct {
	Description string                      `json:"description" yaml:"description"`
	Content     map[string]OpenAPIMediaType `json:"content,omitempty" yaml:"content,omitempty"`
}

// Openapi2Http converts an OpenAPI spec to .http files
func Openapi2Http(opts OpenAPI2HttpOptions) error {
	// Load spec
	spec, err := loadOpenAPISpec(opts.SpecPath)
	if err != nil {
		return fmt.Errorf("failed to load OpenAPI spec: %w", err)
	}

	// Create output directory
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Determine format (default to http)
	format := opts.Format
	if format == "" {
		format = "http"
	}

	// Generate files
	count, err := generateHttpFiles(spec, opts.OutputDir, opts.OrganizeBy, format)
	if err != nil {
		return fmt.Errorf("failed to generate files: %w", err)
	}

	ext := ".http"
	if format == "json" {
		ext = ".json"
	} else if format == "yaml" {
		ext = ".yaml"
	}

	fmt.Fprintf(os.Stderr, "Generated %d %s files in %s\n", count, ext, opts.OutputDir)
	return nil
}

// loadOpenAPISpec loads an OpenAPI spec from a file or URL
func loadOpenAPISpec(path string) (*OpenAPISpec, error) {
	var data []byte
	var err error

	// Check if it's a URL
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		resp, err := http.Get(path)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch spec from URL: %w", err)
		}
		defer resp.Body.Close()

		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
	} else {
		// Read from file
		data, err = os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read spec file: %w", err)
		}
	}

	// Parse as JSON or YAML
	var spec OpenAPISpec

	// Try JSON first
	if err := json.Unmarshal(data, &spec); err != nil {
		// Try YAML
		if err := yaml.Unmarshal(data, &spec); err != nil {
			return nil, fmt.Errorf("failed to parse spec as JSON or YAML: %w", err)
		}
	}

	return &spec, nil
}

// generateHttpFiles generates .http files from the spec
func generateHttpFiles(spec *OpenAPISpec, outputDir, organizeBy, format string) (int, error) {
	count := 0

	// Get base URL
	baseURL := "{{baseUrl}}"
	if len(spec.Servers) > 0 {
		baseURL = "{{baseUrl}}" // We'll suggest it in variables
	}

	// Determine file extension based on format
	ext := ".http"
	if format == "json" {
		ext = ".json"
	} else if format == "yaml" {
		ext = ".yaml"
	}

	// Iterate through paths
	for path, pathItem := range spec.Paths {
		operations := map[string]*OpenAPIOperation{
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

			// Determine filename and directory
			dir, filename := getOutputPath(outputDir, path, method, operation, organizeBy)

			// Replace extension based on format
			filename = strings.TrimSuffix(filename, ".http") + ext

			// Ensure directory exists
			if err := os.MkdirAll(dir, 0755); err != nil {
				return count, err
			}

			fullPath := filepath.Join(dir, filename)

			// Generate content based on format
			var content string
			if format == "json" || format == "yaml" {
				httpReq := OperationToHttpRequest(method, path, operation, baseURL, spec)
				if format == "json" {
					data, err := json.MarshalIndent(httpReq, "", "  ")
					if err != nil {
						return count, err
					}
					content = string(data)
				} else {
					data, err := yaml.Marshal(httpReq)
					if err != nil {
						return count, err
					}
					content = string(data)
				}
			} else {
				content = generateOperationHttpFile(method, path, operation, baseURL, spec)
			}

			// Write file
			if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
				return count, err
			}

			count++
		}
	}

	return count, nil
}

// OperationToHttpRequest converts an OpenAPI operation to types.HttpRequest
func OperationToHttpRequest(method, path string, operation *OpenAPIOperation, baseURL string, spec *OpenAPISpec) types.HttpRequest {
	// Build URL with path parameters
	url := baseURL + path

	// First, convert explicitly defined path parameters
	for _, param := range operation.Parameters {
		if param.In == "path" {
			url = strings.Replace(url, "{"+param.Name+"}", "{{"+param.Name+"}}", 1)
		}
	}

	// Also convert any remaining {param} patterns in the path that weren't explicitly defined
	// This handles cases where path parameters are in the URL but not in the parameters array
	// Only convert single braces {param}, not already converted {{param}}
	result := ""
	remaining := url
	for {
		start := strings.Index(remaining, "{")
		if start == -1 {
			result += remaining
			break
		}

		// Check if it's already a double brace
		if start+1 < len(remaining) && remaining[start+1] == '{' {
			// Skip over {{
			end := strings.Index(remaining[start+2:], "}}")
			if end == -1 {
				result += remaining
				break
			}
			result += remaining[:start+2+end+2]
			remaining = remaining[start+2+end+2:]
			continue
		}

		// Find closing single brace
		end := strings.Index(remaining[start:], "}")
		if end == -1 {
			result += remaining
			break
		}

		paramName := remaining[start+1 : start+end]
		result += remaining[:start] + "{{" + paramName + "}}"
		remaining = remaining[start+end+1:]
	}
	url = result

	// Build headers
	headers := make(map[string]string)
	for _, param := range operation.Parameters {
		if param.In == "header" {
			headers[param.Name] = "{{" + param.Name + "}}"
		}
	}

	// Add Content-Type if there's a body
	var body string
	if operation.RequestBody != nil {
		for contentType, mediaType := range operation.RequestBody.Content {
			headers["Content-Type"] = contentType
			if mediaType.Example != nil {
				if exampleJSON, err := json.MarshalIndent(mediaType.Example, "", "  "); err == nil {
					body = string(exampleJSON)
				}
			} else if mediaType.Schema != nil {
				example := generateExampleFromSchema(mediaType.Schema)
				if exampleJSON, err := json.MarshalIndent(example, "", "  "); err == nil {
					body = string(exampleJSON)
				}
			}
			break
		}
	}

	// Build documentation
	var doc *types.Documentation
	if operation.Summary != "" || operation.Description != "" || len(operation.Tags) > 0 || len(operation.Parameters) > 0 || len(operation.Responses) > 0 {
		doc = &types.Documentation{
			Description: operation.Summary,
			Tags:        operation.Tags,
		}
		if operation.Description != "" && operation.Description != operation.Summary {
			doc.Description = operation.Description
		}

		// Add parameters
		if len(operation.Parameters) > 0 {
			doc.Parameters = make([]types.Parameter, 0, len(operation.Parameters))
			for _, param := range operation.Parameters {
				docParam := types.Parameter{
					Name:        param.Name,
					Required:    param.Required,
					Description: param.Description,
				}
				// Extract type from schema
				if param.Schema != nil {
					if paramType, ok := param.Schema["type"].(string); ok {
						docParam.Type = paramType
					}
				}
				doc.Parameters = append(doc.Parameters, docParam)
			}
		}

		// Add responses
		if len(operation.Responses) > 0 {
			doc.Responses = make([]types.Response, 0, len(operation.Responses))
			for code, resp := range operation.Responses {
				docResp := types.Response{
					Code:        code,
					Description: resp.Description,
				}

				// Extract content type and fields from response schema
				for contentType, mediaType := range resp.Content {
					docResp.ContentType = contentType

					// Try to extract fields from schema
					if mediaType.Schema != nil {
						if props, ok := mediaType.Schema["properties"].(map[string]interface{}); ok {
							docResp.Fields = make([]types.ResponseField, 0)
							for fieldName, fieldSchema := range props {
								if fieldMap, ok := fieldSchema.(map[string]interface{}); ok {
									field := types.ResponseField{
										Name: fieldName,
									}
									if fieldType, ok := fieldMap["type"].(string); ok {
										field.Type = fieldType
									}
									if fieldDesc, ok := fieldMap["description"].(string); ok {
										field.Description = fieldDesc
									}
									docResp.Fields = append(docResp.Fields, field)
								}
							}
						}
					}
					break // Only use first content type
				}

				doc.Responses = append(doc.Responses, docResp)
			}
		}
	}

	// Determine name
	name := operation.Summary
	if name == "" {
		name = method + " " + path
	}

	return types.HttpRequest{
		Name:          name,
		Method:        method,
		URL:           url,
		Headers:       headers,
		Body:          body,
		Documentation: doc,
	}
}

// getOutputPath determines the output directory and filename
func getOutputPath(baseDir, path string, method string, operation *OpenAPIOperation, organizeBy string) (string, string) {
	var dir string
	var filename string

	switch organizeBy {
	case "tags":
		if len(operation.Tags) > 0 {
			tag := sanitizeFilename(operation.Tags[0])
			dir = filepath.Join(baseDir, tag)
		} else {
			dir = filepath.Join(baseDir, "untagged")
		}
		filename = sanitizeFilename(path) + ".http"

	case "paths":
		// Mirror the exact API path structure
		pathParts := strings.Split(strings.Trim(path, "/"), "/")

		if len(pathParts) > 0 {
			// Use all parts except the last one for directory structure
			// Filter out path parameters (e.g., {id})
			dirParts := []string{}
			for i := 0; i < len(pathParts)-1; i++ {
				part := pathParts[i]
				// Skip path parameters
				if !strings.HasPrefix(part, "{") {
					dirParts = append(dirParts, part)
				}
			}

			if len(dirParts) > 0 {
				dir = filepath.Join(baseDir, filepath.Join(dirParts...))
			} else {
				dir = baseDir
			}

			// Filename: {method}_{last_segment}.http
			lastSegment := pathParts[len(pathParts)-1]
			// Remove path parameter braces
			lastSegment = strings.ReplaceAll(lastSegment, "{", "")
			lastSegment = strings.ReplaceAll(lastSegment, "}", "")
			filename = strings.ToLower(method) + "_" + lastSegment + ".http"
		} else {
			dir = baseDir
			filename = strings.ToLower(method) + "_root.http"
		}

	case "flat":
		fallthrough
	default:
		dir = baseDir
		filename = sanitizeFilename(path) + ".http"
	}

	return dir, filename
}

// sanitizeFilename creates a safe filename from a path
func sanitizeFilename(path string) string {
	// Remove leading/trailing slashes
	path = strings.Trim(path, "/")

	// Replace slashes with underscores
	path = strings.ReplaceAll(path, "/", "_")

	// Replace curly braces
	path = strings.ReplaceAll(path, "{", "")
	path = strings.ReplaceAll(path, "}", "")

	// Replace other special characters
	path = strings.ReplaceAll(path, " ", "_")
	path = strings.ReplaceAll(path, ":", "_")

	if path == "" {
		path = "root"
	}

	return path
}

// generateOperationHttpFile generates .http content for an operation
func generateOperationHttpFile(method, path string, operation *OpenAPIOperation, baseURL string, spec *OpenAPISpec) string {
	var sb strings.Builder

	// Add documentation
	if operation.Summary != "" {
		sb.WriteString(fmt.Sprintf("### %s\n", operation.Summary))
	} else {
		sb.WriteString(fmt.Sprintf("### %s %s\n", method, path))
	}

	if operation.Description != "" {
		sb.WriteString(fmt.Sprintf("# @description %s\n", operation.Description))
	}

	if len(operation.Tags) > 0 {
		for _, tag := range operation.Tags {
			sb.WriteString(fmt.Sprintf("# @tag %s\n", tag))
		}
	}

	// Add parameters documentation
	for _, param := range operation.Parameters {
		paramType := "string"
		if param.Schema != nil {
			if t, ok := param.Schema["type"].(string); ok {
				paramType = t
			}
		}

		required := ""
		if param.Required {
			required = "required"
		}

		sb.WriteString(fmt.Sprintf("# @param %s {%s} %s - %s\n",
			param.Name, paramType, required, param.Description))
	}

	// Add response documentation with recursive field extraction
	for code, response := range operation.Responses {
		sb.WriteString(fmt.Sprintf("# @response %s - %s\n", code, response.Description))

		// Extract response fields recursively
		fields := extractResponseFields(response, spec)
		for _, field := range fields {
			required := "optional"
			if field.Required {
				required = "required"
			}

			sb.WriteString(fmt.Sprintf("# @response-field %s {%s} %s",
				field.Name, field.Type, required))

			if field.Deprecated {
				sb.WriteString(" deprecated")
			}

			if field.Description != "" {
				sb.WriteString(fmt.Sprintf(" - %s", field.Description))
			}
			sb.WriteString("\n")

			if field.Example != nil {
				exampleStr := fmt.Sprintf("%v", field.Example)
				if strVal, ok := field.Example.(string); ok {
					exampleStr = fmt.Sprintf("\"%s\"", strVal)
				} else if jsonBytes, err := json.Marshal(field.Example); err == nil {
					exampleStr = string(jsonBytes)
				}
				sb.WriteString(fmt.Sprintf("# @response-example %s %s\n", field.Name, exampleStr))
			}
		}
	}

	sb.WriteString("\n")

	// Build URL with query parameters
	requestPath := path
	queryParams := []string{}

	for _, param := range operation.Parameters {
		switch param.In {
		case "path":
			placeholder := fmt.Sprintf("{{%s}}", param.Name)
			requestPath = strings.ReplaceAll(requestPath, "{"+param.Name+"}", placeholder)
		case "query":
			queryParams = append(queryParams, fmt.Sprintf("%s={{%s}}", param.Name, param.Name))
		}
	}

	fullURL := baseURL + requestPath
	if len(queryParams) > 0 {
		fullURL += "?" + strings.Join(queryParams, "&")
	}

	// Method and URL
	sb.WriteString(fmt.Sprintf("%s %s\n", method, fullURL))

	// Headers
	sb.WriteString("Content-Type: application/json\n")

	// Add header parameters
	for _, param := range operation.Parameters {
		if param.In == "header" {
			sb.WriteString(fmt.Sprintf("%s: {{%s}}\n", param.Name, param.Name))
		}
	}

	// Request body
	if operation.RequestBody != nil && method != "GET" && method != "DELETE" {
		sb.WriteString("\n")

		// Try to generate example from schema
		if content, ok := operation.RequestBody.Content["application/json"]; ok {
			if content.Example != nil {
				// Use provided example
				if exampleJSON, err := json.MarshalIndent(content.Example, "", "  "); err == nil {
					sb.WriteString(string(exampleJSON))
				}
			} else if content.Schema != nil {
				// Generate example from schema
				example := generateExampleFromSchema(content.Schema)
				if exampleJSON, err := json.MarshalIndent(example, "", "  "); err == nil {
					sb.WriteString(string(exampleJSON))
				}
			} else {
				sb.WriteString("{\n  \n}")
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// generateExampleFromSchema generates an example value from a JSON schema
func generateExampleFromSchema(schema map[string]interface{}) interface{} {
	schemaType, ok := schema["type"].(string)
	if !ok {
		return "{}"
	}

	switch schemaType {
	case "object":
		result := make(map[string]interface{})
		if properties, ok := schema["properties"].(map[string]interface{}); ok {
			for key, propSchema := range properties {
				if propMap, ok := propSchema.(map[string]interface{}); ok {
					result[key] = generateExampleFromSchema(propMap)
				}
			}
		}
		return result

	case "array":
		if items, ok := schema["items"].(map[string]interface{}); ok {
			return []interface{}{generateExampleFromSchema(items)}
		}
		return []interface{}{}

	case "string":
		if example, ok := schema["example"].(string); ok {
			return example
		}
		return "string"

	case "integer", "number":
		if example, ok := schema["example"].(float64); ok {
			return example
		}
		return 0

	case "boolean":
		if example, ok := schema["example"].(bool); ok {
			return example
		}
		return false

	default:
		return nil
	}
}

// ResponseField represents a field in a response schema
type ResponseField struct {
	Name        string
	Type        string
	Required    bool
	Description string
	Example     interface{}
	Deprecated  bool
}

// extractResponseFields extracts ALL fields recursively from response schema
func extractResponseFields(response OpenAPIResponse, spec *OpenAPISpec) []ResponseField {
	var fields []ResponseField

	// Get JSON content
	jsonContent, ok := response.Content["application/json"]
	if !ok || jsonContent.Schema == nil {
		return fields
	}

	// Resolve $ref if present
	schema := resolveSchema(jsonContent.Schema, spec)

	// Extract fields recursively
	extractFieldsRecursive(schema, "", &fields, spec, 0, 100)

	return fields
}

// extractFieldsRecursive recursively extracts fields from schema with dot notation
func extractFieldsRecursive(schema map[string]interface{}, prefix string, fields *[]ResponseField, spec *OpenAPISpec, depth int, maxDepth int) {
	// Prevent infinite recursion
	if depth >= maxDepth {
		return
	}

	// Get required fields array
	requiredFields := make(map[string]bool)
	if req, ok := schema["required"].([]interface{}); ok {
		for _, r := range req {
			if reqStr, ok := r.(string); ok {
				requiredFields[reqStr] = true
			}
		}
	}

	// Get properties
	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		return
	}

	for name, propSchemaRaw := range properties {
		propSchema, ok := propSchemaRaw.(map[string]interface{})
		if !ok {
			continue
		}

		fullName := name
		if prefix != "" {
			fullName = prefix + "." + name
		}

		isRequired := requiredFields[name]

		// Resolve $ref if present
		resolvedSchema := resolveSchema(propSchema, spec)

		// Get schema type
		schemaType := getSchemaType(resolvedSchema)

		// Check if this property has nested structure
		hasNestedObjectProps := (schemaType == "object" || hasProperties(resolvedSchema)) && hasProperties(resolvedSchema)
		hasNestedArrayProps := schemaType == "array" && hasArrayOfObjects(resolvedSchema)

		// Only add the field if it's a leaf node (no nested properties to expand)
		// This skips intermediate object/array containers
		if !hasNestedObjectProps && !hasNestedArrayProps {
			fieldType := getFieldType(propSchema)
			*fields = append(*fields, ResponseField{
				Name:        fullName,
				Type:        fieldType,
				Required:    isRequired,
				Description: getStringFromSchema(propSchema, "description"),
				Example:     propSchema["example"],
				Deprecated:  getBoolFromSchema(propSchema, "deprecated") || getBoolFromSchema(resolvedSchema, "deprecated"),
			})
		}

		// Recurse into nested objects
		if hasNestedObjectProps {
			extractFieldsRecursive(resolvedSchema, fullName, fields, spec, depth+1, maxDepth)
		}

		// Recurse into arrays of objects
		if hasNestedArrayProps {
			if items, ok := resolvedSchema["items"].(map[string]interface{}); ok {
				extractFieldsRecursive(items, fullName+"[]", fields, spec, depth+1, maxDepth)
			}
		}
	}
}

// resolveSchema resolves $ref in schema
func resolveSchema(schema map[string]interface{}, spec *OpenAPISpec) map[string]interface{} {
	// Handle direct $ref
	if ref, ok := schema["$ref"].(string); ok {
		// Extract component name from $ref (e.g., "#/components/schemas/User")
		parts := strings.Split(ref, "/")
		if len(parts) >= 4 && parts[0] == "#" && parts[1] == "components" && parts[2] == "schemas" {
			schemaName := strings.Join(parts[3:], "/")
			if spec.Components != nil && spec.Components.Schemas != nil {
				if resolvedSchema, ok := spec.Components.Schemas[schemaName]; ok {
					return resolvedSchema
				}
			}
		}
	}

	// Handle anyOf/oneOf with null (common pattern for nullable)
	if anyOf, ok := schema["anyOf"].([]interface{}); ok {
		for _, s := range anyOf {
			if sMap, ok := s.(map[string]interface{}); ok {
				if sType, ok := sMap["type"].(string); !ok || sType != "null" {
					return resolveSchema(sMap, spec)
				}
			}
		}
	}

	if oneOf, ok := schema["oneOf"].([]interface{}); ok {
		for _, s := range oneOf {
			if sMap, ok := s.(map[string]interface{}); ok {
				if sType, ok := sMap["type"].(string); !ok || sType != "null" {
					return resolveSchema(sMap, spec)
				}
			}
		}
	}

	return schema
}

// getSchemaType gets the type from a schema
func getSchemaType(schema map[string]interface{}) string {
	if typeVal, ok := schema["type"].(string); ok {
		return typeVal
	}
	// Handle array of types (for nullable)
	if typeArr, ok := schema["type"].([]interface{}); ok && len(typeArr) > 0 {
		if typeStr, ok := typeArr[0].(string); ok {
			return typeStr
		}
	}
	return "string"
}

// hasProperties checks if schema has properties
func hasProperties(schema map[string]interface{}) bool {
	_, ok := schema["properties"].(map[string]interface{})
	return ok
}

// hasArrayOfObjects checks if schema is an array of objects
func hasArrayOfObjects(schema map[string]interface{}) bool {
	schemaType := getSchemaType(schema)
	if schemaType != "array" {
		return false
	}

	items, ok := schema["items"].(map[string]interface{})
	if !ok {
		return false
	}

	itemType := getSchemaType(items)
	return itemType == "object" && hasProperties(items)
}

// getFieldType gets a human-readable type description
func getFieldType(schema map[string]interface{}) string {
	schemaType := getSchemaType(schema)

	if schemaType == "array" {
		if items, ok := schema["items"].(map[string]interface{}); ok {
			itemType := getSchemaType(items)
			return fmt.Sprintf("array<%s>", itemType)
		}
		return "array"
	}

	if schemaType == "object" {
		if addProps, ok := schema["additionalProperties"].(map[string]interface{}); ok {
			valueType := getSchemaType(addProps)
			return fmt.Sprintf("object<string, %s>", valueType)
		}
		return "object"
	}

	return schemaType
}

// getStringFromSchema safely gets a string value from schema
func getStringFromSchema(schema map[string]interface{}, key string) string {
	if val, ok := schema[key].(string); ok {
		return val
	}
	return ""
}

// getBoolFromSchema safely gets a bool value from schema
func getBoolFromSchema(schema map[string]interface{}, key string) bool {
	if val, ok := schema[key].(bool); ok {
		return val
	}
	return false
}
