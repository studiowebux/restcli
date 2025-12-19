---
title: Request Chaining
tags:
  - guide
---

# Request Chaining

Execute requests in sequence with automatic dependency resolution and variable extraction.

## Overview

Request chaining allows you to:
1. Define dependencies between requests
2. Automatically extract values from responses
3. Pass extracted values to dependent requests
4. Execute requests in the correct order

## Annotations

### @depends

Specify files this request depends on. Dependencies execute first.

```http
# @depends login.http
# @depends chain/setup.http
GET https://api.example.com/user/profile
```

Multiple dependencies can be listed (space-separated):

```http
# @depends login.http setup.http
```

Or use multiple `@depends` lines:

```http
# @depends login.http
# @depends setup.http
```

### @extract

Extract values from response body using JMESPath and store in session variables.

```http
# @extract token access_token
# @extract userId user.id
POST https://api.example.com/auth/login
```

Format: `@extract <varName> <jmesPath>`

- `varName`: Variable name to store the extracted value
- `jmesPath`: JMESPath expression to extract the value from JSON response

## Basic Example

### Step 1: Login Request

`chain/login.http`:
```http
### Login to API
# @extract token access_token
# @extract userId user.id
POST https://api.example.com/auth/login
Content-Type: application/json

{
  "username": "user@example.com",
  "password": "secret"
}
```

Response:
```json
{
  "access_token": "eyJhbGc...",
  "user": {
    "id": 123,
    "name": "John Doe"
  }
}
```

Variables extracted:
- `token` = "eyJhbGc..."
- `userId` = 123

### Step 2: Dependent Request

`chain/get-profile.http`:
```http
### Get User Profile
# @depends chain/login.http
GET https://api.example.com/users/{{userId}}
Authorization: Bearer {{token}}
```

This request:
1. Waits for `chain/login.http` to execute
2. Uses extracted `token` and `userId` variables
3. Executes with resolved values

## Execution Flow

When you execute a request with dependencies:

1. **Graph Building**: System builds a dependency graph
2. **Topological Sort**: Determines execution order
3. **Sequential Execution**: Executes requests in dependency order
4. **Variable Extraction**: After each request, extracts variables using `@extract`
5. **Session Storage**: Stores extracted variables in session
6. **Variable Resolution**: Subsequent requests use extracted variables

## Multi-Level Dependencies

Chains can have multiple levels:

`chain/setup.http`:
```http
### Setup API
# @extract apiKey key
POST https://api.example.com/setup
```

`chain/login.http`:
```http
### Login
# @depends chain/setup.http
# @extract token access_token
POST https://api.example.com/auth/login
X-API-Key: {{apiKey}}
```

`chain/get-data.http`:
```http
### Get Data
# @depends chain/login.http
GET https://api.example.com/data
Authorization: Bearer {{token}}
X-API-Key: {{apiKey}}
```

Execution order: `setup.http` → `login.http` → `get-data.http`

## JMESPath Extraction Examples

### Simple Field

```http
# @extract userId id
```

Response: `{"id": 123}` → `userId = 123`

### Nested Field

```http
# @extract token user.credentials.access_token
```

Response:
```json
{
  "user": {
    "credentials": {
      "access_token": "abc123"
    }
  }
}
```
Result: `token = "abc123"`

### Array Element

```http
# @extract firstId items[0].id
```

Response: `{"items": [{"id": 1}, {"id": 2}]}` → `firstId = 1`

### Complex Query

```http
# @extract adminIds users[?role=='admin'].id
```

Response:
```json
{
  "users": [
    {"id": 1, "role": "admin"},
    {"id": 2, "role": "user"},
    {"id": 3, "role": "admin"}
  ]
}
```
Result: `adminIds = [1, 3]` (as JSON array string)

## TUI Usage

### Execute Chain

1. Select a file with `@depends` annotation
2. Press `Enter` to execute
3. System automatically:
   - Detects dependencies
   - Builds execution graph
   - Executes chain in order
   - Shows progress: "Executing chain: N requests"

### Chain Progress

Status bar shows:
- `Executing chain: 3 requests` - during execution
- `Chain completed: 3 requests executed` - on success
- `Request 2/3 (login.http) failed: ...` - on failure

### Final Response

After chain completion:
- Response panel shows the final request's response (the one you selected)
- All intermediate responses execute silently
- Extracted variables available in session

### View Extracted Variables

Press `v` to open variable editor and see extracted values in the session variables section.

## Error Handling

### Dependency Parse Error

```
Failed to parse dependency chain/login.http: invalid syntax
```

Fix the syntax error in the dependency file.

### Variable Extraction Error

```
Failed to extract variables from login.http: response is not valid JSON
```

Ensure response is valid JSON when using `@extract`.

### Circular Dependency

```
Circular dependency detected: profile.http depends on itself
```

Remove circular reference from `@depends` annotations.

### Missing Dependency

```
Failed to parse start file chain/missing.http: no such file
```

Ensure all files in `@depends` exist.

## Path Resolution

Dependency paths are resolved relative to the profile's `workdir`:

```json
{
  "profiles": [
    {
      "name": "dev",
      "workdir": "/Users/you/project/requests"
    }
  ]
}
```

If `workdir` is `/Users/you/project/requests`:
- `@depends login.http` → `/Users/you/project/requests/login.http`
- `@depends chain/setup.http` → `/Users/you/project/requests/chain/setup.http`

Absolute paths work too:
- `@depends /path/to/other/request.http`

## Limitations

1. **Single Request Per File**: Only first request in each file is used for chaining
2. **JSON Only**: Variable extraction requires JSON responses
3. **No Parallel Execution**: Dependencies execute sequentially
4. **No Conditional Chains**: All dependencies always execute
5. **Session Scope**: Extracted variables stored in session (cleared on profile switch)

## Best Practices

1. **Organize Chain Files**: Use subdirectory for chain requests (`chain/`, `workflows/`)

2. **Descriptive Names**: Name files clearly (`login.http`, `create-user.http`)

3. **Document Extractions**: Comment what each extraction does

   ```http
   ### Login
   # Extract access token for subsequent requests
   # @extract token access_token
   # Extract user ID for profile lookup
   # @extract userId user.id
   ```

4. **Minimal Dependencies**: Only depend on what's needed

5. **Error Responses**: Handle authentication failures gracefully

6. **Categories**: Tag chain files for filtering

   ```http
   # @category chain
   # @category auth
   ```

## Advanced Example: User Creation Workflow

`chain/admin-login.http`:
```http
### Admin Login
# @extract adminToken access_token
POST https://api.example.com/admin/login
Content-Type: application/json

{
  "username": "admin",
  "password": "{{adminPassword}}"
}
```

`chain/create-user.http`:
```http
### Create User
# @depends chain/admin-login.http
# @extract newUserId id
POST https://api.example.com/users
Authorization: Bearer {{adminToken}}
Content-Type: application/json

{
  "email": "{{userEmail}}",
  "name": "{{userName}}"
}
```

`chain/send-welcome.http`:
```http
### Send Welcome Email
# @depends chain/create-user.http
POST https://api.example.com/emails/welcome
Authorization: Bearer {{adminToken}}
Content-Type: application/json

{
  "userId": {{newUserId}},
  "template": "welcome"
}
```

Execute `send-welcome.http`:
1. Admin login executes → extracts `adminToken`
2. Create user executes → uses `adminToken`, extracts `newUserId`
3. Send welcome executes → uses both `adminToken` and `newUserId`

## Related Features

- [Variables](variables.md) - Variable substitution system
- [Profiles](profiles.md) - Profile and session management
- [Filtering](filtering.md) - JMESPath query syntax
