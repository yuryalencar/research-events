# Globe Event Filters

## Description
A collapsible floating filter panel anchored to the top-left of the globe
homepage. Lets users narrow the globe pins by year, domain, tier, country,
and first deadline month. All filters combine with AND on the backend.
`status` is always fixed to `approved` — never exposed as a filter.

## Behaviour

### Panel visibility
- **Desktop (≥ 768 px):** expanded by default on first load.
- **Mobile (< 768 px):** collapsed by default; a single icon button toggles it open.
- The toggle state is local UI state — it resets to the default on page reload.
- When expanded, the panel overlays the globe (does not push content).

### Filter controls (in panel order)
1. **Year** — two arrow buttons (← prev / → next) with the active year
   displayed between them. Min year: 2000. Max year: current year + 5.
   **Year is mandatory** — every API call always includes a specific year.
   There is no "All years" option and year cannot be cleared or unset.
2. **Domain** — single-select dropdown. Options: "All domains" (no filter) +
   all registered domains (currently: `computer_science`). Extensible — new
   domains added to the backend enum appear here without frontend code
   changes, driven by the `DOMAINS` constant.
3. **Tier** — single-select chip row. Options: "All" (no filter) + `A*`, `A`,
   `B`, `C`, `Unranked`. Only one chip is active at a time (matches backend
   single-value `tier` param).
4. **Country** — single-select dropdown built from a hardcoded comprehensive
   list of world countries (`COUNTRIES` constant in `src/lib/countries.ts`).
   Uses the browser's native type-to-search behaviour (no separate search
   input). Sends the selected country name as-is; backend does
   case-insensitive exact match.
5. **First deadline month** — single-select dropdown. Options: "All months"
   (no filter) + January through December.

### Draft vs. applied state
- Changing any filter control updates **draft state** only — the globe does
  not re-fetch until the user clicks Apply.
- **Apply button:** copies draft → applied filters → triggers `useEvents`
  re-fetch with new params. Year is always present in the applied params.
  On mobile, the panel auto-closes after Apply so the globe is fully visible.
- **Reset button:** restores the draft form to defaults (`year = current year`,
  all optional filters cleared) **without changing the applied state** — the
  user must still click Apply to trigger a re-fetch with the reset values.
- A visual indicator (dot badge on the toggle button) shows when draft
  differs from applied (i.e. any optional filter is set, or year ≠ applied
  year). Dot remains visible even when the panel is collapsed.
- **Globe rotation:** after each filter-triggered Apply completes and at
  least one event is returned, the globe animates to the first result
  (preserving the user's current zoom level). The initial page-load fetch
  does not trigger rotation.

### Panel placement
- `fixed top-4 left-4 z-40` — same z-level as the InfoButton (bottom-left).
- Does not conflict with the event detail Sheet (right side) or InfoButton
  (bottom-left).
- On mobile, when expanded the panel overlays the top-left area of the globe
  but does not cover the bottom Drawer trigger zone.

## Rules
- `year` is **always** included in every `listEvents` call — it is never
  omitted, never undefined in filter state. Initial value: current year.
- `status` is always `approved` — never sent as a user-controlled param.
- `pagination` is always `off` — the globe needs all matching pins.
- Year limits: min 2000, max current year + 5. Prev/next buttons are
  disabled (not hidden) at the bounds.
- Country list is a hardcoded constant (`COUNTRIES` in `src/lib/countries.ts`)
  — approximately 195 entries using standard English country names.
- Domain list is a hardcoded constant (`DOMAINS`) — must stay in sync with
  `allowedDomains` in the backend service.
- Tier list is a hardcoded constant (`TIERS`): `["A*", "A", "B", "C",
  "unranked"]`.
- All user-facing strings go through next-intl (en, pt, es, de).

## Permissions
- Public — no authentication required.
- Rendered only on the globe homepage.

## New hooks
- `useFilters(currentYear)` — manages draft + applied filter state. `year`
  in both draft and applied is always a `number`, never `undefined`. Returns
  setters for each field, `apply()`, `reset()`, `activeFilters`, `isDirty`.
- `useEvents` is modified to accept `EventFilters` (the applied state shape)
  and re-fetches whenever `activeFilters` changes.

## Error cases
| Scenario | Expected behaviour |
|---|---|
| API returns 400 for an invalid filter | `handleApiError` shows a toast; applied filters unchanged |
| API returns network error | `handleApiError` shows a toast; events list left empty |

## Border / corner cases
- Reset while a fetch is in flight — the in-flight fetch result is discarded
  (isMounted guard already in useEvents).
- Year prev/next buttons are disabled (not hidden) at min (2000) and max
  (current year + 5).
- Country dropdown with no selection sends no `country` param to API.
- `first_deadline_month` with no selection sends no param.
- Domain "All domains" selected sends no `domain` param.
- Tier "All" sends no `tier` param.
- `isDirty` is true when applied filters differ from defaults — year ≠
  current year OR any optional filter is set.
- Panel open on mobile when event-detail Drawer opens — panel stays open
  (Drawer slides from bottom, filter panel is top-left, they do not overlap).
- Apply with year only (all optional filters cleared) is valid and must work.

## Definition of done
- [ ] Floating panel visible at top-left on globe homepage only
- [ ] Desktop: expanded by default; mobile: collapsed by default
- [ ] Toggle button collapses/expands the panel on both breakpoints
- [ ] Year prev/next arrows change draft year; disabled at bounds (2000 / current+5)
- [ ] Year is always present in every API call — never omitted
- [ ] Domain single-select dropdown, "All domains" = no filter sent
- [ ] Tier single-select chip row, "All" = no filter sent
- [ ] Country searchable dropdown (≈195 countries), no selection = no filter sent
- [ ] First deadline month dropdown (All + Jan–Dec), no selection = no filter sent
- [ ] Apply button triggers re-fetch with current draft filters (year always included)
- [ ] Reset button restores draft form to defaults without changing applied state
- [ ] Apply button after Reset triggers re-fetch with reset values
- [ ] Active-filter indicator shown on toggle button when draft ≠ applied
- [ ] Globe rotates to first matching event after each filter Apply (not on initial load)
- [ ] "No events" message distinguishes between empty platform and empty filter results
- [ ] `useFilters` has Vitest unit tests covering setters, apply, reset, isDirty,
      and the invariant that year is never undefined
- [ ] `useEvents` tests updated for the new `filters` parameter
- [ ] All strings translated in en, pt, es, de with no missing keys
- [ ] `pnpm typecheck` passes with zero errors
- [ ] `pnpm test` passes with zero failures
