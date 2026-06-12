package repository

//go:generate mockgen -source=event.go -destination=mocks/mock_event.go -package=mocks

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/yuryalencar/research-events/internal/model"
)

// --- Interface ---

// EventRepository defines persistence operations for the Event model.
type EventRepository interface {
	// FindActiveBySlug returns the event with the given slug if it has
	// status=pending or status=approved. Returns ErrNotFound otherwise —
	// including when the only match has status=rejected, since a rejected
	// event's slug is free for reuse (see events-submit.yaml).
	FindActiveBySlug(ctx context.Context, slug string) (model.Event, error)

	// Submit persists a new Event (with its Deadlines) and writes the audit
	// trail for the submission, all within a single transaction.
	//
	// submitter is looked up by email: if a User with that email already exists
	// (any role), it is reused and its name is updated to submitter.Name; otherwise
	// submitter is inserted as-is (role=contributor, password_hash=nil — see
	// service.BuildSubmitterFromInput). The resulting user becomes
	// event.CreatedByID / LastUpdatedByID and each deadline's CreatedByID.
	Submit(ctx context.Context, event model.Event, deadlines []model.Deadline, submitter model.User) (model.Event, error)
}

// --- Types ---

type eventRepository struct {
	db *gorm.DB
}

// --- Constructor ---

func NewEventRepository(db *gorm.DB) EventRepository {
	return &eventRepository{db: db}
}

// --- Public methods ---

func (r *eventRepository) FindActiveBySlug(ctx context.Context, slug string) (model.Event, error) {
	var event model.Event
	err := r.db.WithContext(ctx).
		Where("slug = ? AND status IN ?", slug, []model.EventStatus{model.EventStatusPending, model.EventStatusApproved}).
		First(&event).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Event{}, ErrNotFound
	}
	return event, err
}

func (r *eventRepository) Submit(ctx context.Context, event model.Event, deadlines []model.Deadline, submitter model.User) (model.Event, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		user, err := findOrCreateSubmitter(tx, submitter)
		if err != nil {
			return err
		}

		event.CreatedByID = user.ID
		event.LastUpdatedByID = user.ID
		for i := range deadlines {
			deadlines[i].CreatedByID = user.ID
		}
		event.Deadlines = deadlines

		// Creating event with Deadlines set inserts both in this transaction —
		// GORM populates event.ID and each deadline's EventID/ID automatically.
		if err := tx.Create(&event).Error; err != nil {
			return err
		}

		auditLogs := buildSubmissionAuditLogs(event, user.ID)
		if err := tx.Create(&auditLogs).Error; err != nil {
			return err
		}

		event.CreatedBy = user
		event.LastUpdatedBy = user
		return nil
	})
	if err != nil {
		return model.Event{}, err
	}
	return event, nil
}

// --- Private functions ---

// findOrCreateSubmitter looks up a User by email within tx. If found, its name is
// updated to submitter.Name and the existing record (with its original role) is
// returned. If not found, submitter is inserted as-is.
func findOrCreateSubmitter(tx *gorm.DB, submitter model.User) (model.User, error) {
	var user model.User
	err := tx.Where("email = ?", submitter.Email).First(&user).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		if err := tx.Create(&submitter).Error; err != nil {
			return model.User{}, err
		}
		return submitter, nil
	case err != nil:
		return model.User{}, err
	default:
		if err := tx.Model(&user).Update("name", submitter.Name).Error; err != nil {
			return model.User{}, err
		}
		user.Name = submitter.Name
		return user, nil
	}
}

// buildSubmissionAuditLogs returns one AuditLog for the event creation and one
// per deadline created — written together with the event/deadlines in Submit.
func buildSubmissionAuditLogs(event model.Event, byUserID uint) []model.AuditLog {
	logs := []model.AuditLog{{
		EntityType:  model.AuditEntityEvent,
		EntityID:    event.ID,
		Action:      model.AuditActionCreated,
		ChangedByID: byUserID,
	}}
	for _, d := range event.Deadlines {
		logs = append(logs, model.AuditLog{
			EntityType:  model.AuditEntityDeadline,
			EntityID:    d.ID,
			Action:      model.AuditActionDeadlineAdded,
			ChangedByID: byUserID,
		})
	}
	return logs
}
