# Management Portal (`/manage`)

## Description

A hidden management section accessible only by direct URL — no links from the homepage.
Provides login for admin and moderator roles, then routes each to a role-specific welcome
dashboard. The section is titled "ReSEARCH Manage".

## Routes

| Route | Description |
|---|---|
| `/[locale]/manage` | Login page — email + password card + "become a moderator" banner |
| `/[locale]/manage/admin` | Admin welcome dashboard |
| `/[locale]/manage/moderator` | Moderator welcome dashboard |

## Login Page Behaviour

- Renders a centered card with:
  - Section title: **"ReSEARCH Manage"**
  - Email field (type="email", required)
  - Password field (type="password", required)
  - Submit button ("Sign in")
- Below the login card: a second full-card-style banner — "Want to become a moderator?
  Send us an email." — with a mailto link using the `ADMIN_EMAIL` constant
  (`src/lib/constants.ts`). Never hardcode the email address.
- Client-side validation before submit: both fields must be non-empty. If either is
  empty, show inline field error — no API call.
- On submit: calls `login()` from `src/lib/api/auth.ts`
- On success:
  - Decodes the JWT from `data.token` (base64 payload via `atob()` — no library)
  - Stores decoded claims (`name`, `role`, `email`) in `localStorage` under key
    `manage_user` as a JSON string. The token itself is never stored client-side.
  - Redirects to `/manage/admin` or `/manage/moderator` based on decoded `role`
- On error: renders the translated error message inline below the submit button using
  existing `errors.*` i18n keys. No page redirect.
- If the user already has a valid session (refresh-token call succeeds on `/manage` load):
  redirect immediately to the correct dashboard — do not render the login form.

## Dashboard Pages (`/manage/admin`, `/manage/moderator`)

- On mount: call `refreshToken()` from `src/lib/api/auth.ts`
  - While the call is in-flight: show a full-page loading state (spinner + "Loading…")
  - Success → decode new JWT from response `data.token` → update `localStorage`
    `manage_user` claims → render the dashboard
  - Failure (any error) → clear `localStorage` `manage_user` → redirect to `/manage`
- After load, render:
  - "Welcome, {name}" heading
  - Role badge below the name: "admin" or "moderator" (from decoded JWT)
  - Logout button
- Role mismatch guard: if a user with role `admin` navigates to `/manage/moderator`
  (or vice versa), redirect to their correct dashboard after session validation.
- Logout button:
  - Calls `logout()` from `src/lib/api/auth.ts`
  - Clears `localStorage` `manage_user`
  - Redirects to `/manage`

## JWT Decoding

Decoded client-side from the JWT payload (middle segment, base64url-encoded). No
signature verification — trust is delegated to the backend. Utility function lives at
`src/lib/utils/decodeJwt.ts`.

Claims used: `name` (string), `role` ("admin" | "moderator"), `email` (string).

## Permissions

| Role | Login result |
|---|---|
| admin | 200 → `/manage/admin` |
| moderator | 200 → `/manage/moderator` |
| contributor | 403 → inline `errors.FORBIDDEN` message |
| non-existent / wrong password | 401 → inline `errors.INVALID_CREDENTIALS` message |
| locked account | 423 → inline `errors.ACCOUNT_LOCKED` message |

## Error Cases

| Scenario | Behaviour |
|---|---|
| Empty email or password on submit | Client-side field error, no API call |
| Wrong credentials | Inline `errors.INVALID_CREDENTIALS` |
| Account locked | Inline `errors.ACCOUNT_LOCKED` |
| Contributor login attempt | Inline `errors.FORBIDDEN` |
| Refresh fails on dashboard load | Clear localStorage, redirect to `/manage` |
| Admin visits `/manage/moderator` | Redirect to `/manage/admin` |
| Moderator visits `/manage/admin` | Redirect to `/manage/moderator` |
| Already logged in, visits `/manage` | Redirect to correct dashboard |
| Logout API fails | Still clear localStorage and redirect (best-effort) |

## i18n

New `manage.*` keys added to all four locale files: `en.json`, `pt.json`, `es.json`,
`de.json`. No hardcoded UI strings anywhere in the management pages.

Keys (defined in `en.json` as source of truth):
```
manage.title              — "ReSEARCH Manage"
manage.login.emailLabel   — "Email"
manage.login.passwordLabel — "Password"
manage.login.submitButton — "Sign in"
manage.login.emailRequired — "Email is required"
manage.login.passwordRequired — "Password is required"
manage.moderatorBanner.title — "Want to become a moderator?"
manage.moderatorBanner.description — "Help review event submissions by joining our team."
manage.moderatorBanner.emailLabel — "Send us an email"
manage.dashboard.welcomeHeading — "Welcome, {name}"
manage.dashboard.loading  — "Loading…"
manage.dashboard.logoutButton — "Sign out"
```

## Files to Create / Modify

| File | Action |
|---|---|
| `src/app/[locale]/manage/page.tsx` | Create — login page |
| `src/app/[locale]/manage/admin/page.tsx` | Create — admin dashboard |
| `src/app/[locale]/manage/moderator/page.tsx` | Create — moderator dashboard |
| `src/lib/utils/decodeJwt.ts` | Create — JWT payload decoder |
| `src/lib/utils/decodeJwt.test.ts` | Create — unit tests |
| `src/messages/en.json` | Modify — add `manage.*` keys |
| `src/messages/pt.json` | Modify — add `manage.*` keys |
| `src/messages/es.json` | Modify — add `manage.*` keys |
| `src/messages/de.json` | Modify — add `manage.*` keys |

No new API functions. `login`, `refreshToken`, `logout` from `src/lib/api/auth.ts`
are already implemented and tested.

## Border / Corner Cases

- JWT `role` claim contains an unexpected value (not "admin" or "moderator") → redirect
  to `/manage` and clear localStorage (treat as invalid session)
- `atob()` / JSON.parse throws (malformed JWT) → treat as invalid session, redirect
  to `/manage`
- `localStorage` is unavailable (private browsing on some browsers) → gracefully
  degrade: session is not persisted; user must log in again on each page load
- Logout API call fails → still clear localStorage and redirect (best-effort cleanup)

## Definition of Done

- [ ] `/manage` renders login card with email, password, submit — no homepage link
- [ ] Section title "ReSEARCH Manage" displayed
- [ ] "Become a moderator?" full card below login card links to `mailto:ADMIN_EMAIL`
- [ ] Empty field on submit → client-side error, no API call
- [ ] Valid admin credentials → redirect to `/manage/admin`
- [ ] Valid moderator credentials → redirect to `/manage/moderator`
- [ ] Wrong credentials → inline error, no redirect
- [ ] Locked account → inline error `errors.ACCOUNT_LOCKED`
- [ ] Contributor login → inline error `errors.FORBIDDEN`
- [ ] Already logged in visiting `/manage` → redirect to correct dashboard
- [ ] Dashboard shows full-page loading spinner while refresh-token call is in-flight
- [ ] Dashboard shows "Welcome, {name}" + role from decoded JWT
- [ ] Logout clears localStorage, redirects to `/manage`
- [ ] Expired/missing session on dashboard → redirect to `/manage`
- [ ] Role mismatch on dashboard URL → redirect to correct dashboard
- [ ] `decodeJwt` util has unit tests covering happy path + malformed input
- [ ] All `manage.*` i18n keys present in all 4 locale files, no missing keys
