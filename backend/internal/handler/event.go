package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
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
	Tier       string            `json:"tier"`
	Submitter  submitterRequest  `json:"submitter"`
	Deadlines  []deadlineRequest `json:"deadlines"`
}

type submitterRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type deadlineRequest struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
	Time        *string `json:"time,omitempty"`
	Timezone    *string `json:"timezone,omitempty"`
	IsOptional  bool    `json:"is_optional"`
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

// List handles GET /api/v1/events.
// Public endpoint — no authentication required.
func (h *EventHandler) List(w http.ResponseWriter, r *http.Request) {
	raw := toRawListEventsQuery(r.URL.Query())

	input, err := service.ValidateListEventsQuery(raw, time.Now().Year())
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	events, total, err := h.eventRepo.ListEvents(r.Context(), toListEventsFilter(input))
	if err != nil {
		h.logger.Error("failed to list events", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	data := make([]eventListItemResponse, 0, len(events))
	for _, e := range events {
		data = append(data, toEventListItemResponse(e))
	}

	writeSuccessWithMeta(w, http.StatusOK, "EVENTS_LISTED", data, listMeta{Page: input.Page, Total: total})
}

// --- Private functions ---

// toRawListEventsQuery copies the relevant query string parameters into
// service.RawListEventsQuery. No parsing or validation happens here — that is
// ValidateListEventsQuery's job, so it stays testable without net/http.
func toRawListEventsQuery(q url.Values) service.RawListEventsQuery {
	return service.RawListEventsQuery{
		Year:               q.Get("year"),
		Domain:             q.Get("domain"),
		Country:            q.Get("country"),
		Status:             q.Get("status"),
		Tier:               q.Get("tier"),
		FirstDeadlineMonth: q.Get("first_deadline_month"),
		BBox:               q.Get("bbox"),
		Page:               q.Get("page"),
		PageSize:           q.Get("page_size"),
		Pagination:         q.Get("pagination"),
	}
}

// toListEventsFilter maps the validated service.ListEventsInput to
// repository.ListEventsFilter.
func toListEventsFilter(input service.ListEventsInput) repository.ListEventsFilter {
	var bbox *repository.BBoxFilter
	if input.BBox != nil {
		bbox = &repository.BBoxFilter{
			MinLng: input.BBox.MinLng,
			MinLat: input.BBox.MinLat,
			MaxLng: input.BBox.MaxLng,
			MaxLat: input.BBox.MaxLat,
		}
	}

	return repository.ListEventsFilter{
		Year:               input.Year,
		Status:             input.Status,
		Domain:             input.Domain,
		Country:            input.Country,
		Tier:               input.Tier,
		FirstDeadlineMonth: input.FirstDeadlineMonth,
		BBox:               bbox,
		Page:               input.Page,
		PageSize:           input.PageSize,
		PaginationOff:      input.PaginationOff,
	}
}

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

	deadlines, err := toDeadlineInputs(req.Deadlines)
	if err != nil {
		return service.SubmitEventInput{}, err
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
		Tier:       req.Tier,
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
	ID             uint    `json:"id"`
	Type           string  `json:"type"`
	Description    string  `json:"description"`
	Date           string  `json:"date"`
	Time           *string `json:"time"`
	Timezone       *string `json:"timezone"`
	IsOptional     bool    `json:"is_optional"`
	IsActive       bool    `json:"is_active"`
	SupersededByID *uint   `json:"superseded_by_id"`
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
	Tier       string             `json:"tier"`
	Status     string             `json:"status"`
	CreatedBy  userResponse       `json:"created_by"`
	Deadlines  []deadlineResponse `json:"deadlines"`
	CreatedAt  string             `json:"created_at"`
}

// toEventResponse maps a persisted model.Event to its JSON response shape.
func toEventResponse(e model.Event) eventResponse {
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
		Tier:       e.Tier,
		Status:     string(e.Status),
		CreatedBy: userResponse{
			ID:    e.CreatedBy.ID,
			Name:  e.CreatedBy.Name,
			Email: e.CreatedBy.Email,
		},
		Deadlines: toDeadlineResponses(e.Deadlines),
		CreatedAt: e.CreatedAt.Format(time.RFC3339),
	}
}

// toDeadlineResponses maps a slice of model.Deadline to their JSON response
// shape — shared by toEventResponse and toEventListItemResponse.
func toDeadlineResponses(deadlines []model.Deadline) []deadlineResponse {
	out := make([]deadlineResponse, 0, len(deadlines))
	for _, d := range deadlines {
		out = append(out, deadlineResponse{
			ID:             d.ID,
			Type:           string(d.Type),
			Description:    d.Description,
			Date:           d.Date.Format(dateLayout),
			Time:           d.Time,
			Timezone:       d.Timezone,
			IsOptional:     d.IsOptional,
			IsActive:       d.IsActive,
			SupersededByID: d.SupersededByID,
		})
	}
	return out
}

// listMeta is the "meta" payload for GET /api/v1/events, per
// specs/backend/events-list.yaml responses.200.body.meta.
type listMeta struct {
	Page  int   `json:"page"`
	Total int64 `json:"total"`
}

// eventListItemResponse is one entry of the "data" array for
// GET /api/v1/events, per specs/backend/events-list.yaml responses.200.body.data.
type eventListItemResponse struct {
	ID            uint               `json:"id"`
	Name          string             `json:"name"`
	Slug          string             `json:"slug"`
	Country       string             `json:"country"`
	City          string             `json:"city"`
	Latitude      float64            `json:"latitude"`
	Longitude     float64            `json:"longitude"`
	StartDate     string             `json:"start_date"`
	EndDate       string             `json:"end_date"`
	WebsiteURL    string             `json:"website_url"`
	Domain        string             `json:"domain"`
	Status        string             `json:"status"`
	Tier          string             `json:"tier"`
	Year          int                `json:"year"`
	CreatedBy     userResponse       `json:"created_by"`
	LastUpdatedBy userResponse       `json:"last_updated_by"`
	Deadlines     []deadlineResponse `json:"deadlines"`
	CreatedAt     string             `json:"created_at"`
	UpdatedAt     string             `json:"updated_at"`
}

// toEventListItemResponse maps a persisted model.Event to its JSON response
// shape for GET /api/v1/events.
func toEventListItemResponse(e model.Event) eventListItemResponse {
	return eventListItemResponse{
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
		Tier:       e.Tier,
		Year:       e.Year,
		CreatedBy: userResponse{
			ID:    e.CreatedBy.ID,
			Name:  e.CreatedBy.Name,
			Email: e.CreatedBy.Email,
		},
		LastUpdatedBy: userResponse{
			ID:    e.LastUpdatedBy.ID,
			Name:  e.LastUpdatedBy.Name,
			Email: e.LastUpdatedBy.Email,
		},
		Deadlines: toDeadlineResponses(e.Deadlines),
		CreatedAt: e.CreatedAt.Format(time.RFC3339),
		UpdatedAt: e.UpdatedAt.Format(time.RFC3339),
	}
}
