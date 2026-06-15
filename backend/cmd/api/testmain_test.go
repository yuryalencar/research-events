package main

// Integration test setup for BuildHandler tests that need a real database
// (the GORM OTel tracing plugin only emits spans for real queries).
//
// Mirrors internal/repository/testmain_test.go — see that file for the
// docker-compose / migration setup required before running these tests.

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
	"gorm.io/plugin/opentelemetry/tracing"
)

// testDB is the shared GORM connection used by DB-backed BuildHandler tests.
// Each test opens its own transaction and rolls it back — testDB itself is never mutated.
//
// The GORM OTel tracing plugin is registered once here (via db.Use), so every
// transaction derived from testDB (via Begin) emits DB spans too.
var testDB *gorm.DB

func TestMain(m *testing.M) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = "postgres://postgres:postgres@host.orb.internal:5433/research_events_test?sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(url), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "cmd/api tests: failed to connect to TEST_DATABASE_URL %q: %v\n", url, err)
		os.Exit(1)
	}

	sqlDB, err := db.DB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cmd/api tests: failed to get sql.DB: %v\n", err)
		os.Exit(1)
	}

	_, filename, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(filename), "../../migrations")

	goose.SetLogger(goose.NopLogger())
	if err := goose.Up(sqlDB, migrationsDir); err != nil {
		fmt.Fprintf(os.Stderr, "cmd/api tests: goose migration failed: %v\n", err)
		os.Exit(1)
	}

	// WithoutQueryVariables is REQUIRED — without it, SQL parameter values
	// (e.g. submitter emails) would be embedded verbatim into span attributes.
	//
	// No WithTracerProvider option is passed, so the plugin defaults to
	// otel.GetTracerProvider() — the global *delegating* tracer provider. It
	// starts as a no-op, but each test calls otel.SetTracerProvider(tp) with
	// its own in-memory exporter before making a request; the global proxy
	// retroactively delegates to that tp, so the plugin (registered once,
	// here) ends up emitting its DB spans into whichever tp the current test
	// installed.
	if err := db.Use(tracing.NewPlugin(tracing.WithoutQueryVariables())); err != nil {
		fmt.Fprintf(os.Stderr, "cmd/api tests: failed to register gorm tracing plugin: %v\n", err)
		os.Exit(1)
	}

	testDB = db
	os.Exit(m.Run())
}

// beginTx starts a transaction and returns a GORM DB scoped to it.
// The transaction is always rolled back, keeping the test database clean.
func beginTx(t *testing.T) (*gorm.DB, func()) {
	t.Helper()
	tx := testDB.Begin()
	if tx.Error != nil {
		t.Fatalf("beginTx: failed to start transaction: %v", tx.Error)
	}
	return tx, func() { tx.Rollback() }
}
