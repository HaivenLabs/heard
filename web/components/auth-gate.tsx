"use client";

import { usePathname, useRouter } from "next/navigation";
import type { Route } from "next";
import { ReactNode, useEffect, useState } from "react";
import { getStoredSession, Session } from "../lib/api";

export function AuthGate({ children }: { children: (session: Session) => ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();
  const [session, setSession] = useState<Session | null | undefined>(undefined);

  useEffect(() => {
    const stored = getStoredSession();
    setSession(stored);
    if (!stored) {
      const requested = `${pathname}${window.location.search}`;
      router.replace(`/login?next=${encodeURIComponent(requested)}` as Route);
    }
  }, [pathname, router]);

  if (!session) {
    return (
      <main className="grid min-h-screen place-items-center bg-[#f5efe6] text-ink">
        <div className="flex items-center gap-3 font-body text-sm text-ink/60">
          <span className="h-2.5 w-2.5 animate-pulse rounded-full bg-clay" />
          Checking your heard session
        </div>
      </main>
    );
  }

  return <>{children(session)}</>;
}
