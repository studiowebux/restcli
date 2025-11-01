#!/usr/bin/env -S deno run --allow-read --allow-write --allow-net

import { OpenAPIParser } from "../src/openapi/spec-parser.ts";
import { OpenAPIConverter } from "../src/openapi/converter.ts";

async function main() {
  const args = Deno.args;

  // Parse command line arguments
  let inputPath = "";
  let outputDir = "requests";
  let organizeBy: "tags" | "paths" | "flat" = "tags";
  let showHelp = false;

  for (let i = 0; i < args.length; i++) {
    const arg = args[i];

    if (arg === "--help" || arg === "-h") {
      showHelp = true;
    } else if (arg === "--input" || arg === "-i") {
      if (i + 1 < args.length) {
        inputPath = args[i + 1];
        i++;
      }
    } else if (arg === "--output" || arg === "-o") {
      if (i + 1 < args.length) {
        outputDir = args[i + 1];
        i++;
      }
    } else if (arg === "--organize-by") {
      if (i + 1 < args.length) {
        const value = args[i + 1];
        if (value === "tags" || value === "paths" || value === "flat") {
          organizeBy = value;
        } else {
          console.error(`Invalid --organize-by value: ${value}. Must be: tags, paths, or flat`);
          Deno.exit(1);
        }
        i++;
      }
    } else if (!inputPath) {
      inputPath = arg;
    }
  }

  if (showHelp || !inputPath) {
    console.log(`
restcli-openapi2http - Convert OpenAPI/Swagger specs to .http files

USAGE:
  restcli-openapi2http <input> [OPTIONS]

OPTIONS:
  -i, --input <path|url>       Path to OpenAPI spec file or URL (required)
  -o, --output <dir>           Output directory (default: requests)
  --organize-by <strategy>     Organization strategy: tags, paths, or flat (default: tags)
  -h, --help                   Show this help message

ORGANIZATION STRATEGIES:
  tags    - Group files by OpenAPI tags (e.g., requests/users/, requests/auth/)
  paths   - Group files by URL path segments (e.g., requests/api/, requests/v1/)
  flat    - All files in one directory (e.g., requests/)

EXAMPLES:
  # From local file
  restcli-openapi2http ./swagger.json --output requests/

  # From URL
  restcli-openapi2http https://petstore3.swagger.io/api/v3/openapi.json

  # Organize by paths instead of tags
  restcli-openapi2http ./api.yaml --organize-by paths

  # Using deno
  deno run --allow-read --allow-write --allow-net scripts/openapi2http.ts swagger.json

SUPPORTED FORMATS:
  - OpenAPI 3.0.x (JSON)
  - OpenAPI 3.0.x (YAML)
  - OpenAPI 3.1.x (JSON/YAML)

OUTPUT:
  Generated .http files include:
  - Complete documentation (@description, @param, @response)
  - HTTP method and URL with variable placeholders
  - Headers (including authentication)
  - Request body with examples from schema

For more information, see: docs/OPENAPI.md
`);
    Deno.exit(showHelp ? 0 : 1);
  }

  try {
    console.log(`\n🔄 Loading OpenAPI spec from: ${inputPath}\n`);

    // Load spec
    let spec;
    if (inputPath.startsWith("http://") || inputPath.startsWith("https://")) {
      spec = await OpenAPIParser.fetchUrl(inputPath);
    } else {
      spec = await OpenAPIParser.loadFile(inputPath);
    }

    // Show summary
    console.log(OpenAPIParser.getSummary(spec));
    console.log();

    // Convert
    console.log(`\n🔄 Converting to .http files...\n`);
    const converter = new OpenAPIConverter(spec, {
      outputDir,
      organizeBy,
    });

    const { filesCreated, summary } = await converter.convert();

    // Show results
    console.log(summary);
    console.log();

    // Show next steps
    console.log(`\n💡 Next steps:`);
    console.log(`  1. Review generated files in ${outputDir}/`);
    console.log(`  2. Update variable placeholders in .session.json or .profiles.json`);
    console.log(`  3. Run 'restcli' to test your requests`);
    console.log();

  } catch (error) {
    console.error(`\n❌ Error: ${error instanceof Error ? error.message : String(error)}\n`);
    Deno.exit(1);
  }
}

if (import.meta.main) {
  await main();
}
