package service

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/yuryalencar/research-events/internal/model"
)

// --- Types ---

// LoginResponse is the value returned to the caller after a successful authentication.
// It is constructed by BuildLoginResponse and written to HTTP by the handler.
type LoginResponse struct {
	Token string   `json:"token"`
	Role  string   `json:"role"`
	User  UserInfo `json:"user"`
}

// UserInfo carries the subset of user fields exposed in auth responses.
type UserInfo struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// --- Public functions ---

// FP: pure function
// ValidateLoginInput depends only on its arguments and performs no I/O.
// The same email+password always returns the same result — no hidden state.
// Pure functions are trivially testable: no setup, no teardown, no mocks.
func ValidateLoginInput(email, password string) error {
	if email == "" {
		return errors.New("email is required")
	}
	if password == "" {
		return errors.New("password is required")
	}
	return nil
}

// FP: pure function
// ValidateCredentials wraps bcrypt comparison — deterministic given the same inputs.
// Returning an error for nil storedHash means passwordless contributors (role=contributor)
// can never authenticate through this path, enforcing the auth spec rule at the data level.
func ValidateCredentials(storedHash *string, plainPassword string) error {
	if storedHash == nil {
		return errors.New("account has no password set")
	}
	return bcrypt.CompareHashAndPassword([]byte(*storedHash), []byte(plainPassword))
}

// FP: pure function
// BuildAccessToken is deterministic: the same user, secret, jti, and now always produce
// the same JWT. Non-deterministic inputs (jti UUID, current time) are injected by the
// caller — this keeps the function pure and testable without time-mocking libraries.
//
// HMAC-SHA256 signing is itself deterministic (unlike RSA which uses random padding),
// so the output is identical on every call with the same inputs.
func BuildAccessToken(user model.User, secret, jti string, now time.Time) (string, error) {
	claims := jwt.MapClaims{
		"jti":   jti,
		"sub":   strconv.FormatUint(uint64(user.ID), 10),
		"role":  string(user.Role),
		"name":  user.Name,
		"email": user.Email,
		"iat":   now.Unix(),
		"exp":   now.Add(30 * time.Minute).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

// FP: pure function
// HashRefreshToken computes a SHA-256 hex digest of the plain refresh token.
// SHA-256 is used instead of bcrypt because refresh tokens are already 32 bytes of
// cryptographically random data (256-bit entropy) — bcrypt's slow KDF adds no security
// benefit here, and its cost would add ~100ms latency to every refresh request.
func HashRefreshToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return fmt.Sprintf("%x", sum)
}

// FP: pure function
// VerifyRefreshToken hashes plain and compares it to the stored hash.
// Delegates to HashRefreshToken to ensure the hashing logic is never duplicated —
// function composition keeps the two operations in sync automatically.
func VerifyRefreshToken(plain, hash string) bool {
	return HashRefreshToken(plain) == hash
}

// FP: pure function
// BuildLoginResponse constructs the HTTP response value from pre-computed inputs.
// It never touches the database or HTTP layer — the caller owns those side effects.
// Separating computation from I/O makes each piece independently testable.
func BuildLoginResponse(user model.User, accessToken string) LoginResponse {
	return LoginResponse{
		Token: accessToken,
		Role:  string(user.Role),
		User: UserInfo{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
	}
}

// FP: pure function
// BuildRefreshTokenPayload formats the refresh token string as "{userID}.{randomHex}".
// Embedding the user ID lets the refresh endpoint identify the user without a full-table
// hash lookup — the handler parses the ID then verifies the full token's hash against the DB.
func BuildRefreshTokenPayload(userID uint, randomHex string) string {
	return fmt.Sprintf("%d.%s", userID, randomHex)
}

// FP: pure function
// ParseRefreshTokenUserID extracts the user ID from a refresh token payload produced by
// BuildRefreshTokenPayload. Returns an error if the token is not in the expected format.
func ParseRefreshTokenUserID(token string) (uint, error) {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return 0, errors.New("invalid refresh token format")
	}
	id, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return 0, errors.New("invalid refresh token format")
	}
	return uint(id), nil
}

// FP: no side effects
// BuildUnlockAuditLog computes the AuditLog value for an account unlock operation.
// It returns a value for the caller to persist — writing to any store is the handler's job,
// not this function's. This separation makes the audit entry trivially testable in isolation.
func BuildUnlockAuditLog(targetID uint, adminID uint) model.AuditLog {
	return model.AuditLog{
		EntityType:  model.AuditEntityUser,
		EntityID:    targetID,
		Action:      model.AuditActionUnlocked,
		ChangedByID: adminID,
	}
}
