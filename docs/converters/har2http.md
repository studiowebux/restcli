# HAR to HTTP Converter

Convert HTTP Archive (HAR) files exported from browser developer tools into editable `.http` request files.

## Overview

HAR files capture complete HTTP traffic including headers, bodies, and timing data. Browser developer tools export network activity as HAR files, making them useful for debugging and API exploration.

## Export HAR Files

### Chrome DevTools
1. Open DevTools (F12)
2. Navigate to Network tab
3. Perform HTTP requests
4. Right-click in request list → "Save all as HAR with content"

### Firefox
1. Open Developer Tools (F12)
2. Navigate to Network tab
3. Perform HTTP requests
4. Click gear icon → "Save All As HAR"

### Safari
1. Enable Develop menu (Preferences → Advanced → Show Develop menu)
2. Develop → Show Web Inspector
3. Navigate to Network tab
4. Perform HTTP requests
5. Right-click in request list → "Export HAR"

## Basic Usage

```bash
restcli har2http <har-file>
```

Creates `.http` files in `requests/` directory by default.

## Options

```
--output, -o <directory>    Output directory (default: "requests")
--import-headers            Include sensitive headers (Cookie, Authorization)
--format <type>             Output format: http, json, yaml (default: "http")
--filter <pattern>          Filter requests by URL pattern
```

## Examples

### Import All Requests

```bash
restcli har2http network-log.har
```

Converts all HTTP(S) requests, filters out sensitive headers automatically.

### Import with Headers

```bash
restcli har2http network-log.har --import-headers
```

Includes Authorization, Cookie, X-Auth-Token, X-API-Key headers. Bearer tokens are extracted to `{{token}}` variables.

### Filter by URL Pattern

```bash
restcli har2http network-log.har --filter "api.example.com"
```

Only imports requests matching the URL pattern.

### Custom Output Directory

```bash
restcli har2http network-log.har -o api-requests
```

### JSON or YAML Output

```bash
restcli har2http network-log.har --format json
restcli har2http network-log.har --format yaml
```

JSON and YAML formats create structured request files compatible with restcli's native format. Variable extraction is only available in HTTP format.

## Generated Files

File naming pattern: `<method>-<path>.http`

Path is converted to filename by:
- Replacing `/` with `-`
- Converting to lowercase
- Removing invalid characters

Examples:
- `/posts/1` → `get-posts-1.http`
- `/api/users` → `get-api-users.http`
- `/comments/5/replies` → `get-comments-5-replies.http`

### Example Output (HTTP Format)

```http
### POST /api/users
# Variables:
#   token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9
POST https://api.example.com/api/users
Content-Type: application/json
Authorization: Bearer {{token}}

{"name":"John Doe","email":"john@example.com"}
```

## Variable Extraction

Bearer tokens are automatically detected and templated:

**Original:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9
```

**Converted:**
```
Authorization: Bearer {{token}}
```

Token value is documented in file comments.

## Header Filtering

By default, sensitive headers are excluded:
- Cookie
- Authorization
- X-Auth-Token
- X-API-Key

Use `--import-headers` to include them as variables.

## Workflow Integration

### 1. Capture API Traffic

Use browser to interact with API, export HAR file.

### 2. Import Requests

```bash
restcli har2http captures/session.har -o api-requests --import-headers
```

### 3. Organize

Move generated files to appropriate collections:

```
collections/
  auth/
    post-login.http
  users/
    get-users.http
    post-users.http
```

### 4. Refine

Edit imported requests:
- Replace hardcoded values with variables
- Add environment-specific configurations
- Remove unnecessary headers
- Add documentation comments

### 5. Execute

```bash
restcli exec api-requests/post-users.http
```

## Tips

### Large HAR Files

Filter by domain to reduce output:

```bash
restcli har2http large-capture.har --filter "myapi.com"
```

### Multiple Sessions

Use descriptive output directories:

```bash
restcli har2http session1.har -o requests/session1
restcli har2http session2.har -o requests/session2
```

### Browser Extensions

Some extensions export HAR files automatically:
- Chrome: HTTP Archive Viewer
- Firefox: HAR Export Trigger

### Review Generated Files

Always review imported requests before committing to repository. Remove sensitive data, test tokens, or personal information.

## Common Issues

### Non-HTTP Requests Skipped

Only HTTP(S) requests are converted. WebSocket, data URLs, and other protocols are ignored.

### Empty Request Bodies

Some browsers don't include request bodies in HAR exports. Use "Save all as HAR **with content**" option.

### Duplicate Files

Multiple requests to same endpoint will overwrite each other unless the URLs differ (e.g., query parameters or IDs in path). If you need to preserve all requests, export separate HAR files or use the `--filter` flag to process them separately.

## Related

- [curl2http](/docs/converters/curl2http.md) - Import from curl commands
- [openapi2http](/docs/converters/openapi2http.md) - Import from OpenAPI specs
- [Collections](/docs/features/collections.md) - Organize requests
- [Variables](/docs/features/variables.md) - Template configuration
