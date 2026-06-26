package service_test

// Spec: specs/backend/users-update-password.yaml
// Spec: specs/backend/admin-users-register.yaml
// Spec: specs/backend/admin-users-change-role.yaml
//
// All functions under test are pure — same input always produces the same output,
// no I/O, no global state. Tests need no mocks and no database: just input → output.

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/yuryalencar/research-events/internal/model"
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

// --- ValidateRegisterInput --- (Cycle 2)
// Spec: admin-users-register.yaml — 400 VALIDATION_ERROR for missing fields, bad role, weak password

func TestValidateRegisterInput_ReturnsNilForValidAdminInput(t *testing.T) {
	err := service.ValidateRegisterInput("Jane", "jane@example.com", "Secret@123", "admin")
	assert.NoError(t, err)
}

func TestValidateRegisterInput_ReturnsNilForValidModeratorInput(t *testing.T) {
	err := service.ValidateRegisterInput("Jane", "jane@example.com", "Secret@123", "moderator")
	assert.NoError(t, err)
}

func TestValidateRegisterInput_ReturnsErrorForEmptyName(t *testing.T) {
	err := service.ValidateRegisterInput("", "jane@example.com", "Secret@123", "admin")
	assert.Error(t, err)
}

func TestValidateRegisterInput_ReturnsErrorForEmptyEmail(t *testing.T) {
	err := service.ValidateRegisterInput("Jane", "", "Secret@123", "admin")
	assert.Error(t, err)
}

func TestValidateRegisterInput_ReturnsErrorForEmptyPassword(t *testing.T) {
	err := service.ValidateRegisterInput("Jane", "jane@example.com", "", "admin")
	assert.Error(t, err)
}

func TestValidateRegisterInput_ReturnsErrorForEmptyRole(t *testing.T) {
	err := service.ValidateRegisterInput("Jane", "jane@example.com", "Secret@123", "")
	assert.Error(t, err)
}

func TestValidateRegisterInput_ReturnsErrorForContributorRole(t *testing.T) {
	// Spec: admin-users-register.yaml border_case "role = contributor → 400 VALIDATION_ERROR"
	err := service.ValidateRegisterInput("Jane", "jane@example.com", "Secret@123", "contributor")
	assert.Error(t, err)
}

func TestValidateRegisterInput_ReturnsErrorForInvalidRole(t *testing.T) {
	err := service.ValidateRegisterInput("Jane", "jane@example.com", "Secret@123", "superuser")
	assert.Error(t, err)
}

func TestValidateRegisterInput_ReturnsErrorForWeakPassword(t *testing.T) {
	// Spec: admin-users-register.yaml — same complexity rules as ValidatePasswordComplexity
	err := service.ValidateRegisterInput("Jane", "jane@example.com", "weakpass", "admin")
	assert.Error(t, err)
}

func TestValidateRegisterInput_IsPure_SameInputReturnsSameOutput(t *testing.T) {
	// FP: pure function — deterministic, no hidden state
	err1 := service.ValidateRegisterInput("Jane", "jane@example.com", "Secret@123", "admin")
	err2 := service.ValidateRegisterInput("Jane", "jane@example.com", "Secret@123", "admin")
	assert.Equal(t, err1, err2)
}

// --- BuildRegisterUser --- (Cycle 3)
// Spec: admin-users-register.yaml DoD "Returns 201 + user data on valid registration"

func TestBuildRegisterUser_SetsAllFieldsCorrectly(t *testing.T) {
	hash := "bcrypt-hash"
	u := service.BuildRegisterUser("Jane", "jane@example.com", hash, model.UserRoleModerator)

	assert.Equal(t, "Jane", u.Name)
	assert.Equal(t, "jane@example.com", u.Email)
	require.NotNil(t, u.PasswordHash)
	assert.Equal(t, hash, *u.PasswordHash)
	assert.Equal(t, model.UserRoleModerator, u.Role)
}

func TestBuildRegisterUser_IsPure_SameInputReturnsSameOutput(t *testing.T) {
	// FP: pure function — same args, same struct back, every time
	hash := "bcrypt-hash"
	u1 := service.BuildRegisterUser("Jane", "jane@example.com", hash, model.UserRoleModerator)
	u2 := service.BuildRegisterUser("Jane", "jane@example.com", hash, model.UserRoleModerator)
	assert.Equal(t, u1, u2)
}

// --- BuildRegisterAuditLog --- (Cycle 4)
// Spec: admin-users-register.yaml rule "Write AuditLog: entity_type=user, action=created, diff={name,email,role}"

func TestBuildRegisterAuditLog_SetsEntityTypeUser(t *testing.T) {
	entry := service.BuildRegisterAuditLog(10, 1, "Jane", "jane@example.com", model.UserRoleModerator)
	assert.Equal(t, model.AuditEntityUser, entry.EntityType)
}

func TestBuildRegisterAuditLog_SetsActionCreated(t *testing.T) {
	entry := service.BuildRegisterAuditLog(10, 1, "Jane", "jane@example.com", model.UserRoleModerator)
	assert.Equal(t, model.AuditActionCreated, entry.Action)
}

func TestBuildRegisterAuditLog_SetsDiffWithNameEmailAndRole(t *testing.T) {
	entry := service.BuildRegisterAuditLog(10, 1, "Jane", "jane@example.com", model.UserRoleModerator)
	diffStr := string(entry.Diff)
	assert.Contains(t, diffStr, "jane@example.com")
	assert.Contains(t, diffStr, "Jane")
	assert.Contains(t, diffStr, "moderator")
}

func TestBuildRegisterAuditLog_IsPure_SameInputReturnsSameOutput(t *testing.T) {
	e1 := service.BuildRegisterAuditLog(10, 1, "Jane", "jane@example.com", model.UserRoleModerator)
	e2 := service.BuildRegisterAuditLog(10, 1, "Jane", "jane@example.com", model.UserRoleModerator)
	assert.Equal(t, e1, e2)
}

// --- ValidateRoleChangeInput --- (Cycle 5)
// Spec: admin-users-change-role.yaml — role must be admin|moderator|contributor

func TestValidateRoleChangeInput_AcceptsAdmin(t *testing.T) {
	assert.NoError(t, service.ValidateRoleChangeInput("admin"))
}

func TestValidateRoleChangeInput_AcceptsModerator(t *testing.T) {
	assert.NoError(t, service.ValidateRoleChangeInput("moderator"))
}

func TestValidateRoleChangeInput_AcceptsContributor(t *testing.T) {
	assert.NoError(t, service.ValidateRoleChangeInput("contributor"))
}

func TestValidateRoleChangeInput_RejectsEmptyRole(t *testing.T) {
	assert.Error(t, service.ValidateRoleChangeInput(""))
}

func TestValidateRoleChangeInput_RejectsInvalidRole(t *testing.T) {
	assert.Error(t, service.ValidateRoleChangeInput("superuser"))
}

func TestValidateRoleChangeInput_IsPure_SameInputReturnsSameOutput(t *testing.T) {
	e1 := service.ValidateRoleChangeInput("moderator")
	e2 := service.ValidateRoleChangeInput("moderator")
	assert.Equal(t, e1, e2)
}

// --- BuildRoleChangedAuditLog --- (Cycle 6)
// Spec: admin-users-change-role.yaml rule "Write AuditLog: action=role_changed, diff={role:{before,after}}"

func TestBuildRoleChangedAuditLog_SetsEntityTypeUser(t *testing.T) {
	entry := service.BuildRoleChangedAuditLog(42, 1, model.UserRoleContributor, model.UserRoleModerator)
	assert.Equal(t, model.AuditEntityUser, entry.EntityType)
}

func TestBuildRoleChangedAuditLog_SetsActionRoleChanged(t *testing.T) {
	entry := service.BuildRoleChangedAuditLog(42, 1, model.UserRoleContributor, model.UserRoleModerator)
	assert.Equal(t, model.AuditActionRoleChanged, entry.Action)
}

func TestBuildRoleChangedAuditLog_SetsDiffWithBeforeAndAfterRole(t *testing.T) {
	entry := service.BuildRoleChangedAuditLog(42, 1, model.UserRoleContributor, model.UserRoleModerator)
	diffStr := string(entry.Diff)
	assert.Contains(t, diffStr, "contributor")
	assert.Contains(t, diffStr, "moderator")
}

func TestBuildRoleChangedAuditLog_IsPure_SameInputReturnsSameOutput(t *testing.T) {
	e1 := service.BuildRoleChangedAuditLog(42, 1, model.UserRoleContributor, model.UserRoleModerator)
	e2 := service.BuildRoleChangedAuditLog(42, 1, model.UserRoleContributor, model.UserRoleModerator)
	assert.Equal(t, e1, e2)
}
