# Filtering and Querying

Transform response data using JMESPath expressions or bash/Linux commands.

## Filter vs Query

**Filter**: Narrow down results (subset of data)

**Query**: Transform structure (reshape data)

Both work together in sequence: filter first, then query.

## Supported Formats

Two options available:

1. **JMESPath**: AWS CLI-style expressions for JSON
2. **Bash/Linux commands**: Any shell command using `$(command)` syntax

Both work in CLI flags, request files, and profile defaults.

## JMESPath Syntax

AWS CLI-style expressions for JSON transformation.

### CLI Usage

Filter:

```bash
restcli --filter "items[?status==\`active\`]" request.http
```

Query:

```bash
restcli --query "[].name" request.http
```

Or short form:

```bash
restcli -q "[].name" request.http
```

Combined:

```bash
restcli --filter "users[?age>\`18\`]" --query "[].email" request.http
```

### Request Files

HTTP format:

```text
### Get Active Users
# @filter users[?active==`true`]
# @query [].{name: name, email: email}
GET https://api.example.com/users
```

YAML format:

```yaml
name: Get Active Users
method: GET
url: "https://api.example.com/users"
filter: "users[?active==`true`]"
query: "[].{name: name, email: email}"
```

JSON format:

```json
{
  "name": "Get Active Users",
  "method": "GET",
  "url": "https://api.example.com/users",
  "filter": "users[?active==`true`]",
  "query": "[].{name: name, email: email}"
}
```

### Profile Defaults

In `.profiles.json`:

```json
{
  "name": "API",
  "defaultFilter": "results[?status==`active`]",
  "defaultQuery": "[].{id: id, name: name}"
}
```

Applied to all requests unless overridden.

## JMESPath Examples

### Basic Selection

Select field:

```text
data.users
```

Array element:

```text
users[0]
```

Multiple fields:

```text
{name: name, email: email}
```

### Filtering

Active items:

```text
items[?active==`true`]
```

Age greater than 18:

```text
users[?age>`18`]
```

Status equals value:

```text
results[?status==`completed`]
```

### Projection

Extract field from array:

```text
users[].name
```

Transform array:

```text
users[].{id: id, name: name, email: contact.email}
```

### Functions

Length:

```text
length(items)
```

Sort:

```text
sort_by(users, &age)
```

Max value:

```text
max_by(products, &price)
```

## Bash/Linux Command Syntax

Use `$(command)` to pipe response through any bash/Linux command.

Response is piped to stdin of the command.

### CLI Usage

With jq:

```bash
restcli --query '$(jq ".items[].name")' request.http
```

With grep:

```bash
restcli --query '$(grep -o "id.*")' request.http
```

With filter and query:

```bash
restcli --filter '$(jq ".items")' --query '$(jq ".[].name")' request.http
```

### Request Files

HTTP format:

```text
### Get Names
# @query $(jq '.users[].name')
GET https://api.example.com/users
```

YAML format:

```yaml
name: Get Names
method: GET
url: "https://api.example.com/users"
query: "$(jq '.users[].name')"
```

JSON format:

```json
{
  "name": "Get Names",
  "method": "GET",
  "url": "https://api.example.com/users",
  "query": "$(jq '.users[].name')"
}
```

### Examples

Extract with jq:

```bash
$(jq '.items[].name')
```

Grep pattern:

```bash
$(grep -o '"id":[0-9]*')
```

Count lines:

```bash
$(wc -l)
```

Format with awk:

```bash
$(awk '{print $1, $3}')
```

Sort results:

```bash
$(sort)
```

Unique values:

```bash
$(sort | uniq)
```

Chain commands:

```bash
$(jq '.items[]' | grep active | wc -l)
```

Complex pipeline:

```bash
$(jq -r '.users[].email' | sort | uniq | wc -l)
```

## Priority

Filters and queries apply in this order:

1. CLI flags (highest)
2. Request file
3. Profile defaults (lowest)

Higher priority overrides lower.

## Examples

### Filter Active Users

Response:

```json
{
  "users": [
    {"id": 1, "name": "Alice", "active": true},
    {"id": 2, "name": "Bob", "active": false},
    {"id": 3, "name": "Charlie", "active": true}
  ]
}
```

Filter:

```text
users[?active==`true`]
```

Result:

```json
[
  {"id": 1, "name": "Alice", "active": true},
  {"id": 3, "name": "Charlie", "active": true}
]
```

### Transform Structure

Response:

```json
{
  "data": {
    "users": [
      {"id": 1, "name": "Alice", "contact": {"email": "alice@example.com"}},
      {"id": 2, "name": "Bob", "contact": {"email": "bob@example.com"}}
    ]
  }
}
```

Query:

```text
data.users[].{id: id, name: name, email: contact.email}
```

Result:

```json
[
  {"id": 1, "name": "Alice", "email": "alice@example.com"},
  {"id": 2, "name": "Bob", "email": "bob@example.com"}
]
```

### Filter and Query Combined

Response:

```json
{
  "products": [
    {"id": 1, "name": "Widget", "price": 10, "inStock": true},
    {"id": 2, "name": "Gadget", "price": 20, "inStock": false},
    {"id": 3, "name": "Tool", "price": 15, "inStock": true}
  ]
}
```

Filter:

```text
products[?inStock==`true`]
```

Query:

```text
[].{name: name, price: price}
```

Result:

```json
[
  {"name": "Widget", "price": 10},
  {"name": "Tool", "price": 15}
]
```

### Using jq for Complex Logic

Response:

```json
{
  "items": [
    {"category": "A", "values": [1, 2, 3]},
    {"category": "B", "values": [4, 5, 6]}
  ]
}
```

Query with bash:

```text
$(jq '[.items[] | select(.category == "A") | .values[]] | add')
```

Result:

```text
6
```

### Profile Default Applied

Profile:

```json
{
  "name": "API",
  "defaultFilter": "results[?status==`success`]",
  "defaultQuery": "[].{id: id, message: message}"
}
```

Request:

```text
### Check Results
GET https://api.example.com/results
```

Filter and query applied automatically from profile.

Override in request:

```text
### Check All Results
# @filter ""
# @query ""
GET https://api.example.com/results
```

Empty values disable profile defaults for this request.

## Resources

JMESPath reference: https://jmespath.org
