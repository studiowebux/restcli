#!/usr/bin/env -S deno run --allow-read --allow-write --allow-env

interface ParsedCurl {
  method: string;
  url: string;
  headers: Record<string, string>;
  body?: string;
}

/**
 * Parse a curl command into structured data
 */
function parseCurl(curlCommand: string): ParsedCurl {
  // Remove newlines and extra spaces
  let normalized = curlCommand
    .replace(/\\\n/g, " ") // Handle multiline with \
    .replace(/\s+/g, " ")
    .trim();

  // Remove 'curl' prefix if present
  if (normalized.startsWith("curl ")) {
    normalized = normalized.substring(5);
  }

  const result: ParsedCurl = {
    method: "GET",
    url: "",
    headers: {},
  };

  // Extract URL - try multiple patterns
  let urlFound = false;

  // Pattern 1: Explicit --url flag
  const urlFlagMatch = normalized.match(/--url\s+['"]?([^\s'"]+)['"]?/);
  if (urlFlagMatch) {
    result.url = urlFlagMatch[1];
    urlFound = true;
  }

  // Pattern 2: URL with protocol (http:// or https://)
  if (!urlFound) {
    const httpMatch = normalized.match(/['"]?(https?:\/\/[^\s'"]+)['"]?/);
    if (httpMatch) {
      result.url = httpMatch[1];
      urlFound = true;
    }
  }

  // Pattern 3: localhost without protocol
  if (!urlFound) {
    const localhostMatch = normalized.match(/['"]?(localhost[^\s'"]*)/);
    if (localhostMatch) {
      result.url = "http://" + localhostMatch[1];
      urlFound = true;
    }
  }

  // Pattern 4: First argument that looks like a URL (after curl)
  if (!urlFound) {
    const firstArgMatch = normalized.match(/^['"]?([^\s'"]+)/);
    if (firstArgMatch && !firstArgMatch[1].startsWith("-")) {
      result.url = firstArgMatch[1];
      urlFound = true;
    }
  }

  // Extract method
  const methodMatch = normalized.match(/(?:-X|--request)\s+['"]?([A-Z]+)['"]?/);
  if (methodMatch) {
    result.method = methodMatch[1];
  } else if (normalized.includes("-d ") || normalized.includes("--data")) {
    result.method = "POST"; // curl defaults to POST when -d is used
  }

  // Extract headers
  const headerRegex = /(?:-H|--header)\s+['"]([^'"]+)['"]/g;
  let headerMatch;
  while ((headerMatch = headerRegex.exec(normalized)) !== null) {
    const headerLine = headerMatch[1];
    const separatorIndex = headerLine.indexOf(":");
    if (separatorIndex > 0) {
      const key = headerLine.substring(0, separatorIndex).trim();
      const value = headerLine.substring(separatorIndex + 1).trim();
      result.headers[key] = value;
    }
  }

  // Extract body (data) - handle nested quotes
  const dataRegexes = [
    /(?:-d|--data|--data-raw|--data-binary)\s+'([^']+)'/,  // Single quotes
    /(?:-d|--data|--data-raw|--data-binary)\s+"([^"]+)"/,  // Double quotes
    /(?:-d|--data|--data-raw|--data-binary)\s+(.+?)(?:\s+-|$)/,  // No quotes
  ];

  for (const regex of dataRegexes) {
    const dataMatch = normalized.match(regex);
    if (dataMatch) {
      result.body = dataMatch[1].trim();
      break;
    }
  }

  return result;
}

/**
 * Detect and suggest variable replacements
 */
function detectVariables(url: string, body?: string): {
  url: string;
  body?: string;
  suggestions: Record<string, string>;
} {
  const suggestions: Record<string, string> = {};
  let modifiedUrl = url;
  let modifiedBody = body;

  // Detect base URL
  const baseUrlMatch = url.match(/(https?:\/\/[^/]+)/);
  if (baseUrlMatch) {
    const baseUrl = baseUrlMatch[1];
    suggestions.baseUrl = baseUrl;
    modifiedUrl = url.replace(baseUrl, "{{baseUrl}}");
  }

  // Detect common patterns in body (like IDs)
  if (modifiedBody) {
    try {
      const bodyObj = JSON.parse(modifiedBody);

      // Look for token-like values
      if (bodyObj.token || bodyObj.accessToken) {
        const tokenField = bodyObj.token ? "token" : "accessToken";
        suggestions.token = bodyObj[tokenField];
        bodyObj[tokenField] = "{{token}}";
        modifiedBody = JSON.stringify(bodyObj, null, 2);
      }

      // Look for ID patterns
      if (bodyObj.userId || bodyObj.user_id) {
        const userIdField = bodyObj.userId ? "userId" : "user_id";
        suggestions.userId = bodyObj[userIdField];
        bodyObj[userIdField] = "{{userId}}";
        modifiedBody = JSON.stringify(bodyObj, null, 2);
      }
    } catch {
      // Not JSON, skip
    }
  }

  return { url: modifiedUrl, body: modifiedBody, suggestions };
}

/**
 * Generate .http file content
 */
function generateHttpFile(parsed: ParsedCurl, detectVars: boolean = true): {
  content: string;
  variables: Record<string, string>;
} {
  let { url, body, headers } = parsed;
  let variables: Record<string, string> = {};

  if (detectVars) {
    const detected = detectVariables(url, body);
    url = detected.url;
    body = detected.body;
    variables = detected.suggestions;
  }

  let content = "### ";

  // Generate name from URL path
  const urlObj = new URL(url.replace("{{baseUrl}}", "http://example.com"));
  const pathParts = urlObj.pathname.split("/").filter(p => p);
  if (pathParts.length > 0) {
    content += `${parsed.method} ${pathParts.join("/")}`;
  } else {
    content += `${parsed.method} Request`;
  }

  content += "\n";
  content += `${parsed.method} ${url}\n`;

  // Add headers
  for (const [key, value] of Object.entries(headers)) {
    content += `${key}: ${value}\n`;
  }

  // Add body if present
  if (body) {
    content += "\n";
    // Try to pretty-print JSON
    try {
      const jsonObj = JSON.parse(body);
      content += JSON.stringify(jsonObj, null, 2);
    } catch {
      content += body;
    }
    content += "\n";
  }

  return { content, variables };
}

/**
 * Suggest a filename based on the request
 */
function suggestFilename(parsed: ParsedCurl): string {
  try {
    const urlObj = new URL(parsed.url.replace("{{baseUrl}}", "http://example.com"));
    const pathParts = urlObj.pathname.split("/").filter(p => p);

    if (pathParts.length === 0) {
      return "request.http";
    }

    // Use last path segment as filename
    const lastPart = pathParts[pathParts.length - 1];
    const sanitized = lastPart.replace(/[^a-zA-Z0-9-]/g, "-");

    // Add method prefix if it makes sense
    if (parsed.method !== "GET") {
      return `${parsed.method.toLowerCase()}-${sanitized}.http`;
    }

    return `${sanitized}.http`;
  } catch {
    return "request.http";
  }
}

/**
 * Interactive mode - prompt user for details
 */
async function interactiveMode(parsed: ParsedCurl, result: { content: string; variables: Record<string, string> }): Promise<void> {
  console.log("\n📝 Converted curl to .http format:\n");
  console.log("─".repeat(60));
  console.log(result.content);
  console.log("─".repeat(60));

  if (Object.keys(result.variables).length > 0) {
    console.log("\n💡 Detected variables:");
    for (const [key, value] of Object.entries(result.variables)) {
      console.log(`  ${key}: ${value}`);
    }
  }

  const suggestedName = suggestFilename(parsed);
  console.log(`\n📁 Suggested filename: requests/${suggestedName}`);

  console.log("\nOptions:");
  console.log("  1. Save to suggested location");
  console.log("  2. Enter custom filename");
  console.log("  3. Print to stdout only");
  console.log("  4. Cancel");

  // Read user input
  const buf = new Uint8Array(1024);
  const n = await Deno.stdin.read(buf);
  if (!n) {
    console.log("No input received, printing to stdout only");
    return;
  }

  const choice = new TextDecoder().decode(buf.subarray(0, n)).trim();

  if (choice === "1") {
    const fullPath = `requests/${suggestedName}`;
    await Deno.writeTextFile(fullPath, result.content);
    console.log(`\n✅ Saved to: ${fullPath}`);

    if (Object.keys(result.variables).length > 0) {
      console.log("\n💡 Don't forget to add these to .session.json or .profiles.json:");
      console.log(JSON.stringify(result.variables, null, 2));
    }
  } else if (choice === "2") {
    console.log("\nEnter filename (relative to requests/): ");
    const filenameBuf = new Uint8Array(1024);
    const filenameN = await Deno.stdin.read(filenameBuf);
    if (filenameN) {
      const customName = new TextDecoder().decode(filenameBuf.subarray(0, filenameN)).trim();
      const fullPath = `requests/${customName}`;
      await Deno.writeTextFile(fullPath, result.content);
      console.log(`\n✅ Saved to: ${fullPath}`);
    }
  } else if (choice === "3" || choice === "4") {
    console.log("\n👋 Output printed above");
  }
}

/**
 * Main function
 */
async function main() {
  const args = Deno.args;

  let curlCommand = "";
  let stdoutOnly = false;

  // Check for --stdout flag
  const stdoutFlagIndex = args.indexOf("--stdout");
  if (stdoutFlagIndex !== -1) {
    stdoutOnly = true;
    args.splice(stdoutFlagIndex, 1); // Remove the flag
  }

  if (args.length === 0) {
    // Read from stdin
    // Auto-detect if stdin is piped (non-interactive)
    const isStdinPiped = !Deno.stdin.isTerminal?.() ?? false;

    if (isStdinPiped) {
      stdoutOnly = true; // Auto-enable stdout mode when piped
    } else {
      console.log("📋 Reading from stdin... (paste your curl command and press Ctrl+D)");
    }

    const decoder = new TextDecoder();
    const buf = new Uint8Array(1024 * 10); // 10KB buffer
    const n = await Deno.stdin.read(buf);
    if (n) {
      curlCommand = decoder.decode(buf.subarray(0, n));
    } else {
      console.error("❌ No input provided");
      Deno.exit(1);
    }
  } else {
    // Use command line argument
    curlCommand = args.join(" ");
    // If command line args provided, default to stdout mode (cleaner output)
    if (!stdoutOnly) {
      stdoutOnly = true;
    }
  }

  try {
    const parsed = parseCurl(curlCommand);

    if (!parsed.url) {
      console.error("❌ Could not extract URL from curl command");
      console.error("\n🔍 Debug info:");
      console.error(`  Input length: ${curlCommand.length} characters`);
      console.error(`  First 100 chars: ${curlCommand.substring(0, 100)}`);
      console.error("\n💡 Make sure your curl command includes a URL like:");
      console.error("  curl http://localhost:3000/api/endpoint");
      console.error("  curl https://example.com/api/endpoint");
      Deno.exit(1);
    }

    const result = generateHttpFile(parsed, true);

    if (stdoutOnly) {
      // Auto-save to suggested location
      const suggestedName = suggestFilename(parsed);
      const fullPath = `requests/${suggestedName}`;

      // Create directory if it doesn't exist
      try {
        await Deno.mkdir("requests", { recursive: true });
      } catch {
        // Directory already exists
      }

      await Deno.writeTextFile(fullPath, result.content);

      console.log(`✅ Saved to: ${fullPath}`);

      if (Object.keys(result.variables).length > 0) {
        console.log("\n💡 Detected variables:");
        for (const [key, value] of Object.entries(result.variables)) {
          console.log(`  ${key}: ${value}`);
        }
        console.log("\n📝 Add these to .session.json or .profiles.json");
      }
    } else {
      // Interactive mode
      await interactiveMode(parsed, result);
    }
  } catch (error) {
    console.error("❌ Error:", error instanceof Error ? error.message : String(error));
    Deno.exit(1);
  }
}

if (import.meta.main) {
  await main();
}
