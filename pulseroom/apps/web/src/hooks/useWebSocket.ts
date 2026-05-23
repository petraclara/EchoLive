"use client";

import { useEffect, useRef, useState } from "react";
import type { WSMessage } from "@/lib/types";

export function useWebSocket(url: string | null, onMessage: (msg: WSMessage) => void) {
  const [connected, setConnected] = useState(false);
  const onMessageRef = useRef(onMessage);
  onMessageRef.current = onMessage;

  useEffect(() => {
    if (!url) return;
    const endpoint = url;

    let ws: WebSocket | null = null;
    let closed = false;
    let retry = 1000;
    let timer: ReturnType<typeof setTimeout>;

    const heartbeat = setInterval(() => {
      if (ws?.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: "presence.heartbeat" }));
      }
    }, 30000);

    function connect() {
      ws = new WebSocket(endpoint);
      ws.onopen = () => {
        setConnected(true);
        retry = 1000;
      };
      ws.onmessage = (ev) => {
        try {
          const msg = JSON.parse(ev.data) as WSMessage;
          onMessageRef.current(msg);
        } catch {
          /* ignore */
        }
      };
      ws.onclose = () => {
        setConnected(false);
        if (!closed) {
          timer = setTimeout(connect, retry);
          retry = Math.min(retry * 2, 30000);
        }
      };
      ws.onerror = () => ws?.close();
    }

    connect();

    return () => {
      closed = true;
      clearTimeout(timer);
      clearInterval(heartbeat);
      ws?.close();
    };
  }, [url]);

  return { connected };
}
