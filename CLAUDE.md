# ReSEARCH Events

A collaborative, open-source platform that aggregates research conferences and events in software engineering and computer science. Researchers can discover events on an interactive 3D globe, filter by year/location, and submit new events for admin review.

> **Learning project** — prioritize readable, idiomatic code. Explain WHY patterns are used, mention trade-offs when multiple approaches exist, and add short comments when introducing new Go concepts (goroutines, interfaces, channels).
> 
> **Functional programming goal** — the service layer follows functional programming principles (pure functions, immutability, no side effects). Every function written in `internal/service/` must have a `// FP: <concept>` comment above it naming the principle applied, and Claude must explain that concept the first time it appears in a session.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Frontend | Next.js 15 (App Router), React 19, TypeScript 5, Tailwind CSS v4 |
| Globe / Map | Globe.gl (3D WebGL) + fallback 2D Leaflet map view |
| Location picker | Leaflet embedded in the submission form — submitter drops a pin to set lat/lng directly; no geocoding API needed |
| i18n | next-intl (App Router native, SSR-compatible) |
| Backend | Go 1.22+, `net/http` stdlib (no HTTP framework) |
| ORM | GORM v2 + Goose (migrations) |
| Database | PostgreSQL 16 |
| Auth | JWT (stateless) — admin/moderator only, stored in HTTP-only cookies |
| Deploy | Vercel (frontend) + Fly.io (backend) + managed Postgres |
| Monitoring | Sentry (error tracking + performance) — frontend and backend |
| Tracing | OpenTelemetry SDK for Go → exports traces to Sentry |
| Health check | Custom `GET /health` endpoint (built in Go — learning exercise) |
| Dev tools | pnpm, Docker Compose (local Postgres), Air (Go hot reload) |
| Go testing | `testing` stdlib + testify (assert/require) + gomock (generated mocks) |
| Frontend testing | Vitest (unit tests for logic and hooks only — no component rendering) |

---

## Project Structure

```
/
├── frontend/
│   ├── src/
│   │   ├── app/
│   │   │   └── [locale]/                  # next-intl locale segment (e.g. /en/, /pt/)
│   │   │       ├── layout.tsx
│   │   │       ├── page.tsx               # Globe homepage
│   │   │       ├── events/
│   │   │       │   ├── [id]/page.tsx      # Event detail (deadlines + audit trail)
│   │   │       │   └── submit/page.tsx    # Public submission form (with map picker)
│   │   │       └── admin/
│   │   │           └── review/page.tsx    # Admin review queue
│   │   ├── components/
│   │   │   ├── ui/                        # Primitives: Button, Input, Badge, Modal
│   │   │   ├── globe/                     # Globe.gl wrapper + pin logic
│   │   │   ├── map/                       # Leaflet location picker (submission form only)
│   │   │   ├── events/                    # EventCard, EventFilters, DeadlineList, AuditTrail
│   │   │   └── admin/                     # ReviewCard, ApproveButton
│   │   ├── lib/
│   │   │   ├── api.ts                     # Typed fetch client for Go backend
│   │   │   ├── constants.ts
│   │   │   └── utils.ts
│   │   ├── hooks/                         # useEvents, useGlobeState, useFilters
│   │   ├── types/
│   │   │   └── api.ts                     # Types generated from backend OpenAPI spec
│   │   └── messages/                      # i18n translation files
│   │       ├── en.json                    # English (source of truth)
│   │       ├── pt.json                    # Portuguese
│   │       ├── es.json                    # Spanish
│   │       └── de.json                    # German
│   ├── middleware.ts                       # next-intl locale detection middleware
│   ├── i18n.ts                            # next-intl config
│   ├── tailwind.config.ts
│   └── package.json
│
├── backend/
│   ├── cmd/
│   │   └── api/
│   │       └── main.go                    # Entry point — wires config, DB, routes
│   ├── internal/
│   │   ├── handler/                       # HTTP handlers (one file per resource)
│   │   │   └── health.go                  # GET /health — liveness + dependency checks
│   │   ├── service/                       # Business logic layer
│   │   ├── repository/                    # GORM queries (one file per model)
│   │   ├── model/                         # Domain structs + GORM tags
│   │   │   ├── event.go                   # Event, EventStatus
│   │   │   ├── deadline.go                # Deadline + DeadlineHistory
│   │   │   ├── user.go                    # User (admin | moderator | contributor)
│   │   │   └── audit_log.go               # AuditLog (full change history)
│   │   ├── middleware/                    # JWT auth, CORS, request logging, rate-limit
│   │   ├── health/                        # Health checker — extensible checker interface
│   │   │   ├── health.go                  # HealthChecker, Check interface, response types
│   │   │   ├── db.go                      # DatabaseChecker (pings Postgres)
│   │   │   └── checker.go                 # Registry — add new checkers without changing handler
│   │   ├── observability/                 # OpenTelemetry + Sentry bootstrap
│   │   │   ├── otel.go                    # OTel tracer provider setup
│   │   │   └── sentry.go                  # Sentry SDK init
│   │   └── config/                        # Env parsing via godotenv
│   ├── migrations/                        # Goose SQL files (sequential: 001_, 002_...)
│   ├── go.mod
│   └── go.sum
│
├── ai-sessions/                           # AI session summaries for context recovery
│   └── YYYY-MM-DD-session-title.md        # One file per session
│
├── specs/                                 # Spec files — one per feature, written before any code
│   ├── backend/                           # YAML specs + curl files per endpoint
│   │   ├── auth-login.yaml                # Spec file
│   │   └── auth-login.curl.sh             # curl examples — generated after spec approval
│   └── frontend/                          # Markdown specs for frontend features
├── README.md                              # Index: all specs + sessions (one-line descriptions)
├── docker-compose.yml                     # Local Postgres only
├── Makefile                               # Unified commands for both services
└── CLAUDE.md
```

---

## Commands

```bash
# Frontend
cd frontend && pnpm dev             # Dev server → http://localhost:3000
cd frontend && pnpm build           # Production build
cd frontend && pnpm typecheck       # tsc --noEmit (run before every commit)
cd frontend && pnpm lint            # ESLint

# Backend
cd backend && air                   # Hot reload dev server → :8080
cd backend && go run ./cmd/api      # Run without hot reload
cd backend && go test ./...         # All tests
cd backend && go vet ./...          # Static analysis (must pass, zero warnings)

# Infrastructure
docker compose up -d                # Start local Postgres on :5432
make migrate-up                     # Run pending Goose migrations
make migrate-down                   # Roll back last migration
make generate-types                 # Regenerate frontend/src/types/api.ts from OpenAPI spec
make generate-mocks                 # Regenerate all gomock files from interfaces
```

---

## Data Model (Core)

```
User
  id, name, email, password_hash
  role  (admin | moderator | contributor)
         ← contributor: no password yet; created automatically on first submission
         ← password-less contributors can claim their account later (future feature)
  created_at

Event
  id, name, slug
  country, city
  latitude, longitude       ← set by submitter via Leaflet map picker; never geocoded automatically
  start_date, end_date
  website_url
  domain                    ← extensible string enum (software_engineering | computer_science | ...)
                               platform will expand to other fields (medicine, etc.) in the future
  status                    ← pending | approved | rejected
  year                      ← indexed; default filter = current year
  created_by_id  (FK → User)       ← first submitter; shown as "Added by <name> on <date>"
  last_updated_by_id (FK → User)   ← most recent editor; shown as "Updated by <name> on <date>"
  created_at, updated_at

Deadline (belongs to Event)
  id, event_id
  type         ← abstract | paper | notification | camera_ready | other
  description  ← free text, e.g. "Research track", "Industry innovation track"
  date
  is_optional
  is_active    ← false when superseded by a newer deadline of the same type
  superseded_by_id (FK → Deadline, nullable)  ← points to the replacement deadline
  created_by_id (FK → User)
  created_at
  ← Never update a deadline in place. Always create a new Deadline record,
    set the old one's is_active=false and superseded_by_id=<new id>.
  ← UI: show only is_active=true deadlines by default;
    "view history" toggle reveals the full chain per type.

AuditLog
  id
  entity_type  ← "event" | "deadline"
  entity_id
  action       ← "created" | "updated" | "approved" | "rejected" | "deadline_added" | "deadline_superseded"
  changed_by_id (FK → User)
  diff         ← JSONB — stores before/after values of changed fields
  created_at
  ← Every state change to Event or Deadline must write an AuditLog row.
    This is the source of truth for the full history shown in the UI.
```

---

## Contributor Flow (Event Submission)

1. Public user fills out submission form — provides **name**, **email**, event details, and drops a pin on the Leaflet map to set lat/lng.
2. Backend checks if email exists in `users` table:
   - **Exists** → link `event.created_by_id` to that user (no new user created).
   - **Does not exist** → create a new `User` with `role=contributor`, `name`, `email`, and `password_hash=NULL`. Link event to them.
3. Event is created with `status=pending`. An `AuditLog` row is written (`action=created`).
4. Admin reviews and approves/rejects. Each action writes another `AuditLog` row.
5. Any subsequent update (e.g. deadline change) by any user writes to `AuditLog` and updates `event.last_updated_by_id`.
6. Future: contributor can set a password to claim their account and log in.

---

## API Design

- Base path: `/api/v1/`
- Response envelope (matches `internal/handler/auth.go` `writeSuccess`/`writeSuccessWithMeta`/`writeError`):
  - Success: `{ "code": "SOME_CODE", "data": T }`
  - Success (list endpoints): `{ "code": "SOME_CODE", "data": T[], "meta": { "page": N, "total": N } }`
  - Error: `{ "code": "EVENT_NOT_FOUND", "error": { "message": "..." } }`
- Key endpoints:
  - `GET    /api/v1/events` — filterable by `year`, `domain`, `country`, `bbox`
  - `GET    /api/v1/events/:id` — includes active deadlines + contributor attribution
  - `GET    /api/v1/events/:id/deadlines` — all deadlines grouped by type (active + history)
  - `GET    /api/v1/events/:id/audit` — full audit log for an event
  - `POST   /api/v1/events/submit` — public, no auth; triggers contributor lookup/creation
  - `GET    /api/v1/admin/events?status=pending` — admin only
  - `PATCH  /api/v1/admin/events/:id/review` — approve or reject; writes AuditLog
  - `POST   /api/v1/events/:id/deadlines` — add/update a deadline (supersedes existing)
  - `POST   /api/v1/auth/login` — returns JWT in HTTP-only cookie
- `bbox` param format: `?bbox=minLng,minLat,maxLng,maxLat` — filters events to globe viewport
  - `GET    /health` — public, no auth; returns detailed system status (see Health Check section)

---

## Code Style

### TypeScript
- **Strict mode always** — `tsconfig.json` must have `"strict": true`; never disable it
- No `any` — use `unknown` and narrow with type guards; `@ts-ignore` is banned
- Prefer `interface` for object shapes; `type` for unions/intersections
- Always type function return values explicitly
- Use `satisfies` operator to validate objects against a type without widening
- Enums are banned — use `as const` objects: `const STATUS = { PENDING: 'pending' } as const`
- Never hand-write API types — regenerate `src/types/api.ts` from the OpenAPI spec (`make generate-types`)

### React / Next.js
- ES modules only (`import/export`), never CommonJS
- Destructure imports: `import { useState } from "react"`
- React Server Components by default; add `"use client"` only for interactivity or browser APIs
- Globe.gl and Leaflet must both be loaded with `dynamic(() => import(...), { ssr: false })` — both require browser APIs
- Files: `.tsx` for components, `.ts` for logic/utilities
- Error boundaries on every route segment (`error.tsx` file)

### i18n (next-intl)
- All user-facing strings go through `useTranslations()` (client) or `getTranslations()` (server) — no hardcoded UI strings
- Translation keys use dot notation: `"events.card.deadlineLabel"`
- Supported locales: `en` (English), `pt` (Portuguese), `es` (Spanish), `de` (German)
- `en.json` is the source of truth; all other locales must have every matching key — no missing keys allowed
- Locale segment: `app/[locale]/` — all routes are nested inside this

### Go
- Follow standard Go conventions: `gofmt`, `go vet`, zero warnings
- Errors are values — always handle them, never discard with `_`
- Return `(result, error)` tuples; check the error immediately after the call
- Use `context.Context` as first parameter in every function that does I/O
- Keep `main.go` minimal: parse config → open DB → register routes → start server
- No `init()` functions — always use explicit initialization
- Interfaces belong in the consumer package, not the provider
- Structured logging with `slog` (not `fmt.Println` or `log.Printf`)
- GORM: always use `.WithContext(ctx)`; never `.Find()` without a `WHERE` clause on large tables

### Functional Programming in Go (service layer)

Go is not a functional language, but functional principles apply cleanly to the `internal/service/` layer where business logic lives. This is a deliberate learning goal — the handler and repository layers follow standard Go idioms; the service layer follows FP discipline.

**Scope:** FP rules apply strictly to `internal/service/`. Handler and repository layers follow standard Go idioms.

---

**Core principles enforced in `internal/service/`:**

**1. Pure functions**
A function is pure if: given the same inputs it always returns the same output, and it has no side effects (no mutation of external state, no I/O, no globals). Pure functions are trivially testable — no mocks needed, just input → output.
```go
// FP: pure function
// A pure function depends only on its arguments and has no observable side effects.
// This means: no DB calls, no HTTP, no global state — just data in, data out.
// Pure functions are the easiest code to test and reason about.
func applyEventDefaults(e Event, year int) Event {
    if e.Year == 0 {
        e.Year = year
    }
    return e // returns a new value, never mutates the input
}
```

**2. Immutability**
Never mutate a struct passed as a parameter. Always return a new copy with the changes applied. In Go, structs are value types — returning a modified copy is idiomatic and safe.
```go
// FP: immutability
// Instead of modifying the event we received, we construct and return a new one.
// The caller's original value is untouched. This prevents a whole class of bugs
// where shared state is modified unexpectedly across function boundaries.
func approveEvent(e Event, byUser User, at time.Time) Event {
    return Event{
        ID:              e.ID,
        Status:          StatusApproved,
        LastUpdatedByID: byUser.ID,
        UpdatedAt:       at,
    }
}
```

**3. Separation of computation from side effects**
Side effects (DB writes, HTTP calls, logging) belong in the handler or repository layer — never inside a service function that computes a result. Service functions receive data, compute a result, and return it. The caller decides what to do with it.
```go
// FP: no side effects
// This function only computes the audit entry — it does not write it to the DB.
// The handler receives this value and decides when and how to persist it.
// Separating computation from I/O makes each piece independently testable.
func buildAuditEntry(entity string, id uint, action string, diff map[string]any, by User) AuditLog {
    return AuditLog{EntityType: entity, EntityID: id, Action: action, Diff: diff, ChangedByID: by.ID}
}
```

---

**`// FP: <concept>` comment rule — mandatory for every function in `internal/service/`**

Every function must have an `// FP:` comment directly above its signature. Valid tags:
- `// FP: pure function`
- `// FP: immutability`
- `// FP: no side effects`
- `// FP: function composition`
- `// FP: first-class function`

**Teaching rule — applies every time, not just the first:**
After every `// FP:` comment, Claude must add a 1–2 line prose explanation directly in the code comment explaining:
1. What this concept means in plain language
2. Why it matters / what problem it prevents

The explanation lives in the code, not in chat — so it's always visible when reading the file later.

**Violation examples — Claude must never write these in `internal/service/`:**
```go
// BAD: service method that writes to DB directly (side effect)
func (s *EventService) Submit(ctx context.Context, e Event) error {
    s.db.Create(&e) // ← side effect inside service — move this to repository
    return nil
}

// BAD: mutating the input struct
func enrichEvent(e *Event, year int) {
    e.Year = year // ← mutation — return a new Event instead
}
```

---

## Monitoring & Observability

### Sentry
- Initialised on both frontend (`instrumentation.ts`) and backend (`internal/observability/sentry.go`) at startup
- **Frontend:** captures unhandled errors, React render errors, and slow page loads via `@sentry/nextjs`
- **Backend:** captures panics, unhandled errors, and slow requests via `sentry-go`; integrated as middleware so every request gets a Sentry transaction
- `tracesSampleRate`: `1.0` in development, `0.2` in production — never set to `0`
- Never log sensitive data (passwords, JWT tokens, emails) to Sentry — scrub them in `BeforeSend`
- Every captured error must include: environment, release version, and request context

### OpenTelemetry (Go backend)
- Instrumented at the handler level via OTel HTTP middleware — every request gets a trace span automatically
- Manual spans for: database queries, external HTTP calls, and any operation over 100ms
- Span naming convention: `resource.action` — e.g. `event.submit`, `deadline.supersede`, `db.query`
- Traces exported to Sentry via OTLP exporter (`internal/observability/otel.go`)
- Always pass `context.Context` through the call stack so spans propagate correctly — never start a new root span mid-request
- Required span attributes: `http.method`, `http.route`, `db.operation` (for DB spans), `error` (bool)

### Health Check

`GET /health` — public endpoint, no authentication required.

**Response shape:**
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "timestamp": "2026-01-15T10:00:00Z",
  "uptime": "3h22m10s",
  "checks": {
    "database": {
      "status": "healthy",
      "latency_ms": 4
    },
    "future_dependency": {
      "status": "unhealthy",
      "error": "connection refused"
    }
  }
}
```

**Status rules:**
- Top-level `status` is `"healthy"` only when ALL checks pass — any single failure makes it `"unhealthy"`
- HTTP response code: `200` when healthy, `503` when unhealthy
- Database check: runs `SELECT 1` with a 3-second timeout via `context.WithTimeout`

**Architecture — extensible checker pattern:**
```go
// internal/health/health.go
type Checker interface {
    Name() string
    Check(ctx context.Context) CheckResult
}

type CheckResult struct {
    Status    string `json:"status"`           // "healthy" | "unhealthy"
    LatencyMs int64  `json:"latency_ms,omitempty"`
    Error     string `json:"error,omitempty"`
}
```
New dependencies (cache, external API, etc.) implement `Checker` and register themselves — the handler never changes.

**Test requirements for health check:**
- `TestHealth_AllChecksPass` → returns 200 + `status: healthy`
- `TestHealth_DatabaseUnhealthy` → returns 503 + `status: unhealthy` + db error populated
- `TestHealth_NewCheckerUnhealthy` → adding a failing checker propagates to top-level status
- Mock all checkers with gomock — never hit a real DB in handler tests

---

## Architecture Rules

- Frontend talks to backend only via REST — no direct DB access from Next.js
- Backend validates all input — never trust the frontend
- JWT stored in HTTP-only cookie only — never in localStorage or a JS-readable cookie
- Admin routes (`/admin/*`) protected by JWT middleware on both frontend (redirect) and backend (401)
- Event submissions always created with `status=pending` — never auto-approved
- **Deadlines are immutable once created** — never UPDATE a deadline row; always INSERT a new one and mark the old as `is_active=false`
- **Every state change writes an AuditLog row** — this is non-negotiable; do not skip it even in tests
- The `domain` field is an extensible string enum — never hardcode domain-specific logic outside of validation
- Globe viewport filtering uses bounding box queries on `latitude`/`longitude` columns (ensure they are indexed)

---

## File Patterns

Every file type in this project follows a strict internal structure. Claude must generate all files conforming to these templates — no exceptions. Consistency makes the codebase readable by anyone familiar with the project conventions.

---

### Go — Handler (`internal/handler/*.go`)

```go
package handler

import (
    // stdlib first, then third-party, then internal — separated by blank lines
    "context"
    "encoding/json"
    "net/http"

    "go.uber.org/zap"

    "github.com/yourname/research-events/internal/service"
    "github.com/yourname/research-events/internal/model"
)

// --- Types ---

// EventHandler holds dependencies for event-related HTTP handlers.
// Dependencies are always injected via constructor — never use globals.
type EventHandler struct {
    eventService service.EventService
    logger       *slog.Logger
}

// --- Constructor ---

// NewEventHandler creates a new EventHandler with its required dependencies.
func NewEventHandler(eventService service.EventService, logger *slog.Logger) *EventHandler {
    return &EventHandler{
        eventService: eventService,
        logger:       logger,
    }
}

// --- Public methods (one per route) ---

// Submit handles POST /api/v1/events/submit.
// Public endpoint — no authentication required.
func (h *EventHandler) Submit(w http.ResponseWriter, r *http.Request) {
    // 1. Parse and validate input
    // 2. Call service
    // 3. Write response
}

// --- Private functions (helpers used only in this file) ---

func (h *EventHandler) writeJSON(w http.ResponseWriter, status int, data any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}
```

---

### Go — Service (`internal/service/*.go`)

Interface lives at the top of the file, implementation below. No separate `_interface.go` file.

```go
package service

import (
    "context"

    "github.com/yourname/research-events/internal/model"
    "github.com/yourname/research-events/internal/repository"
)

// --- Interface ---

// EventService defines the business logic contract for events.
// Interfaces belong in the consumer package — here, service defines its own contract
// so the handler layer depends on the interface, not the concrete type.
type EventService interface {
    Submit(ctx context.Context, input SubmitEventInput) (model.Event, error)
    Approve(ctx context.Context, id uint, byUser model.User) (model.Event, error)
}

// --- Types ---

type eventService struct {
    repo   repository.EventRepository
    audits repository.AuditRepository
}

// SubmitEventInput groups all inputs for the Submit operation.
// Input structs are always defined in the service layer, not the handler.
type SubmitEventInput struct {
    Name      string
    Country   string
    City      string
    Latitude  float64
    Longitude float64
    StartDate time.Time
    EndDate   time.Time
    Domain    string
    Website   string
    Submitter SubmitterInput
}

type SubmitterInput struct {
    Name  string
    Email string
}

// --- Constructor ---

func NewEventService(repo repository.EventRepository, audits repository.AuditRepository) EventService {
    return &eventService{repo: repo, audits: audits}
}

// --- Public methods ---

// FP: no side effects
// Submit computes the new event and contributor state; persistence is handled by the caller chain.
// The service layer never calls repo directly for writes — it returns values for the handler to persist.
func (s *eventService) Submit(ctx context.Context, input SubmitEventInput) (model.Event, error) {
    // implementation
}

// --- Private functions ---

// FP: pure function
// validateSubmitInput checks all required fields and returns a descriptive error if any fail.
// Pure: depends only on input, no I/O, always returns the same result for the same input.
func validateSubmitInput(input SubmitEventInput) error {
    // implementation
}
```

---

### Go — Repository (`internal/repository/*.go`)

Same interface-first pattern as service. Repository methods accept and return domain model types — never raw GORM models upward.

```go
package repository

import (
    "context"

    "gorm.io/gorm"

    "github.com/yourname/research-events/internal/model"
)

// --- Interface ---

type EventRepository interface {
    FindByID(ctx context.Context, id uint) (model.Event, error)
    Save(ctx context.Context, event model.Event) (model.Event, error)
    FindByBBox(ctx context.Context, bbox BBoxFilter) ([]model.Event, error)
}

// --- Types ---

type eventRepository struct {
    db *gorm.DB
}

type BBoxFilter struct {
    MinLng, MinLat, MaxLng, MaxLat float64
    Year                           int
}

// --- Constructor ---

func NewEventRepository(db *gorm.DB) EventRepository {
    return &eventRepository{db: db}
}

// --- Public methods ---

func (r *eventRepository) FindByID(ctx context.Context, id uint) (model.Event, error) {
    // always use .WithContext(ctx)
}

// --- Private functions ---
```

---

### Go — Model (`internal/model/*.go`)

```go
package model

import (
    "time"

    "gorm.io/gorm"
)

// --- Types ---

// EventStatus represents the review lifecycle of an event.
// Using a named string type (not iota) keeps values readable in DB and JSON.
type EventStatus string

const (
    EventStatusPending  EventStatus = "pending"
    EventStatusApproved EventStatus = "approved"
    EventStatusRejected EventStatus = "rejected"
)

// Event represents a research conference or event.
type Event struct {
    gorm.Model                          // embeds ID, CreatedAt, UpdatedAt, DeletedAt
    Name            string      `gorm:"not null"`
    Slug            string      `gorm:"uniqueIndex;not null"`
    Country         string      `gorm:"not null"`
    City            string      `gorm:"not null"`
    Latitude        float64     `gorm:"not null"`
    Longitude       float64     `gorm:"not null"`
    StartDate       time.Time   `gorm:"not null"`
    EndDate         time.Time   `gorm:"not null"`
    WebsiteURL      string      `gorm:"not null"`
    Domain          string      `gorm:"not null;index"`
    Status          EventStatus `gorm:"not null;default:pending"`
    Year            int         `gorm:"not null;index"`
    CreatedByID     uint
    CreatedBy       User `gorm:"foreignKey:CreatedByID"`
    LastUpdatedByID uint
    LastUpdatedBy   User `gorm:"foreignKey:LastUpdatedByID"`
}
```

---

### Go — Test (`*_test.go`)

```go
package service_test  // always use external test package (_test suffix) to test the public API

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "go.uber.org/mock/gomock"

    "github.com/yourname/research-events/internal/model"
    "github.com/yourname/research-events/internal/repository/mocks"
)

// One describe block per public function being tested.
// Nested t.Run() per scenario — name describes behaviour, not implementation.

func TestEventService_Submit(t *testing.T) {
    // --- Setup shared across sub-tests ---
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    t.Run("creates contributor when email does not exist", func(t *testing.T) {
        // Arrange
        mockRepo := mocks.NewMockEventRepository(ctrl)
        mockRepo.EXPECT().FindUserByEmail(gomock.Any(), "new@example.com").Return(model.User{}, ErrNotFound)

        svc := NewEventService(mockRepo)

        // Act
        result, err := svc.Submit(context.Background(), validInput)

        // Assert
        require.NoError(t, err)                                  // require stops on failure — use for preconditions
        assert.Equal(t, model.EventStatusPending, result.Status) // assert continues — use for outcome checks
        assert.Equal(t, "New User", result.CreatedBy.Name)
    })

    t.Run("links existing contributor when email already exists", func(t *testing.T) {
        // ...
    })

    t.Run("returns validation error when name is empty", func(t *testing.T) {
        // ...
    })
}
```

---

### TypeScript — Component (`components/**/*.tsx`)

```tsx
// --- Types & Props ---
// All types and interfaces specific to this component live at the top.
// Shared types go in src/types/ — never defined inline mid-file.

interface EventCardProps {
  event: Event
  onSelect: (id: string) => void
}

interface DeadlineRowProps {
  deadline: Deadline
  showHistory: boolean
}

// --- Component ---
// Main exported component comes right after its prop types.
// "use client" directive (if needed) goes at the very top of the file, before imports.

function EventCard({ event, onSelect }: EventCardProps) {
  // state, handlers, and effects follow the hook pattern (see Hook template)

  return (
    <div>
      <DeadlineRow deadline={event.deadlines[0]} showHistory={false} />
    </div>
  )
}

// --- Sub-components ---
// Components used only by this file live below the main component.
// If a sub-component grows complex enough to need its own file, extract it.

function DeadlineRow({ deadline, showHistory }: DeadlineRowProps) {
  return <div>{deadline.date}</div>
}

// --- Export ---
// Named export at the bottom. Never use default exports — they make refactoring harder
// and break auto-import in most editors.

export { EventCard }
export type { EventCardProps }
```

---

### TypeScript — Hook (`hooks/use*.ts`)

```ts
// --- Types ---
interface UseFiltersReturn {
  filters: EventFilters
  setYear: (year: number) => void
  setDomain: (domain: string) => void
  reset: () => void
}

// --- Hook ---
function useFilters(initialYear: number): UseFiltersReturn {
  // 1. State declarations — all useState / useReducer at the top
  const [year, setYear] = useState(initialYear)
  const [domain, setDomain] = useState<string | null>(null)

  // 2. Derived values — useMemo / computed values from state
  const filters = useMemo<EventFilters>(() => ({ year, domain }), [year, domain])

  // 3. Handlers — event handlers and callbacks
  const handleSetYear = useCallback((y: number) => setYear(y), [])
  const reset = useCallback(() => {
    setYear(initialYear)
    setDomain(null)
  }, [initialYear])

  // 4. Effects — useEffect always last, one per concern
  useEffect(() => {
    // side effect here
  }, [filters])

  // 5. Return — always return a named object, never a tuple (except for simple toggle hooks)
  return { filters, setYear: handleSetYear, setDomain, reset }
}

// --- Export ---
export { useFilters }
export type { UseFiltersReturn }
```

---

### TypeScript — Test (`*.test.ts` — Vitest)

```ts
import { describe, it, expect, beforeEach, vi } from 'vitest'

// One describe block per function or hook being tested.
// Nested describe or it() per scenario — name describes behaviour.

describe('useFilters', () => {
  describe('setYear', () => {
    it('updates the year filter and recomputes derived filters', () => {
      // Arrange
      const { result } = renderHook(() => useFilters(2026))

      // Act
      act(() => result.current.setYear(2027))

      // Assert
      expect(result.current.filters.year).toBe(2027)
    })
  })

  describe('reset', () => {
    it('restores all filters to their initial values', () => {
      // ...
    })
  })
})
```

---

### File Naming

| Type | Convention | Example |
|---|---|---|
| Next.js routes | kebab-case (Next.js enforced) | `app/[locale]/events/[id]/page.tsx` |
| React components | PascalCase | `EventCard.tsx`, `DeadlineRow.tsx` |
| Hooks | camelCase with `use` prefix | `useFilters.ts`, `useGlobeState.ts` |
| Utilities / lib | camelCase | `formatDate.ts`, `api.ts` |
| Test files | same name + `.test.ts(x)` | `useFilters.test.ts`, `EventCard.test.tsx` |
| Go source | snake_case | `event_service.go`, `event_handler.go` |
| Go test | same name + `_test.go` | `event_service_test.go` |
| Go mock | `mock_` prefix | `mock_event_repository.go` |
| Backend spec | `kebab-resource-action.yaml` + `.curl.sh` | `events-submit.yaml`, `events-submit.curl.sh` |
| Frontend spec | `kebab-feature-name.md` | `event-export.md`, `globe-filters.md` |
| AI session | `YYYY-MM-DD-kebab-title.md` | `2026-01-15-event-submission-feature.md` |

---

## TDD Practices

This project follows **strict Red-Green-Refactor TDD**. No production code is written before a failing test exists for it.

### The cycle — always in this order
1. **Red** — write a failing test that describes the behaviour you want
2. **Green** — write the minimum production code to make it pass
3. **Refactor** — clean up without changing behaviour; tests must stay green

Never skip step 1. If you are about to write production code without a failing test, stop and write the test first.

---

### Go — libraries and rules

**Libraries:** `testing` stdlib + `github.com/stretchr/testify` (assert/require) + `go.uber.org/mock/gomock` (generated mocks)

**Mock generation**
- Every repository interface must have a generated mock: `mockgen -source=internal/repository/event.go -destination=internal/repository/mocks/mock_event.go`
- Mocks live in `internal/repository/mocks/` and `internal/service/mocks/`
- Never hand-write mocks — always regenerate with `make generate-mocks` after an interface changes
- Run `make generate-mocks` before writing any service or handler test

**Test file structure**
- Tests live alongside source: `event_service.go` → `event_service_test.go`
- Use table-driven tests with `t.Run()` for all non-trivial cases
- Use `require` (stops on failure) for setup/preconditions; `assert` (continues) for multiple assertions in the same case
- Test names describe behaviour, not implementation: `TestEventService_Submit_CreatesContributorWhenEmailNotFound`, not `TestSubmit`

**What to test per layer**

| Layer | Tool | What to test |
|---|---|---|
| `repository/` | Real DB via `dockertest` or test Postgres | Queries return correct rows; constraints enforced |
| `service/` | gomock for repository mocks | Business logic, edge cases, error paths |
| `handler/` | `net/http/httptest` + gomock for service mocks | HTTP status codes, request parsing, response envelope |
| `middleware/` | `net/http/httptest` | JWT validation, CORS headers, rate-limit rejection |

**Required test cases per feature** — every new feature must cover:
- Happy path
- Input validation failure
- Not found / empty result
- Unauthorized / forbidden (where auth applies)
- AuditLog is written (assert the mock was called with the correct args)
- Service functions are pure — test the same input twice and assert identical output (no hidden state)

---

### Frontend — Vitest

**Scope:** logic and hooks only — no component rendering tests (no React Testing Library)

**What to test:**
- Custom hooks (`useFilters`, `useGlobeState`, `useEvents`) — test state transitions and return values
- `lib/api.ts` — mock `fetch` and assert correct URLs, headers, error handling
- Pure utility functions in `lib/utils.ts`
- `as const` enum helpers and type guards in `src/types/`

**What NOT to test with Vitest:**
- Component rendering or DOM structure — these are covered by Playwright E2E
- Next.js routing or middleware — test those with Playwright

**Test file location:** colocated — `useFilters.ts` → `useFilters.test.ts`

**Naming:** same behaviour-first convention — `it('returns only active deadlines for the current year')`

---

### Commit gate

Before every commit all of the following must pass with zero failures:
```bash
go test ./...        # all Go unit tests
go vet ./...         # static analysis
pnpm typecheck       # tsc --noEmit
pnpm test            # Vitest
```

---

## Spec-Driven Development

Every feature starts with a spec. No feature enters the pair programming workflow without an approved spec file in `specs/`. Bugs are exempt from this rule.

### Directory structure

```
specs/
├── backend/          # One YAML file per backend endpoint or domain operation
│   └── auth-login.yaml
└── frontend/         # One Markdown file per frontend feature
    └── event-export.md
```

### Spec naming

| Type | Convention | Example |
|---|---|---|
| Backend spec | `kebab-resource-action.yaml` | `auth-login.yaml`, `events-submit.yaml` |
| Frontend spec | `kebab-feature-name.md` | `event-export.md`, `globe-filters.md` |

---

### Backend spec format (`specs/backend/*.yaml`)

```yaml
endpoint: POST /api/v1/auth/login

description: >
  Authenticates an admin or moderator and returns a JWT stored in an HTTP-only cookie.

request:
  body:
    email: string (required)
    password: string (required)

responses:
  200:
    description: Authentication successful
    body:
      token: string
      expires_in: 3600
  400:
    description: Missing or empty fields
    body:
      error:
        code: VALIDATION_ERROR
        message: "email and password are required"
  401:
    description: Invalid credentials
    body:
      error:
        code: INVALID_CREDENTIALS
        message: "invalid credentials"
    note: >
      Always return 401 for both wrong password AND non-existent email.
      Never reveal whether the email exists in the system.

rules:
  - Passwords hashed with bcrypt at cost 12 — never store plaintext
  - JWT stored in HTTP-only cookie only — never returned in response body
  - Rate-limit to 10 attempts per IP per minute

permissions:
  - Public endpoint — no authentication required to attempt login
  - Only users with role admin or moderator can successfully authenticate

border_cases:
  - Empty email field → 400
  - Empty password field → 400
  - Valid email format but not in DB → 401 (same response as wrong password)
  - Correct email, wrong password → 401
  - Correct credentials but role = contributor → 401 (contributors cannot log in yet)
  - Malformed JSON body → 400

definition_of_done:
  - Returns 200 + cookie on valid admin credentials
  - Returns 401 for wrong password (no information about email existence)
  - Returns 401 for non-existent email (same response shape as wrong password)
  - Returns 401 for contributor role attempting login
  - Returns 400 for missing fields
  - bcrypt cost is 12 (verified in test)
  - JWT is in HTTP-only cookie, not in response body
  - All border cases above have a corresponding test case
```

---

### Frontend spec format (`specs/frontend/*.md`)

```markdown
# Event Export

## Description
Allows admins to export event data for a selected period as a CSV file delivered via email.

## Behaviour
- Only admin role can trigger an export
- Export is asynchronous — returns 202 + jobId immediately, email sent when complete
- Email contains a download link that expires in 24 hours

## Rules
- Maximum 1 active export per admin at a time — reject if one is already running
- If no events match the selected period → export file contains headers only, no data rows
- Results over 50 000 rows are paginated into files of 10 000 rows each
- Export link expires after 24 hours — expired links return 410 Gone

## Permissions
- `admin` role only — moderators and contributors cannot trigger exports

## Error cases
| Scenario | Expected behaviour |
|---|---|
| Export already in progress | 409 Conflict + message |
| No data in selected period | 202 + empty file (headers only) |
| Expired download link | 410 Gone |
| Unauthenticated request | 401 |
| Moderator attempts export | 403 Forbidden |

## Border / corner cases
- What if the admin logs out mid-export? → job continues, email still sent
- What if the email fails to send? → job marked failed, admin can retry from UI
- What if >50k rows AND only 1 active export allowed? → pagination happens transparently

## Definition of done
- [ ] 202 + jobId returned immediately on valid export request
- [ ] Email with download link sent on completion
- [ ] Link expires after exactly 24 hours
- [ ] Second export attempt while first is running → 409
- [ ] Empty period → file with headers only, no error
- [ ] 50 001 rows → two files generated (50 000 + 1)
- [ ] All error cases in the table above have a test
- [ ] Moderator request → 403 (tested)
```

---

### Spec approval checklist

Before a spec is approved and the workflow moves to Phase 1, Claude must verify all five gates:

1. **Permissions mapped** — every role that can and cannot perform this action is explicit
2. **All error cases mapped** — every non-happy-path has a defined HTTP status + response shape
3. **Business rules approved** — the human has explicitly confirmed all rules (e.g. bcrypt cost, expiry time, limits)
4. **Definition of done is checkable** — every item in DoD is a concrete, verifiable statement (not "works correctly")
5. **Border/corner cases reviewed** — Claude must explicitly ask: *"Are there any edge cases not covered here?"* before approving

If any gate fails, Claude revises the spec and presents it again. Claude must not proceed to Phase 1 until all five gates pass and the human says **"spec approved"**.

After the human approves the spec, Claude immediately generates the corresponding `*.curl.sh` file alongside the spec:

```bash
# specs/backend/auth-login.curl.sh
#!/bin/bash
# Auth: Login
# Spec: specs/backend/auth-login.yaml

BASE_URL="${BASE_URL:-http://localhost:8080}"

# 200 — successful login
curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"secret"}' | jq .

# 400 — missing fields
curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":""}' | jq .

# 401 — wrong password
curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"wrong"}' | jq .

# 401 — non-existent email (same response — never reveals existence)
curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"ghost@example.com","password":"secret"}' | jq .
```

Rules for curl files:
- One curl per response code defined in the spec — every case covered, not just the happy path
- `BASE_URL` defaults to `localhost:8080`, overridable via env var
- Pipe through `jq .` for readable output
- Comment above each curl states the expected HTTP status and scenario
- Frontend uses these files as the integration contract reference

---

## Pair Programming Workflow

This is the mandatory process for every feature. Bugs skip Phase 0 (Spec) and start at Phase 1 (Discovery). Claude is the coding partner — the human drives intent and approval, Claude drives implementation.

```
Feature: Phase 0 → 1 → 2 → 3 → 4 → 5 → 6
Bug fix:           Phase 1 → 2 → 3 → 4 → 5 → 6
```

---

### Phase 0 — Spec (features only)

When the human names a feature to build, Claude interviews them to write the spec together. Claude asks all necessary questions in a single message — covering behaviour, rules, permissions, error cases, and definition of done — then drafts the spec file.

The spec interview covers:
- What does this feature do and who uses it?
- What are the inputs and outputs?
- What roles can access this? What roles are explicitly blocked?
- What are the business rules and constraints?
- What happens in every error scenario?
- What border/corner cases exist?
- What does "done" look like — concretely and verifiably?

After drafting, Claude runs the **spec approval checklist** (see Spec-Driven Development section). Once all five gates pass, Claude writes the spec file to `specs/backend/` or `specs/frontend/` and asks: **"Spec written. Do you approve it so we can move to Phase 1?"**

Do not proceed until the human says **"spec approved"** or equivalent.

---

### Phase 1 — Discovery

Claude asks any remaining implementation questions not answered by the spec — focusing on:
- Which existing files, services, or models does this touch?
- Are there any dependencies or constraints not captured in the spec?
- Any implementation preferences for this specific feature?

Do not ask questions already answered in the spec. Do not ask questions one at a time.

---

### Phase 2 — Plan (requires human approval before proceeding)

Claude produces a written plan containing:

1. **Interfaces and signatures** — every function, method, and type that will be created or changed, with full signatures (no implementation yet)
2. **Test list** — every test case derived directly from the spec's error cases, border cases, and definition of done, named using the behaviour-first convention, grouped by layer
3. **File list** — every file that will be created or modified
4. **Cycle breakdown** — the sequence of Red-Green-Refactor cycles, one per logical unit

Claude asks: **"Does this plan look correct? Should I proceed to Phase 3?"**

Do not write any code until the human approves.

---

### Phase 3 — Red (failing tests first)

For each unit in the approved plan, in order:

1. Write the test file — test cases must trace directly to spec items (reference the spec rule or DoD item in the test name or comment)
2. Show the actual output of running the tests — they must all fail
3. Ask: **"Tests are red. Should I proceed to Green?"**

Do not write production code during this phase.

---

### Phase 4 — Green (minimum production code)

For each unit, in order:

1. Write the minimum production code to make the tests pass
2. Show the actual output of running the tests — they must all pass
3. Ask: **"Tests are green. Should I proceed to Refactor?"**

---

### Phase 5 — Refactor

1. Clean up: naming, structure, duplication — no behaviour changes
2. Run the tests again — must still be green
3. Show a brief summary of what was refactored and why
4. Ask: **"Refactor done. Move to the next unit or is this feature complete?"**

---

### Phase 6 — Documentation (requires human approval)

After Phase 5 is complete, Claude asks:
**"Feature is done. Should I write the docs? (Phase 6)"**

The human may decline if the feature is too trivial.

**Bug fix** → add a `bug_fix:` section to the relevant spec file, or create a new spec if none exists:
- Symptom + root cause + fix + how to verify + curl that proves it's fixed

**AI session** → `ai-sessions/YYYY-MM-DD-session-title.md`:
- Goal + decisions made + state at end + context to restore

After any session summary, update `README.md` index with a one-line entry.

Note: all feature contracts live in the spec files — no separate `docs/` folders exist.

---

### Hard rules

- **Never skip a phase.** Features: 0 → 1 → 2 → 3 → 4 → 5 → 6. Bugs: 1 → 2 → 3 → 4 → 5 → 6.
- **No code without an approved spec** — for features. Bugs are exempt.
- **Test cases must trace to spec** — every test references the spec rule, error case, or DoD item it covers.
- **Never bundle units** — each function or method gets its own Red-Green-Refactor cycle.
- **Always show actual test output** — never claim tests pass or fail without showing it.
- **One approval per phase transition** — never move forward without an explicit "yes" or "proceed".
- **If new unknowns appear mid-phase** — stop, return to the appropriate earlier phase, resolve, then continue.
- **FP rule** — every `internal/service/` function gets an `// FP:` comment with inline explanation.

---

## Git Workflow

- Branch naming: `feature/short-description`, `fix/short-description`
- Conventional commits: `feat:`, `fix:`, `refactor:`, `test:`, `docs:`
  - `test:` commits should appear **before** the `feat:` commit they enable (Red before Green)
- One logical change per commit — never bundle unrelated files
- PRs require the full commit gate to pass (see TDD section)

---

## Learning Goals

This project has two explicit learning tracks running in parallel. Claude must support both actively.

### Go
- When introducing a Go concept for the first time (interfaces, goroutines, channels, defer, embedding, generics), add a short comment block explaining what it is and why Go does it this way
- Mention idiomatic Go alternatives when a non-idiomatic approach is used
- Explain error handling patterns (`errors.Is`, `errors.As`, wrapping) when they appear

### Functional Programming
- Every function in `internal/service/` carries an `// FP:` comment (see FP section for full rules)
- When writing a service function, always explain in chat: what FP concept it uses, what it prevents, and what the imperative (non-FP) version would look like — so the contrast is clear
- If a piece of logic could be written in both FP and non-FP style, show both briefly and explain the trade-off before committing to the FP version

### General
- Prefer verbose clarity over terse cleverness — optimize for understanding, not line count
- When there are multiple valid approaches, name them and explain why one was chosen
- Tests are also learning material — test names and structure should be readable as documentation

---

## Deployment Notes

- **Frontend (Vercel):** Set `NEXT_PUBLIC_API_URL` env var pointing to Fly.io backend
- **Backend (Fly.io):** Uses `fly secrets` for `DATABASE_URL`, `JWT_SECRET`
- **Migrations:** Run `make migrate-up` as part of the deploy pipeline before the new binary starts