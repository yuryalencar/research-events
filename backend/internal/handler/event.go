package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/yuryalencar/research-events/internal/model"
	"github.com/yuryalencar/research-events/internal/repository"
	"github.com/yuryalencar/research-events/internal/service"
)

// --- Types ---

// EventHandler holds dependencies for event-related HTTP handlers.
// Dependencies are always injected via constructor — never use globals.
type EventHandler struct {
	eventRepo repository.EventRepository
	logger    *slog.Logger
}

// submitEventRequest is the JSON body for POST /api/v1/events/submit.
type submitEventRequest struct {
	Name       string            `json:"name"`
	Slug       string            `json:"slug"`
	Country    string            `json:"country"`
	City       string            `json:"city"`
	Latitude   float64           `json:"latitude"`
	Longitude  float64           `json:"longitude"`
	StartDate  string            `json:"start_date"`
	EndDate    string            `json:"end_date"`
	WebsiteURL string            `json:"website_url"`
	Domain     string            `json:"domain"`
	Submitter  submitterRequest  `json:"submitter"`
	Deadlines  []deadlineRequest `json:"deadlines"`
}

type submitterRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type deadlineRequest struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Date        string `json:"date"`
	IsOptional  bool   `json:"is_optional"`
}

// dateLayout is the ISO date format used for start_date, end_date, and deadline dates.
const dateLayout = "2006-01-02"

// --- Constructor ---

// NewEventHandler creates a new EventHandler with its required dependencies.
func NewEventHandler(eventRepo repository.EventRepository, logger *slog.Logger) *EventHandler {
	return &EventHandler{
		eventRepo: eventRepo,
		logger:    logger,
	}
}

// --- Public methods (one per route) ---

// Submit handles POST /api/v1/events/submit.
// Public endpoint — no authentication required.
func (h *EventHandler) Submit(w http.ResponseWriter, r *http.Request) {
	var req submitEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	input, err := toSubmitEventInput(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	if err := service.ValidateSubmitEventInput(input); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	_, err = h.eventRepo.FindActiveBySlug(r.Context(), input.Slug)
	switch {
	case err == nil:
		writeError(w, http.StatusConflict, "EVENT_ALREADY_SUBMITTED", "an event with this slug has already been submitted")
		return
	case errors.Is(err, repository.ErrNotFound):
		// Slug is free — continue.
	default:
		h.logger.Error("failed to check slug uniqueness", "slug", input.Slug, "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	event, deadlines, submitter := service.BuildSubmission(input)

	created, err := h.eventRepo.Submit(r.Context(), event, deadlines, submitter)
	if err != nil {
		h.logger.Error("failed to submit event", "slug", input.Slug, "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	writeSuccess(w, http.StatusCreated, "EVENT_SUBMITTED", toEventResponse(created))
}

// --- Private functions ---

// toSubmitEventInput maps the JSON request into service.SubmitEventInput,
// parsing date strings. Returns an error if any date is malformed.
func toSubmitEventInput(req submitEventRequest) (service.SubmitEventInput, error) {
	startDate, err := time.Parse(dateLayout, req.StartDate)
	if err != nil {
		return service.SubmitEventInput{}, errors.New("start_date must be a valid date (YYYY-MM-DD)")
	}
	endDate, err := time.Parse(dateLayout, req.EndDate)
	if err != nil {
		return service.SubmitEventInput{}, errors.New("end_date must be a valid date (YYYY-MM-DD)")
	}

	deadlines := make([]service.DeadlineInput, 0, len(req.Deadlines))
	for i, d := range req.Deadlines {
		date, err := time.Parse(dateLayout, d.Date)
		if err != nil {
			return service.SubmitEventInput{}, errors.New("deadlines[" + strconv.Itoa(i) + "].date must be a valid date (YYYY-MM-DD)")
		}
		deadlines = append(deadlines, service.DeadlineInput{
			Type:        d.Type,
			Description: d.Description,
			Date:        date,
			IsOptional:  d.IsOptional,
		})
	}

	return service.SubmitEventInput{
		Name:       req.Name,
		Slug:       req.Slug,
		Country:    req.Country,
		City:       req.City,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
		StartDate:  startDate,
		EndDate:    endDate,
		WebsiteURL: req.WebsiteURL,
		Domain:     req.Domain,
		Submitter: service.SubmitterInput{
			Name:  req.Submitter.Name,
			Email: req.Submitter.Email,
		},
		Deadlines: deadlines,
	}, nil
}

// userResponse is the subset of User fields exposed in the submission response.
type userResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// deadlineResponse is one entry in the event response's deadlines array.
type deadlineResponse struct {
	ID          uint   `json:"id"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Date        string `json:"date"`
	IsOptional  bool   `json:"is_optional"`
	IsActive    bool   `json:"is_active"`
}

// eventResponse is the "data" payload for a successful submission, per
// specs/backend/events-submit.yaml responses.201.body.data.
type eventResponse struct {
	ID         uint               `json:"id"`
	Name       string             `json:"name"`
	Slug       string             `json:"slug"`
	Country    string             `json:"country"`
	City       string             `json:"city"`
	Latitude   float64            `json:"latitude"`
	Longitude  float64            `json:"longitude"`
	StartDate  string             `json:"start_date"`
	EndDate    string             `json:"end_date"`
	WebsiteURL string             `json:"website_url"`
	Domain     string             `json:"domain"`
	Status     string             `json:"status"`
	CreatedBy  userResponse       `json:"created_by"`
	Deadlines  []deadlineResponse `json:"deadlines"`
	CreatedAt  string             `json:"created_at"`
}

// toEventResponse maps a persisted model.Event to its JSON response shape.
func toEventResponse(e model.Event) eventResponse {
	deadlines := make([]deadlineResponse, 0, len(e.Deadlines))
	for _, d := range e.Deadlines {
		deadlines = append(deadlines, deadlineResponse{
			ID:          d.ID,
			Type:        string(d.Type),
			Description: d.Description,
			Date:        d.Date.Format(dateLayout),
			IsOptional:  d.IsOptional,
			IsActive:    d.IsActive,
		})
	}

	return eventResponse{
		ID:         e.ID,
		Name:       e.Name,
		Slug:       e.Slug,
		Country:    e.Country,
		City:       e.City,
		Latitude:   e.Latitude,
		Longitude:  e.Longitude,
		StartDate:  e.StartDate.Format(dateLayout),
		EndDate:    e.EndDate.Format(dateLayout),
		WebsiteURL: e.WebsiteURL,
		Domain:     e.Domain,
		Status:     string(e.Status),
		CreatedBy: userResponse{
			ID:    e.CreatedBy.ID,
			Name:  e.CreatedBy.Name,
			Email: e.CreatedBy.Email,
		},
		Deadlines: deadlines,
		CreatedAt: e.CreatedAt.Format(time.RFC3339),
	}
}
