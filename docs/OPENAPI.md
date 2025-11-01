# OpenAPI/Swagger Integration

Convert OpenAPI/Swagger specifications to fully documented `.http` request files automatically.

## Overview

The `restcli-openapi2http` tool converts OpenAPI 3.x specifications (JSON or YAML) into `.http` files with complete documentation, making it easy to:

1. **Import entire APIs** from OpenAPI specs
2. **Auto-generate documentation** for all endpoints
3. **Organize requests** by tags, paths, or flat structure
4. **Get started quickly** with examples and placeholders

## Quick Start

### From Local File

```bash
# Using compiled binary
restcli-openapi2http swagger.json --output requests/

# Or with Deno
deno run --allow-read --allow-write --allow-net scripts/openapi2http.ts swagger.json
```

### From URL

```bash
# Fetch from remote URL
restcli-openapi2http https://petstore3.swagger.io/api/v3/openapi.json --output requests/
```

## Features

### ✅ Complete Documentation Extraction

Automatically generates `.http` files with:

- **@description** - From operation summary/description
- **@tag** - From operation tags
- **@param** - From parameters and request body schema
- **@example** - From schema examples
- **@response** - From response definitions

### ✅ Smart Organization

Three organization strategies:

**By Tags** (default):
```
requests/
├── users/
│   ├── create-user.http
│   └── get-user.http
├── auth/
│   └── login.http
└── products/
    ├── list-products.http
    └── get-product.http
```

**By Paths**:
```
requests/
├── api/
│   ├── get-users.http
│   └── post-user.http
└── auth/
    └── post-login.http
```

**Flat**:
```
requests/
├── create-user.http
├── get-user.http
├── login.http
├── list-products.http
└── get-product.http
```

### ✅ Variable Placeholders

- URL paths: `/users/{id}` → `/users/{{id}}`
- Query parameters: `?page={{page}}&limit={{limit}}`
- Request body: Auto-generates with `{{variableName}}` placeholders
- Headers: Security schemes → `Authorization: Bearer {{token}}`

### ✅ Supports OpenAPI 3.x

- OpenAPI 3.0.x (JSON/YAML)
- OpenAPI 3.1.x (JSON/YAML)
- Built-in YAML parser (no external dependencies)

## Usage

### Basic Command

```bash
restcli-openapi2http <input> [OPTIONS]
```

### Options

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--input` | `-i` | Path to OpenAPI spec file or URL | Required |
| `--output` | `-o` | Output directory | `requests` |
| `--organize-by` | - | Organization strategy: `tags`, `paths`, or `flat` | `tags` |
| `--help` | `-h` | Show help message | - |

### Examples

```bash
# From local JSON file
restcli-openapi2http ./swagger.json

# From local YAML file
restcli-openapi2http ./api.yaml --output my-requests/

# From URL
restcli-openapi2http https://api.example.com/openapi.json

# Organize by URL paths instead of tags
restcli-openapi2http swagger.json --organize-by paths

# Flat structure (no subdirectories)
restcli-openapi2http swagger.json --organize-by flat
```

## Generated Output

### Example: Create User Endpoint

**Input (OpenAPI spec):**
```yaml
paths:
  /api/users:
    post:
      summary: Create User
      description: Creates a new user in the system
      tags:
        - users
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - name
                - email
              properties:
                name:
                  type: string
                  description: User's full name
                  example: John Doe
                email:
                  type: string
                  description: User's email address
                  example: john@example.com
      responses:
        '201':
          description: User created successfully
        '400':
          description: Invalid request data
      security:
        - bearerAuth: []
```

**Output (.http file):**
```http
### Create User
# @description Creates a new user in the system
# @tag users
# @param name {string} required - User's full name
# @example name "John Doe"
# @param email {string} required - User's email address
# @example email "john@example.com"
# @response 201 - User created successfully
# @response 400 - Invalid request data
POST {{baseUrl}}/api/users
Content-Type: application/json
Authorization: Bearer {{token}}

{
  "name": "{{name}}",
  "email": "{{email}}"
}

###
```

## Workflow

### 1. Convert Spec

```bash
restcli-openapi2http swagger.json --output requests/
```

**Output:**
```
✅ Conversion complete!

📁 Created 47 .http files
📂 Output directory: requests/
📋 Organization: tags

Files by category:
  • users: 12 files
  • products: 18 files
  • auth: 5 files
  • orders: 12 files

💡 Next steps:
  1. Review generated files in requests/
  2. Update variable placeholders in .session.json or .profiles.json
  3. Run 'restcli' to test your requests
```

### 2. Configure Variables

Create a profile in `.profiles.json`:

```json
[
  {
    "name": "Development",
    "variables": {
      "baseUrl": "http://localhost:3000",
      "token": "dev-token-here",
      "userId": "123"
    },
    "headers": {
      "X-API-Version": "v1"
    }
  }
]
```

### 3. Test in TUI

```bash
restcli
```

- Navigate to generated request files
- Press `m` to view documentation
- Press `Enter` to execute requests
- Use `v` to edit variables, `h` to edit headers

## Best Practices

### 1. Review Generated Files

After conversion:
- Check that variable placeholders match your needs
- Update example values if needed
- Remove unnecessary headers or parameters

### 2. Organize by Domain

Use tags in your OpenAPI spec to organize requests logically:

```yaml
tags:
  - name: users
    description: User management
  - name: auth
    description: Authentication
  - name: products
    description: Product catalog
```

### 3. Provide Examples

Include examples in your OpenAPI schema for better generated files:

```yaml
schema:
  type: object
  properties:
    email:
      type: string
      example: user@example.com  # This becomes @example annotation
```

### 4. Use Security Schemes

Define security schemes for automatic header generation:

```yaml
components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
```

Results in:
```http
Authorization: Bearer {{token}}
```

## Limitations

### Not Yet Supported

- **Nested $ref resolution**: Only top-level component references in request bodies are resolved (e.g., `#/components/schemas/MySchema`)
- **File uploads**: Multipart form data not generated
- **Webhooks**: OpenAPI 3.1 webhooks not converted
- **Callbacks**: Callback operations not supported

### Workarounds

For complex nested schemas with multiple `$ref`:
1. Use a tool to dereference the spec first
2. Or manually edit the generated `.http` files

**Note**: Simple `$ref` to component schemas in request bodies are automatically resolved and will include proper examples from the schema.

## Troubleshooting

### "Invalid OpenAPI spec" Error

**Problem:** Spec validation failed

**Solutions:**
- Ensure spec has required fields: `openapi`, `info`, `paths`
- Validate your spec at https://editor.swagger.io/
- Check for syntax errors in YAML/JSON

### "Failed to parse YAML" Error

**Problem:** YAML parser cannot read the file

**Solutions:**
- Check for YAML syntax errors
- Ensure proper indentation
- Try converting to JSON first

### Missing Variables

**Problem:** Generated files don't have variable placeholders

**Solutions:**
- Add examples to your OpenAPI schema
- Manually edit generated files to add variables
- Use the variable editor (`v`) in the TUI

### Too Many Files

**Problem:** Conversion creates too many files

**Solutions:**
- Use `--organize-by flat` to put all in one directory
- Filter the spec before conversion
- Manually delete unwanted files

## Advanced Usage

### Scripting

Automate API imports:

```bash
#!/bin/bash

# Download latest API spec
curl https://api.example.com/openapi.json -o spec.json

# Convert to requests
restcli-openapi2http spec.json --output requests/

# Commit to git
git add requests/
git commit -m "Update API requests from spec"
```

### CI/CD Integration

Keep requests synchronized with API:

```yaml
# .github/workflows/sync-api.yml
name: Sync API Requests

on:
  schedule:
    - cron: '0 0 * * *'  # Daily

jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: denoland/setup-deno@v1
      - name: Fetch OpenAPI spec
        run: curl https://api.example.com/openapi.json -o spec.json
      - name: Convert to requests
        run: deno run --allow-read --allow-write --allow-net scripts/openapi2http.ts spec.json
      - name: Commit changes
        run: |
          git config user.name "API Sync Bot"
          git add requests/
          git commit -m "chore: sync API requests" || echo "No changes"
          git push
```

## Related Documentation

- [DOCUMENTATION.md](./DOCUMENTATION.md) - How to write documentation manually
- [PROFILES.md](./PROFILES.md) - Profile configuration
- [README.md](../README.md) - Main documentation

## Examples

See `examples/` directory for:
- Sample OpenAPI specs
- Generated `.http` files
- Variable configuration examples
