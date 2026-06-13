package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/yuryalencar/research-events/internal/model"
	"github.com/yuryalencar/research-events/internal/repository"
	"github.com/yuryalencar/research-events/internal/service"
)

// --- Types ---

// AuthHandler handles authentication endpoints: login, refresh-token, and logout.
// Dependencies are injected via constructor — never use globals.
type AuthHandler struct {
	userRepo  repository.UserRepository
	jwtSecret string
	logger    *slog.Logger
}

// tokenResult holds the values produced by issueTokens, ready to be written as
// cookies and included in the response body.
type tokenResult struct {
	accessToken  string
	refreshPlain string
	jtiExp       time.Time
	refreshExp   time.Time
}

// --- Constructor ---

func NewAuthHandler(userRepo repository.UserRepository, jwtSecret string, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		logger:    logger,
	}
}

// --- Public methods ---

// Login handles POST /api/v1/auth/login.
// Public endpoint — no authentication required.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	if err := service.ValidateLoginInput(input.Email, input.Password); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	user, err := h.userRepo.FindByEmail(r.Context(), input.Email)
	if err != nil {
		// Never reveal whether the email exists — always return INVALID_CREDENTIALS.
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid credentials")
		return
	}

	if user.LockedAt != nil {
		writeError(w, http.StatusLocked, "ACCOUNT_LOCKED", "account is locked due to too many failed attempts")
		return
	}

	// Role check before password validation — contributors are blocked without incrementing
	// the failed-attempts counter (the 403 is not an auth failure, it's a permissions error).
	if user.Role == model.UserRoleContributor {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "account does not have permission to authenticate")
		return
	}

	if err := service.ValidateCredentials(user.PasswordHash, input.Password); err != nil {
		// On the 5th consecutive failure, lock the account before incrementing.
		// FailedLoginAttempts reflects the count before this attempt, so +1 is the new total.
		if user.FailedLoginAttempts+1 >= 5 {
			if lockErr := h.userRepo.LockAccount(r.Context(), user.ID); lockErr != nil {
				h.logger.Error("failed to lock account", "user_id", user.ID, "error", lockErr)
			}
		}
		if incErr := h.userRepo.IncrementFailedAttempts(r.Context(), user.ID); incErr != nil {
			h.logger.Error("failed to increment failed attempts", "user_id", user.ID, "error", incErr)
		}
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid credentials")
		return
	}

	tokens, err := h.issueTokens(r.Context(), user)
	if err != nil {
		h.logger.Error("failed to issue tokens on login", "user_id", user.ID, "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	// Counter reset is non-critical — a failure here does not roll back the successful login.
	if err := h.userRepo.ResetFailedAttempts(r.Context(), user.ID); err != nil {
		h.logger.Error("failed to reset failed attempts", "user_id", user.ID, "error", err)
	}

	setAccessTokenCookie(w, tokens.accessToken, tokens.jtiExp)
	setRefreshTokenCookie(w, tokens.refreshPlain, tokens.refreshExp)
	writeSuccess(w, http.StatusOK, "LOGIN_SUCCESS", service.BuildLoginResponse(user, tokens.accessToken))
}

// RefreshToken handles POST /api/v1/auth/refresh-token.
// Rotates both tokens. The refresh token embeds the user ID as a prefix so we can
// identify the user without a full-table hash lookup.
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		writeError(w, http.StatusUnauthorized, "REFRESH_TOKEN_MISSING", "refresh token cookie is missing")
		return
	}

	userID, err := service.ParseRefreshTokenUserID(cookie.Value)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "REFRESH_TOKEN_INVALID", "refresh token is invalid or expired")
		return
	}

	user, err := h.userRepo.FindByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "REFRESH_TOKEN_INVALID", "refresh token is invalid or expired")
		return
	}

	if user.LockedAt != nil {
		writeError(w, http.StatusLocked, "ACCOUNT_LOCKED", "account is locked")
		return
	}

	// Hash mismatch means the token was already rotated — potential reuse/theft.
	// Clear all tokens for this user immediately to force a full re-login.
	if user.RefreshTokenHash == nil || !service.VerifyRefreshToken(cookie.Value, *user.RefreshTokenHash) {
		if clearErr := h.userRepo.ClearTokens(r.Context(), userID); clearErr != nil {
			h.logger.Error("failed to clear tokens on reuse detection", "user_id", userID, "error", clearErr)
		}
		writeError(w, http.StatusUnauthorized, "REFRESH_TOKEN_REUSE", "refresh token has already been used")
		return
	}

	if user.RefreshTokenExpiresAt == nil || time.Now().After(*user.RefreshTokenExpiresAt) {
		writeError(w, http.StatusUnauthorized, "REFRESH_TOKEN_INVALID", "refresh token is invalid or expired")
		return
	}

	tokens, err := h.issueTokens(r.Context(), user)
	if err != nil {
		h.logger.Error("failed to issue tokens on refresh", "user_id", user.ID, "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	setAccessTokenCookie(w, tokens.accessToken, tokens.jtiExp)
	setRefreshTokenCookie(w, tokens.refreshPlain, tokens.refreshExp)
	writeSuccess(w, http.StatusOK, "REFRESH_SUCCESS", service.BuildLoginResponse(user, tokens.accessToken))
}

// Logout handles POST /api/v1/auth/logout.
// Accepts expired tokens so users can always log out gracefully.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("access_token")
	if err != nil {
		writeError(w, http.StatusUnauthorized, "TOKEN_MISSING", "authentication token is missing")
		return
	}

	// Parse without claims validation — an expired token must still be accepted for logout.
	token, err := jwt.Parse(cookie.Value,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			return []byte(h.jwtSecret), nil
		},
		jwt.WithoutClaimsValidation(),
	)
	if err != nil || !token.Valid {
		writeError(w, http.StatusUnauthorized, "TOKEN_INVALID", "authentication token is invalid")
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		writeError(w, http.StatusUnauthorized, "TOKEN_INVALID", "authentication token is invalid")
		return
	}

	subStr, _ := claims["sub"].(string)
	userID64, err := strconv.ParseUint(subStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "TOKEN_INVALID", "authentication token is invalid")
		return
	}

	// ClearTokens is idempotent — clearing already-null fields is safe (no error).
	if err := h.userRepo.ClearTokens(r.Context(), uint(userID64)); err != nil {
		h.logger.Error("failed to clear tokens on logout", "error", err)
	}

	clearAuthCookies(w)
	writeSuccess(w, http.StatusOK, "LOGOUT_SUCCESS", nil)
}

// --- Private methods ---

// issueTokens generates a fresh access + refresh token pair, persists it to the DB,
// and returns the values for the caller to write as cookies and include in the response.
// Extracting this avoids duplicating the identical generation block in Login and RefreshToken.
func (h *AuthHandler) issueTokens(ctx context.Context, user model.User) (tokenResult, error) {
	jti := uuid.New().String()
	randomHex, err := generateRandomHex(32)
	if err != nil {
		return tokenResult{}, err
	}

	now := time.Now()
	accessToken, err := service.BuildAccessToken(user, h.jwtSecret, jti, now)
	if err != nil {
		return tokenResult{}, err
	}

	refreshPlain := service.BuildRefreshTokenPayload(user.ID, randomHex)
	jtiExp := now.Add(30 * time.Minute)
	refreshExp := now.Add(4 * time.Hour)

	if err := h.userRepo.UpdateTokens(ctx, user.ID, jti, jtiExp, service.HashRefreshToken(refreshPlain), refreshExp); err != nil {
		return tokenResult{}, err
	}

	return tokenResult{
		accessToken:  accessToken,
		refreshPlain: refreshPlain,
		jtiExp:       jtiExp,
		refreshExp:   refreshExp,
	}, nil
}

// --- Private helpers ---

func generateRandomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func setAccessTokenCookie(w http.ResponseWriter, token string, exp time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    token,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		Expires:  exp,
		MaxAge:   30 * 60,
	})
}

func setRefreshTokenCookie(w http.ResponseWriter, token string, exp time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/api/v1/auth",
		Expires:  exp,
		MaxAge:   4 * 60 * 60,
	})
}

func clearAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   -1,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/api/v1/auth",
		MaxAge:   -1,
	})
}

func writeSuccess(w http.ResponseWriter, status int, code string, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{"code": code, "data": data})
}

// writeSuccessWithMeta is writeSuccess with an additional top-level "meta" field —
// used by list endpoints to convey pagination info alongside "data".
func writeSuccessWithMeta(w http.ResponseWriter, status int, code string, data any, meta any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{"code": code, "data": data, "meta": meta})
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{
		"code":  code,
		"error": map[string]string{"message": message},
	})
}
