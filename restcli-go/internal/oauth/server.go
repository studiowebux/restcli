package oauth

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// CallbackResult contains the OAuth callback response
type CallbackResult struct {
	Code  string // Authorization code
	State string // State parameter for CSRF protection
	Error string // Error if authorization failed
}

// CallbackServer handles OAuth callbacks
type CallbackServer struct {
	server *http.Server
	result chan CallbackResult
}

// NewCallbackServer creates a new OAuth callback server
func NewCallbackServer(port int) *CallbackServer {
	cs := &CallbackServer{
		result: make(chan CallbackResult, 1),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", cs.handleCallback)

	cs.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	return cs
}

// handleCallback handles the OAuth callback request
func (cs *CallbackServer) handleCallback(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	result := CallbackResult{
		Code:  query.Get("code"),
		State: query.Get("state"),
		Error: query.Get("error"),
	}

	// Send result to channel
	select {
	case cs.result <- result:
	default:
	}

	// Send response to browser
	if result.Error != "" {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head><title>OAuth Error</title></head>
<body>
	<h1>Authentication Failed</h1>
	<p>Error: %s</p>
	<p>You can close this window.</p>
</body>
</html>`, result.Error)
	} else {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `
<!DOCTYPE html>
<html>
<head><title>OAuth Success</title></head>
<body>
	<h1>Authentication Successful!</h1>
	<p>You can close this window and return to the terminal.</p>
	<script>window.close();</script>
</body>
</html>`)
	}
}

// Start starts the callback server
func (cs *CallbackServer) Start() error {
	go func() {
		if err := cs.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Log error but don't crash
			fmt.Printf("Callback server error: %v\n", err)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)
	return nil
}

// WaitForCallback waits for the OAuth callback with timeout
func (cs *CallbackServer) WaitForCallback(timeout time.Duration) (*CallbackResult, error) {
	select {
	case result := <-cs.result:
		return &result, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting for OAuth callback")
	}
}

// Shutdown gracefully shuts down the server
func (cs *CallbackServer) Shutdown(ctx context.Context) error {
	return cs.server.Shutdown(ctx)
}
