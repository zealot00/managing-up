const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

export interface LoginRequest {
  username: string;
  password: string;
}

export interface User {
  id: string;
  username: string;
  role: string;
}

export async function login(req: LoginRequest): Promise<User> {
  const res = await fetch(`${API_BASE_URL}/api/v1/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const data = await res.json();
    throw new Error(data.error?.message || "Login failed");
  }
  const body = await res.json();
  return body.data;
}

export async function logout(): Promise<void> {
  await fetch(`${API_BASE_URL}/api/v1/auth/logout`, {
    method: "POST",
    credentials: "include",
  });
}

export async function getCurrentUser(): Promise<User | null> {
  const res = await fetch(`${API_BASE_URL}/api/v1/auth/me`, {
    credentials: "include",
  });
  if (res.ok) {
    const body = await res.json();
    return body.data;
  }
  return null;
}
