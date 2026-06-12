# Session: Event Submission Feature (`POST /api/v1/events/submit`)

**Date:** 2026-06-12
**Duration:** Full session

---

## Goal

Implement the public event submission feature for the backend, following the full Pair Programming Workflow (Phase 0 Spec → ... → Phase 6 Documentation) and strict Red-Green-Refactor TDD:

- `POST /api/v1/events/submit` — public, no auth, rate-limited
- Contributor lookup/creation by email (no password)
- Slug uniqueness scoped to `pending`/`approved` events only (rejected events free their slug)
- Optional deadlines array on submission
- AuditLog rows for the event and each deadline created
- FP-annotated service layer (`// FP: <concept>` comments)

---

## What was built

### New files
- `specs/backend/events-submit.yaml` — full spec (request/response shapes, rules, border cases, DoD)
- `specs/backend/events-submit.curl.sh` — curl examples for every response code (201, 400, 409)
- `backend/migrations/003_create_events.sql` — `events` table, partial unique index on `slug` (`WHERE deleted_at IS NULL AND status IN ('pending','approved')`)
- `backend/migrations/004_create_deadlines.sql` — `deadlines` table, `type` CHECK constraint, FKs to `events`/`users`/self (`superseded_by_id`)
- `backend/internal/model/event.go` — `Event`, `EventStatus` enum
- `backend/internal/model/deadline.go` — `Deadline`, `DeadlineType` enum
- `backend/internal/service/event.go` — `ValidateSubmitEventInput`, `BuildEventFromInput`, `BuildDeadlinesFromInput`, `BuildSubmitterFromInput`, `BuildSubmission` (all `// FP:`-annotated)
- `backend/internal/service/event_test.go` — 43 tests (validation border cases + builder purity/composition checks)
- `backend/internal/repository/event.go` — `EventRepository` (`FindActiveBySlug`, `Submit`)
- `backend/internal/repository/event_test.go` — 11 integration tests (slug lookup + transactional submission)
- `backend/internal/repository/mocks/mock_event.go` — generated via `mockgen`
- `backend/internal/handler/event.go` — `EventHandler.Submit`, request/response DTOs, date parsing
- `backend/internal/handler/event_test.go` — 14 tests (happy paths, validation errors, 409 duplicate/rejected-slug cases)
- `ai-sessions/2026-06-12-event-submission-feature.md` — this file

### Modified files
- `backend/cmd/api/server.go` — wired `EventRepository`, `EventHandler`, registered `POST /api/v1/events/submit` behind a new 50 req/min rate limiter

---

## Key design decisions

### Partial unique index for slug reuse
`events_slug_idx` is a partial unique index: `UNIQUE (slug) WHERE deleted_at IS NULL AND status IN ('pending','approved')`. This lets a rejected event keep its original `slug` value in the row (for history/audit) while allowing a new submission to reuse that same slug — enforced at the DB level, not just in application code.

### AuditLog construction stays in the repository, not the service
The service layer is pure (no DB access), but an `AuditLog` row needs the **post-insert** `event.ID` / `deadline.ID`. Rather than forcing the pure service to guess IDs or do a two-phase build, `buildSubmissionAuditLogs` (a plain, non-FP-annotated helper) lives in `internal/repository/event.go` and runs after `tx.Create(&event)` inside the same transaction.

### Contributor lookup/creation/rename
`findOrCreateSubmitter` (repository layer): if a `User` with the submitted email exists (any role), its `name` is overwritten with the new submitter name and its existing role is preserved; otherwise a new `User{Role: contributor, PasswordHash: nil}` is created. Verified by `TestEventRepository_Submit_ReusesAndRenamesExistingUser` that an existing `admin` user's role is untouched.

### FP layer (`internal/service/event.go`)
- `ValidateSubmitEventInput` — **pure function**: same input → same error/nil, no I/O, trivially testable.
- `BuildEventFromInput`, `BuildDeadlinesFromInput` — **immutability**: never mutate the input struct/slice; always return new values.
- `BuildSubmitterFromInput` — **pure function**: maps `SubmitterInput` → `model.User` with fixed defaults (`role=contributor`, `password_hash=nil`).
- `BuildSubmission` — **function composition**: combines the three builders above into one call for the handler.

### Rate limit: 50 req/min per IP
New `middleware.NewRateLimiter(50.0/60.0, 50)` instance, separate from the existing 10/min auth limiter — per `events-submit.yaml` rule.

---

## Test count at end of session

| Package | Tests |
|---|---|
| `cmd/api` | 7 |
| `internal/config` | 6 |
| `internal/handler` | 42 (+14 for events) |
| `internal/health` | 7 |
| `internal/middleware` | 18 |
| `internal/repository` | 25 (+11 for events, integration — needs `docker compose up -d`) |
| `internal/service` | 43 (+ ~28 for events) |
| **Total** | **148** |

All passing. `go vet ./...` zero warnings.

---

## State at end of session

`POST /api/v1/events/submit` is fully implemented, tested, and wired into the router. To verify:

```bash
docker compose up -d
make migrate-up
make migrate-test-up
go test ./...               # all 148 tests green
go vet ./...                # zero warnings
cd backend && go run ./cmd/api
bash specs/backend/events-submit.curl.sh
```

Frontend submission form (with Leaflet location picker) is not yet built — only the backend endpoint exists so far.

---

## Context to restore

- `events` table has a **partial unique index** on `slug` — only `pending`/`approved` rows are constrained; `rejected` rows keep their slug value but don't block reuse
- `domain` is an extensible string enum, currently only `computer_science` is allowed (`internal/service/event.go: allowedDomains`)
- Deadlines are always created with `IsActive=true` on submission — superseding logic (for admin/moderator edits) is not yet built
- New 50 req/min rate limiter (`submitRateLimiter` in `cmd/api/server.go`) is separate from the 10 req/min auth limiter
- Next likely features: admin review queue (`GET /api/v1/admin/events?status=pending`, `PATCH /api/v1/admin/events/:id/review`), and the frontend submission form with Leaflet map picker
