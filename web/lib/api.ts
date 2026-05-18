export type Tenant = {
  id: string;
  name: string;
  slug: string;
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
  status: string;
  channel: string;
  qr_asset_url: string;
  qr_svg?: string;
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

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";
export const DEMO_TENANT_ID = process.env.NEXT_PUBLIC_DEMO_TENANT_ID ?? "11111111-1111-1111-1111-111111111111";

type RequestOptions = {
  method?: "GET" | "POST" | "PATCH";
  body?: unknown;
  tenantId?: string;
};

export async function apiFetch<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    method: options.method ?? "GET",
    headers: {
      "Content-Type": "application/json",
      ...(options.tenantId ? { "X-Heard-Tenant-ID": options.tenantId } : {}),
      ...(options.tenantId ? { "X-Heard-Actor-ID": "web-admin" } : {}),
      ...(options.tenantId ? { "X-Heard-Actor-Role": "admin" } : {})
    },
    body: options.body ? JSON.stringify(options.body) : undefined,
    cache: "no-store"
  });

  if (!response.ok) {
    const payload = (await response.json().catch(() => null)) as { error?: { message?: string } } | null;
    throw new Error(payload?.error?.message ?? `Request failed with ${response.status}`);
  }

  return response.json() as Promise<T>;
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
