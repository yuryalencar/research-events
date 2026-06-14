# Session: Cancel a Deadline (`PATCH /api/v1/events/{eventId}/deadlines/{deadlineId}/cancel`)

**Date:** 2026-06-14
**Duration:** Full session

---

## Goal

Implement the "I registered this deadline wrong" flow: any contributor can
cancel an active deadline on an already-approved event, following the full
Pair Programming Workflow (Phase 0 Spec → ... → Phase 6 Documentation) and
strict Red-Green-Refactor TDD, across 7 cycles:

- `PATCH /api/v1/events/{eventId}/deadlines/{deadlineId}/cancel` — public, no auth
- Body: `{ submitter: {name, email} }`
- 200 + full updated event (reusing `toEventListItemResponse`); the cancelled
  deadline is absent from `deadlines`
- Cancelling sets `is_active=false` and leaves `superseded_by_id=NULL` —
  distinguishing "cancelled, no replacement" from "superseded, see
  `superseded_by_id`"
- 404 `EVENT_NOT_FOUND` / `DEADLINE_NOT_FOUND` for missing or non-numeric path
  IDs, or a deadline belonging to a different event (never reveal whether IDs
  are numeric or under which event a deadline exists)
- 409 `EVENT_NOT_APPROVED` when the event is pending/rejected
- 409 `DEADLINE_ALREADY_INACTIVE` when the target deadline is already
  cancelled or superseded (caller cannot distinguish the two from this response)
- Cancelling the last remaining active deadline is allowed → `deadlines: []`
- Audit logging: new `deadline_cancelled` action (entity_type=deadline) plus
  the always-written `updated` row for the `last_updated_by_id` change
- Shared 50 req/min `publicRateLimiter` with `/events/submit` and
  `/events/{id}/deadlines`

---

## What was built

### New files
- `specs/backend/events-deadlines-cancel.yaml` — full spec
- `specs/backend/events-deadlines-cancel.curl.sh` — curl examples for every response code
- `backend/migrations/008_add_deadline_cancelled_audit_action.sql` — adds
  `deadline_cancelled` to the `audit_logs_action_check` CHECK constraint

### Modified files
- `backend/internal/model/audit_log.go` — added
  `AuditActionDeadlineCancelled = "deadline_cancelled"`
- `backend/internal/service/event.go` — extracted `validateSubmitterInput`
  (shared `submitter.name`/`submitter.email` validation), now used by
  `ValidateSubmitEventInput`, `ValidateAddDeadlinesInput`, and the new
  `ValidateCancelDeadlineInput`
- `backend/internal/service/event_deadlines.go` — added `CancelDeadlineInput`
  (`{ Submitter SubmitterInput }`), `ValidateCancelDeadlineInput`
  (`// FP: pure function`), and `ValidateDeadlineCancellable`
  (`// FP: pure function`)
- `backend/internal/service/event_deadlines_test.go` — 8 new tests for
  `ValidateCancelDeadlineInput` and `ValidateDeadlineCancellable`
- `backend/internal/repository/event.go` — added `FindDeadlineByID` and
  `CancelDeadline` to the `EventRepository` interface
- `backend/internal/repository/event_deadlines.go` — `FindDeadlineByID`
  (returns `ErrNotFound` if missing or owned by a different event);
  `CancelDeadline` (transactional: find-or-create submitter, set
  `is_active=false`, update `last_updated_by_id`, write audit rows); extracted
  `buildLastUpdatedByIDAuditLog` (shared by `AddDeadlines` and
  `CancelDeadline`) and `buildCancelDeadlineAuditLogs`
- `backend/internal/repository/event_deadlines_test.go` — 10 new integration
  tests (constraint check, `FindDeadlineByID` x3, `CancelDeadline` x6) against
  a real test Postgres, including `addOneDeadline` setup helper
- `backend/internal/repository/mocks/mock_event.go` — regenerated for the two
  new methods
- `backend/internal/handler/event_deadlines.go` — `CancelDeadline` handler,
  `cancelDeadlineRequest`; extracted `findApprovedEvent` helper (shared by
  `AddDeadlines` and `CancelDeadline`) for the "fetch event, 404 if missing,
  409 if not approved" sequence
- `backend/internal/handler/event_deadlines_test.go` — 12 new handler tests
  covering every response code and border case
- `backend/cmd/api/server.go` — registered
  `PATCH /api/v1/events/{eventId}/deadlines/{deadlineId}/cancel` on
  `publicRateLimiter`

---

## Key design decisions

### "Cancelled" vs "superseded" share `is_active=false`, distinguished by `superseded_by_id`
Cancelling never sets `superseded_by_id` — leaving it `NULL` is what means "no
replacement." `ValidateDeadlineCancellable` treats both as "already inactive"
with the same `DEADLINE_ALREADY_INACTIVE` error, since the cancel endpoint has
no use for distinguishing them and the spec explicitly does not require it.

### `validateSubmitterInput` extracted as a shared pure function
All three endpoints that accept a `submitter: {name, email}` body
(`events/submit`, `events/{id}/deadlines`, and this cancel endpoint) now share
one validation function in `internal/service/event.go`, removing what had
become triplicated inline checks.

### `findApprovedEvent` extracted in the handler layer
`AddDeadlines` and `CancelDeadline` both start with "fetch the event by ID,
404 if missing, 409 if not approved" — identical down to the error codes.
Extracted into one handler-level helper that writes the response itself and
returns `(model.Event, bool)`.

### `eventId`/`deadlineId` parsed in two stages
`:eventId` is parsed and the event is fetched/validated first; `:deadlineId`
is parsed only afterward. This means a non-numeric `:deadlineId` still
exercises (and requires a mock for) the event lookup — reflecting that the
handler always validates the event before looking at the deadline.

### `CancelDeadlineInput` carries only `Submitter`
`EventID`/`DeadlineID` come from path params and are validated by the
repository lookups (`FindByID`, `FindDeadlineByID`), not by
`ValidateCancelDeadlineInput` — so the input struct only holds the one field
that actually needs request-body validation.

---

## FP concepts introduced/reinforced (`internal/service/`)

- **Pure function** (reinforced) — `ValidateCancelDeadlineInput`: given the
  same `CancelDeadlineInput`, always returns the same error (or `nil`), with
  no I/O — event/deadline existence and the deadline's active state are
  checked later, in the repository layer.
- **Pure function** (reinforced) — `ValidateDeadlineCancellable`: given the
  same `model.Deadline`, always returns the same error (or `nil`) based only
  on `IsActive` — does not re-fetch or re-check anything.
- **Pure function** (reinforced) — `validateSubmitterInput`: same input always
  returns the same result; extracting it removed duplicated validation logic
  across three endpoints without changing any endpoint's behavior.

---

## Test count at end of session

| Package | Tests |
|---|---|
| `cmd/api` | 9 |
| `internal/config` | 6 |
| `internal/handler` | 75 (+12) |
| `internal/health` | 7 |
| `internal/middleware` | 18 |
| `internal/repository` | 65 (+10) |
| `internal/service` | 92 (+8) |
| **Total** | **272** |

All passing. `go vet ./...` and `gofmt -l .` (for files in this diff) are clean.

---

## State at end of session

`PATCH /api/v1/events/{eventId}/deadlines/{deadlineId}/cancel` is fully
implemented, tested, and wired into the router. To verify:

```bash
docker compose up -d
make migrate-up
make migrate-test-up
go test ./...               # all 272 tests green
go vet ./...                # zero warnings
cd backend && go run ./cmd/api
bash specs/backend/events-deadlines-cancel.curl.sh
```

---

## Context to restore

- `publicRateLimiter` (50/min) is now shared by `/events/submit`,
  `/events/{id}/deadlines`, and
  `/events/{eventId}/deadlines/{deadlineId}/cancel`.
- `findApprovedEvent` in `internal/handler/event_deadlines.go` is the shared
  "fetch + 404/409" helper — reuse it for any future endpoint that requires an
  approved event (e.g. the deadline update/supersession endpoint).
- `validateSubmitterInput` in `internal/service/event.go` is the shared
  submitter-body validator — reuse it for any future endpoint accepting
  `submitter: {name, email}`.
- "Cancelled" (`is_active=false`, `superseded_by_id=NULL`) and "superseded"
  (`is_active=false`, `superseded_by_id` set) are now both real states in the
  data — any future endpoint reading deadline history must handle both.
- Next likely feature: the deadline **update/supersession** endpoint — marks
  an existing `Deadline` `is_active=false`, sets `superseded_by_id`, inserts
  the replacement row, writes a `deadline_superseded` audit row (per
  `events-deadlines-add.yaml`'s `known_limitations`).
- Other pending features from prior sessions remain open: `GET
  /api/v1/events/:id`, `GET /api/v1/events/:id/audit`, and the admin review
  queue (`GET /api/v1/admin/events?status=pending`, `PATCH
  /api/v1/admin/events/:id/review`).
