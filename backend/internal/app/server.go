package app

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
)

type Server struct {
	cfg   Config
	store *Store
}

type createTenantRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type createLocationRequest struct {
	TenantID string `json:"tenant_id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Timezone string `json:"timezone"`
}

type createSurveyCampaignRequest struct {
	TenantID        string `json:"tenant_id"`
	LocationID      string `json:"location_id"`
	Name            string `json:"name"`
	RestaurantName  string `json:"restaurant_name"`
	Headline        string `json:"headline"`
	Prompt          string `json:"prompt"`
	IncentiveText   string `json:"incentive_text"`
	SMSKeyword      string `json:"sms_keyword"`
	SMSPhone        string `json:"sms_phone"`
	GoogleReviewURL string `json:"google_review_url"`
	YelpReviewURL   string `json:"yelp_review_url"`
}

type createFeedbackLinkRequest struct {
	TenantID   string `json:"tenant_id"`
	LocationID string `json:"location_id"`
	CampaignID string `json:"campaign_id,omitempty"`
	Name       string `json:"name"`
	Channel    string `json:"channel"`
	Token      string `json:"token,omitempty"`
}

type createFeedbackSessionRequest struct {
	Token            string         `json:"token"`
	Channel          string         `json:"channel"`
	GuestName        string         `json:"guest_name"`
	GuestPhone       string         `json:"guest_phone"`
	GuestEmail       string         `json:"guest_email"`
	WantsFollowUp    bool           `json:"wants_follow_up"`
	ContactConsent   bool           `json:"contact_consent"`
	MarketingConsent bool           `json:"marketing_consent"`
	Metadata         map[string]any `json:"metadata"`
}

type submitFeedbackRequest struct {
	FeedbackSessionID string         `json:"feedback_session_id"`
	Rating            int            `json:"rating"`
	Comment           string         `json:"comment"`
	Categories        []string       `json:"categories"`
	GuestName         string         `json:"guest_name"`
	GuestPhone        string         `json:"guest_phone"`
	GuestEmail        string         `json:"guest_email"`
	WantsFollowUp     bool           `json:"wants_follow_up"`
	ContactConsent    bool           `json:"contact_consent"`
	MarketingConsent  bool           `json:"marketing_consent"`
	Metadata          map[string]any `json:"metadata"`
}

type updateRecoveryCaseRequest struct {
	Status string `json:"status"`
}

type actorContext struct {
	TenantID  string
	ActorID   string
	ActorRole string
}

func NewServer(cfg Config, store *Store) *Server {
	return &Server{cfg: cfg, store: store}
}

func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/healthz", s.handleHealth)
	mux.HandleFunc("POST /api/v1/tenants", s.handleCreateTenant)
	mux.HandleFunc("GET /api/v1/tenants/{id}", s.withAdminContext(s.handleGetTenant))
	mux.HandleFunc("GET /api/v1/locations", s.withAdminContext(s.handleListLocations))
	mux.HandleFunc("POST /api/v1/locations", s.withAdminContext(s.handleCreateLocation))
	mux.HandleFunc("GET /api/v1/locations/{id}", s.withAdminContext(s.handleGetLocation))
	mux.HandleFunc("GET /api/v1/survey-campaigns", s.withAdminContext(s.handleListSurveyCampaigns))
	mux.HandleFunc("POST /api/v1/survey-campaigns", s.withAdminContext(s.handleCreateSurveyCampaign))
	mux.HandleFunc("GET /api/v1/survey-campaigns/{id}", s.withAdminContext(s.handleGetSurveyCampaign))
	mux.HandleFunc("GET /api/v1/public/surveys/{token}", s.handlePublicSurvey)
	mux.HandleFunc("POST /api/v1/feedback-links", s.withAdminContext(s.handleCreateFeedbackLink))
	mux.HandleFunc("POST /api/v1/feedback-links/{id}/qr", s.withAdminContext(s.handleGenerateQR))
	mux.HandleFunc("GET /api/v1/feedback-links/resolve/{token}", s.handleResolveFeedbackLink)
	mux.HandleFunc("POST /api/v1/feedback-sessions", s.handleCreateFeedbackSession)
	mux.HandleFunc("POST /api/v1/feedback-responses", s.handleSubmitFeedback)
	mux.HandleFunc("GET /api/v1/feedback-responses", s.withAdminContext(s.handleListFeedbackResponses))
	mux.HandleFunc("GET /api/v1/feedback-responses/{id}", s.withAdminContext(s.handleGetFeedbackResponse))
	mux.HandleFunc("GET /api/v1/recovery-cases", s.withAdminContext(s.handleListRecoveryCases))
	mux.HandleFunc("GET /api/v1/recovery-cases/{id}", s.withAdminContext(s.handleGetRecoveryCase))
	mux.HandleFunc("PATCH /api/v1/recovery-cases/{id}", s.withAdminContext(s.handleUpdateRecoveryCase))
	return s.withCORS(mux)
}

func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := s.cfg.AllowedOrigin
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Heard-Tenant-ID, X-Heard-Actor-ID, X-Heard-Actor-Role")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) withAdminContext(next func(http.ResponseWriter, *http.Request, actorContext)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := actorContext{
			TenantID:  strings.TrimSpace(r.Header.Get("X-Heard-Tenant-ID")),
			ActorID:   defaultString(strings.TrimSpace(r.Header.Get("X-Heard-Actor-ID")), "local-admin"),
			ActorRole: defaultString(strings.TrimSpace(r.Header.Get("X-Heard-Actor-Role")), "admin"),
		}

		if ctx.TenantID == "" && r.Pattern != "GET /api/v1/tenants/{id}" {
			writeError(w, http.StatusUnauthorized, "tenant context is required")
			return
		}
		next(w, r, ctx)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (s *Server) handleCreateTenant(w http.ResponseWriter, r *http.Request) {
	var req createTenantRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	tenant, err := s.store.CreateTenant(r.Context(), "local-admin", "admin", req)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, tenant)
}

func (s *Server) handleGetTenant(w http.ResponseWriter, r *http.Request, ctx actorContext) {
	tenant, err := s.store.GetTenant(r.Context(), r.PathValue("id"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if ctx.TenantID != "" && tenant.ID != ctx.TenantID {
		writeError(w, http.StatusForbidden, "tenant access denied")
		return
	}
	writeJSON(w, http.StatusOK, tenant)
}

func (s *Server) handleCreateLocation(w http.ResponseWriter, r *http.Request, ctx actorContext) {
	var req createLocationRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	location, err := s.store.CreateLocation(r.Context(), ctx.TenantID, ctx.ActorID, ctx.ActorRole, req)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, location)
}

func (s *Server) handleListLocations(w http.ResponseWriter, r *http.Request, ctx actorContext) {
	locations, err := s.store.ListLocations(r.Context(), ctx.TenantID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": locations})
}

func (s *Server) handleGetLocation(w http.ResponseWriter, r *http.Request, ctx actorContext) {
	location, err := s.store.GetLocation(r.Context(), ctx.TenantID, r.PathValue("id"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, location)
}

func (s *Server) handleCreateSurveyCampaign(w http.ResponseWriter, r *http.Request, ctx actorContext) {
	var req createSurveyCampaignRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	campaign, err := s.store.CreateSurveyCampaign(r.Context(), ctx.TenantID, ctx.ActorID, ctx.ActorRole, req)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, campaign)
}

func (s *Server) handleListSurveyCampaigns(w http.ResponseWriter, r *http.Request, ctx actorContext) {
	campaigns, err := s.store.ListSurveyCampaigns(r.Context(), ctx.TenantID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": campaigns})
}

func (s *Server) handleGetSurveyCampaign(w http.ResponseWriter, r *http.Request, ctx actorContext) {
	campaign, err := s.store.GetSurveyCampaign(r.Context(), ctx.TenantID, r.PathValue("id"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, campaign)
}

func (s *Server) handlePublicSurvey(w http.ResponseWriter, r *http.Request) {
	survey, err := s.store.GetPublicSurvey(r.Context(), r.PathValue("token"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, survey)
}

func (s *Server) handleCreateFeedbackLink(w http.ResponseWriter, r *http.Request, ctx actorContext) {
	var req createFeedbackLinkRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	link, err := s.store.CreateFeedbackLink(r.Context(), ctx.TenantID, ctx.ActorID, ctx.ActorRole, req)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, link)
}

func (s *Server) handleGenerateQR(w http.ResponseWriter, r *http.Request, ctx actorContext) {
	link, err := s.store.RegenerateFeedbackLinkQR(r.Context(), ctx.TenantID, r.PathValue("id"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, link)
}

func (s *Server) handleResolveFeedbackLink(w http.ResponseWriter, r *http.Request) {
	link, err := s.store.ResolveFeedbackLink(r.Context(), r.PathValue("token"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, link)
}

func (s *Server) handleCreateFeedbackSession(w http.ResponseWriter, r *http.Request) {
	var req createFeedbackSessionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	session, err := s.store.CreateFeedbackSession(r.Context(), req)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, session)
}

func (s *Server) handleSubmitFeedback(w http.ResponseWriter, r *http.Request) {
	var req submitFeedbackRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	response, err := s.store.SubmitFeedback(r.Context(), req)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, response)
}

func (s *Server) handleListFeedbackResponses(w http.ResponseWriter, r *http.Request, ctx actorContext) {
	items, err := s.store.ListFeedbackResponses(r.Context(), ctx.TenantID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) handleGetFeedbackResponse(w http.ResponseWriter, r *http.Request, ctx actorContext) {
	item, err := s.store.GetFeedbackResponse(r.Context(), ctx.TenantID, r.PathValue("id"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) handleListRecoveryCases(w http.ResponseWriter, r *http.Request, ctx actorContext) {
	items, err := s.store.ListRecoveryCases(r.Context(), ctx.TenantID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) handleGetRecoveryCase(w http.ResponseWriter, r *http.Request, ctx actorContext) {
	item, err := s.store.GetRecoveryCase(r.Context(), ctx.TenantID, r.PathValue("id"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) handleUpdateRecoveryCase(w http.ResponseWriter, r *http.Request, ctx actorContext) {
	var req updateRecoveryCaseRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := s.store.UpdateRecoveryCaseStatus(r.Context(), ctx.TenantID, ctx.ActorID, ctx.ActorRole, r.PathValue("id"), req.Status)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func decodeJSON(r *http.Request, out any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(out)
}

func writeStoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		writeError(w, http.StatusNotFound, "resource not found")
	case strings.Contains(err.Error(), "required"), strings.Contains(err.Error(), "invalid"), strings.Contains(err.Error(), "mismatch"), strings.Contains(err.Error(), "already"):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, err.Error())
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]any{
			"message": message,
			"status":  status,
		},
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
