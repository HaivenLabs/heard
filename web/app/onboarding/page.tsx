"use client";

import Link from "next/link";
import type { Route } from "next";
import { FormEvent, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { AuthGate } from "../../components/auth-gate";
import { apiFetch, FeedbackLink, OnboardingState, Session, SurveyCampaign } from "../../lib/api";
import { slugify } from "../../lib/slug";

export default function OnboardingPage() {
  return <AuthGate>{(session) => <Onboarding session={session} />}</AuthGate>;
}

function Onboarding({ session }: { session: Session }) {
  const router = useRouter();
  const [state, setState] = useState<OnboardingState | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [finishedHere, setFinishedHere] = useState(false);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    void apiFetch<OnboardingState>("/api/v1/onboarding")
      .then((next) => {
        if (next.status === "complete") {
          router.replace("/admin");
          return;
        }
        setState(next);
      })
      .catch((caught) => setError(caught instanceof Error ? caught.message : "Could not resume account setup"));
  }, [router]);

  async function activateWorkspace(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBusy(true);
    setError("");
    const form = new FormData(event.currentTarget);
    const keyName = "heard-onboarding-activation-key";
    const idempotencyKey = window.sessionStorage.getItem(keyName) ?? crypto.randomUUID();
    window.sessionStorage.setItem(keyName, idempotencyKey);
    const source = window.sessionStorage.getItem("heard-onboarding-source") ?? "direct";
    try {
      const activated = await apiFetch<OnboardingState>("/api/v1/onboarding/activations", {
        method: "POST",
        idempotencyKey,
        body: {
          restaurant_name: String(form.get("restaurant_name") ?? ""),
          location_name: String(form.get("location_name") ?? ""),
          timezone: Intl.DateTimeFormat().resolvedOptions().timeZone || "America/Los_Angeles",
          source
        }
      });
      setState(activated);
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : "Could not activate your restaurant workspace");
    } finally {
      setBusy(false);
    }
  }

  async function createFirstCampaign(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!state?.tenant || !state.location) return;
    setBusy(true);
    setError("");
    const form = new FormData(event.currentTarget);
    try {
      const campaign = await apiFetch<SurveyCampaign>("/api/v1/survey-campaigns", {
        method: "POST",
        tenantId: state.tenant.id,
        body: {
          tenant_id: state.tenant.id,
          location_id: state.location.id,
          name: `${state.location.name} guest feedback`,
          restaurant_name: state.tenant.name,
          headline: String(form.get("headline") ?? "How did we do?"),
          prompt: "Tap the face that matches your visit.",
          incentive_text: "",
          sms_keyword: "",
          sms_phone: "",
          google_review_url: "",
          yelp_review_url: ""
        }
      });
      setState({ ...state, campaign, next_step: "feedback_link" });
      await createCampaignLink(campaign, state);
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : "Could not create your first campaign");
    } finally {
      setBusy(false);
    }
  }

  async function createCampaignLink(campaign = state?.campaign, current = state) {
    if (!campaign || !current?.tenant || !current.location) return;
    setBusy(true);
    setError("");
    try {
      const feedbackLink = await apiFetch<FeedbackLink>("/api/v1/feedback-links", {
        method: "POST",
        tenantId: current.tenant.id,
        body: {
          tenant_id: current.tenant.id,
          location_id: current.location.id,
          campaign_id: campaign.id,
          name: `${campaign.name} link`,
          channel: "onboarding",
          slug: slugify(current.location.name)
        }
      });
      setState({ ...current, campaign, feedback_link: feedbackLink, status: "complete", next_step: "complete" });
      setFinishedHere(true);
    } catch (caught) {
      setState({ ...current, campaign, status: "in_progress", next_step: "feedback_link" });
      setError(caught instanceof Error ? caught.message : "Your campaign is safe, but its feedback link could not be created");
    } finally {
      setBusy(false);
    }
  }

  async function copyLink() {
    if (!state?.feedback_link) return;
    await navigator.clipboard.writeText(state.feedback_link.destination_url);
    setCopied(true);
  }

  const step = state?.next_step ?? "workspace";
  return (
    <main className="min-h-screen bg-[#f7f1e6] px-5 py-8 text-ink sm:py-12">
      <div className="mx-auto max-w-5xl">
        <header className="flex items-center justify-between gap-5">
          <Link className="font-display text-2xl font-semibold tracking-[-0.04em]" href="/">heard<span className="text-clay">.</span></Link>
          <p className="font-body text-sm text-ink/48">Signed in as {session.identity.display_name}</p>
        </header>
        <div className="mt-10 grid gap-8 lg:grid-cols-[0.6fr_1fr]">
          <aside>
            <p className="font-body text-xs font-bold uppercase tracking-[0.26em] text-clay">Restaurant setup</p>
            <h1 className="mt-4 font-display text-4xl leading-tight tracking-[-0.05em] sm:text-5xl">From account to live campaign, without the detour.</h1>
            <ol className="mt-8 space-y-4 font-body text-sm text-ink/58">
              <Progress done={step !== "workspace"} number="1" text="Restaurant and first location" />
              <Progress done={step === "feedback_link" || step === "complete"} number="2" text="First feedback campaign" />
              <Progress done={step === "complete"} number="3" text="Shareable guest link" />
            </ol>
            <Link className="mt-9 inline-flex font-body text-sm font-semibold text-ink/48 underline decoration-ink/20 underline-offset-4 hover:text-clay" href={"/walkthrough" as Route}>Want help? Request a walkthrough.</Link>
          </aside>

          <section className="rounded-[2rem] border border-ink/10 bg-[#fffdf8] p-6 shadow-soft sm:p-9">
            {!state && !error ? <div className="space-y-4"><div className="h-8 w-2/3 animate-pulse rounded bg-ink/5" /><div className="h-14 animate-pulse rounded-2xl bg-ink/5" /><div className="h-14 animate-pulse rounded-2xl bg-ink/5" /></div> : null}
            {error ? <p aria-live="assertive" className="mb-5 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 font-body text-sm text-red-700">{error} You can retry without creating duplicates.</p> : null}

            {state && step === "workspace" ? (
              <form onSubmit={activateWorkspace}>
                <p className="font-body text-xs font-bold uppercase tracking-[0.2em] text-olive">Step 1 of 3</p>
                <h2 className="mt-3 font-display text-3xl tracking-[-0.04em]">Tell us where guests visit.</h2>
                <p className="mt-3 font-body text-sm leading-6 text-ink/55">That&apos;s all we need to prepare your workspace. You can add branding, integrations, and more locations later.</p>
                <div className="mt-7 grid gap-4 sm:grid-cols-2">
                  <Field label="Restaurant name" name="restaurant_name" placeholder="Cedar & Salt" />
                  <Field label="First location" name="location_name" placeholder="Downtown" />
                </div>
                <button className="mt-7 h-14 w-full rounded-full bg-clay px-6 font-display text-sm font-semibold tracking-[0.06em] text-white transition hover:bg-[#b95635] disabled:opacity-60" disabled={busy} type="submit">{busy ? "Preparing your workspace..." : "Continue to your campaign"}</button>
              </form>
            ) : null}

            {state && step === "campaign" ? (
              <form onSubmit={createFirstCampaign}>
                <p className="font-body text-xs font-bold uppercase tracking-[0.2em] text-olive">Step 2 of 3</p>
                <h2 className="mt-3 font-display text-3xl tracking-[-0.04em]">Give your first campaign a headline.</h2>
                <p className="mt-3 font-body text-sm leading-6 text-ink/55">We&apos;ve filled in the rest with a short, proven guest experience. Everything stays editable later.</p>
                <div className="mt-7"><Field defaultValue="How did we do?" label="Guest-facing headline" name="headline" placeholder="How did we do?" /></div>
                <div className="mt-5 rounded-2xl bg-[#f5efe6] p-5 font-body text-sm leading-6 text-ink/58"><strong className="text-ink">{state.tenant?.name}</strong><br />{state.location?.name} · Five-face feedback survey<br /><span className="text-xs text-ink/48">Your heard handle: /f/{state.tenant?.slug ?? "your-restaurant"}/</span></div>
                <button className="mt-7 h-14 w-full rounded-full bg-clay px-6 font-display text-sm font-semibold tracking-[0.06em] text-white transition hover:bg-[#b95635] disabled:opacity-60" disabled={busy} type="submit">{busy ? "Creating your campaign..." : "Create feedback campaign"}</button>
              </form>
            ) : null}

            {state && step === "feedback_link" && state.campaign ? (
              <div>
                <p className="font-body text-xs font-bold uppercase tracking-[0.2em] text-olive">Your campaign is safe</p>
                <h2 className="mt-3 font-display text-3xl tracking-[-0.04em]">Finish the shareable link.</h2>
                <p className="mt-3 font-body text-sm leading-6 text-ink/55">Campaign creation finished before the link step was interrupted. Continue without rebuilding anything.</p>
                <button className="mt-7 h-14 w-full rounded-full bg-clay px-6 font-display text-sm font-semibold text-white disabled:opacity-60" disabled={busy} onClick={() => void createCampaignLink()} type="button">{busy ? "Creating your link..." : "Finish campaign link"}</button>
              </div>
            ) : null}

            {state && step === "complete" && state.feedback_link && finishedHere ? (
              <div>
                <p className="font-body text-xs font-bold uppercase tracking-[0.2em] text-olive">Your campaign is live</p>
                <h2 className="mt-3 font-display text-3xl tracking-[-0.04em]">Ready for your first honest answer.</h2>
                <p className="mt-3 font-body text-sm leading-6 text-ink/55">Share this link now. If qurl is unavailable, the guest link still works and you can generate the QR asset later. The link prefix <code className="rounded bg-ink/5 px-1 font-mono text-clay">/f/{state.tenant?.slug ?? "your-restaurant"}/</code> is unique to your restaurant.</p>
                <a className="mt-6 block break-all rounded-2xl bg-[#eef4ff] px-4 py-4 font-body text-sm font-semibold text-[#175cd3]" href={state.feedback_link.destination_url} rel="noreferrer" target="_blank">{state.feedback_link.destination_url}</a>
                {!state.feedback_link.qr_asset_url ? <p className="mt-3 rounded-2xl bg-[#fff7ed] px-4 py-3 font-body text-sm text-[#9a3412]">Your link is ready. QR generation is temporarily unavailable.</p> : null}
                <div className="mt-6 flex flex-wrap gap-3">
                  <button className="rounded-full bg-clay px-6 py-3 font-body text-sm font-semibold text-white" onClick={() => void copyLink()} type="button">{copied ? "Link copied" : "Copy feedback link"}</button>
                  <Link className="rounded-full border border-ink/15 px-6 py-3 font-body text-sm font-semibold" href="/admin">Open workspace</Link>
                </div>
              </div>
            ) : null}
          </section>
        </div>
      </div>
    </main>
  );
}

function Progress({ done, number, text }: { done: boolean; number: string; text: string }) {
  return <li className="flex items-center gap-3"><span className={`grid h-8 w-8 place-items-center rounded-full font-display text-xs ${done ? "bg-olive text-white" : "bg-ink/8 text-ink/50"}`}>{done ? "✓" : number}</span>{text}</li>;
}

function Field({ defaultValue, label, name, placeholder }: { defaultValue?: string; label: string; name: string; placeholder: string }) {
  return <label className="block"><span className="mb-2 block font-body text-sm font-semibold">{label}</span><input className="h-14 w-full rounded-2xl border border-ink/15 bg-white px-4 font-body outline-none transition placeholder:text-ink/30 focus:border-clay focus:ring-4 focus:ring-clay/10" defaultValue={defaultValue} name={name} placeholder={placeholder} required /></label>;
}
