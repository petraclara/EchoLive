import Link from "next/link";
import { ThemeToggle } from "@/components/theme-toggle";

export default function HomePage() {
  return (
    <div className="min-h-screen">
      <header className="mx-auto flex max-w-6xl items-center justify-between px-6 py-6">
        <span className="text-xl font-semibold tracking-tight">PulseRoom</span>
        <div className="flex items-center gap-3">
          <ThemeToggle />
          <Link
            href="/login"
            className="rounded-lg border border-border px-4 py-2 text-sm hover:bg-surface"
          >
            Sign in
          </Link>
          <Link
            href="/register"
            className="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-white hover:bg-primary-hover"
          >
            Get started
          </Link>
        </div>
      </header>

      <main className="mx-auto max-w-6xl px-6 pb-24 pt-16">
        <section className="text-center">
          <p className="mb-4 inline-block rounded-full border border-border bg-surface px-4 py-1 text-sm text-muted">
            Real-time event companion
          </p>
          <h1 className="mx-auto max-w-3xl text-4xl font-bold tracking-tight sm:text-6xl">
            Never miss what&apos;s on stage
          </h1>
          <p className="mx-auto mt-6 max-w-2xl text-lg text-muted">
            Share announcements, links, and schedules instantly to every attendee&apos;s
            phone — no app install, no refresh.
          </p>
          <div className="mt-10 flex flex-wrap justify-center gap-4">
            <Link
              href="/register"
              className="rounded-xl bg-primary px-8 py-3 font-medium text-white hover:bg-primary-hover"
            >
              Create an event
            </Link>
            <Link
              href="/join"
              className="rounded-xl border border-border bg-surface px-8 py-3 font-medium hover:opacity-90"
            >
              Join an event
            </Link>
          </div>
        </section>

        <section className="mt-24 grid gap-8 sm:grid-cols-3">
          {[
            {
              step: "1",
              title: "Create",
              desc: "Set up your event and get a shareable link and join code.",
            },
            {
              step: "2",
              title: "Share",
              desc: "Display the QR code at the venue. Attendees join in seconds.",
            },
            {
              step: "3",
              title: "Broadcast",
              desc: "Push live updates that appear instantly on every phone.",
            },
          ].map((item) => (
            <div
              key={item.step}
              className="rounded-2xl border border-border bg-surface p-6"
            >
              <span className="text-sm font-medium text-primary">Step {item.step}</span>
              <h3 className="mt-2 text-lg font-semibold">{item.title}</h3>
              <p className="mt-2 text-sm text-muted">{item.desc}</p>
            </div>
          ))}
        </section>
      </main>

      <footer className="border-t border-border py-8 text-center text-sm text-muted">
        PulseRoom — built for conferences, workshops, churches, and live events.
      </footer>
    </div>
  );
}
