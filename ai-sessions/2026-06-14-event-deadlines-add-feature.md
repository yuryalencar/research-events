# Session: Add Deadlines to an Approved Event (`POST /api/v1/events/{id}/deadlines`)

**Date:** 2026-06-14
**Duration:** Full session

---

## Goal

Implement the collaborative "keep deadlines fresh" feature: any contributor can
add one or more deadlines to an already-approved event, following the full
Pair Programming Workflow (Phase 0 Spec → ... → Phase 6 Documentation) and
strict Red-Green-Refactor TDD, across 8 cycles:

- `POST /api/v1/events/{id}/deadlines` — public, no auth
- Body: `{ submitter: {name, email}, deadlines: [...] }`, `deadlines` required (min 1)
- 201 + full updated event (reusing `toEventListItemResponse`)
- 404 `EVENT_NOT_FOUND` for missing/non-numeric `:id` (never reveal IDs are numeric)
- 409 `EVENT_NOT_APPROVED` when the event is pending/rejected
- No automatic supersession — duplicate deadlines are allowed (documented limitation)
- Audit logging: `deadline_added` (1 deadline) or `batch_deadlines_added` (>1),
  plus an always-written `updated` row for the `last_updated_by_id` change
- Shared 50 req/min rate limit with `/events/submit`

---

## What was built

### New files
- `specs/backend/events-deadlines-add.yaml` — full spec, including a
  `known_limitations` section documenting the future "update deadline"/supersession
  endpoint and its effect on the event response shape
- `specs/backend/events-deadlines-add.curl.sh` — curl examples for every response code
- `backend/migrations/007_add_batch_deadlines_added_audit_action.sql` — adds
  `batch_deadlines_added` to the `audit_logs_action_check` CHECK constraint
- `backend/internal/service/event_deadlines.go` — moved `DeadlineInput`,
  `allowedDeadlineTypes`, `BuildDeadlinesFromInput`, `validateDeadlineInput` out of
  `event.go` (pure refactor, Cycle 1); added `AddDeadlinesInput`,
  `ValidateAddDeadlinesInput` (`// FP: pure function`), and
  `DetermineDeadlinesAuditAction` (`// FP: pure function`)
- `backend/internal/service/event_deadlines_test.go` — moved
  `TestBuildDeadlinesFromInput_*` here, plus 11 new tests for
  `ValidateAddDeadlinesInput` and `DetermineDeadlinesAuditAction`
- `backend/internal/repository/event_deadlines.go` — `FindByID`, `AddDeadlines`
  (transactional: find-or-create submitter, insert deadlines, update
  `last_updated_by_id`, write audit rows), plus private helpers
  `findByIDWithActiveDeadlines`, `buildAddDeadlinesAuditLogs`, `buildBatchDeadlinesDiff`
- `backend/internal/repository/event_deadlines_test.go` — 11 integration tests
  (constraint check, `FindByID` x2, `AddDeadlines` x9) against a real test Postgres
- `backend/internal/handler/event_deadlines.go` — `AddDeadlines` handler,
  `addDeadlinesRequest`, `toAddDeadlinesInput`, and shared `toDeadlineInputs`
- `backend/internal/handler/event_deadlines_test.go` — 12 handler tests covering
  every response code and border case

### Modified files
- `backend/internal/model/audit_log.go` — added `AuditActionBatchDeadlinesAdded = "batch_deadlines_added"`
- `backend/internal/service/event.go` — removed the deadline-related code now in `event_deadlines.go`
- `backend/internal/repository/event.go` — added `EventRepository.FindByID` and
  `AddDeadlines` to the interface; extracted `preloadEventAssociations` (shared
  `CreatedBy`/`LastUpdatedBy`/active-`Deadlines` preload chain, used by both
  `ListEvents` and `AddDeadlines`'s reload)
- `backend/internal/repository/mocks/mock_event.go` — regenerated for the two new methods
- `backend/internal/handler/event.go` — `toSubmitEventInput` now calls the shared
  `toDeadlineInputs` instead of duplicating the date-parsing loop; removed unused `strconv` import
- `backend/cmd/api/server.go` — registered
  `POST /api/v1/events/{id}/deadlines`; renamed `submitRateLimiter` →
  `publicRateLimiter` (now shared by `/events/submit` and this route) and
  `listRateLimiter` → `publicHighRateLimiter`

---

## Key design decisions

### Everything deadline-related lives in `*_deadlines.go` files
Across service, repository, and handler packages, all deadline-specific types,
validation, and persistence code moved to (or was written directly in)
`event_deadlines.go` / `event_deadlines_test.go`, keeping `event.go` focused on
the event-submission flow.

### One audit row per "logical" change, not per row written
Adding N deadlines writes exactly 2 audit rows total: one
`deadline_added`/`batch_deadlines_added` row describing the deadline change(s),
and one `updated` row for the `last_updated_by_id` change — establishing the
pattern for all future event-update endpoints (per CLAUDE.md's "any update
writes an AuditLog row and updates `last_updated_by_id`" rule).

### `preloadEventAssociations` extracted during refactor
`ListEvents` and `AddDeadlines`'s post-transaction reload both need
`CreatedBy`/`LastUpdatedBy`/active-`Deadlines` preloaded — extracted into one
helper in `internal/repository/event.go` to remove the duplicated preload chain.

### No automatic supersession (documented limitation)
Duplicate deadlines (same type/description/date) are accepted as new active
rows — there's currently no reliable signal to distinguish "this is an update
to an existing deadline" from "this is a genuinely new deadline with a similar
description." The spec's `known_limitations` section documents this for the
future update/supersession endpoint, including the open question of whether
the summary event response should eventually show superseded ↔ replacement
deadline pairs.

### `:id` non-numeric and "not found" are indistinguishable
Both return `404 EVENT_NOT_FOUND` — matches the existing `events-submit.yaml`
convention of never revealing implementation details (here, that IDs are numeric).

---

## FP concepts introduced/reinforced (`internal/service/event_deadlines.go`)

- **Pure function** — `ValidateAddDeadlinesInput`: given the same
  `AddDeadlinesInput`, always returns the same error (or `nil`), with no I/O —
  the event's existence/approval status is checked later, in the repository
  layer, which is the only layer with DB access.
- **Pure function** — `DetermineDeadlinesAuditAction`: maps a deadline count to
  an `model.AuditAction` with no I/O — same input always produces the same
  output, which is what `TestDetermineDeadlinesAuditAction_SameInputTwice_ReturnsIdenticalResult` verifies.
- **Immutability** (reinforced) — `BuildDeadlinesFromInput` (moved, unchanged):
  returns a fresh `[]model.Deadline` slice, never mutating its input.

---

## Test count at end of session

| Package | Tests |
|---|---|
| `cmd/api` | 9 |
| `internal/config` | 6 |
| `internal/handler` | 63 (+12) |
| `internal/health` | 7 |
| `internal/middleware` | 18 |
| `internal/repository` | 55 (+11) |
| `internal/service` | 84 (+11, net of 4 moved) |
| **Total** | **242** |

All passing. `go vet ./...` and `gofmt -l .` (for files in this diff) are clean.

---

## State at end of session

`POST /api/v1/events/{id}/deadlines` is fully implemented, tested, and wired
into the router. To verify:

```bash
docker compose up -d
make migrate-up
make migrate-test-up
go test ./...               # all 242 tests green
go vet ./...                # zero warnings
cd backend && go run ./cmd/api
bash specs/backend/events-deadlines-add.curl.sh
```

---

## Context to restore

- All deadline-specific service/repository/handler code lives in
  `event_deadlines.go` / `event_deadlines_test.go` alongside each package's
  `event.go` — follow this split for any future deadline work (e.g. the
  update/supersession endpoint).
- `publicRateLimiter` (50/min) is now shared by `/events/submit` and
  `/events/{id}/deadlines`; `publicHighRateLimiter` (120/min, burst 30) is
  `/events` only. Both live in `cmd/api/server.go`.
- `preloadEventAssociations` in `internal/repository/event.go` is the single
  place that defines "the full event shape" (CreatedBy, LastUpdatedBy, active
  Deadlines) — reuse it for any new endpoint that returns a full event.
- Next likely feature: the deadline **update/supersession** endpoint —
  marks an existing `Deadline` `is_active=false`, sets `superseded_by_id`,
  inserts the replacement row, writes a `deadline_superseded` audit row. Its
  spec should also resolve the `known_limitations` question in
  `events-deadlines-add.yaml` about whether the summary event response needs
  an optional "previous version" field for superseded deadlines.
- Other pending features from prior sessions remain open: `GET
  /api/v1/events/:id`, `GET /api/v1/events/:id/audit`, and the admin review
  queue (`GET /api/v1/admin/events?status=pending`, `PATCH
  /api/v1/admin/events/:id/review`).
