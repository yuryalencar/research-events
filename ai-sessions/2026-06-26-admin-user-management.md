# Admin User Management — Backend

**Date:** 2026-06-26
**Specs:** `specs/backend/admin-users-register.yaml`, `specs/backend/admin-users-change-role.yaml`

---

## Goal

Add two admin-only endpoints for managing platform users:
- `POST /api/v1/admin/users` — register a new admin or moderator with a hashed password
- `PATCH /api/v1/admin/users/{id}/role` — change any user's role and invalidate their session

---

## Key Decisions

**Contributors excluded from registration endpoint** — Contributors self-register via the event submission flow (passwordless). The admin register endpoint only accepts `role: "admin"` or `role: "moderator"`. Passing `"contributor"` returns 400.

**Email conflict includes soft-deleted users** — The `users` table has a full unique index on `email` (not partial). Soft-deleted rows still occupy the slot. `ExistsByEmail` uses GORM's `.Unscoped()` to check across all rows before attempting an insert, returning 409 instead of a DB-level unique constraint error.

**Session invalidation on role change** — On a successful role change, `ClearTokens` is called to null out `access_token_jti`, `access_token_expires_at`, `refresh_token_hash`, and `refresh_token_expires_at`. This forces the user to re-login and receive a JWT carrying the new role.

**Self-change guard fires before DB lookup** — In `ChangeRole`, the check `targetID == admin.ID` (→ 422 CANNOT_CHANGE_OWN_ROLE) is performed immediately after parsing the path parameter, before any repository call. This matches the pattern from the `Unlock` handler.

**`role_changed` with diff** — Audit log uses `action="role_changed"` with `diff={"role":{"before":"<old>","after":"<new>"}}`, keeping the audit trail self-contained without requiring two separate lookups to reconstruct history.

**All role transitions allowed** — contributor ↔ moderator, contributor/moderator → admin, admin → contributor/moderator are all permitted, as long as the admin is not targeting themselves.

**Rate limiting** — Both endpoints share the existing `rateLimiter` (10 req/min/IP), same as login and other admin endpoints.

---

## What Was Built

### `internal/model/audit_log.go`
- Added `AuditActionRoleChanged AuditAction = "role_changed"`

### `internal/repository/user.go`
- Added `ExistsByEmail(ctx, email) (bool, error)` — checks all users including soft-deleted via `.Unscoped()`
- Added `Create(ctx, user) (model.User, error)` — persists a new user and returns it with DB-assigned ID
- Added `UpdateRole(ctx, userID, newRole) error` — updates only the `role` column

### `internal/repository/mocks/mock_user.go`
- Regenerated to include the three new interface methods

### `internal/service/user.go`
- `ValidateRegisterInput(name, email, password, role string) error` — checks all fields, rejects contributor role, delegates to `ValidatePasswordComplexity`
- `BuildRegisterUser(name, email, passwordHash string, role model.UserRole) model.User` — returns a new User struct, no mutation
- `BuildRegisterAuditLog(newUserID, adminID uint, name, email string, role model.UserRole) model.AuditLog` — action=created, diff={name,email,role}
- `ValidateRoleChangeInput(role string) error` — accepts admin/moderator/contributor
- `BuildRoleChangedAuditLog(targetID, adminID uint, oldRole, newRole model.UserRole) model.AuditLog` — action=role_changed, diff={role:{before,after}}

### `internal/handler/admin_user.go`
- `Register` — parse → validate → email check → hash → build → persist → audit
- `ChangeRole` — self-guard → parse role → validate → fetch user → same-role guard → update → clear tokens → audit

### `cmd/api/server.go`
- Registered `POST /api/v1/admin/users` with `rateLimiter + RequireAuth + RequireRole("admin")`
- Registered `PATCH /api/v1/admin/users/{id}/role` with same middleware chain

---

## Test Count

**Before:** 401 backend tests
**After:** 458 backend tests (+57)

New tests breakdown:
- Repository (integration): 7 — `ExistsByEmail` (3), `Create` (2), `UpdateRole` (2)
- Service (pure): 26 — `ValidateRegisterInput` (10), `BuildRegisterUser` (2), `BuildRegisterAuditLog` (4), `ValidateRoleChangeInput` (6), `BuildRoleChangedAuditLog` (4)
- Handler (gomock): 24 — `Register` (11), `ChangeRole` (13, including 6 table-driven transition cases)

---

## State at End

Feature complete. All 458 backend tests pass, `go vet ./...` clean.

No database migration needed — no new columns or tables were added. The two new endpoints use the existing `users` table and `audit_logs` table.

---

## Context to Restore

The frontend management portal already has role management UI sketched out in `specs/frontend/manage-portal.md` — the backend endpoints built in this session are the contracts that portal will call. The next step is to wire the frontend to these two endpoints.
