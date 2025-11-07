# REST CLI - Terminal HTTP Request Tool

A simple, keyboard-driven TUI for testing HTTP endpoints.

## Features

- File-based request management (`.http` and `.yaml` files)
- Keyboard-driven navigation
- Header profiles for quick account switching
- Variable substitution and support of simple and multi-values fields
- Quick file duplication
- Auto-save session state
- Request/response history with timestamps
- OAuth 2.0 authentication flow (PKCE support)
- OpenAPI/Swagger import - auto-generate .http files
- Interactive documentation viewer with collapsible fields
- Global config directory (`~/.restcli/`)
- Compiled binaries
- YAML format support with JSON schema for autocomplete
- JSON response beautification
- Response scrolling
- CLI mode
- YAML conversion
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
See [OPENAPI.md](./docs/OPENAPI.md) for converting openapi to `.http` files.
See [DOCUMENTATION.md](./docs/DOCUMENTATION.md) for adding documentation to your requests.

## CLI Mode

`restcli` can run in two modes:

### TUI Mode (Interactive)

Run without any arguments to open the interactive TUI:

```bash
# Opens the TUI
restcli
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
- Perfect for piping to tools like `jq`, `grep`, or scripts

**With `--full`:**

- Status line with response time and sizes
- All response headers
- Response body
- Traditional full output format

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
    },
    "workdir": "",
    "editor": "zed",
    "oauth": { }
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
    },
    "workdir": "",
    "editor": "zed",
    "oauth": { }
  }
]
```

Press `p` in the TUI to cycle through profiles.

### Session Data (Ephemeral)

`.session.json` contains **ephemeral** state that is linked to the currently active profile:

- Active profile name
- Temporary runtime variables (auto-extracted tokens, etc.)
- Session state gets cleared when you switch profiles

**Important:** Configure your headers and variables in `.profiles.json`, not `.session.json`.

The TUI auto-extracts `token` or `accessToken` from JSON responses and temporarily stores them in the session.

## Keyboard Shortcuts

### Navigation

- `↑/↓` - Navigate files (circular)
- `Page Up/Down` - Fast scroll (jumps by visible page size)
- `:` - Goto line in hex
- `Ctrl+r` - Search files by name (press `Ctrl+r` again to cycle through matches)

### Actions

- `i` - Inspect request (preview what will be sent without executing)
- `Enter` - Execute request
- `x` - Open file in external editor (configured in profile)
- `X` (Shift+X) - Configure external editor for active profile
- `d` - Duplicate current file
- `s` - Save response/inspection to file (timestamp-based filename)
- `c` - Copy response body/error to clipboard
- `r` - Refresh file list
- `p` - Switch profile (cycles through profiles)
- `v` - Open variable editor (add, edit, delete variables)
- `h` - Open header editor (add, edit, delete headers)
- `P` (Shift+P) - Open .profiles.json in external editor
- `S` (Shift+S) - Open .session.json in external editor
- `m` - View request documentation
- `j` - Scroll response down
- `k` - Scroll response up
- `b` - Toggle response body visibility
- `f` - Toggle fullscreen mode (hide sidebar)
- `H` (Shift+H) - View request history

### OAuth 2.0

- `o` - Start OAuth authentication flow (uses profile's OAuth config)
- `O` (Shift+O) - Configure OAuth settings for current profile

### Utilities

- `?` - Show help
- `ESC` - Clear status message / Cancel search or goto
- `q` - Quit

## Variable Editor

Press `v` to open the interactive variable editor. This allows you to manage **profile variables**:

### List Mode

- Navigate variables with `↑` / `↓`
- `a` - Add new variable
- `e` or `Enter` - Edit selected variable's value
- `d` - Delete selected variable
- `O` (Shift+O) - Open options selector (for multi-value variables)
- `M` (Shift+M) - Manage options (add/edit/delete options for multi-value variables)
- `ESC` - Exit variable editor

### Add Mode

- Type to enter key name
- `Tab` - Move to value field
- `Shift+Tab` - Go back to key field (to fix typos)
- For **Simple variables**:
  - Type value and press `Enter` to save
- For **Multi-value variables**:
  - Press `Tab` again to toggle to Multi-value type
  - Enter each option and press `Enter` (field clears, ready for next)
  - Press `1-9` to set which option is active (marked with ✓)
  - Press `Enter` on empty field to finish and save
  - Press `Tab` to toggle back to Simple (confirms if options exist)
- `Ctrl+k` - Clear current field
- `ESC` - Cancel and return to list

### Edit Mode

- Type to change the value (key cannot be changed)
- `Enter` - Save changes to active profile
- `ESC` - Cancel

### Delete Mode

- `y` - Confirm deletion from profile
- `n` or `ESC` - Cancel

### Multi-Value Variables

Variables can have multiple predefined options with one active value. This is useful for switching between environments, API versions, or any value with a fixed set of choices.

**Creating Multi-Value Variables:**

Edit `.profiles.json` manually to create multi-value variables:

```json
{
  "name": "My Profile",
  "variables": {
    "baseUrl": "https://api.example.com", // Simple variable
    "environment": {
      // Multi-value variable
      "options": ["dev", "staging", "prod"],
      "active": 2,
      "description": "API environment"
    },
    "apiVersion": {
      "options": ["v1", "v2", "v3"],
      "active": 0
    }
  }
}
```

**Using Multi-Value Variables in TUI:**

Multi-value variables are displayed with a `[N options] ◀` indicator showing the currently active value:

**Quick Option Selection (Press `O (Shift+O)`):**

- Navigate with `↑` / `↓` arrow keys
- Press `1-9` for instant selection (first 9 options)
- Press `Enter` to select highlighted option
- Current active option marked with `✓`
- Press `ESC` to cancel

**Manage Options (Press `M (Shift+M)`):**

- `a` - Add new option to the list
- `e` - Edit/rename selected option
- `d` - Delete option (cannot delete active option)
- `Space` - Set selected option as active
- `↑` / `↓` - Navigate options
- `ESC` - Return to variable list

## External Editor Integration

Press `X (Shift+X)` to configure an external editor for the active profile, then press `x` to open the selected file in that editor.

### Configuration

1. Press `X (Shift+X)` to open the editor configuration modal
2. Enter your editor command (e.g., `zed`, `code`, `vim`, `nvim`, `subl`, `etc.`)
3. Press `Enter` to save

The editor setting is saved per-profile in `.profiles.json`:

```json
{
  "name": "My Profile",
  "workdir": "",
  "editor": "zed",
  "headers": { },
  "variables": { },
  "oauth": { }
}
```

### Usage

Once configured, press `x` on any request file to open it in your editor. The editor opens in the background, so the TUI remains running.

## Header Editor

Press `h` to open the interactive header editor. This allows you to manage **profile headers**.

The header editor works identically to the variable editor:

### List Mode

- Navigate headers with `↑` / `↓`
- `a` - Add new header
- `e` or `Enter` - Edit selected header's value
- `d` - Delete selected header
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

- `y` - Confirm deletion from profile
- `n` or `ESC` - Cancel

## OAuth 2.0 Authentication

Press `O (Shift+O)` to configure OAuth 2.0 settings for the active profile. The OAuth configuration supports two modes:

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

Press `H (Shift+H)` to view request history:

- See all previously executed requests with timestamps
- Navigate with `↑`/`↓` arrow keys
- Press `Enter` to view full request/response details
- History is stored in `~/.restcli/.history.json`
- Each entry includes: timestamp, file path, method, URL, status, and response time

## Documentation Viewer (OpenAPI)

Press `m` to view interactive documentation for the current request:

- Shows request parameters, examples, and response schemas
- Navigate with arrow keys
- Press `Space` to expand/collapse nested fields
- Useful for understanding API endpoints with complex request/response structures

## File Organization

- Use duplicate (`d`) to quickly create variations of requests
- Profile headers are merged with request headers (request headers take precedence)
- Files are auto-discovered from `./requests/` directory
- **Text selection**: Use `s` to save or `c` to copy response.
- Use `r` to refresh file list after creating new `.http` files outside the TUI
- **Inspect before executing**: Press `i` to preview the final request:
  - See the actual URL after variable substitution (`{{baseUrl}}` → `http://localhost:3000`)
  - View all headers including those from active profile
  - Check the request body before sending
  - Useful for debugging variable issues or verifying profile headers
- **Quick navigation**:
  - Files are numbered in **hexadecimal** (shown in sidebar)
  - Use `:` followed by hex number (e.g., `:64` to jump to file #100)
  - Use `Ctrl+r` to search by filename, then `Ctrl+r` again to cycle through matches
  - Search is case-insensitive and matches anywhere in the filename
  - Arrow keys wrap around (circular)
  - Page Up/Down for fast scrolling through long lists (jumps by ~1 screen height)
