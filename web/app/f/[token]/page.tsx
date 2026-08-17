"use client";

import Link from "next/link";
import type { Route } from "next";
import { FormEvent, useEffect, useMemo, useState } from "react";
import { RatingFace, RatingValue } from "../../../components/rating-face";
import { apiFetch, FeedbackResponse, FeedbackSession, PublicSurvey } from "../../../lib/api";
import { isValidEmail, isValidPhone } from "../../../lib/contact-validation";

const ratings: Array<{ value: RatingValue; label: string; prompt: string }> = [
  { value: 1, label: "Not good", prompt: "We are sorry this missed the mark." },
  { value: 2, label: "Could be better", prompt: "Thank you for helping us improve." },
  { value: 3, label: "It was okay", prompt: "We would love to make the next visit better." },
  { value: 4, label: "Really good", prompt: "We are glad it was a good visit." },
  { value: 5, label: "Loved it", prompt: "We love hearing that." }
];

const issueTags = ["Food", "Speed", "Service", "Accuracy", "Packaging", "Value"];

export default function FlyerSurveyPage({ params }: { params: Promise<{ token: string }> }) {
  const [token, setToken] = useState("");
  const [survey, setSurvey] = useState<PublicSurvey | null>(null);
  const [session, setSession] = useState<FeedbackSession | null>(null);
  const [submitted, setSubmitted] = useState<FeedbackResponse | null>(null);
  const [rating, setRating] = useState<RatingValue | 0>(0);
  const [comment, setComment] = useState("");
  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  const [guestName, setGuestName] = useState("");
  const [guestPhone, setGuestPhone] = useState("");
  const [guestEmail, setGuestEmail] = useState("");
  const [marketingConsent, setMarketingConsent] = useState(false);
  const [heardFollowUp, setHeardFollowUp] = useState(false);
  const [reviewClicks, setReviewClicks] = useState<string[]>([]);
  const [step, setStep] = useState<"rate" | "details" | "done">("rate");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    void params.then((value) => setToken(value.token));
  }, [params]);

  useEffect(() => {
    if (!token) {
      return;
    }
    void apiFetch<PublicSurvey>(`/api/v1/public/surveys/${token}`)
      .then((payload) => setSurvey(payload))
      .catch((caught) => setError(caught instanceof Error ? caught.message : "This survey link is not available."));
  }, [token]);

  const selectedRating = useMemo(() => ratings.find((item) => item.value === rating), [rating]);
  const isFiveStar = rating === 5;
  const isDemo = token === "demo-heard";
  const isNom = (survey?.campaign.restaurant_name ?? "nom").toLowerCase() === "nom";
  const restaurantName = isNom ? "nom" : survey?.campaign.restaurant_name ?? "your restaurant";

  function chooseRating(value: RatingValue) {
    setRating(value);
    setError("");
    setStep("details");
  }

  function changeRating() {
    setError("");
    setStep("rate");
  }

  function trackReviewClick(destination: string) {
    setReviewClicks((current) => (current.includes(destination) ? current : [...current, destination]));
  }

  async function beginSession() {
    if (session || !survey) {
      return session;
    }
    const entrySurface = isDemo ? "heard_guest_demo" : "printed_flyer_qr";
    const created = await apiFetch<FeedbackSession>("/api/v1/feedback-sessions", {
      method: "POST",
      body: {
        token,
        channel: survey.link.channel || "flyer",
        metadata: {
          campaign_id: survey.campaign.id,
          campaign_type: "flyer_giveaway",
          entry_surface: entrySurface,
          lead_source: isDemo ? "heard_guest_demo" : undefined
        }
      }
    });
    setSession(created);
    return created;
  }

  function validateContact() {
    const email = guestEmail.trim();
    const phone = guestPhone.trim();
    if (!email && !phone) {
      return "Add a phone number or email so we know how to reach you.";
    }
    if (email && !isValidEmail(email)) {
      return "Enter a complete email address, like you@example.com.";
    }
    if (phone && !isValidPhone(phone)) {
      return "Enter a valid phone number with area code, or start international numbers with +.";
    }
    return "";
  }

  async function submitSurvey(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!survey || !rating) {
      setError("Choose a face before finishing.");
      return;
    }
    const contactError = validateContact();
    if (contactError) {
      setError(contactError);
      return;
    }
    if (!isFiveStar && comment.trim().length < 3) {
      setError("Tell us a little about what happened so the team can follow up well.");
      return;
    }

    setBusy(true);
    setError("");
    try {
      const activeSession = await beginSession();
      if (!activeSession) {
        throw new Error("Could not start survey session");
      }
      const response = await apiFetch<FeedbackResponse>("/api/v1/feedback-responses", {
        method: "POST",
        body: {
          feedback_session_id: activeSession.id,
          rating,
          comment,
          categories: selectedTags,
          guest_name: guestName.trim(),
          guest_phone: guestPhone.trim(),
          guest_email: guestEmail.trim(),
          wants_follow_up: rating < 5,
          contact_consent: true,
          marketing_consent: marketingConsent,
          metadata: {
            campaign_id: survey.campaign.id,
            campaign_type: "flyer_giveaway",
            entry_surface: isDemo ? "heard_guest_demo" : "printed_flyer_qr",
            follow_up_required: rating < 5,
            public_review_prompt: rating === 5,
            review_destinations_clicked: reviewClicks,
            incentive_entry: true,
            heard_sales_follow_up_requested: isDemo && heardFollowUp,
            lead_source: isDemo ? "heard_guest_demo" : undefined
          }
        }
      });
      setSubmitted(response);
      setStep("done");
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : "Could not submit survey");
    } finally {
      setBusy(false);
    }
  }

  return (
    <main className="min-h-screen bg-[radial-gradient(circle_at_top,#f9f4e7_0,#eef4ef_48%,#e6efec_100%)] px-4 py-5 text-[#19383a] sm:py-8">
      <div className="mx-auto max-w-2xl">
        <header className="mb-5 flex items-center justify-between gap-4 px-1">
          <div className="flex min-w-0 items-center gap-3">
            {isNom ? (
              <img alt="nom" className="h-14 w-14 shrink-0 rounded-full shadow-[0_8px_24px_rgba(63,105,107,0.2)]" height="56" src="/brands/nom/logo.png" width="56" />
            ) : (
              <span className="grid h-12 w-12 shrink-0 place-items-center rounded-full bg-[#416a6c] font-display text-xl font-bold lowercase text-[#f8f0dc]">{restaurantName.slice(0, 1)}</span>
            )}
            <div className="min-w-0">
              <p className="truncate font-body text-sm font-semibold lowercase tracking-[0.08em] text-[#416a6c]">{restaurantName}</p>
              <h1 className="truncate font-display text-2xl font-semibold tracking-[-0.04em] text-[#19383a]">{survey?.campaign.headline ?? "How did we do?"}</h1>
            </div>
          </div>
          <Link className="shrink-0 rounded-full border border-[#416a6c]/20 bg-white/70 px-4 py-2 font-body text-sm font-semibold lowercase text-[#416a6c] backdrop-blur transition hover:border-[#416a6c]/45" href="/">
            heard<span className="text-[#ed6c5b]">.</span>
          </Link>
        </header>

        <section className="overflow-hidden rounded-[2rem] border border-white/70 bg-white/88 shadow-[0_28px_80px_rgba(38,77,78,0.15)] backdrop-blur">
          {step === "rate" ? (
            <div className="p-5 sm:p-8">
              <p className="font-body text-base leading-7 text-[#496567]">{survey?.campaign.prompt ?? "Tap the face that matches your visit."}</p>
              <div className="mt-7 grid grid-cols-3 gap-2 sm:grid-cols-5 sm:gap-3">
                {ratings.map((item) => (
                  <button
                    aria-label={`${item.label}, ${item.value} out of 5`}
                    className={`group min-w-0 rounded-2xl border px-2 py-3 transition sm:rounded-[1.4rem] ${rating === item.value ? "border-[#416a6c] bg-[#e7f0ec] ring-4 ring-[#416a6c]/10" : "border-[#dce8e3] bg-[#fbfaf5] hover:-translate-y-1 hover:border-[#416a6c]/45 hover:bg-white"}`}
                    key={item.value}
                    onClick={() => chooseRating(item.value)}
                    type="button"
                  >
                    <RatingFace className="mx-auto w-full max-w-[4.8rem] overflow-visible drop-shadow-[0_7px_8px_rgba(25,56,58,0.12)]" rating={item.value} />
                    <span className="mt-1 block min-h-8 font-body text-[11px] font-semibold leading-4 text-[#496567] sm:min-h-0 sm:text-xs">{item.label}</span>
                  </button>
                ))}
              </div>
              <p className="mt-7 rounded-2xl bg-[#f8f0dc] px-4 py-3 font-body text-sm font-semibold leading-6 text-[#6c5844]">{survey?.campaign.incentive_text}</p>
            </div>
          ) : null}

          {step === "details" && selectedRating ? (
            <form noValidate onSubmit={submitSurvey}>
              <div className="border-b border-[#dce8e3] bg-[#f8f0dc]/65 px-5 py-5 sm:px-8">
                <button className="inline-flex items-center gap-2 font-body text-sm font-semibold text-[#416a6c] transition hover:text-[#19383a]" onClick={changeRating} type="button">
                  <span aria-hidden="true">←</span> Change my rating
                </button>
              </div>

              <div className="p-5 sm:p-8">
                <div className="grid items-center gap-5 sm:grid-cols-[7rem_1fr]">
                  <RatingFace className="mx-auto w-28 overflow-visible drop-shadow-[0_12px_14px_rgba(25,56,58,0.14)]" rating={selectedRating.value} />
                  <div className="text-center sm:text-left">
                    <p className="font-body text-xs font-bold uppercase tracking-[0.2em] text-[#ed6c5b]">Thank you for rating us</p>
                    <h2 className="mt-2 font-display text-3xl font-semibold tracking-[-0.045em] text-[#19383a]">{selectedRating.label}.</h2>
                    <p className="mt-2 font-body text-sm leading-6 text-[#597173]">{selectedRating.prompt}</p>
                  </div>
                </div>

                {isFiveStar ? (
                  <div className="mt-7 rounded-[1.5rem] border border-[#dce8e3] bg-[#f5faf7] p-5">
                    <h3 className="font-display text-xl font-semibold tracking-[-0.025em]">Would you share the love?</h3>
                    <p className="mt-1 font-body text-sm leading-6 text-[#597173]">A public review helps more people discover {restaurantName}. This is optional, and you can finish below without leaving a review.</p>
                    <div className="mt-4 grid gap-3 sm:grid-cols-2">
                      {survey?.campaign.google_review_url ? (
                        <a className="rounded-full bg-[#416a6c] px-4 py-3 text-center font-body text-sm font-semibold text-white transition hover:bg-[#315759]" href={survey.campaign.google_review_url} onClick={() => trackReviewClick("google")} rel="noreferrer" target="_blank">
                          Review on Google
                        </a>
                      ) : null}
                      {survey?.campaign.yelp_review_url ? (
                        <a className="rounded-full border border-[#416a6c]/25 bg-white px-4 py-3 text-center font-body text-sm font-semibold text-[#416a6c] transition hover:border-[#416a6c]/55" href={survey.campaign.yelp_review_url} onClick={() => trackReviewClick("yelp")} rel="noreferrer" target="_blank">
                          Review on Yelp
                        </a>
                      ) : null}
                    </div>
                  </div>
                ) : (
                  <div className="mt-7">
                    <label className="block">
                      <span className="mb-2 block font-body text-sm font-semibold text-[#294b4d]">What would have made your visit better?</span>
                      <textarea className="min-h-28 w-full rounded-2xl border border-[#cbdcd6] bg-[#fbfdfb] px-4 py-3 font-body text-base outline-none transition placeholder:text-[#7b9391] focus:border-[#416a6c] focus:ring-4 focus:ring-[#416a6c]/10" onChange={(event) => setComment(event.target.value)} placeholder="Tell the team what happened..." value={comment} />
                    </label>
                    <div className="mt-4 flex flex-wrap gap-2">
                      {issueTags.map((tag) => (
                        <button
                          className={`rounded-full border px-4 py-2 font-body text-sm font-semibold transition ${selectedTags.includes(tag) ? "border-[#416a6c] bg-[#e4efeb] text-[#315759]" : "border-[#cbdcd6] bg-white text-[#597173] hover:border-[#416a6c]/50"}`}
                          key={tag}
                          onClick={() => setSelectedTags((current) => (current.includes(tag) ? current.filter((item) => item !== tag) : [...current, tag]))}
                          type="button"
                        >
                          {tag}
                        </button>
                      ))}
                    </div>
                  </div>
                )}

                <div className="mt-8 border-t border-[#dce8e3] pt-7">
                  <p className="font-body text-xs font-bold uppercase tracking-[0.2em] text-[#ed6c5b]">One last thing</p>
                  <h3 className="mt-2 font-display text-2xl font-semibold tracking-[-0.035em]">Where should we reach you?</h3>
                  <p className="mt-2 font-body text-sm leading-6 text-[#597173]">Add an email or phone number for the giveaway and any follow-up about your experience.</p>

                  <div className="mt-5 grid gap-3 sm:grid-cols-2">
                    <Field autoComplete="name" label="Name" onChange={setGuestName} placeholder="Your name" value={guestName} />
                    <Field autoComplete="tel" inputMode="tel" label="Phone" onChange={setGuestPhone} placeholder="(206) 555-0142" value={guestPhone} />
                    <Field autoComplete="email" className="sm:col-span-2" inputMode="email" label="Email" onChange={setGuestEmail} placeholder="you@example.com" type="email" value={guestEmail} />
                  </div>

                  {isDemo ? (
                    <div className="mt-4 rounded-2xl border border-[#ed6c5b]/20 bg-[#fff7f2] p-4">
                      <p className="font-body text-sm leading-6 text-[#6f5148]">Because this is the heard demo, the contact details you submit are saved so we can understand who tried the experience.</p>
                      <label className="mt-3 flex gap-3 font-body text-sm leading-6 text-[#4f4944]">
                        <input checked={heardFollowUp} className="mt-1 h-4 w-4 accent-[#416a6c]" onChange={(event) => setHeardFollowUp(event.target.checked)} type="checkbox" />
                        I&apos;d like the heard team to follow up about using this at my restaurant. <span className="font-semibold">Optional.</span>
                      </label>
                    </div>
                  ) : (
                    <label className="mt-4 flex gap-3 rounded-2xl bg-[#f2f6f4] px-4 py-3 font-body text-sm leading-6 text-[#425e60]">
                      <input checked={marketingConsent} className="mt-1 h-4 w-4 accent-[#416a6c]" onChange={(event) => setMarketingConsent(event.target.checked)} type="checkbox" />
                      Send me future offers from {restaurantName}.
                    </label>
                  )}

                  {error ? <div aria-live="assertive" className="mt-4 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 font-body text-sm text-red-700" role="alert">{error}</div> : null}

                  <button className="mt-5 h-[3.25rem] w-full rounded-full bg-[#416a6c] px-5 py-3.5 font-body text-sm font-bold text-white shadow-[0_12px_26px_rgba(65,106,108,0.24)] transition hover:-translate-y-0.5 hover:bg-[#315759] disabled:cursor-not-allowed disabled:opacity-60" disabled={busy} type="submit">
                    {busy ? "Saving your response..." : isDemo ? "Finish the demo" : "Finish survey"}
                  </button>
                </div>
              </div>
            </form>
          ) : null}

          {step === "done" && submitted && selectedRating ? (
            <div className="p-7 text-center sm:p-10">
              <RatingFace className="mx-auto w-32 overflow-visible drop-shadow-[0_14px_16px_rgba(25,56,58,0.16)]" rating={selectedRating.value} />
              <p className="mt-5 font-body text-xs font-bold uppercase tracking-[0.2em] text-[#ed6c5b]">Response received</p>
              <h2 className="mt-3 font-display text-3xl font-semibold tracking-[-0.045em]">Thank you for rating us {selectedRating.label.toLowerCase()}.</h2>
              <p className="mx-auto mt-3 max-w-md font-body text-sm leading-7 text-[#597173]">
                {isDemo
                  ? heardFollowUp
                    ? "Your demo response and contact details are saved. The heard team can follow up about bringing this experience to your restaurant."
                    : "Your demo response and contact details are saved. You can explore the restaurant side of heard whenever you are ready."
                  : rating < 5
                    ? `The ${restaurantName} team has your note and can follow up quickly.`
                    : `You are entered. Thanks for helping other guests discover ${restaurantName}.`}
              </p>
              {isDemo ? (
                <Link className="mt-7 inline-flex rounded-full bg-[#416a6c] px-6 py-3 font-body text-sm font-bold lowercase text-white transition hover:bg-[#315759]" href={"/start?source=guest_demo" as Route}>
                  start using heard
                </Link>
              ) : null}
            </div>
          ) : null}
        </section>
      </div>
    </main>
  );
}

function Field({
  autoComplete,
  className = "",
  inputMode,
  label,
  onChange,
  placeholder,
  type = "text",
  value
}: {
  autoComplete: string;
  className?: string;
  inputMode?: "email" | "tel";
  label: string;
  onChange: (value: string) => void;
  placeholder: string;
  type?: "email" | "text";
  value: string;
}) {
  return (
    <label className={`block ${className}`}>
      <span className="mb-2 block font-body text-sm font-semibold text-[#294b4d]">{label}</span>
      <input
        autoComplete={autoComplete}
        className="h-12 w-full rounded-2xl border border-[#cbdcd6] bg-[#fbfdfb] px-4 font-body text-base outline-none transition placeholder:text-[#7b9391] focus:border-[#416a6c] focus:ring-4 focus:ring-[#416a6c]/10"
        inputMode={inputMode}
        onChange={(event) => onChange(event.target.value)}
        placeholder={placeholder}
        type={type}
        value={value}
      />
    </label>
  );
}
