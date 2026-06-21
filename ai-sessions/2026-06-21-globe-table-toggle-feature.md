# Globe / Table View Toggle — Frontend

**Date:** 2026-06-21
**Spec:** `specs/frontend/globe-table-toggle.md`

---

## Goal

Add a floating toggle button on the globe homepage that lets desktop users switch between the 3D globe and a card-list table view. Both views share the same filter state. Mobile users (below `md` = 768 px) always see the globe; a resize while in table mode forces a switch back with a toast.

---

## What was built

### New files

| File | Purpose |
|---|---|
| `src/hooks/useViewMode.ts` | Manages `'globe' \| 'table'` state + `md` resize detection. Calls `onForcedGlobe` on desktop→mobile crossing while in table mode. Uses a ref for the callback so callers can pass closures without triggering effect loops. |
| `src/hooks/useViewMode.test.ts` | 7 tests: initial state, manual toggle both directions, forced switch fires + calls callback, negative cases (already globe, resize upward). |
| `src/components/globe/ViewToggleButton.tsx` | Fixed float button at `top-16 right-4`, `hidden md:flex`. Shows table icon + "Table mode" tooltip in globe mode; globe icon + "Globe mode" tooltip in table mode. Mirrors AddEventButton's hover-tooltip pattern. |
| `src/components/events/EventTableCard.tsx` | Expandable card: summary row (name/year, location, dates, domain/tier badges, next deadline) always visible; expand toggle reveals full active-deadline list sorted by date + "Manage deadlines" button (sessionStorage + router.push same as EventDetailView). |
| `src/components/events/EventTableView.tsx` | Full-screen container: loading spinner and empty-state message match globe's visual language; scrollable `ul` of `EventTableCard` items capped at `max-w-3xl`. |

### Modified files

| File | Change |
|---|---|
| `src/app/[locale]/page.tsx` | Added `useViewMode` with `handleForcedGlobe` (calls `closeDetail()` + shows sonner toast). `handleToggle` also calls `closeDetail()` on manual switch. Globe branch renders GlobeView + loading/empty overlays + InfoButton + EventDetailView; table branch renders EventTableView. FilterPanel and AddEventButton are always rendered. |
| `src/messages/en\|pt\|es\|de.json` | Added `viewToggle` (3 keys) and `eventTable` (5 keys) namespaces to all four locale files. |

---

## Key decisions

- **Globe and table never render simultaneously** — the globe is unmounted in table mode, keeping memory clean and avoiding WebGL/React state conflicts.
- **`onForcedGlobe` callback via ref** — `useViewMode` keeps a ref to the callback so callers can pass closures (e.g. containing `closeDetail` + `toast`) without those refs appearing in the effect dependency array. This prevents infinite re-render loops.
- **Debounce not needed** — the forced switch sets `viewMode = 'globe'`, so subsequent sub-md crossings are no-ops (already in globe mode). Natural state machine prevents multiple toasts per session without explicit debouncing.
- **`hidden md:flex`** — toggle button disappears instantly on resize via CSS, not waiting for a JS event. The `useViewMode` JS handler only fires for the forced globe switch.
- **Deadline list sorted by date** — active deadlines in the expanded card are sorted ascending so the nearest deadline is always first.
- **`SESSION_KEY` duplicated** — the `deadline_management_event` sessionStorage key is defined locally in `EventTableCard` and `EventDetailView`. Not extracted to a shared constant to avoid coupling two independent components to a shared file for a single string.

---

## State at end of session

Feature fully implemented and all tests passing (257/257). Typecheck clean. All DoD items checked in the spec.

- `useViewMode` — complete, tested
- `ViewToggleButton` — complete
- `EventTableCard` — complete (expand/collapse, deadline list, manage link)
- `EventTableView` — complete (loading, empty, list states)
- `page.tsx` wiring — complete
- i18n — all four locales updated

---

## Context to restore next session

No follow-on work required for this feature. If picking up here:
- The pin hover label was also changed this session: `GlobeView.tsx` now uses `.pointLabel()` to show `"Name (Year)"` instead of just the name.
- Next feature to implement is likely backend work or another frontend feature — check `specs/` for any approved specs not yet marked Done.
