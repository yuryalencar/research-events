package handler_test

// Spec: specs/backend/events-submit.yaml

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/yuryalencar/research-events/internal/handler"
	"github.com/yuryalencar/research-events/internal/model"
	"github.com/yuryalencar/research-events/internal/repository"
	"github.com/yuryalencar/research-events/internal/repository/mocks"
)

// validSubmitBody returns a JSON body that passes every validation rule.
// Each test mutates a copy to isolate the field under test.
func validSubmitBody() map[string]any {
	return map[string]any{
		"name":        "International Conference on Model-Driven Engineering",
		"slug":        "MODELS2026",
		"country":     "Brazil",
		"city":        "Recife",
		"latitude":    -8.0476,
		"longitude":   -34.8770,
		"start_date":  "2026-09-21",
		"end_date":    "2026-09-25",
		"website_url": "https://models2026.example.org",
		"domain":      "computer_science",
		"submitter": map[string]any{
			"name":  "Ana Silva",
			"email": "ana@example.com",
		},
	}
}

func submitReq(t *testing.T, body any) *http.Request {
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
	req := httptest.NewRequest(http.MethodPost, "/api/v1/events/submit", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// --- Submit ---

func TestEventHandler_Submit_ReturnsCreatedEventWithoutDeadlines(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)
	repo.EXPECT().FindActiveBySlug(gomock.Any(), "MODELS2026").Return(model.Event{}, repository.ErrNotFound)
	repo.EXPECT().Submit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ any, event model.Event, deadlines []model.Deadline, submitter model.User) (model.Event, error) {
			event.ID = 1
			event.CreatedBy = model.User{Name: submitter.Name, Email: submitter.Email, Role: model.UserRoleContributor}
			event.LastUpdatedBy = event.CreatedBy
			event.Deadlines = deadlines
			return event, nil
		})

	h := handler.NewEventHandler(repo, testLogger)
	rec := httptest.NewRecorder()

	h.Submit(rec, submitReq(t, validSubmitBody()))

	require.Equal(t, http.StatusCreated, rec.Code)

	var body struct {
		Data struct {
			ID        uint   `json:"id"`
			Slug      string `json:"slug"`
			Status    string `json:"status"`
			CreatedBy struct {
				Email string `json:"email"`
			} `json:"created_by"`
			Deadlines []any `json:"deadlines"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.Equal(t, "MODELS2026", body.Data.Slug)
	assert.Equal(t, "pending", body.Data.Status)
	assert.Equal(t, "ana@example.com", body.Data.CreatedBy.Email)
	assert.Empty(t, body.Data.Deadlines)
}

func TestEventHandler_Submit_ReturnsCreatedEventWithDeadlines(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)
	repo.EXPECT().FindActiveBySlug(gomock.Any(), "ICSE2027").Return(model.Event{}, repository.ErrNotFound)
	repo.EXPECT().Submit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ any, event model.Event, deadlines []model.Deadline, submitter model.User) (model.Event, error) {
			event.ID = 2
			for i := range deadlines {
				deadlines[i].ID = uint(i + 1)
				deadlines[i].EventID = event.ID
			}
			event.Deadlines = deadlines
			event.CreatedBy = model.User{Name: submitter.Name, Email: submitter.Email, Role: model.UserRoleContributor}
			return event, nil
		})

	h := handler.NewEventHandler(repo, testLogger)
	body := validSubmitBody()
	body["slug"] = "ICSE2027"
	body["deadlines"] = []map[string]any{
		{"type": "abstract", "description": "Research track abstract", "date": "2026-08-15", "is_optional": true},
		{"type": "paper", "description": "Research track full paper", "date": "2026-08-22"},
	}
	rec := httptest.NewRecorder()

	h.Submit(rec, submitReq(t, body))

	require.Equal(t, http.StatusCreated, rec.Code)

	var resp struct {
		Data struct {
			Deadlines []struct {
				Type        string `json:"type"`
				Description string `json:"description"`
				IsOptional  bool   `json:"is_optional"`
			} `json:"deadlines"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	require.Len(t, resp.Data.Deadlines, 2)
	assert.Equal(t, "abstract", resp.Data.Deadlines[0].Type)
	assert.True(t, resp.Data.Deadlines[0].IsOptional)
	assert.Equal(t, "paper", resp.Data.Deadlines[1].Type)
	assert.False(t, resp.Data.Deadlines[1].IsOptional)
}

func TestEventHandler_Submit_ReturnsDeadlineTimeAndTimezoneInResponse(t *testing.T) {
	// Spec: deadlines-add-time-timezone.yaml — "POST /api/v1/events/submit
	// accepts and returns time/timezone per deadline"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)
	repo.EXPECT().FindActiveBySlug(gomock.Any(), "MODELS2026TZ").Return(model.Event{}, repository.ErrNotFound)
	repo.EXPECT().Submit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ any, event model.Event, deadlines []model.Deadline, submitter model.User) (model.Event, error) {
			event.ID = 1
			for i := range deadlines {
				deadlines[i].ID = uint(i + 1)
				deadlines[i].EventID = event.ID
			}
			event.Deadlines = deadlines
			event.CreatedBy = model.User{Name: submitter.Name, Email: submitter.Email, Role: model.UserRoleContributor}
			return event, nil
		})

	h := handler.NewEventHandler(repo, testLogger)
	body := validSubmitBody()
	body["slug"] = "MODELS2026TZ"
	body["deadlines"] = []map[string]any{
		{"type": "paper", "description": "Research track full paper", "date": "2026-08-22", "time": "23:59", "timezone": "AoE"},
		{"type": "camera_ready", "description": "Research track camera-ready", "date": "2026-10-01"},
	}
	rec := httptest.NewRecorder()

	h.Submit(rec, submitReq(t, body))

	require.Equal(t, http.StatusCreated, rec.Code)

	var resp struct {
		Data struct {
			Deadlines []struct {
				Time     *string `json:"time"`
				Timezone *string `json:"timezone"`
			} `json:"deadlines"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	require.Len(t, resp.Data.Deadlines, 2)
	require.NotNil(t, resp.Data.Deadlines[0].Time)
	require.NotNil(t, resp.Data.Deadlines[0].Timezone)
	assert.Equal(t, "23:59", *resp.Data.Deadlines[0].Time)
	assert.Equal(t, "AoE", *resp.Data.Deadlines[0].Timezone)
	assert.Nil(t, resp.Data.Deadlines[1].Time)
	assert.Nil(t, resp.Data.Deadlines[1].Timezone)
}

func TestEventHandler_Submit_InvalidDeadlineTime_ReturnsValidationError(t *testing.T) {
	// Spec: deadlines-add-time-timezone.yaml border_case `time = "24:00" → 400 VALIDATION_ERROR`
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockEventRepository(ctrl)

	h := handler.NewEventHandler(repo, testLogger)
	body := validSubmitBody()
	body["deadlines"] = []map[string]any{
		{"type": "paper", "description": "Research track full paper", "date": "2026-08-22", "time": "24:00"},
	}
	rec := httptest.NewRecorder()

	h.Submit(rec, submitReq(t, body))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestEventHandler_Submit_ReturnsCreatedForPastDates(t *testing.T) {
	// Spec: events-submit.yaml border_case "start_date/end_date in the past → allowed, 201"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)
	repo.EXPECT().FindActiveBySlug(gomock.Any(), "PASTCONF2020").Return(model.Event{}, repository.ErrNotFound)
	repo.EXPECT().Submit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ any, event model.Event, deadlines []model.Deadline, submitter model.User) (model.Event, error) {
			event.ID = 3
			event.CreatedBy = model.User{Name: submitter.Name, Email: submitter.Email}
			return event, nil
		})

	h := handler.NewEventHandler(repo, testLogger)
	body := validSubmitBody()
	body["slug"] = "PASTCONF2020"
	body["start_date"] = "2020-03-01"
	body["end_date"] = "2020-03-05"
	rec := httptest.NewRecorder()

	h.Submit(rec, submitReq(t, body))

	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestEventHandler_Submit_MalformedJSON_ReturnsValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockEventRepository(ctrl)

	h := handler.NewEventHandler(repo, testLogger)
	rec := httptest.NewRecorder()

	h.Submit(rec, submitReq(t, []byte(`{"name": "Broken JSON"`)))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestEventHandler_Submit_MissingRequiredField_ReturnsValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockEventRepository(ctrl)

	h := handler.NewEventHandler(repo, testLogger)
	body := validSubmitBody()
	delete(body, "name")
	rec := httptest.NewRecorder()

	h.Submit(rec, submitReq(t, body))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestEventHandler_Submit_InvalidEmail_ReturnsValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockEventRepository(ctrl)

	h := handler.NewEventHandler(repo, testLogger)
	body := validSubmitBody()
	body["submitter"] = map[string]any{"name": "Ana Silva", "email": "not-an-email"}
	rec := httptest.NewRecorder()

	h.Submit(rec, submitReq(t, body))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestEventHandler_Submit_InvalidWebsiteURL_ReturnsValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockEventRepository(ctrl)

	h := handler.NewEventHandler(repo, testLogger)
	body := validSubmitBody()
	body["website_url"] = "models2026.example.org"
	rec := httptest.NewRecorder()

	h.Submit(rec, submitReq(t, body))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestEventHandler_Submit_WithTier_PersistsAndReturnsTier(t *testing.T) {
	// Spec: events-submit.yaml DoD "Submission with a valid tier (e.g. 'A*') persists and returns that tier"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)
	repo.EXPECT().FindActiveBySlug(gomock.Any(), "MODELS2026").Return(model.Event{}, repository.ErrNotFound)
	repo.EXPECT().Submit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ any, event model.Event, deadlines []model.Deadline, submitter model.User) (model.Event, error) {
			event.ID = 1
			event.CreatedBy = model.User{Name: submitter.Name, Email: submitter.Email, Role: model.UserRoleContributor}
			event.LastUpdatedBy = event.CreatedBy
			event.Deadlines = deadlines
			return event, nil
		})

	h := handler.NewEventHandler(repo, testLogger)
	body := validSubmitBody()
	body["tier"] = "A*"
	rec := httptest.NewRecorder()

	h.Submit(rec, submitReq(t, body))

	require.Equal(t, http.StatusCreated, rec.Code)

	var resp struct {
		Data struct {
			Tier string `json:"tier"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, "A*", resp.Data.Tier)
}

func TestEventHandler_Submit_WithoutTier_DefaultsToUnranked(t *testing.T) {
	// Spec: events-submit.yaml DoD "Submission without tier returns tier='unranked' in the response"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)
	repo.EXPECT().FindActiveBySlug(gomock.Any(), "MODELS2026").Return(model.Event{}, repository.ErrNotFound)
	repo.EXPECT().Submit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ any, event model.Event, deadlines []model.Deadline, submitter model.User) (model.Event, error) {
			event.ID = 1
			event.CreatedBy = model.User{Name: submitter.Name, Email: submitter.Email, Role: model.UserRoleContributor}
			event.LastUpdatedBy = event.CreatedBy
			event.Deadlines = deadlines
			return event, nil
		})

	h := handler.NewEventHandler(repo, testLogger)
	rec := httptest.NewRecorder()

	h.Submit(rec, submitReq(t, validSubmitBody()))

	require.Equal(t, http.StatusCreated, rec.Code)

	var resp struct {
		Data struct {
			Tier string `json:"tier"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, "unranked", resp.Data.Tier)
}

func TestEventHandler_Submit_InvalidTier_ReturnsValidationError(t *testing.T) {
	// Spec: events-submit.yaml border_case "tier not in the allowed enum (e.g. 'S') → 400 VALIDATION_ERROR"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockEventRepository(ctrl)

	h := handler.NewEventHandler(repo, testLogger)
	body := validSubmitBody()
	body["tier"] = "S"
	rec := httptest.NewRecorder()

	h.Submit(rec, submitReq(t, body))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestEventHandler_Submit_InvalidDomain_ReturnsValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockEventRepository(ctrl)

	h := handler.NewEventHandler(repo, testLogger)
	body := validSubmitBody()
	body["domain"] = "medicine"
	rec := httptest.NewRecorder()

	h.Submit(rec, submitReq(t, body))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestEventHandler_Submit_LatitudeOutOfRange_ReturnsValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockEventRepository(ctrl)

	h := handler.NewEventHandler(repo, testLogger)
	body := validSubmitBody()
	body["latitude"] = 200.0
	rec := httptest.NewRecorder()

	h.Submit(rec, submitReq(t, body))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestEventHandler_Submit_EndDateBeforeStartDate_ReturnsValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockEventRepository(ctrl)

	h := handler.NewEventHandler(repo, testLogger)
	body := validSubmitBody()
	body["start_date"] = "2026-09-25"
	body["end_date"] = "2026-09-21"
	rec := httptest.NewRecorder()

	h.Submit(rec, submitReq(t, body))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestEventHandler_Submit_InvalidDeadlineType_ReturnsValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockEventRepository(ctrl)

	h := handler.NewEventHandler(repo, testLogger)
	body := validSubmitBody()
	body["deadlines"] = []map[string]any{
		{"type": "keynote", "description": "Keynote announcement", "date": "2026-08-22"},
	}
	rec := httptest.NewRecorder()

	h.Submit(rec, submitReq(t, body))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestEventHandler_Submit_DuplicateSlugPending_ReturnsAlreadySubmitted(t *testing.T) {
	// Spec: events-submit.yaml responses 409 EVENT_ALREADY_SUBMITTED
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)
	repo.EXPECT().FindActiveBySlug(gomock.Any(), "MODELS2026").
		Return(model.Event{Status: model.EventStatusPending}, nil)

	h := handler.NewEventHandler(repo, testLogger)
	rec := httptest.NewRecorder()

	h.Submit(rec, submitReq(t, validSubmitBody()))

	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Equal(t, "EVENT_ALREADY_SUBMITTED", responseCode(t, rec))
}

func TestEventHandler_Submit_DuplicateSlugApproved_ReturnsAlreadySubmitted(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)
	repo.EXPECT().FindActiveBySlug(gomock.Any(), "MODELS2026").
		Return(model.Event{Status: model.EventStatusApproved}, nil)

	h := handler.NewEventHandler(repo, testLogger)
	rec := httptest.NewRecorder()

	h.Submit(rec, submitReq(t, validSubmitBody()))

	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Equal(t, "EVENT_ALREADY_SUBMITTED", responseCode(t, rec))
}

func TestEventHandler_Submit_RejectedSlugReusable_ReturnsCreated(t *testing.T) {
	// Spec: events-submit.yaml border_case "slug previously used only by a rejected event → allowed, 201 (slug reusable)"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)
	repo.EXPECT().FindActiveBySlug(gomock.Any(), "MODELS2026").Return(model.Event{}, repository.ErrNotFound)
	repo.EXPECT().Submit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ any, event model.Event, deadlines []model.Deadline, submitter model.User) (model.Event, error) {
			event.ID = 4
			event.CreatedBy = model.User{Name: submitter.Name, Email: submitter.Email}
			return event, nil
		})

	h := handler.NewEventHandler(repo, testLogger)
	rec := httptest.NewRecorder()

	h.Submit(rec, submitReq(t, validSubmitBody()))

	assert.Equal(t, http.StatusCreated, rec.Code)
}
