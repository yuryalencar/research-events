package main

import (
	"net/http"

	"github.com/yuryalencar/research-events/internal/config"
	"github.com/yuryalencar/research-events/internal/handler"
	"github.com/yuryalencar/research-events/internal/health"
	"github.com/yuryalencar/research-events/internal/middleware"
)

// BuildHandler wires routes and middleware into a single http.Handler.
// Exported so server_test.go can call it directly — same-directory test files
// in package main_test have access to exported symbols without importing the package.
func BuildHandler(cfg config.Config, registry *health.Registry) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/health", handler.NewHealthHandler(registry, cfg.AppVersion))

	return middleware.CORS(cfg.CORSAllowedOrigins)(mux)
}
