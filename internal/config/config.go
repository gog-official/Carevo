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
	RedisURL          string
	GeminiAPIKey      string
	GeminiModel       string
	AnthropicAPIKey   string
	AnthropicModel    string
	HuggingFaceAPIKey string
	HuggingFaceModel  string
	FrontendURL       string
	CloudinaryCloud   string
	CloudinaryKey     string
	CloudinarySecret  string
	AdminEmail        string
	SMTPFrom          string
	SMTPPassword      string
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
		RedisURL:          getEnv("REDIS_URL", "localhost:6379"),
		GeminiAPIKey:      getEnv("GEMINI_API_KEY", ""),
		GeminiModel:       getEnv("GEMINI_MODEL", "gemini-2.0-flash"),
		AnthropicAPIKey:   getEnv("ANTHROPIC_API_KEY", ""),
		AnthropicModel:    getEnv("ANTHROPIC_MODEL", "claude-sonnet-4-20250514"),
		HuggingFaceAPIKey: getEnv("HUGGINGFACE_API_KEY", ""),
		HuggingFaceModel:  getEnv("HUGGINGFACE_MODEL", "meta-llama/Llama-3.1-8B-Instruct"),
		FrontendURL:       getEnv("FRONTEND_URL", "http://localhost:3000"),
		CloudinaryCloud:   getEnv("CLOUDINARY_CLOUD", ""),
		CloudinaryKey:     getEnv("CLOUDINARY_KEY", ""),
		CloudinarySecret:  getEnv("CLOUDINARY_SECRET", ""),
		AdminEmail:        getEnv("ADMIN_EMAIL", "siddharthadhakal3722@gmail.com"),
		SMTPFrom:          getEnv("SMTP_FROM", "siddharthadhakal3722@gmail.com"),
		SMTPPassword:      getEnv("SMTP_PASSWORD", ""),
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
