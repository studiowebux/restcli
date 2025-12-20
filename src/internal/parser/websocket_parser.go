package parser

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/studiowebux/restcli/internal/types"
)

// ParseWebSocketFile parses a .ws file with WebSocket connection and message definitions
func ParseWebSocketFile(filePath string) (*types.WebSocketRequest, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	request := &types.WebSocketRequest{
		Headers:  make(map[string]string),
		Messages: []types.WebSocketMessage{},
	}

	var currentMessage *types.WebSocketMessage
	var messageContent []string
	inMessageBody := false

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			if inMessageBody {
				// Empty line in body - include it
				messageContent = append(messageContent, line)
			}
			continue
		}

		// Connection header: WEBSOCKET url
		if strings.HasPrefix(strings.ToUpper(line), "WEBSOCKET ") {
			parts := strings.SplitN(line, " ", 2)
			if len(parts) == 2 {
				request.URL = strings.TrimSpace(parts[1])
			}
			continue
		}

		// New message separator
		if strings.HasPrefix(line, "###") {
			// Save previous message if exists
			if currentMessage != nil {
				if len(messageContent) > 0 {
					currentMessage.Content = strings.Join(messageContent, "\n")
				}
				request.Messages = append(request.Messages, *currentMessage)
			}

			// Start new message
			currentMessage = &types.WebSocketMessage{
				Name:    strings.TrimSpace(strings.TrimPrefix(line, "###")),
				Type:    "text", // Default type
				Timeout: 30,     // Default 30 second timeout
			}
			messageContent = []string{}
			inMessageBody = false
			continue
		}

		// Annotations (must come before message content)
		if strings.HasPrefix(line, "#") && !inMessageBody {
			trimmed := strings.TrimSpace(strings.TrimPrefix(line, "#"))

			// Connection-level annotations
			if currentMessage == nil {
				// @subprotocol annotation
				if strings.HasPrefix(trimmed, "@subprotocol ") {
					subprotocol := strings.TrimSpace(strings.TrimPrefix(trimmed, "@subprotocol"))
					request.Subprotocols = append(request.Subprotocols, subprotocol)
					continue
				}

				// @tls annotations
				if strings.HasPrefix(trimmed, "@tls.") {
					if request.TLS == nil {
						request.TLS = &types.TLSConfig{}
					}
					if strings.HasPrefix(trimmed, "@tls.certFile ") {
						request.TLS.CertFile = strings.TrimSpace(strings.TrimPrefix(trimmed, "@tls.certFile"))
						continue
					}
					if strings.HasPrefix(trimmed, "@tls.keyFile ") {
						request.TLS.KeyFile = strings.TrimSpace(strings.TrimPrefix(trimmed, "@tls.keyFile"))
						continue
					}
					if strings.HasPrefix(trimmed, "@tls.caFile ") {
						request.TLS.CAFile = strings.TrimSpace(strings.TrimPrefix(trimmed, "@tls.caFile"))
						continue
					}
					if strings.HasPrefix(trimmed, "@tls.insecureSkipVerify ") {
						value := strings.TrimSpace(strings.TrimPrefix(trimmed, "@tls.insecureSkipVerify"))
						request.TLS.InsecureSkipVerify = value == "true"
						continue
					}
				}
				continue
			}

			// Message-level annotations
			if strings.HasPrefix(trimmed, "@type ") {
				currentMessage.Type = strings.TrimSpace(strings.TrimPrefix(trimmed, "@type"))
				continue
			}
			if strings.HasPrefix(trimmed, "@timeout ") {
				value := strings.TrimSpace(strings.TrimPrefix(trimmed, "@timeout"))
				if timeout, err := strconv.Atoi(value); err == nil {
					currentMessage.Timeout = timeout
				}
				continue
			}
			continue
		}

		// Header: Key: Value (only at connection level, not in messages)
		if currentMessage == nil && strings.Contains(line, ":") && !inMessageBody {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				request.Headers[key] = value
			}
			continue
		}

		// Message direction and content
		if currentMessage != nil {
			trimmedLine := strings.TrimSpace(line)

			// Send message: > content or > json or > text
			if strings.HasPrefix(trimmedLine, ">") {
				// Save previous message if direction is changing
				if currentMessage.Direction == "receive" && len(messageContent) > 0 {
					currentMessage.Content = strings.Join(messageContent, "\n")
					request.Messages = append(request.Messages, *currentMessage)
					messageContent = []string{}
				}

				currentMessage.Direction = "send"
				content := strings.TrimSpace(strings.TrimPrefix(trimmedLine, ">"))

				// Check if it's a type declaration: > json or > text
				if content == "json" {
					currentMessage.Type = "json"
					inMessageBody = true
					continue
				}
				if content == "text" {
					currentMessage.Type = "text"
					inMessageBody = true
					continue
				}

				// Inline content: > Hello
				if content != "" {
					messageContent = append(messageContent, content)
					// Save immediately for inline send messages
					currentMessage.Content = strings.Join(messageContent, "\n")
					request.Messages = append(request.Messages, *currentMessage)

					// Reset for next message in same block
					currentMessage = &types.WebSocketMessage{
						Name:    currentMessage.Name,
						Type:    currentMessage.Type,
						Timeout: currentMessage.Timeout,
					}
					messageContent = []string{}
					inMessageBody = false
					continue
				}

				// Next lines are body
				inMessageBody = true
				continue
			}

			// Receive/expect message: < content or < json or < text
			if strings.HasPrefix(trimmedLine, "<") {
				// Save previous message if direction is changing
				if currentMessage.Direction == "send" && len(messageContent) > 0 {
					currentMessage.Content = strings.Join(messageContent, "\n")
					request.Messages = append(request.Messages, *currentMessage)

					// Reset for next message
					currentMessage = &types.WebSocketMessage{
						Name:    currentMessage.Name,
						Type:    currentMessage.Type,
						Timeout: currentMessage.Timeout,
					}
					messageContent = []string{}
					inMessageBody = false
				}

				currentMessage.Direction = "receive"
				content := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "<"))

				// Check if it's a type declaration: < json or < text
				if content == "json" {
					currentMessage.Type = "json"
					inMessageBody = true
					continue
				}
				if content == "text" {
					currentMessage.Type = "text"
					inMessageBody = true
					continue
				}

				// Inline content: < Hello
				if content != "" {
					messageContent = append(messageContent, content)
					continue
				}

				// Empty < means just wait for message - save immediately
				if content == "" {
					request.Messages = append(request.Messages, *currentMessage)

					// Reset for next message
					currentMessage = &types.WebSocketMessage{
						Name:    currentMessage.Name,
						Type:    currentMessage.Type,
						Timeout: currentMessage.Timeout,
					}
					messageContent = []string{}
					inMessageBody = false
					continue
				}

				// Next lines are body
				inMessageBody = true
				continue
			}

			// Body content
			if inMessageBody {
				messageContent = append(messageContent, line)
			}
		}
	}

	// Save last message if it has content or direction set
	if currentMessage != nil && (len(messageContent) > 0 || currentMessage.Direction != "") {
		if len(messageContent) > 0 {
			currentMessage.Content = strings.Join(messageContent, "\n")
		}
		// Only append if it's not a duplicate (inline messages are already saved)
		if currentMessage.Direction != "" {
			request.Messages = append(request.Messages, *currentMessage)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Validation
	if request.URL == "" {
		return nil, fmt.Errorf("no WEBSOCKET url found in file")
	}

	// Set default name if not set
	if request.Name == "" {
		request.Name = "WebSocket Connection"
	}

	return request, nil
}
