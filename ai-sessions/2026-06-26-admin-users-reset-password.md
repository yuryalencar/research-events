# Admin Users — Reset Password Endpoint

**Date:** 2026-06-26
**Spec:** `specs/backend/admin-users-reset-password.yaml`

---

## Goal

Add `PATCH /api/v1/admin/users/{id}/password` — allows an admin to reset any user's
password without knowing the current one. Invalidates the target user's session on success.

---

## Key Decisions

**No current password required** — admin authentication is the gate. This differs from
`PATCH /api/v1/users/me/password` which requires the caller's current password as a
second factor. An admin resetting someone else's password is a privileged action already
gated by `RequireRole("admin")`.

**Complexity checked separately from field validation** — `ValidateAdminResetPasswordInput`
only checks field presence and match (→ 400 `VALIDATION_ERROR`). `ValidatePasswordComplexity`
is called afterward by the handler (→ 422 `PASSWORD_TOO_WEAK`). The split keeps the 400/422
distinction clean without needing typed errors or sentinel values.

**Self-change returns 422 `CANNOT_CHANGE_OWN_PASSWORD`** — admins use
`PATCH /api/v1/users/me/password` for their own password. Using this endpoint against
their own ID is rejected before any DB call, same pattern as `ChangeRole`/`Unlock`.

**`UpdatePassword` already existed** — the repository interface and mock already had this
method from a prior session. No mock regeneration was needed.

**Session invalidation via `ClearTokens`** — same pattern as `ChangeRole`. Failure to
clear tokens is logged but does not abort the response; the password update is the
primary operation.

**Audit diff is empty** — `BuildPasswordChangedAuditLog` writes `action=password_changed`
with no diff payload. Storing the hash would leak credential material; storing "password
changed" without the value is sufficient for the audit trail.

---

## What Was Built

| File | Change |
|---|---|
| `specs/backend/admin-users-reset-password.yaml` | New spec |
| `specs/backend/admin-users-reset-password.curl.sh` | curl examples for all response codes |
| `internal/model/audit_log.go` | Added `AuditActionPasswordChanged = "password_changed"` |
| `internal/service/user.go` | Added `ValidateAdminResetPasswordInput` + `BuildPasswordChangedAuditLog` |
| `internal/service/user_test.go` | 7 pure function tests (no mocks) |
| `internal/handler/admin_user.go` | Added `ResetPassword` handler |
| `internal/handler/admin_user_test.go` | 9 handler tests via gomock |
| `cmd/api/server.go` | Registered `PATCH /api/v1/admin/users/{id}/password` with rate limiter |
| `cmd/api/server_test.go` | 1 route registration test |

---

## Test Count

**Before:** 524 backend tests
**After:** 524 + 17 = 541 backend tests

| Package | Tests added |
|---|---|
| `service` | 7 — `ValidateAdminResetPasswordInput` (5), `BuildPasswordChangedAuditLog` (2) |
| `handler` | 9 — `ResetPassword` (success, invalid ID, self-change, bad JSON, mismatch, weak password, not found, FindByID error, UpdatePassword error) |
| `cmd/api` | 1 — route registration |

All 541 tests pass. `go vet ./...` clean.

---

## State at End

Feature complete and ready for frontend integration. No database migration needed —
no new columns or tables were added. The endpoint uses the existing `users` table
(`password_hash` column) and `audit_logs` table.

---

## Context to Restore

The next step is the **frontend admin user management** feature, which depends on all
four backend endpoints now being live:
- `GET    /api/v1/admin/users`               — list with filters
- `POST   /api/v1/admin/users`               — register admin/moderator
- `PATCH  /api/v1/admin/users/{id}/role`     — change role
- `PATCH  /api/v1/admin/users/{id}/password` — reset password (this session)
- `PATCH  /api/v1/admin/users/{id}/unlock`   — unlock locked account

Frontend spec is in `specs/frontend/admin-user-management.md` (to be written at Phase 0
of the frontend feature). The approved UX layout uses independent sections per action
inside the expanded user card — see the last session in `ai-sessions/` for the layout mockup.
