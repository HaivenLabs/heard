alter table feedback_links
    add column if not exists slug text not null default '';

create unique index if not exists idx_feedback_links_tenant_slug
    on feedback_links (tenant_id, slug)
    where slug <> '';
