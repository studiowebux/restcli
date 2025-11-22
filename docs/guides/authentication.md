---
title: Authentication
tags:
  - guide
---

# Authentication

REST CLI supports multiple authentication methods.

## Bearer Tokens

Use variables in Authorization header:

```text
### Get Data
GET https://api.example.com/data
Authorization: Bearer {{token}}
```

Set token via profile:

```json
{
  "name": "API",
  "variables": {
    "token": "your-token-here"
  }
}
```

Or CLI:

```bash
restcli -e token=your-token-here request.http
```

### Auto-extraction

TUI automatically extracts `token` or `accessToken` from JSON responses.

Example response:

```json
{
  "token": "abc123",
  "user": {
    "id": 1
  }
}
```

Token stored in session, available as `{{token}}` in subsequent requests.

## OAuth 2.0 with PKCE

### Configuration

Add to profile in `.profiles.json`:

```json
{
  "name": "OAuth Client",
  "oauth": {
    "authUrl": "https://auth.example.com/authorize",
    "tokenUrl": "https://auth.example.com/token",
    "clientId": "your-client-id",
    "clientSecret": "",
    "scope": "read write",
    "redirectUrl": "http://localhost:8080/callback"
  }
}
```

### Fields

| Field          | Required | Description                            |
| -------------- | -------- | -------------------------------------- |
| `authUrl`      | Yes      | Authorization endpoint                 |
| `tokenUrl`     | Yes      | Token endpoint                         |
| `clientId`     | Yes      | OAuth client ID                        |
| `clientSecret` | No       | Client secret (if not using PKCE)      |
| `scope`        | No       | Requested scopes                       |
| `redirectUrl`  | No       | Callback URL (default: localhost:8080) |

### TUI Flow

1. Press `O` to configure OAuth settings
2. Press `o` to start authorization flow
3. Browser opens to authorization page
4. Grant permissions
5. Token auto-stored in session

### Using Token

OAuth token available as `{{token}}`:

```text
### API Call
GET https://api.example.com/data
Authorization: Bearer {{token}}
```

### PKCE Support

PKCE enabled automatically for public clients.

Set `clientSecret` to empty string or omit for PKCE.

## Mutual TLS (mTLS)

Client certificate authentication.

### Profile Configuration

In `.profiles.json`:

```json
{
  "name": "Secure API",
  "tls": {
    "certFile": "/path/to/client.crt",
    "keyFile": "/path/to/client.key",
    "caFile": "/path/to/ca.crt",
    "insecureSkipVerify": false
  }
}
```

### Per-Request Configuration

HTTP format:

```text
### Secure Request
# @tls.certFile /path/to/client.crt
# @tls.keyFile /path/to/client.key
# @tls.caFile /path/to/ca.crt
GET https://secure.api.example.com/data
```

YAML format:

```yaml
name: Secure Request
method: GET
url: "https://secure.api.example.com/data"
tls:
  certFile: "/path/to/client.crt"
  keyFile: "/path/to/client.key"
  caFile: "/path/to/ca.crt"
```

JSON format:

```json
{
  "name": "Secure Request",
  "method": "GET",
  "url": "https://secure.api.example.com/data",
  "tls": {
    "certFile": "/path/to/client.crt",
    "keyFile": "/path/to/client.key",
    "caFile": "/path/to/ca.crt"
  }
}
```

### TLS Fields

| Field                | Description                                  |
| -------------------- | -------------------------------------------- |
| `certFile`           | Client certificate (PEM)                     |
| `keyFile`            | Private key (PEM)                            |
| `caFile`             | CA certificate for server verification (PEM) |
| `insecureSkipVerify` | Skip server certificate verification         |

### Certificate Generation

Example with OpenSSL:

```bash
# Generate private key
openssl genrsa -out client.key 2048

# Generate certificate signing request
openssl req -new -key client.key -out client.csr

# Generate self-signed certificate (testing)
openssl x509 -req -days 365 -in client.csr -signkey client.key -out client.crt
```

For production, obtain certificates from your CA.

### Priority

TLS configuration applies in this order:

1. Request file `tls` field (highest)
2. Profile `tls` field (lowest)

Request-specific TLS overrides profile TLS.

## API Keys

Use headers with variables:

```text
### API Call
GET https://api.example.com/data
X-API-Key: {{apiKey}}
```

Profile:

```json
{
  "name": "API",
  "variables": {
    "apiKey": "your-api-key"
  }
}
```

Or environment variable:

```text
X-API-Key: {{env.API_KEY}}
```

## Basic Auth

Construct header manually:

```text
### Basic Auth
GET https://api.example.com/data
Authorization: Basic {{basicAuth}}
```

Generate value:

```bash
echo -n "username:password" | base64
```

Or use shell command variable:

```json
{
  "variables": {
    "basicAuth": "$(echo -n 'username:password' | base64)"
  }
}
```

## Custom Authentication

For custom schemes, use headers:

```text
### Custom Auth
GET https://api.example.com/data
X-Custom-Auth: {{authToken}}
X-Signature: {{signature}}
X-Timestamp: {{timestamp}}
```

With shell command for signature:

```json
{
  "variables": {
    "timestamp": "$(date +%s)",
    "signature": "$(echo -n 'data' | openssl dgst -sha256 -hmac 'secret' | cut -d' ' -f2)"
  }
}
```

## Examples

### OAuth Flow

Profile:

```json
{
  "name": "GitHub",
  "oauth": {
    "authUrl": "https://github.com/login/oauth/authorize",
    "tokenUrl": "https://github.com/login/oauth/access_token",
    "clientId": "your-client-id",
    "scope": "repo user"
  }
}
```

Request:

```text
### List Repos
GET https://api.github.com/user/repos
Authorization: Bearer {{token}}
Accept: application/vnd.github.v3+json
```

### mTLS with Custom CA

Profile:

```json
{
  "name": "Internal API",
  "tls": {
    "certFile": "/etc/ssl/certs/client.crt",
    "keyFile": "/etc/ssl/private/client.key",
    "caFile": "/etc/ssl/certs/internal-ca.crt"
  },
  "variables": {
    "baseUrl": "https://internal.api.example.com"
  }
}
```

Request:

```text
### Internal Data
GET {{baseUrl}}/data
```

### Combined Auth

Profile with both OAuth and API key:

```json
{
  "name": "Multi-Auth",
  "headers": {
    "X-API-Key": "{{apiKey}}"
  },
  "variables": {
    "apiKey": "static-key-123"
  },
  "oauth": {
    "authUrl": "https://auth.example.com/authorize",
    "tokenUrl": "https://auth.example.com/token",
    "clientId": "client-id"
  }
}
```

Request uses both:

```text
### Secure Endpoint
GET https://api.example.com/secure-data
Authorization: Bearer {{token}}
```

API key from profile headers, OAuth token from session.
