import { isYamlFormat, parseYamlHttpFile } from "./yaml-parser.ts";

export interface Parameter {
  name: string;
  type: string;
  required: boolean;
  description?: string;
  example?: string;
}

export interface Response {
  code: string;
  description: string;
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
    let methodLineIdx = -1;
    for (let i = 0; i < lines.length; i++) {
      const line = lines[i].trim();
      if (!line) continue;

      // Method must be followed by a URL (http://, https://, /, or {{var}})
      const methodMatch = line.match(/^(GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS)\s+(https?:\/\/.+|\/.*|\{\{.+)$/i);
      if (methodMatch) {
        method = methodMatch[1].toUpperCase();
        url = methodMatch[2].trim();
        methodLineIdx = i;
        bodyStartIndex = i + 1;
        break;
      }
    }

    // Parse lines before method line for name and documentation annotations
    if (methodLineIdx > 0) {
      for (let i = 0; i < methodLineIdx; i++) {
        const line = lines[i].trim();
        if (!line) continue;

        // Check for documentation annotations (# @...)
        const annotationMatch = line.match(/^#\s*@(\w+)\s+(.+)$/);
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
 * Supports: @description, @tag, @param, @example, @response
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
        documentation.responses.push({ code, description });
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
