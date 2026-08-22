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

func TestRegistrationHandoffRequiresWorkspaceDetails(t *testing.T) {
	s := NewServer(Config{AppEnv: "test", PassageBaseURL: "http://identity.internal", PassagePublicURL: "https://auth.heard.example", PassageCallbackURL: "https://heard.example/api/v1/auth/callback", PassageClientID: "heard", WebBaseURL: "https://heard.example"}, nil, stubIdentityProvider{})

	for _, target := range []string{
		"/api/v1/auth/start?provider=google&intent=register&location_name=Downtown",
		"/api/v1/auth/start?provider=google&intent=register&organization_name=Cedar%20Cafe",
	} {
		rec := httptest.NewRecorder()
		s.handlePassageStart(rec, httptest.NewRequest(http.MethodGet, target, nil))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("%s status=%d, want %d", target, rec.Code, http.StatusBadRequest)
		}
		if rec.Header().Get("Location") != "" {
			t.Fatalf("incomplete registration redirected to identity provider: %s", rec.Header().Get("Location"))
		}
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
