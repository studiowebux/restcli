# HTTP TUI - Terminal HTTP Request Tool

A simple, keyboard-driven TUI for testing HTTP endpoints without the bloat of GUI tools.

## Features

- File-based request management (`.http` and `.yaml` files)
- Keyboard-driven navigation
- Header profiles for quick account switching
- Variable substitution
- Quick file duplication
- Auto-save session state
- Request/response history with timestamps
- OAuth 2.0 authentication flow (PKCE support)
- OpenAPI/Swagger import - auto-generate .http files
- Interactive documentation viewer with collapsible fields
- Global config directory (`~/.restcli/`) - use from anywhere
- Compiled binaries - no runtime dependencies
- YAML format support with JSON schema for autocomplete
- JSON response beautification (automatic pretty-printing)
- Response scrolling with vim-style j/k keys
- CLI mode - pipe-friendly output for automation and scripting
- YAML conversion - transform JSON responses to YAML
- Fullscreen mode for focused response viewing

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
pbpaste | restcli-curl2http --output requests/my-request.http
# Or with deno:
pbpaste | deno task curl2http --output requests/my-request.http
```

See [INSTALL.md](./docs/INSTALL.md) for detailed installation and setup guide.
See [PROFILES.md](./docs/PROFILES.md) for detailed profile configuration guide.
See [CURL2HTTP.md](./docs/CURL2HTTP.md) for converting cURL commands to `.http` files.
See [CURL2HTTP.md](./docs/OPENAPI.md) for converting openapi to `.http` files.
See [DOCUMENTATION.md](./docs/DOCUMENTATION.md) for adding documentation to your requests.

## CLI Mode

`restcli` can run in two modes:

### TUI Mode (Interactive)

Run without any arguments to open the interactive TUI:

```bash
restcli           # Opens the TUI
./restcli         # If using local binary
deno task dev     # Using Deno
```

### CLI Mode (Non-interactive)

Run requests directly from the command line:

```bash
# Basic usage - outputs only response body (perfect for piping)
restcli requests/example.http

# Pipe to jq for JSON processing
restcli requests/users/list.http | jq '.users[] | .name'

# Full output mode - shows status, headers, and body
restcli --full requests/example.http
restcli -f requests/example.http

# YAML output - converts JSON response to YAML
restcli --yaml requests/example.http
restcli -y requests/example.http

# Combine flags
restcli --full --yaml requests/example.http

# Use with profile
restcli --profile Admin requests/example.http
restcli -p Admin --yaml requests/example.http

# Show help
restcli --help
restcli -h
```

**CLI Flags:**
- `--help`, `-h`: Show help message
- `--full`, `-f`: Show full output (status line, headers, and body). Default: body only.
- `--yaml`, `-y`: Convert JSON response to YAML format
- `--profile <name>`, `-p <name>`: Use a specific profile for the request

**Default behavior** (without `--full`):
- Only the response body is printed to stdout
- Informational messages are suppressed
- Perfect for piping to tools like `jq`, `grep`, or scripts

**With `--full`:**
- Status line with response time and sizes
- All response headers
- Response body
- Traditional full output format

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

```text
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

```text
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

## Variables and Profiles

Variables use `{{varName}}` syntax in your requests. Headers and variables are configured in **profiles** (`.profiles.json`).

### Profiles

Profiles store your headers and variables permanently. Create profiles in `.profiles.json`:

```json
[
  {
    "name": "User 1",
    "headers": {
      "Authorization": "Bearer {{token}}",
      "X-User-ID": "user1"
    },
    "variables": {
      "baseUrl": "http://localhost:3000",
      "token": "user1-token-here"
    }
  },
  {
    "name": "User 2",
    "headers": {
      "Authorization": "Bearer {{token}}",
      "X-User-ID": "user2"
    },
    "variables": {
      "baseUrl": "http://localhost:3000",
      "token": "user2-token-here"
    }
  }
]
```

Press `p` in the TUI to cycle through profiles, or press `v` to open the variable editor.

### Session Data (Ephemeral)

`.session.json` contains **ephemeral** state that is linked to the currently active profile:
- Active profile name
- Temporary runtime variables (auto-extracted tokens, etc.)
- Session state gets cleared when you switch profiles

**Important:** Configure your headers and variables in `.profiles.json` (permanent), not `.session.json` (ephemeral).

The TUI auto-extracts `token` or `accessToken` from JSON responses and temporarily stores them in the session.

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
- `h` - Open header editor (add, edit, delete headers)
- `m` - View request documentation
- `j` - Scroll response down (useful for long JSON responses)
- `k` - Scroll response up
- `b` - Toggle response body visibility
- `f` - Toggle fullscreen mode (hide sidebar)
- `H` - View request history

### OAuth 2.0
- `o` - Start OAuth authentication flow (uses profile's OAuth config)
- `O` - Configure OAuth settings for current profile

### Utilities
- `?` - Show help
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

## Header Editor

Press `h` to open the interactive header editor. This allows you to manage **profile headers** without editing `.profiles.json` manually.

The header editor works identically to the variable editor:

### List Mode
- Navigate headers with `↑` / `↓`
- `A` - Add new header
- `E` or `Enter` - Edit selected header's value
- `D` - Delete selected header
- `ESC` - Exit header editor

### Add Mode
- Type to enter header name and value
- `Tab` - Switch between name and value fields
- `Enter` - Save header to active profile
- `ESC` - Cancel

### Edit Mode
- Type to change the value (header name cannot be changed)
- `Enter` - Save changes to active profile
- `ESC` - Cancel

### Delete Mode
- `Y` - Confirm deletion from profile
- `N` or `ESC` - Cancel

**Important Notes:**
- Headers are saved to the **active profile** in `.profiles.json`
- Profile headers are merged with request headers (request headers take precedence)
- Common headers: `Authorization`, `Content-Type`, `X-API-Key`, etc.

## OAuth 2.0 Authentication

Press `O` to configure OAuth 2.0 settings for the active profile. The OAuth configuration supports two modes:

### Manual Mode
Provide a complete authorization endpoint URL directly:
- `authEndpoint`: Full authorization URL with all parameters
- `tokenUrl`: Token endpoint for code exchange (required for authorization code flow)
- `responseType`: `code` (default) or `token`

### Auto-Build Mode
Build the authorization URL from components:
- `authUrl`: Base authorization URL
- `tokenUrl`: Token endpoint URL
- `clientId`: OAuth client ID
- `clientSecret`: OAuth client secret (optional)
- `redirectUri`: Callback URL (default: `http://localhost:8888/callback`)
- `scope`: OAuth scopes (default: `openid`)
- `responseType`: `code` (default) or `token`
- `webhookPort`: Local server port (default: 8888)
- `tokenStorageKey`: Variable name to store token (default: `token`)

### OAuth Flow

Press `o` to start the OAuth authentication flow:
1. Local webhook server starts on configured port
2. Browser opens to OAuth provider
3. User completes authentication
4. OAuth provider redirects back to local server
5. Authorization code is exchanged for access token
6. Token is automatically stored in profile variables

The flow supports PKCE (Proof Key for Code Exchange) for enhanced security.

## History Viewer

Press `H` to view request history:
- See all previously executed requests with timestamps
- Navigate with `↑`/`↓` arrow keys
- Press `Enter` to view full request/response details
- History is stored in `~/.restcli/.history.json`
- Each entry includes: timestamp, file path, method, URL, status, and response time

## Documentation Viewer

Press `m` to view interactive documentation for the current request:
- Shows request parameters, examples, and response schemas
- Navigate with arrow keys
- Press `Space` to expand/collapse nested fields
- Useful for understanding API endpoints with complex request/response structures

## File Organization

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
