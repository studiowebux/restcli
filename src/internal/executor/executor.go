package executor

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/studiowebux/restcli/internal/types"
)

const (
	// MaxResponseSize limits response body size to prevent OOM (100MB)
	MaxResponseSize = 100 * 1024 * 1024
)

// Execute performs an HTTP request and returns the result
func Execute(req *types.HttpRequest, tlsConfig *types.TLSConfig, profile *types.Profile) (*types.RequestResult, error) {
	startTime := time.Now()

	// Get timeout from profile or use default
	timeout := 30 // Default timeout in seconds
	if profile != nil {
		timeout = profile.GetRequestTimeout()
	}

	// Handle GraphQL protocol
	if req.Protocol == "graphql" {
		return executeGraphQL(req, tlsConfig, startTime, timeout)
	}

	// Create HTTP request
	var bodyReader io.Reader
	requestSize := 0
	if req.Body != "" {
		bodyReader = bytes.NewBufferString(req.Body)
		requestSize = len(req.Body)
	}

	httpReq, err := http.NewRequest(req.Method, req.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Build HTTP client with optional TLS configuration
	client, err := buildHTTPClient(tlsConfig, time.Duration(timeout)*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to configure HTTP client: %w", err)
	}

	resp, err := client.Do(httpReq)
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		return &types.RequestResult{
			Error:       err.Error(),
			Duration:    duration,
			RequestSize: requestSize,
		}, nil
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return &types.RequestResult{
			Status:      resp.StatusCode,
			StatusText:  resp.Status,
			Error:       fmt.Sprintf("failed to read response body: %v", err),
			Duration:    duration,
			RequestSize: requestSize,
		}, nil
	}

	// Build response headers map
	headers := make(map[string]string)
	for key, values := range resp.Header {
		headers[key] = strings.Join(values, ", ")
	}

	// Note: Escape sequence parsing is now done AFTER filter/query in actions.go and cli.go
	// to ensure parsing happens as the final step

	result := &types.RequestResult{
		Status:       resp.StatusCode,
		StatusText:   resp.Status,
		Headers:      headers,
		Body:         string(bodyBytes),
		Duration:     duration,
		RequestSize:  requestSize,
		ResponseSize: len(bodyBytes),
	}

	return result, nil
}

// ExecuteWithStreaming performs an HTTP request with streaming support
// Auto-detects streaming based on Content-Type and Transfer-Encoding headers
// Calls streamCallback for each chunk received
func ExecuteWithStreaming(ctx context.Context, req *types.HttpRequest, tlsConfig *types.TLSConfig, profile *types.Profile, streamCallback types.StreamCallback) (*types.RequestResult, error) {
	startTime := time.Now()

	// Get max response size from profile or use default
	maxSize := int64(MaxResponseSize) // Default 100MB
	if profile != nil {
		maxSize = profile.GetMaxResponseSize()
	}

	// Get timeout from profile or use default (not used for streaming, but kept for consistency)
	timeout := 30
	if profile != nil {
		timeout = profile.GetRequestTimeout()
	}

	// Handle GraphQL protocol
	if req.Protocol == "graphql" {
		return executeGraphQL(req, tlsConfig, startTime, timeout)
	}

	// Create HTTP request
	var bodyReader io.Reader
	requestSize := 0
	if req.Body != "" {
		bodyReader = bytes.NewBufferString(req.Body)
		requestSize = len(req.Body)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Build HTTP client with optional TLS configuration
	// Use no timeout for streaming requests (timeout is managed by context)
	client, err := buildHTTPClient(tlsConfig, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to configure HTTP client: %w", err)
	}

	resp, err := client.Do(httpReq)
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		return &types.RequestResult{
			Error:       err.Error(),
			Duration:    duration,
			RequestSize: requestSize,
		}, nil
	}
	defer resp.Body.Close()

	// Build response headers map
	headers := make(map[string]string)
	for key, values := range resp.Header {
		headers[key] = strings.Join(values, ", ")
	}

	// Detect if response is streaming
	contentType := resp.Header.Get("Content-Type")
	transferEncoding := resp.Header.Get("Transfer-Encoding")
	isStreaming := strings.Contains(contentType, "text/event-stream") ||
		strings.Contains(contentType, "application/stream+json") ||
		strings.Contains(contentType, "application/x-ndjson") ||
		strings.Contains(contentType, "application/jsonlines") ||
		strings.Contains(transferEncoding, "chunked")

	var bodyBytes []byte
	var readErr error

	if isStreaming {
		// Stream the response (works with or without callback)
		bodyBytes, readErr = streamResponse(ctx, resp.Body, maxSize, streamCallback)
	} else {
		// Non-streaming: read all at once
		bodyBytes, readErr = io.ReadAll(resp.Body)
	}

	if readErr != nil {
		// Check if cancelled
		if ctx.Err() == context.Canceled {
			return &types.RequestResult{
				Status:      resp.StatusCode,
				StatusText:  resp.Status,
				Headers:     headers,
				Body:        string(bodyBytes), // Partial body
				Error:       "Request cancelled",
				Duration:    time.Since(startTime).Milliseconds(),
				RequestSize: requestSize,
				ResponseSize: len(bodyBytes),
			}, nil
		}
		return &types.RequestResult{
			Status:      resp.StatusCode,
			StatusText:  resp.Status,
			Headers:     headers,
			Error:       fmt.Sprintf("failed to read response body: %v", readErr),
			Duration:    time.Since(startTime).Milliseconds(),
			RequestSize: requestSize,
		}, nil
	}

	result := &types.RequestResult{
		Status:       resp.StatusCode,
		StatusText:   resp.Status,
		Headers:      headers,
		Body:         string(bodyBytes),
		Duration:     time.Since(startTime).Milliseconds(),
		RequestSize:  requestSize,
		ResponseSize: len(bodyBytes),
	}

	return result, nil
}

// streamResponse reads the response body in chunks and calls the callback for each chunk
// callback can be nil, in which case chunks are just accumulated
// maxSize limits the total response size to prevent OOM
func streamResponse(ctx context.Context, body io.Reader, maxSize int64, callback types.StreamCallback) ([]byte, error) {
	var fullBody bytes.Buffer
	reader := bufio.NewReader(body)
	buffer := make([]byte, 4096) // 4KB chunks

	for {
		// Check for cancellation
		select {
		case <-ctx.Done():
			if callback != nil {
				callback(nil, true) // Signal done with cancellation
			}
			return fullBody.Bytes(), context.Canceled
		default:
		}

		n, err := reader.Read(buffer)
		if n > 0 {
			chunk := buffer[:n]

			// Check if adding this chunk would exceed max size
			if int64(fullBody.Len())+int64(n) > maxSize {
				return fullBody.Bytes(), fmt.Errorf("response size exceeds maximum allowed size (%d bytes)", maxSize)
			}

			fullBody.Write(chunk)
			if callback != nil {
				callback(chunk, false)
			}
		}

		if err == io.EOF {
			if callback != nil {
				callback(nil, true) // Signal done
			}
			break
		}
		if err != nil {
			return fullBody.Bytes(), err
		}
	}

	return fullBody.Bytes(), nil
}

// buildHTTPClient creates an HTTP client with optional TLS/mTLS configuration
// timeout parameter: 0 = no timeout, > 0 = specific timeout
func buildHTTPClient(tlsConfig *types.TLSConfig, timeout time.Duration) (*http.Client, error) {
	transport := &http.Transport{}

	if tlsConfig != nil {
		tlsCfg := &tls.Config{
			InsecureSkipVerify: tlsConfig.InsecureSkipVerify,
		}

		// Load client certificate if provided (for mTLS)
		if tlsConfig.CertFile != "" && tlsConfig.KeyFile != "" {
			cert, err := tls.LoadX509KeyPair(tlsConfig.CertFile, tlsConfig.KeyFile)
			if err != nil {
				return nil, fmt.Errorf("failed to load client certificate: %w", err)
			}
			tlsCfg.Certificates = []tls.Certificate{cert}
		}

		// Load CA certificate if provided (for server verification)
		if tlsConfig.CAFile != "" {
			caCert, err := os.ReadFile(tlsConfig.CAFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read CA certificate: %w", err)
			}
			caCertPool := x509.NewCertPool()
			if !caCertPool.AppendCertsFromPEM(caCert) {
				return nil, fmt.Errorf("failed to parse CA certificate")
			}
			tlsCfg.RootCAs = caCertPool
		}

		transport.TLSClientConfig = tlsCfg
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}, nil
}

// FormatDuration formats duration in milliseconds to human-readable string
func FormatDuration(ms int64) string {
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	seconds := float64(ms) / 1000.0
	return fmt.Sprintf("%.2fs", seconds)
}

// FormatSize formats byte size to human-readable string
func FormatSize(bytes int) string {
	if bytes < 1024 {
		return fmt.Sprintf("%dB", bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%.2fKB", float64(bytes)/1024.0)
	}
	return fmt.Sprintf("%.2fMB", float64(bytes)/(1024.0*1024.0))
}

// IsSuccessStatus returns true if status code is 2xx
func IsSuccessStatus(status int) bool {
	return status >= 200 && status < 300
}

// ParseEscapeSequences parses common escape sequences in a string AND removes outer JSON quotes
// This is a best-effort parser that handles: \n, \t, \r, \", \\, etc.
// Should be called AFTER filter/query operations to ensure it's the final processing step
// Only applies when # @parsing true is set
func ParseEscapeSequences(s string) string {
	result := s

	// First, try to unquote if it's a JSON-encoded string (removes outer quotes)
	// This handles cases like "Hello\nWorld" -> Hello\nWorld
	if len(result) >= 2 && result[0] == '"' && result[len(result)-1] == '"' {
		// Try to unmarshal as a JSON string
		var unquoted string
		if err := json.Unmarshal([]byte(result), &unquoted); err == nil {
			result = unquoted
			// After unquoting, the escape sequences are already parsed by JSON unmarshaling
			return result
		}
	}

	// If not a quoted JSON string, manually parse escape sequences
	// Important: Process \\ FIRST to handle escaped backslashes correctly
	// Otherwise \\n would become \<newline> instead of \n

	// Replace escaped backslash with a placeholder first
	const placeholder = "\x00BACKSLASH\x00"
	result = strings.ReplaceAll(result, "\\\\", placeholder)

	// Now replace other escape sequences
	result = strings.ReplaceAll(result, "\\n", "\n")
	result = strings.ReplaceAll(result, "\\t", "\t")
	result = strings.ReplaceAll(result, "\\r", "\r")
	result = strings.ReplaceAll(result, "\\\"", "\"")
	result = strings.ReplaceAll(result, "\\'", "'")
	result = strings.ReplaceAll(result, "\\b", "\b")
	result = strings.ReplaceAll(result, "\\f", "\f")

	// Finally replace the placeholder with actual backslash
	result = strings.ReplaceAll(result, placeholder, "\\")

	return result
}

// executeGraphQL handles GraphQL protocol requests
func executeGraphQL(req *types.HttpRequest, tlsConfig *types.TLSConfig, startTime time.Time, timeout int) (*types.RequestResult, error) {
	// Build GraphQL request payload
	graphqlPayload := map[string]interface{}{
		"query": req.Body,
	}

	// Marshal to JSON
	payloadBytes, err := json.Marshal(graphqlPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal GraphQL query: %w", err)
	}

	requestSize := len(payloadBytes)

	// Create HTTP POST request (GraphQL is always POST)
	httpReq, err := http.NewRequest("POST", req.URL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set Content-Type for GraphQL
	httpReq.Header.Set("Content-Type", "application/json")

	// Set other headers from the request
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Build HTTP client with TLS configuration
	client, err := buildHTTPClient(tlsConfig, time.Duration(timeout)*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to configure HTTP client: %w", err)
	}

	resp, err := client.Do(httpReq)
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		return &types.RequestResult{
			Error:       err.Error(),
			Duration:    duration,
			RequestSize: requestSize,
		}, nil
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return &types.RequestResult{
			Status:      resp.StatusCode,
			StatusText:  resp.Status,
			Error:       fmt.Sprintf("failed to read response body: %v", err),
			Duration:    duration,
			RequestSize: requestSize,
		}, nil
	}

	// Build response headers map
	headers := make(map[string]string)
	for key, values := range resp.Header {
		headers[key] = strings.Join(values, ", ")
	}

	// Parse GraphQL response to check for errors
	var graphqlResp struct {
		Data   interface{}            `json:"data"`
		Errors []interface{}          `json:"errors,omitempty"`
	}

	responseBody := string(bodyBytes)
	if err := json.Unmarshal(bodyBytes, &graphqlResp); err == nil {
		// Successfully parsed as GraphQL response
		if len(graphqlResp.Errors) > 0 {
			// GraphQL returned errors - format them nicely
			errorsJSON, _ := json.MarshalIndent(graphqlResp.Errors, "", "  ")
			responseBody = fmt.Sprintf("{\n  \"data\": %s,\n  \"errors\": %s\n}",
				formatJSON(graphqlResp.Data),
				string(errorsJSON))
		} else {
			// No errors, just show formatted data
			dataJSON, _ := json.MarshalIndent(graphqlResp.Data, "", "  ")
			responseBody = string(dataJSON)
		}
	}
	// If parsing fails, just use raw body

	result := &types.RequestResult{
		Status:       resp.StatusCode,
		StatusText:   resp.Status,
		Headers:      headers,
		Body:         responseBody,
		Duration:     duration,
		RequestSize:  requestSize,
		ResponseSize: len(bodyBytes),
	}

	return result, nil
}

// formatJSON formats any data structure as JSON
func formatJSON(data interface{}) string {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", data)
	}
	return string(jsonBytes)
}
