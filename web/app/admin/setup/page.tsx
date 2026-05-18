"use client";

import { FormEvent, useEffect, useState } from "react";
import { apiFetch, DEMO_TENANT_ID, FeedbackLink, Location, Tenant } from "../../../lib/api";

const defaultTenant = {
  name: "North Star Noodles",
  slug: "north-star-noodles"
};

const commonTimezones = [
  "America/Los_Angeles",
  "America/Denver",
  "America/Chicago",
  "America/New_York",
  "America/Phoenix",
  "America/Anchorage",
  "Pacific/Honolulu"
];

export default function SetupPage() {
  const [tenant, setTenant] = useState<Tenant | null>(null);
  const [locations, setLocations] = useState<Location[]>([]);
  const [link, setLink] = useState<FeedbackLink | null>(null);
  const [tenantId, setTenantId] = useState(DEMO_TENANT_ID);
  const [busy, setBusy] = useState<"" | "tenant" | "location" | "link">("");
  const [error, setError] = useState("");

  useEffect(() => {
    const storedTenantId = window.localStorage.getItem("heard-tenant-id");
    if (storedTenantId) {
      setTenantId(storedTenantId);
    }
  }, []);

  async function onCreateTenant(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    setBusy("tenant");
    setError("");
    try {
      const created = await apiFetch<Tenant>("/api/v1/tenants", {
        method: "POST",
        body: {
          name: String(form.get("name") ?? ""),
          slug: String(form.get("slug") ?? "")
        }
      });
      setTenant(created);
      setTenantId(created.id);
      window.localStorage.setItem("heard-tenant-id", created.id);
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : "Could not create tenant");
    } finally {
      setBusy("");
    }
  }

  async function onCreateLocation(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBusy("location");
    setError("");
    const form = new FormData(event.currentTarget);
    try {
      const created = await apiFetch<Location>("/api/v1/locations", {
        method: "POST",
        tenantId,
        body: {
          tenant_id: tenantId,
          name: String(form.get("name") ?? ""),
          slug: String(form.get("slug") ?? ""),
          timezone: String(form.get("timezone") ?? "America/Los_Angeles")
        }
      });
      setLocations((current) => [created, ...current.filter((item) => item.id !== created.id)]);
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : "Could not create location");
    } finally {
      setBusy("");
    }
  }

  async function onCreateLink(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBusy("link");
    setError("");
    const form = new FormData(event.currentTarget);
    try {
      const created = await apiFetch<FeedbackLink>("/api/v1/feedback-links", {
        method: "POST",
        tenantId,
        body: {
          tenant_id: tenantId,
          location_id: String(form.get("location_id") ?? ""),
          name: String(form.get("name") ?? ""),
          channel: String(form.get("channel") ?? "dine-in")
        }
      });
      setLink(created);
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : "Could not create feedback link");
    } finally {
      setBusy("");
    }
  }

  return (
    <main className="min-h-screen bg-[#faf4eb] text-ink">
      <div className="mx-auto max-w-6xl px-6 py-12">
        <div className="mb-10 flex flex-wrap items-end justify-between gap-4">
          <div>
            <p className="font-display text-xs uppercase tracking-[0.32em] text-clay">Admin setup</p>
            <h1 className="mt-4 font-display text-4xl">Build the Slice 1 loop in a few minutes.</h1>
          </div>
          <div className="rounded-full border border-ink/10 bg-white px-4 py-2 font-mono text-xs text-ink/70 shadow-soft">
            Active tenant header: {tenantId}
          </div>
        </div>

        {error ? <div className="mb-6 rounded-3xl border border-red-300 bg-red-50 px-5 py-4 font-body text-sm text-red-700">{error}</div> : null}

        <div className="grid gap-6 lg:grid-cols-3">
          <form className="rounded-[2rem] border border-ink/10 bg-white p-6 shadow-soft" onSubmit={onCreateTenant}>
            <p className="font-display text-sm uppercase tracking-[0.25em] text-clay">1. Tenant</p>
            <div className="mt-5 space-y-4">
              <label className="block">
                <span className="mb-2 block font-body text-sm text-ink/70">Restaurant name</span>
                <input className="w-full rounded-2xl border border-ink/10 bg-mist/40 px-4 py-3 outline-none ring-0 transition focus:border-clay" defaultValue={defaultTenant.name} name="name" />
              </label>
              <label className="block">
                <span className="mb-2 block font-body text-sm text-ink/70">Slug</span>
                <input className="w-full rounded-2xl border border-ink/10 bg-mist/40 px-4 py-3 outline-none ring-0 transition focus:border-clay" defaultValue={defaultTenant.slug} name="slug" />
              </label>
              <button className="w-full rounded-full bg-ink px-5 py-3 font-display text-sm uppercase tracking-[0.24em] text-parchment transition hover:bg-ink/90" disabled={busy === "tenant"} type="submit">
                {busy === "tenant" ? "Creating..." : "Create tenant"}
              </button>
            </div>
            {tenant ? <pre className="mt-5 overflow-auto rounded-2xl bg-ink px-4 py-4 text-xs text-parchment">{JSON.stringify(tenant, null, 2)}</pre> : null}
          </form>

          <form className="rounded-[2rem] border border-ink/10 bg-white p-6 shadow-soft" onSubmit={onCreateLocation}>
            <p className="font-display text-sm uppercase tracking-[0.25em] text-clay">2. Location</p>
            <div className="mt-5 space-y-4">
              <label className="block">
                <span className="mb-2 block font-body text-sm text-ink/70">Location name</span>
                <input className="w-full rounded-2xl border border-ink/10 bg-mist/40 px-4 py-3 outline-none transition focus:border-clay" defaultValue="Downtown" name="name" />
              </label>
              <label className="block">
                <span className="mb-2 block font-body text-sm text-ink/70">Slug</span>
                <input className="w-full rounded-2xl border border-ink/10 bg-mist/40 px-4 py-3 outline-none transition focus:border-clay" defaultValue="downtown" name="slug" />
              </label>
              <label className="block">
                <span className="mb-2 block font-body text-sm text-ink/70">Timezone</span>
                <select className="w-full rounded-2xl border border-ink/10 bg-mist/40 px-4 py-3 outline-none transition focus:border-clay" defaultValue="America/Los_Angeles" name="timezone">
                  {commonTimezones.map((timezone) => (
                    <option key={timezone} value={timezone}>
                      {timezone}
                    </option>
                  ))}
                </select>
              </label>
              <button className="w-full rounded-full bg-olive px-5 py-3 font-display text-sm uppercase tracking-[0.24em] text-parchment transition hover:bg-olive/90" disabled={busy === "location"} type="submit">
                {busy === "location" ? "Creating..." : "Create location"}
              </button>
            </div>
            {locations[0] ? <pre className="mt-5 overflow-auto rounded-2xl bg-ink px-4 py-4 text-xs text-parchment">{JSON.stringify(locations[0], null, 2)}</pre> : null}
          </form>

          <form className="rounded-[2rem] border border-ink/10 bg-white p-6 shadow-soft" onSubmit={onCreateLink}>
            <p className="font-display text-sm uppercase tracking-[0.25em] text-clay">3. Feedback link</p>
            <div className="mt-5 space-y-4">
              <label className="block">
                <span className="mb-2 block font-body text-sm text-ink/70">Location</span>
                <select className="w-full rounded-2xl border border-ink/10 bg-mist/40 px-4 py-3 outline-none transition focus:border-clay" defaultValue={locations[0]?.id ?? ""} name="location_id">
                  {locations.length === 0 ? (
                    <option value="">Create a location first</option>
                  ) : (
                    locations.map((item) => (
                      <option key={item.id} value={item.id}>
                        {item.name} ({item.slug})
                      </option>
                    ))
                  )}
                </select>
              </label>
              <label className="block">
                <span className="mb-2 block font-body text-sm text-ink/70">Link name (admin-facing)</span>
                <input className="w-full rounded-2xl border border-ink/10 bg-mist/40 px-4 py-3 outline-none transition focus:border-clay" defaultValue="Table tent QR" name="name" />
              </label>
              <label className="block">
                <span className="mb-2 block font-body text-sm text-ink/70">Channel</span>
                <select className="w-full rounded-2xl border border-ink/10 bg-mist/40 px-4 py-3 outline-none transition focus:border-clay" defaultValue="dine-in" name="channel">
                  <option value="dine-in">Dine-in</option>
                  <option value="takeout">Takeout</option>
                  <option value="delivery">Delivery</option>
                  <option value="catering">Catering</option>
                </select>
              </label>
              <button className="w-full rounded-full bg-clay px-5 py-3 font-display text-sm uppercase tracking-[0.24em] text-parchment transition hover:bg-clay/90 disabled:cursor-not-allowed disabled:opacity-60" disabled={busy === "link" || locations.length === 0} type="submit">
                {busy === "link" ? "Generating..." : "Create link"}
              </button>
            </div>
            {link ? (
              <div className="mt-5 space-y-4">
                <pre className="overflow-auto rounded-2xl bg-ink px-4 py-4 text-xs text-parchment">{JSON.stringify(link, null, 2)}</pre>
                <a className="inline-flex rounded-full border border-ink/10 bg-mist px-4 py-2 font-display text-xs uppercase tracking-[0.24em] text-ink transition hover:border-clay" href={link.destination_url} target="_blank">
                  Open guest flow
                </a>
                {link.qr_svg ? (
                  <div className="rounded-[1.5rem] bg-parchment p-4" dangerouslySetInnerHTML={{ __html: link.qr_svg }} />
                ) : null}
              </div>
            ) : null}
          </form>
        </div>
      </div>
    </main>
  );
}
