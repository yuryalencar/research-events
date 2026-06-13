package service_test

// Spec: specs/backend/events-list.yaml
//
// ValidateListEventsQuery is pure: given the same RawListEventsQuery and
// currentYear it always returns the same result, with no I/O or hidden state.

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuryalencar/research-events/internal/model"
	"github.com/yuryalencar/research-events/internal/service"
)

const currentYear = 2026

// --- Defaults ---

func TestValidateListEventsQuery_NoParams_ReturnsDefaults(t *testing.T) {
	// Spec: events-list.yaml border_case "No filters at all → status=approved,
	// year=current year, page=1, page_size=20"
	got, err := service.ValidateListEventsQuery(service.RawListEventsQuery{}, currentYear)

	require.NoError(t, err)
	assert.Equal(t, currentYear, got.Year)
	assert.Equal(t, model.EventStatusApproved, got.Status)
	assert.Nil(t, got.Domain)
	assert.Nil(t, got.Country)
	assert.Nil(t, got.Tier)
	assert.Nil(t, got.FirstDeadlineMonth)
	assert.Nil(t, got.BBox)
	assert.Equal(t, 1, got.Page)
	assert.Equal(t, 20, got.PageSize)
	assert.False(t, got.PaginationOff)
}

// --- Year ---

func TestValidateListEventsQuery_YearOverride(t *testing.T) {
	got, err := service.ValidateListEventsQuery(service.RawListEventsQuery{Year: "2025"}, currentYear)

	require.NoError(t, err)
	assert.Equal(t, 2025, got.Year)
}

func TestValidateListEventsQuery_InvalidYear_ReturnsError(t *testing.T) {
	// Spec: events-list.yaml border_case "?year=abc (non-numeric) → 400 VALIDATION_ERROR"
	_, err := service.ValidateListEventsQuery(service.RawListEventsQuery{Year: "abc"}, currentYear)

	require.Error(t, err)
}

func TestValidateListEventsQuery_StatusOmitted_YearStillDefaultsToCurrent(t *testing.T) {
	// Spec: events-list.yaml border_case "?status=pending with no ?year → year still
	// defaults to current year"
	got, err := service.ValidateListEventsQuery(service.RawListEventsQuery{Status: "pending"}, currentYear)

	require.NoError(t, err)
	assert.Equal(t, currentYear, got.Year)
	assert.Equal(t, model.EventStatusPending, got.Status)
}

// --- Status ---

func TestValidateListEventsQuery_StatusOverride(t *testing.T) {
	for _, status := range []string{"pending", "approved", "rejected"} {
		got, err := service.ValidateListEventsQuery(service.RawListEventsQuery{Status: status}, currentYear)

		require.NoError(t, err)
		assert.Equal(t, model.EventStatus(status), got.Status)
	}
}

func TestValidateListEventsQuery_InvalidStatus_ReturnsError(t *testing.T) {
	// Spec: events-list.yaml border_case "?status=foo (not in enum) → 400 VALIDATION_ERROR"
	_, err := service.ValidateListEventsQuery(service.RawListEventsQuery{Status: "foo"}, currentYear)

	require.Error(t, err)
}

// --- Domain ---

func TestValidateListEventsQuery_DomainOverride(t *testing.T) {
	got, err := service.ValidateListEventsQuery(service.RawListEventsQuery{Domain: "computer_science"}, currentYear)

	require.NoError(t, err)
	require.NotNil(t, got.Domain)
	assert.Equal(t, "computer_science", *got.Domain)
}

func TestValidateListEventsQuery_InvalidDomain_ReturnsError(t *testing.T) {
	// Spec: events-list.yaml border_case "?domain=foo (not in enum) → 400 VALIDATION_ERROR"
	_, err := service.ValidateListEventsQuery(service.RawListEventsQuery{Domain: "medicine"}, currentYear)

	require.Error(t, err)
}

// --- Country ---

func TestValidateListEventsQuery_CountryOverride(t *testing.T) {
	got, err := service.ValidateListEventsQuery(service.RawListEventsQuery{Country: "Brazil"}, currentYear)

	require.NoError(t, err)
	require.NotNil(t, got.Country)
	assert.Equal(t, "Brazil", *got.Country)
}

// --- Tier ---

func TestValidateListEventsQuery_TierOverride(t *testing.T) {
	for _, tier := range []string{"A*", "A", "B", "C", "unranked"} {
		got, err := service.ValidateListEventsQuery(service.RawListEventsQuery{Tier: tier}, currentYear)

		require.NoError(t, err)
		require.NotNil(t, got.Tier)
		assert.Equal(t, tier, *got.Tier)
	}
}

func TestValidateListEventsQuery_InvalidTier_ReturnsError(t *testing.T) {
	// Spec: events-list.yaml border_case "?tier=foo (not in enum) → 400 VALIDATION_ERROR"
	_, err := service.ValidateListEventsQuery(service.RawListEventsQuery{Tier: "S"}, currentYear)

	require.Error(t, err)
}

// --- FirstDeadlineMonth ---

func TestValidateListEventsQuery_FirstDeadlineMonthValid(t *testing.T) {
	got, err := service.ValidateListEventsQuery(service.RawListEventsQuery{FirstDeadlineMonth: "6"}, currentYear)

	require.NoError(t, err)
	require.NotNil(t, got.FirstDeadlineMonth)
	assert.Equal(t, 6, *got.FirstDeadlineMonth)
}

func TestValidateListEventsQuery_FirstDeadlineMonthOutOfRange_ReturnsError(t *testing.T) {
	// Spec: events-list.yaml border_case "?first_deadline_month=0 or 13 (out of 1-12 range) → 400 VALIDATION_ERROR"
	for _, month := range []string{"0", "13"} {
		_, err := service.ValidateListEventsQuery(service.RawListEventsQuery{FirstDeadlineMonth: month}, currentYear)

		require.Error(t, err, "month %q should be out of range", month)
	}
}

func TestValidateListEventsQuery_FirstDeadlineMonthNotNumeric_ReturnsError(t *testing.T) {
	_, err := service.ValidateListEventsQuery(service.RawListEventsQuery{FirstDeadlineMonth: "june"}, currentYear)

	require.Error(t, err)
}

// --- BBox ---

func TestValidateListEventsQuery_BBoxValid(t *testing.T) {
	got, err := service.ValidateListEventsQuery(service.RawListEventsQuery{BBox: "-10,-5,10,5"}, currentYear)

	require.NoError(t, err)
	require.NotNil(t, got.BBox)
	assert.Equal(t, service.BBoxInput{MinLng: -10, MinLat: -5, MaxLng: 10, MaxLat: 5}, *got.BBox)
}

func TestValidateListEventsQuery_BBoxWrongCount_ReturnsError(t *testing.T) {
	// Spec: events-list.yaml border_case "?bbox=1,2,3 (wrong number of values) → 400 VALIDATION_ERROR"
	_, err := service.ValidateListEventsQuery(service.RawListEventsQuery{BBox: "1,2,3"}, currentYear)

	require.Error(t, err)
}

func TestValidateListEventsQuery_BBoxInvertedRange_ReturnsError(t *testing.T) {
	// Spec: events-list.yaml border_case "?bbox with minLng >= maxLng or minLat >= maxLat → 400 VALIDATION_ERROR"
	_, err := service.ValidateListEventsQuery(service.RawListEventsQuery{BBox: "10,-5,-10,5"}, currentYear)

	require.Error(t, err)
}

func TestValidateListEventsQuery_BBoxOutOfRange_ReturnsError(t *testing.T) {
	// Spec: events-list.yaml border_case "?bbox with a value out of range (e.g. lat=95) → 400 VALIDATION_ERROR"
	_, err := service.ValidateListEventsQuery(service.RawListEventsQuery{BBox: "-10,-5,10,95"}, currentYear)

	require.Error(t, err)
}

func TestValidateListEventsQuery_BBoxNonNumeric_ReturnsError(t *testing.T) {
	_, err := service.ValidateListEventsQuery(service.RawListEventsQuery{BBox: "a,b,c,d"}, currentYear)

	require.Error(t, err)
}

// --- Pagination ---

func TestValidateListEventsQuery_PageOverride(t *testing.T) {
	got, err := service.ValidateListEventsQuery(service.RawListEventsQuery{Page: "2", PageSize: "10"}, currentYear)

	require.NoError(t, err)
	assert.Equal(t, 2, got.Page)
	assert.Equal(t, 10, got.PageSize)
}

func TestValidateListEventsQuery_PageZeroOrNegative_ReturnsError(t *testing.T) {
	// Spec: events-list.yaml border_case "?page=0 or negative → 400 VALIDATION_ERROR"
	for _, page := range []string{"0", "-1"} {
		_, err := service.ValidateListEventsQuery(service.RawListEventsQuery{Page: page}, currentYear)

		require.Error(t, err, "page %q should be invalid", page)
	}
}

func TestValidateListEventsQuery_PageSizeInvalid_ReturnsError(t *testing.T) {
	// Spec: events-list.yaml border_case "?page_size=0, negative, or > 100 → 400 VALIDATION_ERROR"
	for _, size := range []string{"0", "-1", "101"} {
		_, err := service.ValidateListEventsQuery(service.RawListEventsQuery{PageSize: size}, currentYear)

		require.Error(t, err, "page_size %q should be invalid", size)
	}
}

func TestValidateListEventsQuery_PaginationOff_SetsPaginationOffAndIgnoresPageParams(t *testing.T) {
	// Spec: events-list.yaml border_case "?pagination=off with ?page/?page_size also
	// set → pagination=off wins; page/page_size ignored"
	got, err := service.ValidateListEventsQuery(service.RawListEventsQuery{
		Pagination: "off", Page: "5", PageSize: "999",
	}, currentYear)

	require.NoError(t, err)
	assert.True(t, got.PaginationOff)
	assert.Equal(t, 1, got.Page)
}

func TestValidateListEventsQuery_InvalidPaginationValue_ReturnsError(t *testing.T) {
	// Spec: events-list.yaml border_case "?pagination=maybe (not 'on'/'off') → 400 VALIDATION_ERROR"
	_, err := service.ValidateListEventsQuery(service.RawListEventsQuery{Pagination: "maybe"}, currentYear)

	require.Error(t, err)
}

// --- Purity ---

func TestValidateListEventsQuery_SameInputTwice_ReturnsIdenticalResult(t *testing.T) {
	// FP: pure function — same input always produces the same output, no hidden state.
	raw := service.RawListEventsQuery{Year: "2025", Status: "pending", Tier: "A"}

	got1, err1 := service.ValidateListEventsQuery(raw, currentYear)
	got2, err2 := service.ValidateListEventsQuery(raw, currentYear)

	assert.Equal(t, got1, got2)
	assert.Equal(t, err1, err2)
}
