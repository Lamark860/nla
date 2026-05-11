package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port        string
	Environment string

	// PostgreSQL
	PostgresDSN string

	// JWT
	JWTSecret     string
	JWTExpiration int // hours

	// OpenAI
	OpenAIKey     string
	OpenAIBaseURL string
	OpenAIModel   string
	OpenAIProxy   string
	OpenAITimeout int // seconds
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),

		PostgresDSN: getEnv("POSTGRES_DSN", "postgres://nla:nla_secret@postgres:5432/nla?sslmode=disable"),

		JWTSecret:     getEnv("JWT_SECRET", "change-me-in-production"),
		JWTExpiration: getEnvInt("JWT_EXPIRATION_HOURS", 72),

		OpenAIKey:     getEnv("OPENAI_API_KEY", ""),
		OpenAIBaseURL: getEnv("OPENAI_BASE_URL", "https://api.openai.com/v1/"),
		OpenAIModel:   getEnv("OPENAI_MODEL", "gpt-5.1-2025-11-13"),
		OpenAIProxy:   getEnv("OPENAI_PROXY", ""),
		OpenAITimeout: getEnvInt("OPENAI_TIMEOUT", 200),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
