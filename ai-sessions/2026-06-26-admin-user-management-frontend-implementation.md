# Admin User Management — Frontend Implementation

**Date:** 2026-06-26
**Spec:** `specs/frontend/admin-user-management.md` (approved)

---

## Goal

Implement the frontend for admin user management:
- `/manage/admin/users` — paginated, filterable user list with expandable cards
- `/manage/admin/users/register` — register a new admin/moderator
- A "Manage Users" link in the admin events dashboard (ManageDashboard)

---

## Decisions Made

### Filters are Apply-gated
Draft filter changes (search, role chips, locked/deleted toggles) never trigger a fetch on their own. `apply()` commits draft to applied state and fetches page 1. `goToPage()` reads `appliedFiltersRef` so in-progress draft edits can't change what Prev/Next navigates through. Same pattern as `useReviewEvents`.

### Ref-based stable callbacks in `useAdminUsers`
`useRef` mirrors latest draft state so all `useCallback` handlers have empty dep arrays and remain stable across renders. Prevents the filter handlers being re-created on every keystroke.

### `extractErrorKey` for hook/i18n decoupling
`useUserCard` and `useRegisterUser` store error i18n key strings (e.g. `"errors.ROLE_UNCHANGED"`) rather than resolved messages. The component calls `t(errorKey)` with its own translator. Keeps hooks independent of next-intl.

### `checkPasswordComplexity` extracted to `lib/utils.ts`
The 4-rule complexity check (minLength, hasUppercase, hasLowercase, hasSpecial) was already implemented inline in `useUpdatePassword`. Extracted as a pure function to `lib/utils.ts` and shared by `useUserCard`, `useRegisterUser`, and `useUpdatePassword`. `PasswordField` and `ComplexityItem` UI components were also extracted from `UpdatePasswordCard` to `components/ui/PasswordField.tsx` for the same reason.

### Three independent card sections
Each of Change Role / Reset Password / Unlock Account has its own apply button, confirmation modal, success banner, and error banner. No shared state between sections. `useUserCard` manages all three sections as independent state machines.

### Optimistic updates in `useUserCard`
After a successful role change, `currentRole` is updated in local state immediately. After a successful unlock, `isLocked` flips to false. The list doesn't re-fetch — acceptable for an admin tool where stale UI is low-risk.

### Filter layout: roles and status separated
Role chips (admin / moderator / contributor) and status chips (Locked / Deleted) are rendered on separate rows with their own labels. Initial implementation merged them on one row — corrected after review.

### Navigation chain (management-only)
`Register → Users list → Events dashboard` — no link from management area back to the public globe. Management area is self-contained.

---

## Files Created

| File | Purpose |
|------|---------|
| `src/lib/utils.ts` | Added `checkPasswordComplexity` + `PasswordComplexity` type |
| `src/lib/utils.test.ts` | 6 tests for `checkPasswordComplexity` |
| `src/lib/api/client.ts` | Added `apiPrivateRequestWithMeta` with TOKEN_EXPIRED refresh |
| `src/lib/api/client.test.ts` | 2 new tests (15 total) |
| `src/lib/api/admin.ts` | Added `listAdminUsers`, `registerAdminUser`, `changeUserRole`, `resetUserPassword` |
| `src/lib/api/admin.test.ts` | 5 new tests (7 total) |
| `src/lib/api/errors.ts` | Added 5 new error codes to `KNOWN_CODES` |
| `src/types/api.ts` | Added 9 new types for user management API |
| `src/components/ui/PasswordField.tsx` | Extracted `PasswordField` + `ComplexityItem` from `UpdatePasswordCard` |
| `src/hooks/useAdminUsers.ts` | Draft-filter + pagination hook for user list |
| `src/hooks/useAdminUsers.test.ts` | 7 tests |
| `src/hooks/useUserCard.ts` | Card state machine: 3 independent sections (role / password / unlock) |
| `src/hooks/useUserCard.test.ts` | 11 tests |
| `src/hooks/useRegisterUser.ts` | Register form state machine |
| `src/hooks/useRegisterUser.test.ts` | 7 tests |
| `src/components/manage/users/UserFilters.tsx` | Filter bar: search + role chips (row 1) + status chips (row 2) + Reset/Apply |
| `src/components/manage/users/UserTableCard.tsx` | Expandable user card with 3 independent sections + inline confirm modals |
| `src/components/manage/users/UserListPage.tsx` | Full list page: filters + card list + pagination |
| `src/components/manage/users/RegisterUserForm.tsx` | Register form + confirm modal + success screen |
| `src/app/[locale]/manage/admin/users/page.tsx` | Admin-only page, guarded by `useSessionGuard("admin")` |
| `src/app/[locale]/manage/admin/users/register/page.tsx` | Admin-only page with `onAuthError` redirect |

## Files Modified

| File | Change |
|------|--------|
| `src/components/manage/ManageDashboard.tsx` | Added admin-only "Manage Users" link (top-right), `Link` import |
| `src/components/manage/UpdatePasswordCard.tsx` | Imports `PasswordField`/`ComplexityItem` from `components/ui/PasswordField` |
| `src/hooks/useUpdatePassword.ts` | Imports `checkPasswordComplexity`/`PasswordComplexity` from `lib/utils` |
| `src/messages/en.json` | Added `manage.users.*` namespace + error codes + `reviewDashboard.manageUsersLink` |
| `src/messages/pt.json` | Same |
| `src/messages/es.json` | Same |
| `src/messages/de.json` | Same |

---

## State at End

All cycles complete, typecheck clean. Committed on `main`. Ready for v0.1.3 release.

- No outstanding test failures
- No TypeScript errors
- All 4 locales have matching keys
