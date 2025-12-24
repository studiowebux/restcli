package tui

import (
	"context"
	"errors"
	"net"
	"net/url"
	"strings"
	"syscall"
	"testing"
)

func TestCategorizeRequestError(t *testing.T) {
	tests := []struct {
		name     string
		errStr   string
		wantText string
	}{
		{
			name:     "empty error",
			errStr:   "",
			wantText: "",
		},
		{
			name:     "context deadline exceeded",
			errStr:   "Get \"http://example.com\": context deadline exceeded",
			wantText: "Request timeout - check URL and try increasing timeout in profile settings (default: 30s)",
		},
		{
			name:     "DNS lookup failure",
			errStr:   "dial tcp: lookup nonexistent.example.com: no such host",
			wantText: "DNS resolution failed - verify hostname is correct and network is available",
		},
		{
			name:     "connection refused",
			errStr:   "dial tcp 127.0.0.1:9999: connect: connection refused",
			wantText: "Connection refused - check if server is running and port is correct",
		},
		{
			name:     "connection reset",
			errStr:   "read tcp 127.0.0.1:8080->127.0.0.1:54321: read: connection reset by peer",
			wantText: "Connection reset by server - server may have crashed or network issue occurred",
		},
		{
			name:     "TLS certificate unknown authority",
			errStr:   "x509: certificate signed by unknown authority",
			wantText: "TLS certificate verification failed - certificate is not trusted. Add CA certificate to profile or disable verification (insecure)",
		},
		{
			name:     "TLS certificate expired",
			errStr:   "x509: certificate has expired or is not yet valid",
			wantText: "TLS certificate has expired - contact server administrator or disable verification (insecure)",
		},
		{
			name:     "TLS hostname mismatch",
			errStr:   "x509: certificate is valid for example.com, not example.org",
			wantText: "TLS hostname mismatch - certificate doesn't match the requested hostname",
		},
		{
			name:     "TLS handshake failure",
			errStr:   "tls: handshake failure",
			wantText: "TLS handshake failed - check TLS version compatibility and cipher suites",
		},
		{
			name:     "network unreachable",
			errStr:   "dial tcp: network is unreachable",
			wantText: "Network unreachable - check network connection and firewall settings",
		},
		{
			name:     "too many redirects",
			errStr:   "Get \"http://example.com\": stopped after 10 redirects",
			wantText: "Too many redirects - check server configuration or URL",
		},
		{
			name:     "invalid URL",
			errStr:   "unsupported protocol scheme",
			wantText: "Invalid URL - verify the URL format and protocol (http/https)",
		},
		{
			name:     "proxy error",
			errStr:   "proxyconnect tcp: dial tcp 127.0.0.1:8080: connect: connection refused",
			wantText: "Proxy connection failed - verify proxy settings in profile configuration",
		},
		{
			name:     "EOF error",
			errStr:   "unexpected EOF",
			wantText: "Connection closed unexpectedly - server may have terminated the connection prematurely",
		},
		{
			name:     "generic timeout",
			errStr:   "i/o timeout",
			wantText: "Connection timeout - server took too long to respond, try increasing timeout",
		},
		{
			name:     "malformed HTTP",
			errStr:   "malformed HTTP response",
			wantText: "Malformed HTTP request - check request format, headers, and body",
		},
		{
			name:     "context canceled",
			errStr:   "context canceled",
			wantText: "Request cancelled by user",
		},
		{
			name:     "unknown error",
			errStr:   "something went wrong",
			wantText: "Request failed: something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := categorizeRequestError(tt.errStr)
			if got != tt.wantText {
				t.Errorf("categorizeRequestError() = %q, want %q", got, tt.wantText)
			}
		})
	}
}

func TestCategorizeError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantText string
	}{
		{
			name:     "nil error",
			err:      nil,
			wantText: "",
		},
		{
			name:     "context deadline exceeded",
			err:      context.DeadlineExceeded,
			wantText: "Request timeout - check URL and try increasing timeout in profile settings (default: 30s)",
		},
		{
			name:     "context canceled",
			err:      context.Canceled,
			wantText: "Request cancelled by user",
		},
		{
			name:     "url error with timeout",
			err:      &url.Error{Op: "Get", URL: "http://example.com", Err: context.DeadlineExceeded},
			wantText: "Request timeout - check URL and try increasing timeout in profile settings (default: 30s)",
		},
		{
			name: "net op error with connection refused",
			err: &net.OpError{
				Op:  "dial",
				Net: "tcp",
				Err: syscall.ECONNREFUSED,
			},
			wantText: "Connection refused - check if server is running and port is correct",
		},
		{
			name: "net op error with connection reset",
			err: &net.OpError{
				Op:  "read",
				Net: "tcp",
				Err: syscall.ECONNRESET,
			},
			wantText: "Connection reset by server - server may have crashed or network issue occurred",
		},
		{
			name: "net op error with network unreachable",
			err: &net.OpError{
				Op:  "dial",
				Net: "tcp",
				Err: syscall.ENETUNREACH,
			},
			wantText: "Network unreachable - check network connection and firewall settings",
		},
		{
			name:     "wrapped error",
			err:      errors.New("dial tcp: lookup nonexistent.example.com: no such host"),
			wantText: "DNS resolution failed - verify hostname is correct and network is available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := categorizeError(tt.err)
			if got != tt.wantText {
				t.Errorf("categorizeError() = %q, want %q", got, tt.wantText)
			}
		})
	}
}

func TestCategorizeSSLError(t *testing.T) {
	tests := []struct {
		name     string
		errStr   string
		wantText string
	}{
		{
			name:     "unknown authority",
			errStr:   "x509: certificate signed by unknown authority",
			wantText: "TLS certificate verification failed - certificate is not trusted. Add CA certificate to profile or disable verification (insecure)",
		},
		{
			name:     "certificate expired",
			errStr:   "x509: certificate has expired",
			wantText: "TLS certificate has expired - contact server administrator or disable verification (insecure)",
		},
		{
			name:     "hostname mismatch",
			errStr:   "x509: certificate is valid for example.com, not example.org",
			wantText: "TLS hostname mismatch - certificate doesn't match the requested hostname",
		},
		{
			name:     "handshake failure",
			errStr:   "tls: handshake failure",
			wantText: "TLS handshake failed - check TLS version compatibility and cipher suites",
		},
		{
			name:     "bad certificate",
			errStr:   "tls: bad certificate",
			wantText: "TLS bad certificate - client certificate may be invalid or not accepted by server",
		},
		{
			name:     "certificate required",
			errStr:   "tls: certificate required",
			wantText: "TLS client certificate required - configure client certificate in profile settings",
		},
		{
			name:     "generic TLS error",
			errStr:   "tls: some other error",
			wantText: "TLS/SSL error - check certificate configuration and TLS settings: tls: some other error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := categorizeSSLError(tt.errStr)
			if got != tt.wantText {
				t.Errorf("categorizeSSLError() = %q, want %q", got, tt.wantText)
			}
		})
	}
}

func TestCategorizeErrorFormatsCorrectly(t *testing.T) {
	// Test that all categorized messages don't include redundant prefixes
	errorStrings := []string{
		"context deadline exceeded",
		"no such host",
		"connection refused",
		"x509: certificate signed by unknown authority",
	}

	for _, errStr := range errorStrings {
		got := categorizeRequestError(errStr)
		// Should not have "Request failed:" prefix for known errors
		if strings.HasPrefix(got, "Request failed:") && !strings.Contains(errStr, "something went wrong") {
			t.Errorf("categorizeRequestError(%q) should not have 'Request failed:' prefix for known errors, got %q", errStr, got)
		}
		// Should not have "Error:" prefix
		if strings.HasPrefix(got, "Error:") {
			t.Errorf("categorizeRequestError(%q) should not have 'Error:' prefix, got %q", errStr, got)
		}
	}
}
