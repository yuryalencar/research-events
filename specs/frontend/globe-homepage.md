# Globe Homepage

## Description

The landing page (`app/[locale]/page.tsx`). Renders a 3D globe (Globe.gl)
filling the center of the screen, with one pin per approved event for the
current year. Clicking a pin opens a side panel — on the same screen, no
navigation — showing that event's details. This is a read-only "explore"
view: no filters yet (filtering is a future feature).

## Behaviour

- On mount, the page fetches `GET /api/v1/events` via
  `listEvents({ pagination: "off" })` — relying on the backend's documented
  defaults for the rest (`status=approved`, `year=<current year>`), but with
  pagination disabled so **all** matching events for the current year are
  returned in one response (`meta.page=1`, `meta.total=data.length`).
- Each returned event is rendered as a pin on the globe at
  `(event.latitude, event.longitude)`.
- Clicking a pin:
  - Visually **highlights** that pin (e.g. distinct color and/or larger
    size) so the user can see which event is selected.
  - Opens the detail view for that event — on the same screen, no
    navigation (see "Detail view contents" and "Responsive layout" below).
- Clicking a different pin while a detail view is open: the previous pin's
  highlight is removed, the newly clicked pin is highlighted, and the detail
  view's contents update to the new event (the view stays open).
- Clicking the currently-highlighted pin again, or the detail view's close
  (X) button, closes the detail view and removes the highlight.
- While the initial fetch is in flight, a loading spinner is shown over the
  globe area. Once data arrives (or fails), the spinner is removed.
- If the fetch fails, `handleApiError` shows a translated toast
  (`errors.<CODE>` via the existing `lib/api/errors.ts`); the globe still
  renders, with zero pins.
- Globe.gl is loaded via `dynamic(() => import(...), { ssr: false })` per
  CLAUDE.md (requires browser APIs / WebGL).

### Responsive layout

This page must work on both web and mobile:

- **Desktop / wide viewports**: the detail view is a side panel (drawer)
  sliding in from the right edge. The globe remains visible and interactive
  beside it.
- **Mobile / narrow viewports**: the detail view is a card that slides up
  from the bottom of the screen (bottom sheet), overlaying the lower portion
  of the globe. The globe remains visible and interactive above it.
- Both layouts show the same content (see "Detail view contents") and the
  same close interaction (X button, or re-clicking the highlighted pin).
- The breakpoint follows the project's existing Tailwind CSS v4 breakpoints
  (e.g. side panel at `md:` and above, bottom card below `md:`).

### Detail view contents

For the selected event, the panel shows:
- Name
- Dates: `start_date` – `end_date`
- Location: `city`, `country`
- Website: `website_url` as a clickable link (opens in a new tab)
- Domain (raw string, e.g. `computer_science`)
- Tier: shown as a badge, **only when `tier !== "unranked"`** (e.g. `A*`,
  `A`, `B`, `C`). `"unranked"` shows no badge.
- Active deadlines (`deadlines[]`, already filtered to `is_active=true` by
  the backend): for each, `type`, `description`, `date` (+ `time`/`timezone`
  if present), `is_optional` flag.
  - If `deadlines` is empty, show a translated "no upcoming deadlines"
    message instead of an empty list.
- Attribution: "Added by `created_by.name`" and, if
  `last_updated_by.id !== created_by.id`, "Updated by `last_updated_by.name`".
  **Only `.name` is shown — `created_by.email` and `last_updated_by.email`
  are never rendered**, even though `UserSummary` includes them (privacy:
  contributor emails are not public-facing data).

## Rules

- `listEvents({ pagination: "off" })` is called — the backend applies
  `status=approved` and `year=<current year>` by default, and returns every
  matching event for the current year in one response (no pagination).
- The page must be usable on both web and mobile (see "Responsive layout").
- All user-facing strings (panel/card labels, loading text, "no upcoming
  deadlines", close button, error toasts) go through `next-intl`
  (`useTranslations`) — no hardcoded strings. New keys live under a
  `home` (or `globe`) namespace in `messages/{en,pt,es,de}.json`, all four
  locales kept in parity.
- Pins are rendered individually — no clustering, even if multiple events
  share the same/near coordinates (each gets its own pin; visual overlap is
  acceptable for this version).
- Exactly one pin can be highlighted/selected at a time.
- The globe is interactive (pan/zoom/rotate via Globe.gl defaults); no
  custom camera constraints for this feature.
- The 2D Leaflet fallback map (mentioned in CLAUDE.md tech stack) is **out of
  scope** for this feature — Globe.gl only.

## Permissions

- Public page — no authentication required. Uses `apiRequest`
  (`listEvents`), never `apiPrivateRequest`.

## Error cases

| Scenario | Expected behaviour |
|---|---|
| `GET /api/v1/events` fails (network error, 5xx, etc.) | `handleApiError` shows a translated toast (`errors.<CODE>` / `errors.NETWORK_ERROR` / `errors.UNKNOWN`); globe renders with zero pins, no spinner |
| `GET /api/v1/events` returns `data: []` (no approved events this year) | Globe renders normally with zero pins; no error toast |
| User clicks a pin while the detail view is already open for a different event | Detail view content updates to the new event, highlight moves to the new pin; view stays open |

## Border / corner cases

- Event with `tier: "unranked"` → no tier badge shown.
- Event with `deadlines: []` → detail view shows "no upcoming deadlines"
  message, not an empty list.
- `created_by.id === last_updated_by.id` (never edited after creation) →
  only "Added by" line shown, no "Updated by" line.
- Two events with identical `(latitude, longitude)` → both pins rendered
  individually (per Rules — no clustering); clicking either highlights only
  that one.
- Locale switch while the detail view is open → labels re-render in the new
  locale; event data itself (name, city, etc., which come from the backend
  in whatever language was submitted) is unchanged.
- Page loaded, fetch still pending, user clicks where a pin will appear →
  no-op (no pins exist yet during loading).
- Viewport resized across the mobile/desktop breakpoint while the detail
  view is open → the same selected event's content is shown in the new
  layout (side panel ↔ bottom card), no data refetch.
- `meta.total` larger than the number of pins actually rendered would
  indicate a bug (with `pagination=off`, `data.length` must equal
  `meta.total`) — not expected, but the page always renders one pin per item
  in `data`, never relies on `meta.total` for rendering.

## Definition of done

- [ ] `app/[locale]/page.tsx` renders the globe (via dynamic, `ssr: false`
      import) filling the center of the screen, on both web and mobile
      viewports
- [ ] On mount, `listEvents({ pagination: "off" })` is called and a pin is
      rendered for every returned event at `(latitude, longitude)`
- [ ] Loading spinner shown while the fetch is in flight, removed once
      resolved (success or failure)
- [ ] Fetch failure → `handleApiError` toast shown, globe renders with zero
      pins (no crash)
- [ ] Empty `data: []` → globe renders with zero pins, no error toast
- [ ] Clicking a pin highlights it and opens the detail view (same screen)
      with the event's full details per "Detail view contents"
- [ ] On desktop (`md:` and above), the detail view is a side panel sliding
      in from the right
- [ ] On mobile (below `md:`), the detail view is a card sliding up from the
      bottom
- [ ] Clicking a different pin while open: highlight moves, detail view
      contents update, view stays open
- [ ] Clicking the highlighted pin again, or the detail view's close button,
      closes the view and removes the highlight
- [ ] Tier badge shown only when `tier !== "unranked"`
- [ ] "No upcoming deadlines" message shown when `deadlines: []`
- [ ] "Updated by" line shown only when `last_updated_by.id !== created_by.id`
- [ ] All detail view/loading/error strings present in
      `messages/{en,pt,es,de}.json` under a new namespace, all four locales
      in parity (covered by a key-parity test, matching the pattern of
      `messages/errors.test.ts`)
- [ ] `pnpm typecheck`, `pnpm lint`, `pnpm test` all pass
