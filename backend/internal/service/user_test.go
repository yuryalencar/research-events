package service_test

// Spec: specs/backend/users-update-password.yaml
//
// All functions under test are pure — same input always produces the same output,
// no I/O, no global state. Tests need no mocks and no database: just input → output.

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/yuryalencar/research-events/internal/service"
)

// --- ValidateUpdatePasswordInput ---
// Spec: 400 VALIDATION_ERROR when any required field is missing or empty
// Spec: 400 VALIDATION_ERROR when new_password and confirmation do not match

func TestValidateUpdatePasswordInput(t *testing.T) {
	t.Run("returns error when current_password is empty", func(t *testing.T) {
		err := service.ValidateUpdatePasswordInput("", "NewPass@1", "NewPass@1")

		assert.Error(t, err)
	})

	t.Run("returns error when new_password is empty", func(t *testing.T) {
		err := service.ValidateUpdatePasswordInput("OldPass@1", "", "")

		assert.Error(t, err)
	})

	t.Run("returns error when new_password_confirmation is empty", func(t *testing.T) {
		err := service.ValidateUpdatePasswordInput("OldPass@1", "NewPass@1", "")

		assert.Error(t, err)
	})

	t.Run("returns error when new_password and confirmation do not match", func(t *testing.T) {
		err := service.ValidateUpdatePasswordInput("OldPass@1", "NewPass@1", "Different@2")

		assert.Error(t, err)
	})

	t.Run("returns nil when all fields are present and passwords match", func(t *testing.T) {
		err := service.ValidateUpdatePasswordInput("OldPass@1", "NewPass@1", "NewPass@1")

		assert.NoError(t, err)
	})
}

// --- ValidatePasswordComplexity ---
// Spec: new_password must be ≥ 8 chars, ≥1 uppercase, ≥1 lowercase, ≥1 special char

func TestValidatePasswordComplexity(t *testing.T) {
	t.Run("returns error when password is shorter than 8 characters", func(t *testing.T) {
		err := service.ValidatePasswordComplexity("Ab@1")

		assert.Error(t, err)
	})

	t.Run("returns error when password has no uppercase letter", func(t *testing.T) {
		err := service.ValidatePasswordComplexity("newpass@1")

		assert.Error(t, err)
	})

	t.Run("returns error when password has no lowercase letter", func(t *testing.T) {
		err := service.ValidatePasswordComplexity("NEWPASS@1")

		assert.Error(t, err)
	})

	t.Run("returns error when password has no special character", func(t *testing.T) {
		err := service.ValidatePasswordComplexity("NewPass12")

		assert.Error(t, err)
	})

	t.Run("returns nil for a password meeting all requirements", func(t *testing.T) {
		err := service.ValidatePasswordComplexity("NewPass@1")

		assert.NoError(t, err)
	})
}

// --- HashPassword ---
// Spec: new_password hashed with bcrypt cost 12 before storing

func TestHashPassword(t *testing.T) {
	t.Run("returns a non-empty bcrypt hash", func(t *testing.T) {
		hash, err := service.HashPassword("NewPass@1")

		require.NoError(t, err)
		assert.NotEmpty(t, hash)
	})

	t.Run("produced hash is verifiable with bcrypt", func(t *testing.T) {
		plain := "NewPass@1"
		hash, err := service.HashPassword(plain)
		require.NoError(t, err)

		err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
		assert.NoError(t, err)
	})

	t.Run("same input produces different hashes each call due to bcrypt salting", func(t *testing.T) {
		hash1, err := service.HashPassword("NewPass@1")
		require.NoError(t, err)

		hash2, err := service.HashPassword("NewPass@1")
		require.NoError(t, err)

		// bcrypt embeds a random salt so each call produces a distinct hash
		// even for identical input — the same password must never produce the same ciphertext.
		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("hash is produced at bcrypt cost 12", func(t *testing.T) {
		// Spec: users-update-password.yaml rule "new_password hashed with bcrypt cost 12"
		hash, err := service.HashPassword("NewPass@1")
		require.NoError(t, err)

		cost, err := bcrypt.Cost([]byte(hash))
		require.NoError(t, err)
		assert.Equal(t, 12, cost)
	})
}
