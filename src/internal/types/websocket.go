package types

import "time"

// WebSocketRequest represents a WebSocket connection definition from .ws files
type WebSocketRequest struct {
	Name          string            `json:"name,omitempty" yaml:"name,omitempty"`
	URL           string            `json:"url" yaml:"url"`
	Headers       map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	Subprotocols  []string          `json:"subprotocols,omitempty" yaml:"subprotocols,omitempty"`
	Messages      []WebSocketMessage `json:"messages,omitempty" yaml:"messages,omitempty"`
	TLS           *TLSConfig        `json:"tls,omitempty" yaml:"tls,omitempty"`
	Documentation *Documentation    `json:"documentation,omitempty" yaml:"documentation,omitempty"`
}

// WebSocketMessage represents a message to send or expect in the sequence
type WebSocketMessage struct {
	Name      string `json:"name,omitempty" yaml:"name,omitempty"`       // Message name/label
	Type      string `json:"type" yaml:"type"`                           // "text" | "json" | "binary"
	Content   string `json:"content" yaml:"content"`                     // Message body
	Direction string `json:"direction" yaml:"direction"`                 // "send" | "receive"
	Timeout   int    `json:"timeout,omitempty" yaml:"timeout,omitempty"` // Timeout in seconds
}

// WebSocketResult contains the WebSocket session data
type WebSocketResult struct {
	Messages     []ReceivedMessage `json:"messages"`                 // All received messages
	SentCount    int               `json:"sentCount"`                // Number of messages sent
	ReceivedCount int              `json:"receivedCount"`            // Number of messages received
	Duration     int64             `json:"duration"`                 // Session duration in milliseconds
	Error        string            `json:"error,omitempty"`          // Error message if any
	Timestamp    string            `json:"timestamp,omitempty"`      // Session start time (RFC3339)
	DisconnectReason string        `json:"disconnectReason,omitempty"` // Reason for disconnection
}

// ReceivedMessage represents a single message received during the session
type ReceivedMessage struct {
	Type      string `json:"type"`                // "text" | "json" | "binary" | "ping" | "pong" | "close"
	Content   string `json:"content"`             // Message content
	Timestamp string `json:"timestamp"`           // When received (RFC3339)
	Direction string `json:"direction"`           // "sent" | "received"
	Size      int    `json:"size,omitempty"`      // Message size in bytes
}

// WebSocketConnection represents an active WebSocket connection state
type WebSocketConnection struct {
	URL          string
	Connected    bool
	ConnectedAt  time.Time
	Messages     []ReceivedMessage
	SentCount    int
	ReceivedCount int
	LastPing     time.Time
	LastPong     time.Time
	Error        error
}

// WebSocketCallback is called for each message received during WebSocket session
// done indicates if this is a disconnect/completion event
type WebSocketCallback func(message *ReceivedMessage, done bool)
