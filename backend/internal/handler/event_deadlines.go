package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/yuryalencar/research-events/internal/model"
	"github.com/yuryalencar/research-events/internal/repository"
	"github.com/yuryalencar/research-events/internal/service"
)

// --- Types ---

// addDeadlinesRequest is the JSON body for POST /api/v1/events/{id}/deadlines.
type addDeadlinesRequest struct {
	Submitter submitterRequest  `json:"submitter"`
	Deadlines []deadlineRequest `json:"deadlines"`
}

// cancelDeadlineRequest is the JSON body for
// PATCH /api/v1/events/{eventId}/deadlines/{deadlineId}/cancel.
type cancelDeadlineRequest struct {
	Submitter submitterRequest `json:"submitter"`
}

// --- Public methods ---

// AddDeadlines handles POST /api/v1/events/{id}/deadlines.
// Public endpoint — no authentication required. Per
// specs/backend/events-deadlines-add.yaml, any contributor can add deadlines
// to an already-approved event.
func (h *EventHandler) AddDeadlines(w http.ResponseWriter, r *http.Request) {
	// A non-numeric :id is reported as 404 EVENT_NOT_FOUND, same as a missing
	// event — the spec requires never revealing whether IDs are numeric.
	rawID := r.PathValue("id")
	eventID64, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil {
		writeError(w, http.StatusNotFound, "EVENT_NOT_FOUND", "event not found")
		return
	}
	eventID := uint(eventID64)

	var req addDeadlinesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	input, err := toAddDeadlinesInput(eventID, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	if err := service.ValidateAddDeadlinesInput(input); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	event, ok := h.findApprovedEvent(w, r, eventID, "AddDeadlines")
	if !ok {
		return
	}

	deadlines := service.BuildDeadlinesFromInput(input.Deadlines)
	auditAction := service.DetermineDeadlinesAuditAction(len(deadlines))
	submitter := service.BuildSubmitterFromInput(input.Submitter)

	updated, err := h.eventRepo.AddDeadlines(r.Context(), event, deadlines, submitter, auditAction)
	if err != nil {
		h.logger.Error("failed to add deadlines", "event_id", eventID, "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	writeSuccess(w, http.StatusCreated, "DEADLINES_ADDED", toEventListItemResponse(updated))
}

// CancelDeadline handles PATCH /api/v1/events/{eventId}/deadlines/{deadlineId}/cancel.
// Public endpoint — no authentication required. Per
// specs/backend/events-deadlines-cancel.yaml, any contributor can cancel a
// deadline on an already-approved event.
func (h *EventHandler) CancelDeadline(w http.ResponseWriter, r *http.Request) {
	// A non-numeric :eventId is reported as 404 EVENT_NOT_FOUND, same as a
	// missing event — the spec requires never revealing whether IDs are numeric.
	eventID64, err := strconv.ParseUint(r.PathValue("eventId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusNotFound, "EVENT_NOT_FOUND", "event not found")
		return
	}
	eventID := uint(eventID64)

	var req cancelDeadlineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	input := service.CancelDeadlineInput{
		Submitter: service.SubmitterInput{
			Name:  req.Submitter.Name,
			Email: req.Submitter.Email,
		},
	}

	if err := service.ValidateCancelDeadlineInput(input); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	event, ok := h.findApprovedEvent(w, r, eventID, "CancelDeadline")
	if !ok {
		return
	}

	// A non-numeric :deadlineId is reported as 404 DEADLINE_NOT_FOUND, checked
	// after the event lookup so a bad eventId is always reported first.
	deadlineID64, err := strconv.ParseUint(r.PathValue("deadlineId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusNotFound, "DEADLINE_NOT_FOUND", "deadline not found")
		return
	}
	deadlineID := uint(deadlineID64)

	deadline, err := h.eventRepo.FindDeadlineByID(r.Context(), eventID, deadlineID)
	switch {
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, "DEADLINE_NOT_FOUND", "deadline not found")
		return
	case err != nil:
		h.logger.Error("failed to fetch deadline for CancelDeadline", "event_id", eventID, "deadline_id", deadlineID, "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	if err := service.ValidateDeadlineCancellable(deadline); err != nil {
		writeError(w, http.StatusConflict, "DEADLINE_ALREADY_INACTIVE", err.Error())
		return
	}

	submitter := service.BuildSubmitterFromInput(input.Submitter)

	updated, err := h.eventRepo.CancelDeadline(r.Context(), event, deadline, submitter)
	if err != nil {
		h.logger.Error("failed to cancel deadline", "event_id", eventID, "deadline_id", deadlineID, "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	writeSuccess(w, http.StatusOK, "DEADLINE_CANCELLED", toEventListItemResponse(updated))
}

// --- Private functions ---

// findApprovedEvent fetches the event with eventID and writes the appropriate
// error response if it does not exist, the lookup fails, or it is not
// status=approved — shared by AddDeadlines and CancelDeadline, which both
// require an approved event before touching its deadlines. logOp identifies
// the calling handler in the error log line. Returns the event and true on
// success; on failure it has already written the response and returns
// (model.Event{}, false).
func (h *EventHandler) findApprovedEvent(w http.ResponseWriter, r *http.Request, eventID uint, logOp string) (model.Event, bool) {
	event, err := h.eventRepo.FindByID(r.Context(), eventID)
	switch {
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, "EVENT_NOT_FOUND", "event not found")
		return model.Event{}, false
	case err != nil:
		h.logger.Error("failed to fetch event for "+logOp, "event_id", eventID, "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return model.Event{}, false
	}

	if event.Status != model.EventStatusApproved {
		writeError(w, http.StatusConflict, "EVENT_NOT_APPROVED", "event must be approved for this operation")
		return model.Event{}, false
	}

	return event, true
}

// toAddDeadlinesInput maps the JSON request into service.AddDeadlinesInput,
// parsing each deadline's date string. Returns an error if any date is malformed.
func toAddDeadlinesInput(eventID uint, req addDeadlinesRequest) (service.AddDeadlinesInput, error) {
	deadlines, err := toDeadlineInputs(req.Deadlines)
	if err != nil {
		return service.AddDeadlinesInput{}, err
	}

	return service.AddDeadlinesInput{
		EventID: eventID,
		Submitter: service.SubmitterInput{
			Name:  req.Submitter.Name,
			Email: req.Submitter.Email,
		},
		Deadlines: deadlines,
	}, nil
}

// toDeadlineInputs maps a request's deadlineRequest entries to
// service.DeadlineInput, parsing each date string — shared by
// toSubmitEventInput (event.go) and toAddDeadlinesInput above. Returns an
// error naming the offending index if any date is malformed.
func toDeadlineInputs(deadlines []deadlineRequest) ([]service.DeadlineInput, error) {
	out := make([]service.DeadlineInput, 0, len(deadlines))
	for i, d := range deadlines {
		date, err := time.Parse(dateLayout, d.Date)
		if err != nil {
			return nil, errors.New("deadlines[" + strconv.Itoa(i) + "].date must be a valid date (YYYY-MM-DD)")
		}
		out = append(out, service.DeadlineInput{
			Type:        d.Type,
			Description: d.Description,
			Date:        date,
			Time:        d.Time,
			Timezone:    d.Timezone,
			IsOptional:  d.IsOptional,
		})
	}
	return out, nil
}
