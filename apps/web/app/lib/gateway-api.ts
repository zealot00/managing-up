const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

type Envelope<T> = {
  data: T;
  error: { code: string; message: string } | null;
  meta: { request_id: string };
};

export type GatewayKeyMeta = {
  id: string;
  user_id: string;
  name: string;
  key_prefix: string;
  created_at: string;
  last_used_at?: string | null;
  revoked_at?: string | null;
};

export type GatewayUsageRow = {
  user_id: string;
  username: string;
  provider: string;
  model: string;
  request_count: number;
  prompt_tokens: number;
  completion_tokens: number;
  total_tokens: number;
};

export type GatewayUserUsageRow = {
  user_id: string;
  username: string;
  request_count: number;
  prompt_tokens: number;
  completion_tokens: number;
  total_tokens: number;
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

export async function listGatewayKeys(): Promise<{ items: GatewayKeyMeta[] }> {
  return request<{ items: GatewayKeyMeta[] }>("/api/v1/gateway/keys");
}

export async function createGatewayKey(name: string): Promise<{ key: string; key_meta: GatewayKeyMeta }> {
  return request<{ key: string; key_meta: GatewayKeyMeta }>("/api/v1/gateway/keys", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name }),
  });
}

export async function revokeGatewayKey(id: string): Promise<void> {
  await request<{ status: string }>(`/api/v1/gateway/keys/${id}`, {
    method: "DELETE",
  });
}

export async function getGatewayUsage(params?: {
  from?: string;
  to?: string;
  user_id?: string;
}): Promise<{ items: GatewayUsageRow[] }> {
  const query = new URLSearchParams();
  if (params?.from) query.set("from", params.from);
  if (params?.to) query.set("to", params.to);
  if (params?.user_id) query.set("user_id", params.user_id);
  const qs = query.toString();
  return request<{ items: GatewayUsageRow[] }>(`/api/v1/gateway/usage${qs ? `?${qs}` : ""}`);
}

export async function getGatewayUsageByUsers(params?: {
  from?: string;
  to?: string;
}): Promise<{ items: GatewayUserUsageRow[] }> {
  const query = new URLSearchParams();
  if (params?.from) query.set("from", params.from);
  if (params?.to) query.set("to", params.to);
  const qs = query.toString();
  return request<{ items: GatewayUserUsageRow[] }>(`/api/v1/gateway/usage/users${qs ? `?${qs}` : ""}`);
}
