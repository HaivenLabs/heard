import Link from "next/link";
import { DEMO_TENANT_ID } from "../lib/api";

export default function HomePage() {
  return (
    <main className="min-h-screen bg-ink text-parchment">
      <div className="pointer-events-none absolute inset-0 overflow-hidden">
        <div className="absolute left-[-10rem] top-[-6rem] h-80 w-80 rounded-full bg-clay/40 blur-3xl" />
        <div className="absolute right-[-10rem] top-24 h-96 w-96 rounded-full bg-olive/30 blur-3xl" />
      </div>
      <section className="relative mx-auto flex min-h-screen max-w-6xl flex-col justify-center gap-10 px-6 py-20">
        <p className="w-fit rounded-full border border-parchment/20 bg-parchment/10 px-4 py-2 font-body text-xs uppercase tracking-[0.32em] text-parchment/70">
          Heard local product
        </p>
        <div className="grid gap-10 lg:grid-cols-[1.3fr_0.9fr]">
          <div className="space-y-6">
            <h1 className="max-w-3xl font-display text-5xl leading-tight text-parchment sm:text-6xl">
              Frictionless feedback that turns rough nights into recovery cases.
            </h1>
            <p className="max-w-2xl font-body text-lg leading-8 text-parchment/75">
              This local-first slice wires tenant creation, location setup, QR-backed feedback links, a guest form,
              transactional outbox events, and a recovery inbox into one Docker-runnable loop.
            </p>
            <div className="flex flex-wrap gap-4">
              <Link className="rounded-full bg-parchment px-6 py-3 font-display text-sm uppercase tracking-[0.24em] text-ink transition hover:bg-white" href="/admin/campaigns">
                Build flyer campaign
              </Link>
              <Link className="rounded-full border border-parchment/25 px-6 py-3 font-display text-sm uppercase tracking-[0.24em] text-parchment transition hover:border-parchment/50 hover:bg-parchment/5" href="/admin/recovery">
                Recovery inbox
              </Link>
              <Link className="rounded-full border border-clay/60 bg-clay/10 px-6 py-3 font-display text-sm uppercase tracking-[0.24em] text-parchment transition hover:bg-clay/20" href="/f/demo-heard">
                Guest demo flow
              </Link>
            </div>
          </div>
          <div className="rounded-[2rem] border border-parchment/10 bg-parchment/10 p-8 shadow-soft backdrop-blur">
            <p className="font-display text-sm uppercase tracking-[0.3em] text-parchment/60">Fast start</p>
            <ol className="mt-6 space-y-4 font-body text-sm leading-7 text-parchment/80">
              <li>1. Run `docker compose up --build`.</li>
              <li>2. Open `/admin/campaigns` to create a printed flyer survey campaign.</li>
              <li>3. Open the generated `/f/[token]` route, submit a rating under 5, then watch `/admin/recovery` populate.</li>
            </ol>
            <div className="mt-8 rounded-[1.5rem] bg-ink/60 p-5">
              <p className="font-display text-xs uppercase tracking-[0.3em] text-parchment/50">Seeded local demo</p>
              <p className="mt-3 font-body text-sm leading-7 text-parchment/80">
                Tenant header default: <span className="font-mono text-xs">{DEMO_TENANT_ID}</span>
              </p>
              <p className="font-body text-sm leading-7 text-parchment/80">Guest token: <span className="font-mono text-xs">demo-heard</span></p>
            </div>
          </div>
        </div>
      </section>
    </main>
  );
}
