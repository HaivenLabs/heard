create table if not exists identity_jwks_cache (
    issuer text primary key,
    keys jsonb not null,
    fresh_until timestamptz not null,
    stale_until timestamptz not null,
    updated_at timestamptz not null default now(),
    check (fresh_until < stale_until)
);
