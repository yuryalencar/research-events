package middleware

import "net/http"

// CORS returns a middleware that sets Cross-Origin Resource Sharing headers on every response.
//
// allowedOrigin is written into Access-Control-Allow-Origin.
// Default is "http://localhost:3000" (dev). Set CORS_ALLOWED_ORIGINS to the Vercel URL in production.
//
// When the origin is a specific URL (not "*"), the middleware also sets:
//   - Access-Control-Allow-Credentials: true  — required for HTTP-only cookies to be sent
//   - Vary: Origin                            — tells CDNs to cache per origin, not share one response
//
// The browser enforces that Allow-Credentials: true is incompatible with Allow-Origin: *.
// Never combine them — the browser will block the response silently.
//
// Go middleware pattern: a function that wraps an http.Handler and returns a new http.Handler.
// Each middleware in the chain calls next.ServeHTTP to pass control to the next layer.
func CORS(allowedOrigin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// Wildcard origin cannot be paired with credentials — browsers block that combination.
			// Only set these headers when a specific origin is configured.
			if allowedOrigin != "*" {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Vary", "Origin")
			}

			// OPTIONS is a browser preflight — respond immediately without hitting the real handler.
			// The browser sends this first to check if the actual request (GET/POST/etc.) is allowed.
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
