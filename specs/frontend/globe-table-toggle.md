# Globe / Table View Toggle

## Description

Adds a floating toggle button on the homepage that lets desktop users switch between the existing 3D globe view and a card-list table view of events. Both views share the same filter state. Mobile users (below `md` = 768 px) always see the globe; if they reach table mode via resize, the app switches back automatically with a toast.

---

## Behaviour

### Toggle button
- Fixed float button, top-right corner, directly below the AddEventButton
- Visible only on desktop (≥ `md` / 768 px) — hidden on smaller screens
- Follows the same positioning pattern as AddEventButton (does not shift when the EventDetailView panel opens)
- Tooltip and icon change based on current view:
  - Viewing globe → tooltip "Table mode", table icon
  - Viewing table → tooltip "Globe mode", globe icon

### Globe view (default)
- Unchanged from current behaviour
- Toggle button switches to table view on click

### Table view (card list)
- Replaces the globe in the full-screen area (globe is unmounted / hidden)
- FilterPanel remains visible and functional — filter changes update the table immediately (same `activeFilters` state)
- Long scrollable list — no pagination; loads all events matching current filters (mirrors globe behaviour)
- Each card displays:
  - Event name + year
  - Country, city
  - Domain, tier
  - Start date – End date
  - Next active deadline (label + date), or "No upcoming deadlines" if none
- Each card has an expand toggle
- Expanded state shows:
  - Full list of active deadlines grouped by type (abstract, paper, notification, camera-ready, other)
  - "Manage deadlines" link → navigates to `/[locale]/events/[id]/deadlines`

### Responsive forced switch
- When the viewport shrinks below `md` (768 px) while the user is in table mode:
  - View switches back to globe automatically
  - A toast is shown: *"Table view is not available on small screens"*
- The toggle button disappears on resize below `md` regardless of current view

---

## Rules

- Globe and table share one `useFilters` instance — `activeFilters` is the single source of truth for both
- Table fetches events via the same `useEvents` hook used by the globe (same API call, same filter params)
- Globe is unmounted (not just hidden) when table is active, and vice versa — they are never rendered simultaneously
- The view mode state lives in a `useViewMode` hook that also handles resize detection and the forced switch
- Breakpoint for forced switch: `md` = 768 px (matches Tailwind's default `md` breakpoint)
- Toggle button is rendered only when `window.innerWidth >= 768`; it must also be hidden via CSS (`hidden md:flex`) so it disappears instantly on resize without waiting for a JS event
- Toast is shown only on a forced switch (resize), never on a manual toggle

---

## Permissions

- Public — no authentication required; same as the globe homepage
- The "Manage deadlines" link inside expanded cards is visible to everyone but the target page handles its own auth

---

## Error cases

| Scenario | Expected behaviour |
|---|---|
| Filters return zero events in table mode | Same empty state as globe: "No events match the current filters" message, centred in the table area |
| Events loading in table mode | Same loading spinner as globe, centred in the table area |
| User resizes to < md while in table mode | Auto-switch to globe + toast "Table view is not available on small screens" |
| User manually toggles to table on desktop | No toast — only forced resize triggers it |

---

## Border / corner cases

- Rapid resize across the `md` boundary multiple times → toast should show only once per crossing (debounce or flag)
- Selected event in globe (EventDetailView open) when user switches to table → panel closes, selected event state resets
- Switching back to globe from table → globe resumes at its last camera position (it is re-mounted fresh if unmounted, or restored if kept in DOM)
- No events returned → empty state message must be visible in both views
- Very long event name → card must truncate gracefully, not break layout

---

## i18n keys (en.json additions)

```json
"viewToggle": {
  "tableModeTooltip": "Table mode",
  "globeModeTooltip": "Globe mode",
  "tableUnavailableToast": "Table view is not available on small screens"
},
"eventTable": {
  "noUpcomingDeadlines": "No upcoming deadlines",
  "manageDeadlines": "Manage deadlines",
  "expand": "Expand event details",
  "collapse": "Collapse event details",
  "nextDeadline": "Next deadline"
}
```

All four locales (`en`, `pt`, `es`, `de`) must have every key.

---

## Definition of done

- [x] Toggle button appears below AddEventButton on desktop (≥ 768 px), hidden on mobile
- [x] Tooltip shows "Table mode" + table icon when globe is active
- [x] Tooltip shows "Globe mode" + globe icon when table is active
- [x] Clicking the button switches between globe and table (never both rendered at once)
- [x] Table is a scrollable card list; all event fields listed above are visible on each card
- [x] Each card expands to show active deadlines + "Manage deadlines" link
- [x] "Manage deadlines" link navigates to `/[locale]/events/[id]/deadlines`
- [x] Table updates when `activeFilters` changes (same fetch as globe)
- [x] Loading state visible in table mode
- [x] Empty state visible in table mode when no events match filters
- [x] Resize below `md` while in table mode → auto-switch to globe + toast appears
- [x] Toast appears only on forced resize, not on manual toggle
- [x] Rapid resize across `md` boundary → toast shown at most once per crossing
- [x] Selecting an event on globe then toggling to table → detail panel closes
- [x] All new strings go through `useTranslations` — no hardcoded UI text
- [x] All four locale files updated with matching keys
- [x] `pnpm typecheck` passes
- [x] `pnpm test` passes (useViewMode hook unit tests)
