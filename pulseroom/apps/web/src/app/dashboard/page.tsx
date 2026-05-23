"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, useEffect, useState } from "react";
import { ThemeToggle } from "@/components/theme-toggle";
import { api, clearToken, getToken } from "@/lib/api";
import type { Event } from "@/lib/types";

export default function DashboardPage() {
  const router = useRouter();
  const [events, setEvents] = useState<Event[]>([]);
  const [title, setTitle] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = getToken();
    if (!token) {
      router.replace("/login");
      return;
    }
    api
      .listEvents(token)
      .then(setEvents)
      .catch(() => router.replace("/login"))
      .finally(() => setLoading(false));
  }, [router]);

  async function createEvent(e: FormEvent) {
    e.preventDefault();
    const token = getToken();
    if (!token || !title.trim()) return;
    const { event } = await api.createEvent(token, title.trim());
    setEvents((prev) => [event, ...prev]);
    setTitle("");
    router.push(`/dashboard/events/${event.id}`);
  }

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center text-muted">
        Loading…
      </div>
    );
  }

  return (
    <div className="min-h-screen">
      <header className="border-b border-border">
        <div className="mx-auto flex max-w-4xl items-center justify-between px-6 py-4">
          <Link href="/" className="font-semibold">
            PulseRoom
          </Link>
          <div className="flex items-center gap-3">
            <ThemeToggle />
            <button
              type="button"
              onClick={() => {
                clearToken();
                router.push("/");
              }}
              className="text-sm text-muted hover:underline"
            >
              Sign out
            </button>
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-4xl px-6 py-10">
        <h1 className="text-2xl font-bold">Your events</h1>
        <form onSubmit={createEvent} className="mt-6 flex gap-2">
          <input
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="New event title"
            className="flex-1 rounded-lg border border-border bg-surface px-4 py-2"
          />
          <button
            type="submit"
            className="rounded-lg bg-primary px-5 py-2 text-white hover:bg-primary-hover"
          >
            Create
          </button>
        </form>

        <ul className="mt-8 space-y-3">
          {events.map((ev) => (
            <li key={ev.id}>
              <Link
                href={`/dashboard/events/${ev.id}`}
                className="flex items-center justify-between rounded-xl border border-border bg-surface px-5 py-4 hover:border-primary"
              >
                <div>
                  <p className="font-medium">{ev.title}</p>
                  <p className="text-sm text-muted">Code: {ev.join_code}</p>
                </div>
                <span
                  className={`rounded-full px-3 py-1 text-xs font-medium ${
                    ev.status === "live"
                      ? "bg-green-500/20 text-green-600 dark:text-green-400"
                      : "bg-slate-500/20 text-muted"
                  }`}
                >
                  {ev.status}
                </span>
              </Link>
            </li>
          ))}
          {events.length === 0 && (
            <p className="text-muted">No events yet. Create your first one above.</p>
          )}
        </ul>
      </main>
    </div>
  );
}
