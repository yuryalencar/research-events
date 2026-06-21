# Session: Eager Token Validation + Auth Redirect Bug Fixes

**Date:** 2026-06-21  
**Goal:** Fix two bugs in the Update Password feature and add eager token validation on all protected /manage pages.

---

## Bug 1 Fixed — `errors.errors.TOKEN_MISSING` double-namespace

`UpdatePasswordCard` passed `useTranslations("errors")` to `useUpdatePassword`.  
`handleApiError` calls `t("errors.CODE")` — combined with the scoped `t` this produced `errors.errors.CODE` which is not a key.

**Fix:** Changed to `const tRoot = useTranslations()` (root-level) in `UpdatePasswordCard`.

---

## Bug 2 Fixed — No redirect when token/refresh token both invalid

When the user is "logged in" (localStorage has `manage_user`) but both the access and refresh tokens have expired, `apiPrivateRequest` throws an auth error (`TOKEN_MISSING`, `TOKEN_EXPIRED`, `TOKEN_INVALID`). Previously `handleApiError` showed a toast but never redirected.

**Fix in `useUpdatePassword`:** Added optional `onAuthError?: () => void` parameter.  
After catching an error, if `AUTH_ERROR_CODES.has(err.code)`, calls `onAuthError?.()`.  
Added 3 tests.

**Fix in `UpdatePasswordCard`:** Added `handleAuthError()` which clears `localStorage.manage_user` and calls `router.replace(/${locale}/manage)`. Passed as `onAuthError` to hook.

---

## Feature — Eager token validation on page load (`useSessionGuard`)

**Problem:** All protected pages only checked `localStorage` on mount. If the localStorage stored user data but the token was invalid, the user saw the page normally until they tried to perform an action.

**Solution:**

### Backend — `GET /api/v1/users/me`
- New handler method `UserHandler.Me` — reads `AuthUser` from JWT context (no DB query)
- Returns `{ code: "SESSION_VALID", data: { id, name, email, role } }`
- Route registered with `RequireAuth` (no rate-limit, one call per page load)
- Handler tests: `TestUserHandler_Me_ReturnsAuthenticatedUser`, `TestUserHandler_Me_NoAuthUser`
- Spec: `specs/backend/users-me.yaml`

### Frontend — `validateSession()` in `lib/api/users.ts`
Calls `GET /api/v1/users/me` via `apiPrivateRequest`. Automatic refresh retry is built into `apiPrivateRequest` — if it throws, both tokens are dead.

### Frontend — `useSessionGuard(requiredRole)` hook
Replaces the duplicated `useEffect` in all 6 protected pages:
1. Read localStorage synchronously → redirect if no user or wrong role
2. Set user immediately (instant display, no spinner)
3. Background call to `validateSession()` → on auth error, clear localStorage + redirect

Applied to:
- `manage/admin/page.tsx`
- `manage/moderator/page.tsx`
- `manage/admin/password/page.tsx`
- `manage/moderator/password/page.tsx`
- `manage/admin/events/[slug]/review/page.tsx` (event guard stays local, hook handles user)
- `manage/moderator/events/[slug]/review/page.tsx`

---

## Test results

- Backend: `go test ./...` — all pass
- Frontend: 277 tests pass, `pnpm typecheck` clean

---

## Context to restore

All frontend changes are uncommitted. The backend changes from this session are also uncommitted (backend `user.go` handler + `server.go` route + `user_test.go` new tests).

Files changed:
- `backend/internal/handler/user.go` — added `Me()` method
- `backend/cmd/api/server.go` — registered `GET /api/v1/users/me`
- `backend/internal/handler/user_test.go` — two new Me tests
- `frontend/src/lib/api/users.ts` — added `SessionUser` type + `validateSession()`
- `frontend/src/hooks/useSessionGuard.ts` — new hook
- `frontend/src/hooks/useUpdatePassword.ts` — added `onAuthError` param + `AUTH_ERROR_CODES`
- `frontend/src/hooks/useUpdatePassword.test.ts` — 3 new auth error tests
- `frontend/src/components/manage/UpdatePasswordCard.tsx` — wired `handleAuthError`
- All 6 protected page files updated to use `useSessionGuard`
- `specs/backend/users-me.yaml` + `users-me.curl.sh` — new spec + curl examples
