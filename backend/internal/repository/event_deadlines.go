package repository

import (
	"context"
	"encoding/json"
	"errors"

	"gorm.io/gorm"

	"github.com/yuryalencar/research-events/internal/model"
)

// --- Public methods ---

func (r *eventRepository) FindByID(ctx context.Context, id uint) (model.Event, error) {
	var event model.Event
	err := r.db.WithContext(ctx).First(&event, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Event{}, ErrNotFound
	}
	return event, err
}

func (r *eventRepository) AddDeadlines(ctx context.Context, event model.Event, deadlines []model.Deadline, submitter model.User, auditAction model.AuditAction) (model.Event, error) {
	oldLastUpdatedByID := event.LastUpdatedByID

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		user, err := findOrCreateSubmitter(tx, submitter)
		if err != nil {
			return err
		}

		for i := range deadlines {
			deadlines[i].EventID = event.ID
			deadlines[i].CreatedByID = user.ID
		}
		if err := tx.Create(&deadlines).Error; err != nil {
			return err
		}

		if err := tx.Model(&model.Event{}).Where("id = ?", event.ID).
			Update("last_updated_by_id", user.ID).Error; err != nil {
			return err
		}

		auditLogs, err := buildAddDeadlinesAuditLogs(event, deadlines, auditAction, oldLastUpdatedByID, user.ID)
		if err != nil {
			return err
		}
		return tx.Create(&auditLogs).Error
	})
	if err != nil {
		return model.Event{}, err
	}

	return r.findByIDWithActiveDeadlines(ctx, event.ID)
}

// --- Private functions ---

// findByIDWithActiveDeadlines reloads an event with CreatedBy/LastUpdatedBy
// preloaded and Deadlines containing only is_active=true rows — the shape
// AddDeadlines returns to the handler.
func (r *eventRepository) findByIDWithActiveDeadlines(ctx context.Context, id uint) (model.Event, error) {
	var event model.Event
	err := preloadEventAssociations(r.db.WithContext(ctx)).First(&event, id).Error
	return event, err
}

// buildAddDeadlinesAuditLogs returns the AuditLog rows for AddDeadlines:
//   - auditAction == deadline_added: one row for the single new Deadline
//   - auditAction == batch_deadlines_added: one row for the event, with a diff
//     listing every new deadline's id/type/description/date
//   - always: one "updated" row for the event recording the
//     last_updated_by_id change, per CLAUDE.md's "any update writes an
//     AuditLog row and updates last_updated_by_id" rule.
func buildAddDeadlinesAuditLogs(event model.Event, deadlines []model.Deadline, auditAction model.AuditAction, oldLastUpdatedByID, newLastUpdatedByID uint) ([]model.AuditLog, error) {
	logs := make([]model.AuditLog, 0, 2)

	if auditAction == model.AuditActionBatchDeadlinesAdded {
		diff, err := json.Marshal(buildBatchDeadlinesDiff(deadlines))
		if err != nil {
			return nil, err
		}
		logs = append(logs, model.AuditLog{
			EntityType:  model.AuditEntityEvent,
			EntityID:    event.ID,
			Action:      model.AuditActionBatchDeadlinesAdded,
			ChangedByID: newLastUpdatedByID,
			Diff:        model.JSONB(diff),
		})
	} else {
		logs = append(logs, model.AuditLog{
			EntityType:  model.AuditEntityDeadline,
			EntityID:    deadlines[0].ID,
			Action:      model.AuditActionDeadlineAdded,
			ChangedByID: newLastUpdatedByID,
		})
	}

	updatedDiff, err := json.Marshal(map[string]any{
		"last_updated_by_id": map[string]any{"before": oldLastUpdatedByID, "after": newLastUpdatedByID},
	})
	if err != nil {
		return nil, err
	}
	logs = append(logs, model.AuditLog{
		EntityType:  model.AuditEntityEvent,
		EntityID:    event.ID,
		Action:      model.AuditActionUpdated,
		ChangedByID: newLastUpdatedByID,
		Diff:        model.JSONB(updatedDiff),
	})

	return logs, nil
}

// buildBatchDeadlinesDiff summarizes the newly created deadlines for a
// batch_deadlines_added AuditLog row.
func buildBatchDeadlinesDiff(deadlines []model.Deadline) map[string]any {
	items := make([]map[string]any, 0, len(deadlines))
	for _, d := range deadlines {
		items = append(items, map[string]any{
			"id":          d.ID,
			"type":        d.Type,
			"description": d.Description,
			"date":        d.Date,
		})
	}
	return map[string]any{"deadlines_added": items}
}
