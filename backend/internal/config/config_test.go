package config_test

// Spec: specs/backend/server-bootstrap.yaml
// Rule: "App must refuse to start if DATABASE_URL is missing"
// Rule: "App must refuse to start if JWT_SECRET is missing"
// Rule: "CORS_ALLOWED_ORIGINS defaults to http://localhost:3000; APP_VERSION defaults to dev; PORT defaults to 8080"

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuryalencar/research-events/internal/config"
)

func TestConfig_Load_ReturnsErrorWhenDatabaseURLMissing(t *testing.T) {
	// t.Setenv sets the var for this test only and restores the original on cleanup.
	// Setting to "" is treated as "not provided" — Load must return an error.
	t.Setenv("DATABASE_URL", "")

	_, err := config.Load()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "DATABASE_URL")
}

func TestConfig_Load_ReturnsDatabaseURLFromEnv(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/research_events")
	t.Setenv("JWT_SECRET", "test-secret")

	cfg, err := config.Load()

	require.NoError(t, err)
	assert.Equal(t, "postgres://user:pass@localhost:5432/research_events", cfg.DatabaseURL)
}

func TestConfig_Load_UsesDefaultsWhenOptionalVarsAreEmpty(t *testing.T) {
	// All optional vars set to "" — Load must fall back to their documented defaults.
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("PORT", "")
	t.Setenv("ENV", "")
	t.Setenv("CORS_ALLOWED_ORIGINS", "")
	t.Setenv("APP_VERSION", "")

	cfg, err := config.Load()

	require.NoError(t, err)
	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, "development", cfg.Env)
	assert.Equal(t, "http://localhost:3000", cfg.CORSAllowedOrigins)
	assert.Equal(t, "dev", cfg.AppVersion)
}

func TestConfig_Load_OverridesDefaultsWithEnvVars(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://prod-host/research_events")
	t.Setenv("JWT_SECRET", "prod-secret")
	t.Setenv("PORT", "9090")
	t.Setenv("ENV", "production")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://research-events.vercel.app")
	t.Setenv("APP_VERSION", "2.1.0")

	cfg, err := config.Load()

	require.NoError(t, err)
	assert.Equal(t, "9090", cfg.Port)
	assert.Equal(t, "production", cfg.Env)
	assert.Equal(t, "https://research-events.vercel.app", cfg.CORSAllowedOrigins)
	assert.Equal(t, "2.1.0", cfg.AppVersion)
	assert.Equal(t, "postgres://prod-host/research_events", cfg.DatabaseURL)
}

// Spec: specs/backend/auth-login.yaml
// Rule: "JWT_SECRET loaded from config (env var), never hardcoded"
// Rule: "App must refuse to start if JWT_SECRET is missing"

func TestConfig_Load_ReturnsErrorWhenJWTSecretMissing(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("JWT_SECRET", "")

	_, err := config.Load()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET")
}

func TestConfig_Load_ReturnsJWTSecretFromEnv(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("JWT_SECRET", "supersecret")

	cfg, err := config.Load()

	require.NoError(t, err)
	assert.Equal(t, "supersecret", cfg.JWTSecret)
}

// Spec: specs/backend/observability-opentelemetry.yaml
// Rule: "SENTRY_DSN env var, optional. Empty -> no-op tracer, sentry.Init never called."

func TestConfig_Load_SentryDSNDefaultsToEmptyWhenUnset(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("SENTRY_DSN", "")

	cfg, err := config.Load()

	require.NoError(t, err)
	assert.Equal(t, "", cfg.SentryDSN)
}

func TestConfig_Load_SentryDSNPassedThroughFromEnv(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("SENTRY_DSN", "https://abc123@o0.ingest.de.sentry.io/1")

	cfg, err := config.Load()

	require.NoError(t, err)
	assert.Equal(t, "https://abc123@o0.ingest.de.sentry.io/1", cfg.SentryDSN)
}
