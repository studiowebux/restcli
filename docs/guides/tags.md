---
title: Tags and Filtering
tags:
  - guide
---

# Tags and Filtering

## Overview

Tags allow you to organize and filter request files by function, purpose, or any custom grouping. Filter the file list to show only requests matching specific tags. This uses OpenAPI-compliant `@tag` annotations.

## Adding Tags

Tags work in `.http` files and structured formats (`.yaml`, `.yml`, `.json`, `.jsonc`).

### .http Files

Use the `@tag` annotation in comments before the request:

```http
### Login Endpoint
# @tag auth
# @tag critical
# @tag api
POST https://api.example.com/auth/login
Content-Type: application/json

{
  "username": "{{username}}",
  "password": "{{password}}"
}
```

### YAML Files

Use the `documentation.tags` array:

```yaml
- name: Login Endpoint
  method: POST
  url: https://api.example.com/auth/login
  documentation:
    description: Authenticate user and return access token
    tags:
      - auth
      - critical
      - api
  headers:
    Content-Type: application/json
  body: |
    {
      "username": "{{username}}",
      "password": "{{password}}"
    }
```

### JSON/JSONC Files

Use the `documentation.tags` array:

```json
{
  "name": "Login Endpoint",
  "method": "POST",
  "url": "https://api.example.com/auth/login",
  "documentation": {
    "description": "Authenticate user and return access token",
    "tags": ["auth", "critical", "api"]
  },
  "headers": {
    "Content-Type": "application/json"
  },
  "body": "{\"username\": \"{{username}}\", \"password\": \"{{password}}\"}"
}
```

## Viewing Tags in TUI

Tags appear in the file list sidebar next to each file name:

```
Files
┌─────────────────────────────────────────────┐
│ 0 POST auth/login.http [auth,critical]      │
│ 1 POST auth/refresh.http [auth,api]         │
│ 2 GET  users/list.http [users,admin]        │
│ 3 GET  health.http [monitoring,api]         │
└─────────────────────────────────────────────┘
```

- Up to 2 tags displayed per file
- `...` indicator when more tags exist
- Tags shown in subtle gray color

### Inspect Modal

Press `i` to inspect a request. The inspect modal shows all tags:

```
Request Preview

POST https://api.example.com/auth/login

Headers:
  Content-Type: application/json

Body:
  {
    "username": "user",
    "password": "pass"
  }

Tags:
  auth, critical, api
```

## Filtering by Tag

### Enter Filter Mode

Press `t` to enter tag filter mode. The status bar shows:

```
Tag: █
```

### Type Tag Name

Enter the tag name to filter by (e.g., `auth`):

```
Tag: auth█
```

### Apply Filter

Press `Enter` to apply the filter. The sidebar updates:

```
Files (auth)
┌─────────────────────────────────────────────┐
│ 0 POST auth/login.http [auth,critical]      │
│ 1 POST auth/refresh.http [auth,api]         │
└─────────────────────────────────────────────┘
[2/5]
```

Only files containing the specified tag are shown. The count `[2/5]` indicates 2 filtered files out of 5 total.

### Clear Filter

Press `T` (Shift+T) to clear the active filter.

The sidebar returns to showing all files:

```
Files
┌─────────────────────────────────────────────┐
│ 0 POST auth/login.http [auth,critical]      │
│ 1 POST auth/refresh.http [auth,api]         │
│ 2 GET  users/list.http [users,admin]        │
│ 3 GET  health.http [monitoring,api]         │
│ 4 POST users/create.http [users,critical]   │
└─────────────────────────────────────────────┘
[5/5]
```

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `t` | Enter tag filter mode |
| `Enter` | Apply tag filter |
| `T` | Clear active filter |
| `Esc` | Cancel input |
| `←` / `→` | Move cursor in input |
| `Home` / `Ctrl+A` | Move cursor to start |
| `End` / `Ctrl+E` | Move cursor to end |
| `Backspace` | Delete character before cursor |
| `Delete` / `Ctrl+D` | Delete character at cursor |
| `Ctrl+U` | Delete from cursor to start |
| `Ctrl+K` | Delete from cursor to end |

## Tag Naming Conventions

### By Function
```
auth, users, payments, orders, inventory, reporting
```

### By Criticality
```
critical, safe, experimental, deprecated
```

### By Environment
```
dev, staging, production, local
```

### By Test Type
```
smoke-test, integration, e2e, regression
```

### By Protocol/Type
```
rest, graphql, grpc, websocket
```

### By Access Level
```
public, admin, internal, partner
```

## Practical Examples

### Smoke Test Suite

Tag critical endpoints for quick smoke testing:

```http
### Health Check
# @tag smoke-test
# @tag monitoring
GET https://api.example.com/health

### Login
# @tag smoke-test
# @tag auth
POST https://api.example.com/auth/login
```

Filter by `smoke-test` to see only essential checks.

### Admin Operations

Tag admin-only endpoints:

```http
### Delete User
# @tag admin
# @tag critical
# @tag users
DELETE https://api.example.com/users/{{userId}}
Authorization: Bearer {{admin_token}}
```

Filter by `admin` to see all administrative operations.

### Environment-Specific

Tag requests by environment:

```http
### Dev API Health
# @tag dev
# @tag monitoring
GET https://api.dev.example.com/health

### Prod API Health
# @tag production
# @tag monitoring
# @tag critical
GET https://api.example.com/health
```

Filter by `dev` or `production` to focus on specific environments.

## Tips

1. **Use consistent tag names** - Establish conventions across your team
2. **Tag multiple dimensions** - Combine function, criticality, and environment
3. **Keep tags short** - Easier to type in filter mode
4. **Tags at file level** - Each `.http` file can have multiple requests, tags apply to the file
5. **Case insensitive** - Tag matching ignores case (`Auth` matches `auth`)

## Limitations

- Tag filtering is case-insensitive
- Currently supports single tag filters (multiple tag support planned)
- Tags are stored per file, not per individual request within multi-request files

## Related Features

- [File Formats](file-formats.md) - Supported request file formats
- [TUI Mode](tui-mode.md) - Navigation and keyboard shortcuts
- [History](history.md) - View historical requests with tags
