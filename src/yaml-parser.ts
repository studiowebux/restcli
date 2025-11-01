import { parse as parseYaml } from "@std/yaml";
import { type HttpRequest, type ParsedHttpFile, type Documentation, type Parameter, type Response } from "./parser.ts";

interface YamlParameter {
  name: string;
  type: string;
  required?: boolean;
  description?: string;
  example?: string;
}

interface YamlResponse {
  code?: string;
  description: string;
}

interface YamlDocumentation {
  description?: string;
  tags?: string[];
  parameters?: YamlParameter[];
  responses?: (YamlResponse | Record<string, string>)[];
}

interface YamlRequest {
  name?: string;
  method: string;
  url: string;
  headers?: Record<string, string>;
  body?: string;
  documentation?: YamlDocumentation;
}

interface YamlFile {
  requests?: YamlRequest[];
  // Single request format (no requests array)
  name?: string;
  method?: string;
  url?: string;
  headers?: Record<string, string>;
  body?: string;
  documentation?: YamlDocumentation;
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
        documentation: convertYamlDocumentation(data.documentation),
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
          documentation: convertYamlDocumentation(req.documentation),
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
 * Convert YAML documentation format to internal Documentation format
 */
function convertYamlDocumentation(yamlDoc?: YamlDocumentation): Documentation | undefined {
  if (!yamlDoc) return undefined;

  const documentation: Documentation = {};

  if (yamlDoc.description) {
    documentation.description = yamlDoc.description;
  }

  if (yamlDoc.tags && yamlDoc.tags.length > 0) {
    documentation.tags = yamlDoc.tags;
  }

  if (yamlDoc.parameters && yamlDoc.parameters.length > 0) {
    documentation.parameters = yamlDoc.parameters.map(p => ({
      name: p.name,
      type: p.type,
      required: p.required ?? false,
      description: p.description,
      example: p.example,
    }));
  }

  if (yamlDoc.responses && yamlDoc.responses.length > 0) {
    documentation.responses = yamlDoc.responses.map(r => {
      // Handle both formats: {code: "201", description: "..."} and {"201": "..."}
      if (typeof r === 'object' && 'code' in r && 'description' in r) {
        return { code: r.code!, description: r.description };
      } else if (typeof r === 'object') {
        // Handle {"201": "Created"} format
        const [code, description] = Object.entries(r)[0];
        return { code, description };
      }
      return { code: '200', description: String(r) };
    });
  }

  // Only return documentation if it has content
  const hasContent = documentation.description ||
                     documentation.tags?.length ||
                     documentation.parameters?.length ||
                     documentation.responses?.length;

  return hasContent ? documentation : undefined;
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
