# Management Portal — Frontend

**Date:** 2026-06-19
**Phases completed:** 0 (Spec) → 1 (Discovery) → 2 (Plan) → 3 (Red) → 4 (Green) → 5 (Refactor) → 6 (Docs)

---

## Goal

Add a hidden management section at `/manage` — not linked from the homepage, accessible only by direct URL. Provides login for admin and moderator roles, decodes the JWT client-side to extract user info, and routes each role to its own welcome dashboard.

---

## Files Created

| File | Description |
|------|-------------|
| `src/lib/utils/decodeJwt.ts` | Pure function — decodes JWT payload via `atob()`, no library, throws on malformed input |
| `src/lib/utils/decodeJwt.test.ts` | 5 unit tests: 2 happy paths (admin/moderator claims) + 3 malformed-input cases |
| `src/app/[locale]/manage/page.tsx` | Login page — form with email/password (eye toggle), field validation, API error display, moderator banner, localStorage session check on mount |
| `src/app/[locale]/manage/admin/page.tsx` | Admin dashboard — localStorage session read, role guard, welcome heading + role badge + logout |
| `src/app/[locale]/manage/moderator/page.tsx` | Moderator dashboard — same pattern, role guard flipped |

## Files Modified

| File | Change |
|------|--------|
| `src/lib/api/auth.ts` | Added `credentials: "include"` to `login()` so browser stores HTTP-only cookies from cross-origin login response |
| `src/lib/api/auth.test.ts` | Updated `login` test to assert `credentials: "include"` |
| `src/messages/en.json` | Added `manage.*` i18n keys |
| `src/messages/pt.json` | Added `manage.*` i18n keys (Portuguese) |
| `src/messages/es.json` | Added `manage.*` i18n keys (Spanish) |
| `src/messages/de.json` | Added `manage.*` i18n keys (German) |

---

## Key Decisions

**JWT decoded client-side via `atob()`** — no library. The JWT payload (middle base64url segment) is decoded to extract `name`, `role`, `email`. Signature verification is delegated entirely to the backend.

**localStorage for session persistence** — decoded claims `{name, role, email}` are stored under `manage_user`. The token itself is never stored client-side (lives in HTTP-only cookie). On any page mount, localStorage is read synchronously — no API call needed.

**Refresh token only on `TOKEN_EXPIRED`** — early implementation mistakenly called `refreshToken()` on every page mount, which rotated the token pair on every visit and caused an infinite loop. Corrected: `apiPrivateRequest` already handles TOKEN_EXPIRED → auto-refresh internally. Dashboard pages read from localStorage only; `refreshToken()` is never called proactively.

**`credentials: "include"` on login** — `apiRequest` (used by `login()`) does not include credentials by default. Without this flag, the browser silently drops the HTTP-only cookies set in the cross-origin login response (dev: `:3000` → `:8080`), causing `REFRESH_TOKEN_MISSING` on subsequent authenticated requests.

**Role guard on dashboards** — if an admin navigates to `/manage/moderator` (or vice versa), the localStorage role is checked and they are redirected to their correct dashboard.

---

## Token Refresh Flow (for context)

```
apiPrivateRequest(path)
  └─ sendRequest(path)          → if 200: return data
  └─ catches ApiError(TOKEN_EXPIRED)
       └─ refreshAccessToken()  → POST /api/v1/auth/refresh-token
            └─ browser stores new access_token + refresh_token cookies automatically
       └─ sendRequest(path)     → retry original request
```

No manual refresh calls are needed from page code. `refreshAccessToken` uses `credentials: "include"` so both cookies are updated from the refresh response.

---

## State at End of Session

Feature fully implemented and verified. All 215 tests pass, zero type errors. Ready to extend dashboards with role-specific management UI in future sessions.
