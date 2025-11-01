#!/usr/bin/env -S deno run --allow-read

import { parseHttpFile } from "./src/parser.ts";

console.log("Testing documentation parser...\n");

// Test .http file with annotations
console.log("=== Testing .http file ===");
const httpContent = await Deno.readTextFile("./examples/documented-request.http");
const httpParsed = parseHttpFile(httpContent);

console.log("Request name:", httpParsed.requests[0].name);
console.log("Method:", httpParsed.requests[0].method);
console.log("URL:", httpParsed.requests[0].url);
console.log("\nDocumentation:");
console.log(JSON.stringify(httpParsed.requests[0].documentation, null, 2));

// Test YAML file with documentation section
console.log("\n\n=== Testing YAML file ===");
const yamlContent = await Deno.readTextFile("./examples/documented-request.yaml");
const yamlParsed = parseHttpFile(yamlContent);

console.log("Request name:", yamlParsed.requests[0].name);
console.log("Method:", yamlParsed.requests[0].method);
console.log("URL:", yamlParsed.requests[0].url);
console.log("\nDocumentation:");
console.log(JSON.stringify(yamlParsed.requests[0].documentation, null, 2));

console.log("\n✅ Parser test complete!");
