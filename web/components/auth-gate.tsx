"use client";

import { usePathname, useRouter } from "next/navigation";
import type { Route } from "next";
import { ReactNode, useEffect, useState } from "react";
import { apiFetch, OnboardingState, resolveSession, Session } from "../lib/api";

export function AuthGate({ children, requireOnboardingComplete = false }: { children: (session: Session) => ReactNode; requireOnboardingComplete?: boolean }) {
  const pathname = usePathname();
  const router = useRouter();
  const [session, setSession] = useState<Session | null | undefined>(undefined);

  useEffect(() => {
    let active = true;
    void resolveSession().then(async (current) => {
      if (!active) return;
      if (!current) {
        const requested = `${pathname}${window.location.search}`;
        router.replace(`/login?next=${encodeURIComponent(requested)}` as Route);
        setSession(null);
        return;
      }
      if (requireOnboardingComplete) {
        try {
          const onboarding = await apiFetch<OnboardingState>("/api/v1/onboarding");
          if (!active) return;
          if (onboarding.status !== "complete") {
            router.replace("/onboarding" as Route);
            setSession(null);
            return;
          }
        } catch {
          if (!active) return;
          router.replace("/onboarding" as Route);
          setSession(null);
          return;
        }
      }
      setSession(current);
    });
    return () => { active = false; };
  }, [pathname, requireOnboardingComplete, router]);

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
