# Management Dashboard Homepage

## Description

Shared event review queue for admin and moderator roles, rendered at `/manage/admin`
and `/manage/moderator`. Both pages use the same components and layout — only the
review route and the own-event restriction differ between roles.

## Routes

| Route | Who sees it |
|---|---|
| `/[locale]/manage/admin` | admin |
| `/[locale]/manage/moderator` | moderator |

Both routes already exist as stubs (from the manage-portal feature). This feature
replaces those stubs with the full dashboard implementation.

## Layout

```
┌─────────────────────────────────────────────┐
│ Yury Lima                        [Y] ▼      │  ← header (sticky, full-width)
├─────────────────────────────────────────────┤
│ Status ▼  Tier ▼  Year ____  [ Apply ]      │  ← filter row
├─────────────────────────────────────────────┤
│ ┌───────────────────────────────────────┐   │
│ │ MODELS 2026                [Review]   │   │  ← event card (normal)
│ └───────────────────────────────────────┘   │
│ ┌───────────────────────────────────────┐   │
│ │ ICSE 2026 (grey)  You can't review…  │   │  ← event card (own-event, moderator)
│ └───────────────────────────────────────┘   │
│ ...                                         │
├─────────────────────────────────────────────┤
│ ← Prev   Page 2 of 4 (98 events)   Next →  │  ← pagination row
└─────────────────────────────────────────────┘
```

---

## Header

- **Top-left:** user's full name (from `localStorage.manage_user.name`)
- **Top-right:** Avatar — a circle with the first letter of the user's name
  - Clicking the avatar opens a floating dropdown menu anchored below it
  - Menu contains a single "Sign out" button
  - Sign-out clears `localStorage.manage_user` and calls `logout()`, then redirects
    to `/manage`
  - Menu closes when the user clicks outside of it
- Header is sticky — stays at the top as the list scrolls

---

## Filters

Three filters in a horizontal row (stacked vertically on mobile):

| Filter | Type | Default | Options |
|---|---|---|---|
| Status | Select | `pending` | pending / approved / rejected |
| Tier | Select | all | all / A* / A / B / C / unranked |
| Year | Number input (stepper) | current year | any year |

- An **Apply** button triggers a fetch with the current filter values and resets
  the page to 1
- Filters do NOT fire on change — only on Apply
- On initial load, the page fetches with the defaults: `status=pending`,
  no tier filter, `year=<current year>`, `page=1`, `page_size=30`

---

## Event Card (normal)

One card per event in the list:

- **Left:** event name — rendered as a link that opens `event.website_url` in a new
  tab (`target="_blank" rel="noopener noreferrer"`)
- **Right:** "Review" button — navigates to `/${locale}/manage/${role}/events/${event.slug}/review`
  (role comes from `localStorage.manage_user.role`)

On mobile, the name and button stack vertically (name on top, button below).

## Event Card (own-event — moderator only)

When `event.created_by.id === sessionUser.id` and the current role is `moderator`:

- Entire card is greyed out (lower opacity, muted text colour)
- Event name is **not** a link (plain text)
- Review button is **replaced** by the text "You can't review this submission"
- Admin role never sees this variant — their cards are always normal

---

## Pagination

Below the list:

- **Prev** button — disabled and visually muted when on page 1
- **Page label** — "Page {page} of {totalPages} ({total} events)"
  - `totalPages = Math.ceil(meta.total / 30)`
  - Uses `meta.page` and `meta.total` from the API response
- **Next** button — disabled and visually muted when on the last page
- Clicking Prev/Next fetches the adjacent page with the current applied filters
  (not the draft filter values)

---

## API

Endpoint: `GET /api/v1/events` (public, no auth required)

Parameters sent:
```
status=<applied status>
year=<applied year>
tier=<applied tier>          ← omitted when "all" is selected
page=<current page>
page_size=30
pagination=on
```

Response shape (already typed as `EventListItem[]` + `ApiMeta`):
```json
{
  "code": "EVENTS_LISTED",
  "data": [...],
  "meta": { "page": 2, "total": 98 }
}
```

A new `listEvents(params)` function will be added to `src/lib/api/events.ts`
(create file). Uses `apiRequestWithMeta` from the client.

---

## Session User — localStorage Update

The own-event check requires the user's numeric ID. The current `manage_user`
localStorage entry does not store it. This feature adds `id` to the stored object.

Changes to `src/app/[locale]/manage/page.tsx` (login page):
- Store `id: parseInt(claims.sub, 10)` alongside name, role, email

Updated `SessionUser` interface (shared across all manage pages):
```typescript
interface SessionUser {
  id: number
  name: string
  role: "admin" | "moderator"
  email: string
}
```

---

## Mobile

- Header stays sticky; name truncates with ellipsis if long
- Filter row stacks vertically; each filter is full-width
- Event cards are full-width; name on top, button below
- Pagination row stacks: Prev, label (centered), Next

---

## i18n

New `manage.reviewDashboard.*` keys in all 4 locale files:

```
manage.reviewDashboard.statusLabel         — "Status"
manage.reviewDashboard.statusPending       — "Pending"
manage.reviewDashboard.statusApproved      — "Approved"
manage.reviewDashboard.statusRejected      — "Rejected"
manage.reviewDashboard.tierLabel           — "Tier"
manage.reviewDashboard.tierAll             — "All tiers"
manage.reviewDashboard.yearLabel           — "Year"
manage.reviewDashboard.applyButton         — "Apply"
manage.reviewDashboard.reviewButton        — "Review"
manage.reviewDashboard.cannotReview        — "You can't review this submission"
manage.reviewDashboard.pageLabel           — "Page {page} of {totalPages} ({total} events)"
manage.reviewDashboard.prevButton          — "← Prev"
manage.reviewDashboard.nextButton          — "Next →"
manage.reviewDashboard.loading             — "Loading events…"
manage.reviewDashboard.empty               — "No events match the current filters."
manage.reviewDashboard.errorLoad           — "Failed to load events. Try again."
```

---

## Files to Create / Modify

| File | Action |
|---|---|
| `src/lib/api/events.ts` | **Create** — `listEvents(params)` function |
| `src/lib/api/events.test.ts` | **Create** — unit tests for `listEvents` |
| `src/hooks/useReviewEvents.ts` | **Create** — filter + fetch + pagination state |
| `src/hooks/useReviewEvents.test.ts` | **Create** — unit tests |
| `src/components/manage/ManageHeader.tsx` | **Create** — sticky header with avatar menu |
| `src/components/manage/EventReviewCard.tsx` | **Create** — normal + greyed-out variants |
| `src/components/manage/ManageDashboard.tsx` | **Create** — filter row + list + pagination |
| `src/app/[locale]/manage/admin/page.tsx` | **Modify** — replace stub with ManageDashboard |
| `src/app/[locale]/manage/moderator/page.tsx` | **Modify** — replace stub with ManageDashboard |
| `src/app/[locale]/manage/page.tsx` | **Modify** — store `id` in localStorage on login |
| `src/messages/en.json` | **Modify** — add `manage.reviewDashboard.*` keys |
| `src/messages/pt.json` | **Modify** — add `manage.reviewDashboard.*` keys |
| `src/messages/es.json` | **Modify** — add `manage.reviewDashboard.*` keys |
| `src/messages/de.json` | **Modify** — add `manage.reviewDashboard.*` keys |

---

## Error Cases

| Scenario | Behaviour |
|---|---|
| API returns error on load | Show inline error message + retry not automatic (user re-applies filters) |
| Empty result set | Show "No events match the current filters." message |
| Session missing on mount | Redirect to `/manage` (existing guard, unchanged) |
| Role mismatch on URL | Redirect to correct dashboard (existing guard, unchanged) |

---

## Border / Corner Cases

- Moderator views a list that includes one of their own events → grey card, no link, no review button
- Admin views the same list → all cards normal, all review buttons active
- Year input accepts only integers; non-numeric input is ignored on Apply
- When total is 0, Prev and Next are both disabled and label shows "Page 1 of 1 (0 events)"
- Avatar menu closes on outside click and on Sign out
- Filter Apply while already on the filters' page (no change) → refetch page 1 regardless
- `tier=all` → omit `tier` param from request entirely (don't send `tier=all` to the backend)

---

## Definition of Done

- [ ] Header shows user name (top-left) and avatar with first initial (top-right)
- [ ] Avatar click opens float menu with Sign out button
- [ ] Sign out clears localStorage, calls logout, redirects to `/manage`
- [ ] Filter row shows status (default pending), tier (default all), year (default current year)
- [ ] Apply button fetches page 1 with selected filters
- [ ] 30 events per page (`page_size=30`)
- [ ] Normal card: event name opens `website_url` in new tab; Review button navigates to role-based route
- [ ] Moderator own-event card: entire card grey, no link, "You can't review this submission" shown
- [ ] Admin never sees own-event grey variant
- [ ] Prev/Next buttons navigate pages; both disabled on single-page results
- [ ] Page label shows current page, total pages, total event count
- [ ] Apply resets to page 1
- [ ] Empty state message shown when API returns 0 results
- [ ] Error state shown when API call fails
- [ ] `id` stored in `manage_user` localStorage from login
- [ ] All `manage.reviewDashboard.*` i18n keys in all 4 locale files
- [ ] Layout is mobile-responsive (filters stack, cards stack)
- [ ] `listEvents` has unit tests
- [ ] `useReviewEvents` has unit tests covering: initial load, filter apply, page navigation, own-event detection
