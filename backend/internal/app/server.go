package app

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Server struct {
	cfg      Config
	store    *Store
	identity IdentityProvider
}

type createLocalSessionRequest struct {
	Email string `json:"email"`
}

type createOnboardingActivationRequest struct {
	RestaurantName string `json:"restaurant_name"`
	LocationName   string `json:"location_name"`
	Timezone       string `json:"timezone"`
	Source         string `json:"source"`
}

type createMarketingLeadRequest struct {
	Name           string `json:"name"`
	WorkEmail      string `json:"work_email"`
	Phone          string `json:"phone"`
	RestaurantName string `json:"restaurant_name"`
	LocationCount  string `json:"location_count"`
	Challenge      string `json:"challenge"`
	Source         string `json:"source"`
	ContactConsent bool   `json:"contact_consent"`
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
	Identity  Identity
}

func NewServer(cfg Config, store *Store, identity IdentityProvider) *Server {
	return &Server{cfg: cfg, store: store, identity: identity}
}

func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/healthz", s.handleHealth)
	mux.HandleFunc("POST /api/v1/marketing-leads", s.handleCreateMarketingLead)
	mux.HandleFunc("POST /api/v1/auth/local/session", s.handleCreateLocalSession)
	mux.HandleFunc("POST /api/v1/auth/local/registration", s.handleCreateLocalRegistration)
	mux.HandleFunc("GET /api/v1/session", s.withIdentity("", s.handleGetSession))
	mux.HandleFunc("GET /api/v1/onboarding", s.withIdentity("", s.handleGetOnboarding))
	mux.HandleFunc("POST /api/v1/onboarding/activations", s.withIdentity("tenant:create", s.handleActivateRestaurantWorkspace))
	mux.HandleFunc("POST /api/v1/tenants", s.withIdentity("tenant:create", s.handleCreateTenant))
	mux.HandleFunc("GET /api/v1/tenants/{id}", s.withAdminContext("tenant:read", s.handleGetTenant))
	mux.HandleFunc("GET /api/v1/locations", s.withAdminContext("location:read", s.handleListLocations))
	mux.HandleFunc("POST /api/v1/locations", s.withAdminContext("location:write", s.handleCreateLocation))
	mux.HandleFunc("GET /api/v1/locations/{id}", s.withAdminContext("location:read", s.handleGetLocation))
	mux.HandleFunc("GET /api/v1/survey-campaigns", s.withAdminContext("campaign:read", s.handleListSurveyCampaigns))
	mux.HandleFunc("POST /api/v1/survey-campaigns", s.withAdminContext("campaign:write", s.handleCreateSurveyCampaign))
	mux.HandleFunc("GET /api/v1/survey-campaigns/{id}", s.withAdminContext("campaign:read", s.handleGetSurveyCampaign))
	mux.HandleFunc("GET /api/v1/public/surveys/{token}", s.handlePublicSurvey)
	mux.HandleFunc("POST /api/v1/feedback-links", s.withAdminContext("campaign:write", s.handleCreateFeedbackLink))
	mux.HandleFunc("POST /api/v1/feedback-links/{id}/qr", s.withAdminContext("campaign:write", s.handleGenerateQR))
	mux.HandleFunc("GET /api/v1/feedback-links/resolve/{token}", s.handleResolveFeedbackLink)
	mux.HandleFunc("POST /api/v1/feedback-sessions", s.handleCreateFeedbackSession)
	mux.HandleFunc("POST /api/v1/feedback-responses", s.handleSubmitFeedback)
	mux.HandleFunc("GET /api/v1/feedback-responses", s.withAdminContext("recovery:read", s.handleListFeedbackResponses))
	mux.HandleFunc("GET /api/v1/feedback-responses/{id}", s.withAdminContext("recovery:read", s.handleGetFeedbackResponse))
	mux.HandleFunc("GET /api/v1/recovery-cases", s.withAdminContext("recovery:read", s.handleListRecoveryCases))
	mux.HandleFunc("GET /api/v1/recovery-cases/{id}", s.withAdminContext("recovery:read", s.handleGetRecoveryCase))
	mux.HandleFunc("PATCH /api/v1/recovery-cases/{id}", s.withAdminContext("recovery:write", s.handleUpdateRecoveryCase))
	return s.withCORS(mux)
}

func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := s.cfg.AllowedOrigin
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Idempotency-Key, X-Heard-Tenant-ID")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		requestID := strings.TrimSpace(r.Header.Get("X-Request-ID"))
		if requestID == "" {
			requestID = uuid.NewString()
		}
		w.Header().Set("X-Request-ID", requestID)
		log.Printf(`{"event":"http.request","request_id":%q,"method":%q,"path":%q}`, requestID, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) withIdentity(permission string, next func(http.ResponseWriter, *http.Request, actorContext)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok || s.identity == nil {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		identity, err := s.identity.VerifyToken(r.Context(), token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or expired session")
			return
		}
		if permission != "" && !identity.HasPermission(permission) {
			log.Printf("authorization denied actor_id=%s permission=%s path=%s", identity.UserID, permission, r.URL.Path)
			writeError(w, http.StatusForbidden, "permission denied")
			return
		}
		next(w, r, actorContext{ActorID: identity.UserID, ActorRole: identity.Role, Identity: identity})
	}
}

func (s *Server) withAdminContext(permission string, next func(http.ResponseWriter, *http.Request, actorContext)) http.HandlerFunc {
	return s.withIdentity(permission, func(w http.ResponseWriter, r *http.Request, ctx actorContext) {
		ctx.TenantID = strings.TrimSpace(r.Header.Get("X-Heard-Tenant-ID"))
		if ctx.TenantID == "" {
			writeError(w, http.StatusUnauthorized, "tenant context is required")
			return
		}
		if !ctx.Identity.HasTenant(ctx.TenantID) {
			log.Printf("tenant access denied actor_id=%s tenant_id=%s path=%s", ctx.ActorID, ctx.TenantID, r.URL.Path)
			writeError(w, http.StatusForbidden, "tenant access denied")
			return
		}
		next(w, r, ctx)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (s *Server) handleCreateMarketingLead(w http.ResponseWriter, r *http.Request) {
	var req createMarketingLeadRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	lead, err := s.store.CreateMarketingLead(r.Context(), req)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	log.Printf("marketing lead created lead_id=%s source=%s location_count=%s", lead.ID, lead.Source, lead.LocationCount)
	writeJSON(w, http.StatusCreated, lead)
}

func (s *Server) handleCreateLocalSession(w http.ResponseWriter, r *http.Request) {
	issuer, ok := s.identity.(localSessionIssuer)
	if !ok {
		writeError(w, http.StatusNotFound, "local Passage adapter is disabled")
		return
	}
	var req createLocalSessionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	session, err := issuer.IssueSession(req.Email)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	log.Printf("local Passage session issued actor_id=%s tenant_id=%s", session.Identity.UserID, session.TenantID)
	writeJSON(w, http.StatusCreated, session)
}

func (s *Server) handleCreateLocalRegistration(w http.ResponseWriter, r *http.Request) {
	issuer, ok := s.identity.(localRegistrationIssuer)
	if !ok {
		writeError(w, http.StatusNotFound, "local Passage adapter is disabled")
		return
	}
	var req createLocalSessionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	session, err := issuer.IssueRegistration(req.Email)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	log.Printf("onboarding account resolved actor_id=%s provider=%s", session.Identity.UserID, session.Identity.Provider)
	writeJSON(w, http.StatusCreated, session)
}

func (s *Server) handleGetSession(w http.ResponseWriter, _ *http.Request, ctx actorContext) {
	writeJSON(w, http.StatusOK, ctx.Identity)
}

func (s *Server) handleGetOnboarding(w http.ResponseWriter, r *http.Request, ctx actorContext) {
	state, err := s.store.GetOnboardingState(r.Context(), ctx.Identity)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, state)
}

func (s *Server) handleActivateRestaurantWorkspace(w http.ResponseWriter, r *http.Request, ctx actorContext) {
	var req createOnboardingActivationRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	state, err := s.store.ActivateRestaurantWorkspace(r.Context(), ctx.Identity, ctx.ActorRole, r.Header.Get("Idempotency-Key"), req)
	if err != nil {
		log.Printf("onboarding activation failed actor_id=%s provider=%s error=%q", ctx.ActorID, ctx.Identity.Provider, err.Error())
		writeStoreError(w, err)
		return
	}
	log.Printf("onboarding workspace activated actor_id=%s tenant_id=%s activation_id=%s next_step=%s", ctx.ActorID, state.Tenant.ID, state.ActivationID, state.NextStep)
	writeJSON(w, http.StatusOK, state)
}

func (s *Server) handleCreateTenant(w http.ResponseWriter, r *http.Request, ctx actorContext) {
	var req createTenantRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	tenant, err := s.store.CreateTenant(r.Context(), ctx.ActorID, ctx.ActorRole, req)
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
	writeJSON(w, http.StatusOK, paginateCollection(locations, r))
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
	writeJSON(w, http.StatusOK, paginateCollection(campaigns, r))
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
	writeJSON(w, http.StatusOK, paginateCollection(items, r))
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
	writeJSON(w, http.StatusOK, paginateCollection(items, r))
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
	case errors.Is(err, errValidation):
		writeError(w, http.StatusBadRequest, strings.TrimPrefix(err.Error(), errValidation.Error()+": "))
	case strings.Contains(err.Error(), "required"), strings.Contains(err.Error(), "invalid"), strings.Contains(err.Error(), "mismatch"), strings.Contains(err.Error(), "already"), strings.Contains(err.Error(), "idempotency"):
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

func bearerToken(authorization string) (string, bool) {
	scheme, token, ok := strings.Cut(strings.TrimSpace(authorization), " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(token) == "" {
		return "", false
	}
	return strings.TrimSpace(token), true
}

func collectionPayload[T any](items []T) map[string]any {
	if items == nil {
		items = []T{}
	}
	return map[string]any{"items": items}
}

func paginateCollection[T any](items []T, r *http.Request) map[string]any {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 {
		pageSize = 25
	}
	if pageSize > 100 {
		pageSize = 100
	}
	start := (page - 1) * pageSize
	if start >= len(items) {
		return map[string]any{"items": []T{}, "page": page, "page_size": pageSize, "has_more": false}
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return map[string]any{"items": items[start:end], "page": page, "page_size": pageSize, "has_more": end < len(items)}
}
