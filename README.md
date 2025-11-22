# REST CLI

Keyboard-driven TUI for testing HTTP endpoints.

## Why?

Manage API calls using directory and file structure instead of massive JSON files.

No more merge conflicts. No more outdated collections. Each endpoint is a file. Git-friendly. Team-friendly.

## Features

1. File-based requests (`.http`, `.yaml`, `.json`, `.jsonc`)
2. Variable substitution with `{{varName}}` and shell commands `$(cmd)`
3. Multi-value variables with aliases
4. Profile system for header and variable management
5. OAuth 2.0 with PKCE and mTLS support
6. Response filtering with JMESPath or bash/Linux commands
7. Response pinning and diff for regression testing
8. Request history and embedded documentation
9. CLI mode for scripting (JSON/YAML output)
10. cURL and OpenAPI converters

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

Create a request file `get-user.http`:

```text
### Get User
GET https://jsonplaceholder.typicode.com/users/1
```

Run in TUI mode:

```bash
restcli
```

Run in CLI mode:

```bash
restcli run get-user.http
```

## Key Shortcuts

Full shortcuts reference: `?` in-app or `restcli --help`

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
