# curl2http - Convert cURL to .http Files

Quickly convert cURL commands (from browser DevTools, documentation, etc.) into `.http` request files for the TUI.

## Quick Start

### From Command Line

```bash
deno task curl2http 'curl -X POST http://localhost:3000/auth/login -H "Content-Type: application/json" -d '"'"'{"username":"test","password":"pass"}'"'"''
```

### From Clipboard (macOS)

```bash
pbpaste | deno task curl2http
```

### From Clipboard (Linux)

```bash
xclip -o | deno task curl2http
```

### Direct Execution

```bash
./curl2http.ts 'curl ...'
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

### 📁 Smart Filename Suggestions

Based on the URL path:

| URL                 | Suggested Filename |
| ------------------- | ------------------ |
| `/auth/login`       | `post-login.http`  |
| `/users`            | `users.http`       |
| `/api/v1/posts/123` | `123.http`         |

### 🎯 Interactive Mode

When run, the tool prompts:

```
📝 Converted curl to .http format:
────────────────────────────────────────────
### POST auth/login
POST {{baseUrl}}/auth/login
Content-Type: application/json

{"username":"test","password":"pass"}
────────────────────────────────────────────

💡 Detected variables:
  baseUrl: http://localhost:3000

📁 Suggested filename: requests/post-login.http

Options:
  1. Save to suggested location
  2. Enter custom filename
  3. Print to stdout only
  4. Cancel
```

## Common Use Cases

### Browser DevTools → .http File

1. Open DevTools (F12)
2. Go to Network tab
3. Right-click on a request → "Copy as cURL"
4. Run: `pbpaste | deno task curl2http`
5. Choose option 1 to save

### API Documentation → .http File

Copy curl example from docs:

```bash
echo 'curl https://api.github.com/repos/denoland/deno' | deno task curl2http
```

### Quick Testing

Convert and test in one go:

```bash
# Convert
pbpaste | deno task curl2http

# Then run TUI
deno task dev

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

**Output:**

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

### Clean Up Generated Files

After conversion, you might want to:

1. **Replace auth tokens with variables:**

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

The tool handles these, but you may want to clean up the generated file manually.

### Batch Conversion

Convert multiple requests:

```bash
for curl_cmd in request1.txt request2.txt request3.txt; do
  cat "$curl_cmd" | deno task curl2http
done
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
