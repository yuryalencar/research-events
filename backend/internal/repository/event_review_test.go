package repository_test

// Spec: specs/backend/admin-events-review.yaml

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

// newReviewer persists and returns a new admin User — the "reviewer" for Review tests.
func newReviewer(t *testing.T, tx *gorm.DB, email string) model.User {
	t.Helper()
	user := model.User{Name: "Admin", Email: email, Role: model.UserRoleAdmin}
	require.NoError(t, tx.Create(&user).Error)
	return user
}

func TestEventRepository_Review_UpdatesStatusToApproved(t *testing.T) {
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	pending := newApprovedEvent(t, tx, "MODELS2026", "ana@example.com")
	require.NoError(t, tx.Model(&model.Event{}).Where("id = ?", pending.ID).Update("status", model.EventStatusPending).Error)
	pending.Status = model.EventStatusPending

	reviewer := newReviewer(t, tx, "admin@example.com")
	updated := pending
	updated.Status = model.EventStatusApproved
	updated.LastUpdatedByID = reviewer.ID

	auditLog := model.AuditLog{
		EntityType:  model.AuditEntityEvent,
		EntityID:    pending.ID,
		Action:      model.AuditActionApproved,
		ChangedByID: reviewer.ID,
	}

	got, err := repo.Review(context.Background(), updated, auditLog)

	require.NoError(t, err)
	assert.Equal(t, model.EventStatusApproved, got.Status)
}

func TestEventRepository_Review_UpdatesStatusToRejected(t *testing.T) {
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	pending := newApprovedEvent(t, tx, "MODELS2026", "ana@example.com")
	reviewer := newReviewer(t, tx, "admin@example.com")

	updated := pending
	updated.Status = model.EventStatusRejected
	updated.LastUpdatedByID = reviewer.ID

	reason := "duplicate submission"
	auditLog := model.AuditLog{
		EntityType:  model.AuditEntityEvent,
		EntityID:    pending.ID,
		Action:      model.AuditActionRejected,
		ChangedByID: reviewer.ID,
		Reason:      &reason,
	}

	got, err := repo.Review(context.Background(), updated, auditLog)

	require.NoError(t, err)
	assert.Equal(t, model.EventStatusRejected, got.Status)
}

func TestEventRepository_Review_PersistsEditedFields(t *testing.T) {
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	pending := newApprovedEvent(t, tx, "MODELS2026", "ana@example.com")
	reviewer := newReviewer(t, tx, "admin@example.com")

	updated := pending
	updated.Name = "Fixed Conference Name"
	updated.StartDate = time.Date(2027, 1, 10, 0, 0, 0, 0, time.UTC)
	updated.Year = 2027
	updated.Status = model.EventStatusApproved
	updated.LastUpdatedByID = reviewer.ID

	auditLog := model.AuditLog{
		EntityType:  model.AuditEntityEvent,
		EntityID:    pending.ID,
		Action:      model.AuditActionApproved,
		ChangedByID: reviewer.ID,
	}

	got, err := repo.Review(context.Background(), updated, auditLog)

	require.NoError(t, err)
	assert.Equal(t, "Fixed Conference Name", got.Name)
	assert.True(t, updated.StartDate.Equal(got.StartDate))
	assert.Equal(t, 2027, got.Year)
}

func TestEventRepository_Review_UpdatesLastUpdatedByID(t *testing.T) {
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	pending := newApprovedEvent(t, tx, "MODELS2026", "ana@example.com")
	reviewer := newReviewer(t, tx, "admin@example.com")

	updated := pending
	updated.Status = model.EventStatusApproved
	updated.LastUpdatedByID = reviewer.ID

	auditLog := model.AuditLog{
		EntityType:  model.AuditEntityEvent,
		EntityID:    pending.ID,
		Action:      model.AuditActionApproved,
		ChangedByID: reviewer.ID,
	}

	got, err := repo.Review(context.Background(), updated, auditLog)

	require.NoError(t, err)
	assert.Equal(t, reviewer.ID, got.LastUpdatedByID)
	assert.Equal(t, reviewer.ID, got.LastUpdatedBy.ID)
}

func TestEventRepository_Review_WritesAuditLogWithReason(t *testing.T) {
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	pending := newApprovedEvent(t, tx, "MODELS2026", "ana@example.com")
	reviewer := newReviewer(t, tx, "admin@example.com")

	updated := pending
	updated.Status = model.EventStatusRejected
	updated.LastUpdatedByID = reviewer.ID

	reason := "duplicate submission"
	diff, err := json.Marshal(map[string]any{
		"status": map[string]any{"before": "approved", "after": "rejected"},
	})
	require.NoError(t, err)

	auditLog := model.AuditLog{
		EntityType:  model.AuditEntityEvent,
		EntityID:    pending.ID,
		Action:      model.AuditActionRejected,
		ChangedByID: reviewer.ID,
		Diff:        model.JSONB(diff),
		Reason:      &reason,
	}

	_, err = repo.Review(context.Background(), updated, auditLog)
	require.NoError(t, err)

	var stored model.AuditLog
	require.NoError(t, tx.Where("entity_type = ? AND entity_id = ? AND action = ?",
		model.AuditEntityEvent, pending.ID, model.AuditActionRejected).First(&stored).Error)

	require.NotNil(t, stored.Reason)
	assert.Equal(t, "duplicate submission", *stored.Reason)
	assert.Equal(t, reviewer.ID, stored.ChangedByID)
	assert.JSONEq(t, string(diff), string(stored.Diff))
}

func TestEventRepository_Review_WritesAuditLogWithoutReason(t *testing.T) {
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	pending := newApprovedEvent(t, tx, "MODELS2026", "ana@example.com")
	reviewer := newReviewer(t, tx, "admin@example.com")

	updated := pending
	updated.Status = model.EventStatusApproved
	updated.LastUpdatedByID = reviewer.ID

	auditLog := model.AuditLog{
		EntityType:  model.AuditEntityEvent,
		EntityID:    pending.ID,
		Action:      model.AuditActionApproved,
		ChangedByID: reviewer.ID,
	}

	_, err := repo.Review(context.Background(), updated, auditLog)
	require.NoError(t, err)

	var stored model.AuditLog
	require.NoError(t, tx.Where("entity_type = ? AND entity_id = ? AND action = ?",
		model.AuditEntityEvent, pending.ID, model.AuditActionApproved).First(&stored).Error)

	assert.Nil(t, stored.Reason)
}

func TestEventRepository_Review_ReturnsEventWithActiveDeadlines(t *testing.T) {
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event, deadlines, submitter := newSubmission("MODELS2026", "ana@example.com")
	event.Status = model.EventStatusApproved
	deadlines = oneDeadline(model.DeadlineTypePaper, "Paper deadline", time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
	created, err := repo.Submit(context.Background(), event, deadlines, submitter)
	require.NoError(t, err)

	reviewer := newReviewer(t, tx, "admin@example.com")
	updated := created
	updated.Status = model.EventStatusApproved
	updated.LastUpdatedByID = reviewer.ID

	auditLog := model.AuditLog{
		EntityType:  model.AuditEntityEvent,
		EntityID:    created.ID,
		Action:      model.AuditActionApproved,
		ChangedByID: reviewer.ID,
	}

	got, err := repo.Review(context.Background(), updated, auditLog)

	require.NoError(t, err)
	require.Len(t, got.Deadlines, 1)
	assert.Equal(t, "Paper deadline", got.Deadlines[0].Description)
}
