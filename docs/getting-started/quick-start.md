---
title: Quick Start
---

# Quick Start (CLI)

## First Request

Create a file `get-user.http`:

```text
### Get User
GET https://jsonplaceholder.typicode.com/users/1
```

## TUI Mode

Start the interactive TUI:

```bash
restcli
```

Navigate with arrow keys or `j`/`k`.

Press `Enter` to execute the selected request.

## CLI Mode

Execute directly:

```bash
restcli run get-user.http
```

Or use the shorthand:

```bash
restcli get-user.http
```

File extension is optional. This works too:

```bash
restcli get-user
```

## With Variables

Create `get-post.http`:

```text
### Get Post
GET https://jsonplaceholder.typicode.com/posts/{{postId}}
```

The tool prompts for missing variables in CLI mode.

Or set via flag:

```bash
restcli get-post -e postId=5
```

## With Profile

> The easiest way is to manage the profile(s) using the **TUI**

Create `~/.restcli/.profiles.json`:

```json
[
  {
    "name": "Default",
    "variables": {
      "baseUrl": "https://jsonplaceholder.typicode.com",
      "postId": "1"
    }
  }
]
```

Update `get-post.http`:

```text
### Get Post
GET {{baseUrl}}/posts/{{postId}}
```

Run with profile:

```bash
restcli run get-post -p Default
```

## POST Request

Create `create-post.http`:

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

Execute:

```bash
restcli create-post
```

## Output Formats

JSON output:

```bash
restcli get-user -o json
```

YAML output:

```bash
restcli get-user -o yaml
```

Full output (status, headers, body):

```bash
restcli get-user -f
```

## Save Response

Save to file:

```bash
restcli get-user -s response.json
```

## Next Steps

1. Learn about [file formats](../guides/file-formats.md)
2. Explore [variables](../guides/variables.md)
3. Set up [profiles](../guides/profiles.md)
4. Try [TUI mode](../guides/tui-mode.md) features
