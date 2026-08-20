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
				RestaurantName: "Concurrency Cafe",
				LocationName:   "First Location",
				Source:         "direct",
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
