package converter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// CurlToHttpOptions contains options for curl2http conversion
type CurlToHttpOptions struct {
	CurlCommand    string
	OutputFile     string
	ImportHeaders  bool // If true, include sensitive headers
}

// CurlRequest represents a parsed cURL command
type CurlRequest struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    string
}

// Curl2Http converts a cURL command to a .http file
func Curl2Http(opts CurlToHttpOptions) error {
	// Parse cURL command
	req, err := parseCurl(opts.CurlCommand)
	if err != nil {
		return fmt.Errorf("failed to parse cURL command: %w", err)
	}

	// Filter sensitive headers unless explicitly importing them
	if !opts.ImportHeaders {
		filterSensitiveHeaders(req)
	}

	// Detect variables
	variables := detectVariables(req)

	// Generate .http content
	httpContent := generateHttpFile(req, variables)

	// Determine output filename
	outputFile := opts.OutputFile
	if outputFile == "" {
		outputFile = suggestFilename(req.URL)
	}

	// Write to file or stdout
	if outputFile == "-" {
		fmt.Print(httpContent)
	} else {
		if err := os.WriteFile(outputFile, []byte(httpContent), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Created %s\n", outputFile)
	}

	return nil
}

// parseCurl parses a cURL command string
func parseCurl(curlCmd string) (*CurlRequest, error) {
	req := &CurlRequest{
		Method:  "GET",
		Headers: make(map[string]string),
	}

	// Clean up the command - handle multiline with backslashes
	curlCmd = strings.ReplaceAll(curlCmd, "\\\n", " ")
	curlCmd = strings.ReplaceAll(curlCmd, "\n", " ")
	curlCmd = strings.TrimSpace(curlCmd)

	// Remove 'curl' at the start if present
	curlCmd = regexp.MustCompile(`^curl\s+`).ReplaceAllString(curlCmd, "")

	// Extract URL - try to find after --url or as standalone URL
	var urlFound bool

	// Try --url flag first
	urlFlagRe := regexp.MustCompile(`--url\s+['"]?([^'" ]+)['"]?`)
	if matches := urlFlagRe.FindStringSubmatch(curlCmd); len(matches) > 1 {
		req.URL = strings.Trim(matches[1], `'"`)
		urlFound = true
	}

	// If not found, look for standalone URL (http/https)
	if !urlFound {
		urlRe := regexp.MustCompile(`(https?://[^\s'"\\]+)`)
		if matches := urlRe.FindStringSubmatch(curlCmd); len(matches) > 1 {
			req.URL = strings.Trim(matches[1], `'"`)
			urlFound = true
		}
	}

	if !urlFound || req.URL == "" {
		return nil, fmt.Errorf("could not find URL in cURL command")
	}

	// Extract method - try multiple patterns
	methodPatterns := []string{
		`--request\s+([A-Z]+)`,
		`-X\s+([A-Z]+)`,
	}

	for _, pattern := range methodPatterns {
		methodRe := regexp.MustCompile(pattern)
		if matches := methodRe.FindStringSubmatch(curlCmd); len(matches) > 1 {
			req.Method = strings.ToUpper(matches[1])
			break
		}
	}

	// If no method specified but has data, assume POST
	if req.Method == "GET" && (strings.Contains(curlCmd, "--data") || strings.Contains(curlCmd, "-d")) {
		req.Method = "POST"
	}

	// Extract headers
	headerPatterns := []string{
		`-H\s+['"]([^'"]+)['"]`,
		`--header\s+['"]([^'"]+)['"]`,
	}

	for _, pattern := range headerPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(curlCmd, -1)
		for _, match := range matches {
			if len(match) > 1 {
				headerLine := match[1]
				parts := strings.SplitN(headerLine, ":", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					req.Headers[key] = value
				}
			}
		}
	}

	// Extract body
	bodyPatterns := []string{
		`--data\s+['"]([^'"]+)['"]`,
		`--data-raw\s+['"]([^'"]+)['"]`,
		`--data-binary\s+['"]([^'"]+)['"]`,
		`-d\s+['"]([^'"]+)['"]`,
	}

	for _, pattern := range bodyPatterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(curlCmd); len(matches) > 1 {
			req.Body = matches[1]
			// Unescape common escape sequences
			req.Body = strings.ReplaceAll(req.Body, `\"`, `"`)
			req.Body = strings.ReplaceAll(req.Body, `\n`, "\n")
			req.Body = strings.ReplaceAll(req.Body, `\t`, "\t")
			break
		}
	}

	return req, nil
}

// filterSensitiveHeaders removes or masks sensitive headers
func filterSensitiveHeaders(req *CurlRequest) {
	sensitiveHeaders := []string{
		"authorization",
		"cookie",
		"x-api-key",
		"api-key",
		"apikey",
		"x-auth-token",
		"auth-token",
	}

	for _, sensitive := range sensitiveHeaders {
		for key := range req.Headers {
			if strings.ToLower(key) == sensitive {
				// Replace with variable placeholder
				req.Headers[key] = "{{" + key + "}}"
			}
		}
	}
}

// detectVariables suggests variables for common patterns
func detectVariables(req *CurlRequest) map[string]string {
	variables := make(map[string]string)

	// Parse URL to extract base URL
	if parsedURL, err := url.Parse(req.URL); err == nil {
		baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
		variables["baseUrl"] = baseURL

		// Check for common ID patterns in path
		pathParts := strings.Split(parsedURL.Path, "/")
		for i, part := range pathParts {
			// Numeric IDs
			if matched, _ := regexp.MatchString(`^\d+$`, part); matched {
				if i > 0 {
					varName := pathParts[i-1] + "Id"
					variables[varName] = part
				}
			}
			// UUID patterns
			if matched, _ := regexp.MatchString(`^[a-f0-9-]{36}$`, part); matched {
				if i > 0 {
					varName := pathParts[i-1] + "Id"
					variables[varName] = part
				}
			}
		}
	}

	// Check headers for tokens
	for key, value := range req.Headers {
		keyLower := strings.ToLower(key)
		if strings.Contains(keyLower, "authorization") && !strings.Contains(value, "{{") {
			variables["token"] = strings.TrimPrefix(value, "Bearer ")
		}
	}

	return variables
}

// generateHttpFile generates the .http file content
func generateHttpFile(req *CurlRequest, variables map[string]string) string {
	var sb strings.Builder

	// Comment with suggestions
	if len(variables) > 0 {
		sb.WriteString("# Suggested variables:\n")
		for key, value := range variables {
			sb.WriteString(fmt.Sprintf("# @var %s = %s\n", key, value))
		}
		sb.WriteString("\n")
	}

	// Request name
	sb.WriteString("### Request\n")

	// Replace detected variables in URL
	requestURL := req.URL
	for key, value := range variables {
		if key == "baseUrl" {
			requestURL = strings.Replace(requestURL, value, "{{"+key+"}}", 1)
		} else {
			requestURL = strings.Replace(requestURL, value, "{{"+key+"}}", -1)
		}
	}

	// Method and URL
	sb.WriteString(fmt.Sprintf("%s %s\n", req.Method, requestURL))

	// Headers
	for key, value := range req.Headers {
		sb.WriteString(fmt.Sprintf("%s: %s\n", key, value))
	}

	// Body
	if req.Body != "" {
		sb.WriteString("\n")

		// Try to pretty-print JSON
		var jsonData interface{}
		if err := json.Unmarshal([]byte(req.Body), &jsonData); err == nil {
			if prettyJSON, err := json.MarshalIndent(jsonData, "", "  "); err == nil {
				sb.WriteString(string(prettyJSON))
			} else {
				sb.WriteString(req.Body)
			}
		} else {
			sb.WriteString(req.Body)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// suggestFilename suggests a filename based on the URL path
func suggestFilename(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "request.http"
	}

	// Get last part of path
	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathParts) > 0 && pathParts[len(pathParts)-1] != "" {
		filename := pathParts[len(pathParts)-1]
		// Remove numeric IDs
		if matched, _ := regexp.MatchString(`^\d+$`, filename); !matched {
			return filename + ".http"
		}
		// If last part is ID, use second-to-last
		if len(pathParts) > 1 {
			return pathParts[len(pathParts)-2] + ".http"
		}
	}

	// Fallback to host
	if parsedURL.Host != "" {
		hostname := strings.Split(parsedURL.Host, ":")[0]
		hostname = strings.ReplaceAll(hostname, ".", "_")
		return hostname + ".http"
	}

	return "request.http"
}

// ReadCurlFromStdin reads cURL command from stdin
func ReadCurlFromStdin() (string, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ValidateOutput File validates and creates directory for output file
func ValidateOutputFile(path string) error {
	if path == "" || path == "-" {
		return nil
	}

	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	return nil
}
