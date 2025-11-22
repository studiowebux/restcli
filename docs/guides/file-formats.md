# File Formats

REST CLI supports multiple request file formats.

Each file contains one request. For multiple endpoints, create separate files.

## HTTP Format (.http)

### Basic Structure

```text
### Request Name (optional)
METHOD url
Header: value
Another-Header: value

{
  "body": "for POST/PUT/PATCH"
}
```

### Example

One request per file.

`get-user.http`:

```text
### Get User
GET https://api.example.com/users/1
Accept: application/json
```

`create-user.http`:

```text
### Create User
POST https://api.example.com/users
Content-Type: application/json
Authorization: Bearer {{token}}

{
  "name": "John Doe",
  "email": "john@example.com"
}
```

### Comments

Use `#` for comments:

```text
# This is a comment
### Get User
# @filter users[?active]
# @query [].name
GET https://api.example.com/users
```

### Special Directives

| Directive                   | Purpose                            |
| --------------------------- | ---------------------------------- |
| `# @filter`                 | JMESPath filter or bash command    |
| `# @query`                  | JMESPath query or bash command     |
| `# @parsing`                | Parsing mode for response body     |
| `# @tls.certFile`           | Client certificate path            |
| `# @tls.keyFile`            | Private key path                   |
| `# @tls.caFile`             | CA certificate path                |
| `# @tls.insecureSkipVerify` | Skip TLS verification (true/false) |

## YAML Format (.yaml)

Structured format with full control.

### Single Request

```yaml
name: Get User
method: GET
url: "https://api.example.com/users/{{userId}}"
headers:
  Accept: "application/json"
  Authorization: "Bearer {{token}}"
```

### With Body

```yaml
name: Create User
method: POST
url: "https://api.example.com/users"
headers:
  Content-Type: "application/json"
body: |
  {
    "name": "John Doe",
    "email": "john@example.com"
  }
```

### With Filter and Query

```yaml
name: List Active Users
method: GET
url: "https://api.example.com/users"
filter: "users[?active==`true`]"
query: "[].{name: name, email: email}"
```

## JSON Format (.json)

Structured format with schema validation.

### Single Request

```json
{
  "name": "Get User",
  "method": "GET",
  "url": "https://api.example.com/users/1",
  "headers": {
    "Accept": "application/json"
  }
}
```

### With Schema Reference

```json
{
  "$schema": "../http-request.schema.json",
  "name": "Create User",
  "method": "POST",
  "url": "https://api.example.com/users",
  "headers": {
    "Content-Type": "application/json"
  },
  "body": "{\"name\": \"John\"}"
}
```

## JSONC Format (.jsonc)

JSON with comments support.

### Example

```jsonc
{
  // Single-line comments
  "name": "Get Post",
  "method": "GET",
  "url": "https://api.example.com/posts/1",

  /* Multi-line
   * comments
   */
  "headers": {
    "Accept": "application/json",
  },
}
```

## OpenAPI Format

REST CLI can use OpenAPI spec directly (per endpoint).

### Example

```bash
restcli api.yaml
```

See [OpenAPI converter](../converters/openapi2http.md) for conversion to request files.

## Field Reference

### Required Fields

| Field    | Type   | Description                      |
| -------- | ------ | -------------------------------- |
| `method` | string | HTTP method                      |
| `url`    | string | Request URL (supports variables) |

### Optional Fields

| Field           | Type   | Description                    |
| --------------- | ------ | ------------------------------ |
| `name`          | string | Request name                   |
| `headers`       | object | HTTP headers                   |
| `body`          | string | Request body (POST/PUT/PATCH)  |
| `filter`        | string | JMESPath filter                |
| `query`         | string | JMESPath query or bash command |
| `tls`           | object | TLS configuration              |
| `documentation` | object | Embedded documentation         |

### TLS Object

```json
{
  "certFile": "/path/to/cert.pem",
  "keyFile": "/path/to/key.pem",
  "caFile": "/path/to/ca.pem",
  "insecureSkipVerify": false
}
```

### Documentation Object

See [documentation guide](authentication.md) for structure.

## Variable Substitution

All formats support `{{varName}}` syntax:

```text
GET {{baseUrl}}/users/{{userId}}
Authorization: Bearer {{token}}
```

## Shell Commands

All formats support `$(command)` syntax:

```yaml
url: "https://api.example.com/data?timestamp={{timestamp}}"
```

With variable:

```json
{
  "variables": {
    "timestamp": "$(date +%s)"
  }
}
```

## Format Selection

Choose based on preference:

1. `.http`: Quick, minimal, easy to read
2. `.yaml`: Structured, good for complex requests
3. `.json`: Strict schema, IDE validation
4. `.jsonc`: JSON with comment support
