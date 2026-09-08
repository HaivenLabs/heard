"use client";

import Link from "next/link";
import type { Route } from "next";
import { FormEvent, useEffect, useState } from "react";
import { availableIdentityProviders, identityAuthStartUrl, loginHeardAccount, safeReturnTo } from "../../lib/api";
import { useRouter } from "next/navigation";
import { PublicHeader } from "../../components/public-header";

export default function LoginPage() {
  const router = useRouter();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [googleAvailable, setGoogleAvailable] = useState<boolean | null>(null);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  useEffect(() => {
    availableIdentityProviders().then(providers => setGoogleAvailable(providers.includes("google")));
    const notice = new URLSearchParams(window.location.search).get("auth_notice");
    if (notice === "cancelled") setError("Google sign-in was canceled. You can try again whenever you're ready.");
    if (notice === "provider_error") setError("Google sign-in could not be completed. Please try again.");
    if (notice === "session_expired") setError("Your sign-in window expired. Please try again.");
    if (notice === "service_unavailable") setError("Sign-in is temporarily unavailable. Please try again in a moment.");
  }, []);

  async function continueWithEmail(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBusy(true);
    setError("");
    const next = safeReturnTo(new URLSearchParams(window.location.search).get("next"));
    try {
      await loginHeardAccount({ email, password });
      router.replace(next as Route);
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : "We could not sign you in. Please try again.");
      setBusy(false);
    }
  }

  function continueWithGoogle() {
    if (googleAvailable === false) {
      setError("Google sign-in is temporarily unavailable. Use email and password instead.");
      return;
    }
    setBusy(true);
    setError("");
    const next = safeReturnTo(new URLSearchParams(window.location.search).get("next"));
    window.location.assign(identityAuthStartUrl(next, { intent: "login", provider: "google" }));
  }

  return (
    <main className="relative min-h-screen overflow-hidden bg-parchment text-ink">
      <div className="brand-dot-field pointer-events-none absolute inset-0 opacity-40" />
      <div className="relative"><PublicHeader /></div>
      <div className="relative mx-auto grid min-h-[calc(100vh-7rem)] max-w-7xl items-center gap-14 px-6 pb-16 pt-8 lg:grid-cols-[1.05fr_0.8fr]">
        <section>
          <p className="font-body text-xs font-bold uppercase tracking-[0.28em] text-clay">Your guest recovery workspace</p>
          <h1 className="mt-5 max-w-2xl font-display text-5xl leading-[1.03] tracking-[-0.055em] sm:text-7xl">
            Pick up where your guests left off.
          </h1>
          <p className="mt-7 max-w-xl font-body text-lg leading-8 text-ink/60">
            Review new feedback, follow up with guests, and give every location a clear next move.
          </p>
        </section>

        <section className="rounded-[2rem] border border-ink/10 bg-surface p-7 shadow-[0_32px_90px_rgba(9,40,21,0.16)] sm:p-10">
          <div className="flex items-center justify-between gap-4">
            <div>
              <p className="font-body text-xs tracking-[0.16em] text-clay">Customer sign in</p>
              <h2 className="mt-3 font-display text-3xl tracking-[-0.04em]">Welcome back.</h2>
            </div>
          </div>
          <p className="mt-4 font-body text-sm leading-6 text-ink/58">Choose Google or sign in securely with your email and password.</p>
          <button className="mt-8 flex h-14 w-full items-center justify-center gap-3 rounded-full border border-ink/15 bg-white px-6 font-display text-sm font-semibold text-ink transition hover:border-teal/50 hover:bg-surface disabled:cursor-not-allowed disabled:opacity-55" disabled={busy} onClick={continueWithGoogle} type="button"><span aria-hidden="true" className="font-body text-lg font-bold text-primary">G</span>Continue with Google</button>
          <div className="my-6 flex items-center gap-3 text-xs uppercase tracking-[0.18em] text-ink/35"><span className="h-px flex-1 bg-ink/10" />Or sign in with email and password<span className="h-px flex-1 bg-ink/10" /></div>
          {error ? <p className="rounded-2xl border border-red-200 bg-red-50 px-4 py-3 font-body text-sm text-red-700">{error}</p> : null}
          <form className="mt-5" onSubmit={continueWithEmail}>
            <label className="block"><span className="mb-2 block font-body text-sm font-semibold">Work email</span><input autoComplete="email" className="h-14 w-full rounded-2xl border border-ink/15 bg-white px-4 font-body outline-none transition placeholder:text-ink/30 focus:border-clay focus:ring-4 focus:ring-clay/10" name="email" onChange={(event) => setEmail(event.target.value)} placeholder="you@restaurant.com" required type="email" value={email} /></label>
            <label className="mt-4 block"><span className="mb-2 block font-body text-sm font-semibold">Password</span><input autoComplete="current-password" className="h-14 w-full rounded-2xl border border-ink/15 bg-white px-4 font-body outline-none transition placeholder:text-ink/30 focus:border-clay focus:ring-4 focus:ring-clay/10" name="password" onChange={(event) => setPassword(event.target.value)} required type="password" value={password} /></label>
            <button className="mt-6 flex h-14 w-full items-center justify-center rounded-full bg-primary px-6 font-display text-sm font-semibold tracking-[0.04em] text-white transition hover:bg-spruce disabled:cursor-wait disabled:opacity-65" disabled={busy} type="submit">{busy ? "Signing in..." : "Sign in"}</button>
          </form>
          <p className="mt-6 text-center font-body text-sm leading-6 text-ink/48">New to heard? <Link className="font-semibold text-clay underline decoration-clay/35 underline-offset-4" href={"/start?source=direct" as Route}>Create your account.</Link></p>
        </section>
      </div>
    </main>
  );
}
