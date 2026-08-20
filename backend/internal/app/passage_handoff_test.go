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
	startReq := httptest.NewRequest(http.MethodGet, "http://heard.test/api/v1/auth/start?return_to=%2Fadmin%2Frecovery&intent=register&email=manager%40nom.com", nil)
	startRec := httptest.NewRecorder()
	s.handlePassageStart(startRec, startReq)
	if startRec.Code != 302 {
		t.Fatalf("start=%d", startRec.Code)
	}
	authorize, err := url.Parse(startRec.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	if authorize.Scheme != "http" || authorize.Host != "localhost:3020" || authorize.Path != "/api/v1/oauth/authorize" || authorize.Query().Get("code_challenge_method") != "S256" || authorize.Query().Get("redirect_uri") != "http://heard.test/api/v1/auth/callback" || authorize.Query().Get("intent") != "register" || authorize.Query().Get("login_hint") != "manager@nom.com" || strings.Contains(authorize.String(), "access_token") {
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
