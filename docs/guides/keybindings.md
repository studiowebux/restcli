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
    "ctrl+c": "quit_force"
  },
  "normal": {
    "q": "quit",
    "enter": "execute"
  }
}
```

Format: `"key": "action"` - the key triggers the action.

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

Single keys: `"q": "quit"`, `"enter": "execute"`, `"esc": "cancel"`

Multiple keys (all map to same action): Multiple entries like `"esc": "close_modal"` and `"q": "close_modal"`

Modifiers: `"ctrl+c": "quit_force"`, `"shift+up": "navigate_up"`, `"alt+x": "some_action"`

Multi-key sequences: `"gg": "go_to_top"` (press g twice rapidly)

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
    "ctrl+q": "quit",
    "space": "execute"
  }
}
```

Only override the keys you want to change. Unspecified keys use defaults.

Note: You must map the key to the action, not the action to the key.

## Vim-style Example

```json
{
  "version": "1.0",
  "normal": {
    "k": "navigate_up",
    "j": "navigate_down",
    "gg": "go_to_top",
    "G": "go_to_bottom",
    "ctrl+u": "half_page_up",
    "ctrl+d": "half_page_down",
    "/": "open_search",
    "n": "search_next",
    "N": "search_previous"
  }
}
```

## Emacs-style Example

```json
{
  "version": "1.0",
  "normal": {
    "ctrl+p": "navigate_up",
    "ctrl+n": "navigate_down",
    "ctrl+s": "open_search",
    "ctrl+x": "quit"
  },
  "text_input": {
    "ctrl+b": "text_move_left",
    "ctrl+f": "text_move_right",
    "ctrl+a": "text_move_home",
    "ctrl+e": "text_move_end",
    "ctrl+u": "text_clear_before",
    "ctrl+k": "text_clear_after"
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
    "gg": "go_to_top"
  }
}
```

Press `g` twice rapidly to trigger `go_to_top`.

Supported sequences: `gg` (go to top)

## Conflicts

Keys bound multiple times in same context cause errors.

Context-specific bindings override global bindings (intentional shadowing).

Example conflict (invalid):

```json
{
  "normal": {
    "q": "quit",
    "q": "close_modal"
  }
}
```

This is invalid because the same key `q` is bound twice in the same context.

Fix by choosing different keys:

```json
{
  "normal": {
    "q": "quit",
    "esc": "close_modal"
  }
}
```

Or bind multiple keys to the same action:

```json
{
  "normal": {
    "q": "close_modal",
    "esc": "close_modal"
  }
}
```

## Text Input Context Switching

When typing in text input fields, the system automatically switches to the `text_input` context to prevent single-letter keybinds from intercepting your typing.

For example:
- In stress test config modal, when typing a name, keys like `d`, `l`, `r` are typed as characters
- In profile editor, when typing editor name, space bar inserts a space character
- Special navigation keys (up/down arrows, enter, esc) remain active for field navigation

This context switching is automatic and requires no configuration.

## Implementation Notes

Keybinding system uses action-based registry with context hierarchy.

User config overlays default mappings.

No runtime rebinding. Requires restart to apply changes.

Text input fields automatically switch to `text_input` context to prevent conflicts.
