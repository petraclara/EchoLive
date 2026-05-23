package config

import (
	"os"
	"strings"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	CORSOrigin  string
	Port        string
}

func Load() Config {
	cors := getEnv("CORS_ORIGIN", "")
	if cors == "" {
		cors = getEnv("WEB_APP_URL", "http://localhost:3000")
	}
	return Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://pulseroom:pulseroom@localhost:5432/pulseroom?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		CORSOrigin:  cors,
		Port:        getEnv("PORT", "8080"),
	}
}

// PublicAPIURL is the externally reachable API base URL (https on Render).
func PublicAPIURL() string {
	if u := os.Getenv("API_PUBLIC_URL"); u != "" {
		return strings.TrimRight(u, "/")
	}
	if u := os.Getenv("RENDER_EXTERNAL_URL"); u != "" {
		return strings.TrimRight(u, "/")
	}
	return "http://localhost:8080"
}

// PublicWebURL is the frontend URL used for join links and QR codes.
func PublicWebURL() string {
	if u := os.Getenv("WEB_APP_URL"); u != "" {
		return strings.TrimRight(u, "/")
	}
	return "http://localhost:3000"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
