/**
 * OpenAPI to .http File Converter
 * Converts OpenAPI operations to documented .http request files
 */

import type { OpenAPISpec, Operation, Parameter, Schema } from "./types.ts";
import * as path from "@std/path";

export interface ConversionOptions {
  outputDir: string;
  organizeBy?: "tags" | "paths" | "flat";
  baseUrlVariable?: string;
}

export class OpenAPIConverter {
  private spec: OpenAPISpec;
  private options: ConversionOptions;
  private baseUrl: string;
  private usedFilePaths: Map<string, number>;

  constructor(spec: OpenAPISpec, options: ConversionOptions) {
    this.spec = spec;
    this.options = {
      organizeBy: "paths",
      baseUrlVariable: "baseUrl",
      ...options,
    };
    this.baseUrl = spec.servers?.[0]?.url || "http://localhost";
    this.usedFilePaths = new Map();
  }

  /**
   * Convert the entire spec to .http files
   */
  async convert(): Promise<{ filesCreated: number; summary: string }> {
    let filesCreated = 0;
    const filesByTag: Record<string, number> = {};

    // Reset used paths tracker
    this.usedFilePaths.clear();

    // Ensure output directory exists
    await Deno.mkdir(this.options.outputDir, { recursive: true });

    // Create .session.json with baseUrl if it doesn't exist
    await this.ensureSessionFile();

    // Extract enum parameters and create/update profile
    await this.ensureProfileWithEnums();

    // Process each path
    for (const [urlPath, pathItem] of Object.entries(this.spec.paths)) {
      const methods = ["get", "post", "put", "delete", "patch", "options", "head"] as const;

      for (const method of methods) {
        const operation = pathItem[method];
        if (!operation) continue;

        // Generate .http file
        const content = this.generateHttpFile(urlPath, method, operation);

        // Determine file location with deduplication
        const filePath = this.getUniqueFilePath(urlPath, method, operation);

        // Ensure directory exists
        const dir = path.dirname(filePath);
        await Deno.mkdir(dir, { recursive: true });

        // Write file
        await Deno.writeTextFile(filePath, content);
        filesCreated++;

        // Track by tag
        const tag = operation.tags?.[0] || "untagged";
        filesByTag[tag] = (filesByTag[tag] || 0) + 1;
      }
    }

    // Generate summary
    const summary = this.generateSummary(filesCreated, filesByTag);

    return { filesCreated, summary };
  }

  /**
   * Ensure .session.json exists with baseUrl variable
   */
  private async ensureSessionFile(): Promise<void> {
    const { ConfigManager } = await import("../config.ts");
    const configManager = new ConfigManager();
    const configDir = configManager.getConfigDir();
    const sessionPath = path.join(configDir, ".session.json");

    try {
      // Check if file exists
      await Deno.stat(sessionPath);

      // File exists, try to read and update it
      const content = await Deno.readTextFile(sessionPath);
      const session = JSON.parse(content);

      // Only add baseUrl if it doesn't exist
      if (!session.variables) {
        session.variables = {};
      }

      if (!session.variables[this.options.baseUrlVariable!]) {
        session.variables[this.options.baseUrlVariable!] = this.baseUrl;
        await Deno.writeTextFile(sessionPath, JSON.stringify(session, null, 2) + "\n");
      }
    } catch (_error) {
      // File doesn't exist, create it
      const session = {
        activeProfile: null,
        variables: {
          [this.options.baseUrlVariable!]: this.baseUrl,
        },
      };

      await Deno.writeTextFile(sessionPath, JSON.stringify(session, null, 2) + "\n");
    }
  }

  /**
   * Extract enum parameters and create/update profile with multi-value variables
   */
  private async ensureProfileWithEnums(): Promise<void> {
    const { ConfigManager } = await import("../config.ts");
    const configManager = new ConfigManager();
    const configDir = configManager.getConfigDir();
    const profilesPath = path.join(configDir, ".profiles.json");

    // Collect all enum parameters
    const enumParams: Map<string, string[]> = new Map();

    for (const [, pathItem] of Object.entries(this.spec.paths)) {
      const methods = ["get", "post", "put", "delete", "patch", "options", "head"] as const;

      for (const method of methods) {
        const operation = pathItem[method];
        if (!operation) continue;

        const allParams = this.getAllParameters(operation, "");

        for (const param of allParams) {
          if (param.schema?.enum && param.schema.enum.length > 0) {
            // Convert enum values to strings
            const enumValues = param.schema.enum.map(v => String(v));

            // Merge with existing values if param name already seen
            if (enumParams.has(param.name)) {
              const existing = enumParams.get(param.name)!;
              const merged = Array.from(new Set([...existing, ...enumValues]));
              enumParams.set(param.name, merged);
            } else {
              enumParams.set(param.name, enumValues);
            }
          }
        }
      }
    }

    // If no enum parameters found, skip profile creation
    if (enumParams.size === 0) {
      return;
    }

    // Read or create profiles file
    let profiles: any[] = [];
    try {
      const content = await Deno.readTextFile(profilesPath);
      profiles = JSON.parse(content);
    } catch {
      // File doesn't exist, will create new
    }

    // Find or create "OpenAPI" profile
    let openApiProfile = profiles.find(p => p.name === "OpenAPI");

    if (!openApiProfile) {
      openApiProfile = {
        name: "OpenAPI",
        headers: {},
        variables: {},
      };
      profiles.push(openApiProfile);
    }

    // Ensure variables object exists
    if (!openApiProfile.variables) {
      openApiProfile.variables = {};
    }

    // Add enum parameters as multi-value variables
    for (const [paramName, enumValues] of enumParams.entries()) {
      // Only add if variable doesn't already exist (don't overwrite user changes)
      if (!(paramName in openApiProfile.variables)) {
        openApiProfile.variables[paramName] = {
          options: enumValues,
          active: 0,
          description: `Enum parameter from OpenAPI spec`,
        };
      }
    }

    // Write profiles file
    await Deno.writeTextFile(profilesPath, JSON.stringify(profiles, null, 2) + "\n");
  }

  /**
   * Generate .http file content for an operation
   */
  private generateHttpFile(
    urlPath: string,
    method: string,
    operation: Operation
  ): string {
    let content = "";

    // Title
    const title = operation.summary || operation.operationId || `${method.toUpperCase()} ${urlPath}`;
    content += `### ${title}\n`;

    // Description
    if (operation.description) {
      content += `# @description ${operation.description}\n`;
    }

    // Tags
    if (operation.tags) {
      for (const tag of operation.tags) {
        content += `# @tag ${tag}\n`;
      }
    }

    // Parameters
    const allParams = this.getAllParameters(operation, urlPath);

    // Body parameters (from requestBody)
    const bodyParams = this.extractBodyParameters(operation);

    for (const param of bodyParams) {
      const required = param.required ? "required" : "optional";
      content += `# @param ${param.name} {${param.type}} ${required}`;
      if (param.description) {
        content += ` - ${param.description}`;
      }
      content += "\n";

      if (param.example !== undefined) {
        const exampleStr = typeof param.example === 'string'
          ? `"${param.example}"`
          : JSON.stringify(param.example);
        content += `# @example ${param.name} ${exampleStr}\n`;
      }
    }

    // Path/query/header parameters
    for (const param of allParams) {
      if (param.in === "path" || param.in === "query") {
        const required = param.required ? "required" : "optional";
        const type = param.schema?.type || "string";
        content += `# @param ${param.name} {${type}} ${required}`;
        if (param.description) {
          content += ` - ${param.description}`;
        }

        // Add enum options if available
        if (param.schema?.enum && param.schema.enum.length > 0) {
          const enumValues = param.schema.enum.map(v => `"${v}"`).join(", ");
          content += ` - Options: ${enumValues}`;
        }

        content += "\n";

        if (param.example !== undefined || param.schema?.example !== undefined) {
          const example = param.example || param.schema?.example;
          const exampleStr = typeof example === 'string' ? `"${example}"` : JSON.stringify(example);
          content += `# @example ${param.name} ${exampleStr}\n`;
        }
      }
    }

    // Responses with schema details
    if (operation.responses) {
      for (const [code, response] of Object.entries(operation.responses)) {
        const desc = response.description || "Response";
        content += `# @response ${code} - ${desc}\n`;

        // Extract response schema fields
        const responseFields = this.extractResponseFields(response);
        if (responseFields.length > 0) {
          for (const field of responseFields) {
            const required = field.required ? "required" : "optional";
            content += `# @response-field ${field.name} {${field.type}} ${required}`;
            if (field.deprecated) {
              content += ` deprecated`;
            }
            if (field.description) {
              content += ` - ${field.description}`;
            }
            content += "\n";

            if (field.example !== undefined) {
              const exampleStr = typeof field.example === 'string'
                ? `"${field.example}"`
                : JSON.stringify(field.example);
              content += `# @response-example ${field.name} ${exampleStr}\n`;
            }
          }
        }
      }
    }

    // HTTP Method and URL
    const url = this.buildUrl(urlPath, allParams);
    content += `${method.toUpperCase()} {{${this.options.baseUrlVariable}}}${url}\n`;

    // Headers
    const headers = this.getHeaders(operation, allParams);
    for (const [key, value] of Object.entries(headers)) {
      content += `${key}: ${value}\n`;
    }

    // Body
    const body = this.generateBody(operation);
    if (body) {
      content += `\n${body}\n`;
    }

    content += "\n###\n";
    return content;
  }

  /**
   * Get all parameters for an operation
   */
  private getAllParameters(operation: Operation, urlPath: string): Parameter[] {
    const params: Parameter[] = [];

    // Operation-level parameters
    if (operation.parameters) {
      params.push(...operation.parameters);
    }

    // Path-level parameters
    const pathItem = Object.values(this.spec.paths).find(p =>
      Object.values(p).includes(operation as any)
    );
    if (pathItem?.parameters) {
      params.push(...pathItem.parameters);
    }

    return params;
  }

  /**
   * Extract parameters from request body schema
   */
  private extractBodyParameters(operation: Operation): Array<{
    name: string;
    type: string;
    required: boolean;
    description?: string;
    example?: any;
  }> {
    const params: Array<{
      name: string;
      type: string;
      required: boolean;
      description?: string;
      example?: any;
    }> = [];

    const requestBody = operation.requestBody;
    if (!requestBody) return params;

    const jsonContent = requestBody.content?.["application/json"];
    if (!jsonContent?.schema) return params;

    // Resolve $ref if present
    const schema = this.resolveSchema(jsonContent.schema);
    const required = schema.required || [];

    if (schema.properties) {
      for (const [name, propSchema] of Object.entries(schema.properties)) {
        params.push({
          name,
          type: propSchema.type || "string",
          required: required.includes(name),
          description: propSchema.description,
          example: propSchema.example,
        });
      }
    }

    return params;
  }

  /**
   * Extract fields from response schema (with nested field support)
   */
  private extractResponseFields(response: any): Array<{
    name: string;
    type: string;
    required: boolean;
    description?: string;
    example?: any;
    deprecated?: boolean;
  }> {
    const fields: Array<{
      name: string;
      type: string;
      required: boolean;
      description?: string;
      example?: any;
      deprecated?: boolean;
    }> = [];

    // Get JSON content
    const jsonContent = response.content?.["application/json"];
    if (!jsonContent?.schema) return fields;

    // Resolve $ref if present
    const schema = this.resolveSchema(jsonContent.schema);

    // Extract fields recursively
    this.extractFieldsRecursive(schema, "", fields);

    return fields;
  }

  /**
   * Recursively extract fields from schema
   */
  private extractFieldsRecursive(
    schema: Schema,
    prefix: string,
    fields: Array<{ name: string; type: string; required: boolean; description?: string; example?: any; deprecated?: boolean }>,
    depth = 0,
    maxDepth = 100
  ): void {
    // Prevent infinite recursion or excessive depth
    if (depth >= maxDepth) {
      return;
    }

    const required = schema.required || [];

    if (schema.properties) {
      for (const [name, propSchema] of Object.entries(schema.properties)) {
        const fullName = prefix ? `${prefix}.${name}` : name;
        const isRequired = required.includes(name);

        // Resolve $ref if present
        const resolvedSchema = this.resolveSchema(propSchema);

        // Check schema type (handle both string and array types for nullable)
        const schemaType = Array.isArray(resolvedSchema.type) ? resolvedSchema.type[0] : resolvedSchema.type;

        // Check if this property has nested structure
        const hasNestedObjectProps = (schemaType === "object" || resolvedSchema.properties) && resolvedSchema.properties;
        const hasNestedArrayProps = schemaType === "array" && resolvedSchema.items?.type === "object" && resolvedSchema.items.properties;

        // Only add the field if it's a leaf node (no nested properties to expand)
        // This skips intermediate object/array containers
        if (!hasNestedObjectProps && !hasNestedArrayProps) {
          const fieldType = this.getFieldType(propSchema);
          fields.push({
            name: fullName,
            type: fieldType,
            required: isRequired,
            description: propSchema.description,
            example: propSchema.example,
            deprecated: propSchema.deprecated || resolvedSchema.deprecated,
          });
        }

        // Recurse into nested objects
        if (hasNestedObjectProps) {
          this.extractFieldsRecursive(resolvedSchema, fullName, fields, depth + 1, maxDepth);
        }

        // Recurse into arrays of objects
        if (hasNestedArrayProps && resolvedSchema.items) {
          this.extractFieldsRecursive(resolvedSchema.items, `${fullName}[]`, fields, depth + 1, maxDepth);
        }
      }
    }
  }

  /**
   * Get a human-readable type description for a schema property
   */
  private getFieldType(schema: any): string {
    if (schema.type === "array" && schema.items) {
      const itemType = schema.items.type || "object";
      return `array<${itemType}>`;
    }
    if (schema.type === "object" && schema.additionalProperties) {
      const valueType = schema.additionalProperties.type || "any";
      return `object<string, ${valueType}>`;
    }
    return schema.type || "string";
  }

  /**
   * Build URL with path parameters as variables
   */
  private buildUrl(urlPath: string, params: Parameter[]): string {
    let url = urlPath;

    // Replace path parameters with {{variables}}
    for (const param of params) {
      if (param.in === "path") {
        url = url.replace(`{${param.name}}`, `{{${param.name}}}`);
      }
    }

    // Add query parameters
    const queryParams = params.filter(p => p.in === "query");
    if (queryParams.length > 0) {
      const queryStr = queryParams
        .map(p => `${p.name}={{${p.name}}}`)
        .join("&");
      url += `?${queryStr}`;
    }

    return url;
  }

  /**
   * Get headers for the request
   */
  private getHeaders(operation: Operation, params: Parameter[]): Record<string, string> {
    const headers: Record<string, string> = {};

    // Content-Type from requestBody
    if (operation.requestBody?.content) {
      const contentTypes = Object.keys(operation.requestBody.content);
      if (contentTypes.length > 0) {
        headers["Content-Type"] = contentTypes[0];
      }
    }

    // Header parameters
    for (const param of params) {
      if (param.in === "header") {
        headers[param.name] = `{{${param.name}}}`;
      }
    }

    // Security (basic authentication or bearer token)
    const security = operation.security || this.spec.security;
    if (security && security.length > 0 && this.spec.components?.securitySchemes) {
      for (const secReq of security) {
        for (const [schemeName] of Object.entries(secReq)) {
          const scheme = this.spec.components.securitySchemes[schemeName];

          if (scheme?.type === "http" && scheme.scheme === "bearer") {
            headers["Authorization"] = "Bearer {{token}}";
          } else if (scheme?.type === "apiKey" && scheme.in === "header" && scheme.name) {
            headers[scheme.name] = `{{${scheme.name}}}`;
          }
        }
      }
    }

    return headers;
  }

  /**
   * Generate request body from schema
   */
  private generateBody(operation: Operation): string | null {
    const requestBody = operation.requestBody;
    if (!requestBody) return null;

    const jsonContent = requestBody.content?.["application/json"];
    if (!jsonContent) return null;

    // Use example if available
    if (jsonContent.example) {
      return JSON.stringify(jsonContent.example, null, 2);
    }

    // Generate from schema
    if (jsonContent.schema) {
      const resolvedSchema = this.resolveSchema(jsonContent.schema);
      const body = this.generateBodyFromSchema(resolvedSchema);
      return JSON.stringify(body, null, 2);
    }

    return null;
  }

  /**
   * Resolve $ref in schema
   */
  private resolveSchema(schema: Schema): Schema {
    // Handle direct $ref
    if ('$ref' in schema && typeof schema.$ref === 'string') {
      // Extract component name from $ref (e.g., "#/components/schemas/Post Claim Request")
      const refPath = schema.$ref.split('/');
      if (refPath[0] === '#' && refPath[1] === 'components' && refPath[2] === 'schemas') {
        const schemaName = refPath.slice(3).join('/');
        const resolvedSchema = this.spec.components?.schemas?.[schemaName];
        if (resolvedSchema) {
          return resolvedSchema as Schema;
        }
      }
    }

    // Handle anyOf/oneOf with null (common pattern for nullable in OpenAPI 3.x)
    if ('anyOf' in schema && Array.isArray(schema.anyOf)) {
      // Find the first non-null schema
      const nonNullSchema = schema.anyOf.find(s => {
        if ('type' in s) return s.type !== 'null';
        if ('$ref' in s) return true;
        return false;
      });
      if (nonNullSchema) {
        return this.resolveSchema(nonNullSchema);
      }
    }

    if ('oneOf' in schema && Array.isArray(schema.oneOf)) {
      // Find the first non-null schema
      const nonNullSchema = schema.oneOf.find(s => {
        if ('type' in s) return s.type !== 'null';
        if ('$ref' in s) return true;
        return false;
      });
      if (nonNullSchema) {
        return this.resolveSchema(nonNullSchema);
      }
    }

    return schema;
  }

  /**
   * Generate example body from schema
   */
  private generateBodyFromSchema(schema: Schema): any {
    if (schema.example !== undefined) {
      return schema.example;
    }

    if (schema.type === "object" && schema.properties) {
      const obj: any = {};
      for (const [key, propSchema] of Object.entries(schema.properties)) {
        obj[key] = this.generateBodyFromSchema(propSchema);
      }
      return obj;
    }

    if (schema.type === "array" && schema.items) {
      return [this.generateBodyFromSchema(schema.items)];
    }

    // Generate example values based on type
    switch (schema.type) {
      case "string":
        return schema.default || "string";
      case "number":
      case "integer":
        return schema.default || 0;
      case "boolean":
        return schema.default || false;
      default:
        return null;
    }
  }

  /**
   * Determine file path for the generated .http file
   */
  private getFilePath(urlPath: string, method: string, operation: Operation): string {
    const baseName = this.generateFileName(urlPath, method, operation);

    if (this.options.organizeBy === "flat") {
      return path.join(this.options.outputDir, baseName);
    }

    if (this.options.organizeBy === "tags") {
      const tag = operation.tags?.[0] || "untagged";
      return path.join(this.options.outputDir, tag, baseName);
    }

    // organize by paths - mirror exact API path structure
    const pathParts = urlPath.split("/").filter(p => p && !p.startsWith("{"));
    const dir = pathParts.length > 0 ? pathParts.join("/") : "root";
    return path.join(this.options.outputDir, dir, baseName);
  }

  /**
   * Get unique file path with deduplication
   */
  private getUniqueFilePath(urlPath: string, method: string, operation: Operation): string {
    const basePath = this.getFilePath(urlPath, method, operation);

    // Check if this path was already used
    const count = this.usedFilePaths.get(basePath);

    if (count === undefined) {
      // First time seeing this path
      this.usedFilePaths.set(basePath, 1);
      return basePath;
    }

    // Path conflict - append counter
    this.usedFilePaths.set(basePath, count + 1);

    // Insert counter before .http extension
    const dir = path.dirname(basePath);
    const filename = path.basename(basePath, '.http');
    return path.join(dir, `${filename}-${count}.http`);
  }

  /**
   * Generate filename for .http file
   */
  private generateFileName(urlPath: string, method: string, operation: Operation): string {
    // Use operationId if available
    if (operation.operationId) {
      return `${operation.operationId}.http`;
    }

    // Use summary
    if (operation.summary) {
      const sanitized = operation.summary
        .toLowerCase()
        .replace(/[^a-z0-9]+/g, "-")
        .replace(/^-|-$/g, "");
      return `${sanitized}.http`;
    }

    // Use path + method
    const pathParts = urlPath.split("/").filter(p => p && !p.startsWith("{"));
    const lastPart = pathParts[pathParts.length - 1] || "root";
    return `${method}-${lastPart}.http`;
  }

  /**
   * Generate conversion summary
   */
  private generateSummary(filesCreated: number, filesByTag: Record<string, number>): string {
    const lines = [
      `✅ Conversion complete!`,
      ``,
      `📁 Created ${filesCreated} .http files`,
      `📂 Output directory: ${this.options.outputDir}`,
      `📋 Organization: ${this.options.organizeBy}`,
      `🔗 Base URL: ${this.baseUrl}`,
      ``,
      `Files by category:`,
    ];

    for (const [tag, count] of Object.entries(filesByTag).sort((a, b) => b[1] - a[1])) {
      lines.push(`  • ${tag}: ${count} file${count > 1 ? 's' : ''}`);
    }

    return lines.join('\n');
  }
}
