package middleware_test

// Spec: specs/backend/auth-login.yaml
// Rule: "Rate-limit: 10 attempts per IP per minute"
// Spec: specs/backend/auth-refresh-token.yaml
// Rule: "Rate-limit: 10 attempts per IP per minute"

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuryalencar/research-events/internal/middleware"
)

// okHandler is a minimal next handler that records it was reached.
func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestRateLimiter_AllowsRequestsUnderBurst(t *testing.T) {
	// Spec: rule — requests under the burst limit pass through unblocked.
	// burst=3 means the first 3 back-to-back requests from the same IP all succeed.
	limiter := middleware.NewRateLimiter(100, 3)
	handler := limiter.Limit(okHandler())

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, "request %d should pass", i+1)
	}
}

func TestRateLimiter_BlocksOnLimitExceeded_Returns429(t *testing.T) {
	// Spec: rule — once the burst is exhausted the next request returns 429.
	// rate=0 means no token refill; burst=1 allows exactly one request then blocks.
	limiter := middleware.NewRateLimiter(0, 1)
	handler := limiter.Limit(okHandler())

	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	require.Equal(t, http.StatusOK, rec1.Code, "first request must pass")

	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusTooManyRequests, rec2.Code)

	// Response envelope must include a code so the frontend can show the right message.
	var body map[string]any
	require.NoError(t, json.NewDecoder(rec2.Body).Decode(&body))
	assert.Equal(t, "RATE_LIMIT_EXCEEDED", body["code"])
}

func TestRateLimiter_DifferentIPsHaveIndependentCounters(t *testing.T) {
	// Spec: per-IP limiting — exhausting one IP's budget must not affect another IP.
	limiter := middleware.NewRateLimiter(0, 1)
	handler := limiter.Limit(okHandler())

	// IP A first request — passes.
	reqA1 := httptest.NewRequest(http.MethodPost, "/", nil)
	reqA1.RemoteAddr = "10.0.0.1:9000"
	recA1 := httptest.NewRecorder()
	handler.ServeHTTP(recA1, reqA1)
	require.Equal(t, http.StatusOK, recA1.Code)

	// IP A second request — blocked (burst exhausted).
	reqA2 := httptest.NewRequest(http.MethodPost, "/", nil)
	reqA2.RemoteAddr = "10.0.0.1:9000"
	recA2 := httptest.NewRecorder()
	handler.ServeHTTP(recA2, reqA2)
	assert.Equal(t, http.StatusTooManyRequests, recA2.Code)

	// IP B first request — passes (independent counter, full burst available).
	reqB := httptest.NewRequest(http.MethodPost, "/", nil)
	reqB.RemoteAddr = "10.0.0.2:9000"
	recB := httptest.NewRecorder()
	handler.ServeHTTP(recB, reqB)
	assert.Equal(t, http.StatusOK, recB.Code)
}
