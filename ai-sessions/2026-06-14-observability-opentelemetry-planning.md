# Session: OpenTelemetry Tracing (backend) — Spec + Plan (PARTIAL)

**Date:** 2026-06-14
**Status:** Phase 0 (Spec) and Phase 2 (Plan) complete. **Phase 3 (Red) not started.**
This is a checkpoint, not a finished feature — resume at Phase 3 (or earlier if
the plan needs changes) in the next session.

---

## Goal

Add distributed tracing to the Go backend: every HTTP request gets a root
OTel span (via `otelhttp`), every GORM query gets a child span (via the GORM
OTel tracing plugin), and finished spans are exported to Sentry via the
`sentry-go/otel` (`sentryotel`) bridge. Cross-cutting infra feature, no new
endpoints or response shapes. The user is new to OTel — Claude is acting as
**teacher** throughout, explaining concepts inline as code is written.

---

## Spec (approved)

`/workspace/specs/backend/observability-opentelemetry.yaml` — **approved,
final version**. Key points:

- Export target: **Sentry**, via `github.com/getsentry/sentry-go` +
  `github.com/getsentry/sentry-go/otel` (package `sentryotel`) — **NOT**
  `otlptracehttp` / manual OTLP endpoint derivation (see "course correction"
  below).
- Instrumentation scope for v1: **HTTP requests + DB queries only** (no
  manual spans, no error-capture middleware — that's a future spec).
- `SENTRY_DSN` env var, **optional**. Empty → no-op tracer, `sentry.Init`
  never called, app runs normally. Non-empty + malformed → `InitTracerProvider`
  returns an error → `main.go` fatal-exits (same as bad `DATABASE_URL`).
- Sampling: `ParentBased(TraceIDRatioBased(rate))`, `rate = 1.0` when
  `cfg.Env == "development"`, else `0.2`. Never 0.
- `service.name = "research-events-api"` (constant, not env-configurable),
  `service.version = cfg.AppVersion`, `deployment.environment = cfg.Env`.
- GORM plugin registered with `tracing.WithoutQueryVariables()` — **required**
  to avoid leaking PII (emails etc.) into span attributes.
- Spec includes a step-by-step "create a Sentry project" walkthrough in its
  `setup` section (platform = Go, project name = `research-events-api`,
  enable Tracing/Performance).

### Course correction during Phase 1 (important — do not redo)
Original spec drafted an `otlptracehttp` exporter with a manually-derived
Sentry OTLP endpoint + `x-sentry-auth` header. User researched and confirmed
the correct, Sentry-documented path is the **`sentryotel` bridge**:
`sentry.Init(sentry.ClientOptions{Dsn: ...})` +
`sdktrace.WithSpanProcessor(sentryotel.NewSentrySpanProcessor())` +
`otel.SetTextMapPropagator(sentryotel.NewSentryPropagator())`. The spec was
fully revised to this approach and re-approved. **The final spec file already
reflects this — no OTLP exporter dependency should be added.**

---

## Env files (already updated)

- `backend/.env` (gitignored) — has a real `SENTRY_DSN`:
  `https://cc02869e8a3aa823b16ced9d1beca69f@o4511565147406336.ingest.de.sentry.io/4511565187317840`
  and `TEST_DATABASE_URL=postgres://postgres:postgres@localhost:5433/research_events_test?sslmode=disable`
- `backend/.env.example` — has `TEST_DATABASE_URL` (real, needed for
  `go test ./...`) and a placeholder/simulated `SENTRY_DSN`. **Known minor
  cleanup pending**: the `SENTRY_DSN` comment in `.env.example` still says
  "exported to Sentry via OTLP" — stale wording from before the sentryotel
  revision. Planned as Cycle 8 cleanup (see plan below).

---

## Phase 2 Plan (presented, awaiting "proceed to Phase 3" approval)

### Interfaces / signatures to add or change

```go
// internal/config/config.go
type Config struct {
    // ...existing fields...
    SentryDSN string // new, optional, "" = no-op tracer
}
// Load() adds: SentryDSN: getEnv("SENTRY_DSN", "")  — no validation

// internal/observability/otel.go (currently empty)
func TracesSampleRate(env string) float64
func InitTracerProvider(ctx context.Context, cfg config.Config) (*sdktrace.TracerProvider, error)

// cmd/api/server.go
func BuildHandler(cfg config.Config, db *gorm.DB, registry *health.Registry, logger *slog.Logger, tp trace.TracerProvider) http.Handler
func traced(pattern string, h http.Handler) http.Handler // otelhttp.WithRouteTag wrapper, applied per mux.Handle call
```

`cmd/api/main.go` wiring (no new exported signatures): call
`InitTracerProvider`, `otel.SetTracerProvider(tp)`, register
`db.Use(tracing.NewPlugin(tracing.WithoutQueryVariables()))`, pass `tp` into
`BuildHandler`, and add `tp.Shutdown(ctx)` + conditional `sentry.Flush(...)`
(shared 5s bound, warn-on-timeout) to graceful shutdown.

### Test list

- `internal/config/config_test.go`: `Load_SentryDSNDefaultsToEmptyWhenUnset`,
  `Load_SentryDSNPassedThroughFromEnv`
- `internal/observability/otel_test.go` (new, `package observability_test`):
  - `TestTracesSampleRate_DevelopmentReturns1_0`
  - `TestTracesSampleRate_ProductionReturns0_2`
  - `TestTracesSampleRate_UnknownEnvReturns0_2`
  - `TestInitTracerProvider_EmptyDSN_ReturnsProviderWithoutCallingSentryInit`
  - `TestInitTracerProvider_ValidDSN_InitializesSentryAndRegistersPropagator`
    (uses a syntactically valid but fake DSN, e.g.
    `https://abc123@o0.ingest.de.sentry.io/1` — no network needed)
  - `TestInitTracerProvider_MalformedDSN_ReturnsError`
- `cmd/api/server_test.go`: update all 9 existing `BuildHandler(...)` calls
  to pass a `tp`; add `TestBuildHandler_RequestProducesRootSpanWithHTTPMethodAndRoute`
  (uses `tracetest.NewInMemoryExporter()`)
- `cmd/api/testmain_test.go` (new) — mirrors
  `internal/repository/testmain_test.go` (connects to `TEST_DATABASE_URL`,
  runs goose migrations, exposes `testDB` + `beginTx`)
- `cmd/api/server_db_test.go` (new):
  - `TestBuildHandler_DBQueryProducesChildSpanUnderHTTPSpan`
  - `TestBuildHandler_DBSpanDoesNotLeakRawQueryParameters`

### File list

`internal/config/config.go`, `internal/config/config_test.go`,
`internal/observability/otel.go`, `internal/observability/otel_test.go`,
`cmd/api/server.go`, `cmd/api/server_test.go`, `cmd/api/testmain_test.go`,
`cmd/api/server_db_test.go`, `cmd/api/main.go`, `backend/.env.example`,
`go.mod`/`go.sum`.

New deps: `go.opentelemetry.io/otel`, `go.opentelemetry.io/otel/sdk`,
`go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp`,
`gorm.io/plugin/opentelemetry/tracing`, `github.com/getsentry/sentry-go`,
`github.com/getsentry/sentry-go/otel`.

### Cycle breakdown

1. **Config** — `SentryDSN` field (2 tests)
2. **`TracesSampleRate`** — pure function (3 tests) — good FP/TDD warm-up
3. **`InitTracerProvider` no-op path** — empty DSN (1-2 tests)
4. **`InitTracerProvider` sentryotel + error path** — valid/malformed DSN (2 tests)
5. **`BuildHandler` otelhttp wrap** — HTTP root span (signature change + 1 new test + 9 existing calls updated)
6. **GORM tracing plugin** — DB child span + no-PII-leak (2 tests, needs `testmain_test.go`)
7. **`main.go` wiring + graceful shutdown** — no new Red phase (integration glue, verified via `go build`/`go vet` + manual Sentry check per spec setup step 8)
8. **Docs cleanup** — fix stale "via OTLP" wording in `.env.example`

After Cycle 8: full `go test ./... && go vet ./...`, confirm all border
cases / DoD items covered, then Phase 6.

---

## Context to restore (next session)

- Spec is fully approved — do not re-litigate sentryotel-vs-OTLP.
- Plan above was presented but **not yet approved** — first message back
  should re-confirm or adjust the plan, then ask "proceed to Phase 3?" before
  any test code is written.
- `internal/observability/otel.go` and `sentry.go` are still empty
  (`package observability` only).
- Backend is otherwise feature-complete for v1 (377 tests green as of the
  prior admin-review session) — this OTel feature is the last planned v1
  item.
- Keep the "act as teacher" framing — explain new OTel/Go concepts inline in
  code comments and in chat as each cycle introduces them.
