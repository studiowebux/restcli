---
title: GraphQL Support
tags:
  - guide
---

# GraphQL Support

REST CLI supports GraphQL queries, mutations, and subscriptions with automatic request formatting and response parsing.

## Quick Start

Add `# @protocol graphql` to any request:

```http
### Get Users
# @protocol graphql
POST https://api.example.com/graphql

{ users { id name email } }
```

REST CLI automatically:

- Wraps query in `{"query": "..."}` JSON format
- Sets `Content-Type: application/json`
- Formats GraphQL responses for readability
- Separates `data` from `errors`

## Basic Syntax

### Simple Query

```http
### Get Countries
# @protocol graphql
POST https://countries.trevorblades.com/graphql

{ countries { code name capital } }
```

### Query with Parameters

```http
### Get Specific Country
# @protocol graphql
POST https://countries.trevorblades.com/graphql

{ country(code: "US") { name capital currency } }
```

### With Variables (Inline)

```http
### Get Country by Code
# @protocol graphql
POST https://countries.trevorblades.com/graphql

query GetCountry {
  country(code: "{{countryCode}}") {
    name
    capital
    emoji
  }
}
```

Variables defined in profile or CLI:

```bash
restcli -e countryCode=CA get-country.http
```

### Mutations

```http
### Create User
# @protocol graphql
POST https://api.example.com/graphql
Authorization: Bearer {{token}}

mutation {
  createUser(input: {
    name: "Alice"
    email: "alice@example.com"
  }) {
    id
    name
    createdAt
  }
}
```

## Request Format

### Protocol Annotation

**Required:** Add `# @protocol graphql` before the request.

```http
### Query Name
# @protocol graphql
POST https://your-graphql-endpoint/graphql

{ your query here }
```

### HTTP Method

Use `POST` (GraphQL standard):

```http
POST https://api.example.com/graphql
```

### Headers

Add authentication or custom headers as normal:

```http
### Authenticated Query
# @protocol graphql
POST https://api.example.com/graphql
Authorization: Bearer {{apiKey}}
X-Request-ID: {{requestId}}

{ viewer { login name } }
```

### Query Body

Write GraphQL query directly in the body (no JSON wrapping needed):

```http
{
  users(limit: 10) {
    id
    name
    email
  }
}
```

Or with operation name:

```http
query GetUsers {
  users(limit: 10) {
    id
    name
  }
}
```

Or mutation:

```http
mutation CreatePost {
  createPost(title: "Hello", content: "World") {
    id
    title
  }
}
```

## Response Handling

### Successful Response

REST CLI extracts and formats the `data` field:

**GraphQL Server Response:**

```json
{
  "data": {
    "users": [
      { "id": 1, "name": "Alice" },
      { "id": 2, "name": "Bob" }
    ]
  }
}
```

**Displayed:**

```json
{
  "users": [
    {
      "id": 1,
      "name": "Alice"
    },
    {
      "id": 2,
      "name": "Bob"
    }
  ]
}
```

### Errors

GraphQL errors are formatted for readability:

**GraphQL Server Response:**

```json
{
  "data": null,
  "errors": [
    {
      "message": "Field 'invalidField' not found",
      "locations": [{ "line": 2, "column": 3 }],
      "path": ["user", "invalidField"]
    }
  ]
}
```

**Displayed:**

```json
{
  "data": null,
  "errors": [
    {
      "message": "Field 'invalidField' not found",
      "locations": [
        {
          "line": 2,
          "column": 3
        }
      ],
      "path": ["user", "invalidField"]
    }
  ]
}
```

### Partial Data with Errors

When GraphQL returns both data and errors:

```json
{
  "data": {
    "user": {
      "name": "Alice",
      "email": null
    }
  },
  "errors": [
    {
      "message": "Not authorized to access email",
      "path": ["user", "email"]
    }
  ]
}
```

Both fields are displayed.

## Variables and Profiles

### With REST CLI Variables

```http
### Get User
# @protocol graphql
POST https://api.example.com/graphql
Authorization: Bearer {{apiToken}}

query {
  user(id: {{userId}}) {
    id
    name
  }
}
```

Profile:

```json
{
  "name": "API",
  "variables": {
    "apiToken": "secret123",
    "userId": "42"
  }
}
```

### Interactive Variables for GraphQL

Use interactive variables for dynamic queries:

```json
{
  "name": "GraphQL API",
  "variables": {
    "searchQuery": {
      "value": "",
      "interactive": true
    },
    "apiKey": {
      "value": "sk-...",
      "interactive": true
    }
  }
}
```

Request:

```http
### Search
# @protocol graphql
POST https://api.example.com/graphql
Authorization: Bearer {{apiKey}}

query {
  search(query: "{{searchQuery}}") {
    results {
      title
      description
    }
  }
}
```

On execution, prompts for both `searchQuery` and `apiKey`.

## Filtering and Querying

Apply JMESPath filters to GraphQL responses:

```http
### Get Users (filtered)
# @protocol graphql
# @filter users[?age > `25`]
POST https://api.example.com/graphql

{ users { id name age } }
```

## Advanced Features

### Fragments

```http
### Get Users with Fragments
# @protocol graphql
POST https://api.example.com/graphql

query {
  users {
    ...UserFields
  }
}

fragment UserFields on User {
  id
  name
  email
  createdAt
}
```

### Named Operations

```http
### Multiple Operations
# @protocol graphql
POST https://api.example.com/graphql

query GetUser {
  user(id: 1) { name }
}

query GetPosts {
  posts(limit: 5) { title }
}
```

### Directives

```http
### Conditional Fields
# @protocol graphql
POST https://api.example.com/graphql

query GetUser($withEmail: Boolean!) {
  user(id: 1) {
    name
    email @include(if: $withEmail)
  }
}
```

## Subscriptions

GraphQL subscriptions (WebSocket-based) are planned for future release.

Current workaround: Use SSE-based subscriptions if your API supports them:

```http
### Subscribe to Events
# @streaming true
GET https://api.example.com/graphql/subscriptions?query={{encodedQuery}}
Accept: text/event-stream
```

## Introspection

Query the GraphQL schema:

```http
### Schema Types
# @protocol graphql
POST https://countries.trevorblades.com/graphql

{
  __schema {
    types {
      name
      kind
      description
    }
  }
}
```

```http
### Query Fields
# @protocol graphql
POST https://countries.trevorblades.com/graphql

{
  __type(name: "Query") {
    name
    fields {
      name
      description
      type { name kind }
    }
  }
}
```

## Comparison with HTTP

| Feature             | HTTP   | GraphQL                                    |
| ------------------- | ------ | ------------------------------------------ |
| Protocol annotation | None   | `# @protocol graphql`                      |
| Method              | Any    | Always `POST`                              |
| Body format         | Raw    | Wrapped in `{"query": "..."}`              |
| Response            | Direct | Extracts `data`, formats `errors`          |
| Headers             | As-is  | Auto-adds `Content-Type: application/json` |
