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

### Multi-Value Variables with Aliases

Variables can have multiple options with aliases for quick switching:

```json
{
  "name": "Development",
  "variables": {
    "environment": {
      "options": ["http://localhost:3000", "https://dev.api.com", "https://staging.api.com"],
      "active": 0,
      "description": "API environment",
      "aliases": {
        "local": 0,
        "dev": 1,
        "staging": 2
      }
    },
    "apiVersion": {
      "options": ["v1", "v2", "v3"],
      "active": 1,
      "description": "API version",
      "aliases": {
        "legacy": 0,
        "current": 1,
        "beta": 2
      }
    }
  }
}
```

**Usage with aliases:**

```bash
# Use alias to select option
restcli run api -p Development -e environment=staging -e apiVersion=beta

# Or use the actual value
restcli run api -p Development -e environment=https://staging.api.com
```

**Structure:**

- `options`: Array of possible values
- `active`: Index of currently active option (0-based)
- `description`: Optional description of the variable
- `aliases`: Map of alias names to option indices

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

---

## Filter & Query

Transform and filter response bodies using JMESPath expressions or bash commands.

### JMESPath Syntax

Use AWS CLI-style JMESPath expressions to filter and transform JSON responses:

```bash
# Filter with JMESPath expression
restcli run api --filter "items[?status==\`active\`]"

# Query/transform with JMESPath
restcli run api --query "[].{name: name, id: id}"

# Combine filter and query
restcli run api --filter "users[?age>\`18\`]" --query "[].email"
```

### Bash Command Syntax

Use `$(command)` to pipe the response through any bash command:

```bash
# Use jq for complex transformations
restcli run api --query '$(jq ".items[].name")'

# Use other tools
restcli run api --query '$(grep -o "id.*")'
```

### In Request Files

Add `filter` and `query` to your request files:

**YAML format:**

```yaml
name: List Active Items
method: GET
url: "https://api.example.com/items"
filter: "items[?status==`active`]"
query: "[].{name: name, id: id}"
```

**JSON format:**

```json
{
  "name": "Get User Emails",
  "method": "GET",
  "url": "https://api.example.com/users",
  "query": "$(jq '.users[].email')"
}
```

**HTTP format:**

```text
### Get Active Items
# @filter items[?status==`active`]
# @query [].name
GET https://api.example.com/items
```

### Profile Defaults

Set default filters for all requests in a profile:

```json
{
  "name": "Production",
  "headers": { "Authorization": "Bearer {{token}}" },
  "variables": { "baseUrl": "https://api.prod.com" },
  "defaultFilter": "",
  "defaultQuery": "[].{id: id, name: name}"
}
```

### Priority

Filter and query expressions are applied in this priority order:

1. CLI flags (`--filter`, `--query`)
2. Request file (`filter`, `query` fields)
3. Profile defaults (`defaultFilter`, `defaultQuery`)

---

## mTLS (Mutual TLS)

Configure client certificates for secure API connections that require mutual TLS authentication.

### Profile Configuration

Add TLS configuration to your profile in `.profiles.json`:

```json
{
  "name": "production",
  "headers": {
    "Content-Type": "application/json"
  },
  "tls": {
    "certFile": "/path/to/client.crt",
    "keyFile": "/path/to/client.key",
    "caFile": "/path/to/ca.crt"
  }
}
```

### Per-Request Configuration

**YAML format:**

```yaml
name: Secure API Call
method: GET
url: "https://secure-api.example.com/data"
tls:
  certFile: "/path/to/client.crt"
  keyFile: "/path/to/client.key"
  caFile: "/path/to/ca.crt"
```

**JSON format:**

```json
{
  "name": "Secure Request",
  "method": "GET",
  "url": "https://secure-api.example.com/data",
  "tls": {
    "certFile": "/path/to/client.crt",
    "keyFile": "/path/to/client.key"
  }
}
```

**HTTP format:**

```text
### Secure API Call
# @tls.certFile /path/to/client.crt
# @tls.keyFile /path/to/client.key
# @tls.caFile /path/to/ca.crt
GET https://secure-api.example.com/data
```

### Configuration Options

- `certFile`: Path to client certificate (PEM format)
- `keyFile`: Path to client private key (PEM format)
- `caFile`: Path to CA certificate for server verification (PEM format)
- `insecureSkipVerify`: Skip server certificate verification (for testing only)

### Priority

TLS configuration is applied in this priority order:

1. Request file (`tls` field)
2. Profile configuration (`tls` field)

---

# Completion command

For macOS (Zsh):

**Create completions directory**

```bash
mkdir -p ~/.zsh/completions
```

**Add to ~/.zshrc (if not already there)**

```bash
echo 'fpath=(~/.zsh/completions $fpath)' >> ~/.zshrc
echo 'autoload -Uz compinit && compinit' >> ~/.zshrc
```

**Generate completions**

```bash
restcli completion zsh > ~/.zsh/completions/\_restcli
```

**Reload shell**

```bash
source ~/.zshrc
```
