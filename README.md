# REST CLI - Terminal HTTP Request Tool

A keyboard-driven TUI for testing HTTP endpoints.

> Go version only. Deno version deprecated. Tested on macOS.

## Features

- File-based requests (`.http`, `.yaml`, `.json`)
- Variable substitution with `{{varName}}` and shell commands `$(cmd)`
- Header profiles for quick account switching
- OAuth 2.0 with PKCE support
- OpenAPI/Swagger import
- Request history and documentation viewer
- CLI mode for scripting (JSON/YAML output)

## Key Shortcuts

Press `?` in-app for full shortcuts and features.

**Or**

```bash
restcli --help
```

## HTTP File Format

```text
### Request Name (optional)
METHOD url
Header: value
Another-Header: value

{
  "body": "for POST/PUT"
}
```

### Examples

```text
### Login
POST {{baseUrl}}/auth/login
Content-Type: application/json

{
  "username": "test",
  "password": "pass"
}
```

```text
### Get Profile
GET {{baseUrl}}/users/{{userId}}
Authorization: Bearer {{token}}
```

```yaml
---
name: Get User
method: GET
url: "https://jsonplaceholder.typicode.com/users/{{userId}}"
headers:
  Accept: "application/json"
```

```json
{
  "name": "List Users",
  "method": "GET",
  "url": "https://jsonplaceholder.typicode.com/todo/1"
}
```

```json
{
  "name": "Create User",
  "method": "POST",
  "url": "https://jsonplaceholder.typicode.com/todo/2",
  "headers": { "Content-Type": "application/json" },
  "body": "{\"name\": \"John\"}"
}
```

## Variables and Profiles

Variables use `{{varName}}` syntax in your requests. Headers and variables are configured in **profiles** (`.profiles.json`).

### Shell Command Variables

Variables can execute shell commands using `$(command)` syntax:

```json
{
  "name": "Dynamic Profile",
  "variables": {
    "timestamp": "$(date +%s)",
    "gitBranch": "$(git branch --show-current)",
    "randomId": "$(uuidgen)",
    "apiKey": "$(cat ~/.secrets/api-key)",
    "foo": "bar"
  }
}
```

**Features:**

- Commands execute when variables are resolved (before each request)
- 5-second timeout per command
- Works on Linux, macOS (uses `sh`)
- Supports any shell command that outputs to stdout
- Errors are logged and result in empty string

**Security Note:** Shell commands run with your user permissions. Only use trusted commands in your profiles.

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
    "output": "json",
    "oauth": {}
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
    "output": "text",
    "oauth": {}
  }
]
```

### Session Data (Ephemeral)

`.session.json` contains **ephemeral** state that is linked to the currently active profile:

- Active profile name
- Temporary runtime variables (auto-extracted tokens, etc.)
- Session state gets cleared when you switch profiles

**Important:** Configure your headers and variables in `.profiles.json`, not `.session.json`.

The TUI auto-extracts `token` or `accessToken` from JSON responses and temporarily stores them in the session.
