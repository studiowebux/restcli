---
title: Request History
tags:
  - guide
---

# Request History

REST CLI automatically tracks request and response history for debugging, auditing, and replaying requests.

## Overview

History tracking captures:

- Request details (method, URL, headers, body)
- Response data (status, headers, body)
- Execution metadata (timestamp, duration, sizes)
- Errors (if any)

## Enabling/Disabling History

### Global Setting

History is enabled by default. To disable globally, edit `~/.restcli/.session.json`:

```json
{
  "historyEnabled": false
}
```

Or toggle in TUI:

1. Press `C` to open configuration viewer
2. Navigate to history setting
3. Press `t` to toggle

### Per-Profile Control

Override global setting per profile in `~/.restcli/.profiles.json`:

```json
{
  "name": "Production",
  "headers": {
    "Authorization": "Bearer {{token}}"
  },
  "historyEnabled": false
}
```

**Priority:**

1. Profile setting (if specified)
2. Global setting (if profile doesn't specify)

**Use cases for disabling:**

- **Production profiles** - Avoid logging sensitive production data
- **High-volume testing** - Prevent history buildup during load tests
- **Sensitive APIs** - Don't persist API keys or PII in history
- **Temporary work** - Exploratory testing without logging

## Viewing History

### TUI Mode

Press `H` to open history viewer.

**History List:**
| Key | Action |
| --- | ------ |
| `↑/↓` or `j/k` | Navigate entries (left pane) |
| `Shift+↑/↓` or `J/K` | Scroll preview pane (right pane) |
| `gg` | Go to first entry |
| `G` | Go to last entry |
| `PgUp/PgDn` | Page up/down |
| `Ctrl+u/d` | Half page up/down |
| `Enter` | Load entry into main view |
| `r` | Replay request |
| `p` | Toggle preview pane visibility |
| `C` | Clear all history (with confirmation) |
| `ESC` or `H` or `q` | Close viewer |

Footer shows scroll position: `[current/total] (percentage%)`

**Details View:**

- Full request (method, URL, headers, body)
- Response (status, headers, body)
- Metadata (timestamp, duration, sizes)

### File Location

History stored in: `~/.restcli/restcli.db` (SQLite database, `history` table)

## Clearing History

### Clear All (TUI)

1. Press `H` to open history
2. Press `D` to clear all
3. Confirm with `y` or cancel with `n`

**Confirmation dialog:**

```
┌─ Confirm ─────────────────────────┐
│ Clear all history?                │
│                                   │
│  This will delete all 127 entries │
│  This action cannot be undone.    │
│                                   │
│  [y] Yes  [n] No                  │
└───────────────────────────────────┘
```

## History Entry Structure

```json
{
  "timestamp": "2024-11-23T14:30:22Z",
  "requestFile": "/path/to/requests.http",
  "requestName": "Get User",
  "method": "GET",
  "url": "https://api.example.com/users/123",
  "headers": {
    "Authorization": "Bearer token123",
    "Content-Type": "application/json"
  },
  "body": "",
  "responseStatus": 200,
  "responseStatusText": "200 OK",
  "responseHeaders": {
    "Content-Type": "application/json"
  },
  "responseBody": "{\"id\":123,\"name\":\"Alice\"}",
  "duration": 245,
  "requestSize": 0,
  "responseSize": 32,
  "error": ""
}
```

## Replaying Requests

### From History

1. Press `H` to open history
2. Navigate to desired entry
3. Press `r` to replay

**Replay behavior:**

- Uses exact same request (method, URL, headers, body)
- Resolves variables at replay time (may differ from original)
- Captures new response in history
- Useful for:
  - Debugging API changes
  - Comparing responses over time
  - Re-running failed requests
