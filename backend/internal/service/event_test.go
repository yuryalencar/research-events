package service_test

// Spec: specs/backend/events-submit.yaml
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

// --- Helpers ---

// validSubmitInput returns a SubmitEventInput that passes every validation rule.
// Each test mutates a copy of this baseline to isolate the field under test.
func validSubmitInput() service.SubmitEventInput {
	return service.SubmitEventInput{
		Name:       "International Conference on Model-Driven Engineering",
		Slug:       "MODELS2026",
		Country:    "Brazil",
		City:       "Recife",
		Latitude:   -8.0476,
		Longitude:  -34.8770,
		StartDate:  time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC),
		EndDate:    time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC),
		WebsiteURL: "https://models2026.example.org",
		Domain:     "computer_science",
		Submitter: service.SubmitterInput{
			Name:  "Ana Silva",
			Email: "ana@example.com",
		},
	}
}

// --- ValidateSubmitEventInput ---

func TestValidateSubmitEventInput_ValidInput_ReturnsNil(t *testing.T) {
	err := service.ValidateSubmitEventInput(validSubmitInput())

	require.NoError(t, err)
}

func TestValidateSubmitEventInput_MissingName_ReturnsError(t *testing.T) {
	// Spec: events-submit.yaml border_case "Missing any required field (event or submitter) → 400 VALIDATION_ERROR"
	input := validSubmitInput()
	input.Name = ""

	err := service.ValidateSubmitEventInput(input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestValidateSubmitEventInput_MissingSlug_ReturnsError(t *testing.T) {
	input := validSubmitInput()
	input.Slug = ""

	err := service.ValidateSubmitEventInput(input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "slug")
}

func TestValidateSubmitEventInput_InvalidSlugFormat_ReturnsError(t *testing.T) {
	// Spec: events-submit.yaml border_case "slug contains characters outside [A-Za-z0-9_-] → 400 VALIDATION_ERROR"
	input := validSubmitInput()
	input.Slug = "MODELS 2026!"

	err := service.ValidateSubmitEventInput(input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "slug")
}

func TestValidateSubmitEventInput_MissingCountry_ReturnsError(t *testing.T) {
	input := validSubmitInput()
	input.Country = ""

	err := service.ValidateSubmitEventInput(input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "country")
}

func TestValidateSubmitEventInput_MissingCity_ReturnsError(t *testing.T) {
	input := validSubmitInput()
	input.City = ""

	err := service.ValidateSubmitEventInput(input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "city")
}

func TestValidateSubmitEventInput_LatitudeOutOfRange_ReturnsError(t *testing.T) {
	// Spec: events-submit.yaml border_case "latitude/longitude out of range → 400 VALIDATION_ERROR"
	input := validSubmitInput()
	input.Latitude = 200

	err := service.ValidateSubmitEventInput(input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "latitude")
}

func TestValidateSubmitEventInput_LongitudeOutOfRange_ReturnsError(t *testing.T) {
	input := validSubmitInput()
	input.Longitude = -200

	err := service.ValidateSubmitEventInput(input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "longitude")
}

func TestValidateSubmitEventInput_EndDateBeforeStartDate_ReturnsError(t *testing.T) {
	// Spec: events-submit.yaml border_case "end_date earlier than start_date → 400 VALIDATION_ERROR"
	input := validSubmitInput()
	input.StartDate = time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC)
	input.EndDate = time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC)

	err := service.ValidateSubmitEventInput(input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "end_date")
}

func TestValidateSubmitEventInput_PastDates_ReturnsNil(t *testing.T) {
	// Spec: events-submit.yaml rule "Past dates are allowed"
	input := validSubmitInput()
	input.StartDate = time.Date(2020, 3, 1, 0, 0, 0, 0, time.UTC)
	input.EndDate = time.Date(2020, 3, 5, 0, 0, 0, 0, time.UTC)

	err := service.ValidateSubmitEventInput(input)

	require.NoError(t, err)
}

func TestValidateSubmitEventInput_InvalidWebsiteURL_ReturnsError(t *testing.T) {
	// Spec: events-submit.yaml border_case "website_url not a valid http/https URL → 400 VALIDATION_ERROR"
	input := validSubmitInput()
	input.WebsiteURL = "models2026.example.org"

	err := service.ValidateSubmitEventInput(input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "website_url")
}

func TestValidateSubmitEventInput_InvalidDomain_ReturnsError(t *testing.T) {
	// Spec: events-submit.yaml border_case "domain not in the allowed enum → 400 VALIDATION_ERROR"
	input := validSubmitInput()
	input.Domain = "medicine"

	err := service.ValidateSubmitEventInput(input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "domain")
}

func TestValidateSubmitEventInput_MissingSubmitterName_ReturnsError(t *testing.T) {
	input := validSubmitInput()
	input.Submitter.Name = ""

	err := service.ValidateSubmitEventInput(input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "submitter.name")
}

func TestValidateSubmitEventInput_InvalidSubmitterEmail_ReturnsError(t *testing.T) {
	// Spec: events-submit.yaml border_case "Invalid submitter.email format → 400 VALIDATION_ERROR"
	input := validSubmitInput()
	input.Submitter.Email = "not-an-email"

	err := service.ValidateSubmitEventInput(input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "submitter.email")
}

func TestValidateSubmitEventInput_NoDeadlines_ReturnsNil(t *testing.T) {
	// Spec: events-submit.yaml border_case "deadlines omitted or empty array → allowed, event created with no deadlines"
	input := validSubmitInput()
	input.Deadlines = nil

	err := service.ValidateSubmitEventInput(input)

	require.NoError(t, err)
}

func TestValidateSubmitEventInput_DeadlineMissingDescription_ReturnsError(t *testing.T) {
	// Spec: events-submit.yaml border_case "deadlines entry missing type, description, or date → 400 VALIDATION_ERROR"
	input := validSubmitInput()
	input.Deadlines = []service.DeadlineInput{
		{Type: "paper", Date: time.Date(2026, 8, 22, 0, 0, 0, 0, time.UTC)},
	}

	err := service.ValidateSubmitEventInput(input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "description")
}

func TestValidateSubmitEventInput_DeadlineInvalidType_ReturnsError(t *testing.T) {
	// Spec: events-submit.yaml border_case "deadlines entry with invalid type value → 400 VALIDATION_ERROR"
	input := validSubmitInput()
	input.Deadlines = []service.DeadlineInput{
		{Type: "keynote", Description: "Keynote announcement", Date: time.Date(2026, 8, 22, 0, 0, 0, 0, time.UTC)},
	}

	err := service.ValidateSubmitEventInput(input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "type")
}

func TestValidateSubmitEventInput_ValidDeadlines_ReturnsNil(t *testing.T) {
	input := validSubmitInput()
	input.Deadlines = []service.DeadlineInput{
		{Type: "abstract", Description: "Research track abstract", Date: time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC), IsOptional: true},
		{Type: "paper", Description: "Research track full paper", Date: time.Date(2026, 8, 22, 0, 0, 0, 0, time.UTC)},
	}

	err := service.ValidateSubmitEventInput(input)

	require.NoError(t, err)
}

func TestValidateSubmitEventInput_SameInputTwice_ReturnsIdenticalResult(t *testing.T) {
	// FP: pure function — same input always produces the same output, no hidden state.
	input := validSubmitInput()

	err1 := service.ValidateSubmitEventInput(input)
	err2 := service.ValidateSubmitEventInput(input)

	assert.Equal(t, err1, err2)
}

// --- BuildEventFromInput ---

func TestBuildEventFromInput_SetsStatusPending(t *testing.T) {
	// Spec: events-submit.yaml rule "Event is always created with status=pending — never auto-approved"
	event := service.BuildEventFromInput(validSubmitInput())

	assert.Equal(t, model.EventStatusPending, event.Status)
}

func TestBuildEventFromInput_DerivesYearFromStartDate(t *testing.T) {
	input := validSubmitInput()
	input.StartDate = time.Date(2027, 1, 5, 0, 0, 0, 0, time.UTC)

	event := service.BuildEventFromInput(input)

	assert.Equal(t, 2027, event.Year)
}

func TestBuildEventFromInput_CopiesAllScalarFields(t *testing.T) {
	input := validSubmitInput()

	event := service.BuildEventFromInput(input)

	assert.Equal(t, input.Name, event.Name)
	assert.Equal(t, input.Slug, event.Slug)
	assert.Equal(t, input.Country, event.Country)
	assert.Equal(t, input.City, event.City)
	assert.Equal(t, input.Latitude, event.Latitude)
	assert.Equal(t, input.Longitude, event.Longitude)
	assert.Equal(t, input.StartDate, event.StartDate)
	assert.Equal(t, input.EndDate, event.EndDate)
	assert.Equal(t, input.WebsiteURL, event.WebsiteURL)
	assert.Equal(t, input.Domain, event.Domain)
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

// --- BuildSubmitterFromInput ---

func TestBuildSubmitterFromInput_SetsRoleContributor(t *testing.T) {
	// Spec: events-submit.yaml rule "If no User exists, create one with role=contributor, password_hash=NULL"
	user := service.BuildSubmitterFromInput(service.SubmitterInput{Name: "Ana Silva", Email: "ana@example.com"})

	assert.Equal(t, model.UserRoleContributor, user.Role)
	assert.Nil(t, user.PasswordHash)
	assert.Equal(t, "Ana Silva", user.Name)
	assert.Equal(t, "ana@example.com", user.Email)
}

// --- BuildSubmission ---

func TestBuildSubmission_ComposesEventDeadlinesAndSubmitter(t *testing.T) {
	// FP: function composition — BuildSubmission delegates to the three builders above.
	input := validSubmitInput()
	input.Deadlines = []service.DeadlineInput{
		{Type: "paper", Description: "Research track full paper", Date: time.Date(2026, 8, 22, 0, 0, 0, 0, time.UTC)},
	}

	event, deadlines, submitter := service.BuildSubmission(input)

	assert.Equal(t, input.Name, event.Name)
	assert.Equal(t, model.EventStatusPending, event.Status)
	require.Len(t, deadlines, 1)
	assert.Equal(t, model.DeadlineTypePaper, deadlines[0].Type)
	assert.Equal(t, input.Submitter.Email, submitter.Email)
}

func TestBuildSubmission_SameInputTwice_ReturnsIdenticalResult(t *testing.T) {
	// FP: pure function — same input always produces the same output, no hidden state.
	input := validSubmitInput()

	event1, deadlines1, submitter1 := service.BuildSubmission(input)
	event2, deadlines2, submitter2 := service.BuildSubmission(input)

	assert.Equal(t, event1, event2)
	assert.Equal(t, deadlines1, deadlines2)
	assert.Equal(t, submitter1, submitter2)
}
