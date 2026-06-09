package handler_test

// Spec: specs/backend/auth-login.yaml
// Spec: specs/backend/auth-refresh-token.yaml
// Spec: specs/backend/auth-logout.yaml

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/yuryalencar/research-events/internal/handler"
	"github.com/yuryalencar/research-events/internal/model"
	"github.com/yuryalencar/research-events/internal/repository"
	"github.com/yuryalencar/research-events/internal/repository/mocks"
	"github.com/yuryalencar/research-events/internal/service"
)

const authTestSecret = "auth-handler-test-secret"

var testLogger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

// hashPw creates a bcrypt hash using MinCost (4) to keep tests fast.
// Cost 12 is used in production — see specs/backend/auth-login.yaml rule.
func hashPw(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	require.NoError(t, err)
	return string(hash)
}

// loginReq builds a POST /api/v1/auth/login request with a JSON body.
func loginReq(t *testing.T, email, password string) *http.Request {
	t.Helper()
	body, err := json.Marshal(map[string]string{"email": email, "password": password})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// responseCode decodes the JSON body and returns the top-level "code" field.
func responseCode(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var body map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	return body["code"].(string)
}

// findCookie returns the named Set-Cookie value from a recorder, or "".
func findCookie(rec *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, c := range rec.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// validAdminUser returns an admin model.User with a bcrypt password hash.
func validAdminUser(t *testing.T) model.User {
	t.Helper()
	hash := hashPw(t, "secret")
	return model.User{
		Model:        gorm.Model{ID: 1},
		Name:         "Admin User",
		Email:        "admin@example.com",
		PasswordHash: &hash,
		Role:         model.UserRoleAdmin,
	}
}

// ===== Login =====

func TestAuthHandler_Login_Success_Returns200WithCookiesAndBody(t *testing.T) {
	// Spec: auth-login.yaml — 200 + access_token cookie + refresh_token cookie + body on valid credentials
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUserRepository(ctrl)

	user := validAdminUser(t)
	mockRepo.EXPECT().FindByEmail(gomock.Any(), "admin@example.com").Return(user, nil)
	mockRepo.EXPECT().UpdateTokens(gomock.Any(), user.ID, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mockRepo.EXPECT().ResetFailedAttempts(gomock.Any(), user.ID).Return(nil)

	h := handler.NewAuthHandler(mockRepo, authTestSecret, testLogger)
	rec := httptest.NewRecorder()
	h.Login(rec, loginReq(t, "admin@example.com", "secret"))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "LOGIN_SUCCESS", responseCode(t, rec))
	assert.NotNil(t, findCookie(rec, "access_token"), "access_token cookie must be set")
	assert.NotNil(t, findCookie(rec, "refresh_token"), "refresh_token cookie must be set")
}

func TestAuthHandler_Login_MissingEmail_Returns400(t *testing.T) {
	// Spec: auth-login.yaml border_case "Empty email field → 400 VALIDATION_ERROR"
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUserRepository(ctrl)

	h := handler.NewAuthHandler(mockRepo, authTestSecret, testLogger)
	rec := httptest.NewRecorder()
	h.Login(rec, loginReq(t, "", "secret"))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestAuthHandler_Login_MissingPassword_Returns400(t *testing.T) {
	// Spec: auth-login.yaml border_case "Empty password field → 400 VALIDATION_ERROR"
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUserRepository(ctrl)

	h := handler.NewAuthHandler(mockRepo, authTestSecret, testLogger)
	rec := httptest.NewRecorder()
	h.Login(rec, loginReq(t, "admin@example.com", ""))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestAuthHandler_Login_EmailNotFound_Returns401(t *testing.T) {
	// Spec: auth-login.yaml — non-existent email → 401 INVALID_CREDENTIALS (never reveals existence)
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockRepo.EXPECT().FindByEmail(gomock.Any(), "ghost@example.com").Return(model.User{}, repository.ErrNotFound)

	h := handler.NewAuthHandler(mockRepo, authTestSecret, testLogger)
	rec := httptest.NewRecorder()
	h.Login(rec, loginReq(t, "ghost@example.com", "secret"))

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(t, "INVALID_CREDENTIALS", responseCode(t, rec))
}

func TestAuthHandler_Login_WrongPassword_Returns401AndIncrementsCounter(t *testing.T) {
	// Spec: auth-login.yaml — wrong password → 401 + failed_login_attempts incremented
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUserRepository(ctrl)

	user := validAdminUser(t)
	mockRepo.EXPECT().FindByEmail(gomock.Any(), "admin@example.com").Return(user, nil)
	mockRepo.EXPECT().IncrementFailedAttempts(gomock.Any(), user.ID).Return(nil)

	h := handler.NewAuthHandler(mockRepo, authTestSecret, testLogger)
	rec := httptest.NewRecorder()
	h.Login(rec, loginReq(t, "admin@example.com", "wrongpassword"))

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(t, "INVALID_CREDENTIALS", responseCode(t, rec))
}

func TestAuthHandler_Login_ContributorRole_Returns403(t *testing.T) {
	// Spec: auth-login.yaml — contributor role → 403 FORBIDDEN (no failed_login_attempts increment)
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUserRepository(ctrl)

	hash := hashPw(t, "secret")
	contributor := model.User{
		Model:        gorm.Model{ID: 5},
		Name:         "Contributor",
		Email:        "contrib@example.com",
		PasswordHash: &hash,
		Role:         model.UserRoleContributor,
	}
	mockRepo.EXPECT().FindByEmail(gomock.Any(), "contrib@example.com").Return(contributor, nil)
	// IncrementFailedAttempts must NOT be called for contributor role check

	h := handler.NewAuthHandler(mockRepo, authTestSecret, testLogger)
	rec := httptest.NewRecorder()
	h.Login(rec, loginReq(t, "contrib@example.com", "secret"))

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Equal(t, "FORBIDDEN", responseCode(t, rec))
}

func TestAuthHandler_Login_AccountLocked_Returns423(t *testing.T) {
	// Spec: auth-login.yaml — locked account → 423 ACCOUNT_LOCKED
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUserRepository(ctrl)

	lockedAt := time.Now().Add(-time.Hour)
	hash := hashPw(t, "secret")
	locked := model.User{
		Model:        gorm.Model{ID: 2},
		Email:        "locked@example.com",
		PasswordHash: &hash,
		Role:         model.UserRoleAdmin,
		LockedAt:     &lockedAt,
	}
	mockRepo.EXPECT().FindByEmail(gomock.Any(), "locked@example.com").Return(locked, nil)

	h := handler.NewAuthHandler(mockRepo, authTestSecret, testLogger)
	rec := httptest.NewRecorder()
	h.Login(rec, loginReq(t, "locked@example.com", "secret"))

	assert.Equal(t, http.StatusLocked, rec.Code)
	assert.Equal(t, "ACCOUNT_LOCKED", responseCode(t, rec))
}

func TestAuthHandler_Login_FifthFailedAttemptLocksAccount(t *testing.T) {
	// Spec: auth-login.yaml — 5th failed attempt sets locked_at; return 401 for this attempt
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUserRepository(ctrl)

	hash := hashPw(t, "correct")
	user := model.User{
		Model:               gorm.Model{ID: 3},
		Email:               "admin5@example.com",
		PasswordHash:        &hash,
		Role:                model.UserRoleAdmin,
		FailedLoginAttempts: 4, // one more wrong attempt → lock
	}
	mockRepo.EXPECT().FindByEmail(gomock.Any(), "admin5@example.com").Return(user, nil)
	mockRepo.EXPECT().IncrementFailedAttempts(gomock.Any(), user.ID).Return(nil)
	mockRepo.EXPECT().LockAccount(gomock.Any(), user.ID).Return(nil)

	h := handler.NewAuthHandler(mockRepo, authTestSecret, testLogger)
	rec := httptest.NewRecorder()
	h.Login(rec, loginReq(t, "admin5@example.com", "wrongpassword"))

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(t, "INVALID_CREDENTIALS", responseCode(t, rec))
}

// ===== RefreshToken =====

func TestAuthHandler_RefreshToken_Success_Returns200WithNewCookies(t *testing.T) {
	// Spec: auth-refresh-token.yaml — valid refresh token → 200 + rotated cookies
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUserRepository(ctrl)

	userID := uint(1)
	plain := fmt.Sprintf("%d.%s", userID, "aabbccdd1122334455667788990011223344556677889900aabbccddeeff0011")
	hash := service.HashRefreshToken(plain)
	exp := time.Now().Add(4 * time.Hour)
	jti := "stored-jti"
	hashCopy := hash

	mockRepo.EXPECT().FindByID(gomock.Any(), userID).Return(model.User{
		Model:                 gorm.Model{ID: userID},
		Name:                  "Admin",
		Email:                 "admin@example.com",
		Role:                  model.UserRoleAdmin,
		RefreshTokenHash:      &hashCopy,
		RefreshTokenExpiresAt: &exp,
		AccessTokenJTI:        &jti,
	}, nil)
	mockRepo.EXPECT().UpdateTokens(gomock.Any(), userID, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	h := handler.NewAuthHandler(mockRepo, authTestSecret, testLogger)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh-token", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: plain})
	rec := httptest.NewRecorder()
	h.RefreshToken(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "REFRESH_SUCCESS", responseCode(t, rec))
	assert.NotNil(t, findCookie(rec, "access_token"))
	assert.NotNil(t, findCookie(rec, "refresh_token"))
}

func TestAuthHandler_RefreshToken_MissingCookie_Returns401(t *testing.T) {
	// Spec: auth-refresh-token.yaml — missing cookie → 401 REFRESH_TOKEN_MISSING
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUserRepository(ctrl)

	h := handler.NewAuthHandler(mockRepo, authTestSecret, testLogger)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh-token", nil)
	rec := httptest.NewRecorder()
	h.RefreshToken(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(t, "REFRESH_TOKEN_MISSING", responseCode(t, rec))
}

func TestAuthHandler_RefreshToken_InvalidFormat_Returns401(t *testing.T) {
	// Spec: auth-refresh-token.yaml — tampered/garbage value → 401 REFRESH_TOKEN_INVALID
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUserRepository(ctrl)

	h := handler.NewAuthHandler(mockRepo, authTestSecret, testLogger)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh-token", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "notavalidtoken"})
	rec := httptest.NewRecorder()
	h.RefreshToken(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(t, "REFRESH_TOKEN_INVALID", responseCode(t, rec))
}

func TestAuthHandler_RefreshToken_HashMismatch_Returns401ReuseAndClearsTokens(t *testing.T) {
	// Spec: auth-refresh-token.yaml — hash mismatch (old token reused) → 401 REFRESH_TOKEN_REUSE + clear DB
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUserRepository(ctrl)

	userID := uint(7)
	plain := fmt.Sprintf("%d.%s", userID, "aabbccdd1122334455667788990011223344556677889900aabbccddeeff0011")
	differentHash := service.HashRefreshToken("different-token")

	mockRepo.EXPECT().FindByID(gomock.Any(), userID).Return(model.User{
		Model:            gorm.Model{ID: userID},
		Role:             model.UserRoleAdmin,
		RefreshTokenHash: &differentHash,
	}, nil)
	mockRepo.EXPECT().ClearTokens(gomock.Any(), userID).Return(nil)

	h := handler.NewAuthHandler(mockRepo, authTestSecret, testLogger)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh-token", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: plain})
	rec := httptest.NewRecorder()
	h.RefreshToken(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(t, "REFRESH_TOKEN_REUSE", responseCode(t, rec))
}

func TestAuthHandler_RefreshToken_LockedAccount_Returns423(t *testing.T) {
	// Spec: auth-refresh-token.yaml — locked account → 423 ACCOUNT_LOCKED
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUserRepository(ctrl)

	userID := uint(9)
	plain := fmt.Sprintf("%d.%s", userID, "aabbccdd1122334455667788990011223344556677889900aabbccddeeff0011")
	hash := service.HashRefreshToken(plain)
	exp := time.Now().Add(4 * time.Hour)
	lockedAt := time.Now().Add(-time.Hour)

	mockRepo.EXPECT().FindByID(gomock.Any(), userID).Return(model.User{
		Model:                 gorm.Model{ID: userID},
		Role:                  model.UserRoleAdmin,
		RefreshTokenHash:      &hash,
		RefreshTokenExpiresAt: &exp,
		LockedAt:              &lockedAt,
	}, nil)

	h := handler.NewAuthHandler(mockRepo, authTestSecret, testLogger)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh-token", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: plain})
	rec := httptest.NewRecorder()
	h.RefreshToken(rec, req)

	assert.Equal(t, http.StatusLocked, rec.Code)
	assert.Equal(t, "ACCOUNT_LOCKED", responseCode(t, rec))
}

// ===== Logout =====

func TestAuthHandler_Logout_Success_Returns200AndClearsCookies(t *testing.T) {
	// Spec: auth-logout.yaml — valid token → 200 + both cookies cleared (MaxAge=0)
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockRepo.EXPECT().ClearTokens(gomock.Any(), gomock.Any()).Return(nil)

	h := handler.NewAuthHandler(mockRepo, authTestSecret, testLogger)

	// Build a valid JWT signed with authTestSecret.
	hash := hashPw(t, "secret")
	jti := "logout-jti"
	user := model.User{Model: gorm.Model{ID: 1}, Role: model.UserRoleAdmin, PasswordHash: &hash}
	tokenStr, err := service.BuildAccessToken(user, authTestSecret, jti, time.Now())
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: tokenStr})
	rec := httptest.NewRecorder()
	h.Logout(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "LOGOUT_SUCCESS", responseCode(t, rec))

	accessCookie := findCookie(rec, "access_token")
	refreshCookie := findCookie(rec, "refresh_token")
	require.NotNil(t, accessCookie)
	require.NotNil(t, refreshCookie)
	assert.Equal(t, -1, accessCookie.MaxAge, "access_token cookie must be expired")
	assert.Equal(t, -1, refreshCookie.MaxAge, "refresh_token cookie must be expired")
}

func TestAuthHandler_Logout_ExpiredTokenAllowed_Returns200(t *testing.T) {
	// Spec: auth-logout.yaml — expired token (valid signature) → 200 LOGOUT_SUCCESS (graceful)
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockRepo.EXPECT().ClearTokens(gomock.Any(), gomock.Any()).Return(nil)

	h := handler.NewAuthHandler(mockRepo, authTestSecret, testLogger)

	hash := hashPw(t, "secret")
	user := model.User{Model: gorm.Model{ID: 1}, Role: model.UserRoleAdmin, PasswordHash: &hash}
	// Build a token with exp in the past.
	tokenStr, err := service.BuildAccessToken(user, authTestSecret, "expired-jti", time.Now().Add(-2*time.Hour))
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: tokenStr})
	rec := httptest.NewRecorder()
	h.Logout(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "LOGOUT_SUCCESS", responseCode(t, rec))
}

func TestAuthHandler_Logout_MissingCookie_Returns401(t *testing.T) {
	// Spec: auth-logout.yaml — no cookie → 401 TOKEN_MISSING
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUserRepository(ctrl)

	h := handler.NewAuthHandler(mockRepo, authTestSecret, testLogger)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	rec := httptest.NewRecorder()
	h.Logout(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(t, "TOKEN_MISSING", responseCode(t, rec))
}

func TestAuthHandler_Logout_TamperedToken_Returns401(t *testing.T) {
	// Spec: auth-logout.yaml — tampered token → 401 TOKEN_INVALID
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUserRepository(ctrl)

	h := handler.NewAuthHandler(mockRepo, authTestSecret, testLogger)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "header.payload.badsignature"})
	rec := httptest.NewRecorder()
	h.Logout(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(t, "TOKEN_INVALID", responseCode(t, rec))
}

func TestAuthHandler_Logout_Idempotent_Returns200Twice(t *testing.T) {
	// Spec: auth-logout.yaml — already logged out → 200 (idempotent)
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockRepo.EXPECT().ClearTokens(gomock.Any(), gomock.Any()).Return(nil).Times(2)

	h := handler.NewAuthHandler(mockRepo, authTestSecret, testLogger)

	hash := hashPw(t, "secret")
	user := model.User{Model: gorm.Model{ID: 1}, Role: model.UserRoleAdmin, PasswordHash: &hash}
	tokenStr, err := service.BuildAccessToken(user, authTestSecret, "idem-jti", time.Now())
	require.NoError(t, err)

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
		req.AddCookie(&http.Cookie{Name: "access_token", Value: tokenStr})
		rec := httptest.NewRecorder()
		h.Logout(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, "call %d should return 200", i+1)
	}
}
