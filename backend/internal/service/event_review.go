package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/yuryalencar/research-events/internal/model"
)

// --- Types ---

// EventEditInput carries optional per-field overrides for the "event" object in
// PATCH /api/v1/admin/events/{id}/review's request body, per
// specs/backend/admin-events-review.yaml. A nil pointer means "leave this field
// unchanged" — only non-nil fields are applied by ApplyReview.
type EventEditInput struct {
	Name       *string
	Slug       *string
	Country    *string
	City       *string
	Latitude   *float64
	Longitude  *float64
	StartDate  *time.Time
	EndDate    *time.Time
	WebsiteURL *string
	Domain     *string
	Tier       *string
}

// ReviewEventInput groups all inputs for reviewing an event, mirroring the
// request body defined in specs/backend/admin-events-review.yaml.
type ReviewEventInput struct {
	Action string // "approve" | "reject"
	Reason *string
	Event  *EventEditInput
}

// --- Package-level validation data ---

// reviewDateLayout formats dates inside AuditLog diffs the same way the handler
// formats dates in JSON responses (YYYY-MM-DD) — keeps diffs human-readable.
const reviewDateLayout = "2006-01-02"

// --- Public functions ---

// FP: pure function
// ValidateReviewActionInput depends only on its argument and performs no I/O.
// Given the same input it always returns the same error (or nil): "action" must
// be "approve" or "reject", and a non-empty "reason" is required when rejecting.
func ValidateReviewActionInput(input ReviewEventInput) error {
	switch input.Action {
	case "approve":
		return nil
	case "reject":
		if input.Reason == nil || *input.Reason == "" {
			return fmt.Errorf("reason is required when rejecting an event")
		}
		return nil
	default:
		return fmt.Errorf("action must be 'approve' or 'reject'")
	}
}

// FP: immutability
// ApplyReview never mutates existing — it returns a brand new model.Event with
// any fields from input.Event applied on top of existing, plus the new Status
// (derived from input.Action) and LastUpdatedByID (the reviewer). The caller's
// existing value is left untouched, so it can still be used as the "before"
// side of an audit diff after this call.
func ApplyReview(existing model.Event, input ReviewEventInput, reviewer model.User) model.Event {
	updated := existing
	if input.Event != nil {
		updated = applyEventEdits(updated, *input.Event)
	}

	if input.Action == "approve" {
		updated.Status = model.EventStatusApproved
	} else {
		updated.Status = model.EventStatusRejected
	}
	updated.LastUpdatedByID = reviewer.ID

	return updated
}

// FP: pure function
// ValidateEditedEvent depends only on its argument and performs no I/O. It
// applies the exact same per-field rules as ValidateSubmitEventInput (via the
// shared validateEventFields helper), so an event edited during admin review
// must meet the same bar as a brand new submission.
func ValidateEditedEvent(event model.Event) error {
	return validateEventFields(eventFieldsInput{
		Name:       event.Name,
		Slug:       event.Slug,
		Country:    event.Country,
		City:       event.City,
		Latitude:   event.Latitude,
		Longitude:  event.Longitude,
		StartDate:  event.StartDate,
		EndDate:    event.EndDate,
		WebsiteURL: event.WebsiteURL,
		Domain:     event.Domain,
		Tier:       normalizeTier(event.Tier),
	})
}

// FP: no side effects
// BuildReviewAuditLog only computes the AuditLog row to persist — it does not
// write it. The diff is built from existing vs. updated; the caller decides
// when and how to call AuditRepository.Create (or pass it to
// EventRepository.Review, which writes it in the same transaction as the
// event update).
func BuildReviewAuditLog(existing model.Event, updated model.Event, reviewer model.User, reason *string) (model.AuditLog, error) {
	diff, err := json.Marshal(buildReviewDiff(existing, updated))
	if err != nil {
		return model.AuditLog{}, err
	}

	action := model.AuditActionRejected
	if updated.Status == model.EventStatusApproved {
		action = model.AuditActionApproved
	}

	return model.AuditLog{
		EntityType:  model.AuditEntityEvent,
		EntityID:    existing.ID,
		Action:      action,
		ChangedByID: reviewer.ID,
		Diff:        model.JSONB(diff),
		Reason:      reason,
	}, nil
}

// --- Private functions ---

// FP: immutability
// applyEventEdits never mutates existing — it returns a brand new model.Event
// with only the non-nil fields from edits overridden. Editing StartDate also
// recomputes Year (mirrors BuildEventFromInput at submission), since Year drives
// the public list's default filter and must stay consistent with StartDate.
func applyEventEdits(existing model.Event, edits EventEditInput) model.Event {
	updated := existing

	if edits.Name != nil {
		updated.Name = *edits.Name
	}
	if edits.Slug != nil {
		updated.Slug = *edits.Slug
	}
	if edits.Country != nil {
		updated.Country = *edits.Country
	}
	if edits.City != nil {
		updated.City = *edits.City
	}
	if edits.Latitude != nil {
		updated.Latitude = *edits.Latitude
	}
	if edits.Longitude != nil {
		updated.Longitude = *edits.Longitude
	}
	if edits.StartDate != nil {
		updated.StartDate = *edits.StartDate
		updated.Year = edits.StartDate.Year()
	}
	if edits.EndDate != nil {
		updated.EndDate = *edits.EndDate
	}
	if edits.WebsiteURL != nil {
		updated.WebsiteURL = *edits.WebsiteURL
	}
	if edits.Domain != nil {
		updated.Domain = *edits.Domain
	}
	if edits.Tier != nil {
		updated.Tier = *edits.Tier
	}

	return updated
}

// FP: pure function
// buildReviewDiff depends only on its arguments and performs no I/O. It always
// includes "status" (even when unchanged, e.g. re-approving an already-approved
// event), and includes every other field only when it actually changed between
// existing and updated — so the AuditLog diff stays small and readable.
func buildReviewDiff(existing, updated model.Event) map[string]any {
	diff := map[string]any{
		"status": map[string]any{"before": existing.Status, "after": updated.Status},
	}

	if existing.Name != updated.Name {
		diff["name"] = map[string]any{"before": existing.Name, "after": updated.Name}
	}
	if existing.Slug != updated.Slug {
		diff["slug"] = map[string]any{"before": existing.Slug, "after": updated.Slug}
	}
	if existing.Country != updated.Country {
		diff["country"] = map[string]any{"before": existing.Country, "after": updated.Country}
	}
	if existing.City != updated.City {
		diff["city"] = map[string]any{"before": existing.City, "after": updated.City}
	}
	if existing.Latitude != updated.Latitude {
		diff["latitude"] = map[string]any{"before": existing.Latitude, "after": updated.Latitude}
	}
	if existing.Longitude != updated.Longitude {
		diff["longitude"] = map[string]any{"before": existing.Longitude, "after": updated.Longitude}
	}
	if !existing.StartDate.Equal(updated.StartDate) {
		diff["start_date"] = map[string]any{
			"before": existing.StartDate.Format(reviewDateLayout),
			"after":  updated.StartDate.Format(reviewDateLayout),
		}
		diff["year"] = map[string]any{"before": existing.Year, "after": updated.Year}
	}
	if !existing.EndDate.Equal(updated.EndDate) {
		diff["end_date"] = map[string]any{
			"before": existing.EndDate.Format(reviewDateLayout),
			"after":  updated.EndDate.Format(reviewDateLayout),
		}
	}
	if existing.WebsiteURL != updated.WebsiteURL {
		diff["website_url"] = map[string]any{"before": existing.WebsiteURL, "after": updated.WebsiteURL}
	}
	if existing.Domain != updated.Domain {
		diff["domain"] = map[string]any{"before": existing.Domain, "after": updated.Domain}
	}
	if existing.Tier != updated.Tier {
		diff["tier"] = map[string]any{"before": existing.Tier, "after": updated.Tier}
	}

	return diff
}
