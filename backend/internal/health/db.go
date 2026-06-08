package health

import (
	"context"
	"time"
)

//go:generate mockgen -source=db.go -destination=mocks/mock_db_pinger.go -package=mocks

// DBPinger is the minimal database interface needed for health checking.
// Defining a narrow interface (one method) instead of depending on *sql.DB directly
// lets tests inject a mock without a real database connection.
// *sql.DB satisfies DBPinger automatically — no "implements" keyword needed in Go.
type DBPinger interface {
	PingContext(ctx context.Context) error
}

// DatabaseChecker verifies the Postgres connection is alive.
// It satisfies the Checker interface and is registered in main.go.
type DatabaseChecker struct {
	db DBPinger
}

// NewDatabaseChecker creates a DatabaseChecker.
// In production, pass the *sql.DB obtained from gorm.DB.DB() — it satisfies DBPinger.
// In tests, pass a generated MockDBPinger.
func NewDatabaseChecker(db DBPinger) *DatabaseChecker {
	return &DatabaseChecker{db: db}
}

func (c *DatabaseChecker) Name() string {
	return "database"
}

// Check pings the database and returns the result.
// A 3-second timeout is applied — if the DB does not respond in time,
// the context is cancelled and the check returns unhealthy.
//
// Note: PingContext sends a lightweight round-trip to verify the connection
// is alive, which is equivalent to SELECT 1 for health-checking purposes.
func (c *DatabaseChecker) Check(ctx context.Context) CheckResult {
	// context.WithTimeout creates a child context that auto-cancels after 3 seconds.
	// defer cancel() releases the timer resources even if the ping finishes early.
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	start := time.Now()
	err := c.db.PingContext(ctx)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return CheckResult{Status: StatusUnhealthy, Error: err.Error()}
	}
	return CheckResult{Status: StatusHealthy, LatencyMs: latency}
}
