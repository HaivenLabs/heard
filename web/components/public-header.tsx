import Link from "next/link";
import type { Route } from "next";
import { HeardLogo } from "./heard-logo";

export function PublicHeader() {
  return (
    <header className="relative mx-auto flex max-w-7xl flex-wrap items-center justify-between gap-y-4 px-6 py-6">
      <Link aria-label="heard home" href="/"><HeardLogo /></Link>
      <Link className="font-body text-sm font-semibold text-ink/55 transition hover:text-ink sm:hidden" href="/login">Sign in</Link>
      <nav className="order-last grid w-full grid-cols-2 items-center gap-2 sm:order-none sm:flex sm:w-auto sm:gap-2" aria-label="Primary navigation">
        <Link className="hidden rounded-full px-3 py-2 font-body text-sm text-ink/55 transition hover:text-ink lg:block" href={"/#how-it-works" as Route}>How it works</Link>
        <Link className="hidden rounded-full px-3 py-2 font-body text-sm text-ink/55 transition hover:text-ink lg:block" href={"/walkthrough" as Route}>Walkthrough</Link>
        <Link className="hidden rounded-full px-3 py-2 font-body text-sm text-ink/55 transition hover:text-ink sm:block" href="/login">Sign in</Link>
        <Link className="rounded-full border border-ink/15 bg-white/60 px-3 py-2.5 text-center font-body text-xs font-semibold transition hover:-translate-y-0.5 hover:border-ink/35 hover:bg-white sm:px-4 sm:text-sm" href="/f/demo-heard">Try guest experience</Link>
        <Link className="rounded-full bg-primary px-3 py-2.5 text-center font-display text-xs font-semibold tracking-[0.03em] text-white shadow-[0_10px_24px_rgba(9,40,21,0.2)] transition hover:-translate-y-0.5 hover:bg-spruce sm:px-5 sm:text-sm" href={"/start?source=homepage" as Route}>Start using heard</Link>
      </nav>
    </header>
  );
}
