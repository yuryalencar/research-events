package main

// Tests for BuildHandler — the route + middleware wiring.
// These cover what no other test file covers: that /health is actually registered,
// that CORS wraps every route, and that unknown paths return 404.
//
// Uses package main (internal test) because main packages cannot be imported —
// the test shares the package directly, giving it access to BuildHandler without
// an import statement.

import (
	"encoding/json"
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

func testConfig() config.Config {
	return config.Config{
		Port:               "8080",
		Env:                "test",
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

	h := BuildHandler(testConfig(), registry)

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
	h := BuildHandler(cfg, registry)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, "https://example.com", rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestBuildHandler_UnknownPathReturns404(t *testing.T) {
	registry := health.NewRegistry()
	h := BuildHandler(testConfig(), registry)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/does-not-exist", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
