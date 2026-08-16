"use client";

import Link from "next/link";
import { FormEvent, useState } from "react";
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
    <main className="relative min-h-screen overflow-hidden bg-[#17251d] text-parchment">
      <div className="pointer-events-none absolute inset-0 opacity-55 [background-image:radial-gradient(circle_at_1px_1px,rgba(247,241,227,0.14)_1px,transparent_0)] [background-size:30px_30px]" />
      <div className="pointer-events-none absolute -left-40 top-[-14rem] h-[38rem] w-[38rem] rounded-full bg-clay/35 blur-3xl" />
      <div className="pointer-events-none absolute -right-40 bottom-[-16rem] h-[36rem] w-[36rem] rounded-full bg-olive/35 blur-3xl" />

      <header className="relative mx-auto flex max-w-7xl items-center justify-between px-6 py-6">
        <Link className="font-display text-2xl font-semibold tracking-[-0.05em]" href="/">heard<span className="text-clay">.</span></Link>
        <p className="font-body text-sm text-parchment/60">Already a customer? <Link className="font-semibold text-parchment underline decoration-clay/60 underline-offset-4" href="/login">Sign in</Link></p>
      </header>

      <div className="relative mx-auto grid max-w-7xl items-start gap-12 px-6 pb-16 pt-8 lg:grid-cols-[0.9fr_0.78fr] lg:gap-20 lg:pb-24 lg:pt-16">
        <section className="lg:sticky lg:top-12">
          <p className="font-body text-xs font-bold uppercase tracking-[0.28em] text-clay">
            Built for restaurant operators, by restaurant operators.
          </p>
          <h1 className="mt-6 max-w-2xl font-display text-5xl leading-[0.98] tracking-[-0.055em] sm:text-7xl">Now see what happens on the restaurant side.</h1>
          <p className="mt-7 max-w-xl font-body text-lg leading-8 text-parchment/68">heard gives your team a private signal while there is still time to fix the visit, recover the guest, and learn what keeps going wrong by location.</p>

          <div className="mt-10 space-y-5">
            <Outcome number="01" title="Hear from more guests">One quick, branded experience through QR or link. No app and no account for guests.</Outcome>
            <Outcome number="02" title="Know who needs a human">Contact details, the guest&apos;s words, and the operational issue arrive together.</Outcome>
            <Outcome number="03" title="Turn feedback into a next move">Give every location a focused recovery queue instead of another dashboard nobody checks.</Outcome>
          </div>

          <div className="mt-10 rounded-[1.5rem] border border-parchment/12 bg-parchment/[0.06] p-5">
            <p className="font-body text-xs font-bold uppercase tracking-[0.2em] text-[#a9c27c]">What happens next</p>
            <p className="mt-2 font-body text-sm leading-6 text-parchment/68">A real person from heard will learn how you collect feedback today and tailor the walkthrough to your restaurant. No generic sales maze.</p>
          </div>

          <div className="mt-5 border-l-2 border-clay/70 pl-5">
            <p className="font-body text-xs font-bold uppercase tracking-[0.2em] text-clay">About us</p>
            <p className="mt-2 max-w-xl font-body text-sm leading-6 text-parchment/68">We&apos;re restaurant operators and builders creating the guest feedback system we want in our own restaurants: quick for guests, actionable for teams, and focused on making things right.</p>
          </div>
        </section>

        <section className="rounded-[2rem] border border-parchment/15 bg-[#fffaf0] p-6 text-ink shadow-[0_38px_110px_rgba(0,0,0,0.32)] sm:p-9">
          {lead ? (
            <div className="py-8 text-center">
              <div className="mx-auto grid h-16 w-16 place-items-center rounded-full bg-olive/15 text-3xl text-olive">✓</div>
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
              <p className="font-body text-xs font-bold uppercase tracking-[0.22em] text-clay">See heard for your restaurant</p>
              <h2 className="mt-3 font-display text-4xl tracking-[-0.045em]">Request a walkthrough.</h2>
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

function Outcome({ children, number, title }: { children: React.ReactNode; number: string; title: string }) {
  return <div className="grid grid-cols-[2.5rem_1fr] gap-4"><span className="font-body text-xs font-bold text-clay">{number}</span><div><h2 className="font-display text-xl font-semibold tracking-[-0.02em]">{title}</h2><p className="mt-1 font-body text-sm leading-6 text-parchment/58">{children}</p></div></div>;
}
