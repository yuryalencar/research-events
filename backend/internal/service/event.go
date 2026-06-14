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
	Tier       string // optional; "" normalizes to "unranked" (see normalizeTier)
	Submitter  SubmitterInput
	Deadlines  []DeadlineInput
}

// SubmitterInput carries the contributor's identity from the submission form.
type SubmitterInput struct {
	Name  string
	Email string
}

// --- Package-level validation data ---

// allowedDomains lists the currently accepted values for SubmitEventInput.Domain.
// domain is an extensible enum (see CLAUDE.md) — new values are added here only,
// never as a DB CHECK constraint.
var allowedDomains = map[string]bool{
	"computer_science": true,
}

// allowedTiers mirrors the events_tier_check CHECK constraint in
// migrations/006_add_tier_to_events.sql — this is a closed enum.
var allowedTiers = map[string]bool{
	"A*":       true,
	"A":        true,
	"B":        true,
	"C":        true,
	"unranked": true,
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
	if err := validateEventFields(eventFieldsInput{
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
		Tier:       normalizeTier(input.Tier),
	}); err != nil {
		return err
	}
	if err := validateSubmitterInput(input.Submitter); err != nil {
		return err
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
		Tier:       normalizeTier(input.Tier),
		Status:     model.EventStatusPending,
		Year:       input.StartDate.Year(),
	}
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
// normalizeTier depends only on its argument and always returns the same result
// for the same input. An empty tier (the zero value, meaning "not provided in
// the request") normalizes to "unranked" so validation and persistence never
// have to treat "" as a separate case from the explicit "unranked" value.
func normalizeTier(tier string) string {
	if tier == "" {
		return "unranked"
	}
	return tier
}

// eventFieldsInput groups the Event-level fields that are validated identically
// regardless of where they came from — a new submission (ValidateSubmitEventInput)
// or an admin review edit (ValidateEditedEvent, in event_review.go). Keeping this
// validation in one place means both callers always agree on what a "valid event"
// looks like. Tier must already be normalized (see normalizeTier) before calling.
type eventFieldsInput struct {
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
	Tier       string
}

// FP: pure function
// validateEventFields depends only on its argument and performs no I/O. Given
// the same eventFieldsInput it always returns the same error (or nil) — extracted
// from ValidateSubmitEventInput so the exact same per-field rules apply to admin
// review edits via ValidateEditedEvent, without duplicating any of these checks.
func validateEventFields(f eventFieldsInput) error {
	if f.Name == "" {
		return fmt.Errorf("name is required")
	}
	if f.Slug == "" {
		return fmt.Errorf("slug is required")
	}
	if !slugPattern.MatchString(f.Slug) {
		return fmt.Errorf("slug must contain only letters, digits, hyphens, and underscores")
	}
	if f.Country == "" {
		return fmt.Errorf("country is required")
	}
	if f.City == "" {
		return fmt.Errorf("city is required")
	}
	if f.Latitude < -90 || f.Latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}
	if f.Longitude < -180 || f.Longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}
	if f.EndDate.Before(f.StartDate) {
		return fmt.Errorf("end_date must not be before start_date")
	}
	if !isValidWebsiteURL(f.WebsiteURL) {
		return fmt.Errorf("website_url must be a valid http or https URL")
	}
	if !allowedDomains[f.Domain] {
		return fmt.Errorf("domain %q is not a recognized domain", f.Domain)
	}
	if !allowedTiers[f.Tier] {
		return fmt.Errorf("tier %q is not a recognized tier", f.Tier)
	}
	return nil
}

// FP: pure function
// validateSubmitterInput depends only on its argument and performs no I/O — no
// lookup of whether the email already belongs to a User (that happens in the
// repository layer's findOrCreateSubmitter). Shared by ValidateSubmitEventInput,
// ValidateAddDeadlinesInput, and ValidateCancelDeadlineInput, since every
// endpoint that accepts a `submitter: {name, email}` body validates it the same way.
func validateSubmitterInput(s SubmitterInput) error {
	if s.Name == "" {
		return fmt.Errorf("submitter.name is required")
	}
	if !emailPattern.MatchString(s.Email) {
		return fmt.Errorf("submitter.email must be a valid email address")
	}
	return nil
}

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
