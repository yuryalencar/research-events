package health

import "context"

//go:generate mockgen -source=health.go -destination=mocks/mock_checker.go -package=mocks

// Checker is the contract every health dependency must implement.
//
// Go interfaces are satisfied implicitly — any type with Name() and Check()
// automatically satisfies Checker without an "implements" keyword.
// This lets new checkers (cache, external API, etc.) be added by registering
// them in main.go; the handler and registry never need to change.
type Checker interface {
	Name() string
	Check(ctx context.Context) CheckResult
}

// CheckResult holds the outcome of a single dependency check.
type CheckResult struct {
	Status    string `json:"status"`
	LatencyMs int64  `json:"latency_ms,omitempty"`
	Error     string `json:"error,omitempty"`
}

// HealthResponse is the full JSON body returned by GET /health.
type HealthResponse struct {
	Status    string                 `json:"status"`
	Version   string                 `json:"version"`
	Timestamp string                 `json:"timestamp"`
	Uptime    string                 `json:"uptime"`
	Checks    map[string]CheckResult `json:"checks"`
}

const (
	StatusHealthy   = "healthy"
	StatusUnhealthy = "unhealthy"
)
