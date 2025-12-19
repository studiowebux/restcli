package parser

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/studiowebux/restcli/internal/types"
)

// ParseHTTPFile parses a traditional .http file with ### separators
func ParseHTTPFile(filePath string) ([]types.HttpRequest, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var requests []types.HttpRequest
	var currentRequest *types.HttpRequest
	var bodyLines []string
	inBody := false

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		// New request separator
		if strings.HasPrefix(line, "###") {
			// Save previous request if exists
			if currentRequest != nil {
				if inBody && len(bodyLines) > 0 {
					currentRequest.Body = strings.Join(bodyLines, "\n")
				}
				requests = append(requests, *currentRequest)
			}

			// Start new request
			currentRequest = &types.HttpRequest{
				Name:               strings.TrimSpace(strings.TrimPrefix(line, "###")),
				Headers:            make(map[string]string),
				DocumentationLines: []string{}, // Store raw documentation lines for lazy parsing
			}
			bodyLines = []string{}
			inBody = false
			continue
		}

		// Documentation annotations (before anything else) - store for lazy parsing
		if strings.HasPrefix(line, "#") && currentRequest != nil {
			// Check for @filter and @query annotations first
			trimmed := strings.TrimSpace(strings.TrimPrefix(line, "#"))
			if strings.HasPrefix(trimmed, "@protocol ") {
				currentRequest.Protocol = strings.TrimSpace(strings.TrimPrefix(trimmed, "@protocol "))
				continue
			}
			if strings.HasPrefix(trimmed, "@filter ") {
				currentRequest.Filter = strings.TrimSpace(strings.TrimPrefix(trimmed, "@filter"))
				continue
			}
			if strings.HasPrefix(trimmed, "@query ") {
				currentRequest.Query = strings.TrimSpace(strings.TrimPrefix(trimmed, "@query"))
				continue
			}
			if strings.HasPrefix(trimmed, "@parsing ") {
				value := strings.TrimSpace(strings.TrimPrefix(trimmed, "@parsing"))
				currentRequest.ParseEscapes = value == "true"
				continue
			}
			if strings.HasPrefix(trimmed, "@streaming ") {
				value := strings.TrimSpace(strings.TrimPrefix(trimmed, "@streaming"))
				currentRequest.Streaming = value == "true"
				continue
			}
			if strings.HasPrefix(trimmed, "@confirmation ") {
				value := strings.TrimSpace(strings.TrimPrefix(trimmed, "@confirmation"))
				currentRequest.RequiresConfirmation = value == "true"
				continue
			}
			// Check for @tls.* annotations
			if strings.HasPrefix(trimmed, "@tls.") {
				if currentRequest.TLS == nil {
					currentRequest.TLS = &types.TLSConfig{}
				}
				if strings.HasPrefix(trimmed, "@tls.certFile ") {
					currentRequest.TLS.CertFile = strings.TrimSpace(strings.TrimPrefix(trimmed, "@tls.certFile"))
					continue
				}
				if strings.HasPrefix(trimmed, "@tls.keyFile ") {
					currentRequest.TLS.KeyFile = strings.TrimSpace(strings.TrimPrefix(trimmed, "@tls.keyFile"))
					continue
				}
				if strings.HasPrefix(trimmed, "@tls.caFile ") {
					currentRequest.TLS.CAFile = strings.TrimSpace(strings.TrimPrefix(trimmed, "@tls.caFile"))
					continue
				}
				if strings.HasPrefix(trimmed, "@tls.insecureSkipVerify ") {
					value := strings.TrimSpace(strings.TrimPrefix(trimmed, "@tls.insecureSkipVerify"))
					currentRequest.TLS.InsecureSkipVerify = value == "true"
					continue
				}
			}
			// Check for validation annotations
			if strings.HasPrefix(trimmed, "@expectedStatusCodes ") {
				value := strings.TrimSpace(strings.TrimPrefix(trimmed, "@expectedStatusCodes"))
				currentRequest.ExpectedStatusCodes = ParseStatusCodes(value)
				continue
			}
			if strings.HasPrefix(trimmed, "@expectedBodyExact ") {
				value := strings.TrimSpace(strings.TrimPrefix(trimmed, "@expectedBodyExact"))
				// Remove quotes if present
				if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
					value = value[1 : len(value)-1]
				}
				currentRequest.ExpectedBodyExact = value
				continue
			}
			if strings.HasPrefix(trimmed, "@expectedBody ") {
				value := strings.TrimSpace(strings.TrimPrefix(trimmed, "@expectedBody"))
				// Remove quotes if present
				if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
					value = value[1 : len(value)-1]
				}
				currentRequest.ExpectedBodyContains = value
				continue
			}
			if strings.HasPrefix(trimmed, "@expectedBodyPattern ") {
				value := strings.TrimSpace(strings.TrimPrefix(trimmed, "@expectedBodyPattern"))
				// Remove quotes if present
				if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
					value = value[1 : len(value)-1]
				}
				currentRequest.ExpectedBodyPattern = value
				continue
			}
			if strings.HasPrefix(trimmed, "@expectedBodyField ") {
				value := strings.TrimSpace(strings.TrimPrefix(trimmed, "@expectedBodyField"))
				// Parse field=value format
				if currentRequest.ExpectedBodyFields == nil {
					currentRequest.ExpectedBodyFields = make(map[string]string)
				}
				parts := strings.SplitN(value, "=", 2)
				if len(parts) == 2 {
					field := strings.TrimSpace(parts[0])
					fieldValue := strings.TrimSpace(parts[1])
					currentRequest.ExpectedBodyFields[field] = fieldValue
				}
				continue
			}
			currentRequest.DocumentationLines = append(currentRequest.DocumentationLines, line)
			continue
		}

		// HTTP method and URL (e.g., GET http://example.com)
		if currentRequest != nil && currentRequest.Method == "" {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				method := strings.ToUpper(parts[0])
				validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
				for _, vm := range validMethods {
					if method == vm {
						currentRequest.Method = method
						currentRequest.URL = parts[1]
						break
					}
				}
			}
			continue
		}

		// Empty line after headers starts body (check BEFORE skipping empty lines)
		if currentRequest != nil && currentRequest.Method != "" && strings.TrimSpace(line) == "" && !inBody {
			inBody = true
			continue
		}

		// Skip empty lines when not in body (AFTER checking for body transition)
		if strings.TrimSpace(line) == "" && !inBody {
			continue
		}

		// Headers (Key: Value) - only parse as header if not in body
		if currentRequest != nil && !inBody && currentRequest.Method != "" && strings.Contains(line, ":") {
			// Line must not be indented (body content is often indented)
			if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
				// This looks like body content, not a header
				// Treat as start of body
				inBody = true
				bodyLines = append(bodyLines, line)
				continue
			}

			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// HTTP header names must:
				// - Not be empty
				// - Not start with special chars like quotes, braces, brackets
				// - Not contain spaces, tabs, or other invalid characters
				// - Be valid header name characters (alphanumeric, hyphens, underscores)
				if key == "" || strings.ContainsAny(key, " \t{[\"'") {
					// Invalid header, treat as body start
					inBody = true
					bodyLines = append(bodyLines, line)
					continue
				}

				// Remove surrounding quotes from header value if present
				// e.g., Content-Type: "application/json" -> Content-Type: application/json
				if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
					value = value[1 : len(value)-1]
				}

				// Valid header
				currentRequest.Headers[key] = value
				continue
			}
		}

		// Body content
		if currentRequest != nil && inBody {
			bodyLines = append(bodyLines, line)
		}
	}

	// Save last request
	if currentRequest != nil {
		if inBody && len(bodyLines) > 0 {
			currentRequest.Body = strings.Join(bodyLines, "\n")
		}
		requests = append(requests, *currentRequest)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return requests, nil
}

// ParseDocumentationLines parses documentation from a slice of comment lines
// This is used for lazy loading documentation
func ParseDocumentationLines(lines []string) *types.Documentation {
	if len(lines) == 0 {
		return nil
	}

	doc := &types.Documentation{}
	for _, line := range lines {
		parseDocumentation(line, doc)
	}

	// Only return documentation if it has content
	if doc.Description != "" || len(doc.Parameters) > 0 || len(doc.Responses) > 0 || len(doc.Tags) > 0 {
		return doc
	}
	return nil
}

// parseDocumentation parses documentation annotations from comments
func parseDocumentation(line string, doc *types.Documentation) {
	line = strings.TrimSpace(strings.TrimPrefix(line, "#"))

	// @description
	if strings.HasPrefix(line, "@description") {
		doc.Description = strings.TrimSpace(strings.TrimPrefix(line, "@description"))
		return
	}

	// @category
	if strings.HasPrefix(line, "@category") {
		category := strings.TrimSpace(strings.TrimPrefix(line, "@category"))
		doc.Tags = append(doc.Tags, category)
		return
	}

	// @param name {type} required - description
	if strings.HasPrefix(line, "@param") {
		param := parseParameter(strings.TrimPrefix(line, "@param"))
		if param != nil {
			doc.Parameters = append(doc.Parameters, *param)
		}
		return
	}

	// @response code - description
	if strings.HasPrefix(line, "@response ") {
		response := parseResponse(strings.TrimPrefix(line, "@response"))
		if response != nil {
			doc.Responses = append(doc.Responses, *response)
		}
		return
	}

	// @response-field name {type} required - description
	if strings.HasPrefix(line, "@response-field") {
		// Find the last response and add the field to it
		if len(doc.Responses) > 0 {
			field := parseResponseField(strings.TrimPrefix(line, "@response-field"))
			if field != nil {
				lastIdx := len(doc.Responses) - 1
				doc.Responses[lastIdx].Fields = append(doc.Responses[lastIdx].Fields, *field)
			}
		}
		return
	}

	// @example paramName value
	if strings.HasPrefix(line, "@example") {
		// Format: @example paramName value
		exampleLine := strings.TrimPrefix(line, "@example")
		parts := strings.Fields(exampleLine)
		if len(parts) >= 2 {
			paramName := parts[0]
			exampleValue := strings.Join(parts[1:], " ")

			// Find the parameter and add the example
			for i := range doc.Parameters {
				if doc.Parameters[i].Name == paramName {
					doc.Parameters[i].Example = exampleValue
					break
				}
			}
		}
		return
	}

	// @response-example fieldName "value"
	if strings.HasPrefix(line, "@response-example") {
		// This is typically used to provide example values, we can skip for now
		return
	}

	// If no @ prefix, append to description
	if !strings.HasPrefix(line, "@") && doc.Description == "" {
		doc.Description = strings.TrimSpace(line)
	}
}

// parseParameter parses a parameter annotation
// Format: name {type} required - description
func parseParameter(line string) *types.Parameter {
	line = strings.TrimSpace(line)
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil
	}

	param := &types.Parameter{
		Name: parts[0],
	}

	// Find type in {braces}
	if typeStart := strings.Index(line, "{"); typeStart != -1 {
		if typeEnd := strings.Index(line[typeStart:], "}"); typeEnd != -1 {
			param.Type = line[typeStart+1 : typeStart+typeEnd]
		}
	}

	// Check for required keyword
	param.Required = strings.Contains(line, "required")

	// Extract description after "-"
	if dashIdx := strings.Index(line, "-"); dashIdx != -1 {
		param.Description = strings.TrimSpace(line[dashIdx+1:])
	}

	return param
}

// parseResponse parses a response annotation
// Format: code - description
func parseResponse(line string) *types.Response {
	line = strings.TrimSpace(line)
	parts := strings.SplitN(line, "-", 2)
	if len(parts) < 2 {
		return nil
	}

	return &types.Response{
		Code:        strings.TrimSpace(parts[0]),
		Description: strings.TrimSpace(parts[1]),
	}
}

// parseResponseField parses a response field annotation
// Format: name {type} required - description
func parseResponseField(line string) *types.ResponseField {
	line = strings.TrimSpace(line)
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil
	}

	field := &types.ResponseField{
		Name: parts[0],
	}

	// Find type in {braces}
	if typeStart := strings.Index(line, "{"); typeStart != -1 {
		if typeEnd := strings.Index(line[typeStart:], "}"); typeEnd != -1 {
			field.Type = line[typeStart+1 : typeStart+typeEnd]
		}
	}

	// Check for required keyword
	field.Required = strings.Contains(line, "required")

	// Check for deprecated keyword
	field.Deprecated = strings.Contains(line, "deprecated")

	// Extract description after "-"
	if dashIdx := strings.Index(line, "-"); dashIdx != -1 {
		field.Description = strings.TrimSpace(line[dashIdx+1:])
	}

	return field
}

// ParseStatusCodes parses a comma-separated list of status codes and ranges
// Supports: "200,201,204" or "2xx,3xx" or mixed "200,2xx,404"
func ParseStatusCodes(input string) []int {
	var codes []int
	parts := strings.Split(input, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Check for range notation (2xx, 3xx, 4xx, 5xx)
		if strings.HasSuffix(part, "xx") && len(part) == 3 {
			rangeStart := int(part[0] - '0')
			if rangeStart >= 1 && rangeStart <= 5 {
				// Add all codes in range (e.g., 2xx = 200-299)
				for i := rangeStart * 100; i < (rangeStart+1)*100; i++ {
					codes = append(codes, i)
				}
			}
		} else {
			// Parse as specific code
			var code int
			if _, err := fmt.Sscanf(part, "%d", &code); err == nil {
				codes = append(codes, code)
			}
		}
	}

	return codes
}
