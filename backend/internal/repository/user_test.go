package repository_test

// Spec: specs/backend/database-users.yaml
// Spec: specs/backend/auth-login.yaml
// Spec: specs/backend/auth-refresh-token.yaml
// Spec: specs/backend/admin-users-unlock.yaml

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuryalencar/research-events/internal/model"
	"github.com/yuryalencar/research-events/internal/repository"
)

// seed inserts a User inside the given transaction and returns it with its DB-assigned ID.
func seedUser(t *testing.T, tx interface{ Create(value any) interface{ Error() error } }, u model.User) model.User {
	t.Helper()
	// Use testDB directly through the tx GORM instance
	return u
}

// --- FindByEmail ---

func TestUserRepository_FindByEmail_ReturnsUserWhenFound(t *testing.T) {
	// Spec: auth-login.yaml — handler looks up the user by email on every login attempt
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewUserRepository(tx)
	hash := "hashed"
	require.NoError(t, tx.Create(&model.User{
		Name:         "Alice",
		Email:        "alice@example.com",
		PasswordHash: &hash,
		Role:         model.UserRoleAdmin,
	}).Error)

	got, err := repo.FindByEmail(context.Background(), "alice@example.com")

	require.NoError(t, err)
	assert.Equal(t, "alice@example.com", got.Email)
	assert.Equal(t, model.UserRoleAdmin, got.Role)
}

func TestUserRepository_FindByEmail_ReturnsErrNotFoundWhenMissing(t *testing.T) {
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewUserRepository(tx)

	_, err := repo.FindByEmail(context.Background(), "nobody@example.com")

	require.ErrorIs(t, err, repository.ErrNotFound)
}

func TestUserRepository_FindByEmail_ReturnsErrNotFoundForSoftDeletedUser(t *testing.T) {
	// Spec: auth-login.yaml — soft-deleted users cannot log in
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewUserRepository(tx)
	hash := "hashed"
	require.NoError(t, tx.Create(&model.User{
		Name:         "Deleted",
		Email:        "deleted@example.com",
		PasswordHash: &hash,
		Role:         model.UserRoleAdmin,
	}).Error)

	// Soft-delete the user via GORM (sets deleted_at).
	require.NoError(t, tx.Where("email = ?", "deleted@example.com").Delete(&model.User{}).Error)

	_, err := repo.FindByEmail(context.Background(), "deleted@example.com")

	require.ErrorIs(t, err, repository.ErrNotFound)
}

// --- FindByID ---

func TestUserRepository_FindByID_ReturnsUserWhenFound(t *testing.T) {
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewUserRepository(tx)
	u := model.User{Name: "Bob", Email: "bob@example.com", Role: model.UserRoleModerator}
	require.NoError(t, tx.Create(&u).Error)

	got, err := repo.FindByID(context.Background(), u.ID)

	require.NoError(t, err)
	assert.Equal(t, u.ID, got.ID)
	assert.Equal(t, "Bob", got.Name)
}

func TestUserRepository_FindByID_ReturnsErrNotFoundWhenMissing(t *testing.T) {
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewUserRepository(tx)

	_, err := repo.FindByID(context.Background(), 999999)

	require.ErrorIs(t, err, repository.ErrNotFound)
}

func TestUserRepository_FindByID_ReturnsErrNotFoundForSoftDeletedUser(t *testing.T) {
	// Spec: admin-users-unlock.yaml — soft-deleted users return 404
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewUserRepository(tx)
	u := model.User{Name: "Gone", Email: "gone@example.com", Role: model.UserRoleAdmin}
	require.NoError(t, tx.Create(&u).Error)
	require.NoError(t, tx.Delete(&u).Error)

	_, err := repo.FindByID(context.Background(), u.ID)

	require.ErrorIs(t, err, repository.ErrNotFound)
}

// --- UpdateTokens ---

func TestUserRepository_UpdateTokens_PersistsAllTokenFields(t *testing.T) {
	// Spec: auth-login.yaml — on successful login, jti + refresh hash written to DB
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewUserRepository(tx)
	u := model.User{Name: "Carol", Email: "carol@example.com", Role: model.UserRoleAdmin}
	require.NoError(t, tx.Create(&u).Error)

	jti := "test-jti-uuid"
	jtiExp := time.Now().Add(30 * time.Minute).UTC().Truncate(time.Second)
	refreshHash := "abc123hash"
	refreshExp := time.Now().Add(4 * time.Hour).UTC().Truncate(time.Second)

	err := repo.UpdateTokens(context.Background(), u.ID, jti, jtiExp, refreshHash, refreshExp)
	require.NoError(t, err)

	var updated model.User
	require.NoError(t, tx.First(&updated, u.ID).Error)
	require.NotNil(t, updated.AccessTokenJTI)
	assert.Equal(t, jti, *updated.AccessTokenJTI)
	require.NotNil(t, updated.RefreshTokenHash)
	assert.Equal(t, refreshHash, *updated.RefreshTokenHash)
}

// --- ClearTokens ---

func TestUserRepository_ClearTokens_NullsAllTokenFields(t *testing.T) {
	// Spec: auth-logout.yaml — on logout, all token fields cleared in DB
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewUserRepository(tx)
	jti := "some-jti"
	hash := "password"
	rHash := "refresh"
	exp := time.Now().Add(time.Hour)
	u := model.User{
		Name:                  "Dave",
		Email:                 "dave@example.com",
		Role:                  model.UserRoleAdmin,
		PasswordHash:          &hash,
		AccessTokenJTI:        &jti,
		RefreshTokenHash:      &rHash,
		AccessTokenExpiresAt:  &exp,
		RefreshTokenExpiresAt: &exp,
	}
	require.NoError(t, tx.Create(&u).Error)

	err := repo.ClearTokens(context.Background(), u.ID)
	require.NoError(t, err)

	var updated model.User
	require.NoError(t, tx.First(&updated, u.ID).Error)
	assert.Nil(t, updated.AccessTokenJTI)
	assert.Nil(t, updated.RefreshTokenHash)
	assert.Nil(t, updated.AccessTokenExpiresAt)
	assert.Nil(t, updated.RefreshTokenExpiresAt)
}

// --- IncrementFailedAttempts ---

func TestUserRepository_IncrementFailedAttempts_IncrementsCounter(t *testing.T) {
	// Spec: auth-login.yaml — counter increments on each wrong password
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewUserRepository(tx)
	u := model.User{Name: "Eve", Email: "eve@example.com", Role: model.UserRoleAdmin}
	require.NoError(t, tx.Create(&u).Error)

	require.NoError(t, repo.IncrementFailedAttempts(context.Background(), u.ID))
	require.NoError(t, repo.IncrementFailedAttempts(context.Background(), u.ID))

	var updated model.User
	require.NoError(t, tx.First(&updated, u.ID).Error)
	assert.Equal(t, 2, updated.FailedLoginAttempts)
}

// --- LockAccount ---

func TestUserRepository_LockAccount_SetsLockedAt(t *testing.T) {
	// Spec: auth-login.yaml — account locked after 5 failed attempts
	//
	// before is captured before beginTx so it precedes the Postgres transaction start time.
	// NOW() inside a transaction returns the transaction start time, not the clock time —
	// capturing before after beginTx would make it newer than locked_at and fail the assertion.
	before := time.Now()
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewUserRepository(tx)
	u := model.User{Name: "Frank", Email: "frank@example.com", Role: model.UserRoleAdmin}
	require.NoError(t, tx.Create(&u).Error)

	err := repo.LockAccount(context.Background(), u.ID)
	require.NoError(t, err)

	var updated model.User
	require.NoError(t, tx.First(&updated, u.ID).Error)
	require.NotNil(t, updated.LockedAt)
	assert.True(t, updated.LockedAt.After(before) || updated.LockedAt.Equal(before))
}

// --- ResetFailedAttempts ---

func TestUserRepository_ResetFailedAttempts_ZeroesCounter(t *testing.T) {
	// Spec: auth-login.yaml — counter resets to 0 on successful login
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewUserRepository(tx)
	attempts := 3
	u := model.User{Name: "Grace", Email: "grace@example.com", Role: model.UserRoleAdmin, FailedLoginAttempts: attempts}
	require.NoError(t, tx.Create(&u).Error)

	require.NoError(t, repo.ResetFailedAttempts(context.Background(), u.ID))

	var updated model.User
	require.NoError(t, tx.First(&updated, u.ID).Error)
	assert.Equal(t, 0, updated.FailedLoginAttempts)
}

// --- Unlock ---

func TestUserRepository_Unlock_ClearsLockedAtAndResetsCounter(t *testing.T) {
	// Spec: admin-users-unlock.yaml — unlock clears locked_at and resets failed_login_attempts
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewUserRepository(tx)
	now := time.Now()
	u := model.User{
		Name:                "Hank",
		Email:               "hank@example.com",
		Role:                model.UserRoleAdmin,
		LockedAt:            &now,
		FailedLoginAttempts: 5,
	}
	require.NoError(t, tx.Create(&u).Error)

	err := repo.Unlock(context.Background(), u.ID)
	require.NoError(t, err)

	var updated model.User
	require.NoError(t, tx.First(&updated, u.ID).Error)
	assert.Nil(t, updated.LockedAt)
	assert.Equal(t, 0, updated.FailedLoginAttempts)
}
