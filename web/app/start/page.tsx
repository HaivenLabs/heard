"use client";

import Link from "next/link";
import type { Route } from "next";
import { FormEvent, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { createLocalRegistration, getStoredSession, isProductionAuth, passageHostedUrl, safeReturnTo } from "../../lib/api";

export default function StartPage() {
  const router = useRouter();
  const productionAuth = isProductionAuth();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (getStoredSession()) {
      router.replace("/onboarding" as Route);
    }
  }, [router]);

  async function start(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBusy(true);
    setError("");
    const form = new FormData(event.currentTarget);
    try {
      const source = new URLSearchParams(window.location.search).get("source");
      const attribution = source === "guest_demo" ? "guest_demo" : source === "direct" ? "direct" : "homepage";
      window.sessionStorage.setItem("heard-onboarding-source", attribution);
      const returnTo = safeReturnTo(`/onboarding?source=${encodeURIComponent(attribution)}`, "/onboarding");
      if (isProductionAuth()) {
        try {
          window.location.assign(passageHostedUrl("signup", returnTo));
        } catch (caught) {
          setError(caught instanceof Error ? caught.message : "Account setup is temporarily unavailable");
          setBusy(false);
        }
        return;
      }
      await createLocalRegistration(String(form.get("email") ?? ""));
      router.replace("/onboarding" as Route);
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : "Account setup is temporarily unavailable");
    } finally {
      setBusy(false);
    }
  }

  return (
    <main className="relative min-h-screen overflow-hidden bg-[#f7f1e6] px-6 py-10 text-ink">
      <div className="pointer-events-none absolute inset-0 opacity-40 [background-image:radial-gradient(circle_at_1px_1px,rgba(23,37,29,0.12)_1px,transparent_0)] [background-size:28px_28px]" />
      <div className="relative mx-auto grid min-h-[calc(100vh-5rem)] max-w-6xl items-center gap-14 lg:grid-cols-[1fr_0.82fr]">
        <section>
          <Link className="font-display text-2xl font-semibold tracking-[-0.04em]" href="/">heard<span className="text-clay">.</span></Link>
          <p className="mt-20 font-body text-xs font-bold uppercase tracking-[0.28em] text-clay">Start with one location</p>
          <h1 className="mt-5 max-w-3xl font-display text-5xl leading-[1.02] tracking-[-0.055em] sm:text-7xl">Your first guest feedback campaign is minutes away.</h1>
          <p className="mt-6 max-w-xl font-body text-lg leading-8 text-ink/60">Create your account, add your restaurant and first location, then publish a ready-to-share feedback link. No call or approval required.</p>
          <Link className="mt-8 inline-flex font-body text-sm font-semibold text-ink/55 underline decoration-ink/20 underline-offset-4 hover:text-clay" href="/contact#walkthrough">Unsure? Request a walkthrough instead.</Link>
        </section>

        <section className="rounded-[2rem] border border-ink/10 bg-[#fffdf8] p-7 shadow-[0_32px_90px_rgba(23,37,29,0.16)] sm:p-10">
          <p className="font-body text-xs font-bold uppercase tracking-[0.22em] text-olive">Create your account</p>
          <h2 className="mt-3 font-display text-3xl tracking-[-0.04em]">Let&apos;s get your restaurant live.</h2>
          <p className="mt-3 font-body text-sm leading-6 text-ink/55">{productionAuth ? "Passage will securely create your account, then bring you straight back here." : "We only need your work email to begin. Restaurant details come next."}</p>
          <form className="mt-8 space-y-5" onSubmit={start}>
            {productionAuth ? <div className="rounded-2xl bg-[#f5efe6] px-4 py-4 font-body text-sm leading-6 text-ink/58">You&apos;ll finish account setup in Passage, then return to your restaurant workspace.</div> : <label className="block"><span className="mb-2 block font-body text-sm font-semibold">Work email</span><input autoComplete="email" autoFocus className="h-14 w-full rounded-2xl border border-ink/15 bg-white px-4 font-body outline-none transition placeholder:text-ink/30 focus:border-clay focus:ring-4 focus:ring-clay/10" name="email" placeholder="you@restaurant.com" required type="email" /></label>}
            {error ? <p aria-live="assertive" className="rounded-2xl border border-red-200 bg-red-50 px-4 py-3 font-body text-sm text-red-700">{error}. Your progress is safe; please try again.</p> : null}
            <button className="flex h-14 w-full items-center justify-center rounded-full bg-clay px-6 font-display text-sm font-semibold tracking-[0.07em] text-white transition hover:bg-[#b95635] disabled:cursor-wait disabled:opacity-65" disabled={busy} type="submit">{busy ? "Opening secure account setup..." : productionAuth ? "Create account with Passage" : "Start using heard"}</button>
          </form>
          <p className="mt-6 text-center font-body text-sm text-ink/48">Already have an account? <Link className="font-semibold text-clay underline decoration-clay/35 underline-offset-4" href="/login">Sign in</Link></p>
        </section>
      </div>
    </main>
  );
}
