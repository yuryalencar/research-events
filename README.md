<p align="center">
  <img src="frontend/public/logo-with-opensource.png" alt="ReSEARCH Events logo" width="400" />
</p>

# About ReSEARCH Events

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
| 2026-06-18 | [Deadline Management — Frontend Planning](ai-sessions/2026-06-18-deadline-management-frontend-planning.md) | Spec + plan for single-page deadline management (`/events/[slug]/deadlines`): add/supersede/cancel in one batch, pencil icon in event drawer/sheet, sessionStorage for event state, parallel API submission, same design language as submission wizard. Phases 0–2 complete; Phase 3 (Red) is next. |
| 2026-06-19 | [Deadline Management — Frontend Implementation](ai-sessions/2026-06-19-deadline-management-frontend-implementation.md) | Full implementation of `/events/[slug]/deadlines`: `useDeadlineManage` hook (58 tests), `DeadlineCard` (3 states), `AddDeadlineCard` (2-row layout, Required checkbox), `DeadlineManagePage` (sessionStorage loader + content split), supersede visualization in drawer (strikethrough → new date). 210 tests. |
| 2026-06-19 | [Management Portal — Frontend](ai-sessions/2026-06-19-manage-portal-frontend.md) | Hidden `/manage` login page (email/password + eye toggle, moderator banner), JWT decoded client-side via `atob()`, localStorage session persistence, role-based redirect to `/manage/admin` or `/manage/moderator` welcome dashboards. Fixed `credentials: "include"` on login request. 215 tests. |
| 2026-06-19 | [Management Dashboard — Frontend Planning](ai-sessions/2026-06-19-manage-dashboard-planning.md) | Spec approved for admin/moderator event review queue: shared `ManageDashboard` component, event cards with own-event grey variant for moderators, status/tier/year filters with Apply, 30-per-page pagination, sticky header with avatar menu. Phase 0 done — Phase 1 is next. |
| 2026-06-19 | [Management Dashboard — Frontend Implementation](ai-sessions/2026-06-19-manage-dashboard-implementation.md) | Full implementation of admin/moderator review queue: `useReviewEvents` hook, `ManageHeader` (Welcome + role + avatar menu), `EventReviewCard` (normal + own-event grey), `ManageDashboard` (filters, list, pagination). Custom select arrow (`appearance-none` + ChevronDown), `max-w-4xl` centering. 225 tests. |
| 2026-06-19 | [Management Event Review — Frontend](ai-sessions/2026-06-19-manage-review-frontend.md) | 3-step review wizard: Step 1 edits event details (Leaflet map pre-pinned), Step 2 approve/reject with confirmation modals + ReviewSuccess screen, Step 3 deadline management with auto-filled contributor. Fixed LocationPicker async init race via `latLngRef`. Switched to CartoDB Voyager tiles for English place labels. |
| 2026-06-21 | [Globe / Table View Toggle — Frontend](ai-sessions/2026-06-21-globe-table-toggle-feature.md) | Floating toggle button (desktop-only, `hidden md:flex`) switches homepage between 3D globe and a card-list table view. `useViewMode` hook manages state + forced resize-back-to-globe with sonner toast. `EventTableCard` shows all event fields collapsed, deadlines + manage link expanded. 257 tests. |
| 2026-06-21 | [Update User Password — Backend](ai-sessions/2026-06-21-update-user-password-feature.md) | `PATCH /api/v1/users/me/password` — authenticated, rate-limited (10/min). Validates current password, complexity (8+ chars, upper, lower, special), hashes with bcrypt cost 12. FP service layer: 3 pure functions. Also fixed stale test DB (truncated committed rows that broke `ListEvents_*` integration tests). 401 tests. |
| 2026-06-21 | [Session Guard + Eager Token Validation](ai-sessions/2026-06-21-session-guard-eager-token-validation.md) | Bug fixes: double-namespace `errors.errors.TOKEN_MISSING` (root-level `t` in UpdatePasswordCard), auth error redirect from update-password hook (`onAuthError` callback). New: `GET /api/v1/users/me` (no DB, JWT context only), `validateSession()`, `useSessionGuard` hook applied to all 6 protected pages. 277 frontend tests, full backend suite passes. |
| 2026-06-22 | [Language Selector (Frontend)](ai-sessions/2026-06-22-language-selector-feature.md) | Fixed bottom-right flag button on every page — switches between 🇺🇸 English, 🇧🇷 Português, 🇪🇸 Español, 🇩🇪 Deutsch while preserving the current URL path. Created `src/i18n/navigation.ts` (missing `createNavigation` module for next-intl v4 locale-aware routing). |
| 2026-06-22 | [Year Filter: "From Year" Semantics](ai-sessions/2026-06-22-year-filter-from-semantics.md) | `?year=2026` now returns year >= 2026; omitting year returns all events. Backend: `*int` nullable year + updated `firstDeadlineMonth` subquery. Frontend: `number \| undefined` year type, globe locks year (min = currentYear−2), submission page min = currentYear, "From year" label in all 4 locales. 278 tests. |
| 2026-06-24 | [Globe Event Clustering (Frontend)](ai-sessions/2026-06-24-globe-event-clustering.md) | `supercluster` groups nearby pins into violet cluster pins (dynamic, zoom-aware); clicking opens a multi-event drawer immediately. New `useGlobeClusters` hook, `altitudeToZoom` pure function, `ClusterEventDrawer` (Dialog/Drawer responsive). InfoModal violet legend row. 4 new i18n keys across all locales. 299 tests. |
| 2026-06-24 | [Globe Loading Message Progression (Frontend)](ai-sessions/2026-06-24-globe-loading-progression.md) | `useLoadingMessage` hook cycles through 4 messages while the backend cold-starts: "Loading events…" → "Waking up our server…" → "First load takes a moment…" → "Almost there, hang tight…". 2500ms initial delay, 1500ms between steps, resets on re-fetch. 307 tests. |
| 2026-06-26 | [Admin User Management — Backend](ai-sessions/2026-06-26-admin-user-management.md) | `POST /api/v1/admin/users` (register admin/moderator) + `PATCH /api/v1/admin/users/{id}/role` (role change with session invalidation). `ExistsByEmail`, `Create`, `UpdateRole` in repository; 5 FP service functions; `AuditActionRoleChanged`. 458 backend tests. |

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
| Event Submission Wizard (Frontend) | [event-submission-wizard.md](specs/frontend/event-submission-wizard.md) | Done |
| Deadline Management (Frontend) | [deadline-management.md](specs/frontend/deadline-management.md) | Done |
| Management Portal (Frontend) | [manage-portal.md](specs/frontend/manage-portal.md) | Done |
| Management Dashboard (Frontend) | [manage-dashboard.md](specs/frontend/manage-dashboard.md) | Done |
| Management Event Review (Frontend) | [manage-review.md](specs/frontend/manage-review.md) | Done |
| Globe / Table View Toggle (Frontend) | [globe-table-toggle.md](specs/frontend/globe-table-toggle.md) | Done |
| Users: Update Password | [users-update-password.yaml](specs/backend/users-update-password.yaml) | Done |
| Users: Get Current Session (Me) | [users-me.yaml](specs/backend/users-me.yaml) | Done |
| Language Selector (Frontend) | [language-selector.md](specs/frontend/language-selector.md) | Done |
| Events: Year filter — "from year" semantics (Backend) | [events-list-year-from-semantics.yaml](specs/backend/events-list-year-from-semantics.yaml) | Done |
| Year filter — "from year" semantics (Frontend) | [year-filter-from-semantics.md](specs/frontend/year-filter-from-semantics.md) | Done |
| Globe Event Clustering (Frontend) | [globe-event-clustering.md](specs/frontend/globe-event-clustering.md) | Done |
| Globe Loading Message Progression (Frontend) | [globe-loading-progression.md](specs/frontend/globe-loading-progression.md) | Done |
| Admin: Register User | [admin-users-register.yaml](specs/backend/admin-users-register.yaml) | Done |
| Admin: Change User Role | [admin-users-change-role.yaml](specs/backend/admin-users-change-role.yaml) | Done |

---

## Releases

| Version | Date | Highlights |
|---------|------|------------|
| [v0.1.0](https://github.com/yuryalencar/research-events/releases/tag/v0.1.0) | 2026-06-19 | Initial public release — Go API (auth, event submission, deadlines, admin review, OTel tracing), Globe homepage, 3D pins, event detail, submission wizard, management portal + dashboard, deadline management. |
| [v0.1.1](https://github.com/yuryalencar/research-events/releases/tag/v0.1.1) | 2026-06-24 | Globe event clustering (`supercluster`), multi-event cluster drawer, year-from filter semantics, language selector, table view toggle, session guard, update password. |
| [v0.1.2](https://github.com/yuryalencar/research-events/releases/tag/v0.1.2) | 2026-06-24 | Globe loading message progression — `useLoadingMessage` hook cycles through 4 messages while the backend cold-starts, improving perceived UX on first visit. |

---

## License

[MIT](./LICENSE) © 2026 Yury Lima
