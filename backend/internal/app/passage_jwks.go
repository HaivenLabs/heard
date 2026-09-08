package app

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type passageJWKSProvider struct {
	issuer, audience, jwksURL string
	client                    *http.Client
	now                       func() time.Time
	ttl                       time.Duration
	staleTTL                  time.Duration
	mu                        sync.RWMutex
	keys                      map[string]ed25519.PublicKey
	expires                   time.Time
	staleExpires              time.Time
	retryAfter                time.Time
	lastRefreshFailed         bool
	cache                     IdentityJWKSCache
}
type passageHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
	KeyID     string `json:"kid"`
}
type passageClaims struct {
	Issuer         string          `json:"iss"`
	Audience       json.RawMessage `json:"aud"`
	Subject        string          `json:"sub"`
	OrganizationID string          `json:"org_id"`
	Roles          []string        `json:"roles"`
	Products       []string        `json:"products"`
	ExpiresAt      int64           `json:"exp"`
	IssuedAt       int64           `json:"iat"`
}
type passageJWKS struct {
	Keys []struct {
		KeyType   string `json:"kty"`
		Curve     string `json:"crv"`
		Use       string `json:"use"`
		Algorithm string `json:"alg"`
		KeyID     string `json:"kid"`
		X         string `json:"x"`
	} `json:"keys"`
}

func newPassageJWKSProvider(cfg Config, client *http.Client, now func() time.Time) (*passageJWKSProvider, error) {
	if cfg.PassageIssuer == "" || cfg.PassageAudience != "heard" {
		return nil, errors.New("PASSAGE_ISSUER and PASSAGE_AUDIENCE=heard are required")
	}
	base, err := url.Parse(cfg.PassageBaseURL)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return nil, errors.New("PASSAGE_BASE_URL must be an absolute URL")
	}
	if cfg.IsProductionLike() && base.Scheme != "https" {
		return nil, errors.New("PASSAGE_BASE_URL must use https in staging and production")
	}
	issuerURL, issuerErr := url.Parse(cfg.PassageIssuer)
	if issuerErr != nil || issuerURL.Scheme == "" || issuerURL.Host == "" {
		return nil, errors.New("PASSAGE_ISSUER must be an absolute URL")
	}
	if cfg.IsProductionLike() && issuerURL.Scheme != "https" {
		return nil, errors.New("PASSAGE_ISSUER must use https in staging and production")
	}
	base.Path = "/.well-known/jwks.json"
	base.RawQuery = ""
	base.Fragment = ""
	if client == nil {
		trustedHost := base.Host
		client = &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 || req.URL.Host != trustedHost || (cfg.IsProductionLike() && req.URL.Scheme != "https") {
				return errors.New("untrusted Passage JWKS redirect")
			}
			return nil
		}}
	}
	ttl := time.Duration(cfg.PassageJWKSCacheSeconds) * time.Second
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	staleTTL := time.Duration(cfg.PassageJWKSStaleSeconds) * time.Second
	if staleTTL <= 0 {
		staleTTL = 15 * time.Minute
	}
	return &passageJWKSProvider{issuer: cfg.PassageIssuer, audience: "heard", jwksURL: base.String(), client: client, now: now, ttl: ttl, staleTTL: staleTTL, keys: map[string]ed25519.PublicKey{}}, nil
}
func (p *passageJWKSProvider) VerifyToken(ctx context.Context, token string) (Identity, error) {
	if len(token) > 16*1024 {
		return Identity{}, errInvalidToken
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Identity{}, errInvalidToken
	}
	headerRaw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return Identity{}, errInvalidToken
	}
	var h passageHeader
	if json.Unmarshal(headerRaw, &h) != nil || h.Algorithm != "EdDSA" || h.Type != "JWT" || h.KeyID == "" {
		return Identity{}, errInvalidToken
	}
	key, err := p.key(ctx, h.KeyID, false)
	if err != nil {
		if errors.Is(err, errIdentityUnavailable) {
			return Identity{}, err
		}
		return Identity{}, errInvalidToken
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !ed25519.Verify(key, []byte(parts[0]+"."+parts[1]), sig) {
		key, err = p.key(ctx, h.KeyID, true)
		if errors.Is(err, errIdentityUnavailable) {
			return Identity{}, err
		}
		if err != nil || !ed25519.Verify(key, []byte(parts[0]+"."+parts[1]), sig) {
			return Identity{}, errInvalidToken
		}
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Identity{}, errInvalidToken
	}
	var c passageClaims
	if json.Unmarshal(payload, &c) != nil || c.Issuer != p.issuer || c.Subject == "" || c.OrganizationID == "" || c.ExpiresAt <= p.now().Unix() || !audienceContains(c.Audience, p.audience) || !contains(c.Products, "heard") {
		return Identity{}, errInvalidToken
	}
	if _, err = uuid.Parse(c.Subject); err != nil {
		return Identity{}, errInvalidToken
	}
	if _, err = uuid.Parse(c.OrganizationID); err != nil {
		return Identity{}, errInvalidToken
	}
	role, ok := selectHeardRole(c.Roles)
	if !ok {
		return Identity{}, errInvalidToken
	}
	return Identity{UserID: c.Subject, Role: role, TenantIDs: []string{c.OrganizationID}, Permissions: heardPermissions(role), Provider: "passage"}, nil
}
func audienceContains(raw json.RawMessage, want string) bool {
	var one string
	if json.Unmarshal(raw, &one) == nil {
		return one == want
	}
	var many []string
	if json.Unmarshal(raw, &many) != nil {
		return false
	}
	return contains(many, want)
}
func contains(values []string, want string) bool {
	for _, v := range values {
		if v == want {
			return true
		}
	}
	return false
}
func selectHeardRole(roles []string) (string, bool) {
	for _, candidate := range []string{"owner", "admin", "manager", "member", "viewer"} {
		if contains(roles, candidate) {
			return candidate, true
		}
	}
	return "", false
}
func heardPermissions(role string) []string {
	switch role {
	case "owner":
		return []string{"tenant:read", "tenant:create", "tenant:write", "location:read", "location:write", "campaign:read", "campaign:write", "recovery:read", "recovery:write", "outbox:replay"}
	case "admin":
		return []string{"tenant:read", "tenant:create", "tenant:write", "location:read", "location:write", "campaign:read", "campaign:write", "recovery:read", "recovery:write"}
	case "manager":
		return []string{"tenant:read", "location:read", "location:write", "campaign:read", "campaign:write", "recovery:read", "recovery:write"}
	case "member":
		return []string{"tenant:read", "location:read", "campaign:read", "recovery:read", "recovery:write"}
	case "viewer":
		return []string{"tenant:read", "location:read", "campaign:read", "recovery:read"}
	}
	return nil
}
func (p *passageJWKSProvider) key(ctx context.Context, kid string, force bool) (ed25519.PublicKey, error) {
	p.mu.RLock()
	key, ok := p.keys[kid]
	now := p.now()
	fresh := now.Before(p.expires)
	stale := now.Before(p.staleExpires)
	retryPending := now.Before(p.retryAfter)
	p.mu.RUnlock()
	if ok && fresh && !force {
		return key, nil
	}
	// A verified, already-known key is safe to use for a bounded stale window:
	// token signature and expiry checks still run below. Unknown or changed keys
	// always require a successful refresh and therefore fail closed.
	if ok && stale && !force && retryPending {
		return key, nil
	}
	if err := p.refresh(ctx); err != nil {
		p.recordRefreshFailure()
		if ok && stale && !force {
			return key, nil
		}
		return nil, fmt.Errorf("%w: %v", errIdentityUnavailable, err)
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	key, ok = p.keys[kid]
	if !ok {
		return nil, errors.New("unknown signing key")
	}
	return key, nil
}

func (p *passageJWKSProvider) recordRefreshFailure() {
	p.mu.Lock()
	p.lastRefreshFailed = true
	// Avoid every authenticated request becoming a network probe during an
	// outage. A forced refresh remains available for a signature mismatch.
	p.retryAfter = p.now().Add(15 * time.Second)
	p.mu.Unlock()
}
func (p *passageJWKSProvider) refresh(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.jwksURL, nil)
	if err != nil {
		return err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("jwks status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	var set passageJWKS
	if json.Unmarshal(body, &set) != nil {
		return errors.New("invalid jwks")
	}
	keys := map[string]ed25519.PublicKey{}
	for _, j := range set.Keys {
		if j.KeyType != "OKP" || j.Curve != "Ed25519" || j.Algorithm != "EdDSA" || j.Use != "sig" || j.KeyID == "" {
			continue
		}
		raw, err := base64.RawURLEncoding.DecodeString(j.X)
		if err == nil && len(raw) == ed25519.PublicKeySize {
			keys[j.KeyID] = ed25519.PublicKey(raw)
		}
	}
	if len(keys) == 0 {
		return errors.New("jwks has no usable keys")
	}
	p.mu.Lock()
	p.keys = keys
	p.expires = p.now().Add(p.ttl)
	p.staleExpires = p.expires.Add(p.staleTTL)
	p.retryAfter = time.Time{}
	p.lastRefreshFailed = false
	cache := p.cache
	record := IdentityJWKSCacheRecord{Issuer: p.issuer, Keys: encodePublicKeys(keys), FreshUntil: p.expires, StaleUntil: p.staleExpires}
	p.mu.Unlock()
	if cache != nil {
		if err := cache.SaveIdentityJWKS(ctx, record); err != nil {
			// A cache write must not prevent verification with a freshly retrieved
			// public key; database health remains observable independently.
			log.Printf(`{"event":"identity.jwks_cache_persistence_failed","error_type":%q}`, fmt.Sprintf("%T", err))
		}
	}
	return nil
}

func (p *passageJWKSProvider) InitializeIdentityCache(ctx context.Context, cache IdentityJWKSCache) {
	p.mu.Lock()
	p.cache = cache
	p.mu.Unlock()
	record, found, err := cache.LoadIdentityJWKS(ctx, p.issuer)
	if err != nil {
		log.Printf(`{"event":"identity.jwks_cache_restore_failed","error_type":%q}`, fmt.Sprintf("%T", err))
		return
	}
	if !found || record.Issuer != p.issuer || !p.now().Before(record.StaleUntil) {
		return
	}
	keys := decodePublicKeys(record.Keys)
	if len(keys) == 0 {
		return
	}
	p.mu.Lock()
	p.keys = keys
	p.expires = record.FreshUntil
	p.staleExpires = record.StaleUntil
	p.lastRefreshFailed = !p.now().Before(record.FreshUntil)
	p.mu.Unlock()
}

func encodePublicKeys(keys map[string]ed25519.PublicKey) map[string]string {
	encoded := make(map[string]string, len(keys))
	for kid, key := range keys {
		encoded[kid] = base64.RawURLEncoding.EncodeToString(key)
	}
	return encoded
}

func decodePublicKeys(encoded map[string]string) map[string]ed25519.PublicKey {
	keys := make(map[string]ed25519.PublicKey, len(encoded))
	for kid, value := range encoded {
		raw, err := base64.RawURLEncoding.DecodeString(value)
		if err == nil && kid != "" && len(raw) == ed25519.PublicKeySize {
			keys[kid] = ed25519.PublicKey(raw)
		}
	}
	return keys
}

func (p *passageJWKSProvider) IdentityHealth() IdentityHealth {
	p.mu.RLock()
	defer p.mu.RUnlock()
	now := p.now()
	status := "unknown"
	if len(p.keys) > 0 {
		switch {
		case now.Before(p.expires):
			status = "operational"
		case now.Before(p.staleExpires):
			status = "degraded"
		default:
			status = "unavailable"
		}
	} else if p.lastRefreshFailed {
		status = "unavailable"
	}
	return IdentityHealth{Status: status, CachedKeyCount: len(p.keys), FreshUntil: p.expires, StaleUntil: p.staleExpires, LastRefreshFailed: p.lastRefreshFailed}
}

// RefreshIdentityHealth lets a health probe detect an identity outage without
// making every probe a network call. It observes the same bounded retry window
// used by session verification.
func (p *passageJWKSProvider) RefreshIdentityHealth(ctx context.Context) {
	p.mu.RLock()
	now := p.now()
	fresh := now.Before(p.expires)
	retryPending := now.Before(p.retryAfter)
	p.mu.RUnlock()
	if fresh || retryPending {
		return
	}
	if err := p.refresh(ctx); err != nil {
		p.recordRefreshFailure()
	}
}
