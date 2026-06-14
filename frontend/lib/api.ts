import type { ActivityLog, Dashboard, LoginSession, Settings, TemporaryToken, UploadRecord, UploadSession } from "@/types/api";

const API_BASES = [
  process.env.NEXT_PUBLIC_API_BASE,
  "http://127.0.0.1:8080",
  "http://localhost:8080"
].filter(Boolean) as string[];

let activeAPIBase = API_BASES[0];

async function request<T>(path: string, init?: RequestInit): Promise<T> {
	const method = init?.method?.toUpperCase() ?? "GET";
	const csrf = typeof window !== "undefined" ? window.localStorage.getItem("pt_csrf") : null;
	const session = typeof window !== "undefined" ? window.localStorage.getItem("pt_session") : null;
	const requestInit: RequestInit = {
		credentials: "include",
		headers: {
			"Content-Type": "application/json",
			...(session ? { Authorization: `Bearer ${session}` } : {}),
			...(method !== "GET" && csrf ? { "X-CSRF-Token": csrf } : {}),
			...(init?.headers ?? {})
		},
		...init
	};
	let lastError: unknown;
	for (const base of orderedAPIBases()) {
		for (let attempt = 0; attempt < 3; attempt++) {
			try {
				const response = await fetch(`${base}${path}`, requestInit);
				activeAPIBase = base;
				if (!response.ok) {
					throw new Error(await response.text());
				}
				const text = await response.text();
				return (text ? JSON.parse(text) : undefined) as T;
			} catch (error) {
				lastError = error;
				await delay(400);
			}
		}
	}
	throw lastError instanceof Error ? lastError : new Error("Failed to fetch");
}

function orderedAPIBases(): string[] {
	return [activeAPIBase, ...API_BASES.filter((base) => base !== activeAPIBase)];
}

function delay(ms: number): Promise<void> {
	return new Promise((resolve) => window.setTimeout(resolve, ms));
}

export async function login(username: string, password: string): Promise<LoginSession> {
  const session = await request<LoginSession>("/api/auth/login", {
    method: "POST",
    body: JSON.stringify({ username, password })
  });
  window.localStorage.setItem("pt_csrf", session.csrf_token);
  window.localStorage.setItem("pt_session", session.session_id);
  return session;
}

export function logout(): Promise<{ ok: boolean }> {
  const result = request<{ ok: boolean }>("/api/auth/logout", { method: "POST", body: "{}" });
  window.localStorage.removeItem("pt_csrf");
  window.localStorage.removeItem("pt_session");
  return result;
}

export function getDashboard(): Promise<Dashboard> {
  return request<Dashboard>("/api/dashboard");
}

export function createTemporaryToken(): Promise<TemporaryToken> {
  return request<TemporaryToken>("/api/tokens", { method: "POST", body: "{}" });
}

export function getSettings(): Promise<Settings> {
  return request<Settings>("/api/settings");
}

export function saveSettings(settings: Settings): Promise<Settings> {
  return request<Settings>("/api/settings", {
    method: "PUT",
    body: JSON.stringify(settings)
  });
}

export function getLogs(): Promise<ActivityLog[]> {
  return request<ActivityLog[]>("/api/logs");
}

export function logsCSVURL(): string {
	return `${activeAPIBase}/api/logs.csv`;
}

export function setup(username: string, password: string): Promise<void> {
  return request<void>("/api/setup", {
    method: "POST",
    body: JSON.stringify({ username, password })
  });
}

export async function createUploadSession(file: File, deviceName: string, uploadToken: string | null): Promise<UploadSession> {
  return request<UploadSession>("/api/upload-sessions", {
    method: "POST",
    headers: uploadToken ? { "X-Upload-Token": uploadToken } : {},
    body: JSON.stringify({
      filename: file.name,
      filesize: file.size,
      device_name: deviceName,
      duplicate_policy: "skip"
    })
  });
}

export async function uploadFile(file: File, deviceName: string, uploadToken: string | null, onProgress: (sent: number) => void): Promise<UploadRecord> {
  const session = await createUploadSession(file, deviceName, uploadToken);
  let offset = 0;
  let index = 0;
  while (offset < file.size) {
    const chunk = file.slice(offset, offset + session.chunk_size);
		const response = await fetch(`${activeAPIBase}/api/upload-sessions/${session.id}/chunks/${index}`, {
      method: "PUT",
      headers: uploadToken
        ? { "X-Upload-Token": uploadToken }
        : {
            "X-CSRF-Token": window.localStorage.getItem("pt_csrf") ?? "",
            Authorization: `Bearer ${window.localStorage.getItem("pt_session") ?? ""}`
          },
      body: chunk
    });
    if (!response.ok) throw new Error(await response.text());
    offset += chunk.size;
    index += 1;
    onProgress(offset);
  }
  return request<UploadRecord>(`/api/upload-sessions/${session.id}/complete`, {
    method: "POST",
    headers: uploadToken ? { "X-Upload-Token": uploadToken } : {},
    body: JSON.stringify({ duplicate_policy: "skip" })
  });
}

export function wsURL(): string {
  const url = new URL(activeAPIBase);
  url.protocol = url.protocol === "https:" ? "wss:" : "ws:";
  url.pathname = "/api/ws";
  return url.toString();
}
