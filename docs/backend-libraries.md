# Backend Libraries

Every library used in the Go backend — what it does, why it was chosen over alternatives,
and where it appears in the codebase.

---

## Database

### `gorm.io/gorm` + `gorm.io/driver/postgres`

**What it is:** GORM is the most widely used ORM (Object-Relational Mapper) in Go.
An ORM maps Go structs to database tables and generates SQL for you.

**Why it was chosen:**
- Reduces boilerplate for common queries (find by ID, save, list with filters)
- Built-in support for associations (`Event` → `User`, `Event` → `[]Deadline`)
- Auto-migrations and hooks (before/after save, etc.)
- `gorm.Model` embeds `ID`, `CreatedAt`, `UpdatedAt`, `DeletedAt` automatically

**The alternative** would be `database/sql` (stdlib) + raw SQL strings. More control,
zero magic, but much more boilerplate for simple CRUD operations. We use raw SQL only
in `DatabaseChecker` (SELECT 1) where GORM would be overkill.

**Key rule in this project:** always call `.WithContext(ctx)` on every GORM query:
```go
db.WithContext(ctx).Where("status = ?", "pending").Find(&events)
```
This ensures the DB query is cancelled if the HTTP request times out or the client disconnects.

**Where:** `internal/repository/`, `cmd/api/main.go`

---

### `github.com/jackc/pgx/v5`

**What it is:** The PostgreSQL driver for Go. GORM uses it under the hood via
`gorm.io/driver/postgres` — you never call pgx directly, but it must be present.

**Why pgx over `lib/pq`:** pgx is the modern, actively maintained Postgres driver.
`lib/pq` is in maintenance mode. pgx supports newer Postgres features, has better
performance, and is recommended by the GORM docs.

**Where:** `go.mod` (indirect dependency via gorm driver)

---

### `github.com/pressly/goose/v3`

**What it is:** A database migration tool. It runs `.sql` files in
`backend/migrations/` in order, tracking which have already been applied.

**Why migrations matter:** the database schema evolves as features are added.
Without a migration tool, every developer (and every deploy environment) would
need to manually run ALTER TABLE statements. Goose tracks state in a
`goose_db_version` table so each migration runs exactly once.

**Why Goose over alternatives:**
- `golang-migrate` — popular but requires a separate binary or more complex setup
- GORM AutoMigrate — convenient but dangerous in production (can't express all schema changes, e.g. dropping columns, renaming)
- Goose uses plain SQL files — readable, diffable, and not tied to any ORM

**Usage:**
```bash
make migrate-up    # run pending migrations
make migrate-down  # roll back the last migration
```

**Migration naming convention:** `001_create_events.sql`, `002_create_users.sql` — sequential prefix ensures order.

**Where:** `backend/migrations/`

---

## Authentication

### `github.com/golang-jwt/jwt/v5`

**What it is:** A library for creating and verifying JSON Web Tokens (JWT).

**What a JWT is:** a self-contained token that encodes identity claims (e.g. user ID, role)
and is signed with a secret key. The server can verify the token without a database
lookup — the signature proves it was issued by us.

**Why JWT (stateless auth):**
- No server-side session storage needed
- Works well with stateless deploys (Fly.io can run multiple instances)
- Token is stored in an HTTP-only cookie — not accessible to JavaScript (XSS protection)

**Why HTTP-only cookie (not localStorage):**
- localStorage is readable by any JavaScript on the page → XSS risk
- HTTP-only cookies are sent automatically by the browser but are invisible to JS

**Key rule:** JWT is stored in HTTP-only cookie only — never returned in the response body.

**Where:** `internal/middleware/` (JWT auth middleware, added in a future feature)

---

## Configuration

### `github.com/joho/godotenv`

**What it is:** Loads a `.env` file into environment variables at startup.

**Why it exists:** Go's standard library has no built-in `.env` file support.
In production, env vars are set by the platform (Fly.io secrets). In development,
you'd have to `export DATABASE_URL=...` manually before every `go run`. Godotenv
automates this by reading `backend/.env` and calling `os.Setenv` for each line.

**Key behaviour in this project:**
```go
_ = godotenv.Load()  // silently ignored if .env is missing — expected in production
```

**Where:** `internal/config/config.go`

---

## Observability

### `github.com/getsentry/sentry-go`

**What it is:** Sentry's Go SDK. Captures panics, unhandled errors, and slow
requests, then sends them to the Sentry dashboard.

**Why error tracking matters:** in production, errors happen silently — no one
is watching the logs 24/7. Sentry aggregates errors, groups duplicates,
shows stack traces, and can alert on-call engineers.

**What it captures in this project:**
- Panics (automatically via the Sentry middleware)
- Unhandled errors passed to `sentry.CaptureException`
- Slow requests (performance monitoring)

**What it must NOT capture:** passwords, JWT tokens, email addresses — scrubbed
in `BeforeSend`.

**Where:** `internal/observability/sentry.go` (initialised at startup, added in the observability feature)

---

### `go.opentelemetry.io/otel` (+ exporter packages)

**What it is:** OpenTelemetry is the industry standard for distributed tracing.
A "trace" is a tree of "spans" — each span records one operation (HTTP request,
DB query, external call) with timing, attributes, and errors.

**Why tracing matters:** logs tell you what happened; traces tell you *where time was spent*.
If a request takes 500ms, a trace shows you: 2ms parsing, 480ms in a DB query, 18ms encoding.

**How it works in this project:**
- OTel HTTP middleware wraps every handler → each request gets a root span automatically
- Manual spans added for DB queries and operations over 100ms
- Traces exported to Sentry via OTLP (so you see traces and errors in one place)

**Span naming convention:** `resource.action` — e.g. `event.submit`, `deadline.supersede`, `db.query`

**Where:** `internal/observability/otel.go` (added in the observability feature)

---

## Testing

### `github.com/stretchr/testify`

**What it is:** The most popular Go testing library. Adds `assert` and `require`
on top of the standard `testing` package.

**Why not just use `testing` stdlib:**
The stdlib only has `t.Error`, `t.Fatal`, and `t.Log`. Writing assertions manually
is verbose and produces poor failure messages:
```go
// stdlib — tells you "expected true, got false", nothing more
if result.Status != "healthy" {
    t.Errorf("expected healthy, got %s", result.Status)
}

// testify — tells you exactly what differed, including the values
assert.Equal(t, "healthy", result.Status)
// Output: "Not equal: expected 'healthy', actual 'unhealthy'"
```

**`assert` vs `require`:**
- `assert.X` — test continues after failure (collect all failures in one run)
- `require.X` — test stops immediately (use for preconditions where continuing makes no sense)

```go
require.NoError(t, err)          // stop here if err != nil — rest of test is meaningless
assert.Equal(t, "healthy", body.Status)  // continue even if this fails
assert.NotEmpty(t, body.Uptime)
```

**Where:** all `*_test.go` files

---

### `go.uber.org/mock` (mockgen + gomock)

**What it is:** A mock generation library. `mockgen` reads a Go interface and
generates a mock implementation. `gomock` provides the runtime to set expectations
and verify they were met.

**Why mocks matter:**
Service layer tests must not hit a real database — that would be slow, fragile, and
require Postgres to be running. Mocks replace real dependencies with controlled fakes:
```go
mockRepo.EXPECT().FindUserByEmail(gomock.Any(), "new@example.com").Return(model.User{}, ErrNotFound)
```
This says: "when FindUserByEmail is called with any context and this email, return not-found".
The test is deterministic, fast, and database-free.

**Why `go.uber.org/mock` over `github.com/golang/mock`:**
`golang/mock` is no longer maintained. Uber forked it into `go.uber.org/mock` and
continues active development. Same API, same `mockgen` tool.

**Regenerating mocks after interface changes:**
```bash
make generate-mocks  # runs go generate ./...
```
Never hand-write mocks — they go stale and introduce subtle bugs.

**Where:** `internal/health/mocks/`, future `internal/repository/mocks/`, `internal/service/mocks/`

---

## Standard library highlights

These are not third-party libraries but Go stdlib packages worth understanding:

| Package | Purpose | Used in |
|---|---|---|
| `net/http` | HTTP server, router, handlers — no framework needed | All handlers, middleware |
| `log/slog` | Structured JSON logging (Go 1.21+) | `cmd/api/main.go` |
| `context` | Deadlines, cancellation, request-scoped values | Every I/O function |
| `encoding/json` | JSON encode/decode | Handlers (`writeJSON`) |
| `database/sql` | Raw DB interface — GORM wraps this | `DatabaseChecker` |
| `os/signal` | Receive OS signals (SIGTERM, SIGINT) | `cmd/api/main.go` |
| `net/http/httptest` | In-memory HTTP server for tests | All handler tests |
| `time` | Clocks, durations, formatting | Health check, middleware |
