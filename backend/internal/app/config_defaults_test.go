package app

import "testing"

func TestLoadConfigDefaultsHeardAPIToDedicatedHostPort(t *testing.T) {
	t.Setenv("APP_PORT", "")

	if got := LoadConfig().AppPort; got != "8082" {
		t.Fatalf("default APP_PORT = %q, want 8082", got)
	}
}
