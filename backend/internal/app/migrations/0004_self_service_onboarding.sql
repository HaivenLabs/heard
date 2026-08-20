create table if not exists onboarding_activations (
	id uuid primary key,
	actor_provider text not null,
	actor_id text not null,
	idempotency_key text not null,
	tenant_id uuid not null unique references tenants(id) on delete cascade,
	location_id uuid not null unique references locations(id) on delete cascade,
	source text not null,
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	constraint onboarding_activations_actor_unique unique (actor_provider, actor_id),
	constraint onboarding_activations_source_check check (source in ('homepage', 'guest_demo', 'direct'))
);

create index if not exists idx_onboarding_activations_actor
	on onboarding_activations(actor_provider, actor_id);

