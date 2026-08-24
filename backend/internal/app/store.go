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
	demoTenantID        = "11111111-1111-1111-1111-111111111111"
	demoTenantSlug      = "nom-demo"
	demoLocationID      = "22222222-2222-2222-2222-222222222222"
	demoFeedbackLinkID  = "33333333-3333-3333-3333-333333333333"
	demoCampaignID      = "44444444-4444-4444-4444-444444444444"
	demoGoogleReviewURL = "https://maps.app.goo.gl/D3cEeXBEtGaKF2Lz8"
	demoYelpReviewURL   = "https://www.yelp.com/biz/nom-san-juan-capistrano"
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
		values ($1, 'nom', $2)
		on conflict (id) do update set name = excluded.name, slug = excluded.slug
	`, demoTenantID, demoTenantSlug); err != nil {
		return err
	}
	if _, err := s.pool.Exec(ctx, `
		insert into locations (id, tenant_id, name, slug, timezone)
		values ($1, $2, 'nom - Takeout', 'nom-takeout', 'America/Los_Angeles')
		on conflict (id) do update set name = excluded.name
	`, demoLocationID, demoTenantID); err != nil {
		return err
	}
	if _, err := s.pool.Exec(ctx, `
		insert into survey_campaigns (
			id, tenant_id, location_id, name, restaurant_name, headline, prompt, incentive_text,
			sms_keyword, sms_phone, google_review_url, yelp_review_url, logo_url, theme, status
		)
		values (
			$1, $2, $3, 'Takeout bag gift card survey', 'nom', 'How did we do?',
			'Tap the face that matches your visit.', 'Complete this survey for a chance to win a $100 nom gift card.',
			'WIN', '(877) 426-0492', $4, $5, '/brands/nom/logo.png', 'teal', 'active'
		)
		on conflict (id) do update set
			restaurant_name = excluded.restaurant_name,
			headline = excluded.headline,
			prompt = excluded.prompt,
			incentive_text = excluded.incentive_text,
			google_review_url = excluded.google_review_url,
			yelp_review_url = excluded.yelp_review_url,
			logo_url = excluded.logo_url,
			theme = excluded.theme
	`, demoCampaignID, demoTenantID, demoLocationID, demoGoogleReviewURL, demoYelpReviewURL); err != nil {
		return err
	}
	if _, err := s.pool.Exec(ctx, `
		update survey_campaigns
		set google_review_url = $1, yelp_review_url = $2
		where (google_review_url = '' or yelp_review_url = '') and lower(restaurant_name) = 'nom'
	`, demoGoogleReviewURL, demoYelpReviewURL); err != nil {
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

func (s *Store) CreateMarketingLead(ctx context.Context, req createMarketingLeadRequest) (MarketingLead, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.WorkEmail = strings.ToLower(strings.TrimSpace(req.WorkEmail))
	req.Phone = strings.TrimSpace(req.Phone)
	req.RestaurantName = strings.TrimSpace(req.RestaurantName)
	req.LocationCount = strings.TrimSpace(req.LocationCount)
	req.Challenge = strings.TrimSpace(req.Challenge)
	req.Source = defaultString(req.Source, "marketing_site")
	if err := validateMarketingLeadRequest(req); err != nil {
		return MarketingLead{}, err
	}

	lead := MarketingLead{
		ID:             uuid.NewString(),
		Name:           req.Name,
		WorkEmail:      req.WorkEmail,
		Phone:          req.Phone,
		RestaurantName: req.RestaurantName,
		LocationCount:  req.LocationCount,
		Challenge:      req.Challenge,
		Source:         req.Source,
		ContactConsent: req.ContactConsent,
		Status:         "new",
	}
	if err := s.pool.QueryRow(ctx, `
		insert into marketing_leads (
			id, name, work_email, phone, restaurant_name, location_count, challenge, source, contact_consent, status
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		returning created_at
	`, lead.ID, lead.Name, lead.WorkEmail, lead.Phone, lead.RestaurantName, lead.LocationCount, lead.Challenge, lead.Source, lead.ContactConsent, lead.Status).Scan(&lead.CreatedAt); err != nil {
		return MarketingLead{}, err
	}
	return lead, nil
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

func (s *Store) TenantHandleAvailability(ctx context.Context, rawHandle string) (string, bool, error) {
	handle := optionalSlug(rawHandle)
	if len(handle) < 2 || len(handle) > 80 {
		return "", false, errors.New("restaurant handle must be between 2 and 80 characters")
	}
	var exists bool
	if err := s.pool.QueryRow(ctx, `select exists(select 1 from tenants where slug = $1)`, handle).Scan(&exists); err != nil {
		return "", false, err
	}
	return handle, !exists, nil
}

func (s *Store) UpdateTenantHandle(ctx context.Context, tenantID, actorID, actorRole, rawHandle string) (Tenant, error) {
	handle, available, err := s.TenantHandleAvailability(ctx, rawHandle)
	if err != nil {
		return Tenant{}, err
	}
	tenant, err := s.GetTenant(ctx, tenantID)
	if err != nil {
		return Tenant{}, err
	}
	if tenant.Slug == handle {
		return tenant, nil
	}
	if !available {
		return Tenant{}, errors.New("that restaurant handle is already taken")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Tenant{}, err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `update tenants set slug = $1 where id = $2`, handle, tenantID); err != nil {
		return Tenant{}, err
	}
	if _, err := tx.Exec(ctx, `
		update feedback_links
		set destination_url = $1 || '/f/' || $2 || '/' || slug, qr_asset_url = '', qr_svg = ''
		where tenant_id = $3 and slug <> ''
	`, strings.TrimRight(s.cfg.WebBaseURL, "/"), handle, tenantID); err != nil {
		return Tenant{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Tenant{}, err
	}
	tenant.Slug = handle
	_ = s.writeAudit(ctx, pgx.Tx(nil), tenantID, actorID, actorRole, "tenant.handle_updated", "tenant", tenantID, map[string]any{"slug": handle})
	return tenant, nil
}

func (s *Store) GetOnboardingState(ctx context.Context, identity Identity) (OnboardingState, error) {
	state := OnboardingState{}
	var tenant Tenant
	var location Location
	err := s.pool.QueryRow(ctx, `
		select a.id::text, a.source,
			t.id::text, t.name, t.slug, t.created_at,
			l.id::text, l.tenant_id::text, l.name, l.slug, l.timezone, l.created_at
		from onboarding_activations a
		join tenants t on t.id = a.tenant_id
		join locations l on l.id = a.location_id and l.tenant_id = a.tenant_id
		where a.actor_provider = $1 and a.actor_id = $2
	`, identity.Provider, identity.UserID).Scan(
		&state.ActivationID, &state.Source,
		&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.CreatedAt,
		&location.ID, &location.TenantID, &location.Name, &location.Slug, &location.Timezone, &location.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		for _, tenantID := range identity.TenantIDs {
			existingTenant, tenantErr := s.GetTenant(ctx, tenantID)
			if tenantErr == nil {
				locations, locationsErr := s.ListLocations(ctx, tenantID)
				if locationsErr != nil {
					return OnboardingState{}, locationsErr
				}
				if len(locations) > 0 {
					state.Tenant = &existingTenant
					state.Location = &locations[0]
					state.Status = "complete"
					state.NextStep = "complete"
					return state, nil
				}
			}
			if !errors.Is(tenantErr, pgx.ErrNoRows) {
				return OnboardingState{}, tenantErr
			}
		}
		state.resolveProgress()
		return state, nil
	}
	if err != nil {
		return OnboardingState{}, err
	}
	state.Tenant = &tenant
	state.Location = &location

	campaigns, err := s.ListSurveyCampaigns(ctx, tenant.ID)
	if err != nil {
		return OnboardingState{}, err
	}
	for index := len(campaigns) - 1; index >= 0; index-- {
		if campaigns[index].LocationID == location.ID {
			campaign := campaigns[index]
			state.Campaign = &campaign
			break
		}
	}
	if state.Campaign != nil {
		link := FeedbackLink{}
		err = s.pool.QueryRow(ctx, `
			select id::text, tenant_id::text, location_id::text, coalesce(campaign_id::text, ''), name, token, slug, status, channel,
				qr_asset_url, qr_svg, destination_url, created_at
			from feedback_links
			where tenant_id = $1 and campaign_id = $2
			order by created_at asc
			limit 1
		`, tenant.ID, state.Campaign.ID).Scan(
			&link.ID, &link.TenantID, &link.LocationID, &link.CampaignID, &link.Name, &link.Token, &link.Slug, &link.Status,
			&link.Channel, &link.QRAssetURL, &link.QRSVG, &link.Destination, &link.CreatedAt,
		)
		if err == nil {
			isolateStoredQURL(&link, s.cfg.IsLocalRuntime())
			state.FeedbackLink = &link
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return OnboardingState{}, err
		}
	}
	state.resolveProgress()
	return state, nil
}

func (s *Store) ActivateRestaurantWorkspace(ctx context.Context, identity Identity, actorRole, idempotencyKey string, req createOnboardingActivationRequest) (OnboardingState, error) {
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if len(idempotencyKey) < 8 || len(idempotencyKey) > 200 {
		return OnboardingState{}, errors.New("idempotency key must be between 8 and 200 characters")
	}
	req.RestaurantName = strings.TrimSpace(req.RestaurantName)
	req.RestaurantHandle = optionalSlug(req.RestaurantHandle)
	req.LocationName = strings.TrimSpace(req.LocationName)
	req.Timezone = defaultString(strings.TrimSpace(req.Timezone), "America/Los_Angeles")
	req.Source = defaultString(strings.TrimSpace(req.Source), "direct")
	if req.RestaurantName == "" || req.RestaurantHandle == "" || req.LocationName == "" {
		return OnboardingState{}, errors.New("restaurant name, handle, and location name are required")
	}
	switch req.Source {
	case "homepage", "guest_demo", "direct":
	default:
		return OnboardingState{}, errors.New("invalid onboarding source")
	}
	if identity.UserID == "" || identity.Provider == "" || len(identity.TenantIDs) != 1 {
		return OnboardingState{}, errors.New("a single verified Passage account context is required")
	}
	tenantID := identity.TenantIDs[0]
	if _, err := uuid.Parse(tenantID); err != nil {
		return OnboardingState{}, errors.New("verified Passage account context is invalid")
	}

	existing, err := s.GetOnboardingState(ctx, identity)
	if err != nil {
		return OnboardingState{}, err
	}
	if existing.ActivationID != "" {
		return existing, nil
	}

	activationID := uuid.NewSHA1(uuid.NameSpaceURL, []byte("heard-onboarding:"+identity.Provider+":"+identity.UserID)).String()
	locationID := uuid.NewSHA1(uuid.NameSpaceURL, []byte("heard-first-location:"+tenantID)).String()
	eventID := uuid.NewSHA1(uuid.NameSpaceURL, []byte("heard-workspace-activated:"+activationID)).String()
	now := time.Now().UTC()
	locationSlug := slugify("", req.LocationName)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return OnboardingState{}, err
	}
	defer tx.Rollback(ctx)
	if _, err := claimTenantHandle(ctx, tx, tenantID, req.RestaurantName, req.RestaurantHandle); err != nil {
		return OnboardingState{}, err
	}
	if _, err := tx.Exec(ctx, `
		insert into locations (id, tenant_id, name, slug, timezone) values ($1, $2, $3, $4, $5)
		on conflict (id) do nothing
	`, locationID, tenantID, req.LocationName, locationSlug, req.Timezone); err != nil {
		return OnboardingState{}, err
	}
	result, err := tx.Exec(ctx, `
		insert into onboarding_activations (id, actor_provider, actor_id, idempotency_key, tenant_id, location_id, source)
		values ($1, $2, $3, $4, $5, $6, $7)
		on conflict (actor_provider, actor_id) do nothing
	`, activationID, identity.Provider, identity.UserID, idempotencyKey, tenantID, locationID, req.Source)
	if err != nil {
		return OnboardingState{}, err
	}
	if result.RowsAffected() == 1 {
		event := RestaurantWorkspaceActivatedEvent{
			EventID: eventID, EventType: "restaurant-workspace-activated", EventVersion: 1,
			TenantID: tenantID, LocationID: locationID, ActivationID: activationID, OccurredAt: now,
		}
		payload, err := json.Marshal(event)
		if err != nil {
			return OnboardingState{}, err
		}
		if _, err := tx.Exec(ctx, `
			insert into outbox_events (id, tenant_id, event_type, event_version, aggregate_type, aggregate_id, payload, occurred_at)
			values ($1, $2, $3, $4, 'onboarding_activation', $5, $6, $7)
			on conflict (id) do nothing
		`, event.EventID, event.TenantID, event.EventType, event.EventVersion, activationID, payload, event.OccurredAt); err != nil {
			return OnboardingState{}, err
		}
		if err := s.writeAudit(ctx, tx, tenantID, identity.UserID, actorRole, "restaurant_workspace.activated", "onboarding_activation", activationID, map[string]any{
			"location_id": locationID,
			"source":      req.Source,
		}); err != nil {
			return OnboardingState{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return OnboardingState{}, err
	}
	return s.GetOnboardingState(ctx, identity)
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
		RestaurantName:  defaultString(strings.TrimSpace(req.RestaurantName), "nom"),
		Headline:        defaultString(strings.TrimSpace(req.Headline), "How did we do?"),
		Prompt:          defaultString(strings.TrimSpace(req.Prompt), "Tap the face that matches your visit."),
		IncentiveText:   defaultString(strings.TrimSpace(req.IncentiveText), "Complete this survey for a chance to win a $100 gift card."),
		SMSKeyword:      strings.TrimSpace(req.SMSKeyword),
		SMSPhone:        strings.TrimSpace(req.SMSPhone),
		GoogleReviewURL: defaultString(strings.TrimSpace(req.GoogleReviewURL), demoGoogleReviewURL),
		YelpReviewURL:   defaultString(strings.TrimSpace(req.YelpReviewURL), demoYelpReviewURL),
		LogoURL:         defaultString(strings.TrimSpace(req.LogoURL), "/brands/nom/logo.png"),
		Theme:           defaultString(strings.TrimSpace(req.Theme), "teal"),
		Status:          "active",
	}

	if err := s.pool.QueryRow(ctx, `
		insert into survey_campaigns (
			id, tenant_id, location_id, name, restaurant_name, headline, prompt, incentive_text,
			sms_keyword, sms_phone, google_review_url, yelp_review_url, logo_url, theme, status
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		returning created_at
	`, campaign.ID, campaign.TenantID, campaign.LocationID, campaign.Name, campaign.RestaurantName, campaign.Headline,
		campaign.Prompt, campaign.IncentiveText, campaign.SMSKeyword, campaign.SMSPhone, campaign.GoogleReviewURL,
		campaign.YelpReviewURL, campaign.LogoURL, campaign.Theme, campaign.Status,
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
			sms_keyword, sms_phone, google_review_url, yelp_review_url, logo_url, theme, status, created_at
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
			sms_keyword, sms_phone, google_review_url, yelp_review_url, logo_url, theme, status, created_at
		from survey_campaigns
		where tenant_id = $1 and id = $2
	`, tenantID, campaignID)
	return scanSurveyCampaign(row)
}

func (s *Store) UpdateSurveyCampaign(ctx context.Context, tenantID, actorID, actorRole, campaignID string, req updateSurveyCampaignRequest) (SurveyCampaign, error) {
	if strings.TrimSpace(req.LocationID) == "" || strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.RestaurantName) == "" || strings.TrimSpace(req.Headline) == "" || strings.TrimSpace(req.Prompt) == "" {
		return SurveyCampaign{}, errors.New("location, campaign name, restaurant name, headline, and prompt are required")
	}
	if _, err := s.GetLocation(ctx, tenantID, req.LocationID); err != nil {
		return SurveyCampaign{}, errors.New("location not found for tenant")
	}
	campaign, err := s.GetSurveyCampaign(ctx, tenantID, campaignID)
	if err != nil {
		return SurveyCampaign{}, err
	}
	campaign.LocationID = req.LocationID
	campaign.Name = strings.TrimSpace(req.Name)
	campaign.RestaurantName = strings.TrimSpace(req.RestaurantName)
	campaign.Headline = strings.TrimSpace(req.Headline)
	campaign.Prompt = strings.TrimSpace(req.Prompt)
	campaign.IncentiveText = strings.TrimSpace(req.IncentiveText)
	campaign.SMSKeyword = strings.TrimSpace(req.SMSKeyword)
	campaign.SMSPhone = strings.TrimSpace(req.SMSPhone)
	campaign.GoogleReviewURL = strings.TrimSpace(req.GoogleReviewURL)
	campaign.YelpReviewURL = strings.TrimSpace(req.YelpReviewURL)
	campaign.LogoURL = strings.TrimSpace(req.LogoURL)
	campaign.Theme = strings.TrimSpace(req.Theme)
	if campaign.Theme == "" {
		campaign.Theme = "teal"
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return SurveyCampaign{}, err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `
		update survey_campaigns
		set location_id = $1, name = $2, restaurant_name = $3, headline = $4, prompt = $5, incentive_text = $6,
			sms_keyword = $7, sms_phone = $8, google_review_url = $9, yelp_review_url = $10, logo_url = $11, theme = $12
		where id = $13 and tenant_id = $14
	`, campaign.LocationID, campaign.Name, campaign.RestaurantName, campaign.Headline, campaign.Prompt, campaign.IncentiveText,
		campaign.SMSKeyword, campaign.SMSPhone, campaign.GoogleReviewURL, campaign.YelpReviewURL, campaign.LogoURL, campaign.Theme,
		campaign.ID, tenantID); err != nil {
		return SurveyCampaign{}, err
	}
	if _, err := tx.Exec(ctx, `
		update feedback_links
		set location_id = $1
		where tenant_id = $2 and campaign_id = $3
	`, campaign.LocationID, tenantID, campaign.ID); err != nil {
		return SurveyCampaign{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return SurveyCampaign{}, err
	}
	_ = s.writeAudit(ctx, pgx.Tx(nil), tenantID, actorID, actorRole, "survey_campaign.updated", "survey_campaign", campaign.ID, map[string]any{
		"location_id": campaign.LocationID,
		"name":        campaign.Name,
	})
	return campaign, nil
}

func (s *Store) ListFeedbackLinks(ctx context.Context, tenantID, campaignID string) ([]FeedbackLink, error) {
	query := `
		select id::text, tenant_id::text, location_id::text, coalesce(campaign_id::text, ''), name, token, coalesce(slug, ''), status, channel, qr_asset_url, qr_svg, destination_url, created_at
		from feedback_links
		where tenant_id = $1`
	args := []any{tenantID}
	if campaignID = strings.TrimSpace(campaignID); campaignID != "" {
		query += " and campaign_id = $2"
		args = append(args, campaignID)
	}
	query += " order by created_at desc"
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var links []FeedbackLink
	for rows.Next() {
		link, err := scanFeedbackLink(rows)
		if err != nil {
			return nil, err
		}
		isolateStoredQURL(&link, s.cfg.IsLocalRuntime())
		links = append(links, link)
	}
	return links, rows.Err()
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
	slug := optionalSlug(req.Slug)

	link := FeedbackLink{
		ID:         uuid.NewString(),
		TenantID:   tenantID,
		LocationID: req.LocationID,
		CampaignID: strings.TrimSpace(req.CampaignID),
		Name:       defaultString(strings.TrimSpace(req.Name), "Feedback QR"),
		Token:      token,
		Slug:       slug,
		Status:     "active",
		Channel:    defaultString(req.Channel, "qr"),
	}
	if link.Slug != "" {
		tenant, err := s.GetTenant(ctx, tenantID)
		if err != nil {
			return FeedbackLink{}, err
		}
		link.Destination = strings.TrimRight(s.cfg.WebBaseURL, "/") + "/f/" + tenant.Slug + "/" + link.Slug
	} else {
		link.Destination = strings.TrimRight(s.cfg.WebBaseURL, "/") + "/f/" + link.Token
	}

	if qr, err := s.qurlProvider.GenerateFeedbackQR(ctx, link.Destination); err == nil {
		link.QRAssetURL = qr.AssetURL
		link.QRSVG = ""
	}

	if err := s.pool.QueryRow(ctx, `
		insert into feedback_links (id, tenant_id, location_id, campaign_id, name, token, slug, status, channel, destination_url, qr_asset_url, qr_svg)
		values ($1, $2, $3, nullif($4, '')::uuid, $5, $6, $7, $8, $9, $10, $11, $12)
		returning created_at
	`, link.ID, link.TenantID, link.LocationID, link.CampaignID, link.Name, link.Token, link.Slug, link.Status, link.Channel, link.Destination, link.QRAssetURL, link.QRSVG).Scan(&link.CreatedAt); err != nil {
		return FeedbackLink{}, err
	}
	_ = s.writeAudit(ctx, pgx.Tx(nil), tenantID, actorID, actorRole, "feedback_link.created", "feedback_link", link.ID, map[string]any{"location_id": link.LocationID, "channel": link.Channel, "slug": link.Slug})
	return link, nil
}

func (s *Store) UpdateFeedbackLink(ctx context.Context, tenantID, actorID, actorRole, linkID string, req updateFeedbackLinkRequest) (FeedbackLink, error) {
	link, err := s.ResolveFeedbackLinkByID(ctx, tenantID, linkID)
	if err != nil {
		return FeedbackLink{}, err
	}
	cleanSlug := optionalSlug(req.Slug)
	if cleanSlug == "" {
		return FeedbackLink{}, errors.New("a survey path is required")
	}
	campaignID := strings.TrimSpace(req.CampaignID)
	if campaignID != "" {
		if _, err := s.GetSurveyCampaign(ctx, tenantID, campaignID); err != nil {
			return FeedbackLink{}, errors.New("campaign not found for tenant")
		}
		link.CampaignID = campaignID
	}
	tenant, err := s.GetTenant(ctx, tenantID)
	if err != nil {
		return FeedbackLink{}, err
	}
	link.Slug = cleanSlug
	link.Destination = strings.TrimRight(s.cfg.WebBaseURL, "/") + "/f/" + tenant.Slug + "/" + link.Slug
	if qr, err := s.qurlProvider.GenerateFeedbackQR(ctx, link.Destination); err == nil {
		link.QRAssetURL = qr.AssetURL
		link.QRSVG = ""
	}
	if _, err := s.pool.Exec(ctx, `
		update feedback_links
		set slug = $1, destination_url = $2, qr_asset_url = $3, qr_svg = $4, campaign_id = nullif($5, '')::uuid
		where id = $6 and tenant_id = $7
	`, link.Slug, link.Destination, link.QRAssetURL, link.QRSVG, link.CampaignID, link.ID, tenantID); err != nil {
		return FeedbackLink{}, err
	}
	_ = s.writeAudit(ctx, pgx.Tx(nil), tenantID, actorID, actorRole, "feedback_link.updated", "feedback_link", link.ID, map[string]any{"slug": link.Slug})
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
	link.QRSVG = ""
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
		select id::text, tenant_id::text, location_id::text, coalesce(campaign_id::text, ''), name, token, slug, status, channel, qr_asset_url, qr_svg, destination_url, created_at
		from feedback_links
		where token = $1 and status = 'active'
	`, token).Scan(
		&link.ID,
		&link.TenantID,
		&link.LocationID,
		&link.CampaignID,
		&link.Name,
		&link.Token,
		&link.Slug,
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
		select id::text, tenant_id::text, location_id::text, coalesce(campaign_id::text, ''), name, token, slug, status, channel, qr_asset_url, qr_svg, destination_url, created_at
		from feedback_links
		where id = $1 and tenant_id = $2
	`, linkID, tenantID).Scan(
		&link.ID,
		&link.TenantID,
		&link.LocationID,
		&link.CampaignID,
		&link.Name,
		&link.Token,
		&link.Slug,
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
	return s.publicSurveyForLink(ctx, link)
}

func (s *Store) GetPublicSurveyByPath(ctx context.Context, handle, slug string) (PublicSurvey, error) {
	link, err := s.resolveFeedbackLinkByHandleSlug(ctx, handle, slug)
	if err != nil {
		return PublicSurvey{}, err
	}
	return s.publicSurveyForLink(ctx, link)
}

func (s *Store) resolveFeedbackLinkByHandleSlug(ctx context.Context, handle, slug string) (FeedbackLink, error) {
	link := FeedbackLink{}
	err := s.pool.QueryRow(ctx, `
		select fl.id::text, fl.tenant_id::text, fl.location_id::text, coalesce(fl.campaign_id::text, ''), fl.name, fl.token, fl.slug, fl.status, fl.channel, fl.qr_asset_url, fl.qr_svg, fl.destination_url, fl.created_at
		from feedback_links fl
		join tenants t on t.id = fl.tenant_id
		where t.slug = $1 and fl.slug = $2 and fl.status = 'active'
	`, handle, slug).Scan(
		&link.ID,
		&link.TenantID,
		&link.LocationID,
		&link.CampaignID,
		&link.Name,
		&link.Token,
		&link.Slug,
		&link.Status,
		&link.Channel,
		&link.QRAssetURL,
		&link.QRSVG,
		&link.Destination,
		&link.CreatedAt,
	)
	if err == nil {
		isolateStoredQURL(&link, s.cfg.IsLocalRuntime())
	}
	return link, err
}

func (s *Store) publicSurveyForLink(ctx context.Context, link FeedbackLink) (PublicSurvey, error) {
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
	if err := validateFeedbackSessionInput(req); err != nil {
		return FeedbackSession{}, err
	}
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
	if err := validateSubmitFeedbackInput(req); err != nil {
		return FeedbackResponse{}, err
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
	// Campaign classification is authoritative server state. Never trust public
	// metadata to opt a flyer campaign out of its contact/recovery rules.
	var flyerCampaign bool
	if session.FeedbackLinkID != "" {
		err = tx.QueryRow(ctx, `
			select exists(
				select 1 from feedback_links fl
				join survey_campaigns sc on sc.id = fl.campaign_id
				where fl.id = $1 and fl.status = 'active'
			)`, session.FeedbackLinkID).Scan(&flyerCampaign)
		if err != nil {
			return FeedbackResponse{}, err
		}
	}

	categories := req.Categories
	if categories == nil {
		categories = []string{}
	} else {
		for index := range categories {
			categories[index] = strings.TrimSpace(categories[index])
		}
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
	if flyerCampaign {
		response.Metadata = enforceCampaignMetadata(response.Metadata, true, response.Rating)
	}
	requiresContact := flyerCampaign
	if err := validateContactDetails(response.GuestEmail, response.GuestPhone, requiresContact); err != nil {
		return FeedbackResponse{}, err
	}
	if flyerCampaign && !response.ContactConsent {
		return FeedbackResponse{}, validationError("transactional contact consent is required for giveaway entry")
	}
	if flyerCampaign && response.Rating < 5 && len(response.Comment) < 3 {
		return FeedbackResponse{}, validationError("tell us a little about what happened")
	}

	categoriesJSON, err := json.Marshal(response.Categories)
	if err != nil {
		return FeedbackResponse{}, err
	}
	metadata, err := json.Marshal(response.Metadata)
	if err != nil {
		return FeedbackResponse{}, err
	}
	feedbackLinkID, err := nullableUUID(response.FeedbackLinkID)
	if err != nil {
		return FeedbackResponse{}, err
	}
	experienceID, err := nullableUUID(response.ExperienceID)
	if err != nil {
		return FeedbackResponse{}, err
	}

	if _, err := tx.Exec(ctx, `
		insert into feedback_responses (
			id, tenant_id, location_id, feedback_session_id, feedback_link_id, experience_id, rating, sentiment, comment,
			categories, guest_name, guest_phone, guest_email, wants_follow_up, contact_consent, marketing_consent, metadata, submitted_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`, response.ID, response.TenantID, response.LocationID, response.FeedbackSessionID, feedbackLinkID, experienceID,
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

func nullableUUID(value string) (*uuid.UUID, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := uuid.Parse(value)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID %q: %w", value, err)
	}
	return &parsed, nil
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

	var currentStatus string
	if err := s.pool.QueryRow(ctx, `select status from recovery_cases where id = $1 and tenant_id = $2`, caseID, tenantID).Scan(&currentStatus); err != nil {
		return RecoveryCase{}, err
	}
	if !recoveryStatusTransitionAllowed(currentStatus, status) {
		return RecoveryCase{}, errors.New("invalid recovery case status transition")
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
	var eventType string
	var payload []byte
	err = tx.QueryRow(ctx, `
		select id::text, tenant_id::text, event_type, payload
		from outbox_events
		where status = 'pending' and available_at <= now()
		order by created_at asc
		limit 1
		for update skip locked
	`).Scan(&eventID, &tenantID, &eventType, &payload)
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
	retry := func(processErr error) (bool, error) {
		_, updateErr := tx.Exec(ctx, `
			update outbox_events
			set status = case when attempts >= 8 then 'failed' else 'pending' end,
				available_at = now() + make_interval(secs => least(3600, greatest(5, attempts * attempts * 5))),
				last_error = $2
			where id = $1
		`, eventID, processErr.Error())
		if updateErr != nil {
			return false, updateErr
		}
		if commitErr := tx.Commit(ctx); commitErr != nil {
			return false, commitErr
		}
		return true, processErr
	}

	if eventType == "restaurant-workspace-activated" {
		var event RestaurantWorkspaceActivatedEvent
		if err := json.Unmarshal(payload, &event); err != nil || event.ActivationID == "" {
			if err == nil {
				err = errors.New("workspace activation event is incomplete")
			}
			_, _ = tx.Exec(ctx, `update outbox_events set status = 'failed', last_error = $2 where id = $1`, eventID, err.Error())
			if commitErr := tx.Commit(ctx); commitErr != nil {
				return false, commitErr
			}
			return true, err
		}
		if _, err := tx.Exec(ctx, `
			update outbox_events set status = 'processed', processed_at = now(), last_error = '' where id = $1
		`, eventID); err != nil {
			return false, err
		}
		if err := tx.Commit(ctx); err != nil {
			return false, err
		}
		return true, nil
	}

	var event FeedbackSubmittedEvent
	if eventType != "feedback-submitted" {
		err = fmt.Errorf("unsupported event type %q", eventType)
	} else {
		err = json.Unmarshal(payload, &event)
	}
	if err != nil {
		_, _ = tx.Exec(ctx, `update outbox_events set status = 'failed', last_error = $2 where id = $1`, eventID, err.Error())
		if commitErr := tx.Commit(ctx); commitErr != nil {
			return false, commitErr
		}
		return true, err
	}

	response, err := s.GetFeedbackResponse(ctx, event.TenantID, event.FeedbackResponseID)
	if err != nil {
		return retry(err)
	}

	if shouldCreateRecoveryCaseForResponse(response) {
		var exists bool
		if err := tx.QueryRow(ctx, `select exists(select 1 from recovery_cases where feedback_response_id = $1)`, event.FeedbackResponseID).Scan(&exists); err != nil {
			return retry(err)
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
				return retry(err)
			}
			if err := s.writeAudit(ctx, tx, tenantID, "worker", "system", "recovery_case.created", "feedback_response", response.ID, map[string]any{
				"reason": recoveryReasonForResponse(response),
			}); err != nil {
				return retry(err)
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

func (s *Store) RequeueFailedOutboxEvent(ctx context.Context, tenantID, eventID, actorID, actorRole string) (OutboxRequeueResult, error) {
	if _, err := uuid.Parse(tenantID); err != nil {
		return OutboxRequeueResult{}, validationError("tenant context is invalid")
	}
	if _, err := uuid.Parse(eventID); err != nil {
		return OutboxRequeueResult{}, validationError("outbox event ID is invalid")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return OutboxRequeueResult{}, err
	}
	defer tx.Rollback(ctx)

	result := OutboxRequeueResult{}
	err = tx.QueryRow(ctx, `
		update outbox_events
		set status = 'pending', attempts = 0, last_error = '', available_at = now(), processed_at = null
		where id = $1 and tenant_id = $2 and status = 'failed'
		returning id::text, tenant_id::text, status, attempts
	`, eventID, tenantID).Scan(&result.EventID, &result.TenantID, &result.Status, &result.Attempts)
	if err != nil {
		return OutboxRequeueResult{}, err
	}
	if err := s.writeAudit(ctx, tx, tenantID, actorID, actorRole, "outbox_event.requeued", "outbox_event", eventID, map[string]any{
		"status": "pending",
	}); err != nil {
		return OutboxRequeueResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return OutboxRequeueResult{}, err
	}
	return result, nil
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

type feedbackLinkScanner interface {
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
		&campaign.LogoURL,
		&campaign.Theme,
		&campaign.Status,
		&campaign.CreatedAt,
	)
	return campaign, err
}

func scanFeedbackLink(row feedbackLinkScanner) (FeedbackLink, error) {
	var link FeedbackLink
	err := row.Scan(
		&link.ID,
		&link.TenantID,
		&link.LocationID,
		&link.CampaignID,
		&link.Name,
		&link.Token,
		&link.Slug,
		&link.Status,
		&link.Channel,
		&link.QRAssetURL,
		&link.QRSVG,
		&link.Destination,
		&link.CreatedAt,
	)
	return link, err
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

func optionalSlug(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return slugify(value, "")
}

// claimTenantHandle reserves a clean, heard-unique handle for a tenant. It
// tries the slugified restaurant name first and falls back to suffixed
// variants (nom, nom-2, nom-3, ...) until an unused one is found. The insert
// uses on conflict do nothing so concurrent claims of the same handle resolve
// atomically against the tenants.slug unique constraint; if the tenant row
// already exists (a retry), the stored handle is kept.
func claimTenantHandle(ctx context.Context, tx pgx.Tx, tenantID, name, handle string) (string, error) {
	result, err := tx.Exec(ctx, `insert into tenants (id, name, slug) values ($1, $2, $3) on conflict do nothing`, tenantID, name, handle)
	if err != nil {
		return "", err
	}
	if result.RowsAffected() == 1 {
		return handle, nil
	}
	var stored string
	err = tx.QueryRow(ctx, `select slug from tenants where id = $1`, tenantID).Scan(&stored)
	if err == nil {
		return stored, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return "", errors.New("that restaurant handle is already taken")
	}
	return "", err
}
