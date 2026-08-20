create table if not exists public_rate_limit_buckets (
	bucket_key text primary key,
	request_count integer not null,
	expires_at timestamptz not null
);

create index if not exists idx_public_rate_limit_buckets_expiry
	on public_rate_limit_buckets(expires_at);
