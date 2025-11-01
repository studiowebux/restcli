/**
 * OpenAPI 3.0 Type Definitions
 * Subset needed for converting to .http files
 */

export interface OpenAPISpec {
  openapi: string; // Version (e.g., "3.0.0")
  info: Info;
  servers?: Server[];
  paths: Paths;
  components?: Components;
  security?: SecurityRequirement[];
  tags?: Tag[];
}

export interface Info {
  title: string;
  version: string;
  description?: string;
  termsOfService?: string;
  contact?: Contact;
  license?: License;
}

export interface Contact {
  name?: string;
  url?: string;
  email?: string;
}

export interface License {
  name: string;
  url?: string;
}

export interface Server {
  url: string;
  description?: string;
  variables?: Record<string, ServerVariable>;
}

export interface ServerVariable {
  default: string;
  enum?: string[];
  description?: string;
}

export interface Paths {
  [path: string]: PathItem;
}

export interface PathItem {
  summary?: string;
  description?: string;
  get?: Operation;
  post?: Operation;
  put?: Operation;
  delete?: Operation;
  patch?: Operation;
  options?: Operation;
  head?: Operation;
  trace?: Operation;
  parameters?: Parameter[];
}

export interface Operation {
  summary?: string;
  description?: string;
  operationId?: string;
  tags?: string[];
  parameters?: Parameter[];
  requestBody?: RequestBody;
  responses: Responses;
  security?: SecurityRequirement[];
  deprecated?: boolean;
}

export interface Parameter {
  name: string;
  in: "query" | "header" | "path" | "cookie";
  description?: string;
  required?: boolean;
  deprecated?: boolean;
  schema?: Schema;
  example?: any;
}

export interface RequestBody {
  description?: string;
  required?: boolean;
  content: Content;
}

export interface Content {
  [mediaType: string]: MediaType;
}

export interface MediaType {
  schema?: Schema;
  example?: any;
  examples?: Record<string, Example>;
}

export interface Example {
  summary?: string;
  description?: string;
  value?: any;
}

export interface Schema {
  type?: string;
  format?: string;
  properties?: Record<string, Schema>;
  items?: Schema;
  required?: string[];
  enum?: any[];
  default?: any;
  example?: any;
  description?: string;
  nullable?: boolean;
  readOnly?: boolean;
  writeOnly?: boolean;
  deprecated?: boolean;
  $ref?: string; // Reference to component schema
  anyOf?: Schema[]; // Union types (commonly used for nullable)
  oneOf?: Schema[]; // Exclusive union types
  allOf?: Schema[]; // Composition
  additionalProperties?: Schema | boolean; // For map-like objects
}

export interface Responses {
  [statusCode: string]: Response;
}

export interface Response {
  description: string;
  content?: Content;
  headers?: Record<string, Header>;
}

export interface Header {
  description?: string;
  required?: boolean;
  schema?: Schema;
}

export interface Components {
  schemas?: Record<string, Schema>;
  responses?: Record<string, Response>;
  parameters?: Record<string, Parameter>;
  examples?: Record<string, Example>;
  requestBodies?: Record<string, RequestBody>;
  headers?: Record<string, Header>;
  securitySchemes?: Record<string, SecurityScheme>;
}

export interface SecurityScheme {
  type: "apiKey" | "http" | "oauth2" | "openIdConnect";
  description?: string;
  name?: string; // For apiKey
  in?: "query" | "header" | "cookie"; // For apiKey
  scheme?: string; // For http (e.g., "bearer")
  bearerFormat?: string; // For http bearer
}

export interface SecurityRequirement {
  [name: string]: string[];
}

export interface Tag {
  name: string;
  description?: string;
}
