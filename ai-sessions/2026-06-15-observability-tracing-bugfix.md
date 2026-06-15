# Session: OpenTelemetry Tracing — Sentry Double-Sampling Bug Fix

**Date:** 2026-06-15
**Status:** Complete. Bug fix workflow (Phase 1 → 6), no spec changes needed
beyond a `bug_fix:` entry.

---

## Context

The OpenTelemetry tracing feature (spec'd and planned in the prior session,
`2026-06-14-observability-opentelemetry-planning.md`) had since been fully
implemented: `InitTracerProvider`, `otelhttp` wrapping in `BuildHandler`, the
GORM tracing plugin, and graceful-shutdown flushing were all in place and
`go test ./...` / `go vet ./...` were green (388 tests).

The user reported that after running the backend locally with a real
`SENTRY_DSN` and generating traffic, **nothing appeared in Sentry** — neither
the Issues onboarding screen nor Explore > Traces / Insights > Backend.

---

## Goal

Find and fix why traces weren't reaching Sentry despite the feature appearing
complete and tests passing.

---

## Root cause

`internal/observability/otel.go` called `sentry.Init` with `EnableTracing:
true` but never set `TracesSampleRate` (zero value `0.0`).

The `sentryotel` span processor starts a *Sentry* transaction for every OTel
root span via `sentry.StartTransaction(..., WithSpanSampled(SampledUndefined))`
(no incoming `sentry-trace` header on a fresh request → "undefined" explicit
decision). sentry-go's `(*Span).sample()` then falls through to
`ClientOptions.TracesSampleRate`, sees `0.0`, and silently drops the
transaction — **independently of** the OTel `ParentBased(TraceIDRatioBased(...))`
sampler that already decided to keep the span. Two independent sampling gates
were stacked, and the second was effectively "always reject everything".

---

## Fix

Set `TracesSampleRate: 1.0` in the `sentry.Init` `ClientOptions` (deliberately
**not** `TracesSampleRate(cfg.Env)` — that would compound, e.g.
`0.2 × 0.2 = 0.04` effective rate in production). The OTel sampler remains the
single source of truth for which spans are created; sentry-go now passes
through everything it receives.

Added `TestInitTracerProvider_ValidDSN_InitializesSentryAndRegistersPropagator`
assertion: `sentry.CurrentHub().Client().Options().TracesSampleRate == 1.0`.
Followed Red → Green: test failed (`expected: 1, actual: 0`) before the fix,
passed after.

---

## State at end

- `go test ./...` and `go vet ./...` green (389 tests).
- `specs/backend/observability-opentelemetry.yaml` updated with a `bug_fix:`
  section (symptom, root cause, fix, how to verify).
- Manually verified: backend starts, connects to Postgres, `GET /health`
  returns 200. Network calls to Sentry can't be verified from this sandbox —
  user to confirm traces now appear in Explore > Traces after restarting
  locally and hitting `/health` a few times.

## Out of scope / noted for future

- **Error capture to Sentry Issues** (panic recovery + `CaptureException` via
  `sentryhttp` middleware) is a separate, not-yet-built feature. Currently
  only tracing/performance data is sent — no Issues will appear regardless of
  this fix. Discussed with user; deferred, not scoped into a spec yet.

## Context to restore (next session)

- OTel tracing feature is now feature-complete and the sampling bug is fixed.
- If picking up observability work again, the next item is the error-capture
  middleware (`sentryhttp`) — would need its own Phase 0 spec.

---

## Follow-up: refactor pass

After the bug fix, did a refactor review of the whole observability solution
(`otel.go`, `sentry.go`, `cmd/api/{main,server}.go`, test files, `go.mod`,
`docs/backend-libraries.md`). Four items raised, all actioned:

1. **`sentryotel.NewSentrySpanProcessor` is deprecated** in
   `sentry-go/otel` v0.46.2 — doc comment says "will be removed in 0.47.0,
   prefer `sentryotlp.NewTraceExporter`". Checked: v0.46.2 is the latest
   released version, and `sentryotlp` doesn't exist in any released version
   yet — there is nothing to migrate to today. Added a "Deprecation watch"
   section to `docs/backend-libraries.md` warning against `go get -u` on
   `sentry-go`/`sentry-go/otel` until `sentryotlp` ships.
2. **`go mod tidy`** — `sentry-go`, `sentry-go/otel`,
   `go.opentelemetry.io/otel{,/sdk,/trace}`, `otelhttp`, and
   `gorm.io/plugin/opentelemetry` were directly imported but listed under
   `// indirect`; tidy moved them to the direct `require` block (plus their
   transitive deps). `go build`/`go vet`/`go test ./...` all still green.
3. **Dropped the unused `ctx context.Context` param** from
   `InitTracerProvider` — never read in the body. Updated `main.go` and all
   `otel_test.go` call sites.
4. **Removed `internal/observability/sentry.go`** — empty placeholder
   (`package observability` only), nothing referenced it. Updated the
   `docs/backend-libraries.md` "Where" note accordingly. Also removed
   `.gitkeep` files from directories that already contain real files:
   `ai-sessions/`, `specs/backend/`, `backend/migrations/`,
   `backend/internal/{middleware,repository,config,service}/`.
