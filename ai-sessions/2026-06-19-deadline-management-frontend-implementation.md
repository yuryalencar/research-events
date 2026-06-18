# Deadline Management — Frontend Implementation

**Date:** 2026-06-19
**Phases completed:** 3 (Red) → 4 (Green) → 5 (Refactor) → 6 (Docs)
**Continued from:** `ai-sessions/2026-06-18-deadline-management-frontend-planning.md`

---

## Goal

Implement the full Deadline Management frontend: a single page at `/[locale]/events/[slug]/deadlines` where any contributor can add new deadlines, supersede existing ones, and cancel active ones — all in one batch submission. Entry point is a pencil icon in the event detail drawer/sheet on the globe homepage.

---

## Files Created

| File | Description |
|------|-------------|
| `src/hooks/useDeadlineManage.ts` | Main hook — full state machine for the manage page |
| `src/hooks/useDeadlineManage.test.ts` | 58 tests across 3 describe blocks (state, validate, submitChanges) |
| `src/components/events/deadlines/DeadlineManageSuccess.tsx` | Success screen with summary (added/updated/cancelled counts) |
| `src/components/events/deadlines/AddDeadlineCard.tsx` | Form card for new deadlines (2-row layout: type/description/remove + date/time/timezone/required) |
| `src/components/events/deadlines/DeadlineCard.tsx` | Existing deadline card — 3 states: default, superseding, pendingCancel |
| `src/components/events/deadlines/DeadlineManagePage.tsx` | Client orchestrator: sessionStorage loader + DeadlineManageContent |
| `src/app/[locale]/events/[id]/deadlines/page.tsx` | Thin server wrapper route |

## Files Modified

| File | Change |
|------|--------|
| `src/components/events/EventDetailContent.tsx` | Added "Deadlines" section header with pencil icon; supersede visualization (strikethrough → new date); removed per-item pencil buttons; DeadlineList now accepts all deadlines and filters internally |
| `src/components/events/EventDetailView.tsx` | Added `handleManageDeadlines` — writes event to sessionStorage then navigates to `/[locale]/events/[slug]/deadlines` |
| `src/messages/en.json` | Added `deadlines.manage.*` keys (page, cards, buttons) + `eventDetail.deadlinesTitle` |
| `src/messages/pt.json` | Same |
| `src/messages/es.json` | Same |
| `src/messages/de.json` | Same |
| `src/messages/globeHomepage.test.ts` | Updated `EXPECTED_EVENT_DETAIL_KEYS` with `manageDeadlinesLabel` and `deadlinesTitle` |

---

## Key Decisions

**`useRef` + version counter for errors** — `validate()` mutates `errorsRef` synchronously; `setErrorVersion` triggers re-renders. Callers read the new errors immediately without waiting for a re-render cycle. Same pattern as `useSubmitWizard`.

**Flat dotted error keys** — `"contributor.name"`, `"supersede.{id}.date"`, `"new.{localId}.type"` etc. Single `Record<string, string>` propagated to all child components; each card derives its own prefix.

**Unchanged supersede detection** — `submitChanges` compares `editValues` against the original API values (null normalized to `""`) before adding to the parallel promises. Unchanged supersede edits are silently skipped.

**`Promise.all` parallel submission** — cancels + supersedes + `addDeadlines` all fire in parallel. A single `apiError` string captures any failure.

**sessionStorage state passing** — homepage writes the full `EventListItem` under `"deadline_management_event"` before navigating. Manage page reads it in `useEffect` on mount and redirects to `/${locale}` if missing or corrupt.

**Loader/content split in `DeadlineManagePage`** — `DeadlineManagePage` (loader) reads sessionStorage in `useEffect` and only renders `DeadlineManageContent` once the event is confirmed. Avoids calling `useDeadlineManage` conditionally.

**`"use client"` on `EventDetailContent`** — required because `EventDetailView` (client) passes `handleManageDeadlines` (a function) to it across the serialization boundary.

**`AddDeadlineCard` two-row layout** — 7-column single-row grid was too wide on desktop. Row 1: type + description + X (top-right, absolutely positioned). Row 2: date + time + timezone + Required checkbox. Required checkbox inverts `isOptional` (`checked={!isOptional}`).

**Supersede visualization in drawer** — `DeadlineList` now receives all deadlines (active + inactive). For each active deadline it finds the inactive predecessor via `d.superseded_by_id === activeDeadline.id` and renders `~~old date~~ → new date`.

**Back button matches submission wizard** — `flex justify-between border-t border-border pt-4` footer, same button classes as Step2Details.

---

## State at End

Feature complete. All 10 cycles done:

1. `useDeadlineManage` — state management (58 tests)
2. `useDeadlineManage` — validate
3. `useDeadlineManage` — submitChanges
4. `EventDetailContent` + `EventDetailView` — entry point
5. `DeadlineManageSuccess`
6. `AddDeadlineCard`
7. `DeadlineCard`
8. `DeadlineManagePage`
9. Route `app/[locale]/events/[id]/deadlines/page.tsx`
10. i18n — all 4 locales in parity

**Tests:** 210 passing, 0 failing.
**Typecheck:** clean.

---

## Context to Restore

- The ARM64 native binding error (`Cannot find module '@rolldown/binding-linux-arm64-gnu'`) reappears occasionally; fix with `CI=true pnpm install`.
- The `globeHomepage.test.ts` file has a `EXPECTED_EVENT_DETAIL_KEYS` array that must be updated whenever new `eventDetail.*` i18n keys are added.
- `useDeadlineManage` is initialized with a full `EventListItem` (including deadlines). The hook owns all UI state; child components receive slices + callbacks only.
- The `"deadline_management_event"` sessionStorage key is the contract between `EventDetailView` and `DeadlineManagePage`.
