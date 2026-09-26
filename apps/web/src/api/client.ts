const baseURL = (
  import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080"
).replace(/\/$/, "");
type ErrorEnvelope = { error?: { code?: string; message?: string } };
export async function requestJSON<T>(
  path: string,
  init: RequestInit = {},
): Promise<T> {
  const headers = new Headers(init.headers);
  headers.set("Accept", "application/json");
  if (init.body && !(init.body instanceof FormData))
    headers.set("Content-Type", "application/json");
  const token = authStorage.get();
  if (token) headers.set("Authorization", `Bearer ${token}`);
  const response = await fetch(`${baseURL}${path}`, { ...init, headers });
  if (!response.ok) {
    const body = (await response.json().catch(() => ({}))) as ErrorEnvelope;
    if (response.status === 401 && path !== "/api/v1/auth/login") {
      authStorage.clear();
      window.dispatchEvent(new Event(unauthorizedEvent));
    }
    throw new Error(
      body.error?.message ?? `Request failed with status ${response.status}`,
    );
  }
  if (response.status === 204) return undefined as T;
  return response.json() as Promise<T>;
}
export function getJSON<T>(path: string): Promise<T> {
  return requestJSON<T>(path);
}
import { authStorage, unauthorizedEvent } from "../features/auth/storage";
