package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConfigurePassageGoogleSendsCredentialsOnceFromBackend(t *testing.T) {
	var received map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/api/v1/internal/applications/heard/identity-providers/google" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer configuration-token-at-least-32-characters" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Error(err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	cfg := Config{PassageBaseURL: server.URL, PassageClientID: "heard", GoogleClientID: "google-client", GoogleClientSecret: "google-secret", GoogleRedirectURL: "http://localhost:3020/api/v1/auth/external/google/callback", PassageProviderConfigurationToken: "configuration-token-at-least-32-characters"}
	if err := ConfigurePassageGoogle(context.Background(), cfg, server.Client()); err != nil {
		t.Fatal(err)
	}
	if received["provider_client_id"] != "google-client" || received["provider_client_secret"] != "google-secret" || received["redirect_uri"] != cfg.GoogleRedirectURL {
		t.Fatalf("received=%v", received)
	}
}

func TestConfigurePassageGoogleSkipsWhenCredentialsAreAbsent(t *testing.T) {
	if err := ConfigurePassageGoogle(context.Background(), Config{}, nil); err != nil {
		t.Fatal(err)
	}
}
