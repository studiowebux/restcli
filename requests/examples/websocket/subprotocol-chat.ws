# WebSocket Chat Subprotocol Example
# Start the subprotocol server: deno run --allow-net server/websocket/subprotocol-server.ts
# This example uses the "chat" subprotocol

WEBSOCKET ws://localhost:8082
# @subprotocol chat

### Welcome Message
# Server sends welcome message after connection
<

### Send Chat Message
> {"type": "message", "text": "Hello from chat client"}
<

### Ping Server
> {"type": "ping"}
<

### Another Chat Message
> {"type": "message", "text": "This is the chat protocol"}
<
