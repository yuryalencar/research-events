package main

import (
	"log/slog"
	"net/http"

	"gorm.io/gorm"

	"github.com/yuryalencar/research-events/internal/config"
	"github.com/yuryalencar/research-events/internal/handler"
	"github.com/yuryalencar/research-events/internal/health"
	"github.com/yuryalencar/research-events/internal/middleware"
	"github.com/yuryalencar/research-events/internal/repository"
)

// BuildHandler wires all routes and middleware into a single http.Handler.
// Exported so server_test.go can call it directly without starting a real server.
//
// db may be nil in tests that only exercise routes which do not reach the DB layer
// (e.g. requests that fail at validation or cookie-check before any repo call).
func BuildHandler(cfg config.Config, db *gorm.DB, registry *health.Registry, logger *slog.Logger) http.Handler {
	// --- Repositories ---
	userRepo := repository.NewUserRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	eventRepo := repository.NewEventRepository(db)

	// --- Handlers ---
	authHandler := handler.NewAuthHandler(userRepo, cfg.JWTSecret, logger)
	adminUserHandler := handler.NewAdminUserHandler(userRepo, auditRepo, logger)
	eventHandler := handler.NewEventHandler(eventRepo, logger)

	// --- Middleware ---
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret, userRepo)

	// 10 requests/minute per IP — applied to login and refresh-token.
	// See specs/backend/auth-login.yaml and auth-refresh-token.yaml rule sections.
	rateLimiter := middleware.NewRateLimiter(10.0/60.0, 10)

	// 50 requests/minute per IP — applied to public event submission.
	// See specs/backend/events-submit.yaml rule "rate-limit 50 requests per minute per IP".
	submitRateLimiter := middleware.NewRateLimiter(50.0/60.0, 50)

	// --- Routes ---
	mux := http.NewServeMux()

	// Health endpoint — no auth, no rate-limit.
	mux.Handle("GET /health", handler.NewHealthHandler(registry, cfg.AppVersion))

	// Public auth endpoints.
	mux.Handle("POST /api/v1/auth/login",
		rateLimiter.Limit(http.HandlerFunc(authHandler.Login)))

	mux.Handle("POST /api/v1/auth/refresh-token",
		rateLimiter.Limit(http.HandlerFunc(authHandler.RefreshToken)))

	// Logout accepts expired tokens (graceful) — no rate-limit needed (not a brute-force target).
	mux.Handle("POST /api/v1/auth/logout",
		http.HandlerFunc(authHandler.Logout))

	// Public event submission — no auth, rate-limited per events-submit.yaml.
	mux.Handle("POST /api/v1/events/submit",
		submitRateLimiter.Limit(http.HandlerFunc(eventHandler.Submit)))

	// Admin-only endpoints — RequireAuth then RequireRole("admin").
	mux.Handle("PATCH /api/v1/admin/users/{id}/unlock",
		authMiddleware.RequireAuth(
			middleware.RequireRole("admin")(
				http.HandlerFunc(adminUserHandler.Unlock),
			),
		),
	)

	// CORS wraps the entire mux — every response gets the appropriate headers.
	return middleware.CORS(cfg.CORSAllowedOrigins)(mux)
}
