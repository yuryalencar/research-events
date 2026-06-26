package repository

//go:generate mockgen -source=user.go -destination=mocks/mock_user.go -package=mocks

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/yuryalencar/research-events/internal/model"
)

// --- Errors ---

// ErrNotFound is returned when a query matches no rows.
// Callers compare with errors.Is(err, repository.ErrNotFound) — never check the string.
var ErrNotFound = errors.New("record not found")

// --- Interface ---

// UserRepository defines all persistence operations for the User model.
// Interfaces belong in the consumer package — here, the repository package defines
// its own contract so service and handler layers depend on the interface, not the
// concrete GORM implementation.
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (model.User, error)
	FindByID(ctx context.Context, id uint) (model.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	Create(ctx context.Context, user model.User) (model.User, error)
	List(ctx context.Context, filter ListUsersFilter) ([]model.User, int64, error)
	UpdateRole(ctx context.Context, userID uint, newRole model.UserRole) error
	UpdateTokens(ctx context.Context, userID uint, jti string, jtiExp time.Time, refreshHash string, refreshExp time.Time) error
	ClearTokens(ctx context.Context, userID uint) error
	IncrementFailedAttempts(ctx context.Context, userID uint) error
	LockAccount(ctx context.Context, userID uint) error
	ResetFailedAttempts(ctx context.Context, userID uint) error
	Unlock(ctx context.Context, userID uint) error
	UpdatePassword(ctx context.Context, userID uint, newHash string) error
}

// --- Types ---

// ListUsersFilter carries the validated filters for UserRepository.List.
// Roles is an OR filter — empty means all roles.
// Locked nil means no filter; true/false filters by locked_at IS [NOT] NULL.
// IncludeDeleted true adds soft-deleted users to the result set.
type ListUsersFilter struct {
	Roles  []model.UserRole
	Search string
	Locked *bool

	IncludeDeleted bool

	Page          int
	PageSize      int
	PaginationOff bool
}

type userRepository struct {
	db *gorm.DB
}

// --- Constructor ---

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// --- Public methods ---

func (r *userRepository) List(ctx context.Context, filter ListUsersFilter) ([]model.User, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.User{})

	if filter.IncludeDeleted {
		q = q.Unscoped()
	}

	if len(filter.Roles) > 0 {
		q = q.Where("role IN ?", filter.Roles)
	}

	if filter.Search != "" {
		pattern := "%" + filter.Search + "%"
		q = q.Where("name ILIKE ? OR email ILIKE ?", pattern, pattern)
	}

	if filter.Locked != nil {
		if *filter.Locked {
			q = q.Where("locked_at IS NOT NULL")
		} else {
			q = q.Where("locked_at IS NULL")
		}
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	q = q.Order("created_at DESC")

	if !filter.PaginationOff {
		offset := (filter.Page - 1) * filter.PageSize
		q = q.Limit(filter.PageSize).Offset(offset)
	}

	var users []model.User
	if err := q.Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Unscoped().Model(&model.User{}).
		Where("email = ?", email).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *userRepository) Create(ctx context.Context, user model.User) (model.User, error) {
	err := r.db.WithContext(ctx).Create(&user).Error
	return user, err
}

func (r *userRepository) UpdateRole(ctx context.Context, userID uint, newRole model.UserRole) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).
		Update("role", newRole).Error
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ? AND deleted_at IS NULL", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.User{}, ErrNotFound
	}
	return user, err
}

func (r *userRepository) FindByID(ctx context.Context, id uint) (model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.User{}, ErrNotFound
	}
	return user, err
}

func (r *userRepository) UpdateTokens(ctx context.Context, userID uint, jti string, jtiExp time.Time, refreshHash string, refreshExp time.Time) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Updates(map[string]any{
		"access_token_jti":         jti,
		"access_token_expires_at":  jtiExp,
		"refresh_token_hash":       refreshHash,
		"refresh_token_expires_at": refreshExp,
	}).Error
}

func (r *userRepository) ClearTokens(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Updates(map[string]any{
		"access_token_jti":         nil,
		"access_token_expires_at":  nil,
		"refresh_token_hash":       nil,
		"refresh_token_expires_at": nil,
	}).Error
}

func (r *userRepository) IncrementFailedAttempts(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).
		UpdateColumn("failed_login_attempts", gorm.Expr("failed_login_attempts + 1")).Error
}

func (r *userRepository) LockAccount(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).
		UpdateColumn("locked_at", gorm.Expr("NOW()")).Error
}

func (r *userRepository) ResetFailedAttempts(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).
		UpdateColumn("failed_login_attempts", 0).Error
}

func (r *userRepository) Unlock(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Updates(map[string]any{
		"locked_at":             nil,
		"failed_login_attempts": 0,
	}).Error
}

func (r *userRepository) UpdatePassword(ctx context.Context, userID uint, newHash string) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).
		Update("password_hash", newHash).Error
}
