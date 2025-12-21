# WebSocket Authentication Example
# Start the auth server: deno run --allow-net server/websocket/auth-server.ts
# Valid token: demo-token-12345

WEBSOCKET ws://localhost:8081
Authorization: Bearer demo-token-12345

### Welcome Message
# After connecting, server sends auth success confirmation
<

### Send Authenticated Message
> {"type": "test", "message": "Hello from authenticated client"}
<

### JSON Echo Test
> {"action": "ping", "data": "test data"}
<

### Text Message
> Simple text message
<
