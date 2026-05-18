"use client";

import { useEffect, useMemo, useState, useTransition } from "react";
import { apiFetch, DEMO_TENANT_ID, FeedbackResponse, RecoveryCase, formatRelativeDate } from "../../../lib/api";

const STATUS_OPTIONS = ["new", "open", "assigned", "waiting_on_guest", "resolved", "closed"] as const;

export default function RecoveryPage() {
  const [tenantId, setTenantId] = useState(DEMO_TENANT_ID);
  const [cases, setCases] = useState<RecoveryCase[]>([]);
  const [selectedId, setSelectedId] = useState("");
  const [response, setResponse] = useState<FeedbackResponse | null>(null);
  const [error, setError] = useState("");
  const [isPending, startTransition] = useTransition();

  useEffect(() => {
    const storedTenantId = window.localStorage.getItem("heard-tenant-id");
    if (storedTenantId) {
      setTenantId(storedTenantId);
    }
  }, []);

  useEffect(() => {
    startTransition(() => {
      void refreshCases(tenantId);
    });
  }, [tenantId]);

  const selectedCase = useMemo(() => cases.find((item) => item.id === selectedId) ?? cases[0] ?? null, [cases, selectedId]);

  useEffect(() => {
    if (!selectedCase) {
      setResponse(null);
      return;
    }
    void apiFetch<FeedbackResponse>(`/api/v1/feedback-responses/${selectedCase.feedback_response_id}`, {
      tenantId
    })
      .then((item) => {
        setResponse(item);
        setSelectedId(selectedCase.id);
      })
      .catch((caught) => {
        setError(caught instanceof Error ? caught.message : "Could not load feedback response");
      });
  }, [selectedCase?.id, selectedCase?.feedback_response_id, tenantId]);

  async function refreshCases(activeTenantId: string) {
    try {
      setError("");
      const payload = await apiFetch<{ items: RecoveryCase[] }>("/api/v1/recovery-cases", { tenantId: activeTenantId });
      setCases(payload.items);
      setSelectedId((current) => current || payload.items[0]?.id || "");
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : "Could not load recovery cases");
    }
  }

  async function updateStatus(status: string) {
    if (!selectedCase) {
      return;
    }
    try {
      await apiFetch<RecoveryCase>(`/api/v1/recovery-cases/${selectedCase.id}`, {
        method: "PATCH",
        tenantId,
        body: { status }
      });
      await refreshCases(tenantId);
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : "Could not update status");
    }
  }

  return (
    <main className="min-h-screen bg-[#fbf6ef] px-6 py-10 text-ink">
      <div className="mx-auto max-w-7xl">
        <div className="mb-8 flex flex-wrap items-end justify-between gap-4">
          <div>
            <p className="font-display text-xs uppercase tracking-[0.32em] text-clay">Recovery inbox</p>
            <h1 className="mt-4 font-display text-4xl">Negative feedback lands here after the worker clears the outbox.</h1>
          </div>
          <label className="rounded-full border border-ink/10 bg-white px-4 py-3 shadow-soft">
            <span className="mb-2 block font-body text-xs uppercase tracking-[0.22em] text-ink/50">Tenant header</span>
            <input className="w-72 bg-transparent font-mono text-xs outline-none" onChange={(event) => setTenantId(event.target.value)} value={tenantId} />
          </label>
        </div>

        {error ? <div className="mb-6 rounded-3xl border border-red-300 bg-red-50 px-5 py-4 font-body text-sm text-red-700">{error}</div> : null}

        <div className="grid gap-6 lg:grid-cols-[0.95fr_1.35fr]">
          <section className="rounded-[2rem] border border-ink/10 bg-white p-5 shadow-soft">
            <div className="mb-4 flex items-center justify-between">
              <p className="font-display text-sm uppercase tracking-[0.25em] text-clay">Queue</p>
              <button className="rounded-full border border-ink/10 px-4 py-2 font-display text-xs uppercase tracking-[0.2em] text-ink/70 transition hover:border-clay" onClick={() => void refreshCases(tenantId)} type="button">
                Refresh
              </button>
            </div>
            <div className="space-y-3">
              {cases.length === 0 ? (
                <div className="rounded-[1.5rem] bg-mist/60 px-5 py-8 font-body text-sm leading-7 text-ink/70">
                  No recovery cases yet. Submit a 1 or 2 star response from the guest flow and the worker will create one.
                </div>
              ) : null}
              {cases.map((item) => (
                <button
                  className={`w-full rounded-[1.5rem] border px-4 py-4 text-left transition ${selectedCase?.id === item.id ? "border-clay bg-clay/10" : "border-ink/10 bg-[#fffdf9] hover:border-ink/20"}`}
                  key={item.id}
                  onClick={() => setSelectedId(item.id)}
                  type="button"
                >
                  <div className="flex items-center justify-between gap-3">
                    <p className="font-display text-sm uppercase tracking-[0.18em] text-ink">{item.status.replaceAll("_", " ")}</p>
                    <span className="rounded-full bg-ink px-3 py-1 font-display text-[10px] uppercase tracking-[0.22em] text-parchment">{item.priority}</span>
                  </div>
                  <p className="mt-3 font-body text-sm text-ink/70">{item.feedback_preview || "No written comment."}</p>
                  <div className="mt-4 flex items-center justify-between font-body text-xs uppercase tracking-[0.14em] text-ink/45">
                    <span>{item.sentiment}</span>
                    <span>{formatRelativeDate(item.created_at)}</span>
                  </div>
                </button>
              ))}
            </div>
          </section>

          <section className="rounded-[2rem] border border-ink/10 bg-ink p-7 text-parchment shadow-soft">
            {!selectedCase ? (
              <div className="rounded-[1.5rem] border border-parchment/10 bg-parchment/5 px-6 py-10 font-body text-sm text-parchment/70">
                Pick a case to inspect the original feedback, guest contact state, and current recovery status.
              </div>
            ) : (
              <div className="space-y-6">
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <div>
                    <p className="font-display text-xs uppercase tracking-[0.32em] text-parchment/45">Case detail</p>
                    <h2 className="mt-3 font-display text-3xl">Case {selectedCase.id.slice(0, 8)}</h2>
                  </div>
                  <select className="rounded-full border border-parchment/15 bg-parchment/10 px-4 py-3 font-display text-xs uppercase tracking-[0.2em] text-parchment outline-none" defaultValue={selectedCase.status} onChange={(event) => void updateStatus(event.target.value)}>
                    {STATUS_OPTIONS.map((option) => (
                      <option key={option} value={option}>
                        {option.replaceAll("_", " ")}
                      </option>
                    ))}
                  </select>
                </div>

                <div className="grid gap-4 md:grid-cols-3">
                  <StatCard label="Sentiment" value={selectedCase.sentiment} />
                  <StatCard label="Rating" value={`${selectedCase.rating}/5`} />
                  <StatCard label="Guest" value={selectedCase.guest_name || "Anonymous"} />
                </div>

                <div className="rounded-[1.75rem] border border-parchment/10 bg-parchment/5 p-5">
                  <p className="font-display text-xs uppercase tracking-[0.28em] text-parchment/45">Feedback preview</p>
                  <p className="mt-4 font-body text-base leading-8 text-parchment/90">{selectedCase.feedback_preview || "No written comment was included."}</p>
                </div>

                <div className="grid gap-4 md:grid-cols-2">
                  <InfoCard label="Phone" value={selectedCase.guest_phone || "Not provided"} />
                  <InfoCard label="Email" value={selectedCase.guest_email || "Not provided"} />
                  <InfoCard label="Created" value={formatRelativeDate(selectedCase.created_at)} />
                  <InfoCard label="Rule" value={selectedCase.created_reason.replaceAll("_", " ")} />
                </div>

                <div className="rounded-[1.75rem] border border-parchment/10 bg-parchment/5 p-5">
                  <p className="font-display text-xs uppercase tracking-[0.28em] text-parchment/45">Original feedback record</p>
                  <pre className="mt-4 overflow-auto rounded-[1.25rem] bg-black/25 p-4 text-xs text-parchment/80">{JSON.stringify(response, null, 2)}</pre>
                </div>
              </div>
            )}
            {isPending ? <p className="mt-4 font-body text-xs uppercase tracking-[0.2em] text-parchment/50">Refreshing cases...</p> : null}
          </section>
        </div>
      </div>
    </main>
  );
}

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-[1.5rem] border border-parchment/10 bg-parchment/5 p-5">
      <p className="font-display text-xs uppercase tracking-[0.24em] text-parchment/45">{label}</p>
      <p className="mt-3 font-display text-2xl">{value}</p>
    </div>
  );
}

function InfoCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-[1.5rem] border border-parchment/10 bg-parchment/5 p-5">
      <p className="font-display text-xs uppercase tracking-[0.24em] text-parchment/45">{label}</p>
      <p className="mt-3 font-body text-sm leading-7 text-parchment/85">{value}</p>
    </div>
  );
}
