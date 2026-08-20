alter table survey_campaigns
	add column if not exists logo_url text not null default '/brands/nom/logo.png',
	add column if not exists theme text not null default 'teal';
