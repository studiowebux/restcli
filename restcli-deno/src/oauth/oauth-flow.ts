/**
 * OAuth 2.0 Flow Controller
 * Orchestrates the complete OAuth authentication flow
 */

import type { OAuthConfig } from "./oauth-config.ts";
import { getOAuthDefaults } from "./oauth-config.ts";
import { generatePKCEParams, type PKCEParams } from "./oauth-pkce.ts";
import { startOAuthWebhookServer, type OAuthCallbackResult } from "./oauth-server.ts";

export interface OAuthFlowResult {
  success: boolean;
  accessToken?: string;
  refreshToken?: string;
  idToken?: string;
  expiresIn?: number;
  error?: string;
}

/**
 * Build authorization URL from config and PKCE params
 */
function buildAuthorizationUrl(config: OAuthConfig, pkce: PKCEParams): string {
  const defaults = getOAuthDefaults(config);

  // If manual endpoint provided, use it and append PKCE params
  if (defaults.authEndpoint) {
    const url = new URL(defaults.authEndpoint);
    url.searchParams.set("code_challenge", pkce.codeChallenge);
    url.searchParams.set("code_challenge_method", "S256");
    url.searchParams.set("state", pkce.state);
    return url.toString();
  }

  // Auto-build from components
  const url = new URL(defaults.authUrl);
  url.searchParams.set("client_id", defaults.clientId);
  url.searchParams.set("redirect_uri", defaults.redirectUri);
  url.searchParams.set("response_type", defaults.responseType);
  url.searchParams.set("scope", defaults.scope);
  url.searchParams.set("code_challenge", pkce.codeChallenge);
  url.searchParams.set("code_challenge_method", "S256");
  url.searchParams.set("state", pkce.state);

  return url.toString();
}

/**
 * Exchange authorization code for access token
 */
async function exchangeCodeForToken(
  config: OAuthConfig,
  code: string,
  pkce: PKCEParams
): Promise<OAuthFlowResult> {
  const defaults = getOAuthDefaults(config);

  // Validate tokenUrl is provided
  if (!defaults.tokenUrl) {
    return {
      success: false,
      error: "Missing tokenUrl: Authorization code flow requires a token endpoint (OAuth 2.0 spec). Configure tokenUrl in your profile's OAuth settings.",
    };
  }

  const body = new URLSearchParams({
    grant_type: "authorization_code",
    code,
    redirect_uri: defaults.redirectUri,
    client_id: defaults.clientId,
    code_verifier: pkce.codeVerifier,
  });

  // Add client secret if provided
  if (defaults.clientSecret) {
    body.set("client_secret", defaults.clientSecret);
  }

  try {
    const response = await fetch(defaults.tokenUrl, {
      method: "POST",
      headers: {
        "Content-Type": "application/x-www-form-urlencoded",
      },
      body: body.toString(),
    });

    if (!response.ok) {
      const errorText = await response.text();
      return {
        success: false,
        error: `Token exchange failed: ${response.status} ${errorText}`,
      };
    }

    const data = await response.json();

    return {
      success: true,
      accessToken: data.access_token,
      refreshToken: data.refresh_token,
      idToken: data.id_token,
      expiresIn: data.expires_in,
    };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : String(error),
    };
  }
}

/**
 * Open URL in macOS default browser
 */
async function openBrowser(url: string): Promise<void> {
  const command = new Deno.Command("open", {
    args: [url],
    stdout: "null",
    stderr: "null",
  });

  const process = command.spawn();
  await process.status;
}

/**
 * Execute complete OAuth flow
 */
export async function executeOAuthFlow(
  config: OAuthConfig,
  onStatusChange?: (status: string) => void
): Promise<OAuthFlowResult> {
  const defaults = getOAuthDefaults(config);

  try {
    // Step 1: Generate PKCE parameters
    onStatusChange?.("Generating security parameters...");
    const pkce = await generatePKCEParams();

    // Step 2: Build authorization URL
    onStatusChange?.("Building authorization URL...");
    const authUrl = buildAuthorizationUrl(config, pkce);

    // Step 3: Start local webhook server
    onStatusChange?.("Starting local server...");
    const webhookPromise = startOAuthWebhookServer({
      port: defaults.webhookPort,
      expectedState: pkce.state,
    });

    // Step 4: Open browser
    onStatusChange?.("Opening browser...");
    await openBrowser(authUrl);

    // Step 5: Wait for callback
    onStatusChange?.("Waiting for authentication...");
    const callbackResult: OAuthCallbackResult = await webhookPromise;

    // Step 6: Handle callback result
    if (callbackResult.error) {
      return {
        success: false,
        error: callbackResult.errorDescription || callbackResult.error,
      };
    }

    // If using implicit flow (response_type=token), we already have the token
    if (callbackResult.accessToken) {
      return {
        success: true,
        accessToken: callbackResult.accessToken,
      };
    }

    // If using authorization code flow, exchange code for token
    if (callbackResult.code) {
      onStatusChange?.("Exchanging code for token...");
      return await exchangeCodeForToken(config, callbackResult.code, pkce);
    }

    return {
      success: false,
      error: "No authorization code or access token received",
    };

  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : String(error),
    };
  }
}
