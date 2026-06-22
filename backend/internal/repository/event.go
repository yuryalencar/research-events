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

	// ListEvents returns events matching filter, sorted by start_date ascending,
	// along with the total count of matching rows (independent of pagination).
	// Each event's CreatedBy and LastUpdatedBy are preloaded, and Deadlines
	// includes only is_active=true rows.
	ListEvents(ctx context.Context, filter ListEventsFilter) ([]model.Event, int64, error)

	// FindByID returns the event with the given ID, regardless of status.
	// Returns ErrNotFound if no event has that ID.
	FindByID(ctx context.Context, id uint) (model.Event, error)

	// FindDeadlineByID returns the deadline with the given ID if it belongs to
	// eventID. Returns ErrNotFound if no deadline has that ID, or if it exists
	// but belongs to a different event.
	FindDeadlineByID(ctx context.Context, eventID, deadlineID uint) (model.Deadline, error)

	// AddDeadlines persists new Deadlines on event, all within a single transaction:
	// find-or-create submitter (same rules as Submit), insert deadlines with
	// CreatedByID=submitter's ID, update event.LastUpdatedByID to submitter's ID, and
	// write the AuditLog row(s) for auditAction (deadline_added or
	// batch_deadlines_added) plus an "updated" row for the LastUpdatedByID change.
	// Returns the event reloaded with CreatedBy/LastUpdatedBy preloaded and
	// Deadlines containing all is_active=true rows (old and new).
	AddDeadlines(ctx context.Context, event model.Event, deadlines []model.Deadline, submitter model.User, auditAction model.AuditAction) (model.Event, error)

	// CancelDeadline marks deadline as is_active=false (superseded_by_id left nil)
	// within a single transaction: find-or-create submitter (same rules as
	// Submit/AddDeadlines), update event.LastUpdatedByID to submitter's ID, and
	// write a deadline_cancelled AuditLog row plus an "updated" row for the
	// LastUpdatedByID change. Returns the event reloaded with CreatedBy/LastUpdatedBy
	// preloaded and Deadlines containing all remaining is_active=true rows.
	CancelDeadline(ctx context.Context, event model.Event, deadline model.Deadline, submitter model.User) (model.Event, error)

	// SupersedeDeadline replaces oldDeadline with newDeadline within a single
	// transaction: find-or-create submitter (same rules as Submit/AddDeadlines/
	// CancelDeadline), insert newDeadline (is_active=true, superseded_by_id=nil),
	// update oldDeadline (is_active=false, superseded_by_id=<newDeadline's ID>),
	// update event.LastUpdatedByID to submitter's ID, and write a
	// deadline_superseded AuditLog row (entity=oldDeadline, diff records
	// before/after for date/time/timezone/is_active/superseded_by_id) plus an
	// "updated" row for the LastUpdatedByID change. Returns the event reloaded
	// with CreatedBy/LastUpdatedBy preloaded and Deadlines containing every
	// is_active=true row plus any is_active=false row that has
	// superseded_by_id set (i.e. oldDeadline and newDeadline both appear).
	SupersedeDeadline(ctx context.Context, event model.Event, oldDeadline model.Deadline, newDeadline model.Deadline, submitter model.User) (model.Event, error)

	// Review persists updated (the full reviewed Event — status plus any edited
	// fields and last_updated_by_id, as built by service.ApplyReview) and writes
	// auditLog, all within a single transaction. Returns the event reloaded with
	// CreatedBy/LastUpdatedBy preloaded and Deadlines containing all
	// is_active=true rows, per admin-events-review.yaml's 200 response
	// (eventListItemResponse shape).
	Review(ctx context.Context, updated model.Event, auditLog model.AuditLog) (model.Event, error)
}

// ListEventsFilter groups the filters for EventRepository.ListEvents, mirroring
// the validated fields of service.ListEventsInput.
type ListEventsFilter struct {
	Year    *int // nil = no year constraint; non-nil = events.year >= *Year
	Status  model.EventStatus
	Domain  *string
	Country *string
	Tier    *string
	BBox    *BBoxFilter

	// FirstDeadlineMonth, when set, restricts results to events whose earliest
	// is_active=true deadline of type "abstract" or "paper" falls in this
	// calendar month of Year.
	FirstDeadlineMonth *int

	// Page and PageSize are 1-indexed; ignored when PaginationOff is true.
	Page          int
	PageSize      int
	PaginationOff bool
}

// BBoxFilter is the parsed map-viewport bounding box. Bounds are inclusive on
// all four edges.
type BBoxFilter struct {
	MinLng, MinLat, MaxLng, MaxLat float64
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

func (r *eventRepository) ListEvents(ctx context.Context, filter ListEventsFilter) ([]model.Event, int64, error) {
	var total int64
	if err := applyListEventsFilters(r.db.WithContext(ctx).Model(&model.Event{}), filter).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query := preloadEventAssociations(applyListEventsFilters(r.db.WithContext(ctx), filter)).
		Order("start_date ASC")

	if !filter.PaginationOff {
		query = query.Limit(filter.PageSize).Offset((filter.Page - 1) * filter.PageSize)
	}

	var events []model.Event
	if err := query.Find(&events).Error; err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

// --- Private functions ---

// applyListEventsFilters applies the WHERE clauses shared by ListEvents' count
// and select queries. db must not yet have Preload/Order/Find called on it —
// returning a fresh *gorm.DB from each call site keeps the count query and the
// select query independent (avoiding state leaking between them).
func applyListEventsFilters(db *gorm.DB, filter ListEventsFilter) *gorm.DB {
	db = db.Where("status = ?", filter.Status)
	if filter.Year != nil {
		db = db.Where("year >= ?", *filter.Year)
	}
	if filter.Domain != nil {
		db = db.Where("domain = ?", *filter.Domain)
	}
	if filter.Country != nil {
		db = db.Where("LOWER(country) = LOWER(?)", *filter.Country)
	}
	if filter.Tier != nil {
		db = db.Where("tier = ?", *filter.Tier)
	}
	if filter.BBox != nil {
		db = db.Where("longitude BETWEEN ? AND ? AND latitude BETWEEN ? AND ?",
			filter.BBox.MinLng, filter.BBox.MaxLng, filter.BBox.MinLat, filter.BBox.MaxLat)
	}
	if filter.FirstDeadlineMonth != nil {
		if filter.Year != nil {
			db = db.Where(`EXISTS (
				SELECT 1 FROM deadlines d
				WHERE d.event_id = events.id
					AND d.is_active = true
					AND d.type IN ('abstract', 'paper')
					AND d.date = (
						SELECT MIN(d2.date) FROM deadlines d2
						WHERE d2.event_id = events.id
							AND d2.is_active = true
							AND d2.type IN ('abstract', 'paper')
					)
					AND EXTRACT(MONTH FROM d.date) = ?
					AND EXTRACT(YEAR FROM d.date) >= ?
			)`, *filter.FirstDeadlineMonth, *filter.Year)
		} else {
			db = db.Where(`EXISTS (
				SELECT 1 FROM deadlines d
				WHERE d.event_id = events.id
					AND d.is_active = true
					AND d.type IN ('abstract', 'paper')
					AND d.date = (
						SELECT MIN(d2.date) FROM deadlines d2
						WHERE d2.event_id = events.id
							AND d2.is_active = true
							AND d2.type IN ('abstract', 'paper')
					)
					AND EXTRACT(MONTH FROM d.date) = ?
			)`, *filter.FirstDeadlineMonth)
		}
	}
	return db
}

// preloadEventAssociations adds the CreatedBy/LastUpdatedBy/Deadlines
// preloads shared by every query that returns a full Event to the API —
// ListEvents, AddDeadlines's reload, CancelDeadline's reload, and
// SupersedeDeadline's reload all build on this. Deadlines includes every
// is_active=true row plus any is_active=false row that has
// superseded_by_id set, so a just-superseded deadline is returned alongside
// its replacement (per events-deadlines-supersede.yaml) while a plain
// cancelled deadline (is_active=false, superseded_by_id=NULL) stays hidden.
func preloadEventAssociations(db *gorm.DB) *gorm.DB {
	return db.
		Preload("CreatedBy").
		Preload("LastUpdatedBy").
		Preload("Deadlines", "is_active = ? OR superseded_by_id IS NOT NULL", true)
}

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
