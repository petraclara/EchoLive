package config

import (
	"os"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	CORSOrigin  string
	Port        string
}

func Load() Config {
	return Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://pulseroom:pulseroom@localhost:5432/pulseroom?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		CORSOrigin:  getEnv("CORS_ORIGIN", "http://localhost:3000"),
		Port:        getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
