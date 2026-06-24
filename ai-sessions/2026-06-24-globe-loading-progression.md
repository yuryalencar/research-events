# Globe Loading Message Progression

**Date:** 2026-06-24
**Spec:** [specs/frontend/globe-loading-progression.md](../specs/frontend/globe-loading-progression.md)

---

## Goal

Improve UX on the first page load when the backend is cold-starting (Render free tier spins down after 15 min of inactivity). Instead of a static "Loading events…" spinner, the loading overlay progressively cycles through messages so users understand the delay and stay on the page.

---

## Decisions made

- **Timing:** 2500ms before the first message switch, then 1500ms between each subsequent message (adjusted from 1000ms after review — 1s was too fast to read comfortably).
- **No loop:** progression stops at the last message and stays there.
- **Reset on re-fetch:** when `isLoading` flips `false → true` (new filter fetch), timers reset and the sequence restarts from message 1.
- **Scope limited to globe overlay only** — no other loading states were affected.
- **Hook abstraction:** all timer logic lives in `useLoadingMessage(isLoading, { messages, initialDelay?, interval? })` so the page component stays declarative.
- **Stable message array:** callers should not recreate the array on every render; the page constructs it inline (stable because locale translations don't change mid-session).

---

## Message sequence (en)

| # | Delay from start | Message |
|---|---|---|
| 1 | 0 ms | "Loading events…" |
| 2 | 2500 ms | "Waking up our server…" |
| 3 | 4000 ms | "First load takes a moment…" |
| 4 | 5500 ms | "Almost there, hang tight…" |

---

## Files created / modified

| File | Change |
|---|---|
| `specs/frontend/globe-loading-progression.md` | New spec |
| `hooks/useLoadingMessage.ts` | New hook |
| `hooks/useLoadingMessage.test.ts` | 8 tests |
| `messages/globeHomepage.test.ts` | Updated `EXPECTED_HOME_KEYS`: replaced `"loading"` with `loadingStep1–4` |
| `messages/en.json` | Replaced `home.loading` with 4 step keys |
| `messages/pt.json` | Same |
| `messages/es.json` | Same |
| `messages/de.json` | Same |
| `app/[locale]/page.tsx` | Wired hook, replaced `t("loading")` with `loadingMessage` |

---

## State at end of session

Feature complete and tested. 307 frontend tests pass, typecheck clean.

v0.1.2 release — includes globe event clustering (v0.1.1) + loading progression.

---

## Context to restore

- `useLoadingMessage` defaults: `initialDelay=2500`, `interval=1500`
- The old `home.loading` key no longer exists — it is replaced by `loadingStep1`–`loadingStep4`
- `globeHomepage.test.ts` is the i18n parity guard for the `home` namespace
