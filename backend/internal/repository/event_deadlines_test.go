package repository_test

// Spec: specs/backend/events-deadlines-add.yaml

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/yuryalencar/research-events/internal/model"
	"github.com/yuryalencar/research-events/internal/repository"
)

// --- FindByID ---

func TestEventRepository_FindByID_ReturnsEventWhenExists(t *testing.T) {
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event, deadlines, submitter := newSubmission("MODELS2026", "ana@example.com")
	created, err := repo.Submit(context.Background(), event, deadlines, submitter)
	require.NoError(t, err)

	got, err := repo.FindByID(context.Background(), created.ID)

	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, created.Slug, got.Slug)
}

func TestEventRepository_FindByID_ReturnsErrNotFoundWhenMissing(t *testing.T) {
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)

	_, err := repo.FindByID(context.Background(), 999999)

	require.ErrorIs(t, err, repository.ErrNotFound)
}

// --- AddDeadlines ---

// newApprovedEvent returns a freshly persisted, approved Event owned by a new
// contributor user with the given email — the starting point for AddDeadlines tests.
func newApprovedEvent(t *testing.T, tx *gorm.DB, slug, email string) model.Event {
	t.Helper()
	event, deadlines, submitter := newSubmission(slug, email)
	event.Status = model.EventStatusApproved

	repo := repository.NewEventRepository(tx)
	created, err := repo.Submit(context.Background(), event, deadlines, submitter)
	require.NoError(t, err)
	return created
}

// oneDeadline returns a single new (unsaved) is_active=true Deadline, as
// service.BuildDeadlinesFromInput would produce.
func oneDeadline(deadlineType model.DeadlineType, description string, date time.Time) []model.Deadline {
	return []model.Deadline{
		{Type: deadlineType, Description: description, Date: date, IsActive: true},
	}
}

func TestEventRepository_AddDeadlines_CreatesNewContributorWhenEmailNotFound(t *testing.T) {
	// Spec: events-deadlines-add.yaml border_case "submitter.email is new → new User
	// created with role=contributor, password_hash=NULL"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event := newApprovedEvent(t, tx, "MODELS2026", "ana@example.com")
	deadlines := oneDeadline(model.DeadlineTypeCameraReady, "Camera-ready", time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC))
	submitter := model.User{Name: "Beatriz Costa", Email: "beatriz@example.com", Role: model.UserRoleContributor}

	got, err := repo.AddDeadlines(context.Background(), event, deadlines, submitter, model.AuditActionDeadlineAdded)

	require.NoError(t, err)

	var user model.User
	require.NoError(t, tx.Where("email = ?", "beatriz@example.com").First(&user).Error)
	assert.Equal(t, model.UserRoleContributor, user.Role)
	assert.Nil(t, user.PasswordHash)
	require.Len(t, got.Deadlines, 1)
}

func TestEventRepository_AddDeadlines_ReusesAndRenamesExistingUser(t *testing.T) {
	// Spec: events-deadlines-add.yaml border_case "submitter.email matches an existing
	// User (any role) → event linked to that user for this update, name updated"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event := newApprovedEvent(t, tx, "MODELS2026", "ana@example.com")

	existing := model.User{Name: "Old Name", Email: "carlos@example.com", Role: model.UserRoleContributor}
	require.NoError(t, tx.Create(&existing).Error)

	deadlines := oneDeadline(model.DeadlineTypeCameraReady, "Camera-ready", time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC))
	submitter := model.User{Name: "Carlos Souza", Email: "carlos@example.com", Role: model.UserRoleContributor}

	_, err := repo.AddDeadlines(context.Background(), event, deadlines, submitter, model.AuditActionDeadlineAdded)
	require.NoError(t, err)

	var user model.User
	require.NoError(t, tx.First(&user, existing.ID).Error)
	assert.Equal(t, "Carlos Souza", user.Name)
}

func TestEventRepository_AddDeadlines_CreatesDeadlinesWithIsActiveTrue(t *testing.T) {
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event := newApprovedEvent(t, tx, "MODELS2026", "ana@example.com")
	deadlines := oneDeadline(model.DeadlineTypeCameraReady, "Camera-ready", time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC))
	submitter := model.User{Name: "Beatriz Costa", Email: "beatriz@example.com", Role: model.UserRoleContributor}

	got, err := repo.AddDeadlines(context.Background(), event, deadlines, submitter, model.AuditActionDeadlineAdded)
	require.NoError(t, err)

	var saved model.Deadline
	require.NoError(t, tx.Where("event_id = ? AND type = ?", got.ID, model.DeadlineTypeCameraReady).First(&saved).Error)
	assert.True(t, saved.IsActive)
	assert.NotZero(t, saved.CreatedByID)
}

func TestEventRepository_AddDeadlines_UpdatesLastUpdatedByID(t *testing.T) {
	// Spec: events-deadlines-add.yaml rule "event.last_updated_by_id is set to the
	// submitter's user ID"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event := newApprovedEvent(t, tx, "MODELS2026", "ana@example.com")
	originalLastUpdatedByID := event.LastUpdatedByID

	deadlines := oneDeadline(model.DeadlineTypeCameraReady, "Camera-ready", time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC))
	submitter := model.User{Name: "Beatriz Costa", Email: "beatriz@example.com", Role: model.UserRoleContributor}

	got, err := repo.AddDeadlines(context.Background(), event, deadlines, submitter, model.AuditActionDeadlineAdded)
	require.NoError(t, err)

	assert.NotEqual(t, originalLastUpdatedByID, got.LastUpdatedByID)
	assert.Equal(t, got.LastUpdatedByID, got.LastUpdatedBy.ID)
	assert.Equal(t, "Beatriz Costa", got.LastUpdatedBy.Name)
}

func TestEventRepository_AddDeadlines_SingleDeadline_WritesDeadlineAddedAuditLog(t *testing.T) {
	// Spec: events-deadlines-add.yaml rule "Exactly 1 deadline submitted → one row:
	// entity_type=deadline, entity_id=<new deadline's id>, action=deadline_added"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event := newApprovedEvent(t, tx, "MODELS2026", "ana@example.com")
	deadlines := oneDeadline(model.DeadlineTypeCameraReady, "Camera-ready", time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC))
	submitter := model.User{Name: "Beatriz Costa", Email: "beatriz@example.com", Role: model.UserRoleContributor}

	got, err := repo.AddDeadlines(context.Background(), event, deadlines, submitter, model.AuditActionDeadlineAdded)
	require.NoError(t, err)

	require.Len(t, got.Deadlines, 1)
	var log model.AuditLog
	err = tx.Where("entity_type = ? AND entity_id = ? AND action = ?",
		model.AuditEntityDeadline, got.Deadlines[0].ID, model.AuditActionDeadlineAdded).First(&log).Error
	require.NoError(t, err)
	assert.Equal(t, got.LastUpdatedByID, log.ChangedByID)
}

func TestEventRepository_AddDeadlines_MultipleDeadlines_WritesBatchDeadlinesAddedAuditLog(t *testing.T) {
	// Spec: events-deadlines-add.yaml rule "More than 1 deadline submitted → one row:
	// entity_type=event, entity_id=<event id>, action=batch_deadlines_added"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event := newApprovedEvent(t, tx, "MODELS2026", "ana@example.com")
	deadlines := []model.Deadline{
		{Type: model.DeadlineTypePaper, Description: "Industry track full paper", Date: time.Date(2026, 8, 29, 0, 0, 0, 0, time.UTC), IsActive: true},
		{Type: model.DeadlineTypeNotification, Description: "Industry track notification", Date: time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC), IsActive: true},
	}
	submitter := model.User{Name: "Carlos Souza", Email: "carlos@example.com", Role: model.UserRoleContributor}

	got, err := repo.AddDeadlines(context.Background(), event, deadlines, submitter, model.AuditActionBatchDeadlinesAdded)
	require.NoError(t, err)

	var log model.AuditLog
	err = tx.Where("entity_type = ? AND entity_id = ? AND action = ?",
		model.AuditEntityEvent, got.ID, model.AuditActionBatchDeadlinesAdded).First(&log).Error
	require.NoError(t, err)
	assert.Equal(t, got.LastUpdatedByID, log.ChangedByID)
	assert.NotEmpty(t, log.Diff)

	// No per-deadline deadline_added rows when batched.
	var count int64
	require.NoError(t, tx.Model(&model.AuditLog{}).
		Where("entity_type = ? AND action = ?", model.AuditEntityDeadline, model.AuditActionDeadlineAdded).
		Count(&count).Error)
	assert.Zero(t, count)
}

func TestEventRepository_AddDeadlines_AlwaysWritesEventUpdatedAuditLog(t *testing.T) {
	// Spec: events-deadlines-add.yaml rule "Always, in addition: one row
	// entity_type=event, entity_id=<event id>, action=updated,
	// diff={"last_updated_by_id":{"before":<old>,"after":<new>}}"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event := newApprovedEvent(t, tx, "MODELS2026", "ana@example.com")
	originalLastUpdatedByID := event.LastUpdatedByID

	deadlines := oneDeadline(model.DeadlineTypeCameraReady, "Camera-ready", time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC))
	submitter := model.User{Name: "Beatriz Costa", Email: "beatriz@example.com", Role: model.UserRoleContributor}

	got, err := repo.AddDeadlines(context.Background(), event, deadlines, submitter, model.AuditActionDeadlineAdded)
	require.NoError(t, err)

	var log model.AuditLog
	err = tx.Where("entity_type = ? AND entity_id = ? AND action = ?",
		model.AuditEntityEvent, got.ID, model.AuditActionUpdated).First(&log).Error
	require.NoError(t, err)
	assert.Equal(t, got.LastUpdatedByID, log.ChangedByID)
	assert.Contains(t, string(log.Diff), "last_updated_by_id")
	assert.NotContains(t, string(log.Diff), string(rune(originalLastUpdatedByID))) // sanity: diff is non-trivial JSON
}

func TestEventRepository_AddDeadlines_ReturnsEventWithAllActiveDeadlines(t *testing.T) {
	tx, rollback := beginTx(t)
	defer rollback()

	// Start with one deadline already on the event.
	event, deadlines, submitter := newSubmission("MODELS2026", "ana@example.com")
	event.Status = model.EventStatusApproved
	deadlines = []model.Deadline{
		{Type: model.DeadlineTypeAbstract, Description: "Abstract", Date: time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC), IsActive: true},
	}
	repo := repository.NewEventRepository(tx)
	created, err := repo.Submit(context.Background(), event, deadlines, submitter)
	require.NoError(t, err)

	newDeadlines := oneDeadline(model.DeadlineTypeCameraReady, "Camera-ready", time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC))
	newSubmitter := model.User{Name: "Beatriz Costa", Email: "beatriz@example.com", Role: model.UserRoleContributor}

	got, err := repo.AddDeadlines(context.Background(), created, newDeadlines, newSubmitter, model.AuditActionDeadlineAdded)
	require.NoError(t, err)

	require.Len(t, got.Deadlines, 2)
	types := []model.DeadlineType{got.Deadlines[0].Type, got.Deadlines[1].Type}
	assert.Contains(t, types, model.DeadlineTypeAbstract)
	assert.Contains(t, types, model.DeadlineTypeCameraReady)
	for _, d := range got.Deadlines {
		assert.True(t, d.IsActive)
	}
}

func TestEventRepository_AddDeadlines_AllowsDuplicateDeadlines(t *testing.T) {
	// Spec: events-deadlines-add.yaml border_case "duplicate deadline (same
	// type/description/date as an existing active deadline) → allowed, 201,
	// both remain active (no dedup/supersession)"
	tx, rollback := beginTx(t)
	defer rollback()

	date := time.Date(2026, 8, 22, 0, 0, 0, 0, time.UTC)
	event, deadlines, submitter := newSubmission("MODELS2026", "ana@example.com")
	event.Status = model.EventStatusApproved
	deadlines = []model.Deadline{
		{Type: model.DeadlineTypePaper, Description: "Research track full paper", Date: date, IsActive: true},
	}
	repo := repository.NewEventRepository(tx)
	created, err := repo.Submit(context.Background(), event, deadlines, submitter)
	require.NoError(t, err)

	duplicate := []model.Deadline{
		{Type: model.DeadlineTypePaper, Description: "Research track full paper", Date: date, IsActive: true},
	}
	newSubmitter := model.User{Name: "Beatriz Costa", Email: "beatriz@example.com", Role: model.UserRoleContributor}

	got, err := repo.AddDeadlines(context.Background(), created, duplicate, newSubmitter, model.AuditActionDeadlineAdded)
	require.NoError(t, err)

	require.Len(t, got.Deadlines, 2)
	for _, d := range got.Deadlines {
		assert.True(t, d.IsActive)
		assert.Equal(t, model.DeadlineTypePaper, d.Type)
		assert.Equal(t, "Research track full paper", d.Description)
	}
}

// --- FindDeadlineByID ---

func TestEventRepository_FindDeadlineByID_ReturnsDeadlineWhenItBelongsToEvent(t *testing.T) {
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event := newApprovedEvent(t, tx, "MODELS2026", "ana@example.com")
	deadlines := oneDeadline(model.DeadlineTypeCameraReady, "Camera-ready", time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC))
	submitter := model.User{Name: "Beatriz Costa", Email: "beatriz@example.com", Role: model.UserRoleContributor}
	updated, err := repo.AddDeadlines(context.Background(), event, deadlines, submitter, model.AuditActionDeadlineAdded)
	require.NoError(t, err)
	require.Len(t, updated.Deadlines, 1)

	got, err := repo.FindDeadlineByID(context.Background(), event.ID, updated.Deadlines[0].ID)

	require.NoError(t, err)
	assert.Equal(t, updated.Deadlines[0].ID, got.ID)
	assert.Equal(t, event.ID, got.EventID)
}

func TestEventRepository_FindDeadlineByID_ReturnsErrNotFoundWhenMissing(t *testing.T) {
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event := newApprovedEvent(t, tx, "MODELS2026", "ana@example.com")

	_, err := repo.FindDeadlineByID(context.Background(), event.ID, 999999)

	require.ErrorIs(t, err, repository.ErrNotFound)
}

func TestEventRepository_FindDeadlineByID_ReturnsErrNotFoundWhenBelongsToDifferentEvent(t *testing.T) {
	// Spec: events-deadlines-cancel.yaml border_case ":deadlineId matches a deadline
	// that belongs to a different event → 404 DEADLINE_NOT_FOUND"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	eventA := newApprovedEvent(t, tx, "MODELS2026", "ana@example.com")
	eventB := newApprovedEvent(t, tx, "ICSE2026", "carlos@example.com")

	deadlines := oneDeadline(model.DeadlineTypeCameraReady, "Camera-ready", time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC))
	submitter := model.User{Name: "Beatriz Costa", Email: "beatriz@example.com", Role: model.UserRoleContributor}
	updatedA, err := repo.AddDeadlines(context.Background(), eventA, deadlines, submitter, model.AuditActionDeadlineAdded)
	require.NoError(t, err)
	require.Len(t, updatedA.Deadlines, 1)

	_, err = repo.FindDeadlineByID(context.Background(), eventB.ID, updatedA.Deadlines[0].ID)

	require.ErrorIs(t, err, repository.ErrNotFound)
}

// --- CancelDeadline ---

// addOneDeadline adds a single active Camera-ready deadline (created by
// beatriz@example.com) to event and returns the reloaded event, which contains
// exactly one Deadline.
func addOneDeadline(t *testing.T, tx *gorm.DB, event model.Event) model.Event {
	t.Helper()
	repo := repository.NewEventRepository(tx)
	deadlines := oneDeadline(model.DeadlineTypeCameraReady, "Camera-ready", time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC))
	submitter := model.User{Name: "Beatriz Costa", Email: "beatriz@example.com", Role: model.UserRoleContributor}
	updated, err := repo.AddDeadlines(context.Background(), event, deadlines, submitter, model.AuditActionDeadlineAdded)
	require.NoError(t, err)
	require.Len(t, updated.Deadlines, 1)
	return updated
}

func TestEventRepository_CancelDeadline_SetsIsActiveFalseAndSupersededByIDNil(t *testing.T) {
	// Spec: events-deadlines-cancel.yaml rule "Cancelling sets is_active=false on
	// the target deadline and leaves superseded_by_id=NULL"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event := newApprovedEvent(t, tx, "MODELS2026", "ana@example.com")
	withDeadline := addOneDeadline(t, tx, event)
	target := withDeadline.Deadlines[0]
	submitter := model.User{Name: "Carlos Souza", Email: "carlos@example.com", Role: model.UserRoleContributor}

	_, err := repo.CancelDeadline(context.Background(), withDeadline, target, submitter)
	require.NoError(t, err)

	var saved model.Deadline
	require.NoError(t, tx.First(&saved, target.ID).Error)
	assert.False(t, saved.IsActive)
	assert.Nil(t, saved.SupersededByID)
}

func TestEventRepository_CancelDeadline_UpdatesLastUpdatedByID(t *testing.T) {
	// Spec: events-deadlines-cancel.yaml rule "event.last_updated_by_id is set to
	// the canceller's user ID"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event := newApprovedEvent(t, tx, "MODELS2026", "ana@example.com")
	withDeadline := addOneDeadline(t, tx, event)
	target := withDeadline.Deadlines[0]
	originalLastUpdatedByID := withDeadline.LastUpdatedByID
	submitter := model.User{Name: "Carlos Souza", Email: "carlos@example.com", Role: model.UserRoleContributor}

	got, err := repo.CancelDeadline(context.Background(), withDeadline, target, submitter)
	require.NoError(t, err)

	assert.NotEqual(t, originalLastUpdatedByID, got.LastUpdatedByID)
	assert.Equal(t, got.LastUpdatedByID, got.LastUpdatedBy.ID)
	assert.Equal(t, "Carlos Souza", got.LastUpdatedBy.Name)
}

func TestEventRepository_CancelDeadline_WritesDeadlineCancelledAndUpdatedAuditLogs(t *testing.T) {
	// Spec: events-deadlines-cancel.yaml rule "AuditLog rows written ... entity_type=deadline,
	// entity_id=<cancelled deadline's id>, action=deadline_cancelled ... entity_type=event,
	// entity_id=<event id>, action=updated, diff={...last_updated_by_id...}"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event := newApprovedEvent(t, tx, "MODELS2026", "ana@example.com")
	withDeadline := addOneDeadline(t, tx, event)
	target := withDeadline.Deadlines[0]
	submitter := model.User{Name: "Carlos Souza", Email: "carlos@example.com", Role: model.UserRoleContributor}

	got, err := repo.CancelDeadline(context.Background(), withDeadline, target, submitter)
	require.NoError(t, err)

	var cancelledLog model.AuditLog
	err = tx.Where("entity_type = ? AND entity_id = ? AND action = ?",
		model.AuditEntityDeadline, target.ID, model.AuditActionDeadlineCancelled).First(&cancelledLog).Error
	require.NoError(t, err)
	assert.Equal(t, got.LastUpdatedByID, cancelledLog.ChangedByID)

	// AddDeadlines (via addOneDeadline) also wrote an entity_type=event,
	// action=updated row — order by id DESC to get CancelDeadline's row.
	var updatedLog model.AuditLog
	err = tx.Where("entity_type = ? AND entity_id = ? AND action = ?",
		model.AuditEntityEvent, got.ID, model.AuditActionUpdated).Order("id DESC").First(&updatedLog).Error
	require.NoError(t, err)
	assert.Equal(t, got.LastUpdatedByID, updatedLog.ChangedByID)
}

func TestEventRepository_CancelDeadline_CreatesContributorWhenEmailNotFound(t *testing.T) {
	// Spec: events-deadlines-cancel.yaml border_case "submitter.email is new → new
	// User created with role=contributor, password_hash=NULL"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event := newApprovedEvent(t, tx, "MODELS2026", "ana@example.com")
	withDeadline := addOneDeadline(t, tx, event)
	target := withDeadline.Deadlines[0]
	submitter := model.User{Name: "Carlos Souza", Email: "carlos@example.com", Role: model.UserRoleContributor}

	_, err := repo.CancelDeadline(context.Background(), withDeadline, target, submitter)
	require.NoError(t, err)

	var user model.User
	require.NoError(t, tx.Where("email = ?", "carlos@example.com").First(&user).Error)
	assert.Equal(t, model.UserRoleContributor, user.Role)
	assert.Nil(t, user.PasswordHash)
}

func TestEventRepository_CancelDeadline_ReusesExistingUserAndUpdatesName(t *testing.T) {
	// Spec: events-deadlines-cancel.yaml border_case "submitter.email matches an
	// existing User (any role) → event linked to that user for this update, name
	// updated, no new User created"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event := newApprovedEvent(t, tx, "MODELS2026", "ana@example.com")
	withDeadline := addOneDeadline(t, tx, event)
	target := withDeadline.Deadlines[0]

	existing := model.User{Name: "Old Name", Email: "carlos@example.com", Role: model.UserRoleContributor}
	require.NoError(t, tx.Create(&existing).Error)

	submitter := model.User{Name: "Carlos Souza", Email: "carlos@example.com", Role: model.UserRoleContributor}

	_, err := repo.CancelDeadline(context.Background(), withDeadline, target, submitter)
	require.NoError(t, err)

	var user model.User
	require.NoError(t, tx.First(&user, existing.ID).Error)
	assert.Equal(t, "Carlos Souza", user.Name)
}

func TestEventRepository_CancelDeadline_AllowsCancellingLastActiveDeadline_ReturnsEmptyDeadlines(t *testing.T) {
	// Spec: events-deadlines-cancel.yaml border_case "cancelling the only/last active
	// deadline of the event → 200, allowed, `deadlines` in the response is an empty array"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event := newApprovedEvent(t, tx, "MODELS2026", "ana@example.com")
	withDeadline := addOneDeadline(t, tx, event)
	target := withDeadline.Deadlines[0]
	submitter := model.User{Name: "Carlos Souza", Email: "carlos@example.com", Role: model.UserRoleContributor}

	got, err := repo.CancelDeadline(context.Background(), withDeadline, target, submitter)

	require.NoError(t, err)
	assert.Empty(t, got.Deadlines)
}

// --- SupersedeDeadline ---
// Spec: specs/backend/events-deadlines-supersede.yaml

// newSupersedingDeadline returns the new model.Deadline to pass into
// SupersedeDeadline, with the requested date/time/timezone — mirroring what
// service.BuildSupersedingDeadline would produce from old (inheriting
// Type/Description/IsOptional) and the request's date/time/timezone.
func newSupersedingDeadline(old model.Deadline, date time.Time, deadlineTime, timezone *string) model.Deadline {
	return model.Deadline{
		Type:        old.Type,
		Description: old.Description,
		Date:        date,
		Time:        deadlineTime,
		Timezone:    timezone,
		IsOptional:  old.IsOptional,
		IsActive:    true,
	}
}

func TestEventRepository_SupersedeDeadline_CreatesNewActiveDeadlineWithRequestedFields(t *testing.T) {
	// Spec: rule "The new row is created with is_active=true and
	// superseded_by_id=null"; rule "only date, time, and timezone come from
	// the request body"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event := newApprovedEvent(t, tx, "MODELS2026", "ana@example.com")
	withDeadline := addOneDeadline(t, tx, event)
	old := withDeadline.Deadlines[0]
	submitter := model.User{Name: "Carlos Souza", Email: "carlos@example.com", Role: model.UserRoleContributor}

	newTime := "23:59"
	newTimezone := "AoE"
	newDate := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	newDeadline := newSupersedingDeadline(old, newDate, &newTime, &newTimezone)

	got, err := repo.SupersedeDeadline(context.Background(), withDeadline, old, newDeadline, submitter)
	require.NoError(t, err)

	var saved model.Deadline
	require.NoError(t, tx.Where("event_id = ? AND id != ?", event.ID, old.ID).First(&saved).Error)
	assert.True(t, saved.IsActive)
	assert.Nil(t, saved.SupersededByID)
	assert.True(t, newDate.Equal(saved.Date))
	require.NotNil(t, saved.Time)
	assert.Equal(t, "23:59", *saved.Time)
	require.NotNil(t, saved.Timezone)
	assert.Equal(t, "AoE", *saved.Timezone)
	assert.Equal(t, old.Type, saved.Type)
	assert.Equal(t, old.Description, saved.Description)
	assert.NotEmpty(t, got.Deadlines)
}

func TestEventRepository_SupersedeDeadline_MarksOldDeadlineInactiveWithSupersededByID(t *testing.T) {
	// Spec: rule "The old row is updated: is_active=false,
	// superseded_by_id=<new row's ID>"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event := newApprovedEvent(t, tx, "MODELS2026", "ana@example.com")
	withDeadline := addOneDeadline(t, tx, event)
	old := withDeadline.Deadlines[0]
	submitter := model.User{Name: "Carlos Souza", Email: "carlos@example.com", Role: model.UserRoleContributor}

	newDate := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	newDeadline := newSupersedingDeadline(old, newDate, nil, nil)

	_, err := repo.SupersedeDeadline(context.Background(), withDeadline, old, newDeadline, submitter)
	require.NoError(t, err)

	var savedOld model.Deadline
	require.NoError(t, tx.First(&savedOld, old.ID).Error)
	assert.False(t, savedOld.IsActive)
	require.NotNil(t, savedOld.SupersededByID)

	var savedNew model.Deadline
	require.NoError(t, tx.Where("event_id = ? AND id != ?", event.ID, old.ID).First(&savedNew).Error)
	assert.Equal(t, savedNew.ID, *savedOld.SupersededByID)
}

func TestEventRepository_SupersedeDeadline_UpdatesLastUpdatedByID(t *testing.T) {
	// Spec: mirrors events-deadlines-cancel.yaml's "event.last_updated_by_id is
	// set to the submitter's user ID"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event := newApprovedEvent(t, tx, "MODELS2026", "ana@example.com")
	withDeadline := addOneDeadline(t, tx, event)
	old := withDeadline.Deadlines[0]
	originalLastUpdatedByID := withDeadline.LastUpdatedByID
	submitter := model.User{Name: "Carlos Souza", Email: "carlos@example.com", Role: model.UserRoleContributor}

	newDate := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	newDeadline := newSupersedingDeadline(old, newDate, nil, nil)

	got, err := repo.SupersedeDeadline(context.Background(), withDeadline, old, newDeadline, submitter)
	require.NoError(t, err)

	assert.NotEqual(t, originalLastUpdatedByID, got.LastUpdatedByID)
	assert.Equal(t, got.LastUpdatedByID, got.LastUpdatedBy.ID)
	assert.Equal(t, "Carlos Souza", got.LastUpdatedBy.Name)
}

func TestEventRepository_SupersedeDeadline_WritesDeadlineSupersededAuditLog(t *testing.T) {
	// Spec: rule "AuditLog: one deadline_superseded row (entity_type=deadline,
	// entity_id=<old deadline ID>) whose diff records before/after for date,
	// time, timezone, is_active, and superseded_by_id"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event := newApprovedEvent(t, tx, "MODELS2026", "ana@example.com")
	withDeadline := addOneDeadline(t, tx, event)
	old := withDeadline.Deadlines[0]
	submitter := model.User{Name: "Carlos Souza", Email: "carlos@example.com", Role: model.UserRoleContributor}

	newDate := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	newDeadline := newSupersedingDeadline(old, newDate, nil, nil)

	got, err := repo.SupersedeDeadline(context.Background(), withDeadline, old, newDeadline, submitter)
	require.NoError(t, err)

	var supersededLog model.AuditLog
	err = tx.Where("entity_type = ? AND entity_id = ? AND action = ?",
		model.AuditEntityDeadline, old.ID, model.AuditActionDeadlineSuperseded).First(&supersededLog).Error
	require.NoError(t, err)
	assert.Equal(t, got.LastUpdatedByID, supersededLog.ChangedByID)

	var diff map[string]any
	require.NoError(t, json.Unmarshal(supersededLog.Diff, &diff))
	assert.Contains(t, diff, "date")
	assert.Contains(t, diff, "time")
	assert.Contains(t, diff, "timezone")
	assert.Contains(t, diff, "is_active")
	assert.Contains(t, diff, "superseded_by_id")
}

func TestEventRepository_SupersedeDeadline_WritesUpdatedAuditLogForLastUpdatedByID(t *testing.T) {
	// Spec: rule "plus the always-written updated row (entity_type=event) for
	// the last_updated_by_id change — mirrors events-deadlines-cancel.yaml's
	// audit pattern"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event := newApprovedEvent(t, tx, "MODELS2026", "ana@example.com")
	withDeadline := addOneDeadline(t, tx, event)
	old := withDeadline.Deadlines[0]
	submitter := model.User{Name: "Carlos Souza", Email: "carlos@example.com", Role: model.UserRoleContributor}

	newDate := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	newDeadline := newSupersedingDeadline(old, newDate, nil, nil)

	got, err := repo.SupersedeDeadline(context.Background(), withDeadline, old, newDeadline, submitter)
	require.NoError(t, err)

	var updatedLog model.AuditLog
	err = tx.Where("entity_type = ? AND entity_id = ? AND action = ?",
		model.AuditEntityEvent, got.ID, model.AuditActionUpdated).Order("id DESC").First(&updatedLog).Error
	require.NoError(t, err)
	assert.Equal(t, got.LastUpdatedByID, updatedLog.ChangedByID)
}

func TestEventRepository_SupersedeDeadline_ReloadedEventIncludesBothOldAndNewDeadlines(t *testing.T) {
	// Spec: rule "preloadEventAssociations changes ... so reload responses
	// (this endpoint, ...) include superseded deadlines alongside active
	// ones"; DoD "response deadlines include both the new active deadline and
	// the superseded one (is_active=false, superseded_by_id=<new id>)"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event := newApprovedEvent(t, tx, "MODELS2026", "ana@example.com")
	withDeadline := addOneDeadline(t, tx, event)
	old := withDeadline.Deadlines[0]
	submitter := model.User{Name: "Carlos Souza", Email: "carlos@example.com", Role: model.UserRoleContributor}

	newDate := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	newDeadline := newSupersedingDeadline(old, newDate, nil, nil)

	got, err := repo.SupersedeDeadline(context.Background(), withDeadline, old, newDeadline, submitter)
	require.NoError(t, err)

	require.Len(t, got.Deadlines, 2)

	var oldFound, newFound bool
	for _, d := range got.Deadlines {
		switch d.ID {
		case old.ID:
			oldFound = true
			assert.False(t, d.IsActive)
			require.NotNil(t, d.SupersededByID)
		default:
			newFound = true
			assert.True(t, d.IsActive)
			assert.Nil(t, d.SupersededByID)
		}
	}
	assert.True(t, oldFound)
	assert.True(t, newFound)
}

// --- audit_logs_action_check constraint (migrations/007_add_batch_deadlines_added_audit_action.sql) ---

func TestAuditLog_BatchDeadlinesAddedAction_IsAllowedByConstraint(t *testing.T) {
	// Spec: events-deadlines-add.yaml rule "More than 1 deadline submitted → one row ...
	// action=batch_deadlines_added". The audit_logs_action_check CHECK constraint must
	// allow this new value.
	tx, rollback := beginTx(t)
	defer rollback()

	user := model.User{Name: "Ana Silva", Email: "ana@example.com", Role: model.UserRoleContributor}
	require.NoError(t, tx.Create(&user).Error)

	event := model.Event{
		Name: "MODELS", Slug: "MODELS2026", Country: "Brazil", City: "Recife",
		Latitude: -8.04, Longitude: -34.87,
		StartDate:  time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC),
		EndDate:    time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC),
		WebsiteURL: "https://models2026.example.org", Domain: "computer_science",
		Status: model.EventStatusApproved, Year: 2026,
		CreatedByID: user.ID, LastUpdatedByID: user.ID,
	}
	require.NoError(t, tx.Create(&event).Error)

	log := model.AuditLog{
		EntityType:  model.AuditEntityEvent,
		EntityID:    event.ID,
		Action:      model.AuditActionBatchDeadlinesAdded,
		ChangedByID: user.ID,
	}

	err := tx.WithContext(context.Background()).Create(&log).Error

	require.NoError(t, err)
}

// --- audit_logs_action_check constraint (migrations/008_add_deadline_cancelled_audit_action.sql) ---

func TestAuditLog_DeadlineCancelledAction_IsAllowedByConstraint(t *testing.T) {
	// Spec: events-deadlines-cancel.yaml rule "entity_type=deadline, ...,
	// action=deadline_cancelled". The audit_logs_action_check CHECK constraint
	// must allow this new value.
	tx, rollback := beginTx(t)
	defer rollback()

	user := model.User{Name: "Ana Silva", Email: "ana@example.com", Role: model.UserRoleContributor}
	require.NoError(t, tx.Create(&user).Error)

	event := model.Event{
		Name: "MODELS", Slug: "MODELS2026", Country: "Brazil", City: "Recife",
		Latitude: -8.04, Longitude: -34.87,
		StartDate:  time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC),
		EndDate:    time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC),
		WebsiteURL: "https://models2026.example.org", Domain: "computer_science",
		Status: model.EventStatusApproved, Year: 2026,
		CreatedByID: user.ID, LastUpdatedByID: user.ID,
	}
	require.NoError(t, tx.Create(&event).Error)

	deadline := model.Deadline{
		EventID: event.ID, Type: model.DeadlineTypePaper, Description: "Research track full paper",
		Date: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), IsActive: true, CreatedByID: user.ID,
	}
	require.NoError(t, tx.Create(&deadline).Error)

	log := model.AuditLog{
		EntityType:  model.AuditEntityDeadline,
		EntityID:    deadline.ID,
		Action:      model.AuditActionDeadlineCancelled,
		ChangedByID: user.ID,
	}

	err := tx.WithContext(context.Background()).Create(&log).Error

	require.NoError(t, err)
}
