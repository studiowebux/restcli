package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

// Hop-by-hop headers that must not be forwarded
var hopHeaders = map[string]bool{
	"Connection":          true,
	"Keep-Alive":          true,
	"Proxy-Authenticate":  true,
	"Proxy-Authorization": true,
	"Proxy-Connection":    true,
	"Te":                  true,
	"Trailers":            true,
	"Transfer-Encoding":   true,
	"Upgrade":             true,
}

// ProxyLog represents a captured HTTP transaction
type ProxyLog struct {
	ID          int
	Timestamp   time.Time
	Method      string
	URL         string
	ReqHeaders  http.Header
	ReqBody     []byte
	Status      int
	StatusText  string
	RespHeaders http.Header
	RespBody    []byte
	Duration    time.Duration
}

// Proxy represents an HTTP debug proxy server
type Proxy struct {
	Port      int
	logs      []*ProxyLog
	logMutex  sync.RWMutex
	nextID    int
	server    *http.Server
	stopChan  chan struct{}
	maxLogs   int // Maximum number of logs to keep in memory
	notifyCh  chan struct{} // Channel to notify when new log arrives
}

// NewProxy creates a new debug proxy
func NewProxy(port int) *Proxy {
	return &Proxy{
		Port:     port,
		logs:     make([]*ProxyLog, 0),
		nextID:   1,
		stopChan: make(chan struct{}),
		maxLogs:  1000, // Keep last 1000 requests
		notifyCh: make(chan struct{}, 100), // Buffered channel for notifications
	}
}

// Start starts the proxy server
func (p *Proxy) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", p.handleProxy)

	p.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", p.Port),
		Handler: mux,
	}

	go func() {
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Proxy server error: %v\n", err)
		}
	}()

	return nil
}

// Stop stops the proxy server
func (p *Proxy) Stop() error {
	close(p.stopChan)
	if p.server != nil {
		return p.server.Close()
	}
	return nil
}

// handleProxy handles incoming proxy requests
func (p *Proxy) handleProxy(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Handle CONNECT method for HTTPS tunneling
	if r.Method == http.MethodConnect {
		p.handleConnect(w, r, startTime)
		return
	}

	// Build full URL for proxy request
	targetURL := r.URL.String()
	if r.URL.Scheme == "" {
		// For proxy requests, the URL should include scheme and host
		// If missing, construct from Host header
		if r.URL.Host == "" && r.Host != "" {
			targetURL = "http://" + r.Host + r.URL.Path
			if r.URL.RawQuery != "" {
				targetURL += "?" + r.URL.RawQuery
			}
		} else {
			http.Error(w, "Invalid proxy request: missing scheme and host", http.StatusBadRequest)
			return
		}
	}

	// Create log entry
	log := &ProxyLog{
		ID:         p.getNextID(),
		Timestamp:  startTime,
		Method:     r.Method,
		URL:        targetURL,
		ReqHeaders: r.Header.Clone(),
	}

	// Read request body
	var bodyBytes []byte
	if r.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(r.Body)
		if err == nil {
			log.ReqBody = bodyBytes
		}
	}

	// Create new request to target server
	var bodyReader io.Reader
	if len(bodyBytes) > 0 {
		bodyReader = bytes.NewReader(bodyBytes)
	}
	proxyReq, err := http.NewRequest(r.Method, targetURL, bodyReader)
	if err != nil {
		log.Status = http.StatusInternalServerError
		log.StatusText = fmt.Sprintf("Error creating proxy request: %v", err)
		log.Duration = time.Since(startTime)
		p.addLog(log)
		http.Error(w, log.StatusText, http.StatusInternalServerError)
		return
	}

	// Copy headers (skip hop-by-hop headers)
	for name, values := range r.Header {
		if hopHeaders[name] {
			continue
		}
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}

	// Execute request with transport that supports keep-alive
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}
	resp, err := client.Do(proxyReq)
	if err != nil {
		log.Status = http.StatusBadGateway
		log.StatusText = fmt.Sprintf("Error forwarding request: %v", err)
		log.Duration = time.Since(startTime)
		p.addLog(log)
		http.Error(w, log.StatusText, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Status = http.StatusInternalServerError
		log.StatusText = fmt.Sprintf("Error reading response: %v", err)
		log.Duration = time.Since(startTime)
		p.addLog(log)
		http.Error(w, log.StatusText, http.StatusInternalServerError)
		return
	}

	// Record response in log
	log.Status = resp.StatusCode
	log.StatusText = resp.Status
	log.RespHeaders = resp.Header.Clone()
	log.RespBody = respBody
	log.Duration = time.Since(startTime)

	// Add to logs
	p.addLog(log)

	// Copy response headers to client (skip hop-by-hop headers)
	for name, values := range resp.Header {
		if hopHeaders[name] {
			continue
		}
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// Write status and body to client
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)
}

// getNextID returns the next log ID and increments the counter
func (p *Proxy) getNextID() int {
	p.logMutex.Lock()
	defer p.logMutex.Unlock()
	id := p.nextID
	p.nextID++
	return id
}

// addLog adds a log entry and maintains the maximum log limit
func (p *Proxy) addLog(log *ProxyLog) {
	p.logMutex.Lock()
	defer p.logMutex.Unlock()

	p.logs = append(p.logs, log)

	// Keep only the last maxLogs entries
	if len(p.logs) > p.maxLogs {
		p.logs = p.logs[len(p.logs)-p.maxLogs:]
	}

	// Notify listeners (non-blocking)
	select {
	case p.notifyCh <- struct{}{}:
	default:
		// Channel full, skip notification
	}
}

// NotifyChannel returns the notification channel
func (p *Proxy) NotifyChannel() <-chan struct{} {
	return p.notifyCh
}

// GetLogs returns a copy of all logs
func (p *Proxy) GetLogs() []*ProxyLog {
	p.logMutex.RLock()
	defer p.logMutex.RUnlock()

	// Create a copy to avoid race conditions
	logs := make([]*ProxyLog, len(p.logs))
	copy(logs, p.logs)
	return logs
}

// GetLogCount returns the number of captured logs
func (p *Proxy) GetLogCount() int {
	p.logMutex.RLock()
	defer p.logMutex.RUnlock()
	return len(p.logs)
}

// ClearLogs clears all captured logs and resets the ID counter
func (p *Proxy) ClearLogs() {
	p.logMutex.Lock()
	defer p.logMutex.Unlock()
	p.logs = make([]*ProxyLog, 0)
	p.nextID = 0
}

// ExportHAR exports captured traffic to HAR format
func (p *Proxy) ExportHAR(filename string) error {
	// TODO(#TODO-004): Implement HAR export similar to HAR import converter
	// See TODO.md for HAR 1.2 spec and implementation details
	return fmt.Errorf("HAR export not yet implemented")
}

// FormatSize formats byte size as human-readable string
func FormatSize(bytes int) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatDuration formats duration as human-readable string
func FormatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%dns", d.Nanoseconds())
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%.2fµs", float64(d.Microseconds()))
	}
	if d < time.Second {
		return fmt.Sprintf("%.2fms", float64(d.Microseconds())/1000.0)
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

// handleConnect handles HTTPS CONNECT requests (tunneling)
func (p *Proxy) handleConnect(w http.ResponseWriter, r *http.Request, startTime time.Time) {
	// Log CONNECT request (we can't see the actual HTTPS content)
	log := &ProxyLog{
		ID:         p.getNextID(),
		Timestamp:  startTime,
		Method:     "CONNECT",
		URL:        "https://" + r.Host, // CONNECT uses Host for target
		ReqHeaders: r.Header.Clone(),
		ReqBody:    nil, // No body in CONNECT
	}

	// Establish connection to target
	targetConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		log.Status = http.StatusBadGateway
		log.StatusText = fmt.Sprintf("Failed to connect to %s: %v", r.Host, err)
		log.Duration = time.Since(startTime)
		p.addLog(log)
		http.Error(w, log.StatusText, http.StatusBadGateway)
		return
	}
	defer targetConn.Close()

	// Hijack the client connection
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		log.Status = http.StatusInternalServerError
		log.StatusText = "Hijacking not supported"
		log.Duration = time.Since(startTime)
		p.addLog(log)
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		log.Status = http.StatusInternalServerError
		log.StatusText = fmt.Sprintf("Failed to hijack connection: %v", err)
		log.Duration = time.Since(startTime)
		p.addLog(log)
		http.Error(w, log.StatusText, http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()

	// Send 200 Connection Established to client
	_, err = clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	if err != nil {
		log.Status = http.StatusInternalServerError
		log.StatusText = fmt.Sprintf("Failed to send response: %v", err)
		log.Duration = time.Since(startTime)
		p.addLog(log)
		return
	}

	// Log successful CONNECT
	log.Status = http.StatusOK
	log.StatusText = "200 Connection Established (HTTPS tunnel)"
	log.Duration = time.Since(startTime)
	p.addLog(log)

	// Bidirectional copy between client and target
	errCh := make(chan error, 2)
	go func() {
		_, err := io.Copy(targetConn, clientConn)
		errCh <- err
	}()
	go func() {
		_, err := io.Copy(clientConn, targetConn)
		errCh <- err
	}()

	// Wait for either direction to finish
	<-errCh
}
