"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { useAdminSession } from "../../components/admin-session";
import { apiFetch, Location, RecoveryCase, Session, SurveyCampaign, Tenant } from "../../lib/api";

type ConsoleData = {
  locations: Location[];
  campaigns: SurveyCampaign[];
  cases: RecoveryCase[];
  tenant: Tenant;
};

export default function ConsolePage() {
  const session = useAdminSession();
  return <Console session={session} />;
}

function Console({ session }: { session: Session }) {
  const [data, setData] = useState<ConsoleData | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    Promise.all([
      apiFetch<Tenant>(`/api/v1/tenants/${session.tenant_id}`, { tenantId: session.tenant_id }),
      apiFetch<{ items: Location[] }>("/api/v1/locations", { tenantId: session.tenant_id }),
      apiFetch<{ items: SurveyCampaign[] }>("/api/v1/survey-campaigns", { tenantId: session.tenant_id }),
      apiFetch<{ items: RecoveryCase[] }>("/api/v1/recovery-cases", { tenantId: session.tenant_id })
    ])
      .then(([tenant, locations, campaigns, cases]) => setData({ tenant, locations: locations.items, campaigns: campaigns.items, cases: cases.items }))
      .catch((caught) => setError(caught instanceof Error ? caught.message : "Could not load your workspace"));
  }, [session.tenant_id]);

  const openCases = data?.cases.filter((item) => !["resolved", "closed", "archived"].includes(item.status)).length ?? 0;

  return (
    <main className="mx-auto max-w-7xl px-5 py-10 sm:py-14">
        <div className="grid items-end gap-8 lg:grid-cols-[1fr_auto]">
          <div>
            <p className="font-body text-xs uppercase tracking-[0.28em] text-clay">{data?.tenant.name ?? "Your workspace"} · Free</p>
            <h1 className="mt-4 max-w-3xl font-display text-4xl tracking-[-0.05em] sm:text-6xl">
              Good evening, {session.identity.display_name.split(" ")[0]}.
            </h1>
            <p className="mt-4 max-w-2xl font-body text-base leading-7 text-ink/58">
              Launch a feedback campaign tonight. Every response under five lands in recovery so your team knows who needs a human follow-up.
            </p>
          </div>
          <Link className="inline-flex h-[3.25rem] items-center justify-center rounded-full bg-primary px-6 py-3.5 font-display text-sm font-semibold uppercase tracking-[0.16em] text-white shadow-[0_14px_30px_rgba(9,40,21,0.25)] transition hover:-translate-y-0.5 hover:bg-spruce" href="/admin/campaigns?new=1">
            Create campaign
          </Link>
        </div>

        {error ? <div className="mt-8 rounded-2xl border border-red-200 bg-red-50 px-5 py-4 font-body text-sm text-red-700">{error}</div> : null}

        <section className="mt-10 grid gap-4 sm:grid-cols-3">
          <Metric label="Active locations" loading={!data} value={data?.locations.length ?? 0} />
          <Metric label="Campaigns" loading={!data} value={data?.campaigns.length ?? 0} />
          <Metric accent label="Need a response" loading={!data} value={openCases} />
        </section>

        <section className="mt-8 grid gap-6 lg:grid-cols-[1.35fr_0.75fr]">
          <div className="overflow-hidden rounded-[2rem] border border-ink/10 bg-surface shadow-soft">
            <div className="flex items-center justify-between border-b border-ink/8 px-6 py-5">
              <div>
                <p className="font-body text-xs uppercase tracking-[0.22em] text-ink/40">Campaigns</p>
                <h2 className="mt-1 font-display text-2xl tracking-[-0.035em]">What guests can see</h2>
              </div>
              <Link className="font-body text-sm font-semibold text-clay" href="/admin/campaigns?new=1">Build another</Link>
            </div>
            <div className="p-4">
              {!data ? <LoadingRows /> : null}
              {data?.campaigns.length === 0 ? (
                <div className="rounded-[1.5rem] bg-canvas px-6 py-10 text-center">
                  <p className="font-display text-2xl">Your first campaign starts here.</p>
                  <p className="mx-auto mt-2 max-w-md font-body text-sm leading-6 text-ink/55">Choose a location, tune the flyer copy, and heard will create the survey link through the qurl boundary.</p>
                </div>
              ) : null}
              {data?.campaigns.slice(0, 4).map((campaign) => (
                <Link className="flex flex-wrap items-center justify-between gap-4 rounded-[1.4rem] px-4 py-4 transition hover:bg-ink/[0.035] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-clay" href={`/admin/campaigns?campaign=${encodeURIComponent(campaign.id)}`} key={campaign.id}>
                  <div>
                    <p className="font-display text-lg font-semibold">{campaign.name}</p>
                    <p className="mt-1 font-body text-sm text-ink/45">{campaign.restaurant_name.toLowerCase() === "nom" ? "nom" : campaign.restaurant_name} · {campaign.headline}</p>
                  </div>
                  <span className="rounded-full bg-olive/12 px-3 py-1.5 font-body text-xs font-semibold capitalize text-olive">{campaign.status}</span>
                </Link>
              ))}
            </div>
          </div>

          <aside className="rounded-[2rem] bg-primary p-7 text-parchment shadow-soft">
            <p className="font-body text-xs uppercase tracking-[0.25em] text-parchment/45">Tonight&apos;s loop</p>
            <h2 className="mt-4 font-display text-3xl tracking-[-0.04em]">From table to recovery.</h2>
            <ol className="mt-7 space-y-5 font-body text-sm leading-6 text-parchment/68">
              <li className="flex gap-4"><Step number="1" /> Create the campaign and guest-facing link.</li>
              <li className="flex gap-4"><Step number="2" /> Open the survey and submit a rating below five.</li>
              <li className="flex gap-4"><Step number="3" /> Find the guest in the recovery inbox.</li>
            </ol>
            <Link className="mt-8 inline-flex rounded-full border border-parchment/20 px-5 py-3 font-body text-sm font-semibold transition hover:bg-parchment/10" href="/admin/recovery">Open recovery inbox</Link>
          </aside>
        </section>
    </main>
  );
}

function Metric({ accent = false, label, loading, value }: { accent?: boolean; label: string; loading: boolean; value: number }) {
  return (
    <div className={`rounded-[1.6rem] border p-6 ${accent ? "border-primary/25 bg-primary text-white" : "border-ink/10 bg-surface"}`}>
      <p className={`font-body text-xs uppercase tracking-[0.2em] ${accent ? "text-white/65" : "text-ink/42"}`}>{label}</p>
      <p className="mt-4 font-display text-5xl tracking-[-0.06em]">{loading ? "–" : value}</p>
    </div>
  );
}

function LoadingRows() {
  return <div className="space-y-3">{[0, 1, 2].map((item) => <div className="h-16 animate-pulse rounded-[1.4rem] bg-ink/5" key={item} />)}</div>;
}

function Step({ number }: { number: string }) {
  return <span className="grid h-7 w-7 shrink-0 place-items-center rounded-full bg-parchment/10 font-display text-xs text-parchment">{number}</span>;
}
