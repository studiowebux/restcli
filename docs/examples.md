---
title: Examples
---


# Examples

Practical examples for common use cases.

## Basic Requests

### Simple GET

```text
### Get User
GET https://jsonplaceholder.typicode.com/users/1
```

Execute:

```bash
restcli get-user.http
```

### POST with JSON Body

```text
### Create Post
POST https://jsonplaceholder.typicode.com/posts
Content-Type: application/json

{
  "title": "New Post",
  "body": "Content here",
  "userId": 1
}
```

### PUT Request

```text
### Update Post
PUT https://jsonplaceholder.typicode.com/posts/1
Content-Type: application/json

{
  "id": 1,
  "title": "Updated Title",
  "body": "Updated content",
  "userId": 1
}
```

### DELETE Request

```text
### Delete Post
DELETE https://jsonplaceholder.typicode.com/posts/1
```

## Variables

### Basic Substitution

Request:

```text
### Get Post
GET https://jsonplaceholder.typicode.com/posts/{{postId}}
```

Execute:

```bash
restcli get-post.http -e postId=5
```

### Multiple Variables

Request:

```text
### Get User Post
GET {{baseUrl}}/users/{{userId}}/posts/{{postId}}
```

Profile:

```json
{
  "name": "API",
  "variables": {
    "baseUrl": "https://jsonplaceholder.typicode.com",
    "userId": "1",
    "postId": "1"
  }
}
```

Execute:

```bash
restcli -p API get-user-post.http
```

### Shell Command Variables

Request:

```text
### Create Event
POST https://api.example.com/events
Content-Type: application/json

{
  "timestamp": {{timestamp}},
  "branch": "{{branch}}",
  "uuid": "{{uuid}}"
}
```

Profile:

```json
{
  "name": "Events",
  "variables": {
    "timestamp": "$(date +%s)",
    "branch": "$(git branch --show-current)",
    "uuid": "$(uuidgen)"
  }
}
```

### Environment Variables

Request:

```text
### Get Todo
GET https://jsonplaceholder.typicode.com/todos/{{env.TODO_ID}}
```

Execute:

```bash
export TODO_ID=5
restcli get-todo.http
```

Or with file:

```bash
restcli --env-file .env get-todo.http
```

## Filtering and Querying

### JMESPath Filter

Request:

```text
### Get Users
# @filter users[?id > `5`]
GET https://jsonplaceholder.typicode.com/users
```

Or CLI:

```bash
restcli get-users.http --filter "users[?id > \`5\`]"
```

### JMESPath Query

Request:

```text
### Get User Names
# @query [].name
GET https://jsonplaceholder.typicode.com/users
```

### Filter and Query Combined

Request:

```text
### Get Active User Emails
# @filter users[?active == `true`]
# @query [].email
GET https://api.example.com/users
```

### Bash Command Query

Request (JSON):

```json
{
  "name": "Get Names with jq",
  "method": "GET",
  "url": "https://jsonplaceholder.typicode.com/users",
  "query": "$(jq '.[].name')"
}
```

## Authentication

### Bearer Token

Request:

```text
### Get Protected Data
GET https://api.example.com/protected
Authorization: Bearer {{token}}
```

Profile:

```json
{
  "name": "Auth",
  "variables": {
    "token": "your-token-here"
  }
}
```

### API Key Header

Request:

```text
### API Call
GET https://api.example.com/data
X-API-Key: {{apiKey}}
```

Profile:

```json
{
  "name": "API",
  "headers": {
    "X-API-Key": "{{apiKey}}"
  },
  "variables": {
    "apiKey": "secret-key-123"
  }
}
```

### Basic Auth

Request:

```text
### Basic Auth
GET https://api.example.com/data
Authorization: Basic {{basicAuth}}
```

Profile:

```json
{
  "variables": {
    "basicAuth": "$(echo -n 'user:pass' | base64)"
  }
}
```

## mTLS

### Profile Configuration

Profile:

```json
{
  "name": "Secure",
  "variables": {
    "baseUrl": "https://secure.api.example.com"
  },
  "tls": {
    "certFile": "/path/to/client.crt",
    "keyFile": "/path/to/client.key",
    "caFile": "/path/to/ca.crt"
  }
}
```

Request:

```text
### Secure Call
GET {{baseUrl}}/data
```

### Per-Request TLS

YAML:

```yaml
name: Secure Request
method: GET
url: "https://secure.api.example.com/data"
tls:
  certFile: "/path/to/client.crt"
  keyFile: "/path/to/client.key"
  caFile: "/path/to/ca.crt"
```

## File Formats

### HTTP Format

```text
### Get User
GET https://jsonplaceholder.typicode.com/users/1
Accept: application/json

### Create User
POST https://jsonplaceholder.typicode.com/users
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com"
}
```

### YAML Format

```yaml
name: Get User
method: GET
url: "https://jsonplaceholder.typicode.com/users/1"
headers:
  Accept: "application/json"
```

### JSON Format

```json
{
  "name": "Get User",
  "method": "GET",
  "url": "https://jsonplaceholder.typicode.com/users/1",
  "headers": {
    "Accept": "application/json"
  }
}
```

### JSONC Format

```jsonc
{
  // This is a comment
  "name": "Get Post",
  "method": "GET",
  "url": "https://jsonplaceholder.typicode.com/posts/1",

  /* Multi-line
   * comment
   */
  "headers": {
    "Accept": "application/json"
  }
}
```

## Multi-Value Variables

Profile:

```json
{
  "name": "Multi-Env",
  "variables": {
    "baseUrl": {
      "options": [
        "http://localhost:3000",
        "https://dev.api.example.com",
        "https://api.example.com"
      ],
      "active": 0,
      "description": "API environment",
      "aliases": {
        "local": 0,
        "dev": 1,
        "prod": 2
      }
    },
    "apiVersion": {
      "options": ["v1", "v2", "v3"],
      "active": 1,
      "aliases": {
        "legacy": 0,
        "current": 1,
        "beta": 2
      }
    }
  }
}
```

Request:

```text
### API Call
GET {{baseUrl}}/{{apiVersion}}/users
```

Execute with alias:

```bash
restcli -p Multi-Env -e baseUrl=prod -e apiVersion=beta api-call.http
```

## Workflows

### API Testing Workflow

1. Create profile:

```json
{
  "name": "Testing",
  "variables": {
    "baseUrl": "http://localhost:3000"
  }
}
```

2. Create login request:

```text
### Login
POST {{baseUrl}}/auth/login
Content-Type: application/json

{
  "username": "test",
  "password": "pass"
}
```

3. Execute and extract token (TUI auto-extracts)

4. Use token in next request:

```text
### Get Profile
GET {{baseUrl}}/profile
Authorization: Bearer {{token}}
```

### CI/CD Integration

Script:

```bash
#!/bin/bash

# Run health check
restcli health.http -o json > health.json

# Check status
if ! jq -e '.status == "ok"' health.json > /dev/null; then
  echo "Health check failed"
  exit 1
fi

# Run tests
restcli test-suite.http -p CI -o json > results.json

# Validate
jq -e '.success == true' results.json
```

### Data Migration

Extract IDs:

```bash
restcli get-users.http -o json --query '[].id' > user-ids.json
```

Process each:

```bash
cat user-ids.json | jq -r '.[]' | while read id; do
  restcli migrate-user.http -e userId=$id
done
```

### Response Comparison

TUI workflow:

1. Execute request: `Enter`
2. Pin response: `w`
3. Make code change
4. Execute again: `Enter`
5. View diff: `W`

### Documentation Workflow

Create request with docs:

```yaml
name: Create User
method: POST
url: "{{baseUrl}}/users"
documentation:
  description: "Create a new user account"
  tags:
    - Users
    - POST
  parameters:
    - name: name
      type: string
      required: true
      description: "User's full name"
      example: "John Doe"
    - name: email
      type: string
      required: true
      description: "User's email"
      example: "john@example.com"
  responses:
    - code: "201"
      description: "User created"
      contentType: "application/json"
      fields:
        - name: id
          type: number
          required: true
          description: "User ID"
        - name: name
          type: string
          required: true
        - name: email
          type: string
          required: true
      example: |
        {
          "id": 1,
          "name": "John Doe",
          "email": "john@example.com"
        }
```

View in TUI: press `m`

## Advanced

### Stdin Body

```bash
cat payload.json | restcli create-user.http
```

### Piping Responses

```bash
restcli get-data.http -o json | jq '.items[].name'
```

### Chaining Requests

```bash
TOKEN=$(restcli login.http -o json | jq -r '.token')
restcli -e token=$TOKEN get-profile.http
```

### Conditional Execution

```bash
if restcli health.http --query '.status' | grep -q 'ok'; then
  restcli deploy.http
else
  echo "Service unhealthy"
  exit 1
fi
```

### Loop Processing

```bash
for env in dev staging prod; do
  restcli check-status.http -e environment=$env -o json >> status-$env.json
done
```
