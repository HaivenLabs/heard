alter table survey_campaigns
	add column if not exists rating_face_set text not null default 'heard';

alter table survey_campaigns
	drop constraint if exists survey_campaigns_rating_face_set_check;

alter table survey_campaigns
	add constraint survey_campaigns_rating_face_set_check
	check (rating_face_set in ('heard', 'clay', 'glass', 'minimal', 'retro'));
