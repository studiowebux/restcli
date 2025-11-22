---
title: curl2http Converter
description: Convert cURL commands to REST CLI request files.
tags:
  - converter
---

# curl2http Converter

Convert cURL commands to REST CLI request files.

## Basic Usage

```bash
restcli curl2http 'curl https://api.example.com/users'
```

From clipboard:

```bash
pbpaste | restcli curl2http
```

## Flags

### Output File

```bash
restcli curl2http -o request.http
```

Short: `-o`

### Import Headers

```bash
restcli curl2http --import-headers
```

Converts cURL headers to profile headers format.

### Output Format

```bash
restcli curl2http -f json
```

Options: `http`, `yaml`, `json`, `jsonc`

Default: `http`

Short: `-f`

## Examples

### Simple GET

Copy this cURL command:

```bash
curl https://api.example.com/users
```

Convert from clipboard:

```bash
pbpaste | restcli curl2http -o get-users.http
```

Result:

```text
### Request
GET https://api.example.com/users
```

### With Headers

Copy:

```bash
curl -H "Authorization: Bearer token123" -H "Accept: application/json" https://api.example.com/users
```

Convert:

```bash
pbpaste | restcli curl2http -o get-users.http
```

Result:

```text
### Request
GET https://api.example.com/users
Authorization: Bearer token123
Accept: application/json
```

### POST with Body

Copy:

```bash
curl -X POST https://api.example.com/users -H "Content-Type: application/json" -d '{"name":"John","email":"john@example.com"}'
```

Convert:

```bash
pbpaste | restcli curl2http -o create-user.http
```

Result:

```text
### Request
POST https://api.example.com/users
Content-Type: application/json

{
  "name": "John",
  "email": "john@example.com"
}
```

### YAML Output

```bash
pbpaste | restcli curl2http -f yaml -o get-users.yaml
```

Result:

```yaml
name: Request
method: GET
url: "https://api.example.com/users"
```

### JSON Output

```bash
pbpaste | restcli curl2http -f json -o create-user.json
```

Result:

```json
{
  "name": "Request",
  "method": "POST",
  "url": "https://api.example.com/users",
  "headers": {
    "Content-Type": "application/json"
  },
  "body": "{\"name\":\"John\"}"
}
```

### Import Headers

```bash
pbpaste | restcli curl2http --import-headers
```

Extracts headers to profile format:

```json
{
  "headers": {
    "Authorization": "Bearer token123",
    "X-API-Key": "key456"
  }
}
```

## Workflow

### From Browser DevTools

1. Open browser DevTools
2. Go to Network tab
3. Right-click request
4. Select "Copy as cURL"
5. Paste and convert:

```bash
pbpaste | restcli curl2http -o request.http
```

### From Documentation

Copy cURL example from API docs:

```bash
pbpaste | restcli curl2http -f yaml -o api-call.yaml
```

## Supported cURL Flags

| Flag     | Support | Description                       |
| -------- | ------- | --------------------------------- |
| `-X`     | Yes     | HTTP method                       |
| `-H`     | Yes     | Headers                           |
| `-d`     | Yes     | Request body                      |
| `--data` | Yes     | Request body                      |
| `-u`     | Partial | Converted to Authorization header |
| `--user` | Partial | Converted to Authorization header |

## Limitations

1. Complex cURL scripts may not convert perfectly
2. Shell variables in cURL not converted to REST CLI variables
3. Some advanced cURL features unsupported

For complex cases, manually edit converted files.
