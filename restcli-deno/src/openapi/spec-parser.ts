/**
 * OpenAPI Spec Parser
 * Loads and validates OpenAPI specs from JSON or YAML
 */

import { parseYaml } from "./yaml-parser.ts";
import type { OpenAPISpec } from "./types.ts";

export class OpenAPIParser {
  /**
   * Parse OpenAPI spec from string content
   * Auto-detects JSON or YAML format
   */
  static parse(content: string): OpenAPISpec {
    const trimmed = content.trim();

    let parsed: any;

    // Detect format
    if (trimmed.startsWith('{')) {
      // JSON
      try {
        parsed = JSON.parse(content);
      } catch (error) {
        throw new Error(`Failed to parse JSON: ${error instanceof Error ? error.message : String(error)}`);
      }
    } else {
      // YAML
      try {
        parsed = parseYaml(content);
      } catch (error) {
        throw new Error(`Failed to parse YAML: ${error instanceof Error ? error.message : String(error)}`);
      }
    }

    // Validate it's an OpenAPI spec
    return this.validate(parsed);
  }

  /**
   * Load OpenAPI spec from file
   */
  static async loadFile(filePath: string): Promise<OpenAPISpec> {
    try {
      const content = await Deno.readTextFile(filePath);
      return this.parse(content);
    } catch (error) {
      throw new Error(`Failed to load file ${filePath}: ${error instanceof Error ? error.message : String(error)}`);
    }
  }

  /**
   * Fetch OpenAPI spec from URL
   */
  static async fetchUrl(url: string): Promise<OpenAPISpec> {
    try {
      const response = await fetch(url);
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
      const content = await response.text();
      return this.parse(content);
    } catch (error) {
      throw new Error(`Failed to fetch ${url}: ${error instanceof Error ? error.message : String(error)}`);
    }
  }

  /**
   * Validate that the parsed object is a valid OpenAPI spec
   */
  private static validate(obj: any): OpenAPISpec {
    if (!obj || typeof obj !== 'object') {
      throw new Error('Invalid OpenAPI spec: must be an object');
    }

    // Check required fields
    if (!obj.openapi || typeof obj.openapi !== 'string') {
      throw new Error('Invalid OpenAPI spec: missing or invalid "openapi" version field');
    }

    if (!obj.info || typeof obj.info !== 'object') {
      throw new Error('Invalid OpenAPI spec: missing or invalid "info" object');
    }

    if (!obj.info.title || typeof obj.info.title !== 'string') {
      throw new Error('Invalid OpenAPI spec: missing or invalid "info.title"');
    }

    if (!obj.info.version || typeof obj.info.version !== 'string') {
      throw new Error('Invalid OpenAPI spec: missing or invalid "info.version"');
    }

    if (!obj.paths || typeof obj.paths !== 'object') {
      throw new Error('Invalid OpenAPI spec: missing or invalid "paths" object');
    }

    // Warn if OpenAPI version is not 3.x
    const version = obj.openapi;
    if (!version.startsWith('3.')) {
      console.warn(`⚠️  OpenAPI version ${version} detected. This parser is optimized for OpenAPI 3.x`);
    }

    return obj as OpenAPISpec;
  }

  /**
   * Get a user-friendly summary of the spec
   */
  static getSummary(spec: OpenAPISpec): string {
    const pathCount = Object.keys(spec.paths).length;
    const operationCount = this.countOperations(spec);
    const serverUrl = spec.servers?.[0]?.url || 'No server specified';

    return `
📊 OpenAPI Spec Summary

Title: ${spec.info.title}
Version: ${spec.info.version}
${spec.info.description ? `Description: ${spec.info.description}\n` : ''}
OpenAPI Version: ${spec.openapi}
Server: ${serverUrl}

📁 Endpoints: ${pathCount} paths, ${operationCount} operations
${spec.tags ? `\n🏷️  Tags: ${spec.tags.map(t => t.name).join(', ')}` : ''}
    `.trim();
  }

  /**
   * Count total operations in the spec
   */
  private static countOperations(spec: OpenAPISpec): number {
    let count = 0;
    const methods = ['get', 'post', 'put', 'delete', 'patch', 'options', 'head', 'trace'];

    for (const pathItem of Object.values(spec.paths)) {
      for (const method of methods) {
        if ((pathItem as any)[method]) {
          count++;
        }
      }
    }

    return count;
  }
}
