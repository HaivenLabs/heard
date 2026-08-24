"use client";

import { ChangeEvent, FormEvent, useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import { useAdminSession } from "../../../components/admin-session";
import { RatingFace, RatingValue } from "../../../components/rating-face";
import { apiFetch, FeedbackLink, Location, Session, SurveyCampaign } from "../../../lib/api";
import { slugify } from "../../../lib/slug";

const defaultCampaign = {
  name: "Guest feedback campaign",
  restaurantName: "",
  headline: "How did we do?",
  prompt: "Tap the face that matches your visit.",
  incentive: "",
  smsKeyword: "",
  smsPhone: "",
  googleReviewURL: "",
  yelpReviewURL: "",
  logoURL: "",
  pathSlug: "feedback"
};

const emptyCampaignForm = {
  location_id: "",
  restaurant_name: defaultCampaign.restaurantName,
  name: defaultCampaign.name,
  logo_url: defaultCampaign.logoURL,
  headline: defaultCampaign.headline,
  prompt: defaultCampaign.prompt,
  incentive_text: defaultCampaign.incentive,
  sms_keyword: defaultCampaign.smsKeyword,
  sms_phone: defaultCampaign.smsPhone,
  google_review_url: defaultCampaign.googleReviewURL,
  yelp_review_url: defaultCampaign.yelpReviewURL,
  path_slug: defaultCampaign.pathSlug
};

export default function CampaignBuilderPage() {
  const session = useAdminSession();
  return <CampaignBuilder session={session} />;
}

function CampaignBuilder({ session }: { session: Session }) {
  const searchParams = useSearchParams();
  const tenantId = session.tenant_id;
  const campaignID = searchParams.get("campaign")?.trim() ?? "";
  const creatingNew = searchParams.get("new") === "1";
  const [locations, setLocations] = useState<Location[]>([]);
  const [campaign, setCampaign] = useState<SurveyCampaign | null>(null);
  const [link, setLink] = useState<FeedbackLink | null>(null);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [copied, setCopied] = useState(false);
  const [origin, setOrigin] = useState("http://localhost:3010");
  const [tenantSlug, setTenantSlug] = useState("");

  const [formData, setFormData] = useState(emptyCampaignForm);

  useEffect(() => {
    if (!tenantId) return;
    void apiFetch<{ id: string; name: string; slug: string }>(`/api/v1/tenants/${tenantId}`, { tenantId })
      .then((tenant) => {
        setTenantSlug(tenant.slug);
        setFormData((previous) => ({ ...previous, restaurant_name: previous.restaurant_name || tenant.name }));
      })
      .catch(() => {});
  }, [tenantId]);

  useEffect(() => {
    if (typeof window !== "undefined") {
      setOrigin(window.location.origin);
    }
  }, []);

  useEffect(() => {
    void apiFetch<{ items: Location[] }>("/api/v1/locations", { tenantId })
      .then((payload) => {
        setLocations(payload.items);
        if (payload.items.length > 0 && !formData.location_id) {
          setFormData((prev) => ({ ...prev, location_id: payload.items[0].id }));
        }
      })
      .catch((caught) => setError(caught instanceof Error ? caught.message : "Could not load locations"));
  }, [tenantId]);

  useEffect(() => {
    if (!tenantId || creatingNew) return;
    const campaignRequest: Promise<{ campaign?: SurveyCampaign; feedback_link?: FeedbackLink }> = campaignID
      ? apiFetch<SurveyCampaign>(`/api/v1/survey-campaigns/${encodeURIComponent(campaignID)}`, { tenantId }).then((item) => ({ campaign: item }))
      : apiFetch<{ campaign?: SurveyCampaign; feedback_link?: FeedbackLink }>("/api/v1/onboarding");
    void campaignRequest
      .then(async (payload) => {
        if (payload.campaign) {
          setCampaign(payload.campaign);
          setFormData((prev) => ({
            ...prev,
            location_id: payload.campaign?.location_id || prev.location_id,
            restaurant_name: payload.campaign?.restaurant_name || prev.restaurant_name,
            name: payload.campaign?.name || prev.name,
            logo_url: payload.campaign?.logo_url || prev.logo_url,
            headline: payload.campaign?.headline || prev.headline,
            prompt: payload.campaign?.prompt || prev.prompt,
            incentive_text: payload.campaign?.incentive_text || prev.incentive_text,
            sms_keyword: payload.campaign?.sms_keyword || prev.sms_keyword,
            sms_phone: payload.campaign?.sms_phone || prev.sms_phone,
            google_review_url: payload.campaign?.google_review_url || prev.google_review_url,
            yelp_review_url: payload.campaign?.yelp_review_url || prev.yelp_review_url
          }));
          const links = await apiFetch<{ items: FeedbackLink[] }>(`/api/v1/feedback-links?campaign_id=${encodeURIComponent(payload.campaign.id)}`, { tenantId });
          if (links.items[0]) setLink(links.items[0]);
        }
        if (payload.feedback_link) setLink(payload.feedback_link);
      })
      .catch(() => {});
  }, [campaignID, creatingNew, tenantId]);

  useEffect(() => {
    if (!creatingNew) return;
    setCampaign(null);
    setLink(null);
    setFormData((previous) => ({ ...emptyCampaignForm, location_id: previous.location_id }));
  }, [creatingNew]);

  function handleInputChange(e: ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  }

  function handleLogoUpload(e: ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (event) => {
        const result = event.target?.result as string;
        setFormData((prev) => ({ ...prev, logo_url: result }));
      };
      reader.readAsDataURL(file);
    }
  }

  function removeLogo() {
    setFormData((prev) => ({ ...prev, logo_url: "" }));
  }

  async function saveCampaign(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBusy(true);
    setError("");

    try {
      const campaignBody = {
        location_id: formData.location_id,
        name: formData.name,
        restaurant_name: formData.restaurant_name,
        headline: formData.headline,
        prompt: formData.prompt,
        incentive_text: formData.incentive_text,
        sms_keyword: formData.sms_keyword,
        sms_phone: formData.sms_phone,
        google_review_url: formData.google_review_url,
        yelp_review_url: formData.yelp_review_url,
        logo_url: formData.logo_url
      };
      const savedCampaign = await apiFetch<SurveyCampaign>(campaign ? `/api/v1/survey-campaigns/${campaign.id}` : "/api/v1/survey-campaigns", {
        method: campaign ? "PATCH" : "POST",
        tenantId,
        body: campaign ? campaignBody : { ...campaignBody, tenant_id: tenantId }
      });

      const linkBody = {
        tenant_id: tenantId,
        campaign_id: savedCampaign.id,
        slug: formData.path_slug
      };

      let createdLink: FeedbackLink;
      if (link) {
        createdLink = await apiFetch<FeedbackLink>(`/api/v1/feedback-links/${link.id}`, {
          method: "PATCH",
          tenantId,
          body: { campaign_id: savedCampaign.id, slug: formData.path_slug }
        });
      } else {
        createdLink = await apiFetch<FeedbackLink>("/api/v1/feedback-links", {
          method: "POST",
          tenantId,
          body: {
            ...linkBody,
            location_id: formData.location_id,
            name: `${savedCampaign.name} QR`,
            channel: "flyer"
          }
        });
      }

      setCampaign(savedCampaign);
      setLink(createdLink);
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : "Could not save survey campaign");
    } finally {
      setBusy(false);
    }
  }

  const restaurantHandle = tenantSlug || slugify(formData.restaurant_name || "restaurant");
  const cleanPathSlug = formData.path_slug.trim().toLowerCase().replace(/[^a-z0-9/-]+/g, "").replace(/^\/+|\/+$/g, "");
  const activeSurveyUrl = `${origin}/f/${restaurantHandle}/${cleanPathSlug}`;

  async function copySurveyLink() {
    await navigator.clipboard.writeText(activeSurveyUrl);
    setCopied(true);
    setTimeout(() => setCopied(false), 2500);
  }

  return (
    <main className="px-5 py-10 text-ink sm:py-14">
      <div className="mx-auto max-w-7xl">
        <div className="mb-8 flex flex-wrap items-end justify-between gap-5">
          <div>
            <p className="font-body text-xs font-bold uppercase tracking-[0.28em] text-clay">Campaign builder</p>
            <h1 className="mt-4 font-display text-4xl tracking-[-0.05em] sm:text-5xl">{campaign ? "Edit your guest feedback campaign." : "Create a flyer guests will actually scan."}</h1>
            <p className="mt-3 max-w-2xl font-body text-sm leading-7 text-ink/55">{campaign ? "Tune the guest-facing message, then save changes to this campaign and its existing flyer link." : "Choose the location, tune the guest-facing message, and preview the complete takeout flyer in real-time."}</p>
          </div>
        </div>

        <div className="grid gap-6 lg:grid-cols-[minmax(0,0.95fr)_minmax(420px,1.05fr)]">
          <form className="space-y-5" onSubmit={saveCampaign}>
            {error ? <div className="rounded-2xl border border-red-200 bg-red-50 px-4 py-3 font-body text-sm text-red-700">{error}</div> : null}

            <section className="rounded-[2rem] border border-ink/10 bg-[#fffdf8] p-6 shadow-soft">
              <div className="grid gap-4 sm:grid-cols-2">
                <label className="block sm:col-span-2">
                  <span className="mb-1 block text-sm font-medium">Location</span>
                  <select className="h-12 w-full rounded-2xl border border-ink/15 bg-white px-4 font-body text-sm outline-none transition focus:border-clay focus:ring-4 focus:ring-clay/10" name="location_id" onChange={handleInputChange} value={formData.location_id} required>
                    {locations.map((location) => (
                      <option key={location.id} value={location.id}>{location.name}</option>
                    ))}
                  </select>
                </label>

                <TextInput label="Restaurant name" name="restaurant_name" onChange={handleInputChange} required value={formData.restaurant_name} />
                <TextInput label="Campaign name" name="name" onChange={handleInputChange} required value={formData.name} />

                {/* Logo Image File Uploader */}
                <div className="sm:col-span-2">
                  <span className="mb-1 block text-sm font-medium">Restaurant logo</span>
                  <div className="flex flex-wrap items-center gap-4 rounded-2xl border border-ink/15 bg-white p-4">
                    {formData.logo_url ? (
                      <div className="relative h-16 w-16 overflow-hidden rounded-full border border-ink/10 bg-[#f7f1e6] shadow-sm shrink-0">
                        <img alt="Logo preview" className="h-full w-full object-cover" src={formData.logo_url} />
                      </div>
                    ) : (
                      <div className="grid h-16 w-16 place-items-center rounded-full bg-[#416a6c] font-display text-xl font-bold lowercase text-[#f8f0dc] shrink-0">
                        {(formData.restaurant_name || "R").slice(0, 1)}
                      </div>
                    )}
                    <div className="flex flex-1 flex-wrap items-center gap-3">
                      <label className="cursor-pointer rounded-full bg-clay/10 px-4 py-2 font-body text-xs font-semibold text-clay transition hover:bg-clay/20">
                        <span>{formData.logo_url ? "Change logo image" : "Upload logo image"}</span>
                        <input accept="image/*" className="hidden" onChange={handleLogoUpload} type="file" />
                      </label>
                      {formData.logo_url ? (
                        <button className="rounded-full px-3 py-2 font-body text-xs font-semibold text-red-600 hover:bg-red-50" onClick={removeLogo} type="button">
                          Remove
                        </button>
                      ) : null}
                    </div>
                  </div>
                </div>

                <TextInput label="Flyer headline" name="headline" onChange={handleInputChange} required value={formData.headline} />
                <TextInput label="Survey prompt" name="prompt" onChange={handleInputChange} required value={formData.prompt} />

                <label className="block sm:col-span-2">
                  <span className="mb-1 block text-sm font-medium">Gift card / giveaway copy</span>
                  <textarea className="min-h-24 w-full rounded-2xl border border-ink/15 bg-white px-4 py-3 font-body text-sm outline-none transition focus:border-clay focus:ring-4 focus:ring-clay/10" name="incentive_text" onChange={handleInputChange} value={formData.incentive_text} />
                </label>

                <TextInput label="SMS keyword" name="sms_keyword" onChange={handleInputChange} value={formData.sms_keyword} />
                <TextInput label="SMS number shown on flyer" name="sms_phone" onChange={handleInputChange} value={formData.sms_phone} />

                <label className="block sm:col-span-2">
                  <span className="mb-1 block text-sm font-medium">Google Maps review URL</span>
                  <input className="h-12 w-full rounded-2xl border border-ink/15 bg-white px-4 font-body text-sm outline-none transition focus:border-clay focus:ring-4 focus:ring-clay/10" name="google_review_url" onChange={handleInputChange} type="url" value={formData.google_review_url} />
                </label>

                <label className="block sm:col-span-2">
                  <span className="mb-1 block text-sm font-medium">Yelp review URL</span>
                  <input className="h-12 w-full rounded-2xl border border-ink/15 bg-white px-4 font-body text-sm outline-none transition focus:border-clay focus:ring-4 focus:ring-clay/10" name="yelp_review_url" onChange={handleInputChange} type="url" value={formData.yelp_review_url} />
                </label>
              </div>
            </section>

            {/* Unified Survey Link & Path Settings Box */}
            <section className="rounded-[2rem] border border-ink/10 bg-[#fffdf8] p-6 shadow-soft">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <span className="font-body text-xs font-bold uppercase tracking-[0.2em] text-olive">Your survey link</span>
                <span className="inline-flex items-center gap-1.5 rounded-full bg-emerald-50 px-3 py-1 font-body text-xs font-semibold text-emerald-700 border border-emerald-200">
                  <span>✓</span> heard handle: @{restaurantHandle}
                </span>
              </div>
              <p className="mt-1.5 font-body text-xs leading-5 text-ink/55">The heard handle <code className="rounded bg-ink/5 px-1 font-mono text-clay">/f/{restaurantHandle}</code> is unique to your restaurant and secured. Everything after it is yours to edit.</p>

              {/* Editable Path input after restaurant handle */}
              <div className="mt-4">
                <span className="mb-1.5 block text-xs font-semibold text-ink/75">Survey path</span>
                <div className="flex flex-wrap items-center rounded-2xl border border-ink/15 bg-white px-4 py-2.5 shadow-inner">
                  <span className="font-body text-xs font-bold text-ink/45 select-none shrink-0 py-1">
                    {origin}/f/<span className="text-clay font-bold">{restaurantHandle}</span>/
                  </span>
                  <input className="h-9 min-w-36 flex-1 font-body text-sm font-bold outline-none text-ink/90 placeholder:text-ink/25" name="path_slug" onChange={handleInputChange} placeholder="sjc/takeout" value={formData.path_slug} />
                </div>
              </div>

              {/* Live URL & Instant Action Controls */}
              <div className="mt-4 rounded-2xl border border-ink/10 bg-[#f7f1e6]/60 p-4">
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <div className="min-w-0 flex-1">
                    <p className="font-body text-[11px] font-bold uppercase tracking-wider text-ink/45">Your survey link</p>
                    <p className="mt-1 break-all font-mono text-xs font-semibold text-clay">{activeSurveyUrl}</p>
                  </div>
                  <div className="flex gap-2 shrink-0">
                    <button className="rounded-full bg-clay/10 px-4 py-2 font-body text-xs font-semibold text-clay transition hover:bg-clay/20" onClick={copySurveyLink} type="button">
                      {copied ? "Copied!" : "Copy link"}
                    </button>
                    <a className="rounded-full bg-ink/5 px-4 py-2 font-body text-xs font-semibold text-ink/75 transition hover:bg-ink/10" href={activeSurveyUrl} rel="noreferrer" target="_blank">
                      Test survey ↗
                    </a>
                  </div>
                </div>
              </div>
            </section>

            {/* Action Submit Button */}
            <button className="h-13 w-full rounded-full bg-clay px-5 py-4 font-display text-sm font-semibold uppercase tracking-[0.12em] text-white shadow-[0_14px_30px_rgba(203,104,67,0.24)] transition hover:-translate-y-0.5 hover:bg-[#b95635] disabled:cursor-not-allowed disabled:opacity-60" disabled={busy || locations.length === 0} type="submit">
              {busy ? "Saving campaign..." : campaign ? "Save campaign changes" : "Create flyer survey link"}
            </button>
          </form>

          {/* Real-time Live Flyer Preview Card */}
          <aside className="space-y-4">
            <section className="sticky top-6 rounded-[2rem] border border-ink/10 bg-[#fffdf8] p-6 shadow-soft">
              <p className="mb-4 font-body text-xs font-bold uppercase tracking-[0.2em] text-ink/45">Live flyer preview</p>
              <div className="mx-auto max-w-md rounded-lg border border-[#cfd6e4] bg-[#fffaf5] p-6 text-center shadow-sm">
                {formData.logo_url ? (
                  <img alt={formData.restaurant_name} className="mx-auto mb-4 h-16 w-16 rounded-full object-cover" height="64" src={formData.logo_url} width="64" />
                ) : (
                  <span className="mx-auto mb-4 grid h-16 w-16 place-items-center rounded-full bg-[#416a6c] font-display text-2xl font-bold lowercase text-[#f8f0dc]">
                    {(formData.restaurant_name || "R").slice(0, 1)}
                  </span>
                )}
                <p className="text-sm font-medium text-[#667085]">{formData.restaurant_name || "Restaurant Name"}</p>
                <h2 className="mt-3 text-4xl font-black uppercase tracking-normal text-[#111827]">{formData.headline || "How did we do?"}</h2>
                <div className="mt-5 grid grid-cols-5 gap-2" aria-label="Survey rating options">
                  {([1, 2, 3, 4, 5] as RatingValue[]).map((value) => (
                    <RatingFace className="w-full overflow-visible" key={value} rating={value} />
                  ))}
                </div>
                <p className="mt-6 text-lg font-black uppercase tracking-normal text-[#111827]">{formData.incentive_text || "Giveaway copy"}</p>
                <div className="mt-6 grid items-center gap-4 sm:grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)]">
                  <div className="min-w-0 rounded-lg border-2 border-[#111827] bg-white p-3">
                    {link?.qr_asset_url ? (
                      <img alt="Generated qurl QR asset" className="mx-auto aspect-square max-w-40 object-contain" src={link.qr_asset_url} />
                    ) : (
                      <div className="grid aspect-square place-items-center rounded-md bg-[#edf2f7] p-3 text-center text-sm text-[#667085]">
                        qurl not configured
                      </div>
                    )}
                  </div>
                  <p className="font-black uppercase text-[#111827]">or</p>
                  <div className="min-w-0 text-left">
                    <p className="text-2xl font-black uppercase"><span className="text-[#f25f4c]">Text</span> {formData.sms_keyword || "WIN"}</p>
                    <p className="text-2xl font-black uppercase">to</p>
                    <p className="break-words text-2xl font-black text-[#f25f4c]">{formData.sms_phone || "(800) 000-0000"}</p>
                  </div>
                </div>
              </div>
            </section>
          </aside>
        </div>
      </div>
    </main>
  );
}

function TextInput({ label, name, onChange, required = false, value }: { label: string; name: string; onChange: (e: ChangeEvent<HTMLInputElement>) => void; required?: boolean; value: string }) {
  return (
    <label className="block">
      <span className="mb-1 block text-sm font-medium">{label}</span>
      <input className="h-12 w-full rounded-2xl border border-ink/15 bg-white px-4 font-body text-sm outline-none transition focus:border-clay focus:ring-4 focus:ring-clay/10" name={name} onChange={onChange} required={required} value={value} />
    </label>
  );
}
