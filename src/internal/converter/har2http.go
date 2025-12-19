package converter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/studiowebux/restcli/internal/config"
	"github.com/studiowebux/restcli/internal/types"
	"gopkg.in/yaml.v3"
)

// Har2HttpOptions contains options for har2http conversion
type Har2HttpOptions struct {
	HarFile       string
	OutputDir     string
	ImportHeaders bool   // If true, include sensitive headers
	Format        string // http, json, yaml (default: http)
	Filter        string // Filter by URL pattern (optional)
}

// HARFile represents the HAR file structure
type HARFile struct {
	Log HARLog `json:"log"`
}

// HARLog represents the log section of HAR
type HARLog struct {
	Version string      `json:"version"`
	Creator HARCreator  `json:"creator"`
	Entries []HAREntry  `json:"entries"`
}

// HARCreator represents the tool that created the HAR
type HARCreator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// HAREntry represents a single HTTP request/response
type HAREntry struct {
	Request  HARRequest  `json:"request"`
	Response HARResponse `json:"response"`
}

// HARRequest represents the request part of an entry
type HARRequest struct {
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	HTTPVersion string            `json:"httpVersion"`
	Headers     []HARHeader       `json:"headers"`
	QueryString []HARQueryParam   `json:"queryString"`
	PostData    *HARPostData      `json:"postData,omitempty"`
}

// HARResponse represents the response part of an entry
type HARResponse struct {
	Status      int         `json:"status"`
	StatusText  string      `json:"statusText"`
	HTTPVersion string      `json:"httpVersion"`
	Headers     []HARHeader `json:"headers"`
	Content     HARContent  `json:"content"`
}

// HARHeader represents a single header
type HARHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// HARQueryParam represents a query parameter
type HARQueryParam struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// HARPostData represents POST data
type HARPostData struct {
	MimeType string      `json:"mimeType"`
	Text     string      `json:"text"`
	Params   []HARParam  `json:"params,omitempty"`
}

// HARParam represents a form parameter
type HARParam struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// HARContent represents response content
type HARContent struct {
	Size     int    `json:"size"`
	MimeType string `json:"mimeType"`
	Text     string `json:"text"`
}

// Har2Http converts a HAR file to .http files
func Har2Http(opts Har2HttpOptions) error {
	// Read HAR file
	data, err := os.ReadFile(opts.HarFile)
	if err != nil {
		return fmt.Errorf("failed to read HAR file: %w", err)
	}

	// Parse HAR
	var har HARFile
	if err := json.Unmarshal(data, &har); err != nil {
		return fmt.Errorf("failed to parse HAR file: %w", err)
	}

	if len(har.Log.Entries) == 0 {
		return fmt.Errorf("no entries found in HAR file")
	}

	// Create output directory
	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = "requests"
	}
	if err := os.MkdirAll(outputDir, config.DirPermissions); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Determine format
	format := opts.Format
	if format == "" {
		format = "http"
	}

	// Convert each entry
	converted := 0
	for i, entry := range har.Log.Entries {
		// Filter by URL pattern if specified
		if opts.Filter != "" && !strings.Contains(entry.Request.URL, opts.Filter) {
			continue
		}

		// Skip non-HTTP(S) requests
		if !strings.HasPrefix(entry.Request.URL, "http://") && !strings.HasPrefix(entry.Request.URL, "https://") {
			continue
		}

		// Convert entry
		if err := convertEntry(entry, i, outputDir, format, opts.ImportHeaders); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to convert entry %d: %v\n", i, err)
			continue
		}
		converted++
	}

	fmt.Fprintf(os.Stderr, "Converted %d/%d entries to %s/\n", converted, len(har.Log.Entries), outputDir)
	return nil
}

// convertEntry converts a single HAR entry to a request file
func convertEntry(entry HAREntry, index int, outputDir, format string, importHeaders bool) error {
	req := entry.Request

	// Build headers map
	headers := make(map[string]string)
	for _, h := range req.Headers {
		// Skip pseudo-headers
		if strings.HasPrefix(h.Name, ":") {
			continue
		}
		headers[h.Name] = h.Value
	}

	// Filter sensitive headers unless explicitly importing
	if !importHeaders {
		sensitiveHeaders := []string{"Cookie", "Authorization", "X-Auth-Token", "X-API-Key"}
		for _, name := range sensitiveHeaders {
			delete(headers, name)
		}
	}

	// Get body
	body := ""
	if req.PostData != nil {
		body = req.PostData.Text
	}

	// Detect variables
	variables := make(map[string]string)
	if authHeader, ok := headers["Authorization"]; ok {
		if strings.HasPrefix(authHeader, "Bearer ") {
			variables["token"] = strings.TrimPrefix(authHeader, "Bearer ")
			headers["Authorization"] = "Bearer {{token}}"
		}
	}

	// Generate filename
	filename := suggestFilenameFromURL(req.URL, req.Method, index)

	var content string
	var ext string

	switch format {
	case "json":
		httpReq := harToHttpRequest(req, headers, body, variables)
		data, err := json.MarshalIndent(httpReq, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		content = string(data)
		ext = ".json"
	case "yaml":
		httpReq := harToHttpRequest(req, headers, body, variables)
		data, err := yaml.Marshal(httpReq)
		if err != nil {
			return fmt.Errorf("failed to marshal YAML: %w", err)
		}
		content = string(data)
		ext = ".yaml"
	default: // "http"
		content = generateHttpFileFromHAR(req, headers, body, variables)
		ext = ".http"
	}

	// Replace extension based on format
	if format != "http" {
		filename = strings.TrimSuffix(filename, ".http") + ext
	}

	outputPath := filepath.Join(outputDir, filename)
	if err := os.WriteFile(outputPath, []byte(content), config.FilePermissions); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// harToHttpRequest converts HAR request to HttpRequest type
func harToHttpRequest(req HARRequest, headers map[string]string, body string, variables map[string]string) types.HttpRequest {
	httpReq := types.HttpRequest{
		Method:  req.Method,
		URL:     req.URL,
		Headers: headers,
		Body:    body,
	}

	// Add documentation for variables
	if len(variables) > 0 {
		var docLines []string
		docLines = append(docLines, fmt.Sprintf("### %s %s", req.Method, extractPath(req.URL)))
		docLines = append(docLines, "")
		docLines = append(docLines, "Variables:")
		for k, v := range variables {
			docLines = append(docLines, fmt.Sprintf("  %s: %s", k, v))
		}
		httpReq.DocumentationLines = docLines
	}

	return httpReq
}

// generateHttpFileFromHAR generates .http format content from HAR request
func generateHttpFileFromHAR(req HARRequest, headers map[string]string, body string, variables map[string]string) string {
	var sb strings.Builder

	// Add title comment
	path := extractPath(req.URL)
	sb.WriteString(fmt.Sprintf("### %s %s\n", req.Method, path))

	// Add variables as comments if any
	if len(variables) > 0 {
		sb.WriteString("# Variables:\n")
		for k, v := range variables {
			sb.WriteString(fmt.Sprintf("#   %s: %s\n", k, v))
		}
	}

	// Request line
	sb.WriteString(fmt.Sprintf("%s %s\n", req.Method, req.URL))

	// Headers
	for name, value := range headers {
		sb.WriteString(fmt.Sprintf("%s: %s\n", name, value))
	}

	// Body
	if body != "" {
		sb.WriteString("\n")
		sb.WriteString(body)
		sb.WriteString("\n")
	}

	return sb.String()
}

// suggestFilenameFromURL generates a filename from URL and method
func suggestFilenameFromURL(urlStr, method string, index int) string {
	path := extractPath(urlStr)

	// Clean path for filename
	filename := strings.ReplaceAll(path, "/", "-")
	filename = strings.Trim(filename, "-")
	filename = strings.ToLower(filename)

	// Remove invalid characters
	filename = regexp.MustCompile(`[^a-z0-9-_]`).ReplaceAllString(filename, "-")

	// Prepend method
	filename = strings.ToLower(method) + "-" + filename

	// Handle empty or root path
	if filename == strings.ToLower(method)+"-" || filename == "" {
		filename = fmt.Sprintf("%s-request-%d", strings.ToLower(method), index)
	}

	// Add extension
	filename += ".http"

	return filename
}

// extractPath extracts the path from a URL
func extractPath(urlStr string) string {
	// Find start of path (after host)
	parts := strings.SplitN(urlStr, "://", 2)
	if len(parts) != 2 {
		return "/"
	}

	pathStart := strings.Index(parts[1], "/")
	if pathStart == -1 {
		return "/"
	}

	path := parts[1][pathStart:]

	// Remove query string
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}

	// Remove fragment
	if idx := strings.Index(path, "#"); idx != -1 {
		path = path[:idx]
	}

	return path
}
