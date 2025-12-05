---
title: Analytics
tags:
  - guide
---

# Analytics

Track request performance and statistics per endpoint.

## Overview

Analytics feature provides insights into:

- Request frequency and patterns
- Response times (avg, min, max)
- Success/error rates
- Status code distribution
- Data transfer volumes

## Enabling Analytics

Analytics is opt-in per profile (disabled by default).

### Via Profile Configuration

Edit `.profiles.json`:

```json
{
  "name": "Dev",
  "analyticsEnabled": true,
  "variables": {
    "baseUrl": "https://dev.api.example.com"
  }
}
```

### Via TUI

1. Press `p` to switch profiles
2. Select profile and press `e` to edit
3. Navigate to analytics field
4. Press `SPACE` to toggle

## Viewing Analytics

Press `A` in TUI mode to open analytics viewer.

### Split View

Left pane: List of endpoints with summary stats (scrollable)
Right pane: Detailed breakdown (scrollable)

Scroll indicators (▲/▼) appear when content exceeds viewport height.

```text
┌─────────────────────────┬─────────────────────────┐
│ Analytics            ▲  │ Details              ▲  │
│                         │                         │
│ GET /users | Calls: 150 │ GET /users              │
│ POST /auth | Calls: 45  │                         │
│                      ▼  │ Summary                 │
│                         │ Total Calls:    150     │
│                         │ Success:        148     │
│                         │ Errors:         2       │
│                         │                         │
│                         │ Timing                  │
│                         │ Average:        125ms   │
│                         │ Min:            45ms    │
│                         │ Max:            890ms   │
│                         │                      ▼  │
└─────────────────────────┴─────────────────────────┘
```

### Keyboard Shortcuts

| Key                  | Action                                 |
| -------------------- | -------------------------------------- |
| `↑/↓` or `j/k`       | Navigate entries (left pane)           |
| `Shift+↑/↓` or `J/K` | Scroll detail pane (right pane)        |
| `gg`                 | Go to first entry                      |
| `G`                  | Go to last entry                       |
| `PgUp/PgDn`          | Page up/down                           |
| `Ctrl+u/d`           | Half page up/down                      |
| `Enter`              | Load associated request file           |
| `p`                  | Toggle detail pane visibility          |
| `t`                  | Toggle grouping (per-file <-> by path) |
| `C`                  | Clear all analytics data               |
| `ESC` or `q`         | Close viewer                           |

Footer shows scroll position: `[current/total] (percentage%)`

## Grouping Modes

### Per File (Default)

Stats tracked separately for each request file.

Use when:

- Different files represent different use cases
- You want file-level insights

### By Normalized Path

Aggregates stats across files with same endpoint.

Example: `/users/123` and `/users/456` → grouped as `/users/{id}`

Use when:

- Multiple files hit same endpoint
- Files are renamed or reorganized
- You want endpoint-level insights

Toggle: Press `t` in analytics viewer

## Tracked Metrics

### Per Request

- File path
- Normalized endpoint path
- HTTP method
- Status code
- Request body size (bytes)
- Response body size (bytes)
- Duration (milliseconds)
- Timestamp

### Aggregated Stats

- Total calls
- Success count (2xx status codes)
- Error count (4xx/5xx status codes)
- Average/min/max duration
- Total request/response data transferred
- Status code distribution

## Data Storage

Analytics stored in SQLite database:

```text
~/.restcli/restcli.db
```

**Database Structure:**

- Table: `analytics`
- Shared database with history (separate tables)
- Auto-creates on first use when analytics enabled

## Privacy

- Analytics stored locally only
- No external data transmission
- Can be cleared anytime (`C` in analytics viewer)
- Per-file tracking enables selective history

## Performance

- Minimal overhead (async writes)
- SQLite indexing for fast queries
- Request execution not blocked

## Use Cases

### API Development

- Monitor endpoint performance during development
- Identify slow endpoints
- Track error patterns

### Testing

- Verify request frequency
- Analyze response times
- Monitor data transfer

### Debugging

- Correlate errors with specific endpoints
- Track success rates over time
- Identify performance regressions

## Clearing Analytics

### All Data

In analytics viewer (`A`), press `C` then confirm with `y`.

### Manual Database Reset

```bash
# Delete entire database (removes both analytics AND history)
rm ~/.restcli/restcli.db

# Or use sqlite3 to delete only analytics table
sqlite3 ~/.restcli/restcli.db "DELETE FROM analytics;"
```
