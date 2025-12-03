---
title: CLI Mode
tags:
  - guide
---

# CLI Mode

Execute requests from the command line.

## Basic Usage

```bash
restcli <file>
```

Or explicit:

```bash
restcli run <file>
```

File extension is optional:

```bash
restcli get-user
```

Finds `get-user.http`, `get-user.yaml`, `get-user.json`, or `get-user.jsonc`.

## Flags

### Profile

```bash
restcli -p Dev request.http
```

Short: `-p`
Long: `--profile`

### Output Format

```bash
restcli -o json request.http
```

Options: `json`, `yaml`, `text`

Short: `-o`
Long: `--output`

### Full Output

```bash
restcli -f request.http
```

Includes status line, headers, and body.

Short: `-f`
Long: `--full`

### Save Response

```bash
restcli -s response.json request.http
```

Short: `-s`
Long: `--save`

### Override Body

```bash
restcli -b '{"key":"value"}' request.http
```

Short: `-b`
Long: `--body`

### Set Variables

```bash
restcli -e userId=123 -e token=abc request.http
```

Repeatable flag for multiple variables.

Short: `-e`
Long: `--extra-vars`

### Environment File

```bash
restcli --env-file .env request.http
```

Load variables from file.

### Filter

```bash
restcli --filter "items[?status==\`active\`]" request.http
```

JMESPath filter expression.

### Query

```bash
restcli -q "[].name" request.http
```

JMESPath query or bash command.

Short: `-q`
Long: `--query`

## Stdin Body

Pipe data directly:

```bash
cat payload.json | restcli request.http
```

Or:

```bash
echo '{"data":"value"}' | restcli post-data.http
```

Stdin overrides body in request file.

## Combining Flags

```bash
restcli run api.http \
  -p Production \
  -o json \
  -e apiKey=secret123 \
  -e version=v2 \
  --filter "results[?active]" \
  -s output.json
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Request failed or error |
| 2 | Missing variables |

## Scripting

### Check Response

```bash
restcli health.http -o json | jq '.status == "ok"'
```

### Variable Extraction

```bash
TOKEN=$(restcli login.http -o json | jq -r '.token')
restcli -e token=$TOKEN get-data.http
```

### Loop Requests

```bash
for id in {1..10}; do
  restcli get-user.http -e userId=$id -o json >> results.jsonl
done
```

### Conditional Execution

```bash
if restcli health.http -q '.status' | grep -q 'ok'; then
  restcli deploy.http
fi
```

## Interactive Variables

If variables are missing and not provided via flags, CLI prompts for input:

```bash
restcli get-user.http
# Prompts: Enter value for userId:
```

Disable prompts in scripts by setting all variables via flags or profiles.

## Examples

Basic request:

```bash
restcli get-user.http
```

With profile and variables:

```bash
restcli -p Dev -e userId=5 get-user.http
```

JSON output saved to file:

```bash
restcli api.http -o json -s result.json
```

Filter and query:

```bash
restcli users.http \
  --filter "users[?age > \`18\`]" \
  --query "[].{name: name, email: email}"
```

Pipe stdin body:

```bash
cat new-user.json | restcli create-user.http -o json
```

Full output with headers:

```bash
restcli -f auth.http
```

## Confirmation Prompts

Requests with `@confirmation` directive require confirmation in CLI mode:

```http
### Delete User
# @confirmation
DELETE https://api.example.com/users/{{userId}}
```

When executed in CLI:

```bash
restcli delete-user.http -p prod
# Prompts: Request 'Delete User' requires confirmation.
# Method: DELETE
# URL: https://api.example.com/users/123
#
# Proceed? [y/N]:
```

Prevents accidental execution of critical endpoints.

## Shell Completion

Generate completions for your shell:

### Bash

```bash
# Generate completion script
restcli completion bash > /etc/bash_completion.d/restcli

# Or add to ~/.bashrc
source <(restcli completion bash)
```

### Zsh

```bash
# Create completions directory
mkdir -p ~/.zsh/completions

# Add to ~/.zshrc (if not already present)
if type restcli &>/dev/null; then
    echo 'fpath=(~/.zsh/completions $fpath)' >> ~/.zshrc
    echo 'autoload -Uz compinit && compinit' >> ~/.zshrc
fi

# Generate completions
restcli completion zsh > ~/.zsh/completions/_restcli

# Reload shell
source ~/.zshrc
```

### Features

Completion provides:
- **Profile names** for `--profile` flag (from `.profiles.json`)
- **Request files** for file argument (from profile's workdir or current directory)
- **Flag names** and options

Example:

```bash
restcli run <TAB>          # Shows .http files
restcli -p <TAB>           # Shows profile names
restcli run api -p <TAB>   # Shows profiles
```

**Note**: File completion is profile-aware. With `--profile dev`, it scans the `dev` profile's workdir for valid files.
