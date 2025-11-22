---
title: openapi2http Converter
description: Convert OpenAPI/Swagger specifications to REST CLI request files.
tags:
  - converter
---

# openapi2http Converter

Convert OpenAPI/Swagger specifications to REST CLI request files.

## Basic Usage

```bash
restcli openapi2http spec.yaml
```

Or JSON:

```bash
restcli openapi2http swagger.json
```

## Flags

### Output Directory

```bash
restcli openapi2http spec.yaml -o requests/
```

Creates request files in target directory.

Short: `-o`

### Organization

```bash
restcli openapi2http spec.yaml --organize-by tags
```

Options:

1. `tags`: Group by OpenAPI tags
2. `paths`: Group by path structure
3. `flat`: All files in output directory

Default: `flat`

### Output Format

```bash
restcli openapi2http spec.yaml -f yaml
```

Options: `http`, `yaml`, `json`, `jsonc`
Default: `http`
Short: `-f`

## Examples

### Basic Conversion

Input OpenAPI spec:

```yaml
openapi: 3.0.0
paths:
  /users:
    get:
      summary: List users
      operationId: listUsers
  /users/{id}:
    get:
      summary: Get user
      operationId: getUser
      parameters:
        - name: id
          in: path
          required: true
```

Command:

```bash
restcli openapi2http api.yaml -o requests/
```

Output files:

1. `requests/listUsers.http`
2. `requests/getUser.http`

### Organized by Tags

OpenAPI spec with tags:

```yaml
openapi: 3.0.0
paths:
  /users:
    get:
      tags: [Users]
      summary: List users
  /posts:
    get:
      tags: [Posts]
      summary: List posts
```

Command:

```bash
restcli openapi2http api.yaml --organize-by tags -o requests/
```

Output structure:

```text
requests/
├── Users/
│   └── listUsers.http
└── Posts/
    └── listPosts.http
```

### Organized by Paths

Command:

```bash
restcli openapi2http api.yaml --organize-by paths -o requests/
```

Output structure:

```text
requests/
├── users/
│   ├── list.http
│   └── get.http
└── posts/
    └── list.http
```

### YAML Format

```bash
restcli openapi2http spec.yaml -f yaml -o requests/
```

Generated file:

```yaml
name: List Users
method: GET
url: "{{baseUrl}}/users"
documentation:
  description: "List all users"
  tags:
    - Users
  responses:
    - code: 200
      description: "Success"
      contentType: "application/json"
```

### JSON Format

```bash
restcli openapi2http spec.json -f json -o requests/
```

Generated file:

```json
{
  "name": "List Users",
  "method": "GET",
  "url": "{{baseUrl}}/users",
  "documentation": {
    "description": "List all users",
    "tags": ["Users"],
    "responses": [
      {
        "code": 200,
        "description": "Success",
        "contentType": "application/json"
      }
    ]
  }
}
```

## Generated Content

### Variables

Converter creates `{{baseUrl}}` variable:

```text
GET {{baseUrl}}/users
```

Set in profile:

```json
{
  "variables": {
    "baseUrl": "https://api.example.com"
  }
}
```

### Path Parameters

```yaml
/users/{id}:
  get:
    parameters:
      - name: id
        in: path
```

Generates:

```text
GET {{baseUrl}}/users/{{id}}
```

### Query Parameters

```yaml
/users:
  get:
    parameters:
      - name: limit
        in: query
      - name: offset
        in: query
```

Generates:

```text
GET {{baseUrl}}/users?limit={{limit}}&offset={{offset}}
```

### Request Body

```yaml
/users:
  post:
    requestBody:
      content:
        application/json:
          schema:
            type: object
            properties:
              name:
                type: string
              email:
                type: string
```

Generates:

```text
POST {{baseUrl}}/users
Content-Type: application/json

{
  "name": "{{name}}",
  "email": "{{email}}"
}
```

### Headers

```yaml
/users:
  get:
    parameters:
      - name: Authorization
        in: header
        required: true
```

Generates:

```text
GET {{baseUrl}}/users
Authorization: {{Authorization}}
```

### Documentation

Full documentation embedded in request files:

```yaml
documentation:
  description: "Create a new user"
  tags:
    - Users
  parameters:
    - name: name
      type: string
      required: true
      description: "User's full name"
      example: "John Doe"
    - name: email
      type: string
      required: true
      description: "User's email address"
      example: "john@example.com"
  responses:
    - code: 201
      description: "User created successfully"
      contentType: "application/json"
      example: |
        {
          "id": 1,
          "name": "John Doe",
          "email": "john@example.com"
        }
    - code: 400
      description: "Invalid input"
```

## Workflow

### Generate Requests

```bash
restcli openapi2http api.yaml --organize-by tags -o requests/
```

### Create Profile

Create `.profiles.json`:

```json
[
  {
    "name": "Dev",
    "variables": {
      "baseUrl": "https://dev.api.example.com"
    }
  },
  {
    "name": "Prod",
    "variables": {
      "baseUrl": "https://api.example.com"
    }
  }
]
```

## Limitations

1. Complex schemas may need manual adjustment
2. Authentication configuration not auto-generated
3. Some OpenAPI features unsupported
4. Generated examples use placeholder values
