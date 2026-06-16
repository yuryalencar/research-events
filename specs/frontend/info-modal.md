# Info Modal (floating button)

## Description
A floating info button fixed to the bottom-left of the globe homepage. When
clicked it opens a centered modal (desktop) or bottom sheet (mobile) with
static information about the project: pin legend, event submission flow,
moderator contact, author credit, and open-source contribution.

## Behaviour
- Button is always visible on the globe homepage, **except** when the event
  detail Drawer is open on mobile — in that case the button is hidden.
- Clicking the button opens the info modal.
- Modal closes on: explicit close button click, backdrop click, or Escape key
  (same pattern as EventDetailView Sheet/Drawer).
- On desktop (≥ 768 px): centered Dialog overlay.
- On mobile (< 768 px): bottom Drawer (same Vaul Drawer used by EventDetailView).
- Modal content is scrollable when it overflows the viewport.

## Content sections (in order)
1. **Version** — label + value from `APP_VERSION` constant.
2. **Map legend** — three rows with a colored dot + label + one-line description:
   - Yellow dot → upcoming / ongoing conference (end date not yet passed)
   - Red dot → past conference (end date has already passed)
   - Pink dot → currently selected conference (clicked pin)
3. **How events are added** — prose paragraph explaining the full flow:
   submitter provides details → submission enters pending state → moderator
   reviews → moderator cannot review their own submissions → collaborative
   tool for the research community.
4. **Become a moderator** — short paragraph + mailto link to `ADMIN_EMAIL`.
5. **About the author** — "Created by Yury Lima" with a link to
   `GITHUB_PROFILE_URL`.
6. **Contribute** — "ReSEARCH Events is open source" + link to
   `GITHUB_REPO_URL`. Invites the community to collaborate.

All external links open in a new tab with `rel="noopener noreferrer"`.

## Rules
- All user-facing strings go through next-intl (en, pt, es, de).
- Button is hidden (not merely invisible) when the mobile Drawer is open —
  use the `selectedEvent` state already available in `page.tsx` to
  conditionally render the button.
- Constants (`APP_VERSION`, `GITHUB_REPO_URL`, `GITHUB_PROFILE_URL`,
  `ADMIN_EMAIL`) live in `src/lib/constants.ts`.
- No API calls — all content is static.
- Follows the existing component file template (types → component →
  sub-components → export).

## Permissions
- Public — no authentication required.
- Visible to all users regardless of role.
- Rendered only on the globe homepage (`app/[locale]/page.tsx`) — not on event detail, submit, or admin pages.

## Error cases
| Scenario | Expected behaviour |
|---|---|
| External link fails to load | Browser default — no special handling needed (static href) |

## Border / corner cases
- Button hidden while event detail Drawer is open on mobile.
- Modal is scrollable if content overflows (long translations).
- Button does not overlap with event detail Sheet on desktop (Sheet slides
  from the right; button is bottom-left — no overlap).
- All four locales must have every translation key — no missing keys allowed.

## Definition of done
- [ ] Floating `ⓘ` button visible at bottom-left of globe homepage
- [ ] Button hidden when mobile event-detail Drawer is open
- [ ] Clicking button opens centered Dialog on desktop (≥ 768 px)
- [ ] Clicking button opens bottom Drawer on mobile (< 768 px)
- [ ] Backdrop click closes the modal
- [ ] Escape key closes the modal
- [ ] All 6 content sections rendered in correct order
- [ ] Colored dots (yellow / red / pink) match the exact hex values in `lib/events.ts`
- [ ] Mailto link uses `ADMIN_EMAIL` constant
- [ ] Author and repo links open in new tab with `rel="noopener noreferrer"`
- [ ] All strings translated in en, pt, es, de with no missing keys
- [ ] `pnpm typecheck` passes with zero errors
