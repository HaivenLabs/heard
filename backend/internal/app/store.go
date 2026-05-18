package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	cfg          Config
	pool         *pgxpool.Pool
	qurlProvider QURLProvider
}

const (
	demoTenantID       = "11111111-1111-1111-1111-111111111111"
	demoLocationID     = "22222222-2222-2222-2222-222222222222"
	demoFeedbackLinkID = "33333333-3333-3333-3333-333333333333"
	demoCampaignID     = "44444444-4444-4444-4444-444444444444"
)

func NewStore(ctx context.Context, cfg Config) (*Store, error) {
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	return &Store{
		cfg:          cfg,
		pool:         pool,
		qurlProvider: NewQURLProvider(cfg),
	}, nil
}

func (s *Store) Close() {
	s.pool.Close()
}

func (s *Store) SeedDemoData(ctx context.Context) error {
	if !s.cfg.DemoSeedEnabled {
		return nil
	}

	if _, err := s.pool.Exec(ctx, `
		insert into tenants (id, name, slug)
		values ($1, 'Sunday Hearth', 'sunday-hearth')
		on conflict (id) do nothing
	`, demoTenantID); err != nil {
		return err
	}
	if _, err := s.pool.Exec(ctx, `
		insert into locations (id, tenant_id, name, slug, timezone)
		values ($1, $2, 'Nom - Takeout', 'nom-takeout', 'America/Los_Angeles')
		on conflict (id) do nothing
	`, demoLocationID, demoTenantID); err != nil {
		return err
	}
	if _, err := s.pool.Exec(ctx, `
		insert into survey_campaigns (
			id, tenant_id, location_id, name, restaurant_name, headline, prompt, incentive_text,
			sms_keyword, sms_phone, google_review_url, yelp_review_url, status
		)
		values (
			$1, $2, $3, 'Takeout bag gift card survey', 'Nom', 'How did we do?',
			'Tap the face that matches your visit.', 'Complete this survey for a chance to win a $100 Nom gift card.',
			'WIN', '(877) 426-0492', 'https://www.google.com/maps/search/?api=1&query=Nom+restaurant',
			'https://www.yelp.com/search?find_desc=Nom', 'active'
		)
		on conflict (id) do nothing
	`, demoCampaignID, demoTenantID, demoLocationID); err != nil {
		return err
	}
	_, err := s.pool.Exec(ctx, `
		insert into feedback_links (id, tenant_id, location_id, campaign_id, name, token, status, channel, destination_url, qr_asset_url, qr_svg)
		values ($1, $2, $3, $4, 'Takeout flyer QR', 'demo-heard', 'active', 'flyer', $5, $5, $6)
		on conflict (id) do update set
			campaign_id = excluded.campaign_id,
			channel = excluded.channel,
			destination_url = excluded.destination_url,
			qr_asset_url = excluded.qr_asset_url,
			qr_svg = excluded.qr_svg
	`, demoFeedbackLinkID, demoTenantID, demoLocationID, demoCampaignID, strings.TrimRight(s.cfg.WebBaseURL, "/")+"/f/demo-heard", "")
	return err
}

func (s *Store) CreateTenant(ctx context.Context, actorID, actorRole string, req createTenantRequest) (Tenant, error) {
	tenant := Tenant{
		ID:   uuid.NewString(),
		Name: strings.TrimSpace(req.Name),
		Slug: slugify(req.Slug, req.Name),
	}
	if tenant.Name == "" {
		return Tenant{}, errors.New("tenant name is required")
	}

	if err := s.pool.QueryRow(ctx, `
		insert into tenants (id, name, slug)
		values ($1, $2, $3)
		returning created_at
	`, tenant.ID, tenant.Name, tenant.Slug).Scan(&tenant.CreatedAt); err != nil {
		return Tenant{}, err
	}
	_ = s.writeAudit(ctx, pgx.Tx(nil), "", actorID, actorRole, "tenant.created", "tenant", tenant.ID, map[string]any{"slug": tenant.Slug})
	return tenant, nil
}

func (s *Store) GetTenant(ctx context.Context, tenantID string) (Tenant, error) {
	tenant := Tenant{}
	err := s.pool.QueryRow(ctx, `
		select id::text, name, slug, created_at
		from tenants
		where id = $1
	`, tenantID).Scan(&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.CreatedAt)
	return tenant, err
}

func (s *Store) CreateLocation(ctx context.Context, tenantID, actorID, actorRole string, req createLocationRequest) (Location, error) {
	if tenantID == "" || tenantID != req.TenantID {
		return Location{}, errors.New("tenant mismatch")
	}
	location := Location{
		ID:       uuid.NewString(),
		TenantID: tenantID,
		Name:     strings.TrimSpace(req.Name),
		Slug:     slugify(req.Slug, req.Name),
		Timezone: defaultString(req.Timezone, "America/Los_Angeles"),
	}
	if location.Name == "" {
		return Location{}, errors.New("location name is required")
	}

	if err := s.pool.QueryRow(ctx, `
		insert into locations (id, tenant_id, name, slug, timezone)
		values ($1, $2, $3, $4, $5)
		returning created_at
	`, location.ID, location.TenantID, location.Name, location.Slug, location.Timezone).Scan(&location.CreatedAt); err != nil {
		return Location{}, err
	}
	_ = s.writeAudit(ctx, pgx.Tx(nil), tenantID, actorID, actorRole, "location.created", "location", location.ID, map[string]any{"slug": location.Slug})
	return location, nil
}

func (s *Store) GetLocation(ctx context.Context, tenantID, locationID string) (Location, error) {
	location := Location{}
	err := s.pool.QueryRow(ctx, `
		select id::text, tenant_id::text, name, slug, timezone, created_at
		from locations
		where id = $1 and tenant_id = $2
	`, locationID, tenantID).Scan(&location.ID, &location.TenantID, &location.Name, &location.Slug, &location.Timezone, &location.CreatedAt)
	return location, err
}

func (s *Store) ListLocations(ctx context.Context, tenantID string) ([]Location, error) {
	rows, err := s.pool.Query(ctx, `
		select id::text, tenant_id::text, name, slug, timezone, created_at
		from locations
		where tenant_id = $1
		order by created_at desc
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []Location
	for rows.Next() {
		var location Location
		if err := rows.Scan(&location.ID, &location.TenantID, &location.Name, &location.Slug, &location.Timezone, &location.CreatedAt); err != nil {
			return nil, err
		}
		locations = append(locations, location)
	}
	return locations, rows.Err()
}

func (s *Store) CreateSurveyCampaign(ctx context.Context, tenantID, actorID, actorRole string, req createSurveyCampaignRequest) (SurveyCampaign, error) {
	if tenantID == "" || tenantID != req.TenantID {
		return SurveyCampaign{}, errors.New("tenant mismatch")
	}
	if _, err := s.GetLocation(ctx, tenantID, req.LocationID); err != nil {
		return SurveyCampaign{}, errors.New("location not found for tenant")
	}

	campaign := SurveyCampaign{
		ID:              uuid.NewString(),
		TenantID:        tenantID,
		LocationID:      req.LocationID,
		Name:            defaultString(strings.TrimSpace(req.Name), "Takeout flyer survey"),
		RestaurantName:  defaultString(strings.TrimSpace(req.RestaurantName), "Nom"),
		Headline:        defaultString(strings.TrimSpace(req.Headline), "How did we do?"),
		Prompt:          defaultString(strings.TrimSpace(req.Prompt), "Tap the face that matches your visit."),
		IncentiveText:   defaultString(strings.TrimSpace(req.IncentiveText), "Complete this survey for a chance to win a $100 gift card."),
		SMSKeyword:      strings.TrimSpace(req.SMSKeyword),
		SMSPhone:        strings.TrimSpace(req.SMSPhone),
		GoogleReviewURL: strings.TrimSpace(req.GoogleReviewURL),
		YelpReviewURL:   strings.TrimSpace(req.YelpReviewURL),
		Status:          "active",
	}

	if err := s.pool.QueryRow(ctx, `
		insert into survey_campaigns (
			id, tenant_id, location_id, name, restaurant_name, headline, prompt, incentive_text,
			sms_keyword, sms_phone, google_review_url, yelp_review_url, status
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		returning created_at
	`, campaign.ID, campaign.TenantID, campaign.LocationID, campaign.Name, campaign.RestaurantName, campaign.Headline,
		campaign.Prompt, campaign.IncentiveText, campaign.SMSKeyword, campaign.SMSPhone, campaign.GoogleReviewURL,
		campaign.YelpReviewURL, campaign.Status,
	).Scan(&campaign.CreatedAt); err != nil {
		return SurveyCampaign{}, err
	}
	_ = s.writeAudit(ctx, pgx.Tx(nil), tenantID, actorID, actorRole, "survey_campaign.created", "survey_campaign", campaign.ID, map[string]any{
		"location_id": campaign.LocationID,
		"name":        campaign.Name,
	})
	return campaign, nil
}

func (s *Store) ListSurveyCampaigns(ctx context.Context, tenantID string) ([]SurveyCampaign, error) {
	rows, err := s.pool.Query(ctx, `
		select id::text, tenant_id::text, location_id::text, name, restaurant_name, headline, prompt, incentive_text,
			sms_keyword, sms_phone, google_review_url, yelp_review_url, status, created_at
		from survey_campaigns
		where tenant_id = $1
		order by created_at desc
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var campaigns []SurveyCampaign
	for rows.Next() {
		campaign, err := scanSurveyCampaign(rows)
		if err != nil {
			return nil, err
		}
		campaigns = append(campaigns, campaign)
	}
	return campaigns, rows.Err()
}

func (s *Store) GetSurveyCampaign(ctx context.Context, tenantID, campaignID string) (SurveyCampaign, error) {
	row := s.pool.QueryRow(ctx, `
		select id::text, tenant_id::text, location_id::text, name, restaurant_name, headline, prompt, incentive_text,
			sms_keyword, sms_phone, google_review_url, yelp_review_url, status, created_at
		from survey_campaigns
		where tenant_id = $1 and id = $2
	`, tenantID, campaignID)
	return scanSurveyCampaign(row)
}

func (s *Store) CreateFeedbackLink(ctx context.Context, tenantID, actorID, actorRole string, req createFeedbackLinkRequest) (FeedbackLink, error) {
	if tenantID == "" || tenantID != req.TenantID {
		return FeedbackLink{}, errors.New("tenant mismatch")
	}
	if _, err := s.GetLocation(ctx, tenantID, req.LocationID); err != nil {
		return FeedbackLink{}, errors.New("location not found for tenant")
	}
	if strings.TrimSpace(req.CampaignID) != "" {
		if _, err := s.GetSurveyCampaign(ctx, tenantID, req.CampaignID); err != nil {
			return FeedbackLink{}, errors.New("campaign not found for tenant")
		}
	}
	token := strings.TrimSpace(req.Token)
	if token == "" {
		var err error
		token, err = createOpaqueToken()
		if err != nil {
			return FeedbackLink{}, err
		}
	}

	link := FeedbackLink{
		ID:         uuid.NewString(),
		TenantID:   tenantID,
		LocationID: req.LocationID,
		CampaignID: strings.TrimSpace(req.CampaignID),
		Name:       defaultString(strings.TrimSpace(req.Name), "Feedback QR"),
		Token:      token,
		Status:     "active",
		Channel:    defaultString(req.Channel, "qr"),
	}
	link.Destination = strings.TrimRight(s.cfg.WebBaseURL, "/") + "/f/" + link.Token

	if qr, err := s.qurlProvider.GenerateFeedbackQR(ctx, link.Destination); err == nil {
		link.QRAssetURL = qr.AssetURL
		link.QRSVG = qr.SVG
	}

	if err := s.pool.QueryRow(ctx, `
		insert into feedback_links (id, tenant_id, location_id, campaign_id, name, token, status, channel, destination_url, qr_asset_url, qr_svg)
		values ($1, $2, $3, nullif($4, '')::uuid, $5, $6, $7, $8, $9, $10, $11)
		returning created_at
	`, link.ID, link.TenantID, link.LocationID, link.CampaignID, link.Name, link.Token, link.Status, link.Channel, link.Destination, link.QRAssetURL, link.QRSVG).Scan(&link.CreatedAt); err != nil {
		return FeedbackLink{}, err
	}
	_ = s.writeAudit(ctx, pgx.Tx(nil), tenantID, actorID, actorRole, "feedback_link.created", "feedback_link", link.ID, map[string]any{"location_id": link.LocationID, "channel": link.Channel})
	return link, nil
}

func (s *Store) RegenerateFeedbackLinkQR(ctx context.Context, tenantID, linkID string) (FeedbackLink, error) {
	link, err := s.ResolveFeedbackLinkByID(ctx, tenantID, linkID)
	if err != nil {
		return FeedbackLink{}, err
	}
	qr, err := s.qurlProvider.GenerateFeedbackQR(ctx, link.Destination)
	if err != nil {
		return FeedbackLink{}, err
	}
	link.QRAssetURL = qr.AssetURL
	link.QRSVG = qr.SVG
	if _, err := s.pool.Exec(ctx, `
		update feedback_links
		set qr_asset_url = $1, qr_svg = $2
		where id = $3 and tenant_id = $4
	`, link.QRAssetURL, link.QRSVG, link.ID, tenantID); err != nil {
		return FeedbackLink{}, err
	}
	return link, nil
}

func (s *Store) ResolveFeedbackLink(ctx context.Context, token string) (FeedbackLink, error) {
	link := FeedbackLink{}
	err := s.pool.QueryRow(ctx, `
		select id::text, tenant_id::text, location_id::text, coalesce(campaign_id::text, ''), name, token, status, channel, qr_asset_url, qr_svg, destination_url, created_at
		from feedback_links
		where token = $1 and status = 'active'
	`, token).Scan(
		&link.ID,
		&link.TenantID,
		&link.LocationID,
		&link.CampaignID,
		&link.Name,
		&link.Token,
		&link.Status,
		&link.Channel,
		&link.QRAssetURL,
		&link.QRSVG,
		&link.Destination,
		&link.CreatedAt,
	)
	return link, err
}

func (s *Store) ResolveFeedbackLinkByID(ctx context.Context, tenantID, linkID string) (FeedbackLink, error) {
	link := FeedbackLink{}
	err := s.pool.QueryRow(ctx, `
		select id::text, tenant_id::text, location_id::text, coalesce(campaign_id::text, ''), name, token, status, channel, qr_asset_url, qr_svg, destination_url, created_at
		from feedback_links
		where id = $1 and tenant_id = $2
	`, linkID, tenantID).Scan(
		&link.ID,
		&link.TenantID,
		&link.LocationID,
		&link.CampaignID,
		&link.Name,
		&link.Token,
		&link.Status,
		&link.Channel,
		&link.QRAssetURL,
		&link.QRSVG,
		&link.Destination,
		&link.CreatedAt,
	)
	return link, err
}

func (s *Store) GetPublicSurvey(ctx context.Context, token string) (PublicSurvey, error) {
	link, err := s.ResolveFeedbackLink(ctx, token)
	if err != nil {
		return PublicSurvey{}, err
	}
	if link.CampaignID == "" {
		return PublicSurvey{}, errors.New("feedback link is not attached to a survey campaign")
	}
	campaign, err := s.GetSurveyCampaign(ctx, link.TenantID, link.CampaignID)
	if err != nil {
		return PublicSurvey{}, err
	}
	return PublicSurvey{Link: link, Campaign: campaign}, nil
}

func (s *Store) CreateFeedbackSession(ctx context.Context, req createFeedbackSessionRequest) (FeedbackSession, error) {
	link, err := s.ResolveFeedbackLink(ctx, req.Token)
	if err != nil {
		return FeedbackSession{}, err
	}

	session := FeedbackSession{
		ID:               uuid.NewString(),
		TenantID:         link.TenantID,
		LocationID:       link.LocationID,
		FeedbackLinkID:   link.ID,
		Status:           "started",
		Source:           "qr",
		Channel:          defaultString(req.Channel, link.Channel),
		GuestName:        strings.TrimSpace(req.GuestName),
		GuestPhone:       strings.TrimSpace(req.GuestPhone),
		GuestEmail:       strings.TrimSpace(req.GuestEmail),
		WantsFollowUp:    req.WantsFollowUp,
		ContactConsent:   req.ContactConsent,
		MarketingConsent: req.MarketingConsent,
		Metadata:         defaultMetadata(req.Metadata),
	}

	metadata, err := json.Marshal(session.Metadata)
	if err != nil {
		return FeedbackSession{}, err
	}

	if err := s.pool.QueryRow(ctx, `
		insert into feedback_sessions (
			id, tenant_id, location_id, feedback_link_id, status, source, channel, guest_name, guest_phone, guest_email,
			wants_follow_up, contact_consent, marketing_consent, metadata
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		returning created_at, completed_at
	`, session.ID, session.TenantID, session.LocationID, session.FeedbackLinkID, session.Status, session.Source, session.Channel,
		session.GuestName, session.GuestPhone, session.GuestEmail, session.WantsFollowUp, session.ContactConsent, session.MarketingConsent, metadata,
	).Scan(&session.CreatedAt, &session.CompletedAt); err != nil {
		return FeedbackSession{}, err
	}
	return session, nil
}

func (s *Store) SubmitFeedback(ctx context.Context, req submitFeedbackRequest) (FeedbackResponse, error) {
	if req.Rating < 1 || req.Rating > 5 {
		return FeedbackResponse{}, errors.New("rating must be between 1 and 5")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return FeedbackResponse{}, err
	}
	defer tx.Rollback(ctx)

	var session FeedbackSession
	var rawMetadata []byte
	err = tx.QueryRow(ctx, `
		select id::text, tenant_id::text, location_id::text, coalesce(feedback_link_id::text, ''), coalesce(experience_id::text, ''), status, source, channel,
			guest_name, guest_phone, guest_email, wants_follow_up, contact_consent, marketing_consent, metadata, created_at, completed_at
		from feedback_sessions
		where id = $1
	`, req.FeedbackSessionID).Scan(
		&session.ID,
		&session.TenantID,
		&session.LocationID,
		&session.FeedbackLinkID,
		&session.ExperienceID,
		&session.Status,
		&session.Source,
		&session.Channel,
		&session.GuestName,
		&session.GuestPhone,
		&session.GuestEmail,
		&session.WantsFollowUp,
		&session.ContactConsent,
		&session.MarketingConsent,
		&rawMetadata,
		&session.CreatedAt,
		&session.CompletedAt,
	)
	if err != nil {
		return FeedbackResponse{}, err
	}
	if session.CompletedAt != nil {
		return FeedbackResponse{}, errors.New("feedback session already completed")
	}
	if err := json.Unmarshal(rawMetadata, &session.Metadata); err != nil {
		session.Metadata = map[string]any{}
	}

	categories := req.Categories
	if categories == nil {
		categories = []string{}
	}

	response := FeedbackResponse{
		ID:                uuid.NewString(),
		TenantID:          session.TenantID,
		LocationID:        session.LocationID,
		FeedbackSessionID: session.ID,
		FeedbackLinkID:    session.FeedbackLinkID,
		ExperienceID:      session.ExperienceID,
		Rating:            req.Rating,
		Sentiment:         sentimentFromRating(req.Rating),
		Comment:           strings.TrimSpace(req.Comment),
		Categories:        categories,
		GuestName:         fallbackTrim(req.GuestName, session.GuestName),
		GuestPhone:        fallbackTrim(req.GuestPhone, session.GuestPhone),
		GuestEmail:        fallbackTrim(req.GuestEmail, session.GuestEmail),
		WantsFollowUp:     req.WantsFollowUp || session.WantsFollowUp,
		ContactConsent:    req.ContactConsent || session.ContactConsent,
		MarketingConsent:  req.MarketingConsent || session.MarketingConsent,
		Metadata:          mergeMetadata(session.Metadata, req.Metadata),
		SubmittedAt:       time.Now().UTC(),
	}

	categoriesJSON, err := json.Marshal(response.Categories)
	if err != nil {
		return FeedbackResponse{}, err
	}
	metadata, err := json.Marshal(response.Metadata)
	if err != nil {
		return FeedbackResponse{}, err
	}

	if _, err := tx.Exec(ctx, `
		insert into feedback_responses (
			id, tenant_id, location_id, feedback_session_id, feedback_link_id, experience_id, rating, sentiment, comment,
			categories, guest_name, guest_phone, guest_email, wants_follow_up, contact_consent, marketing_consent, metadata, submitted_at
		)
		values ($1, $2, $3, $4, nullif($5, ''), nullif($6, ''), $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`, response.ID, response.TenantID, response.LocationID, response.FeedbackSessionID, response.FeedbackLinkID, response.ExperienceID,
		response.Rating, response.Sentiment, response.Comment, categoriesJSON, response.GuestName, response.GuestPhone, response.GuestEmail,
		response.WantsFollowUp, response.ContactConsent, response.MarketingConsent, metadata, response.SubmittedAt,
	); err != nil {
		return FeedbackResponse{}, err
	}

	if _, err := tx.Exec(ctx, `
		update feedback_sessions
		set status = 'completed', completed_at = $2, guest_name = $3, guest_phone = $4, guest_email = $5,
			wants_follow_up = $6, contact_consent = $7, marketing_consent = $8, metadata = $9
		where id = $1
	`, response.FeedbackSessionID, response.SubmittedAt, response.GuestName, response.GuestPhone, response.GuestEmail,
		response.WantsFollowUp, response.ContactConsent, response.MarketingConsent, metadata,
	); err != nil {
		return FeedbackResponse{}, err
	}

	event := FeedbackSubmittedEvent{
		EventID:            uuid.NewString(),
		EventType:          "feedback-submitted",
		EventVersion:       1,
		TenantID:           response.TenantID,
		LocationID:         response.LocationID,
		FeedbackResponseID: response.ID,
		FeedbackSessionID:  response.FeedbackSessionID,
		FeedbackLinkID:     response.FeedbackLinkID,
		ExperienceID:       response.ExperienceID,
		Sentiment:          response.Sentiment,
		Rating:             response.Rating,
		OccurredAt:         response.SubmittedAt,
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return FeedbackResponse{}, err
	}

	if _, err := tx.Exec(ctx, `
		insert into outbox_events (id, tenant_id, event_type, event_version, aggregate_type, aggregate_id, payload, occurred_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8)
	`, event.EventID, event.TenantID, event.EventType, event.EventVersion, "feedback_response", response.ID, payload, response.SubmittedAt); err != nil {
		return FeedbackResponse{}, err
	}

	if err := s.writeAudit(ctx, tx, response.TenantID, "guest", "guest", "feedback.submitted", "feedback_response", response.ID, map[string]any{
		"sentiment": response.Sentiment,
		"rating":    response.Rating,
	}); err != nil {
		return FeedbackResponse{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return FeedbackResponse{}, err
	}
	return response, nil
}

func (s *Store) ListFeedbackResponses(ctx context.Context, tenantID string) ([]FeedbackResponse, error) {
	rows, err := s.pool.Query(ctx, `
		select id::text, tenant_id::text, location_id::text, feedback_session_id::text, coalesce(feedback_link_id::text, ''), coalesce(experience_id::text, ''),
			rating, sentiment, comment, categories, guest_name, guest_phone, guest_email, wants_follow_up, contact_consent, marketing_consent, metadata, submitted_at
		from feedback_responses
		where tenant_id = $1
		order by submitted_at desc
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []FeedbackResponse
	for rows.Next() {
		var item FeedbackResponse
		var rawCategories []byte
		var rawMetadata []byte
		if err := rows.Scan(
			&item.ID,
			&item.TenantID,
			&item.LocationID,
			&item.FeedbackSessionID,
			&item.FeedbackLinkID,
			&item.ExperienceID,
			&item.Rating,
			&item.Sentiment,
			&item.Comment,
			&rawCategories,
			&item.GuestName,
			&item.GuestPhone,
			&item.GuestEmail,
			&item.WantsFollowUp,
			&item.ContactConsent,
			&item.MarketingConsent,
			&rawMetadata,
			&item.SubmittedAt,
		); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(rawCategories, &item.Categories)
		_ = json.Unmarshal(rawMetadata, &item.Metadata)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) GetFeedbackResponse(ctx context.Context, tenantID, responseID string) (FeedbackResponse, error) {
	var item FeedbackResponse
	var rawCategories []byte
	var rawMetadata []byte
	err := s.pool.QueryRow(ctx, `
		select id::text, tenant_id::text, location_id::text, feedback_session_id::text, coalesce(feedback_link_id::text, ''), coalesce(experience_id::text, ''),
			rating, sentiment, comment, categories, guest_name, guest_phone, guest_email, wants_follow_up, contact_consent, marketing_consent, metadata, submitted_at
		from feedback_responses
		where tenant_id = $1 and id = $2
	`, tenantID, responseID).Scan(
		&item.ID,
		&item.TenantID,
		&item.LocationID,
		&item.FeedbackSessionID,
		&item.FeedbackLinkID,
		&item.ExperienceID,
		&item.Rating,
		&item.Sentiment,
		&item.Comment,
		&rawCategories,
		&item.GuestName,
		&item.GuestPhone,
		&item.GuestEmail,
		&item.WantsFollowUp,
		&item.ContactConsent,
		&item.MarketingConsent,
		&rawMetadata,
		&item.SubmittedAt,
	)
	if err != nil {
		return FeedbackResponse{}, err
	}
	_ = json.Unmarshal(rawCategories, &item.Categories)
	_ = json.Unmarshal(rawMetadata, &item.Metadata)
	return item, nil
}

func (s *Store) ListRecoveryCases(ctx context.Context, tenantID string) ([]RecoveryCase, error) {
	rows, err := s.pool.Query(ctx, `
		select id::text, tenant_id::text, location_id::text, feedback_response_id::text, status, priority, sentiment, rating,
			guest_name, guest_phone, guest_email, feedback_preview, created_reason, created_at, updated_at
		from recovery_cases
		where tenant_id = $1
		order by created_at desc
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []RecoveryCase
	for rows.Next() {
		var item RecoveryCase
		if err := rows.Scan(
			&item.ID,
			&item.TenantID,
			&item.LocationID,
			&item.FeedbackResponseID,
			&item.Status,
			&item.Priority,
			&item.Sentiment,
			&item.Rating,
			&item.GuestName,
			&item.GuestPhone,
			&item.GuestEmail,
			&item.FeedbackPreview,
			&item.CreatedReason,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) GetRecoveryCase(ctx context.Context, tenantID, caseID string) (RecoveryCase, error) {
	item := RecoveryCase{}
	err := s.pool.QueryRow(ctx, `
		select id::text, tenant_id::text, location_id::text, feedback_response_id::text, status, priority, sentiment, rating,
			guest_name, guest_phone, guest_email, feedback_preview, created_reason, created_at, updated_at
		from recovery_cases
		where tenant_id = $1 and id = $2
	`, tenantID, caseID).Scan(
		&item.ID,
		&item.TenantID,
		&item.LocationID,
		&item.FeedbackResponseID,
		&item.Status,
		&item.Priority,
		&item.Sentiment,
		&item.Rating,
		&item.GuestName,
		&item.GuestPhone,
		&item.GuestEmail,
		&item.FeedbackPreview,
		&item.CreatedReason,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	return item, err
}

func (s *Store) UpdateRecoveryCaseStatus(ctx context.Context, tenantID, actorID, actorRole, caseID, status string) (RecoveryCase, error) {
	status = strings.TrimSpace(status)
	switch status {
	case "new", "open", "assigned", "waiting_on_guest", "waiting_on_internal_action", "offer_pending", "resolved", "closed", "spam", "duplicate", "archived":
	default:
		return RecoveryCase{}, errors.New("invalid recovery case status")
	}

	item := RecoveryCase{}
	err := s.pool.QueryRow(ctx, `
		update recovery_cases
		set status = $1, updated_at = now()
		where id = $2 and tenant_id = $3
		returning id::text, tenant_id::text, location_id::text, feedback_response_id::text, status, priority, sentiment, rating,
			guest_name, guest_phone, guest_email, feedback_preview, created_reason, created_at, updated_at
	`, status, caseID, tenantID).Scan(
		&item.ID,
		&item.TenantID,
		&item.LocationID,
		&item.FeedbackResponseID,
		&item.Status,
		&item.Priority,
		&item.Sentiment,
		&item.Rating,
		&item.GuestName,
		&item.GuestPhone,
		&item.GuestEmail,
		&item.FeedbackPreview,
		&item.CreatedReason,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return RecoveryCase{}, err
	}
	_ = s.writeAudit(ctx, pgx.Tx(nil), tenantID, actorID, actorRole, "recovery_case.status_updated", "recovery_case", caseID, map[string]any{"status": status})
	return item, nil
}

func (s *Store) ProcessNextOutboxEvent(ctx context.Context) (bool, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	var eventID string
	var tenantID string
	var payload []byte
	err = tx.QueryRow(ctx, `
		select id::text, tenant_id::text, payload
		from outbox_events
		where status = 'pending' and available_at <= now()
		order by created_at asc
		limit 1
		for update skip locked
	`).Scan(&eventID, &tenantID, &payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if _, err := tx.Exec(ctx, `
		update outbox_events
		set status = 'processing', attempts = attempts + 1
		where id = $1
	`, eventID); err != nil {
		return false, err
	}

	var event FeedbackSubmittedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		_, _ = tx.Exec(ctx, `update outbox_events set status = 'failed', last_error = $2 where id = $1`, eventID, err.Error())
		if commitErr := tx.Commit(ctx); commitErr != nil {
			return false, commitErr
		}
		return true, nil
	}

	response, err := s.GetFeedbackResponse(ctx, event.TenantID, event.FeedbackResponseID)
	if err != nil {
		return false, err
	}

	if shouldCreateRecoveryCaseForResponse(response) {
		var exists bool
		if err := tx.QueryRow(ctx, `select exists(select 1 from recovery_cases where feedback_response_id = $1)`, event.FeedbackResponseID).Scan(&exists); err != nil {
			return false, err
		}
		if !exists {
			if _, err := tx.Exec(ctx, `
				insert into recovery_cases (
					id, tenant_id, location_id, feedback_response_id, status, priority, sentiment, rating,
					guest_name, guest_phone, guest_email, feedback_preview, created_reason
				)
				values ($1, $2, $3, $4, 'new', $5, $6, $7, $8, $9, $10, $11, $12)
			`, uuid.NewString(), response.TenantID, response.LocationID, response.ID, priorityFromSentiment(response.Sentiment), response.Sentiment,
				response.Rating, response.GuestName, response.GuestPhone, response.GuestEmail, previewComment(response.Comment), recoveryReasonForResponse(response),
			); err != nil {
				return false, err
			}
			if err := s.writeAudit(ctx, tx, tenantID, "worker", "system", "recovery_case.created", "feedback_response", response.ID, map[string]any{
				"reason": recoveryReasonForResponse(response),
			}); err != nil {
				return false, err
			}
		}
	}

	if _, err := tx.Exec(ctx, `
		update outbox_events
		set status = 'processed', processed_at = now(), last_error = ''
		where id = $1
	`, eventID); err != nil {
		return false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Store) writeAudit(ctx context.Context, tx pgx.Tx, tenantID, actorID, actorRole, action, resourceType, resourceID string, details map[string]any) error {
	detailsJSON, err := json.Marshal(defaultMetadata(details))
	if err != nil {
		return err
	}

	query := `
		insert into audit_entries (id, tenant_id, actor_id, actor_role, action, resource_type, resource_id, details)
		values ($1, nullif($2, '')::uuid, $3, $4, $5, $6, $7, $8)
	`

	execFn := s.pool.Exec
	if tx != nil {
		execFn = tx.Exec
	}
	_, err = execFn(ctx, query, uuid.NewString(), tenantID, defaultString(actorID, "system"), defaultString(actorRole, "system"), action, resourceType, resourceID, detailsJSON)
	return err
}

func defaultMetadata(value map[string]any) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	return value
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func fallbackTrim(primary, fallback string) string {
	if strings.TrimSpace(primary) != "" {
		return strings.TrimSpace(primary)
	}
	return strings.TrimSpace(fallback)
}

func mergeMetadata(a, b map[string]any) map[string]any {
	out := map[string]any{}
	for key, value := range a {
		out[key] = value
	}
	for key, value := range b {
		out[key] = value
	}
	return out
}

type surveyCampaignScanner interface {
	Scan(dest ...any) error
}

func scanSurveyCampaign(row surveyCampaignScanner) (SurveyCampaign, error) {
	var campaign SurveyCampaign
	err := row.Scan(
		&campaign.ID,
		&campaign.TenantID,
		&campaign.LocationID,
		&campaign.Name,
		&campaign.RestaurantName,
		&campaign.Headline,
		&campaign.Prompt,
		&campaign.IncentiveText,
		&campaign.SMSKeyword,
		&campaign.SMSPhone,
		&campaign.GoogleReviewURL,
		&campaign.YelpReviewURL,
		&campaign.Status,
		&campaign.CreatedAt,
	)
	return campaign, err
}

func slugify(explicit, fallback string) string {
	source := strings.ToLower(strings.TrimSpace(explicit))
	if source == "" {
		source = strings.ToLower(strings.TrimSpace(fallback))
	}
	replacer := strings.NewReplacer(" ", "-", "_", "-", "/", "-", ".", "-", "&", "and", "'", "")
	source = replacer.Replace(source)
	source = strings.Trim(source, "-")
	if source == "" {
		source = fmt.Sprintf("item-%s", strings.ToLower(uuid.NewString()[:8]))
	}
	return source
}
