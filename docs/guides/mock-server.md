---
title: Mock Server
tags:
  - guide
---

# Mock Server

Run a configurable HTTP mock server for testing.

## Configuration

Create `.mock.yaml` or `.mock.json` in your project:

```yaml
port: 8080
host: localhost
logging: true

routes:
  - name: Get Users
    method: GET
    path: /api/users
    status: 200
    headers:
      Content-Type: application/json
    body: |
      [
        {"id": 1, "name": "John Doe"},
        {"id": 2, "name": "Jane Smith"}
      ]

  - name: Create User
    method: POST
    path: /api/users
    status: 201
    headers:
      Content-Type: application/json
    body: '{"id": 3, "created": true}'

  - name: Delayed Response
    method: GET
    path: /api/slow
    status: 200
    delay: 2000
    body: '{"message": "Delayed by 2 seconds"}'

  - name: Error Response
    method: GET
    path: /api/error
    status: 500
    body: '{"error": "Internal server error"}'

  - name: Catch-all
    method: GET
    path: /api/
    pathType: prefix
    status: 404
    body: '{"error": "Not found"}'
```

## Route Configuration

### Required Fields

- `method`: HTTP method (GET, POST, PUT, DELETE, etc.)
- `path`: URL path to match
- `status`: HTTP status code to return

### Optional Fields

- `name`: Descriptive name for the route
- `pathType`: Match type - `exact` (default), `prefix`, or `regex`
- `headers`: Response headers (map of key-value pairs)
- `body`: Inline response body (string)
- `bodyFile`: Path to file containing response body (relative to config file)
- `delay`: Response delay in milliseconds
- `description`: Route documentation

### Path Matching

**Exact Match** (default)
```yaml
path: /api/users/1
pathType: exact
```

**Prefix Match**
```yaml
path: /api/
pathType: prefix
```
Matches `/api/users`, `/api/posts`, etc.

**Regex Match**
```yaml
path: /api/users/[0-9]+
pathType: regex
```
Matches `/api/users/1`, `/api/users/123`, etc.

### Body from File

```yaml
routes:
  - name: Large Response
    method: GET
    path: /api/data
    status: 200
    headers:
      Content-Type: application/json
    bodyFile: responses/data.json
```

File path is relative to the config file location.

## CLI Usage

### Start Server

Auto-discover config:
```bash
restcli mock start
```

Searches for `.mock.yaml`, `.mock.yml`, or `.mock.json` in:
- `mocks/` directory
- Current directory
- `../mocks/` (parent directory's mocks folder)
- `..` (parent directory)

Specify config file:
```bash
restcli mock start path/to/config.mock.yaml
```

Server runs in foreground. Press Ctrl+C to stop.

### Stop/Logs

These commands require the TUI:
```bash
# Start TUI and press 'M' for mock server management
restcli
```

## TUI Management

Press `M` in TUI to open mock server modal.

### When Server Stopped

- Shows available mock config files
- Press `s` to start server (uses first found config)
- Press `ESC` to close

### When Server Running

- Shows server address and config file
- Displays recent 10 requests with:
  - Timestamp
  - Method and path
  - Status code (colored by result)
  - Response time in milliseconds
  - Matched route name
- **Auto-refreshes every 500ms** - logs update in real-time
- Press `s` to stop server
- Press `c` to clear logs
- Press `ESC` to close

### Scrolling Logs

- `↑`/`k` or `↓`/`j`: Line up/down
- `PgUp`/`PgDn`: Page up/down
- `g`: Go to top
- `G`: Go to bottom

## Configuration Examples

### REST API Mock

```yaml
port: 8080
host: localhost
logging: true

routes:
  - name: List Items
    method: GET
    path: /items
    status: 200
    headers:
      Content-Type: application/json
    body: '[{"id": 1, "name": "Item 1"}]'

  - name: Get Item
    method: GET
    path: /items/[0-9]+
    pathType: regex
    status: 200
    headers:
      Content-Type: application/json
    body: '{"id": 1, "name": "Item 1"}'

  - name: Create Item
    method: POST
    path: /items
    status: 201
    headers:
      Content-Type: application/json
      Location: /items/2
    body: '{"id": 2, "name": "New Item"}'

  - name: Update Item
    method: PUT
    path: /items/[0-9]+
    pathType: regex
    status: 200
    body: '{"id": 1, "name": "Updated"}'

  - name: Delete Item
    method: DELETE
    path: /items/[0-9]+
    pathType: regex
    status: 204
```

### Testing Error Scenarios

```yaml
port: 8080
host: localhost
logging: true

routes:
  - name: Rate Limited
    method: GET
    path: /api/rate-limited
    status: 429
    headers:
      Retry-After: "60"
    body: '{"error": "Too many requests"}'

  - name: Unauthorized
    method: GET
    path: /api/protected
    status: 401
    body: '{"error": "Unauthorized"}'

  - name: Forbidden
    method: GET
    path: /api/admin
    status: 403
    body: '{"error": "Forbidden"}'

  - name: Validation Error
    method: POST
    path: /api/users
    status: 422
    body: '{"errors": [{"field": "email", "message": "Invalid format"}]}'

  - name: Server Error
    method: GET
    path: /api/error
    status: 500
    body: '{"error": "Internal server error"}'
```

### Latency Simulation

```yaml
port: 8080
host: localhost
logging: true

routes:
  - name: Fast Response
    method: GET
    path: /api/fast
    status: 200
    body: '{"speed": "fast"}'

  - name: Slow Response
    method: GET
    path: /api/slow
    status: 200
    delay: 2000
    body: '{"speed": "slow"}'

  - name: Very Slow
    method: GET
    path: /api/very-slow
    status: 200
    delay: 5000
    body: '{"speed": "very slow"}'
```

## Testing Against Mock Server

After starting the mock server, create `.http` files to test against it:

```http
### Test Get Users
GET http://localhost:8080/api/users

### Test Create User
POST http://localhost:8080/api/users
Content-Type: application/json

{
  "name": "New User",
  "email": "new@example.com"
}

### Test Error Response
GET http://localhost:8080/api/error

### Test Slow Endpoint
GET http://localhost:8080/api/slow
```

Run in CLI:
```bash
restcli test-get-users.http
```

Or use TUI to execute and view results interactively.

## Route Priority

Routes are matched in **configuration file order** - first match wins.

Order your routes from most specific to least specific:

Example:
```yaml
routes:
  # This matches first for exact /api/users
  - method: GET
    path: /api/users
    status: 200
    body: '{"users": []}'

  # This matches /api/users/123
  - method: GET
    path: /api/users/[0-9]+
    pathType: regex
    status: 200
    body: '{"user": {}}'

  # This matches anything under /api/ not matched above
  - method: GET
    path: /api/
    pathType: prefix
    status: 404
    body: '{"error": "Not found"}'
```

## Tips

**Organize by Environment**
```
mocks/
  dev.mock.yaml
  staging.mock.yaml
  errors.mock.yaml
```

**Use Body Files for Large Responses**
```
mocks/
  api.mock.yaml
  responses/
    users.json
    posts.json
```

**Enable Logging for Debugging**
```yaml
logging: true
```

Logs show in TUI modal and server console.

**Test Different Ports**
```yaml
port: 3000  # Avoid conflicts with other services
```

**Combine with Request Files**

Store mock configs and test requests together:
```
project/
  mocks/
    api.mock.yaml
  tests/
    test-api.http
```
