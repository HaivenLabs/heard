package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLocalPassageIssuesAndVerifiesSession(t *testing.T) {
	provider := newLocalPassageProvider("test-secret", func() time.Time {
		return time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)
	})

	session, err := provider.IssueSession("owner@northstar.test")
	if err != nil {
		t.Fatalf("issue session: %v", err)
	}
	identity, err := provider.VerifyToken(context.Background(), session.AccessToken)
	if err != nil {
		t.Fatalf("verify token: %v", err)
	}

	if identity.Email != "owner@northstar.test" {
		t.Fatalf("got email %q", identity.Email)
	}
	if !identity.HasTenant(demoTenantID) {
		t.Fatalf("expected identity to belong to demo tenant")
	}
	if !identity.HasPermission("campaign:write") {
		t.Fatalf("expected campaign:write permission")
	}
}

func TestLocalPassageRejectsTamperedToken(t *testing.T) {
	provider := newLocalPassageProvider("test-secret", time.Now)
	session, err := provider.IssueSession("owner@northstar.test")
	if err != nil {
		t.Fatalf("issue session: %v", err)
	}

	if _, err := provider.VerifyToken(context.Background(), session.AccessToken+"tampered"); err == nil {
		t.Fatal("expected tampered token to be rejected")
	}
}

func TestAdminContextRequiresBearerToken(t *testing.T) {
	server := &Server{identity: stubIdentityProvider{err: errUnauthenticated}}
	handler := server.withAdminContext("campaign:read", func(w http.ResponseWriter, _ *http.Request, _ actorContext) {
		w.WriteHeader(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/survey-campaigns", nil))

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestAdminContextRejectsCrossTenantRequest(t *testing.T) {
	server := &Server{identity: stubIdentityProvider{identity: Identity{
		UserID:      "user-1",
		Role:        "owner",
		TenantIDs:   []string{demoTenantID},
		Permissions: []string{"campaign:read"},
	}}}
	handler := server.withAdminContext("campaign:read", func(w http.ResponseWriter, _ *http.Request, _ actorContext) {
		w.WriteHeader(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/survey-campaigns", nil)
	request.Header.Set("Authorization", "Bearer valid")
	request.Header.Set("X-Heard-Tenant-ID", "22222222-2222-2222-2222-222222222222")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestAdminContextRejectsMissingPermission(t *testing.T) {
	server := &Server{identity: stubIdentityProvider{identity: Identity{
		UserID:      "user-1",
		Role:        "viewer",
		TenantIDs:   []string{demoTenantID},
		Permissions: []string{"campaign:read"},
	}}}
	handler := server.withAdminContext("campaign:write", func(w http.ResponseWriter, _ *http.Request, _ actorContext) {
		w.WriteHeader(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodPost, "/api/v1/survey-campaigns", nil)
	request.Header.Set("Authorization", "Bearer valid")
	request.Header.Set("X-Heard-Tenant-ID", demoTenantID)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestLocalPassageCannotRunInProduction(t *testing.T) {
	_, err := NewIdentityProvider(Config{
		AppEnv:             "production",
		PassageMode:        "local",
		LocalPassageSecret: "test-secret",
	})
	if err == nil {
		t.Fatal("expected local Passage mode to be rejected in production")
	}
}

func TestCollectionPayloadEncodesEmptyItemsAsArray(t *testing.T) {
	payload, err := json.Marshal(collectionPayload([]RecoveryCase(nil)))
	if err != nil {
		t.Fatalf("marshal collection: %v", err)
	}
	if string(payload) != `{"items":[]}` {
		t.Fatalf("got %s, want empty items array", payload)
	}
}

type stubIdentityProvider struct {
	identity Identity
	err      error
}

func (s stubIdentityProvider) VerifyToken(context.Context, string) (Identity, error) {
	return s.identity, s.err
}
