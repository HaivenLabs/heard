package app

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Server struct {
	cfg           Config
	store         *Store
	identity      IdentityProvider
	passageHTTP   *http.Client
	limiter       publicWriteLimiter
	clientAddress clientAddressResolver
}

const (
	publicRequestBodyLimit  int64 = 32 * 1024
	defaultRequestBodyLimit int64 = 64 * 1024
)

type createLocalSessionRequest struct {
	Email string `json:"email"`
}

type heardLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type heardRegistrationRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Source   string `json:"source"`
}

type createOnboardingActivationRequest struct {
	RestaurantName   string `json:"restaurant_name"`
	RestaurantHandle string `json:"restaurant_handle"`
	LocationName     string `json:"location_name"`
	Timezone         string `json:"timezone"`
	Source           string `json:"source"`
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

type updateTenantHandleRequest struct {
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
	LogoURL         string `json:"logo_url"`
	Theme           string `json:"theme"`
}

type updateSurveyCampaignRequest struct {
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
	LogoURL         string `json:"logo_url"`
	Theme           string `json:"theme"`
}

type createFeedbackLinkRequest struct {
	TenantID   string `json:"tenant_id"`
	LocationID string `json:"location_id"`
	CampaignID string `json:"campaign_id,omitempty"`
	Name       string `json:"name"`
	Channel    string `json:"channel"`
	Token      string `json:"token,omitempty"`
	Slug       string `json:"slug,omitempty"`
}

type updateFeedbackLinkRequest struct {
	Slug       string `json:"slug"`
	CampaignID string `json:"campaign_id,omitempty"`
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
	var limiter publicWriteLimiter = newPublicRateLimiter(cfg.PublicWriteRateLimit, time.Duration(cfg.PublicWriteRateWindowSeconds)*time.Second, time.Now)
	if cfg.IsProductionLike() {
		limiter = unavailablePublicWriteLimiter{}
		if store != nil {
			limiter = newPostgresPublicRateLimiter(store.pool, cfg.PublicWriteRateLimit, cfg.PublicWriteRateWindowSeconds)
		}
	}
	return &Server{
		cfg: cfg, store: store, identity: identity,
		passageHTTP:   &http.Client{Timeout: 5 * time.Second},
		limiter:       limiter,
		clientAddress: newClientAddressResolver(cfg.TrustedProxyCIDRs),
	}
}

func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/healthz", s.handleHealth)
	mux.HandleFunc("POST /api/v1/marketing-leads", s.withPublicWriteControls("marketing-lead", s.handleCreateMarketingLead))
	mux.HandleFunc("POST /api/v1/auth/register", s.withPublicWriteControls("auth-registration", s.handleHeardRegistration))
	mux.HandleFunc("POST /api/v1/auth/login", s.withPublicWriteControls("auth-login", s.handleHeardLogin))
	if _, ok := s.identity.(localSessionIssuer); ok && s.cfg.IsLocalRuntime() {
		mux.HandleFunc("POST /api/v1/auth/local/session", s.handleCreateLocalSession)
	}
	if _, ok := s.identity.(localRegistrationIssuer); ok && s.cfg.IsLocalRuntime() {
		mux.HandleFunc("POST /api/v1/auth/local/registration", s.handleCreateLocalRegistration)
	}
	mux.HandleFunc("GET /api/v1/auth/start", s.handlePassageStart)
	mux.HandleFunc("GET /api/v1/auth/providers", s.handleIdentityProviders)
	mux.HandleFunc("GET /api/v1/auth/callback", s.handlePassageCallback)
	mux.HandleFunc("GET /api/v1/session", s.withIdentity("", s.handleGetSession))
	mux.HandleFunc("GET /api/v1/onboarding", s.withIdentity("", s.handleGetOnboarding))
	mux.HandleFunc("POST /api/v1/onboarding/activations", s.withIdentity("tenant:create", s.handleActivateRestaurantWorkspace))
	mux.HandleFunc("POST /api/v1/tenants", s.withIdentity("tenant:create", s.handleCreateTenant))
	mux.HandleFunc("GET /api/v1/tenants/{id}", s.withAdminContext("tenant:read", s.handleGetTenant))
	mux.HandleFunc("PATCH /api/v1/tenants/{id}", s.withAdminContext("tenant:write", s.handleUpdateTenantHandle))
	mux.HandleFunc("GET /api/v1/tenant-handles/{handle}/availability", s.withIdentity("", s.handleTenantHandleAvailability))
	mux.HandleFunc("GET /api/v1/locations", s.withAdminContext("location:read", s.handleListLocations))
	mux.HandleFunc("POST /api/v1/locations", s.withAdminContext("location:write", s.handleCreateLocation))
	mux.HandleFunc("GET /api/v1/locations/{id}", s.withAdminContext("location:read", s.handleGetLocation))
	mux.HandleFunc("GET /api/v1/survey-campaigns", s.withAdminContext("campaign:read", s.handleListSurveyCampaigns))
	mux.HandleFunc("POST /api/v1/survey-campaigns", s.withAdminContext("campaign:write", s.handleCreateSurveyCampaign))
	mux.HandleFunc("GET /api/v1/survey-campaigns/{id}", s.withAdminContext("campaign:read", s.handleGetSurveyCampaign))
	mux.HandleFunc("PATCH /api/v1/survey-campaigns/{id}", s.withAdminContext("campaign:write", s.handleUpdateSurveyCampaign))
	mux.HandleFunc("GET /api/v1/public/surveys/{token}", s.handlePublicSurvey)
	mux.HandleFunc("GET /api/v1/public/surveys/by-path/{handle}/{slug...}", s.handlePublicSurveyByPath)
	mux.HandleFunc("POST /api/v1/feedback-links", s.withAdminContext("campaign:write", s.handleCreateFeedbackLink))
	mux.HandleFunc("GET /api/v1/feedback-links", s.withAdminContext("campaign:read", s.handleListFeedbackLinks))
	mux.HandleFunc("PATCH /api/v1/feedback-links/{id}", s.withAdminContext("campaign:write", s.handleUpdateFeedbackLink))
	mux.HandleFunc("POST /api/v1/feedback-links/{id}/qr", s.withAdminContext("campaign:write", s.handleGenerateQR))
	mux.HandleFunc("GET /api/v1/feedback-links/resolve/{token}", s.handleResolveFeedbackLink)
	mux.HandleFunc("POST /api/v1/feedback-sessions", s.withPublicWriteControls("feedback-session", s.handleCreateFeedbackSession))
	mux.HandleFunc("POST /api/v1/feedback-responses", s.withPublicWriteControls("feedback-response", s.handleSubmitFeedback))
	mux.HandleFunc("GET /api/v1/feedback-responses", s.withAdminContext("recovery:read", s.handleListFeedbackResponses))
	mux.HandleFunc("GET /api/v1/feedback-responses/{id}", s.withAdminContext("recovery:read", s.handleGetFeedbackResponse))
	mux.HandleFunc("GET /api/v1/recovery-cases", s.withAdminContext("recovery:read", s.handleListRecoveryCases))
	mux.HandleFunc("GET /api/v1/recovery-cases/{id}", s.withAdminContext("recovery:read", s.handleGetRecoveryCase))
	mux.HandleFunc("PATCH /api/v1/recovery-cases/{id}", s.withAdminContext("recovery:write", s.handleUpdateRecoveryCase))
	mux.HandleFunc("POST /api/v1/outbox-events/{id}/requeue", s.withOwnerContext("outbox:replay", s.handleRequeueOutboxEvent))
	return s.withCORS(mux)
}

func (s *Server) withPublicWriteControls(route string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.limiter != nil {
			client := s.clientAddress.Resolve(r)
			allowed, retryAfter, err := s.limiter.Allow(r.Context(), route+":"+client)
			if err != nil {
				log.Printf(`{"event":"public_write.rate_limit_unavailable","route":%q,"error_type":%q}`, route, fmt.Sprintf("%T", err))
				writeError(w, http.StatusServiceUnavailable, "request protection is temporarily unavailable")
				return
			}
			if !allowed {
				seconds := int(retryAfter.Round(time.Second) / time.Second)
				if seconds < 1 {
					seconds = 1
				}
				w.Header().Set("Retry-After", strconv.Itoa(seconds))
				log.Printf(`{"event":"public_write.rate_limited","route":%q}`, route)
				writeError(w, http.StatusTooManyRequests, "too many requests; retry later")
				return
			}
		}
		next(w, r)
	}
}

func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := s.cfg.AllowedOrigin
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		if origin != "*" {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
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
		if !ok {
			if cookie, err := r.Cookie("heard_session"); err == nil && cookie.Value != "" {
				token, ok = cookie.Value, true
			}
		}
		if !ok || s.identity == nil {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		identity, err := s.identity.VerifyToken(r.Context(), token)
		if err != nil {
			if errors.Is(err, errIdentityUnavailable) {
				log.Printf("identity provider unavailable path=%s", r.URL.Path)
				writeError(w, http.StatusServiceUnavailable, "identity service unavailable")
				return
			}
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

func (s *Server) withOwnerContext(permission string, next func(http.ResponseWriter, *http.Request, actorContext)) http.HandlerFunc {
	return s.withAdminContext(permission, func(w http.ResponseWriter, r *http.Request, ctx actorContext) {
		if ctx.ActorRole != "owner" {
			log.Printf("owner access denied actor_id=%s tenant_id=%s path=%s", ctx.ActorID, ctx.TenantID, r.URL.Path)
			writeError(w, http.StatusForbidden, "owner access required")
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
	if err := decodeJSONRequest(w, r, &req, publicRequestBodyLimit); err != nil {
		writeRequestDecodeError(w, err)
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
		writeError(w, http.StatusNotFound, "local identity adapter is disabled")
		return
	}
	var req createLocalSessionRequest
	if err := decodeJSONRequest(w, r, &req, defaultRequestBodyLimit); err != nil {
		writeRequestDecodeError(w, err)
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
		writeError(w, http.StatusNotFound, "local identity adapter is disabled")
		return
	}
	var req createLocalSessionRequest
	if err := decodeJSONRequest(w, r, &req, defaultRequestBodyLimit); err != nil {
		writeRequestDecodeError(w, err)
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

func (s *Server) handleHeardRegistration(w http.ResponseWriter, r *http.Request) {
	var input heardRegistrationRequest
	if err := decodeJSONRequest(w, r, &input, defaultRequestBodyLimit); err != nil {
		writeRequestDecodeError(w, err)
		return
	}
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.Source = strings.TrimSpace(input.Source)
	if !isValidEmailFormat(input.Email) || len(input.Password) < 12 || (input.Source != "homepage" && input.Source != "guest_demo" && input.Source != "direct") {
		writeError(w, http.StatusBadRequest, "enter a valid email and password")
		return
	}
	payload, _ := json.Marshal(map[string]string{"email": input.Email, "password": input.Password})
	endpoint := strings.TrimRight(s.cfg.PassageBaseURL, "/") + "/api/v1/auth/register"
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, endpoint, strings.NewReader(string(payload)))
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "account creation is temporarily unavailable")
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if key := strings.TrimSpace(r.Header.Get("Idempotency-Key")); key != "" {
		req.Header.Set("Idempotency-Key", key)
	} else {
		req.Header.Set("Idempotency-Key", "heard-register-"+handoffRandom(12))
	}
	resp, err := s.passageHTTP.Do(req)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "account creation is temporarily unavailable")
		return
	}
	defer resp.Body.Close()
	raw, readErr := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if readErr != nil {
		writeError(w, http.StatusServiceUnavailable, "account creation is temporarily unavailable")
		return
	}
	if resp.StatusCode != http.StatusCreated {
		var upstream struct {
			Message string `json:"message"`
			Error   struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		_ = json.Unmarshal(raw, &upstream)
		message := strings.TrimSpace(upstream.Error.Message)
		if message == "" {
			message = strings.TrimSpace(upstream.Message)
		}
		if message == "" || strings.Contains(strings.ToLower(message), "passage") {
			message = "we could not create your account right now"
		}
		status := resp.StatusCode
		if status < http.StatusBadRequest || status > 499 {
			status = http.StatusServiceUnavailable
		}
		writeError(w, status, message)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"email": input.Email, "verification_delivery": "requested"})
}

func (s *Server) handleHeardLogin(w http.ResponseWriter, r *http.Request) {
	var input heardLoginRequest
	if err := decodeJSONRequest(w, r, &input, defaultRequestBodyLimit); err != nil {
		writeRequestDecodeError(w, err)
		return
	}
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if !isValidEmailFormat(input.Email) || len(strings.TrimSpace(input.Password)) == 0 {
		writeError(w, http.StatusBadRequest, "enter your email and password")
		return
	}
	payload, _ := json.Marshal(map[string]string{"email": input.Email, "password": input.Password})
	request, err := http.NewRequestWithContext(r.Context(), http.MethodPost, strings.TrimRight(s.cfg.PassageBaseURL, "/")+"/api/v1/auth/login", strings.NewReader(string(payload)))
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "sign-in is temporarily unavailable")
		return
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := s.passageHTTP.Do(request)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "sign-in is temporarily unavailable")
		return
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		writeError(w, http.StatusUnauthorized, "email or password is incorrect")
		return
	}
	var sessionCookie, csrfCookie *http.Cookie
	for _, cookie := range response.Cookies() {
		switch cookie.Name {
		case "passage_session":
			sessionCookie = cookie
		case "passage_csrf":
			csrfCookie = cookie
		}
	}
	if sessionCookie == nil || csrfCookie == nil {
		writeError(w, http.StatusServiceUnavailable, "sign-in is temporarily unavailable")
		return
	}
	tokenRequest, _ := http.NewRequestWithContext(r.Context(), http.MethodPost, strings.TrimRight(s.cfg.PassageBaseURL, "/")+"/api/v1/product-token", strings.NewReader(`{"audience":"heard"}`))
	tokenRequest.Header.Set("Content-Type", "application/json")
	tokenRequest.Header.Set("Cookie", "passage_session="+sessionCookie.Value+"; passage_csrf="+csrfCookie.Value)
	tokenRequest.Header.Set("X-CSRF-Token", csrfCookie.Value)
	tokenResponse, err := s.passageHTTP.Do(tokenRequest)
	if err != nil || tokenResponse == nil {
		writeError(w, http.StatusServiceUnavailable, "sign-in is temporarily unavailable")
		return
	}
	defer tokenResponse.Body.Close()
	var token struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}
	if tokenResponse.StatusCode != http.StatusOK || json.NewDecoder(io.LimitReader(tokenResponse.Body, 1<<16)).Decode(&token) != nil || token.TokenType != "Bearer" || token.AccessToken == "" {
		writeError(w, http.StatusUnauthorized, "your account is not ready for heard yet")
		return
	}
	if _, err := s.identity.VerifyToken(r.Context(), token.AccessToken); err != nil {
		writeError(w, http.StatusServiceUnavailable, "sign-in is temporarily unavailable")
		return
	}
	maxAge := token.ExpiresIn
	if maxAge <= 0 || maxAge > 600 {
		maxAge = 300
	}
	http.SetCookie(w, &http.Cookie{Name: "heard_session", Value: token.AccessToken, Path: "/", HttpOnly: true, Secure: s.cfg.IsProductionLike(), SameSite: http.SameSiteLaxMode, MaxAge: maxAge})
	writeJSON(w, http.StatusOK, map[string]bool{"signed_in": true})
}

func (s *Server) handlePassageStart(w http.ResponseWriter, r *http.Request) {
	returnTo := safeReturnPath(r.URL.Query().Get("return_to"))
	provider := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("provider")))
	if provider != "" && provider != "google" {
		writeError(w, http.StatusBadRequest, "sign-in method is unavailable")
		return
	}
	intent := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("intent")))
	organizationName := strings.TrimSpace(r.URL.Query().Get("organization_name"))
	locationName := strings.TrimSpace(r.URL.Query().Get("location_name"))
	if intent == "register" && (len(organizationName) > 120 || len(locationName) > 120) {
		writeError(w, http.StatusBadRequest, "account setup request is invalid")
		return
	}
	if intent == "register" && organizationName == "" {
		organizationName = "Heard workspace"
	}
	state := handoffRandom(24)
	verifier := handoffRandom(48)
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	secure := s.cfg.IsProductionLike()
	for _, cookie := range []*http.Cookie{{Name: "heard_oauth_state", Value: state, Path: "/api/v1/auth", HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode, MaxAge: 600}, {Name: "heard_oauth_verifier", Value: verifier, Path: "/api/v1/auth", HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode, MaxAge: 600}, {Name: "heard_oauth_return", Value: base64.RawURLEncoding.EncodeToString([]byte(returnTo)), Path: "/api/v1/auth", HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode, MaxAge: 600}} {
		http.SetCookie(w, cookie)
	}
	http.SetCookie(w, &http.Cookie{Name: "heard_oauth_intent", Value: intent, Path: "/api/v1/auth", HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode, MaxAge: 600})
	if intent == "register" && locationName != "" {
		draft, _ := json.Marshal(map[string]string{"restaurant_name": organizationName, "location_name": locationName, "source": strings.TrimSpace(r.URL.Query().Get("source"))})
		http.SetCookie(w, &http.Cookie{Name: "heard_onboarding_draft", Value: base64.RawURLEncoding.EncodeToString(draft), Path: "/api/v1/auth", HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode, MaxAge: 600})
	}
	publicURL := s.cfg.PassagePublicURL
	if publicURL == "" {
		publicURL = s.cfg.PassageBaseURL
	}
	authorize, err := url.Parse(strings.TrimRight(publicURL, "/") + "/api/v1/oauth/authorize")
	if err != nil {
		writeError(w, 500, "account service is not configured")
		return
	}
	q := authorize.Query()
	q.Set("client_id", s.cfg.PassageClientID)
	q.Set("redirect_uri", s.cfg.PassageCallbackURL)
	q.Set("state", state)
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")
	if intent == "register" || intent == "login" {
		q.Set("intent", intent)
	}
	if intent == "register" {
		q.Set("organization_name", organizationName)
	}
	email := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("email")))
	if email != "" && isValidEmailFormat(email) {
		q.Set("login_hint", email)
	}
	authorize.RawQuery = q.Encode()
	destination := authorize
	if provider != "" {
		external, parseErr := url.Parse(strings.TrimRight(publicURL, "/") + "/api/v1/auth/external/" + provider + "/start")
		if parseErr != nil {
			writeError(w, http.StatusInternalServerError, "account service is not configured")
			return
		}
		externalQuery := external.Query()
		externalQuery.Set("return_to", authorize.RequestURI())
		external.RawQuery = externalQuery.Encode()
		destination = external
	}
	http.Redirect(w, r, destination.String(), http.StatusFound)
}

func (s *Server) handleIdentityProviders(w http.ResponseWriter, r *http.Request) {
	endpoint := strings.TrimRight(s.cfg.PassageBaseURL, "/") + "/api/v1/auth/providers?client_id=" + url.QueryEscape(s.cfg.PassageClientID)
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, endpoint, nil)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "sign-in methods are temporarily unavailable")
		return
	}
	resp, err := s.passageHTTP.Do(req)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "sign-in methods are temporarily unavailable")
		return
	}
	defer resp.Body.Close()
	raw, readErr := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if readErr != nil || resp.StatusCode != http.StatusOK {
		writeError(w, http.StatusServiceUnavailable, "sign-in methods are temporarily unavailable")
		return
	}
	var payload struct {
		Providers []string `json:"providers"`
	}
	if json.Unmarshal(raw, &payload) != nil {
		writeError(w, http.StatusBadGateway, "sign-in methods are temporarily unavailable")
		return
	}
	allowed := make([]string, 0, 1)
	for _, provider := range payload.Providers {
		if provider == "google" {
			allowed = append(allowed, provider)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"providers": allowed})
}

func (s *Server) handlePassageCallback(w http.ResponseWriter, r *http.Request) {
	code, state, providerError := r.URL.Query().Get("code"), r.URL.Query().Get("state"), r.URL.Query().Get("error")
	if providerError == "provider_session_expired" {
		s.clearOAuthCookies(w)
		http.Redirect(w, r, strings.TrimRight(s.cfg.WebBaseURL, "/")+"/login?auth_notice=session_expired", http.StatusFound)
		return
	}
	stateCookie, stateErr := r.Cookie("heard_oauth_state")
	verifierCookie, verifierErr := r.Cookie("heard_oauth_verifier")
	returnCookie, _ := r.Cookie("heard_oauth_return")
	intentCookie, _ := r.Cookie("heard_oauth_intent")
	if stateErr != nil || subtle.ConstantTimeCompare([]byte(state), []byte(stateCookie.Value)) != 1 {
		writeError(w, 400, "account sign-in request is invalid or expired")
		return
	}
	if providerError != "" {
		s.clearOAuthCookies(w)
		intent := "login"
		if intentCookie != nil && intentCookie.Value == "register" {
			intent = "register"
		}
		path, notice := "/login", "provider_error"
		if intent == "register" {
			path = "/start"
		}
		switch providerError {
		case "access_denied":
			notice = "cancelled"
		case "identity_not_registered":
			if intent == "login" {
				registration := url.URL{Path: "/api/v1/auth/start"}
				query := registration.Query()
				query.Set("provider", "google")
				query.Set("intent", "register")
				query.Set("return_to", "/onboarding?auth_notice=google_account_created")
				registration.RawQuery = query.Encode()
				http.Redirect(w, r, strings.TrimRight(s.cfg.WebBaseURL, "/")+registration.String(), http.StatusFound)
				return
			}
			path, notice = "/start", "account_not_found"
		}
		destination := strings.TrimRight(s.cfg.WebBaseURL, "/") + path + "?auth_notice=" + url.QueryEscape(notice)
		http.Redirect(w, r, destination, http.StatusFound)
		return
	}
	if code == "" || verifierErr != nil {
		writeError(w, 400, "account sign-in request is invalid or expired")
		return
	}
	payload, _ := json.Marshal(map[string]string{"grant_type": "authorization_code", "code": code, "client_id": s.cfg.PassageClientID, "redirect_uri": s.cfg.PassageCallbackURL, "code_verifier": verifierCookie.Value})
	endpoint := strings.TrimRight(s.cfg.PassageBaseURL, "/") + "/api/v1/oauth/token"
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, endpoint, strings.NewReader(string(payload)))
	if err != nil {
		writeError(w, 502, "account service is temporarily unavailable")
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.passageHTTP.Do(req)
	if err != nil {
		log.Printf("Passage handoff exchange unavailable: %v", err)
		writeError(w, 503, "account service is temporarily unavailable")
		return
	}
	defer resp.Body.Close()
	raw, readErr := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if readErr != nil || resp.StatusCode != 200 {
		log.Printf("Passage handoff exchange failed status=%d", resp.StatusCode)
		writeError(w, 401, "account sign-in request is invalid or expired")
		return
	}
	var exchanged struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if json.Unmarshal(raw, &exchanged) != nil || exchanged.AccessToken == "" || exchanged.TokenType != "Bearer" {
		writeError(w, 502, "account service returned an invalid response")
		return
	}
	identity, err := s.identity.VerifyToken(r.Context(), exchanged.AccessToken)
	if err != nil {
		log.Printf("Passage handoff token verification failed")
		writeError(w, 502, "account service returned an invalid response")
		return
	}
	secure := s.cfg.IsProductionLike()
	maxAge := exchanged.ExpiresIn
	if maxAge <= 0 || maxAge > 600 {
		maxAge = 300
	}
	http.SetCookie(w, &http.Cookie{Name: "heard_session", Value: exchanged.AccessToken, Path: "/", HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode, MaxAge: maxAge})
	if draftCookie, draftErr := r.Cookie("heard_onboarding_draft"); draftErr == nil && s.store != nil {
		if rawDraft, decodeErr := base64.RawURLEncoding.DecodeString(draftCookie.Value); decodeErr == nil {
			var draft map[string]string
			if json.Unmarshal(rawDraft, &draft) == nil {
				_, activationErr := s.store.ActivateRestaurantWorkspace(r.Context(), identity, identity.Role, "oauth-onboarding-"+identity.UserID, createOnboardingActivationRequest{RestaurantName: draft["restaurant_name"], LocationName: draft["location_name"], Source: defaultString(draft["source"], "direct")})
				if activationErr == nil {
					http.SetCookie(w, &http.Cookie{Name: "heard_onboarding_draft", Value: "", Path: "/api/v1/auth", HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode, MaxAge: -1, Expires: time.Unix(1, 0)})
				} else {
					log.Printf("Heard onboarding activation failed after identity handoff: %v", activationErr)
				}
			}
		}
	}
	s.clearOAuthCookies(w)
	returnTo := "/onboarding"
	if returnCookie != nil {
		if decoded, err := base64.RawURLEncoding.DecodeString(returnCookie.Value); err == nil {
			returnTo = safeReturnPath(string(decoded))
		}
	}
	destination := strings.TrimRight(s.cfg.WebBaseURL, "/") + returnTo
	http.Redirect(w, r, destination, http.StatusFound)
}

func (s *Server) clearOAuthCookies(w http.ResponseWriter) {
	secure := s.cfg.IsProductionLike()
	for _, name := range []string{"heard_oauth_state", "heard_oauth_verifier", "heard_oauth_return", "heard_oauth_intent"} {
		http.SetCookie(w, &http.Cookie{Name: name, Value: "", Path: "/api/v1/auth", HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode, MaxAge: -1, Expires: time.Unix(1, 0)})
	}
}

func handoffRandom(size int) string {
	b := make([]byte, size)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
func safeReturnPath(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "/onboarding"
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.IsAbs() || parsed.Host != "" || !strings.HasPrefix(parsed.Path, "/") || strings.HasPrefix(value, "//") {
		return "/onboarding"
	}
	return parsed.RequestURI()
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
	if err := decodeJSONRequest(w, r, &req, defaultRequestBodyLimit); err != nil {
		writeRequestDecodeError(w, err)
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
	if err := decodeJSONRequest(w, r, &req, defaultRequestBodyLimit); err != nil {
		writeRequestDecodeError(w, err)
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

func (s *Server) handleTenantHandleAvailability(w http.ResponseWriter, r *http.Request, _ actorContext) {
	handle, available, err := s.store.TenantHandleAvailability(r.Context(), r.PathValue("handle"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"handle": handle, "available": available})
}

func (s *Server) handleUpdateTenantHandle(w http.ResponseWriter, r *http.Request, ctx actorContext) {
	if r.PathValue("id") != ctx.TenantID {
		writeError(w, http.StatusForbidden, "tenant access denied")
		return
	}
	var req updateTenantHandleRequest
	if err := decodeJSONRequest(w, r, &req, defaultRequestBodyLimit); err != nil {
		writeRequestDecodeError(w, err)
		return
	}
	tenant, err := s.store.UpdateTenantHandle(r.Context(), ctx.TenantID, ctx.ActorID, ctx.ActorRole, req.Slug)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tenant)
}

func (s *Server) handleCreateLocation(w http.ResponseWriter, r *http.Request, ctx actorContext) {
	var req createLocationRequest
	if err := decodeJSONRequest(w, r, &req, defaultRequestBodyLimit); err != nil {
		writeRequestDecodeError(w, err)
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
	if err := decodeJSONRequest(w, r, &req, defaultRequestBodyLimit); err != nil {
		writeRequestDecodeError(w, err)
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

func (s *Server) handleUpdateSurveyCampaign(w http.ResponseWriter, r *http.Request, ctx actorContext) {
	var req updateSurveyCampaignRequest
	if err := decodeJSONRequest(w, r, &req, defaultRequestBodyLimit); err != nil {
		writeRequestDecodeError(w, err)
		return
	}
	campaign, err := s.store.UpdateSurveyCampaign(r.Context(), ctx.TenantID, ctx.ActorID, ctx.ActorRole, r.PathValue("id"), req)
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

func (s *Server) handlePublicSurveyByPath(w http.ResponseWriter, r *http.Request) {
	handle := strings.ToLower(strings.Trim(strings.TrimSpace(r.PathValue("handle")), "/"))
	slug := strings.ToLower(strings.Trim(strings.TrimSpace(r.PathValue("slug")), "/"))
	if handle == "" || slug == "" {
		writeError(w, http.StatusNotFound, "survey not found")
		return
	}
	survey, err := s.store.GetPublicSurveyByPath(r.Context(), handle, slug)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, survey)
}

func (s *Server) handleCreateFeedbackLink(w http.ResponseWriter, r *http.Request, ctx actorContext) {
	var req createFeedbackLinkRequest
	if err := decodeJSONRequest(w, r, &req, defaultRequestBodyLimit); err != nil {
		writeRequestDecodeError(w, err)
		return
	}
	link, err := s.store.CreateFeedbackLink(r.Context(), ctx.TenantID, ctx.ActorID, ctx.ActorRole, req)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, link)
}

func (s *Server) handleListFeedbackLinks(w http.ResponseWriter, r *http.Request, ctx actorContext) {
	links, err := s.store.ListFeedbackLinks(r.Context(), ctx.TenantID, r.URL.Query().Get("campaign_id"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, paginateCollection(links, r))
}

func (s *Server) handleUpdateFeedbackLink(w http.ResponseWriter, r *http.Request, ctx actorContext) {
	var req updateFeedbackLinkRequest
	if err := decodeJSONRequest(w, r, &req, defaultRequestBodyLimit); err != nil {
		writeRequestDecodeError(w, err)
		return
	}
	link, err := s.store.UpdateFeedbackLink(r.Context(), ctx.TenantID, ctx.ActorID, ctx.ActorRole, r.PathValue("id"), req)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, link)
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
	if err := decodeJSONRequest(w, r, &req, publicRequestBodyLimit); err != nil {
		writeRequestDecodeError(w, err)
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
	if err := decodeJSONRequest(w, r, &req, publicRequestBodyLimit); err != nil {
		writeRequestDecodeError(w, err)
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
	if err := decodeJSONRequest(w, r, &req, defaultRequestBodyLimit); err != nil {
		writeRequestDecodeError(w, err)
		return
	}
	item, err := s.store.UpdateRecoveryCaseStatus(r.Context(), ctx.TenantID, ctx.ActorID, ctx.ActorRole, r.PathValue("id"), req.Status)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) handleRequeueOutboxEvent(w http.ResponseWriter, r *http.Request, ctx actorContext) {
	result, err := s.store.RequeueFailedOutboxEvent(r.Context(), ctx.TenantID, r.PathValue("id"), ctx.ActorID, ctx.ActorRole)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	log.Printf(`{"event":"outbox.requeued","event_id":%q,"tenant_id":%q,"actor_id":%q}`, result.EventID, result.TenantID, ctx.ActorID)
	writeJSON(w, http.StatusOK, result)
}

type requestDecodeError struct {
	status  int
	message string
	err     error
}

func (e *requestDecodeError) Error() string { return e.err.Error() }

func decodeJSONRequest(w http.ResponseWriter, r *http.Request, out any, maxBytes int64) error {
	defer r.Body.Close()
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return &requestDecodeError{status: http.StatusUnsupportedMediaType, message: "Content-Type must be application/json", err: errors.New("unsupported media type")}
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return &requestDecodeError{status: http.StatusRequestEntityTooLarge, message: fmt.Sprintf("request body exceeds %d KiB", maxBytes/1024), err: err}
		}
		return &requestDecodeError{status: http.StatusBadRequest, message: "request body must contain one valid JSON object", err: err}
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return &requestDecodeError{status: http.StatusBadRequest, message: "request body must contain one valid JSON object", err: errors.New("multiple JSON values")}
	}
	return nil
}

func requestErrorStatus(err error) int {
	var requestErr *requestDecodeError
	if errors.As(err, &requestErr) {
		return requestErr.status
	}
	return http.StatusBadRequest
}

func writeRequestDecodeError(w http.ResponseWriter, err error) {
	var requestErr *requestDecodeError
	if errors.As(err, &requestErr) {
		writeError(w, requestErr.status, requestErr.message)
		return
	}
	writeError(w, http.StatusBadRequest, "request body must contain one valid JSON object")
}

func writeStoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		writeError(w, http.StatusNotFound, "resource not found")
	case errors.Is(err, errValidation):
		writeError(w, http.StatusBadRequest, strings.TrimPrefix(err.Error(), errValidation.Error()+": "))
	case isUniqueViolation(err):
		writeError(w, http.StatusConflict, "that path is already in use — choose a different one")
	case strings.Contains(err.Error(), "required"), strings.Contains(err.Error(), "invalid"), strings.Contains(err.Error(), "mismatch"), strings.Contains(err.Error(), "already"), strings.Contains(err.Error(), "idempotency"):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		log.Printf("store operation failed error_type=%T", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
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
