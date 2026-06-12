package service

import (
	"fmt"
	"net/url"
	"regexp"
	"time"

	"github.com/yuryalencar/research-events/internal/model"
)

// --- Types ---

// SubmitEventInput groups all inputs for the Submit operation, mirroring the
// request body defined in specs/backend/events-submit.yaml.
type SubmitEventInput struct {
	Name       string
	Slug       string
	Country    string
	City       string
	Latitude   float64
	Longitude  float64
	StartDate  time.Time
	EndDate    time.Time
	WebsiteURL string
	Domain     string
	Submitter  SubmitterInput
	Deadlines  []DeadlineInput
}

// SubmitterInput carries the contributor's identity from the submission form.
type SubmitterInput struct {
	Name  string
	Email string
}

// DeadlineInput carries one deadline entry from the submission form.
// Type is a plain string here (not model.DeadlineType) because it must be
// validated against the allowed enum before it can be trusted as a domain value.
type DeadlineInput struct {
	Type        string
	Description string
	Date        time.Time
	IsOptional  bool
}

// --- Package-level validation data ---

// allowedDomains lists the currently accepted values for SubmitEventInput.Domain.
// domain is an extensible enum (see CLAUDE.md) — new values are added here only,
// never as a DB CHECK constraint.
var allowedDomains = map[string]bool{
	"computer_science": true,
}

// allowedDeadlineTypes mirrors the deadlines_type_check CHECK constraint in
// migrations/004_create_deadlines.sql — this is a closed enum.
var allowedDeadlineTypes = map[string]bool{
	"abstract":     true,
	"paper":        true,
	"notification": true,
	"camera_ready": true,
	"other":        true,
}

// slugPattern restricts slugs to URL-safe characters: letters, digits, hyphens, underscores.
var slugPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// emailPattern is a simple format check (not a full RFC 5322 validator) —
// good enough to catch typos without rejecting valid addresses.
var emailPattern = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

// --- Public functions ---

// FP: pure function
// ValidateSubmitEventInput depends only on its argument and performs no I/O.
// Given the same input it always returns the same error (or nil) — there is no
// hidden state, no clock, no database lookup. This makes every validation rule
// trivially testable: no mocks, no setup, just input → output. It prevents the
// class of bugs where validation behaves differently depending on when or how
// many times it runs.
func ValidateSubmitEventInput(input SubmitEventInput) error {
	if input.Name == "" {
		return fmt.Errorf("name is required")
	}
	if input.Slug == "" {
		return fmt.Errorf("slug is required")
	}
	if !slugPattern.MatchString(input.Slug) {
		return fmt.Errorf("slug must contain only letters, digits, hyphens, and underscores")
	}
	if input.Country == "" {
		return fmt.Errorf("country is required")
	}
	if input.City == "" {
		return fmt.Errorf("city is required")
	}
	if input.Latitude < -90 || input.Latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}
	if input.Longitude < -180 || input.Longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}
	if input.EndDate.Before(input.StartDate) {
		return fmt.Errorf("end_date must not be before start_date")
	}
	if !isValidWebsiteURL(input.WebsiteURL) {
		return fmt.Errorf("website_url must be a valid http or https URL")
	}
	if !allowedDomains[input.Domain] {
		return fmt.Errorf("domain %q is not a recognized domain", input.Domain)
	}
	if input.Submitter.Name == "" {
		return fmt.Errorf("submitter.name is required")
	}
	if !emailPattern.MatchString(input.Submitter.Email) {
		return fmt.Errorf("submitter.email must be a valid email address")
	}
	for i, deadline := range input.Deadlines {
		if err := validateDeadlineInput(deadline); err != nil {
			return fmt.Errorf("deadlines[%d]: %w", i, err)
		}
	}
	return nil
}

// FP: immutability
// BuildEventFromInput never modifies input — it reads from it and constructs a
// brand new model.Event value. The caller's input is left untouched, so the same
// SubmitEventInput could be passed to multiple builders (as BuildSubmission does)
// without one builder's output affecting another's.
func BuildEventFromInput(input SubmitEventInput) model.Event {
	return model.Event{
		Name:       input.Name,
		Slug:       input.Slug,
		Country:    input.Country,
		City:       input.City,
		Latitude:   input.Latitude,
		Longitude:  input.Longitude,
		StartDate:  input.StartDate,
		EndDate:    input.EndDate,
		WebsiteURL: input.WebsiteURL,
		Domain:     input.Domain,
		Status:     model.EventStatusPending,
		Year:       input.StartDate.Year(),
	}
}

// FP: immutability
// BuildDeadlinesFromInput returns a new slice of model.Deadline — it never
// appends to or reuses the backing array of input. Each element is a fresh
// value built from the corresponding DeadlineInput, with IsActive always true
// for a newly submitted deadline (see migrations/004_create_deadlines.sql).
func BuildDeadlinesFromInput(input []DeadlineInput) []model.Deadline {
	deadlines := make([]model.Deadline, 0, len(input))
	for _, d := range input {
		deadlines = append(deadlines, model.Deadline{
			Type:        model.DeadlineType(d.Type),
			Description: d.Description,
			Date:        d.Date,
			IsOptional:  d.IsOptional,
			IsActive:    true,
		})
	}
	return deadlines
}

// FP: pure function
// BuildSubmitterFromInput depends only on its argument and returns a new
// model.User value. role=contributor and password_hash=nil are the defaults for
// any user created via public submission — see events-submit.yaml rules.
func BuildSubmitterFromInput(submitter SubmitterInput) model.User {
	return model.User{
		Name:         submitter.Name,
		Email:        submitter.Email,
		PasswordHash: nil,
		Role:         model.UserRoleContributor,
	}
}

// FP: function composition
// BuildSubmission combines BuildEventFromInput, BuildDeadlinesFromInput, and
// BuildSubmitterFromInput into a single call. Composition lets each builder stay
// small and independently testable (as in the tests above) while still giving
// the handler one function to call. None of the three builders depends on the
// others' output, so composing them here introduces no hidden ordering requirements.
func BuildSubmission(input SubmitEventInput) (model.Event, []model.Deadline, model.User) {
	event := BuildEventFromInput(input)
	deadlines := BuildDeadlinesFromInput(input.Deadlines)
	submitter := BuildSubmitterFromInput(input.Submitter)
	return event, deadlines, submitter
}

// --- Private functions ---

// FP: pure function
// isValidWebsiteURL depends only on its argument — no I/O, no network request to
// check the URL actually resolves. It only checks the URL is well-formed and uses
// http or https, which is all the spec requires.
func isValidWebsiteURL(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}
	return (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}

// FP: pure function
// validateDeadlineInput checks a single deadline entry. Returning a wrapped error
// from ValidateSubmitEventInput (with the index) keeps this function focused on
// one deadline at a time — function composition over a loop, rather than one
// large function trying to validate everything at once.
func validateDeadlineInput(d DeadlineInput) error {
	if d.Type == "" {
		return fmt.Errorf("type is required")
	}
	if !allowedDeadlineTypes[d.Type] {
		return fmt.Errorf("type %q is not a recognized deadline type", d.Type)
	}
	if d.Description == "" {
		return fmt.Errorf("description is required")
	}
	if d.Date.IsZero() {
		return fmt.Errorf("date is required")
	}
	return nil
}
