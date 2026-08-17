import Link from "next/link";
import type { Route } from "next";
import { BrandBackdrop } from "../components/brand-backdrop";
import { PublicHeader } from "../components/public-header";
import { RatingFace, RatingValue } from "../components/rating-face";

const ratingValues: RatingValue[] = [1, 2, 3, 4, 5];

export default function HomePage() {
  return (
    <main className="relative overflow-hidden bg-[#f7f1e6] text-ink">
      <BrandBackdrop />
      <section className="relative min-h-[94vh] border-b border-ink/10">
        <PublicHeader />

        <div className="relative mx-auto grid max-w-7xl items-center gap-14 px-6 pb-20 pt-16 lg:grid-cols-[1.06fr_0.8fr] lg:pt-24">
          <div>
            <h1 className="max-w-4xl font-display text-6xl leading-[0.96] tracking-[-0.065em] sm:text-7xl lg:text-[6.4rem]">The guest experience, in your hands</h1>
            <p className="mt-7 max-w-2xl font-display text-2xl leading-tight tracking-[-0.03em] text-clay sm:text-3xl">Fix the visit before it becomes the review.</p>
            <p className="mt-5 max-w-2xl font-body text-lg leading-8 text-ink/62">heard gives restaurants a guest feedback experience people actually finish, then turns every rough visit into a clear recovery task while the relationship can still be saved.</p>
            <div className="mt-10 flex flex-wrap items-center gap-4">
              <Link className="rounded-full bg-clay px-7 py-4 font-display text-sm font-semibold tracking-[0.06em] text-white shadow-[0_16px_36px_rgba(203,104,67,0.28)] transition hover:-translate-y-0.5 hover:bg-[#b95635]" href={"/contact" as Route}>See heard for your restaurant</Link>
              <Link className="rounded-full border border-ink/15 bg-white/45 px-7 py-4 font-body text-sm font-semibold transition hover:border-ink/35 hover:bg-white/75" href="/f/demo-heard">Try guest experience</Link>
            </div>
            <p className="mt-4 font-body text-sm text-ink/42">A restaurant-specific walkthrough. No generic sales maze.</p>
          </div>

          <div className="relative mx-auto w-full max-w-lg lg:mr-0">
            <div className="absolute -left-8 -top-8 h-full w-full rotate-[-3deg] rounded-[2.5rem] bg-olive/20" />
            <div className="relative rounded-[2.5rem] bg-[#17251d] p-7 text-parchment shadow-[0_40px_100px_rgba(23,37,29,0.28)] sm:p-9">
              <div className="flex items-start justify-between gap-4">
                <div>
                  <p className="font-body text-xs uppercase tracking-[0.22em] text-parchment/42">New guest feedback</p>
                  <p className="mt-2 font-display text-2xl">How did we do?</p>
                </div>
                <span className="rounded-full bg-[#9fba73]/15 px-3 py-1.5 font-body text-xs font-semibold text-[#c4d99d]">Just received</span>
              </div>
              <div className="mt-8 grid grid-cols-5 gap-2">
                {ratingValues.map((rating) => <RatingFace className={`w-full overflow-visible ${rating === 2 ? "drop-shadow-[0_0_12px_rgba(247,194,111,0.25)]" : "opacity-45"}`} key={rating} rating={rating} />)}
              </div>
              <div className="mt-7 rounded-[1.5rem] border border-clay/28 bg-clay/10 p-5">
                <div className="flex items-center justify-between gap-3">
                  <p className="font-body text-xs font-bold uppercase tracking-[0.18em] text-clay">Needs follow-up</p>
                  <span className="font-body text-xs text-parchment/45">Just now</span>
                </div>
                <p className="mt-3 font-display text-xl">Pickup took too long.</p>
                <p className="mt-2 font-body text-sm leading-6 text-parchment/62">The guest asked to hear from you. Service and speed are tagged so your team has the context to respond.</p>
                <div className="mt-4 flex gap-2">
                  <span className="rounded-full bg-parchment/8 px-3 py-1.5 font-body text-xs text-parchment/65">Speed</span>
                  <span className="rounded-full bg-parchment/8 px-3 py-1.5 font-body text-xs text-parchment/65">Takeout</span>
                </div>
                <div className="mt-5 flex items-center justify-between gap-4 border-t border-parchment/10 pt-4 font-body text-xs">
                  <span className="font-semibold text-parchment">Sent to Recovery inbox</span>
                  <span className="text-parchment/42">Ready for your team</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="mx-auto max-w-7xl px-6 py-24" id="how-it-works">
        <div className="grid gap-10 lg:grid-cols-[0.7fr_1fr] lg:gap-20">
          <div>
            <p className="font-body text-xs font-bold uppercase tracking-[0.26em] text-clay">From signal to save</p>
            <h2 className="mt-5 font-display text-5xl leading-[1.02] tracking-[-0.055em] sm:text-6xl">One scan. One honest answer. One clear next move.</h2>
          </div>
          <div className="grid gap-6 sm:grid-cols-3">
            <Promise number="01" title="Ask simply">A branded, mobile-first survey opens from a QR code or link. Guests do not need an app or account.</Promise>
            <Promise number="02" title="Route the signal">Feedback, contact details, and the operational issue arrive together instead of living in separate tools.</Promise>
            <Promise number="03" title="Recover personally">Your team gets a focused queue and enough context to respond like a human while the visit is still fresh.</Promise>
          </div>
        </div>
      </section>

      <section className="border-y border-ink/10 bg-[#e9e4d7]">
        <div className="mx-auto grid max-w-7xl gap-5 px-6 py-20 md:grid-cols-3">
          <Result eyebrow="Collect" title="Feedback guests will actually give">Keep the first interaction quick, visual, and unmistakably yours.</Result>
          <Result eyebrow="Recover" title="A human next step, not another metric">Know what happened, how to reach the guest, and who should act.</Result>
          <Result eyebrow="Improve" title="Patterns your operators can use">See recurring issues by location and turn them into focused changes.</Result>
        </div>
      </section>

      <section className="relative bg-[#17251d] px-6 py-24 text-parchment">
        <div className="pointer-events-none absolute inset-0 opacity-40 [background-image:radial-gradient(circle_at_1px_1px,rgba(247,241,227,0.13)_1px,transparent_0)] [background-size:30px_30px]" />
        <div className="relative mx-auto flex max-w-5xl flex-col items-center text-center">
          <p className="font-body text-xs font-bold uppercase tracking-[0.26em] text-clay">Make the next rough visit recoverable</p>
          <h2 className="mt-5 font-display text-5xl tracking-[-0.055em] sm:text-7xl">See heard with your restaurant in mind.</h2>
          <p className="mt-6 max-w-2xl font-body text-lg leading-8 text-parchment/62">Tell us how many locations you run and what you want to fix. We&apos;ll shape the walkthrough around that.</p>
          <Link className="mt-9 rounded-full bg-clay px-8 py-4 font-display text-sm font-semibold tracking-[0.06em] text-white transition hover:-translate-y-0.5 hover:bg-[#b95635]" href={"/contact" as Route}>Request a walkthrough</Link>
          <div className="mt-12 flex flex-wrap justify-center gap-5 font-body text-sm text-parchment/45">
            <Link className="hover:text-parchment" href="/f/demo-heard">Guest demo</Link>
            <Link className="hover:text-parchment" href="/login">Customer sign in</Link>
          </div>
        </div>
      </section>
    </main>
  );
}

function Promise({ children, number, title }: { children: React.ReactNode; number: string; title: string }) {
  return <article className="border-t border-ink/15 pt-5"><p className="font-body text-xs font-bold text-clay">{number}</p><h3 className="mt-4 font-display text-2xl tracking-[-0.035em]">{title}</h3><p className="mt-3 font-body text-sm leading-7 text-ink/55">{children}</p></article>;
}

function Result({ children, eyebrow, title }: { children: React.ReactNode; eyebrow: string; title: string }) {
  return <article className="rounded-[2rem] border border-ink/10 bg-[#f8f3e9] p-7 shadow-[0_20px_50px_rgba(17,24,39,0.06)]"><p className="font-body text-xs font-bold uppercase tracking-[0.2em] text-olive">{eyebrow}</p><h3 className="mt-4 font-display text-3xl leading-tight tracking-[-0.04em]">{title}</h3><p className="mt-4 font-body text-sm leading-7 text-ink/55">{children}</p></article>;
}
