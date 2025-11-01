# curl2http - Convert cURL to .http Files

Quickly convert cURL commands (from browser DevTools, documentation, etc.) into `.http` request files for the TUI.

## Quick Start

### Using Compiled Binary (Recommended)

**From Clipboard (macOS):**
```bash
pbpaste | restcli-curl2http --output requests/my-request.http
```

**From Clipboard (Linux):**
```bash
xclip -o | restcli-curl2http --output requests/my-request.http
```

**From Command Line:**
```bash
restcli-curl2http --output requests/login.http 'curl -X POST http://localhost:3000/auth/login -H "Content-Type: application/json" -d '"'"'{"username":"test","password":"pass"}'"'"''
```

### Using Deno (Development)

**From Clipboard:**
```bash
pbpaste | deno task curl2http --output requests/my-request.http
```

**From Command Line:**
```bash
deno task curl2http --output requests/login.http 'curl -X POST http://localhost:3000/auth/login ...'
```

## Features

### ✅ Automatic Parsing

- Extracts HTTP method (`-X`, `--request`)
- Extracts URL
- Extracts all headers (`-H`, `--header`)
- Extracts request body (`-d`, `--data`, `--data-raw`)
- Handles multiline curl commands (with `\`)

### 🔍 Smart Variable Detection

Automatically detects and suggests variables:

**Input:**

```bash
curl http://localhost:3000/api/users -H "Authorization: Bearer eyJhbGc..."
```

**Output:**

```http
### GET api/users
GET {{baseUrl}}/api/users
Authorization: Bearer eyJhbGc...
```

**Detected variables:**

```json
{
  "baseUrl": "http://localhost:3000"
}
```

### 🔒 Security Header Filtering

By default, sensitive headers are excluded from generated `.http` files for security:

**Filtered headers:**
- `Authorization`
- `Cookie`
- `X-API-Key`
- `X-Auth-Token`
- `API-Key`
- `Auth-Token`
- `Bearer`
- `X-Session-Token`
- `X-CSRF-Token`

**Example:**

```bash
curl http://localhost:3000/api/users -H "Authorization: Bearer secret123"
```

**Output:**

```http
### GET api/users
GET {{baseUrl}}/api/users
```

```
🔒 Excluded sensitive headers (use --import-headers to include):
  Authorization: Bearer secret123

💡 Add these to your profile headers in .profiles.json instead
```

**To include sensitive headers:**

```bash
pbpaste | restcli-curl2http --output requests/file.http --import-headers
```

This is useful for:
- Generating quick test files where you'll manually replace tokens with `{{token}}`
- Documenting API endpoints with actual auth examples
- One-off testing scenarios

**Best practice:** Keep auth credentials in `.profiles.json` for reusability and security.

### 📁 Output File Specification

Use the `--output` (or `-o`) flag to specify where to save the `.http` file:

```bash
# Save to specific file
pbpaste | restcli-curl2http --output requests/login.http

# Save to directory (auto-generates filename)
pbpaste | restcli-curl2http --output requests/

# Short form
pbpaste | restcli-curl2http -o requests/my-request.http
```

**Auto-generated filenames** are based on the URL path:

| URL                 | Generated Filename |
| ------------------- | ------------------ |
| `/auth/login`       | `post-login.http`  |
| `/users`            | `users.http`       |
| `/api/v1/posts/123` | `123.http`         |

## Common Use Cases

### Browser DevTools → .http File

1. Open DevTools (F12)
2. Go to Network tab
3. Right-click on a request → "Copy as cURL"
4. Run: `pbpaste | restcli-curl2http --output requests/my-request.http`
5. File is saved and ready to use in TUI

### API Documentation → .http File

Copy curl example from docs:

```bash
echo 'curl https://api.github.com/repos/denoland/deno' | restcli-curl2http --output requests/github.http
```

### Quick Testing

Convert and test in one go:

```bash
# Convert
pbpaste | restcli-curl2http --output requests/new-request.http

# Then run TUI
restcli

# Navigate to the new file and execute
```

## Examples

### Simple GET

**Input:**

```bash
curl https://jsonplaceholder.typicode.com/posts/1
```

**Output:**

```http
### GET posts/1
GET {{baseUrl}}/posts/1
```

### POST with JSON

**Input:**

```bash
curl -X POST https://api.example.com/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer token123" \
  -d '{"name":"John","email":"john@example.com"}'
```

**Output:**

```http
### POST users
POST {{baseUrl}}/users
Content-Type: application/json
Authorization: Bearer token123

{
  "name": "John",
  "email": "john@example.com"
}
```

### Complex Headers

**Input:**

```bash
curl 'https://api.example.com/data' \
  -H 'accept: application/json' \
  -H 'accept-language: en-US,en;q=0.9' \
  -H 'authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...' \
  -H 'cache-control: no-cache' \
  -H 'user-agent: Mozilla/5.0'
```

**Output (default - filters sensitive headers):**

```http
### GET data
GET {{baseUrl}}/data
accept: application/json
accept-language: en-US,en;q=0.9
cache-control: no-cache
user-agent: Mozilla/5.0
```

```
🔒 Excluded sensitive headers (use --import-headers to include):
  authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

💡 Add these to your profile headers in .profiles.json instead
```

**Output (with --import-headers):**

```http
### GET data
GET {{baseUrl}}/data
accept: application/json
accept-language: en-US,en;q=0.9
authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
cache-control: no-cache
user-agent: Mozilla/5.0
```

## Tips

### Security Best Practices

1. **By default, sensitive headers are excluded** - this prevents accidentally committing credentials to git

2. **Use profile headers for auth** - store credentials in `.profiles.json` (add to `.gitignore`):
   ```json
   {
     "name": "Production",
     "headers": {
       "Authorization": "Bearer {{prodToken}}"
     },
     "variables": {
       "prodToken": "your-secret-token"
     }
   }
   ```

3. **Only use --import-headers when needed** - for quick testing or when you'll replace tokens with variables

### Clean Up Generated Files

After conversion, you might want to:

1. **Replace auth tokens with variables** (if using --import-headers):

   ```http
   Authorization: Bearer {{token}}
   ```

2. **Simplify headers** - remove unnecessary ones like `user-agent`, `accept-language`

3. **Add to profile** - move common headers to `.profiles.json`

4. **Use variables** - replace hard-coded values:
   ```http
   POST {{baseUrl}}/users/{{userId}}/posts
   ```

### Browser Copy Tips

When copying from browser DevTools, the curl command might include:

- Compressed data (`--compressed`)
- Cookies (`--cookie`)
- Lots of headers
- Authorization tokens

The tool automatically:
- Filters out sensitive headers like `Authorization` and `Cookie` (unless `--import-headers` is used)
- Handles common curl flags
- Pretty-prints JSON bodies

You may still want to clean up unnecessary headers like `user-agent` and `accept-language` manually.

### Batch Conversion

Convert multiple requests:

```bash
# From multiple files
for curl_file in request1.txt request2.txt request3.txt; do
  filename=$(basename "$curl_file" .txt)
  cat "$curl_file" | restcli-curl2http --output "requests/${filename}.http"
done

# From clipboard with different endpoints
pbpaste | restcli-curl2http --output requests/endpoint1.http
pbpaste | restcli-curl2http --output requests/endpoint2.http
```

## Limitations

- **Cookies**: Extracted as headers, but you might want to replace with `{{cookie}}` variables
- **File uploads**: `--form` and multipart data not yet supported
- **Authentication**: Basic auth (`-u`) not yet supported (use headers instead)
- **Complex escaping**: Very complex shell escaping might not parse correctly

## Troubleshooting

### "Could not extract URL"

Make sure the curl command includes a valid URL:

```bash
# ❌ Won't work
curl -X POST

# ✅ Works
curl -X POST http://localhost:3000
```

### Body Not Detected

Ensure the data flag is recognized:

```bash
# Supported
-d "..."
--data "..."
--data-raw "..."
--data-binary "..."
```

### Nested Quotes Issue

If you have complex JSON with nested quotes, use single quotes on the outside:

```bash
curl -d '{"key":"value with \"quotes\""}' http://...
```

Or escape properly:

```bash
curl -d "{\"key\":\"value\"}" http://...
```

---

**Integration:** Works seamlessly with the HTTP TUI - convert curl commands, save to `requests/`, and test immediately!
