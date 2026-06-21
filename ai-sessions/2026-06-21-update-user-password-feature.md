# Session: Update User Password — Backend

**Date:** 2026-06-21
**Feature:** `PATCH /api/v1/users/me/password`
**Spec:** `specs/backend/users-update-password.yaml`

---

## Goal

Add an authenticated endpoint allowing any logged-in user (admin or moderator) to change their own password. Contributors are excluded for now — they have no JWT until the future account-claim flow is implemented.

---

## Decisions made

- **Endpoint path:** `/api/v1/users/me/password` — `me` scopes it to the requester's own account; no user ID in the path prevents changing another user's password.
- **Validation order:** fields present → passwords match → complexity → DB lookup → current password verified → hash → persist. DB is only touched when all input is already valid.
- **No AuditLog entry:** AuditLog tracks event and deadline entities only. Password changes are user mutations, not domain events.
- **Same rate limit as login (10/min/IP):** both are brute-force targets; consistent limit avoids a weaker backdoor.
- **Rate limiter wraps RequireAuth** in server.go: unauthenticated requests still consume a token, preventing rate-limit bypass via missing JWT.
- **`ValidateCredentials` reused** from auth service — same bcrypt compare logic, no duplication.
- **bcrypt cost 12** — same as the rest of the auth flow.
- **400 (not 401) for wrong current password** — the user is authenticated; wrong current password is an input validation failure, not an auth failure.

---

## Files created or modified

| File | Change |
|---|---|
| `internal/service/user.go` | New — `ValidateUpdatePasswordInput`, `ValidatePasswordComplexity`, `HashPassword` (all FP: pure / no side effects) |
| `internal/service/user_test.go` | New — 14 tests covering all three service functions including bcrypt cost 12 assertion |
| `internal/repository/user.go` | Added `UpdatePassword(ctx, userID, newHash)` to interface + implementation |
| `internal/repository/mocks/mock_user.go` | Regenerated — includes `UpdatePassword` mock |
| `internal/repository/user_test.go` | Added 2 integration tests: `UpdatePassword_PersistsNewHash` and `UpdatePassword_NewHashVerifiableWithBcrypt` |
| `internal/handler/user.go` | New — `UserHandler` + `UpdatePassword` handler |
| `internal/handler/user_test.go` | New — 13 handler tests (success, 401, malformed JSON, missing fields ×3, mismatch, complexity ×4, wrong current password) |
| `cmd/api/server.go` | Registered `PATCH /api/v1/users/me/password` with rate limiter + RequireAuth |
| `cmd/api/server_test.go` | Added `TestBuildHandler_UpdatePasswordRateLimited_Returns429AfterBurst` |
| `specs/backend/users-update-password.yaml` | All DoD items checked |
| `specs/backend/users-update-password.curl.sh` | Generated with curl examples for all response codes |

---

## FP concepts introduced

- **`ValidateUpdatePasswordInput`** — `// FP: pure function` — depends only on its three string arguments; no I/O, no globals. Returns the same result for the same input every time.
- **`ValidatePasswordComplexity`** — `// FP: pure function` — iterates over runes, accumulates boolean flags, returns an error or nil. No state beyond its arguments.
- **`HashPassword`** — `// FP: no side effects` — only observable effect is the returned hash; no mutation of external state, no logging, no DB call.

---

## Test coverage

401 tests total after this session (up from 377 before the globe/table toggle and this feature).

All 14 DoD items checked with specific tests:

| DoD item | Test | Layer |
|---|---|---|
| 200 PASSWORD_UPDATED | `TestUserHandler_UpdatePassword_Success` | handler |
| 400 missing/empty fields (×3) | `TestUserHandler_UpdatePassword_MissingFields` | handler |
| 400 passwords don't match | `TestUserHandler_UpdatePassword_PasswordsMismatch` | handler |
| 400 complexity: too short | `TestUserHandler_UpdatePassword_ComplexityFailures/too_short` | handler |
| 400 complexity: no uppercase | `TestUserHandler_UpdatePassword_ComplexityFailures/no_uppercase` | handler |
| 400 complexity: no lowercase | `TestUserHandler_UpdatePassword_ComplexityFailures/no_lowercase` | handler |
| 400 complexity: no special | `TestUserHandler_UpdatePassword_ComplexityFailures/no_special_character` | handler |
| 400 INVALID_CURRENT_PASSWORD | `TestUserHandler_UpdatePassword_WrongCurrentPassword` | handler |
| 401 no JWT | `TestUserHandler_UpdatePassword_NoAuthUser` | handler |
| 429 rate limit | `TestBuildHandler_UpdatePasswordRateLimited_Returns429AfterBurst` | server |
| bcrypt cost 12 | `TestHashPassword/hash_is_produced_at_bcrypt_cost_12` | service |
| new password verifiable with bcrypt | `TestUserRepository_UpdatePassword_NewHashVerifiableWithBcrypt` | repository |
| rate limit 10/min | `TestBuildHandler_UpdatePasswordRateLimited_Returns429AfterBurst` | server |
| all border cases covered | all of the above | — |

---

## Bug found and fixed during Phase 6

The integration test database (port 5433) had 8 committed events and 2 users from previous development sessions. These caused `ListEvents_*` and `FindDeadlineByID_*` repository tests to fail — tests expecting 1 row found 9. Root cause: the transaction-isolation strategy (begin tx + rollback) protects tests from each other but not from pre-existing committed rows. Fix: `TRUNCATE events, users, deadlines, audit_logs RESTART IDENTITY CASCADE` against the test DB. All repository tests now pass cleanly.

---

## State at end of session

- Feature fully implemented and tested (Phase 3–5 complete)
- All 401 tests green across all packages
- Test DB clean
- Spec DoD fully checked
- Phase 6 docs written

## Context to restore next session

- Next backend feature to build: TBD
- No outstanding uncommitted work
- Test DB is clean — do not seed it with real-looking events to avoid repeating this issue
