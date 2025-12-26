/*
Package executor handles HTTP request execution with support for multiple protocols.

# Overview

The executor package provides HTTP request execution capabilities including:
  - Standard HTTP/HTTPS requests
  - Server-Sent Events (SSE) streaming
  - WebSocket connections (interactive and scripted)
  - Variable resolution
  - TLS/mTLS configuration

# Request Types

HTTP Requests (http.go):
  - Standard RESTful requests (GET, POST, PUT, DELETE, etc.)
  - Request/response body handling
  - Header management
  - Timeout configuration
  - Redirect handling

Streaming Requests (streaming.go):
  - Server-Sent Events (SSE)
  - Real-time event delivery via callbacks
  - Context-based cancellation
  - Connection management

WebSocket Requests (websocket.go):
  - Scripted message sequences (send/receive patterns)
  - Interactive sessions with user input
  - Binary and text message support
  - Subprotocol negotiation

# Variable Resolution

Variables in requests are resolved through parser.VariableResolver:
  - Environment variables
  - Profile variables
  - Request dependencies (chain execution)
  - Dynamic values (timestamps, UUIDs, etc.)

# TLS Configuration

TLS support includes:
  - Custom CA certificates
  - Client certificates (mTLS)
  - InsecureSkipVerify for development
  - Certificate validation

# Error Handling

Errors are categorized as:
  - Network errors (connection failures, timeouts)
  - TLS errors (certificate issues)
  - Protocol errors (malformed responses)
  - Application errors (unexpected status codes)

Errors include detailed context for troubleshooting.

# Example Usage - HTTP Request

	request := &types.HttpRequest{
		Method: "POST",
		URL:    "https://api.example.com/users",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: `{"name": "John Doe"}`,
	}

	profile := &types.Profile{
		RequestTimeout: 30 * time.Second,
	}

	result, err := Execute(request, nil, profile)
	if err != nil {
		return err
	}

	fmt.Printf("Status: %d\n", result.Status)
	fmt.Printf("Body: %s\n", result.Body)

# Example Usage - WebSocket

	wsRequest := &types.WebSocketRequest{
		URL: "wss://api.example.com/ws",
		Messages: []types.WebSocketMessage{
			{
				Name:      "greeting",
				Type:      "text",
				Content:   "Hello",
				Direction: "send",
			},
			{
				Name:      "response",
				Direction: "receive",
				Timeout:   5,
			},
		},
	}

	result, err := ExecuteWebSocket(ctx, wsRequest, nil, func(msg *types.ReceivedMessage, done bool) {
		if done {
			fmt.Println("Connection closed")
			return
		}
		fmt.Printf("Message: %s\n", msg.Content)
	})

# Example Usage - Interactive WebSocket

	sendChan := make(chan string)

	go ExecuteWebSocketInteractive(
		ctx,
		"wss://api.example.com/ws",
		headers,
		subprotocols,
		tlsConfig,
		resolver,
		sendChan,
		callback,
	)

	// Send messages
	sendChan <- "user message"

# Thread Safety

The Execute functions are safe to call concurrently.
Each execution creates isolated HTTP clients and connections.

# Resource Management

All functions properly clean up resources:
  - HTTP connections are closed
  - WebSocket connections are gracefully terminated
  - Context cancellation is respected
  - Timeouts are enforced
*/
package executor
