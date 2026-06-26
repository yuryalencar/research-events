package service

import (
	"encoding/json"
	"errors"
	"unicode"

	"golang.org/x/crypto/bcrypt"

	"github.com/yuryalencar/research-events/internal/model"
)

// --- Public functions ---

// FP: pure function
// ValidateUpdatePasswordInput checks that all three fields are non-empty and that
// new_password matches new_password_confirmation. Validation is ordered so that
// "fields missing" is reported before "passwords do not match" — the first error
// found is returned; the caller should fix one issue at a time.
// Pure: depends only on its arguments, performs no I/O, no global state.
func ValidateUpdatePasswordInput(current, newPw, confirm string) error {
	if current == "" || newPw == "" || confirm == "" {
		return errors.New("current_password, new_password and new_password_confirmation are required")
	}
	if newPw != confirm {
		return errors.New("new_password and new_password_confirmation do not match")
	}
	return nil
}

// FP: pure function
// ValidatePasswordComplexity enforces the password strength rules defined in the spec:
// minimum 8 characters, at least one uppercase letter, one lowercase letter, and one
// special (non-alphanumeric) character. All four rules are checked independently so the
// error message always describes the complete requirement, not whichever rule failed first.
// Pure: deterministic given the same input, no I/O, no side effects.
func ValidatePasswordComplexity(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters and contain at least one uppercase letter, one lowercase letter, and one special character")
	}

	var hasUpper, hasLower, hasSpecial bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case !unicode.IsLetter(ch) && !unicode.IsDigit(ch):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasSpecial {
		return errors.New("password must be at least 8 characters and contain at least one uppercase letter, one lowercase letter, and one special character")
	}

	return nil
}

// FP: pure function
// ValidateRegisterInput checks all four registration fields in one pass.
// Combining field-presence and role-validity here keeps the handler thin: one call, one error.
// Contributor is explicitly rejected because contributors self-register via event submission.
func ValidateRegisterInput(name, email, password, role string) error {
	if name == "" || email == "" || password == "" || role == "" {
		return errors.New("name, email, password and role are required")
	}
	if role != string(model.UserRoleAdmin) && role != string(model.UserRoleModerator) {
		return errors.New("role must be one of: admin, moderator")
	}
	return ValidatePasswordComplexity(password)
}

// FP: immutability
// BuildRegisterUser constructs a new User value from validated inputs.
// It never mutates any argument — the caller receives a fresh struct to hand to the repository.
func BuildRegisterUser(name, email, passwordHash string, role model.UserRole) model.User {
	return model.User{
		Name:         name,
		Email:        email,
		PasswordHash: &passwordHash,
		Role:         role,
	}
}

// FP: no side effects
// BuildRegisterAuditLog computes the AuditLog entry for a new user registration.
// The diff records the initial values so the audit trail is self-contained.
// Persistence is the handler's responsibility — this function only builds the value.
func BuildRegisterAuditLog(newUserID, adminID uint, name, email string, role model.UserRole) model.AuditLog {
	diff, _ := json.Marshal(map[string]any{
		"name":  name,
		"email": email,
		"role":  string(role),
	})
	return model.AuditLog{
		EntityType:  model.AuditEntityUser,
		EntityID:    newUserID,
		Action:      model.AuditActionCreated,
		ChangedByID: adminID,
		Diff:        model.JSONB(diff),
	}
}

// FP: pure function
// ValidateRoleChangeInput checks that the requested role is one of the three valid values.
// Unlike ValidateRegisterInput, contributor is allowed here — it is a valid downgrade target.
func ValidateRoleChangeInput(role string) error {
	switch model.UserRole(role) {
	case model.UserRoleAdmin, model.UserRoleModerator, model.UserRoleContributor:
		return nil
	}
	return errors.New("role must be one of: admin, moderator, contributor")
}

// FP: no side effects
// BuildRoleChangedAuditLog computes the AuditLog entry for a role change operation.
// The diff captures the before/after transition so history is fully reconstructible.
func BuildRoleChangedAuditLog(targetID, adminID uint, oldRole, newRole model.UserRole) model.AuditLog {
	diff, _ := json.Marshal(map[string]any{
		"role": map[string]string{
			"before": string(oldRole),
			"after":  string(newRole),
		},
	})
	return model.AuditLog{
		EntityType:  model.AuditEntityUser,
		EntityID:    targetID,
		Action:      model.AuditActionRoleChanged,
		ChangedByID: adminID,
		Diff:        model.JSONB(diff),
	}
}

// FP: no side effects
// HashPassword wraps bcrypt.GenerateFromPassword at cost 12 — the same cost
// used by the login flow (see specs/backend/auth-login.yaml, rules section).
// It returns the hash for the caller to persist; writing to any store is the
// handler's responsibility, keeping this function free of side effects.
func HashPassword(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
