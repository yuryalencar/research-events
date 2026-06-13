# Session: Events List Feature (`GET /api/v1/events`)

**Date:** 2026-06-13
**Duration:** Full session

---

## Goal

Implement the public event listing feature for the backend, following the full
Pair Programming Workflow (Phase 0 Spec → ... → Phase 6 Documentation) and
strict Red-Green-Refactor TDD, across 14 cycles:

- `GET /api/v1/events` — public, no auth, filterable, paginated
- Filters: `year`, `domain`, `country`, `status`, `tier`, `first_deadline_month`, `bbox`
- Pagination (`page`/`page_size`, default 20, max 100) with `pagination=off` escape hatch
- Each event includes only `is_active=true` deadlines, plus `created_by`/`last_updated_by`
- New `tier` column on `events` (CORE-style ranking, `NOT NULL DEFAULT 'unranked'`)
- FP-annotated query validation in the service layer

---

## What was built

### New files
- `specs/backend/events-list.yaml` — full spec (query params, response shape, rules, border cases, DoD)
- `specs/backend/events-list.curl.sh` — curl examples for every 200/400 case + 429 note
- `backend/migrations/006_add_tier_to_events.sql` — adds `tier VARCHAR NOT NULL DEFAULT 'unranked'` with `CHECK (tier IN ('A*','A','B','C','unranked'))`
- `backend/internal/service/event_list.go` — `RawListEventsQuery`, `ListEventsInput`, `BBoxInput`, `ValidateListEventsQuery` (+ private `parseBBox`, `parsePagination`), all `// FP: pure function`
- `backend/internal/service/event_list_test.go` — 24 tests covering every default/validation rule
- `backend/internal/repository/event_test.go` — +17 `TestEventRepository_ListEvents_*` integration tests
- `backend/internal/handler/event_list_test.go` — 6 handler tests (defaults, validation, query parsing, pagination=off, field mapping, repo error)

### Modified files
- `backend/internal/model/event.go` — added `Tier string` field (`gorm:"not null;default:unranked;index"`)
- `backend/internal/repository/event.go` — added `EventRepository.ListEvents`, `ListEventsFilter`, `BBoxFilter`, `applyListEventsFilters` (shared WHERE-clause builder for count + select queries)
- `backend/internal/repository/mocks/mock_event.go` — regenerated (`go generate ./...`) to add `ListEvents`
- `backend/internal/handler/event.go` — added `List` handler, `toRawListEventsQuery`, `toListEventsFilter`, `eventListItemResponse`, `toEventListItemResponse`, `listMeta`; refactored deadline-mapping into shared `toDeadlineResponses` (used by both `toEventResponse` and `toEventListItemResponse`)
- `backend/internal/handler/auth.go` — added `writeSuccessWithMeta` alongside `writeSuccess`
- `backend/cmd/api/server.go` — registered `GET /api/v1/events` behind a new 120 req/min (burst 30) rate limiter — the highest limit in the app, since this is the primary globe/list endpoint
- `backend/cmd/api/server_test.go` — +2 tests: route registration and 429-after-burst-30

---

## Key design decisions

### `ListEventsFilter`/`BBoxFilter` live in the repository package
Mirrors the existing repository-template pattern (`BBoxFilter` was already named in CLAUDE.md's repo template). The handler maps `service.ListEventsInput` → `repository.ListEventsFilter` via `toListEventsFilter`, the same pattern `Submit` already uses to go from service output to repository calls directly.

### Separate response type for list items
`eventListItemResponse` is distinct from `eventResponse` (used by `Submit`) because the list spec adds `year`, `last_updated_by`, and `updated_at` — fields the submit response doesn't expose. Shared deadline-mapping was extracted into `toDeadlineResponses` to avoid duplication between the two.

### Pagination defaults must always be present
GORM's `Limit(0)` returns **zero rows**, not "no limit". `ListEventsFilter.Page`/`PageSize` must always be populated (1/20 by default) even for `pagination=off` calls — `PaginationOff` is the actual switch that skips `Limit`/`Offset`.

### `first_deadline_month` via correlated EXISTS+MIN subquery
Filters events whose **earliest** `is_active=true` deadline of type `abstract`/`paper` falls in the given month *and* year — implemented as a single correlated subquery in `applyListEventsFilters` rather than a join, keeping the count query and select query symmetric.

### GORM gotcha: zero-value `bool` + `gorm:"default"` tag
A `Deadline{IsActive: false}` on `Create` is silently converted to the column default (`true`) by GORM. Test data that needs an inactive deadline must be created `IsActive: true` then superseded via `tx.Model(&d).Update("is_active", false)` — which is also exactly how the application supersedes deadlines per the "deadlines are immutable" architecture rule.

### Rate limit: 120 req/min, burst 30
New `middleware.NewRateLimiter(120.0/60.0, 30)` — the highest limit in the app, since `GET /api/v1/events` backs both the globe and list views (read-heavy, public).

---

## FP concepts introduced/reinforced (`internal/service/event_list.go`)

- **Pure function** — `ValidateListEventsQuery`, `parseBBox`, `parsePagination`: given the same `RawListEventsQuery` + `currentYear`, always return the same `ListEventsInput`/error, with no I/O. `currentYear` is passed in (derived from `time.Now()` by the caller) specifically so the function itself never touches the clock — this is what makes `TestValidateListEventsQuery_SameInputTwice_ReturnsIdenticalResult` meaningful.

---

## Test count at end of session

| Package | Tests |
|---|---|
| `cmd/api` | 9 (+2) |
| `internal/config` | 6 |
| `internal/handler` | 51 (+6) |
| `internal/health` | 7 |
| `internal/middleware` | 18 |
| `internal/repository` | 43 (+17) |
| `internal/service` | 73 (+24) |
| **Total** | **207** |

All passing. `go vet ./...` zero warnings.

---

## State at end of session

`GET /api/v1/events` is fully implemented, tested, and wired into the router. To verify:

```bash
docker compose up -d
make migrate-up
make migrate-test-up
go test ./...               # all 207 tests green
go vet ./...                # zero warnings
cd backend && go run ./cmd/api
bash specs/backend/events-list.curl.sh
```

---

## Context to restore

- `tier` is a closed enum at the DB level (`CHECK` constraint) but an open `string`
  in Go — never `nil`, defaults to `"unranked"`.
- `ListEventsFilter.Page`/`PageSize` must always be set (defaults 1/20), even when
  `PaginationOff=true` — `Limit(0)` in GORM returns zero rows, not "all rows".
- `applyListEventsFilters` returns a fresh `*gorm.DB` per call — required so the
  `Count` query and the `Find` query (with `Preload`/`Order`/`Limit`) don't share
  mutated query state.
- `GET /api/v1/events` now has the highest per-IP rate limit in the app
  (120/min, burst 30) — `listRateLimiter` in `cmd/api/server.go`.
- Next likely features: `GET /api/v1/events/:id`, `GET /api/v1/events/:id/deadlines`,
  `GET /api/v1/events/:id/audit`, and the admin review queue
  (`GET /api/v1/admin/events?status=pending`, `PATCH /api/v1/admin/events/:id/review`).
