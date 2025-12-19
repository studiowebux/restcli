# REST CLI

Keyboard-driven TUI for testing HTTP endpoints with Vim-style navigation.

## Breaking Changes (V0.0.28)

**Database Schema Update**: The database schema has been updated to include `profile_name` across all tables (history, analytics, stress test). Existing databases from previous versions may encounter migration warnings. For a clean start, remove old database files from `~/.restcli/data/` before upgrading.

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
6. **Categories and filtering** - organize requests with categories, filter by category in TUI

### Execution & Control
7. **Request chaining** with dependency resolution and automatic variable extraction
8. **Streaming support** for SSE and real-time responses
9. **GraphQL & HTTP protocols** with automatic detection
10. **Request cancellation** (ESC to abort in-progress requests)
11. **Confirmation modals** for critical endpoints
12. **Concurrent request blocking** prevents accidental request

### Security & Auth
13. **OAuth 2.0** with PKCE flow and token auto-extraction
14. **mTLS support** with client certificates
15. **Variable interpolation in TLS paths** for dynamic cert loading

### Analysis & Debugging
16. **Inline filter editor** with bookmark system - filter responses while viewing JSON structure
17. **Response filtering** with JMESPath or bash commands
18. **Response pinning and diff** for regression testing
19. **Error detail modals** with full stack traces
20. **Embedded documentation** viewer with collapsible trees
21. **Analytics tracking** with per-endpoint stats and aggregated metrics
22. **Request history** with persistent storage and search
23. **HTTP method color coding** in file list for quick identification

### Performance Testing
24. **Stress testing** with configurable concurrency and load
25. **Ramp-up control** for gradual load increase
26. **Real-time metrics** (latency, RPS, percentiles P50/P95/P99)
27. **Test result persistence** with historical comparison
28. **One-click re-run** for saved test configurations

### Automation
29. **CLI mode** for scripting (JSON/YAML output)
30. **cURL converter** (convert cURL to request files)
31. **OpenAPI converter** (generate requests from specs)

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

## Analytics & Performance Testing

### View Analytics

Press `A` in the TUI to view request analytics:
- Per-endpoint statistics (count, avg/min/max latency, error rate)
- Aggregated metrics across all requests
- Historical tracking with timestamps
- Grouped by file or normalized path

### Run Stress Tests

Press `S` (Shift+s) to access stress testing:

1. **Create a new test** - Press `n` to configure:
   - Request file and endpoint to test
   - Concurrent connections (workers)
   - Total requests to send
   - Ramp-up duration (gradual load increase)
   - Test duration limit

2. **Run the test** - Ctrl+S to save config and start
   - Real-time progress with live metrics
   - Latency percentiles (P50, P95, P99)
   - Requests per second (RPS)
   - Success/error counts

3. **View results** - Automatic on completion:
   - Historical test runs with comparison
   - Detailed statistics and metrics
   - Press `r` on any run to re-execute
   - Press `l` to load saved configurations

See [docs/guides/stress-testing.md](docs/guides/stress-testing.md) for detailed usage.

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
8. [Categories](docs/guides/categories.md)
9. [Request Chaining](docs/guides/chaining.md)
10. [Authentication](docs/guides/authentication.md)
11. [Filtering & Querying](docs/guides/filtering.md)
12. [Analytics](docs/guides/analytics.md)
13. [Stress Testing](docs/guides/stress-testing.md)
14. [Examples](docs/examples.md)

## License

See [LICENSE](LICENSE)
