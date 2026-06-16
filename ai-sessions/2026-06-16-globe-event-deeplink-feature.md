# Globe Event Deep-link via URL Slug

**Date:** 2026-06-16
**Spec:** `specs/frontend/globe-event-deeplink.md`

---

## Goal

Add `?event=SLUG` URL deep-linking to the globe homepage so that:
- Clicking a pin adds `?event=<slug>` to the URL (shareable/bookmarkable)
- Closing the panel removes the param
- Loading the page with `?event=SLUG` behaves identically to clicking that pin
- The globe rotates (without changing zoom) to the selected event

---

## Decisions made

### No extra API call for slug resolution
The spec considered two approaches: (a) fetch the event by slug from the API on page load, or (b) resolve the slug from the in-memory events list returned by the existing `listEvents` call. We chose (b) because `listEvents` is already called on mount and returns all current-year approved events. A separate fetch would duplicate data and add latency.

### `router.replace` not `router.push`
The `?event=` param is a selection pointer, not a navigation step. Using `replace` means the browser back button leaves the page rather than looping through panel-open states. If we had used `push`, every pin click would add a history entry.

### `useSelectedEvent(events)` owns all URL sync
URL read and write logic was placed inside `useSelectedEvent` rather than in `page.tsx` or a separate hook. This keeps all selection logic — state, URL, and slug resolution — in one place. The hook now accepts `events: EventListItem[]` as a parameter to support slug lookup.

### Slug resolution runs at most once (ref guard)
A `slugResolvedRef = useRef(false)` prevents the on-load `useEffect` from re-running every time `searchParams` or `removeEventParam` references change. Without the guard, calling `router.replace` inside the effect (for the "not found" case) would trigger another render and re-run the effect in an infinite loop.

### Globe rotation preserves current zoom
`globe.pointOfView()` called as a getter first to read the current `altitude`, then as a setter passing `{ lat, lng, altitude }` with the same altitude. This re-orients the globe without snapping the user's zoom level.

### `GlobeViewProps`: `selectedEventId` → `selectedEvent`
The prop was changed from `selectedEventId: number | null` to `selectedEvent: EventListItem | null` so the component has access to `latitude`/`longitude` for rotation without needing a second prop. Pin color/radius comparisons use `selectedEvent?.id`.

---

## Files changed

| File | Change |
|---|---|
| `src/hooks/useSelectedEvent.ts` | Added `events` param, `useRouter`/`useSearchParams`/`usePathname`, URL write helpers, slug resolution effect |
| `src/hooks/useSelectedEvent.test.ts` | Added 8 new tests (URL writes + URL reads), `vi.mock('next/navigation')`, `setSearchParams` helper |
| `src/components/globe/GlobeView.tsx` | `selectedEventId` → `selectedEvent` prop, `globe.pointOfView()` on selection |
| `src/app/[locale]/page.tsx` | Pass `events` to hook; pass `selectedEvent` to GlobeView |
| `specs/frontend/globe-event-deeplink.md` | New spec file |

---

## State at end of session

Feature complete. All 101 frontend tests pass. TypeScript clean.

**To restore context:** The globe homepage is fully implemented (see `2026-06-15-globe-homepage-feature.md`). This session added URL deep-linking on top of it. The next logical feature would be either search/filter on the globe, or the public event detail page (`/events/[id]`).
