const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

type Envelope<T> = {
  data: T;
  error: { code: string; message: string } | null;
  meta: { request_id: string };
};

export type MCPServer = {
  id: string;
  name: string;
  description?: string;
  transport_type: string;
  command?: string;
  args?: string[];
  env?: string[];
  url?: string;
  headers?: string[];
  status: string;
  rejection_reason?: string;
  approved_by?: string;
  approved_at?: string;
  is_enabled: boolean;
  created_at: string;
  updated_at: string;
};

export type CreateMCPServerRequest = {
  name: string;
  description?: string;
  transport_type: string;
  command?: string;
  args?: string[];
  env?: string[];
  url?: string;
  headers?: string[];
};

export type UpdateMCPServerRequest = {
  name?: string;
  description?: string;
  transport_type?: string;
  command?: string;
  args?: string[];
  env?: string[];
  url?: string;
  headers?: string[];
  status?: string;
  is_enabled?: boolean;
};

export type ApproveMCPServerRequest = {
  decision: "approved" | "rejected";
  approver: string;
  note?: string;
};

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    credentials: "include",
    cache: "no-store",
    ...init,
  });

  const body = (await response.json()) as Envelope<T>;
  if (!response.ok || body.error) {
    throw new Error(body.error?.message || `Request failed: ${response.status}`);
  }
  return body.data;
}

export async function listMCPServers(): Promise<{ items: MCPServer[] }> {
  return request<{ items: MCPServer[] }>("/api/v1/mcp-servers");
}

export async function getMCPServer(id: string): Promise<MCPServer> {
  return request<MCPServer>(`/api/v1/mcp-servers/${id}`);
}

export async function createMCPServer(
  req: CreateMCPServerRequest
): Promise<MCPServer> {
  return request<MCPServer>("/api/v1/mcp-servers", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
}

export async function updateMCPServer(
  id: string,
  req: UpdateMCPServerRequest
): Promise<MCPServer> {
  return request<MCPServer>(`/api/v1/mcp-servers/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
}

export async function deleteMCPServer(id: string): Promise<void> {
  await request<{ deleted: boolean }>(`/api/v1/mcp-servers/${id}`, {
    method: "DELETE",
  });
}

export async function approveMCPServer(
  id: string,
  req: ApproveMCPServerRequest
): Promise<MCPServer> {
  return request<MCPServer>(`/api/v1/mcp-servers/${id}/approve`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
}
