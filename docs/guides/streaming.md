---
title: Streaming Responses
tags:
  - guide
---

# Streaming Responses

REST CLI automatically detects and handles streaming HTTP responses.

## Auto-Detection

Streaming is detected based on response headers:

- `Content-Type: text/event-stream` (Server-Sent Events)
- `Content-Type: application/stream+json` (JSON streaming)
- `Transfer-Encoding: chunked` (Chunked transfer encoding)

No configuration required - it just works!

## Supported Streaming Types

### Server-Sent Events (SSE)

SSE streams require the `@streaming true` directive for real-time display:

```text
### Stream Events
# @streaming true
GET https://api.example.com/events
Accept: text/event-stream
```

**IMPORTANT:** Without `# @streaming true`, the request will wait for completion (which never happens for infinite SSE streams!).

Response:

```
Content-Type: text/event-stream

data: {"event": "user_joined", "user": "alice"}

data: {"event": "message", "text": "Hello!"}

data: {"event": "user_left", "user": "bob"}
```

### Chunked Transfer Encoding

HTTP/1.1 chunked responses are handled transparently:

```text
### Large Dataset
GET https://api.example.com/data/export
```

Server responds with `Transfer-Encoding: chunked`, data arrives progressively.

### JSON Streaming

Newline-delimited JSON streams:

```text
### Stream JSON
POST https://api.openai.com/v1/chat/completions
Authorization: Bearer {{apiKey}}
Content-Type: application/json

{
  "model": "gpt-4",
  "stream": true,
  "messages": [{"role": "user", "content": "Hello"}]
}
```

Response arrives as progressive chunks.

## Behavior

### TUI Mode

**Real-Time Streaming:**
- **Progressive display** - data appears in real-time as it arrives
- **Auto-scroll** - automatically scrolls to show latest data
- **Status indicator** - shows "Streaming..." with stop instructions
- **Cancelation** - press `q` to stop streaming at any time
- **Infinite streams** - no timeout, streams run until stopped or completed
- Auto-detection happens behind the scenes
- Works with all filtering and parsing features

**How it works:**
1. Execute a streaming request (SSE, chunked transfer, etc.)
2. Data appears on screen immediately as chunks arrive
3. Viewport auto-scrolls to show the latest content
4. Press `q` to stop the stream whenever you want
5. Stream completes with "Stream completed" status

### CLI Mode

**Current Behavior:**
- Waits for complete response
- Auto-detection ensures proper handling
- Full response printed when done

**Output:**
```bash
restcli stream-events.http
# Waits for stream to complete, then shows full response
```

## Enabling Streaming

### The `@streaming` Directive

Use `# @streaming true` to enable real-time streaming display:

```text
### My SSE Endpoint
# @streaming true
GET https://example.com/events
```

**When to use:**
- **Infinite streams**: SSE, WebSocket upgrades, long-polling
- **Real-time data**: Live logs, metrics, notifications
- **When you need immediate feedback**: Don't want to wait for completion

**When NOT to use:**
- Regular API requests
- Finite streams that complete quickly
- When you need filtering/querying (not yet supported for streaming)

## Technical Details

### Detection Logic

Auto-detection happens for non-streaming requests:

```go
// For requests WITHOUT @streaming directive:
Content-Type: text/event-stream           → Buffered
Content-Type: application/stream+json     → Buffered
Transfer-Encoding: chunked                → Buffered

// For requests WITH @streaming true:
All responses                             → Real-time display
```

### Implementation

- Uses `bufio.Reader` for efficient chunk reading
- 4KB buffer size for optimal performance
- **Asynchronous processing** - request runs in goroutine
- **Real-time updates** - chunks sent via channel to UI
- **Context-based cancellation** - clean shutdown on user request
- **Auto-scroll** - viewport scrolls to bottom on each update
- No timeout for SSE streams - runs indefinitely until stopped

## Examples

### OpenAI Streaming

```text
### Chat with Streaming
POST https://api.openai.com/v1/chat/completions
Authorization: Bearer {{apiKey}}
Content-Type: application/json

{
  "model": "gpt-4",
  "stream": true,
  "messages": [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "{{userPrompt}}"}
  ]
}
```

Variable `apiKey` can be marked as [interactive](./variables.md#interactive-variables) to prompt securely.

### SSE Endpoint

```text
### Watch Events
# @streaming true
GET https://your-api.com/events/stream
Authorization: Bearer {{token}}
Accept: text/event-stream
```

With `@streaming true`, events appear in real-time as they arrive. Press `q` to stop.

### Large File Download

```text
### Download Report
GET https://api.example.com/reports/2024/export.csv
```

If server uses chunked encoding, handles efficiently.

## Limitations

- **CLI Mode:** Currently waits for completion (TUI has real-time streaming)
- **No manual toggle:** Always auto-detects (cannot force on/off)
- **Memory:** Very long streams accumulate in memory - stop periodically if needed

## Best Practices

1. **Infinite Streams:** Use `q` to stop SSE streams when you have enough data
2. **Memory Management:** For very long streams, stop and restart periodically
3. **Testing:** Test streaming endpoints - they work like curl now!
4. **LLM APIs:** Perfect for OpenAI, Anthropic, and similar streaming APIs with real-time token display

## Keyboard Controls

When streaming is active:
- **`q`** - Stop the stream immediately
- **`j`/`k` or arrow keys** - Scroll through streamed data (auto-scroll pauses)
- **`G`** - Jump to bottom (resume auto-scroll)
- **`g`** - Jump to top

## Related

- [Variables](./variables.md) - Use interactive variables for prompts
- [Filtering](./filtering.md) - Apply JMESPath to streaming responses
- [Profiles](./profiles.md) - Store API keys and configuration
