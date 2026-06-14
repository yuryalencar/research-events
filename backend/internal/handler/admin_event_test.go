package handler_test

// Spec: specs/backend/admin-events-review.yaml

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/yuryalencar/research-events/internal/handler"
	"github.com/yuryalencar/research-events/internal/middleware"
	"github.com/yuryalencar/research-events/internal/model"
	"github.com/yuryalencar/research-events/internal/repository"
	"github.com/yuryalencar/research-events/internal/repository/mocks"
)

// reviewMux registers the Review handler on its real route pattern so that
// r.PathValue("id") is populated by the Go 1.22 ServeMux.
func reviewMux(h *handler.AdminEventHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /api/v1/admin/events/{id}/review", h.Review)
	return mux
}

// reviewReq builds a PATCH request with body and the reviewer's AuthUser injected into context.
func reviewReq(t *testing.T, id any, role string, userID uint, body any) *http.Request {
	t.Helper()
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	return rawReviewReq(id, role, userID, raw)
}

// rawReviewReq is reviewReq for callers that need to send a raw (possibly malformed) body.
func rawReviewReq(id any, role string, userID uint, raw []byte) *http.Request {
	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/admin/events/%v/review", id), bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	return req.WithContext(middleware.WithAuthUser(req.Context(), middleware.AuthUser{
		ID:   userID,
		Role: role,
		Name: "Reviewer",
	}))
}

// existingEvent returns a model.Event that passes ValidateEditedEvent — the
// "existing" event being reviewed. Each test mutates a copy to isolate the
// field under test.
func existingEvent() model.Event {
	return model.Event{
		Model:           gorm.Model{ID: 7},
		Name:            "International Conference on Software Engineering",
		Slug:            "ICSE2026",
		Country:         "USA",
		City:            "Boston",
		Latitude:        42.36,
		Longitude:       -71.05,
		StartDate:       time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC),
		EndDate:         time.Date(2026, 5, 18, 0, 0, 0, 0, time.UTC),
		WebsiteURL:      "https://icse2026.example.org",
		Domain:          "computer_science",
		Tier:            "A",
		Status:          model.EventStatusPending,
		Year:            2026,
		CreatedByID:     3,
		LastUpdatedByID: 3,
	}
}

// --- 200 ---

func TestAdminEventHandler_Review_ApproveNoEdits_Returns200(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)
	existing := existingEvent()
	repo.EXPECT().FindByID(gomock.Any(), existing.ID).Return(existing, nil)
	repo.EXPECT().Review(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ any, updated model.Event, auditLog model.AuditLog) (model.Event, error) {
			assert.Equal(t, model.EventStatusApproved, updated.Status)
			assert.Equal(t, uint(1), updated.LastUpdatedByID)
			assert.Equal(t, model.AuditActionApproved, auditLog.Action)
			return updated, nil
		})

	h := handler.NewAdminEventHandler(repo, testLogger)
	rec := httptest.NewRecorder()

	reviewMux(h).ServeHTTP(rec, reviewReq(t, existing.ID, "admin", 1, map[string]any{"action": "approve"}))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "EVENT_REVIEWED", responseCode(t, rec))
}

func TestAdminEventHandler_Review_ApproveWithEdits_RecomputesYear(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)
	existing := existingEvent()
	repo.EXPECT().FindByID(gomock.Any(), existing.ID).Return(existing, nil)
	repo.EXPECT().Review(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ any, updated model.Event, auditLog model.AuditLog) (model.Event, error) {
			assert.Equal(t, "Fixed Conference Name", updated.Name)
			assert.Equal(t, 2027, updated.Year)
			return updated, nil
		})

	h := handler.NewAdminEventHandler(repo, testLogger)
	rec := httptest.NewRecorder()

	body := map[string]any{
		"action": "approve",
		"reason": "Fixed typo and start date",
		"event": map[string]any{
			"name":       "Fixed Conference Name",
			"start_date": "2027-01-10",
			"end_date":   "2027-01-18",
		},
	}

	reviewMux(h).ServeHTTP(rec, reviewReq(t, existing.ID, "admin", 1, body))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "EVENT_REVIEWED", responseCode(t, rec))
}

func TestAdminEventHandler_Review_RejectWithReason_Returns200(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)
	existing := existingEvent()
	existing.Status = model.EventStatusApproved
	repo.EXPECT().FindByID(gomock.Any(), existing.ID).Return(existing, nil)
	repo.EXPECT().Review(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ any, updated model.Event, auditLog model.AuditLog) (model.Event, error) {
			assert.Equal(t, model.EventStatusRejected, updated.Status)
			assert.Equal(t, model.AuditActionRejected, auditLog.Action)
			require.NotNil(t, auditLog.Reason)
			assert.Equal(t, "duplicate submission", *auditLog.Reason)
			return updated, nil
		})

	h := handler.NewAdminEventHandler(repo, testLogger)
	rec := httptest.NewRecorder()

	body := map[string]any{"action": "reject", "reason": "duplicate submission"}

	reviewMux(h).ServeHTTP(rec, reviewReq(t, existing.ID, "admin", 1, body))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "EVENT_REVIEWED", responseCode(t, rec))
}

func TestAdminEventHandler_Review_AdminReviewingOwnEvent_Returns200(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)
	existing := existingEvent()
	existing.CreatedByID = 1 // same as reviewer's ID
	repo.EXPECT().FindByID(gomock.Any(), existing.ID).Return(existing, nil)
	repo.EXPECT().Review(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ any, updated model.Event, _ model.AuditLog) (model.Event, error) {
			return updated, nil
		})

	h := handler.NewAdminEventHandler(repo, testLogger)
	rec := httptest.NewRecorder()

	reviewMux(h).ServeHTTP(rec, reviewReq(t, existing.ID, "admin", 1, map[string]any{"action": "approve"}))

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminEventHandler_Review_SlugUnchanged_NoCollisionCheck_Returns200(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)
	existing := existingEvent()
	repo.EXPECT().FindByID(gomock.Any(), existing.ID).Return(existing, nil)
	// No FindActiveBySlug expectation — slug unchanged means no collision check.
	repo.EXPECT().Review(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ any, updated model.Event, _ model.AuditLog) (model.Event, error) {
			return updated, nil
		})

	h := handler.NewAdminEventHandler(repo, testLogger)
	rec := httptest.NewRecorder()

	body := map[string]any{
		"action": "approve",
		"event":  map[string]any{"slug": existing.Slug},
	}

	reviewMux(h).ServeHTTP(rec, reviewReq(t, existing.ID, "admin", 1, body))

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminEventHandler_Review_SlugCollidesWithRejectedEvent_Returns200(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)
	existing := existingEvent()
	repo.EXPECT().FindByID(gomock.Any(), existing.ID).Return(existing, nil)
	repo.EXPECT().FindActiveBySlug(gomock.Any(), "REUSABLE2026").Return(model.Event{}, repository.ErrNotFound)
	repo.EXPECT().Review(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ any, updated model.Event, _ model.AuditLog) (model.Event, error) {
			return updated, nil
		})

	h := handler.NewAdminEventHandler(repo, testLogger)
	rec := httptest.NewRecorder()

	body := map[string]any{
		"action": "approve",
		"event":  map[string]any{"slug": "REUSABLE2026"},
	}

	reviewMux(h).ServeHTTP(rec, reviewReq(t, existing.ID, "admin", 1, body))

	require.Equal(t, http.StatusOK, rec.Code)
}

// --- 400 ---

func TestAdminEventHandler_Review_InvalidAction_Returns400(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)

	h := handler.NewAdminEventHandler(repo, testLogger)
	rec := httptest.NewRecorder()

	reviewMux(h).ServeHTTP(rec, reviewReq(t, 7, "admin", 1, map[string]any{"action": "approveee"}))

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestAdminEventHandler_Review_RejectMissingReason_Returns400(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)

	h := handler.NewAdminEventHandler(repo, testLogger)
	rec := httptest.NewRecorder()

	reviewMux(h).ServeHTTP(rec, reviewReq(t, 7, "admin", 1, map[string]any{"action": "reject"}))

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestAdminEventHandler_Review_InvalidEventLatitude_Returns400(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)
	existing := existingEvent()
	repo.EXPECT().FindByID(gomock.Any(), existing.ID).Return(existing, nil)

	h := handler.NewAdminEventHandler(repo, testLogger)
	rec := httptest.NewRecorder()

	body := map[string]any{
		"action": "approve",
		"event":  map[string]any{"latitude": 999},
	}

	reviewMux(h).ServeHTTP(rec, reviewReq(t, existing.ID, "admin", 1, body))

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestAdminEventHandler_Review_EndDateBeforeStartDate_Returns400(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)
	existing := existingEvent()
	repo.EXPECT().FindByID(gomock.Any(), existing.ID).Return(existing, nil)

	h := handler.NewAdminEventHandler(repo, testLogger)
	rec := httptest.NewRecorder()

	body := map[string]any{
		"action": "approve",
		"event":  map[string]any{"end_date": "2020-01-01"},
	}

	reviewMux(h).ServeHTTP(rec, reviewReq(t, existing.ID, "admin", 1, body))

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestAdminEventHandler_Review_MalformedJSON_Returns400(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)

	h := handler.NewAdminEventHandler(repo, testLogger)
	rec := httptest.NewRecorder()

	reviewMux(h).ServeHTTP(rec, rawReviewReq(7, "admin", 1, []byte(`{"action": "approve"`)))

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestAdminEventHandler_Review_NonNumericID_Returns400(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)

	h := handler.NewAdminEventHandler(repo, testLogger)
	rec := httptest.NewRecorder()

	reviewMux(h).ServeHTTP(rec, reviewReq(t, "abc", "admin", 1, map[string]any{"action": "approve"}))

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

// --- 403 ---

func TestAdminEventHandler_Review_ModeratorReviewingOwnEvent_Returns403(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)
	existing := existingEvent()
	existing.CreatedByID = 1 // same as reviewer's ID
	repo.EXPECT().FindByID(gomock.Any(), existing.ID).Return(existing, nil)

	h := handler.NewAdminEventHandler(repo, testLogger)
	rec := httptest.NewRecorder()

	reviewMux(h).ServeHTTP(rec, reviewReq(t, existing.ID, "moderator", 1, map[string]any{"action": "approve"}))

	require.Equal(t, http.StatusForbidden, rec.Code)
	assert.Equal(t, "CANNOT_REVIEW_OWN_EVENT", responseCode(t, rec))
}

// --- 404 ---

func TestAdminEventHandler_Review_EventNotFound_Returns404(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)
	repo.EXPECT().FindByID(gomock.Any(), uint(999999)).Return(model.Event{}, repository.ErrNotFound)

	h := handler.NewAdminEventHandler(repo, testLogger)
	rec := httptest.NewRecorder()

	reviewMux(h).ServeHTTP(rec, reviewReq(t, 999999, "admin", 1, map[string]any{"action": "approve"}))

	require.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "EVENT_NOT_FOUND", responseCode(t, rec))
}

// --- 409 ---

func TestAdminEventHandler_Review_SlugCollision_Returns409(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)
	existing := existingEvent()
	repo.EXPECT().FindByID(gomock.Any(), existing.ID).Return(existing, nil)
	repo.EXPECT().FindActiveBySlug(gomock.Any(), "TAKEN2026").
		Return(model.Event{Model: gorm.Model{ID: 99}, Status: model.EventStatusApproved}, nil)

	h := handler.NewAdminEventHandler(repo, testLogger)
	rec := httptest.NewRecorder()

	body := map[string]any{
		"action": "approve",
		"event":  map[string]any{"slug": "TAKEN2026"},
	}

	reviewMux(h).ServeHTTP(rec, reviewReq(t, existing.ID, "admin", 1, body))

	require.Equal(t, http.StatusConflict, rec.Code)
	assert.Equal(t, "SLUG_ALREADY_EXISTS", responseCode(t, rec))
}
