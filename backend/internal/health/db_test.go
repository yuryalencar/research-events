package health_test

// Spec: specs/backend/server-bootstrap.yaml
// Rule: "DatabaseChecker runs SELECT 1 with a 3-second context timeout"
// DoD:  "GET /health returns 503 + unhealthy body when DB is down"

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/yuryalencar/research-events/internal/health"
	"github.com/yuryalencar/research-events/internal/health/mocks"
)

func TestDatabaseChecker_Name_ReturnsDatabaseString(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockPinger := mocks.NewMockDBPinger(ctrl)

	checker := health.NewDatabaseChecker(mockPinger)

	assert.Equal(t, "database", checker.Name())
}

func TestDatabaseChecker_Check_ReturnsHealthyWhenPingSucceeds(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockPinger := mocks.NewMockDBPinger(ctrl)
	mockPinger.EXPECT().PingContext(gomock.Any()).Return(nil)

	checker := health.NewDatabaseChecker(mockPinger)
	result := checker.Check(context.Background())

	assert.Equal(t, health.StatusHealthy, result.Status)
	assert.Empty(t, result.Error)
	assert.GreaterOrEqual(t, result.LatencyMs, int64(0))
}

func TestDatabaseChecker_Check_ReturnsUnhealthyWhenPingFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockPinger := mocks.NewMockDBPinger(ctrl)
	mockPinger.EXPECT().PingContext(gomock.Any()).Return(errors.New("connection refused"))

	checker := health.NewDatabaseChecker(mockPinger)
	result := checker.Check(context.Background())

	// Spec: GET /health returns 503 when DB is down — error must be populated
	assert.Equal(t, health.StatusUnhealthy, result.Status)
	assert.Equal(t, "connection refused", result.Error)
}

func TestDatabaseChecker_Check_ReturnsUnhealthyWhenContextCancelled(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockPinger := mocks.NewMockDBPinger(ctrl)
	// A cancelled context causes PingContext to return context.Canceled
	mockPinger.EXPECT().PingContext(gomock.Any()).Return(context.Canceled)

	checker := health.NewDatabaseChecker(mockPinger)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before the check runs

	result := checker.Check(ctx)

	assert.Equal(t, health.StatusUnhealthy, result.Status)
	assert.NotEmpty(t, result.Error)
}
