package tui

import (
	"context"
	"crypto/x509"
	"errors"
	"net"
	"net/url"
	"strings"
	"syscall"
)

// categorizeRequestError analyzes error strings from HTTP requests and provides
// actionable, user-friendly error messages based on the error type.
func categorizeRequestError(errStr string) string {
	if errStr == "" {
		return ""
	}

	errLower := strings.ToLower(errStr)

	// Context cancellation (user cancelled or timeout)
	if strings.Contains(errLower, "context canceled") ||
		strings.Contains(errLower, "context cancelled") {
		return "Request cancelled by user"
	}

	if strings.Contains(errLower, "context deadline exceeded") ||
		strings.Contains(errLower, "deadline exceeded") {
		return "Request timeout - check URL and try increasing timeout in profile settings (default: 30s)"
	}

	// Proxy errors (check before connection errors since proxy errors often contain "connection refused")
	if strings.Contains(errLower, "proxy") {
		return "Proxy connection failed - verify proxy settings in profile configuration"
	}

	// DNS resolution errors
	if strings.Contains(errLower, "no such host") ||
		strings.Contains(errLower, "dns") ||
		strings.Contains(errLower, "dial tcp: lookup") {
		return "DNS resolution failed - verify hostname is correct and network is available"
	}

	// Connection refused (server not running)
	if strings.Contains(errLower, "connection refused") ||
		strings.Contains(errLower, "connect: connection refused") {
		return "Connection refused - check if server is running and port is correct"
	}

	// Connection reset
	if strings.Contains(errLower, "connection reset") {
		return "Connection reset by server - server may have crashed or network issue occurred"
	}

	// Network unreachable
	if strings.Contains(errLower, "network is unreachable") ||
		strings.Contains(errLower, "no route to host") {
		return "Network unreachable - check network connection and firewall settings"
	}

	// TLS/SSL errors
	if strings.Contains(errLower, "tls") ||
		strings.Contains(errLower, "ssl") ||
		strings.Contains(errLower, "certificate") ||
		strings.Contains(errLower, "x509") {
		return categorizeSSLError(errStr)
	}

	// Too many redirects
	if strings.Contains(errLower, "stopped after") && strings.Contains(errLower, "redirect") {
		return "Too many redirects - check server configuration or URL"
	}

	// Invalid URL
	if strings.Contains(errLower, "invalid url") ||
		strings.Contains(errLower, "unsupported protocol") {
		return "Invalid URL - verify the URL format and protocol (http/https)"
	}

	// EOF errors (connection closed unexpectedly)
	if strings.Contains(errLower, "eof") ||
		strings.Contains(errLower, "unexpected eof") {
		return "Connection closed unexpectedly - server may have terminated the connection prematurely"
	}

	// Timeout (generic)
	if strings.Contains(errLower, "timeout") ||
		strings.Contains(errLower, "timed out") {
		return "Connection timeout - server took too long to respond, try increasing timeout"
	}

	// Protocol errors
	if strings.Contains(errLower, "malformed http") ||
		strings.Contains(errLower, "bad request") {
		return "Malformed HTTP request - check request format, headers, and body"
	}

	// Return original error with a generic prefix if we can't categorize it
	return "Request failed: " + errStr
}

// categorizeSSLError provides specific guidance for TLS/SSL certificate errors
func categorizeSSLError(errStr string) string {
	errLower := strings.ToLower(errStr)

	// Certificate verification errors
	if strings.Contains(errLower, "certificate is not trusted") ||
		strings.Contains(errLower, "unknown authority") ||
		strings.Contains(errLower, "certificate signed by unknown authority") {
		return "TLS certificate verification failed - certificate is not trusted. Add CA certificate to profile or disable verification (insecure)"
	}

	if strings.Contains(errLower, "certificate has expired") ||
		strings.Contains(errLower, "expired") {
		return "TLS certificate has expired - contact server administrator or disable verification (insecure)"
	}

	if strings.Contains(errLower, "certificate is valid for") ||
		strings.Contains(errLower, "name mismatch") ||
		strings.Contains(errLower, "doesn't match") {
		return "TLS hostname mismatch - certificate doesn't match the requested hostname"
	}

	if strings.Contains(errLower, "handshake failure") ||
		strings.Contains(errLower, "handshake") {
		return "TLS handshake failed - check TLS version compatibility and cipher suites"
	}

	if strings.Contains(errLower, "bad certificate") {
		return "TLS bad certificate - client certificate may be invalid or not accepted by server"
	}

	if strings.Contains(errLower, "certificate required") {
		return "TLS client certificate required - configure client certificate in profile settings"
	}

	// Generic TLS error
	return "TLS/SSL error - check certificate configuration and TLS settings: " + errStr
}

// categorizeError is a helper that wraps categorizeRequestError for use with Go error types.
// It handles nil errors and unwraps the error chain to get the root cause.
func categorizeError(err error) string {
	if err == nil {
		return ""
	}

	// Unwrap to get root cause
	var rootErr error = err
	for {
		unwrapped := errors.Unwrap(rootErr)
		if unwrapped == nil {
			break
		}
		rootErr = unwrapped
	}

	// Check for specific error types
	switch e := rootErr.(type) {
	case *url.Error:
		return categorizeURLError(e)
	case *net.OpError:
		return categorizeNetError(e)
	case x509.CertificateInvalidError:
		return "TLS certificate is invalid: " + e.Error()
	case x509.UnknownAuthorityError:
		return "TLS certificate signed by unknown authority - add CA certificate to profile or disable verification (insecure)"
	}

	// Check for context errors
	if errors.Is(err, context.DeadlineExceeded) {
		return "Request timeout - check URL and try increasing timeout in profile settings (default: 30s)"
	}
	if errors.Is(err, context.Canceled) {
		return "Request cancelled by user"
	}

	// Fall back to string-based categorization
	return categorizeRequestError(err.Error())
}

// categorizeURLError provides specific handling for url.Error types
func categorizeURLError(e *url.Error) string {
	if e.Timeout() {
		return "Request timeout - check URL and try increasing timeout in profile settings (default: 30s)"
	}
	if e.Temporary() {
		return "Temporary network error - retry the request: " + e.Error()
	}
	return categorizeError(e.Err)
}

// categorizeNetError provides specific handling for net.OpError types
func categorizeNetError(e *net.OpError) string {
	if e.Timeout() {
		return "Connection timeout - server took too long to respond, try increasing timeout"
	}

	// Check for specific syscall errors
	if errno, ok := e.Err.(syscall.Errno); ok {
		switch errno {
		case syscall.ECONNREFUSED:
			return "Connection refused - check if server is running and port is correct"
		case syscall.ECONNRESET:
			return "Connection reset by server - server may have crashed or network issue occurred"
		case syscall.ENETUNREACH:
			return "Network unreachable - check network connection and firewall settings"
		case syscall.EHOSTUNREACH:
			return "Host unreachable - check if server is online and accessible"
		}
	}

	return categorizeError(e.Err)
}
