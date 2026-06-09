package model

import (
	"time"

	"gorm.io/gorm"
)

// --- Types ---

// UserRole is a named string type — not iota — so values are readable in DB and JSON.
// Using a named type (not plain string) lets the compiler reject misuse like
// passing an arbitrary string where a UserRole is expected.
type UserRole string

const (
	UserRoleAdmin       UserRole = "admin"
	UserRoleModerator   UserRole = "moderator"
	UserRoleContributor UserRole = "contributor"
)

// User represents a platform user across all three roles.
//
// Contributors are created automatically when an event is submitted — they have no
// password (PasswordHash is nil) and cannot log in until they claim their account.
// Only admin and moderator roles can authenticate via /auth/login.
//
// Auth fields (AccessTokenJTI, RefreshTokenHash, etc.) store the current valid token
// pair for stateful JWT validation — one session per user at a time.
// See specs/backend/auth-jti-design.md for the design rationale.
type User struct {
	gorm.Model // embeds ID, CreatedAt, UpdatedAt, DeletedAt (soft delete)

	Name         string   `gorm:"not null"`
	Email        string   `gorm:"uniqueIndex;not null"`
	PasswordHash *string  // nil for passwordless contributors
	Role         UserRole `gorm:"not null;default:contributor"`

	// Stateful JWT — only one valid token pair per user at a time.
	// AccessTokenJTI stores the jti claim of the current valid access token.
	// Comparing the incoming JWT's jti against this value is how we detect revoked tokens.
	AccessTokenJTI        *string    `gorm:"column:access_token_jti"`
	AccessTokenExpiresAt  *time.Time
	RefreshTokenHash      *string    // SHA-256 hex of the current refresh token
	RefreshTokenExpiresAt *time.Time

	// Account lockout — set after 5 consecutive failed login attempts.
	// LockedAt is nil when the account is active; set to the lock time when locked.
	// Only another admin can clear it via PATCH /api/v1/admin/users/:id/unlock.
	FailedLoginAttempts int        `gorm:"not null;default:0"`
	LockedAt            *time.Time
}
