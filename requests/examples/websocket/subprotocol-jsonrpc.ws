# WebSocket JSON-RPC Subprotocol Example
# Start the subprotocol server: deno run --allow-net server/websocket/subprotocol-server.ts
# This example uses the "json-rpc" subprotocol

WEBSOCKET ws://localhost:8082
# @subprotocol json-rpc

### Server Ready Notification
# Server sends ready notification after connection
<

### Echo Method
> {"jsonrpc": "2.0", "method": "echo", "params": {"message": "test"}, "id": 1}
<

### Ping Method
> {"jsonrpc": "2.0", "method": "ping", "id": 2}
<

### Time Method
> {"jsonrpc": "2.0", "method": "time", "id": 3}
<

### Invalid Method (Error Response)
> {"jsonrpc": "2.0", "method": "unknown", "id": 4}
<
