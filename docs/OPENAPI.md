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
```

### From URL

```bash
# Fetch from remote URL
restcli-openapi2http https://petstore3.swagger.io/api/v3/openapi.json --output requests/
```

## Features

### Complete Documentation Extraction

Automatically generates `.http` files with:

- **@description** - From operation summary/description
- **@tag** - From operation tags
- **@param** - From parameters and request body schema
- **@example** - From schema examples
- **@response** - From response definitions

### Smart Organization

Three organization strategies:

**By Tags** (default):

```text
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

```text
requests/
├── api/
│   ├── get-users.http
│   └── post-user.http
└── auth/
    └── post-login.http
```

**Flat**:

```text
requests/
├── create-user.http
├── get-user.http
├── login.http
├── list-products.http
└── get-product.http
```

### Variable Placeholders

- URL paths: `/users/{id}` → `/users/{{id}}`
- Query parameters: `?page={{page}}&limit={{limit}}`
- Request body: Auto-generates with `{{variableName}}` placeholders
- Headers: Security schemes → `Authorization: Bearer {{token}}`

## Usage

### Basic Command

```bash
restcli-openapi2http <input> [OPTIONS]
```

### Options

| Option          | Short | Description                                       | Default    |
| --------------- | ----- | ------------------------------------------------- | ---------- |
| `--input`       | `-i`  | Path to OpenAPI spec file or URL                  | Required   |
| `--output`      | `-o`  | Output directory                                  | `requests` |
| `--organize-by` |       | Organization strategy: `tags`, `paths`, or `flat` | `tags`     |
| `--help`        | `-h`  | Show help message                                 | -          |
