import { type HttpRequest, applyVariables } from "./parser.ts";

export interface RequestResult {
  status: number;
  statusText: string;
  headers: Record<string, string>;
  body: string;
  duration: number;
  error?: string;
}

export class RequestExecutor {
  private baseUrl: string;

  constructor(baseUrl: string = "") {
    this.baseUrl = baseUrl;
  }

  async execute(
    request: HttpRequest,
    variables: Record<string, string>,
    profileHeaders: Record<string, string> = {}
  ): Promise<RequestResult> {
    const start = performance.now();

    try {
      // Apply variables
      const substituted = applyVariables(request, variables);

      // Build full URL
      let url = substituted.url;
      if (this.baseUrl && !url.startsWith("http")) {
        url = this.baseUrl + url;
      }

      // Merge headers: request headers override profile headers
      const headers = { ...profileHeaders, ...substituted.headers };

      // Make request
      const response = await fetch(url, {
        method: substituted.method,
        headers,
        body: substituted.body,
      });

      const duration = performance.now() - start;

      // Get response headers
      const responseHeaders: Record<string, string> = {};
      response.headers.forEach((value, key) => {
        responseHeaders[key] = value;
      });

      // Get body
      const body = await response.text();

      return {
        status: response.status,
        statusText: response.statusText,
        headers: responseHeaders,
        body,
        duration,
      };
    } catch (error) {
      const duration = performance.now() - start;
      return {
        status: 0,
        statusText: "Error",
        headers: {},
        body: "",
        duration,
        error: error instanceof Error ? error.message : String(error),
      };
    }
  }
}
