"use client";

import Link from "next/link";
import type { Route } from "next";
import { FormEvent, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { createLocalSession, getStoredSession, identityAuthStartUrl, isProductionAuth, safeReturnTo } from "../../lib/api";

export default function LoginPage() {
  const router = useRouter();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (getStoredSession()) {
      router.replace("/admin");
    }
  }, [router]);

  async function signIn(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBusy(true);
    setError("");
    const form = new FormData(event.currentTarget);
    const email = String(form.get("email") ?? "").trim();
    const next = safeReturnTo(new URLSearchParams(window.location.search).get("next"));
    if (isProductionAuth()) {
      window.location.assign(identityAuthStartUrl(next, { intent: "login", email }));
      return;
    }
    try {
      await createLocalSession(email);
      router.replace(next as Route);
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : "Could not start your session");
    } finally {
      setBusy(false);
    }
  }

  return (
    <main className="relative min-h-screen overflow-hidden bg-[#17251d] text-parchment">
      <div className="pointer-events-none absolute inset-0 opacity-50 [background-image:radial-gradient(circle_at_1px_1px,rgba(247,241,227,0.14)_1px,transparent_0)] [background-size:28px_28px]" />
      <div className="pointer-events-none absolute -right-40 top-[-12rem] h-[36rem] w-[36rem] rounded-full bg-clay/35 blur-3xl" />
      <div className="relative mx-auto grid min-h-screen max-w-6xl items-center gap-14 px-6 py-12 lg:grid-cols-[1.05fr_0.8fr]">
        <section>
          <Link className="font-display text-2xl font-semibold tracking-[-0.04em]" href="/">heard<span className="text-clay">.</span></Link>
          <p className="mt-20 font-body text-xs uppercase tracking-[0.32em] text-parchment/50">Your guest recovery workspace</p>
          <h1 className="mt-5 max-w-2xl font-display text-5xl leading-[1.03] tracking-[-0.055em] sm:text-7xl">
            Pick up where your guests left off.
          </h1>
          <p className="mt-7 max-w-xl font-body text-lg leading-8 text-parchment/68">
            Review new feedback, follow up with guests, and give every location a clear next move.
          </p>
        </section>

        <section className="rounded-[2rem] border border-parchment/15 bg-[#fffaf0] p-7 text-ink shadow-[0_36px_100px_rgba(0,0,0,0.3)] sm:p-10">
          <div className="flex items-center justify-between gap-4">
            <div>
              <p className="font-body text-xs tracking-[0.16em] text-clay">Customer sign in</p>
              <h2 className="mt-3 font-display text-3xl tracking-[-0.04em]">Welcome back.</h2>
            </div>
          </div>
          <p className="mt-4 font-body text-sm leading-6 text-ink/58">
            Use the work email connected to your heard account.
          </p>

          <form className="mt-8 space-y-5" onSubmit={signIn}>
            <label className="block">
              <span className="mb-2 block font-body text-sm font-semibold">Work email</span>
              <input autoComplete="email" autoFocus className="h-14 w-full rounded-2xl border border-ink/15 bg-white px-4 font-body outline-none transition placeholder:text-ink/30 focus:border-clay focus:ring-4 focus:ring-clay/10" name="email" placeholder="you@restaurant.com" required type="email" />
            </label>
            {error ? <p className="rounded-2xl border border-red-200 bg-red-50 px-4 py-3 font-body text-sm text-red-700">{error}</p> : null}
            <button className="flex h-14 w-full items-center justify-center rounded-full bg-clay px-6 font-display text-sm font-semibold tracking-[0.08em] text-white transition hover:bg-[#b95635] disabled:cursor-wait disabled:opacity-65" disabled={busy} type="submit">
              {busy ? "Signing in..." : "Continue with email"}
            </button>
          </form>
          <p className="mt-6 text-center font-body text-sm leading-6 text-ink/48">New to heard? <Link className="font-semibold text-clay underline decoration-clay/35 underline-offset-4" href={"/start?source=direct" as Route}>Create your account.</Link></p>
        </section>
      </div>
    </main>
  );
}
