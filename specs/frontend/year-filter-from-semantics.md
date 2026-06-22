# Year filter: "from year" semantics

## What changes
- year filter means "this year and onwards", not "exactly this year"
- year can be cleared (undefined) to fetch all events with no year constraint
- Default on page load: current year (same as before)

## Behaviour
- Year stepper shows current year value OR "All years" when cleared
- A clear button (×) next to the stepper sets year to undefined
- Decrement / increment arrows work as before when a year is set
- Decrement is disabled when year is undefined
- reset() restores year to current year (same default as page load)

## Label changes (all 4 locales)
- filters.yearLabel: "Year" → "From year"
- manage.filters.yearLabel: "Year" → "From year"
- New key filters.allYears: "All years" (placeholder shown when year is undefined)
- New key manage.filters.allYears: "All years"
- prevYear / nextYear aria-labels unchanged

## Screens affected
- FilterPanel.tsx (globe / table homepage)
- ManageDashboard.tsx (admin/moderator review queue)

## Type changes
- EventFilters.year: number → number | undefined
- ReviewFilters.year: number → number | undefined
- useFilters: setYear(year: number | undefined)
- useReviewEvents: setDraftYear(year: number | undefined)
- toListEventsParams: omits year key entirely when undefined

## Definition of done
- [ ] GET /events?year=2026 returns 2026+ events
- [ ] GET /events with no year returns all events
- [ ] Globe/table page loads with current year pre-filled (same as before)
- [ ] Clearing year shows "All years" and fetches all events
- [ ] Reset restores year to current year
- [ ] "From year" label shown on both globe FilterPanel and ManageDashboard
- [ ] All 4 locales have the new keys (en/pt/es/de)
- [ ] TypeScript strict mode passes (pnpm typecheck)
- [ ] All affected tests updated
