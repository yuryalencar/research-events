package service_test

// Spec: specs/backend/admin-users-list.yaml
//
// ValidateListUsersQuery is pure: given the same RawListUsersQuery it always
// returns the same ListUsersInput (or error), with no I/O or hidden state.

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuryalencar/research-events/internal/model"
	"github.com/yuryalencar/research-events/internal/service"
)

// --- Defaults ---

func TestValidateListUsersQuery_NoParams_ReturnsDefaults(t *testing.T) {
	// Spec: all params omitted → page=1, pageSize=20, paginationOff=false, no filters
	got, err := service.ValidateListUsersQuery(service.RawListUsersQuery{})

	require.NoError(t, err)
	assert.Empty(t, got.Roles)
	assert.Equal(t, "", got.Search)
	assert.Nil(t, got.Locked)
	assert.False(t, got.IncludeDeleted)
	assert.Equal(t, 1, got.Page)
	assert.Equal(t, 20, got.PageSize)
	assert.False(t, got.PaginationOff)
}

// --- Roles ---

func TestValidateListUsersQuery_ValidSingleRole_ReturnsParsedRole(t *testing.T) {
	// Spec: roles=admin → only admin role in filter
	got, err := service.ValidateListUsersQuery(service.RawListUsersQuery{Roles: "admin"})

	require.NoError(t, err)
	assert.Equal(t, []model.UserRole{model.UserRoleAdmin}, got.Roles)
}

func TestValidateListUsersQuery_ValidMultipleRoles_ReturnsBothRoles(t *testing.T) {
	// Spec: roles=admin,moderator → OR filter includes both
	got, err := service.ValidateListUsersQuery(service.RawListUsersQuery{Roles: "admin,moderator"})

	require.NoError(t, err)
	assert.ElementsMatch(t, []model.UserRole{model.UserRoleAdmin, model.UserRoleModerator}, got.Roles)
}

func TestValidateListUsersQuery_AllThreeRoles_ReturnsAllRoles(t *testing.T) {
	// Spec: roles=admin,moderator,contributor → all three roles (same as no filter)
	got, err := service.ValidateListUsersQuery(service.RawListUsersQuery{Roles: "admin,moderator,contributor"})

	require.NoError(t, err)
	assert.ElementsMatch(t, []model.UserRole{
		model.UserRoleAdmin,
		model.UserRoleModerator,
		model.UserRoleContributor,
	}, got.Roles)
}

func TestValidateListUsersQuery_UnknownRole_ReturnsError(t *testing.T) {
	// Spec: roles contains unrecognised value → 400
	_, err := service.ValidateListUsersQuery(service.RawListUsersQuery{Roles: "superuser"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "superuser")
}

func TestValidateListUsersQuery_MixedValidAndInvalidRole_ReturnsError(t *testing.T) {
	// Spec: any unknown role in the list → 400, even if others are valid
	_, err := service.ValidateListUsersQuery(service.RawListUsersQuery{Roles: "admin,superuser"})

	require.Error(t, err)
}

// --- Search ---

func TestValidateListUsersQuery_SearchProvided_PassthroughAsIs(t *testing.T) {
	// Spec: search=john → input.Search == "john" (no transformation here — DB does ILIKE)
	got, err := service.ValidateListUsersQuery(service.RawListUsersQuery{Search: "john"})

	require.NoError(t, err)
	assert.Equal(t, "john", got.Search)
}

func TestValidateListUsersQuery_SearchOmitted_ReturnsEmptyString(t *testing.T) {
	got, err := service.ValidateListUsersQuery(service.RawListUsersQuery{})

	require.NoError(t, err)
	assert.Equal(t, "", got.Search)
}

// --- Locked ---

func TestValidateListUsersQuery_LockedTrue_ReturnsTruePointer(t *testing.T) {
	// Spec: locked=true → filter to users with locked_at IS NOT NULL
	got, err := service.ValidateListUsersQuery(service.RawListUsersQuery{Locked: "true"})

	require.NoError(t, err)
	require.NotNil(t, got.Locked)
	assert.True(t, *got.Locked)
}

func TestValidateListUsersQuery_LockedFalse_ReturnsFalsePointer(t *testing.T) {
	// Spec: locked=false → filter to users with locked_at IS NULL
	got, err := service.ValidateListUsersQuery(service.RawListUsersQuery{Locked: "false"})

	require.NoError(t, err)
	require.NotNil(t, got.Locked)
	assert.False(t, *got.Locked)
}

func TestValidateListUsersQuery_LockedOmitted_ReturnsNil(t *testing.T) {
	// Spec: locked param omitted → both locked and unlocked included (no filter)
	got, err := service.ValidateListUsersQuery(service.RawListUsersQuery{})

	require.NoError(t, err)
	assert.Nil(t, got.Locked)
}

func TestValidateListUsersQuery_LockedInvalidValue_ReturnsError(t *testing.T) {
	// Spec: locked=maybe is not a valid boolean → 400
	_, err := service.ValidateListUsersQuery(service.RawListUsersQuery{Locked: "maybe"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "locked")
}

// --- IncludeDeleted ---

func TestValidateListUsersQuery_IncludeDeletedTrue_ReturnsTrue(t *testing.T) {
	// Spec: include_deleted=true → soft-deleted users included
	got, err := service.ValidateListUsersQuery(service.RawListUsersQuery{IncludeDeleted: "true"})

	require.NoError(t, err)
	assert.True(t, got.IncludeDeleted)
}

func TestValidateListUsersQuery_IncludeDeletedFalse_ReturnsFalse(t *testing.T) {
	got, err := service.ValidateListUsersQuery(service.RawListUsersQuery{IncludeDeleted: "false"})

	require.NoError(t, err)
	assert.False(t, got.IncludeDeleted)
}

func TestValidateListUsersQuery_IncludeDeletedOmitted_DefaultsFalse(t *testing.T) {
	// Spec: default false — soft-deleted users excluded
	got, err := service.ValidateListUsersQuery(service.RawListUsersQuery{})

	require.NoError(t, err)
	assert.False(t, got.IncludeDeleted)
}

func TestValidateListUsersQuery_IncludeDeletedInvalidValue_ReturnsError(t *testing.T) {
	// Spec: include_deleted=yes is not a valid boolean → 400
	_, err := service.ValidateListUsersQuery(service.RawListUsersQuery{IncludeDeleted: "yes"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "include_deleted")
}

// --- Pagination ---

func TestValidateListUsersQuery_PageZero_ReturnsError(t *testing.T) {
	// Spec: page min 1 — page=0 → 400
	_, err := service.ValidateListUsersQuery(service.RawListUsersQuery{Page: "0"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "page")
}

func TestValidateListUsersQuery_PageNonInteger_ReturnsError(t *testing.T) {
	_, err := service.ValidateListUsersQuery(service.RawListUsersQuery{Page: "abc"})

	require.Error(t, err)
}

func TestValidateListUsersQuery_PageSizeExceedsMax_ReturnsError(t *testing.T) {
	// Spec: page_size max 100 — page_size=101 → 400
	_, err := service.ValidateListUsersQuery(service.RawListUsersQuery{PageSize: "101"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "page_size")
}

func TestValidateListUsersQuery_PageSizeZero_ReturnsError(t *testing.T) {
	_, err := service.ValidateListUsersQuery(service.RawListUsersQuery{PageSize: "0"})

	require.Error(t, err)
}

func TestValidateListUsersQuery_PaginationOff_SetsPaginationOffTrue(t *testing.T) {
	// Spec: pagination=off → all rows returned; page and page_size ignored
	got, err := service.ValidateListUsersQuery(service.RawListUsersQuery{Pagination: "off"})

	require.NoError(t, err)
	assert.True(t, got.PaginationOff)
}

func TestValidateListUsersQuery_PaginationOn_SetsPaginationOffFalse(t *testing.T) {
	got, err := service.ValidateListUsersQuery(service.RawListUsersQuery{Pagination: "on"})

	require.NoError(t, err)
	assert.False(t, got.PaginationOff)
}

func TestValidateListUsersQuery_PaginationInvalid_ReturnsError(t *testing.T) {
	_, err := service.ValidateListUsersQuery(service.RawListUsersQuery{Pagination: "yes"})

	require.Error(t, err)
}

func TestValidateListUsersQuery_PaginationOffIgnoresPageAndPageSize(t *testing.T) {
	// Spec: pagination=off → page/page_size irrelevant; output defaults to page=1
	got, err := service.ValidateListUsersQuery(service.RawListUsersQuery{
		Pagination: "off",
		Page:       "999",
		PageSize:   "5",
	})

	require.NoError(t, err)
	assert.True(t, got.PaginationOff)
	assert.Equal(t, 1, got.Page)
}
