package middleware_test

// Spec: specs/backend/server-bootstrap.yaml + specs/backend/auth-middleware.yaml
// Rule: "Sets Access-Control-Allow-Origin on every response"
// Rule: "Handles OPTIONS preflight requests — returns 204 with appropriate headers"
// Rule: "Sets Access-Control-Allow-Credentials: true for specific origins (required for HTTP-only cookies)"
// Rule: "Never sets Allow-Credentials when origin is * (browser rejects that combination)"
// Rule: "Default origin is http://localhost:3000 (not * — wildcard blocks cookies)"

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuryalencar/research-events/internal/middleware"
)

// nextHandler is a simple handler that records whether it was called.
func nextHandler(called *bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*called = true
		w.WriteHeader(http.StatusOK)
	})
}

func TestCORS_SetsAccessControlAllowOriginOnEveryResponse(t *testing.T) {
	called := false
	handler := middleware.CORS("*")(nextHandler(&called))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Spec: "Sets Access-Control-Allow-Origin on every response"
	assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.True(t, called, "next handler should be called for non-OPTIONS requests")
}

func TestCORS_UsesConfiguredOriginInsteadOfWildcard(t *testing.T) {
	called := false
	handler := middleware.CORS("https://research-events.vercel.app")(nextHandler(&called))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, "https://research-events.vercel.app", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.True(t, called)
}

func TestCORS_PreflightOptionsReturns204AndDoesNotCallNext(t *testing.T) {
	called := false
	handler := middleware.CORS("*")(nextHandler(&called))

	req := httptest.NewRequest(http.MethodOptions, "/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Spec: "Handles OPTIONS preflight requests — returns 204 with appropriate headers"
	require.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.False(t, called, "next handler must NOT be called for OPTIONS preflight")
}

func TestCORS_NonOptionsRequestCallsNextHandler(t *testing.T) {
	called := false
	handler := middleware.CORS("*")(nextHandler(&called))

	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodPatch} {
		called = false
		req := httptest.NewRequest(method, "/api/v1/events", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.True(t, called, "next handler should be called for %s", method)
	}
}

// Spec: specs/backend/auth-middleware.yaml
// Rule: "Cookies require Allow-Credentials: true + specific origin — wildcard blocks them (browser rule)"

func TestCORS_SetsAllowCredentialsForSpecificOrigin(t *testing.T) {
	handler := middleware.CORS("https://research-events.vercel.app")(nextHandler(new(bool)))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/events", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCORS_DoesNotSetAllowCredentialsForWildcard(t *testing.T) {
	// Browser rejects Allow-Credentials: true + Allow-Origin: * — never combine them.
	handler := middleware.CORS("*")(nextHandler(new(bool)))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/events", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCORS_SetsVaryOriginHeader(t *testing.T) {
	// Vary: Origin tells CDNs/proxies to cache responses per origin when Allow-Origin is dynamic.
	handler := middleware.CORS("https://research-events.vercel.app")(nextHandler(new(bool)))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/events", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, "Origin", rec.Header().Get("Vary"))
}
