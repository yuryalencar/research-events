# Session: Admin/Moderator Event Review (`PATCH /api/v1/admin/events/{id}/review`)

**Date:** 2026-06-14
**Duration:** Full Pair Programming Workflow (Phase 0 Spec approved in a prior
session; this session covered Phase 1–6, across 3 cycles)

---

## Goal

Implement the admin/moderator review action that approves or rejects a
pending (or previously reviewed) event submission, optionally editing any of
the event's own fields in the same request:

- `PATCH /api/v1/admin/events/{id}/review` — `RequireAuth` +
  `RequireRole("admin", "moderator")`
- Body: `{ action: "approve" | "reject", reason?: string, event?: {...partial fields...} }`
- `action` sets `event.status` and writes one `AuditLog` row
  (`action="approved"` or `"rejected"`, `entity_type="event"`)
- `reason` is required (non-empty) for `action="reject"`, optional for
  `"approve"`; stored on the new nullable `audit_logs.reason` column
- Optional `event` object is a partial update validated with the same
  per-field rules as `events-submit.yaml`; editing `start_date` recomputes
  `year`
- Editing `slug` to collide with another `pending`/`approved` event ->
  409 `SLUG_ALREADY_EXISTS`; colliding with a `rejected` event's slug is
  allowed (slug reuse rule)
- `last_updated_by_id` always set to the reviewer, regardless of whether
  `event` fields were edited
- Re-review is always allowed — any status can transition to any other,
  including no-op transitions (status before == after), each writing a new
  `AuditLog` row
- Moderators get 403 `CANNOT_REVIEW_OWN_EVENT` when `created_by_id` is their
  own user ID; admins are exempt
- 200 response uses the existing `eventListItemResponse` shape

Plan (approved earlier): Cycle 1 = service layer, Cycle 2 = repository +
migration + model, Cycle 3 = handler + routing, Cycle 4 = refactor (folded
into the cycles above — no separate cycle was needed).

---

## What was built

### New files
- `backend/migrations/010_add_reason_to_audit_logs.sql` — adds nullable
  `audit_logs.reason TEXT` column (usable by future actions, not just this
  endpoint)
- `backend/internal/service/event_review.go` — `EventEditInput`,
  `ReviewEventInput`, `ValidateReviewActionInput` (`// FP: pure function`),
  `ApplyReview` (`// FP: immutability`), `ValidateEditedEvent`
  (`// FP: pure function`), `BuildReviewAuditLog` (`// FP: no side effects`),
  plus private helpers `applyEventEdits` (`// FP: immutability`) and
  `buildReviewDiff` (`// FP: pure function`)
- `backend/internal/service/event_review_test.go` — 33 tests
- `backend/internal/repository/event_review.go` — `Review(ctx, updated,
  auditLog)`: transactional save of the reviewed event + new `AuditLog` row,
  reloaded via `findByIDWithActiveDeadlines`
- `backend/internal/repository/event_review_test.go` — 7 integration tests
  against the real test Postgres
- `backend/internal/handler/admin_event.go` — `AdminEventHandler`, `Review`
  handler, `reviewEventRequest`/`eventEditRequest`, `toEventEditInput`
- `backend/internal/handler/admin_event_test.go` — 15 handler tests covering
  every response code and border case

### Modified files
- `backend/internal/repository/event.go` — added `Review` to
  `EventRepository`
- `backend/internal/repository/mocks/mock_event.go` — regenerated for
  `Review`
- `backend/internal/handler/event.go` — extracted shared `parseDateField(field,
  raw string) (time.Time, error)` helper (used by both
  `toSubmitEventInput` and the new `toEventEditInput`)
- `backend/cmd/api/server.go` — registered
  `PATCH /api/v1/admin/events/{id}/review` behind `RequireAuth` +
  `RequireRole("admin", "moderator")`
- `backend/.env` / `backend/.env.example` — added `TEST_DATABASE_URL` (the
  Postgres connection string `go test ./...` uses for repository tests)

---

## Key design decisions

### GORM association overwrite — `Omit(clause.Associations)`
`updated` (the event returned by `ApplyReview`) still carries the
`CreatedBy`/`LastUpdatedBy`/`Deadlines` associations populated from the
reload that produced `existing`. A plain `tx.Save(&updated)` causes GORM to
re-save those associations, which overwrites the explicitly-set
`LastUpdatedByID` foreign key with the (stale) association struct's ID.
Fixed with `tx.Omit(clause.Associations).Save(&updated)` in
`internal/repository/event_review.go`.

### Shared `parseDateField` helper
`toSubmitEventInput` (Submit) and `toEventEditInput` (Review) both needed
"parse `YYYY-MM-DD`, or return `'<field> must be a valid date (YYYY-MM-DD)'`"
— extracted into `parseDateField(field, raw string) (time.Time, error)` in
`internal/handler/event.go` so both handlers report identical error messages
and there's a single place to change the date format.

### Reusing `ValidateEditedEvent` for partial updates
Rather than a separate "partial edit" validator, `ApplyReview` merges the
optional `event` overrides onto `existing` to produce a complete `model.Event`,
then `ValidateEditedEvent` runs the exact same per-field checks as
`ValidateSubmitEventInput` — so a reviewer's edit can never produce an event
that wouldn't have passed submission validation.

---

## FP concepts introduced/reinforced (`internal/service/`)

- **Pure function** (reinforced) — `ValidateReviewActionInput` and
  `ValidateEditedEvent`: given the same input, always return the same error
  (or `nil`), no I/O.
- **Immutability** (reinforced) — `ApplyReview` and `applyEventEdits`: never
  mutate `existing`; return a brand-new `model.Event` with the reviewer's
  edits and `LastUpdatedByID` applied. The imperative alternative — mutating
  a copy of `existing` field-by-field in place — makes it easy to forget a
  field or accidentally leave a stale association populated (which is exactly
  the bug the `Omit(clause.Associations)` fix above had to work around at the
  repository layer).
- **No side effects** — `BuildReviewAuditLog`: computes the `AuditLog` row
  (including the `diff` JSONB via `buildReviewDiff`) without writing it; the
  repository's `Review` method persists it inside the same transaction as the
  event update.

---

## Test count at end of session

| Package | Tests |
|---|---|
| `cmd/api` | 9 |
| `internal/config` | 6 |
| `internal/handler` | 111 |
| `internal/health` | 7 |
| `internal/middleware` | 18 |
| `internal/repository` | 81 |
| `internal/service` | 145 |
| **Total** | **377** |

All passing (`go test ./...`). `go vet ./...` and `gofmt -l .` clean (only the
3 pre-existing unrelated files flagged, unchanged by this diff).

---

## State at end of session

`PATCH /api/v1/admin/events/{id}/review` is fully implemented, tested, and
wired into the router. To verify:

```bash
docker compose up -d
make migrate-up
make migrate-test-up
go test ./...               # all 377 tests green
go vet ./...                # zero warnings
cd backend && go run ./cmd/api
bash specs/backend/admin-events-review.curl.sh
```

---

## Context to restore

- `audit_logs.reason` (nullable `TEXT`, migration 010) is now available for
  any future audit-writing action, not just review.
- `parseDateField` in `internal/handler/event.go` is the shared date-parsing
  helper — any future handler that accepts `YYYY-MM-DD` fields should use it
  instead of calling `time.Parse` directly.
- `internal/repository/event_review.go`'s `Omit(clause.Associations)` pattern
  is the template for any future repository method that `Save()`s a
  `model.Event` carrying populated associations.
- Other pending features from prior sessions remain open: `GET
  /api/v1/events/:id`, `GET /api/v1/events/:id/audit`, and
  `GET /api/v1/admin/events?status=pending`.
