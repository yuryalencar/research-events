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

func (r *eventRepository) FindDeadlineByID(ctx context.Context, eventID, deadlineID uint) (model.Deadline, error) {
	var deadline model.Deadline
	err := r.db.WithContext(ctx).Where("event_id = ?", eventID).First(&deadline, deadlineID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Deadline{}, ErrNotFound
	}
	return deadline, err
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

func (r *eventRepository) CancelDeadline(ctx context.Context, event model.Event, deadline model.Deadline, submitter model.User) (model.Event, error) {
	oldLastUpdatedByID := event.LastUpdatedByID

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		user, err := findOrCreateSubmitter(tx, submitter)
		if err != nil {
			return err
		}

		if err := tx.Model(&model.Deadline{}).Where("id = ?", deadline.ID).
			Update("is_active", false).Error; err != nil {
			return err
		}

		if err := tx.Model(&model.Event{}).Where("id = ?", event.ID).
			Update("last_updated_by_id", user.ID).Error; err != nil {
			return err
		}

		auditLogs, err := buildCancelDeadlineAuditLogs(event, deadline, oldLastUpdatedByID, user.ID)
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

func (r *eventRepository) SupersedeDeadline(ctx context.Context, event model.Event, oldDeadline model.Deadline, newDeadline model.Deadline, submitter model.User) (model.Event, error) {
	oldLastUpdatedByID := event.LastUpdatedByID

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		user, err := findOrCreateSubmitter(tx, submitter)
		if err != nil {
			return err
		}

		newDeadline.EventID = event.ID
		newDeadline.CreatedByID = user.ID
		if err := tx.Create(&newDeadline).Error; err != nil {
			return err
		}

		if err := tx.Model(&model.Deadline{}).Where("id = ?", oldDeadline.ID).
			Updates(map[string]any{"is_active": false, "superseded_by_id": newDeadline.ID}).Error; err != nil {
			return err
		}

		if err := tx.Model(&model.Event{}).Where("id = ?", event.ID).
			Update("last_updated_by_id", user.ID).Error; err != nil {
			return err
		}

		auditLogs, err := buildSupersedeDeadlineAuditLogs(event, oldDeadline, newDeadline, oldLastUpdatedByID, user.ID)
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
// preloaded and Deadlines containing every is_active=true row plus any
// is_active=false row that has superseded_by_id set — the shape
// AddDeadlines, CancelDeadline, and SupersedeDeadline return to the handler.
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

	updatedLog, err := buildLastUpdatedByIDAuditLog(event.ID, oldLastUpdatedByID, newLastUpdatedByID)
	if err != nil {
		return nil, err
	}
	logs = append(logs, updatedLog)

	return logs, nil
}

// buildLastUpdatedByIDAuditLog returns the "updated" AuditLog row that records
// an event.LastUpdatedByID change — written alongside every operation that
// touches an event's deadlines (AddDeadlines, CancelDeadline), per CLAUDE.md's
// "any update writes an AuditLog row and updates last_updated_by_id" rule.
func buildLastUpdatedByIDAuditLog(eventID uint, oldLastUpdatedByID, newLastUpdatedByID uint) (model.AuditLog, error) {
	diff, err := json.Marshal(map[string]any{
		"last_updated_by_id": map[string]any{"before": oldLastUpdatedByID, "after": newLastUpdatedByID},
	})
	if err != nil {
		return model.AuditLog{}, err
	}
	return model.AuditLog{
		EntityType:  model.AuditEntityEvent,
		EntityID:    eventID,
		Action:      model.AuditActionUpdated,
		ChangedByID: newLastUpdatedByID,
		Diff:        model.JSONB(diff),
	}, nil
}

// buildCancelDeadlineAuditLogs returns the two AuditLog rows for CancelDeadline:
// one deadline_cancelled row for the cancelled Deadline, and one "updated" row
// for the event.LastUpdatedByID change.
func buildCancelDeadlineAuditLogs(event model.Event, deadline model.Deadline, oldLastUpdatedByID, newLastUpdatedByID uint) ([]model.AuditLog, error) {
	updatedLog, err := buildLastUpdatedByIDAuditLog(event.ID, oldLastUpdatedByID, newLastUpdatedByID)
	if err != nil {
		return nil, err
	}
	return []model.AuditLog{
		{
			EntityType:  model.AuditEntityDeadline,
			EntityID:    deadline.ID,
			Action:      model.AuditActionDeadlineCancelled,
			ChangedByID: newLastUpdatedByID,
		},
		updatedLog,
	}, nil
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

// buildSupersedeDeadlineAuditLogs returns the two AuditLog rows for
// SupersedeDeadline: one deadline_superseded row for the superseded
// Deadline (with a before/after diff of the fields that changed), and one
// "updated" row for the event.LastUpdatedByID change.
func buildSupersedeDeadlineAuditLogs(event model.Event, oldDeadline, newDeadline model.Deadline, oldLastUpdatedByID, newLastUpdatedByID uint) ([]model.AuditLog, error) {
	diff, err := json.Marshal(buildDeadlineSupersedeDiff(oldDeadline, newDeadline))
	if err != nil {
		return nil, err
	}

	updatedLog, err := buildLastUpdatedByIDAuditLog(event.ID, oldLastUpdatedByID, newLastUpdatedByID)
	if err != nil {
		return nil, err
	}

	return []model.AuditLog{
		{
			EntityType:  model.AuditEntityDeadline,
			EntityID:    oldDeadline.ID,
			Action:      model.AuditActionDeadlineSuperseded,
			ChangedByID: newLastUpdatedByID,
			Diff:        model.JSONB(diff),
		},
		updatedLog,
	}, nil
}

// buildDeadlineSupersedeDiff records the before/after values of the fields
// that change when oldDeadline is superseded by newDeadline, per
// events-deadlines-supersede.yaml's AuditLog rule.
func buildDeadlineSupersedeDiff(old, newDeadline model.Deadline) map[string]any {
	return map[string]any{
		"date":             map[string]any{"before": old.Date, "after": newDeadline.Date},
		"time":             map[string]any{"before": old.Time, "after": newDeadline.Time},
		"timezone":         map[string]any{"before": old.Timezone, "after": newDeadline.Timezone},
		"is_active":        map[string]any{"before": old.IsActive, "after": false},
		"superseded_by_id": map[string]any{"before": old.SupersededByID, "after": newDeadline.ID},
	}
}
