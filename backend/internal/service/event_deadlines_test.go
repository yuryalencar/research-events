package service_test

// Spec: specs/backend/events-submit.yaml, specs/backend/events-deadlines-add.yaml
//
// All functions in internal/service/ are pure — same input always produces same output,
// no I/O, no global state. Tests here need no mocks and no database: just input → output.

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuryalencar/research-events/internal/model"
	"github.com/yuryalencar/research-events/internal/service"
)

// --- ValidateAddDeadlinesInput ---

// validAddDeadlinesInput returns an AddDeadlinesInput that passes every validation rule.
// Each test mutates a copy of this baseline to isolate the field under test.
func validAddDeadlinesInput() service.AddDeadlinesInput {
	return service.AddDeadlinesInput{
		EventID: 1,
		Submitter: service.SubmitterInput{
			Name:  "Beatriz Costa",
			Email: "beatriz@example.com",
		},
		Deadlines: []service.DeadlineInput{
			{Type: "camera_ready", Description: "Research track camera-ready", Date: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)},
		},
	}
}

func TestValidateAddDeadlinesInput_ValidInput_ReturnsNil(t *testing.T) {
	err := service.ValidateAddDeadlinesInput(validAddDeadlinesInput())

	require.NoError(t, err)
}

func TestValidateAddDeadlinesInput_EmptyDeadlines_ReturnsError(t *testing.T) {
	// Spec: events-deadlines-add.yaml border_case "deadlines omitted or empty array → 400 VALIDATION_ERROR"
	input := validAddDeadlinesInput()
	input.Deadlines = nil

	err := service.ValidateAddDeadlinesInput(input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "deadlines")
}

func TestValidateAddDeadlinesInput_DeadlineMissingDescription_ReturnsError(t *testing.T) {
	// Spec: events-deadlines-add.yaml border_case "a deadlines entry missing type, description, or date → 400 VALIDATION_ERROR"
	input := validAddDeadlinesInput()
	input.Deadlines = []service.DeadlineInput{
		{Type: "paper", Date: time.Date(2026, 8, 22, 0, 0, 0, 0, time.UTC)},
	}

	err := service.ValidateAddDeadlinesInput(input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "description")
}

func TestValidateAddDeadlinesInput_DeadlineInvalidType_ReturnsError(t *testing.T) {
	// Spec: events-deadlines-add.yaml border_case "a deadlines entry with an invalid type value → 400 VALIDATION_ERROR"
	input := validAddDeadlinesInput()
	input.Deadlines = []service.DeadlineInput{
		{Type: "keynote", Description: "Keynote announcement", Date: time.Date(2026, 8, 22, 0, 0, 0, 0, time.UTC)},
	}

	err := service.ValidateAddDeadlinesInput(input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "type")
}

func TestValidateAddDeadlinesInput_MissingSubmitterName_ReturnsError(t *testing.T) {
	// Spec: events-deadlines-add.yaml border_case "submitter.name missing → 400 VALIDATION_ERROR"
	input := validAddDeadlinesInput()
	input.Submitter.Name = ""

	err := service.ValidateAddDeadlinesInput(input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "submitter.name")
}

func TestValidateAddDeadlinesInput_InvalidSubmitterEmail_ReturnsError(t *testing.T) {
	// Spec: events-deadlines-add.yaml border_case "submitter.email missing or invalid format → 400 VALIDATION_ERROR"
	input := validAddDeadlinesInput()
	input.Submitter.Email = "not-an-email"

	err := service.ValidateAddDeadlinesInput(input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "submitter.email")
}

func TestValidateAddDeadlinesInput_MultipleDeadlines_ReturnsNil(t *testing.T) {
	input := validAddDeadlinesInput()
	input.Deadlines = []service.DeadlineInput{
		{Type: "paper", Description: "Industry track full paper", Date: time.Date(2026, 8, 29, 0, 0, 0, 0, time.UTC)},
		{Type: "notification", Description: "Industry track notification", Date: time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)},
	}

	err := service.ValidateAddDeadlinesInput(input)

	require.NoError(t, err)
}

func TestValidateAddDeadlinesInput_SameInputTwice_ReturnsIdenticalResult(t *testing.T) {
	// FP: pure function — same input always produces the same output, no hidden state.
	input := validAddDeadlinesInput()

	err1 := service.ValidateAddDeadlinesInput(input)
	err2 := service.ValidateAddDeadlinesInput(input)

	assert.Equal(t, err1, err2)
}

// --- ValidateCancelDeadlineInput ---

// validCancelDeadlineInput returns a CancelDeadlineInput that passes every
// validation rule. Each test mutates a copy of this baseline to isolate the
// field under test.
func validCancelDeadlineInput() service.CancelDeadlineInput {
	return service.CancelDeadlineInput{
		Submitter: service.SubmitterInput{
			Name:  "Beatriz Costa",
			Email: "beatriz@example.com",
		},
	}
}

func TestValidateCancelDeadlineInput_ValidInput_ReturnsNil(t *testing.T) {
	err := service.ValidateCancelDeadlineInput(validCancelDeadlineInput())

	require.NoError(t, err)
}

func TestValidateCancelDeadlineInput_MissingSubmitterName_ReturnsError(t *testing.T) {
	// Spec: events-deadlines-cancel.yaml border_case "submitter.name missing → 400 VALIDATION_ERROR"
	input := validCancelDeadlineInput()
	input.Submitter.Name = ""

	err := service.ValidateCancelDeadlineInput(input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "submitter.name")
}

func TestValidateCancelDeadlineInput_InvalidSubmitterEmail_ReturnsError(t *testing.T) {
	// Spec: events-deadlines-cancel.yaml border_case "submitter.email missing or invalid format → 400 VALIDATION_ERROR"
	input := validCancelDeadlineInput()
	input.Submitter.Email = "not-an-email"

	err := service.ValidateCancelDeadlineInput(input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "submitter.email")
}

func TestValidateCancelDeadlineInput_SameInputTwice_ReturnsIdenticalResult(t *testing.T) {
	// FP: pure function — same input always produces the same output, no hidden state.
	input := validCancelDeadlineInput()

	err1 := service.ValidateCancelDeadlineInput(input)
	err2 := service.ValidateCancelDeadlineInput(input)

	assert.Equal(t, err1, err2)
}

// --- ValidateDeadlineCancellable ---

func TestValidateDeadlineCancellable_ActiveDeadline_ReturnsNil(t *testing.T) {
	deadline := model.Deadline{IsActive: true}

	err := service.ValidateDeadlineCancellable(deadline)

	require.NoError(t, err)
}

func TestValidateDeadlineCancellable_AlreadyInactiveDeadline_ReturnsError(t *testing.T) {
	// Spec: events-deadlines-cancel.yaml border_case "target deadline is already
	// is_active=false (previously cancelled) → 409 DEADLINE_ALREADY_INACTIVE"
	deadline := model.Deadline{IsActive: false}

	err := service.ValidateDeadlineCancellable(deadline)

	require.Error(t, err)
}

func TestValidateDeadlineCancellable_AlreadySupersededDeadline_ReturnsError(t *testing.T) {
	// Spec: events-deadlines-cancel.yaml border_case "target deadline is already
	// is_active=false (superseded, superseded_by_id set) → 409 DEADLINE_ALREADY_INACTIVE"
	supersededBy := uint(99)
	deadline := model.Deadline{IsActive: false, SupersededByID: &supersededBy}

	err := service.ValidateDeadlineCancellable(deadline)

	require.Error(t, err)
}

func TestValidateDeadlineCancellable_SameInputTwice_ReturnsIdenticalResult(t *testing.T) {
	// FP: pure function — same input always produces the same output, no hidden state.
	deadline := model.Deadline{IsActive: false}

	err1 := service.ValidateDeadlineCancellable(deadline)
	err2 := service.ValidateDeadlineCancellable(deadline)

	assert.Equal(t, err1, err2)
}

// --- DetermineDeadlinesAuditAction ---

func TestDetermineDeadlinesAuditAction_SingleDeadline_ReturnsDeadlineAdded(t *testing.T) {
	// Spec: events-deadlines-add.yaml rule "Exactly 1 deadline submitted → one row:
	// entity_type=deadline, ..., action=deadline_added"
	action := service.DetermineDeadlinesAuditAction(1)

	assert.Equal(t, model.AuditActionDeadlineAdded, action)
}

func TestDetermineDeadlinesAuditAction_MultipleDeadlines_ReturnsBatchDeadlinesAdded(t *testing.T) {
	// Spec: events-deadlines-add.yaml rule "More than 1 deadline submitted → one row:
	// entity_type=event, ..., action=batch_deadlines_added"
	action := service.DetermineDeadlinesAuditAction(2)

	assert.Equal(t, model.AuditActionBatchDeadlinesAdded, action)
}

func TestDetermineDeadlinesAuditAction_SameInputTwice_ReturnsIdenticalResult(t *testing.T) {
	// FP: pure function — same input always produces the same output, no hidden state.
	action1 := service.DetermineDeadlinesAuditAction(3)
	action2 := service.DetermineDeadlinesAuditAction(3)

	assert.Equal(t, action1, action2)
}

// --- BuildDeadlinesFromInput ---

func TestBuildDeadlinesFromInput_EmptyInput_ReturnsEmptySlice(t *testing.T) {
	// Spec: events-submit.yaml border_case "deadlines omitted or empty array → allowed, event created with no deadlines"
	deadlines := service.BuildDeadlinesFromInput(nil)

	assert.Empty(t, deadlines)
}

func TestBuildDeadlinesFromInput_SetsIsActiveTrue(t *testing.T) {
	input := []service.DeadlineInput{
		{Type: "paper", Description: "Research track full paper", Date: time.Date(2026, 8, 22, 0, 0, 0, 0, time.UTC)},
	}

	deadlines := service.BuildDeadlinesFromInput(input)

	require.Len(t, deadlines, 1)
	assert.True(t, deadlines[0].IsActive)
}

func TestBuildDeadlinesFromInput_DefaultsIsOptionalFalse(t *testing.T) {
	input := []service.DeadlineInput{
		{Type: "paper", Description: "Research track full paper", Date: time.Date(2026, 8, 22, 0, 0, 0, 0, time.UTC)},
	}

	deadlines := service.BuildDeadlinesFromInput(input)

	require.Len(t, deadlines, 1)
	assert.False(t, deadlines[0].IsOptional)
}

func TestBuildDeadlinesFromInput_PreservesProvidedFields(t *testing.T) {
	date := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	input := []service.DeadlineInput{
		{Type: "abstract", Description: "Research track abstract", Date: date, IsOptional: true},
	}

	deadlines := service.BuildDeadlinesFromInput(input)

	require.Len(t, deadlines, 1)
	assert.Equal(t, model.DeadlineTypeAbstract, deadlines[0].Type)
	assert.Equal(t, "Research track abstract", deadlines[0].Description)
	assert.Equal(t, date, deadlines[0].Date)
	assert.True(t, deadlines[0].IsOptional)
}
