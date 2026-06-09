package repository_test

// Integration test setup for the repository layer.
//
// Requirements before running:
//   1. docker compose up -d           (starts postgres_test on :5433)
//   2. make migrate-test-up           (runs goose migrations against the test DB)
//   3. go test ./internal/repository/...
//
// TEST_DATABASE_URL defaults to the docker-compose test instance.
// Override it to point at any Postgres 16 instance with the migrations applied.

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// testDB is the shared GORM connection used by all repository integration tests.
// Each test opens its own transaction and rolls it back — testDB itself is never mutated.
var testDB *gorm.DB

func TestMain(m *testing.M) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		// Default to the docker-compose test instance.
		// Inside OrbStack's Linux VM (this workspace) the Mac's Docker ports are
		// reachable via host.orb.internal rather than localhost.
		url = "postgres://postgres:postgres@host.orb.internal:5433/research_events_test?sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(url), &gorm.Config{
		// Silence query logs during tests — only errors are shown.
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "repository tests: failed to connect to TEST_DATABASE_URL %q: %v\n", url, err)
		os.Exit(1)
	}

	sqlDB, err := db.DB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "repository tests: failed to get sql.DB: %v\n", err)
		os.Exit(1)
	}

	// Locate the migrations directory relative to this file.
	// runtime.Caller(0) returns the absolute path of the current source file,
	// so this works regardless of where `go test` is invoked from.
	_, filename, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(filename), "../../migrations")

	goose.SetLogger(goose.NopLogger())
	if err := goose.Up(sqlDB, migrationsDir); err != nil {
		fmt.Fprintf(os.Stderr, "repository tests: goose migration failed: %v\n", err)
		os.Exit(1)
	}

	testDB = db
	os.Exit(m.Run())
}

// beginTx starts a transaction and returns a GORM DB scoped to it.
// Call the returned rollback func in defer — the transaction is always rolled back,
// keeping the test database clean between runs.
//
// Usage:
//
//	tx, rollback := beginTx(t)
//	defer rollback()
//	repo := NewUserRepository(tx)
func beginTx(t *testing.T) (*gorm.DB, func()) {
	t.Helper()
	tx := testDB.Begin()
	if tx.Error != nil {
		t.Fatalf("beginTx: failed to start transaction: %v", tx.Error)
	}
	return tx, func() { tx.Rollback() }
}
