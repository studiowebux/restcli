# Installation Guide

## Quick Start

### Compiled Binary (Production)

```bash
# Build the binaries
deno task build:all
chmod +x bin/restcli*

# This creates:
# - restcli          (TUI & CLI runner)
# - restcli-curl2http (cURL converter)

# Move to your PATH (optional)
mv bin/restcli* /usr/local/bin/

# Or create symlinks
ln -s $(pwd)/bin/restcli /usr/local/bin/restcli
ln -s $(pwd)/bin/restcli-curl2http /usr/local/bin/restcli-curl2http
```

## First Run

When you run `restcli` for the first time:

1. **If `~/.restcli/` doesn't exist and no local `requests/` folder:**
   - Automatically initializes `~/.restcli/`
   - Creates example configuration files
   - Creates sample request files
   - You can start using it immediately!

2. **If `~/.restcli/` already exists:**
   - Uses global configuration automatically
   - Works from any directory!

## Manual Initialization

If you want to manually initialize or migrate:

```bash
restcli init  # if binary is in PATH
```

## Configuration Directory Structure

After initialization, your `~/.restcli/` will contain:

```
~/.restcli/
├── .session.json          # Active session, variables, selected profile
├── .profiles.json         # Header profiles for different environments
└── history/               # Request/response history (if enabled)
```

## Uninstall

```bash
# Remove global config
rm -rf ~/.restcli

# Remove binaries (if installed globally)
sudo rm /usr/local/bin/restcli*

# Or just delete from your project directory
rm restcli*
```
