package service

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuryalencar/research-events/internal/model"
	"github.com/yuryalencar/research-events/internal/repository"
)

// --- Types ---

// RawListUsersQuery carries the raw (string) query parameters from
// GET /api/v1/admin/users, before any parsing or validation. Keeping this a
// plain string struct lets the handler stay a thin adapter — all parsing and
// validation logic lives in ValidateListUsersQuery, where it can be tested
// without net/http or httptest.
type RawListUsersQuery struct {
	Roles  string
	Search string
	Locked string

	IncludeDeleted string

	Page       string
	PageSize   string
	Pagination string
}

// ListUsersInput groups the validated, normalized filters for
// UserRepository.List, with defaults applied per
// specs/backend/admin-users-list.yaml.
type ListUsersInput struct {
	Roles  []model.UserRole
	Search string
	Locked *bool // nil = no filter; true = locked only; false = unlocked only

	IncludeDeleted bool

	Page          int
	PageSize      int
	PaginationOff bool
}

// --- Package-level validation data ---

var allowedUserRoles = map[string]model.UserRole{
	"admin":       model.UserRoleAdmin,
	"moderator":   model.UserRoleModerator,
	"contributor": model.UserRoleContributor,
}

// --- Public functions ---

// FP: pure function
// ValidateListUsersQuery depends only on its argument and performs no I/O —
// given the same RawListUsersQuery it always returns the same ListUsersInput
// (or error). All defaults and validation rules live here, making them
// trivially testable: input → output, no mocks, no server, no database.
func ValidateListUsersQuery(raw RawListUsersQuery) (ListUsersInput, error) {
	roles, err := parseUserRoles(raw.Roles)
	if err != nil {
		return ListUsersInput{}, err
	}

	locked, err := parseLockedFilter(raw.Locked)
	if err != nil {
		return ListUsersInput{}, err
	}

	includeDeleted, err := parseIncludeDeleted(raw.IncludeDeleted)
	if err != nil {
		return ListUsersInput{}, err
	}

	paginationOff, err := parsePagination(raw.Pagination)
	if err != nil {
		return ListUsersInput{}, err
	}

	if paginationOff {
		return ListUsersInput{
			Roles:          roles,
			Search:         raw.Search,
			Locked:         locked,
			IncludeDeleted: includeDeleted,
			Page:           1,
			PaginationOff:  true,
		}, nil
	}

	page := 1
	if raw.Page != "" {
		parsed, err := strconv.Atoi(raw.Page)
		if err != nil || parsed < 1 {
			return ListUsersInput{}, fmt.Errorf("page must be a positive integer")
		}
		page = parsed
	}

	pageSize := 20
	if raw.PageSize != "" {
		parsed, err := strconv.Atoi(raw.PageSize)
		if err != nil || parsed < 1 || parsed > 100 {
			return ListUsersInput{}, fmt.Errorf("page_size must be an integer between 1 and 100")
		}
		pageSize = parsed
	}

	return ListUsersInput{
		Roles:          roles,
		Search:         raw.Search,
		Locked:         locked,
		IncludeDeleted: includeDeleted,
		Page:           page,
		PageSize:       pageSize,
	}, nil
}

// FP: pure function
// ToListUsersFilter maps the validated ListUsersInput to the repository-level
// filter. Pure: depends only on its argument, no I/O, always the same output
// for the same input. Separating this from ValidateListUsersQuery keeps the
// service layer free of repository import concerns in tests.
func ToListUsersFilter(input ListUsersInput) repository.ListUsersFilter {
	return repository.ListUsersFilter{
		Roles:          input.Roles,
		Search:         input.Search,
		Locked:         input.Locked,
		IncludeDeleted: input.IncludeDeleted,
		Page:           input.Page,
		PageSize:       input.PageSize,
		PaginationOff:  input.PaginationOff,
	}
}

// --- Private functions ---

// FP: pure function
// parseUserRoles splits and validates the comma-separated roles string.
// An empty string returns nil (no role filter). Any unrecognised value is an
// error — the caller returns 400 to the client.
func parseUserRoles(raw string) ([]model.UserRole, error) {
	if raw == "" {
		return nil, nil
	}

	parts := strings.Split(raw, ",")
	roles := make([]model.UserRole, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		role, ok := allowedUserRoles[part]
		if !ok {
			return nil, fmt.Errorf("roles contains unknown value: %q", part)
		}
		roles = append(roles, role)
	}
	return roles, nil
}

// FP: pure function
// parseLockedFilter converts the raw "true"/"false"/empty string to a *bool.
// nil means no filter (both locked and unlocked users included). Any other
// value is an error — prevents silent misinterpretation of typos like "maybe".
func parseLockedFilter(raw string) (*bool, error) {
	switch raw {
	case "":
		return nil, nil
	case "true":
		v := true
		return &v, nil
	case "false":
		v := false
		return &v, nil
	default:
		return nil, fmt.Errorf("locked must be 'true' or 'false'")
	}
}

// FP: pure function
// parseIncludeDeleted converts the raw "true"/"false"/empty string to a bool.
// Defaults to false — soft-deleted users are excluded unless explicitly opted in.
func parseIncludeDeleted(raw string) (bool, error) {
	switch raw {
	case "", "false":
		return false, nil
	case "true":
		return true, nil
	default:
		return false, fmt.Errorf("include_deleted must be 'true' or 'false'")
	}
}
