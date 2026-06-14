package service

import (
	"fmt"
	"regexp"
	"time"

	"github.com/yuryalencar/research-events/internal/model"
)

// --- Types ---

// AddDeadlinesInput groups all inputs for adding one or more deadlines to an
// already-approved event, mirroring the request body defined in
// specs/backend/events-deadlines-add.yaml.
type AddDeadlinesInput struct {
	EventID   uint
	Submitter SubmitterInput
	Deadlines []DeadlineInput
}

// SupersedeDeadlineInput groups all inputs for superseding a single deadline
// on an already-approved event, mirroring the request body defined in
// specs/backend/events-deadlines-supersede.yaml. EventID and DeadlineID come
// from path params, validated by repository lookups — only the request
// body's submitter, date, time, and timezone are validated here. Per the
// spec, time and timezone always come from the request and are never
// inherited from the deadline being superseded.
type SupersedeDeadlineInput struct {
	Submitter SubmitterInput
	Date      time.Time
	Time      *string
	Timezone  *string
}

// CancelDeadlineInput groups all inputs for cancelling a single deadline on an
// already-approved event, mirroring the request body defined in
// specs/backend/events-deadlines-cancel.yaml. EventID and DeadlineID come from
// path params, not the body — the handler validates those separately by
// looking the event and deadline up, so only Submitter needs validation here.
type CancelDeadlineInput struct {
	Submitter SubmitterInput
}

// DeadlineInput carries one deadline entry from the submission form.
// Type is a plain string here (not model.DeadlineType) because it must be
// validated against the allowed enum before it can be trusted as a domain value.
type DeadlineInput struct {
	Type        string
	Description string
	Date        time.Time

	// Time and Timezone are independently optional — see
	// specs/backend/deadlines-add-time-timezone.yaml. Time uses 24h "HH:MM"
	// (e.g. "23:59"); Timezone is a free string (e.g. "AoE", "UTC-3").
	Time     *string
	Timezone *string

	IsOptional bool
}

// --- Package-level validation data ---

// allowedDeadlineTypes mirrors the deadlines_type_check CHECK constraint in
// migrations/004_create_deadlines.sql — this is a closed enum.
var allowedDeadlineTypes = map[string]bool{
	"abstract":     true,
	"paper":        true,
	"notification": true,
	"camera_ready": true,
	"other":        true,
}

// deadlineTimePattern matches a 24-hour, zero-padded "HH:MM" string
// (00:00-23:59), per specs/backend/deadlines-add-time-timezone.yaml rule
// "time, if provided, must match HH:MM 24-hour format (00-23 : 00-59), zero-padded".
var deadlineTimePattern = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)

// --- Public functions ---

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
			Time:        d.Time,
			Timezone:    d.Timezone,
			IsOptional:  d.IsOptional,
			IsActive:    true,
		})
	}
	return deadlines
}

// FP: pure function
// ValidateAddDeadlinesInput depends only on its argument and performs no I/O —
// no database lookup to check the event exists or is approved (that happens in
// the repository layer, which has access to the DB). Given the same input it
// always returns the same error (or nil), which is what makes it trivially
// testable: no mocks, no setup, just input → output. This mirrors
// ValidateSubmitEventInput in event.go, scoped to the fields
// events-deadlines-add.yaml actually requires (submitter + deadlines).
func ValidateAddDeadlinesInput(input AddDeadlinesInput) error {
	if err := validateSubmitterInput(input.Submitter); err != nil {
		return err
	}
	if len(input.Deadlines) == 0 {
		return fmt.Errorf("deadlines must contain at least one entry")
	}
	for i, deadline := range input.Deadlines {
		if err := validateDeadlineInput(deadline); err != nil {
			return fmt.Errorf("deadlines[%d]: %w", i, err)
		}
	}
	return nil
}

// FP: pure function
// DetermineDeadlinesAuditAction depends only on count and always returns the same
// model.AuditAction for the same count — no I/O, no lookup of the actual deadlines.
// It encodes the events-deadlines-add.yaml rule that a single new deadline gets its
// own AuditActionDeadlineAdded row, while a batch of two or more is recorded as one
// AuditActionBatchDeadlinesAdded row instead of one row per deadline.
func DetermineDeadlinesAuditAction(count int) model.AuditAction {
	if count == 1 {
		return model.AuditActionDeadlineAdded
	}
	return model.AuditActionBatchDeadlinesAdded
}

// FP: pure function
// ValidateCancelDeadlineInput depends only on its argument and performs no I/O —
// no database lookup to check the event/deadline exist or the deadline is still
// active (those checks happen in the repository layer, which is the only layer
// with DB access). Given the same input it always returns the same error (or
// nil), so it is trivially testable: no mocks, no setup, just input → output.
// This mirrors ValidateAddDeadlinesInput, scoped to the only field
// events-deadlines-cancel.yaml's request body has: submitter.
func ValidateCancelDeadlineInput(input CancelDeadlineInput) error {
	if err := validateSubmitterInput(input.Submitter); err != nil {
		return err
	}
	return nil
}

// FP: pure function
// ValidateDeadlineCancellable depends only on its argument and performs no I/O.
// Given the same Deadline it always returns the same error (or nil) — it does
// not re-fetch the deadline or check anything beyond the IsActive flag already
// loaded by the repository. is_active=false covers two cases the caller cannot
// tell apart from this function alone: the deadline was already cancelled
// (SupersededByID == nil), or it was already superseded by a newer deadline
// (SupersededByID != nil) — both are "already inactive" and cannot be
// cancelled again.
func ValidateDeadlineCancellable(d model.Deadline) error {
	if !d.IsActive {
		return fmt.Errorf("deadline is already inactive")
	}
	return nil
}

// FP: pure function
// ValidateSupersedeDeadlineInput depends only on its argument and performs no
// I/O — no database lookup to confirm the event/deadline exist or that the
// deadline is still active (those checks belong to the repository layer,
// which alone has DB access). Given the same input it always returns the
// same error (or nil), so it is trivially testable: no mocks, no setup, just
// input → output. The time/timezone checks reuse deadlineTimePattern and the
// "empty timezone is invalid" rule already shared with
// ValidateAddDeadlinesInput.
func ValidateSupersedeDeadlineInput(input SupersedeDeadlineInput) error {
	if err := validateSubmitterInput(input.Submitter); err != nil {
		return err
	}
	if input.Date.IsZero() {
		return fmt.Errorf("date is required")
	}
	if input.Time != nil && !deadlineTimePattern.MatchString(*input.Time) {
		return fmt.Errorf("time must be in HH:MM 24-hour format")
	}
	if input.Timezone != nil && *input.Timezone == "" {
		return fmt.Errorf("timezone must not be empty")
	}
	return nil
}

// FP: immutability
// BuildSupersedingDeadline returns a brand new model.Deadline — it never
// mutates old, the Deadline being replaced. Per
// specs/backend/events-deadlines-supersede.yaml, the new row inherits Type,
// Description, and IsOptional from old unchanged, but Date/Time/Timezone come
// only from input (never copied from old, even if input omits them — that is
// what keeps the "never inherit time/timezone" rule a one-way data flow
// instead of a conditional fallback). The new row always starts
// IsActive=true with SupersededByID=nil; the caller (repository layer) is
// responsible for marking old as is_active=false/superseded_by_id=<new ID>
// once the new row has been persisted and its ID is known.
func BuildSupersedingDeadline(old model.Deadline, input SupersedeDeadlineInput) model.Deadline {
	return model.Deadline{
		Type:        old.Type,
		Description: old.Description,
		Date:        input.Date,
		Time:        input.Time,
		Timezone:    input.Timezone,
		IsOptional:  old.IsOptional,
		IsActive:    true,
	}
}

// --- Private functions ---

// FP: pure function
// validateDeadlineInput checks a single deadline entry. Returning a wrapped error
// from the caller (with the index) keeps this function focused on
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
	if d.Time != nil && !deadlineTimePattern.MatchString(*d.Time) {
		return fmt.Errorf("time must be in HH:MM 24-hour format")
	}
	if d.Timezone != nil && *d.Timezone == "" {
		return fmt.Errorf("timezone must not be empty")
	}
	return nil
}
