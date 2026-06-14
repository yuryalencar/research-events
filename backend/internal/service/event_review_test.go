package service_test

// Spec: specs/backend/admin-events-review.yaml
//
// All functions in internal/service/ are pure — same input always produces same output,
// no I/O, no global state. Tests here need no mocks and no database: just input → output.

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/yuryalencar/research-events/internal/model"
	"github.com/yuryalencar/research-events/internal/service"
)

// --- Helpers ---

// approvedEvent returns a model.Event that passes ValidateEditedEvent and has
// status=approved — the baseline "existing" event for review tests. Each test
// mutates a copy to isolate the field under test.
func approvedEvent() model.Event {
	return model.Event{
		Model:           gorm.Model{ID: 7},
		Name:            "International Conference on Model-Driven Engineering",
		Slug:            "MODELS2026",
		Country:         "Brazil",
		City:            "Recife",
		Latitude:        -8.0476,
		Longitude:       -34.8770,
		StartDate:       time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC),
		EndDate:         time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC),
		WebsiteURL:      "https://models2026.example.org",
		Domain:          "computer_science",
		Tier:            "A",
		Status:          model.EventStatusApproved,
		Year:            2026,
		CreatedByID:     3,
		LastUpdatedByID: 3,
	}
}

// adminReviewer returns a model.User with role=admin — the "reviewer" for review tests.
func adminReviewer() model.User {
	return model.User{Model: gorm.Model{ID: 1}, Name: "Admin", Email: "admin@example.com", Role: model.UserRoleAdmin}
}

func ptrString(s string) *string { return &s }

// --- ValidateReviewActionInput ---

func TestValidateReviewActionInput_ApproveWithoutReason_ReturnsNil(t *testing.T) {
	err := service.ValidateReviewActionInput(service.ReviewEventInput{Action: "approve"})

	require.NoError(t, err)
}

func TestValidateReviewActionInput_ApproveWithReason_ReturnsNil(t *testing.T) {
	err := service.ValidateReviewActionInput(service.ReviewEventInput{
		Action: "approve",
		Reason: ptrString("looks good"),
	})

	require.NoError(t, err)
}

func TestValidateReviewActionInput_RejectWithReason_ReturnsNil(t *testing.T) {
	err := service.ValidateReviewActionInput(service.ReviewEventInput{
		Action: "reject",
		Reason: ptrString("duplicate submission"),
	})

	require.NoError(t, err)
}

func TestValidateReviewActionInput_RejectWithoutReason_ReturnsError(t *testing.T) {
	err := service.ValidateReviewActionInput(service.ReviewEventInput{Action: "reject"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "reason is required")
}

func TestValidateReviewActionInput_RejectWithEmptyReason_ReturnsError(t *testing.T) {
	err := service.ValidateReviewActionInput(service.ReviewEventInput{
		Action: "reject",
		Reason: ptrString(""),
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "reason is required")
}

func TestValidateReviewActionInput_InvalidAction_ReturnsError(t *testing.T) {
	err := service.ValidateReviewActionInput(service.ReviewEventInput{Action: "approveee"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "action must be 'approve' or 'reject'")
}

func TestValidateReviewActionInput_MissingAction_ReturnsError(t *testing.T) {
	err := service.ValidateReviewActionInput(service.ReviewEventInput{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "action must be 'approve' or 'reject'")
}

func TestValidateReviewActionInput_SameInputTwice_ReturnsIdenticalResult(t *testing.T) {
	input := service.ReviewEventInput{Action: "reject"}

	err1 := service.ValidateReviewActionInput(input)
	err2 := service.ValidateReviewActionInput(input)

	assert.Equal(t, err1, err2)
}

// --- ApplyReview ---

func TestApplyReview_NoEventEdits_OnlyStatusAndLastUpdatedByIDChange(t *testing.T) {
	existing := approvedEvent()
	existing.Status = model.EventStatusPending

	updated := service.ApplyReview(existing, service.ReviewEventInput{Action: "approve"}, adminReviewer())

	assert.Equal(t, model.EventStatusApproved, updated.Status)
	assert.Equal(t, adminReviewer().ID, updated.LastUpdatedByID)
	assert.Equal(t, existing.Name, updated.Name)
	assert.Equal(t, existing.Slug, updated.Slug)
	assert.Equal(t, existing.Year, updated.Year)
}

func TestApplyReview_ApproveAction_SetsStatusApproved(t *testing.T) {
	existing := approvedEvent()
	existing.Status = model.EventStatusPending

	updated := service.ApplyReview(existing, service.ReviewEventInput{Action: "approve"}, adminReviewer())

	assert.Equal(t, model.EventStatusApproved, updated.Status)
}

func TestApplyReview_RejectAction_SetsStatusRejected(t *testing.T) {
	existing := approvedEvent()

	updated := service.ApplyReview(existing, service.ReviewEventInput{
		Action: "reject",
		Reason: ptrString("duplicate"),
	}, adminReviewer())

	assert.Equal(t, model.EventStatusRejected, updated.Status)
}

func TestApplyReview_PartialEdits_OnlyOverridesProvidedFields(t *testing.T) {
	existing := approvedEvent()

	updated := service.ApplyReview(existing, service.ReviewEventInput{
		Action: "approve",
		Event: &service.EventEditInput{
			Name: ptrString("Fixed Conference Name"),
		},
	}, adminReviewer())

	assert.Equal(t, "Fixed Conference Name", updated.Name)
	// Every other field is unchanged.
	assert.Equal(t, existing.Slug, updated.Slug)
	assert.Equal(t, existing.Country, updated.Country)
	assert.Equal(t, existing.City, updated.City)
	assert.Equal(t, existing.Latitude, updated.Latitude)
	assert.Equal(t, existing.Longitude, updated.Longitude)
	assert.True(t, existing.StartDate.Equal(updated.StartDate))
	assert.True(t, existing.EndDate.Equal(updated.EndDate))
	assert.Equal(t, existing.WebsiteURL, updated.WebsiteURL)
	assert.Equal(t, existing.Domain, updated.Domain)
	assert.Equal(t, existing.Tier, updated.Tier)
	assert.Equal(t, existing.Year, updated.Year)
}

func TestApplyReview_StartDateEdited_RecomputesYear(t *testing.T) {
	existing := approvedEvent() // Year: 2026, StartDate: 2026-09-21
	newStart := time.Date(2027, 1, 10, 0, 0, 0, 0, time.UTC)

	updated := service.ApplyReview(existing, service.ReviewEventInput{
		Action: "approve",
		Event: &service.EventEditInput{
			StartDate: &newStart,
		},
	}, adminReviewer())

	assert.True(t, updated.StartDate.Equal(newStart))
	assert.Equal(t, 2027, updated.Year)
}

func TestApplyReview_DoesNotMutateExistingEvent(t *testing.T) {
	existing := approvedEvent()
	originalName := existing.Name
	originalStatus := existing.Status

	_ = service.ApplyReview(existing, service.ReviewEventInput{
		Action: "reject",
		Reason: ptrString("duplicate"),
		Event: &service.EventEditInput{
			Name: ptrString("Changed Name"),
		},
	}, adminReviewer())

	assert.Equal(t, originalName, existing.Name)
	assert.Equal(t, originalStatus, existing.Status)
}

func TestApplyReview_SetsLastUpdatedByIDToReviewer(t *testing.T) {
	existing := approvedEvent()
	reviewer := model.User{Model: gorm.Model{ID: 42}, Role: model.UserRoleModerator}

	updated := service.ApplyReview(existing, service.ReviewEventInput{Action: "approve"}, reviewer)

	assert.Equal(t, uint(42), updated.LastUpdatedByID)
}

// --- ValidateEditedEvent ---

func TestValidateEditedEvent_ValidEvent_ReturnsNil(t *testing.T) {
	err := service.ValidateEditedEvent(approvedEvent())

	require.NoError(t, err)
}

func TestValidateEditedEvent_EmptyName_ReturnsError(t *testing.T) {
	event := approvedEvent()
	event.Name = ""

	err := service.ValidateEditedEvent(event)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestValidateEditedEvent_InvalidSlugCharacters_ReturnsError(t *testing.T) {
	event := approvedEvent()
	event.Slug = "not a valid slug!"

	err := service.ValidateEditedEvent(event)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "slug must contain only letters, digits, hyphens, and underscores")
}

func TestValidateEditedEvent_LatitudeOutOfRange_ReturnsError(t *testing.T) {
	event := approvedEvent()
	event.Latitude = 999

	err := service.ValidateEditedEvent(event)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "latitude must be between -90 and 90")
}

func TestValidateEditedEvent_LongitudeOutOfRange_ReturnsError(t *testing.T) {
	event := approvedEvent()
	event.Longitude = -999

	err := service.ValidateEditedEvent(event)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "longitude must be between -180 and 180")
}

func TestValidateEditedEvent_EndDateBeforeStartDate_ReturnsError(t *testing.T) {
	event := approvedEvent()
	event.EndDate = event.StartDate.AddDate(0, 0, -1)

	err := service.ValidateEditedEvent(event)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "end_date must not be before start_date")
}

func TestValidateEditedEvent_InvalidWebsiteURL_ReturnsError(t *testing.T) {
	event := approvedEvent()
	event.WebsiteURL = "not-a-url"

	err := service.ValidateEditedEvent(event)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "website_url must be a valid http or https URL")
}

func TestValidateEditedEvent_UnrecognizedDomain_ReturnsError(t *testing.T) {
	event := approvedEvent()
	event.Domain = "medicine"

	err := service.ValidateEditedEvent(event)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "is not a recognized domain")
}

func TestValidateEditedEvent_UnrecognizedTier_ReturnsError(t *testing.T) {
	event := approvedEvent()
	event.Tier = "S"

	err := service.ValidateEditedEvent(event)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "is not a recognized tier")
}

// --- BuildReviewAuditLog ---

func TestBuildReviewAuditLog_Approve_SetsActionApproved(t *testing.T) {
	existing := approvedEvent()
	existing.Status = model.EventStatusPending
	updated := existing
	updated.Status = model.EventStatusApproved

	log, err := service.BuildReviewAuditLog(existing, updated, adminReviewer(), nil)

	require.NoError(t, err)
	assert.Equal(t, model.AuditActionApproved, log.Action)
}

func TestBuildReviewAuditLog_Reject_SetsActionRejected(t *testing.T) {
	existing := approvedEvent()
	updated := existing
	updated.Status = model.EventStatusRejected

	log, err := service.BuildReviewAuditLog(existing, updated, adminReviewer(), ptrString("duplicate"))

	require.NoError(t, err)
	assert.Equal(t, model.AuditActionRejected, log.Action)
}

func TestBuildReviewAuditLog_DiffAlwaysIncludesStatusBeforeAfter(t *testing.T) {
	existing := approvedEvent()
	existing.Status = model.EventStatusPending
	updated := existing
	updated.Status = model.EventStatusApproved

	log, err := service.BuildReviewAuditLog(existing, updated, adminReviewer(), nil)
	require.NoError(t, err)

	var diff map[string]map[string]any
	require.NoError(t, json.Unmarshal(log.Diff, &diff))

	require.Contains(t, diff, "status")
	assert.Equal(t, string(model.EventStatusPending), diff["status"]["before"])
	assert.Equal(t, string(model.EventStatusApproved), diff["status"]["after"])
}

func TestBuildReviewAuditLog_DiffIncludesChangedEventFields(t *testing.T) {
	existing := approvedEvent()
	updated := existing
	updated.Name = "New Name"
	updated.Tier = "B"

	log, err := service.BuildReviewAuditLog(existing, updated, adminReviewer(), nil)
	require.NoError(t, err)

	var diff map[string]map[string]any
	require.NoError(t, json.Unmarshal(log.Diff, &diff))

	require.Contains(t, diff, "name")
	assert.Equal(t, existing.Name, diff["name"]["before"])
	assert.Equal(t, "New Name", diff["name"]["after"])

	require.Contains(t, diff, "tier")
	assert.Equal(t, existing.Tier, diff["tier"]["before"])
	assert.Equal(t, "B", diff["tier"]["after"])
}

func TestBuildReviewAuditLog_DiffOmitsUnchangedFields(t *testing.T) {
	existing := approvedEvent()
	updated := existing
	updated.Name = "New Name"

	log, err := service.BuildReviewAuditLog(existing, updated, adminReviewer(), nil)
	require.NoError(t, err)

	var diff map[string]map[string]any
	require.NoError(t, json.Unmarshal(log.Diff, &diff))

	assert.NotContains(t, diff, "slug")
	assert.NotContains(t, diff, "country")
	assert.NotContains(t, diff, "city")
	assert.NotContains(t, diff, "website_url")
	assert.NotContains(t, diff, "domain")
	assert.NotContains(t, diff, "tier")
	assert.NotContains(t, diff, "start_date")
	assert.NotContains(t, diff, "end_date")
	assert.NotContains(t, diff, "year")
}

func TestBuildReviewAuditLog_StartDateChange_DiffIncludesStartDateAndYear(t *testing.T) {
	existing := approvedEvent() // StartDate: 2026-09-21, Year: 2026
	updated := existing
	updated.StartDate = time.Date(2027, 1, 10, 0, 0, 0, 0, time.UTC)
	updated.Year = 2027

	log, err := service.BuildReviewAuditLog(existing, updated, adminReviewer(), nil)
	require.NoError(t, err)

	var diff map[string]map[string]any
	require.NoError(t, json.Unmarshal(log.Diff, &diff))

	require.Contains(t, diff, "start_date")
	require.Contains(t, diff, "year")
	assert.Equal(t, float64(2026), diff["year"]["before"])
	assert.Equal(t, float64(2027), diff["year"]["after"])
}

func TestBuildReviewAuditLog_SetsReasonWhenProvided(t *testing.T) {
	existing := approvedEvent()
	updated := existing
	updated.Status = model.EventStatusRejected

	log, err := service.BuildReviewAuditLog(existing, updated, adminReviewer(), ptrString("duplicate submission"))
	require.NoError(t, err)

	require.NotNil(t, log.Reason)
	assert.Equal(t, "duplicate submission", *log.Reason)
}

func TestBuildReviewAuditLog_ReasonNilWhenNotProvided(t *testing.T) {
	existing := approvedEvent()
	updated := existing

	log, err := service.BuildReviewAuditLog(existing, updated, adminReviewer(), nil)
	require.NoError(t, err)

	assert.Nil(t, log.Reason)
}

func TestBuildReviewAuditLog_SetsChangedByIDToReviewerAndEntity(t *testing.T) {
	existing := approvedEvent()
	updated := existing
	reviewer := model.User{Model: gorm.Model{ID: 9}, Role: model.UserRoleAdmin}

	log, err := service.BuildReviewAuditLog(existing, updated, reviewer, nil)
	require.NoError(t, err)

	assert.Equal(t, uint(9), log.ChangedByID)
	assert.Equal(t, model.AuditEntityEvent, log.EntityType)
	assert.Equal(t, existing.ID, log.EntityID)
}
