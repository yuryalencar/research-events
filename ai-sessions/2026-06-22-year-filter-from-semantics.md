# Year filter: "from year" semantics

**Date:** 2026-06-22
**Session goal:** Change `?year=` from exact-match to "from year" semantics on both backend and frontend, and update all related labels and constraints.

---

## What changed

### Backend

- `ListEventsFilter.Year`: `int` → `*int` (nil = no constraint; non-nil = `year >= *Year`)
- `ListEventsInput.Year`: `int` → `*int`; `ValidateListEventsQuery` no longer defaults to current year when param is omitted
- `applyListEventsFilters`: rewrote year clause to `WHERE year >= ?` when non-nil; also updated `firstDeadlineMonth` subquery — when year is provided the subquery now scopes to `EXTRACT(YEAR FROM d.date) >= ?`, when year is omitted the YEAR extraction is dropped entirely
- Handler mapping in `event_list.go` was already `*int` compatible — no change needed there
- All tests updated: new cases for `nil` year (returns all), non-nil year (returns `>=`), and `firstDeadlineMonth` without a year

### Frontend

- `EventFilters.year`: `number` → `number | undefined` (undefined = fetch all events)
- `ReviewFilters.year`: `number` → `number | undefined`
- `useFilters.setYear`: now accepts `number | undefined`
- `useEvents.toListEventsParams`: omits `year` key entirely when undefined
- `useReviewEvents`: same pattern — year omitted from fetch params when undefined
- `ManageDashboard`: number input shows empty/placeholder "All years" when year is undefined; × clear button sets year to undefined
- `FilterPanel` (globe): year stepper has no × clear button — globe ALWAYS has a year set; minimum year is `currentYear - 2` (not a static constant)
- `Step1Search` (submission wizard): year input has `min={new Date().getFullYear()}` — only current year and above allowed

### Label changes (all 4 locales: en / pt / es / de)

- `filters.yearLabel`: "Year" → "From year" / "A partir do ano" / "Desde el año" / "Ab dem Jahr"
- `filters.allYears` (new): "All years" / "Todos os anos" / "Todos los años" / "Alle Jahre"
- `manage.reviewDashboard.yearLabel`: same → "From year" in each locale
- `manage.reviewDashboard.allYears` (new): same "All years" equivalents
- `submit.step1.yearLabel`: same "From year" change in each locale

---

## Decisions made

- Globe page keeps year non-nullable (always shows a year) — user cannot clear it to "all years" on the globe view
- Submission Step 1 min year = `currentYear` (not `currentYear - 2`) — searching past conferences while submitting a new one is not a valid use case
- `FilterPanel` lower bound = `currentYear - 2` — lets users scroll back up to 2 past years to see recent conferences
- `firstDeadlineMonth` subquery updated to match `>=` semantics only when a year is provided; without a year it matches deadlines in any year

---

## State at end

- Typecheck: clean (0 errors)
- Tests: 278/278 passing
- Not yet committed

## Context to restore

- Specs: `specs/backend/events-list-year-from-semantics.yaml`, `specs/frontend/year-filter-from-semantics.md`
- Backend files changed: `internal/service/event_list.go`, `internal/repository/event.go`, plus their test files and `internal/handler/event_list_test.go`
- Frontend files changed: `useFilters.ts`, `useEvents.ts`, `useReviewEvents.ts`, `FilterPanel.tsx`, `ManageDashboard.tsx`, `Step1Search.tsx`, all 4 locale JSON files
