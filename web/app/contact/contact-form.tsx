"use client";

import Link from "next/link";
import { FormEvent, useState } from "react";
import { BrandBackdrop } from "../../components/brand-backdrop";
import { PublicHeader } from "../../components/public-header";
import { apiFetch, MarketingLead } from "../../lib/api";

type LeadSource = "marketing_site" | "guest_demo";

export default function ContactForm({ source }: { source: LeadSource }) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [lead, setLead] = useState<MarketingLead | null>(null);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const name = String(form.get("name") ?? "").trim();
    const workEmail = String(form.get("work_email") ?? "").trim();
    const phone = String(form.get("phone") ?? "").trim();
    const restaurantName = String(form.get("restaurant_name") ?? "").trim();
    const locationCount = String(form.get("location_count") ?? "");
    const challenge = String(form.get("challenge") ?? "").trim();

    const validationError = validateLead({ name, workEmail, phone, restaurantName, locationCount });
    if (validationError) {
      setError(validationError);
      return;
    }

    setBusy(true);
    setError("");
    try {
      const created = await apiFetch<MarketingLead>("/api/v1/marketing-leads", {
        method: "POST",
        auth: false,
        body: {
          name,
          work_email: workEmail,
          phone,
          restaurant_name: restaurantName,
          location_count: locationCount,
          challenge,
          source,
          contact_consent: true
        }
      });
      setLead(created);
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : "We could not save your request. Please try again.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <main className="relative min-h-screen overflow-hidden bg-[#f7f1e6] text-ink">
      <BrandBackdrop />
      <PublicHeader />

      <div className="relative mx-auto grid max-w-7xl items-start gap-12 px-6 pb-20 pt-10 lg:grid-cols-[0.88fr_0.78fr] lg:gap-20 lg:pb-24 lg:pt-20">
        <section className="lg:sticky lg:top-32">
          <h1 className="max-w-2xl font-display text-5xl leading-[0.98] tracking-[-0.055em] sm:text-7xl">Built for restaurant operators, by restaurant operators.</h1>
          <p className="mt-7 max-w-xl font-display text-2xl leading-tight tracking-[-0.03em] text-clay sm:text-3xl">Now see what happens on the restaurant side.</p>
          <p className="mt-6 max-w-xl font-body text-lg leading-8 text-ink/62">We&apos;re restaurant operators and builders creating the feedback system we want in our own restaurants: quick for guests, clear for teams, and focused on making things right.</p>
          <p className="mt-8 max-w-xl border-l-2 border-olive/55 pl-5 font-body text-sm leading-7 text-ink/55"><strong className="font-semibold text-ink">What happens next:</strong> a real person from heard will learn how you collect feedback today and tailor the walkthrough to your restaurant.</p>
        </section>

        <section className="rounded-[2rem] border border-ink/10 bg-[#fffaf0]/95 p-6 text-ink shadow-[0_28px_80px_rgba(23,37,29,0.14)] backdrop-blur-sm sm:p-9" id="walkthrough">
          {lead ? (
            <div className="py-8 text-center">
              <div className="mx-auto grid h-16 w-16 place-items-center rounded-full bg-olive/15 text-3xl text-olive">&#10003;</div>
              <p className="mt-6 font-body text-xs font-bold uppercase tracking-[0.22em] text-clay">Request received</p>
              <h2 className="mt-3 font-display text-4xl tracking-[-0.045em]">We&apos;ll take it from here.</h2>
              <p className="mx-auto mt-4 max-w-md font-body text-base leading-7 text-ink/60">The heard team has your request for {lead.restaurant_name}. We&apos;ll reach out using the contact details you provided.</p>
              <div className="mt-8 flex flex-wrap justify-center gap-3">
                <Link className="rounded-full bg-ink px-6 py-3 font-body text-sm font-semibold text-parchment" href="/">Back to heard</Link>
                <Link className="rounded-full border border-ink/15 px-6 py-3 font-body text-sm font-semibold" href="/f/demo-heard">Try the guest demo again</Link>
              </div>
            </div>
          ) : (
            <>
              <h2 className="font-display text-4xl leading-tight tracking-[-0.045em]">See heard for your restaurant</h2>
              <p className="mt-3 font-display text-xl tracking-[-0.02em] text-clay">Request a walkthrough built around your operation.</p>
              <p className="mt-3 font-body text-sm leading-6 text-ink/55">Tell us just enough to make the conversation useful. Phone is optional.</p>

              <form className="mt-8 space-y-5" noValidate onSubmit={submit}>
                <LeadField autoComplete="name" label="Your name" name="name" placeholder="Avery Chen" />
                <LeadField autoComplete="email" inputMode="email" label="Work email" name="work_email" placeholder="avery@yourrestaurant.com" type="email" />
                <LeadField autoComplete="organization" label="Restaurant or brand" name="restaurant_name" placeholder="Cedar & Salt" />

                <div className="grid gap-5 sm:grid-cols-2">
                  <label className="block">
                    <span className="mb-2 block font-body text-sm font-semibold">Locations</span>
                    <select className="h-14 w-full rounded-2xl border border-ink/15 bg-white px-4 font-body text-sm outline-none transition focus:border-clay focus:ring-4 focus:ring-clay/10" defaultValue="" name="location_count">
                      <option disabled value="">Choose one</option>
                      <option value="1">1 location</option>
                      <option value="2-4">2-4 locations</option>
                      <option value="5-19">5-19 locations</option>
                      <option value="20-49">20-49 locations</option>
                      <option value="50+">50+ locations</option>
                    </select>
                  </label>
                  <LeadField autoComplete="tel" inputMode="tel" label="Phone (optional)" name="phone" placeholder="(206) 555-0142" />
                </div>

                <label className="block">
                  <span className="mb-2 block font-body text-sm font-semibold">What would you most like to fix? <span className="font-normal text-ink/38">Optional</span></span>
                  <textarea className="min-h-28 w-full rounded-2xl border border-ink/15 bg-white px-4 py-3 font-body text-base outline-none transition placeholder:text-ink/28 focus:border-clay focus:ring-4 focus:ring-clay/10" maxLength={1000} name="challenge" placeholder="More guest feedback, faster recovery, clearer location-level issues..." />
                </label>

                {error ? <div aria-live="assertive" className="rounded-2xl border border-red-200 bg-red-50 px-4 py-3 font-body text-sm text-red-700" role="alert">{error}</div> : null}

                <button className="h-14 w-full rounded-full bg-clay px-6 font-display text-sm font-semibold tracking-[0.06em] text-white shadow-[0_14px_30px_rgba(203,104,67,0.24)] transition hover:-translate-y-0.5 hover:bg-[#b95635] disabled:cursor-wait disabled:opacity-65" disabled={busy} type="submit">{busy ? "Saving your request..." : "Request my walkthrough"}</button>
                <p className="text-center font-body text-xs leading-5 text-ink/40">By submitting, you agree that heard may contact you about this request. We&apos;ll only use these details to follow up about heard.</p>
              </form>
            </>
          )}
        </section>
      </div>
    </main>
  );
}

function validateLead({ name, workEmail, phone, restaurantName, locationCount }: { name: string; workEmail: string; phone: string; restaurantName: string; locationCount: string }) {
  if (name.length < 2) return "Enter your name.";
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(workEmail)) return "Enter a complete work email address.";
  if (restaurantName.length < 2) return "Enter your restaurant or brand name.";
  if (!locationCount) return "Choose the number of restaurant locations.";
  if (phone) {
    const digits = phone.replace(/\D/g, "").length;
    if (!/^\+?[()\d\s.-]+$/.test(phone) || digits < 10 || digits > 15) return "Enter a complete phone number, including area code.";
  }
  return "";
}

function LeadField({ autoComplete, inputMode, label, name, placeholder, type = "text" }: { autoComplete: string; inputMode?: "email" | "tel"; label: string; name: string; placeholder: string; type?: string }) {
  return <label className="block"><span className="mb-2 block font-body text-sm font-semibold">{label}</span><input autoComplete={autoComplete} className="h-14 w-full rounded-2xl border border-ink/15 bg-white px-4 font-body text-base outline-none transition placeholder:text-ink/28 focus:border-clay focus:ring-4 focus:ring-clay/10" inputMode={inputMode} name={name} placeholder={placeholder} type={type} /></label>;
}
