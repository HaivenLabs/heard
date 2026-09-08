"use client";

import Link from "next/link";
import type { Route } from "next";
import { FormEvent, useEffect, useState } from "react";
import { availableIdentityProviders, identityAuthStartUrl, registerHeardAccount, safeReturnTo } from "../../lib/api";
import { PublicHeader } from "../../components/public-header";

export default function StartPage() {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [googleAvailable, setGoogleAvailable] = useState<boolean | null>(null);
  const [verificationRequested, setVerificationRequested] = useState(false);

  useEffect(() => {
    availableIdentityProviders().then(providers => setGoogleAvailable(providers.includes("google")));
    const notice = new URLSearchParams(window.location.search).get("auth_notice");
    if (notice === "cancelled") setError("Google sign-up was canceled. No account was created.");
    if (notice === "account_not_found") setError("That Google account doesn't have a Heard account yet. Create one to continue.");
    if (notice === "provider_error") setError("Google sign-up could not be completed. Please try again.");
    if (notice === "session_expired") setError("Your sign-up window expired. Please try again.");
    if (notice === "service_unavailable") setError("Account setup is temporarily unavailable. Please try again in a moment.");
  }, []);

  function continueWithGoogle() {
    setError("");
    if (googleAvailable === false) {
      setError("Google sign-up is temporarily unavailable. Create an account with email and password instead.");
      return;
    }
    setBusy(true);
    const source = new URLSearchParams(window.location.search).get("source");
    const attribution = source === "guest_demo" ? "guest_demo" : source === "direct" ? "direct" : "homepage";
    window.sessionStorage.setItem("heard-onboarding-source", attribution);
    const returnTo = safeReturnTo(`/onboarding?source=${encodeURIComponent(attribution)}`, "/onboarding");
    window.location.assign(identityAuthStartUrl(returnTo, { intent: "register", provider: "google", source: attribution }));
  }

  async function createAccount(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBusy(true);
    try {
      const source = new URLSearchParams(window.location.search).get("source");
      await registerHeardAccount({
        email,
        password,
        source: source === "guest_demo" ? "guest_demo" : source === "direct" ? "direct" : "homepage"
      });
      setVerificationRequested(true);
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : "We could not create your account. Please try again.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <main className="relative min-h-screen overflow-hidden bg-parchment text-ink">
      <div className="brand-dot-field pointer-events-none absolute inset-0 opacity-40" />
      <div className="relative"><PublicHeader /></div>
      <div className="relative mx-auto grid min-h-[calc(100vh-7rem)] max-w-7xl items-center gap-14 px-6 pb-16 pt-8 lg:grid-cols-[1fr_0.82fr]">
        <section>
          <p className="font-body text-xs font-bold uppercase tracking-[0.28em] text-clay">Start with one location</p>
          <h1 className="mt-5 max-w-3xl font-display text-5xl leading-[1.02] tracking-[-0.055em] sm:text-7xl">Your first guest feedback campaign is minutes away.</h1>
          <p className="mt-6 max-w-xl font-body text-lg leading-8 text-ink/60">Create your account, add your restaurant and first location, then publish a ready-to-share feedback link. No call or approval required.</p>
          <Link className="mt-8 inline-flex font-body text-sm font-semibold text-ink/55 underline decoration-ink/20 underline-offset-4 hover:text-clay" href={"/walkthrough" as Route}>Unsure? Request a walkthrough instead.</Link>
        </section>

        <section className="rounded-[2rem] border border-ink/10 bg-surface p-7 shadow-[0_32px_90px_rgba(9,40,21,0.16)] sm:p-10">
          <p className="font-body text-xs font-bold uppercase tracking-[0.22em] text-olive">Create your account</p>
          <h2 className="mt-3 font-display text-3xl tracking-[-0.04em]">Let&apos;s get your restaurant live.</h2>
          <p className="mt-3 font-body text-sm leading-6 text-ink/55">Choose how you&apos;d like to sign up. We&apos;ll ask about your restaurant once your account is ready.</p>
          {verificationRequested ? (
            <div className="mt-8 rounded-3xl border border-olive/20 bg-olive/10 px-5 py-6">
              <p className="font-body text-xs font-bold uppercase tracking-[0.18em] text-olive">Check your inbox</p>
              <p className="mt-3 font-body text-sm leading-6 text-ink/70">We sent a verification link to <strong>{email}</strong>. Open it to continue setting up your restaurant.</p>
            </div>
          ) : <form className="mt-8" onSubmit={createAccount}>
            <button className="flex h-14 w-full items-center justify-center gap-3 rounded-full border border-ink/15 bg-white px-6 font-display text-sm font-semibold text-ink transition hover:border-teal/50 hover:bg-surface disabled:cursor-not-allowed disabled:opacity-55" disabled={busy} onClick={continueWithGoogle} type="button"><span aria-hidden="true" className="font-body text-lg font-bold text-primary">G</span>Sign up with Google</button>
            <div className="my-6 flex items-center gap-3 text-xs uppercase tracking-[0.18em] text-ink/35"><span className="h-px flex-1 bg-ink/10" />Or sign up with email and password<span className="h-px flex-1 bg-ink/10" /></div>
            <label className="block"><span className="mb-2 block font-body text-sm font-semibold">Work email</span><input autoComplete="email" className="h-14 w-full rounded-2xl border border-ink/15 bg-white px-4 font-body outline-none transition placeholder:text-ink/30 focus:border-clay focus:ring-4 focus:ring-clay/10" name="email" onChange={(event) => setEmail(event.target.value)} placeholder="you@restaurant.com" required type="email" value={email} /></label>
            <label className="mt-4 block"><span className="mb-2 block font-body text-sm font-semibold">Password</span><input autoComplete="new-password" className="h-14 w-full rounded-2xl border border-ink/15 bg-white px-4 font-body outline-none transition placeholder:text-ink/30 focus:border-clay focus:ring-4 focus:ring-clay/10" minLength={12} name="password" onChange={(event) => setPassword(event.target.value)} placeholder="At least 12 characters" required type="password" value={password} /><span className="mt-2 block font-body text-xs leading-5 text-ink/48">Use 12 or more characters. A passphrase is easiest to remember.</span></label>
            {error ? <p aria-live="assertive" className="mb-5 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 font-body text-sm text-red-700">{error}</p> : null}
            <button className="flex h-14 w-full items-center justify-center rounded-full bg-primary px-6 font-display text-sm font-semibold tracking-[0.04em] text-white transition hover:bg-spruce disabled:cursor-wait disabled:opacity-65" disabled={busy} type="submit">{busy ? "Creating your account..." : "Create your account"}</button>
          </form>}
          <p className="mt-6 text-center font-body text-sm text-ink/48">Already have an account? <Link className="font-semibold text-clay underline decoration-clay/35 underline-offset-4" href="/login">Sign in</Link></p>
        </section>
      </div>
    </main>
  );
}
