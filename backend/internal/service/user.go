package service

import (
	"errors"
	"unicode"

	"golang.org/x/crypto/bcrypt"
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
