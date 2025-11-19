# curl2http - Convert cURL to .http Files

Quickly convert cURL commands (from browser DevTools, documentation, etc.) into `.http` request files for the TUI.

## Quick Start

### Using Compiled Binary (Recommended)

**From Clipboard (macOS):**

```bash
pbpaste | restcli-curl2http --output requests/my-request.http
```

**From Command Line:**

```bash
restcli-curl2http --output requests/login.http 'curl -X POST http://localhost:3000/auth/login -H "Content-Type: application/json" -d '"'"'{"username":"test","password":"pass"}'"'"''
```

## Features

### Automatic Parsing

- Extracts HTTP method (`-X`, `--request`)
- Extracts URL
- Extracts all headers (`-H`, `--header`)
- Extracts request body (`-d`, `--data`, `--data-raw`)
- Handles multiline curl commands (with `\`)

### Smart Variable Detection

Automatically detects and suggests variables:

**Input:**

```bash
curl http://localhost:3000/api/users -H "Authorization: Bearer eyJhbGc..."
```

**Output:**

```text
### GET api/users
GET {{baseUrl}}/api/users
Authorization: Bearer eyJhbGc...
```

**Detected variables:**

```json
{
  "baseUrl": "http://localhost:3000"
}
```

### Security Header Filtering

By default, sensitive headers are excluded from generated `.http` files for security:

**Filtered headers:**

- `Authorization`
- `Cookie`
- `X-API-Key`
- `X-Auth-Token`
- `API-Key`
- `Auth-Token`
- `Bearer`
- `X-Session-Token`
- `X-CSRF-Token`

**Example:**

```bash
curl http://localhost:3000/api/users -H "Authorization: Bearer secret123"
```

**Output:**

```text
### GET api/users
GET {{baseUrl}}/api/users
```

```bash
🔒 Excluded sensitive headers (use --import-headers to include):
  Authorization: Bearer secret123

💡 Add these to your profile headers in .profiles.json instead
```

**To include sensitive headers:**

```bash
pbpaste | restcli-curl2http --output requests/file.http --import-headers
```

### Output File Specification

Use the `--output` (or `-o`) flag to specify where to save the `.http` file:

```bash
# Save to specific file
pbpaste | restcli-curl2http --output requests/login.http

# Save to directory (auto-generates filename)
pbpaste | restcli-curl2http --output requests/

# Short form
pbpaste | restcli-curl2http -o requests/my-request.http
```

**Auto-generated filenames** are based on the URL path:

| URL                 | Generated Filename |
| ------------------- | ------------------ |
| `/auth/login`       | `post-login.http`  |
| `/users`            | `users.http`       |
| `/api/v1/posts/123` | `123.http`         |
