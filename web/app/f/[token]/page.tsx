"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import { apiFetch, FeedbackResponse, FeedbackSession, PublicSurvey } from "../../../lib/api";

const ratings = [
  { value: 1, face: "😡", label: "Terrible" },
  { value: 2, face: "🙁", label: "Bad" },
  { value: 3, face: "😐", label: "Okay" },
  { value: 4, face: "🙂", label: "Good" },
  { value: 5, face: "😍", label: "Loved it" }
];

const issueTags = ["Food", "Speed", "Service", "Accuracy", "Packaging", "Value"];

export default function FlyerSurveyPage({ params }: { params: Promise<{ token: string }> }) {
  const [token, setToken] = useState("");
  const [survey, setSurvey] = useState<PublicSurvey | null>(null);
  const [session, setSession] = useState<FeedbackSession | null>(null);
  const [submitted, setSubmitted] = useState<FeedbackResponse | null>(null);
  const [rating, setRating] = useState(0);
  const [comment, setComment] = useState("");
  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  const [guestName, setGuestName] = useState("");
  const [guestPhone, setGuestPhone] = useState("");
  const [guestEmail, setGuestEmail] = useState("");
  const [marketingConsent, setMarketingConsent] = useState(false);
  const [reviewClicks, setReviewClicks] = useState<string[]>([]);
  const [step, setStep] = useState<"rate" | "review" | "details" | "done">("rate");
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
  const needsContact = guestPhone.trim() || guestEmail.trim();

  function chooseRating(value: number) {
    setRating(value);
    setError("");
    setStep(value === 5 ? "review" : "details");
  }

  function trackReviewClick(destination: string) {
    setReviewClicks((current) => (current.includes(destination) ? current : [...current, destination]));
  }

  async function beginSession() {
    if (session || !survey) {
      return session;
    }
    const created = await apiFetch<FeedbackSession>("/api/v1/feedback-sessions", {
      method: "POST",
      body: {
        token,
        channel: survey.link.channel || "flyer",
        metadata: {
          campaign_id: survey.campaign.id,
          campaign_type: "flyer_giveaway",
          entry_surface: "printed_flyer_qr"
        }
      }
    });
    setSession(created);
    return created;
  }

  async function submitSurvey() {
    if (!survey) {
      return;
    }
    if (!rating) {
      setError("Choose a face before finishing.");
      return;
    }
    if (!needsContact) {
      setError("Add a phone number or email so the restaurant can reach you if you win.");
      return;
    }
    if (!isFiveStar && comment.trim().length < 3) {
      setError("Tell the restaurant what happened so they can follow up well.");
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
          guest_name: guestName,
          guest_phone: guestPhone,
          guest_email: guestEmail,
          wants_follow_up: rating < 5,
          contact_consent: true,
          marketing_consent: marketingConsent,
          metadata: {
            campaign_id: survey.campaign.id,
            campaign_type: "flyer_giveaway",
            follow_up_required: rating < 5,
            public_review_prompt: rating === 5,
            review_destinations_clicked: reviewClicks,
            incentive_entry: true
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
    <main className="min-h-screen bg-[#f7f9fc] px-4 py-5 text-[#14213d]">
      <div className="mx-auto max-w-xl">
        <header className="mb-5 flex items-center justify-between">
          <div>
            <p className="text-sm font-semibold text-[#2f80ed]">{survey?.campaign.restaurant_name ?? "Heard"}</p>
            <h1 className="text-2xl font-bold tracking-normal">{survey?.campaign.headline ?? "How did we do?"}</h1>
          </div>
          <Link className="rounded-md border border-[#d0d7e2] bg-white px-3 py-2 text-sm font-medium text-[#475467]" href="/">
            Heard
          </Link>
        </header>

        <section className="rounded-lg border border-[#d8dee9] bg-white p-5 shadow-sm">
          {step === "rate" ? (
            <div>
              <p className="text-base text-[#475467]">{survey?.campaign.prompt ?? "Tap the face that matches your visit."}</p>
              <div className="mt-6 grid grid-cols-5 gap-2">
                {ratings.map((item) => (
                  <button
                    aria-label={`${item.value} out of 5, ${item.label}`}
                    className="grid aspect-square place-items-center rounded-md border border-[#d8dee9] bg-[#fffaf5] text-4xl transition hover:border-[#2f80ed] hover:bg-[#eef4ff] focus:outline-none focus:ring-2 focus:ring-[#2f80ed]"
                    key={item.value}
                    onClick={() => chooseRating(item.value)}
                    type="button"
                  >
                    {item.face}
                  </button>
                ))}
              </div>
              <p className="mt-6 rounded-md bg-[#fff7ed] px-4 py-3 text-sm font-medium text-[#9a3412]">{survey?.campaign.incentive_text}</p>
            </div>
          ) : null}

          {step === "review" ? (
            <div>
              <p className="text-5xl">{selectedRating?.face}</p>
              <h2 className="mt-3 text-2xl font-bold tracking-normal">Glad we earned a 5.</h2>
              <p className="mt-2 text-sm leading-6 text-[#475467]">A public review helps the restaurant a ton. Open Google or Yelp, then come back here to enter the gift card drawing.</p>
              <div className="mt-5 grid gap-3 sm:grid-cols-2">
                {survey?.campaign.google_review_url ? (
                  <a className="rounded-md bg-[#14213d] px-4 py-3 text-center text-sm font-semibold text-white" href={survey.campaign.google_review_url} onClick={() => trackReviewClick("google")} target="_blank">
                    Review on Google
                  </a>
                ) : null}
                {survey?.campaign.yelp_review_url ? (
                  <a className="rounded-md border border-[#d0d7e2] px-4 py-3 text-center text-sm font-semibold text-[#14213d]" href={survey.campaign.yelp_review_url} onClick={() => trackReviewClick("yelp")} target="_blank">
                    Review on Yelp
                  </a>
                ) : null}
              </div>
              <button className="mt-5 w-full rounded-md bg-[#f25f4c] px-4 py-3 text-sm font-semibold text-white" onClick={() => setStep("details")} type="button">
                Continue to gift card entry
              </button>
            </div>
          ) : null}

          {step === "details" ? (
            <div>
              <p className="text-5xl">{selectedRating?.face}</p>
              <h2 className="mt-3 text-2xl font-bold tracking-normal">{isFiveStar ? "Enter the gift card drawing" : "Tell us what happened"}</h2>
              {!isFiveStar ? (
                <>
                  <textarea className="mt-5 min-h-28 w-full rounded-md border border-[#d0d7e2] px-3 py-3 text-base outline-none focus:border-[#2f80ed]" onChange={(event) => setComment(event.target.value)} placeholder="What could we have done better?" value={comment} />
                  <div className="mt-4 flex flex-wrap gap-2">
                    {issueTags.map((tag) => (
                      <button
                        className={`rounded-md border px-3 py-2 text-sm font-medium ${selectedTags.includes(tag) ? "border-[#2f80ed] bg-[#eef4ff] text-[#175cd3]" : "border-[#d0d7e2] bg-white text-[#475467]"}`}
                        key={tag}
                        onClick={() => setSelectedTags((current) => (current.includes(tag) ? current.filter((item) => item !== tag) : [...current, tag]))}
                        type="button"
                      >
                        {tag}
                      </button>
                    ))}
                  </div>
                </>
              ) : null}

              <div className="mt-5 grid gap-3 sm:grid-cols-2">
                <input className="h-12 rounded-md border border-[#d0d7e2] px-3 text-base outline-none focus:border-[#2f80ed]" onChange={(event) => setGuestName(event.target.value)} placeholder="Name" value={guestName} />
                <input className="h-12 rounded-md border border-[#d0d7e2] px-3 text-base outline-none focus:border-[#2f80ed]" inputMode="tel" onChange={(event) => setGuestPhone(event.target.value)} placeholder="Phone" value={guestPhone} />
                <input className="h-12 rounded-md border border-[#d0d7e2] px-3 text-base outline-none focus:border-[#2f80ed] sm:col-span-2" inputMode="email" onChange={(event) => setGuestEmail(event.target.value)} placeholder="Email" value={guestEmail} />
              </div>

              <label className="mt-4 flex gap-3 rounded-md bg-[#f2f4f7] px-3 py-3 text-sm leading-6 text-[#344054]">
                <input checked={marketingConsent} className="mt-1" onChange={(event) => setMarketingConsent(event.target.checked)} type="checkbox" />
                Send me future offers from this restaurant.
              </label>

              {error ? <div className="mt-4 rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{error}</div> : null}

              <button className="mt-5 h-12 w-full rounded-md bg-[#14213d] px-4 text-sm font-semibold text-white disabled:cursor-not-allowed disabled:opacity-60" disabled={busy} onClick={() => void submitSurvey()} type="button">
                {busy ? "Submitting..." : "Finish survey"}
              </button>
            </div>
          ) : null}

          {step === "done" && submitted ? (
            <div>
              <p className="text-5xl">{rating === 5 ? "😍" : "✓"}</p>
              <h2 className="mt-3 text-2xl font-bold tracking-normal">You are entered.</h2>
              <p className="mt-2 text-sm leading-6 text-[#475467]">
                {rating < 5 ? "The restaurant has your note and can follow up quickly." : "Thanks for helping other guests find a good meal."}
              </p>
            </div>
          ) : null}
        </section>
      </div>
    </main>
  );
}
