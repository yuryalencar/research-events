# Go Backend — Direct Dependencies

All packages added to `go.mod` with rationale. Update this file whenever a new direct dependency is introduced.

---

## Production dependencies

| Package | Version | Why |
|---|---|---|
| `gorm.io/gorm` | v1.31+ | ORM — model definitions, query builder, soft deletes, migrations DSL |
| `gorm.io/driver/postgres` | v1.6+ | GORM Postgres driver — uses pgx v5 under the hood |
| `github.com/joho/godotenv` | v1.5+ | Loads `.env` file into `os.Environ` at startup — dev convenience, no-op in production |
| `github.com/golang-jwt/jwt/v5` | v5.3+ | Signs and verifies HS256 JWTs for stateful access tokens. v5 is the current stable; earlier versions had security advisories. |
| `github.com/google/uuid` | v1.6+ | Generates UUID v4 for the JWT `jti` (token ID) claim. Used to uniquely identify each token issuance for revocation checks. |
| `github.com/pressly/goose/v3` | v3.27+ | SQL migration runner — sequential numbered files, up/down support, Postgres dialect. Referenced in `Makefile` (`make migrate-up`). |
| `golang.org/x/time` | v0.15+ | Provides `rate.Limiter` (token bucket algorithm) for per-IP rate limiting on auth endpoints. Token bucket is smoother than a fixed-window counter and prevents burst abuse at window boundaries. |

## Test-only dependencies

| Package | Version | Why |
|---|---|---|
| `github.com/stretchr/testify` | v1.11+ | `assert` (continues after failure) and `require` (stops on failure) — cleaner test assertions than raw `t.Error`/`t.Fatal` |
| `go.uber.org/mock` | v0.6+ | gomock — generates type-safe mocks from interfaces via `mockgen`. Used to mock repository interfaces in service and handler tests without hitting a real database. |

---

## Indirect dependencies (selected — managed by `go mod tidy`)

These are pulled in transitively and do not need manual management. Notable ones:

| Package | Pulled in by |
|---|---|
| `github.com/jackc/pgx/v5` | `gorm.io/driver/postgres` |
| `golang.org/x/crypto` | `gorm.io/driver/postgres` (bcrypt via pgx), also used directly for `bcrypt.CompareHashAndPassword` |
| `golang.org/x/sync` | `goose/v3` |

---

## Decisions log

**Why `golang-jwt/jwt/v5` instead of hand-rolling JWT?**
JWT has subtle security requirements (algorithm confusion, signature verification ordering). A battle-tested library is always preferred over a custom implementation for auth primitives.

**Why `google/uuid` instead of `crypto/rand` directly?**
UUID v4 is the standard format for `jti` claims (RFC 7519). `google/uuid` handles the version bits and formatting correctly. `crypto/rand` + hex encoding would work but loses the semantic meaning and standard format.

**Why `goose` instead of GORM AutoMigrate?**
AutoMigrate is convenient for development but cannot handle destructive changes, renaming, or data migrations safely. Goose gives explicit, versioned, reviewable SQL that runs the same in dev and production.

**Why `golang.org/x/time/rate` instead of a third-party rate-limit library?**
`rate.Limiter` (token bucket) is maintained by the Go team, has no transitive dependencies, and is sufficient for per-IP in-memory limiting. Third-party libraries add operational complexity without benefit at this scale.
