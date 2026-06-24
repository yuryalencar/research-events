# ReSEARCH Events — Release Notes

---

## v0.1.1 — June 24, 2026

### Globe Event Clustering

Fixes the UX problem where multiple events at close or identical coordinates on the 3D globe would overlap into a single, unclickable pin.

**Clustering**
- Nearby events are now automatically grouped into violet cluster pins using [supercluster](https://github.com/mapbox/supercluster)
- Clusters are dynamic — they split into individual pins as the user zooms in, and reform on zoom out
- Minimum cluster size is 2 events; single events are never wrapped in a cluster
- Cluster pins are sized proportionally to the number of events they contain (2 → small, 11+ → large)
- Clustering responds to camera movement with a 150 ms debounce so re-computation does not fire on every animation frame during a drag

**Interaction**
- Clicking a cluster opens a multi-event drawer immediately — no zoom-in step required
- The drawer shows a compact, scrollable list of every event in the cluster, sorted by start date
- Clicking a row closes the drawer and opens the existing event detail panel for the selected event
- The cluster drawer closes automatically when a filter is applied or the view is toggled to table mode

**Visual**
- Cluster pins use violet (`#a78bfa`) to distinguish them from individual event pins (yellow/pink/red)
- The Info Modal legend gains a violet row explaining what cluster pins represent

**Internationalisation**
- 4 new i18n keys added across all 4 locales (English, Portuguese, Spanish, German):
  - Drawer title with event count
  - Cluster pin tooltip label
  - Info Modal legend label and description

---

## v0.1.0 — June 23, 2026

The first public release of ReSEARCH Events — a collaborative, open-source platform that aggregates research conferences and events in software engineering and computer science, built with Next.js 15, Go, and PostgreSQL.

### Backend

**Server & Infrastructure**
- HTTP server bootstrap with Go `net/http` stdlib — no framework
- `GET /health` endpoint with extensible checker pattern (database latency, status rollup)
- CORS middleware with `Access-Control-Allow-Credentials` support for cross-origin cookie authentication
- JWT authentication (stateless HS256) stored in HTTP-only cookies with refresh-token rotation
- Rate limiting and account lockout (5 failed attempts locks the account)
- OpenTelemetry tracing with Sentry export — every request gets a trace span automatically
- Graceful shutdown with 30-second drain window

**API Endpoints**
- `POST /api/v1/auth/login` — admin/moderator login, sets `access_token` + `refresh_token` cookies
- `POST /api/v1/auth/refresh-token` — silent token rotation
- `POST /api/v1/auth/logout` — clears both cookies server-side
- `POST /api/v1/events/submit` — public submission, auto-creates contributor accounts
- `GET  /api/v1/events` — filterable by year, domain, tier, country, bbox, with pagination
- `POST /api/v1/events/{id}/deadlines` — add deadlines to an approved event
- `PATCH /api/v1/events/{id}/deadlines/{deadlineId}/cancel` — cancel a deadline
- `POST /api/v1/events/{id}/deadlines/{deadlineId}/supersede` — immutable deadline update (creates new, marks old inactive)
- `PATCH /api/v1/admin/events/{id}/review` — approve or reject, writes audit log
- `GET  /api/v1/users/me` — current session user
- `PATCH /api/v1/users/me/password` — change own password
- `PATCH /api/v1/admin/users/{id}/unlock` — unlock a locked account

**Data Model**
- Event — name, slug, location (lat/lng), dates, domain, tier, status (pending/approved/rejected), year index
- Deadline — immutable once created; supersession chain with `is_active` + `superseded_by_id`
- AuditLog — full JSONB diff history on every state change
- User — admin / moderator / contributor roles; password-less contributor accounts

### Frontend

**Globe & Discovery**
- Interactive 3D globe (Globe.gl / WebGL) with event pins coloured by status (upcoming/past/selected)
- 2D Leaflet fallback map for environments without WebGL
- Table view alternative for desktop (toggle between globe and card list)
- Collapsible filter panel — year (>= semantics), domain, tier, country, first-deadline month
- Event detail side panel with deadlines, attribution, and manage-deadlines shortcut
- Deep-link support via `?event=SLUG` query param — globe rotates and opens the event on load

**Event Submission**
- 3-step public wizard: duplicate check → event details → deadlines
- Leaflet map pin picker for exact lat/lng — no geocoding API required
- Contributor account auto-creation on first submission (password-less)

**Deadline Management**
- Single-page deadline manager: add new, supersede existing, cancel — all in one submit
- Supersede history toggle — full chain visible per deadline type

**Admin / Moderator Portal**
- Login page with JWT session (HTTP-only cookies)
- Event review queue with status/tier/year filters and pagination
- 3-step review wizard: edit details → approve/reject decision with notes → manage deadlines
- Moderators cannot review events they submitted themselves
- Update password page with complexity rules
- Language selector (flag button) on every page

**Internationalisation**
- Full i18n with next-intl — English, Portuguese, Spanish, German
- All user-facing strings (including error messages and toast notifications) translated in all 4 locales

**API Client**
- Typed fetch client (`apiRequest` / `apiPrivateRequest`) with automatic token refresh on `TOKEN_EXPIRED`
- Centralised error-to-toast mapping for all backend error codes
- Next.js rewrites proxy all `/api/*` calls through Vercel — eliminates cross-origin cookie restrictions

**SEO**
- Title template, meta description, Open Graph, and Twitter card on all public pages
- hreflang alternates for all 4 locales
- `robots.txt` — disallows all `/*/manage` routes
- `sitemap.xml` — one entry per locale for the globe homepage

### Infrastructure & Deployment
- Frontend: Vercel (Next.js native, edge CDN)
- Backend: Render (always-on Go binary)
- Database: Neon serverless PostgreSQL (pooled connection string via PgBouncer)
- Monitoring: Sentry (error tracking + performance) on both frontend and backend
- Tracing: OpenTelemetry → Sentry on the backend
- Migrations: Goose (sequential SQL files)
- Local dev: Docker Compose (Postgres), Air (Go hot reload), `pnpm dev`

### Known Limitations (planned for v0.2.0)
- Event detail page (`/events/[id]`) is not yet implemented
- Contributor account claiming (setting a password after passwordless creation) is not yet implemented
- Dynamic event pages are not yet in the sitemap
- Domain is currently limited to `computer_science`; the platform is architected to expand to other fields
