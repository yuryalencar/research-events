package handler_test

// Spec: specs/backend/events-deadlines-add.yaml

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/yuryalencar/research-events/internal/handler"
	"github.com/yuryalencar/research-events/internal/model"
	"github.com/yuryalencar/research-events/internal/repository"
	"github.com/yuryalencar/research-events/internal/repository/mocks"
)

// validAddDeadlinesBody returns a JSON body that passes every validation rule.
// Each test mutates a copy to isolate the field under test.
func validAddDeadlinesBody() map[string]any {
	return map[string]any{
		"submitter": map[string]any{
			"name":  "Beatriz Costa",
			"email": "beatriz@example.com",
		},
		"deadlines": []map[string]any{
			{"type": "camera_ready", "description": "Research track camera-ready", "date": "2026-10-01"},
		},
	}
}

func addDeadlinesReq(t *testing.T, id string, body any) *http.Request {
	t.Helper()
	var raw []byte
	switch v := body.(type) {
	case []byte:
		raw = v
	default:
		var err error
		raw, err = json.Marshal(body)
		require.NoError(t, err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/"+id+"/deadlines", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", id)
	return req
}

// approvedEvent returns a minimal approved model.Event with the given ID, ready
// to be returned by a mocked FindByID call.
func approvedEvent(id uint) model.Event {
	return model.Event{
		Model:  gorm.Model{ID: id},
		Status: model.EventStatusApproved,
	}
}

// --- AddDeadlines ---

func TestEventHandler_AddDeadlines_ReturnsUpdatedEventOnSingleDeadline(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)
	repo.EXPECT().FindByID(gomock.Any(), uint(1)).Return(approvedEvent(1), nil)
	repo.EXPECT().AddDeadlines(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), model.AuditActionDeadlineAdded).
		DoAndReturn(func(_ any, event model.Event, deadlines []model.Deadline, submitter model.User, _ model.AuditAction) (model.Event, error) {
			for i := range deadlines {
				deadlines[i].ID = uint(i + 1)
				deadlines[i].EventID = event.ID
			}
			event.Deadlines = deadlines
			event.LastUpdatedBy = model.User{Name: submitter.Name, Email: submitter.Email, Role: model.UserRoleContributor}
			return event, nil
		})

	h := handler.NewEventHandler(repo, testLogger)
	rec := httptest.NewRecorder()

	h.AddDeadlines(rec, addDeadlinesReq(t, "1", validAddDeadlinesBody()))

	require.Equal(t, http.StatusCreated, rec.Code)

	var resp struct {
		Data struct {
			ID            uint   `json:"id"`
			Status        string `json:"status"`
			LastUpdatedBy struct {
				Email string `json:"email"`
			} `json:"last_updated_by"`
			Deadlines []struct {
				Type        string `json:"type"`
				Description string `json:"description"`
			} `json:"deadlines"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, uint(1), resp.Data.ID)
	assert.Equal(t, "approved", resp.Data.Status)
	assert.Equal(t, "beatriz@example.com", resp.Data.LastUpdatedBy.Email)
	require.Len(t, resp.Data.Deadlines, 1)
	assert.Equal(t, "camera_ready", resp.Data.Deadlines[0].Type)
}

func TestEventHandler_AddDeadlines_ReturnsUpdatedEventOnMultipleDeadlines(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)
	repo.EXPECT().FindByID(gomock.Any(), uint(2)).Return(approvedEvent(2), nil)
	repo.EXPECT().AddDeadlines(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), model.AuditActionBatchDeadlinesAdded).
		DoAndReturn(func(_ any, event model.Event, deadlines []model.Deadline, submitter model.User, _ model.AuditAction) (model.Event, error) {
			for i := range deadlines {
				deadlines[i].ID = uint(i + 1)
				deadlines[i].EventID = event.ID
			}
			event.Deadlines = deadlines
			return event, nil
		})

	h := handler.NewEventHandler(repo, testLogger)
	body := validAddDeadlinesBody()
	body["deadlines"] = []map[string]any{
		{"type": "paper", "description": "Industry track full paper", "date": "2026-08-29"},
		{"type": "notification", "description": "Industry track notification", "date": "2026-09-12"},
	}
	rec := httptest.NewRecorder()

	h.AddDeadlines(rec, addDeadlinesReq(t, "2", body))

	require.Equal(t, http.StatusCreated, rec.Code)

	var resp struct {
		Data struct {
			Deadlines []struct {
				Type string `json:"type"`
			} `json:"deadlines"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	require.Len(t, resp.Data.Deadlines, 2)
}

func TestEventHandler_AddDeadlines_ReturnsNotFoundForNonNumericID(t *testing.T) {
	// Spec: events-deadlines-add.yaml border_case ":id is not a positive integer (e.g. "abc") → 404 EVENT_NOT_FOUND"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)

	h := handler.NewEventHandler(repo, testLogger)
	rec := httptest.NewRecorder()

	h.AddDeadlines(rec, addDeadlinesReq(t, "abc", validAddDeadlinesBody()))

	require.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "EVENT_NOT_FOUND", responseCode(t, rec))
}

func TestEventHandler_AddDeadlines_ReturnsNotFoundWhenEventDoesNotExist(t *testing.T) {
	// Spec: events-deadlines-add.yaml border_case ":id does not match any event → 404 EVENT_NOT_FOUND"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)
	repo.EXPECT().FindByID(gomock.Any(), uint(999)).Return(model.Event{}, repository.ErrNotFound)

	h := handler.NewEventHandler(repo, testLogger)
	rec := httptest.NewRecorder()

	h.AddDeadlines(rec, addDeadlinesReq(t, "999", validAddDeadlinesBody()))

	require.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "EVENT_NOT_FOUND", responseCode(t, rec))
}

func TestEventHandler_AddDeadlines_ReturnsConflictWhenEventIsPending(t *testing.T) {
	// Spec: events-deadlines-add.yaml border_case "event exists with status=pending → 409 EVENT_NOT_APPROVED"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	event := approvedEvent(3)
	event.Status = model.EventStatusPending

	repo := mocks.NewMockEventRepository(ctrl)
	repo.EXPECT().FindByID(gomock.Any(), uint(3)).Return(event, nil)

	h := handler.NewEventHandler(repo, testLogger)
	rec := httptest.NewRecorder()

	h.AddDeadlines(rec, addDeadlinesReq(t, "3", validAddDeadlinesBody()))

	require.Equal(t, http.StatusConflict, rec.Code)
	assert.Equal(t, "EVENT_NOT_APPROVED", responseCode(t, rec))
}

func TestEventHandler_AddDeadlines_ReturnsConflictWhenEventIsRejected(t *testing.T) {
	// Spec: events-deadlines-add.yaml border_case "event exists with status=rejected → 409 EVENT_NOT_APPROVED"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	event := approvedEvent(4)
	event.Status = model.EventStatusRejected

	repo := mocks.NewMockEventRepository(ctrl)
	repo.EXPECT().FindByID(gomock.Any(), uint(4)).Return(event, nil)

	h := handler.NewEventHandler(repo, testLogger)
	rec := httptest.NewRecorder()

	h.AddDeadlines(rec, addDeadlinesReq(t, "4", validAddDeadlinesBody()))

	require.Equal(t, http.StatusConflict, rec.Code)
	assert.Equal(t, "EVENT_NOT_APPROVED", responseCode(t, rec))
}

func TestEventHandler_AddDeadlines_ReturnsValidationErrorForEmptyDeadlines(t *testing.T) {
	// Spec: events-deadlines-add.yaml border_case "deadlines omitted or empty array → 400 VALIDATION_ERROR"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)

	h := handler.NewEventHandler(repo, testLogger)
	body := validAddDeadlinesBody()
	body["deadlines"] = []map[string]any{}
	rec := httptest.NewRecorder()

	h.AddDeadlines(rec, addDeadlinesReq(t, "1", body))

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestEventHandler_AddDeadlines_ReturnsValidationErrorForInvalidDeadlineType(t *testing.T) {
	// Spec: events-deadlines-add.yaml border_case "a deadlines entry with an invalid type value → 400 VALIDATION_ERROR"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)

	h := handler.NewEventHandler(repo, testLogger)
	body := validAddDeadlinesBody()
	body["deadlines"] = []map[string]any{
		{"type": "keynote", "description": "Keynote announcement", "date": "2026-08-22"},
	}
	rec := httptest.NewRecorder()

	h.AddDeadlines(rec, addDeadlinesReq(t, "1", body))

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestEventHandler_AddDeadlines_ReturnsValidationErrorForUnparsableDate(t *testing.T) {
	// Spec: events-deadlines-add.yaml border_case "a deadlines entry with an unparsable date → 400 VALIDATION_ERROR"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)

	h := handler.NewEventHandler(repo, testLogger)
	body := validAddDeadlinesBody()
	body["deadlines"] = []map[string]any{
		{"type": "paper", "description": "Research track full paper", "date": "not-a-date"},
	}
	rec := httptest.NewRecorder()

	h.AddDeadlines(rec, addDeadlinesReq(t, "1", body))

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestEventHandler_AddDeadlines_ReturnsValidationErrorForMissingSubmitterName(t *testing.T) {
	// Spec: events-deadlines-add.yaml border_case "submitter.name missing → 400 VALIDATION_ERROR"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)

	h := handler.NewEventHandler(repo, testLogger)
	body := validAddDeadlinesBody()
	body["submitter"] = map[string]any{"name": "", "email": "beatriz@example.com"}
	rec := httptest.NewRecorder()

	h.AddDeadlines(rec, addDeadlinesReq(t, "1", body))

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestEventHandler_AddDeadlines_ReturnsValidationErrorForInvalidSubmitterEmail(t *testing.T) {
	// Spec: events-deadlines-add.yaml border_case "submitter.email missing or invalid format → 400 VALIDATION_ERROR"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)

	h := handler.NewEventHandler(repo, testLogger)
	body := validAddDeadlinesBody()
	body["submitter"] = map[string]any{"name": "Beatriz Costa", "email": "not-an-email"}
	rec := httptest.NewRecorder()

	h.AddDeadlines(rec, addDeadlinesReq(t, "1", body))

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestEventHandler_AddDeadlines_ReturnsValidationErrorForMalformedJSON(t *testing.T) {
	// Spec: events-deadlines-add.yaml border_case "malformed JSON body → 400 VALIDATION_ERROR"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)

	h := handler.NewEventHandler(repo, testLogger)
	rec := httptest.NewRecorder()

	h.AddDeadlines(rec, addDeadlinesReq(t, "1", []byte("{not json")))

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}
