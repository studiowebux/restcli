# Header Profiles Guide

Header profiles let you quickly switch between different user accounts/roles without editing individual request files.

## Quick Start

1. **Copy the example files:**

```bash
cp .profiles.json.example .profiles.json
cp .session.json.example .session.json
```

2. **Customize for your game:**
   Edit `.profiles.json` with your actual headers and `.session.json` with your tokens.

3. **Switch profiles in TUI:**
   Press `p` to cycle through profiles. The active profile is shown in the header.

## How It Works

### Profile Definition (`.profiles.json`)

Each profile has a name, optional variables, and headers that will be automatically added to requests:

```json
{
  "name": "Dev - Player 1",
  "variables": {
    "baseUrl": "http://localhost:3000",
    "token": "dev_player1_token_here"
  },
  "headers": {
    "Authorization": "Bearer {{token}}",
    "X-User-ID": "player-001",
    "X-Environment": "dev"
  }
}
```

### Variables Priority

Variables are resolved in this order (later overrides earlier):

1. **Global variables** (`.session.json`)
2. **Profile variables** (`.profiles.json` for active profile)

**Example:**

`.session.json`:

```json
{
  "variables": {
    "apiVersion": "v1",
    "token": "global_fallback_token"
  }
}
```

`.profiles.json` (active profile):

```json
{
  "name": "Dev",
  "variables": {
    "baseUrl": "http://localhost:3000",
    "token": "dev_specific_token"
  }
}
```

**Result:**

- `{{apiVersion}}` → `"v1"` (from global)
- `{{baseUrl}}` → `"http://localhost:3000"` (from profile)
- `{{token}}` → `"dev_specific_token"` (profile overrides global)

### Execution Flow

When you execute a request:

1. **Profile headers** from the active profile are loaded
2. **Variables** are substituted (e.g., `{{token1}}` → actual token)
3. **Request headers** from your `.http` file are merged (override profile headers if same key)
4. Request is sent with the combined headers

## Example: Environment Switching

### Setup Profiles for Multiple Environments

**`.profiles.json`:**

```json
[
  {
    "name": "Dev - Player",
    "variables": {
      "baseUrl": "http://localhost:3000",
      "token": "dev_player_token"
    },
    "headers": {
      "Authorization": "Bearer {{token}}",
      "X-User-ID": "player-001",
      "X-Environment": "dev"
    }
  },
  {
    "name": "Dev - Admin",
    "variables": {
      "baseUrl": "http://localhost:3000",
      "token": "dev_admin_token"
    },
    "headers": {
      "Authorization": "Bearer {{token}}",
      "X-User-ID": "admin-001",
      "X-Admin": "true",
      "X-Environment": "dev"
    }
  },
  {
    "name": "Staging - Player",
    "variables": {
      "baseUrl": "https://staging.example.com",
      "token": "staging_player_token"
    },
    "headers": {
      "Authorization": "Bearer {{token}}",
      "X-User-ID": "player-001",
      "X-Environment": "staging"
    }
  },
  {
    "name": "Production - Player",
    "variables": {
      "baseUrl": "https://api.example.com",
      "token": "prod_player_token"
    },
    "headers": {
      "Authorization": "Bearer {{token}}",
      "X-User-ID": "player-001",
      "X-Environment": "production"
    }
  }
]
```

### Create Reusable Requests

**`requests/auth/login.http`:**

```http
### Login
POST {{baseUrl}}/auth/login
Content-Type: application/json

{
  "username": "player1",
  "password": "test123"
}
```

**Key benefit:** This SAME file works for all environments!

### Manual Token Setup

Edit `.profiles.json` to add your tokens for each environment:

```json
{
  "name": "Dev - Player",
  "variables": {
    "baseUrl": "http://localhost:3000",
    "token": "eyJhbGc...get_this_from_dev_login"
  },
  "headers": {
    "Authorization": "Bearer {{token}}"
  }
}
```

Or use global variables in `.session.json` for shared values:

```json
{
  "variables": {
    "apiVersion": "v1",
    "timeout": "5000"
  }
}
```

### Test Across Environments

Create a test request:

**`requests/player/inventory.http`:**

```http
### Get Player Inventory
GET {{baseUrl}}/player/inventory
```

Now in the TUI:

1. Navigate to this request (or press `:` + line number)
2. Press `i` to inspect → see it targets `http://localhost:3000` (Dev)
3. Press Enter to execute → get dev inventory
4. Press `p` to switch to "Staging - Player"
5. Press `i` again → now see it targets `https://staging.example.com`
6. Press Enter → get staging inventory
7. Press `p` to switch to "Production - Player"
8. Execute → get production inventory

**The request file never changes!** Just switch profiles with `p` and the URL, token, and headers all update automatically.

## Header Precedence

If both profile and request define the same header, **request wins**:

**Profile:**

```json
{
  "headers": {
    "Authorization": "Bearer {{token}}"
  }
}
```

**Request:**

```http
GET /api/test
Authorization: Bearer custom-override-token
```

**Actual request sent:** `Authorization: Bearer custom-override-token`

This is useful for testing edge cases where you need to override the profile temporarily.

## Tips

- Use descriptive profile names (shown in TUI header bar)
- Keep commonly-used tokens in variables
- Use profiles for different:
  - User roles (player, admin, moderator)
  - Account states (new user, banned user, premium user)
  - Character types (warrior, mage, archer)
  - Testing scenarios (valid auth, expired auth, no auth)
- Variables auto-update when responses contain `token` or `accessToken` fields

## Example Use Cases

### Environment Switching (Most Common)

```json
[
  {
    "name": "Dev",
    "variables": {
      "baseUrl": "http://localhost:3000",
      "token": "dev_token"
    },
    "headers": { "X-Environment": "dev" }
  },
  {
    "name": "Staging",
    "variables": {
      "baseUrl": "https://staging.example.com",
      "token": "staging_token"
    },
    "headers": { "X-Environment": "staging" }
  },
  {
    "name": "Production",
    "variables": {
      "baseUrl": "https://api.example.com",
      "token": "prod_token"
    },
    "headers": { "X-Environment": "production" }
  }
]
```

### Multi-Character Testing

```json
[
  {
    "name": "Warrior",
    "variables": { "characterId": "char-warrior-001" },
    "headers": { "X-Character-ID": "{{characterId}}" }
  },
  {
    "name": "Mage",
    "variables": { "characterId": "char-mage-001" },
    "headers": { "X-Character-ID": "{{characterId}}" }
  }
]
```

### User Role Testing

```json
[
  {
    "name": "Basic User",
    "variables": { "token": "basic_user_token" },
    "headers": {
      "Authorization": "Bearer {{token}}",
      "X-Permissions": "read"
    }
  },
  {
    "name": "Admin",
    "variables": { "token": "admin_token" },
    "headers": {
      "Authorization": "Bearer {{token}}",
      "X-Permissions": "read,write,delete,admin"
    }
  }
]
```

---

**Need help?** Check the main README.md for keyboard shortcuts and general usage.
