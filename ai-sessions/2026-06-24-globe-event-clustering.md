# Globe Event Clustering

**Date:** 2026-06-24
**Spec:** `specs/frontend/globe-event-clustering.md`

---

## Goal

Fix the UX problem where multiple events at close or identical lat/lng coordinates on the 3D globe overlap into a single unclickable pin. The solution merges two options: supercluster-based grouping (Option B) and a multi-event drawer (Option C).

---

## Decisions Made

**Architecture**
- Clustering logic lives in a new `useGlobeClusters(events, zoom)` hook outside `GlobeView` — keeps the rendering component dumb and the logic independently testable.
- `GlobeView` reports camera altitude changes via `onZoomChange(zoom)` callback, debounced at 150ms using `globe.controls().addEventListener("change", ...)` so re-clustering doesn't fire on every animation frame during a drag.
- `ClusterEventDrawer` lives in `components/globe/` (globe-specific interaction, not a general event component).
- Cluster pins use `pointsData` only — no `htmlElementsData` mixing — because the spec requires only a scaled radius + tooltip label, not rendered HTML badges.

**Clustering behaviour**
- Library: `supercluster@8.0.1` with `@types/supercluster@7.1.3`
- Cluster radius: `40px` at each zoom level (supercluster default)
- Max zoom for clustering: `16` — beyond this, all pins become individual
- Minimum cluster size: 2 events
- Dynamic: clusters split as the user zooms in; the camera change is debounced at 150ms

**Altitude → zoom conversion** (`lib/globe.ts`)
- Formula: `Math.round(-Math.log2(altitude) * 1.2 + 4)`
- Constants chosen so altitude `10` (fully out) → zoom `0`, altitude `0.0001` (max in) → zoom `20`
- Result clamped to `[0, 20]`

**Visual**
- Cluster color: `#a78bfa` (violet) — distinct from yellow/pink/red individual pins
- Radius scale: `2 events → 1.4`, `3–5 → 1.8`, `6–10 → 2.4`, `11+ → 3.0`
- Cluster label (tooltip): plain `"N events"` string — globe.gl tooltips are not React-rendered

**Interaction**
- Clicking a cluster → `ClusterEventDrawer` opens immediately (no zoom-in step)
- Drawer shows compact list: event name + date range, sorted by start date ascending
- Clicking a row → drawer closes, `EventDetailView` opens for the selected event
- Drawer closes when a new fetch starts (filter change) via `isLoading` check in the focus-point `useEffect`
- View toggle to table mode also closes any open cluster drawer

**InfoModal**
- Violet `PinLegendRow` added after the pink row with i18n keys `info.violetPin.label` / `info.violetPin.description`

**i18n**
- 4 new keys added to all 4 locales (`en`, `pt`, `es`, `de`):
  - `home.cluster.drawerTitle` — drawer header with `{count}` interpolation
  - `home.cluster.eventCount` — pin tooltip count label
  - `info.violetPin.label` — legend label
  - `info.violetPin.description` — legend description
- `globeHomepage.test.ts` `EXPECTED_HOME_KEYS` updated to include the two new `cluster.*` keys

**One test fix during Green**
- The "splits a cluster on zoom" test originally used Vienna + Tokyo at zoom 0, which don't cluster (87px apart > 40px radius). Corrected to use same-lat/lng events at zoom 1 splitting at zoom 17 (beyond `maxZoom: 16`), which correctly tests the supercluster `maxZoom` behaviour.

---

## Files Changed

| File | Change |
|---|---|
| `src/lib/events.ts` | Added `PIN_COLOR_CLUSTER`, `getClusterPinRadius` |
| `src/lib/globe.ts` | New — `altitudeToZoom` pure function |
| `src/lib/globe.test.ts` | New — tests for `altitudeToZoom` |
| `src/lib/events.test.ts` | Added tests for `getClusterPinRadius` |
| `src/hooks/useGlobeClusters.ts` | New — `useGlobeClusters`, `isCluster`, `ClusterPoint`, `GlobePoint` |
| `src/hooks/useGlobeClusters.test.ts` | New — 8 hook tests |
| `src/components/globe/GlobeView.tsx` | New props: `globePoints`, `onClusterClick`, `onZoomChange`; camera change listener |
| `src/components/globe/ClusterEventDrawer.tsx` | New — Dialog/Drawer multi-event picker |
| `src/components/globe/InfoModal.tsx` | Violet `PinLegendRow` added |
| `src/app/[locale]/page.tsx` | Wired `zoom`, `useGlobeClusters`, `clusterEvents`, `ClusterEventDrawer` |
| `src/messages/en|pt|es|de.json` | 4 new i18n keys each |
| `src/messages/globeHomepage.test.ts` | `EXPECTED_HOME_KEYS` updated |
| `package.json` | Added `supercluster`, `@types/supercluster` |

---

## State at End

- Feature fully implemented across all 8 cycles + refactor
- `cursor-pointer` applied to `ClusterEventDrawer` event rows (minor UI fix)
- Typecheck: clean (zero errors)
- Tests: 299 passing, 0 failing
- `specs/frontend/globe-event-clustering.md` — spec complete, all DoD items satisfied

---

## Context to Restore

- The `GlobePoint = EventListItem | ClusterPoint` union type is defined in and exported from `hooks/useGlobeClusters.ts` — `GlobeView` imports it from there
- `globe.controls()` returns the Three.js `OrbitControls` instance — this is how we attach the camera change listener without modifying globe.gl internals
- `supercluster` requires `@types/supercluster` separately (v8 ships no built-in types); `'cluster' in feature.properties` is the correct type-narrowing check (direct `.cluster` access fails TypeScript)
- The initial `zoom` state in `page.tsx` is hardcoded to `2` (matches globe.gl default altitude ~2.5); GlobeView also emits the actual initial zoom immediately on mount
