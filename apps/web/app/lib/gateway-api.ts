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
  client_name: string;
  provider: string;
  model: string;
  request_count: number;
  prompt_tokens: number;
  completion_tokens: number;
  total_tokens: number;
  total_cost: number;
};

export type GatewayUserUsageRow = {
  user_id: string;
  username: string;
  request_count: number;
  prompt_tokens: number;
  completion_tokens: number;
  total_tokens: number;
  total_cost: number;
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

export type GatewayProviderKey = {
  id: string;
  user_id: string;
  provider: string;
  model: string;
  key_prefix: string;
  is_enabled: boolean;
  monthly_limit: number;
  created_at: string;
  updated_at: string;
};

export type UserBudget = {
  id: string;
  user_id: string;
  monthly_limit: number;
  daily_limit: number;
  used_this_month: number;
  used_today: number;
  reset_at: string;
  updated_at: string;
};

export async function listProviderKeys(): Promise<{ items: GatewayProviderKey[] }> {
  return request<{ items: GatewayProviderKey[] }>("/api/v1/gateway/providers");
}

export async function createProviderKey(data: {
  provider: string;
  api_key: string;
  model?: string;
  monthly_limit?: number;
}): Promise<{ item: GatewayProviderKey }> {
  return request<{ item: GatewayProviderKey }>("/api/v1/gateway/providers", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
}

export async function getProviderKey(id: string): Promise<{ item: GatewayProviderKey }> {
  return request<{ item: GatewayProviderKey }>(`/api/v1/gateway/providers/${id}`);
}

export async function updateProviderKey(id: string, data: {
  provider?: string;
  api_key?: string;
  model?: string;
  monthly_limit?: number;
  is_enabled?: boolean;
}): Promise<{ item: GatewayProviderKey }> {
  return request<{ item: GatewayProviderKey }>(`/api/v1/gateway/providers/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
}

export async function deleteProviderKey(id: string): Promise<void> {
  await request<{ status: string }>(`/api/v1/gateway/providers/${id}`, {
    method: "DELETE",
  });
}

export async function toggleProviderKey(id: string, enabled: boolean): Promise<{ item: GatewayProviderKey }> {
  return request<{ item: GatewayProviderKey }>(`/api/v1/gateway/providers/${id}/toggle`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ enabled }),
  });
}

export async function getBudget(): Promise<{ item: UserBudget }> {
  return request<{ item: UserBudget }>("/api/v1/gateway/budget");
}

export async function updateBudget(data: {
  monthly_limit?: number;
  daily_limit?: number;
}): Promise<{ item: UserBudget }> {
  return request<{ item: UserBudget }>("/api/v1/gateway/budget", {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
}

// Gateway Session Types
export type GatewaySession = {
  id: string;
  session_type: string;
  agent_id: string;
  correlation_id: string;
  task_intent: Record<string, unknown>;
  risk_level: string;
  policy_decision: Record<string, unknown> | null;
  status: string;
  started_at: string;
  ended_at: string | null;
  metadata: Record<string, unknown>;
};

// Snapshot Types
export type SkillCapabilitySnapshot = {
  id: string;
  skill_id: string;
  version: string;
  snapshot_type: string;
  dataset_id: string | null;
  run_id: string | null;
  metrics: Record<string, number>;
  overall_score: number;
  passed: boolean;
  gate_policy_id: string | null;
  evaluated_at: string;
  created_at: string;
};

// Gateway Session APIs
export async function listGatewaySessions(params?: {
  agent_id?: string;
  limit?: number;
}): Promise<{ items: GatewaySession[] }> {
  const query = new URLSearchParams();
  if (params?.agent_id) query.set("agent_id", params.agent_id);
  if (params?.limit) query.set("limit", String(params.limit));
  const qs = query.toString();
  return request<{ items: GatewaySession[] }>(`/api/v1/gateway/sessions${qs ? `?${qs}` : ""}`);
}

export async function getGatewaySession(id: string): Promise<{ item: GatewaySession }> {
  return request<{ item: GatewaySession }>(`/api/v1/gateway/sessions/${id}`);
}

// Snapshot APIs
export async function getSnapshot(params: {
  skill_id: string;
  version: string;
}): Promise<{ found: boolean; snapshot: SkillCapabilitySnapshot | null }> {
  const query = new URLSearchParams();
  query.set("skill_id", params.skill_id);
  query.set("version", params.version);
  return request<{ found: boolean; snapshot: SkillCapabilitySnapshot | null }>(`/api/v1/snapshots?${query.toString()}`);
}

export async function listSnapshots(params: {
  skill_id: string;
  limit?: number;
}): Promise<{ items: SkillCapabilitySnapshot[] }> {
  const query = new URLSearchParams();
  query.set("skill_id", params.skill_id);
  if (params?.limit) query.set("limit", String(params.limit));
  return request<{ items: SkillCapabilitySnapshot[] }>(`/api/v1/snapshots/list?${query.toString()}`);
}
