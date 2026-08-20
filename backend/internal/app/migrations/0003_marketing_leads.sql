create table if not exists marketing_leads (
	id uuid primary key,
	name text not null,
	work_email text not null,
	phone text not null default '',
	restaurant_name text not null,
	location_count text not null,
	challenge text not null default '',
	source text not null default 'marketing_site',
	contact_consent boolean not null,
	status text not null default 'new',
	created_at timestamptz not null default now(),
	constraint marketing_leads_location_count_check check (location_count in ('1', '2-4', '5-19', '20-49', '50+')),
	constraint marketing_leads_source_check check (source in ('marketing_site', 'guest_demo')),
	constraint marketing_leads_contact_consent_check check (contact_consent = true),
	constraint marketing_leads_status_check check (status in ('new'))
);

create index if not exists idx_marketing_leads_created on marketing_leads(created_at desc);
create index if not exists idx_marketing_leads_email on marketing_leads(lower(work_email));
