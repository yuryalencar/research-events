package repository

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/yuryalencar/research-events/internal/model"
)

// --- Public methods ---

func (r *eventRepository) Review(ctx context.Context, updated model.Event, auditLog model.AuditLog) (model.Event, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// updated.CreatedBy/LastUpdatedBy/Deadlines are still populated from the
		// reload that produced "existing" (see ApplyReview's caller). Without
		// Omit(clause.Associations), GORM would re-save those associations and
		// overwrite our explicit LastUpdatedByID with the association struct's ID.
		if err := tx.Omit(clause.Associations).Save(&updated).Error; err != nil {
			return err
		}
		return tx.Create(&auditLog).Error
	})
	if err != nil {
		return model.Event{}, err
	}

	return r.findByIDWithActiveDeadlines(ctx, updated.ID)
}
