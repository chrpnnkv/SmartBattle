import { CORE_HTTP_BASE } from "./config";

export class HttpError extends Error {
  status: number;
  body: unknown;
  constructor(status: number, body: unknown) {
    super(`HTTP ${status}`);
    this.status = status;
    this.body = body;
  }
}

function joinUrl(base: string, path: string) {
  if (!base) return path; 
  return `${base.replace(/\/$/, "")}${path.startsWith("/") ? "" : "/"}${path}`;
}

export async function http<T>(
  path: string,
  options: RequestInit & { auth?: "teacher" | "none" } = {}
): Promise<T> {
  const url = joinUrl(CORE_HTTP_BASE, path);

  const headers = new Headers(options.headers || {});
  if (!headers.has("Content-Type") && options.body) headers.set("Content-Type", "application/json");

  if (options.auth === "teacher") {
  const token = localStorage.getItem("teacherToken");
  if (token && token !== "null" && token !== "undefined") {
    headers.set("Authorization", `Bearer ${token}`);
  }
}

  const res = await fetch(url, { ...options, headers });

  const contentType = res.headers.get("content-type") || "";
  const body = contentType.includes("application/json") ? await res.json().catch(() => null) : await res.text().catch(() => "");

  if (!res.ok) throw new HttpError(res.status, body);
  return body as T;
}
