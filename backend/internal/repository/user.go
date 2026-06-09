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
	UpdateTokens(ctx context.Context, userID uint, jti string, jtiExp time.Time, refreshHash string, refreshExp time.Time) error
	ClearTokens(ctx context.Context, userID uint) error
	IncrementFailedAttempts(ctx context.Context, userID uint) error
	LockAccount(ctx context.Context, userID uint) error
	ResetFailedAttempts(ctx context.Context, userID uint) error
	Unlock(ctx context.Context, userID uint) error
}

// --- Types ---

type userRepository struct {
	db *gorm.DB
}

// --- Constructor ---

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// --- Public methods ---

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
