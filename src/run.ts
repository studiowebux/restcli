import { parseHttpFile, applyVariables } from "./parser.ts";
import { RequestExecutor } from "./executor.ts";
import { SessionManager } from "./session.ts";
import { HistoryManager } from "./history.ts";
import { ConfigManager } from "./config.ts";

/**
 * CLI runner for executing HTTP requests without TUI
 * Usage: deno task run <path-to-http-file> [--profile <profile-name>]
 */
async function main() {
  const args = Deno.args;

  if (args.length === 0) {
    console.error("Usage: deno task run <path-to-http-file> [--profile <profile-name>]");
    Deno.exit(1);
  }

  // Parse command line arguments
  let filePath = "";
  let profileOverride: string | null = null;

  for (let i = 0; i < args.length; i++) {
    const arg = args[i];
    if (arg === "--profile" || arg === "-p") {
      if (i + 1 < args.length) {
        profileOverride = args[i + 1];
        i++; // Skip next arg
      } else {
        console.error("Error: --profile requires a profile name");
        Deno.exit(1);
      }
    } else if (!filePath) {
      filePath = arg;
    }
  }

  if (!filePath) {
    console.error("Error: No file path specified");
    Deno.exit(1);
  }

  try {
    // Check if config directory exists, use it if available
    const configManager = new ConfigManager();
    const isInitialized = await configManager.isInitialized();

    let baseDir = ".";
    if (isInitialized) {
      baseDir = configManager.getConfigDir();
    }

    // Load session
    const sessionManager = new SessionManager(baseDir);
    await sessionManager.load();

    // Override profile if specified
    if (profileOverride) {
      const profiles = sessionManager.getProfiles();
      const profile = profiles.find(p => p.name === profileOverride);
      if (!profile) {
        console.error(`Error: Profile "${profileOverride}" not found`);
        console.error("Available profiles:");
        profiles.forEach(p => console.error(`  - ${p.name}`));
        Deno.exit(1);
      }
      sessionManager.setActiveProfile(profileOverride);
      console.log(`Using profile: ${profileOverride}\n`);
    }

    // Read and parse file
    const content = await Deno.readTextFile(filePath);
    const parsed = parseHttpFile(content);

    if (parsed.requests.length === 0) {
      console.error("No requests found in file");
      Deno.exit(1);
    }

    console.log(`Found ${parsed.requests.length} request(s) in ${filePath}\n`);

    // Execute first request
    const request = parsed.requests[0];
    console.log(`Executing: ${request.name || "Unnamed Request"}`);
    console.log(`${request.method} ${request.url}\n`);

    const executor = new RequestExecutor();
    const variables = sessionManager.getVariables();
    const profileHeaders = sessionManager.getActiveHeaders();

    const result = await executor.execute(request, variables, profileHeaders);

    // Save to history if enabled
    if (sessionManager.isHistoryEnabled()) {
      const historyManager = new HistoryManager(baseDir);
      const substituted = applyVariables(request, variables);
      const mergedHeaders = { ...profileHeaders, ...substituted.headers };

      const historyPath = await historyManager.save({
        timestamp: "", // Will be set by save method
        requestFile: filePath,
        requestName: request.name,
        method: substituted.method,
        url: substituted.url,
        headers: mergedHeaders,
        body: substituted.body,
        responseStatus: result.status,
        responseStatusText: result.statusText,
        responseHeaders: result.headers,
        responseBody: result.body,
        duration: result.duration,
        error: result.error,
      });
      console.log(`📝 History saved to: ${historyPath}\n`);
    }

    // Display result
    if (result.error) {
      console.error(`❌ Error: ${result.error}`);
      Deno.exit(1);
    }

    const statusColor = result.status >= 200 && result.status < 300
      ? "\x1b[32m"
      : result.status >= 400
      ? "\x1b[31m"
      : "\x1b[33m";

    console.log(
      `${statusColor}${result.status} ${result.statusText}\x1b[0m | ${Math.round(result.duration)}ms\n`
    );

    console.log("Headers:");
    for (const [key, value] of Object.entries(result.headers)) {
      console.log(`  ${key}: ${value}`);
    }

    console.log("\nBody:");
    console.log(result.body);

    // Try to extract token
    try {
      const json = JSON.parse(result.body);
      if (json.token) {
        sessionManager.setVariable("token", json.token);
        await sessionManager.save();
        console.log("\n✓ Saved token to session");
      }
      if (json.accessToken) {
        sessionManager.setVariable("token", json.accessToken);
        await sessionManager.save();
        console.log("\n✓ Saved accessToken to session");
      }
    } catch {
      // Not JSON or no token
    }
  } catch (error) {
    console.error(`Error: ${error instanceof Error ? error.message : String(error)}`);
    Deno.exit(1);
  }
}

if (import.meta.main) {
  await main();
}
