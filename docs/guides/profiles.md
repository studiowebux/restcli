---
title: Profiles
tags:
  - guide
---

# Profiles

Profiles store persistent configuration for different environments or users.

## Configuration File

Create `.profiles.json` in your project directory:

```json
[
  {
    "name": "Development",
    "headers": {
      "Authorization": "Bearer {{token}}",
      "X-Environment": "dev"
    },
    "variables": {
      "baseUrl": "https://dev.api.example.com",
      "token": "dev-token-123"
    },
    "workdir": "",
    "editor": "vim",
    "output": "json"
  },
  {
    "name": "Production",
    "headers": {
      "Authorization": "Bearer {{token}}",
      "X-Environment": "prod"
    },
    "variables": {
      "baseUrl": "https://api.example.com",
      "token": "prod-token-456"
    },
    "workdir": "",
    "editor": "vim",
    "output": "json"
  }
]
```

## Fields

### name (required)

Profile identifier.

```json
{
  "name": "Development"
}
```

### headers (optional)

Default headers for all requests.

```json
{
  "headers": {
    "Authorization": "Bearer {{token}}",
    "Content-Type": "application/json",
    "X-User-ID": "{{userId}}"
  }
}
```

Headers support variable substitution.

### variables (optional)

Profile variables.

Simple:

```json
{
  "variables": {
    "baseUrl": "https://api.example.com",
    "userId": "123"
  }
}
```

Multi-value:

```json
{
  "variables": {
    "environment": {
      "options": ["dev", "staging", "prod"],
      "active": 0,
      "aliases": {
        "d": 0,
        "s": 1,
        "p": 2
      }
    }
  }
}
```

Shell commands:

```json
{
  "variables": {
    "timestamp": "$(date +%s)",
    "branch": "$(git branch --show-current)"
  }
}
```

### workdir

Working directory for requests.

```json
{
  "workdir": "/path/to/requests"
}
```

If set, TUI uses this directory for file operations.

### editor

External editor command.

```json
{
  "editor": "vim"
}
```

Or:

```json
{
  "editor": "code -w"
}
```

Used when pressing `x` in TUI.

### output

Default output format for CLI mode.

```json
{
  "output": "json"
}
```

Options: `json`, `yaml`, `text`

### oauth (optional)

OAuth 2.0 configuration.

```json
{
  "oauth": {
    "authUrl": "https://auth.example.com/authorize",
    "tokenUrl": "https://auth.example.com/token",
    "clientId": "your-client-id",
    "scope": "read write",
    "redirectUrl": "http://localhost:8080/callback"
  }
}
```

See [authentication guide](authentication.md) for details.

### defaultFilter (optional)

Default JMESPath filter for all requests.

```json
{
  "defaultFilter": "items[?active==`true`]"
}
```

### defaultQuery (optional)

Default JMESPath query or bash command.

```json
{
  "defaultQuery": "[].{id: id, name: name}"
}
```

Or bash:

```json
{
  "defaultQuery": "$(jq '.items[].name')"
}
```

### tls (optional)

Default TLS configuration.

```json
{
  "tls": {
    "certFile": "/path/to/client.crt",
    "keyFile": "/path/to/client.key",
    "caFile": "/path/to/ca.crt",
    "insecureSkipVerify": false
  }
}
```

See [authentication guide](authentication.md) for mTLS details.

## Using Profiles

### CLI Mode

Select profile with `-p` flag:

```bash
restcli -p Development request.http
```

### TUI Mode

Press `p` to open profile switcher.

Select profile from list.

Press `e` to edit profile configuration.

Press `d` to duplicate the selected profile (copies all settings including workdir, headers, variables, and OAuth).

Press `D` to delete the selected profile (requires confirmation; cannot delete active or last profile).

## Sessions

Session data in `.session.json` tracks ephemeral state.

### Structure

```json
{
  "activeProfile": "Development",
  "variables": {
    "token": "auto-extracted-token",
    "refreshToken": "auto-extracted-refresh"
  }
}
```

### Auto-extraction

TUI automatically extracts `token` or `accessToken` from JSON responses.

Stored in session, available as `{{token}}` in requests.

### Profile Linking

Session clears when switching profiles.

Each profile has independent session state.

### Important

Do not manually edit `.session.json` for configuration.

Use `.profiles.json` for persistent settings.

Session is for runtime state only.

## Examples

### Multiple Environments

```json
[
  {
    "name": "Local",
    "variables": {
      "baseUrl": "http://localhost:3000"
    }
  },
  {
    "name": "Development",
    "variables": {
      "baseUrl": "https://dev.api.example.com"
    }
  },
  {
    "name": "Staging",
    "variables": {
      "baseUrl": "https://staging.api.example.com"
    }
  },
  {
    "name": "Production",
    "variables": {
      "baseUrl": "https://api.example.com"
    }
  }
]
```

### Multiple Users

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
      "token": "user1-token"
    }
  },
  {
    "name": "User 2",
    "headers": {
      "Authorization": "Bearer {{token}}",
      "X-User-ID": "user2"
    },
    "variables": {
      "baseUrl": "http://localhost:3000",
      "token": "user2-token"
    }
  }
]
```

### With OAuth and Filters

```json
[
  {
    "name": "API Client",
    "headers": {
      "Content-Type": "application/json"
    },
    "variables": {
      "baseUrl": "https://api.example.com"
    },
    "oauth": {
      "authUrl": "https://auth.example.com/authorize",
      "tokenUrl": "https://auth.example.com/token",
      "clientId": "client-123",
      "scope": "api.read api.write"
    },
    "defaultFilter": "results[?status==`active`]",
    "defaultQuery": "[].{id: id, name: name, created: createdAt}"
  }
]
```

## Priority

Profile settings apply in this order:

1. Request file settings (highest)
2. CLI flags
3. Profile settings
4. Session data (lowest)

Request-specific settings override profile defaults.
