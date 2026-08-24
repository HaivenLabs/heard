"use client";

import { FormEvent, useEffect, useState } from "react";
import { useAdminSession } from "../../../components/admin-session";
import { apiFetch, Tenant } from "../../../lib/api";
import { slugify } from "../../../lib/slug";

export default function SettingsPage() {
  const session = useAdminSession();
  const [tenant, setTenant] = useState<Tenant | null>(null);
  const [handle, setHandle] = useState("");
  const [availability, setAvailability] = useState<"available" | "taken" | "checking" | "idle">("idle");
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");

  useEffect(() => {
    void apiFetch<Tenant>(`/api/v1/tenants/${session.tenant_id}`, { tenantId: session.tenant_id }).then((item) => { setTenant(item); setHandle(item.slug); }).catch(() => setMessage("Could not load restaurant settings."));
  }, [session.tenant_id]);

  useEffect(() => {
    const normalized = slugify(handle);
    if (!tenant || normalized === tenant.slug) { setAvailability("idle"); return; }
    if (normalized.length < 2) return;
    setAvailability("checking");
    const timeout = window.setTimeout(() => void apiFetch<{ available: boolean }>(`/api/v1/tenant-handles/${encodeURIComponent(normalized)}/availability`).then((result) => setAvailability(result.available ? "available" : "taken")).catch(() => setAvailability("idle")), 250);
    return () => window.clearTimeout(timeout);
  }, [handle, tenant]);

  async function save(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!tenant) return;
    setBusy(true); setMessage("");
    try {
      const updated = await apiFetch<Tenant>(`/api/v1/tenants/${tenant.id}`, { method: "PATCH", tenantId: tenant.id, body: { slug: handle } });
      setTenant(updated); setHandle(updated.slug); setMessage("Public handle updated. Reprint any QR codes that used the old link.");
    } catch (caught) { setMessage(caught instanceof Error ? caught.message : "Could not update the public handle."); } finally { setBusy(false); }
  }

  return <main className="mx-auto max-w-3xl px-5 py-10 sm:py-14"><p className="font-body text-xs uppercase tracking-[0.28em] text-clay">Restaurant settings</p><h1 className="mt-4 font-display text-4xl tracking-[-0.05em]">Your public restaurant handle.</h1><p className="mt-3 max-w-2xl font-body leading-7 text-ink/58">Choose the prefix guests see in every survey link. It must be unique across heard.</p><form className="mt-8 rounded-[2rem] border border-ink/10 bg-[#fffdf8] p-6 shadow-soft sm:p-8" onSubmit={save}><label className="block"><span className="mb-2 block font-body text-sm font-semibold">Restaurant handle</span><div className="flex items-center rounded-2xl border border-ink/15 bg-white px-4 focus-within:border-clay focus-within:ring-4 focus-within:ring-clay/10"><span className="font-mono text-sm text-ink/45">/f/</span><input className="h-14 min-w-0 flex-1 bg-transparent px-1 font-body outline-none" onChange={(event) => setHandle(event.target.value)} required value={handle} /></div></label><p aria-live="polite" className={`mt-3 font-body text-sm ${availability === "taken" ? "text-red-700" : availability === "available" ? "text-olive" : "text-ink/55"}`}>{availability === "checking" ? "Checking availability…" : availability === "available" ? "That handle is available." : availability === "taken" ? "That handle is already taken." : "Changing this updates active survey destinations. Existing printed QR codes need to be replaced."}</p>{message ? <p aria-live="polite" className="mt-4 rounded-2xl bg-[#f5efe6] px-4 py-3 font-body text-sm text-ink/70">{message}</p> : null}<button className="mt-6 h-13 rounded-full bg-clay px-6 font-display text-sm font-semibold uppercase tracking-[0.12em] text-white disabled:opacity-60" disabled={busy || availability === "taken" || availability === "checking"} type="submit">{busy ? "Saving…" : "Save public handle"}</button></form></main>;
}
