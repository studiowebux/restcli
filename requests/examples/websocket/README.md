# WebSocket Examples

Example WebSocket configurations demonstrating various features and use cases.

## Available Examples

### Basic Echo Server
**File:** `local-echo.ws`
**Server:** `server/websocket/echo-server.ts`

Simple echo server that reflects back all messages sent to it. Good for basic WebSocket testing.

```bash
# Start server
deno run --allow-net server/websocket/echo-server.ts

# Run in restcli
restcli requests/examples/websocket/local-echo.ws
```

### Authentication
**File:** `auth-example.ws`
**Server:** `server/websocket/auth-server.ts`

Demonstrates WebSocket connections with Bearer token authentication. The server validates the Authorization header before accepting the connection.

Valid token: `demo-token-12345`

```bash
# Start server
deno run --allow-net server/websocket/auth-server.ts

# Run in restcli
restcli requests/examples/websocket/auth-example.ws
```

Features:
- Bearer token authentication
- Connection rejection for invalid tokens
- Authenticated message echoing

### Subprotocol: Chat
**File:** `subprotocol-chat.ws`
**Server:** `server/websocket/subprotocol-server.ts`

Demonstrates the "chat" subprotocol with structured message types.

```bash
# Start server
deno run --allow-net server/websocket/subprotocol-server.ts

# Run in restcli
restcli requests/examples/websocket/subprotocol-chat.ws
```

Message format:
```json
{
  "type": "message",
  "text": "Your message here"
}
```

Supported message types:
- `message` - Chat messages
- `ping` - Ping/pong for keepalive

### Subprotocol: JSON-RPC
**File:** `subprotocol-jsonrpc.ws`
**Server:** `server/websocket/subprotocol-server.ts`

Demonstrates JSON-RPC 2.0 protocol over WebSocket.

```bash
# Start server
deno run --allow-net server/websocket/subprotocol-server.ts

# Run in restcli
restcli requests/examples/websocket/subprotocol-jsonrpc.ws
```

Supported methods:
- `echo` - Echoes back the parameters
- `ping` - Returns current timestamp
- `time` - Returns current server time

Request format:
```json
{
  "jsonrpc": "2.0",
  "method": "echo",
  "params": {"message": "test"},
  "id": 1
}
```

## Running Examples

All WebSocket examples can be executed in restcli's TUI mode:

```bash
# Interactive TUI
restcli

# Direct execution
restcli requests/examples/websocket/local-echo.ws
```

## Features Demonstrated

- Basic send/receive messaging
- Authentication with headers
- Subprotocol negotiation
- Structured message protocols
- Error handling
- Connection management
