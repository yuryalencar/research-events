# Globe Event Filters — Session Summary

**Date:** 2026-06-16
**Feature:** Collapsible floating filter panel on the globe homepage

---

## Goal

Add a floating filter panel (top-left, `z-40`) to the globe homepage that lets users narrow visible events by year, domain, tier, country, and first deadline month. Filters are applied in draft/applied two-stage fashion — the globe only re-fetches when Apply is clicked.

---

## Decisions Made

### Architecture

- **Draft vs. applied state** — `useFilters` holds two parallel `EventFilters` objects: `draftFilters` (what the form shows) and `activeFilters` (what was last applied and what `useEvents` fetches against). Changing a control updates draft only; Apply promotes draft → active.
- **`reset()` resets draft only** — original spec said reset also applied immediately. User revised this mid-session: reset should only restore the form to defaults without triggering a re-fetch. The user must click Apply after resetting. Tests updated to match.
- **Year is mandatory** — `EventFilters.year` is always a `number`, never `undefined`. Every `listEvents` call always includes `year`. There is no "All years" option.
- **Countries in a separate file** — `src/lib/countries.ts` (not `constants.ts`) per user's explicit request. ~195 countries, hardcoded alphabetically.
- **`toListEventsParams`** — pure mapping function in `useEvents.ts` converting camelCase `EventFilters` → snake_case `ListEventsParams` (e.g. `firstDeadlineMonth` → `first_deadline_month`). Only includes optional fields when set, so query string stays clean.

### Re-fetch loop fix

`useTranslations()` returns a new function reference on every render. Putting `t` in the `useEffect` dep array caused a re-fetch after every state update. Fix: store `t` in a `useRef` (`tRef.current = t`) and access via `tRef.current` inside the effect. Individual filter primitives (`year`, `domain`, etc.) are destructured and used as dep array entries instead of the filter object (reference equality issue).

### UI choices

- **Month names** — computed inside component via `useLocale()` + `toLocaleString(locale, { month: "long" })`. First letter capitalised because pt/de return lowercase. Wrapped in `useMemo` on `locale`.
- **Country picker** — initially implemented as a `size={4}` visible list with a separate text search input. Simplified to a standard `<select>` dropdown: browser-native type-to-search covers the use case; separate input was removed.
- **Mobile auto-close** — `handleApply()` calls `apply()` then checks `window.matchMedia(DESKTOP_QUERY).matches`; if on mobile, closes the panel so the globe is fully visible.
- **Active-filter dot** — small blue dot on the toggle button when `isDirty` is true (draft ≠ active). Shows even when panel is collapsed.

### Globe rotation after filter apply

When filters are applied and the fetch completes, the globe rotates to `events[0]` (first result). An `isFirstLoad` ref skips the initial page-load fetch so the globe only moves on user-triggered applies. `GlobeView` receives a `focusPoint?: { lat, lng }` prop and rotates in a separate `useEffect` (preserving current altitude).

### "No events" message distinction

Two messages in the `home` i18n namespace:
- `noEvents` — shown when no filters beyond current-year defaults are active ("No events registered yet.")
- `noEventsFiltered` — shown when at least one non-default filter is active ("No events match the current filters. Try adjusting or resetting them.")

---

## Files Created / Modified

| File | Change |
|------|--------|
| `src/lib/countries.ts` | New — ~195 countries constant |
| `src/lib/constants.ts` | Added `MIN_FILTER_YEAR`, `DOMAINS`, `TIERS` |
| `src/hooks/useFilters.ts` | New — draft/active dual-state hook |
| `src/hooks/useFilters.test.ts` | New — 25 tests (setters, apply, reset, isDirty, year invariant) |
| `src/hooks/useEvents.ts` | Updated — accepts `EventFilters`, maps to `ListEventsParams`, `useRef` for `t` |
| `src/hooks/useEvents.test.ts` | Updated — 8 tests (pagination always off, optional params, re-fetch on change) |
| `src/components/globe/FilterPanel.tsx` | New — collapsible panel with 5 filter controls |
| `src/components/globe/GlobeView.tsx` | Added `focusPoint` prop + rotation effect |
| `src/app/[locale]/page.tsx` | Wired `useFilters`, `FilterPanel`, `focusPoint`, `hasNonDefaultFilters` |
| `src/messages/*.json` (×4) | Added `filters` namespace + `home.noEventsFiltered` |
| `src/messages/globeHomepage.test.ts` | Updated `EXPECTED_HOME_KEYS` to include `noEventsFiltered` |
| `specs/frontend/globe-filters.md` | Spec file (written Phase 0, updated Phase 6) |

---

## State at End of Session

- All 6 cycles complete: countries constant, translations, `useFilters`, `useEvents` update, `FilterPanel`, page wiring
- Phases 3–5 complete (Red → Green → Refactor); Phase 6 (this doc)
- 129 tests, 0 failures; `pnpm typecheck` clean

---

## Context to Restore

- `useFilters` draft/active pattern: never apply on reset, always require explicit Apply click
- `useEvents` dep array uses destructured primitives, `t` via `useRef` — do not add `t` back to dep array
- `COUNTRIES` lives in `src/lib/countries.ts`, not `constants.ts`
- `toListEventsParams` is a private helper in `useEvents.ts` — keep it there
- `focusPoint` in `GlobeView` is separate from `selectedEvent` — one rotates on filter apply, the other on pin click
