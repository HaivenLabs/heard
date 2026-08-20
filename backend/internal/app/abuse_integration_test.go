package app

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPostgresPublicRateLimiterIsAtomicAcrossConcurrentCallers(t *testing.T) {
	databaseURL := os.Getenv("HEARD_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set HEARD_TEST_DATABASE_URL to run PostgreSQL rate limiter integration tests")
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

	const limit = 5
	key := fmt.Sprintf("feedback-response:test-%d", time.Now().UnixNano())
	defer store.pool.Exec(ctx, `delete from public_rate_limit_buckets where bucket_key = $1`, rateLimitBucketKey(key))
	limiter := newPostgresPublicRateLimiter(store.pool, limit, 60)
	var allowed atomic.Int32
	var wait sync.WaitGroup
	errorsCh := make(chan error, 20)
	for index := 0; index < 20; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			ok, _, err := limiter.Allow(ctx, key)
			if err != nil {
				errorsCh <- err
				return
			}
			if ok {
				allowed.Add(1)
			}
		}()
	}
	wait.Wait()
	close(errorsCh)
	for err := range errorsCh {
		t.Fatalf("rate limiter request: %v", err)
	}
	if got := allowed.Load(); got != limit {
		t.Fatalf("allowed requests = %d, want %d", got, limit)
	}
}
