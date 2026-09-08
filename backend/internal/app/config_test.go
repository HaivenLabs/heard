package app

import "testing"

func TestValidateAPIConfigFailsClosedForUnknownEnvironment(t *testing.T) {
	for _, environment := range []string{"", "prod", "preview", "Production"} {
		cfg := validLocalAPIConfig()
		cfg.AppEnv = environment
		if err := cfg.ValidateAPI(); err == nil {
			t.Fatalf("APP_ENV %q should be rejected", environment)
		}
	}
}

func TestValidateAPIConfigAllowsDocumentedLocalDockerRuntime(t *testing.T) {
	cfg := validLocalAPIConfig()
	if err := cfg.ValidateAPI(); err != nil {
		t.Fatalf("documented local Docker config was rejected: %v", err)
	}
}

func TestValidateAPIConfigRejectsDevelopmentFeaturesOutsideLocalRuntime(t *testing.T) {
	for _, environment := range []string{"staging", "production"} {
		t.Run(environment, func(t *testing.T) {
			cfg := validLocalAPIConfig()
			cfg.AppEnv = environment
			if err := cfg.ValidateAPI(); err == nil {
				t.Fatal("expected local Passage and demo seed to be rejected")
			}
		})
	}
}

func TestValidateAPIConfigRejectsPublicOriginWithLocalAuth(t *testing.T) {
	cfg := validLocalAPIConfig()
	cfg.PublicAppURL = "https://heard.example"
	if err := cfg.ValidateAPI(); err == nil {
		t.Fatal("expected public origin with local authentication to be rejected")
	}
}

func TestValidateAPIConfigRequiresOneCanonicalBrowserOrigin(t *testing.T) {
	cfg := validProductionAPIConfig()
	for field, value := range map[string]string{
		"web":      "https://other.heard.example",
		"cors":     "https://other.heard.example",
		"callback": "https://other.heard.example/api/v1/auth/callback",
	} {
		t.Run(field, func(t *testing.T) {
			candidate := cfg
			switch field {
			case "web":
				candidate.WebBaseURL = value
			case "cors":
				candidate.AllowedOrigin = value
			case "callback":
				candidate.PassageCallbackURL = value
			}
			if err := candidate.ValidateAPI(); err == nil {
				t.Fatalf("expected mismatched %s origin to be rejected", field)
			}
		})
	}
}

func TestValidateAPIConfigRequiresHTTPSCanonicalOriginOutsideLocalRuntime(t *testing.T) {
	cfg := validProductionAPIConfig()
	cfg.PublicAppURL = "http://heard.example"
	cfg.WebBaseURL = "http://heard.example"
	cfg.AllowedOrigin = "http://heard.example"
	cfg.PassageCallbackURL = "http://heard.example/api/v1/auth/callback"
	if err := cfg.ValidateAPI(); err == nil {
		t.Fatal("expected cleartext canonical origin to be rejected")
	}
}

func TestValidateAPIConfigRejectsPublicComposeBindWithDevelopmentFeatures(t *testing.T) {
	cfg := validLocalAPIConfig()
	cfg.LocalBindAddress = "0.0.0.0"
	if err := cfg.ValidateAPI(); err == nil {
		t.Fatal("expected public Compose bind with local auth and demo seed to be rejected")
	}
}

func TestValidateAPIConfigRejectsCleartextProductionIdentityURLs(t *testing.T) {
	cfg := validLocalAPIConfig()
	cfg.AppEnv = "production"
	cfg.PassageMode = "jwks"
	cfg.DemoSeedEnabled = false
	cfg.PassageBaseURL = "https://passage.example"
	cfg.PassagePublicURL = "http://passage.example"
	cfg.PassageIssuer = "https://passage.example"
	cfg.PassageCallbackURL = "https://heard.example/api/v1/auth/callback"
	cfg.PassageClientID = "heard"
	if err := cfg.ValidateAPI(); err == nil {
		t.Fatal("expected public cleartext Passage URL to be rejected")
	}
}

func TestValidateAPIConfigRequiresBoundedJWKSOutageWindow(t *testing.T) {
	cfg := validLocalAPIConfig()
	cfg.PassageMode = "jwks"
	cfg.DemoSeedEnabled = false
	cfg.PassageBaseURL = "http://localhost:8081"
	cfg.PassagePublicURL = "http://localhost:8081"
	cfg.PassageIssuer = "http://localhost:8081"
	cfg.PassageCallbackURL = "http://localhost:3010/api/v1/auth/callback"
	cfg.PassageClientID = "heard"
	cfg.PassageJWKSCacheSeconds = 30
	cfg.PassageJWKSStaleSeconds = 0
	if err := cfg.ValidateAPI(); err == nil {
		t.Fatal("expected unbounded JWKS outage window to be rejected")
	}
}

func validLocalAPIConfig() Config {
	return Config{
		AppEnv:                       "docker",
		LocalBindAddress:             "127.0.0.1",
		PassageMode:                  "local",
		LocalPassageSecret:           "local-test-secret",
		PublicAppURL:                 "http://localhost:3010",
		WebBaseURL:                   "http://localhost:3010",
		AllowedOrigin:                "http://localhost:3010",
		PassageCallbackURL:           "http://localhost:3010/api/v1/auth/callback",
		DemoSeedEnabled:              true,
		WorkerPollIntervalMS:         1500,
		PublicWriteRateLimit:         30,
		PublicWriteRateWindowSeconds: 60,
		WorkerBatchSize:              50,
	}
}

func validProductionAPIConfig() Config {
	return Config{
		AppEnv:                       "production",
		PassageMode:                  "jwks",
		PublicAppURL:                 "https://heard.example",
		WebBaseURL:                   "https://heard.example",
		AllowedOrigin:                "https://heard.example",
		PassageCallbackURL:           "https://heard.example/api/v1/auth/callback",
		PassageBaseURL:               "https://auth.heard.example",
		PassagePublicURL:             "https://auth.heard.example",
		PassageIssuer:                "https://auth.heard.example",
		PassageClientID:              "heard",
		DemoSeedEnabled:              false,
		WorkerPollIntervalMS:         1500,
		WorkerBatchSize:              50,
		PublicWriteRateLimit:         30,
		PublicWriteRateWindowSeconds: 60,
		PassageJWKSCacheSeconds:      300,
		PassageJWKSStaleSeconds:      900,
	}
}
