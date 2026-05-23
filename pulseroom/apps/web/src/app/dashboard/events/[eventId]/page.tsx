"use client";

import Link from "next/link";
import { QRCodeSVG } from "qrcode.react";
import { useParams, useRouter } from "next/navigation";
import { FormEvent, useCallback, useEffect, useState } from "react";
import { ThemeToggle } from "@/components/theme-toggle";
import { useWebSocket } from "@/hooks/useWebSocket";
import { api, getToken } from "@/lib/api";
import type { Announcement, Event, WSMessage } from "@/lib/types";
import { wsEventUrl } from "@/lib/ws-url";

export default function EventControlPage() {
  const { eventId } = useParams<{ eventId: string }>();
  const router = useRouter();
  const [event, setEvent] = useState<Event | null>(null);
  const [joinUrl, setJoinUrl] = useState("");
  const [count, setCount] = useState(0);
  const [announcements, setAnnouncements] = useState<Announcement[]>([]);
  const [body, setBody] = useState("");
  const [annType, setAnnType] = useState("info");
  const [resourceTitle, setResourceTitle] = useState("");
  const [resourceUrl, setResourceUrl] = useState("");

  const token = getToken();
  const wsUrl = token && eventId ? wsEventUrl(eventId, token) : null;

  const load = useCallback(async () => {
    if (!token) {
      router.replace("/login");
      return;
    }
    const data = await api.getEvent(token, eventId);
    setEvent(data.event);
    setJoinUrl(data.join_url);
    setCount(data.attendee_count);
    const anns = await api.listAnnouncements(token, eventId);
    setAnnouncements(anns);
  }, [eventId, router, token]);

  useEffect(() => {
    load().catch(() => router.replace("/login"));
  }, [load, router]);

  const onWS = useCallback((msg: WSMessage) => {
    if (msg.type === "presence.count") {
      const p = msg.payload as { count: number };
      setCount(p.count);
    }
    if (msg.type === "announcement.created") {
      setAnnouncements((prev) => [...prev, msg.payload as Announcement]);
    }
    if (msg.type === "announcement.pinned") {
      const p = msg.payload as Announcement;
      setAnnouncements((prev) =>
        prev.map((a) => ({ ...a, is_pinned: a.id === p.id })),
      );
    }
    if (msg.type === "announcement.deleted") {
      const p = msg.payload as { id: string };
      setAnnouncements((prev) => prev.filter((a) => a.id !== p.id));
    }
  }, []);

  const { connected } = useWebSocket(wsUrl, onWS);

  async function sendAnnouncement(e: FormEvent) {
    e.preventDefault();
    if (!token || !body.trim()) return;
    await api.createAnnouncement(token, eventId, { body: body.trim(), type: annType });
    setBody("");
    await load();
  }

  async function setStatus(status: string) {
    if (!token) return;
    const updated = await api.updateEvent(token, eventId, { status });
    setEvent(updated);
  }

  async function pin(id: string) {
    if (!token) return;
    await api.pinAnnouncement(token, eventId, id);
    await load();
  }

  async function addResource(e: FormEvent) {
    e.preventDefault();
    if (!token || !resourceTitle.trim()) return;
    await api.createResource(token, eventId, {
      title: resourceTitle.trim(),
      url: resourceUrl.trim() || undefined,
    });
    setResourceTitle("");
    setResourceUrl("");
  }

  if (!event) {
    return (
      <div className="flex min-h-screen items-center justify-center text-muted">
        Loading…
      </div>
    );
  }

  return (
    <div className="min-h-screen pb-20">
      <header className="sticky top-0 z-10 border-b border-border bg-background/90 backdrop-blur">
        <div className="mx-auto flex max-w-5xl items-center justify-between px-4 py-3">
          <Link href="/dashboard" className="text-sm text-muted hover:underline">
            ← Events
          </Link>
          <div className="flex items-center gap-2">
            <span
              className={`h-2 w-2 rounded-full ${connected ? "bg-green-500" : "bg-amber-500"}`}
            />
            <span className="text-sm text-muted">{count} attendees</span>
            <ThemeToggle />
          </div>
        </div>
      </header>

      <main className="mx-auto grid max-w-5xl gap-6 px-4 py-6 lg:grid-cols-3">
        <div className="space-y-6 lg:col-span-2">
          <div className="flex flex-wrap items-center gap-3">
            <h1 className="text-2xl font-bold">{event.title}</h1>
            <span className="rounded-full border border-border px-3 py-0.5 text-sm">
              {event.join_code}
            </span>
          </div>

          <div className="flex flex-wrap gap-2">
            {(["draft", "live", "ended"] as const).map((s) => (
              <button
                key={s}
                type="button"
                onClick={() => setStatus(s)}
                className={`rounded-lg px-4 py-2 text-sm capitalize ${
                  event.status === s
                    ? "bg-primary text-white"
                    : "border border-border bg-surface"
                }`}
              >
                {s}
              </button>
            ))}
          </div>

          <form onSubmit={sendAnnouncement} className="rounded-2xl border border-border bg-surface p-4">
            <label className="text-sm font-medium">Live announcement</label>
            <textarea
              value={body}
              onChange={(e) => setBody(e.target.value)}
              rows={3}
              placeholder="Share a link, room change, or key info…"
              className="mt-2 w-full resize-none rounded-lg border border-border bg-background px-4 py-3"
            />
            <div className="mt-3 flex flex-wrap items-center gap-2">
              {["info", "alert", "link"].map((t) => (
                <button
                  key={t}
                  type="button"
                  onClick={() => setAnnType(t)}
                  className={`rounded-lg px-3 py-1 text-sm capitalize ${
                    annType === t ? "bg-primary text-white" : "border border-border"
                  }`}
                >
                  {t}
                </button>
              ))}
              <button
                type="submit"
                className="ml-auto rounded-lg bg-primary px-6 py-2 text-sm font-medium text-white hover:bg-primary-hover"
              >
                Send live
              </button>
            </div>
          </form>

          <div className="space-y-3">
            <h2 className="font-semibold">Feed preview</h2>
            {[...announcements].reverse().map((a) => (
              <div
                key={a.id}
                className={`rounded-xl border p-4 ${
                  a.is_pinned ? "border-amber-500/50 bg-amber-500/10" : "border-border bg-surface"
                }`}
              >
                <p className="whitespace-pre-wrap">{a.body}</p>
                <div className="mt-2 flex gap-2">
                  <button
                    type="button"
                    onClick={() => pin(a.id)}
                    className="text-xs text-primary hover:underline"
                  >
                    {a.is_pinned ? "Pinned" : "Pin"}
                  </button>
                  <span className="text-xs text-muted">
                    {new Date(a.created_at).toLocaleTimeString()}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>

        <aside className="space-y-6">
          <div className="rounded-2xl border border-border bg-surface p-5 text-center">
            <p className="text-sm font-medium">Join QR code</p>
            <div className="mt-4 flex justify-center rounded-xl bg-white p-4">
              <QRCodeSVG value={joinUrl} size={180} />
            </div>
            <p className="mt-4 break-all text-xs text-muted">{joinUrl}</p>
            <Link
              href={`/join/${event.join_code}`}
              target="_blank"
              className="mt-3 inline-block text-sm text-primary hover:underline"
            >
              Open attendee view
            </Link>
          </div>

          <form onSubmit={addResource} className="rounded-2xl border border-border bg-surface p-4">
            <p className="font-medium">Add resource</p>
            <input
              value={resourceTitle}
              onChange={(e) => setResourceTitle(e.target.value)}
              placeholder="Title"
              className="mt-2 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
            />
            <input
              value={resourceUrl}
              onChange={(e) => setResourceUrl(e.target.value)}
              placeholder="URL (optional)"
              className="mt-2 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm"
            />
            <button
              type="submit"
              className="mt-3 w-full rounded-lg border border-border py-2 text-sm hover:bg-background"
            >
              Add link
            </button>
          </form>
        </aside>
      </main>
    </div>
  );
}
