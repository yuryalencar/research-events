package service_test

// Spec: specs/backend/auth-login.yaml
// Spec: specs/backend/auth-refresh-token.yaml
// Spec: specs/backend/admin-users-unlock.yaml
//
// All functions in internal/service/ are pure — same input always produces same output,
// no I/O, no global state. Tests here need no mocks and no database: just input → output.

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/yuryalencar/research-events/internal/model"
	"github.com/yuryalencar/research-events/internal/service"
)

// --- ValidateLoginInput ---

func TestValidateLoginInput_MissingEmail_ReturnsError(t *testing.T) {
	// Spec: auth-login.yaml border_case "Empty email field → 400 VALIDATION_ERROR"
	err := service.ValidateLoginInput("", "password123")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "email")
}

func TestValidateLoginInput_MissingPassword_ReturnsError(t *testing.T) {
	// Spec: auth-login.yaml border_case "Empty password field → 400 VALIDATION_ERROR"
	err := service.ValidateLoginInput("user@example.com", "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "password")
}

func TestValidateLoginInput_ValidInput_ReturnsNil(t *testing.T) {
	err := service.ValidateLoginInput("user@example.com", "secret")

	require.NoError(t, err)
}

// --- ValidateCredentials ---

func TestValidateCredentials_MatchingPassword_ReturnsNil(t *testing.T) {
	// Spec: auth-login.yaml rule "Passwords hashed with bcrypt cost 12"
	hash, err := bcrypt.GenerateFromPassword([]byte("correct"), 12)
	require.NoError(t, err)
	stored := string(hash)

	err = service.ValidateCredentials(&stored, "correct")

	require.NoError(t, err)
}

func TestValidateCredentials_WrongPassword_ReturnsError(t *testing.T) {
	// Spec: auth-login.yaml — wrong password → 401 INVALID_CREDENTIALS
	hash, err := bcrypt.GenerateFromPassword([]byte("correct"), 12)
	require.NoError(t, err)
	stored := string(hash)

	err = service.ValidateCredentials(&stored, "wrong")

	require.Error(t, err)
}

func TestValidateCredentials_NilHash_ReturnsError(t *testing.T) {
	// Spec: auth-login.yaml — contributors have nil PasswordHash and can never authenticate
	err := service.ValidateCredentials(nil, "anypassword")

	require.Error(t, err)
}

// --- BuildAccessToken ---

func TestBuildAccessToken_ContainsExpectedClaims(t *testing.T) {
	// Spec: auth-middleware.yaml — JWT contains jti, sub, role, name, email, exp
	user := model.User{
		Name:  "Alice",
		Email: "alice@example.com",
		Role:  model.UserRoleAdmin,
	}
	user.ID = 42
	jti := "test-jti-uuid"
	secret := "test-secret"
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	tokenStr, err := service.BuildAccessToken(user, secret, jti, now)
	require.NoError(t, err)

	// Parse and verify the token to inspect claims.
	// WithTimeFunc overrides the clock the parser uses for expiry checks — without it,
	// a fixed past `now` in tests would cause jwt.Parse to reject the token as expired.
	parsed, err := jwt.NewParser(jwt.WithTimeFunc(func() time.Time { return now })).
		Parse(tokenStr, func(t *jwt.Token) (any, error) {
			return []byte(secret), nil
		})
	require.NoError(t, err)
	require.True(t, parsed.Valid)

	claims, ok := parsed.Claims.(jwt.MapClaims)
	require.True(t, ok)
	assert.Equal(t, jti, claims["jti"])
	assert.Equal(t, "42", claims["sub"])
	assert.Equal(t, string(model.UserRoleAdmin), claims["role"])
	assert.Equal(t, "Alice", claims["name"])
	assert.Equal(t, "alice@example.com", claims["email"])
	// exp should be now + 30 minutes
	exp := time.Unix(int64(claims["exp"].(float64)), 0).UTC()
	assert.Equal(t, now.Add(30*time.Minute), exp)
}

func TestBuildAccessToken_IsDeterministic(t *testing.T) {
	// FP: pure function — same inputs must always produce the same token.
	user := model.User{Name: "Bob", Email: "bob@example.com", Role: model.UserRoleModerator}
	user.ID = 7
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	token1, err1 := service.BuildAccessToken(user, "secret", "jti-abc", now)
	token2, err2 := service.BuildAccessToken(user, "secret", "jti-abc", now)

	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.Equal(t, token1, token2)
}

// --- HashRefreshToken / VerifyRefreshToken ---

func TestHashRefreshToken_IsDeterministic(t *testing.T) {
	// FP: pure function — SHA-256 of the same input must always produce the same hash.
	hash1 := service.HashRefreshToken("my-random-token")
	hash2 := service.HashRefreshToken("my-random-token")

	assert.Equal(t, hash1, hash2)
	assert.Len(t, hash1, 64) // SHA-256 hex is always 64 characters
}

func TestHashRefreshToken_DifferentInputsDifferentHashes(t *testing.T) {
	hash1 := service.HashRefreshToken("token-a")
	hash2 := service.HashRefreshToken("token-b")

	assert.NotEqual(t, hash1, hash2)
}

func TestVerifyRefreshToken_ValidPlain_ReturnsTrue(t *testing.T) {
	// Spec: auth-refresh-token.yaml — SHA-256 of cookie value must match stored hash
	plain := "my-secret-refresh-token"
	hash := service.HashRefreshToken(plain)

	assert.True(t, service.VerifyRefreshToken(plain, hash))
}

func TestVerifyRefreshToken_WrongPlain_ReturnsFalse(t *testing.T) {
	// Spec: auth-refresh-token.yaml — tampered token → 401 REFRESH_TOKEN_INVALID
	hash := service.HashRefreshToken("original-token")

	assert.False(t, service.VerifyRefreshToken("tampered-token", hash))
}

// --- BuildLoginResponse ---

func TestBuildLoginResponse_ContainsExpectedFields(t *testing.T) {
	// Spec: auth-login.yaml response 200 — body contains token, role, user{id,name,email}
	user := model.User{Name: "Carol", Email: "carol@example.com", Role: model.UserRoleAdmin}
	user.ID = 99

	resp := service.BuildLoginResponse(user, "jwt-token-string")

	assert.Equal(t, "jwt-token-string", resp.Token)
	assert.Equal(t, string(model.UserRoleAdmin), resp.Role)
	assert.Equal(t, uint(99), resp.User.ID)
	assert.Equal(t, "Carol", resp.User.Name)
	assert.Equal(t, "carol@example.com", resp.User.Email)
}

// --- BuildUnlockAuditLog ---

func TestBuildUnlockAuditLog_SetsCorrectFields(t *testing.T) {
	// Spec: admin-users-unlock.yaml rule "Write AuditLog: entity_type=user, action=unlocked"
	entry := service.BuildUnlockAuditLog(42, 7)

	assert.Equal(t, model.AuditEntityUser, entry.EntityType)
	assert.Equal(t, uint(42), entry.EntityID)
	assert.Equal(t, model.AuditActionUnlocked, entry.Action)
	assert.Equal(t, uint(7), entry.ChangedByID)
}
