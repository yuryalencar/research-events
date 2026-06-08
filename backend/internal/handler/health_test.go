package handler_test

// Spec: specs/backend/server-bootstrap.yaml
// DoD: TestHealth_AllChecksPass → 200 + healthy
//      TestHealth_DatabaseUnhealthy → 503 + unhealthy + error populated
//      TestHealth_NewCheckerUnhealthy → failing checker propagates to top-level status

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/yuryalencar/research-events/internal/handler"
	"github.com/yuryalencar/research-events/internal/health"
	"github.com/yuryalencar/research-events/internal/health/mocks"
)

func TestHealthHandler_AllChecksPass_Returns200WithHealthyStatus(t *testing.T) {
	ctrl := gomock.NewController(t)

	// Arrange — one checker that reports healthy
	mockChecker := mocks.NewMockChecker(ctrl)
	mockChecker.EXPECT().Name().Return("database").AnyTimes()
	mockChecker.EXPECT().Check(gomock.Any()).Return(health.CheckResult{
		Status:    health.StatusHealthy,
		LatencyMs: 4,
	})

	registry := health.NewRegistry()
	registry.Register(mockChecker)

	h := handler.NewHealthHandler(registry, "1.0.0")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	// Act
	h.ServeHTTP(rec, req)

	// Assert — HTTP layer
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	// Assert — response body
	var body health.HealthResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))

	assert.Equal(t, health.StatusHealthy, body.Status)
	assert.Equal(t, "1.0.0", body.Version)
	assert.NotEmpty(t, body.Timestamp)
	assert.NotEmpty(t, body.Uptime)
	assert.Equal(t, health.StatusHealthy, body.Checks["database"].Status)
	assert.Equal(t, int64(4), body.Checks["database"].LatencyMs)
}

func TestHealthHandler_DatabaseUnhealthy_Returns503WithErrorPopulated(t *testing.T) {
	ctrl := gomock.NewController(t)

	// Arrange — checker reports DB is down
	mockChecker := mocks.NewMockChecker(ctrl)
	mockChecker.EXPECT().Name().Return("database").AnyTimes()
	mockChecker.EXPECT().Check(gomock.Any()).Return(health.CheckResult{
		Status: health.StatusUnhealthy,
		Error:  "connection refused",
	})

	registry := health.NewRegistry()
	registry.Register(mockChecker)

	h := handler.NewHealthHandler(registry, "1.0.0")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	// Act
	h.ServeHTTP(rec, req)

	// Assert — 503 when any checker fails
	require.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var body health.HealthResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))

	assert.Equal(t, health.StatusUnhealthy, body.Status)
	assert.Equal(t, health.StatusUnhealthy, body.Checks["database"].Status)
	assert.Equal(t, "connection refused", body.Checks["database"].Error)
}

func TestHealthHandler_NewCheckerUnhealthy_PropagatesTopLevelStatus(t *testing.T) {
	ctrl := gomock.NewController(t)

	// Arrange — DB healthy, a second checker (e.g. future cache) unhealthy
	mockDB := mocks.NewMockChecker(ctrl)
	mockDB.EXPECT().Name().Return("database").AnyTimes()
	mockDB.EXPECT().Check(gomock.Any()).Return(health.CheckResult{
		Status: health.StatusHealthy,
	})

	mockCache := mocks.NewMockChecker(ctrl)
	mockCache.EXPECT().Name().Return("cache").AnyTimes()
	mockCache.EXPECT().Check(gomock.Any()).Return(health.CheckResult{
		Status: health.StatusUnhealthy,
		Error:  "cache unavailable",
	})

	registry := health.NewRegistry()
	registry.Register(mockDB)
	registry.Register(mockCache)

	h := handler.NewHealthHandler(registry, "1.0.0")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	// Act
	h.ServeHTTP(rec, req)

	// Assert — one failing checker is enough to make the top-level status unhealthy
	require.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var body health.HealthResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))

	assert.Equal(t, health.StatusUnhealthy, body.Status)
	assert.Equal(t, health.StatusHealthy, body.Checks["database"].Status)
	assert.Equal(t, health.StatusUnhealthy, body.Checks["cache"].Status)
	assert.Equal(t, "cache unavailable", body.Checks["cache"].Error)
}
