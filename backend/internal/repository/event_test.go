package repository_test

// Spec: specs/backend/events-submit.yaml

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/yuryalencar/research-events/internal/model"
	"github.com/yuryalencar/research-events/internal/repository"
)

// --- FindActiveBySlug ---

func TestEventRepository_FindActiveBySlug_ReturnsEventWhenPending(t *testing.T) {
	// Spec: events-submit.yaml rule "slug ... must be unique among events with status=pending or status=approved"
	tx, rollback := beginTx(t)
	defer rollback()

	user := model.User{Name: "Ana Silva", Email: "ana@example.com", Role: model.UserRoleContributor}
	require.NoError(t, tx.Create(&user).Error)

	event := model.Event{
		Name: "MODELS", Slug: "MODELS2026", Country: "Brazil", City: "Recife",
		Latitude: -8.04, Longitude: -34.87,
		StartDate:  time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC),
		EndDate:    time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC),
		WebsiteURL: "https://models2026.example.org", Domain: "computer_science",
		Status: model.EventStatusPending, Year: 2026,
		CreatedByID: user.ID, LastUpdatedByID: user.ID,
	}
	require.NoError(t, tx.Create(&event).Error)

	repo := repository.NewEventRepository(tx)

	got, err := repo.FindActiveBySlug(context.Background(), "MODELS2026")

	require.NoError(t, err)
	assert.Equal(t, event.ID, got.ID)
}

func TestEventRepository_FindActiveBySlug_ReturnsEventWhenApproved(t *testing.T) {
	tx, rollback := beginTx(t)
	defer rollback()

	user := model.User{Name: "Ana Silva", Email: "ana@example.com", Role: model.UserRoleContributor}
	require.NoError(t, tx.Create(&user).Error)

	event := model.Event{
		Name: "ICSE", Slug: "ICSE2027", Country: "USA", City: "Boston",
		Latitude: 42.36, Longitude: -71.05,
		StartDate:  time.Date(2027, 5, 10, 0, 0, 0, 0, time.UTC),
		EndDate:    time.Date(2027, 5, 18, 0, 0, 0, 0, time.UTC),
		WebsiteURL: "https://icse2027.example.org", Domain: "computer_science",
		Status: model.EventStatusApproved, Year: 2027,
		CreatedByID: user.ID, LastUpdatedByID: user.ID,
	}
	require.NoError(t, tx.Create(&event).Error)

	repo := repository.NewEventRepository(tx)

	got, err := repo.FindActiveBySlug(context.Background(), "ICSE2027")

	require.NoError(t, err)
	assert.Equal(t, event.ID, got.ID)
}

func TestEventRepository_FindActiveBySlug_ReturnsErrNotFoundWhenOnlyRejected(t *testing.T) {
	// Spec: events-submit.yaml border_case "slug previously used only by a rejected event → allowed, 201 (slug reusable)"
	tx, rollback := beginTx(t)
	defer rollback()

	user := model.User{Name: "Ana Silva", Email: "ana@example.com", Role: model.UserRoleContributor}
	require.NoError(t, tx.Create(&user).Error)

	event := model.Event{
		Name: "Rejected Conf", Slug: "REJECTED2026", Country: "Brazil", City: "Recife",
		Latitude: -8.04, Longitude: -34.87,
		StartDate:  time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC),
		EndDate:    time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC),
		WebsiteURL: "https://rejected.example.org", Domain: "computer_science",
		Status: model.EventStatusRejected, Year: 2026,
		CreatedByID: user.ID, LastUpdatedByID: user.ID,
	}
	require.NoError(t, tx.Create(&event).Error)

	repo := repository.NewEventRepository(tx)

	_, err := repo.FindActiveBySlug(context.Background(), "REJECTED2026")

	require.ErrorIs(t, err, repository.ErrNotFound)
}

func TestEventRepository_FindActiveBySlug_ReturnsErrNotFoundWhenMissing(t *testing.T) {
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)

	_, err := repo.FindActiveBySlug(context.Background(), "DOES-NOT-EXIST")

	require.ErrorIs(t, err, repository.ErrNotFound)
}

// --- Tier column (migrations/006_add_tier_to_events.sql) ---

func TestEventRepository_TierColumn_DefaultsToUnranked(t *testing.T) {
	// Spec: events-list.yaml schema_changes "tier ... NOT NULL DEFAULT 'unranked'"
	tx, rollback := beginTx(t)
	defer rollback()

	event, deadlines, submitter := newSubmission("MODELS2026", "ana@example.com")
	// Tier intentionally left as the zero value ("") to simulate a row inserted
	// without specifying tier — the column default should apply.
	event.Tier = ""

	repo := repository.NewEventRepository(tx)
	got, err := repo.Submit(context.Background(), event, deadlines, submitter)

	require.NoError(t, err)

	var tier string
	require.NoError(t, tx.Model(&model.Event{}).Where("id = ?", got.ID).Pluck("tier", &tier).Error)
	assert.Equal(t, "unranked", tier)
}

func TestEventRepository_TierColumn_RejectsInvalidValue(t *testing.T) {
	// Spec: events-list.yaml schema_changes "CHECK constraint: tier IN ('A*','A','B','C','unranked')"
	tx, rollback := beginTx(t)
	defer rollback()

	event, _, submitter := newSubmission("MODELS2026", "ana@example.com")
	require.NoError(t, tx.Create(&submitter).Error)
	event.CreatedByID = submitter.ID
	event.LastUpdatedByID = submitter.ID
	event.Tier = "S-tier"

	err := tx.Create(&event).Error

	require.Error(t, err)
}

// --- ListEvents ---

// mustCreateUser creates and returns a persisted User with the given email.
func mustCreateUser(t *testing.T, tx *gorm.DB, email string) model.User {
	t.Helper()
	user := model.User{Name: "Test User", Email: email, Role: model.UserRoleContributor}
	require.NoError(t, tx.Create(&user).Error)
	return user
}

// baseListEvent returns a fresh (unsaved) Event with sensible defaults
// (status=approved, year=2026, domain=computer_science, country=Brazil,
// tier=unranked) for ListEvents tests. Callers override only the fields
// relevant to the behaviour under test.
func baseListEvent(slug string, userID uint) model.Event {
	return model.Event{
		Name: "Event " + slug, Slug: slug, Country: "Brazil", City: "Recife",
		Latitude: -8.04, Longitude: -34.87,
		StartDate:  time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC),
		EndDate:    time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC),
		WebsiteURL: "https://example.org/" + slug, Domain: "computer_science",
		Status: model.EventStatusApproved, Year: 2026, Tier: "unranked",
		CreatedByID: userID, LastUpdatedByID: userID,
	}
}

func TestEventRepository_ListEvents_FiltersByYearAndStatus(t *testing.T) {
	// Spec: events-list.yaml definition_of_done "?year=2025 alone → approved events for 2025."
	tx, rollback := beginTx(t)
	defer rollback()

	user := mustCreateUser(t, tx, "ana@example.com")

	match := baseListEvent("MATCH2026", user.ID)
	require.NoError(t, tx.Create(&match).Error)

	wrongYear := baseListEvent("WRONGYEAR2025", user.ID)
	wrongYear.Year = 2025
	require.NoError(t, tx.Create(&wrongYear).Error)

	pending := baseListEvent("PENDING2026", user.ID)
	pending.Status = model.EventStatusPending
	require.NoError(t, tx.Create(&pending).Error)

	repo := repository.NewEventRepository(tx)

	got, total, err := repo.ListEvents(context.Background(), repository.ListEventsFilter{
		Year: 2026, Status: model.EventStatusApproved, Page: 1, PageSize: 20,
	})

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, match.ID, got[0].ID)
	assert.Equal(t, int64(1), total)
}

func TestEventRepository_ListEvents_FiltersByDomain(t *testing.T) {
	// Spec: events-list.yaml definition_of_done "?domain=computer_science ... filter correctly"
	tx, rollback := beginTx(t)
	defer rollback()

	user := mustCreateUser(t, tx, "ana@example.com")

	csEvent := baseListEvent("CS2026", user.ID)
	csEvent.Domain = "computer_science"
	require.NoError(t, tx.Create(&csEvent).Error)

	otherEvent := baseListEvent("OTHER2026", user.ID)
	otherEvent.Domain = "software_engineering"
	require.NoError(t, tx.Create(&otherEvent).Error)

	repo := repository.NewEventRepository(tx)
	domain := "computer_science"

	got, total, err := repo.ListEvents(context.Background(), repository.ListEventsFilter{
		Year: 2026, Status: model.EventStatusApproved, Page: 1, PageSize: 20, Domain: &domain,
	})

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, csEvent.ID, got[0].ID)
	assert.Equal(t, int64(1), total)
}

func TestEventRepository_ListEvents_FiltersByCountryCaseInsensitiveExactMatch(t *testing.T) {
	// Spec: events-list.yaml border_case "?country=Brazil with an event whose country
	// is 'Brazilian Republic' → NOT included (exact match only, not substring)."
	tx, rollback := beginTx(t)
	defer rollback()

	user := mustCreateUser(t, tx, "ana@example.com")

	brazil := baseListEvent("BRAZIL2026", user.ID)
	brazil.Country = "Brazil"
	require.NoError(t, tx.Create(&brazil).Error)

	brazilianRepublic := baseListEvent("BRAZILIANREP2026", user.ID)
	brazilianRepublic.Country = "Brazilian Republic"
	require.NoError(t, tx.Create(&brazilianRepublic).Error)

	repo := repository.NewEventRepository(tx)
	country := "brazil" // lower-case — must still match "Brazil" exactly (case-insensitive)

	got, total, err := repo.ListEvents(context.Background(), repository.ListEventsFilter{
		Year: 2026, Status: model.EventStatusApproved, Page: 1, PageSize: 20, Country: &country,
	})

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, brazil.ID, got[0].ID)
	assert.Equal(t, int64(1), total)
}

func TestEventRepository_ListEvents_FiltersByTier(t *testing.T) {
	// Spec: events-list.yaml definition_of_done "?tier=A returns only events with tier='A'"
	tx, rollback := beginTx(t)
	defer rollback()

	user := mustCreateUser(t, tx, "ana@example.com")

	tierA := baseListEvent("TIERA2026", user.ID)
	tierA.Tier = "A"
	require.NoError(t, tx.Create(&tierA).Error)

	unranked := baseListEvent("UNRANKED2026", user.ID)
	require.NoError(t, tx.Create(&unranked).Error)

	repo := repository.NewEventRepository(tx)
	tier := "A"

	got, total, err := repo.ListEvents(context.Background(), repository.ListEventsFilter{
		Year: 2026, Status: model.EventStatusApproved, Page: 1, PageSize: 20, Tier: &tier,
	})

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, tierA.ID, got[0].ID)
	assert.Equal(t, int64(1), total)
}

func TestEventRepository_ListEvents_SortedByStartDateAscending(t *testing.T) {
	// Spec: events-list.yaml rule "Results sorted by start_date ascending (soonest upcoming first)."
	tx, rollback := beginTx(t)
	defer rollback()

	user := mustCreateUser(t, tx, "ana@example.com")

	later := baseListEvent("LATER2026", user.ID)
	later.StartDate = time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	later.EndDate = time.Date(2026, 11, 5, 0, 0, 0, 0, time.UTC)
	require.NoError(t, tx.Create(&later).Error)

	earlier := baseListEvent("EARLIER2026", user.ID)
	earlier.StartDate = time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	earlier.EndDate = time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)
	require.NoError(t, tx.Create(&earlier).Error)

	repo := repository.NewEventRepository(tx)

	got, _, err := repo.ListEvents(context.Background(), repository.ListEventsFilter{
		Year: 2026, Status: model.EventStatusApproved, Page: 1, PageSize: 20,
	})

	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, earlier.ID, got[0].ID)
	assert.Equal(t, later.ID, got[1].ID)
}

func TestEventRepository_ListEvents_NoMatches_ReturnsEmptySliceAndZeroCount(t *testing.T) {
	// Spec: events-list.yaml border_case "No events match filters → 200 with data: []
	// and meta.total: 0 (not an error)."
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)

	got, total, err := repo.ListEvents(context.Background(), repository.ListEventsFilter{
		Year: 2026, Status: model.EventStatusApproved, Page: 1, PageSize: 20,
	})

	require.NoError(t, err)
	assert.Empty(t, got)
	assert.Equal(t, int64(0), total)
}

func TestEventRepository_ListEvents_FiltersByBBox(t *testing.T) {
	// Spec: events-list.yaml definition_of_done "?bbox=minLng,minLat,maxLng,maxLat
	// returns only events within the viewport."
	tx, rollback := beginTx(t)
	defer rollback()

	user := mustCreateUser(t, tx, "ana@example.com")

	// Recife, Brazil — inside the bbox below.
	inside := baseListEvent("INSIDE2026", user.ID)
	inside.Latitude, inside.Longitude = -8.04, -34.87
	require.NoError(t, tx.Create(&inside).Error)

	// Boston, USA — outside the bbox below.
	outside := baseListEvent("OUTSIDE2026", user.ID)
	outside.Latitude, outside.Longitude = 42.36, -71.05
	require.NoError(t, tx.Create(&outside).Error)

	repo := repository.NewEventRepository(tx)

	got, total, err := repo.ListEvents(context.Background(), repository.ListEventsFilter{
		Year: 2026, Status: model.EventStatusApproved, Page: 1, PageSize: 20,
		BBox: &repository.BBoxFilter{MinLng: -40, MinLat: -10, MaxLng: -30, MaxLat: -5},
	})

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, inside.ID, got[0].ID)
	assert.Equal(t, int64(1), total)
}

func TestEventRepository_ListEvents_BBoxBoundaryValuesAreInclusive(t *testing.T) {
	// Spec: events-list.yaml rule "bbox filters on latitude/longitude columns"
	// — an event sitting exactly on the viewport edge is still "within" it.
	tx, rollback := beginTx(t)
	defer rollback()

	user := mustCreateUser(t, tx, "ana@example.com")

	onCorner := baseListEvent("CORNER2026", user.ID)
	onCorner.Latitude, onCorner.Longitude = -5, -30
	require.NoError(t, tx.Create(&onCorner).Error)

	repo := repository.NewEventRepository(tx)

	got, _, err := repo.ListEvents(context.Background(), repository.ListEventsFilter{
		Year: 2026, Status: model.EventStatusApproved, Page: 1, PageSize: 20,
		BBox: &repository.BBoxFilter{MinLng: -40, MinLat: -10, MaxLng: -30, MaxLat: -5},
	})

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, onCorner.ID, got[0].ID)
}

func TestEventRepository_ListEvents_FirstDeadlineMonth_MatchesEarliestAbstractOrPaperDeadline(t *testing.T) {
	// Spec: events-list.yaml definition_of_done "?first_deadline_month=<m> returns
	// only events whose earliest active abstract/paper deadline falls in month m."
	tx, rollback := beginTx(t)
	defer rollback()

	user := mustCreateUser(t, tx, "ana@example.com")

	event := baseListEvent("FDM2026", user.ID)
	event.Deadlines = []model.Deadline{
		{Type: model.DeadlineTypeAbstract, Description: "Abstract", Date: time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC), IsActive: true, CreatedByID: user.ID},
		{Type: model.DeadlineTypePaper, Description: "Paper", Date: time.Date(2026, 8, 22, 0, 0, 0, 0, time.UTC), IsActive: true, CreatedByID: user.ID},
	}
	require.NoError(t, tx.Create(&event).Error)

	repo := repository.NewEventRepository(tx)
	month := 6

	got, total, err := repo.ListEvents(context.Background(), repository.ListEventsFilter{
		Year: 2026, Status: model.EventStatusApproved, Page: 1, PageSize: 20, FirstDeadlineMonth: &month,
	})

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, event.ID, got[0].ID)
	assert.Equal(t, int64(1), total)
}

func TestEventRepository_ListEvents_FirstDeadlineMonth_ExcludesEventWithOnlyNotificationDeadlineInMonth(t *testing.T) {
	// Spec: events-list.yaml border_case "?first_deadline_month=6 with an event that
	// has only a 'notification' deadline in June (no abstract/paper deadline) →
	// event NOT included."
	tx, rollback := beginTx(t)
	defer rollback()

	user := mustCreateUser(t, tx, "ana@example.com")

	event := baseListEvent("FDMNOTIFICATION2026", user.ID)
	event.Deadlines = []model.Deadline{
		{Type: model.DeadlineTypeNotification, Description: "Notification", Date: time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC), IsActive: true, CreatedByID: user.ID},
	}
	require.NoError(t, tx.Create(&event).Error)

	repo := repository.NewEventRepository(tx)
	month := 6

	got, total, err := repo.ListEvents(context.Background(), repository.ListEventsFilter{
		Year: 2026, Status: model.EventStatusApproved, Page: 1, PageSize: 20, FirstDeadlineMonth: &month,
	})

	require.NoError(t, err)
	assert.Empty(t, got)
	assert.Equal(t, int64(0), total)
}

func TestEventRepository_ListEvents_FirstDeadlineMonth_ExcludesEventWhenEarliestDeadlineYearMismatches(t *testing.T) {
	// Spec: events-list.yaml border_case "?year=2026&first_deadline_month=6 with an
	// event whose earliest abstract/paper deadline is June 2025 (not 2026) → event
	// NOT included (year mismatch)."
	tx, rollback := beginTx(t)
	defer rollback()

	user := mustCreateUser(t, tx, "ana@example.com")

	event := baseListEvent("FDMYEARMISMATCH2026", user.ID)
	event.Deadlines = []model.Deadline{
		{Type: model.DeadlineTypeAbstract, Description: "Abstract", Date: time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC), IsActive: true, CreatedByID: user.ID},
	}
	require.NoError(t, tx.Create(&event).Error)

	repo := repository.NewEventRepository(tx)
	month := 6

	got, total, err := repo.ListEvents(context.Background(), repository.ListEventsFilter{
		Year: 2026, Status: model.EventStatusApproved, Page: 1, PageSize: 20, FirstDeadlineMonth: &month,
	})

	require.NoError(t, err)
	assert.Empty(t, got)
	assert.Equal(t, int64(0), total)
}

func TestEventRepository_ListEvents_FirstDeadlineMonth_UsesEarliestWhenLaterDeadlineMatchesMonth(t *testing.T) {
	// Spec: events-list.yaml rule "Filters events whose earliest is_active=true
	// deadline of type 'abstract' or 'paper' falls in the given calendar month" —
	// a later deadline matching the requested month must NOT cause a match if it
	// isn't the earliest one.
	tx, rollback := beginTx(t)
	defer rollback()

	user := mustCreateUser(t, tx, "ana@example.com")

	event := baseListEvent("FDMNOTEARLIEST2026", user.ID)
	event.Deadlines = []model.Deadline{
		{Type: model.DeadlineTypeAbstract, Description: "Abstract", Date: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), IsActive: true, CreatedByID: user.ID},
		{Type: model.DeadlineTypePaper, Description: "Paper", Date: time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC), IsActive: true, CreatedByID: user.ID},
	}
	require.NoError(t, tx.Create(&event).Error)

	repo := repository.NewEventRepository(tx)
	month := 6

	got, total, err := repo.ListEvents(context.Background(), repository.ListEventsFilter{
		Year: 2026, Status: model.EventStatusApproved, Page: 1, PageSize: 20, FirstDeadlineMonth: &month,
	})

	require.NoError(t, err)
	assert.Empty(t, got)
	assert.Equal(t, int64(0), total)
}

func TestEventRepository_ListEvents_Pagination_ReturnsCorrectSliceAndTotal(t *testing.T) {
	// Spec: events-list.yaml definition_of_done "Pagination: ?page=2&page_size=10
	// returns the correct slice; meta.total reflects the full filtered count."
	tx, rollback := beginTx(t)
	defer rollback()

	user := mustCreateUser(t, tx, "ana@example.com")

	first := baseListEvent("PAGE1A2026", user.ID)
	first.StartDate = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	first.EndDate = time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	require.NoError(t, tx.Create(&first).Error)

	second := baseListEvent("PAGE1B2026", user.ID)
	second.StartDate = time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	second.EndDate = time.Date(2026, 2, 5, 0, 0, 0, 0, time.UTC)
	require.NoError(t, tx.Create(&second).Error)

	third := baseListEvent("PAGE2A2026", user.ID)
	third.StartDate = time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	third.EndDate = time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)
	require.NoError(t, tx.Create(&third).Error)

	repo := repository.NewEventRepository(tx)

	page1, total1, err := repo.ListEvents(context.Background(), repository.ListEventsFilter{
		Year: 2026, Status: model.EventStatusApproved, Page: 1, PageSize: 2,
	})
	require.NoError(t, err)
	require.Len(t, page1, 2)
	assert.Equal(t, first.ID, page1[0].ID)
	assert.Equal(t, second.ID, page1[1].ID)
	assert.Equal(t, int64(3), total1)

	page2, total2, err := repo.ListEvents(context.Background(), repository.ListEventsFilter{
		Year: 2026, Status: model.EventStatusApproved, Page: 2, PageSize: 2,
	})
	require.NoError(t, err)
	require.Len(t, page2, 1)
	assert.Equal(t, third.ID, page2[0].ID)
	assert.Equal(t, int64(3), total2)
}

func TestEventRepository_ListEvents_PaginationOff_ReturnsAllMatchingRows(t *testing.T) {
	// Spec: events-list.yaml definition_of_done "?pagination=off returns all
	// matching rows in one response with meta.page=1."
	tx, rollback := beginTx(t)
	defer rollback()

	user := mustCreateUser(t, tx, "ana@example.com")

	first := baseListEvent("OFF1_2026", user.ID)
	first.StartDate = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	first.EndDate = time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	require.NoError(t, tx.Create(&first).Error)

	second := baseListEvent("OFF2_2026", user.ID)
	second.StartDate = time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	second.EndDate = time.Date(2026, 2, 5, 0, 0, 0, 0, time.UTC)
	require.NoError(t, tx.Create(&second).Error)

	repo := repository.NewEventRepository(tx)

	// Even though Page/PageSize are set to a tiny page, PaginationOff must win
	// and return every matching row.
	got, total, err := repo.ListEvents(context.Background(), repository.ListEventsFilter{
		Year: 2026, Status: model.EventStatusApproved, Page: 1, PageSize: 1, PaginationOff: true,
	})

	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, first.ID, got[0].ID)
	assert.Equal(t, second.ID, got[1].ID)
	assert.Equal(t, int64(2), total)
}

func TestEventRepository_ListEvents_IncludesCreatedByAndLastUpdatedBy(t *testing.T) {
	// Spec: events-list.yaml definition_of_done "Each returned event includes
	// created_by and last_updated_by."
	tx, rollback := beginTx(t)
	defer rollback()

	creator := mustCreateUser(t, tx, "creator@example.com")
	updater := mustCreateUser(t, tx, "updater@example.com")

	event := baseListEvent("ATTRIBUTION2026", creator.ID)
	event.LastUpdatedByID = updater.ID
	require.NoError(t, tx.Create(&event).Error)

	repo := repository.NewEventRepository(tx)

	got, _, err := repo.ListEvents(context.Background(), repository.ListEventsFilter{
		Year: 2026, Status: model.EventStatusApproved, Page: 1, PageSize: 20,
	})

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "creator@example.com", got[0].CreatedBy.Email)
	assert.Equal(t, "updater@example.com", got[0].LastUpdatedBy.Email)
}

func TestEventRepository_ListEvents_OnlyIncludesActiveDeadlines(t *testing.T) {
	// Spec: events-list.yaml rule "Each event's deadlines array includes only
	// is_active=true rows; full history remains available via
	// GET /api/v1/events/:id/deadlines."
	tx, rollback := beginTx(t)
	defer rollback()

	user := mustCreateUser(t, tx, "ana@example.com")

	event := baseListEvent("DEADLINES2026", user.ID)
	event.Deadlines = []model.Deadline{
		{Type: model.DeadlineTypePaper, Description: "Old paper deadline", Date: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC), IsActive: true, CreatedByID: user.ID},
		{Type: model.DeadlineTypePaper, Description: "Current paper deadline", Date: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), IsActive: true, CreatedByID: user.ID},
	}
	require.NoError(t, tx.Create(&event).Error)

	// GORM skips zero-value fields when a column has a `default` tag, so a
	// create-time IsActive: false would silently become true (the column
	// default). Superseding the old deadline via an explicit UPDATE — exactly
	// how the application does it per the "Deadlines are immutable" rule —
	// avoids that gotcha.
	require.NoError(t, tx.Model(&event.Deadlines[0]).Update("is_active", false).Error)

	repo := repository.NewEventRepository(tx)

	got, _, err := repo.ListEvents(context.Background(), repository.ListEventsFilter{
		Year: 2026, Status: model.EventStatusApproved, Page: 1, PageSize: 20,
	})

	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Len(t, got[0].Deadlines, 1)
	assert.Equal(t, "Current paper deadline", got[0].Deadlines[0].Description)
	assert.True(t, got[0].Deadlines[0].IsActive)
}

// --- Submit ---

// newSubmission returns a fresh (unsaved) Event, its Deadlines, and Submitter,
// the same shapes service.BuildSubmission returns to the handler.
func newSubmission(slug, email string) (model.Event, []model.Deadline, model.User) {
	event := model.Event{
		Name: "MODELS", Slug: slug, Country: "Brazil", City: "Recife",
		Latitude: -8.04, Longitude: -34.87,
		StartDate:  time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC),
		EndDate:    time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC),
		WebsiteURL: "https://models2026.example.org", Domain: "computer_science",
		Status: model.EventStatusPending, Year: 2026,
	}
	submitter := model.User{Name: "Ana Silva", Email: email, Role: model.UserRoleContributor}
	return event, nil, submitter
}

func TestEventRepository_Submit_CreatesNewContributorWhenEmailNotFound(t *testing.T) {
	// Spec: events-submit.yaml rule "If no User exists, create one with role=contributor, password_hash=NULL"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event, deadlines, submitter := newSubmission("MODELS2026", "new@example.com")

	got, err := repo.Submit(context.Background(), event, deadlines, submitter)

	require.NoError(t, err)
	assert.NotZero(t, got.CreatedByID)
	assert.Equal(t, "new@example.com", got.CreatedBy.Email)
	assert.Equal(t, model.UserRoleContributor, got.CreatedBy.Role)
	assert.Nil(t, got.CreatedBy.PasswordHash)
}

func TestEventRepository_Submit_ReusesAndRenamesExistingUser(t *testing.T) {
	// Spec: events-submit.yaml rule "If a User with that email exists, link event.created_by_id to
	// that user and update their name to the submitted submitter.name (always overwritten)"
	tx, rollback := beginTx(t)
	defer rollback()

	existing := model.User{Name: "Old Name", Email: "existing@example.com", Role: model.UserRoleAdmin}
	require.NoError(t, tx.Create(&existing).Error)

	repo := repository.NewEventRepository(tx)
	event, deadlines, submitter := newSubmission("MODELS2026", "existing@example.com")
	submitter.Name = "New Name"

	got, err := repo.Submit(context.Background(), event, deadlines, submitter)

	require.NoError(t, err)
	assert.Equal(t, existing.ID, got.CreatedByID)
	assert.Equal(t, "New Name", got.CreatedBy.Name)
	assert.Equal(t, model.UserRoleAdmin, got.CreatedBy.Role) // role is never changed by submission
}

func TestEventRepository_Submit_CreatesEventWithStatusPending(t *testing.T) {
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event, deadlines, submitter := newSubmission("MODELS2026", "ana@example.com")

	got, err := repo.Submit(context.Background(), event, deadlines, submitter)

	require.NoError(t, err)
	assert.NotZero(t, got.ID)
	assert.Equal(t, model.EventStatusPending, got.Status)
}

func TestEventRepository_Submit_CreatesDeadlinesWithIsActiveTrue(t *testing.T) {
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event, _, submitter := newSubmission("MODELS2026", "ana@example.com")
	deadlines := []model.Deadline{
		{Type: model.DeadlineTypePaper, Description: "Research track full paper", Date: time.Date(2026, 8, 22, 0, 0, 0, 0, time.UTC), IsActive: true},
	}

	got, err := repo.Submit(context.Background(), event, deadlines, submitter)

	require.NoError(t, err)
	require.Len(t, got.Deadlines, 1)
	assert.True(t, got.Deadlines[0].IsActive)
	assert.Equal(t, got.ID, got.Deadlines[0].EventID)
	assert.NotZero(t, got.Deadlines[0].CreatedByID)
}

func TestEventRepository_Submit_NoDeadlinesWhenNoneProvided(t *testing.T) {
	// Spec: events-submit.yaml border_case "deadlines omitted or empty array → allowed, event created with no deadlines"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event, deadlines, submitter := newSubmission("MODELS2026", "ana@example.com")

	got, err := repo.Submit(context.Background(), event, deadlines, submitter)

	require.NoError(t, err)
	assert.Empty(t, got.Deadlines)
}

func TestEventRepository_Submit_WritesAuditLogForEventCreation(t *testing.T) {
	// Spec: events-submit.yaml rule "AuditLog row written for the Event (entity_type=event, action=created, ...)"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event, deadlines, submitter := newSubmission("MODELS2026", "ana@example.com")

	got, err := repo.Submit(context.Background(), event, deadlines, submitter)
	require.NoError(t, err)

	var log model.AuditLog
	err = tx.Where("entity_type = ? AND entity_id = ? AND action = ?",
		model.AuditEntityEvent, got.ID, model.AuditActionCreated).First(&log).Error

	require.NoError(t, err)
	assert.Equal(t, got.CreatedByID, log.ChangedByID)
}

func TestEventRepository_Submit_WritesAuditLogPerDeadline(t *testing.T) {
	// Spec: events-submit.yaml rule "AuditLog row written for each Deadline created
	// (entity_type=deadline, action=deadline_added, ...)"
	tx, rollback := beginTx(t)
	defer rollback()

	repo := repository.NewEventRepository(tx)
	event, _, submitter := newSubmission("MODELS2026", "ana@example.com")
	deadlines := []model.Deadline{
		{Type: model.DeadlineTypeAbstract, Description: "Abstract", Date: time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC), IsActive: true},
		{Type: model.DeadlineTypePaper, Description: "Paper", Date: time.Date(2026, 8, 22, 0, 0, 0, 0, time.UTC), IsActive: true},
	}

	got, err := repo.Submit(context.Background(), event, deadlines, submitter)
	require.NoError(t, err)
	require.Len(t, got.Deadlines, 2)

	var count int64
	err = tx.Model(&model.AuditLog{}).
		Where("entity_type = ? AND action = ? AND entity_id IN ?",
			model.AuditEntityDeadline, model.AuditActionDeadlineAdded,
			[]uint{got.Deadlines[0].ID, got.Deadlines[1].ID}).
		Count(&count).Error

	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}
