import {
  ApiError,
  type BackendInfo,
  type ChannelConfig,
  type ChannelStatusMap,
  type ChannelTestResult,
  type ChatMessage,
  type FileContext,
  type HealthErrorEntry,
  type HealthSummary,
  type IdentityFiles,
  type IdentitySaveResponse,
  type MCPPreset,
  type MCPStatusMap,
  type MCPTestResponse,
  type MediaAttachment,
  type RecentFileEntry,
  type MemoryEntry,
  type MemorySettings,
  type MemoryStats,
  type Reminder,
  type RemindersResponse,
  type SSEAskUser,
  type SSEChunk,
  type SSEError,
  type SSEStreamEnd,
  type SSEThinking,
  type SSEToolResult,
  type SSEToolStart,
  type SecurityAuditResponse,
  type Session,
  type SessionListResponse,
  type Settings,
  type Skill,
  type VersionInfo,
} from "./types";
import type { TaskStatus, TaskPriority, DocumentType } from "$lib/types/pawkit";
import { BACKEND_URL, API_PREFIX } from "./config";

/** Default timeout for REST requests (30 seconds). */
const REQUEST_TIMEOUT_MS = 30_000;

/** Timeout for streaming requests (5 minutes). */
const STREAM_TIMEOUT_MS = 5 * 60_000;

/** Timeout between stream chunks before we consider it stalled (60 seconds). */
const STREAM_STALL_TIMEOUT_MS = 60_000;

/**
 * Map raw fetch/network errors to user-friendly messages.
 */
export function friendlyErrorMessage(err: unknown): string {
  if (err instanceof ApiError) {
    if (err.status === 0) return "Could not reach the backend. Is it running?";
    if (err.status === 401) return "Session expired. Please sign in again.";
    if (err.status === 502 || err.status === 503)
      return "Backend is starting up or temporarily unavailable. Try again in a moment.";
    if (err.status === 504) return "Request timed out. The backend may be overloaded.";
    if (err.detail) return err.detail;
    return err.message;
  }
  if (err instanceof DOMException && err.name === "AbortError")
    return "Request timed out. The backend may be unresponsive.";
  if (err instanceof TypeError)
    return "Could not reach the backend. Check your connection and make sure it's running.";
  if (err instanceof Error) return err.message;
  return "An unexpected error occurred.";
}

export class PocketPawClient {
  private baseUrl: string;
  private apiBase: string;
  private token: string | null;

  constructor(baseUrl?: string, token?: string) {
    this.baseUrl = (baseUrl ?? BACKEND_URL).replace(/\/+$/, "");
    this.apiBase = `${this.baseUrl}${API_PREFIX}`;
    this.token = token ?? null;
  }

  setToken(token: string) {
    this.token = token;
  }

  /** Returns the API base URL (e.g. "http://localhost:8888/api/v1") used for direct fetch calls */
  getApiBase(): string {
    return this.apiBase;
  }

  /** Returns the WebSocket URL derived from the base URL */
  getWsUrl(): string {
    return this.baseUrl.replace(/^http/, "ws") + `${API_PREFIX}/ws`;
  }

  // ---------------------------------------------------------------------------
  // Internal helpers
  // ---------------------------------------------------------------------------

  private headers(extra?: Record<string, string>): Record<string, string> {
    const h: Record<string, string> = { "Content-Type": "application/json", ...extra };
    if (this.token) h["Authorization"] = `Bearer ${this.token}`;
    return h;
  }

  private async request<T>(
    method: string,
    path: string,
    body?: unknown,
    _retried = false,
  ): Promise<T> {
    const url = `${this.apiBase}${path}`;
    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS);
    let res: Response;
    try {
      res = await fetch(url, {
        method,
        headers: this.headers(),
        body: body != null ? JSON.stringify(body) : undefined,
        signal: controller.signal,
      });
    } catch (err) {
      clearTimeout(timer);
      throw new ApiError(0, friendlyErrorMessage(err));
    }
    clearTimeout(timer);
    if (!res.ok) {
      // On 401, try refreshing the token and retrying once
      if (res.status === 401 && !_retried) {
        try {
          const { readTokens } = await import("$lib/auth/token-store");
          const { refreshAccessToken } = await import("$lib/auth/token-refresh");
          const tokens = await readTokens();
          if (tokens) {
            const newTokens = await refreshAccessToken(tokens);
            this.setToken(newTokens.access_token);
            return this.request<T>(method, path, body, true);
          }
        } catch {
          // Refresh failed — fall through to original error
        }
      }
      let detail: string | undefined;
      try {
        const json = await res.json();
        detail = json.detail ?? json.message ?? JSON.stringify(json);
      } catch {
        detail = await res.text().catch(() => undefined);
      }
      throw new ApiError(res.status, `${method} ${path} failed: ${res.status}`, detail);
    }
    const text = await res.text();
    if (!text) return undefined as T;
    return JSON.parse(text) as T;
  }

  private get<T>(path: string) {
    return this.request<T>("GET", path);
  }

  private post<T>(path: string, body?: unknown) {
    return this.request<T>("POST", path, body);
  }

  private put<T>(path: string, body?: unknown) {
    return this.request<T>("PUT", path, body);
  }

  private del<T>(path: string, body?: unknown) {
    return this.request<T>("DELETE", path, body);
  }

  /** Like request(), but targets /api/mission-control instead of /api/v1 */
  private async mcRequest<T>(
    method: string,
    path: string,
    body?: unknown,
  ): Promise<T> {
    const url = `${this.baseUrl}/api/mission-control${path}`;
    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS);
    let res: Response;
    try {
      res = await fetch(url, {
        method,
        headers: this.headers(),
        body: body != null ? JSON.stringify(body) : undefined,
        signal: controller.signal,
      });
    } catch (err) {
      clearTimeout(timer);
      throw new ApiError(0, friendlyErrorMessage(err));
    }
    clearTimeout(timer);
    if (!res.ok) {
      let detail: string | undefined;
      try {
        const json = await res.json();
        detail = json.detail ?? json.message ?? JSON.stringify(json);
      } catch {
        detail = await res.text().catch(() => undefined);
      }
      throw new ApiError(res.status, `${method} /mc${path} failed: ${res.status}`, detail);
    }
    const text = await res.text();
    if (!text) return undefined as T;
    return JSON.parse(text) as T;
  }

  private async dwRequest<T>(
    method: string,
    path: string,
    body?: unknown,
  ): Promise<T> {
    const url = `${this.baseUrl}/api/deep-work${path}`;
    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS);
    let res: Response;
    try {
      res = await fetch(url, {
        method,
        headers: this.headers(),
        body: body != null ? JSON.stringify(body) : undefined,
        signal: controller.signal,
      });
    } catch (err) {
      clearTimeout(timer);
      throw new ApiError(0, friendlyErrorMessage(err));
    }
    clearTimeout(timer);
    if (!res.ok) {
      let detail: string | undefined;
      try {
        const json = await res.json();
        detail = json.detail ?? json.message ?? JSON.stringify(json);
      } catch {
        detail = await res.text().catch(() => undefined);
      }
      throw new ApiError(res.status, `${method} /dw${path} failed: ${res.status}`, detail);
    }
    const text = await res.text();
    if (!text) return undefined as T;
    return JSON.parse(text) as T;
  }

  // ---------------------------------------------------------------------------
  // Auth
  // ---------------------------------------------------------------------------

  async login(token: string): Promise<void> {
    this.token = token;
    await this.post("/auth/login", { token });
  }

  /**
   * Call the login endpoint with `credentials: "include"` so the browser
   * stores the session cookie for the backend origin.  The WebSocket handler
   * validates this cookie, avoiding the need to pass tokens in the URL.
   */
  async loginForSession(token: string): Promise<void> {
    const url = `${this.apiBase}/auth/login`;
    const res = await fetch(url, {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ token }),
    });
    if (!res.ok) {
      let detail: string | undefined;
      try {
        const json = await res.json();
        detail = json.detail ?? json.message ?? JSON.stringify(json);
      } catch {
        detail = await res.text().catch(() => undefined);
      }
      throw new ApiError(res.status, `POST /auth/login failed: ${res.status}`, detail);
    }
  }

  async logout(): Promise<void> {
    await this.post("/auth/logout");
    this.token = null;
  }

  async getSessionToken(token: string): Promise<string> {
    this.token = token;
    const res = await this.post<{ session_token: string }>("/auth/session", {});
    return res.session_token;
  }

  async regenerateToken(): Promise<string> {
    const res = await this.post<{ token: string }>("/token/regenerate");
    return res.token;
  }

  // TRUNCATION_MARKER_DO_NOT_COMMIT
