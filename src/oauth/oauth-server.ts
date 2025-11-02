/**
 * Local OAuth Callback Webhook Server
 * Listens for OAuth provider redirects and extracts authorization codes/tokens
 */

export interface OAuthCallbackResult {
  code?: string;
  accessToken?: string;
  error?: string;
  errorDescription?: string;
  state?: string;
}

export interface WebhookServerOptions {
  port: number;
  timeout?: number; // milliseconds (default: 5 minutes)
  expectedState?: string; // For CSRF validation
}

/**
 * Start local webhook server and wait for OAuth callback
 */
export async function startOAuthWebhookServer(
  options: WebhookServerOptions,
): Promise<OAuthCallbackResult> {
  const { port, timeout = 5 * 60 * 1000, expectedState } = options;

  return new Promise((resolve, reject) => {
    let abortController: AbortController | null = new AbortController();
    let timeoutId: number | null = null;

    // Success HTML page to show in browser
    const successHtml = `
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Authentication Successful</title>
  <style>
    body {
      font-family: system-ui, -apple-system, sans-serif;
      display: flex;
      align-items: center;
      justify-content: center;
      height: 100vh;
      margin: 0;
      background: #000;
    }
    .container {
      background: #fff;
      padding: 3rem;
      border: 2px solid #000;
      text-align: center;
      max-width: 400px;
    }
    h1 { color: #000; margin-top: 0; }
    p { color: #333; line-height: 1.6; }
    .checkmark {
      font-size: 4rem;
      color: #000;
      margin-bottom: 1rem;
    }
  </style>
</head>
<body>
  <div class="container">
    <div class="checkmark">✓</div>
    <h1>Authentication Successful!</h1>
    <p>You can now close this window and return to the terminal.</p>
  </div>
</body>
</html>`;

    // Error HTML page
    const errorHtml = (error: string) => `
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Authentication Error</title>
  <style>
    body {
      font-family: system-ui, -apple-system, sans-serif;
      display: flex;
      align-items: center;
      justify-content: center;
      height: 100vh;
      margin: 0;
      background: #000;
    }
    .container {
      background: #fff;
      padding: 3rem;
      border: 2px solid #000;
      text-align: center;
      max-width: 400px;
    }
    h1 { color: #000; margin-top: 0; }
    p { color: #333; line-height: 1.6; }
    .error-icon {
      font-size: 4rem;
      color: #000;
      margin-bottom: 1rem;
    }
  </style>
</head>
<body>
  <div class="container">
    <div class="error-icon">✗</div>
    <h1>Authentication Failed</h1>
    <p>${error}</p>
    <p>Please close this window and try again.</p>
  </div>
</body>
</html>`;

    const cleanup = () => {
      if (timeoutId !== null) {
        clearTimeout(timeoutId);
      }
      if (abortController) {
        abortController.abort();
        abortController = null;
      }
    };

    // Start HTTP server
    const handler = (
      req: Request,
      connInfo: Deno.ServeHandlerInfo,
    ): Response => {
      const url = new URL(req.url);

      // Only handle /callback endpoint
      if (url.pathname !== "/callback") {
        return new Response("Not Found", { status: 404 });
      }

      // Extract query parameters
      const code = url.searchParams.get("code");
      const accessToken = url.searchParams.get("access_token");
      const error = url.searchParams.get("error");
      const errorDescription = url.searchParams.get("error_description");
      const state = url.searchParams.get("state");

      // Validate state for CSRF protection
      if (expectedState && state !== expectedState) {
        const result: OAuthCallbackResult = {
          error: "invalid_state",
          errorDescription: "State parameter mismatch (potential CSRF attack)",
        };

        // Wait for connection to close, then cleanup
        connInfo.completed.then(async () => {
          await new Promise((resolve) =>
            setTimeout(() => {
              cleanup();
              return resolve(true);
            }, 1000)
          );
          resolve(result);
        });

        return new Response(errorHtml("Invalid state parameter"), {
          status: 400,
          headers: { "Content-Type": "text/html" },
        });
      }

      // Check for errors
      if (error) {
        const result: OAuthCallbackResult = {
          error,
          errorDescription: errorDescription || undefined,
          state: state || undefined,
        };

        // Wait for connection to close, then cleanup
        connInfo.completed.then(() => {
          cleanup();
          resolve(result);
        });

        return new Response(errorHtml(errorDescription || error), {
          status: 400,
          headers: { "Content-Type": "text/html" },
        });
      }

      // Success - got authorization code or access token
      const result: OAuthCallbackResult = {
        code: code || undefined,
        accessToken: accessToken || undefined,
        state: state || undefined,
      };

      // Wait for connection to close, then cleanup
      connInfo.completed.then(async () => {
        await new Promise((resolve) =>
          setTimeout(() => {
            cleanup();
            return resolve(true);
          }, 1000)
        );

        resolve(result);
      });

      return new Response(successHtml, {
        status: 200,
        headers: { "Content-Type": "text/html" },
      });
    };

    // Start server
    try {
      const server = Deno.serve(
        {
          port,
          hostname: "127.0.0.1", // localhost only for security
          signal: abortController.signal,
          onListen: () => {
            // Server started successfully
          },
        },
        handler,
      );

      // Set timeout
      timeoutId = setTimeout(() => {
        cleanup();
        reject(new Error("OAuth flow timed out after 5 minutes"));
      }, timeout);

      // Wait for server to finish (will be aborted on cleanup)
      server.finished.catch(() => {
        // Ignore abort errors
      });
    } catch (error) {
      cleanup();
      if (
        error instanceof Error &&
        error.message.includes("Address already in use")
      ) {
        reject(
          new Error(
            `Port ${port} is already in use. Try a different port in oauth.webhookPort`,
          ),
        );
      } else {
        reject(error);
      }
    }
  });
}
