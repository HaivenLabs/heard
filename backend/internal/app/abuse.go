package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const maxTrackedRateLimitKeys = 10000

type publicRateLimitEntry struct {
	count int
	reset time.Time
}

type publicWriteLimiter interface {
	Allow(context.Context, string) (bool, time.Duration, error)
}

type unavailablePublicWriteLimiter struct{}

var errPublicWriteLimiterUnavailable = errors.New("shared public write limiter is unavailable")

func (unavailablePublicWriteLimiter) Allow(context.Context, string) (bool, time.Duration, error) {
	return false, 0, errPublicWriteLimiterUnavailable
}

type memoryPublicRateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	now     func() time.Time
	entries map[string]publicRateLimitEntry
}

func newPublicRateLimiter(limit int, window time.Duration, now func() time.Time) *memoryPublicRateLimiter {
	if now == nil {
		now = time.Now
	}
	return &memoryPublicRateLimiter{limit: limit, window: window, now: now, entries: map[string]publicRateLimitEntry{}}
}

func (l *memoryPublicRateLimiter) Allow(_ context.Context, key string) (bool, time.Duration, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	entry, exists := l.entries[key]
	if !exists || !now.Before(entry.reset) {
		if len(l.entries) >= maxTrackedRateLimitKeys {
			l.evict(now)
		}
		entry = publicRateLimitEntry{reset: now.Add(l.window)}
	}
	if entry.count >= l.limit {
		return false, maxDuration(time.Second, entry.reset.Sub(now)), nil
	}
	entry.count++
	l.entries[key] = entry
	return true, entry.reset.Sub(now), nil
}

func (l *memoryPublicRateLimiter) evict(now time.Time) {
	var oldestKey string
	var oldestReset time.Time
	for key, entry := range l.entries {
		if !now.Before(entry.reset) {
			delete(l.entries, key)
			continue
		}
		if oldestKey == "" || entry.reset.Before(oldestReset) {
			oldestKey, oldestReset = key, entry.reset
		}
	}
	if len(l.entries) >= maxTrackedRateLimitKeys && oldestKey != "" {
		delete(l.entries, oldestKey)
	}
}

type postgresPublicRateLimiter struct {
	pool           *pgxpool.Pool
	limit          int
	windowSeconds  int
	cleanupCounter atomic.Uint64
}

func newPostgresPublicRateLimiter(pool *pgxpool.Pool, limit, windowSeconds int) *postgresPublicRateLimiter {
	return &postgresPublicRateLimiter{pool: pool, limit: limit, windowSeconds: windowSeconds}
}

func (l *postgresPublicRateLimiter) Allow(ctx context.Context, key string) (bool, time.Duration, error) {
	if l.cleanupCounter.Add(1)%128 == 0 {
		if _, err := l.pool.Exec(ctx, `
			delete from public_rate_limit_buckets
			where ctid in (
				select ctid from public_rate_limit_buckets
				where expires_at <= now()
				limit 500
			)
		`); err != nil {
			return false, 0, err
		}
	}

	var count int
	var retrySeconds int
	err := l.pool.QueryRow(ctx, `
		insert into public_rate_limit_buckets (bucket_key, request_count, expires_at)
		values ($1, 1, now() + make_interval(secs => $2))
		on conflict (bucket_key) do update set
			request_count = case
				when public_rate_limit_buckets.expires_at <= now() then 1
				when public_rate_limit_buckets.request_count >= 2147483647 then 2147483647
				else public_rate_limit_buckets.request_count + 1
			end,
			expires_at = case
				when public_rate_limit_buckets.expires_at <= now() then now() + make_interval(secs => $2)
				else public_rate_limit_buckets.expires_at
			end
		returning request_count, greatest(1, ceil(extract(epoch from (expires_at - now()))))::integer
	`, rateLimitBucketKey(key), l.windowSeconds).Scan(&count, &retrySeconds)
	if err != nil {
		return false, 0, err
	}
	return count <= l.limit, time.Duration(retrySeconds) * time.Second, nil
}

func rateLimitBucketKey(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}

type clientAddressResolver struct {
	trusted []netip.Prefix
}

func newClientAddressResolver(rawPrefixes []string) clientAddressResolver {
	resolver := clientAddressResolver{}
	for _, raw := range rawPrefixes {
		if prefix, err := netip.ParsePrefix(strings.TrimSpace(raw)); err == nil {
			resolver.trusted = append(resolver.trusted, prefix)
		}
	}
	return resolver
}

func (r clientAddressResolver) Resolve(request *http.Request) string {
	peer, ok := parseRequestAddress(request.RemoteAddr)
	if !ok {
		return "unknown"
	}
	if !r.isTrusted(peer) {
		return peer.String()
	}
	forwarded := strings.Split(request.Header.Get("X-Forwarded-For"), ",")
	if len(forwarded) == 1 && strings.TrimSpace(forwarded[0]) == "" {
		return peer.String()
	}
	if len(forwarded) > 20 {
		return peer.String()
	}
	candidates := make([]netip.Addr, 0, len(forwarded))
	for _, raw := range forwarded {
		candidate, err := netip.ParseAddr(strings.TrimSpace(raw))
		if err != nil {
			return peer.String()
		}
		candidates = append(candidates, candidate.Unmap())
	}
	current := peer
	for index := len(candidates) - 1; index >= 0 && r.isTrusted(current); index-- {
		current = candidates[index]
	}
	return current.String()
}

func (r clientAddressResolver) isTrusted(address netip.Addr) bool {
	for _, prefix := range r.trusted {
		if prefix.Contains(address) {
			return true
		}
	}
	return false
}

func parseRequestAddress(remoteAddress string) (netip.Addr, bool) {
	host, _, err := net.SplitHostPort(strings.TrimSpace(remoteAddress))
	if err != nil {
		host = strings.TrimSpace(remoteAddress)
	}
	address, err := netip.ParseAddr(host)
	return address.Unmap(), err == nil
}

func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}
