"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { FormEvent, useState } from "react";
import { api, setSession } from "@/lib/api";

export default function JoinPage() {
  const router = useRouter();
  const [code, setCode] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      const res = await api.join(code.toUpperCase().trim());
      setSession(res.event_id, res.session_token);
      router.push(`/e/${res.event_id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not join event");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="mx-auto flex min-h-screen max-w-md flex-col justify-center px-6">
      <Link href="/" className="mb-8 text-sm text-muted hover:underline">
        ← Home
      </Link>
      <h1 className="text-2xl font-bold">Join an event</h1>
      <p className="mt-2 text-sm text-muted">
        Enter the 6-character code from the screen or organizer.
      </p>
      <form onSubmit={onSubmit} className="mt-8 space-y-4">
        <input
          value={code}
          onChange={(e) => setCode(e.target.value.toUpperCase())}
          maxLength={6}
          required
          placeholder="e.g. LIVE42"
          className="w-full rounded-xl border border-border bg-surface px-4 py-4 text-center text-2xl font-mono tracking-widest"
        />
        {error && <p className="text-center text-sm text-red-500">{error}</p>}
        <button
          type="submit"
          disabled={loading || code.length < 6}
          className="w-full rounded-xl bg-primary py-4 font-medium text-white hover:bg-primary-hover disabled:opacity-50"
        >
          {loading ? "Joining…" : "Join event"}
        </button>
      </form>
    </div>
  );
}
