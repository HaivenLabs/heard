package app

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

func TestOnboardingActivationIsIdempotentUnderConcurrency(t *testing.T) {
	databaseURL := os.Getenv("HEARD_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set HEARD_TEST_DATABASE_URL to run PostgreSQL onboarding integration tests")
	}
	ctx := context.Background()
	store, err := NewStore(ctx, Config{DatabaseURL: databaseURL, WebBaseURL: "http://heard.test"})
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	defer store.Close()
	if err := store.RunMigrations(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	provider := newLocalPassageProvider("integration-secret", time.Now)
	session, err := provider.IssueRegistration(fmt.Sprintf("onboarding-%d@heard.test", time.Now().UnixNano()))
	if err != nil {
		t.Fatalf("registration: %v", err)
	}
	defer store.pool.Exec(ctx, `delete from tenants where id = $1`, session.TenantID)

	const attempts = 12
	states := make(chan OnboardingState, attempts)
	errorsCh := make(chan error, attempts)
	var wait sync.WaitGroup
	for index := 0; index < attempts; index++ {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			state, err := store.ActivateRestaurantWorkspace(ctx, session.Identity, "owner", fmt.Sprintf("retry-key-%03d", index), createOnboardingActivationRequest{
				RestaurantName:   "Concurrency Cafe",
				RestaurantHandle: "concurrency-cafe",
				LocationName:     "First Location",
				Source:           "direct",
			})
			if err != nil {
				errorsCh <- err
				return
			}
			states <- state
		}(index)
	}
	wait.Wait()
	close(states)
	close(errorsCh)
	for err := range errorsCh {
		t.Fatalf("concurrent activation: %v", err)
	}

	activationID := ""
	for state := range states {
		if activationID == "" {
			activationID = state.ActivationID
		}
		if state.ActivationID != activationID || state.Tenant == nil || state.Tenant.ID != session.TenantID {
			t.Fatalf("activation was not stable across retries: %#v", state)
		}
	}
	var activations, locations, events int
	if err := store.pool.QueryRow(ctx, `select count(*) from onboarding_activations where actor_id = $1`, session.Identity.UserID).Scan(&activations); err != nil {
		t.Fatal(err)
	}
	if err := store.pool.QueryRow(ctx, `select count(*) from locations where tenant_id = $1`, session.TenantID).Scan(&locations); err != nil {
		t.Fatal(err)
	}
	if err := store.pool.QueryRow(ctx, `select count(*) from outbox_events where aggregate_id = $1`, activationID).Scan(&events); err != nil {
		t.Fatal(err)
	}
	if activations != 1 || locations != 1 || events != 1 {
		t.Fatalf("got activations=%d locations=%d events=%d; want one each", activations, locations, events)
	}
}

func TestOnboardingNeverLeaksWorkspaceAcrossEmails(t *testing.T) {
	databaseURL := os.Getenv("HEARD_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set HEARD_TEST_DATABASE_URL to run PostgreSQL onboarding integration tests")
	}
	ctx := context.Background()
	store, err := NewStore(ctx, Config{DatabaseURL: databaseURL, WebBaseURL: "http://heard.test"})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.RunMigrations(ctx); err != nil {
		t.Fatal(err)
	}

	provider := newLocalPassageProvider("integration-secret", time.Now)
	stamp := time.Now().UnixNano()
	first, err := provider.IssueRegistration(fmt.Sprintf("first-%d@heard.test", stamp))
	if err != nil {
		t.Fatal(err)
	}
	second, err := provider.IssueRegistration(fmt.Sprintf("second-%d@heard.test", stamp))
	if err != nil {
		t.Fatal(err)
	}
	defer store.pool.Exec(ctx, `delete from tenants where id in ($1,$2)`, first.TenantID, second.TenantID)

	created, err := store.ActivateRestaurantWorkspace(ctx, first.Identity, "owner", "first-workspace-key", createOnboardingActivationRequest{RestaurantName: "First Cafe", RestaurantHandle: "first-cafe", LocationName: "Downtown", Source: "direct"})
	if err != nil {
		t.Fatal(err)
	}
	if created.Tenant == nil || created.Tenant.Name != "First Cafe" {
		t.Fatalf("first workspace was not created: %#v", created)
	}
	secondState, err := store.GetOnboardingState(ctx, second.Identity)
	if err != nil {
		t.Fatal(err)
	}
	if secondState.Tenant != nil || secondState.Status == "complete" || secondState.NextStep != "workspace" {
		t.Fatalf("second email inherited another workspace: %#v", secondState)
	}
}

func TestSurveyCampaignUpdateIsTenantScopedAndPreservesIdentity(t *testing.T) {
	databaseURL := os.Getenv("HEARD_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set HEARD_TEST_DATABASE_URL to run PostgreSQL campaign integration tests")
	}
	ctx := context.Background()
	store, err := NewStore(ctx, Config{DatabaseURL: databaseURL, WebBaseURL: "http://heard.test"})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.RunMigrations(ctx); err != nil {
		t.Fatal(err)
	}

	provider := newLocalPassageProvider("integration-secret", time.Now)
	session, err := provider.IssueRegistration(fmt.Sprintf("campaign-update-%d@heard.test", time.Now().UnixNano()))
	if err != nil {
		t.Fatal(err)
	}
	defer store.pool.Exec(ctx, `delete from tenants where id = $1`, session.TenantID)
	state, err := store.ActivateRestaurantWorkspace(ctx, session.Identity, "owner", "campaign-update-key", createOnboardingActivationRequest{
		RestaurantName:   "Editable Cafe",
		RestaurantHandle: "editable-cafe",
		LocationName:     "Main Street",
		Source:           "direct",
	})
	if err != nil || state.Location == nil {
		t.Fatalf("activate workspace: state=%#v err=%v", state, err)
	}
	created, err := store.CreateSurveyCampaign(ctx, session.TenantID, session.Identity.UserID, "owner", createSurveyCampaignRequest{
		TenantID:       session.TenantID,
		LocationID:     state.Location.ID,
		Name:           "Lunch feedback",
		RestaurantName: "Editable Cafe",
		Headline:       "How was lunch?",
		Prompt:         "Tell us about it.",
	})
	if err != nil {
		t.Fatalf("create campaign: %v", err)
	}
	link, err := store.CreateFeedbackLink(ctx, session.TenantID, session.Identity.UserID, "owner", createFeedbackLinkRequest{
		TenantID: session.TenantID, LocationID: state.Location.ID, CampaignID: created.ID, Name: "Lunch flyer", Channel: "flyer", Slug: "lunch-feedback",
	})
	if err != nil {
		t.Fatalf("create feedback link: %v", err)
	}
	secondLocation, err := store.CreateLocation(ctx, session.TenantID, session.Identity.UserID, "owner", createLocationRequest{TenantID: session.TenantID, Name: "Harbor", Slug: "harbor"})
	if err != nil {
		t.Fatalf("create second location: %v", err)
	}

	updated, err := store.UpdateSurveyCampaign(ctx, session.TenantID, session.Identity.UserID, "owner", created.ID, updateSurveyCampaignRequest{
		LocationID:      secondLocation.ID,
		Name:            "Dinner feedback",
		RestaurantName:  "Editable Cafe",
		Headline:        "How was dinner?",
		Prompt:          "Tell us about your dinner.",
		IncentiveText:   "Share feedback for a chance to win.",
		SMSKeyword:      "DINNER",
		SMSPhone:        "(555) 010-0123",
		GoogleReviewURL: "https://example.com/google",
		YelpReviewURL:   "https://example.com/yelp",
	})
	if err != nil {
		t.Fatalf("update campaign: %v", err)
	}
	if updated.ID != created.ID || updated.Name != "Dinner feedback" || updated.Headline != "How was dinner?" || updated.SMSKeyword != "DINNER" {
		t.Fatalf("campaign update did not preserve identity and apply fields: %#v", updated)
	}
	stored, err := store.GetSurveyCampaign(ctx, session.TenantID, created.ID)
	if err != nil || stored.Prompt != "Tell us about your dinner." {
		t.Fatalf("updated campaign was not persisted: campaign=%#v err=%v", stored, err)
	}
	updatedLink, err := store.ResolveFeedbackLinkByID(ctx, session.TenantID, link.ID)
	if err != nil || updatedLink.LocationID != secondLocation.ID {
		t.Fatalf("campaign link did not move with the campaign: link=%#v err=%v", updatedLink, err)
	}
	if _, err := store.UpdateSurveyCampaign(ctx, "22222222-2222-2222-2222-222222222222", session.Identity.UserID, "owner", created.ID, updateSurveyCampaignRequest{
		LocationID: state.Location.ID, Name: "Blocked", RestaurantName: "Blocked", Headline: "Blocked", Prompt: "Blocked",
	}); err == nil {
		t.Fatal("expected cross-tenant campaign update to fail")
	}
	if _, available, err := store.TenantHandleAvailability(ctx, "editable-cafe-new"); err != nil || !available {
		t.Fatalf("expected candidate handle to be available: available=%t err=%v", available, err)
	}
	updatedTenant, err := store.UpdateTenantHandle(ctx, session.TenantID, session.Identity.UserID, "owner", "editable-cafe-new")
	if err != nil || updatedTenant.Slug != "editable-cafe-new" {
		t.Fatalf("update tenant handle: tenant=%#v err=%v", updatedTenant, err)
	}
	if _, available, err := store.TenantHandleAvailability(ctx, "editable-cafe-new"); err != nil || available {
		t.Fatalf("expected claimed handle to be unavailable: available=%t err=%v", available, err)
	}
}
