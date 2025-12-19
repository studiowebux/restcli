package mock

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Server represents the mock HTTP server
type Server struct {
	config     *Config
	httpServer *http.Server
	logs       []RequestLog
	logsMutex  sync.RWMutex
	workdir    string
	notifyCh   chan struct{} // Channel to notify when new log arrives
}

// NewServer creates a new mock server
func NewServer(config *Config, workdir string) *Server {
	if config.Port == 0 {
		config.Port = 8080
	}
	if config.Host == "" {
		config.Host = "localhost"
	}

	return &Server{
		config:   config,
		logs:     make([]RequestLog, 0),
		workdir:  workdir,
		notifyCh: make(chan struct{}, 100), // Buffered channel for notifications
	}
}

// Start starts the mock server
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRequest)

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Mock server error: %v", err)
		}
	}()

	return nil
}

// Stop stops the mock server
func (s *Server) Stop() error {
	if s.httpServer == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.httpServer.Shutdown(ctx)
}

// handleRequest handles incoming HTTP requests
func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Read request body
	bodyBytes, _ := io.ReadAll(r.Body)
	r.Body.Close()
	requestBody := string(bodyBytes)

	// Find matching route
	route := s.findMatchingRoute(r.Method, r.URL.Path)

	var status int
	var responseBody string
	var matchedRule string

	if route == nil {
		// No matching route - return 404
		status = http.StatusNotFound
		responseBody = fmt.Sprintf("Mock server: No route configured for %s %s", r.Method, r.URL.Path)
		matchedRule = "none"
	} else {
		// Apply delay if configured
		if route.Delay > 0 {
			time.Sleep(time.Duration(route.Delay) * time.Millisecond)
		}

		// Set status
		status = route.Status
		if status == 0 {
			status = http.StatusOK
		}

		// Set headers
		for key, value := range route.Headers {
			w.Header().Set(key, value)
		}

		// Get response body
		if route.BodyFile != "" {
			// Load from file
			filePath := route.BodyFile
			if !filepath.IsAbs(filePath) {
				filePath = filepath.Join(s.workdir, filePath)
			}
			bodyBytes, err := os.ReadFile(filePath)
			if err != nil {
				status = http.StatusInternalServerError
				responseBody = fmt.Sprintf("Mock server: Failed to read body file %s: %v", route.BodyFile, err)
			} else {
				responseBody = string(bodyBytes)
			}
		} else {
			responseBody = route.Body
		}

		matchedRule = route.Name
		if matchedRule == "" {
			matchedRule = fmt.Sprintf("%s %s", route.Method, route.Path)
		}
	}

	// Write response
	w.WriteHeader(status)
	w.Write([]byte(responseBody))

	duration := time.Since(start)

	// Log request
	if s.config.Logging {
		s.logRequest(RequestLog{
			Timestamp:   start,
			Method:      r.Method,
			Path:        r.URL.Path,
			Headers:     flattenHeaders(r.Header),
			Body:        requestBody,
			MatchedRule: matchedRule,
			Status:      status,
			Duration:    duration,
		})
	}
}

// findMatchingRoute finds the first route that matches the method and path
func (s *Server) findMatchingRoute(method, path string) *Route {
	for _, route := range s.config.Routes {
		if !strings.EqualFold(route.Method, method) {
			continue
		}

		pathType := route.PathType
		if pathType == "" {
			pathType = "exact"
		}

		matched := false
		switch pathType {
		case "exact":
			matched = route.Path == path
		case "prefix":
			matched = strings.HasPrefix(path, route.Path)
		case "regex":
			if re, err := regexp.Compile(route.Path); err == nil {
				matched = re.MatchString(path)
			}
		}

		if matched {
			return &route
		}
	}

	return nil
}

// logRequest adds a request to the log
func (s *Server) logRequest(log RequestLog) {
	s.logsMutex.Lock()
	defer s.logsMutex.Unlock()

	s.logs = append(s.logs, log)

	// Keep only last 1000 logs
	if len(s.logs) > 1000 {
		s.logs = s.logs[len(s.logs)-1000:]
	}

	// Notify listeners (non-blocking)
	select {
	case s.notifyCh <- struct{}{}:
	default:
		// Channel full, skip notification
	}
}

// NotifyChannel returns the notification channel
func (s *Server) NotifyChannel() <-chan struct{} {
	return s.notifyCh
}

// GetLogs returns all logged requests
func (s *Server) GetLogs() []RequestLog {
	s.logsMutex.RLock()
	defer s.logsMutex.RUnlock()

	// Return a copy
	logs := make([]RequestLog, len(s.logs))
	copy(logs, s.logs)
	return logs
}

// ClearLogs clears all logged requests
func (s *Server) ClearLogs() {
	s.logsMutex.Lock()
	defer s.logsMutex.Unlock()

	s.logs = make([]RequestLog, 0)
}

// GetAddress returns the server address
func (s *Server) GetAddress() string {
	return fmt.Sprintf("http://%s:%d", s.config.Host, s.config.Port)
}

// flattenHeaders converts http.Header to map[string]string (first value only)
func flattenHeaders(headers http.Header) map[string]string {
	result := make(map[string]string)
	for key, values := range headers {
		if len(values) > 0 {
			result[key] = values[0]
		}
	}
	return result
}
