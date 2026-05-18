create extension if not exists pgcrypto;

create table if not exists tenants (
	id uuid primary key,
	name text not null,
	slug text not null unique,
	created_at timestamptz not null default now()
);

create table if not exists locations (
	id uuid primary key,
	tenant_id uuid not null references tenants(id) on delete cascade,
	name text not null,
	slug text not null,
	timezone text not null default 'America/Los_Angeles',
	created_at timestamptz not null default now(),
	unique (tenant_id, slug)
);

create table if not exists experiences (
	id uuid primary key,
	tenant_id uuid not null references tenants(id) on delete cascade,
	location_id uuid not null references locations(id) on delete cascade,
	source_system text not null,
	source_entity_id text not null,
	metadata jsonb not null default '{}'::jsonb,
	created_at timestamptz not null default now(),
	unique (tenant_id, location_id, source_system, source_entity_id)
);

create table if not exists feedback_links (
	id uuid primary key,
	tenant_id uuid not null references tenants(id) on delete cascade,
	location_id uuid not null references locations(id) on delete cascade,
	name text not null,
	token text not null unique,
	status text not null default 'active',
	channel text not null default 'qr',
	destination_url text not null,
	qr_asset_url text not null default '',
	qr_svg text not null default '',
	created_at timestamptz not null default now()
);

create table if not exists feedback_sessions (
	id uuid primary key,
	tenant_id uuid not null references tenants(id) on delete cascade,
	location_id uuid not null references locations(id) on delete cascade,
	feedback_link_id uuid references feedback_links(id) on delete set null,
	experience_id uuid references experiences(id) on delete set null,
	status text not null default 'started',
	source text not null default 'qr',
	channel text not null default 'dine-in',
	guest_name text not null default '',
	guest_phone text not null default '',
	guest_email text not null default '',
	wants_follow_up boolean not null default false,
	contact_consent boolean not null default false,
	marketing_consent boolean not null default false,
	metadata jsonb not null default '{}'::jsonb,
	created_at timestamptz not null default now(),
	completed_at timestamptz
);

create table if not exists feedback_responses (
	id uuid primary key,
	tenant_id uuid not null references tenants(id) on delete cascade,
	location_id uuid not null references locations(id) on delete cascade,
	feedback_session_id uuid not null unique references feedback_sessions(id) on delete cascade,
	feedback_link_id uuid references feedback_links(id) on delete set null,
	experience_id uuid references experiences(id) on delete set null,
	rating integer not null,
	sentiment text not null,
	comment text not null default '',
	categories jsonb not null default '[]'::jsonb,
	guest_name text not null default '',
	guest_phone text not null default '',
	guest_email text not null default '',
	wants_follow_up boolean not null default false,
	contact_consent boolean not null default false,
	marketing_consent boolean not null default false,
	metadata jsonb not null default '{}'::jsonb,
	submitted_at timestamptz not null default now()
);

create table if not exists recovery_cases (
	id uuid primary key,
	tenant_id uuid not null references tenants(id) on delete cascade,
	location_id uuid not null references locations(id) on delete cascade,
	feedback_response_id uuid not null unique references feedback_responses(id) on delete cascade,
	status text not null,
	priority text not null,
	sentiment text not null,
	rating integer not null,
	guest_name text not null default '',
	guest_phone text not null default '',
	guest_email text not null default '',
	feedback_preview text not null default '',
	created_reason text not null default 'negative_feedback_rule',
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now()
);

create table if not exists audit_entries (
	id uuid primary key,
	tenant_id uuid references tenants(id) on delete set null,
	actor_id text not null,
	actor_role text not null,
	action text not null,
	resource_type text not null,
	resource_id uuid not null,
	details jsonb not null default '{}'::jsonb,
	created_at timestamptz not null default now()
);

create table if not exists outbox_events (
	id uuid primary key,
	tenant_id uuid not null references tenants(id) on delete cascade,
	event_type text not null,
	event_version integer not null,
	aggregate_type text not null,
	aggregate_id uuid not null,
	payload jsonb not null,
	status text not null default 'pending',
	attempts integer not null default 0,
	last_error text not null default '',
	occurred_at timestamptz not null default now(),
	available_at timestamptz not null default now(),
	processed_at timestamptz,
	created_at timestamptz not null default now()
);

create index if not exists idx_locations_tenant_id on locations(tenant_id);
create index if not exists idx_feedback_links_tenant_id on feedback_links(tenant_id);
create index if not exists idx_feedback_links_token on feedback_links(token);
create index if not exists idx_feedback_sessions_tenant_created on feedback_sessions(tenant_id, created_at desc);
create index if not exists idx_feedback_responses_tenant_submitted on feedback_responses(tenant_id, submitted_at desc);
create index if not exists idx_recovery_cases_tenant_created on recovery_cases(tenant_id, created_at desc);
create index if not exists idx_outbox_events_pending on outbox_events(status, available_at);
