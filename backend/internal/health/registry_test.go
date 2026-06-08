package health_test

// Spec: specs/backend/server-bootstrap.yaml
// Rule: "Top-level status is healthy only if ALL registered checks pass"

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/yuryalencar/research-events/internal/health"
	"github.com/yuryalencar/research-events/internal/health/mocks"
)

func TestRegistry_RunAll_AllCheckersPass_ReturnsHealthy(t *testing.T) {
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

	// Act
	results, allHealthy := registry.RunAll(context.Background())

	// Assert
	assert.True(t, allHealthy, "expected allHealthy=true when all checkers pass")
	assert.Equal(t, health.StatusHealthy, results["database"].Status)
	assert.Equal(t, int64(4), results["database"].LatencyMs)
}

func TestRegistry_RunAll_OneCheckerFails_ReturnsUnhealthy(t *testing.T) {
	ctrl := gomock.NewController(t)

	// Arrange — one healthy checker, one unhealthy checker
	mockDB := mocks.NewMockChecker(ctrl)
	mockDB.EXPECT().Name().Return("database").AnyTimes()
	mockDB.EXPECT().Check(gomock.Any()).Return(health.CheckResult{
		Status: health.StatusHealthy,
	})

	mockCache := mocks.NewMockChecker(ctrl)
	mockCache.EXPECT().Name().Return("cache").AnyTimes()
	mockCache.EXPECT().Check(gomock.Any()).Return(health.CheckResult{
		Status: health.StatusUnhealthy,
		Error:  "connection refused",
	})

	registry := health.NewRegistry()
	registry.Register(mockDB)
	registry.Register(mockCache)

	// Act
	results, allHealthy := registry.RunAll(context.Background())

	// Assert — a single failing checker must make allHealthy false
	assert.False(t, allHealthy, "expected allHealthy=false when any checker fails")
	assert.Equal(t, health.StatusHealthy, results["database"].Status)
	assert.Equal(t, health.StatusUnhealthy, results["cache"].Status)
	assert.Equal(t, "connection refused", results["cache"].Error)
}

func TestRegistry_RunAll_NoCheckers_ReturnsHealthy(t *testing.T) {
	// Arrange — empty registry (valid startup state before any checker is registered)
	registry := health.NewRegistry()

	// Act
	results, allHealthy := registry.RunAll(context.Background())

	// Assert — no checkers means nothing is failing
	assert.True(t, allHealthy)
	assert.Empty(t, results)
}
