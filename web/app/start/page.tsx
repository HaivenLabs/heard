"use client";

import Link from "next/link";
import type { Route } from "next";
import { useEffect, useState } from "react";
import { availableIdentityProviders, identityAuthStartUrl, safeReturnTo } from "../../lib/api";
import { PublicHeader } from "../../components/public-header";

export default function StartPage() {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [restaurantName, setRestaurantName] = useState("");
  const [locationName, setLocationName] = useState("");
  const [googleAvailable, setGoogleAvailable] = useState<boolean | null>(null);

  useEffect(() => { availableIdentityProviders().then(providers => setGoogleAvailable(providers.includes("google"))); }, []);

  function continueWith(method: "email" | "google") {
    setError("");
    if (!restaurantName.trim() || !locationName.trim()) {
      setError("Restaurant and first location are required");
      return;
    }
    setBusy(true);
    const source = new URLSearchParams(window.location.search).get("source");
    const attribution = source === "guest_demo" ? "guest_demo" : source === "direct" ? "direct" : "homepage";
    window.sessionStorage.setItem("heard-onboarding-source", attribution);
    const returnTo = safeReturnTo(`/onboarding?source=${encodeURIComponent(attribution)}`, "/onboarding");
    window.location.assign(identityAuthStartUrl(returnTo, { intent: "register", provider: method === "google" ? "google" : undefined, organizationName: restaurantName, locationName, source: attribution }));
  }

  return (
    <main className="relative min-h-screen overflow-hidden bg-[#f7f1e6] text-ink">
      <div className="pointer-events-none absolute inset-0 opacity-40 [background-image:radial-gradient(circle_at_1px_1px,rgba(23,37,29,0.12)_1px,transparent_0)] [background-size:28px_28px]" />
      <div className="relative"><PublicHeader /></div>
      <div className="relative mx-auto grid min-h-[calc(100vh-7rem)] max-w-7xl items-center gap-14 px-6 pb-16 pt-8 lg:grid-cols-[1fr_0.82fr]">
        <section>
          <p className="font-body text-xs font-bold uppercase tracking-[0.28em] text-clay">Start with one location</p>
          <h1 className="mt-5 max-w-3xl font-display text-5xl leading-[1.02] tracking-[-0.055em] sm:text-7xl">Your first guest feedback campaign is minutes away.</h1>
          <p className="mt-6 max-w-xl font-body text-lg leading-8 text-ink/60">Create your account, add your restaurant and first location, then publish a ready-to-share feedback link. No call or approval required.</p>
          <Link className="mt-8 inline-flex font-body text-sm font-semibold text-ink/55 underline decoration-ink/20 underline-offset-4 hover:text-clay" href={"/walkthrough" as Route}>Unsure? Request a walkthrough instead.</Link>
        </section>

        <section className="rounded-[2rem] border border-ink/10 bg-[#fffdf8] p-7 shadow-[0_32px_90px_rgba(23,37,29,0.16)] sm:p-10">
          <p className="font-body text-xs font-bold uppercase tracking-[0.22em] text-olive">Create your account</p>
          <h2 className="mt-3 font-display text-3xl tracking-[-0.04em]">Let&apos;s get your restaurant live.</h2>
          <p className="mt-3 font-body text-sm leading-6 text-ink/55">{googleAvailable === true ? "These workspace details apply whether you continue with Google or with email and password." : "Add your workspace details, then create your account securely with email and password."}</p>
          <div className="mt-8 grid gap-4 sm:grid-cols-2">
            <label className="block"><span className="mb-2 block font-body text-sm font-semibold">Restaurant name</span><input className="h-14 w-full rounded-2xl border border-ink/15 bg-white px-4 font-body outline-none transition placeholder:text-ink/30 focus:border-clay focus:ring-4 focus:ring-clay/10" name="restaurant_name" onChange={(event) => setRestaurantName(event.target.value)} placeholder="Cedar & Salt" required value={restaurantName} /></label>
            <label className="block"><span className="mb-2 block font-body text-sm font-semibold">First location</span><input className="h-14 w-full rounded-2xl border border-ink/15 bg-white px-4 font-body outline-none transition placeholder:text-ink/30 focus:border-clay focus:ring-4 focus:ring-clay/10" name="location_name" onChange={(event) => setLocationName(event.target.value)} placeholder="Downtown" required value={locationName} /></label>
          </div>
          {googleAvailable === true ? (
            <>
              <button className="mt-8 flex h-14 w-full items-center justify-center gap-3 rounded-full border border-ink/15 bg-white px-6 font-display text-sm font-semibold text-ink transition hover:border-clay/50 hover:bg-[#fffdf8] disabled:cursor-not-allowed disabled:opacity-55" disabled={busy} onClick={() => continueWith("google")} type="button">
                <span aria-hidden="true" className="font-body text-lg font-bold text-[#4285f4]">G</span>
                Sign up with Google
              </button>
              <div className="my-6 flex items-center gap-3 text-xs uppercase tracking-[0.18em] text-ink/35"><span className="h-px flex-1 bg-ink/10" />or<span className="h-px flex-1 bg-ink/10" /></div>
            </>
          ) : <div className="mt-8" />}
          {error ? <p aria-live="assertive" className="mb-5 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 font-body text-sm text-red-700">{error}. Your progress is safe; please try again.</p> : null}
          <button className="flex h-14 w-full items-center justify-center rounded-full bg-clay px-6 font-display text-sm font-semibold tracking-[0.04em] text-white transition hover:bg-[#b95635] disabled:cursor-wait disabled:opacity-65" disabled={busy} onClick={() => continueWith("email")} type="button">Continue with email and password</button>
          <p className="mt-6 text-center font-body text-sm text-ink/48">Already have an account? <Link className="font-semibold text-clay underline decoration-clay/35 underline-offset-4" href="/login">Sign in</Link></p>
        </section>
      </div>
    </main>
  );
}
