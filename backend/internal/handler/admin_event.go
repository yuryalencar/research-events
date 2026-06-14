package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"gorm.io/gorm"

	"github.com/yuryalencar/research-events/internal/middleware"
	"github.com/yuryalencar/research-events/internal/model"
	"github.com/yuryalencar/research-events/internal/repository"
	"github.com/yuryalencar/research-events/internal/service"
)

// --- Types ---

// AdminEventHandler handles admin-scoped event review endpoints.
// Must be protected by RequireAuth + RequireRole("admin", "moderator") in the
// route registration.
type AdminEventHandler struct {
	eventRepo repository.EventRepository
	logger    *slog.Logger
}

// reviewEventRequest is the JSON body for PATCH /api/v1/admin/events/{id}/review.
type reviewEventRequest struct {
	Action string            `json:"action"`
	Reason *string           `json:"reason"`
	Event  *eventEditRequest `json:"event"`
}

// eventEditRequest is the optional partial update for the event's own fields,
// per specs/backend/admin-events-review.yaml. A nil pointer means "leave this
// field unchanged".
type eventEditRequest struct {
	Name       *string  `json:"name"`
	Slug       *string  `json:"slug"`
	Country    *string  `json:"country"`
	City       *string  `json:"city"`
	Latitude   *float64 `json:"latitude"`
	Longitude  *float64 `json:"longitude"`
	StartDate  *string  `json:"start_date"`
	EndDate    *string  `json:"end_date"`
	WebsiteURL *string  `json:"website_url"`
	Domain     *string  `json:"domain"`
	Tier       *string  `json:"tier"`
}

// --- Constructor ---

// NewAdminEventHandler creates a new AdminEventHandler with its required dependencies.
func NewAdminEventHandler(eventRepo repository.EventRepository, logger *slog.Logger) *AdminEventHandler {
	return &AdminEventHandler{
		eventRepo: eventRepo,
		logger:    logger,
	}
}

// --- Public methods (one per route) ---

// Review handles PATCH /api/v1/admin/events/{id}/review.
// Requires admin or moderator role — enforced at route level via
// RequireAuth + RequireRole("admin", "moderator").
func (h *AdminEventHandler) Review(w http.ResponseWriter, r *http.Request) {
	// 1. Parse :id from the URL path parameter.
	rawID := r.PathValue("id")
	id64, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", ":id must be a valid integer")
		return
	}
	id := uint(id64)

	// 2. Get the authenticated reviewer from context (set by RequireAuth middleware).
	authUser, ok := middleware.AuthUserFromContext(r.Context())
	if !ok {
		// Should never happen when RequireAuth is in the chain — guard for safety.
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	// 3. Parse and validate the request body.
	var req reviewEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	edit, err := toEventEditInput(req.Event)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	input := service.ReviewEventInput{Action: req.Action, Reason: req.Reason, Event: edit}
	if err := service.ValidateReviewActionInput(input); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// 4. Fetch the event being reviewed.
	existing, err := h.eventRepo.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "EVENT_NOT_FOUND", "event not found")
			return
		}
		h.logger.Error("failed to fetch event for review", "event_id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	// 5. Moderators cannot review an event they submitted themselves — admins are exempt.
	if authUser.Role == string(model.UserRoleModerator) && existing.CreatedByID == authUser.ID {
		writeError(w, http.StatusForbidden, "CANNOT_REVIEW_OWN_EVENT", "moderators cannot review an event they submitted themselves")
		return
	}

	reviewer := model.User{Model: gorm.Model{ID: authUser.ID}, Role: model.UserRole(authUser.Role)}
	updated := service.ApplyReview(existing, input, reviewer)

	// 6. Validate the edited event fields (same rules as submission).
	if input.Event != nil {
		if err := service.ValidateEditedEvent(updated); err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
	}

	// 7. If the slug changed, it must not collide with another pending/approved event.
	if updated.Slug != existing.Slug {
		other, err := h.eventRepo.FindActiveBySlug(r.Context(), updated.Slug)
		switch {
		case err == nil:
			if other.ID != existing.ID {
				writeError(w, http.StatusConflict, "SLUG_ALREADY_EXISTS", "an event with this slug already exists")
				return
			}
		case errors.Is(err, repository.ErrNotFound):
			// Slug is free — continue.
		default:
			h.logger.Error("failed to check slug uniqueness", "slug", updated.Slug, "error", err)
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
			return
		}
	}

	// 8. Compute the AuditLog row (pure computation → persist).
	auditLog, err := service.BuildReviewAuditLog(existing, updated, reviewer, input.Reason)
	if err != nil {
		h.logger.Error("failed to build review audit log", "event_id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	result, err := h.eventRepo.Review(r.Context(), updated, auditLog)
	if err != nil {
		h.logger.Error("failed to persist event review", "event_id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	writeSuccess(w, http.StatusOK, "EVENT_REVIEWED", toEventListItemResponse(result))
}

// --- Private functions ---

// toEventEditInput maps the optional "event" object into service.EventEditInput,
// parsing date strings. Returns (nil, nil) when req is nil — meaning no event
// fields were submitted for editing.
func toEventEditInput(req *eventEditRequest) (*service.EventEditInput, error) {
	if req == nil {
		return nil, nil
	}

	edit := &service.EventEditInput{
		Name:       req.Name,
		Slug:       req.Slug,
		Country:    req.Country,
		City:       req.City,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
		WebsiteURL: req.WebsiteURL,
		Domain:     req.Domain,
		Tier:       req.Tier,
	}

	if req.StartDate != nil {
		startDate, err := parseDateField("start_date", *req.StartDate)
		if err != nil {
			return nil, err
		}
		edit.StartDate = &startDate
	}
	if req.EndDate != nil {
		endDate, err := parseDateField("end_date", *req.EndDate)
		if err != nil {
			return nil, err
		}
		edit.EndDate = &endDate
	}

	return edit, nil
}
