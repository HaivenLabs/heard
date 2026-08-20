package app

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestRequeueFailedOutboxEventIsOwnerAuditedAndTenantScoped(t *testing.T) {
	databaseURL := os.Getenv("HEARD_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set HEARD_TEST_DATABASE_URL to run PostgreSQL outbox integration tests")
	}
	ctx := context.Background()
	store, err := NewStore(ctx, Config{DatabaseURL: databaseURL})
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	defer store.Close()
	if err := store.RunMigrations(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	tenantID := uuid.NewString()
	otherTenantID := uuid.NewString()
	eventID := uuid.NewString()
	for _, tenant := range []string{tenantID, otherTenantID} {
		if _, err := store.pool.Exec(ctx, `insert into tenants (id, name, slug) values ($1, 'Outbox Test', $2)`, tenant, "outbox-"+tenant[:8]); err != nil {
			t.Fatalf("insert tenant: %v", err)
		}
		defer store.pool.Exec(ctx, `delete from tenants where id = $1`, tenant)
	}
	if _, err := store.pool.Exec(ctx, `
		insert into outbox_events (id, tenant_id, event_type, event_version, aggregate_type, aggregate_id, payload, status, attempts, last_error)
		values ($1, $2, 'feedback-submitted', 1, 'feedback_response', $3, '{}', 'failed', 8, 'test failure')
	`, eventID, tenantID, uuid.NewString()); err != nil {
		t.Fatalf("insert failed event: %v", err)
	}

	if _, err := store.RequeueFailedOutboxEvent(ctx, otherTenantID, eventID, "owner-2", "owner"); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("cross-tenant requeue error = %v, want not found", err)
	}
	result, err := store.RequeueFailedOutboxEvent(ctx, tenantID, eventID, "owner-1", "owner")
	if err != nil {
		t.Fatalf("requeue failed event: %v", err)
	}
	if result.Status != "pending" || result.Attempts != 0 {
		t.Fatalf("unexpected requeue result: %#v", result)
	}
	var auditCount int
	if err := store.pool.QueryRow(ctx, `select count(*) from audit_entries where tenant_id = $1 and resource_id = $2 and action = 'outbox_event.requeued'`, tenantID, eventID).Scan(&auditCount); err != nil {
		t.Fatal(err)
	}
	if auditCount != 1 {
		t.Fatalf("audit count = %d, want 1", auditCount)
	}
}
