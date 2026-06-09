package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/yuryalencar/research-events/internal/repository"
)

// --- Types ---

// authContextKey is an unexported type used as the context key for AuthUser.
// Using a dedicated unexported type prevents collisions with keys from other packages.
type authContextKey struct{}

// AuthUser holds the authenticated user's identity, extracted from the JWT claims
// and confirmed against the database. Set by RequireAuth, read by RequireRole and handlers.
type AuthUser struct {
	ID    uint
	Name  string
	Email string
	Role  string
}

// AuthMiddleware validates the access token on every protected request.
type AuthMiddleware struct {
	secret   string
	userRepo repository.UserRepository
}

// --- Constructor ---

func NewAuthMiddleware(secret string, userRepo repository.UserRepository) *AuthMiddleware {
	return &AuthMiddleware{secret: secret, userRepo: userRepo}
}

// --- Public methods ---

// RequireAuth validates the access_token cookie and attaches AuthUser to context.
//
// Validation order (each step short-circuits on failure):
//  1. Cookie present          → missing       → 401 TOKEN_MISSING
//  2. Signature valid         → invalid       → 401 TOKEN_INVALID
//  3. Token not expired       → expired       → 401 TOKEN_EXPIRED  (distinct: frontend retries with refresh)
//  4. User found in DB        → not found     → 401 TOKEN_INVALID
//  5. JTI matches DB value    → mismatch      → 401 TOKEN_INVALID  (revoked / logged out)
//  6. Account not locked      → locked        → 423 ACCOUNT_LOCKED
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("access_token")
		if err != nil {
			writeAuthJSON(w, http.StatusUnauthorized, "TOKEN_MISSING", "authentication token is missing")
			return
		}

		// Parse with claims validation disabled so we can distinguish TOKEN_EXPIRED from
		// TOKEN_INVALID — jwt.Parse would otherwise lump both into the same error.
		token, err := jwt.Parse(cookie.Value,
			func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrTokenSignatureInvalid
				}
				return []byte(m.secret), nil
			},
			jwt.WithoutClaimsValidation(),
		)
		if err != nil || !token.Valid {
			writeAuthJSON(w, http.StatusUnauthorized, "TOKEN_INVALID", "authentication token is invalid")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			writeAuthJSON(w, http.StatusUnauthorized, "TOKEN_INVALID", "authentication token is invalid")
			return
		}

		// Manual expiry check — returns a distinct code so the frontend knows to call /auth/refresh-token.
		expTime, err := claims.GetExpirationTime()
		if err != nil || expTime == nil {
			writeAuthJSON(w, http.StatusUnauthorized, "TOKEN_INVALID", "authentication token is invalid")
			return
		}
		if expTime.Time.Before(time.Now()) {
			writeAuthJSON(w, http.StatusUnauthorized, "TOKEN_EXPIRED", "authentication token has expired")
			return
		}

		subStr, _ := claims["sub"].(string)
		userID64, err := strconv.ParseUint(subStr, 10, 64)
		if err != nil {
			writeAuthJSON(w, http.StatusUnauthorized, "TOKEN_INVALID", "authentication token is invalid")
			return
		}

		user, err := m.userRepo.FindByID(r.Context(), uint(userID64))
		if err != nil {
			writeAuthJSON(w, http.StatusUnauthorized, "TOKEN_INVALID", "authentication token is invalid")
			return
		}

		jti, _ := claims["jti"].(string)
		if user.AccessTokenJTI == nil || *user.AccessTokenJTI != jti {
			writeAuthJSON(w, http.StatusUnauthorized, "TOKEN_INVALID", "authentication token is invalid")
			return
		}

		if user.LockedAt != nil {
			writeAuthJSON(w, http.StatusLocked, "ACCOUNT_LOCKED", "account is locked")
			return
		}

		next.ServeHTTP(w, r.WithContext(WithAuthUser(r.Context(), AuthUser{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Role:  string(user.Role),
		})))
	})
}

// RequireRole returns a middleware that allows only the listed roles through.
// Must always be chained after RequireAuth — it reads AuthUser from context.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	permitted := make(map[string]bool, len(roles))
	for _, r := range roles {
		permitted[r] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := AuthUserFromContext(r.Context())
			if !ok || !permitted[user.Role] {
				writeAuthJSON(w, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// AuthUserFromContext retrieves the AuthUser set by RequireAuth.
// Returns false if RequireAuth was not in the chain for this request.
func AuthUserFromContext(ctx context.Context) (AuthUser, bool) {
	user, ok := ctx.Value(authContextKey{}).(AuthUser)
	return user, ok
}

// WithAuthUser stores an AuthUser in ctx. Used internally by RequireAuth and
// exported so tests and integration helpers can inject a user without running the full chain.
func WithAuthUser(ctx context.Context, user AuthUser) context.Context {
	return context.WithValue(ctx, authContextKey{}, user)
}

// --- Private helpers ---

func writeAuthJSON(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{
		"code":  code,
		"error": map[string]string{"message": message},
	})
}
