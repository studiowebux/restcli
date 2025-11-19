# Profiles Guide

Header profiles let you quickly switch between different user accounts/roles without editing individual request files.

## How It Works

### Profile Definition (`.profiles.json`)

Each profile has a name, optional variables, and headers that will be automatically added to requests:

```json
{
  "name": "Dev",
  "workdir": "/Users/.../requests",
  "oauth": {},
  "editor": "",
  "variables": {
    "baseUrl": "http://localhost:3000",
    "token": "dev_specific_token"
  },
  "headers": {}
}
```
