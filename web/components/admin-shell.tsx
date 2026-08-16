"use client";

import Link from "next/link";
import type { Route } from "next";
import { usePathname, useRouter } from "next/navigation";
import { ReactNode } from "react";
import { clearStoredSession, Session } from "../lib/api";

const navigation: Array<{ href: Route; label: string }> = [
  { href: "/admin", label: "Overview" },
  { href: "/admin/campaigns", label: "Campaigns" },
  { href: "/admin/recovery", label: "Recovery" }
];

export function AdminShell({ children, session }: { children: ReactNode; session: Session }) {
  const pathname = usePathname();
  const router = useRouter();

  function signOut() {
    clearStoredSession();
    router.replace("/login");
  }

  return (
    <div className="min-h-screen bg-[#f5efe6] text-ink">
      <header className="sticky top-0 z-20 border-b border-ink/10 bg-[#fffdf8]/90 backdrop-blur-xl">
        <div className="mx-auto flex max-w-7xl items-center justify-between gap-5 px-5 py-4">
          <div className="flex items-center gap-8">
            <Link className="font-display text-xl font-semibold tracking-[-0.04em]" href="/admin">
              heard<span className="text-clay">.</span>
            </Link>
            <nav className="hidden items-center gap-1 sm:flex" aria-label="Management console">
              {navigation.map((item) => {
                const active = item.href === "/admin" ? pathname === item.href : pathname.startsWith(item.href);
                return (
                  <Link
                    className={`rounded-full px-4 py-2 font-body text-sm transition ${active ? "bg-ink text-parchment" : "text-ink/60 hover:bg-ink/5 hover:text-ink"}`}
                    href={item.href}
                    key={item.href}
                  >
                    {item.label}
                  </Link>
                );
              })}
            </nav>
          </div>
          <div className="flex items-center gap-3">
            <div className="hidden text-right md:block">
              <p className="font-body text-sm font-semibold">{session.identity.display_name}</p>
              <p className="font-body text-xs capitalize text-ink/45">{session.identity.role} · North Star Noodles</p>
            </div>
            <button className="rounded-full border border-ink/15 px-4 py-2 font-body text-sm text-ink/65 transition hover:border-clay hover:text-clay" onClick={signOut} type="button">
              Sign out
            </button>
          </div>
        </div>
        <nav className="flex gap-1 overflow-x-auto px-4 pb-3 sm:hidden" aria-label="Management console">
          {navigation.map((item) => {
            const active = item.href === "/admin" ? pathname === item.href : pathname.startsWith(item.href);
            return <Link className={`whitespace-nowrap rounded-full px-4 py-2 font-body text-sm ${active ? "bg-ink text-parchment" : "bg-ink/5 text-ink/60"}`} href={item.href} key={item.href}>{item.label}</Link>;
          })}
        </nav>
      </header>
      {children}
    </div>
  );
}
