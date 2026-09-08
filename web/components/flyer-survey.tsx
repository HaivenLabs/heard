"use client";

import Link from "next/link";
import type { Route } from "next";
import { FormEvent, useEffect, useMemo, useState } from "react";
import { RatingFace, RatingValue } from "./rating-face";
import { HeardLogo } from "./heard-logo";
import { apiFetch, FeedbackResponse, FeedbackSession, PublicSurvey } from "../lib/api";
import { isValidEmail, isValidPhone } from "../lib/contact-validation";

const ratings: Array<{ value: RatingValue; label: string; prompt: string }> = [
  { value: 1, label: "Not good", prompt: "We are sorry this missed the mark." },
  { value: 2, label: "Could be better", prompt: "Thank you for helping us improve." },
  { value: 3, label: "It was okay", prompt: "We would love to make the next visit better." },
  { value: 4, label: "Really good", prompt: "We are glad it was a good visit." },
  { value: 5, label: "Loved it", prompt: "We love hearing that." }
];

const issueTags = ["Food", "Speed", "Service", "Accuracy", "Packaging", "Value"];

export type FlyerSurveyResolve =
  | { kind: "token"; token: string }
  | { kind: "path"; handle: string; slug: string };

export function FlyerSurvey({ resolve }: { resolve: FlyerSurveyResolve }) {
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

  const surveyEndpoint = useMemo(() => {
    if (resolve.kind === "path") {
      return `/api/v1/public/surveys/by-path/${resolve.handle}/${resolve.slug}`;
    }
    return `/api/v1/public/surveys/${resolve.token}`;
  }, [resolve]);

  useEffect(() => {
    setSurvey(null);
    setSession(null);
    setSubmitted(null);
    setRating(0);
    setComment("");
    setSelectedTags([]);
    setGuestName("");
    setGuestPhone("");
    setGuestEmail("");
    setMarketingConsent(false);
    setHeardFollowUp(false);
    setReviewClicks([]);
    setStep("rate");
    setBusy(false);
    setError("");
    void apiFetch<PublicSurvey>(surveyEndpoint)
      .then((payload) => setSurvey(payload))
      .catch((caught) => setError(caught instanceof Error ? caught.message : "This survey link is not available."));
  }, [surveyEndpoint]);

  const selectedRating = useMemo(() => ratings.find((item) => item.value === rating), [rating]);
  const isFiveStar = rating === 5;
  const isDemo = resolve.kind === "token" && resolve.token === "demo-heard";
  const isNom = (survey?.campaign.restaurant_name ?? "nom").toLowerCase() === "nom";
  const restaurantName = isNom ? "nom" : survey?.campaign.restaurant_name ?? "your restaurant";
  const googleReviewUrl = survey?.campaign.google_review_url || `https://www.google.com/search?q=${encodeURIComponent(restaurantName + " reviews")}`;
  const yelpReviewUrl = survey?.campaign.yelp_review_url || `https://www.yelp.com/search?find_desc=${encodeURIComponent(restaurantName)}`;

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
        token: survey.link.token,
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
    <main className="survey-canvas min-h-screen px-4 py-5 text-ink sm:py-8">
      <div className="mx-auto max-w-2xl">
        <header className="mb-5 flex items-center justify-between gap-4 px-1">
          <div className="flex min-w-0 items-center gap-3">
            {survey?.campaign.logo_url || isNom ? (
              <img alt={restaurantName} className="h-14 w-14 shrink-0 rounded-full object-cover shadow-soft" height="56" src={survey?.campaign.logo_url || "/brands/nom/logo.png"} width="56" />
            ) : (
              <span className="grid h-12 w-12 shrink-0 place-items-center rounded-full bg-primary font-display text-xl font-bold lowercase text-parchment">{restaurantName.slice(0, 1)}</span>
            )}
            <div className="min-w-0">
              <p className="truncate font-body text-sm font-semibold lowercase tracking-[0.08em] text-primary">{restaurantName}</p>
              <h1 className="truncate font-display text-2xl font-semibold tracking-[-0.04em] text-ink">{survey?.campaign.headline ?? "How did we do?"}</h1>
            </div>
          </div>
          <Link aria-label="heard home" className="shrink-0 rounded-full border border-primary/20 bg-white/70 px-3 py-1.5 backdrop-blur transition hover:border-primary/45" href="/"><HeardLogo className="scale-75 origin-right" /></Link>
        </header>

        <section className="overflow-hidden rounded-[2rem] border border-white/70 bg-white/88 shadow-soft backdrop-blur">
          {step === "rate" ? (
            <div className="p-5 sm:p-8">
              <p className="font-body text-base leading-7 text-ink/65">{survey?.campaign.prompt ?? "Tap the face that matches your visit."}</p>
              <div className="mt-7 grid grid-cols-3 gap-2 sm:grid-cols-5 sm:gap-3">
                {ratings.map((item) => (
                  <button
                    aria-label={`${item.label}, ${item.value} out of 5`}
                    className={`group min-w-0 rounded-2xl border px-2 py-3 transition sm:rounded-[1.4rem] ${rating === item.value ? "border-primary bg-mist ring-4 ring-primary/10" : "border-sand bg-parchment hover:-translate-y-1 hover:border-primary/45 hover:bg-white"}`}
                    key={item.value}
                    onClick={() => chooseRating(item.value)}
                    type="button"
                  >
                    <RatingFace className="mx-auto w-full max-w-[4.8rem] overflow-visible drop-shadow-sm" faceSet={survey?.campaign?.rating_face_set ?? "heard"} rating={item.value} />
                    <span className="mt-1 block min-h-8 font-body text-[11px] font-semibold leading-4 text-ink/65 sm:min-h-0 sm:text-xs">{item.label}</span>
                  </button>
                ))}
              </div>
              <p className="mt-7 rounded-2xl bg-parchment px-4 py-3 font-body text-sm font-semibold leading-6 text-ink">{survey?.campaign.incentive_text}</p>
            </div>
          ) : null}

          {step === "details" && selectedRating ? (
            <form noValidate onSubmit={submitSurvey}>
              <div className="border-b border-sand bg-parchment/65 px-5 py-5 sm:px-8">
                <button className="inline-flex items-center gap-2 font-body text-sm font-semibold text-primary transition hover:text-spruce" onClick={changeRating} type="button">
                  <span aria-hidden="true">←</span> Change my rating
                </button>
              </div>

              <div className="p-5 sm:p-8">
                <div className="grid items-center gap-5 sm:grid-cols-[7rem_1fr]">
                  <RatingFace className="mx-auto w-28 overflow-visible drop-shadow-md" faceSet={survey?.campaign?.rating_face_set ?? "heard"} rating={selectedRating.value} />
                  <div className="text-center sm:text-left">
                    <p className="font-body text-xs font-bold uppercase tracking-[0.2em] text-spruce">Thank you for rating us</p>
                    <h2 className="mt-2 font-display text-3xl font-semibold tracking-[-0.045em] text-ink">{selectedRating.label}.</h2>
                    <p className="mt-2 font-body text-sm leading-6 text-ink/60">{selectedRating.prompt}</p>
                  </div>
                </div>

                {isFiveStar ? (
                  <div className="mt-7 rounded-[1.5rem] border border-sand bg-surface p-5">
                    <h3 className="font-display text-xl font-semibold tracking-[-0.025em]">Would you share the love?</h3>
                    <p className="mt-1 font-body text-sm leading-6 text-ink/60">A public review helps more people discover {restaurantName}. This is optional, and you can finish below without leaving a review.</p>
                    <div className="mt-4 grid gap-3 sm:grid-cols-2">
                      <a className="rounded-full bg-primary px-4 py-3 text-center font-body text-sm font-semibold text-white transition hover:bg-spruce" href={googleReviewUrl} onClick={() => trackReviewClick("google")} rel="noreferrer" target="_blank">
                        Review on Google
                      </a>
                      <a className="rounded-full border border-primary/25 bg-white px-4 py-3 text-center font-body text-sm font-semibold text-primary transition hover:border-primary/55" href={yelpReviewUrl} onClick={() => trackReviewClick("yelp")} rel="noreferrer" target="_blank">
                        Review on Yelp
                      </a>
                    </div>
                  </div>
                ) : (
                  <div className="mt-7">
                    <label className="block">
                      <span className="mb-2 block font-body text-sm font-semibold text-ink">What would have made your visit better?</span>
                      <textarea className="min-h-28 w-full rounded-2xl border border-sand bg-surface px-4 py-3 font-body text-base outline-none transition placeholder:text-ink/45 focus:border-primary focus:ring-4 focus:ring-primary/10" onChange={(event) => setComment(event.target.value)} placeholder="Tell the team what happened..." value={comment} />
                    </label>
                    <div className="mt-4 flex flex-wrap gap-2">
                      {issueTags.map((tag) => (
                        <button
                          className={`rounded-full border px-4 py-2 font-body text-sm font-semibold transition ${selectedTags.includes(tag) ? "border-primary bg-mist text-spruce" : "border-sand bg-white text-ink/60 hover:border-primary/50"}`}
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

                <div className="mt-8 border-t border-sand pt-7">
                  <p className="font-body text-xs font-bold uppercase tracking-[0.2em] text-spruce">One last thing</p>
                  <h3 className="mt-2 font-display text-2xl font-semibold tracking-[-0.035em]">Where should we reach you?</h3>
                  <p className="mt-2 font-body text-sm leading-6 text-ink/60">Add an email or phone number for the giveaway and any follow-up about your experience.</p>

                  <div className="mt-5 grid gap-3 sm:grid-cols-2">
                    <Field autoComplete="name" label="Name" onChange={setGuestName} placeholder="Your name" value={guestName} />
                    <Field autoComplete="tel" inputMode="tel" label="Phone" onChange={setGuestPhone} placeholder="(206) 555-0142" value={guestPhone} />
                    <Field autoComplete="email" className="sm:col-span-2" inputMode="email" label="Email" onChange={setGuestEmail} placeholder="you@example.com" type="email" value={guestEmail} />
                  </div>

                  {isDemo ? (
                    <div className="mt-4 rounded-2xl border border-red-400/20 bg-red-50 p-4">
                      <p className="font-body text-sm leading-6 text-ink/70">Because this is the heard demo, the contact details you submit are saved so we can understand who tried the experience.</p>
                      <label className="mt-3 flex gap-3 font-body text-sm leading-6 text-ink/70">
                        <input checked={heardFollowUp} className="mt-1 h-4 w-4 accent-primary" onChange={(event) => setHeardFollowUp(event.target.checked)} type="checkbox" />
                        I&apos;d like the heard team to follow up about using this at my restaurant. <span className="font-semibold">Optional.</span>
                      </label>
                    </div>
                  ) : (
                    <label className="mt-4 flex gap-3 rounded-2xl bg-parchment px-4 py-3 font-body text-sm leading-6 text-ink/65">
                      <input checked={marketingConsent} className="mt-1 h-4 w-4 accent-primary" onChange={(event) => setMarketingConsent(event.target.checked)} type="checkbox" />
                      Send me future offers from {restaurantName}.
                    </label>
                  )}

                  {error ? <div aria-live="assertive" className="mt-4 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 font-body text-sm text-red-700" role="alert">{error}</div> : null}

                  <button className="mt-5 h-[3.25rem] w-full rounded-full bg-primary px-5 py-3.5 font-body text-sm font-bold text-white shadow-soft transition hover:-translate-y-0.5 hover:bg-spruce disabled:cursor-not-allowed disabled:opacity-60" disabled={busy} type="submit">
                    {busy ? "Saving your response..." : isDemo ? "Finish the demo" : "Finish survey"}
                  </button>
                </div>
              </div>
            </form>
          ) : null}

          {step === "done" && submitted && selectedRating ? (
            <div className="p-7 text-center sm:p-10">
              <RatingFace className="mx-auto w-32 overflow-visible drop-shadow-md" faceSet={survey?.campaign?.rating_face_set ?? "heard"} rating={selectedRating.value} />
              <p className="mt-5 font-body text-xs font-bold uppercase tracking-[0.2em] text-spruce">Response received</p>
              <h2 className="mt-3 font-display text-3xl font-semibold tracking-[-0.045em]">Thank you for rating us {selectedRating.label.toLowerCase()}.</h2>
              <p className="mx-auto mt-3 max-w-md font-body text-sm leading-7 text-ink/60">
                {isDemo
                  ? heardFollowUp
                    ? "Your demo response and contact details are saved. The heard team can follow up about bringing this experience to your restaurant."
                    : "Your demo response and contact details are saved. You can explore the restaurant side of heard whenever you are ready."
                  : rating < 5
                    ? `The ${restaurantName} team has your note and can follow up quickly.`
                    : `You are entered. Thanks for helping other guests discover ${restaurantName}.`}
              </p>
              {isDemo ? (
                <Link className="mt-7 inline-flex rounded-full bg-primary px-6 py-3 font-body text-sm font-bold lowercase text-white transition hover:bg-spruce" href={"/start?source=guest_demo" as Route}>
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
      <span className="mb-2 block font-body text-sm font-semibold text-ink">{label}</span>
      <input
        autoComplete={autoComplete}
        className="h-12 w-full rounded-2xl border border-sand bg-surface px-4 font-body text-base outline-none transition placeholder:text-ink/45 focus:border-primary focus:ring-4 focus:ring-primary/10"
        inputMode={inputMode}
        onChange={(event) => onChange(event.target.value)}
        placeholder={placeholder}
        type={type}
        value={value}
      />
    </label>
  );
}
