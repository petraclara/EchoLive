/** Converts http(s) API base URL to ws(s) for WebSocket connections. */
export function toWebSocketBase(url: string): string {
  return url.replace(/^https:\/\//, "wss://").replace(/^http:\/\//, "ws://").replace(/\/$/, "");
}

export function wsEventUrl(eventId: string, token: string): string {
  const api = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
  const base = process.env.NEXT_PUBLIC_WS_URL
    ? toWebSocketBase(process.env.NEXT_PUBLIC_WS_URL)
    : toWebSocketBase(api);
  return `${base}/ws/events/${eventId}?token=${token}`;
}
