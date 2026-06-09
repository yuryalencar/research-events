package middleware_test

// Spec: specs/backend/auth-middleware.yaml
// Rule: "RequireAuth validates access token and attaches AuthUser to context"
// Rule: "RequireRole checks role from context — must be chained after RequireAuth"
// Rule: "TOKEN_EXPIRED is a distinct code so frontend knows to attempt /auth/refresh-token"

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/yuryalencar/research-events/internal/middleware"
	"github.com/yuryalencar/research-events/internal/model"
	"github.com/yuryalencar/research-events/internal/repository/mocks"
)

const testJWTSecret = "test-jwt-secret-for-middleware"

// buildToken creates a signed JWT for use in middleware tests.
func buildToken(t *testing.T, secret, jti string, userID uint, role model.UserRole, exp time.Time) string {
	t.Helper()
	claims := jwt.MapClaims{
		"jti":   jti,
		"sub":   strconv.FormatUint(uint64(userID), 10),
		"role":  string(role),
		"name":  "Test User",
		"email": "testuser@example.com",
		"iat":   time.Now().Unix(),
		"exp":   exp.Unix(),
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	require.NoError(t, err)
	return tok
}

// captureUser returns a handler that captures the AuthUser injected by RequireAuth.
func captureUser(captured *middleware.AuthUser, ok *bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*captured, *ok = middleware.AuthUserFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})
}

// assertCode decodes the JSON body and asserts the top-level "code" field value.
func assertCode(t *testing.T, rec *httptest.ResponseRecorder, expected string) {
	t.Helper()
	var body map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.Equal(t, expected, body["code"])
}

// --- RequireAuth ---

func TestRequireAuth_ValidToken_AttachesAuthUserToContext(t *testing.T) {
	// Spec: auth-middleware.yaml — valid token → AuthUser attached, request proceeds
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUserRepository(ctrl)

	jti := "valid-jti-uuid"
	userID := uint(1)
	tok := buildToken(t, testJWTSecret, jti, userID, model.UserRoleAdmin, time.Now().Add(30*time.Minute))

	mockRepo.EXPECT().FindByID(gomock.Any(), userID).Return(model.User{
		Model:          gorm.Model{ID: userID},
		Name:           "Test User",
		Email:          "testuser@example.com",
		Role:           model.UserRoleAdmin,
		AccessTokenJTI: &jti,
	}, nil)

	var captured middleware.AuthUser
	var userOK bool

	auth := middleware.NewAuthMiddleware(testJWTSecret, mockRepo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: tok})
	rec := httptest.NewRecorder()

	auth.RequireAuth(captureUser(&captured, &userOK)).ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.True(t, userOK, "AuthUser must be present in context")
	assert.Equal(t, userID, captured.ID)
	assert.Equal(t, "admin", captured.Role)
	assert.Equal(t, "testuser@example.com", captured.Email)
}

func TestRequireAuth_MissingCookie_Returns401TokenMissing(t *testing.T) {
	// Spec: auth-middleware.yaml — missing cookie → 401 TOKEN_MISSING
	// No DB call expected: middleware returns before touching the repository.
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUserRepository(ctrl)

	auth := middleware.NewAuthMiddleware(testJWTSecret, mockRepo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events", nil)
	rec := httptest.NewRecorder()

	auth.RequireAuth(okHandler()).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assertCode(t, rec, "TOKEN_MISSING")
}

func TestRequireAuth_ExpiredToken_Returns401TokenExpired(t *testing.T) {
	// Spec: auth-middleware.yaml — expired → 401 TOKEN_EXPIRED (distinct from INVALID)
	// Expiry is checked before the DB lookup — no FindByID call expected.
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUserRepository(ctrl)

	tok := buildToken(t, testJWTSecret, "jti-1", 1, model.UserRoleAdmin, time.Now().Add(-time.Hour))

	auth := middleware.NewAuthMiddleware(testJWTSecret, mockRepo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: tok})
	rec := httptest.NewRecorder()

	auth.RequireAuth(okHandler()).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assertCode(t, rec, "TOKEN_EXPIRED")
}

func TestRequireAuth_InvalidSignature_Returns401TokenInvalid(t *testing.T) {
	// Spec: auth-middleware.yaml — invalid signature → 401 TOKEN_INVALID
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUserRepository(ctrl)

	tok := buildToken(t, "wrong-secret", "jti-1", 1, model.UserRoleAdmin, time.Now().Add(time.Hour))

	auth := middleware.NewAuthMiddleware(testJWTSecret, mockRepo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: tok})
	rec := httptest.NewRecorder()

	auth.RequireAuth(okHandler()).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assertCode(t, rec, "TOKEN_INVALID")
}

func TestRequireAuth_JTIMismatch_Returns401TokenInvalid(t *testing.T) {
	// Spec: auth-middleware.yaml — JTI mismatch (revoked or rotated) → 401 TOKEN_INVALID
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUserRepository(ctrl)

	jtiInToken := "jti-in-token"
	jtiInDB := "different-jti-in-db"
	userID := uint(5)
	tok := buildToken(t, testJWTSecret, jtiInToken, userID, model.UserRoleAdmin, time.Now().Add(time.Hour))

	mockRepo.EXPECT().FindByID(gomock.Any(), userID).Return(model.User{
		Role:           model.UserRoleAdmin,
		AccessTokenJTI: &jtiInDB,
	}, nil)

	auth := middleware.NewAuthMiddleware(testJWTSecret, mockRepo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: tok})
	rec := httptest.NewRecorder()

	auth.RequireAuth(okHandler()).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assertCode(t, rec, "TOKEN_INVALID")
}

func TestRequireAuth_LockedAccount_Returns423(t *testing.T) {
	// Spec: auth-middleware.yaml — locked account → 423 ACCOUNT_LOCKED
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUserRepository(ctrl)

	jti := "jti-locked"
	userID := uint(3)
	lockedAt := time.Now().Add(-time.Hour)
	tok := buildToken(t, testJWTSecret, jti, userID, model.UserRoleAdmin, time.Now().Add(time.Hour))

	mockRepo.EXPECT().FindByID(gomock.Any(), userID).Return(model.User{
		Role:           model.UserRoleAdmin,
		AccessTokenJTI: &jti,
		LockedAt:       &lockedAt,
	}, nil)

	auth := middleware.NewAuthMiddleware(testJWTSecret, mockRepo)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: tok})
	rec := httptest.NewRecorder()

	auth.RequireAuth(okHandler()).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusLocked, rec.Code) // 423
	assertCode(t, rec, "ACCOUNT_LOCKED")
}

// --- RequireRole ---

func TestRequireRole_MatchingRole_PassesThrough(t *testing.T) {
	// Spec: auth-middleware.yaml — role in permitted list → request proceeds
	authUser := middleware.AuthUser{ID: 1, Role: "admin", Name: "Admin", Email: "admin@example.com"}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/events", nil)
	req = req.WithContext(middleware.WithAuthUser(req.Context(), authUser))
	rec := httptest.NewRecorder()

	middleware.RequireRole("admin", "moderator")(okHandler()).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireRole_WrongRole_Returns403Forbidden(t *testing.T) {
	// Spec: auth-middleware.yaml — role not in permitted list → 403 FORBIDDEN
	authUser := middleware.AuthUser{ID: 2, Role: "moderator", Name: "Mod", Email: "mod@example.com"}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/events", nil)
	req = req.WithContext(middleware.WithAuthUser(req.Context(), authUser))
	rec := httptest.NewRecorder()

	middleware.RequireRole("admin")(okHandler()).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assertCode(t, rec, "FORBIDDEN")
}
