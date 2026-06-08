package middleware

import "net/http"

// CORS returns a middleware that sets Cross-Origin Resource Sharing headers on every response.
//
// allowedOrigins is written directly into Access-Control-Allow-Origin.
// Pass "*" in development; set a specific origin (e.g. "https://app.vercel.app") in production
// via the CORS_ALLOWED_ORIGINS environment variable.
//
// Go middleware pattern: a function that wraps an http.Handler and returns a new http.Handler.
// Each middleware in the chain calls next.ServeHTTP to pass control to the next layer.
func CORS(allowedOrigins string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigins)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

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
