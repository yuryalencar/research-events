# Session: Supersede a Deadline (`POST /api/v1/events/{eventId}/deadlines/{deadlineId}/supersede`)

**Date:** 2026-06-14
**Duration:** Full session (spec approved in a prior session; this session covered Phase 1–6)

---

## Goal

Implement the "extend/replace a deadline" flow: any contributor can replace an
active deadline on an already-approved event with a new one carrying an
updated date (and optionally time/timezone), following the full Pair
Programming Workflow (Phase 1 Discovery → ... → Phase 6 Documentation) and
strict Red-Green-Refactor TDD, across 4 cycles:

- `POST /api/v1/events/{eventId}/deadlines/{deadlineId}/supersede` — public, no auth
- Body: `{ submitter: {name, email}, date: "YYYY-MM-DD", time?: "HH:MM", timezone?: string }`
- The new `Deadline` row inherits `type`, `description`, and `is_optional`
  from the old row — only `date`, `time`, `timezone` come from the request,
  and `time`/`timezone` are **never inherited** from the old row (omitting
  them means `time=null, timezone=null` on the new row even if the old row
  had values)
- New row created with `is_active=true, superseded_by_id=nil`; old row
  updated to `is_active=false, superseded_by_id=<new row's ID>`
- 200 + full updated event (`toEventListItemResponse`); `deadlines` now
  includes **both** the new active deadline and the just-superseded one
  (`is_active=false, superseded_by_id=<new id>`)
- `deadlineResponse` gains `superseded_by_id` (`number | null`) — returned
  everywhere deadlines are returned (submit, list, add, cancel, supersede)
- `preloadEventAssociations` changed from `is_active = true` to
  `is_active = true OR superseded_by_id IS NOT NULL`
- 404 `EVENT_NOT_FOUND` / `DEADLINE_NOT_FOUND` for missing/non-numeric path
  IDs or a deadline belonging to a different event
- 409 `EVENT_NOT_APPROVED` when the event is pending/rejected
- 409 `DEADLINE_ALREADY_INACTIVE` when the target deadline is already
  cancelled or superseded
- No validation that the new date is later than the old date — same or
  earlier dates are accepted
- Audit logging: one `deadline_superseded` row (entity=old deadline, diff
  records before/after for `date`/`time`/`timezone`/`is_active`/
  `superseded_by_id`) plus the always-written `updated` row for
  `last_updated_by_id`
- Shared 50 req/min `publicRateLimiter` with `/events/submit`,
  `/events/{id}/deadlines`, and the cancel endpoint

---

## What was built

### New files
- `specs/backend/events-deadlines-supersede.curl.sh` — curl examples for every
  response code and border case

### Modified files
- `backend/internal/service/event_deadlines.go` — added
  `SupersedeDeadlineInput`, `ValidateSupersedeDeadlineInput`
  (`// FP: pure function`), and `BuildSupersedingDeadline`
  (`// FP: immutability`)
- `backend/internal/service/event_deadlines_test.go` — 12 new tests for
  `ValidateSupersedeDeadlineInput` and `BuildSupersedingDeadline`
- `backend/internal/repository/event.go` — added `SupersedeDeadline` to the
  `EventRepository` interface; `preloadEventAssociations` now preloads
  `is_active = true OR superseded_by_id IS NOT NULL`
- `backend/internal/repository/event_deadlines.go` — `SupersedeDeadline`
  (transactional: find-or-create submitter, insert new deadline, mark old
  deadline inactive + set `superseded_by_id`, update `last_updated_by_id`,
  write audit rows); `findByIDWithActiveDeadlines` doc comment updated;
  added `buildSupersedeDeadlineAuditLogs` and `buildDeadlineSupersedeDiff`
- `backend/internal/repository/event_deadlines_test.go` — 7 new integration
  tests against the real test Postgres (constraint/preload check +
  `SupersedeDeadline` x6) plus a `newSupersedingDeadline` helper
- `backend/internal/repository/event_test.go` — 1 new test confirming the
  updated preload returns both active and superseded deadlines
- `backend/internal/repository/mocks/mock_event.go` — regenerated for
  `SupersedeDeadline`
- `backend/internal/handler/event.go` — `deadlineResponse` gains
  `superseded_by_id`; `toDeadlineResponses` maps it
- `backend/internal/handler/event_deadlines.go` — `Supersede` handler,
  `supersedeDeadlineRequest`, `toSupersedeDeadlineInput`
- `backend/internal/handler/event_deadlines_test.go` — 16 new handler tests
  covering every response code and border case
- `backend/cmd/api/server.go` — registered
  `POST /api/v1/events/{eventId}/deadlines/{deadlineId}/supersede` on
  `publicRateLimiter`

---

## Key design decisions

### Time/timezone are a one-way data flow from the request, never from `old`
`BuildSupersedingDeadline` takes `Date`, `Time`, `Timezone` only from
`input` (the request body) — `old.Time`/`old.Timezone` are never read. This
makes the "never inherit time/timezone" spec rule structurally enforced: the
function simply has no path by which `old`'s time/timezone could end up on
the new row.

### `preloadEventAssociations` widened to include superseded rows
Changing the `Deadlines` preload condition to
`is_active = true OR superseded_by_id IS NOT NULL` is the single change that
makes every reload endpoint (list, add, cancel, supersede) return superseded
deadlines alongside active ones — no per-endpoint changes were needed beyond
this shared helper.

### `SupersedeDeadline` reuses the `CancelDeadline` transaction shape
Find-or-create submitter → mutate deadline rows → update
`event.last_updated_by_id` → write audit rows → reload via
`findByIDWithActiveDeadlines`. The only new step is inserting the replacement
row before marking the old one inactive, so the new row's ID is available for
`old.superseded_by_id` and the audit diff.

### Audit diff captures the full before/after deadline state
`buildDeadlineSupersedeDiff` records `date`, `time`, `timezone`, `is_active`,
and `superseded_by_id` before/after on the **old** deadline's
`deadline_superseded` audit row — sufficient to reconstruct what changed
without needing to look up the new row.

---

## FP concepts introduced/reinforced (`internal/service/`)

- **Pure function** (reinforced) — `ValidateSupersedeDeadlineInput`: given the
  same `SupersedeDeadlineInput`, always returns the same error (or `nil`),
  with no I/O. Reuses `validateSubmitterInput` and `deadlineTimePattern` from
  prior features.
- **Immutability** (reinforced) — `BuildSupersedingDeadline`: never mutates
  `old`; returns a brand-new `model.Deadline` with `Date`/`Time`/`Timezone`
  taken only from `input`. The contrast with an imperative version (mutating
  a copy of `old` in place and overwriting fields) is that immutability makes
  "what data flows from where" visible directly in the return expression —
  there's no intermediate mutable state where a stray `old.Time` reference
  could sneak back in.

---

## Test count at end of session

| Package | Tests |
|---|---|
| `cmd/api` | 9 |
| `internal/config` | 6 |
| `internal/handler` | 96 |
| `internal/health` | 7 |
| `internal/middleware` | 18 |
| `internal/repository` | 74 |
| `internal/service` | 112 |
| **Total** | **322** |

All passing (`go test ./...`). `go vet ./...` and `gofmt -l .` clean (only the
3 pre-existing unrelated files flagged, unchanged by this diff).

---

## State at end of session

`POST /api/v1/events/{eventId}/deadlines/{deadlineId}/supersede` is fully
implemented, tested, and wired into the router. To verify:

```bash
docker compose up -d
make migrate-up
make migrate-test-up
go test ./...               # all 322 tests green
go vet ./...                # zero warnings
cd backend && go run ./cmd/api
bash specs/backend/events-deadlines-supersede.curl.sh
```

---

## Context to restore

- `preloadEventAssociations` (in `internal/repository/event.go`) now preloads
  `Deadlines` with `is_active = ? OR superseded_by_id IS NOT NULL` — any
  future endpoint reading `event.Deadlines` will see superseded rows too, not
  just active ones.
- `deadlineResponse.superseded_by_id` is now part of the public API contract
  for every endpoint that returns deadlines.
- `BuildSupersedingDeadline` / `ValidateSupersedeDeadlineInput` /
  `SupersedeDeadline` follow the same shape as the cancel feature's
  equivalents — reuse this shape for any future deadline-mutation endpoint.
- Other pending features from prior sessions remain open: `GET
  /api/v1/events/:id`, `GET /api/v1/events/:id/audit`, and the admin review
  queue (`GET /api/v1/admin/events?status=pending`, `PATCH
  /api/v1/admin/events/:id/review`).
