---
title: Variables
tags:
  - guide
---

# Variables

Variables enable dynamic request configuration.

## Basic Syntax

Use `{{varName}}` in requests:

```text
### Get User
GET {{baseUrl}}/users/{{userId}}
Authorization: Bearer {{token}}
```

## Where Variables Work

Variables can be used in:

- **URLs**: `{{baseUrl}}/users/{{userId}}`
- **Headers**: `Authorization: Bearer {{token}}`
- **Request Body**: `{"name": "{{userName}}"}`
- **TLS Paths**: `# @tls.certFile {{certsPath}}/client.pem`
- **Filters**: `# @filter items[?type=='{{itemType}}']`
- **Queries**: `# @query {{jmespathQuery}}`

### TLS Path Variables Example

```text
### Secure API Call
# @tls.certFile {{certsPath}}/client.crt
# @tls.keyFile {{certsPath}}/client.key
# @tls.caFile {{caPath}}/ca.crt
GET https://secure-api.example.com/data
```

With profile:

```json
{
  "name": "Secure",
  "variables": {
    "certsPath": "/etc/ssl/mycerts",
    "caPath": "/etc/ssl/ca"
  }
}
```

## Setting Variables

### CLI Flags

```bash
restcli -e baseUrl=https://api.example.com -e userId=5 get-user.http
```

### Profiles

In `.profiles.json`:

```json
{
  "name": "Dev",
  "variables": {
    "baseUrl": "https://dev.api.example.com",
    "userId": "1",
    "token": "dev-token-123"
  }
}
```

### Request Files

YAML format:

```yaml
name: Get User
method: GET
url: "{{baseUrl}}/users/{{userId}}"
variables:
  userId: "5"
```

JSON format:

```json
{
  "name": "Get User",
  "method": "GET",
  "url": "{{baseUrl}}/users/{{userId}}",
  "variables": {
    "userId": "5"
  }
}
```

### Session

Variables auto-extracted from responses to `.session.json`.

TUI automatically extracts `token` or `accessToken` from JSON responses.

## Environment Variables

Use `{{env.VAR_NAME}}` syntax:

```text
GET https://api.example.com/data
X-API-Key: {{env.API_KEY}}
```

Load from file:

```bash
restcli --env-file .env request.http
```

`.env` format:

```text
API_KEY=secret123
BASE_URL=https://api.example.com
```

## Shell Commands

Execute commands with `$(command)` syntax:

```json
{
  "variables": {
    "timestamp": "$(date +%s)",
    "branch": "$(git branch --show-current)",
    "uuid": "$(uuidgen)",
    "secret": "$(cat ~/.secrets/api-key)"
  }
}
```

### Execution Rules

1. Commands run when variables resolve (before each request)
2. 5 second timeout per command
3. Uses `sh` shell
4. Output captured from stdout
5. Errors logged, result is empty string

### Security

Shell commands run with your user permissions.

Only use trusted commands.

Validate commands in profiles before committing to repositories.

## Interactive Variables

Variables that always prompt for input at execution time, useful for dynamic values like LLM prompts, user inputs, or secrets.

### Marking Variables as Interactive

In TUI mode, press `v` to open the variable editor, select a variable, and press `i` to toggle interactive mode.

Interactive variables are marked with `[interactive]` in the variable list.

### Behavior

**TUI Mode:**
- Modal prompts appear automatically before request execution
- Shows variable name and default value (if any)
- Enter value and press Enter to continue
- ESC to cancel execution

**CLI Mode:**
- Always prompts for interactive variables (unless provided via `-e` flag)
- Cannot run in non-interactive mode without explicit values
- Bypassed by providing value: `restcli -e promptVar=value request.http`

### Example

Profile variable marked as interactive:

```json
{
  "name": "LLM API",
  "variables": {
    "userPrompt": {
      "value": "Default question...",
      "interactive": true
    },
    "apiKey": {
      "value": "sk-...",
      "interactive": true
    }
  }
}
```

Request:

```text
POST https://api.openai.com/v1/chat/completions
Authorization: Bearer {{apiKey}}
Content-Type: application/json

{
  "model": "gpt-4",
  "messages": [{"role": "user", "content": "{{userPrompt}}"}]
}
```

On execution:
1. Prompts for `userPrompt` value
2. Prompts for `apiKey` value
3. Executes request with entered values

Values are stored in session (ephemeral) for the current execution only.

### Use Cases

- **LLM prompts:** Change query text each time without editing files
- **Secrets:** Prompt for API keys/tokens instead of storing in profiles
- **User input:** Dynamic form data, search terms, IDs
- **One-time values:** Verification codes, session tokens

### Combined with Multi-Value

Interactive variables can also be multi-value:

```json
{
  "llmModel": {
    "options": ["gpt-4", "gpt-3.5-turbo", "claude-3"],
    "active": 0,
    "interactive": true
  }
}
```

This prompts for selection every time, even though there's a default.

## Multi-Value Variables

Variables with multiple options and aliases.

### Structure

```json
{
  "variables": {
    "environment": {
      "options": [
        "http://localhost:3000",
        "https://dev.api.com",
        "https://prod.api.com"
      ],
      "active": 0,
      "description": "API environment",
      "aliases": {
        "local": 0,
        "dev": 1,
        "prod": 2
      }
    }
  }
}
```

### Fields

| Field         | Type   | Description                      |
| ------------- | ------ | -------------------------------- |
| `options`     | array  | Available values                 |
| `active`      | number | Currently active index (0-based) |
| `description` | string | Variable description             |
| `aliases`     | object | Name to index mapping            |

### Using Aliases

CLI mode:

```bash
restcli -e environment=prod api.http
```

Or by value:

```bash
restcli -e environment=https://prod.api.com api.http
```

### TUI Editor

Press `v` to open variable editor.

For multi-value variables:

| Key | Action            |
| --- | ----------------- |
| `s` | Set active option |
| `l` | List all values   |
| `L` | Set by alias      |
| `e` | Edit variable     |

## Priority

Variables resolve in this order:

1. CLI flags (`-e`)
2. Request file (`variables` field)
3. Profile (`.profiles.json`)
4. Session (`.session.json`)

Higher priority overwrites lower.

## Examples

### Basic Substitution

Request:

```text
GET {{baseUrl}}/users/{{userId}}
```

Variables:

```json
{
  "baseUrl": "https://api.example.com",
  "userId": "123"
}
```

Result:

```text
GET https://api.example.com/users/123
```

### Shell Command

Request:

```text
POST https://api.example.com/events
Content-Type: application/json

{
  "timestamp": {{timestamp}},
  "branch": "{{branch}}"
}
```

Variables:

```json
{
  "timestamp": "$(date +%s)",
  "branch": "$(git branch --show-current)"
}
```

Executed command output replaces variables.

### Multi-Value

Profile:

```json
{
  "name": "API",
  "variables": {
    "apiVersion": {
      "options": ["v1", "v2", "v3"],
      "active": 1,
      "aliases": {
        "legacy": 0,
        "current": 1,
        "beta": 2
      }
    }
  }
}
```

Request:

```text
GET https://api.example.com/{{apiVersion}}/users
```

Default uses index 1: `v2`

Override with alias:

```bash
restcli -e apiVersion=beta api.http
```

Result:

```text
GET https://api.example.com/v3/users
```

### Environment Variables

Request:

```text
GET https://api.example.com/data
Authorization: Bearer {{env.API_TOKEN}}
X-User-ID: {{env.USER_ID}}
```

Shell:

```bash
export API_TOKEN=abc123
export USER_ID=user-456
restcli request.http
```

Or with file:

```bash
restcli --env-file .env request.http
```

## Missing Variables

CLI mode prompts for missing variables:

```bash
restcli get-user.http
# Prompts: Enter value for userId:
```

TUI mode shows error and highlights missing variables.

Provide all variables via flags or profiles to avoid prompts in scripts.
