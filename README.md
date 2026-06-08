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

### 2. Start the database

```bash
docker compose up -d
```

This starts a PostgreSQL 16 instance on port `5432` with:
- user: `postgres`
- password: `postgres`
- database: `research_events`

### 3. Backend

```bash
cd backend
cp .env.example .env   # fill in your values
go mod download
go run ./cmd/api
```

The API will be available at `http://localhost:8080`.

### 4. Frontend

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
docker compose up -d    # start local Postgres
make migrate-up         # run pending migrations
make migrate-down       # roll back last migration
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

## Specs

| Feature | Spec | Status |
|---------|------|--------|
| Server Bootstrap + Health Check | [server-bootstrap.yaml](specs/backend/server-bootstrap.yaml) | Done |

---

## License

[MIT](./LICENSE) © 2026 Yury Lima
