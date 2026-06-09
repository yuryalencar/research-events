package repository

//go:generate mockgen -source=audit.go -destination=mocks/mock_audit.go -package=mocks

import (
	"context"

	"gorm.io/gorm"

	"github.com/yuryalencar/research-events/internal/model"
)

// --- Interface ---

// AuditRepository defines persistence operations for the AuditLog model.
type AuditRepository interface {
	Create(ctx context.Context, log model.AuditLog) error
}

// --- Types ---

type auditRepository struct {
	db *gorm.DB
}

// --- Constructor ---

func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &auditRepository{db: db}
}

// --- Public methods ---

func (r *auditRepository) Create(ctx context.Context, log model.AuditLog) error {
	return r.db.WithContext(ctx).Create(&log).Error
}
