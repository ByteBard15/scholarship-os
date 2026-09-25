const baseURL = (
  import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080"
).replace(/\/$/, "");

type ErrorEnvelope = { error?: { code?: string; message?: string } };

export async function getJSON<T>(path: string): Promise<T> {
  const response = await fetch(`${baseURL}${path}`, {
    headers: { Accept: "application/json" },
  });
  if (!response.ok) {
    const body = (await response.json().catch(() => ({}))) as ErrorEnvelope;
    throw new Error(
      body.error?.message ?? `Request failed with status ${response.status}`,
    );
  }
  return response.json() as Promise<T>;
}
