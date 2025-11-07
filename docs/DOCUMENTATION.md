# Request Documentation

You can add documentation to your HTTP requests to describe endpoints, parameters, and expected responses. This documentation can be viewed in the TUI and used for generating API documentation.

## Formats

Documentation is supported in both `.http` and `.yaml` file formats.

### .http File Format (Annotations)

Use `# @` comments before your request to add documentation:

```text
### Create User
# @description Creates a new user in the system with the provided information
# @tag users
# @tag authentication
# @param name {string} required - User's full name
# @param email {string} required - User's email address
# @param age {number} optional - User's age
# @example name "John Doe"
# @example email "john@example.com"
# @example age 30
# @response 201 - User created successfully
# @response 400 - Invalid request data
# @response 409 - User already exists
POST {{baseUrl}}/api/v1/users
Content-Type: application/json
Authorization: Bearer {{token}}

{
  "name": "{{name}}",
  "email": "{{email}}",
  "age": {{age}}
}
```

**Supported Annotations:**

- `@description` - Description of what the endpoint does
- `@tag` - Category or tag for organizing endpoints (can have multiple)
- `@param` - Parameter definition
  - Format: `@param <name> {<type>} <required|optional> - <description>`
- `@example` - Example value for a parameter
  - Format: `@example <paramName> <value>`
- `@response` - Expected response
  - Format: `@response <code> - <description>`

### YAML File Format (Structured)

Add a `documentation` section to your YAML request:

```yaml
---
name: Create User
method: POST
url: "{{baseUrl}}/api/v1/users"
documentation:
  description: Creates a new user in the system with the provided information
  tags:
    - users
    - authentication
  parameters:
    - name: name
      type: string
      required: true
      description: User's full name
      example: "John Doe"
    - name: email
      type: string
      required: true
      description: User's email address
      example: "john@example.com"
    - name: age
      type: number
      required: false
      description: User's age
      example: 30
  responses:
    - "201": "User created successfully"
    - "400": "Invalid request data"
    - "409": "User already exists"
headers:
  Content-Type: application/json
  Authorization: "Bearer {{token}}"
body: |
  {
    "name": "{{name}}",
    "email": "{{email}}",
    "age": {{age}}
  }
```

**YAML Documentation Structure:**

```yaml
documentation:
  description: string           # What the endpoint does
  tags:                         # Categories/tags
    - string
  parameters:                   # Request parameters
    - name: string              # Parameter name
      type: string              # Type (string, number, boolean, object, array)
      required: boolean         # Is required?
      description: string       # What this parameter does
      example: any              # Example value
  responses:                    # Expected responses
    - "code": "description"     # HTTP status code and description
```

## Parameter Types

Common parameter types:
- `string` - Text values
- `number` - Numeric values
- `boolean` - true/false
- `object` - JSON objects
- `array` - Arrays/lists

## Viewing Documentation

In the TUI, documentation for the current request is displayed in the documentation panel (press `m` to toggle).
