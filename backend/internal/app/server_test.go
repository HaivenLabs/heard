package app

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestWriteStoreErrorMapsValidationErrorsToBadRequest(t *testing.T) {
	recorder := httptest.NewRecorder()

	writeStoreError(recorder, fmt.Errorf("%w: enter a valid email address", errValidation))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestDecodeJSONRequestEnforcesContentType(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/feedback-sessions", strings.NewReader(`{"token":"opaque"}`))
	recorder := httptest.NewRecorder()
	var payload createFeedbackSessionRequest

	err := decodeJSONRequest(recorder, request, &payload, publicRequestBodyLimit)
	if err == nil || requestErrorStatus(err) != http.StatusUnsupportedMediaType {
		t.Fatalf("error = %v, want unsupported media type", err)
	}
}

func TestDecodeJSONRequestRejectsOversizedBody(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/feedback-sessions", bytes.NewReader([]byte(`{"token":"too-large"}`)))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	var payload createFeedbackSessionRequest

	err := decodeJSONRequest(recorder, request, &payload, 8)
	if err == nil || requestErrorStatus(err) != http.StatusRequestEntityTooLarge {
		t.Fatalf("error = %v, want payload too large", err)
	}
}

func TestPublicWriteRateLimitReturnsRetryAfter(t *testing.T) {
	now := time.Date(2026, time.August, 16, 12, 0, 0, 0, time.UTC)
	server := &Server{
		limiter:       newPublicRateLimiter(1, time.Minute, func() time.Time { return now }),
		clientAddress: newClientAddressResolver(nil),
	}
	handler := server.withPublicWriteControls("feedback-session", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	first := httptest.NewRecorder()
	handler.ServeHTTP(first, httptest.NewRequest(http.MethodPost, "/api/v1/feedback-sessions", nil))
	second := httptest.NewRecorder()
	handler.ServeHTTP(second, httptest.NewRequest(http.MethodPost, "/api/v1/feedback-sessions", nil))

	if first.Code != http.StatusCreated || second.Code != http.StatusTooManyRequests {
		t.Fatalf("statuses = %d, %d", first.Code, second.Code)
	}
	if second.Header().Get("Retry-After") == "" {
		t.Fatal("rate-limited response must include Retry-After")
	}
}

func TestProductionPublicWriteFailsClosedWithoutSharedLimiterStore(t *testing.T) {
	server := NewServer(Config{
		AppEnv:                       "production",
		PublicWriteRateLimit:         30,
		PublicWriteRateWindowSeconds: 60,
	}, nil, nil)
	handler := server.withPublicWriteControls("marketing-lead", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/marketing-leads", nil))
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want fail-closed 503", recorder.Code)
	}
}

func TestClientAddressIgnoresForwardedHeaderFromUntrustedPeer(t *testing.T) {
	resolver := newClientAddressResolver([]string{"10.0.0.0/8"})
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	request.RemoteAddr = "203.0.113.40:51000"
	request.Header.Set("X-Forwarded-For", "198.51.100.17")
	if got := resolver.Resolve(request); got != "203.0.113.40" {
		t.Fatalf("resolved address = %q", got)
	}
}

func TestClientAddressUsesFirstUntrustedHopBehindTrustedProxy(t *testing.T) {
	resolver := newClientAddressResolver([]string{"10.0.0.0/8"})
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	request.RemoteAddr = "10.0.0.8:51000"
	request.Header.Set("X-Forwarded-For", "198.51.100.17, 10.1.1.4")
	if got := resolver.Resolve(request); got != "198.51.100.17" {
		t.Fatalf("resolved address = %q", got)
	}
}
