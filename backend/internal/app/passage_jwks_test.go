package app

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type rotatingKeys struct {
	mu   sync.RWMutex
	pub  ed25519.PublicKey
	priv ed25519.PrivateKey
	hits atomic.Int32
}

func (k *rotatingKeys) rotate() {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	k.mu.Lock()
	k.pub, k.priv = pub, priv
	k.mu.Unlock()
}
func (k *rotatingKeys) serve(w http.ResponseWriter, _ *http.Request) {
	k.hits.Add(1)
	k.mu.RLock()
	x := base64.RawURLEncoding.EncodeToString(k.pub)
	k.mu.RUnlock()
	_ = json.NewEncoder(w).Encode(map[string]any{"keys": []any{map[string]string{"kty": "OKP", "crv": "Ed25519", "use": "sig", "alg": "EdDSA", "kid": "passage-ed25519-1", "x": x}}})
}
func (k *rotatingKeys) token(t *testing.T, claims map[string]any) string {
	t.Helper()
	header, _ := json.Marshal(map[string]string{"alg": "EdDSA", "typ": "JWT", "kid": "passage-ed25519-1"})
	payload, _ := json.Marshal(claims)
	signed := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	k.mu.RLock()
	sig := ed25519.Sign(k.priv, []byte(signed))
	k.mu.RUnlock()
	return signed + "." + base64.RawURLEncoding.EncodeToString(sig)
}
func baseClaims(now time.Time) map[string]any {
	return map[string]any{"iss": "https://passage.test", "aud": "heard", "sub": "11111111-1111-4111-8111-111111111111", "org_id": "22222222-2222-4222-8222-222222222222", "roles": []string{"manager"}, "products": []string{"heard"}, "iat": now.Unix(), "exp": now.Add(5 * time.Minute).Unix()}
}
func providerFor(t *testing.T, server string, now func() time.Time) *passageJWKSProvider {
	t.Helper()
	p, err := newPassageJWKSProvider(Config{AppEnv: "test", PassageBaseURL: server, PassageIssuer: "https://passage.test", PassageAudience: "heard", PassageJWKSCacheSeconds: 300}, nil, now)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestPassageVerifierMapsTrustedOrganizationRoleAndPermissions(t *testing.T) {
	now := time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC)
	keys := &rotatingKeys{}
	keys.rotate()
	server := httptest.NewServer(http.HandlerFunc(keys.serve))
	defer server.Close()
	p := providerFor(t, server.URL, func() time.Time { return now })
	identity, err := p.VerifyToken(context.Background(), keys.token(t, baseClaims(now)))
	if err != nil {
		t.Fatal(err)
	}
	if identity.Provider != "passage" || identity.Role != "manager" || !identity.HasTenant("22222222-2222-4222-8222-222222222222") || identity.HasTenant("33333333-3333-4333-8333-333333333333") {
		t.Fatalf("bad identity: %#v", identity)
	}
	if !identity.HasPermission("campaign:write") || identity.HasPermission("tenant:create") {
		t.Fatalf("bad permission mapping: %#v", identity.Permissions)
	}
}

func TestPassageVerifierRejectsExpiryAudienceTenantAndProductAccess(t *testing.T) {
	now := time.Now().UTC()
	keys := &rotatingKeys{}
	keys.rotate()
	server := httptest.NewServer(http.HandlerFunc(keys.serve))
	defer server.Close()
	tests := []func(map[string]any){func(c map[string]any) { c["exp"] = now.Add(-time.Second).Unix() }, func(c map[string]any) { c["aud"] = "qurl" }, func(c map[string]any) { c["org_id"] = "not-a-uuid" }, func(c map[string]any) { c["products"] = []string{"qurl"} }}
	for n, mutate := range tests {
		claims := baseClaims(now)
		mutate(claims)
		p := providerFor(t, server.URL, func() time.Time { return now })
		if _, err := p.VerifyToken(context.Background(), keys.token(t, claims)); err == nil {
			t.Fatalf("case %d accepted", n)
		}
	}
}

func TestPassageVerifierCachesJWKSAndRefreshesOnRotation(t *testing.T) {
	now := time.Now().UTC()
	keys := &rotatingKeys{}
	keys.rotate()
	server := httptest.NewServer(http.HandlerFunc(keys.serve))
	defer server.Close()
	p := providerFor(t, server.URL, func() time.Time { return now })
	if _, err := p.VerifyToken(context.Background(), keys.token(t, baseClaims(now))); err != nil {
		t.Fatal(err)
	}
	if _, err := p.VerifyToken(context.Background(), keys.token(t, baseClaims(now))); err != nil {
		t.Fatal(err)
	}
	if got := keys.hits.Load(); got != 1 {
		t.Fatalf("cache fetched %d times", got)
	}
	keys.rotate()
	if _, err := p.VerifyToken(context.Background(), keys.token(t, baseClaims(now))); err != nil {
		t.Fatalf("rotation rejected: %v", err)
	}
	if got := keys.hits.Load(); got != 2 {
		t.Fatalf("rotation fetches=%d", got)
	}
}

func TestPassageVerifierFailsClosedWhenJWKSUnavailable(t *testing.T) {
	now := time.Now().UTC()
	keys := &rotatingKeys{}
	keys.rotate()
	server := httptest.NewServer(http.HandlerFunc(keys.serve))
	p := providerFor(t, server.URL, func() time.Time { return now })
	token := keys.token(t, baseClaims(now))
	server.Close()
	if _, err := p.VerifyToken(context.Background(), token); !errors.Is(err, errIdentityUnavailable) {
		t.Fatal("unavailable uncached JWKS accepted")
	}
}

func TestPassageVerifierUsesBoundedStaleCacheDuringPassageOutage(t *testing.T) {
	now := time.Now().UTC()
	clock := now
	keys := &rotatingKeys{}
	keys.rotate()
	server := httptest.NewServer(http.HandlerFunc(keys.serve))
	p, err := newPassageJWKSProvider(Config{AppEnv: "test", PassageBaseURL: server.URL, PassageIssuer: "https://passage.test", PassageAudience: "heard", PassageJWKSCacheSeconds: 1, PassageJWKSStaleSeconds: 5}, nil, func() time.Time { return clock })
	if err != nil {
		t.Fatal(err)
	}
	token := keys.token(t, baseClaims(now))
	if _, err = p.VerifyToken(context.Background(), token); err != nil {
		t.Fatal(err)
	}
	server.Close()
	if _, err = p.VerifyToken(context.Background(), token); err != nil {
		t.Fatalf("fresh cache rejected: %v", err)
	}
	clock = clock.Add(2 * time.Second)
	if _, err = p.VerifyToken(context.Background(), token); err != nil {
		t.Fatalf("bounded stale cache rejected an unexpired session: %v", err)
	}
	if health := p.IdentityHealth(); health.Status != "degraded" || !health.LastRefreshFailed {
		t.Fatalf("unexpected outage health: %#v", health)
	}
	clock = clock.Add(5 * time.Second)
	if _, err = p.VerifyToken(context.Background(), token); !errors.Is(err, errIdentityUnavailable) {
		t.Fatalf("stale cache survived its bounded outage window: %v", err)
	}
}

func TestPassageVerifierRejectsExpiredTokenDuringStaleKeyFallback(t *testing.T) {
	now := time.Now().UTC()
	clock := now
	keys := &rotatingKeys{}
	keys.rotate()
	server := httptest.NewServer(http.HandlerFunc(keys.serve))
	p, err := newPassageJWKSProvider(Config{AppEnv: "test", PassageBaseURL: server.URL, PassageIssuer: "https://passage.test", PassageAudience: "heard", PassageJWKSCacheSeconds: 1, PassageJWKSStaleSeconds: 60}, nil, func() time.Time { return clock })
	if err != nil {
		t.Fatal(err)
	}
	claims := baseClaims(now)
	claims["exp"] = now.Add(time.Second).Unix()
	token := keys.token(t, claims)
	if _, err = p.VerifyToken(context.Background(), token); err != nil {
		t.Fatal(err)
	}
	server.Close()
	clock = clock.Add(2 * time.Second)
	if _, err = p.VerifyToken(context.Background(), token); !errors.Is(err, errInvalidToken) {
		t.Fatalf("expired token accepted during identity outage: %v", err)
	}
}

func TestPassageHealthProbeReportsUnavailableIdentityWithoutKeys(t *testing.T) {
	p, err := newPassageJWKSProvider(Config{AppEnv: "test", PassageBaseURL: "http://127.0.0.1:1", PassageIssuer: "https://passage.test", PassageAudience: "heard", PassageJWKSCacheSeconds: 60, PassageJWKSStaleSeconds: 60}, &http.Client{Timeout: 10 * time.Millisecond}, time.Now)
	if err != nil {
		t.Fatal(err)
	}
	p.RefreshIdentityHealth(context.Background())
	if health := p.IdentityHealth(); health.Status != "unavailable" || !health.LastRefreshFailed {
		t.Fatalf("unexpected initial outage health: %#v", health)
	}
}

func TestPassageVerifierRestoresBoundedPublicKeyCacheAfterRestart(t *testing.T) {
	now := time.Now().UTC()
	keys := &rotatingKeys{}
	keys.rotate()
	server := httptest.NewServer(http.HandlerFunc(keys.serve))
	cache := &memoryIdentityJWKSCache{}
	config := Config{AppEnv: "test", PassageBaseURL: server.URL, PassageIssuer: "https://passage.test", PassageAudience: "heard", PassageJWKSCacheSeconds: 60, PassageJWKSStaleSeconds: 60}
	first, err := newPassageJWKSProvider(config, nil, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	first.InitializeIdentityCache(context.Background(), cache)
	token := keys.token(t, baseClaims(now))
	if _, err = first.VerifyToken(context.Background(), token); err != nil {
		t.Fatal(err)
	}
	server.Close()

	restarted, err := newPassageJWKSProvider(config, nil, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	restarted.InitializeIdentityCache(context.Background(), cache)
	if _, err = restarted.VerifyToken(context.Background(), token); err != nil {
		t.Fatalf("restored public key cache rejected valid session: %v", err)
	}
}

type memoryIdentityJWKSCache struct {
	record IdentityJWKSCacheRecord
	found  bool
}

func (c *memoryIdentityJWKSCache) LoadIdentityJWKS(_ context.Context, _ string) (IdentityJWKSCacheRecord, bool, error) {
	return c.record, c.found, nil
}

func (c *memoryIdentityJWKSCache) SaveIdentityJWKS(_ context.Context, record IdentityJWKSCacheRecord) error {
	c.record = record
	c.found = true
	return nil
}

func TestProductionRequiresRemotePassage(t *testing.T) {
	if _, err := NewIdentityProvider(Config{AppEnv: "production", PassageMode: "local", LocalPassageSecret: "long-enough-secret"}); err == nil {
		t.Fatal("production local adapter accepted")
	}
}
