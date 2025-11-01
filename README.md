# HTTP TUI - Terminal HTTP Request Tool

A simple, keyboard-driven TUI for testing HTTP endpoints without the bloat of GUI tools.

## Features

- 📁 File-based request management (`.http` and `.yaml` files)
- 🎯 Keyboard-driven navigation
- 👤 Header profiles for quick account switching
- 🔄 Variable substitution
- 📋 Quick file duplication
- 💾 Auto-save session state
- 📜 Request/response history with timestamps
- 🏠 Global config directory (`~/.restcli/`) - use from anywhere
- 🔨 Compiled binaries - no runtime dependencies
- 🧾 YAML format support with JSON schema for autocomplete
- ✨ JSON response beautification (automatic pretty-printing)
- 📄 Response scrolling with vim-style j/k keys

## Quick Start

### Using Compiled Binary (Recommended)

```bash
# Build the binary
deno task build

# Run it (auto-initializes ~/.restcli/ on first run)
./restcli

# Or install globally
sudo mv restcli /usr/local/bin/
restcli  # Use from anywhere!
```

### Using Deno

```bash
# Run the TUI
deno task dev

# Run a specific request file
deno task run requests/example.http

# Convert cURL to .http (from browser DevTools, docs, etc.)
pbpaste | deno task curl2http
```

See [INSTALL.md](./docs/INSTALL.md) for detailed installation and setup guide.
See [PROFILES.md](./docs/PROFILES.md) for detailed profile configuration guide.
See [HEX-REFERENCE.md](./docs/HEX-REFERENCE.md) for hexadecimal numbering explanation.
See [CURL2HTTP.md](./docs/CURL2HTTP.md) for converting cURL commands to `.http` files.

## File Structure

```
.
├── requests/                    # Your .http files (supports nested dirs!)
│   ├── auth/
│   │   └── login.http
│   ├── users/
│   │   ├── admin/
│   │   │   └── list.http
│   │   └── player/
│   │       └── profile.http
│   └── examples/
│       ├── get-example.http
│       └── post-example.http
├── .session.json               # Auto-saved variables and active profile
├── .profiles.json              # Header profiles for switching users
└── tui.ts                      # The TUI app
```

The TUI will display files with their relative paths (e.g., `auth/login.http`, `users/admin/list.http`) making it easy to organize your 100+ endpoints by feature, domain, or user type.

## HTTP File Format

```http
### Request Name (optional)
METHOD url
Header: value
Another-Header: value

{
  "body": "for POST/PUT"
}

###
```

### Example

```http
### Login
POST {{baseUrl}}/auth/login
Content-Type: application/json

{
  "username": "test",
  "password": "pass"
}

### Get Profile
GET {{baseUrl}}/users/{{userId}}
Authorization: Bearer {{token}}
```

## Variables

Variables use `{{varName}}` syntax and are stored in `.session.json`:

```json
{
  "variables": {
    "baseUrl": "http://localhost:3000",
    "token": "eyJhbGc...",
    "userId": "123"
  }
}
```

The TUI auto-extracts `token` or `accessToken` from JSON responses.

## Header Profiles

Create profiles in `.profiles.json` to quickly switch between users:

```json
[
  {
    "name": "User 1",
    "headers": {
      "Authorization": "Bearer {{token}}",
      "X-User-ID": "user1"
    }
  },
  {
    "name": "User 2",
    "headers": {
      "Authorization": "Bearer {{token2}}",
      "X-User-ID": "user2"
    }
  }
]
```

Press `p` in the TUI to cycle through profiles.

## Keyboard Shortcuts

### Navigation
- `↑/↓` - Navigate files (circular: wraps from top to bottom and vice versa)
- `Page Up/Down` - Fast scroll (jumps by visible page size)
- `:` - Goto line in hex (e.g., `:64` jumps to file #100, `:FF` to #255)
- `Ctrl+R` - Search files by name (press `Ctrl+R` again to cycle through matches)

### Actions
- `i` - Inspect request (preview what will be sent without executing)
- `Enter` - Execute request
- `d` - Duplicate current file
- `s` - Save response/inspection to file (timestamp-based filename)
- `c` - Copy response body/error to clipboard
- `r` - Refresh file list
- `p` - Switch profile (cycles through profiles)
- `v` - Open variable editor (add, edit, delete variables)
- `j` - Scroll response down (useful for long JSON responses)
- `k` - Scroll response up

### Utilities
- `ESC` - Clear status message / Cancel search or goto
- `q` - Quit

## Variable Editor

Press `v` to open the interactive variable editor. This allows you to manage **profile variables** without editing `.profiles.json` manually:

### List Mode
- Navigate variables with `↑` / `↓`
- `A` - Add new variable
- `E` or `Enter` - Edit selected variable's value
- `D` - Delete selected variable
- `ESC` - Exit variable editor

### Add Mode
- Type to enter key and value
- `Tab` - Switch between key and value fields
- `Enter` - Save variable to active profile
- `ESC` - Cancel

### Edit Mode
- Type to change the value (key cannot be changed)
- `Enter` - Save changes to active profile
- `ESC` - Cancel

### Delete Mode
- `Y` - Confirm deletion from profile
- `N` or `ESC` - Cancel

**Important Notes:**
- Variables are saved to the **active profile** in `.profiles.json`
- Session variables (`.session.json`) are temporary state that gets cleared when switching profiles
- Profile variables are permanent configuration
- Long values are automatically truncated to prevent overlap

## Tips

- Organize requests by feature/domain in subdirectories
- Use duplicate (`d`) to quickly create variations of requests
- Profile headers are merged with request headers (request headers take precedence)
- Files are auto-discovered from `./requests/` directory
- **Text selection**: Use `s` to save or `c` to copy response instead of selecting with mouse (avoids copying TUI structure)
- Use `r` to refresh file list after creating new `.http` files outside the TUI
- **Inspect before executing**: Press `i` to preview the final request:
  - See the actual URL after variable substitution (`{{baseUrl}}` → `http://localhost:3000`)
  - View all headers including those from active profile
  - Check the request body before sending
  - Useful for debugging variable issues or verifying profile headers
- **Quick navigation**:
  - Files are numbered in **hexadecimal** (shown in sidebar) to save space
    - Examples: `1` = file #1, `A` = file #10, `64` = file #100, `FF` = file #255, `3E8` = file #1000
  - Use `:` followed by hex number (e.g., `:64` to jump to file #100)
  - Use `Ctrl+R` to search by filename, then `Ctrl+R` again to cycle through matches
  - Search is case-insensitive and matches anywhere in the filename
  - Arrow keys wrap around (circular): press Up at top to jump to bottom, Down at bottom to jump to top
  - Page Up/Down for fast scrolling through long lists (jumps by ~1 screen height)
