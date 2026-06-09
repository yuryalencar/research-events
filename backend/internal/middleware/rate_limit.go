package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

// --- Types ---

// RateLimiter holds a per-IP token bucket limiter map.
//
// Token bucket algorithm: each IP starts with b tokens (burst). Every request consumes
// one token. Tokens refill at r per second. When the bucket is empty the request is
// rejected with 429. This is smoother than a fixed-window counter — it prevents burst
// abuse at window boundaries without penalising steady legitimate traffic.
type RateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	r        rate.Limit
	b        int
}

// --- Constructor ---

// NewRateLimiter creates a RateLimiter where each IP gets a token bucket that refills
// at r tokens/second with a maximum burst of b tokens.
//
// For 10 requests/minute use: NewRateLimiter(10.0/60.0, 10)
func NewRateLimiter(r float64, b int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		r:        rate.Limit(r),
		b:        b,
	}
}

// --- Public methods ---

// Limit returns a middleware that enforces the per-IP rate limit.
// Requests that exceed the limit receive 429 with code RATE_LIMIT_EXCEEDED.
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := extractIP(r.RemoteAddr)

		if !rl.getLimiter(ip).Allow() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]any{
				"code":  "RATE_LIMIT_EXCEEDED",
				"error": map[string]string{"message": "too many requests"},
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// --- Private functions ---

// getLimiter returns the existing limiter for ip, or creates one if none exists.
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if l, ok := rl.limiters[ip]; ok {
		return l
	}
	l := rate.NewLimiter(rl.r, rl.b)
	rl.limiters[ip] = l
	return l
}

// extractIP strips the port from an addr in "host:port" format.
// Falls back to the full addr string if parsing fails (e.g. Unix sockets in tests).
func extractIP(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}
