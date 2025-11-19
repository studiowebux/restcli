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

# Structured JSON output (includes status, headers, body, duration, sizes)
restcli --output json requests/example.http | jq '.status'
restcli -o json requests/example.http

# YAML structured output
restcli --output yaml requests/example.http
restcli -o yaml requests/example.http

# Save response to file (auto-detects format from extension)
restcli --save response.json requests/example.http  # JSON format
restcli --save response.yaml requests/example.http  # YAML format
restcli --save response.txt requests/example.http   # Text format

# Or explicitly specify format (overrides extension)
restcli -s output.txt -o json requests/example.http

# Override request body from command line
restcli --body '{"key":"value"}' requests/example.http
restcli -b '{"user":"admin"}' requests/example.http

# Pipe body from another command
echo '{"dynamic":"data"}' | restcli requests/example.http
cat payload.json | restcli requests/create.http

# Use with profile
restcli --profile Admin requests/example.http
restcli -p Admin --yaml requests/example.http

# Show help
restcli --help
restcli -h
```

**CLI Flags:**

- `--help`, `-h`: Show help message
- `--full`, `-f`: Show full output (status line, headers, and body)
- `--body-only`: Show only response body (explicit override)
- `--output <format>`, `-o <format>`: Output format: `json`, `yaml`, or `text` (default: `text`)
  - `json`: Structured output with status, headers, body, duration, and sizes
  - `yaml`: Same structured data in YAML format
  - `text`: Body only (current behavior)
- `--save <file>`, `-s <file>`: Save response to file instead of stdout
- `--body <json>`, `-b <json>`: Override request body with inline JSON
- `--yaml`, `-y`: Convert JSON response to YAML format (deprecated, use `-o yaml`)
- `--profile <name>`, `-p <name>`: Use a specific profile for the request

**Smart Output Detection:**

The CLI automatically detects whether output is to a terminal (TTY) or being piped:

- **When piped** (e.g., `restcli file.http | jq`): Body-only output (perfect for JSON processing)
- **In terminal**: Full output with status and headers by default
- Override with `--full` or `--body-only` flags

**Output Format Priority:**

When multiple output format sources are present:

1. `--output` flag (highest priority)
2. File extension in `--save` flag (`.json`, `.yaml`, `.txt`)
3. Profile's `output` setting
4. Default (`text`)

**Body Override Priority:**

When multiple body sources are present:

1. `--body` flag (highest priority)
2. Stdin (if piped)
3. Request file body (default)

**Profile Output Configuration:**

Set a default output format in your profile's `.profiles.json`:

```json
{
  "name": "My Profile",
  "output": "json",
  "headers": {},
  "variables": {}
}
```

This affects:

- **CLI mode**: Default format when no `--output` flag is used
- **TUI `s` shortcut**: Format used when saving responses with `s` key
- Supported values: `"json"`, `"yaml"`, `"text"` (default)

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
- `R` (Shift+R) - **Rename current file**
- `s` - Save response/inspection to file (timestamp-based filename)
- `c` - Copy response body/error to clipboard
- `r` - **Refresh file list and reload profiles/session**
- `n` - **Create new profile interactively**
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
  "headers": {},
  "variables": {},
  "oauth": {}
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
