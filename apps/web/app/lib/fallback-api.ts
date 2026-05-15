const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

type Envelope<T> = {
  data: T;
  error: { code: string; message: string } | null;
  meta: { request_id: string };
};

export type FallbackTarget = {
  id: string;
  chain_id: string;
  provider: string;
  model: string;
  weight: number;
  priority: number;
  is_enabled: boolean;
};

export type FallbackChain = {
  id: string;
  model: string;
  is_enabled: boolean;
  targets: FallbackTarget[];
  created_at: string;
  updated_at: string;
};

export type CreateFallbackChainRequest = {
  model: string;
  is_enabled: boolean;
  targets: Omit<FallbackTarget, "id" | "chain_id">[];
};

export type UpdateFallbackChainRequest = {
  model?: string;
  is_enabled?: boolean;
  targets?: FallbackTarget[];
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

export async function listFallbackChains(): Promise<{ items: FallbackChain[] }> {
  return request<{ items: FallbackChain[] }>("/api/v1/admin/fallback-chains");
}

export async function createFallbackChain(
  req: CreateFallbackChainRequest
): Promise<FallbackChain> {
  return request<FallbackChain>("/api/v1/admin/fallback-chains", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
}

export async function updateFallbackChain(
  id: string,
  req: UpdateFallbackChainRequest
): Promise<FallbackChain> {
  return request<FallbackChain>(`/api/v1/admin/fallback-chains/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
}

export async function deleteFallbackChain(id: string): Promise<void> {
  await request<{ deleted: boolean }>(`/api/v1/admin/fallback-chains/${id}`, {
    method: "DELETE",
  });
}
