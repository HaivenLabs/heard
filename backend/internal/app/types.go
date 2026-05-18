package app

import "time"

type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

type Location struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Timezone  string    `json:"timezone"`
	CreatedAt time.Time `json:"created_at"`
}

type SurveyCampaign struct {
	ID              string    `json:"id"`
	TenantID        string    `json:"tenant_id"`
	LocationID      string    `json:"location_id"`
	Name            string    `json:"name"`
	RestaurantName  string    `json:"restaurant_name"`
	Headline        string    `json:"headline"`
	Prompt          string    `json:"prompt"`
	IncentiveText   string    `json:"incentive_text"`
	SMSKeyword      string    `json:"sms_keyword"`
	SMSPhone        string    `json:"sms_phone"`
	GoogleReviewURL string    `json:"google_review_url"`
	YelpReviewURL   string    `json:"yelp_review_url"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}

type FeedbackLink struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	LocationID  string    `json:"location_id"`
	CampaignID  string    `json:"campaign_id,omitempty"`
	Name        string    `json:"name"`
	Token       string    `json:"token"`
	Status      string    `json:"status"`
	Channel     string    `json:"channel"`
	QRAssetURL  string    `json:"qr_asset_url"`
	QRSVG       string    `json:"qr_svg,omitempty"`
	Destination string    `json:"destination_url"`
	CreatedAt   time.Time `json:"created_at"`
}

type PublicSurvey struct {
	Link     FeedbackLink   `json:"link"`
	Campaign SurveyCampaign `json:"campaign"`
}

type FeedbackSession struct {
	ID               string         `json:"id"`
	TenantID         string         `json:"tenant_id"`
	LocationID       string         `json:"location_id"`
	FeedbackLinkID   string         `json:"feedback_link_id,omitempty"`
	ExperienceID     string         `json:"experience_id,omitempty"`
	Status           string         `json:"status"`
	Source           string         `json:"source"`
	Channel          string         `json:"channel"`
	GuestName        string         `json:"guest_name,omitempty"`
	GuestPhone       string         `json:"guest_phone,omitempty"`
	GuestEmail       string         `json:"guest_email,omitempty"`
	WantsFollowUp    bool           `json:"wants_follow_up"`
	ContactConsent   bool           `json:"contact_consent"`
	MarketingConsent bool           `json:"marketing_consent"`
	Metadata         map[string]any `json:"metadata"`
	CreatedAt        time.Time      `json:"created_at"`
	CompletedAt      *time.Time     `json:"completed_at,omitempty"`
}

type FeedbackResponse struct {
	ID                string         `json:"id"`
	TenantID          string         `json:"tenant_id"`
	LocationID        string         `json:"location_id"`
	FeedbackSessionID string         `json:"feedback_session_id"`
	FeedbackLinkID    string         `json:"feedback_link_id,omitempty"`
	ExperienceID      string         `json:"experience_id,omitempty"`
	Rating            int            `json:"rating"`
	Sentiment         string         `json:"sentiment"`
	Comment           string         `json:"comment"`
	Categories        []string       `json:"categories"`
	GuestName         string         `json:"guest_name,omitempty"`
	GuestPhone        string         `json:"guest_phone,omitempty"`
	GuestEmail        string         `json:"guest_email,omitempty"`
	WantsFollowUp     bool           `json:"wants_follow_up"`
	ContactConsent    bool           `json:"contact_consent"`
	MarketingConsent  bool           `json:"marketing_consent"`
	Metadata          map[string]any `json:"metadata"`
	SubmittedAt       time.Time      `json:"submitted_at"`
}

type RecoveryCase struct {
	ID                 string    `json:"id"`
	TenantID           string    `json:"tenant_id"`
	LocationID         string    `json:"location_id"`
	FeedbackResponseID string    `json:"feedback_response_id"`
	Status             string    `json:"status"`
	Priority           string    `json:"priority"`
	Sentiment          string    `json:"sentiment"`
	Rating             int       `json:"rating"`
	GuestName          string    `json:"guest_name,omitempty"`
	GuestPhone         string    `json:"guest_phone,omitempty"`
	GuestEmail         string    `json:"guest_email,omitempty"`
	FeedbackPreview    string    `json:"feedback_preview"`
	CreatedReason      string    `json:"created_reason"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type FeedbackSubmittedEvent struct {
	EventID            string    `json:"event_id"`
	EventType          string    `json:"event_type"`
	EventVersion       int       `json:"event_version"`
	TenantID           string    `json:"tenant_id"`
	LocationID         string    `json:"location_id"`
	FeedbackResponseID string    `json:"feedback_response_id"`
	FeedbackSessionID  string    `json:"feedback_session_id"`
	FeedbackLinkID     string    `json:"feedback_link_id,omitempty"`
	ExperienceID       string    `json:"experience_id,omitempty"`
	Sentiment          string    `json:"sentiment"`
	Rating             int       `json:"rating"`
	OccurredAt         time.Time `json:"occurred_at"`
}
