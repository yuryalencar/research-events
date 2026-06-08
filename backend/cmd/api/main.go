package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/yuryalencar/research-events/internal/config"
	"github.com/yuryalencar/research-events/internal/health"
)

func main() {
	// slog is Go's structured logger (added in 1.21). It writes JSON key-value pairs,
	// which are easy to parse in log aggregation tools (Datadog, Fly.io logs, etc.).
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// --- 1. Config ---

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// --- 2. Database ---

	// gorm.Open creates the connection pool — it does NOT ping yet.
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		logger.Error("failed to open database connection", "error", err)
		os.Exit(1)
	}

	// db.DB() returns the underlying *sql.DB so we can call PingContext directly.
	// GORM wraps *sql.DB — we need the raw handle for health checking and closing.
	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("failed to get sql.DB from gorm", "error", err)
		os.Exit(1)
	}

	// --- 3. Ping database at startup ---

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()

	if err := sqlDB.PingContext(pingCtx); err != nil {
		logger.Error("database ping failed — check DATABASE_URL and that Postgres is running", "error", err)
		os.Exit(1)
	}

	logger.Info("database connection established")

	// --- 4. Health checker registry ---

	registry := health.NewRegistry()
	registry.Register(health.NewDatabaseChecker(sqlDB))

	// --- 5. Routes + middleware ---

	// buildHandler is defined in server.go — extracted so tests can call it
	// without starting a real server or waiting for OS signals.
	httpHandler := BuildHandler(cfg, registry)

	// --- 6. HTTP server ---

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: httpHandler,
	}

	// --- 7. Start server in a goroutine ---

	// A goroutine is a lightweight concurrent function. "go func()" launches it
	// immediately and returns — main() continues without waiting for it to finish.
	// Goroutines are much cheaper than OS threads (start at ~8 KB vs 1–8 MB),
	// so Go programs routinely run thousands of them.
	//
	// We need the server in a goroutine so that main() can reach the signal-waiting
	// code below. If we called srv.ListenAndServe() directly, it would block forever
	// and we'd never get to the shutdown logic.
	go func() {
		logger.Info("server starting", "port", cfg.Port, "env", cfg.Env, "version", cfg.AppVersion)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// http.ErrServerClosed is the expected "error" when Shutdown() is called —
			// it is not a real error, so we only log unexpected ones.
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// --- 8. Block until shutdown signal ---

	// A channel is a typed pipe for communication between goroutines.
	// make(chan os.Signal, 1) creates a buffered channel that holds 1 signal value.
	// The buffer of 1 ensures the OS can deliver the signal even if we are not
	// reading from the channel at the exact moment it arrives.
	//
	// signal.Notify registers quit to receive SIGTERM (sent by Fly.io on deploy)
	// and SIGINT (sent by Ctrl+C during local development).
	//
	// <-quit blocks — the goroutine sleeps here until a signal arrives.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	sig := <-quit

	logger.Info("shutdown signal received — draining in-flight requests", "signal", sig.String())

	// --- 9. Graceful shutdown ---

	// context.WithTimeout gives in-flight requests 30 seconds to complete.
	// After that, Shutdown() closes connections forcefully.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown error", "error", err)
		os.Exit(1)
	}

	// --- 10. Close DB connection pool ---

	if err := sqlDB.Close(); err != nil {
		logger.Error("failed to close database connection", "error", err)
	}

	logger.Info("server stopped gracefully")
}
