import { isYamlFormat, parseYamlHttpFile } from "./yaml-parser.ts";

export interface HttpRequest {
  name?: string;
  method: string;
  url: string;
  headers: Record<string, string>;
  body?: string;
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

    // If we found a method line and there are lines before it, the first non-empty line is the name
    if (methodLineIdx > 0) {
      for (let i = 0; i < methodLineIdx; i++) {
        const line = lines[i].trim();
        if (line) {
          name = line;
          break;
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

    requests.push({ name, method, url, headers, body });
  }

  return { requests };
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
