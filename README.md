# REST CLI

Keyboard-driven TUI for testing HTTP endpoints with Vim-style navigation.

## Why?

Manage API calls using directory and file structure instead of massive JSON files.

No more merge conflicts. No more outdated collections. Each endpoint is a file. Git-friendly. Team-friendly.

## Features

### Core
1. **File-based requests** (`.http`, `.yaml`, `.json`, `.jsonc`, `openapi`) - one endpoint per file
2. **Variable substitution** with `{{varName}}` and shell commands `$(cmd)`
3. **Multi-value variables** with aliases (e.g., `dev`, `staging`, `prod`)
4. **Profile system** for environment-specific headers and variables
5. **Request history** with split view and live preview

### Execution & Control
6. **Streaming support** for SSE and real-time responses
7. **GraphQL & HTTP protocols** with automatic detection
8. **Request cancellation** (ESC to abort in-progress requests)
9. **Confirmation modals** for critical endpoints
10. **Concurrent request blocking** prevents accidental request

### Security & Auth
11. **OAuth 2.0** with PKCE flow and token auto-extraction
12. **mTLS support** with client certificates
13. **Variable interpolation in TLS paths** for dynamic cert loading

### Analysis & Debugging
14. **Response filtering** with JMESPath or bash commands
15. **Response pinning and diff** for regression testing
16. **Error detail modals** with full stack traces
17. **Embedded documentation** viewer with collapsible trees

### Automation
18. **CLI mode** for scripting (JSON/YAML output)
19. **cURL converter** (convert cURL to request files)
20. **OpenAPI converter** (generate requests from specs)

## Installation

### Pre-built binaries

Download from releases or use the binary in `bin/`.

### From source

```bash
git clone https://github.com/studiowebux/restcli
cd restcli/src
go build -o ../bin/restcli ./cmd/restcli
mv ../bin/restcli /usr/local/bin/
```

## Quick Start

### 1. Create a request file

`get-user.http`:
```text
### Get User
GET https://jsonplaceholder.typicode.com/users/1
```

### 2. Run in TUI mode

```bash
restcli
```

Navigate with `j`/`k`, press `Enter` to execute, `ESC` to cancel.

### 3. Or use CLI mode for automation

```bash
restcli run get-user.http
restcli run get-user.http --json  # Output as JSON
```

### 4. Add variables and profiles

```text
### Get User
GET {{baseUrl}}/users/{{userId}}
Authorization: Bearer {{token}}
```

Create `.profiles.json`:
```json
{
  "profiles": [
    {
      "name": "dev",
      "variables": {
        "baseUrl": "https://api.dev.example.com",
        "userId": "1"
      }
    }
  ]
}
```

Run with profile:
```bash
restcli -p dev
```

## Key Shortcuts

Full reference: Press `?` in-app or see [docs/reference/keyboard-shortcuts.md](docs/reference/keyboard-shortcuts.md)

## Documentation

Complete documentation at [docs/](docs/)

1. [Installation](docs/getting-started/installation.md)
2. [Quick Start](docs/getting-started/quick-start.md)
3. [TUI Mode](docs/guides/tui-mode.md)
4. [CLI Mode](docs/guides/cli-mode.md)
5. [File Formats](docs/guides/file-formats.md)
6. [Variables](docs/guides/variables.md)
7. [Profiles](docs/guides/profiles.md)
8. [Authentication](docs/guides/authentication.md)
9. [Filtering & Querying](docs/guides/filtering.md)
10. [Examples](docs/examples.md)

## License

See [LICENSE](LICENSE)
