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

func TestLocalPassageRegistrationCreatesDedicatedAccountContext(t *testing.T) {
	provider := newLocalPassageProvider("test-secret", func() time.Time {
		return time.Date(2026, time.August, 16, 12, 0, 0, 0, time.UTC)
	})

	session, err := provider.IssueRegistration("new-owner@cedar.test")
	if err != nil {
		t.Fatalf("issue registration: %v", err)
	}
	if session.TenantID == "" || session.TenantID == demoTenantID {
		t.Fatalf("expected a dedicated local Passage account context, got %q", session.TenantID)
	}
	if !session.Identity.HasTenant(session.TenantID) {
		t.Fatal("expected registration identity to carry its local Passage tenant membership")
	}

	retry, err := provider.IssueRegistration("NEW-OWNER@CEDAR.TEST")
	if err != nil {
		t.Fatalf("retry registration: %v", err)
	}
	if retry.TenantID != session.TenantID {
		t.Fatalf("registration context changed across retry: %q != %q", retry.TenantID, session.TenantID)
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

func TestOwnerContextRejectsNonOwnerWithPermission(t *testing.T) {
	server := &Server{identity: stubIdentityProvider{identity: Identity{
		UserID:      "user-1",
		Role:        "admin",
		TenantIDs:   []string{demoTenantID},
		Permissions: []string{"outbox:replay"},
	}}}
	handler := server.withOwnerContext("outbox:replay", func(w http.ResponseWriter, _ *http.Request, _ actorContext) {
		w.WriteHeader(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodPost, "/api/v1/outbox-events/00000000-0000-0000-0000-000000000000/requeue", nil)
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

func TestLocalMintRoutesAreAbsentWithoutLocalAdapter(t *testing.T) {
	server := &Server{cfg: Config{AppEnv: "production"}, identity: stubIdentityProvider{}}
	for _, path := range []string{"/api/v1/auth/local/session", "/api/v1/auth/local/registration"} {
		request := httptest.NewRequest(http.MethodPost, path, nil)
		recorder := httptest.NewRecorder()
		server.Router().ServeHTTP(recorder, request)
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("%s status = %d, want route absent", path, recorder.Code)
		}
	}
}

func TestOnboardingNextStep(t *testing.T) {
	tests := []struct {
		state OnboardingState
		want  string
	}{
		{state: OnboardingState{}, want: "workspace"},
		{state: OnboardingState{ActivationID: "activation", Tenant: &Tenant{}, Location: &Location{}}, want: "campaign"},
		{state: OnboardingState{ActivationID: "activation", Tenant: &Tenant{}, Location: &Location{}, Campaign: &SurveyCampaign{}}, want: "feedback_link"},
		{state: OnboardingState{ActivationID: "activation", Tenant: &Tenant{}, Location: &Location{}, Campaign: &SurveyCampaign{}, FeedbackLink: &FeedbackLink{}}, want: "complete"},
	}
	for _, item := range tests {
		item.state.resolveProgress()
		if item.state.NextStep != item.want {
			t.Fatalf("got next step %q, want %q", item.state.NextStep, item.want)
		}
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
