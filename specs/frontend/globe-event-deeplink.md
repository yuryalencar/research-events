# Globe Event Deep-link via URL Slug

## Description

When a user selects an event on the globe homepage, the URL is updated to
`?event=SLUG` so the selection can be bookmarked or shared. When the panel
closes the param is removed. Loading the page with `?event=SLUG` in the URL
behaves identically to clicking that event's pin: the globe rotates to the
event and the detail panel opens.

No extra API call is made — the slug is resolved against the events already
returned by `listEvents` (same fetch as today). This also adds globe
rotation/zoom when any event is selected (click or URL), which is new
behaviour not present in the current implementation.

## Behaviour

### Clicking a pin

1. Globe rotates and zooms to that event's `(latitude, longitude)` (new —
   currently the globe does not move on selection).
2. Detail panel opens — same Sheet / Drawer as today.
3. URL is updated with `router.replace` to append `?event=<slug>`.
   `router.replace` (not `push`) so the back button never goes to a
   "panel-open" state — the slug is a pointer, not a navigation step.

### Closing the panel

Any existing close gesture (X button, click-outside, Escape, re-clicking the
highlighted pin) closes the panel AND removes `?event=` from the URL via
`router.replace`. No other behaviour changes.

### Page loads with `?event=SLUG` in the URL

1. Events are fetched exactly as today (`listEvents({ pagination: "off" })`),
   same loading spinner.
2. Once events arrive, find the event where `event.slug === SLUG`.
3. **Found** → call `selectEvent(event)` — identical result to clicking the
   pin: globe rotates to the event, panel opens, URL stays with `?event=SLUG`.
4. **Not found** → call `router.replace` to remove `?event=` from the URL.
   No error message, no toast — silent removal.
5. While events are still loading, no action is taken on the slug — wait for
   the events list to resolve first.

### Globe rotation

When any event becomes selected (click or URL load), the globe animates
smoothly to that event using `globe.pointOfView({ lat, lng, altitude })` with
the library's default animated transition. When the panel closes the globe
does not move — it stays at its current orientation.

## Rules

- `router.replace` is always used (never `router.push`) — the `?event=` param
  is not a history entry.
- The slug is resolved from the in-memory events list returned by the existing
  `listEvents` call — no additional API request is made for slug lookup.
- If events load and the slug is not present in the list, the param is removed
  silently — no error toast, no visible indication to the user.
- Globe rotation applies on every selection, regardless of how the event was
  selected (click or URL).
- All other behaviours of the globe homepage (loading spinner, zero-pin empty
  state, error toast, panel content, responsive layout, close gestures) remain
  exactly as specified in `specs/frontend/globe-homepage.md`.
- No new i18n keys are introduced — this feature adds no new visible strings.

## Permissions

- Public page — no authentication required. Same as globe homepage today.

## Error cases

| Scenario | Expected behaviour |
|---|---|
| `?event=unknown-slug` on page load, slug not in events list | `router.replace` removes `?event=` silently; page shows globe with no selection |
| `?event=unknown-slug` on page load, events fetch fails | Events list is empty; no event found; `router.replace` removes param silently. Existing error toast from fetch failure still shown |
| `?event=SLUG` on page load, events still loading | No action until `isLoading` becomes false and the events list is populated |
| URL has `?event=SLUG` and user clicks a different pin | URL updates to `?event=NEW-SLUG`, panel content switches — same toggle behaviour as today |
| URL has `?event=SLUG` and user clicks the same pin (toggle) | Panel closes, URL param removed |

## Border / corner cases

- Globe in mid-rotation when user closes panel — globe stays at current
  position of the rotation, no snap-back.
- User edits URL manually to `?event=garbage` while page is loaded — treated
  as an unknown slug: param removed silently on next event selection/load.
  (URL changes after mount are not watched — only the initial URL and explicit
  `selectEvent`/`closeDetail` calls update the URL.)
- `?event=SLUG` combined with other query params (e.g. future filter params) —
  `router.replace` must preserve any other params while only modifying
  `event`. Use `URLSearchParams` to mutate only the `event` key.
- Locale prefix in URL (`/en/?event=SLUG`, `/pt/?event=SLUG`) — `useRouter`
  and `useSearchParams` from `next/navigation` operate on the path after the
  locale prefix; locale is unaffected.
- Two browser tabs open to the same URL — URL change in one tab does not
  affect the other (standard browser behaviour, no special handling needed).

## Definition of done

- [ ] Clicking a pin adds `?event=<slug>` to the URL (router.replace, no new
      history entry)
- [ ] Globe rotates+zooms to the selected event on every selection (click or
      URL load)
- [ ] Closing the panel (X button, click-outside, Escape, re-click) removes
      `?event=` from the URL
- [ ] Page loads with `?event=SLUG` for a known event → event is selected,
      panel opens, globe rotates to it
- [ ] Page loads with `?event=unknown` → param removed silently, no panel, no
      error toast
- [ ] Slug resolution waits for events to load (no action while
      `isLoading=true`)
- [ ] `router.replace` used everywhere — back button never sees a
      "panel-open" intermediate state
- [ ] Other query params (if any) are preserved when setting/removing `?event=`
- [ ] `pnpm typecheck`, `pnpm lint`, `pnpm test` all pass with zero failures
