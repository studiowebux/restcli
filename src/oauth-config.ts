/**
 * OAuth 2.0 Configuration
 * Supports manual endpoint override or auto-build from components
 */

export interface OAuthConfig {
  enabled: boolean;

  // Manual endpoint override (takes precedence over auto-build)
  authEndpoint?: string;

  // Auto-build components (used if authEndpoint not provided)
  authUrl?: string;
  tokenUrl?: string;
  clientId?: string;
  clientSecret?: string;
  redirectUri?: string;
  scope?: string;
  responseType?: "code" | "token";

  // Webhook server configuration
  webhookPort?: number;

  // Token storage
  tokenStorageKey?: string; // Variable name to store token in (default: "token")
}

export interface OAuthValidationError {
  field: string;
  message: string;
}

/**
 * Validate OAuth configuration
 */
export function validateOAuthConfig(config: OAuthConfig): OAuthValidationError[] {
  const errors: OAuthValidationError[] = [];

  if (!config.enabled) {
    return errors; // Not enabled, skip validation
  }

  // If manual endpoint provided, still need tokenUrl for code exchange
  if (config.authEndpoint) {
    try {
      new URL(config.authEndpoint);
    } catch {
      errors.push({
        field: "authEndpoint",
        message: "Invalid URL format"
      });
    }

    // Authorization code flow ALWAYS requires tokenUrl for code exchange
    // Implicit flow (response_type=token) returns token directly in callback
    const responseType = config.responseType || "code";
    if (responseType === "code" && !config.tokenUrl) {
      errors.push({
        field: "tokenUrl",
        message: "Required for authorization code flow (OAuth 2.0 spec)"
      });
    }

    return errors;
  }

  // Auto-build mode: validate all required fields
  if (!config.authUrl) {
    errors.push({
      field: "authUrl",
      message: "authUrl is required (or provide authEndpoint)"
    });
  }

  if (!config.tokenUrl) {
    errors.push({
      field: "tokenUrl",
      message: "tokenUrl is required"
    });
  }

  if (!config.clientId) {
    errors.push({
      field: "clientId",
      message: "clientId is required"
    });
  }

  if (!config.redirectUri) {
    errors.push({
      field: "redirectUri",
      message: "redirectUri is required"
    });
  }

  // Validate URLs
  const urlFields: Array<{ field: keyof OAuthConfig; value?: string }> = [
    { field: "authUrl", value: config.authUrl },
    { field: "tokenUrl", value: config.tokenUrl },
    { field: "redirectUri", value: config.redirectUri },
  ];

  for (const { field, value } of urlFields) {
    if (value) {
      try {
        new URL(value);
      } catch {
        errors.push({
          field,
          message: "Invalid URL format"
        });
      }
    }
  }

  return errors;
}

/**
 * Get default values for optional OAuth config fields
 */
export function getOAuthDefaults(config: OAuthConfig): Required<OAuthConfig> {
  return {
    enabled: config.enabled,
    authEndpoint: config.authEndpoint || "",
    authUrl: config.authUrl || "",
    tokenUrl: config.tokenUrl || "",
    clientId: config.clientId || "",
    clientSecret: config.clientSecret || "",
    redirectUri: config.redirectUri || "http://localhost:8888/callback",
    scope: config.scope || "openid",
    responseType: config.responseType || "code",
    webhookPort: config.webhookPort || 8888,
    tokenStorageKey: config.tokenStorageKey || "token",
  };
}
