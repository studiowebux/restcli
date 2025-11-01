# Installation Guide

## Quick Start

### Option 1: Using Deno (Development)

```bash
# Clone or download the repository
git clone <repo-url>
cd http

# Run directly with Deno
deno task dev        # Start TUI
deno task run <file> # Run a request file
deno task curl2http  # Convert cURL commands
```

### Option 2: Compiled Binary (Production)

```bash
# Build the binaries
deno task build:all
chmod +x bin/restcli*

# This creates:
# - restcli          (TUI & CLI runner)
# - restcli-curl2http (cURL converter)

# Move to your PATH (optional)
sudo mv bin/restcli* /usr/local/bin/

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

2. **If running from project directory with `requests/` folder:**
   - Uses local configuration (backward compatible)
   - Shows a tip about migrating to `~/.restcli/`

3. **If `~/.restcli/` already exists:**
   - Uses global configuration automatically
   - Works from any directory!

## Manual Initialization

If you want to manually initialize or migrate:

```bash
# Initialize fresh config
deno task init
# or
restcli init  # if binary is in PATH

# Migrate from current directory to ~/.restcli/
deno task init --migrate
# or
restcli init --migrate

# Force re-initialization
deno task init --force
```

## Configuration Directory Structure

After initialization, your `~/.restcli/` will contain:

```
~/.restcli/
├── .session.json          # Active session, variables, selected profile
├── .profiles.json         # Header profiles for different environments
├── requests/              # Your .http request files
│   └── example.http      # Example request file
└── history/              # Request/response history (if enabled)
```

## Working from Anywhere

Once `~/.restcli/` is initialized:

```bash
# You can run from any directory!
cd ~/projects/my-app
restcli  # Uses ~/.restcli/ config

cd /tmp
restcli  # Still uses ~/.restcli/ config

# Add requests from anywhere
restcli-curl2http 'curl http://api.example.com/users' -o ~/.restcli/requests/users.http
```

## Local Development Mode

For project-specific configurations, create a local `requests/` folder:

```bash
mkdir requests
# Add your .http files
# Create local .session.json and .profiles.json

deno task dev  # Uses local config instead of ~/.restcli/
```

This is useful for:
- Project-specific API endpoints
- Sharing request collections via git
- Testing different configurations

## Migration

To migrate from local to global config:

```bash
# From your project directory
deno task init --migrate

# This copies:
# - .session.json → ~/.restcli/.session.json
# - .profiles.json → ~/.restcli/.profiles.json
# - requests/ → ~/.restcli/requests/
# - history/ → ~/.restcli/history/
```

After migration, you can delete the local files if desired.

## Building from Source

Requirements:
- [Deno](https://deno.land/) 1.37+

```bash
# Install dependencies (automatic on first run)
deno cache src/tui.ts

# Build all binaries
deno task build:all
chmod +x bin/restcli*

# Or build individually
deno task build  # Just the TUI & CLI
deno compile --allow-read --allow-write --allow-env --output restcli-curl2http scripts/curl2http.ts
```

## Distribution

To distribute the compiled binaries:

1. Build for your platform: `deno task build:all`
2. Binaries are platform-specific (macOS, Linux, Windows)
3. Users just need to:
   - Download the binary
   - Make it executable (`chmod +x restcli`)
   - Run it (auto-initializes on first run)

No Deno installation required for end users!

## Uninstall

```bash
# Remove global config
rm -rf ~/.restcli

# Remove binaries (if installed globally)
sudo rm /usr/local/bin/restcli*

# Or just delete from your project directory
rm restcli*
```

## Troubleshooting

### "No such device" error
This happens when running in non-interactive mode (background, piped, etc.). The TUI requires a terminal. Use `restcli` with a filepath for scripting instead.

### Config not found
If `~/.restcli/` exists but seems empty:
```bash
restcli init --force  # Recreate example files
```

### Permission denied
```bash
chmod +x bin/restcli*  # Make binaries executable
```

### Wrong config being used
Check priority order:
1. Local `requests/` folder (if exists) → uses current directory
2. `~/.restcli/` (if exists) → uses global config
3. Neither exists → auto-initializes `~/.restcli/`
