---
title: WebSocket
tags:
  - guide
  - websocket
  - real-time
---

# WebSocket Support

Interactive WebSocket client with dual-pane TUI, message history, and variable resolution.

## Quick Start

```websocket
WEBSOCKET ws://localhost:8080
Authorization: Bearer token123

### Ping
> {"type": "ping"}
<
```

Execute: `restcli file.ws`

## File Format

### Connection

```websocket
WEBSOCKET url
Header-Name: value
# @subprotocol protocol-name
```

Connection-level configuration. Headers sent during WebSocket handshake. Subprotocol negotiates application protocol.

### Messages

```websocket
### Message Name
# @type text|json|binary
# @timeout 30
> content
<
```

- `###` separates messages
- `>` sends message
- `<` receives/waits for response  
- Type: text (default), json, binary
- Timeout in seconds (optional)

### TLS Configuration

```websocket
# @tls.certFile path/to/cert.pem
# @tls.keyFile path/to/key.pem
# @tls.caFile path/to/ca.pem
# @tls.insecureSkipVerify true
```

Client certificate authentication for wss:// connections. Inherits from profile if not specified.

## TUI Mode

Interactive dual-pane interface:

**Left:** Message history with timestamps, direction indicators  
**Right:** Predefined messages from .ws file

Real-time updates. Connection status with color indicators:
- Green (●) Connected
- Yellow (◐) Connecting
- Gray (○) Disconnected
- Red (✖) Error

## Keyboard Shortcuts

### Navigation
- `j`/`k`, `↑`/`↓` - Navigate messages or scroll history
- `Tab` - Switch focus between panes
- `gg` - Jump to top
- `G` - Jump to bottom
- `Ctrl+d` - Page down
- `Ctrl+u` - Page up

### Actions
- `Enter` - Send selected message
- `r` - Connect/reconnect
- `d` - Disconnect
- `i` - Compose custom message (when connected)
- `c` - Copy last message to clipboard
- `C` - Clear message history (with confirmation)
- `e` - Export history to JSON
- `/` - Search messages
- `q`, `Esc` - Close WebSocket modal

### Search Mode
- Type to filter messages
- `Enter` - Keep filter active
- `Esc` - Clear filter

### Composer Mode
- Type custom message
- `Enter` - Send
- `Esc` - Cancel

## Variable Resolution

Messages support variable substitution from profiles, sessions, and environment:

```websocket
### Auth Message
> {"token": "{{auth_token}}", "user": "{{username}}"}
<
```

Resolution occurs at send time. Errors shown as system messages.

## Profile Integration

### Headers

Profile headers merge with .ws file headers. File headers override profile headers.

```yaml
# profile.yaml
headers:
  Authorization: Bearer default-token
```

```websocket
# File headers override profile
WEBSOCKET wss://api.example.com
Authorization: Bearer {{api_key}}
```

### TLS

```websocket
# @tls.certFile client.pem
# @tls.keyFile client-key.pem
# @tls.caFile ca.pem
```

Profile TLS configuration inherited if not specified in .ws file.

## Message Types

### Text

```websocket
### Simple Text
> Hello, server!
<
```

Default type. Sent as WebSocket text frame.

### JSON

```websocket
### JSON Data
# @type json
> {"action": "subscribe", "channel": "updates"}
<
```

Validated and formatted in message history.

### Binary

```websocket
### Binary Data
# @type binary
> base64-encoded-data
<
```

Sent as WebSocket binary frame.

## Subprotocols

Negotiate application protocols during handshake:

```websocket
WEBSOCKET ws://localhost:8082
# @subprotocol chat

### Chat Message
> {"type": "message", "text": "Hello"}
<
```

Server must support requested subprotocol. Common protocols: chat, json-rpc, graphql-ws.

## Advanced Features

### Message Search

Press `/` to filter message history:
- Case-insensitive search
- Searches content, direction, type
- `Enter` keeps filter
- `Esc` clears filter

### Message Export

Press `e` to export message history:
- JSON format with timestamps
- Filename: `websocket-messages-YYYYMMDD-HHMMSS.json`
- Includes all sent/received/system messages
- Preserves direction and type

### Persistent Connections

Single connection for multiple messages:
- Bidirectional real-time communication
- Reconnect with `r` key
- Clean disconnect with `d` key

### Timeout

```websocket
### Long Operation
# @timeout 60
> {"action": "process"}
<
```

Per-message timeout in seconds. Default: 30s.

## Examples

### Echo Server

```websocket
WEBSOCKET ws://localhost:8080

### Echo Test
> Hello, WebSocket!
<

### JSON Echo
> {"message": "test"}
<
```

### Authenticated Connection

```websocket
WEBSOCKET wss://api.example.com
Authorization: Bearer {{access_token}}

### Subscribe
> {"action": "subscribe", "topic": "notifications"}
<
```

### Chat Protocol

```websocket
WEBSOCKET ws://localhost:8082
# @subprotocol chat

### Join
> {"type": "join", "user": "{{username}}"}
<

### Send Message
> {"type": "message", "text": "Hello"}
<
```

### JSON-RPC

```websocket
WEBSOCKET ws://localhost:8082
# @subprotocol json-rpc

### Echo Method
> {"jsonrpc": "2.0", "method": "echo", "params": {"msg": "test"}, "id": 1}
<

### Ping
> {"jsonrpc": "2.0", "method": "ping", "id": 2}
<
```

## See Also

- [File Formats](file-formats.md)
- [Keyboard Shortcuts](../reference/keyboard-shortcuts.md)
- [Variables](variables.md)
- [Profiles](profiles.md)
