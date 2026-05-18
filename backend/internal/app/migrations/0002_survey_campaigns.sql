create table if not exists survey_campaigns (
	id uuid primary key,
	tenant_id uuid not null references tenants(id) on delete cascade,
	location_id uuid not null references locations(id) on delete cascade,
	name text not null,
	restaurant_name text not null,
	headline text not null,
	prompt text not null,
	incentive_text text not null default '',
	sms_keyword text not null default '',
	sms_phone text not null default '',
	google_review_url text not null default '',
	yelp_review_url text not null default '',
	status text not null default 'active',
	created_at timestamptz not null default now()
);

alter table feedback_links
	add column if not exists campaign_id uuid references survey_campaigns(id) on delete set null;

create index if not exists idx_survey_campaigns_tenant_created on survey_campaigns(tenant_id, created_at desc);
create index if not exists idx_feedback_links_campaign_id on feedback_links(campaign_id);
