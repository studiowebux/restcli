package executor

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/studiowebux/restcli/internal/types"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// TestExecuteWebSocket_SuccessfulConnection tests basic connection establishment
func TestExecuteWebSocket_SuccessfulConnection(t *testing.T) {
	// Create test WebSocket server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("Failed to upgrade connection: %v", err)
		}
		defer conn.Close()

		// Just accept the connection and wait for close
		_, _, _ = conn.ReadMessage()
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	req := &types.WebSocketRequest{
		URL:      wsURL,
		Headers:  map[string]string{},
		Messages: []types.WebSocketMessage{},
	}

	ctx := context.Background()
	result, err := ExecuteWebSocket(ctx, req, nil, nil)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Error != "" {
		t.Errorf("Expected no error in result, got: %s", result.Error)
	}

	if result.DisconnectReason != "Completed successfully" {
		t.Errorf("Expected 'Completed successfully', got: %s", result.DisconnectReason)
	}

	if result.Duration < 0 {
		t.Errorf("Expected non-negative duration, got: %d", result.Duration)
	}
}

// TestExecuteWebSocket_SendTextMessage tests sending text messages
func TestExecuteWebSocket_SendTextMessage(t *testing.T) {
	receivedMessages := []string{}
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			mu.Lock()
			receivedMessages = append(receivedMessages, string(message))
			mu.Unlock()
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	req := &types.WebSocketRequest{
		URL: wsURL,
		Messages: []types.WebSocketMessage{
			{
				Name:      "greeting",
				Type:      "text",
				Content:   "Hello, WebSocket!",
				Direction: "send",
			},
		},
	}

	ctx := context.Background()
	result, err := ExecuteWebSocket(ctx, req, nil, nil)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Error != "" {
		t.Errorf("Expected no error in result, got: %s", result.Error)
	}

	if result.SentCount != 1 {
		t.Errorf("Expected 1 sent message, got: %d", result.SentCount)
	}

	// Wait a bit for server to receive
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(receivedMessages) != 1 {
		t.Errorf("Expected 1 received message on server, got: %d", len(receivedMessages))
	} else if receivedMessages[0] != "Hello, WebSocket!" {
		t.Errorf("Expected 'Hello, WebSocket!', got: %s", receivedMessages[0])
	}
}

// TestExecuteWebSocket_SendJSONMessage tests sending and validating JSON messages
func TestExecuteWebSocket_SendJSONMessage(t *testing.T) {
	receivedJSON := []string{}
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			mu.Lock()
			receivedJSON = append(receivedJSON, string(message))
			mu.Unlock()
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	req := &types.WebSocketRequest{
		URL: wsURL,
		Messages: []types.WebSocketMessage{
			{
				Name:      "json-data",
				Type:      "json",
				Content:   `{"message":"test","count":42}`,
				Direction: "send",
			},
		},
	}

	ctx := context.Background()
	result, err := ExecuteWebSocket(ctx, req, nil, nil)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Error != "" {
		t.Errorf("Expected no error in result, got: %s", result.Error)
	}

	if result.SentCount != 1 {
		t.Errorf("Expected 1 sent message, got: %d", result.SentCount)
	}

	// Verify JSON was received
	time.Sleep(100 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	if len(receivedJSON) != 1 {
		t.Errorf("Expected 1 JSON message, got: %d", len(receivedJSON))
	} else {
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(receivedJSON[0]), &parsed); err != nil {
			t.Errorf("Received invalid JSON: %v", err)
		}
	}
}

// TestExecuteWebSocket_InvalidJSON tests JSON validation
func TestExecuteWebSocket_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		_, _, _ = conn.ReadMessage()
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	req := &types.WebSocketRequest{
		URL: wsURL,
		Messages: []types.WebSocketMessage{
			{
				Name:      "invalid-json",
				Type:      "json",
				Content:   `{invalid json}`,
				Direction: "send",
			},
		},
	}

	ctx := context.Background()
	result, err := ExecuteWebSocket(ctx, req, nil, nil)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Error == "" {
		t.Error("Expected error for invalid JSON, got none")
	}

	if !strings.Contains(result.Error, "invalid JSON") {
		t.Errorf("Expected 'invalid JSON' error, got: %s", result.Error)
	}
}

// TestExecuteWebSocket_ReceiveMessage tests receiving messages from server
func TestExecuteWebSocket_ReceiveMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Send a message immediately
		conn.WriteMessage(websocket.TextMessage, []byte("Server says hello"))

		// Wait for close
		_, _, _ = conn.ReadMessage()
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	req := &types.WebSocketRequest{
		URL: wsURL,
		Messages: []types.WebSocketMessage{
			{
				Name:      "wait-for-greeting",
				Type:      "text",
				Direction: "receive",
				Timeout:   5,
			},
		},
	}

	ctx := context.Background()
	result, err := ExecuteWebSocket(ctx, req, nil, nil)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Error != "" {
		t.Errorf("Expected no error in result, got: %s", result.Error)
	}

	if result.ReceivedCount != 1 {
		t.Errorf("Expected 1 received message, got: %d", result.ReceivedCount)
	}

	if len(result.Messages) != 1 {
		t.Errorf("Expected 1 message in result, got: %d", len(result.Messages))
	} else {
		if result.Messages[0].Content != "Server says hello" {
			t.Errorf("Expected 'Server says hello', got: %s", result.Messages[0].Content)
		}
		if result.Messages[0].Direction != "received" {
			t.Errorf("Expected direction 'received', got: %s", result.Messages[0].Direction)
		}
	}
}

// TestExecuteWebSocket_ReceiveTimeout tests timeout when waiting for message
func TestExecuteWebSocket_ReceiveTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Don't send anything - let it timeout
		time.Sleep(3 * time.Second)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	req := &types.WebSocketRequest{
		URL: wsURL,
		Messages: []types.WebSocketMessage{
			{
				Name:      "wait-for-message",
				Type:      "text",
				Direction: "receive",
				Timeout:   1, // 1 second timeout
			},
		},
	}

	ctx := context.Background()
	result, err := ExecuteWebSocket(ctx, req, nil, nil)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Error == "" {
		t.Error("Expected timeout error, got none")
	}

	if !strings.Contains(result.Error, "Timeout waiting for message") {
		t.Errorf("Expected timeout error, got: %s", result.Error)
	}
}

// TestExecuteWebSocket_ContextCancellation tests cancellation during execution
func TestExecuteWebSocket_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Wait for close
		_, _, _ = conn.ReadMessage()
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	req := &types.WebSocketRequest{
		URL: wsURL,
		Messages: []types.WebSocketMessage{
			{
				Name:      "wait-forever",
				Type:      "text",
				Direction: "receive",
				Timeout:   30,
			},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after 100ms
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	result, err := ExecuteWebSocket(ctx, req, nil, nil)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.DisconnectReason != "Cancelled by user" {
		t.Errorf("Expected 'Cancelled by user', got: %s", result.DisconnectReason)
	}
}

// TestExecuteWebSocket_InvalidURL tests connection failure with invalid URL
func TestExecuteWebSocket_InvalidURL(t *testing.T) {
	req := &types.WebSocketRequest{
		URL:      "not-a-valid-url",
		Messages: []types.WebSocketMessage{},
	}

	ctx := context.Background()
	result, err := ExecuteWebSocket(ctx, req, nil, nil)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Error == "" {
		t.Error("Expected error for invalid URL, got none")
	}

	if !strings.Contains(result.Error, "Invalid URL") && !strings.Contains(result.Error, "malformed ws") {
		t.Errorf("Expected 'Invalid URL' or 'malformed ws' error, got: %s", result.Error)
	}
}

// TestExecuteWebSocket_ConnectionRefused tests connection failure
func TestExecuteWebSocket_ConnectionRefused(t *testing.T) {
	req := &types.WebSocketRequest{
		URL:      "ws://localhost:9999", // Non-existent server
		Messages: []types.WebSocketMessage{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result, err := ExecuteWebSocket(ctx, req, nil, nil)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Error == "" {
		t.Error("Expected connection error, got none")
	}

	if !strings.Contains(result.Error, "Connection failed") {
		t.Errorf("Expected 'Connection failed' error, got: %s", result.Error)
	}
}

// TestExecuteWebSocket_Callback tests callback invocation
func TestExecuteWebSocket_Callback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Echo messages back
		for {
			msgType, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			conn.WriteMessage(msgType, message)
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	callbackMessages := []types.ReceivedMessage{}
	doneCount := 0
	var mu sync.Mutex

	callback := func(msg *types.ReceivedMessage, done bool) {
		mu.Lock()
		defer mu.Unlock()
		if done {
			doneCount++
		} else if msg != nil {
			callbackMessages = append(callbackMessages, *msg)
		}
	}

	req := &types.WebSocketRequest{
		URL: wsURL,
		Messages: []types.WebSocketMessage{
			{
				Name:      "send-test",
				Type:      "text",
				Content:   "test message",
				Direction: "send",
			},
			{
				Name:      "receive-echo",
				Type:      "text",
				Direction: "receive",
				Timeout:   5,
			},
		},
	}

	ctx := context.Background()
	result, err := ExecuteWebSocket(ctx, req, nil, callback)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Error != "" {
		t.Errorf("Expected no error in result, got: %s", result.Error)
	}

	mu.Lock()
	defer mu.Unlock()

	// Should have: connect message, sent message, received message
	if len(callbackMessages) < 3 {
		t.Errorf("Expected at least 3 callback messages (connect, sent, received), got: %d", len(callbackMessages))
	}

	// Should have 1 done callback
	if doneCount != 1 {
		t.Errorf("Expected 1 done callback, got: %d", doneCount)
	}

	// Verify connect message
	if callbackMessages[0].Direction != "system" {
		t.Errorf("Expected first message to be system connect, got: %s", callbackMessages[0].Direction)
	}

	// Verify sent message
	foundSent := false
	for _, msg := range callbackMessages {
		if msg.Direction == "sent" && msg.Content == "test message" {
			foundSent = true
			break
		}
	}
	if !foundSent {
		t.Error("Did not find sent message in callbacks")
	}

	// Verify received message
	foundReceived := false
	for _, msg := range callbackMessages {
		if msg.Direction == "received" && msg.Content == "test message" {
			foundReceived = true
			break
		}
	}
	if !foundReceived {
		t.Error("Did not find received echo message in callbacks")
	}
}

// TestExecuteWebSocket_SendReceiveSequence tests a complex send/receive sequence
func TestExecuteWebSocket_SendReceiveSequence(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Read first message and reply with specific response
		_, msg1, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if string(msg1) == "ping" {
			conn.WriteMessage(websocket.TextMessage, []byte("pong"))
		}

		// Read second message and reply
		_, msg2, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if string(msg2) == "hello" {
			conn.WriteMessage(websocket.TextMessage, []byte("world"))
		}

		// Wait for close
		_, _, _ = conn.ReadMessage()
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	req := &types.WebSocketRequest{
		URL: wsURL,
		Messages: []types.WebSocketMessage{
			{Name: "send-ping", Type: "text", Content: "ping", Direction: "send"},
			{Name: "receive-pong", Type: "text", Direction: "receive", Timeout: 5},
			{Name: "send-hello", Type: "text", Content: "hello", Direction: "send"},
			{Name: "receive-world", Type: "text", Direction: "receive", Timeout: 5},
		},
	}

	ctx := context.Background()
	result, err := ExecuteWebSocket(ctx, req, nil, nil)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Error != "" {
		t.Errorf("Expected no error in result, got: %s", result.Error)
	}

	if result.SentCount != 2 {
		t.Errorf("Expected 2 sent messages, got: %d", result.SentCount)
	}

	if result.ReceivedCount != 2 {
		t.Errorf("Expected 2 received messages, got: %d", result.ReceivedCount)
	}

	// Verify message sequence
	expectedContents := []string{"ping", "pong", "hello", "world"}
	if len(result.Messages) != len(expectedContents) {
		t.Errorf("Expected %d messages, got: %d", len(expectedContents), len(result.Messages))
	} else {
		for i, expected := range expectedContents {
			if result.Messages[i].Content != expected {
				t.Errorf("Message %d: expected '%s', got '%s'", i, expected, result.Messages[i].Content)
			}
		}
	}
}

// TestExecuteWebSocketInteractive_BasicFunctionality tests interactive mode
func TestExecuteWebSocketInteractive_BasicFunctionality(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Echo messages
		for {
			msgType, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			conn.WriteMessage(msgType, message)
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	sendChan := make(chan string, 10)
	receivedMessages := []types.ReceivedMessage{}
	var mu sync.Mutex

	callback := func(msg *types.ReceivedMessage, done bool) {
		if msg != nil {
			mu.Lock()
			receivedMessages = append(receivedMessages, *msg)
			mu.Unlock()
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start interactive session
	errChan := make(chan error, 1)
	go func() {
		err := ExecuteWebSocketInteractive(ctx, wsURL, map[string]string{}, []string{}, nil, nil, sendChan, callback)
		errChan <- err
	}()

	// Wait for connection
	time.Sleep(200 * time.Millisecond)

	// Send a message
	sendChan <- "test interactive"

	// Wait for echo
	time.Sleep(200 * time.Millisecond)

	// Cancel context to close
	cancel()

	// Wait for completion
	err := <-errChan
	if err != nil && err != context.Canceled {
		t.Errorf("Expected no error or context.Canceled, got: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	// Should have connect message, sent message, and received echo
	if len(receivedMessages) < 3 {
		t.Errorf("Expected at least 3 messages, got: %d", len(receivedMessages))
	}

	// Verify we got the echo back
	foundEcho := false
	for _, msg := range receivedMessages {
		if msg.Direction == "received" && msg.Content == "test interactive" {
			foundEcho = true
			break
		}
	}
	if !foundEcho {
		t.Error("Did not receive echo of sent message")
	}
}

// TestExecuteWebSocketInteractive_ContextCancellation tests cancellation in interactive mode
func TestExecuteWebSocketInteractive_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		_, _, _ = conn.ReadMessage()
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	sendChan := make(chan string, 10)
	doneCount := 0
	var mu sync.Mutex

	callback := func(msg *types.ReceivedMessage, done bool) {
		if done {
			mu.Lock()
			doneCount++
			mu.Unlock()
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Start interactive session
	errChan := make(chan error, 1)
	go func() {
		err := ExecuteWebSocketInteractive(ctx, wsURL, map[string]string{}, []string{}, nil, nil, sendChan, callback)
		errChan <- err
	}()

	// Wait for connection
	time.Sleep(100 * time.Millisecond)

	// Cancel context
	cancel()

	// Wait for completion
	err := <-errChan
	if err != nil {
		t.Errorf("Expected no error on graceful close, got: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if doneCount != 1 {
		t.Errorf("Expected 1 done callback, got: %d", doneCount)
	}
}

// TestSendMessage_BinaryMessage tests sending binary messages
func TestSendMessage_BinaryMessage(t *testing.T) {
	receivedType := 0
	receivedData := []byte{}
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		msgType, data, err := conn.ReadMessage()
		if err == nil {
			mu.Lock()
			receivedType = msgType
			receivedData = data
			mu.Unlock()
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	req := &types.WebSocketRequest{
		URL: wsURL,
		Messages: []types.WebSocketMessage{
			{
				Name:      "binary-data",
				Type:      "binary",
				Content:   "binary content here",
				Direction: "send",
			},
		},
	}

	ctx := context.Background()
	result, err := ExecuteWebSocket(ctx, req, nil, nil)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Error != "" {
		t.Errorf("Expected no error in result, got: %s", result.Error)
	}

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if receivedType != websocket.BinaryMessage {
		t.Errorf("Expected binary message type (%d), got: %d", websocket.BinaryMessage, receivedType)
	}

	if string(receivedData) != "binary content here" {
		t.Errorf("Expected 'binary content here', got: %s", string(receivedData))
	}
}

// TestExecuteWebSocketInteractive_EmptyMessage tests handling of empty messages
func TestExecuteWebSocketInteractive_EmptyMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		_, _, _ = conn.ReadMessage()
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	sendChan := make(chan string, 10)
	sentCount := 0
	var mu sync.Mutex

	callback := func(msg *types.ReceivedMessage, done bool) {
		if msg != nil && msg.Direction == "sent" {
			mu.Lock()
			sentCount++
			mu.Unlock()
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		err := ExecuteWebSocketInteractive(ctx, wsURL, map[string]string{}, []string{}, nil, nil, sendChan, callback)
		errChan <- err
	}()

	time.Sleep(100 * time.Millisecond)

	// Send empty string - should be skipped
	sendChan <- ""

	time.Sleep(100 * time.Millisecond)

	cancel()
	<-errChan

	mu.Lock()
	defer mu.Unlock()

	if sentCount != 0 {
		t.Errorf("Expected 0 sent messages (empty message should be skipped), got: %d", sentCount)
	}
}

// TestBuildWebSocketTLSConfig tests TLS configuration building
func TestBuildWebSocketTLSConfig(t *testing.T) {
	tests := []struct {
		name      string
		tlsConfig *types.TLSConfig
		wantErr   bool
	}{
		{
			name: "insecure skip verify",
			tlsConfig: &types.TLSConfig{
				InsecureSkipVerify: true,
			},
			wantErr: false,
		},
		{
			name: "invalid cert file",
			tlsConfig: &types.TLSConfig{
				CertFile: "/nonexistent/cert.pem",
				KeyFile:  "/nonexistent/key.pem",
			},
			wantErr: true,
		},
		{
			name: "invalid CA file",
			tlsConfig: &types.TLSConfig{
				CAFile: "/nonexistent/ca.pem",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := buildWebSocketTLSConfig(tt.tlsConfig)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if config == nil {
					t.Error("Expected TLS config, got nil")
				}
				if config.InsecureSkipVerify != tt.tlsConfig.InsecureSkipVerify {
					t.Errorf("Expected InsecureSkipVerify=%v, got %v",
						tt.tlsConfig.InsecureSkipVerify, config.InsecureSkipVerify)
				}
			}
		})
	}
}

// TestExecuteWebSocket_HeadersPropagation tests that headers are sent correctly
func TestExecuteWebSocket_HeadersPropagation(t *testing.T) {
	receivedHeaders := http.Header{}
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		receivedHeaders = r.Header.Clone()
		mu.Unlock()

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		_, _, _ = conn.ReadMessage()
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	req := &types.WebSocketRequest{
		URL: wsURL,
		Headers: map[string]string{
			"X-Custom-Header": "test-value",
			"Authorization":   "Bearer token123",
		},
		Messages: []types.WebSocketMessage{},
	}

	ctx := context.Background()
	result, err := ExecuteWebSocket(ctx, req, nil, nil)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Error != "" {
		t.Errorf("Expected no error in result, got: %s", result.Error)
	}

	mu.Lock()
	defer mu.Unlock()

	if receivedHeaders.Get("X-Custom-Header") != "test-value" {
		t.Errorf("Expected X-Custom-Header='test-value', got: %s", receivedHeaders.Get("X-Custom-Header"))
	}

	if receivedHeaders.Get("Authorization") != "Bearer token123" {
		t.Errorf("Expected Authorization='Bearer token123', got: %s", receivedHeaders.Get("Authorization"))
	}
}
