package main

// Tests for BuildHandler — the route + middleware wiring.
// These cover what no other test file covers: that routes are actually registered,
// CORS wraps every response, and unknown paths return 404.
//
// Uses package main (internal test) because main packages cannot be imported —
// the test shares the package directly, giving it access to BuildHandler without
// an import statement.
//
// db is passed as nil — existing tests only exercise /health which has no DB dependency.
// Auth route tests send requests that fail at the middleware/validation layer before any
// DB call, so nil repositories are safe.

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/yuryalencar/research-events/internal/config"
	"github.com/yuryalencar/research-events/internal/health"
	"github.com/yuryalencar/research-events/internal/health/mocks"
)

// discardLogger silences all log output during tests.
var discardLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

func testConfig() config.Config {
	return config.Config{
		Port:               "8080",
		Env:                "test",
		JWTSecret:          "test-secret",
		CORSAllowedOrigins: "*",
		AppVersion:         "test-version",
	}
}

func TestBuildHandler_HealthRouteRegistered_Returns200WhenHealthy(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockChecker := mocks.NewMockChecker(ctrl)
	mockChecker.EXPECT().Name().Return("database").AnyTimes()
	mockChecker.EXPECT().Check(gomock.Any()).Return(health.CheckResult{Status: health.StatusHealthy})

	registry := health.NewRegistry()
	registry.Register(mockChecker)

	h := BuildHandler(testConfig(), nil, registry, discardLogger)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var body health.HealthResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.Equal(t, health.StatusHealthy, body.Status)
	assert.Equal(t, "test-version", body.Version)
}

func TestBuildHandler_CORSHeaderPresentOnEveryResponse(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockChecker := mocks.NewMockChecker(ctrl)
	mockChecker.EXPECT().Name().Return("database").AnyTimes()
	mockChecker.EXPECT().Check(gomock.Any()).Return(health.CheckResult{Status: health.StatusHealthy})

	registry := health.NewRegistry()
	registry.Register(mockChecker)

	cfg := testConfig()
	cfg.CORSAllowedOrigins = "https://example.com"
	h := BuildHandler(cfg, nil, registry, discardLogger)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, "https://example.com", rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestBuildHandler_UnknownPathReturns404(t *testing.T) {
	h := BuildHandler(testConfig(), nil, health.NewRegistry(), discardLogger)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/does-not-exist", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestBuildHandler_AuthLoginRouteRegistered(t *testing.T) {
	// Verifies the route is wired — empty body returns 400 (validation), not 404.
	h := BuildHandler(testConfig(), nil, health.NewRegistry(), discardLogger)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.NotEqual(t, http.StatusNotFound, rec.Code, "POST /api/v1/auth/login must be registered")
}

func TestBuildHandler_AuthRefreshTokenRouteRegistered(t *testing.T) {
	// No cookie → 401 REFRESH_TOKEN_MISSING, not 404.
	h := BuildHandler(testConfig(), nil, health.NewRegistry(), discardLogger)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh-token", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.NotEqual(t, http.StatusNotFound, rec.Code, "POST /api/v1/auth/refresh-token must be registered")
}

func TestBuildHandler_AuthLogoutRouteRegistered(t *testing.T) {
	// No cookie → 401 TOKEN_MISSING, not 404.
	h := BuildHandler(testConfig(), nil, health.NewRegistry(), discardLogger)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.NotEqual(t, http.StatusNotFound, rec.Code, "POST /api/v1/auth/logout must be registered")
}

func TestBuildHandler_EventsListRouteRegistered(t *testing.T) {
	// ?year=abc fails validation before any repository call, so db=nil is safe —
	// 400 (not 404) proves the route is wired.
	h := BuildHandler(testConfig(), nil, health.NewRegistry(), discardLogger)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/events?year=abc", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code, "GET /api/v1/events must be registered")
}

func TestBuildHandler_EventsListRateLimited_Returns429AfterBurst(t *testing.T) {
	// Spec: events-list.yaml border_case "31st request in the same minute from one
	// IP (burst 30 exhausted) -> 429 RATE_LIMIT_EXCEEDED".
	h := BuildHandler(testConfig(), nil, health.NewRegistry(), discardLogger)

	var lastCode int
	for i := 0; i < 31; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/events?year=abc", nil)
		req.RemoteAddr = "203.0.113.7:1234"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		lastCode = rec.Code
	}

	assert.Equal(t, http.StatusTooManyRequests, lastCode)
}

func TestBuildHandler_AdminUnlockRouteRegistered(t *testing.T) {
	// No token cookie → 401 TOKEN_MISSING from RequireAuth, not 404.
	h := BuildHandler(testConfig(), nil, health.NewRegistry(), discardLogger)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/users/1/unlock", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.NotEqual(t, http.StatusNotFound, rec.Code, "PATCH /api/v1/admin/users/{id}/unlock must be registered")
}
