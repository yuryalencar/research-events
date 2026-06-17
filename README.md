# ReSEARCH Events

A collaborative, open-source platform that aggregates research conferences and events in software engineering and computer science. Researchers can discover events on an interactive 3D globe, filter by year and location, and submit new events for admin review.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Frontend | Next.js 16, React 19, TypeScript 6, Tailwind CSS v4 |
| Globe / Map | Globe.gl (3D WebGL) + fallback 2D Leaflet map |
| i18n | next-intl — English, Portuguese, Spanish, German |
| Backend | Go 1.26, `net/http` stdlib |
| Database | PostgreSQL 16 + GORM v2 + Goose migrations |
| Auth | JWT in HTTP-only cookies (admin/moderator only) |
| Monitoring | Sentry + OpenTelemetry |

---

## Prerequisites

| Tool | Version | Install |
|---|---|---|
| Node.js | 22+ | [nodejs.org](https://nodejs.org) |
| pnpm | 10+ | `npm install -g pnpm` |
| Go | 1.26+ | [go.dev/dl](https://go.dev/dl) |
| Docker | any | [docker.com](https://www.docker.com) |

---

## Setup

### 1. Clone the repo

```bash
git clone https://github.com/yuryalencar/research-events.git
cd research-events
```

### 2. Install tools

```bash
make install-tools   # installs goose migration runner — run once per machine
```

### 3. Start the databases

```bash
docker compose up -d
```

This starts two PostgreSQL 16 instances:
- `postgres` on port `5432` — dev database (`research_events`)
- `postgres_test` on port `5433` — integration test database (`research_events_test`)

Both use `postgres`/`postgres` credentials.

### 4. Run migrations

```bash
make migrate-up         # applies migrations to dev DB
make migrate-test-up    # applies migrations to test DB (needed to run `go test ./...`)
```

### 5. Backend

```bash
cd backend
cp .env.example .env   # set DATABASE_URL and JWT_SECRET at minimum
go mod download
go run ./cmd/api
```

Required env vars:
- `DATABASE_URL` — Postgres connection string
- `JWT_SECRET` — secret for signing JWTs (any long random string)

The API will be available at `http://localhost:8080`.

### 6. Frontend

```bash
cd frontend
pnpm install
pnpm dev
```

The app will be available at `http://localhost:3000` (redirects to `/en/` by default).

---

## Development Commands

### Frontend

```bash
cd frontend
pnpm dev          # dev server → http://localhost:3000
pnpm build        # production build
pnpm typecheck    # tsc --noEmit (run before every commit)
pnpm lint         # ESLint
pnpm test         # Vitest
```

### Backend

```bash
cd backend
go run ./cmd/api      # run without hot reload
air                   # hot reload (install: go install github.com/air-verse/air@latest)
go test ./...         # all tests
go vet ./...          # static analysis
```

### Infrastructure

```bash
make install-tools      # install goose binary — run once per machine
docker compose up -d    # start local Postgres (5432) and test Postgres (5433)
make migrate-up         # run pending migrations on dev DB
make migrate-test-up    # run pending migrations on test DB
make migrate-down       # roll back last migration on dev DB
make generate-mocks     # regenerate gomock files
make generate-types     # regenerate frontend types from OpenAPI spec
```

---

## Project Structure

```
/
├── frontend/           # Next.js app
│   └── src/
│       ├── app/[locale]/   # routes (Globe, Event detail, Submit, Admin)
│       ├── components/     # UI primitives, Globe, Map, Events, Admin
│       ├── hooks/          # useEvents, useFilters, useGlobeState
│       ├── lib/            # API client, utils, constants
│       ├── types/          # generated from OpenAPI spec
│       └── messages/       # i18n translations (en, pt, es, de)
│
├── backend/            # Go API
│   ├── cmd/api/        # entry point
│   └── internal/
│       ├── handler/    # HTTP handlers
│       ├── service/    # business logic (functional programming layer)
│       ├── repository/ # GORM queries
│       ├── model/      # domain structs
│       ├── middleware/ # JWT, CORS, rate-limit
│       ├── health/     # GET /health extensible checker
│       └── config/     # env parsing
│
├── migrations/         # Goose SQL files
├── specs/              # feature specs (written before any code)
├── docs/               # learning notes (Go concepts, FP, backend libraries)
└── ai-sessions/        # session summaries for context recovery
```

---

## Sessions

| Date | Session | Summary |
|------|---------|---------|
| 2026-06-08 | [Server Bootstrap + Health Check](ai-sessions/2026-06-08-server-bootstrap-health-check.md) | Go server wired: config, DB ping, CORS, `GET /health` extensible checker, graceful shutdown. 21 tests. |
| 2026-06-09 | [Auth Feature](ai-sessions/2026-06-09-auth-feature.md) | Login, refresh-token, logout, account unlock. Stateful JWT (JTI), token rotation, rate limiting, lockout. Users + audit_logs migrations. 94 tests. |
| 2026-06-12 | [Event Submission Feature](ai-sessions/2026-06-12-event-submission-feature.md) | `POST /api/v1/events/submit` — public submission with contributor lookup/creation, optional deadlines, slug reuse via partial unique index, FP service layer. Events + deadlines migrations. 148 tests. |
| 2026-06-13 | [Events List Feature](ai-sessions/2026-06-13-events-list-feature.md) | `GET /api/v1/events` — public, filterable (year/domain/country/status/tier/first_deadline_month/bbox) + paginated listing with active deadlines and attribution. New `tier` column + 120 req/min rate limiter. 207 tests. |
| 2026-06-14 | [Add Deadlines to an Approved Event](ai-sessions/2026-06-14-event-deadlines-add-feature.md) | `POST /api/v1/events/{id}/deadlines` — public, lets any contributor add one or more deadlines to an approved event. Single vs. batch audit actions, `batch_deadlines_added` migration, shared 50 req/min rate limiter. 242 tests. |
| 2026-06-14 | [Cancel a Deadline](ai-sessions/2026-06-14-event-deadlines-cancel-feature.md) | `PATCH /api/v1/events/{eventId}/deadlines/{deadlineId}/cancel` — public, lets any contributor cancel an active deadline (`is_active=false`, `superseded_by_id=NULL`). New `deadline_cancelled` audit action + migration, shared 50 req/min rate limiter. 272 tests. |
| 2026-06-14 | [Deadlines: time + timezone fields](ai-sessions/2026-06-14-deadlines-time-timezone-feature.md) | Cross-cutting prerequisite for deadline supersession — adds optional `time` (HH:MM) and `timezone` (free string) to `Deadline`, returned/accepted by submit, add-deadlines, list, and cancel-reload. New migration 009. 287 tests. |
| 2026-06-14 | [Supersede a Deadline](ai-sessions/2026-06-14-event-deadlines-supersede-feature.md) | `POST /api/v1/events/{eventId}/deadlines/{deadlineId}/supersede` — public, lets any contributor replace an active deadline with a new date/time/timezone; old row marked `is_active=false, superseded_by_id=<new id>`, new `superseded_by_id` field on every deadline response. Shared 50 req/min rate limiter. 322 tests. |
| 2026-06-14 | [Admin/Moderator Event Review](ai-sessions/2026-06-14-admin-events-review-feature.md) | `PATCH /api/v1/admin/events/{id}/review` — admin/moderator only, approve/reject an event with an optional partial field edit (same validation as submission) and reviewer reason. New nullable `audit_logs.reason` column, re-review always allowed, moderators blocked from reviewing their own submissions. 377 tests. |
| 2026-06-14 | [OpenTelemetry Tracing — Spec + Plan (partial)](ai-sessions/2026-06-14-observability-opentelemetry-planning.md) | Spec approved (HTTP + DB spans -> Sentry via `sentryotel` bridge) and Phase 2 plan presented. Phase 3 (Red) not started — resume from this checkpoint. |
| 2026-06-15 | [OpenTelemetry Tracing — Sentry Double-Sampling Bug Fix](ai-sessions/2026-06-15-observability-tracing-bugfix.md) | Feature was implemented (388 tests) but no traces reached Sentry. Root cause: unset `TracesSampleRate` (0.0) on `sentry.Init` caused sentry-go to drop every transaction independently of the OTel sampler. Fixed by setting `TracesSampleRate: 1.0`. 389 tests. |
| 2026-06-15 | [API Client + Error Handling (Frontend)](ai-sessions/2026-06-15-frontend-api-client-error-handling-feature.md) | First frontend feature: typed `lib/api/*` client (`apiRequest`/`apiPrivateRequest` with refresh-and-retry, `ApiError`), centralized `errorMessageKey`/`handleApiError` -> sonner toasts, `errors` i18n namespace (25 keys, 4 locales), hand-written `types/api.ts`, shadcn/ui + sonner setup. 57 tests. Also fixed `pnpm lint` (Next 16 dropped `next lint`; added flat `eslint.config.mjs`, pinned `eslint` to `^9` for `eslint-plugin-react` compatibility). |
| 2026-06-15 | [Globe Homepage (Frontend)](ai-sessions/2026-06-15-globe-homepage-feature.md) | 3D globe (Globe.gl) plots approved events as pins (yellow/pink/red for default/selected/past), starfield background, Sheet (desktop)/Drawer (mobile) detail view with localized domain + tier + date-range display, scrollable deadline list, loading/empty-state overlays, `proxy.ts` rename, hydration fix, 8 seeded events. 93 tests. |
| 2026-06-16 | [Globe Event Deep-link via URL Slug (Frontend)](ai-sessions/2026-06-16-globe-event-deeplink-feature.md) | `?event=SLUG` deep-linking on the globe homepage — clicking a pin sets the param, closing removes it, loading with the param auto-selects from the existing events list (no extra API call). Globe rotates to selected event (preserving zoom). `useSelectedEvent` extended with URL sync + slug resolution. 101 tests. |
| 2026-06-16 | [Info Modal — Floating Info Button (Frontend)](ai-sessions/2026-06-16-info-modal-feature.md) | Floating `ⓘ` button (bottom-left, `z-40`) opens a Dialog (desktop) / Drawer (mobile) with 6 sections: version, pin legend, submission flow, moderator contact, author credit, open-source invite. New `ui/dialog.tsx` primitive, full i18n (en/pt/es/de), hidden on mobile when event Drawer is open. |
| 2026-06-16 | [Globe Event Filters (Frontend)](ai-sessions/2026-06-16-globe-filters-feature.md) | Collapsible floating filter panel (top-left): year stepper, domain dropdown, tier chips, country native-select, deadline month dropdown. Draft/applied two-stage state, Apply rotates globe to first result, Reset restores form only, mobile auto-close, locale-aware month names, `noEventsFiltered` empty-state message. `useFilters` + `useEvents` updated. 129 tests. |
| 2026-06-17 | [Event Submission Wizard (Frontend)](ai-sessions/2026-06-17-event-submission-wizard-frontend.md) | 3-step wizard: Step 1 duplicate check (server-side paginated pending events table), Step 2 event details form + Leaflet map pin, Step 3 optional deadlines. Floating `+` button on globe homepage with custom CSS hover tooltip. `useEventSearch` + `useSubmitWizard` hooks. Fixed Leaflet "Map container already initialized" (async `cancelled` flag), added coordinates for all 195 countries. Full i18n (en/pt/es/de). 152 tests. |

## Specs

| Feature | Spec | Status |
|---------|------|--------|
| Server Bootstrap + Health Check | [server-bootstrap.yaml](specs/backend/server-bootstrap.yaml) | Done |
| Auth: Login | [auth-login.yaml](specs/backend/auth-login.yaml) | Done |
| Auth: Refresh Token | [auth-refresh-token.yaml](specs/backend/auth-refresh-token.yaml) | Done |
| Auth: Logout | [auth-logout.yaml](specs/backend/auth-logout.yaml) | Done |
| Auth: JWT Middleware + Role Guard | [auth-middleware.yaml](specs/backend/auth-middleware.yaml) | Done |
| Admin: Unlock User Account | [admin-users-unlock.yaml](specs/backend/admin-users-unlock.yaml) | Done |
| Database: Users table | [database-users.yaml](specs/backend/database-users.yaml) | Done |
| Events: Submit | [events-submit.yaml](specs/backend/events-submit.yaml) | Done |
| Events: List | [events-list.yaml](specs/backend/events-list.yaml) | Done |
| Events: Add Deadlines | [events-deadlines-add.yaml](specs/backend/events-deadlines-add.yaml) | Done |
| Events: Cancel a Deadline | [events-deadlines-cancel.yaml](specs/backend/events-deadlines-cancel.yaml) | Done |
| Deadlines: time + timezone fields | [deadlines-add-time-timezone.yaml](specs/backend/deadlines-add-time-timezone.yaml) | Done |
| Events: Supersede a Deadline | [events-deadlines-supersede.yaml](specs/backend/events-deadlines-supersede.yaml) | Done |
| Admin: Review an Event | [admin-events-review.yaml](specs/backend/admin-events-review.yaml) | Done |
| OpenTelemetry Tracing | [observability-opentelemetry.yaml](specs/backend/observability-opentelemetry.yaml) | Done |
| API Client + Error Handling (Frontend) | [api-client-error-handling.md](specs/frontend/api-client-error-handling.md) | Done |
| Globe Homepage (Frontend) | [globe-homepage.md](specs/frontend/globe-homepage.md) | Done |
| Globe Event Deep-link (Frontend) | [globe-event-deeplink.md](specs/frontend/globe-event-deeplink.md) | Done |
| Info Modal — Floating Info Button (Frontend) | [info-modal.md](specs/frontend/info-modal.md) | Done |
| Globe Event Filters (Frontend) | [globe-filters.md](specs/frontend/globe-filters.md) | Done |

---

## License

[MIT](./LICENSE) © 2026 Yury Lima
