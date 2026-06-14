# Session: Deadlines — Add `time` + `timezone` Fields (cross-cutting)

**Date:** 2026-06-14
**Duration:** Full session

---

## Goal

Add optional time-of-day precision to deadlines: a `time` (HH:MM, 24h) and a
`timezone` (free string, e.g. "AoE", "UTC-3") field on `Deadline`, both
nullable. This is a prerequisite for the upcoming deadline-supersede endpoint
but applies to every endpoint that accepts or returns deadlines:

- `POST /api/v1/events/submit` — `deadlines[].time`/`timezone` optional in request and response
- `POST /api/v1/events/{id}/deadlines` — same
- `GET /api/v1/events` — `deadlines[].time`/`timezone` in response
- `PATCH /api/v1/events/{eventId}/deadlines/{deadlineId}/cancel` — reloaded event's `deadlines[].time`/`timezone`

`date` remains a separate date-only field — `time`/`timezone` are
independently optional, no "both or neither" rule. Pre-migration rows have
both `null`.

Followed the full Pair Programming Workflow (Phase 0 Spec → ... → Phase 6
Documentation) and strict Red-Green-Refactor TDD, across 3 cycles.

---

## What was built

### New files
- `specs/backend/deadlines-add-time-timezone.yaml` — full spec
- `specs/backend/deadlines-add-time-timezone.curl.sh` — curl examples for every border case
- `backend/migrations/009_add_time_timezone_to_deadlines.sql` — adds nullable
  `time VARCHAR(5)` and `timezone VARCHAR(50)` columns to `deadlines`

### Modified files
- `backend/internal/model/deadline.go` — `Deadline` gains `Time *string`,
  `Timezone *string`
- `backend/internal/repository/event_test.go` — 2 new integration tests
  (`TestEventRepository_Submit_PersistsDeadlineTimeAndTimezone`,
  `TestEventRepository_Submit_PersistsDeadlineWithNilTimeAndTimezone`)
- `backend/internal/service/event_deadlines.go` —
  `DeadlineInput` gains `Time *string`, `Timezone *string`;
  `deadlineTimePattern` (HH:MM, 24h, zero-padded) added as package-level
  regex; `validateDeadlineInput` (`// FP: pure function`, reinforced)
  validates `time` format and rejects an explicit empty `timezone`;
  `BuildDeadlinesFromInput` (`// FP: immutability`, reinforced) maps both
  fields through unchanged
- `backend/internal/service/event_deadlines_test.go` — 8 new tests covering
  every border case in the spec (both set, time-only, timezone-only, bad
  formats, empty timezone, immutable mapping)
- `backend/internal/handler/event.go` — `deadlineRequest` and
  `deadlineResponse` gain `Time *string`, `Timezone *string`;
  `toDeadlineResponses` maps them through
- `backend/internal/handler/event_deadlines.go` — `toDeadlineInputs` maps
  `time`/`timezone` from the request through to `service.DeadlineInput`
- `backend/internal/handler/event_test.go` — 2 new tests for `Submit`
  (round-trip + invalid time → 400)
- `backend/internal/handler/event_deadlines_test.go` — 3 new tests
  (`AddDeadlines` round-trip + empty timezone → 400; `CancelDeadline` reload
  includes time/timezone on remaining deadlines)

---

## Key design decisions

### `date` stays separate; `time`/`timezone` are independent, nullable additions
Rather than combining into a single `timestamptz`, `time` and `timezone` are
plain nullable columns. A deadline may have date-only, date+time, or
date+time+timezone. This keeps the existing `date`-based queries
(`first_deadline_month`, sorting) unchanged.

### One shared validation/regex, reused everywhere
`deadlineTimePattern` and the two new checks live in
`validateDeadlineInput`, which is already shared by `ValidateAddDeadlinesInput`
(used by both `events/submit` and `events/{id}/deadlines`). No per-endpoint
duplication was needed.

### `toDeadlineResponses`/`toDeadlineInputs` already centralized
Both mapping functions were already shared across `toEventResponse`,
`toEventListItemResponse`, and the add-deadlines path from prior sessions —
adding the two new fields there automatically covered `submit`,
`add-deadlines`, `list`, and `cancel`'s reload with one change each.

---

## FP concepts reinforced (`internal/service/`)

- **Pure function** (reinforced) — `validateDeadlineInput`: the new
  `time`/`timezone` checks depend only on the `DeadlineInput` argument, no
  I/O, same input always produces the same error (or `nil`).
- **Immutability** (reinforced) — `BuildDeadlinesFromInput`: `Time`/`Timezone`
  pointers are copied into a new `model.Deadline` value; the input slice and
  its elements are never mutated.

---

## Test count at end of session

| Package | Tests |
|---|---|
| `cmd/api` | 9 |
| `internal/config` | 6 |
| `internal/handler` | 80 (+5) |
| `internal/health` | 7 |
| `internal/middleware` | 18 |
| `internal/repository` | 67 (+2) |
| `internal/service` | 100 (+8) |
| **Total** | **287** |

All passing. `go vet ./...` clean; `gofmt -l .` shows no new issues (3
pre-existing unrelated files only).

---

## State at end of session

Migration 009 + all four affected endpoints now accept/return
`time`/`timezone` per deadline. To verify:

```bash
docker compose up -d
make migrate-up
make migrate-test-up
go test ./...               # all 287 tests green
go vet ./...                # zero warnings
cd backend && go run ./cmd/api
bash specs/backend/deadlines-add-time-timezone.curl.sh
```

---

## Context to restore

- `model.Deadline.Time *string` (HH:MM, 24h) and `Timezone *string` (free
  string) are now available on every deadline, `nil` for pre-migration rows.
- `deadlineRequest`/`deadlineResponse` in `internal/handler/event.go` and
  `service.DeadlineInput` in `internal/service/event_deadlines.go` both carry
  the new fields — any future endpoint accepting or returning a deadline
  (e.g. the supersede endpoint) should reuse these types/mappers as-is.
- `deadlineTimePattern` in `internal/service/event_deadlines.go` is the
  shared HH:MM validation regex — reuse it rather than duplicating.
- Next: write the Phase 0 spec for `POST
  /api/v1/events/{eventId}/deadlines/{deadlineId}/supersede`, now that
  `time`/`timezone` are available for the replacement deadline's request body
  and response.
- Other pending features from prior sessions remain open: `GET
  /api/v1/events/:id`, `GET /api/v1/events/:id/audit`, and the admin review
  queue (`GET /api/v1/admin/events?status=pending`, `PATCH
  /api/v1/admin/events/:id/review`).
