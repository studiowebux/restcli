---
title: TUI Mode
tags:
  - guide
---

# TUI Mode

Keyboard-driven interface for testing HTTP endpoints.

## Start TUI

```bash
restcli
```

## Panel System

Two panels: sidebar (file list) and response viewer.

Green border shows focused panel.

Switch focus: `TAB`

Each panel scrolls independently.

## Navigation

### Basic

| Key        | Action          |
| ---------- | --------------- |
| `j` / Down | Move down       |
| `k` / Up   | Move up         |
| `Enter`    | Execute request |
| `TAB`      | Switch panel    |

### Vim-style

| Key      | Action         |
| -------- | -------------- |
| `gg`     | Jump to top    |
| `G`      | Jump to bottom |
| `Ctrl+U` | Half page up   |
| `Ctrl+D` | Half page down |

### Direct Jump

`:` followed by hex line number to jump directly.

Example: `:1A` jumps to line 26.

## File Operations

| Key      | Action                   |
| -------- | ------------------------ |
| `Enter`  | Execute request          |
| `Esc`    | Cancel running request   |
| `i`      | Inspect request          |
| `x`      | Edit in external editor  |
| `X`      | Edit in inline editor    |
| `d`      | Duplicate file           |
| `D`      | Delete file              |
| `F`      | Create new file          |
| `R`      | Rename file              |
| `r`      | Refresh file list        |
| `Ctrl+P` | MRU (most recently used) |

**Request Cancellation**: Press `Esc` while a request is in progress to cancel it. Works for both regular and streaming requests.

### Creating Files

Press `F` to create a new file.

Select format: `.http`, `.yaml`, `.json`, `.jsonc`

Enter filename (extension added automatically).

## Search

| Key      | Action           |
| -------- | ---------------- |
| `/`      | Start search     |
| `n`      | Next match       |
| `N`      | Previous match   |
| `Ctrl+R` | Alternative next |

Search is context-aware:

1. Sidebar focused: searches filenames
2. Response focused: searches response body

**Search behavior:**

- Text search is **case-insensitive**
- Regex patterns are **case-sensitive** by default
- Use `(?i)pattern` for case-insensitive regex matching

Example: `(?i)error` matches "Error", "ERROR", "error"

## Response Operations

| Key | Action                    |
| --- | ------------------------- |
| `s` | Save to file              |
| `c` | Copy to clipboard         |
| `b` | Toggle body visibility    |
| `B` | Toggle headers visibility |
| `f` | Fullscreen mode           |
| `w` | Pin response              |
| `W` | Diff with pinned          |

### Pinning and Diff

Pin current response with `w`.

Execute another request.

Press `W` to see diff between pinned and current.

Useful for API regression testing.

## Modals and Editors

| Key | Action               |
| --- | -------------------- |
| `v` | Variable editor      |
| `h` | Header editor        |
| `p` | Profile switcher     |
| `m` | Documentation viewer |
| `H` | History viewer       |
| `C` | Configuration viewer |
| `?` | Help                 |

### Variable Editor

Press `v` to open.

Multi-value variable support:

| Key | Action                   |
| --- | ------------------------ |
| `m` | Modify selected variable |
| `s` | Set active option        |
| `a` | Add new variable         |
| `e` | Edit variable            |
| `d` | Delete variable          |
| `l` | List all values          |
| `L` | Set value by alias       |

### Documentation Viewer

Press `m` to view embedded request documentation.

Tree navigation:

1. Expand/collapse sections
2. View parameter schemas
3. See response examples
4. Read field descriptions

### History Viewer

Press `H` to view request history.

- Left pane: List of all history entries with timestamps, methods, URLs, and status codes
- Right pane: Live preview of the selected entry's response

Navigate with `j`/`k` keys - the preview updates automatically as you move through the list.

Actions:
- `Enter`: Load selected response into the main view
- `r`: Replay the request (re-execute)
- `p`: Toggle preview pane visibility (hide sensitive data)
- `C`: Clear all history

**Privacy**: Use `p` to hide the response preview pane when viewing sensitive data or secrets.

History persists across sessions.

## Profile Switching

Press `p` to switch profiles.

Select from list.

Press `e` on a profile to edit.

Session clears when switching profiles.

## OAuth Flow

Press `o` to start OAuth 2.0 flow.

Press `O` to configure OAuth settings.

## Configuration

Press `C` to view current configuration:

1. Active profile
2. Variables
3. Headers
4. TLS settings
5. Filters and queries

Press `P` to view profile config.

Press `S` to view session config.

## Text Input

In modal text inputs:

| Key      | Action               |
| -------- | -------------------- |
| `Ctrl+V` | Paste from clipboard |
| `Ctrl+K` | Clear input          |

## Error Details

When an error occurs, a truncated message appears in the footer.

Press `e` to view the full error message in a scrollable modal.

The error modal supports:
- Text wrapping for long error messages
- `j`/`k` scrolling
- `g`/`G` for top/bottom navigation

## Shortcuts

Press `?` for complete list.

Search within help: `/` to search, `n`/`N` to navigate.

## Exit

Press `q` to quit.
