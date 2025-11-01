import { isYamlFormat, parseYamlHttpFile } from "./yaml-parser.ts";

export interface Parameter {
  name: string;
  type: string;
  required: boolean;
  description?: string;
  example?: string;
  deprecated?: boolean;
}

export interface ResponseField {
  name: string;
  type: string;
  required: boolean;
  description?: string;
  example?: any;
  deprecated?: boolean;
}

export interface Response {
  code: string;
  description: string;
  contentType?: string;
  fields?: ResponseField[];
  example?: string;
}

export interface Documentation {
  description?: string;
  tags?: string[];
  parameters?: Parameter[];
  responses?: Response[];
}

export interface HttpRequest {
  name?: string;
  method: string;
  url: string;
  headers: Record<string, string>;
  body?: string;
  documentation?: Documentation;
}

export interface ParsedHttpFile {
  requests: HttpRequest[];
}

/**
 * Parse HTTP file - auto-detects format (traditional HTTP or YAML/JSON)
 */
export function parseHttpFile(content: string): ParsedHttpFile {
  if (isYamlFormat(content)) {
    return parseYamlHttpFile(content);
  }
  return parseTraditionalHttpFile(content);
}

/**
 * Parse traditional HTTP file format
 * Supports:
 * ### Request Name
 * METHOD url
 * Header: value
 *
 * body
 *
 * ###
 */
export function parseTraditionalHttpFile(content: string): ParsedHttpFile {
  const requests: HttpRequest[] = [];
  const sections = content.split(/^###\s*/m).filter(s => s.trim());

  for (const section of sections) {
    const lines = section.split('\n');
    let name: string | undefined;
    let method = '';
    let url = '';
    const headers: Record<string, string> = {};
    let bodyStartIndex = -1;
    const documentation: Documentation = {};

    // Find method and URL first
    // Skip the first line if it looks like a title (e.g., "POST /v1/mobile/account/claim")
    // The actual method line should have variables ({{...}}) or full URLs (http://...)
    let methodLineIdx = -1;
    for (let i = 0; i < lines.length; i++) {
      const line = lines[i].trim();
      if (!line) continue;

      // Check if this line starts with a method
      const methodMatch = line.match(/^(GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS)\s+(.+)$/i);
      if (methodMatch) {
        const potentialUrl = methodMatch[2].trim();

        // Only treat it as the actual method line if:
        // 1. It contains variables ({{...}})
        // 2. It starts with http:// or https://
        // 3. It starts with / and we're past the first line (to allow titles on line 0)
        if (
          potentialUrl.includes('{{') ||
          potentialUrl.startsWith('http://') ||
          potentialUrl.startsWith('https://') ||
          (potentialUrl.startsWith('/') && i > 0)
        ) {
          method = methodMatch[1].toUpperCase();
          url = potentialUrl;
          methodLineIdx = i;
          bodyStartIndex = i + 1;
          break;
        }
      }
    }

    // Parse lines before method line for name and documentation annotations
    if (methodLineIdx > 0) {
      for (let i = 0; i < methodLineIdx; i++) {
        const line = lines[i].trim();
        if (!line) continue;

        // Check for documentation annotations (# @...)
        // Note: Use [\w-]+ to support hyphenated annotation types like @response-field
        const annotationMatch = line.match(/^#\s*@([\w-]+)\s+(.+)$/);
        if (annotationMatch) {
          const [, annotationType, annotationValue] = annotationMatch;
          parseAnnotation(annotationType, annotationValue, documentation);
        } else if (line.startsWith('#')) {
          // Skip regular comments
          continue;
        } else if (!name) {
          // First non-comment, non-annotation line is the name
          name = line;
        }
      }
    }

    if (!method || !url) continue;

    // Parse headers until empty line or end
    for (let i = bodyStartIndex; i < lines.length; i++) {
      const line = lines[i].trim();
      if (!line) {
        bodyStartIndex = i + 1;
        break;
      }

      const headerMatch = line.match(/^([^:]+):\s*(.+)$/);
      if (headerMatch) {
        headers[headerMatch[1].trim()] = headerMatch[2].trim();
        bodyStartIndex = i + 1;
      } else {
        // Not a header, must be start of body
        bodyStartIndex = i;
        break;
      }
    }

    // Rest is body
    const bodyLines = lines.slice(bodyStartIndex).filter(l => l.trim() !== '###');
    const body = bodyLines.join('\n').trim() || undefined;

    // Only include documentation if it has content
    const hasDocumentation = documentation.description ||
                             documentation.tags?.length ||
                             documentation.parameters?.length ||
                             documentation.responses?.length;

    requests.push({
      name,
      method,
      url,
      headers,
      body,
      documentation: hasDocumentation ? documentation : undefined
    });
  }

  return { requests };
}

/**
 * Parse a documentation annotation
 * Supports: @description, @tag, @param, @example, @response, @response-field, @response-example
 */
function parseAnnotation(type: string, value: string, documentation: Documentation): void {
  switch (type) {
    case 'description':
      documentation.description = value;
      break;

    case 'tag':
      if (!documentation.tags) documentation.tags = [];
      documentation.tags.push(value);
      break;

    case 'param': {
      // Format: name {type} required|optional - description
      const paramMatch = value.match(/^(\w+)\s+\{(\w+)\}\s+(required|optional)(?:\s+-\s+(.+))?$/);
      if (paramMatch) {
        const [, name, type, requiredStr, description] = paramMatch;
        if (!documentation.parameters) documentation.parameters = [];
        documentation.parameters.push({
          name,
          type,
          required: requiredStr === 'required',
          description,
        });
      }
      break;
    }

    case 'example': {
      // Format: paramName value
      const exampleMatch = value.match(/^(\w+)\s+(.+)$/);
      if (exampleMatch && documentation.parameters) {
        const [, paramName, exampleValue] = exampleMatch;
        const param = documentation.parameters.find(p => p.name === paramName);
        if (param) {
          // Remove quotes if present
          param.example = exampleValue.replace(/^["'](.+)["']$/, '$1');
        }
      }
      break;
    }

    case 'response': {
      // Format: code - description
      const responseMatch = value.match(/^(\d{3})\s+-\s+(.+)$/);
      if (responseMatch) {
        const [, code, description] = responseMatch;
        if (!documentation.responses) documentation.responses = [];
        documentation.responses.push({ code, description, fields: [] });
      }
      break;
    }

    case 'response-field': {
      // Format: fieldName {type} required|optional [deprecated] - description
      const fieldMatch = value.match(/^(\S+)\s+\{([^}]+)\}\s+(required|optional)(?:\s+(deprecated))?(?:\s+-\s+(.+))?$/);
      if (fieldMatch && documentation.responses && documentation.responses.length > 0) {
        const [, name, type, requiredStr, deprecatedStr, description] = fieldMatch;
        const lastResponse = documentation.responses[documentation.responses.length - 1];
        if (!lastResponse.fields) lastResponse.fields = [];
        lastResponse.fields.push({
          name,
          type,
          required: requiredStr === 'required',
          deprecated: deprecatedStr === 'deprecated',
          description,
        });
      }
      break;
    }

    case 'response-example': {
      // Format: fieldName value
      const exampleMatch = value.match(/^(\S+)\s+(.+)$/);
      if (exampleMatch && documentation.responses && documentation.responses.length > 0) {
        const [, fieldName, exampleValue] = exampleMatch;
        const lastResponse = documentation.responses[documentation.responses.length - 1];
        if (lastResponse.fields) {
          const field = lastResponse.fields.find(f => f.name === fieldName);
          if (field) {
            // Try to parse as JSON, otherwise treat as string
            try {
              field.example = JSON.parse(exampleValue);
            } catch {
              // Remove quotes if present
              field.example = exampleValue.replace(/^["'](.+)["']$/, '$1');
            }
          }
        }
      }
      break;
    }
  }
}

/**
 * Substitute variables in a string
 * {{varName}} -> value from vars
 */
export function substituteVariables(
  text: string,
  vars: Record<string, string>
): string {
  return text.replace(/\{\{(\w+)\}\}/g, (_, varName) => {
    return vars[varName] ?? `{{${varName}}}`;
  });
}

/**
 * Apply variable substitution to a request
 */
export function applyVariables(
  request: HttpRequest,
  vars: Record<string, string>
): HttpRequest {
  const substituted: HttpRequest = {
    ...request,
    url: substituteVariables(request.url, vars),
    headers: {},
  };

  for (const [key, value] of Object.entries(request.headers)) {
    substituted.headers[key] = substituteVariables(value, vars);
  }

  if (request.body) {
    substituted.body = substituteVariables(request.body, vars);
  }

  return substituted;
}

/**
 * Serialize request back to HTTP file format
 */
export function serializeRequest(request: HttpRequest): string {
  let result = '###';
  if (request.name) {
    result += ` ${request.name}`;
  }
  result += `\n${request.method} ${request.url}\n`;

  for (const [key, value] of Object.entries(request.headers)) {
    result += `${key}: ${value}\n`;
  }

  if (request.body) {
    result += `\n${request.body}\n`;
  }

  return result + '\n';
}
