# Keyboard Shortcuts Reference

Complete keyboard shortcuts for TUI mode.

## Navigation

### Basic Movement

| Key            | Action                 |
| -------------- | ---------------------- |
| `j` or `Down`  | Move down one line     |
| `k` or `Up`    | Move up one line       |
| `h` or `Left`  | Move left (in modals)  |
| `l` or `Right` | Move right (in modals) |
| `PageDown`     | Page down              |
| `PageUp`       | Page up                |
| `Home`         | Jump to first item     |
| `End`          | Jump to last item      |

### Vim-style Navigation

| Key      | Action         |
| -------- | -------------- |
| `gg`     | Jump to top    |
| `G`      | Jump to bottom |
| `Ctrl+U` | Half page up   |
| `Ctrl+D` | Half page down |
| `Ctrl+B` | Full page up   |
| `Ctrl+F` | Full page down |

### Direct Jump

| Key | Action                        |
| --- | ----------------------------- |
| `:` | Enter hex line number to jump |

Example: `:1A` jumps to line 26 (hex 1A).

### Panel Focus

| Key   | Action                |
| ----- | --------------------- |
| `TAB` | Switch between panels |

Green border indicates focused panel.

## File Operations

| Key      | Action                        |
| -------- | ----------------------------- |
| `Enter`  | Execute selected request      |
| `i`      | Inspect request details       |
| `x`      | Edit in external editor       |
| `X`      | Edit in inline editor         |
| `d`      | Duplicate file                |
| `D`      | Delete file                   |
| `F`      | Create new file               |
| `R`      | Rename file                   |
| `r`      | Refresh file list             |
| `Ctrl+P` | Open MRU (most recently used) |

## Search

| Key      | Action                           |
| -------- | -------------------------------- |
| `/`      | Start search (files or response) |
| `n`      | Next match                       |
| `N`      | Previous match                   |
| `Ctrl+R` | Alternative next match           |

Search is context-aware based on focused panel and supports Regexes.

## Response Operations

| Key | Action                         |
| --- | ------------------------------ |
| `s` | Save response to file          |
| `c` | Copy response to clipboard     |
| `b` | Toggle body visibility         |
| `B` | Toggle headers visibility      |
| `f` | Fullscreen mode                |
| `w` | Pin current response           |
| `W` | Show diff with pinned response |

## Configuration

| Key | Action               |
| --- | -------------------- |
| `v` | Open variable editor |
| `h` | Open header editor   |
| `p` | Switch profile       |
| `n` | Create new profile   |
| `C` | View configuration   |
| `P` | View profile config  |
| `S` | View session config  |

## Authentication

| Key | Action                   |
| --- | ------------------------ |
| `o` | Start OAuth flow         |
| `O` | Configure OAuth settings |

## Documentation and History

| Key | Action                       |
| --- | ---------------------------- |
| `m` | Open documentation viewer    |
| `H` | Open history viewer          |
| `r` | Replay selected history item |

## Help and Info

| Key | Action                  |
| --- | ----------------------- |
| `?` | Show help and shortcuts |
| `q` | Quit application        |

## Modal Operations

### General Modal Keys

| Key         | Action         |
| ----------- | -------------- |
| `Esc`       | Close modal    |
| `Enter`     | Confirm/Submit |
| `Tab`       | Next field     |
| `Shift+Tab` | Previous field |

### Text Input

| Key         | Action                   |
| ----------- | ------------------------ |
| `Ctrl+V`    | Paste from clipboard     |
| `Ctrl+K`    | Clear input              |
| `Backspace` | Delete character         |
| `Delete`    | Delete character forward |

## Variable Editor

| Key | Action                           |
| --- | -------------------------------- |
| `m` | Modify selected variable         |
| `s` | Set active option (multi-value)  |
| `a` | Add new variable                 |
| `e` | Edit variable                    |
| `d` | Delete variable                  |
| `l` | List all values (multi-value)    |
| `L` | Set value by alias (multi-value) |

## Profile Switcher

| Key     | Action                     |
| ------- | -------------------------- |
| `Enter` | Switch to selected profile |
| `e`     | Edit selected profile      |
| `d`     | Delete profile             |
| `n`     | Create new profile         |

## Documentation Viewer

| Key     | Action                  |
| ------- | ----------------------- |
| `j`/`k` | Navigate tree           |
| `Enter` | Expand/collapse section |
| `Space` | Toggle section          |
| `Esc`   | Close viewer            |

## History Viewer

| Key     | Action                  |
| ------- | ----------------------- |
| `j`/`k` | Navigate history        |
| `r`     | Replay selected request |
| `d`     | Delete history entry    |
| `Esc`   | Close viewer            |

## Diff Viewer

| Key     | Action      |
| ------- | ----------- |
| `j`/`k` | Scroll diff |
| `Esc`   | Close diff  |

## Quick Reference

### Essential Keys

| Key     | Action          |
| ------- | --------------- |
| `Enter` | Execute request |
| `TAB`   | Switch panel    |
| `?`     | Help            |
| `q`     | Quit            |
| `/`     | Search          |
| `v`     | Variables       |
| `h`     | Headers         |
| `p`     | Profiles        |
| `s`     | Save            |
| `o`     | OAuth           |

### File Management

| Key | Action    |
| --- | --------- |
| `i` | Inspect   |
| `x` | Edit      |
| `d` | Duplicate |
| `D` | Delete    |
| `F` | New file  |
| `R` | Rename    |
| `r` | Refresh   |

### Response Actions

| Key | Action         |
| --- | -------------- |
| `s` | Save           |
| `c` | Copy           |
| `b` | Toggle body    |
| `B` | Toggle headers |
| `f` | Fullscreen     |
| `w` | Pin            |
| `W` | Diff           |

### Navigation Shortcuts

| Key      | Action         |
| -------- | -------------- |
| `j`/`k`  | Up/Down        |
| `gg`     | Top            |
| `G`      | Bottom         |
| `Ctrl+D` | Half page down |
| `Ctrl+U` | Half page up   |
| `:`      | Jump to line   |

## Tips

1. Press `?` anytime for in-app help
2. Use `/` to search within help modal
3. TAB switches focus between panels
4. Green border shows active panel
5. Vim-style navigation works throughout
6. Most actions work on selected item
7. Modals close with `Esc`
8. Search context changes with focused panel

## Context-Specific Keys

### File List Focused

1. `Enter`: Execute request
2. `i`: Inspect request
3. `x`: Edit file
4. `/`: Search filenames

### Response Focused

1. `s`: Save response
2. `c`: Copy response
3. `/`: Search response content
4. `f`: Fullscreen

### Modal Open

1. `Esc`: Close modal
2. `Enter`: Confirm
3. `Tab`: Next field
4. Navigation keys work

## Accessibility

All functionality is keyboard-accessible.

No mouse required.

Focus indicated by green border.

Visual feedback for all actions.
