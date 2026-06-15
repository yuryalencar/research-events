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

**Where:** `internal/observability/otel.go` — `sentry.Init` is called from
`InitTracerProvider` (non-empty `SentryDSN` path) so the error-reporting
client and the tracing span processor share one initialization step. Panic
recovery / `sentry.CaptureException` / `BeforeSend` scrubbing are not yet
implemented — that's a separate future feature.

---

### `go.opentelemetry.io/otel`

**What it is:** the vendor-neutral OpenTelemetry tracing API. OpenTelemetry is
the industry standard for distributed tracing. A "trace" is a tree of "spans"
— each span records one operation (HTTP request, DB query, external call)
with timing, attributes, and errors.

**Why tracing matters:** logs tell you what happened; traces tell you *where
time was spent*. If a request takes 500ms, a trace shows you: 2ms parsing,
480ms in a DB query, 18ms encoding.

**What we use it for:** `otel.SetTextMapPropagator(...)` — registers how trace
context (trace ID, span ID, sampling decision, baggage) is encoded onto/decoded
from HTTP headers so a trace stays connected across services.

**How it works in this project:**
- `InitTracerProvider` builds a `*sdktrace.TracerProvider` once at startup
- OTel HTTP middleware (`otelhttp`, added in a later cycle) wraps every
  handler → each request gets a root span automatically
- Manual spans/GORM plugin spans added for DB queries and operations over 100ms
- Finished spans are forwarded to Sentry via the `sentryotel` bridge (see below)
  — **not** via an OTLP exporter

**Span naming convention:** `resource.action` — e.g. `event.submit`, `deadline.supersede`, `db.query`

**Where:** `internal/observability/otel.go`

---

### `go.opentelemetry.io/otel/sdk/trace` (`sdktrace`)

**What it is:** the SDK implementation of the OTel tracing API above.

**What it provides:**
- `TracerProvider` — the factory every tracer/span is created from; owns the
  sampler and span processors
- `ParentBased` / `TraceIDRatioBased` — composable samplers. `ParentBased`
  means "if this span has a parent, inherit its sampling decision; otherwise
  apply the wrapped sampler" — this keeps a whole trace either fully sampled
  or fully dropped, never half-and-half. `TraceIDRatioBased(rate)` samples a
  fraction (`rate`) of new traces, derived deterministically from the trace ID
  so the decision is consistent across services.
- `SpanProcessor` — interface invoked when a span starts/ends; this is the
  extension point used to forward spans to Sentry (`sentryotel.NewSentrySpanProcessor()`)

**Where:** `internal/observability/otel.go` (`InitTracerProvider`, `TracesSampleRate`)

---

### `github.com/getsentry/sentry-go/otel` (`sentryotel`)

**What it is:** the official bridge between OpenTelemetry and Sentry's
performance/tracing product, maintained by Sentry.

**What it provides:**
- `NewSentrySpanProcessor()` — an `sdktrace.SpanProcessor` that forwards every
  finished OTel span into Sentry as a transaction/span
- `NewSentryPropagator()` — a `propagation.TextMapPropagator` that understands
  Sentry's `sentry-trace` and `baggage` headers, so a trace stays linked when
  it crosses a service that uses Sentry's own SDKs

**Why this instead of an OTLP exporter:** Sentry's OTLP ingestion requires
manually deriving an OTLP endpoint + `x-sentry-auth` header from the DSN — an
undocumented, fragile path. `sentryotel` is the Sentry-documented integration:
spans go through the same `sentry.Init`-configured client used for error
reporting, so traces and errors share one pipeline.

**Deprecation watch:** as of `sentry-go/otel` v0.46.2 (the latest released
version, and what this project pins), `NewSentrySpanProcessor` carries a
`Deprecated:` doc comment pointing at a future `sentryotlp.NewTraceExporter`
and says it "will be removed in 0.47.0". That package does not exist in any
released version yet, so there is nothing to migrate to today. **Do not run
`go get -u` on `github.com/getsentry/sentry-go` / `github.com/getsentry/sentry-go/otel`
without checking whether `sentryotlp` has shipped and re-reading this section**
— a bump to 0.47.0 as-is would remove the function this integration depends on.

**Where:** `internal/observability/otel.go` (`InitTracerProvider`, non-empty `SentryDSN` path)

---

### `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp`

**What it is:** OTel's official instrumentation for `net/http`. `otelhttp.NewHandler`
wraps an `http.Handler` so every incoming request automatically gets a root span.

**What it provides:**
- A root span per request, named from the matched route (Go 1.22+'s
  `http.ServeMux` sets `r.Pattern` after dispatch, e.g. `"GET /health"`, and
  otelhttp uses that to rename the span)
- Standard HTTP attributes (`http.request.method`, `http.response.status_code`, etc.)
- `otelhttp.WithTracerProvider(tp)` — wires the handler to our `InitTracerProvider` output

**Course correction:** the original plan assumed `otelhttp.WithRouteTag` would set
an `http.route` span attribute. That option does not exist in v0.69.0 — only the
span *name* is derived from `r.Pattern`, not an attribute. We added a small
`traced` middleware (`cmd/api/server.go`) that runs after the mux has dispatched
the request and sets `http.route` explicitly from `r.Pattern`.

**Where:** `cmd/api/server.go` (`BuildHandler` — outermost layer, wraps the whole mux)

---

### `gorm.io/plugin/opentelemetry/tracing`

**What it is:** GORM's official OpenTelemetry plugin. Registered once via
`db.Use(...)`, it hooks into GORM's callback chain (`Create`, `Query`, `Update`,
`Delete`, `Row`, `Raw`) and emits a child span for every query.

**What it provides:**
- A span per query, named `"{operation} {table}"` (e.g. `"select events"`),
  with `db.operation.name`, `db.query.text`, and `db.collection.name` attributes
- `trace.SpanKindClient` spans, automatically parented to whatever span is
  already in the query's `context.Context` — i.e. the per-request root span
  from `otelhttp`
- `tracing.WithoutQueryVariables()` — **required** in this project. Without it,
  the plugin interpolates bound parameters (e.g. a submitter's email address)
  directly into the `db.query.text` attribute, which would leak PII into traces.
  With it, only the parameterized SQL (`?` placeholders) is recorded.

**Tracer provider resolution:** if no `tracing.WithTracerProvider(...)` option
is passed, the plugin calls `otel.GetTracerProvider()` *once*, at registration
time. Because OTel's global provider is a *delegating proxy*, a later call to
`otel.SetTracerProvider(tp)` retroactively redirects the plugin's spans to `tp`
— this is why `main.go` calls `otel.SetTracerProvider(tp)` before `db.Use(...)`
is even relevant, and why tests can register the plugin once in `TestMain` and
still redirect its output per-test via `otel.SetTracerProvider`.

**Where:** `cmd/api/main.go` (registered on the real `db` after the startup
ping), `cmd/api/testmain_test.go` (registered on `testDB` for integration tests)

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
