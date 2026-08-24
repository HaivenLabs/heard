package app

import (
	"errors"
	"fmt"
	"net/netip"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	AppEnv                            string
	AppPort                           string
	LocalBindAddress                  string
	DatabaseURL                       string
	PublicAppURL                      string
	WebBaseURL                        string
	AllowedOrigin                     string
	DemoSeedEnabled                   bool
	QURLBaseURL                       string
	PassageMode                       string
	LocalPassageSecret                string
	PassageBaseURL                    string
	PassagePublicURL                  string
	PassageIssuer                     string
	PassageAudience                   string
	PassageJWKSCacheSeconds           int
	PassageCallbackURL                string
	PassageClientID                   string
	GoogleClientID                    string
	GoogleClientSecret                string
	GoogleRedirectURL                 string
	PassageProviderConfigurationToken string
	WorkerPollIntervalMS              int
	WorkerBatchSize                   int
	PublicWriteRateLimit              int
	PublicWriteRateWindowSeconds      int
	TrustedProxyCIDRs                 []string
}

func LoadConfig() Config {
	passageBaseURL := getEnv("PASSAGE_BASE_URL", "http://localhost:8081")
	return Config{
		AppEnv:                            getEnv("APP_ENV", ""),
		AppPort:                           getEnv("APP_PORT", "8082"),
		LocalBindAddress:                  getEnv("HEARD_BIND_ADDRESS", ""),
		DatabaseURL:                       getEnv("DATABASE_URL", "postgres://heard:heard@localhost:5432/heard?sslmode=disable"),
		PublicAppURL:                      getEnv("PUBLIC_APP_URL", "http://localhost:3010"),
		WebBaseURL:                        getEnv("WEB_BASE_URL", "http://localhost:3010"),
		AllowedOrigin:                     getEnv("ALLOWED_ORIGIN", "http://localhost:3010"),
		DemoSeedEnabled:                   getEnv("HEARD_SEED_DEMO", "true") == "true",
		QURLBaseURL:                       getEnv("QURL_BASE_URL", ""),
		PassageMode:                       getEnv("PASSAGE_MODE", ""),
		LocalPassageSecret:                getEnv("LOCAL_PASSAGE_SECRET", ""),
		PassageBaseURL:                    passageBaseURL,
		PassagePublicURL:                  getEnv("PASSAGE_PUBLIC_URL", passageBaseURL),
		PassageIssuer:                     getEnv("PASSAGE_ISSUER", "http://localhost:8081"),
		PassageAudience:                   getEnv("PASSAGE_AUDIENCE", "heard"),
		PassageJWKSCacheSeconds:           getEnvInt("PASSAGE_JWKS_CACHE_SECONDS", 300),
		PassageCallbackURL:                getEnv("PASSAGE_CALLBACK_URL", "http://localhost:3010/api/v1/auth/callback"),
		PassageClientID:                   getEnv("PASSAGE_CLIENT_ID", "heard"),
		GoogleClientID:                    getEnv("HEARD_GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:                getEnv("HEARD_GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:                 getEnv("HEARD_GOOGLE_REDIRECT_URL", "http://localhost:3020/api/v1/auth/external/google/callback"),
		PassageProviderConfigurationToken: getEnv("PASSAGE_PROVIDER_CONFIGURATION_TOKEN", ""),
		WorkerPollIntervalMS:              getEnvInt("WORKER_POLL_INTERVAL_MS", 1500),
		WorkerBatchSize:                   getEnvInt("WORKER_BATCH_SIZE", 50),
		PublicWriteRateLimit:              getEnvInt("PUBLIC_WRITE_RATE_LIMIT", 30),
		PublicWriteRateWindowSeconds:      getEnvInt("PUBLIC_WRITE_RATE_WINDOW_SECONDS", 60),
		TrustedProxyCIDRs:                 splitEnvList(getEnv("TRUSTED_PROXY_CIDRS", "")),
	}
}

func (cfg Config) ValidateRuntime() error {
	switch cfg.AppEnv {
	case "local", "docker", "test", "staging", "production":
	default:
		return fmt.Errorf("APP_ENV must be one of local, docker, test, staging, or production; got %q", cfg.AppEnv)
	}
	if cfg.DemoSeedEnabled && cfg.IsProductionLike() {
		return errors.New("HEARD_SEED_DEMO cannot be enabled in staging or production")
	}
	if cfg.DemoSeedEnabled && cfg.LocalBindAddress != "" && !isLoopbackAddress(cfg.LocalBindAddress) {
		return errors.New("HEARD_SEED_DEMO cannot be enabled when the local Compose bind address is public")
	}
	if cfg.WorkerPollIntervalMS < 50 || cfg.WorkerPollIntervalMS > 300000 {
		return errors.New("WORKER_POLL_INTERVAL_MS must be between 50 and 300000")
	}
	if cfg.WorkerBatchSize < 1 || cfg.WorkerBatchSize > 1000 {
		return errors.New("WORKER_BATCH_SIZE must be between 1 and 1000")
	}
	return nil
}

func (cfg Config) ValidateAPI() error {
	if err := cfg.ValidateRuntime(); err != nil {
		return err
	}
	if cfg.PublicWriteRateLimit < 1 || cfg.PublicWriteRateLimit > 10000 {
		return errors.New("PUBLIC_WRITE_RATE_LIMIT must be between 1 and 10000")
	}
	if cfg.PublicWriteRateWindowSeconds < 1 || cfg.PublicWriteRateWindowSeconds > 86400 {
		return errors.New("PUBLIC_WRITE_RATE_WINDOW_SECONDS must be between 1 and 86400")
	}
	for _, raw := range cfg.TrustedProxyCIDRs {
		if _, err := netip.ParsePrefix(raw); err != nil {
			return fmt.Errorf("TRUSTED_PROXY_CIDRS contains invalid network %q", raw)
		}
	}
	if strings.TrimSpace(cfg.QURLBaseURL) != "" {
		qurlURL, err := url.Parse(cfg.QURLBaseURL)
		if err != nil || qurlURL.Host == "" || (qurlURL.Scheme != "https" && qurlURL.Scheme != "http") {
			return errors.New("QURL_BASE_URL must be an absolute HTTPS URL")
		}
		if qurlURL.Scheme == "http" && (!cfg.IsLocalRuntime() || !isExplicitLocalProviderURL(cfg.QURLBaseURL)) {
			return errors.New("QURL_BASE_URL must use HTTPS unless it is an explicit local provider")
		}
	}
	if cfg.PassageMode == "local" {
		if !cfg.IsLocalRuntime() {
			return errors.New("local Passage adapter is allowed only in local, docker, or test runtimes")
		}
		if len(cfg.LocalPassageSecret) < 12 {
			return errors.New("LOCAL_PASSAGE_SECRET must be at least 12 characters")
		}
		if cfg.LocalBindAddress != "" && !isLoopbackAddress(cfg.LocalBindAddress) {
			return errors.New("PASSAGE_MODE=local requires a loopback HEARD_BIND_ADDRESS")
		}
		for name, raw := range map[string]string{"PUBLIC_APP_URL": cfg.PublicAppURL, "WEB_BASE_URL": cfg.WebBaseURL, "ALLOWED_ORIGIN": cfg.AllowedOrigin} {
			if !isLoopbackURL(raw) {
				return fmt.Errorf("%s must use a loopback host while PASSAGE_MODE=local", name)
			}
		}
	} else if cfg.PassageMode == "jwks" {
		if strings.TrimSpace(cfg.PassageClientID) == "" {
			return errors.New("PASSAGE_CLIENT_ID is required")
		}
		for name, raw := range map[string]string{
			"PASSAGE_BASE_URL": cfg.PassageBaseURL, "PASSAGE_PUBLIC_URL": cfg.PassagePublicURL,
			"PASSAGE_ISSUER": cfg.PassageIssuer, "PASSAGE_CALLBACK_URL": cfg.PassageCallbackURL,
		} {
			parsed, err := url.Parse(strings.TrimSpace(raw))
			if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") {
				return fmt.Errorf("%s must be an absolute HTTP(S) URL", name)
			}
			if cfg.IsProductionLike() && parsed.Scheme != "https" {
				return fmt.Errorf("%s must use HTTPS in staging and production", name)
			}
		}
	} else {
		return fmt.Errorf("PASSAGE_MODE must be local or jwks; got %q", cfg.PassageMode)
	}
	googleConfigured := strings.TrimSpace(cfg.GoogleClientID) != "" || strings.TrimSpace(cfg.GoogleClientSecret) != ""
	if googleConfigured {
		if strings.TrimSpace(cfg.GoogleClientID) == "" || strings.TrimSpace(cfg.GoogleClientSecret) == "" || strings.TrimSpace(cfg.GoogleRedirectURL) == "" {
			return errors.New("HEARD_GOOGLE_CLIENT_ID, HEARD_GOOGLE_CLIENT_SECRET, and HEARD_GOOGLE_REDIRECT_URL must be configured together")
		}
		if len(strings.TrimSpace(cfg.PassageProviderConfigurationToken)) < 32 {
			return errors.New("PASSAGE_PROVIDER_CONFIGURATION_TOKEN must be at least 32 characters when Google sign-in is configured")
		}
		redirect, err := url.Parse(strings.TrimSpace(cfg.GoogleRedirectURL))
		if err != nil || redirect.Host == "" || (redirect.Scheme != "https" && !(cfg.IsLocalRuntime() && redirect.Scheme == "http" && isLoopbackURL(cfg.GoogleRedirectURL))) {
			return errors.New("HEARD_GOOGLE_REDIRECT_URL must use HTTPS except on local loopback")
		}
	}
	return nil
}

func (cfg Config) IsLocalRuntime() bool {
	return cfg.AppEnv == "local" || cfg.AppEnv == "docker" || cfg.AppEnv == "test"
}

func (cfg Config) IsProductionLike() bool {
	return cfg.AppEnv == "staging" || cfg.AppEnv == "production"
}

func isLoopbackURL(raw string) bool {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "localhost" {
		return true
	}
	address, err := netip.ParseAddr(host)
	return err == nil && address.IsLoopback()
}

func isExplicitLocalProviderURL(raw string) bool {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return false
	}
	if strings.EqualFold(parsed.Hostname(), "host.docker.internal") {
		return true
	}
	return isLoopbackURL(raw)
}

func isLoopbackAddress(raw string) bool {
	address, err := netip.ParseAddr(strings.TrimSpace(raw))
	return err == nil && address.IsLoopback()
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	value := getEnv(key, "")
	if value == "" {
		return fallback
	}
	n := 0
	for _, r := range value {
		if r < '0' || r > '9' {
			return fallback
		}
		n = (n * 10) + int(r-'0')
	}
	return n
}

func splitEnvList(value string) []string {
	var values []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			values = append(values, item)
		}
	}
	return values
}
