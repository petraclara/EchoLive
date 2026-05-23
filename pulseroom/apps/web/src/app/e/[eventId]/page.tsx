"use client";

import { useParams, useRouter } from "next/navigation";
import { useCallback, useEffect, useMemo, useState } from "react";
import { ThemeToggle } from "@/components/theme-toggle";
import { useWebSocket } from "@/hooks/useWebSocket";
import { api, getSession, setSession } from "@/lib/api";
import type { Announcement, PublicEvent, WSMessage } from "@/lib/types";
import { wsEventUrl } from "@/lib/ws-url";

type Tab = "updates" | "agenda" | "resources";

export default function AttendeeFeedPage() {
  const { eventId } = useParams<{ eventId: string }>();
  const router = useRouter();
  const [event, setEvent] = useState<PublicEvent | null>(null);
  const [tab, setTab] = useState<Tab>("updates");
  const [announcements, setAnnouncements] = useState<Announcement[]>([]);
  const [pinned, setPinned] = useState<Announcement | undefined>();

  const session = getSession(eventId);
  const wsUrl = useMemo(() => {
    if (!session || !eventId) return null;
    return wsEventUrl(eventId, session);
  }, [eventId, session]);

  const bootstrap = useCallback(async () => {
    try {
      if (!session) {
        router.replace("/join");
        return;
      }
      const pub = await api.publicEvent(eventId);
      setEvent(pub);
      setAnnouncements(pub.announcements || []);
      setPinned(pub.pinned);
    } catch {
      router.replace("/join");
    }
  }, [eventId, router, session]);

  useEffect(() => {
    bootstrap();
  }, [bootstrap]);

  const onWS = useCallback((msg: WSMessage) => {
    if (msg.type === "announcement.created") {
      const a = msg.payload as Announcement;
      setAnnouncements((prev) => [...prev, a]);
    }
    if (msg.type === "announcement.pinned") {
      setPinned(msg.payload as Announcement);
      const p = msg.payload as Announcement;
      setAnnouncements((prev) =>
        prev.map((a) => ({ ...a, is_pinned: a.id === p.id })),
      );
    }
    if (msg.type === "announcement.deleted") {
      const p = msg.payload as { id: string };
      setAnnouncements((prev) => prev.filter((a) => a.id !== p.id));
    }
    if (msg.type === "event.status") {
      bootstrap();
    }
  }, [bootstrap]);

  const { connected } = useWebSocket(wsUrl, onWS);

  async function rejoin() {
    const code = prompt("Enter event code:");
    if (!code) return;
    const res = await api.join(code);
    setSession(res.event_id, res.session_token);
    router.push(`/e/${res.event_id}`);
  }

  if (!event) {
    return (
      <div className="flex min-h-screen items-center justify-center text-muted">
        Loading event…
      </div>
    );
  }

  const feed = [...announcements].sort(
    (a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime(),
  );

  return (
    <div className="mx-auto flex min-h-screen max-w-lg flex-col">
      <header className="sticky top-0 z-10 border-b border-border bg-background/95 px-4 py-4 backdrop-blur">
        <div className="flex items-start justify-between gap-2">
          <div>
            <h1 className="text-lg font-bold leading-tight">{event.title}</h1>
            <p className="text-xs text-muted">
              {event.status === "live" ? (
                <span className="text-green-600 dark:text-green-400">● Live</span>
              ) : (
                event.status
              )}
              {connected ? " · connected" : " · reconnecting…"}
            </p>
          </div>
          <ThemeToggle />
        </div>

        {pinned && (
          <div className="mt-3 rounded-xl border border-amber-500/40 bg-amber-500/10 px-4 py-3">
            <p className="text-xs font-medium uppercase text-amber-600 dark:text-amber-400">
              Pinned
            </p>
            <p className="mt-1 whitespace-pre-wrap text-sm font-medium">{pinned.body}</p>
          </div>
        )}
      </header>

      <div className="flex border-b border-border">
        {(["updates", "agenda", "resources"] as Tab[]).map((t) => (
          <button
            key={t}
            type="button"
            onClick={() => setTab(t)}
            className={`flex-1 py-3 text-sm capitalize ${
              tab === t ? "border-b-2 border-primary font-medium text-primary" : "text-muted"
            }`}
          >
            {t}
          </button>
        ))}
      </div>

      <main className="flex-1 overflow-y-auto px-4 py-4 pb-24">
        {tab === "updates" && (
          <div className="space-y-3">
            {[...feed].reverse().map((a) => (
              <article
                key={a.id}
                className={`rounded-xl border p-4 ${
                  a.type === "alert"
                    ? "border-red-500/30 bg-red-500/10"
                    : "border-border bg-surface"
                }`}
              >
                <p className="whitespace-pre-wrap">{a.body}</p>
                {a.link_url && (
                  <a
                    href={a.link_url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="mt-2 inline-block text-sm text-primary underline"
                  >
                    Open link
                  </a>
                )}
                <time className="mt-2 block text-xs text-muted">
                  {new Date(a.created_at).toLocaleTimeString()}
                </time>
              </article>
            ))}
            {feed.length === 0 && (
              <p className="py-12 text-center text-muted">
                Waiting for live updates from the organizer…
              </p>
            )}
          </div>
        )}

        {tab === "agenda" && (
          <ul className="space-y-3">
            {event.agenda?.map((item) => (
              <li key={item.id} className="rounded-xl border border-border bg-surface p-4">
                <p className="font-medium">{item.title}</p>
                {item.speaker && (
                  <p className="text-sm text-muted">{item.speaker}</p>
                )}
              </li>
            ))}
            {(!event.agenda || event.agenda.length === 0) && (
              <p className="py-12 text-center text-muted">No agenda items yet.</p>
            )}
          </ul>
        )}

        {tab === "resources" && (
          <ul className="space-y-3">
            {event.resources?.map((r) => (
              <li key={r.id}>
                <a
                  href={r.url || "#"}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="block rounded-xl border border-border bg-surface p-4 hover:border-primary"
                >
                  <p className="font-medium">{r.title}</p>
                  {r.url && <p className="truncate text-xs text-muted">{r.url}</p>}
                </a>
              </li>
            ))}
            {(!event.resources || event.resources.length === 0) && (
              <p className="py-12 text-center text-muted">No resources shared yet.</p>
            )}
          </ul>
        )}
      </main>

      <footer className="fixed bottom-0 left-0 right-0 border-t border-border bg-background/95 px-4 py-3 text-center">
        <button type="button" onClick={rejoin} className="text-xs text-muted hover:underline">
          Join different event
        </button>
      </footer>
    </div>
  );
}
