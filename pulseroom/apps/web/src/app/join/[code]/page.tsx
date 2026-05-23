"use client";

import { useParams, useRouter } from "next/navigation";
import { useEffect } from "react";
import { api, setSession } from "@/lib/api";

export default function JoinCodePage() {
  const { code } = useParams<{ code: string }>();
  const router = useRouter();

  useEffect(() => {
    if (!code) return;
    api
      .join(code.toUpperCase())
      .then((res) => {
        setSession(res.event_id, res.session_token);
        router.replace(`/e/${res.event_id}`);
      })
      .catch(() => router.replace("/join"));
  }, [code, router]);

  return (
    <div className="flex min-h-screen items-center justify-center text-muted">
      Joining event…
    </div>
  );
}
