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

	// --- Handlers ---
	authHandler := handler.NewAuthHandler(userRepo, cfg.JWTSecret, logger)
	adminUserHandler := handler.NewAdminUserHandler(userRepo, auditRepo, logger)

	// --- Middleware ---
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret, userRepo)

	// 10 requests/minute per IP — applied to login and refresh-token.
	// See specs/backend/auth-login.yaml and auth-refresh-token.yaml rule sections.
	rateLimiter := middleware.NewRateLimiter(10.0/60.0, 10)

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
