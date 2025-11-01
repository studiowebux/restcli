# OAuth Flow Integration - Implementation Plan

## Goal
Automate OAuth2/Cognito authentication flow within the TUI, automatically capturing and storing tokens without manual copy-paste.

## User Flow

1. User triggers OAuth flow (new keyboard shortcut or command)
2. TUI starts local webhook server on configurable port (default: 8888)
3. TUI opens system browser to OAuth provider URL
4. User completes authentication in browser (Cognito login)
5. OAuth provider redirects to `http://localhost:8888/callback` with auth code/tokens
6. TUI's webhook receives the callback, extracts tokens
7. TUI updates active profile with new authorization header/tokens
8. TUI closes webhook server and shows success message
9. User can now make authenticated requests

## Configuration Structure

Add to `.profiles.json`:
```json
{
  "name": "Production",
  "oauth": {
    "enabled": true,
    "provider": "cognito",  // or "generic"
    "authUrl": "https://your-domain.auth.region.amazoncognito.com/oauth2/authorize",
    "tokenUrl": "https://your-domain.auth.region.amazoncognito.com/oauth2/token",
    "clientId": "your-client-id",
    "clientSecret": "your-client-secret",  // optional
    "redirectUri": "http://localhost:8888/callback",
    "scope": "openid profile email",
    "responseType": "code",  // or "token" for implicit flow
    "webhookPort": 8888,
    "tokenStorageKey": "Authorization",  // Which header/variable to update
    "tokenPrefix": "Bearer"  // Prefix for token (e.g., "Bearer ")
  },
  "headers": {
    "Authorization": "Bearer {{token}}"
  },
  "variables": {
    "baseUrl": "https://api.example.com",
    "token": ""  // Will be auto-filled by OAuth flow
  }
}
```

## Implementation Steps

### 1. OAuth Configuration Parser
**File**: `src/oauth-config.ts`
- Interface for OAuth configuration
- Validation of required fields
- Support for Cognito-specific and generic OAuth2

### 2. Local Webhook Server
**File**: `src/oauth-server.ts`
- Lightweight HTTP server using Deno's native `Deno.serve()`
- Listen on configured port
- Handle `/callback` endpoint
- Parse query parameters or POST body
- Extract authorization code or tokens
- Return success HTML page to browser
- Timeout after 5 minutes
- Cleanup and close server

### 3. OAuth Flow Controller
**File**: `src/oauth-flow.ts`
- Build authorization URL with PKCE (for enhanced security)
- Open system browser using `Deno.Command`:
  - macOS: `open <url>`
  - Linux: `xdg-open <url>`
  - Windows: `start <url>`
- Start webhook server
- Handle authorization code exchange for access token (if using authorization code flow)
- Update profile with received tokens
- Save profile to disk

### 4. TUI Integration
**File**: `src/tui.ts`
- Add new keyboard shortcut: `o` or `a` for "OAuth/Authenticate"
- Show OAuth flow status (waiting for auth, processing, success/error)
- Display remaining timeout
- Allow ESC to cancel flow and close webhook
- Update status bar with auth state

### 5. Token Management
- Store access token in profile variables or headers
- Optionally store refresh token for later use
- Add token expiry tracking (future enhancement)
- Add refresh flow (future enhancement)

## Security Considerations

1. **PKCE (Proof Key for Code Exchange)**:
   - Generate code_verifier (random string)
   - Generate code_challenge (SHA-256 hash of verifier)
   - Send challenge with auth request
   - Send verifier with token exchange

2. **Local Server Security**:
   - Only bind to localhost (127.0.0.1)
   - Validate state parameter to prevent CSRF
   - Timeout webhook server after 5 minutes
   - Close server immediately after successful callback

3. **Token Storage**:
   - Store in `.profiles.json` (user-readable but local)
   - Warn user that tokens are stored in plain text
   - Document rotation/expiry best practices

## Example Usage

### Keyboard Shortcut Flow:
1. User presses `o` (OAuth)
2. TUI shows: "Starting OAuth flow... Opening browser..."
3. Browser opens to Cognito login
4. User logs in
5. TUI shows: "✓ Authentication successful! Token received."
6. Status bar updates: "Authenticated as: user@example.com"

### Profile-Based Flow:
```bash
# Start TUI with profile that has OAuth configured
restcli --profile Production

# Press 'o' to authenticate
# Browser opens automatically
# Complete login
# Return to TUI - now authenticated
```

## Error Handling

1. **Timeout**: "OAuth flow timed out after 5 minutes. Press 'o' to retry."
2. **Canceled**: "OAuth flow canceled by user."
3. **Invalid Configuration**: "OAuth not configured for this profile. See docs/OAUTH.md"
4. **Network Error**: "Failed to exchange authorization code: <error>"
5. **Port In Use**: "Port 8888 already in use. Configure different port in oauth.webhookPort"

## Future Enhancements (Phase 2)

1. **Token Refresh**:
   - Auto-refresh when token expires
   - Refresh before making requests if expiry detected

2. **Multiple Providers**:
   - Cognito
   - Auth0
   - Okta
   - Generic OAuth2/OIDC

3. **Device Flow**:
   - For devices without browser access
   - Show code and polling status

4. **Token Introspection**:
   - Show token details (user, scopes, expiry)
   - Warn when token expires soon

## Files to Create/Modify

### New Files:
- `src/oauth-config.ts` - Configuration types and validation
- `src/oauth-server.ts` - Local webhook HTTP server
- `src/oauth-flow.ts` - OAuth flow orchestration
- `src/oauth-pkce.ts` - PKCE helper functions
- `docs/OAUTH.md` - User documentation

### Modified Files:
- `src/tui.ts` - Add keyboard shortcut and UI integration
- `src/parser.ts` - Add OAuth config to Profile type
- `http-request.schema.json` - Add OAuth schema for autocomplete

## Testing Plan

1. **Manual Testing**:
   - Test with Cognito (your use case)
   - Test with generic OAuth2 provider
   - Test timeout scenarios
   - Test cancellation
   - Test invalid configurations

2. **Integration Testing**:
   - Mock OAuth server for automated tests
   - Test token storage and retrieval
   - Test profile switching with different auth states

## Documentation

Create `docs/OAUTH.md` with:
- Configuration examples for common providers
- Step-by-step setup guide
- Security best practices
- Troubleshooting guide
- FAQ

## Timeline Estimate

- OAuth configuration & validation: 1-2 hours
- Local webhook server: 2-3 hours
- OAuth flow controller: 3-4 hours
- TUI integration: 2-3 hours
- Testing & debugging: 2-3 hours
- Documentation: 1-2 hours

**Total: 11-17 hours**

## Decision Points

1. **Store refresh tokens?**
   - Pros: Auto-refresh capability
   - Cons: Security risk if .profiles.json is compromised
   - **Decision**: Store optionally, warn user

2. **PKCE always or optional?**
   - **Decision**: Always use PKCE for security

3. **Support implicit flow?**
   - **Decision**: Support both code and implicit, prefer code

4. **Token encryption?**
   - **Decision**: Not in v1, document as future enhancement
