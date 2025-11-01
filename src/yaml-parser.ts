import { parse as parseYaml } from "@std/yaml";
import { type HttpRequest, type ParsedHttpFile } from "./parser.ts";

interface YamlRequest {
  name?: string;
  method: string;
  url: string;
  headers?: Record<string, string>;
  body?: string;
}

interface YamlFile {
  requests?: YamlRequest[];
  // Single request format (no requests array)
  name?: string;
  method?: string;
  url?: string;
  headers?: Record<string, string>;
  body?: string;
}

/**
 * Parse YAML/JSON format HTTP request file
 */
export function parseYamlHttpFile(content: string): ParsedHttpFile {
  try {
    const data = parseYaml(content) as YamlFile;

    const requests: HttpRequest[] = [];

    // Check if it's a single request (no requests array)
    if (data.method && data.url) {
      requests.push({
        name: data.name,
        method: data.method.toUpperCase(),
        url: data.url,
        headers: data.headers || {},
        body: data.body,
      });
    }
    // Multiple requests format
    else if (data.requests && Array.isArray(data.requests)) {
      for (const req of data.requests) {
        requests.push({
          name: req.name,
          method: req.method.toUpperCase(),
          url: req.url,
          headers: req.headers || {},
          body: req.body,
        });
      }
    } else {
      throw new Error("Invalid YAML format: must have either method+url or requests array");
    }

    return { requests };
  } catch (error) {
    throw new Error(`Failed to parse YAML: ${error instanceof Error ? error.message : String(error)}`);
  }
}

/**
 * Detect if content is YAML/JSON format or traditional HTTP format
 * YAML files typically start with --- or have key: value syntax
 * JSON files start with { or [
 */
export function isYamlFormat(content: string): boolean {
  const trimmed = content.trim();

  // Check for YAML document start marker
  if (trimmed.startsWith("---")) {
    return true;
  }

  // Check for JSON
  if (trimmed.startsWith("{") || trimmed.startsWith("[")) {
    return true;
  }

  // Check for YAML key-value patterns (method:, url:, requests:, etc.)
  const yamlPatterns = [
    /^method\s*:/m,
    /^url\s*:/m,
    /^requests\s*:/m,
    /^name\s*:/m,
    /^headers\s*:/m,
  ];

  return yamlPatterns.some(pattern => pattern.test(trimmed));
}

/**
 * Convert HttpRequest to YAML string
 */
export function requestToYaml(request: HttpRequest, includeRequestsWrapper: boolean = false): string {
  const req: YamlRequest = {
    method: request.method,
    url: request.url,
  };

  if (request.name) {
    req.name = request.name;
  }

  if (Object.keys(request.headers).length > 0) {
    req.headers = request.headers;
  }

  if (request.body) {
    req.body = request.body;
  }

  if (includeRequestsWrapper) {
    return `---
requests:
  - ${yamlStringify(req).split('\n').join('\n    ')}`;
  }

  return `---\n${yamlStringify(req)}`;
}

// Simple YAML stringifier for our use case
function yamlStringify(obj: unknown, indent: number = 0): string {
  const spaces = " ".repeat(indent);

  if (typeof obj === "string") {
    // Multi-line strings
    if (obj.includes("\n")) {
      return `|\n${spaces}  ${obj.split("\n").join(`\n${spaces}  `)}`;
    }
    // Check if string needs quoting
    if (obj.includes(":") || obj.includes("#") || obj.includes("{{")) {
      return `"${obj.replace(/"/g, '\\"')}"`;
    }
    return obj;
  }

  if (typeof obj === "number" || typeof obj === "boolean") {
    return String(obj);
  }

  if (Array.isArray(obj)) {
    return obj.map(item => `\n${spaces}- ${yamlStringify(item, indent + 2)}`).join("");
  }

  if (typeof obj === "object" && obj !== null) {
    const entries = Object.entries(obj);
    return entries.map(([key, value]) => {
      const valueStr = yamlStringify(value, indent + 2);
      if (typeof value === "object" && value !== null && !Array.isArray(value)) {
        return `\n${spaces}${key}:${valueStr}`;
      }
      if (typeof value === "string" && value.includes("\n")) {
        return `\n${spaces}${key}: ${valueStr}`;
      }
      return `\n${spaces}${key}: ${valueStr}`;
    }).join("").substring(1); // Remove leading newline
  }

  return String(obj);
}
