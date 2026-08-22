export type Tenant = {
  id: string;
  name: string;
  slug: string;
  created_at: string;
};

export type MarketingLead = {
  id: string;
  name: string;
  work_email: string;
  phone: string;
  restaurant_name: string;
  location_count: string;
  challenge: string;
  source: "marketing_site" | "guest_demo";
  contact_consent: true;
  status: "new";
  created_at: string;
};

export type Location = {
  id: string;
  tenant_id: string;
  name: string;
  slug: string;
  timezone: string;
  created_at: string;
};

export type FeedbackLink = {
  id: string;
  tenant_id: string;
  location_id: string;
  campaign_id?: string;
  name: string;
  token: string;
  slug?: string;
  status: string;
  channel: string;
  qr_asset_url: string;
  destination_url: string;
  created_at: string;
};

export type SurveyCampaign = {
  id: string;
  tenant_id: string;
  location_id: string;
  name: string;
  restaurant_name: string;
  headline: string;
  prompt: string;
  incentive_text: string;
  sms_keyword: string;
  sms_phone: string;
  google_review_url: string;
  yelp_review_url: string;
  logo_url?: string;
  theme?: string;
  status: string;
  created_at: string;
};

export type PublicSurvey = {
  link: FeedbackLink;
  campaign: SurveyCampaign;
};

export type FeedbackSession = {
  id: string;
  tenant_id: string;
  location_id: string;
  feedback_link_id?: string;
  status: string;
  source: string;
  channel: string;
  guest_name?: string;
  guest_phone?: string;
  guest_email?: string;
  wants_follow_up: boolean;
  contact_consent: boolean;
  marketing_consent: boolean;
  metadata: Record<string, unknown>;
  created_at: string;
};

export type FeedbackResponse = {
  id: string;
  tenant_id: string;
  location_id: string;
  feedback_session_id: string;
  feedback_link_id?: string;
  rating: number;
  sentiment: string;
  comment: string;
  categories: string[];
  guest_name?: string;
  guest_phone?: string;
  guest_email?: string;
  wants_follow_up: boolean;
  contact_consent: boolean;
  marketing_consent: boolean;
  metadata: Record<string, unknown>;
  submitted_at: string;
};

export type RecoveryCase = {
  id: string;
  tenant_id: string;
  location_id: string;
  feedback_response_id: string;
  status: string;
  priority: string;
  sentiment: string;
  rating: number;
  guest_name?: string;
  guest_phone?: string;
  guest_email?: string;
  feedback_preview: string;
  created_reason: string;
  created_at: string;
  updated_at: string;
};

export type Identity = {
  user_id: string;
  email: string;
  display_name: string;
  role: string;
  tenant_ids: string[];
  permissions: string[];
  provider: string;
};

export type Session = {
  access_token: string;
  expires_at: string;
  identity: Identity;
  tenant_id: string;
};

export type OnboardingState = {
  activation_id?: string;
  status: "not_started" | "in_progress" | "complete";
  next_step: "workspace" | "campaign" | "feedback_link" | "complete";
  source?: "homepage" | "guest_demo" | "direct";
  tenant?: Tenant;
  location?: Location;
  campaign?: SurveyCampaign;
  feedback_link?: FeedbackLink;
};

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "";
export const DEMO_TENANT_ID = process.env.NEXT_PUBLIC_DEMO_TENANT_ID ?? "11111111-1111-1111-1111-111111111111";

/** Local bearer sessions require an explicit auth-mode opt-in, never merely a local runtime label. */
export function isProductionAuth(): boolean {
  return process.env.NEXT_PUBLIC_PASSAGE_MODE !== "local";
}

export function safeReturnTo(value: string | null | undefined, fallback = "/admin"): string {
  if (!value || !value.startsWith("/") || value.startsWith("//") || value.includes("\\")) return fallback;
  return value;
}

export function identityAuthStartUrl(returnTo: string, options?: { intent?: "register" | "login"; email?: string; provider?: "google"; organizationName?: string; locationName?: string; source?: string }): string {
  const params = new URLSearchParams();
  params.set("return_to", safeReturnTo(returnTo, "/admin"));
  if (options?.intent) params.set("intent", options.intent);
  if (options?.email) params.set("email", options.email);
  if (options?.provider) params.set("provider", options.provider);
  if (options?.organizationName) params.set("organization_name", options.organizationName);
  if (options?.locationName) params.set("location_name", options.locationName);
  if (options?.source) params.set("source", options.source);
  return `/api/v1/auth/start?${params.toString()}`;
}

export async function availableIdentityProviders(): Promise<string[]> {
  try {
    const response = await fetch(`${API_BASE_URL}/api/v1/auth/providers`, { cache: "no-store", credentials: "include" });
    if (!response.ok) return [];
    const payload = await response.json() as { providers?: string[] };
    return Array.isArray(payload.providers) ? payload.providers : [];
  } catch {
    return [];
  }
}

type RequestOptions = {
  method?: "GET" | "POST" | "PATCH";
  body?: unknown;
  tenantId?: string;
  auth?: boolean;
  idempotencyKey?: string;
};

const SESSION_STORAGE_KEY = "heard-session-v1";

export function getStoredSession(): Session | null {
  if (typeof window === "undefined" || isProductionAuth()) {
    return null;
  }
  const raw = window.localStorage.getItem(SESSION_STORAGE_KEY);
  if (!raw) {
    return null;
  }
  try {
    const session = JSON.parse(raw) as Session;
    if (!session.access_token || new Date(session.expires_at).getTime() <= Date.now() || (isProductionAuth() && session.identity.provider === "passage-local")) {
      clearStoredSession();
      return null;
    }
    return session;
  } catch {
    clearStoredSession();
    return null;
  }
}

export function storeSession(session: Session) {
  if (isProductionAuth()) return;
  window.localStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify(session));
}

export function clearStoredSession() {
  if (typeof window !== "undefined") {
    window.localStorage.removeItem(SESSION_STORAGE_KEY);
  }
}

export async function apiFetch<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const session = options.auth === false ? null : getStoredSession();
  const response = await fetch(`${API_BASE_URL}${path}`, {
    method: options.method ?? "GET",
    headers: {
      "Content-Type": "application/json",
      ...(session ? { Authorization: `Bearer ${session.access_token}` } : {}),
      ...(options.idempotencyKey ? { "Idempotency-Key": options.idempotencyKey } : {}),
      ...(options.tenantId ? { "X-Heard-Tenant-ID": options.tenantId } : {})
    },
    body: options.body ? JSON.stringify(options.body) : undefined,
    cache: "no-store",
    credentials: "include"
  });

  if (!response.ok) {
    const payload = (await response.json().catch(() => null)) as { error?: { message?: string } } | null;
    if (response.status === 401 && options.auth !== false && typeof window !== "undefined") {
      clearStoredSession();
      const requested = `${window.location.pathname}${window.location.search}`;
      const next = requested.startsWith("/admin") || requested.startsWith("/onboarding") ? requested : "/admin";
      window.location.assign(`/login?next=${encodeURIComponent(next)}`);
    }
    throw new Error(payload?.error?.message ?? `Request failed with ${response.status}`);
  }

  return response.json() as Promise<T>;
}

export async function resolveSession(): Promise<Session | null> {
  if (!isProductionAuth()) return getStoredSession();
  try {
    const identity = await apiFetch<Identity>("/api/v1/session", { auth: false });
    return { access_token: "", expires_at: "", identity, tenant_id: identity.tenant_ids[0] ?? "" };
  } catch {
    return null;
  }
}

export async function createLocalSession(email: string): Promise<Session> {
  const session = await apiFetch<Session>("/api/v1/auth/local/session", {
    method: "POST",
    auth: false,
    body: { email }
  });
  storeSession(session);
  return session;
}

export async function createLocalRegistration(email: string): Promise<Session> {
  const session = await apiFetch<Session>("/api/v1/auth/local/registration", {
    method: "POST",
    auth: false,
    body: { email }
  });
  storeSession(session);
  return session;
}

export function formatRelativeDate(value: string): string {
  const date = new Date(value);
  return new Intl.DateTimeFormat("en-US", {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit"
  }).format(date);
}
