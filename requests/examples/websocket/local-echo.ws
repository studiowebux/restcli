# Local WebSocket Echo Server Example
#
# This example connects to the local echo server and exchanges messages.
#
# Before running:
#   1. Start the echo server: deno run --allow-net server/websocket/echo-server.ts
#   2. Execute this request in REST CLI
#
# Expected behavior:
#   - Connects to ws://localhost:8080
#   - Sends a text greeting
#   - Receives the echo response
#   - Sends a JSON message
#   - Receives the JSON echo
#   - Disconnects gracefully

WEBSOCKET ws://localhost:8080

### Text Greeting
# @type text
# @timeout 5
> Hello from REST CLI!
<

### JSON Echo Test
# @type json
# @timeout 5
> json
{
  "action": "ping",
  "timestamp": "2025-12-19T10:00:00Z",
  "data": {
    "message": "Testing WebSocket JSON support",
    "client": "REST CLI"
  }
}
<

### Another Text Message
# @type text
> This is a second message to test multiple exchanges
<
