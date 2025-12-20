package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseWebSocketFile_BasicConnection(t *testing.T) {
	content := `WEBSOCKET ws://localhost:8080

### Hello
> Hello, World!
<
`
	tmpFile := createTempWSFile(t, content)
	defer os.Remove(tmpFile)

	result, err := ParseWebSocketFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseWebSocketFile failed: %v", err)
	}

	if result.URL != "ws://localhost:8080" {
		t.Errorf("Expected URL ws://localhost:8080, got %s", result.URL)
	}

	// Should create 2 messages: one send, one receive
	if len(result.Messages) != 2 {
		t.Fatalf("Expected 2 messages (send + receive), got %d", len(result.Messages))
	}

	// First message: send
	sendMsg := result.Messages[0]
	if sendMsg.Name != "Hello" {
		t.Errorf("Expected message name 'Hello', got '%s'", sendMsg.Name)
	}
	if sendMsg.Direction != "send" {
		t.Errorf("Expected direction 'send', got '%s'", sendMsg.Direction)
	}
	if sendMsg.Content != "Hello, World!" {
		t.Errorf("Expected content 'Hello, World!', got '%s'", sendMsg.Content)
	}
	if sendMsg.Type != "text" {
		t.Errorf("Expected type 'text', got '%s'", sendMsg.Type)
	}

	// Second message: receive
	recvMsg := result.Messages[1]
	if recvMsg.Direction != "receive" {
		t.Errorf("Expected direction 'receive', got '%s'", recvMsg.Direction)
	}
	if recvMsg.Timeout != 30 {
		t.Errorf("Expected timeout 30, got %d", recvMsg.Timeout)
	}
}

func TestParseWebSocketFile_Headers(t *testing.T) {
	content := `WEBSOCKET ws://localhost:8080
Authorization: Bearer token123
X-Custom-Header: value

### Message
> test
<
`
	tmpFile := createTempWSFile(t, content)
	defer os.Remove(tmpFile)

	result, err := ParseWebSocketFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseWebSocketFile failed: %v", err)
	}

	if len(result.Headers) != 2 {
		t.Fatalf("Expected 2 headers, got %d", len(result.Headers))
	}

	if result.Headers["Authorization"] != "Bearer token123" {
		t.Errorf("Authorization header mismatch: %s", result.Headers["Authorization"])
	}
	if result.Headers["X-Custom-Header"] != "value" {
		t.Errorf("Custom header mismatch: %s", result.Headers["X-Custom-Header"])
	}
}

func TestParseWebSocketFile_JSONMessage(t *testing.T) {
	content := `WEBSOCKET ws://localhost:8080

### JSON Test
# @type json
# @timeout 60
> json
{
  "action": "ping",
  "data": {
    "message": "test"
  }
}
<
`
	tmpFile := createTempWSFile(t, content)
	defer os.Remove(tmpFile)

	result, err := ParseWebSocketFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseWebSocketFile failed: %v", err)
	}

	// Should create 2 messages: send JSON + receive
	if len(result.Messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(result.Messages))
	}

	// First message: send JSON
	sendMsg := result.Messages[0]
	if sendMsg.Type != "json" {
		t.Errorf("Expected type 'json', got '%s'", sendMsg.Type)
	}
	if sendMsg.Direction != "send" {
		t.Errorf("Expected direction 'send', got '%s'", sendMsg.Direction)
	}
	if sendMsg.Timeout != 60 {
		t.Errorf("Expected timeout 60, got %d", sendMsg.Timeout)
	}

	expectedContent := `{
  "action": "ping",
  "data": {
    "message": "test"
  }
}`
	if sendMsg.Content != expectedContent {
		t.Errorf("JSON content mismatch.\nExpected:\n%s\n\nGot:\n%s", expectedContent, sendMsg.Content)
	}

	// Second message: receive
	recvMsg := result.Messages[1]
	if recvMsg.Direction != "receive" {
		t.Errorf("Expected direction 'receive', got '%s'", recvMsg.Direction)
	}
}

func TestParseWebSocketFile_MultipleMessages(t *testing.T) {
	content := `WEBSOCKET ws://localhost:8080

### First Message
> Hello
<

### Second Message
# @type json
> {"action": "ping"}
<

### Third Message
> Goodbye
<
`
	tmpFile := createTempWSFile(t, content)
	defer os.Remove(tmpFile)

	result, err := ParseWebSocketFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseWebSocketFile failed: %v", err)
	}

	// 3 blocks * 2 messages each (send + receive) = 6 messages
	// But also need to account for the ### separator behavior
	// Let's check actual count
	if len(result.Messages) < 6 {
		t.Fatalf("Expected at least 6 messages, got %d", len(result.Messages))
	}

	// Find messages by content to verify parsing
	foundHello := false
	foundPing := false
	foundGoodbye := false

	for _, msg := range result.Messages {
		if msg.Content == "Hello" && msg.Direction == "send" {
			foundHello = true
			if msg.Name != "First Message" {
				t.Errorf("Hello message name mismatch: %s", msg.Name)
			}
		}
		if msg.Content == "{\"action\": \"ping\"}" && msg.Direction == "send" {
			foundPing = true
			if msg.Name != "Second Message" {
				t.Errorf("Ping message name mismatch: %s", msg.Name)
			}
			if msg.Type != "json" {
				t.Errorf("Ping message type should be json, got: %s", msg.Type)
			}
		}
		if msg.Content == "Goodbye" && msg.Direction == "send" {
			foundGoodbye = true
			if msg.Name != "Third Message" {
				t.Errorf("Goodbye message name mismatch: %s", msg.Name)
			}
		}
	}

	if !foundHello {
		t.Error("Did not find 'Hello' send message")
	}
	if !foundPing {
		t.Error("Did not find ping JSON send message")
	}
	if !foundGoodbye {
		t.Error("Did not find 'Goodbye' send message")
	}
}

func TestParseWebSocketFile_ReceiveMessage(t *testing.T) {
	content := `WEBSOCKET ws://localhost:8080

### Send and Receive
> Hello
<

### Just Receive
<
`
	tmpFile := createTempWSFile(t, content)
	defer os.Remove(tmpFile)

	result, err := ParseWebSocketFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseWebSocketFile failed: %v", err)
	}

	// First block: send + receive = 2 messages
	// Second block: receive only = 1 message
	// Total = 3 messages minimum
	if len(result.Messages) < 3 {
		t.Fatalf("Expected at least 3 messages, got %d", len(result.Messages))
	}

	// Check we have send and receive messages
	hasSend := false
	receiveCount := 0

	for _, msg := range result.Messages {
		if msg.Direction == "send" {
			hasSend = true
			if msg.Content != "Hello" {
				t.Errorf("Send message content should be 'Hello', got: %s", msg.Content)
			}
		}
		if msg.Direction == "receive" {
			receiveCount++
		}
	}

	if !hasSend {
		t.Error("Should have at least one send message")
	}
	if receiveCount < 2 {
		t.Errorf("Should have at least 2 receive messages, got %d", receiveCount)
	}
}

func TestParseWebSocketFile_Subprotocols(t *testing.T) {
	content := `WEBSOCKET ws://localhost:8080
# @subprotocol chat
# @subprotocol v2

### Test
> test
<
`
	tmpFile := createTempWSFile(t, content)
	defer os.Remove(tmpFile)

	result, err := ParseWebSocketFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseWebSocketFile failed: %v", err)
	}

	if len(result.Subprotocols) != 2 {
		t.Fatalf("Expected 2 subprotocols, got %d", len(result.Subprotocols))
	}

	if result.Subprotocols[0] != "chat" {
		t.Errorf("First subprotocol should be 'chat', got: %s", result.Subprotocols[0])
	}
	if result.Subprotocols[1] != "v2" {
		t.Errorf("Second subprotocol should be 'v2', got: %s", result.Subprotocols[1])
	}
}

func TestParseWebSocketFile_TLSConfig(t *testing.T) {
	content := `WEBSOCKET wss://secure.example.com
# @tls.certFile /path/to/cert.pem
# @tls.keyFile /path/to/key.pem
# @tls.caFile /path/to/ca.pem
# @tls.insecureSkipVerify true

### Test
> test
<
`
	tmpFile := createTempWSFile(t, content)
	defer os.Remove(tmpFile)

	result, err := ParseWebSocketFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseWebSocketFile failed: %v", err)
	}

	if result.TLS == nil {
		t.Fatal("Expected TLS config, got nil")
	}

	if result.TLS.CertFile != "/path/to/cert.pem" {
		t.Errorf("CertFile mismatch: %s", result.TLS.CertFile)
	}
	if result.TLS.KeyFile != "/path/to/key.pem" {
		t.Errorf("KeyFile mismatch: %s", result.TLS.KeyFile)
	}
	if result.TLS.CAFile != "/path/to/ca.pem" {
		t.Errorf("CAFile mismatch: %s", result.TLS.CAFile)
	}
	if !result.TLS.InsecureSkipVerify {
		t.Error("InsecureSkipVerify should be true")
	}
}

func TestParseWebSocketFile_MissingURL(t *testing.T) {
	content := `### Message
> test
<
`
	tmpFile := createTempWSFile(t, content)
	defer os.Remove(tmpFile)

	_, err := ParseWebSocketFile(tmpFile)
	if err == nil {
		t.Fatal("Expected error for missing URL, got nil")
	}

	expectedError := "no WEBSOCKET url found in file"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestParseWebSocketFile_EmptyFile(t *testing.T) {
	content := ``
	tmpFile := createTempWSFile(t, content)
	defer os.Remove(tmpFile)

	_, err := ParseWebSocketFile(tmpFile)
	if err == nil {
		t.Fatal("Expected error for empty file, got nil")
	}
}

func TestParseWebSocketFile_InlineAndMultilineContent(t *testing.T) {
	content := `WEBSOCKET ws://localhost:8080

### Inline
> Inline content
<

### Multiline
>
Line 1
Line 2
Line 3
<
`
	tmpFile := createTempWSFile(t, content)
	defer os.Remove(tmpFile)

	result, err := ParseWebSocketFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseWebSocketFile failed: %v", err)
	}

	// Each block creates send + receive messages
	if len(result.Messages) < 4 {
		t.Fatalf("Expected at least 4 messages, got %d", len(result.Messages))
	}

	// Find the inline send message
	foundInline := false
	foundMultiline := false

	expectedMultiline := `Line 1
Line 2
Line 3`

	for _, msg := range result.Messages {
		if msg.Direction == "send" && msg.Content == "Inline content" {
			foundInline = true
		}
		if msg.Direction == "send" && msg.Content == expectedMultiline {
			foundMultiline = true
		}
	}

	if !foundInline {
		t.Error("Did not find inline content send message")
	}
	if !foundMultiline {
		t.Errorf("Did not find multiline content send message")
	}
}

func TestParseWebSocketFile_DefaultName(t *testing.T) {
	content := `WEBSOCKET ws://localhost:8080

### Test
> test
<
`
	tmpFile := createTempWSFile(t, content)
	defer os.Remove(tmpFile)

	result, err := ParseWebSocketFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseWebSocketFile failed: %v", err)
	}

	if result.Name != "WebSocket Connection" {
		t.Errorf("Expected default name 'WebSocket Connection', got '%s'", result.Name)
	}
}

// Helper function to create temporary .ws files for testing
func createTempWSFile(t *testing.T, content string) string {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.ws")

	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	return tmpFile
}
