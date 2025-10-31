import { parseHttpFile } from "./parser.ts";
import { RequestExecutor } from "./executor.ts";
import { SessionManager } from "./session.ts";

/**
 * CLI runner for executing HTTP requests without TUI
 * Usage: deno task run <path-to-http-file>
 */
async function main() {
  const args = Deno.args;

  if (args.length === 0) {
    console.error("Usage: deno task run <path-to-http-file>");
    Deno.exit(1);
  }

  const filePath = args[0];

  try {
    // Load session
    const sessionManager = new SessionManager();
    await sessionManager.load();

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
