---
title: Categories and Filtering
tags:
  - guide
---

# Categories and Filtering

## Overview

Categories allow you to organize and filter request files by function, purpose, or any custom grouping. Filter the file list to show only requests matching specific categories.

## Adding Categories

Categories work in `.http` files and structured formats (`.yaml`, `.yml`, `.json`, `.jsonc`).

**Note:** OpenAPI files use `tags` field which is OpenAPI-specific metadata. For regular request files, use `@category`.

### .http Files

Use the `@category` annotation in comments before the request:

```http
### Login Endpoint
# @category auth
# @category critical
# @category api
POST https://api.example.com/auth/login
Content-Type: application/json

{
  "username": "{{username}}",
  "password": "{{password}}"
}
```

### YAML Files

Use the `documentation.tags` array (stored as tags internally but represents categories):

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

## Viewing Categories in TUI

Categories appear in the file list sidebar next to each file name:

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

Press `i` to inspect a request. The inspect modal shows all categories:

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

Categories:
  auth, critical, api
```

## Filtering by Category

### Enter Filter Mode

Press `t` to enter category filter mode. The status bar shows:

```
Category: █
```

### Type Category Name

Enter the category name to filter by (e.g., `auth`):

```
Category: auth█
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
| `t` | Enter category filter mode |
| `Enter` | Apply category filter |
| `T` | Clear active filter |
| `Esc` | Cancel input |
| `←` / `→` | Move cursor in input |
| `Home` / `Ctrl+A` | Move cursor to start |
| `End` / `Ctrl+E` | Move cursor to end |
| `Backspace` | Delete character before cursor |
| `Delete` / `Ctrl+D` | Delete character at cursor |
| `Ctrl+U` | Delete from cursor to start |
| `Ctrl+K` | Delete from cursor to end |

## Category Naming Conventions

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

Categorize critical endpoints for quick smoke testing:

```http
### Health Check
# @category smoke-test
# @category monitoring
GET https://api.example.com/health

### Login
# @category smoke-test
# @category auth
POST https://api.example.com/auth/login
```

Filter by `smoke-test` to see only essential checks.

### Admin Operations

Categorize admin-only endpoints:

```http
### Delete User
# @category admin
# @category critical
# @category users
DELETE https://api.example.com/users/{{userId}}
Authorization: Bearer {{admin_token}}
```

Filter by `admin` to see all administrative operations.

### Environment-Specific

Categorize requests by environment:

```http
### Dev API Health
# @category dev
# @category monitoring
GET https://api.dev.example.com/health

### Prod API Health
# @category production
# @category monitoring
# @category critical
GET https://api.example.com/health
```

Filter by `dev` or `production` to focus on specific environments.

## Tips

1. **Use consistent category names** - Establish conventions across your team
2. **Categorize multiple dimensions** - Combine function, criticality, and environment
3. **Keep categories short** - Easier to type in filter mode
4. **Categories at file level** - Each `.http` file can have multiple requests, categories apply to the file
5. **Case insensitive** - Category matching ignores case (`Auth` matches `auth`)

## Limitations

- Category filtering is case-insensitive
- Currently supports single category filters (multiple category support planned)
- Categories are stored per file, not per individual request within multi-request files
- OpenAPI files use `tags` field which is separate from `@category` annotations

## Related Features

- [File Formats](file-formats.md) - Supported request file formats
- [TUI Mode](tui-mode.md) - Navigation and keyboard shortcuts
- [History](history.md) - View historical requests with categories
