package executor

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/studiowebux/restcli/internal/parser"
	"github.com/studiowebux/restcli/internal/types"
)

// ExecuteWebSocket connects to a WebSocket endpoint and executes the message sequence
func ExecuteWebSocket(ctx context.Context, req *types.WebSocketRequest, tlsConfig *types.TLSConfig, callback types.WebSocketCallback) (*types.WebSocketResult, error) {
	startTime := time.Now()

	result := &types.WebSocketResult{
		Messages:  []types.ReceivedMessage{},
		Timestamp: startTime.Format(time.RFC3339),
	}

	// Parse and validate URL
	u, err := url.Parse(req.URL)
	if err != nil {
		result.Error = fmt.Sprintf("Invalid URL: %v", err)
		return result, nil
	}

	// Build WebSocket dialer with TLS config if needed
	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
	}

	// Configure TLS if needed
	if tlsConfig != nil && (u.Scheme == "wss" || u.Scheme == "https") {
		tlsClientConfig, err := buildWebSocketTLSConfig(tlsConfig)
		if err != nil {
			result.Error = fmt.Sprintf("TLS configuration error: %v", err)
			return result, nil
		}
		dialer.TLSClientConfig = tlsClientConfig
	}

	// Prepare headers
	headers := http.Header{}
	for key, value := range req.Headers {
		headers.Set(key, value)
	}

	// Set subprotocols if specified
	dialer.Subprotocols = req.Subprotocols

	// Connect to WebSocket
	conn, resp, err := dialer.DialContext(ctx, req.URL, headers)
	if err != nil {
		errMsg := fmt.Sprintf("Connection failed: %v", err)
		if resp != nil {
			errMsg = fmt.Sprintf("Connection failed (HTTP %d): %v", resp.StatusCode, err)
		}
		result.Error = errMsg
		result.Duration = time.Since(startTime).Milliseconds()
		return result, nil
	}
	defer conn.Close()

	// Notify connection established
	if callback != nil {
		connectMsg := &types.ReceivedMessage{
			Type:      "connect",
			Content:   fmt.Sprintf("Connected to %s", req.URL),
			Timestamp: time.Now().Format(time.RFC3339),
			Direction: "system",
		}
		callback(connectMsg, false)
	}

	// Channel for receiving messages
	receiveChan := make(chan types.ReceivedMessage, 100)
	receiveErrChan := make(chan error, 1)

	// Start receive goroutine
	go receiveMessages(conn, receiveChan, receiveErrChan)

	// Execute message sequence
	for _, msg := range req.Messages {
		select {
		case <-ctx.Done():
			result.DisconnectReason = "Cancelled by user"
			result.Duration = time.Since(startTime).Milliseconds()
			return result, nil
		default:
		}

		if msg.Direction == "send" {
			// Send message
			if err := sendMessage(conn, &msg); err != nil {
				result.Error = fmt.Sprintf("Failed to send message '%s': %v", msg.Name, err)
				result.Duration = time.Since(startTime).Milliseconds()
				return result, nil
			}

			result.SentCount++

			// Record sent message
			sentMsg := types.ReceivedMessage{
				Type:      msg.Type,
				Content:   msg.Content,
				Timestamp: time.Now().Format(time.RFC3339),
				Direction: "sent",
				Size:      len(msg.Content),
			}
			result.Messages = append(result.Messages, sentMsg)

			if callback != nil {
				callback(&sentMsg, false)
			}

		} else if msg.Direction == "receive" {
			// Wait for expected message with timeout
			timeout := time.Duration(msg.Timeout) * time.Second
			timer := time.NewTimer(timeout)

			select {
			case receivedMsg := <-receiveChan:
				timer.Stop()
				result.ReceivedCount++
				result.Messages = append(result.Messages, receivedMsg)

				if callback != nil {
					callback(&receivedMsg, false)
				}

				// Optional: validate received message matches expected
				// For now, we just record it

			case err := <-receiveErrChan:
				timer.Stop()
				result.Error = fmt.Sprintf("Receive error: %v", err)
				result.Duration = time.Since(startTime).Milliseconds()
				return result, nil

			case <-timer.C:
				result.Error = fmt.Sprintf("Timeout waiting for message '%s' (%ds)", msg.Name, msg.Timeout)
				result.Duration = time.Since(startTime).Milliseconds()
				return result, nil

			case <-ctx.Done():
				timer.Stop()
				result.DisconnectReason = "Cancelled by user"
				result.Duration = time.Since(startTime).Milliseconds()
				return result, nil
			}
		}
	}

	// Graceful close
	err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		// Ignore close errors as connection might already be closed
	}

	result.Duration = time.Since(startTime).Milliseconds()
	result.DisconnectReason = "Completed successfully"

	// Notify completion
	if callback != nil {
		callback(nil, true)
	}

	return result, nil
}

// sendMessage sends a message through the WebSocket connection
func sendMessage(conn *websocket.Conn, msg *types.WebSocketMessage) error {
	var messageType int

	switch strings.ToLower(msg.Type) {
	case "text":
		messageType = websocket.TextMessage
	case "json":
		messageType = websocket.TextMessage
		// Validate JSON
		var js json.RawMessage
		if err := json.Unmarshal([]byte(msg.Content), &js); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}
	case "binary":
		messageType = websocket.BinaryMessage
	default:
		messageType = websocket.TextMessage
	}

	return conn.WriteMessage(messageType, []byte(msg.Content))
}

// receiveMessages continuously receives messages from the WebSocket connection
func receiveMessages(conn *websocket.Conn, msgChan chan<- types.ReceivedMessage, errChan chan<- error) {
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				errChan <- err
			}
			return
		}

		var msgTypeStr string
		switch messageType {
		case websocket.TextMessage:
			msgTypeStr = "text"
		case websocket.BinaryMessage:
			msgTypeStr = "binary"
		case websocket.PingMessage:
			msgTypeStr = "ping"
		case websocket.PongMessage:
			msgTypeStr = "pong"
		case websocket.CloseMessage:
			msgTypeStr = "close"
		default:
			msgTypeStr = "unknown"
		}

		receivedMsg := types.ReceivedMessage{
			Type:      msgTypeStr,
			Content:   string(message),
			Timestamp: time.Now().Format(time.RFC3339),
			Direction: "received",
			Size:      len(message),
		}

		msgChan <- receivedMsg
	}
}

// buildWebSocketTLSConfig creates a TLS configuration for WebSocket connections
func buildWebSocketTLSConfig(tlsConfig *types.TLSConfig) (*tls.Config, error) {
	config := &tls.Config{
		InsecureSkipVerify: tlsConfig.InsecureSkipVerify,
	}

	// Load client certificate if specified
	if tlsConfig.CertFile != "" && tlsConfig.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(tlsConfig.CertFile, tlsConfig.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		config.Certificates = []tls.Certificate{cert}
	}

	// Load CA certificate if specified
	if tlsConfig.CAFile != "" {
		caCert, err := os.ReadFile(tlsConfig.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		config.RootCAs = caCertPool
	}

	return config, nil
}

// ExecuteWebSocketInteractive establishes a persistent WebSocket connection
// and listens for messages to send via sendChan
func ExecuteWebSocketInteractive(ctx context.Context, url string, headers map[string]string, subprotocols []string, tlsConfig *types.TLSConfig, resolver *parser.VariableResolver, sendChan <-chan string, callback types.WebSocketCallback) error {
	startTime := time.Now()

	// Build WebSocket dialer
	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
	}

	// Configure TLS if needed
	if tlsConfig != nil && (strings.HasPrefix(url, "wss://") || strings.HasPrefix(url, "https://")) {
		tlsClientConfig, err := buildWebSocketTLSConfig(tlsConfig)
		if err != nil {
			return fmt.Errorf("TLS configuration error: %w", err)
		}
		dialer.TLSClientConfig = tlsClientConfig
	}

	// Prepare headers
	headerMap := make(map[string][]string)
	for key, value := range headers {
		headerMap[key] = []string{value}
	}

	// Set subprotocols
	dialer.Subprotocols = subprotocols

	// Connect to WebSocket
	conn, resp, err := dialer.DialContext(ctx, url, headerMap)
	if err != nil {
		errMsg := fmt.Sprintf("Connection failed: %v", err)
		if resp != nil {
			errMsg = fmt.Sprintf("Connection failed (HTTP %d): %v", resp.StatusCode, err)
		}
		return fmt.Errorf("%s", errMsg)
	}
	defer conn.Close()

	// Notify connection established
	if callback != nil {
		connectMsg := &types.ReceivedMessage{
			Type:      "system",
			Content:   fmt.Sprintf("Connected to %s", url),
			Timestamp: time.Now().Format(time.RFC3339),
			Direction: "system",
		}
		callback(connectMsg, false)
	}

	// Channels for coordination
	receiveChan := make(chan types.ReceivedMessage, 100)
	receiveErrChan := make(chan error, 1)
	done := make(chan struct{})

	// Start receive goroutine
	go func() {
		defer close(done)
		receiveMessages(conn, receiveChan, receiveErrChan)
	}()

	// Main loop - handle both sending and receiving
	for {
		select {
		case <-ctx.Done():
			// Context cancelled - close connection gracefully
			_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if callback != nil {
				callback(nil, true)
			}
			return nil

		case message := <-sendChan:
			// User wants to send a message
			if message == "" {
				continue
			}

			// Resolve variables in message if resolver is provided
			resolvedMessage := message
			if resolver != nil {
				resolved, err := resolver.Resolve(message)
				if err != nil {
					if callback != nil {
						errorMsg := &types.ReceivedMessage{
							Type:      "system",
							Content:   fmt.Sprintf("Variable resolution failed: %v", err),
							Timestamp: time.Now().Format(time.RFC3339),
							Direction: "system",
						}
						callback(errorMsg, false)
					}
					continue
				}
				resolvedMessage = resolved
			}

			// Send as text message
			err := conn.WriteMessage(websocket.TextMessage, []byte(resolvedMessage))
			if err != nil {
				if callback != nil {
					errorMsg := &types.ReceivedMessage{
						Type:      "system",
						Content:   fmt.Sprintf("Failed to send: %v", err),
						Timestamp: time.Now().Format(time.RFC3339),
						Direction: "system",
					}
					callback(errorMsg, false)
				}
				continue
			}

			// Record sent message
			if callback != nil {
				sentMsg := types.ReceivedMessage{
					Type:      "text",
					Content:   message,
					Timestamp: time.Now().Format(time.RFC3339),
					Direction: "sent",
					Size:      len(message),
				}
				callback(&sentMsg, false)
			}

		case receivedMsg := <-receiveChan:
			// Received message from WebSocket
			if callback != nil {
				callback(&receivedMsg, false)
			}

		case err := <-receiveErrChan:
			// Error receiving from WebSocket
			if callback != nil {
				errorMsg := &types.ReceivedMessage{
					Type:      "system",
					Content:   fmt.Sprintf("Receive error: %v", err),
					Timestamp: time.Now().Format(time.RFC3339),
					Direction: "system",
				}
				callback(errorMsg, false)
			}
			return err

		case <-done:
			// Receive goroutine finished (connection closed)
			duration := time.Since(startTime).Milliseconds()
			if callback != nil {
				disconnectMsg := &types.ReceivedMessage{
					Type:      "system",
					Content:   fmt.Sprintf("Disconnected after %dms", duration),
					Timestamp: time.Now().Format(time.RFC3339),
					Direction: "system",
				}
				callback(disconnectMsg, false)
				callback(nil, true)
			}
			return nil
		}
	}
}

