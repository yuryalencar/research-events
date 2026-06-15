# Session: API Client + Error Handling (Frontend)

**Date:** 2026-06-15
**Status:** Complete. Feature workflow (Phase 0 → 6), all 11 cycles done.

---

## Context

First frontend feature of the project. Builds the foundational infrastructure
every later feature (submission form, admin review, etc.) depends on: a typed
fetch client for the Go backend, centralized error-to-toast handling, and the
shadcn/ui + sonner setup needed to render those toasts.

Spec: `specs/frontend/api-client-error-handling.md` (approved before this
session's scope; phases 3-6 happened in this session).

---

## What was built

### shadcn/ui + sonner setup
- `components.json`, `src/app/globals.css` (Tailwind v4 CSS-first config,
  `@theme inline`, oklch "Neutral" palette), `src/lib/utils.ts` (`cn()` via
  `clsx` + `tailwind-merge`), `src/components/ui/sonner.tsx` (`Toaster`
  wrapper, `theme="light"`).
- Mounted `<Toaster />` and `<NextIntlClientProvider>` in
  `src/app/[locale]/layout.tsx`.
- shadcn's CLI (`@2.3.0` and `@4.11`) couldn't be used directly — v4.11's new
  "preset" registry 404s and doesn't map to "Neutral"; v2.3.0 doesn't support
  Tailwind v4's CSS-first config. Built the equivalent setup by hand following
  shadcn's documented Tailwind v4 conventions.

### `src/types/api.ts`
Hand-written request/response types for all 11 endpoints, transcribed from
`specs/backend/*.yaml` — a deliberate, temporary exception to "never
hand-write API types" (no OpenAPI spec/`make generate-types` yet).

### `src/lib/api/client.ts`
- `apiRequest<T>` / `apiRequestWithMeta<T>` — public requests, never send
  cookies.
- `apiPrivateRequest<T>` — `credentials: "include"`; on `401 TOKEN_EXPIRED`,
  calls `POST /api/v1/auth/refresh-token` once and retries the original
  request once. Other 401 variants (`TOKEN_MISSING`, `TOKEN_INVALID`) throw
  immediately, no refresh.
- `ApiError` (`code`, `status`, `message`). Network failures and unparsable
  bodies become `ApiError("NETWORK_ERROR", 0, ...)`.
- 13 tests (`client.test.ts`).

### `src/lib/api/errors.ts`
- `errorMessageKey(code)` — pure mapping from backend error code to
  `errors.<CODE>` i18n key, falling back to `errors.UNKNOWN` for unmapped
  codes.
- `handleApiError(error, t)` — shows a sonner toast with the translated
  message; `VALIDATION_ERROR` always shows the generic message, never the raw
  backend field-specific text.
- 28 tests (`errors.test.ts`).

### `src/messages/{en,pt,es,de}.json`
New `errors` namespace, 25 keys (24 backend error codes + `UNKNOWN`), added to
all four locales with full translations. Key-parity test
(`src/messages/errors.test.ts`, 4 tests) checks `pt`/`es`/`de` mirror `en`
exactly.

### Endpoint wrapper functions
- `lib/api/auth.ts` — `login`, `refreshToken`, `logout` (3 tests)
- `lib/api/events.ts` — `listEvents`, `submitEvent`, `addDeadlines`,
  `cancelDeadline`, `supersedeDeadline` (6 tests, including query-string
  building for `listEvents`)
- `lib/api/admin.ts` — `reviewEvent`, `unlockUser` (2 tests)
- `lib/api/health.ts` — `getHealth` (1 test, `GET /health`, no `/api/v1`
  prefix)

All endpoint functions go through `apiRequest`/`apiRequestWithMeta`/
`apiPrivateRequest` — `fetch` is never called elsewhere.

---

## Process

Strict Red-Green-Refactor TDD, one cycle at a time with explicit approval
between phases (per the approved Phase 2 plan's 11-cycle breakdown). Every
cycle's Refactor phase was explicitly skipped ("nothing to refactor") — no
duplication or cleanup needed across the whole feature.

---

## Lint/build setup fix (bonus, not in original spec)

While running the commit gate, `pnpm lint` (`next lint`) failed — Next 16
removed the `next lint` command entirely, and there was no `eslint.config.js`
for a direct `eslint` invocation.

Fixed:
- Added `eslint.config.mjs` (flat config) spreading `eslint-config-next`'s
  exported config array.
- Changed `package.json`'s `lint` script from `next lint` to `eslint .`.
- **Downgraded `eslint` from `^10.4.1` to `^9.39.4`.** `eslint-plugin-react@7.37.5`
  (pulled in by `eslint-config-next@16.2.7`, latest available) calls the
  removed `context.getFilename()` API and crashes under ESLint 10.
  `eslint-config-next@16.2.7`'s own `peerDependencies` require `eslint
  >=9.0.0` — v9 is the actually-supported version today, not a step backward.
  Documented in a comment at the top of `eslint.config.mjs` with the
  "remove this note and bump back to ^10" condition (once `eslint-plugin-react`
  ships a fix and `eslint-config-next` raises its peer range).
- Fixed two `import/no-anonymous-default-export` warnings (`eslint.config.mjs`,
  `postcss.config.mjs`) by assigning to a named `config` const before export.

---

## State at end

- `pnpm typecheck`, `pnpm lint`, `pnpm test` (57 tests), and `pnpm build` all
  pass cleanly.
- All Definition of Done items in
  `specs/frontend/api-client-error-handling.md` verified complete.

## Context to restore (next session)

- This is the foundation layer. Next frontend features (submission form, event
  detail page, admin review queue, globe) will consume `lib/api/*` and
  `lib/api/errors.ts` directly — no further client infrastructure needed.
- Watch the `eslint` pin: re-check `eslint-plugin-react` releases periodically
  and bump `eslint` back to `^10` + remove the note in `eslint.config.mjs`
  once compatible.
