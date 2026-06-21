package main

import (
	"log/slog"
	"net/http"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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
//
// tp is the OTel TracerProvider (see internal/observability.InitTracerProvider).
// otelhttp.NewHandler wraps the whole mux as the outermost layer, so every
// request gets a root span. Go 1.22+'s http.ServeMux records the matched
// pattern on the request as r.Pattern after dispatch; otelhttp reads that
// field to rename the span (e.g. "GET /health"), but — as of v0.69 — does
// not turn it into an "http.route" span attribute on its own. traced (below)
// fills that gap.
func BuildHandler(cfg config.Config, db *gorm.DB, registry *health.Registry, logger *slog.Logger, tp trace.TracerProvider) http.Handler {
	// --- Repositories ---
	userRepo := repository.NewUserRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	eventRepo := repository.NewEventRepository(db)

	// --- Handlers ---
	authHandler := handler.NewAuthHandler(userRepo, cfg.JWTSecret, logger)
	adminUserHandler := handler.NewAdminUserHandler(userRepo, auditRepo, logger)
	adminEventHandler := handler.NewAdminEventHandler(eventRepo, logger)
	eventHandler := handler.NewEventHandler(eventRepo, logger)
	userHandler := handler.NewUserHandler(userRepo, logger)

	// --- Middleware ---
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret, userRepo)

	// 10 requests/minute per IP — applied to login and refresh-token.
	// See specs/backend/auth-login.yaml and auth-refresh-token.yaml rule sections.
	rateLimiter := middleware.NewRateLimiter(10.0/60.0, 10)

	// 50 requests/minute per IP — shared by public write endpoints (event
	// submission and adding deadlines to an approved event).
	// See specs/backend/events-submit.yaml and events-deadlines-add.yaml,
	// both specifying "rate-limit 50 requests per minute per IP".
	publicRateLimiter := middleware.NewRateLimiter(50.0/60.0, 50)

	// 120 requests/minute per IP, burst 30 — applied to the events list endpoint.
	// See specs/backend/events-list.yaml rule "Highest limit in the app since this
	// is the primary endpoint (globe + list views)".
	publicHighRateLimiter := middleware.NewRateLimiter(120.0/60.0, 30)

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
		publicRateLimiter.Limit(http.HandlerFunc(eventHandler.Submit)))

	// Public event list — no auth, rate-limited per events-list.yaml.
	mux.Handle("GET /api/v1/events",
		publicHighRateLimiter.Limit(http.HandlerFunc(eventHandler.List)))

	// Public deadline additions to an approved event — no auth, rate-limited
	// per events-deadlines-add.yaml.
	mux.Handle("POST /api/v1/events/{id}/deadlines",
		publicRateLimiter.Limit(http.HandlerFunc(eventHandler.AddDeadlines)))

	// Public deadline cancellation on an approved event — no auth, rate-limited
	// per events-deadlines-cancel.yaml.
	mux.Handle("PATCH /api/v1/events/{eventId}/deadlines/{deadlineId}/cancel",
		publicRateLimiter.Limit(http.HandlerFunc(eventHandler.CancelDeadline)))

	// Public deadline supersession on an approved event — no auth, rate-limited
	// per events-deadlines-supersede.yaml.
	mux.Handle("POST /api/v1/events/{eventId}/deadlines/{deadlineId}/supersede",
		publicRateLimiter.Limit(http.HandlerFunc(eventHandler.Supersede)))

	// Authenticated user — update own password.
	// Rate-limited same as login (10/min/IP) per specs/backend/users-update-password.yaml.
	// Rate limiter wraps RequireAuth so unauthenticated attempts still consume a token.
	mux.Handle("PATCH /api/v1/users/me/password",
		rateLimiter.Limit(
			authMiddleware.RequireAuth(
				http.HandlerFunc(userHandler.UpdatePassword),
			),
		),
	)

	// Admin-only endpoints — RequireAuth then RequireRole("admin").
	mux.Handle("PATCH /api/v1/admin/users/{id}/unlock",
		authMiddleware.RequireAuth(
			middleware.RequireRole("admin")(
				http.HandlerFunc(adminUserHandler.Unlock),
			),
		),
	)

	// Admin/moderator endpoint — RequireAuth then RequireRole("admin", "moderator").
	mux.Handle("PATCH /api/v1/admin/events/{id}/review",
		authMiddleware.RequireAuth(
			middleware.RequireRole("admin", "moderator")(
				http.HandlerFunc(adminEventHandler.Review),
			),
		),
	)

	// CORS wraps the entire mux — every response gets the appropriate headers.
	corsHandler := middleware.CORS(cfg.CORSAllowedOrigins)(mux)

	// otelhttp wraps everything else — outermost layer, so it creates the root
	// span before any other middleware runs, and sees the final status code
	// for every request. traced runs inside that span (see below) to record
	// the matched route once the mux has dispatched the request.
	return otelhttp.NewHandler(traced(corsHandler), "http.server", otelhttp.WithTracerProvider(tp))
}

// traced records the "http.route" attribute on the current span (the root
// span created by otelhttp.NewHandler) once the request has been dispatched.
//
// Required because otelhttp v0.69 reads r.Pattern (set by http.ServeMux during
// next.ServeHTTP) only to rename the span — it does not add an "http.route"
// span attribute. Per the observability spec, "http.route" is a required
// attribute on every request span, so we add it ourselves here.
func traced(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)

		if route := routeFromPattern(r.Pattern); route != "" {
			trace.SpanFromContext(r.Context()).SetAttributes(attribute.String("http.route", route))
		}
	})
}

// routeFromPattern strips the leading "METHOD " (if any) from a ServeMux
// pattern, e.g. "GET /health" -> "/health". Patterns registered without a
// method (rare in this codebase) already start with "/" and pass through
// unchanged.
func routeFromPattern(pattern string) string {
	if idx := strings.IndexByte(pattern, '/'); idx >= 0 {
		return pattern[idx:]
	}
	return ""
}
