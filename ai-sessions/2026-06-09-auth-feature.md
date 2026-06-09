# Session: Auth Feature (Login, Refresh Token, Logout, Account Unlock)

**Date:** 2026-06-09
**Duration:** Full session

---

## Goal

Implement the complete authentication feature for the backend:
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh-token`
- `POST /api/v1/auth/logout`
- `PATCH /api/v1/admin/users/{id}/unlock`
- JWT auth middleware + role guard
- Per-IP rate limiting
- Users table migration + audit_logs table migration

---

## What was built

### New files
- `specs/backend/auth-login.yaml`, `auth-refresh-token.yaml`, `auth-logout.yaml`, `auth-middleware.yaml`
- `specs/backend/database-users.yaml`, `admin-users-unlock.yaml`
- `specs/backend/auth-jti-design.md` — explains JTI vs full token storage decision
- `specs/backend/go-dependencies.md` — dependency rationale doc
- `backend/migrations/001_create_users.sql`
- `backend/migrations/002_create_audit_logs.sql`
- `backend/internal/model/user.go`, `audit_log.go`
- `backend/internal/repository/user.go`, `audit.go`, `mocks/`
- `backend/internal/service/auth.go` — all pure functions with `// FP:` comments
- `backend/internal/middleware/rate_limit.go`, `auth.go`
- `backend/internal/handler/auth.go`, `admin_user.go`

### Modified files
- `backend/internal/config/config.go` — added required `JWTSecret` field
- `backend/internal/middleware/cors.go` — added `Allow-Credentials: true` + `Vary: Origin`; default origin changed from `*` to `http://localhost:3000`
- `backend/cmd/api/server.go` — new signature `BuildHandler(cfg, db, registry, logger)`; all routes wired
- `backend/cmd/api/main.go` — passes `db` and `logger` to `BuildHandler`
- `docker-compose.yml` — added `postgres_test` service on port 5433
- `Makefile` — added `install-tools`, `migrate-test-up`, `migrate-test-down`

---

## Key design decisions

### Stateful JWT (JTI in DB column)
Each access token contains a `jti` (UUID v4) claim. The server stores only the JTI — not the full token — in `users.access_token_jti`. On every authenticated request, `RequireAuth` verifies the JWT signature locally then checks `jti == DB.access_token_jti` for revocation. See `specs/backend/auth-jti-design.md` for full rationale.

### Refresh token format: `{userID}.{randomHex}`
The refresh token is semi-opaque: `fmt.Sprintf("%d.%s", userID, randomHex64)`. Embedding the user ID avoids a full-table hash lookup on refresh, while the 64-byte random hex provides 256 bits of entropy. The server computes `SHA-256(fullToken)` and stores that as `refresh_token_hash`. This was a deliberate deviation from the original "fully opaque" spec decision — a fully opaque token would have required a `FindByRefreshTokenHash` full-table scan, and token reuse detection (clearing DB tokens on reuse) would have been impossible without knowing the user ID. The trade-off is documented in `service.ParseRefreshTokenUserID`.

### CORS update: `*` → `http://localhost:3000`
HTTP-only cookies cannot be set cross-origin when `Allow-Origin: *` (browser rule: `Allow-Credentials: true` is incompatible with `*`). The CORS default was changed from `*` to `http://localhost:3000`. In production, `CORS_ALLOWED_ORIGINS` env var must be set to the Vercel frontend URL.

### Rate limiter: in-memory token bucket
Used `golang.org/x/time/rate` (token bucket, 10 req/min per IP). In-memory — counters reset on server restart. Acceptable for this scale; a Redis-backed limiter would be needed for multi-instance deployments.

### Account lockout
After 5 consecutive wrong passwords, `locked_at` is set. Only another admin can clear it via `PATCH /api/v1/admin/users/{id}/unlock`. Self-unlock is blocked (422 CANNOT_UNLOCK_SELF). Every unlock writes an `AuditLog` row with `entity_type=user, action=unlocked`.

### Goose CLI: `go install`, not `go run`
The goose CLI binary pulls in MySQL, ClickHouse, and other drivers that aren't in the project's `go.sum`. Using `go install github.com/pressly/goose/v3/cmd/goose@v3.27.1` installs the binary to `GOPATH/bin` outside the module. The Makefile uses `$(GOPATH)/bin/goose`. Run `make install-tools` once per machine.

### OrbStack workspace ↔ Docker networking
The workspace (`localhost`) cannot reach Docker containers directly — Docker runs in the Mac and exposes ports to the Mac's `localhost`. From inside the OrbStack Linux VM (this workspace), Docker containers are reachable via `host.orb.internal`. The test `TestMain` default URL uses `host.orb.internal:5433` instead of `localhost:5433`.

---

## Test count at end of session

| Package | Tests |
|---|---|
| `cmd/api` | 7 |
| `internal/config` | 6 |
| `internal/handler` | 28 |
| `internal/health` | 7 |
| `internal/middleware` | 18 |
| `internal/repository` | 14 (integration, needs `docker compose up -d`) |
| `internal/service` | 14 |
| **Total** | **94** |

All passing. `go vet ./...` zero warnings. Clean build.

---

## State at end of session

The auth feature is fully implemented and tested. The server boots, connects to DB, and serves all auth endpoints. To run the full stack locally:

```bash
make install-tools          # once per machine — installs goose binary
docker compose up -d        # starts postgres (5432) and postgres_test (5433)
make migrate-up             # applies migrations to dev DB
make migrate-test-up        # applies migrations to test DB
go test ./...               # all 94 tests green
cd backend && go run ./cmd/api  # server on :8080
```

Next feature to build: event submission (`POST /api/v1/events/submit`) — spec not yet written.

---

## Context to restore

- `JWT_SECRET` env var is now required at startup — `.env` must include it
- `CORS_ALLOWED_ORIGINS` default is now `http://localhost:3000` (was `*`) — frontend dev server on 3000 works without extra config
- `BuildHandler` signature changed from `(cfg, registry)` to `(cfg, db, registry, logger)`
- Refresh token format is `{userID}.{randomHex}` — not a JWT, not fully opaque
- `service/` functions are all pure — side effects (DB, random) live in handlers
- `internal/repository/testmain_test.go` uses `host.orb.internal:5433` as default test DB URL
