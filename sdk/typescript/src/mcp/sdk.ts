/**
 * Managing-Up MCP SDK for TypeScript/Node.js
 */

export enum TransportType {
  STDIO = "stdio",
  HTTP = "http",
  SSE = "sse",
}

export interface MCPServerConfig {
  name: string;
  version: string;
  description?: string;
  transportType: TransportType;
  url?: string;
  command?: string;
  args?: string[];
  env?: Record<string, string>;
}

export interface Metrics {
  requests_total: number;
  requests_success: number;
  requests_error: number;
  latency_seconds: number;
  tool_calls: Record<string, number>;
}

export interface ToolHandler {
  (args: Record<string, unknown>): Promise<string>;
}

export interface ToolDefinition {
  name: string;
  description: string;
  inputSchema: Record<string, unknown>;
  handler: ToolHandler;
}

export class MCPServer {
  private config: MCPServerConfig;
  private platformUrl: string;
  private token: string;
  private serverId?: string;
  private tools: Map<string, ToolDefinition> = new Map();
  private metrics: Metrics = {
    requests_total: 0,
    requests_success: 0,
    requests_error: 0,
    latency_seconds: 0,
    tool_calls: {},
  };

  constructor(config: MCPServerConfig) {
    this.config = config;
    this.platformUrl = process.env.MANAGING_UP_PLATFORM_URL || "http://localhost:8080";
    this.token = process.env.MANAGING_UP_TOKEN || "";
  }

  static fromEnv(): MCPServer {
    return new MCPServer({
      name: process.env.MANAGING_UP_NAME || "typescript-mcp-server",
      version: process.env.MANAGING_UP_VERSION || "1.0.0",
      description: process.env.MANAGING_UP_DESCRIPTION,
      transportType: (process.env.MANAGING_UP_TRANSPORT_TYPE as TransportType) || TransportType.HTTP,
      url: process.env.MANAGING_UP_URL,
    });
  }

  addTool(definition: ToolDefinition): void {
    this.tools.set(definition.name, definition);
  }

  async register(): Promise<{ id?: string; name?: string; status?: string } | null> {
    const url = `${this.platformUrl}/api/v1/mcp-servers`;
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
    };
    if (this.token) {
      headers["Authorization"] = `Bearer ${this.token}`;
    }

    const payload = {
      name: this.config.name,
      version: this.config.version,
      description: this.config.description || "",
      transport_type: this.config.transportType,
      url: this.config.url || "",
      status: "pending",
    };

    try {
      const response = await fetch(url, {
        method: "POST",
        headers,
        body: JSON.stringify(payload),
      });

      if (response.status === 201) {
        const data = await response.json();
        this.serverId = data.data?.id;
        console.log(`Registered with Managing-Up platform: ${this.serverId}`);
        return data.data;
      } else {
        console.warn(`Registration failed with status ${response.status}`);
        return null;
      }
    } catch (error) {
      console.warn(`Registration failed: ${error}`);
      return null;
    }
  }

  recordRequest(success: boolean, latencyMs: number): void {
    this.metrics.requests_total++;
    if (success) {
      this.metrics.requests_success++;
    } else {
      this.metrics.requests_error++;
    }
    this.metrics.latency_seconds += latencyMs / 1000;
  }

  recordToolCall(toolName: string): void {
    if (!this.metrics.tool_calls[toolName]) {
      this.metrics.tool_calls[toolName] = 0;
    }
    this.metrics.tool_calls[toolName]++;
  }

  getMetrics(): Metrics {
    return { ...this.metrics };
  }

  async handleRequest(request: {
    method: string;
    params?: { name?: string; arguments?: Record<string, unknown> };
    id?: string | number;
  }): Promise<{ jsonrpc: string; id?: string | number; result?: unknown; error?: { code: number; message: string } }> {
    const startTime = Date.now();

    try {
      if (request.method === "tools/list") {
        const tools = Array.from(this.tools.values()).map((t) => ({
          name: t.name,
          description: t.description,
          inputSchema: t.inputSchema,
        }));
        return { jsonrpc: "2.0", id: request.id, result: { tools } };
      }

      if (request.method === "tools/call") {
        const toolName = request.params?.name;
        const arguments_ = request.params?.arguments || {};

        if (toolName && this.tools.has(toolName)) {
          const tool = this.tools.get(toolName)!;
          const result = await tool.handler(arguments_);
          this.recordToolCall(toolName);
          this.recordRequest(true, Date.now() - startTime);
          return {
            jsonrpc: "2.0",
            id: request.id,
            result: { content: [{ type: "text", text: String(result) }] },
          };
        } else {
          this.recordRequest(false, Date.now() - startTime);
          return {
            jsonrpc: "2.0",
            id: request.id,
            error: { code: -32601, message: `Tool not found: ${toolName}` },
          };
        }
      }

      return {
        jsonrpc: "2.0",
        id: request.id,
        error: { code: -32601, message: "Method not found" },
      };
    } catch (error) {
      this.recordRequest(false, Date.now() - startTime);
      return {
        jsonrpc: "2.0",
        id: request.id,
        error: { code: -32603, message: String(error) },
      };
    }
  }

  async serveStdio(): Promise<void> {
    const readline = await import("readline");
    const rl = readline.createInterface({
      input: process.stdin,
      output: process.stdout,
      terminal: false,
    });

    for await (const line of rl) {
      try {
        const request = JSON.parse(line.trim());
        const response = await this.handleRequest(request);
        if (response) {
          console.log(JSON.stringify(response));
        }
      } catch {
        // Ignore parse errors
      }
    }
  }

  createHandler(): (request: Request) => Promise<Response> {
    return async (request: Request): Promise<Response> => {
      if (request.method === "GET") {
        return new Response(JSON.stringify({ tools: Array.from(this.tools.keys()) }), {
          headers: { "Content-Type": "application/json" },
        });
      }

      try {
        const body = await request.json();
        const response = await this.handleRequest(body);
        return new Response(JSON.stringify(response), {
          headers: { "Content-Type": "application/json" },
        });
      } catch (error) {
        return new Response(
          JSON.stringify({ jsonrpc: "2.0", error: { code: -32700, message: "Parse error" } }),
          { status: 400, headers: { "Content-Type": "application/json" } }
        );
      }
    };
  }
}

export default MCPServer;
