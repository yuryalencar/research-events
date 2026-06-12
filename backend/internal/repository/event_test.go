package repository_test

// Spec: specs/backend/events-submit.yaml

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuryalencar/research-events/internal/model"
	"github.com/yuryalencar/research-events/internal/repository"
)

// --- FindActiveBySlug ---

func TestEventRepository_FindActiveBySlug_ReturnsEventWhenPending(t *testing.T) {
	// Spec: events-submit.yaml rule "slug ... must be unique among events with status=pending or status=approved"
	tx, rollback := beginTx(t)
	defer rollback()

	user := model.User{Name: "Ana Silva", Email: "ana@example.com", Role: model.UserRoleContributor}
	require.NoError(t, tx.Create(&user).Error)

	event := model.Event{
		Name: "MODELS", Slug: "MODELS2026", Country: "Brazil", City: "Recife",
		Latitude: -8.04, Longitude: -34.87,
		StartDate: time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC),
		WebsiteURL: "https://models2026.example.org", Domain: "computer_science",
		Status: model.EventStatusPending, Year: 2026,
		CreatedByID: user.ID, LastUpdatedByID: user.ID,
	}
	require.NoError(t, tx.Create(&event).Error)

	repo := repository.NewEventRepository(tx)

	got, err := repo.FindActiveBySlug(context.Background(), "MODELS2026")

	require.NoError(t, err)
	assert.Equal(t, event.ID, got.ID)
}

func TestEventRepository_FindActiveBySlug_ReturnsEventWhenApproved(t *testing.T) {
	tx, rollback := beginTx(t)
	defer rollback()

	user := model.User{Name: "Ana Silva", Email: "ana@example.com", Role: model.UserRoleContributor}
	require.NoError(t, tx.Create(&user).Error)

	event := model.Event{
		Name: "ICSE", Slug: "ICSE2027", Country: "USA", City: "Boston",
		Latitude: 42.36, Longitude: -71.05,
		StartDate: time.Date(2027, 5, 10, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2027, 5, 18, 0, 0, 0, 0, time.UTC),
		WebsiteURL: "https://icse2027.example.org", Domain: "computer_science",
		Status: model.EventStatusApproved, Year: 2027,
		CreatedByID: user.ID, LastUpdatedByID: user.ID,
	}
	require.NoError(t, tx.Create(&event).Error)

	repo := repository.NewEventRepository(tx)

	got, err := repo.FindActiveBySlug(context.Background(), "ICSE2027")

	require.NoError(t, err)
	assert.Equal(t, event.ID, got.ID)
}

func TestEventRepository_FindActiveBySlug_ReturnsErrNotFoundWhenOnlyRejected(t *testing.T) {
	// Spec: events-submit.yaml border_case "slug previously used only by a rejected event → allowed, 201 (slug reusable)"
	tx, rollback := beginTx(t)
	defer rollback()

	user := model.User{Name: "Ana Silva", Email: "ana@example.com", Role: model.UserRoleContributor}
	require.NoError(t, tx.Create(&user).Error)

	event := model.Event{
		Name: "Rejected Conf", Slug: "REJECTED2026", Country: "Brazil", City: "Recife",
		Latitude: -8.04, Longitude: -34.87,
		StartDate: time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC),
		WebsiteURL: "https://rejected.example.org", Domain: "computer_science",
		Status: model.EventStatusRejected, Year: 2026,
		CreatedByID: user.ID, LastUpdatedByID: user.ID,
	}
	require.NoError(t, tx.Create(&event).Error)

	repo := repository.NewEventRepository(tx)

	_, err := repo.FindActiveBySlug(context.Background(), "REJECTED2026")

	require.ErrorIs(t, err, repository.ErrNotFound)
}

func TestEventRepository_FindActiveBySlug_ReturnsErrNotFoundWhenMissing(t *testing.T) {
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)

	_, err := repo.FindActiveBySlug(context.Background(), "DOES-NOT-EXIST")

	require.ErrorIs(t, err, repository.ErrNotFound)
}

// --- Submit ---

// newSubmission returns a fresh (unsaved) Event, its Deadlines, and Submitter,
// the same shapes service.BuildSubmission returns to the handler.
func newSubmission(slug, email string) (model.Event, []model.Deadline, model.User) {
	event := model.Event{
		Name: "MODELS", Slug: slug, Country: "Brazil", City: "Recife",
		Latitude: -8.04, Longitude: -34.87,
		StartDate: time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC),
		WebsiteURL: "https://models2026.example.org", Domain: "computer_science",
		Status: model.EventStatusPending, Year: 2026,
	}
	submitter := model.User{Name: "Ana Silva", Email: email, Role: model.UserRoleContributor}
	return event, nil, submitter
}

func TestEventRepository_Submit_CreatesNewContributorWhenEmailNotFound(t *testing.T) {
	// Spec: events-submit.yaml rule "If no User exists, create one with role=contributor, password_hash=NULL"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event, deadlines, submitter := newSubmission("MODELS2026", "new@example.com")

	got, err := repo.Submit(context.Background(), event, deadlines, submitter)

	require.NoError(t, err)
	assert.NotZero(t, got.CreatedByID)
	assert.Equal(t, "new@example.com", got.CreatedBy.Email)
	assert.Equal(t, model.UserRoleContributor, got.CreatedBy.Role)
	assert.Nil(t, got.CreatedBy.PasswordHash)
}

func TestEventRepository_Submit_ReusesAndRenamesExistingUser(t *testing.T) {
	// Spec: events-submit.yaml rule "If a User with that email exists, link event.created_by_id to
	// that user and update their name to the submitted submitter.name (always overwritten)"
	tx, rollback := beginTx(t)
	defer rollback()

	existing := model.User{Name: "Old Name", Email: "existing@example.com", Role: model.UserRoleAdmin}
	require.NoError(t, tx.Create(&existing).Error)

	repo := repository.NewEventRepository(tx)
	event, deadlines, submitter := newSubmission("MODELS2026", "existing@example.com")
	submitter.Name = "New Name"

	got, err := repo.Submit(context.Background(), event, deadlines, submitter)

	require.NoError(t, err)
	assert.Equal(t, existing.ID, got.CreatedByID)
	assert.Equal(t, "New Name", got.CreatedBy.Name)
	assert.Equal(t, model.UserRoleAdmin, got.CreatedBy.Role) // role is never changed by submission
}

func TestEventRepository_Submit_CreatesEventWithStatusPending(t *testing.T) {
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event, deadlines, submitter := newSubmission("MODELS2026", "ana@example.com")

	got, err := repo.Submit(context.Background(), event, deadlines, submitter)

	require.NoError(t, err)
	assert.NotZero(t, got.ID)
	assert.Equal(t, model.EventStatusPending, got.Status)
}

func TestEventRepository_Submit_CreatesDeadlinesWithIsActiveTrue(t *testing.T) {
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event, _, submitter := newSubmission("MODELS2026", "ana@example.com")
	deadlines := []model.Deadline{
		{Type: model.DeadlineTypePaper, Description: "Research track full paper", Date: time.Date(2026, 8, 22, 0, 0, 0, 0, time.UTC), IsActive: true},
	}

	got, err := repo.Submit(context.Background(), event, deadlines, submitter)

	require.NoError(t, err)
	require.Len(t, got.Deadlines, 1)
	assert.True(t, got.Deadlines[0].IsActive)
	assert.Equal(t, got.ID, got.Deadlines[0].EventID)
	assert.NotZero(t, got.Deadlines[0].CreatedByID)
}

func TestEventRepository_Submit_NoDeadlinesWhenNoneProvided(t *testing.T) {
	// Spec: events-submit.yaml border_case "deadlines omitted or empty array → allowed, event created with no deadlines"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event, deadlines, submitter := newSubmission("MODELS2026", "ana@example.com")

	got, err := repo.Submit(context.Background(), event, deadlines, submitter)

	require.NoError(t, err)
	assert.Empty(t, got.Deadlines)
}

func TestEventRepository_Submit_WritesAuditLogForEventCreation(t *testing.T) {
	// Spec: events-submit.yaml rule "AuditLog row written for the Event (entity_type=event, action=created, ...)"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event, deadlines, submitter := newSubmission("MODELS2026", "ana@example.com")

	got, err := repo.Submit(context.Background(), event, deadlines, submitter)
	require.NoError(t, err)

	var log model.AuditLog
	err = tx.Where("entity_type = ? AND entity_id = ? AND action = ?",
		model.AuditEntityEvent, got.ID, model.AuditActionCreated).First(&log).Error

	require.NoError(t, err)
	assert.Equal(t, got.CreatedByID, log.ChangedByID)
}

func TestEventRepository_Submit_WritesAuditLogPerDeadline(t *testing.T) {
	// Spec: events-submit.yaml rule "AuditLog row written for each Deadline created
	// (entity_type=deadline, action=deadline_added, ...)"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event, _, submitter := newSubmission("MODELS2026", "ana@example.com")
	deadlines := []model.Deadline{
		{Type: model.DeadlineTypeAbstract, Description: "Abstract", Date: time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC), IsActive: true},
		{Type: model.DeadlineTypePaper, Description: "Paper", Date: time.Date(2026, 8, 22, 0, 0, 0, 0, time.UTC), IsActive: true},
	}

	got, err := repo.Submit(context.Background(), event, deadlines, submitter)
	require.NoError(t, err)
	require.Len(t, got.Deadlines, 2)

	var count int64
	err = tx.Model(&model.AuditLog{}).
		Where("entity_type = ? AND action = ? AND entity_id IN ?",
			model.AuditEntityDeadline, model.AuditActionDeadlineAdded,
			[]uint{got.Deadlines[0].ID, got.Deadlines[1].ID}).
		Count(&count).Error

	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}
