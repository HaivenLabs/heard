package app

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestPassageBrowserHandoffUsesPKCEAndHttpOnlySession(t *testing.T) {
	now := time.Now().UTC()
	keys := &rotatingKeys{}
	keys.rotate()
	var exchanges atomic.Int32
	var passage *httptest.Server
	passage = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/jwks.json":
			keys.serve(w, r)
		case "/api/v1/oauth/token":
			exchanges.Add(1)
			raw, _ := io.ReadAll(r.Body)
			var in map[string]string
			_ = json.Unmarshal(raw, &in)
			if in["grant_type"] != "authorization_code" || in["client_id"] != "heard" || in["redirect_uri"] != "http://heard.test/api/v1/auth/callback" || len(in["code_verifier"]) < 43 {
				http.Error(w, "bad exchange", 400)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": keys.token(t, baseClaims(now)), "token_type": "Bearer", "expires_in": 300})
		default:
			http.NotFound(w, r)
		}
	}))
	defer passage.Close()
	provider := providerFor(t, passage.URL, func() time.Time { return now })
	s := NewServer(Config{AppEnv: "test", PassageBaseURL: passage.URL, PassagePublicURL: "http://localhost:3020", PassageCallbackURL: "http://heard.test/api/v1/auth/callback", PassageClientID: "heard", WebBaseURL: "http://heard.test"}, nil, provider)
	startReq := httptest.NewRequest(http.MethodGet, "http://heard.test/api/v1/auth/start?return_to=%2Fadmin%2Frecovery&intent=register&email=manager%40nom.com&organization_name=Manager%27s+Cafe&location_name=Downtown", nil)
	startRec := httptest.NewRecorder()
	s.handlePassageStart(startRec, startReq)
	if startRec.Code != 302 {
		t.Fatalf("start=%d", startRec.Code)
	}
	authorize, err := url.Parse(startRec.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	if authorize.Scheme != "http" || authorize.Host != "localhost:3020" || authorize.Path != "/api/v1/oauth/authorize" || authorize.Query().Get("code_challenge_method") != "S256" || authorize.Query().Get("redirect_uri") != "http://heard.test/api/v1/auth/callback" || authorize.Query().Get("intent") != "register" || authorize.Query().Get("organization_name") != "Manager's Cafe" || authorize.Query().Get("login_hint") != "manager@nom.com" || strings.Contains(authorize.String(), "access_token") {
		t.Fatalf("unsafe authorize URL: %s", authorize)
	}
	var state, verifier, returnValue *http.Cookie
	for _, c := range startRec.Result().Cookies() {
		if !c.HttpOnly || c.SameSite != http.SameSiteLaxMode {
			t.Fatalf("unsafe transient cookie %#v", c)
		}
		switch c.Name {
		case "heard_oauth_state":
			state = c
		case "heard_oauth_verifier":
			verifier = c
		case "heard_oauth_return":
			returnValue = c
		}
	}
	if state == nil || verifier == nil || returnValue == nil {
		t.Fatal("missing PKCE cookies")
	}
	callbackReq := httptest.NewRequest(http.MethodGet, "http://heard.test/api/v1/auth/callback?code=one-time-code&state="+url.QueryEscape(state.Value), nil)
	callbackReq.AddCookie(state)
	callbackReq.AddCookie(verifier)
	callbackReq.AddCookie(returnValue)
	callbackRec := httptest.NewRecorder()
	s.handlePassageCallback(callbackRec, callbackReq)
	if callbackRec.Code != 302 || callbackRec.Header().Get("Location") != "http://heard.test/admin/recovery" {
		t.Fatalf("callback=%d location=%s", callbackRec.Code, callbackRec.Header().Get("Location"))
	}
	if exchanges.Load() != 1 {
		t.Fatalf("exchanges=%d", exchanges.Load())
	}
	var session *http.Cookie
	for _, c := range callbackRec.Result().Cookies() {
		if c.Name == "heard_session" {
			session = c
		}
	}
	if session == nil || !session.HttpOnly || session.SameSite != http.SameSiteLaxMode || session.Value == "" {
		t.Fatalf("unsafe session cookie %#v", session)
	}
	handler := s.withIdentity("campaign:read", func(w http.ResponseWriter, _ *http.Request, _ actorContext) { w.WriteHeader(204) })
	gatewayReq := httptest.NewRequest(http.MethodGet, "/api/v1/survey-campaigns", nil)
	gatewayReq.AddCookie(session)
	gatewayRec := httptest.NewRecorder()
	handler.ServeHTTP(gatewayRec, gatewayReq)
	if gatewayRec.Code != 204 {
		t.Fatalf("cookie gateway=%d", gatewayRec.Code)
	}
}

func TestPassageCallbackRejectsStateMismatchWithoutExchange(t *testing.T) {
	var exchanges atomic.Int32
	passage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { exchanges.Add(1) }))
	defer passage.Close()
	s := NewServer(Config{AppEnv: "test", PassageBaseURL: passage.URL, PassageCallbackURL: "http://heard.test/api/v1/auth/callback", PassageClientID: "heard", WebBaseURL: "http://heard.test"}, nil, stubIdentityProvider{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/callback?code=code&state=attacker", nil)
	req.AddCookie(&http.Cookie{Name: "heard_oauth_state", Value: "expected"})
	req.AddCookie(&http.Cookie{Name: "heard_oauth_verifier", Value: strings.Repeat("v", 43)})
	rec := httptest.NewRecorder()
	s.handlePassageCallback(rec, req)
	if rec.Code != 400 || exchanges.Load() != 0 {
		t.Fatalf("state mismatch status=%d exchanges=%d", rec.Code, exchanges.Load())
	}
}

func TestPassageCallbackRecoversExpiredProviderSessionWithoutState(t *testing.T) {
	s := NewServer(Config{AppEnv: "test", WebBaseURL: "http://heard.test"}, nil, stubIdentityProvider{})
	rec := httptest.NewRecorder()
	s.handlePassageCallback(rec, httptest.NewRequest(http.MethodGet, "/api/v1/auth/callback?error=provider_session_expired", nil))
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "http://heard.test/login?auth_notice=session_expired" {
		t.Fatalf("expired provider session status=%d location=%s", rec.Code, rec.Header().Get("Location"))
	}
}

func TestPassageCallbackRedirectsProviderErrorsToBrandedAuthPages(t *testing.T) {
	tests := []struct {
		name, providerError, intent, destination string
		startsRegistration                       bool
	}{
		{"login canceled", "access_denied", "login", "http://heard.test/login?auth_notice=cancelled", false},
		{"signup canceled", "access_denied", "register", "http://heard.test/start?auth_notice=cancelled", false},
		{"unknown login continues into Google registration", "identity_not_registered", "login", "", true},
		{"provider failure", "provider_failed", "login", "http://heard.test/login?auth_notice=provider_error", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := NewServer(Config{AppEnv: "test", WebBaseURL: "http://heard.test"}, nil, stubIdentityProvider{})
			req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/callback?error="+url.QueryEscape(test.providerError)+"&state=expected", nil)
			for _, cookie := range []*http.Cookie{
				{Name: "heard_oauth_state", Value: "expected"},
				{Name: "heard_oauth_verifier", Value: strings.Repeat("v", 43)},
				{Name: "heard_oauth_return", Value: base64.RawURLEncoding.EncodeToString([]byte("/admin"))},
				{Name: "heard_oauth_intent", Value: test.intent},
			} {
				req.AddCookie(cookie)
			}
			rec := httptest.NewRecorder()
			s.handlePassageCallback(rec, req)
			if rec.Code != http.StatusFound {
				t.Fatalf("status=%d location=%s", rec.Code, rec.Header().Get("Location"))
			}
			if test.startsRegistration {
				destination, err := url.Parse(rec.Header().Get("Location"))
				if err != nil || destination.Path != "/api/v1/auth/start" || destination.Query().Get("provider") != "google" || destination.Query().Get("intent") != "register" || destination.Query().Get("return_to") != "/onboarding?auth_notice=google_account_created" {
					t.Fatalf("registration continuation=%s", rec.Header().Get("Location"))
				}
				return
			}
			if rec.Header().Get("Location") != test.destination {
				t.Fatalf("status=%d location=%s", rec.Code, rec.Header().Get("Location"))
			}
		})
	}
}

func TestPassageStartRejectsOpenReturnRedirect(t *testing.T) {
	s := NewServer(Config{AppEnv: "test", PassageBaseURL: "http://passage.test", PassageCallbackURL: "http://heard.test/api/v1/auth/callback", PassageClientID: "heard", WebBaseURL: "http://heard.test"}, nil, stubIdentityProvider{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/start?return_to=https%3A%2F%2Fevil.example%2Fpwn", nil)
	rec := httptest.NewRecorder()
	s.handlePassageStart(rec, req)
	var encoded string
	for _, c := range rec.Result().Cookies() {
		if c.Name == "heard_oauth_return" {
			encoded = c.Value
		}
	}
	decoded, _ := base64.RawURLEncoding.DecodeString(encoded)
	if string(decoded) != "/onboarding" {
		t.Fatalf("unsafe return path %q", decoded)
	}
}

func TestGoogleHandoffSkipsPassageUIAndPreservesHeardPKCE(t *testing.T) {
	s := NewServer(Config{AppEnv: "test", PassageBaseURL: "http://identity.internal", PassagePublicURL: "https://auth.heard.example", PassageCallbackURL: "https://heard.example/api/v1/auth/callback", PassageClientID: "heard", WebBaseURL: "https://heard.example"}, nil, stubIdentityProvider{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/start?provider=google&intent=login&return_to=%2Fadmin", nil)
	rec := httptest.NewRecorder()
	s.handlePassageStart(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("start=%d body=%s", rec.Code, rec.Body.String())
	}
	destination, err := url.Parse(rec.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	if destination.Host != "auth.heard.example" || destination.Path != "/api/v1/auth/external/google/start" {
		t.Fatalf("Google handoff exposed the platform UI: %s", destination)
	}
	continuation, err := url.Parse(destination.Query().Get("return_to"))
	if err != nil {
		t.Fatal(err)
	}
	if continuation.Path != "/api/v1/oauth/authorize" || continuation.Query().Get("client_id") != "heard" || continuation.Query().Get("code_challenge_method") != "S256" || continuation.Query().Get("intent") != "login" {
		t.Fatalf("invalid headless continuation: %s", continuation)
	}
}

func TestRegistrationHandoffDefersWorkspaceDetailsUntilOnboarding(t *testing.T) {
	s := NewServer(Config{AppEnv: "test", PassageBaseURL: "http://identity.internal", PassagePublicURL: "https://auth.heard.example", PassageCallbackURL: "https://heard.example/api/v1/auth/callback", PassageClientID: "heard", WebBaseURL: "https://heard.example"}, nil, stubIdentityProvider{})

	rec := httptest.NewRecorder()
	s.handlePassageStart(rec, httptest.NewRequest(http.MethodGet, "/api/v1/auth/start?provider=google&intent=register&return_to=%2Fonboarding", nil))
	if rec.Code != http.StatusFound {
		t.Fatalf("registration start status=%d", rec.Code)
	}
	destination, err := url.Parse(rec.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	continuation, err := url.Parse(destination.Query().Get("return_to"))
	if err != nil || continuation.Query().Get("organization_name") != "Heard workspace" {
		t.Fatalf("registration did not use a neutral workspace context: %s", destination)
	}
}

func TestIdentityProvidersExposeOnlyConfiguredConsumerMethods(t *testing.T) {
	identity := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/auth/providers" || r.URL.Query().Get("client_id") != "heard" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"providers": []string{"google", "internal-test-provider"}})
	}))
	defer identity.Close()
	s := NewServer(Config{AppEnv: "test", PassageBaseURL: identity.URL, PassageClientID: "heard"}, nil, stubIdentityProvider{})
	rec := httptest.NewRecorder()
	s.handleIdentityProviders(rec, httptest.NewRequest(http.MethodGet, "/api/v1/auth/providers", nil))
	if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != `{"providers":["google"]}` {
		t.Fatalf("provider discovery status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestHeardRegistrationForwardsOnlyIdentityCredentials(t *testing.T) {
	var received map[string]string
	identity := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/auth/register" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Idempotency-Key") == "" {
			http.Error(w, "missing idempotency key", http.StatusBadRequest)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{"email": received["email"], "verification_delivery": "requested"})
	}))
	defer identity.Close()

	s := NewServer(Config{AppEnv: "test", PassageBaseURL: identity.URL, PassageClientID: "heard"}, nil, stubIdentityProvider{})
	s.limiter = nil
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{"email":"owner@example.com","password":"a secure password","source":"homepage"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated || received["email"] != "owner@example.com" || received["password"] != "a secure password" || len(received) != 2 {
		t.Fatalf("registration status=%d received=%#v", rec.Code, received)
	}
}

func TestHeardLoginExchangesIdentitySessionForHeardCookie(t *testing.T) {
	identity := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/auth/login":
			var received map[string]string
			if err := json.NewDecoder(r.Body).Decode(&received); err != nil || received["email"] != "owner@example.com" || received["password"] != "a secure password" {
				http.Error(w, "invalid login", http.StatusBadRequest)
				return
			}
			http.SetCookie(w, &http.Cookie{Name: "passage_session", Value: "identity-session", Path: "/", HttpOnly: true})
			http.SetCookie(w, &http.Cookie{Name: "passage_csrf", Value: "identity-csrf", Path: "/", HttpOnly: true})
			w.WriteHeader(http.StatusOK)
		case "/api/v1/product-token":
			if !strings.Contains(r.Header.Get("Cookie"), "passage_session=identity-session") || r.Header.Get("X-CSRF-Token") != "identity-csrf" {
				http.Error(w, "missing identity session", http.StatusUnauthorized)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"access_token": "heard-token", "token_type": "Bearer", "expires_in": 300})
		default:
			http.NotFound(w, r)
		}
	}))
	defer identity.Close()

	s := NewServer(Config{AppEnv: "test", PassageBaseURL: identity.URL, PassageClientID: "heard"}, nil, stubIdentityProvider{})
	s.limiter = nil
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"owner@example.com","password":"a secure password"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != `{"signed_in":true}` {
		t.Fatalf("login status=%d body=%s", rec.Code, rec.Body.String())
	}
	var heardSession *http.Cookie
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == "passage_session" || cookie.Name == "passage_csrf" {
			t.Fatalf("identity cookie leaked to browser: %#v", cookie)
		}
		if cookie.Name == "heard_session" {
			heardSession = cookie
		}
	}
	if heardSession == nil || heardSession.Value != "heard-token" || !heardSession.HttpOnly || heardSession.SameSite != http.SameSiteLaxMode {
		t.Fatalf("unsafe heard session: %#v", heardSession)
	}
}

func TestConsumerAuthRoutesDoNotExposeProviderNames(t *testing.T) {
	s := NewServer(Config{AppEnv: "test", PassageBaseURL: "://invalid", PassageCallbackURL: "http://heard.test/api/v1/auth/callback", PassageClientID: "heard", WebBaseURL: "http://heard.test"}, nil, stubIdentityProvider{})

	rec := httptest.NewRecorder()
	s.Router().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/auth/start", nil))
	if strings.Contains(strings.ToLower(rec.Body.String()), "passage") {
		t.Fatalf("provider name leaked through heard auth error: %s", rec.Body.String())
	}

	legacy := httptest.NewRecorder()
	s.Router().ServeHTTP(legacy, httptest.NewRequest(http.MethodGet, "/api/v1/auth/passage/start", nil))
	if legacy.Code != http.StatusNotFound {
		t.Fatalf("provider-named consumer route remains exposed: %d", legacy.Code)
	}
}
