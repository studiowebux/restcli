---
title: Keybindings
tags:
  - guide
  - customization
---

# Keybindings

Custom keybinding support for all TUI modes.

## Configuration File

Location: `~/.restcli/keybinds.json`

Auto-created on first launch with default mappings.

## Structure

```json
{
  "version": "1.0",
  "global": {
    "quit_force": "ctrl+c"
  },
  "normal": {
    "quit": "q",
    "execute": "enter"
  }
}
```

## Contexts

Keybindings organized by context:

- `global` - Available everywhere
- `normal` - Main mode
- `search` - Search input
- `goto` - Goto line input
- `variables` - Variable editor
- `headers` - Header editor
- `profiles` - Profile manager
- `documentation` - Documentation viewer
- `history` - History browser
- `analytics` - Analytics viewer
- `stress_test` - Stress test modes
- `help` - Help viewer
- `inspect` - Request inspector
- `websocket` - WebSocket interface
- `modal` - Generic modals
- `text_input` - Text input fields
- `confirm` - Confirmation dialogs

## Key Format

Single keys: `"q"`, `"enter"`, `"esc"`

Multiple keys (any triggers action): `"esc,q"`

Modifiers: `"ctrl+c"`, `"shift+up"`, `"alt+x"`

Multi-key sequences: `"gg"` (press g twice)

## Action Reference

### Global

| Action | Default | Description |
| --- | --- | --- |
| `quit_force` | `ctrl+c` | Force quit |

### Normal Mode

| Action | Default | Description |
| --- | --- | --- |
| `quit` | `q` | Quit application |
| `execute` | `enter` | Execute request |
| `switch_focus` | `tab` | Switch panel focus |
| `navigate_up` | `up,k` | Move up |
| `navigate_down` | `down,j` | Move down |
| `page_up` | `pgup` | Page up |
| `page_down` | `pgdown` | Page down |
| `half_page_up` | `ctrl+u` | Half page up |
| `half_page_down` | `ctrl+d` | Half page down |
| `go_to_top` | `gg,home` | Jump to top |
| `go_to_bottom` | `G,end` | Jump to bottom |
| `open_goto` | `:` | Open goto line |
| `open_search` | `/` | Open search |
| `open_editor` | `x` | Open in editor |
| `configure_editor` | `X` | Configure editor |
| `duplicate_file` | `d` | Duplicate file |
| `delete_file` | `D` | Delete file |
| `rename_file` | `R` | Rename file |
| `create_file` | `F` | Create file |
| `refresh_files` | `r` | Refresh list |
| `save_response` | `s` | Save response |
| `copy_to_clipboard` | `c` | Copy response |
| `toggle_body` | `b` | Toggle body |
| `toggle_headers` | `B` | Toggle headers |
| `toggle_fullscreen` | `f` | Toggle fullscreen |
| `pin_response` | `w` | Pin for comparison |
| `show_diff` | `W` | Show diff |
| `filter_response` | `J` | Filter with JMESPath |
| `open_inspect` | `i` | Request inspector |
| `open_variables` | `v` | Variable editor |
| `open_headers` | `h` | Header editor |
| `open_help` | `?` | Help viewer |
| `open_history` | `H` | History browser |
| `open_analytics` | `A` | Analytics viewer |
| `open_stress_test` | `S` | Stress test |
| `open_profiles` | `p` | Profile manager |
| `open_documentation` | `m` | Documentation |

### Variable Editor

| Action | Default | Description |
| --- | --- | --- |
| `close_modal` | `esc,v,q` | Close editor |
| `var_add` | `a` | Add variable |
| `var_edit` | `e` | Edit variable |
| `var_delete` | `d` | Delete variable |
| `var_manage` | `m` | Manage options |

### Header Editor

| Action | Default | Description |
| --- | --- | --- |
| `close_modal` | `esc,h,q` | Close editor |
| `header_add` | `C` | Add header |
| `header_edit` | `enter` | Edit header |
| `header_delete` | `r` | Delete header |

### WebSocket

| Action | Default | Description |
| --- | --- | --- |
| `close_modal` | `esc,q` | Close interface |
| `switch_pane` | `tab` | Switch pane |
| `ws_send` | `enter` | Send message |
| `ws_disconnect` | `d` | Disconnect |
| `ws_clear` | `C` | Clear history |

### Text Input

| Action | Default | Description |
| --- | --- | --- |
| `text_submit` | `enter` | Submit |
| `text_cancel` | `esc` | Cancel |
| `text_paste` | `ctrl+v` | Paste |
| `text_backspace` | `backspace` | Delete before |
| `text_delete` | `delete` | Delete at cursor |
| `text_move_left` | `left` | Move cursor left |
| `text_move_right` | `right` | Move cursor right |
| `text_move_home` | `home,ctrl+a` | Move to start |
| `text_move_end` | `end,ctrl+e` | Move to end |
| `text_clear_after` | `ctrl+k` | Clear after cursor |

### Modals

| Action | Default | Description |
| --- | --- | --- |
| `close_modal` | `esc,q` | Close modal |
| `navigate_up` | `up,k` | Move up |
| `navigate_down` | `down,j` | Move down |

## Customization

Edit `~/.restcli/keybinds.json`:

```json
{
  "version": "1.0",
  "normal": {
    "quit": "ctrl+q",
    "execute": "space"
  }
}
```

Only override the keys you want to change. Unspecified keys use defaults.

## Vim-style Example

```json
{
  "version": "1.0",
  "normal": {
    "navigate_up": "k",
    "navigate_down": "j",
    "go_to_top": "gg",
    "go_to_bottom": "G",
    "half_page_up": "ctrl+u",
    "half_page_down": "ctrl+d",
    "open_search": "/",
    "search_next": "n",
    "search_previous": "N"
  }
}
```

## Emacs-style Example

```json
{
  "version": "1.0",
  "normal": {
    "navigate_up": "ctrl+p",
    "navigate_down": "ctrl+n",
    "open_search": "ctrl+s",
    "quit": "ctrl+x,ctrl+c"
  },
  "text_input": {
    "text_move_left": "ctrl+b",
    "text_move_right": "ctrl+f",
    "text_move_home": "ctrl+a",
    "text_move_end": "ctrl+e",
    "text_clear_before": "ctrl+u",
    "text_clear_after": "ctrl+k"
  }
}
```

## Validation

Restart restcli to apply changes.

Errors logged to stderr on startup:

```
warning: keybinds config error, using defaults: ...
```

Invalid keys are ignored. Valid keys load successfully.

## Reset to Defaults

Delete or rename `~/.restcli/keybinds.json` to restore defaults.

New example file created on next launch.

## Multi-Key Sequences

Bind actions to multi-key sequences:

```json
{
  "normal": {
    "go_to_top": "gg"
  }
}
```

Press `g` twice to trigger `go_to_top`.

Currently supported: `gg` sequence.

## Conflicts

Keys bound multiple times in same context cause errors.

Context-specific bindings override global bindings (intentional shadowing).

Example conflict (invalid):

```json
{
  "normal": {
    "quit": "q",
    "close_modal": "q"
  }
}
```

Fix by choosing different keys or using multiple triggers:

```json
{
  "normal": {
    "quit": "q",
    "close_modal": "esc"
  }
}
```

## Implementation Notes

Keybinding system uses action-based registry with context hierarchy.

User config overlays default mappings.

No runtime rebinding. Requires restart to apply changes.
