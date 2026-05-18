package app

import "os"

type Config struct {
	AppEnv               string
	AppPort              string
	DatabaseURL          string
	PublicAppURL         string
	WebBaseURL           string
	AllowedOrigin        string
	DemoSeedEnabled      bool
	QURLBaseURL          string
	WorkerPollIntervalMS int
}

func LoadConfig() Config {
	return Config{
		AppEnv:               getEnv("APP_ENV", "local"),
		AppPort:              getEnv("APP_PORT", "8080"),
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://heard:heard@localhost:5432/heard?sslmode=disable"),
		PublicAppURL:         getEnv("PUBLIC_APP_URL", "http://localhost:3010"),
		WebBaseURL:           getEnv("WEB_BASE_URL", "http://localhost:3010"),
		AllowedOrigin:        getEnv("ALLOWED_ORIGIN", "http://localhost:3010"),
		DemoSeedEnabled:      getEnv("HEARD_SEED_DEMO", "true") == "true",
		QURLBaseURL:          getEnv("QURL_BASE_URL", ""),
		WorkerPollIntervalMS: getEnvInt("WORKER_POLL_INTERVAL_MS", 1500),
	}
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
