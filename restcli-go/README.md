# REST CLI - Go

## Installation

### From Source

```bash
git clone https://github.com/studiowebux/restcli-go
cd restcli-go
go build -o ./bin/restcli ./cmd/restcli
```

### Pre-built Binaries

Download from the [Releases](https://github.com/studiowebux/restcli-go/releases) page.

## Quick Start

### TUI Mode

```bash
# Start interactive TUI
restcli
```

### CLI Mode

```bash
# Execute a request
restcli request.http

# With profile
restcli -p Dev request.http

# Output as JSON
restcli -o json request.http

# Pipe body from stdin
cat payload.json | restcli request.http

# Override body
restcli -b '{"key":"value"}' request.http

# Save response
restcli -s response.json request.http

# Full output with headers
restcli -f request.http
```

### Converters

```bash
# Convert cURL to .http
curl https://api.example.com | pbpaste | restcli curl2http -o request.http

# From OpenAPI spec
restcli openapi2http swagger.json -o requests/

# Organize by tags
restcli openapi2http api.yaml --organize-by tags

# Organize by paths
restcli openapi2http api.yaml --organize-by paths
```

## Development

### Build

```bash
go build -o ./bin/restcli ./cmd/restcli
```

## License

MIT

## Acknowledgments

- Original TypeScript version: [@studiowebux/restcli](https://github.com/studiowebux/restcli)
- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea)
