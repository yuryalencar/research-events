package service

import (
	"fmt"
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

// allowedDeadlineTypes mirrors the deadlines_type_check CHECK constraint in
// migrations/004_create_deadlines.sql — this is a closed enum.
var allowedDeadlineTypes = map[string]bool{
	"abstract":     true,
	"paper":        true,
	"notification": true,
	"camera_ready": true,
	"other":        true,
}

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
	if input.Submitter.Name == "" {
		return fmt.Errorf("submitter.name is required")
	}
	if !emailPattern.MatchString(input.Submitter.Email) {
		return fmt.Errorf("submitter.email must be a valid email address")
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
	return nil
}
