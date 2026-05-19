package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL       string
	JWTSecret         string
	JWTAccessTTL      time.Duration
	JWTRefreshTTL     time.Duration
	Port              string
}

func Load() (*Config, error) {
	godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, errMissing("DATABASE_URL")
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, errMissing("JWT_SECRET")
	}

	accessTTL, err := time.ParseDuration(getEnv("JWT_ACCESS_TTL", "15m"))
	if err != nil {
		return nil, err
	}

	refreshTTL, err := time.ParseDuration(getEnv("JWT_REFRESH_TTL", "720h"))
	if err != nil {
		return nil, err
	}

	return &Config{
		DatabaseURL:       dbURL,
		JWTSecret:         secret,
		JWTAccessTTL:      accessTTL,
		JWTRefreshTTL:     refreshTTL,
		Port:              getEnv("PORT", "8080"),
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func errMissing(key string) error {
	return missingEnvError{key: key}
}

type missingEnvError struct{ key string }

func (e missingEnvError) Error() string {
	return "missing required environment variable: " + e.key
}
