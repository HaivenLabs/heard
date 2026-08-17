"use client";

import { FormEvent, useEffect, useState } from "react";
import { useAdminSession } from "../../../components/admin-session";
import { RatingFace, RatingValue } from "../../../components/rating-face";
import { apiFetch, FeedbackLink, Location, Session, SurveyCampaign } from "../../../lib/api";

const defaultCampaign = {
  name: "Takeout bag gift card survey",
  restaurantName: "nom",
  headline: "How did we do?",
  prompt: "Tap the face that matches your visit.",
  incentive: "Complete this survey for a chance to win a $100 nom gift card.",
  smsKeyword: "WIN",
  smsPhone: "(877) 426-0492",
  googleReviewURL: "https://www.google.com/maps/search/?api=1&query=Nom+restaurant",
  yelpReviewURL: "https://www.yelp.com/search?find_desc=Nom"
};

export default function CampaignBuilderPage() {
  const session = useAdminSession();
  return <CampaignBuilder session={session} />;
}

function CampaignBuilder({ session }: { session: Session }) {
  const tenantId = session.tenant_id;
  const [locations, setLocations] = useState<Location[]>([]);
  const [campaign, setCampaign] = useState<SurveyCampaign | null>(null);
  const [link, setLink] = useState<FeedbackLink | null>(null);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    void apiFetch<{ items: Location[] }>("/api/v1/locations", { tenantId })
      .then((payload) => setLocations(payload.items))
      .catch((caught) => setError(caught instanceof Error ? caught.message : "Could not load locations"));
  }, [tenantId]);

  async function createCampaign(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBusy(true);
    setError("");
    const form = new FormData(event.currentTarget);
    const locationId = String(form.get("location_id") ?? "");

    try {
      const createdCampaign = await apiFetch<SurveyCampaign>("/api/v1/survey-campaigns", {
        method: "POST",
        tenantId,
        body: {
          tenant_id: tenantId,
          location_id: locationId,
          name: String(form.get("name") ?? ""),
          restaurant_name: String(form.get("restaurant_name") ?? ""),
          headline: String(form.get("headline") ?? ""),
          prompt: String(form.get("prompt") ?? ""),
          incentive_text: String(form.get("incentive_text") ?? ""),
          sms_keyword: String(form.get("sms_keyword") ?? ""),
          sms_phone: String(form.get("sms_phone") ?? ""),
          google_review_url: String(form.get("google_review_url") ?? ""),
          yelp_review_url: String(form.get("yelp_review_url") ?? "")
        }
      });

      const createdLink = await apiFetch<FeedbackLink>("/api/v1/feedback-links", {
        method: "POST",
        tenantId,
        body: {
          tenant_id: tenantId,
          location_id: locationId,
          campaign_id: createdCampaign.id,
          name: `${createdCampaign.name} QR`,
          channel: "flyer"
        }
      });

      setCampaign(createdCampaign);
      setLink(createdLink);
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : "Could not create survey campaign");
    } finally {
      setBusy(false);
    }
  }

  return (
    <main className="px-5 py-10 text-ink sm:py-14">
        <div className="mx-auto max-w-7xl">
          <div className="mb-8 flex flex-wrap items-end justify-between gap-5">
          <div>
              <p className="font-body text-xs font-bold uppercase tracking-[0.28em] text-clay">Campaign builder</p>
              <h1 className="mt-4 font-display text-4xl tracking-[-0.05em] sm:text-5xl">Create a flyer guests will actually scan.</h1>
              <p className="mt-3 max-w-2xl font-body text-sm leading-7 text-ink/55">Choose the location, tune the guest-facing message, and preview the complete takeout flyer before publishing it.</p>
          </div>
            <a className="inline-flex rounded-full border border-ink/15 bg-[#fffdf8] px-5 py-3 font-body text-sm font-semibold text-ink/65 transition hover:border-clay hover:text-clay" href={link?.destination_url ?? "/f/demo-heard"}>Open guest survey</a>
          </div>

          <div className="grid gap-6 lg:grid-cols-[minmax(0,0.95fr)_minmax(420px,1.05fr)]">
        <form className="space-y-5" onSubmit={createCampaign}>
          {error ? <div className="rounded-2xl border border-red-200 bg-red-50 px-4 py-3 font-body text-sm text-red-700">{error}</div> : null}

          <section className="rounded-[2rem] border border-ink/10 bg-[#fffdf8] p-6 shadow-soft">
            <div className="grid gap-4 sm:grid-cols-2">
              <label className="block sm:col-span-2">
                <span className="mb-1 block text-sm font-medium">Location</span>
                <select className="h-12 w-full rounded-2xl border border-ink/15 bg-white px-4 font-body text-sm outline-none transition focus:border-clay focus:ring-4 focus:ring-clay/10" name="location_id" required>
                  {locations.map((location) => (
                    <option key={location.id} value={location.id}>{location.name}</option>
                  ))}
                </select>
              </label>
              <TextInput defaultValue={defaultCampaign.restaurantName} label="Restaurant name" name="restaurant_name" />
              <TextInput defaultValue={defaultCampaign.name} label="Campaign name" name="name" />
              <TextInput defaultValue={defaultCampaign.headline} label="Flyer headline" name="headline" />
              <TextInput defaultValue={defaultCampaign.prompt} label="Survey prompt" name="prompt" />
              <label className="block sm:col-span-2">
                <span className="mb-1 block text-sm font-medium">Gift card / giveaway copy</span>
                <textarea className="min-h-24 w-full rounded-2xl border border-ink/15 bg-white px-4 py-3 font-body text-sm outline-none transition focus:border-clay focus:ring-4 focus:ring-clay/10" defaultValue={defaultCampaign.incentive} name="incentive_text" />
              </label>
              <TextInput defaultValue={defaultCampaign.smsKeyword} label="SMS keyword" name="sms_keyword" />
              <TextInput defaultValue={defaultCampaign.smsPhone} label="SMS number shown on flyer" name="sms_phone" />
              <label className="block sm:col-span-2">
                <span className="mb-1 block text-sm font-medium">Google Maps review URL</span>
                <input className="h-12 w-full rounded-2xl border border-ink/15 bg-white px-4 font-body text-sm outline-none transition focus:border-clay focus:ring-4 focus:ring-clay/10" defaultValue={defaultCampaign.googleReviewURL} name="google_review_url" type="url" />
              </label>
              <label className="block sm:col-span-2">
                <span className="mb-1 block text-sm font-medium">Yelp review URL</span>
                <input className="h-12 w-full rounded-2xl border border-ink/15 bg-white px-4 font-body text-sm outline-none transition focus:border-clay focus:ring-4 focus:ring-clay/10" defaultValue={defaultCampaign.yelpReviewURL} name="yelp_review_url" type="url" />
              </label>
            </div>
          </section>

          <button className="h-13 w-full rounded-full bg-clay px-5 py-4 font-display text-sm font-semibold uppercase tracking-[0.12em] text-white shadow-[0_14px_30px_rgba(203,104,67,0.24)] transition hover:-translate-y-0.5 hover:bg-[#b95635] disabled:cursor-not-allowed disabled:opacity-60" disabled={busy || locations.length === 0} type="submit">
            {busy ? "Creating campaign..." : "Create flyer survey link"}
          </button>
        </form>

        <aside className="space-y-4">
          <section className="rounded-[2rem] border border-ink/10 bg-[#fffdf8] p-6 shadow-soft">
            <div className="mx-auto max-w-md rounded-lg border border-[#cfd6e4] bg-[#fffaf5] p-6 text-center shadow-sm">
              <img alt="nom" className="mx-auto mb-4 h-16 w-16 rounded-full" height="64" src="/brands/nom/logo.png" width="64" />
              <p className="text-sm font-medium text-[#667085]">{campaign?.restaurant_name.toLowerCase() === "nom" ? "nom" : campaign?.restaurant_name ?? defaultCampaign.restaurantName}</p>
              <h2 className="mt-3 text-4xl font-black uppercase tracking-normal text-[#111827]">{campaign?.headline ?? defaultCampaign.headline}</h2>
              <div className="mt-5 grid grid-cols-5 gap-2" aria-label="Survey rating options">
                {([1, 2, 3, 4, 5] as RatingValue[]).map((value) => (
                  <RatingFace className="w-full overflow-visible" key={value} rating={value} />
                ))}
              </div>
              <p className="mt-6 text-lg font-black uppercase tracking-normal text-[#111827]">{campaign?.incentive_text ?? defaultCampaign.incentive}</p>
              <div className="mt-6 grid items-center gap-4 sm:grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)]">
                <div className="min-w-0 rounded-lg border-2 border-[#111827] bg-white p-3">
                  {link?.qr_svg ? (
                    <div className="mx-auto max-w-40" dangerouslySetInnerHTML={{ __html: link.qr_svg }} />
                  ) : link?.qr_asset_url ? (
                    <img alt="Generated qurl QR asset" className="mx-auto aspect-square max-w-40 object-contain" src={link.qr_asset_url} />
                  ) : (
                    <div className="grid aspect-square place-items-center rounded-md bg-[#edf2f7] p-3 text-center text-sm text-[#667085]">
                      qurl not configured
                    </div>
                  )}
                </div>
                <p className="font-black uppercase text-[#111827]">or</p>
                <div className="min-w-0 text-left">
                  <p className="text-2xl font-black uppercase"><span className="text-[#f25f4c]">Text</span> {campaign?.sms_keyword ?? defaultCampaign.smsKeyword}</p>
                  <p className="text-2xl font-black uppercase">to</p>
                  <p className="break-words text-2xl font-black text-[#f25f4c]">{campaign?.sms_phone ?? defaultCampaign.smsPhone}</p>
                </div>
              </div>
            </div>
          </section>

          {link ? (
            <section className="rounded-[2rem] border border-ink/10 bg-[#fffdf8] p-6 shadow-soft">
              <p className="text-sm font-medium">Campaign link</p>
              <a className="mt-2 block break-all rounded-md bg-[#eef4ff] px-3 py-3 text-sm text-[#175cd3]" href={link.destination_url} target="_blank">{link.destination_url}</a>
              {!link.qr_asset_url && !link.qr_svg ? (
                <p className="mt-3 rounded-md bg-[#fff7ed] px-3 py-2 text-sm text-[#9a3412]">
                  QR asset generation is disabled because heard does not have `QURL_BASE_URL` configured.
                </p>
              ) : null}
            </section>
          ) : null}
        </aside>
          </div>
        </div>
    </main>
  );
}

function TextInput({ defaultValue, label, name }: { defaultValue: string; label: string; name: string }) {
  return (
    <label className="block">
      <span className="mb-1 block text-sm font-medium">{label}</span>
      <input className="h-12 w-full rounded-2xl border border-ink/15 bg-white px-4 font-body text-sm outline-none transition focus:border-clay focus:ring-4 focus:ring-clay/10" defaultValue={defaultValue} name={name} />
    </label>
  );
}
