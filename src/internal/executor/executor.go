package executor

import (
	"bytes"
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

// Execute performs an HTTP request and returns the result
func Execute(req *types.HttpRequest, tlsConfig *types.TLSConfig) (*types.RequestResult, error) {
	startTime := time.Now()

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
	client, err := buildHTTPClient(tlsConfig)
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

// buildHTTPClient creates an HTTP client with optional TLS/mTLS configuration
func buildHTTPClient(tlsConfig *types.TLSConfig) (*http.Client, error) {
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
		Timeout:   30 * time.Second,
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

// IsClientErrorStatus returns true if status code is 4xx
func IsClientErrorStatus(status int) bool {
	return status >= 400 && status < 500
}

// IsServerErrorStatus returns true if status code is 5xx
func IsServerErrorStatus(status int) bool {
	return status >= 500 && status < 600
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
