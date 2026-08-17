package app

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	errUnauthenticated = errors.New("authentication required")
	errInvalidToken    = errors.New("invalid access token")
)

type Identity struct {
	UserID      string   `json:"user_id"`
	Email       string   `json:"email"`
	DisplayName string   `json:"display_name"`
	Role        string   `json:"role"`
	TenantIDs   []string `json:"tenant_ids"`
	Permissions []string `json:"permissions"`
	Provider    string   `json:"provider"`
}

func (i Identity) HasTenant(tenantID string) bool {
	for _, candidate := range i.TenantIDs {
		if subtle.ConstantTimeCompare([]byte(candidate), []byte(tenantID)) == 1 {
			return true
		}
	}
	return false
}

func (i Identity) HasPermission(permission string) bool {
	for _, candidate := range i.Permissions {
		if candidate == permission || candidate == "*" {
			return true
		}
	}
	return false
}

type Session struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
	Identity    Identity  `json:"identity"`
	TenantID    string    `json:"tenant_id"`
}

type IdentityProvider interface {
	VerifyToken(ctx context.Context, accessToken string) (Identity, error)
}

type localSessionIssuer interface {
	IssueSession(email string) (Session, error)
}

type localRegistrationIssuer interface {
	IssueRegistration(email string) (Session, error)
}

func NewIdentityProvider(cfg Config) (IdentityProvider, error) {
	switch cfg.PassageMode {
	case "jwks":
		return newPassageJWKSProvider(cfg, nil, time.Now)
	case "local":
		if cfg.AppEnv == "production" {
			return nil, errors.New("local Passage adapter cannot run in production")
		}
		if len(cfg.LocalPassageSecret) < 12 {
			return nil, errors.New("LOCAL_PASSAGE_SECRET must be at least 12 characters")
		}
		return newLocalPassageProvider(cfg.LocalPassageSecret, time.Now), nil
	default:
		return nil, fmt.Errorf("unsupported PASSAGE_MODE %q", cfg.PassageMode)
	}
}

type localPassageProvider struct {
	secret []byte
	now    func() time.Time
}

type localPassageClaims struct {
	Subject     string   `json:"sub"`
	Email       string   `json:"email"`
	DisplayName string   `json:"name"`
	Role        string   `json:"role"`
	TenantIDs   []string `json:"tenant_ids"`
	Permissions []string `json:"permissions"`
	ExpiresAt   int64    `json:"exp"`
	Issuer      string   `json:"iss"`
}

func newLocalPassageProvider(secret string, now func() time.Time) *localPassageProvider {
	return &localPassageProvider{secret: []byte(secret), now: now}
}

func (p *localPassageProvider) IssueSession(rawEmail string) (Session, error) {
	return p.issueSession(rawEmail, demoTenantID)
}

func (p *localPassageProvider) IssueRegistration(rawEmail string) (Session, error) {
	email := strings.ToLower(strings.TrimSpace(rawEmail))
	if email == "" || !strings.Contains(email, "@") {
		return Session{}, errors.New("a valid email is required")
	}
	tenantID := uuid.NewSHA1(uuid.NameSpaceURL, []byte("local-passage-workspace:"+email)).String()
	return p.issueSession(email, tenantID)
}

func (p *localPassageProvider) issueSession(rawEmail, tenantID string) (Session, error) {
	email := strings.ToLower(strings.TrimSpace(rawEmail))
	if email == "" || !strings.Contains(email, "@") {
		return Session{}, errors.New("a valid email is required")
	}

	now := p.now().UTC()
	expiresAt := now.Add(8 * time.Hour)
	identity := Identity{
		UserID:      uuid.NewSHA1(uuid.NameSpaceURL, []byte("local-passage:"+email)).String(),
		Email:       email,
		DisplayName: displayNameFromEmail(email),
		Role:        "owner",
		TenantIDs:   []string{tenantID},
		Permissions: []string{"tenant:read", "tenant:create", "location:read", "location:write", "campaign:read", "campaign:write", "recovery:read", "recovery:write"},
		Provider:    "passage-local",
	}
	claims := localPassageClaims{
		Subject:     identity.UserID,
		Email:       identity.Email,
		DisplayName: identity.DisplayName,
		Role:        identity.Role,
		TenantIDs:   identity.TenantIDs,
		Permissions: identity.Permissions,
		ExpiresAt:   expiresAt.Unix(),
		Issuer:      "heard-local-passage",
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return Session{}, fmt.Errorf("encode local Passage claims: %w", err)
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	token := "local." + encodedPayload + "." + p.sign(encodedPayload)

	return Session{AccessToken: token, ExpiresAt: expiresAt, Identity: identity, TenantID: tenantID}, nil
}

func (p *localPassageProvider) VerifyToken(_ context.Context, accessToken string) (Identity, error) {
	parts := strings.Split(accessToken, ".")
	if len(parts) != 3 || parts[0] != "local" {
		return Identity{}, errInvalidToken
	}
	expected := p.sign(parts[1])
	if subtle.ConstantTimeCompare([]byte(parts[2]), []byte(expected)) != 1 {
		return Identity{}, errInvalidToken
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Identity{}, errInvalidToken
	}
	var claims localPassageClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return Identity{}, errInvalidToken
	}
	if claims.Issuer != "heard-local-passage" || claims.Subject == "" || claims.ExpiresAt <= p.now().UTC().Unix() {
		return Identity{}, errInvalidToken
	}
	return Identity{
		UserID:      claims.Subject,
		Email:       claims.Email,
		DisplayName: claims.DisplayName,
		Role:        claims.Role,
		TenantIDs:   claims.TenantIDs,
		Permissions: claims.Permissions,
		Provider:    "passage-local",
	}, nil
}

func (p *localPassageProvider) sign(payload string) string {
	mac := hmac.New(sha256.New, p.secret)
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func displayNameFromEmail(email string) string {
	name := strings.SplitN(email, "@", 2)[0]
	name = strings.NewReplacer(".", " ", "_", " ", "-", " ").Replace(name)
	words := strings.Fields(name)
	for index, word := range words {
		words[index] = strings.ToUpper(word[:1]) + word[1:]
	}
	if len(words) == 0 {
		return "Restaurant owner"
	}
	return strings.Join(words, " ")
}
