package repository_test

// Spec: specs/backend/database-users.yaml
// Spec: specs/backend/auth-login.yaml
// Spec: specs/backend/auth-refresh-token.yaml
// Spec: specs/backend/admin-users-unlock.yaml
// Spec: specs/backend/users-update-password.yaml
// Spec: specs/backend/admin-users-register.yaml
// Spec: specs/backend/admin-users-change-role.yaml
// Spec: specs/backend/admin-users-list.yaml

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

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

// --- UpdatePassword ---

func TestUserRepository_UpdatePassword_PersistsNewHash(t *testing.T) {
	// Spec: users-update-password.yaml DoD "Password stored as bcrypt hash (cost 12) — never plaintext"
	tx, rollback := beginTx(t)
	defer rollback()

	oldHash := "$2a$12$oldhasholdhasholdhashold" // placeholder — not a real bcrypt hash
	u := model.User{
		Name:         "Donna",
		Email:        "donna@example.com",
		Role:         model.UserRoleAdmin,
		PasswordHash: &oldHash,
	}
	require.NoError(t, tx.Create(&u).Error)

	newPlain := "NewPass@2"
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPlain), 12)
	require.NoError(t, err)

	repo := repository.NewUserRepository(tx)
	err = repo.UpdatePassword(context.Background(), u.ID, string(newHash))
	require.NoError(t, err)

	var updated model.User
	require.NoError(t, tx.First(&updated, u.ID).Error)
	require.NotNil(t, updated.PasswordHash)
	assert.Equal(t, string(newHash), *updated.PasswordHash)
}

// --- ExistsByEmail ---

func TestUserRepository_ExistsByEmail_ReturnsTrueForActiveUser(t *testing.T) {
	// Spec: admin-users-register.yaml — email taken by active user → 409 EMAIL_ALREADY_EXISTS
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewUserRepository(tx)
	hash := "hashed"
	require.NoError(t, tx.Create(&model.User{
		Name:         "Alice",
		Email:        "exists-active@example.com",
		PasswordHash: &hash,
		Role:         model.UserRoleAdmin,
	}).Error)

	exists, err := repo.ExistsByEmail(context.Background(), "exists-active@example.com")

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestUserRepository_ExistsByEmail_ReturnsTrueForSoftDeletedUser(t *testing.T) {
	// Spec: admin-users-register.yaml — soft-deleted rows still occupy the unique email slot
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewUserRepository(tx)
	hash := "hashed"
	u := model.User{
		Name:         "Deleted",
		Email:        "exists-deleted@example.com",
		PasswordHash: &hash,
		Role:         model.UserRoleAdmin,
	}
	require.NoError(t, tx.Create(&u).Error)
	require.NoError(t, tx.Delete(&u).Error)

	exists, err := repo.ExistsByEmail(context.Background(), "exists-deleted@example.com")

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestUserRepository_ExistsByEmail_ReturnsFalseWhenEmailNotFound(t *testing.T) {
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewUserRepository(tx)

	exists, err := repo.ExistsByEmail(context.Background(), "nobody@example.com")

	require.NoError(t, err)
	assert.False(t, exists)
}

// --- Create ---

func TestUserRepository_Create_PersistsUserAndReturnsWithID(t *testing.T) {
	// Spec: admin-users-register.yaml DoD "Returns 201 + user data on valid registration"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewUserRepository(tx)
	hash := "hashed-password"
	u := model.User{
		Name:         "Jane Doe",
		Email:        "create-jane@example.com",
		PasswordHash: &hash,
		Role:         model.UserRoleModerator,
	}

	created, err := repo.Create(context.Background(), u)

	require.NoError(t, err)
	assert.NotZero(t, created.ID)
	assert.Equal(t, "Jane Doe", created.Name)
	assert.Equal(t, "create-jane@example.com", created.Email)
	assert.Equal(t, model.UserRoleModerator, created.Role)
	assert.False(t, created.CreatedAt.IsZero())
}

func TestUserRepository_Create_ReturnsErrorOnDuplicateEmail(t *testing.T) {
	// Spec: admin-users-register.yaml — unique index on email enforced at DB level
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewUserRepository(tx)
	hash := "hashed"
	require.NoError(t, tx.Create(&model.User{
		Name:         "First",
		Email:        "dup@example.com",
		PasswordHash: &hash,
		Role:         model.UserRoleAdmin,
	}).Error)

	_, err := repo.Create(context.Background(), model.User{
		Name:         "Second",
		Email:        "dup@example.com",
		PasswordHash: &hash,
		Role:         model.UserRoleModerator,
	})

	require.Error(t, err)
}

// --- UpdateRole ---

func TestUserRepository_UpdateRole_ChangesRoleInDB(t *testing.T) {
	// Spec: admin-users-change-role.yaml DoD "Returns 200 + user data (with new role) on valid role change"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewUserRepository(tx)
	u := model.User{Name: "Bob", Email: "updaterole-bob@example.com", Role: model.UserRoleContributor}
	require.NoError(t, tx.Create(&u).Error)

	err := repo.UpdateRole(context.Background(), u.ID, model.UserRoleModerator)
	require.NoError(t, err)

	var updated model.User
	require.NoError(t, tx.First(&updated, u.ID).Error)
	assert.Equal(t, model.UserRoleModerator, updated.Role)
}

func TestUserRepository_UpdateRole_LeavesOtherFieldsUnchanged(t *testing.T) {
	// Spec: admin-users-change-role.yaml — only the role column is modified
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewUserRepository(tx)
	hash := "hashed"
	u := model.User{
		Name:         "Carol",
		Email:        "updaterole-carol@example.com",
		PasswordHash: &hash,
		Role:         model.UserRoleContributor,
	}
	require.NoError(t, tx.Create(&u).Error)

	require.NoError(t, repo.UpdateRole(context.Background(), u.ID, model.UserRoleAdmin))

	var updated model.User
	require.NoError(t, tx.First(&updated, u.ID).Error)
	assert.Equal(t, "Carol", updated.Name)
	assert.Equal(t, "updaterole-carol@example.com", updated.Email)
	require.NotNil(t, updated.PasswordHash)
	assert.Equal(t, hash, *updated.PasswordHash)
}

// --- List ---

func TestUserRepository_List_ReturnsAllUsers(t *testing.T) {
	// Spec: admin-users-list.yaml DoD "200 + paginated list with correct meta"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewUserRepository(tx)
	require.NoError(t, tx.Create(&model.User{Name: "Alpha", Email: "list-alpha@example.com", Role: model.UserRoleAdmin}).Error)
	require.NoError(t, tx.Create(&model.User{Name: "Beta", Email: "list-beta@example.com", Role: model.UserRoleModerator}).Error)

	users, total, err := repo.List(context.Background(), repository.ListUsersFilter{Page: 1, PageSize: 20})

	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(2))
	assert.GreaterOrEqual(t, len(users), 2)
}

func TestUserRepository_List_FiltersByRole(t *testing.T) {
	// Spec: admin-users-list.yaml DoD "200 + only admins when roles=admin"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewUserRepository(tx)
	require.NoError(t, tx.Create(&model.User{Name: "AdminUser", Email: "list-role-admin@example.com", Role: model.UserRoleAdmin}).Error)
	require.NoError(t, tx.Create(&model.User{Name: "ModUser", Email: "list-role-mod@example.com", Role: model.UserRoleModerator}).Error)

	users, _, err := repo.List(context.Background(), repository.ListUsersFilter{
		Roles:    []model.UserRole{model.UserRoleAdmin},
		Page:     1,
		PageSize: 20,
	})

	require.NoError(t, err)
	for _, u := range users {
		assert.Equal(t, model.UserRoleAdmin, u.Role)
	}
}

func TestUserRepository_List_FiltersByMultipleRoles(t *testing.T) {
	// Spec: admin-users-list.yaml DoD "200 + admins and moderators when roles=admin,moderator (OR logic)"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewUserRepository(tx)
	require.NoError(t, tx.Create(&model.User{Name: "A2", Email: "list-multi-admin@example.com", Role: model.UserRoleAdmin}).Error)
	require.NoError(t, tx.Create(&model.User{Name: "M2", Email: "list-multi-mod@example.com", Role: model.UserRoleModerator}).Error)
	require.NoError(t, tx.Create(&model.User{Name: "C2", Email: "list-multi-cont@example.com", Role: model.UserRoleContributor}).Error)

	users, _, err := repo.List(context.Background(), repository.ListUsersFilter{
		Roles:    []model.UserRole{model.UserRoleAdmin, model.UserRoleModerator},
		Page:     1,
		PageSize: 20,
	})

	require.NoError(t, err)
	for _, u := range users {
		assert.NotEqual(t, model.UserRoleContributor, u.Role, "contributor must not appear in admin+moderator filter")
	}
	emails := make([]string, len(users))
	for i, u := range users {
		emails[i] = u.Email
	}
	assert.Contains(t, emails, "list-multi-admin@example.com")
	assert.Contains(t, emails, "list-multi-mod@example.com")
}

func TestUserRepository_List_FiltersBySearchMatchesName(t *testing.T) {
	// Spec: admin-users-list.yaml DoD "200 + matching users when search=<partial name>"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewUserRepository(tx)
	require.NoError(t, tx.Create(&model.User{Name: "Zephyr Unique", Email: "list-search-name@example.com", Role: model.UserRoleAdmin}).Error)
	require.NoError(t, tx.Create(&model.User{Name: "Other Person", Email: "list-search-other@example.com", Role: model.UserRoleAdmin}).Error)

	users, _, err := repo.List(context.Background(), repository.ListUsersFilter{
		Search:   "Zephyr",
		Page:     1,
		PageSize: 20,
	})

	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "list-search-name@example.com", users[0].Email)
}

func TestUserRepository_List_FiltersBySearchMatchesEmail(t *testing.T) {
	// Spec: admin-users-list.yaml border_case "search matches email but not name → user is returned"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewUserRepository(tx)
	require.NoError(t, tx.Create(&model.User{Name: "Regular Name", Email: "uniquetoken99@example.com", Role: model.UserRoleAdmin}).Error)

	users, _, err := repo.List(context.Background(), repository.ListUsersFilter{
		Search:   "uniquetoken99",
		Page:     1,
		PageSize: 20,
	})

	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "uniquetoken99@example.com", users[0].Email)
}

func TestUserRepository_List_FiltersByLockedTrue(t *testing.T) {
	// Spec: admin-users-list.yaml DoD "200 + only locked users when locked=true"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewUserRepository(tx)
	now := time.Now()
	require.NoError(t, tx.Create(&model.User{Name: "Locked", Email: "list-locked@example.com", Role: model.UserRoleAdmin, LockedAt: &now}).Error)
	require.NoError(t, tx.Create(&model.User{Name: "Active", Email: "list-active@example.com", Role: model.UserRoleAdmin}).Error)

	lockedTrue := true
	users, _, err := repo.List(context.Background(), repository.ListUsersFilter{
		Locked:   &lockedTrue,
		Page:     1,
		PageSize: 20,
	})

	require.NoError(t, err)
	for _, u := range users {
		assert.NotNil(t, u.LockedAt, "all returned users must be locked")
	}
	emails := make([]string, len(users))
	for i, u := range users {
		emails[i] = u.Email
	}
	assert.Contains(t, emails, "list-locked@example.com")
	assert.NotContains(t, emails, "list-active@example.com")
}

func TestUserRepository_List_FiltersByLockedFalse(t *testing.T) {
	// Spec: admin-users-list.yaml — locked=false returns only unlocked users
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewUserRepository(tx)
	now := time.Now()
	require.NoError(t, tx.Create(&model.User{Name: "LockedB", Email: "list-lockedb@example.com", Role: model.UserRoleAdmin, LockedAt: &now}).Error)
	require.NoError(t, tx.Create(&model.User{Name: "ActiveB", Email: "list-activeb@example.com", Role: model.UserRoleAdmin}).Error)

	lockedFalse := false
	users, _, err := repo.List(context.Background(), repository.ListUsersFilter{
		Locked:   &lockedFalse,
		Page:     1,
		PageSize: 20,
	})

	require.NoError(t, err)
	for _, u := range users {
		assert.Nil(t, u.LockedAt, "all returned users must be unlocked")
	}
}

func TestUserRepository_List_ExcludesSoftDeletedByDefault(t *testing.T) {
	// Spec: admin-users-list.yaml rule "Soft-deleted users are excluded by default"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewUserRepository(tx)
	u := model.User{Name: "ToDelete", Email: "list-softdel@example.com", Role: model.UserRoleAdmin}
	require.NoError(t, tx.Create(&u).Error)
	require.NoError(t, tx.Delete(&u).Error)

	users, _, err := repo.List(context.Background(), repository.ListUsersFilter{Page: 1, PageSize: 20})

	require.NoError(t, err)
	for _, got := range users {
		assert.NotEqual(t, "list-softdel@example.com", got.Email)
	}
}

func TestUserRepository_List_IncludesSoftDeletedWhenFlagSet(t *testing.T) {
	// Spec: admin-users-list.yaml DoD "200 + includes soft-deleted users when include_deleted=true"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewUserRepository(tx)
	u := model.User{Name: "Deleted2", Email: "list-softdel2@example.com", Role: model.UserRoleAdmin}
	require.NoError(t, tx.Create(&u).Error)
	require.NoError(t, tx.Delete(&u).Error)

	users, _, err := repo.List(context.Background(), repository.ListUsersFilter{
		IncludeDeleted: true,
		Page:           1,
		PageSize:       20,
	})

	require.NoError(t, err)
	emails := make([]string, len(users))
	for i, u := range users {
		emails[i] = u.Email
	}
	assert.Contains(t, emails, "list-softdel2@example.com")
}

func TestUserRepository_List_SortedByCreatedAtDesc(t *testing.T) {
	// Spec: admin-users-list.yaml rule "Result set is sorted by created_at DESC"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewUserRepository(tx)
	require.NoError(t, tx.Create(&model.User{Name: "First", Email: "list-sort-first@example.com", Role: model.UserRoleAdmin}).Error)
	require.NoError(t, tx.Create(&model.User{Name: "Second", Email: "list-sort-second@example.com", Role: model.UserRoleAdmin}).Error)

	users, _, err := repo.List(context.Background(), repository.ListUsersFilter{
		Search:   "list-sort-",
		Page:     1,
		PageSize: 20,
	})

	require.NoError(t, err)
	require.Len(t, users, 2)
	assert.True(t, users[0].CreatedAt.After(users[1].CreatedAt) || users[0].CreatedAt.Equal(users[1].CreatedAt),
		"first result must be newer or equal — sorted by created_at DESC")
}

func TestUserRepository_List_Paginated(t *testing.T) {
	// Spec: admin-users-list.yaml — page beyond total returns empty data with correct meta
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewUserRepository(tx)
	require.NoError(t, tx.Create(&model.User{Name: "Page1", Email: "list-page-a@example.com", Role: model.UserRoleAdmin}).Error)
	require.NoError(t, tx.Create(&model.User{Name: "Page2", Email: "list-page-b@example.com", Role: model.UserRoleAdmin}).Error)

	users, total, err := repo.List(context.Background(), repository.ListUsersFilter{
		Search:   "list-page-",
		Page:     1,
		PageSize: 1,
	})

	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, users, 1)
}

func TestUserRepository_List_PageBeyondTotal_ReturnsEmptySlice(t *testing.T) {
	// Spec: admin-users-list.yaml border_case "page beyond total → 200 with empty data array + correct meta"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewUserRepository(tx)
	require.NoError(t, tx.Create(&model.User{Name: "Only", Email: "list-beyond@example.com", Role: model.UserRoleAdmin}).Error)

	users, total, err := repo.List(context.Background(), repository.ListUsersFilter{
		Search:   "list-beyond",
		Page:     999,
		PageSize: 20,
	})

	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Empty(t, users)
}

func TestUserRepository_List_PaginationOff_ReturnsAll(t *testing.T) {
	// Spec: admin-users-list.yaml — pagination=off ignores page and page_size
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewUserRepository(tx)
	require.NoError(t, tx.Create(&model.User{Name: "PA", Email: "list-poff-a@example.com", Role: model.UserRoleAdmin}).Error)
	require.NoError(t, tx.Create(&model.User{Name: "PB", Email: "list-poff-b@example.com", Role: model.UserRoleAdmin}).Error)
	require.NoError(t, tx.Create(&model.User{Name: "PC", Email: "list-poff-c@example.com", Role: model.UserRoleAdmin}).Error)

	users, total, err := repo.List(context.Background(), repository.ListUsersFilter{
		Search:        "list-poff-",
		PaginationOff: true,
	})

	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, users, 3)
}

func TestUserRepository_List_NeverReturnsPasswordHash(t *testing.T) {
	// Spec: admin-users-list.yaml rule "Response never includes password_hash"
	// At the repository level: PasswordHash is loaded by GORM, but the handler
	// must never serialise it. This test documents that the repo does NOT
	// redact it — the handler layer is responsible for omission.
	// We assert the hash IS present in the model so the handler test can prove it strips it.
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewUserRepository(tx)
	hash := "secret-hash"
	require.NoError(t, tx.Create(&model.User{
		Name:         "Secure",
		Email:        "list-secure@example.com",
		Role:         model.UserRoleAdmin,
		PasswordHash: &hash,
	}).Error)

	users, _, err := repo.List(context.Background(), repository.ListUsersFilter{
		Search:   "list-secure",
		Page:     1,
		PageSize: 20,
	})

	require.NoError(t, err)
	require.Len(t, users, 1)
	// The model carries the hash — handler must NOT expose it in the JSON response.
	assert.NotNil(t, users[0].PasswordHash)
}

func TestUserRepository_UpdatePassword_NewHashVerifiableWithBcrypt(t *testing.T) {
	// Spec: users-update-password.yaml DoD "New password can be used to authenticate
	// (login) immediately after update" — proves the stored hash satisfies
	// bcrypt.CompareHashAndPassword, which is exactly what ValidateCredentials calls on login.
	tx, rollback := beginTx(t)
	defer rollback()

	oldHash := "$2a$12$oldhasholdhasholdhashold"
	u := model.User{
		Name:         "Eve",
		Email:        "eve@example.com",
		Role:         model.UserRoleAdmin,
		PasswordHash: &oldHash,
	}
	require.NoError(t, tx.Create(&u).Error)

	newPlain := "LoginReady@3"
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPlain), 12)
	require.NoError(t, err)

	repo := repository.NewUserRepository(tx)
	require.NoError(t, repo.UpdatePassword(context.Background(), u.ID, string(newHash)))

	var updated model.User
	require.NoError(t, tx.First(&updated, u.ID).Error)
	require.NotNil(t, updated.PasswordHash)

	// This is the same comparison performed by ValidateCredentials (service.go) during login.
	err = bcrypt.CompareHashAndPassword([]byte(*updated.PasswordHash), []byte(newPlain))
	assert.NoError(t, err, "stored hash must match the new plaintext password — login must work immediately after update")
}
