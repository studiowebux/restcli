# Request Schema Reference

Complete schema for request files (`.yaml`, `.json`, `.jsonc` formats).

## Root Structure

Request files contain a single request object.

## SingleRequest

### Required Fields

| Field    | Type   | Description                      |
| -------- | ------ | -------------------------------- |
| `method` | string | HTTP method                      |
| `url`    | string | Request URL (supports variables) |

### Optional Fields

| Field           | Type          | Description                     |
| --------------- | ------------- | ------------------------------- |
| `name`          | string        | Request name                    |
| `headers`       | object        | HTTP headers                    |
| `body`          | string        | Request body                    |
| `filter`        | string        | JMESPath filter or bash command |
| `query`         | string        | JMESPath query or bash command  |
| `tls`           | TLSConfig     | TLS configuration               |
| `documentation` | Documentation | Embedded documentation          |

### method

HTTP method.

Values: `GET`, `POST`, `PUT`, `DELETE`, `PATCH`, `HEAD`, `OPTIONS`

```json
{
  "method": "GET"
}
```

### url

Request URL. Supports variable substitution.

```json
{
  "url": "https://api.example.com/users/{{userId}}"
}
```

### name (optional)

Descriptive name for the request.

```json
{
  "name": "Get User Details"
}
```

### headers (optional)

HTTP headers as key-value pairs.

```json
{
  "headers": {
    "Authorization": "Bearer {{token}}",
    "Content-Type": "application/json",
    "Accept": "application/json"
  }
}
```

### body (optional)

Request body for `POST`, `PUT`, `PATCH`.

```json
{
  "body": "{\"name\": \"John\", \"email\": \"john@example.com\"}"
}
```

YAML supports multi-line strings:

```yaml
body: |
  {
    "name": "John",
    "email": "john@example.com"
  }
```

### filter (optional)

JMESPath expression to filter response.

```json
{
  "filter": "users[?active==`true`]"
}
```

### query (optional)

JMESPath query or bash command to transform response.

JMESPath:

```json
{
  "query": "[].{id: id, name: name}"
}
```

Bash command:

```json
{
  "query": "$(jq '.items[].name')"
}
```

### tls (optional)

TLS/mTLS configuration. See TLSConfig below.

### documentation (optional)

Embedded API documentation. See Documentation below.

## TLSConfig

### Fields

| Field                | Type    | Default | Description                             |
| -------------------- | ------- | ------- | --------------------------------------- |
| `certFile`           | string  |         | Client certificate path (PEM)           |
| `keyFile`            | string  |         | Private key path (PEM)                  |
| `caFile`             | string  |         | CA certificate path (PEM)               |
| `insecureSkipVerify` | boolean | false   | Skip server verification (testing only) |

### Example

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

## Documentation

### Fields

| Field         | Type   | Description             |
| ------------- | ------ | ----------------------- |
| `description` | string | Endpoint description    |
| `tags`        | array  | Tags/categories         |
| `parameters`  | array  | Request parameters      |
| `responses`   | array  | Response specifications |

### Example

```json
{
  "documentation": {
    "description": "Create a new user account",
    "tags": ["Users", "Authentication"],
    "parameters": [
      {
        "name": "name",
        "type": "string",
        "required": true,
        "description": "User's full name",
        "example": "John Doe"
      }
    ],
    "responses": [
      {
        "code": "201",
        "description": "User created successfully",
        "contentType": "application/json",
        "fields": [
          {
            "name": "id",
            "type": "number",
            "required": true,
            "description": "User ID"
          }
        ],
        "example": "{\"id\": 1, \"name\": \"John Doe\"}"
      }
    ]
  }
}
```

## Parameter

Request parameter documentation.

### Required Fields

| Field  | Type   | Description    |
| ------ | ------ | -------------- |
| `name` | string | Parameter name |
| `type` | string | Parameter type |

### Optional Fields

| Field         | Type    | Default | Description   |
| ------------- | ------- | ------- | ------------- |
| `required`    | boolean | false   | Is required   |
| `deprecated`  | boolean | false   | Is deprecated |
| `description` | string  |         | Description   |
| `example`     | string  |         | Example value |

### Example

```json
{
  "name": "userId",
  "type": "number",
  "required": true,
  "deprecated": false,
  "description": "Unique user identifier",
  "example": "123"
}
```

## Response

Response documentation.

### Required Fields

| Field         | Type   | Description                                 |
| ------------- | ------ | ------------------------------------------- |
| `code`        | string | HTTP status code (pattern: `[1-5][0-9]{2}`) |
| `description` | string | Response description                        |

### Optional Fields

| Field         | Type   | Description               |
| ------------- | ------ | ------------------------- |
| `contentType` | string | Content type              |
| `fields`      | array  | Response field schemas    |
| `example`     | string | Complete response example |

### Example

```json
{
  "code": "200",
  "description": "Success",
  "contentType": "application/json",
  "fields": [
    {
      "name": "id",
      "type": "number",
      "required": true,
      "description": "User ID"
    },
    {
      "name": "name",
      "type": "string",
      "required": true,
      "description": "User's name"
    }
  ],
  "example": "{\"id\": 1, \"name\": \"John\"}"
}
```

### Shorthand

Simple responses can use shorthand:

```json
{
  "responses": [{ "200": "Success" }, { "404": "Not Found" }]
}
```

## ResponseField

Response field documentation.

### Required Fields

| Field  | Type   | Description                        |
| ------ | ------ | ---------------------------------- |
| `name` | string | Field name (supports dot notation) |
| `type` | string | Field type                         |

### Optional Fields

| Field         | Type    | Default | Description   |
| ------------- | ------- | ------- | ------------- |
| `required`    | boolean | false   | Is required   |
| `deprecated`  | boolean | false   | Is deprecated |
| `description` | string  |         | Description   |
| `example`     | string  |         | Example value |

### Example

```json
{
  "name": "user.profile.avatar",
  "type": "string",
  "required": false,
  "deprecated": false,
  "description": "User avatar URL",
  "example": "https://example.com/avatar.jpg"
}
```

Dot notation for nested fields:

```json
{
  "fields": [
    {
      "name": "user",
      "type": "object",
      "required": true
    },
    {
      "name": "user.id",
      "type": "number",
      "required": true
    },
    {
      "name": "user.profile",
      "type": "object"
    },
    {
      "name": "user.profile.name",
      "type": "string"
    }
  ]
}
```

## Complete Example

```json
{
  "$schema": "https://raw.githubusercontent.com/studiowebux/restcli/main/http-request.schema.json",
  "name": "Create User",
  "method": "POST",
  "url": "{{baseUrl}}/users",
  "headers": {
    "Content-Type": "application/json",
    "Authorization": "Bearer {{token}}"
  },
  "body": "{\"name\": \"{{name}}\", \"email\": \"{{email}}\"}",
  "filter": "user",
  "query": "{id: id, name: name, created: createdAt}",
  "tls": {
    "certFile": "/path/to/cert.pem",
    "keyFile": "/path/to/key.pem"
  },
  "documentation": {
    "description": "Create a new user account",
    "tags": ["Users", "POST"],
    "parameters": [
      {
        "name": "name",
        "type": "string",
        "required": true,
        "description": "User's full name",
        "example": "John Doe"
      },
      {
        "name": "email",
        "type": "string",
        "required": true,
        "description": "User's email",
        "example": "john@example.com"
      }
    ],
    "responses": [
      {
        "code": "201",
        "description": "User created successfully",
        "contentType": "application/json",
        "fields": [
          {
            "name": "id",
            "type": "number",
            "required": true,
            "description": "Generated user ID"
          },
          {
            "name": "name",
            "type": "string",
            "required": true,
            "description": "User's name"
          },
          {
            "name": "email",
            "type": "string",
            "required": true,
            "description": "User's email"
          },
          {
            "name": "createdAt",
            "type": "string",
            "required": true,
            "description": "ISO timestamp"
          }
        ],
        "example": "{\"id\": 1, \"name\": \"John Doe\", \"email\": \"john@example.com\", \"createdAt\": \"2025-01-01T00:00:00Z\"}"
      },
      {
        "code": "400",
        "description": "Invalid input",
        "contentType": "application/json",
        "example": "{\"error\": \"Email already exists\"}"
      }
    ]
  }
}
```

## Schema URL

Reference schema in JSON files:

```json
{
  "$schema": "https://raw.githubusercontent.com/studiowebux/restcli/main/http-request.schema.json"
}
```

Or local:

```json
{
  "$schema": "../http-request.schema.json"
}
```

Enables IDE validation and autocomplete.
