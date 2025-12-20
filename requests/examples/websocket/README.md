# WebSocket Examples

This directory contains example WebSocket connection files (`.ws`) for testing REST CLI's WebSocket client functionality.

## Prerequisites

Deno runtime for running the local echo server.

## Quick Start

1. Start the echo server:
```bash
deno run --allow-net server/websocket/echo-server.ts
```

2. Open REST CLI TUI:
```bash
restcli
```

3. Navigate to `requests/examples/websocket/local-echo.ws`

4. Execute the WebSocket connection

## File Format (.ws)

WebSocket files use a simple text format:

```
WEBSOCKET ws://localhost:8080

### Message Name
# @type text|json|binary
# @timeout 30
> message content
<
```

### Syntax Elements

**Connection Header**
```
WEBSOCKET <url>
```
Defines the WebSocket endpoint to connect to. Supports `ws://` and `wss://` schemes.

**Headers** (optional)
```
Authorization: Bearer token123
Custom-Header: value
```
Connection-level headers sent during WebSocket handshake.

**Annotations** (optional)

Connection-level:
- `# @subprotocol <name>` - Request specific subprotocol
- `# @tls.certFile <path>` - Client certificate for mTLS
- `# @tls.keyFile <path>` - Private key for mTLS
- `# @tls.caFile <path>` - Custom CA certificate
- `# @tls.insecureSkipVerify true|false` - Skip TLS verification

Message-level:
- `# @type text|json|binary` - Message type (default: text)
- `# @timeout <seconds>` - Receive timeout (default: 30)

**Message Separator**
```
### Message Name
```
Starts a new message definition with an optional descriptive name.

**Send Direction**
```
> message content
```
Indicates a message to send. Content can be inline or on following lines.

For JSON messages:
```
> json
{
  "key": "value"
}
```

**Receive Direction**
```
<
```
Indicates expecting a message from the server. Waits up to `@timeout` seconds.

### Message Flow

Messages are executed sequentially:
1. `> ...` - Send message to server
2. `<` - Wait for response from server
3. Repeat

Connection closes automatically after all messages are processed.

## Examples

### local-echo.ws

Simple echo server test demonstrating:
- Text messages
- JSON messages
- Multiple message exchanges
- Automatic echoing

### Usage Patterns

**Simple Text Echo**
```
WEBSOCKET ws://localhost:8080

### Greeting
> Hello, World!
<
```

**JSON Request/Response**
```
WEBSOCKET ws://localhost:8080

### JSON Echo
# @type json
> json
{
  "action": "ping",
  "timestamp": "2025-12-19T10:00:00Z"
}
<
```

**Authenticated Connection**
```
WEBSOCKET wss://api.example.com/ws
Authorization: Bearer {{authToken}}

### Subscribe
> json
{"action": "subscribe", "channel": "updates"}
<
```

**mTLS Connection**
```
WEBSOCKET wss://secure.example.com/ws
# @tls.certFile certs/client.crt
# @tls.keyFile certs/client.key
# @tls.caFile certs/ca.crt

### Secure Handshake
> Hello
<
```

**Custom Timeout**
```
WEBSOCKET ws://slow-server.com

### Long Wait
# @timeout 120
> Request data
<
```

## Server Setup

The included `echo-server.ts` provides a simple WebSocket server for testing:

```bash
# Default port 8080
deno run --allow-net server/websocket/echo-server.ts

# Custom port
deno run --allow-net server/websocket/echo-server.ts 9000
```

Features:
- Echoes all received messages
- Logs connection events
- Handles multiple concurrent connections
- Graceful shutdown on Ctrl+C

## Notes

Phase 1 (Current):
- Parser and executor foundation complete
- Example files created
- Local echo server available

Phase 2 (TUI Integration):
- Split-pane modal for live messaging
- Real-time message history
- Interactive message composition
- Connection status display

WebSocket support is completely isolated from HTTP/REST functionality. `.ws` files are recognized but require Phase 2 TUI integration for execution.
