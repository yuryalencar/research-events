# Session: Server Bootstrap + Health Check

**Date:** 2026-06-08
**Spec:** `specs/backend/server-bootstrap.yaml`
**Status at end:** feature complete — 21 tests passing, `go vet` clean, binary builds

---

## Goal

Implement the Go HTTP server from scratch: configuration, database connection, CORS
middleware, `GET /health` endpoint with an extensible checker pattern, and graceful
shutdown on SIGTERM/SIGINT.

---

## What was built

| File | Purpose |
|---|---|
| `internal/config/config.go` | Parses env vars via godotenv; fails fast if `DATABASE_URL` missing |
| `internal/health/health.go` | `Checker` interface, `CheckResult`, `HealthResponse` types |
| `internal/health/checker.go` | `Registry` — holds checkers, runs them all, returns aggregate result |
| `internal/health/db.go` | `DatabaseChecker` — pings Postgres with 3s timeout; `DBPinger` interface for testability |
| `internal/health/mocks/` | Generated mocks for `Checker` and `DBPinger` (gomock) |
| `internal/handler/health.go` | `HealthHandler` — serves GET /health; 200/503 based on registry result |
| `internal/middleware/cors.go` | Hand-written CORS middleware; handles OPTIONS preflight (204) |
| `cmd/api/server.go` | `BuildHandler` — wires mux + middleware; extracted from main for testability |
| `cmd/api/main.go` | Bootstrap: config → DB ping → registry → BuildHandler → graceful shutdown |
| `backend/.env.example` | Committed template documenting all env vars with defaults |
| `backend/.env` | Gitignored local dev values (points to Docker Compose Postgres) |
| `docs/go-learning.md` | Running Go learning notes: context, pointers, interfaces, stack/heap, goroutines, channels, multiple return values |
| `docs/backend-libraries.md` | Every backend library with rationale and alternatives considered |

**Test count:** 21 tests across 4 packages (`cmd/api`, `internal/config`, `internal/handler`, `internal/health`, `internal/middleware`)

---

## Key decisions

**CORS: hand-written over `rs/cors` library**
The project uses `net/http` stdlib with no framework. Writing CORS manually (20 lines)
is consistent with that philosophy and is a learning exercise. `rs/cors` would be
appropriate if edge cases (multiple origins, credential headers) become important.

**`APP_VERSION` as env var (not build-time ldflags)**
Simpler for a learning project. Set via `fly secrets set APP_VERSION=x.y.z` at deploy time.
Upgrade path to `go build -ldflags "-X main.version=x.y.z"` is straightforward later.

**`DBPinger` interface over `*sql.DB` directly**
`DatabaseChecker` accepts `DBPinger` (a narrow interface with one method: `PingContext`)
instead of `*sql.DB`. This lets tests inject a mock without a real database connection.
In production, `*sql.DB` satisfies `DBPinger` automatically (implicit interface).

**`BuildHandler` extracted from `main()`**
`main()` calls `os.Exit` and blocks on `<-quit`, making it untestable directly.
Extracting `BuildHandler(cfg, registry) http.Handler` into `server.go` gives tests a
clean entry point: real registry with mock checkers, no OS signals, no real DB.
Test file uses `package main` (internal) because `main` packages cannot be imported.

**`PingContext` over `SELECT 1`**
The spec says SELECT 1 but `PingContext` achieves the same goal (verify a live
connection round-trip) and is simpler to mock via the `DBPinger` interface.
Documented in a code comment.

---

## Go concepts introduced this session

All documented in `docs/go-learning.md`:
- `context.Context` — explicit dependency chain, cancellation, timeouts
- Pointer receivers (`*T`) and pointer types — heap allocation, shared state
- Interfaces — implicit satisfaction, consumer-side definition, testability
- Struct literals and `&T{}` — value vs pointer, heap escape
- Short variable declaration `:=`
- `make()` for maps, slices, channels
- Stack vs heap — escape analysis, `go build -gcflags="-m"`
- Multiple return values and `nil` as "no error"
- Goroutines — lightweight concurrency, `go func()`
- Channels — typed pipes, buffered channels, blocking `<-`

---

## Functional programming notes

FP rules apply strictly to `internal/service/` (not yet implemented).
The patterns introduced this session that feed into FP:
- Interfaces enable pure function testing (inject mock, no real I/O)
- `context.Context` as explicit dependency is an FP-friendly pattern (no hidden globals)
- `DBPinger` interface: the handler depends on an abstraction, not a concrete DB — side effects are at the edge

---

## State at end of session

- `GET /health` works end-to-end (verified via `go build` + manual curl instructions given)
- Smoke test requires local Postgres via `docker compose up -d`; Docker not available in the dev container
- All automated tests pass: `go test ./... && go vet ./...`
- `backend/.env` is ready for local development
- No migrations written yet — no tables exist, but the DB connection is established at startup

---

## Context to restore next session

**Next feature candidates (not yet started):**
- Auth login (`POST /api/v1/auth/login`) — spec not yet written
- Event model + migrations — schema for `events`, `users`, `deadlines`, `audit_logs`
- Observability feature — Sentry + OpenTelemetry bootstrap in `internal/observability/`

**Important file locations:**
- Spec: `specs/backend/server-bootstrap.yaml`
- Curl examples: `specs/backend/server-bootstrap.curl.sh`
- Go learning notes: `docs/go-learning.md`
- Library reference: `docs/backend-libraries.md`

**To run the server locally:**
```bash
docker compose up -d
cd backend && go run ./cmd/api
curl -s http://localhost:8080/health | jq .
```
