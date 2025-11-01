# Changelog

## Recent Updates

### Auto-Initialization & Global Config
- ✅ Automatic initialization on first run
- ✅ Global config directory (`~/.restcli/`)
- ✅ Use from any directory after initialization
- ✅ Backward compatible with local development
- ✅ Manual migration with `init --migrate`

### Profile Enhancements
- ✅ Override profile for CLI runs: `--profile` or `-p` flag
- ✅ Per-profile working directories with `workdir` field
- ✅ Profile-specific variables that override global variables

### History System
- ✅ Request/response history with timestamps
- ✅ Saved as `{request-name}_{timestamp}.json` in `history/`
- ✅ Can be disabled via `.session.json` with `historyEnabled: false`
- ✅ Includes full request details, response, and duration

### YAML Support
- ✅ Auto-detect YAML/JSON format in request files
- ✅ Support single request format:
  ```yaml
  ---
  name: My Request
  method: GET
  url: "{{baseUrl}}/endpoint"
  headers:
    Authorization: "Bearer {{token}}"
  ```
- ✅ Support multiple requests format:
  ```yaml
  ---
  requests:
    - name: First Request
      method: GET
      url: "..."
    - name: Second Request
      method: POST
      url: "..."
  ```
- ✅ JSON schema for IDE autocomplete (`http-request.schema.json`)

### cURL Converter Improvements
- ✅ Specify output location: `--output` or `-o` flag
- ✅ Support directory paths with auto-filename generation
- ✅ Fixed TypeScript compilation error

### Build System
- ✅ Compile to standalone binaries (no Deno required)
- ✅ Three binaries: `restcli`, `restcli-run`, `restcli-curl2http`
- ✅ ~70MB binaries with all dependencies embedded
- ✅ Build tasks: `build`, `build:all`

### CLI Improvements
- ✅ `run.ts` now uses global config automatically
- ✅ `tui.ts` detects and uses `~/.restcli/` when available
- ✅ Helpful tips shown when config not initialized

## Usage Examples

### Global Config
```bash
# First run - auto-initializes
./restcli

# Manual init
deno task init

# Migrate existing config
deno task init --migrate
```

### Profile Override
```bash
# Use specific profile for one request
deno task run requests/auth/login.http --profile "Production - Admin"
./restcli-run requests/auth/login.http -p "Dev - Player 1"
```

### Custom Output Location
```bash
# Save to specific file
pbpaste | deno task curl2http -o requests/api/users.http

# Save to directory (auto-generates filename)
pbpaste | deno task curl2http --output requests/api/
```

### History Control
```json
{
  "variables": {...},
  "activeProfile": "Dev - Player 1",
  "historyEnabled": false  // Disable history
}
```

### Per-Profile Workdir
```json
{
  "name": "Admin Profile",
  "workdir": "requests/admin",
  "variables": {...},
  "headers": {...}
}
```

## File Structure

### Global Config (`~/.restcli/`)
```
~/.restcli/
├── .session.json          # Session state
├── .profiles.json         # Profiles configuration
├── requests/              # Request files (.http, .yaml)
│   ├── example.http
│   └── api/
└── history/               # Request/response history
    └── example_2025-10-31T19-54-09-924.json
```

### Local Development
```
./
├── .session.json          # Local session (optional)
├── .profiles.json         # Local profiles (optional)
├── requests/              # Local requests
└── history/               # Local history
```

## Breaking Changes
None - all changes are backward compatible!

## Migration Guide

### From Local to Global
```bash
# Migrate existing config to ~/.restcli/
deno task init --migrate

# Verify migration
ls -la ~/.restcli/

# Delete local files if desired
rm -rf .session.json .profiles.json requests/ history/
```

### Configuration Priority
1. Local `requests/` folder exists → uses current directory
2. `~/.restcli/` exists → uses global config
3. Neither exists → auto-initializes `~/.restcli/`

## Technical Details

### New Files
- `src/config.ts` - ConfigManager for global config
- `src/init.ts` - Initialization script
- `src/history.ts` - History management
- `src/yaml-parser.ts` - YAML/JSON parser
- `http-request.schema.json` - JSON schema for validation
- `INSTALL.md` - Installation guide

### Modified Files
- `src/tui.ts` - Auto-detect and use global config
- `src/run.ts` - Auto-detect and use global config, add profile override
- `src/session.ts` - Add history flag, workdir support, profile variables
- `src/parser.ts` - Auto-detect YAML format
- `scripts/curl2http.ts` - Add output location flag, fix TypeScript error
- `deno.json` - Add build tasks, init task, YAML dependency
- `README.md` - Updated with new features

### Dependencies
- `@std/yaml` - YAML parsing support

### Build Output
- `restcli` (70MB) - Main TUI
- `restcli-run` (70MB) - CLI runner
- `restcli-curl2http` (69MB) - cURL converter
