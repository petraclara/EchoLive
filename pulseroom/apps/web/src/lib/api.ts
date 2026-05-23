const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
  ) {
    super(message);
  }
}

async function request<T>(
  path: string,
  options: RequestInit = {},
  token?: string,
): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };
  if (token) headers.Authorization = `Bearer ${token}`;

  const res = await fetch(`${API_URL}/v1${path}`, {
    ...options,
    headers,
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new ApiError(body.error || res.statusText, res.status);
  }

  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

export const api = {
  register: (data: { email: string; password: string; name: string }) =>
    request<{ token: string }>("/auth/register", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  login: (data: { email: string; password: string }) =>
    request<{ token: string }>("/auth/login", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  me: (token: string) => request<{ id: string; email: string; name: string }>("/auth/me", {}, token),

  listEvents: (token: string) => request<import("./types").Event[]>("/events", {}, token),

  createEvent: (token: string, title: string) =>
    request<{ event: import("./types").Event; join_url: string }>(
      "/events",
      { method: "POST", body: JSON.stringify({ title }) },
      token,
    ),

  getEvent: (token: string, eventId: string) =>
    request<{
      event: import("./types").Event;
      join_url: string;
      attendee_count: number;
    }>(`/events/${eventId}`, {}, token),

  updateEvent: (
    token: string,
    eventId: string,
    data: { title?: string; status?: string },
  ) =>
    request<import("./types").Event>(
      `/events/${eventId}`,
      { method: "PATCH", body: JSON.stringify(data) },
      token,
    ),

  listAnnouncements: (token: string, eventId: string) =>
    request<import("./types").Announcement[]>(
      `/events/${eventId}/announcements`,
      {},
      token,
    ),

  createAnnouncement: (
    token: string,
    eventId: string,
    data: { body: string; type?: string; link_url?: string },
  ) =>
    request<import("./types").Announcement>(
      `/events/${eventId}/announcements`,
      { method: "POST", body: JSON.stringify(data) },
      token,
    ),

  pinAnnouncement: (token: string, eventId: string, announcementId: string) =>
    request<import("./types").Announcement>(
      `/events/${eventId}/announcements/${announcementId}/pin`,
      { method: "POST" },
      token,
    ),

  deleteAnnouncement: (token: string, eventId: string, announcementId: string) =>
    request<void>(
      `/events/${eventId}/announcements/${announcementId}`,
      { method: "DELETE" },
      token,
    ),

  listResources: (token: string, eventId: string) =>
    request<import("./types").Resource[]>(`/events/${eventId}/resources`, {}, token),

  createResource: (
    token: string,
    eventId: string,
    data: { title: string; url?: string; kind?: string },
  ) =>
    request<import("./types").Resource>(
      `/events/${eventId}/resources`,
      { method: "POST", body: JSON.stringify(data) },
      token,
    ),

  listAgenda: (token: string, eventId: string) =>
    request<import("./types").AgendaItem[]>(`/events/${eventId}/agenda`, {}, token),

  createAgenda: (
    token: string,
    eventId: string,
    data: { title: string; speaker?: string },
  ) =>
    request<import("./types").AgendaItem>(
      `/events/${eventId}/agenda`,
      { method: "POST", body: JSON.stringify(data) },
      token,
    ),

  join: (code: string) =>
    request<{
      event_id: string;
      session_token: string;
      event: import("./types").PublicEvent;
      ws_url: string;
    }>("/join", { method: "POST", body: JSON.stringify({ code }) }),

  publicEvent: (eventId: string) =>
    request<import("./types").PublicEvent>(`/events/${eventId}/public`),
};

export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("pulseroom_token");
}

export function setToken(token: string) {
  localStorage.setItem("pulseroom_token", token);
}

export function clearToken() {
  localStorage.removeItem("pulseroom_token");
}

export function getSession(eventId: string): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(`pulseroom_session_${eventId}`);
}

export function setSession(eventId: string, token: string) {
  localStorage.setItem(`pulseroom_session_${eventId}`, token);
}
