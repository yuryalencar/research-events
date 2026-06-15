# API Client + Error Handling

## Description

Foundational frontend infrastructure: a typed fetch client for the Go backend
(public and private/cookie-authenticated requests), a centralized error
handler that maps backend error codes to translated, user-friendly messages
shown via toast, and the shadcn/ui + sonner setup needed to render those
toasts with a clean, minimalistic design. This is infrastructure other
features (submission form, admin review, etc.) will build on.

## Behaviour

### Response envelope (read from running backend code, `internal/handler/auth.go`)

- Success: `{ "code": "SOME_CODE", "data": T }` or, for list endpoints,
  `{ "code": "SOME_CODE", "data": T[], "meta": {...} }`
- Error: `{ "code": "SOME_ERROR_CODE", "error": { "message": "..." } }`

  Note: this differs slightly from the envelope shape documented at the top
  of `CLAUDE.md` (`{ "error": { "code", "message" } }`) — the **actual**
  backend implementation puts `code` at the top level. The client and types
  follow the real implementation.

### `lib/api/client.ts` — low-level request functions

- `apiRequest<T>(path, init?)` — public request. Prepends
  `NEXT_PUBLIC_API_URL`, sets `Content-Type: application/json` for bodies with
  content, parses the JSON envelope.
  - 2xx → returns `data` (and `meta` when present) typed as `T`
  - non-2xx → throws `ApiError` (see below)
  - network failure / non-JSON body → throws `ApiError` with code `"NETWORK_ERROR"`
- `apiPrivateRequest<T>(path, init?)` — same as above, plus:
  - `credentials: "include"` (sends the HTTP-only `access_token` /
    `refresh_token` cookies)
  - On a `401 TOKEN_EXPIRED` response: calls `POST /api/v1/auth/refresh-token`
    once (also `credentials: "include"`), and if that succeeds, retries the
    original request exactly once. If the refresh call itself fails, or the
    retried request also 401s, the **original** 401 `ApiError` is thrown (the
    caller doesn't need to special-case the retry).
  - Any other status code (including other 401 variants like `TOKEN_MISSING`,
    `TOKEN_INVALID`) is thrown immediately, no refresh attempted.

### `ApiError`

```ts
class ApiError extends Error {
  readonly code: string      // e.g. "EVENT_NOT_FOUND", "NETWORK_ERROR"
  readonly status: number    // HTTP status; 0 for network errors
  readonly message: string   // raw backend message (English) — not shown to users directly
}
```

### `lib/api/errors.ts` — error-to-message mapping + toast

- `errorMessageKey(code: string): string` — pure function, maps a backend
  error `code` to an i18n key under the `errors` namespace
  (`errors.<CODE>`). Unknown/unmapped codes → `errors.UNKNOWN`.
- `handleApiError(error: unknown, t: (key: string) => string): void` —
  - If `error instanceof ApiError`: looks up `errorMessageKey(error.code)`,
    translates it via `t`, and shows it with `toast.error(...)` (sonner).
  - If `error` is anything else (unexpected throw): shows the translated
    `errors.UNKNOWN` message.
  - `VALIDATION_ERROR` always maps to a generic translated message
    (`errors.VALIDATION_ERROR`, e.g. "Please check your input and try
    again") — the raw backend validation message (`error.message`, English,
    field-specific) is never shown to the user. (Per-field inline validation
    is a separate future feature.)

### Endpoint functions, grouped by resource, derived from `specs/backend/*.yaml`

| File | Function | Endpoint | Client |
|---|---|---|---|
| `lib/api/events.ts` | `listEvents(params)` | `GET /api/v1/events` | public |
| `lib/api/events.ts` | `submitEvent(input)` | `POST /api/v1/events/submit` | public |
| `lib/api/events.ts` | `addDeadlines(eventId, input)` | `POST /api/v1/events/{id}/deadlines` | public |
| `lib/api/events.ts` | `cancelDeadline(eventId, deadlineId)` | `PATCH /api/v1/events/{eventId}/deadlines/{deadlineId}/cancel` | public |
| `lib/api/events.ts` | `supersedeDeadline(eventId, deadlineId, input)` | `POST /api/v1/events/{eventId}/deadlines/{deadlineId}/supersede` | public |
| `lib/api/auth.ts` | `login(input)` | `POST /api/v1/auth/login` | public (sets cookies) |
| `lib/api/auth.ts` | `refreshToken()` | `POST /api/v1/auth/refresh-token` | private (used internally by client, and exported for explicit use) |
| `lib/api/auth.ts` | `logout()` | `POST /api/v1/auth/logout` | private |
| `lib/api/admin.ts` | `reviewEvent(id, input)` | `PATCH /api/v1/admin/events/{id}/review` | private |
| `lib/api/admin.ts` | `unlockUser(id)` | `PATCH /api/v1/admin/users/{id}/unlock` | private |
| `lib/api/health.ts` | `getHealth()` | `GET /health` | public (no `/api/v1` prefix) |

### Types (`types/api.ts`)

`make generate-types` has no OpenAPI spec to generate from yet (out of
scope for this feature). Request/response types for the endpoints above are
**hand-written** in `types/api.ts` from the YAML specs, with a top-of-file
comment flagging them for replacement once `make generate-types` exists. This
is a deliberate, temporary exception to the "never hand-write API types" rule.

### shadcn/ui + sonner setup

- Initialize shadcn/ui for this Next.js 15 + Tailwind v4 project
  (`components.json`, `lib/utils.ts` with `cn()` via `clsx` + `tailwind-merge`,
  base theme tokens in `app/globals.css`, minimal/neutral palette).
- Add the shadcn `sonner` component (`components/ui/sonner.tsx`, thin wrapper
  around the `sonner` package) and mount `<Toaster />` once in
  `app/[locale]/layout.tsx`.
- No other shadcn components are added in this feature — just the
  setup + `sonner`. Future features add components as needed.

## Rules

- All user-facing strings (toast messages) go through `next-intl` — no
  hardcoded English strings.
- `errors` namespace added to `messages/en.json` (source of truth) with one
  key per backend error code below; `pt.json`, `es.json`, `de.json` get
  matching keys (no missing keys).
- Private requests always use `credentials: "include"`; public requests never
  send cookies.
- The refresh-and-retry logic lives only in `apiPrivateRequest` — endpoint
  functions never handle 401/refresh themselves.
- `apiRequest`/`apiPrivateRequest` are the *only* places `fetch` is called for
  backend API access — endpoint functions always go through them.

## Permissions

- `apiRequest` (public client): usable from any component, any role
  (including unauthenticated visitors).
- `apiPrivateRequest` (private client): intended for admin/moderator-only
  pages (`/admin/*`); the backend itself enforces role checks — the frontend
  client does not duplicate that logic, it only handles the auth
  cookie + refresh mechanics.

## Error cases

| Backend `code` | HTTP status | `errors.<CODE>` message shown (en) |
|---|---|---|
| `VALIDATION_ERROR` | 400 | "Please check your input and try again." |
| `UNAUTHORIZED` | 401 | "You need to be logged in to do that." |
| `TOKEN_MISSING` | 401 | "You need to be logged in to do that." |
| `TOKEN_EXPIRED` | 401 | "Your session has expired. Please log in again." |
| `TOKEN_INVALID` | 401 | "Your session is invalid. Please log in again." |
| `INVALID_CREDENTIALS` | 401 | "Incorrect email or password." |
| `ACCOUNT_LOCKED` | 423 | "This account is locked. Contact an administrator." |
| `REFRESH_TOKEN_MISSING` | 401 | "Your session has expired. Please log in again." |
| `REFRESH_TOKEN_INVALID` | 401 | "Your session has expired. Please log in again." |
| `REFRESH_TOKEN_REUSE` | 401 | "Your session is no longer valid. Please log in again." |
| `FORBIDDEN` | 403 | "You don't have permission to do that." |
| `CANNOT_REVIEW_OWN_EVENT` | 403 | "You can't review an event you submitted yourself." |
| `CANNOT_UNLOCK_SELF` | 422 | "You can't unlock your own account." |
| `EVENT_NOT_FOUND` | 404 | "Event not found." |
| `DEADLINE_NOT_FOUND` | 404 | "Deadline not found." |
| `USER_NOT_FOUND` | 404 | "User not found." |
| `EVENT_NOT_APPROVED` | 409/422 | "This event hasn't been approved yet." |
| `EVENT_ALREADY_SUBMITTED` | 409 | "An event with this name/slug has already been submitted." |
| `DEADLINE_ALREADY_INACTIVE` | 409 | "This deadline is no longer active." |
| `SLUG_ALREADY_EXISTS` | 409 | "An event with this name already exists." |
| `USER_NOT_LOCKED` | 409 | "This account is not locked." |
| `RATE_LIMIT_EXCEEDED` | 429 | "Too many requests. Please wait a moment and try again." |
| `INTERNAL_ERROR` | 500 | "Something went wrong on our end. Please try again later." |
| `NETWORK_ERROR` (client-generated, request never reached the backend) | — | "Couldn't connect to the server. Check your connection and try again." |
| `UNKNOWN` (fallback for any unmapped code) | any | "Something went wrong. Please try again." |

## Border / corner cases

- Backend returns a `code` not in the table above (e.g. a new code added
  later without a frontend update) → `errors.UNKNOWN`.
- `fetch` itself throws (offline, DNS failure, CORS) → `ApiError("NETWORK_ERROR", 0, ...)` → `errors.NETWORK_ERROR`.
- Response body is not valid JSON (e.g. backend returns an HTML error page) →
  treated as `NETWORK_ERROR` (can't read `code`/`message`).
- `apiPrivateRequest` gets `401 TOKEN_EXPIRED` → refresh succeeds → retried
  request still fails (e.g. `403 FORBIDDEN`, the user's role changed) → the
  **second** error (`FORBIDDEN`) is thrown, not the original `TOKEN_EXPIRED`.
- `apiPrivateRequest` gets `401 TOKEN_EXPIRED` → refresh call itself returns
  `401 REFRESH_TOKEN_REUSE` (or any error) → original `TOKEN_EXPIRED`
  `ApiError` is thrown (per Behaviour above) — caller's `handleApiError` shows
  "Your session has expired."
- `apiPrivateRequest` gets `401 TOKEN_MISSING` or `401 TOKEN_INVALID` → no
  refresh attempted, thrown immediately.
- `handleApiError` called with a value that is not an `ApiError` (e.g. a
  plain `Error` from a bug elsewhere) → `errors.UNKNOWN`.
- `GET /api/v1/events` (list, has `meta`) vs other endpoints (no `meta`) —
  `apiRequest`'s return type only includes `meta` when the generic type
  parameter says so; `listEvents` is typed to return `{ data, meta }`, others
  return `data` directly.

## Definition of done

- [ ] shadcn/ui initialized for the project (`components.json`, `lib/utils.ts` `cn()`, theme tokens in `app/globals.css`)
- [ ] `sonner` installed; shadcn `components/ui/sonner.tsx` added; `<Toaster />` mounted in `app/[locale]/layout.tsx`
- [ ] `lib/api/client.ts`: `apiRequest`, `apiPrivateRequest`, `ApiError` implemented per Behaviour
- [ ] `apiPrivateRequest` refresh-and-retry-once logic implemented and unit tested (mocked `fetch`)
- [ ] `lib/api/errors.ts`: `errorMessageKey`, `handleApiError` implemented and unit tested
- [ ] `lib/api/{events,auth,admin,health}.ts`: typed functions for all 11 endpoints in the table above
- [ ] `types/api.ts`: hand-written request/response types for all 11 endpoints, with the "replace via `make generate-types`" comment
- [ ] `messages/en.json` gets a new `errors` namespace with all 24 keys from the Error cases table
- [ ] `messages/pt.json`, `es.json`, `de.json` mirror every `errors.*` key (translated, no missing keys)
- [ ] Vitest tests cover every row of Error cases + Border/corner cases above
- [ ] `pnpm typecheck`, `pnpm lint`, `pnpm test` all pass
