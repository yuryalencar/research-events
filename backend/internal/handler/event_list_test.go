package handler_test

// Spec: specs/backend/events-list.yaml

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/yuryalencar/research-events/internal/handler"
	"github.com/yuryalencar/research-events/internal/model"
	"github.com/yuryalencar/research-events/internal/repository"
	"github.com/yuryalencar/research-events/internal/repository/mocks"
)

// listResponse mirrors the GET /api/v1/events response envelope:
// { "code": "...", "data": [...], "meta": { "page": N, "total": N } }.
type listResponse struct {
	Code string                  `json:"code"`
	Data []eventListItemResponse `json:"data"`
	Meta struct {
		Page  int   `json:"page"`
		Total int64 `json:"total"`
	} `json:"meta"`
}

// eventListItemResponse mirrors one entry of the "data" array per
// events-list.yaml responses.200.body.data.
type eventListItemResponse struct {
	ID         uint    `json:"id"`
	Name       string  `json:"name"`
	Slug       string  `json:"slug"`
	Country    string  `json:"country"`
	City       string  `json:"city"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	StartDate  string  `json:"start_date"`
	EndDate    string  `json:"end_date"`
	WebsiteURL string  `json:"website_url"`
	Domain     string  `json:"domain"`
	Status     string  `json:"status"`
	Tier       string  `json:"tier"`
	Year       int     `json:"year"`
	CreatedBy  struct {
		ID    uint   `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"created_by"`
	LastUpdatedBy struct {
		ID    uint   `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"last_updated_by"`
	Deadlines []struct {
		ID          uint   `json:"id"`
		Type        string `json:"type"`
		Description string `json:"description"`
		Date        string `json:"date"`
		IsOptional  bool   `json:"is_optional"`
		IsActive    bool   `json:"is_active"`
	} `json:"deadlines"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func listReq(target string) *http.Request {
	return httptest.NewRequest(http.MethodGet, target, nil)
}

func decodeListResponse(t *testing.T, rec *httptest.ResponseRecorder) listResponse {
	t.Helper()
	var body listResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	return body
}

// --- List ---

func TestEventHandler_List_NoParams_UsesDefaultFilters(t *testing.T) {
	// Spec: events-list.yaml border_case "No filters at all → status=approved,
	// year=current year, page=1, page_size=20."
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	currentYear := time.Now().Year()

	repo := mocks.NewMockEventRepository(ctrl)
	repo.EXPECT().ListEvents(gomock.Any(), repository.ListEventsFilter{
		Year: currentYear, Status: model.EventStatusApproved, Page: 1, PageSize: 20,
	}).Return([]model.Event{}, int64(0), nil)

	h := handler.NewEventHandler(repo, testLogger)
	rec := httptest.NewRecorder()

	h.List(rec, listReq("/api/v1/events"))

	require.Equal(t, http.StatusOK, rec.Code)
	body := decodeListResponse(t, rec)
	assert.Empty(t, body.Data)
	assert.Equal(t, 1, body.Meta.Page)
	assert.Equal(t, int64(0), body.Meta.Total)
}

func TestEventHandler_List_InvalidYear_ReturnsValidationErrorWithoutCallingRepository(t *testing.T) {
	// Spec: events-list.yaml border_case "?year=abc (non-numeric) → 400 VALIDATION_ERROR."
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockEventRepository(ctrl)
	// No ListEvents expectation — validation must fail before any repository call.

	h := handler.NewEventHandler(repo, testLogger)
	rec := httptest.NewRecorder()

	h.List(rec, listReq("/api/v1/events?year=abc"))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "VALIDATION_ERROR", responseCode(t, rec))
}

func TestEventHandler_List_QueryParams_AreParsedAndPassedToFilter(t *testing.T) {
	// Spec: events-list.yaml rules "All filters combine with AND (year AND domain
	// AND country AND status AND tier AND first_deadline_month AND bbox)."
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	domain := "computer_science"
	country := "Brazil"
	tier := "A"
	month := 6

	repo := mocks.NewMockEventRepository(ctrl)
	repo.EXPECT().ListEvents(gomock.Any(), repository.ListEventsFilter{
		Year: 2025, Status: model.EventStatusPending,
		Domain: &domain, Country: &country, Tier: &tier,
		FirstDeadlineMonth: &month,
		BBox:               &repository.BBoxFilter{MinLng: -10, MinLat: -5, MaxLng: 10, MaxLat: 5},
		Page:               2, PageSize: 10,
	}).Return([]model.Event{}, int64(0), nil)

	h := handler.NewEventHandler(repo, testLogger)
	rec := httptest.NewRecorder()

	h.List(rec, listReq("/api/v1/events?year=2025&status=pending&domain=computer_science"+
		"&country=Brazil&tier=A&first_deadline_month=6&bbox=-10,-5,10,5&page=2&page_size=10"))

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestEventHandler_List_PaginationOff_SetsMetaPageToOne(t *testing.T) {
	// Spec: events-list.yaml rule "pagination=off returns every matching row in
	// one response; meta.page=1 and meta.total=len(data) in that case."
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	currentYear := time.Now().Year()

	repo := mocks.NewMockEventRepository(ctrl)
	repo.EXPECT().ListEvents(gomock.Any(), repository.ListEventsFilter{
		Year: currentYear, Status: model.EventStatusApproved, Page: 1, PaginationOff: true,
	}).Return([]model.Event{{}, {}}, int64(2), nil)

	h := handler.NewEventHandler(repo, testLogger)
	rec := httptest.NewRecorder()

	h.List(rec, listReq("/api/v1/events?pagination=off"))

	require.Equal(t, http.StatusOK, rec.Code)
	body := decodeListResponse(t, rec)
	assert.Len(t, body.Data, 2)
	assert.Equal(t, 1, body.Meta.Page)
	assert.Equal(t, int64(2), body.Meta.Total)
}

func TestEventHandler_List_MapsEventFieldsIncludingTierYearAndAttribution(t *testing.T) {
	// Spec: events-list.yaml definition_of_done "Each returned event includes
	// created_by and last_updated_by." / "Each returned event includes its tier
	// (never null — 'unranked' by default)."
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	currentYear := time.Now().Year()
	startDate := time.Date(currentYear, 9, 21, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(currentYear, 9, 25, 0, 0, 0, 0, time.UTC)

	event := model.Event{
		Name: "MODELS", Slug: "MODELS2026", Country: "Brazil", City: "Recife",
		Latitude: -8.04, Longitude: -34.87,
		StartDate: startDate, EndDate: endDate,
		WebsiteURL: "https://models2026.example.org", Domain: "computer_science",
		Status: model.EventStatusApproved, Year: currentYear, Tier: "A",
		CreatedBy:     model.User{Name: "Ana Silva", Email: "ana@example.com"},
		LastUpdatedBy: model.User{Name: "Bob Souza", Email: "bob@example.com"},
		Deadlines: []model.Deadline{
			{Type: model.DeadlineTypePaper, Description: "Paper", Date: startDate, IsOptional: false, IsActive: true},
		},
	}
	event.ID = 7

	repo := mocks.NewMockEventRepository(ctrl)
	repo.EXPECT().ListEvents(gomock.Any(), gomock.Any()).Return([]model.Event{event}, int64(1), nil)

	h := handler.NewEventHandler(repo, testLogger)
	rec := httptest.NewRecorder()

	h.List(rec, listReq("/api/v1/events"))

	require.Equal(t, http.StatusOK, rec.Code)
	body := decodeListResponse(t, rec)
	require.Len(t, body.Data, 1)

	got := body.Data[0]
	assert.Equal(t, uint(7), got.ID)
	assert.Equal(t, "MODELS2026", got.Slug)
	assert.Equal(t, "A", got.Tier)
	assert.Equal(t, currentYear, got.Year)
	assert.Equal(t, "ana@example.com", got.CreatedBy.Email)
	assert.Equal(t, "bob@example.com", got.LastUpdatedBy.Email)
	require.Len(t, got.Deadlines, 1)
	assert.Equal(t, "Paper", got.Deadlines[0].Description)
	assert.True(t, got.Deadlines[0].IsActive)
}

func TestEventHandler_List_RepositoryError_ReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	currentYear := time.Now().Year()

	repo := mocks.NewMockEventRepository(ctrl)
	repo.EXPECT().ListEvents(gomock.Any(), repository.ListEventsFilter{
		Year: currentYear, Status: model.EventStatusApproved, Page: 1, PageSize: 20,
	}).Return(nil, int64(0), assert.AnError)

	h := handler.NewEventHandler(repo, testLogger)
	rec := httptest.NewRecorder()

	h.List(rec, listReq("/api/v1/events"))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, "INTERNAL_ERROR", responseCode(t, rec))
}
