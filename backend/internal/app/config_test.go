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

func validLocalAPIConfig() Config {
	return Config{
		AppEnv:                       "docker",
		LocalBindAddress:             "127.0.0.1",
		PassageMode:                  "local",
		LocalPassageSecret:           "local-test-secret",
		PublicAppURL:                 "http://localhost:3010",
		WebBaseURL:                   "http://localhost:3010",
		AllowedOrigin:                "http://localhost:3010",
		DemoSeedEnabled:              true,
		WorkerPollIntervalMS:         1500,
		PublicWriteRateLimit:         30,
		PublicWriteRateWindowSeconds: 60,
		WorkerBatchSize:              50,
	}
}
