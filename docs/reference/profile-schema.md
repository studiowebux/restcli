# Profile Schema Reference

Complete schema for `.profiles.json`.

## Root Structure

Array of profile objects:

```json
[
  {
    /* Profile 1 */
  },
  {
    /* Profile 2 */
  }
]
```

## Profile Object

### Required Fields

| Field  | Type   | Description        |
| ------ | ------ | ------------------ |
| `name` | string | Profile identifier |

### Optional Fields

| Field           | Type        | Description                       |
| --------------- | ----------- | --------------------------------- |
| `headers`       | object      | Default headers                   |
| `variables`     | object      | Variables (simple or multi-value) |
| `workdir`       | string      | Working directory                 |
| `editor`        | string      | External editor command           |
| `output`        | string      | Default output format             |
| `oauth`         | OAuthConfig | OAuth configuration               |
| `defaultFilter` | string      | Default JMESPath filter           |
| `defaultQuery`  | string      | Default query                     |
| `tls`           | TLSConfig   | Default TLS configuration         |

## name (required)

Profile identifier.

```json
{
  "name": "Development"
}
```

## headers (optional)

Default headers for all requests in this profile.

```json
{
  "headers": {
    "Authorization": "Bearer {{token}}",
    "Content-Type": "application/json",
    "X-Environment": "dev",
    "X-User-ID": "{{userId}}"
  }
}
```

Headers support variable substitution.

## variables (optional)

Profile variables. Can be simple strings or multi-value objects.

### Simple Variables

```json
{
  "variables": {
    "baseUrl": "https://api.example.com",
    "userId": "123",
    "token": "abc"
  }
}
```

### Multi-Value Variables

```json
{
  "variables": {
    "environment": {
      "options": ["dev", "staging", "prod"],
      "active": 0,
      "description": "API environment",
      "aliases": {
        "d": 0,
        "s": 1,
        "p": 2
      }
    }
  }
}
```

### Shell Command Variables

```json
{
  "variables": {
    "timestamp": "$(date +%s)",
    "branch": "$(git branch --show-current)",
    "uuid": "$(uuidgen)"
  }
}
```

### Mixed Variables

```json
{
  "variables": {
    "baseUrl": "https://api.example.com",
    "timestamp": "$(date +%s)",
    "environment": {
      "options": ["dev", "prod"],
      "active": 0,
      "aliases": { "d": 0, "p": 1 }
    }
  }
}
```

## workdir

Working directory for file operations in TUI.

```json
{
  "workdir": "/path/to/requests"
}
```

If empty or omitted, uses current directory.

## editor

External editor command for editing files.

```json
{
  "editor": "vim"
}
```

Or with flags:

```json
{
  "editor": "code -w"
}
```

Used when pressing `x` in TUI.

## output

Default output format for CLI mode.

```json
{
  "output": "json"
}
```

Values: `json`, `yaml`, `text`

## oauth (optional)

OAuth 2.0 configuration.

### OAuthConfig Fields

| Field          | Type   | Required | Description            |
| -------------- | ------ | -------- | ---------------------- |
| `authUrl`      | string | Yes      | Authorization endpoint |
| `tokenUrl`     | string | Yes      | Token endpoint         |
| `clientId`     | string | Yes      | OAuth client ID        |
| `clientSecret` | string | No       | Client secret          |
| `scope`        | string | No       | Requested scopes       |
| `redirectUrl`  | string | No       | Callback URL           |

### Example

```json
{
  "oauth": {
    "authUrl": "https://auth.example.com/authorize",
    "tokenUrl": "https://auth.example.com/token",
    "clientId": "client-123",
    "clientSecret": "",
    "scope": "read write",
    "redirectUrl": "http://localhost:8080/callback"
  }
}
```

Empty `clientSecret` enables PKCE.

## defaultFilter (optional)

Default JMESPath filter applied to all responses.

```json
{
  "defaultFilter": "results[?status==`active`]"
}
```

Can be overridden per request.

## defaultQuery (optional)

Default JMESPath query or bash command.

JMESPath:

```json
{
  "defaultQuery": "[].{id: id, name: name}"
}
```

Bash command:

```json
{
  "defaultQuery": "$(jq '.items[].name')"
}
```

Can be overridden per request.

## tls (optional)

Default TLS/mTLS configuration.

### TLSConfig Fields

| Field                | Type    | Default | Description                 |
| -------------------- | ------- | ------- | --------------------------- |
| `certFile`           | string  |         | Client certificate path     |
| `keyFile`            | string  |         | Private key path            |
| `caFile`             | string  |         | CA certificate path         |
| `insecureSkipVerify` | boolean | false   | Skip verification (testing) |

### Example

```json
{
  "tls": {
    "certFile": "/etc/ssl/certs/client.crt",
    "keyFile": "/etc/ssl/private/client.key",
    "caFile": "/etc/ssl/certs/ca.crt",
    "insecureSkipVerify": false
  }
}
```

## Multi-Value Variable Schema

### Fields

| Field         | Type   | Required | Description            |
| ------------- | ------ | -------- | ---------------------- |
| `options`     | array  | Yes      | Available values       |
| `active`      | number | Yes      | Active index (0-based) |
| `description` | string | No       | Variable description   |
| `aliases`     | object | No       | Alias to index mapping |

### Example

```json
{
  "apiVersion": {
    "options": ["v1", "v2", "v3"],
    "active": 1,
    "description": "API version to use",
    "aliases": {
      "legacy": 0,
      "current": 1,
      "beta": 2
    }
  }
}
```

## Complete Examples

### Development Profile

```json
{
  "name": "Development",
  "headers": {
    "Authorization": "Bearer {{token}}",
    "X-Environment": "dev"
  },
  "variables": {
    "baseUrl": "https://dev.api.example.com",
    "token": "dev-token-123",
    "userId": "1"
  },
  "workdir": "./requests",
  "editor": "vim",
  "output": "json",
  "defaultFilter": "",
  "defaultQuery": ""
}
```

### Production Profile with OAuth

```json
{
  "name": "Production",
  "headers": {
    "Content-Type": "application/json"
  },
  "variables": {
    "baseUrl": "https://api.example.com"
  },
  "editor": "code -w",
  "output": "json",
  "oauth": {
    "authUrl": "https://auth.example.com/oauth/authorize",
    "tokenUrl": "https://auth.example.com/oauth/token",
    "clientId": "prod-client-id",
    "scope": "api.read api.write",
    "redirectUrl": "http://localhost:8080/callback"
  },
  "defaultFilter": "results",
  "defaultQuery": "[].{id: id, status: status}"
}
```

### Multi-Environment Profile

```json
{
  "name": "Multi-Env",
  "headers": {
    "Authorization": "Bearer {{token}}"
  },
  "variables": {
    "baseUrl": {
      "options": [
        "http://localhost:3000",
        "https://dev.api.example.com",
        "https://staging.api.example.com",
        "https://api.example.com"
      ],
      "active": 0,
      "description": "API environment",
      "aliases": {
        "local": 0,
        "dev": 1,
        "staging": 2,
        "prod": 3
      }
    },
    "apiVersion": {
      "options": ["v1", "v2"],
      "active": 1,
      "aliases": {
        "stable": 0,
        "latest": 1
      }
    },
    "timestamp": "$(date +%s)"
  },
  "editor": "vim",
  "output": "json"
}
```

### Secure Profile with mTLS

```json
{
  "name": "Secure API",
  "headers": {
    "Content-Type": "application/json"
  },
  "variables": {
    "baseUrl": "https://secure.api.example.com"
  },
  "tls": {
    "certFile": "/etc/ssl/certs/client.crt",
    "keyFile": "/etc/ssl/private/client.key",
    "caFile": "/etc/ssl/certs/internal-ca.crt",
    "insecureSkipVerify": false
  },
  "output": "json",
  "defaultFilter": "data",
  "defaultQuery": "{id: id, status: status, timestamp: createdAt}"
}
```

### Complete Multi-Profile File

```json
[
  {
    "name": "Local",
    "variables": {
      "baseUrl": "http://localhost:3000"
    },
    "output": "text"
  },
  {
    "name": "Development",
    "headers": {
      "Authorization": "Bearer {{token}}"
    },
    "variables": {
      "baseUrl": "https://dev.api.example.com",
      "token": "dev-token"
    },
    "oauth": {
      "authUrl": "https://dev-auth.example.com/authorize",
      "tokenUrl": "https://dev-auth.example.com/token",
      "clientId": "dev-client-id"
    }
  },
  {
    "name": "Production",
    "headers": {
      "Authorization": "Bearer {{token}}"
    },
    "variables": {
      "baseUrl": "https://api.example.com"
    },
    "oauth": {
      "authUrl": "https://auth.example.com/authorize",
      "tokenUrl": "https://auth.example.com/token",
      "clientId": "prod-client-id",
      "scope": "api.read api.write"
    },
    "tls": {
      "certFile": "/etc/ssl/certs/prod-client.crt",
      "keyFile": "/etc/ssl/private/prod-client.key"
    },
    "output": "json",
    "defaultQuery": "[].{id: id, name: name}"
  }
]
```

## Session File

`.session.json` stores ephemeral state (not for manual editing).

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

### Fields

| Field           | Type   | Description                        |
| --------------- | ------ | ---------------------------------- |
| `activeProfile` | string | Currently active profile name      |
| `variables`     | object | Runtime variables (auto-extracted) |

Session clears when switching profiles.

Use profiles for persistent configuration, not session.
