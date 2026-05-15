const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

type Envelope<T> = {
  data: T;
  error: { code: string; message: string } | null;
  meta: { request_id: string };
};

async function readEnvelope<T>(path: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    credentials: "include",
    cache: "no-store",
  });

  if (!response.ok) {
    const body = await response.json().catch(() => null);
    const errorMsg = body?.error?.message || `API request failed with status ${response.status}`;
    throw new Error(errorMsg);
  }

  const body = (await response.json()) as Envelope<T>;
  if (body.error) {
    throw new Error(body.error.message);
  }
  return body.data;
}

export type UserProfile = {
  id: string;
  username: string;
  role: string;
  created_at: string;
};

export type UserPreferences = {
  user_id: string;
  language: string;
  sidebar_collapsed: boolean;
  updated_at: string;
};

export type ChangePasswordRequest = {
  current_password: string;
  new_password: string;
};

export type UpdatePreferencesRequest = {
  language?: string;
  sidebar_collapsed?: boolean;
};

export async function getUserProfile(): Promise<UserProfile> {
  return readEnvelope<UserProfile>("/api/v1/user/profile");
}

export async function changePassword(req: ChangePasswordRequest): Promise<{ status: string }> {
  return readEnvelope<{ status: string }>("/api/v1/user/password", {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
}

export async function getUserPreferences(): Promise<UserPreferences> {
  return readEnvelope<UserPreferences>("/api/v1/user/preferences");
}

export async function updateUserPreferences(req: UpdatePreferencesRequest): Promise<UserPreferences> {
  return readEnvelope<UserPreferences>("/api/v1/user/preferences", {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
}
