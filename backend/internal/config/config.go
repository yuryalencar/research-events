package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all values parsed from environment variables at startup.
type Config struct {
	Port               string
	DatabaseURL        string
	JWTSecret          string
	Env                string
	CORSAllowedOrigins string
	AppVersion         string
}

// Load reads environment variables and returns a validated Config.
// It attempts to load a .env file first — if the file is missing that is silently
// ignored, which is the expected behaviour in production (no .env file present).
// Returns an error if any required variable is missing.
func Load() (Config, error) {
	// godotenv.Load populates os env from .env — safe to ignore "file not found"
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return Config{}, errors.New("DATABASE_URL is required but not set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return Config{}, errors.New("JWT_SECRET is required but not set")
	}

	return Config{
		Port:               getEnv("PORT", "8080"),
		DatabaseURL:        dbURL,
		JWTSecret:          jwtSecret,
		Env:                getEnv("ENV", "development"),
		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),
		AppVersion:         getEnv("APP_VERSION", "dev"),
	}, nil
}

// getEnv returns the env var value, or defaultValue if the var is empty or unset.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
