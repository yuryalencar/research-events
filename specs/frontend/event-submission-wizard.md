# Event Submission Wizard

## Description
A 3-step wizard flow that lets any public user submit a new research conference for admin
review. Entry point is a floating "+" button at the top-right of the globe homepage. The
flow navigates to `/events/submit` (a full page) and guides the user through:

1. **Step 1 — Duplicate check**: search existing pending events to avoid re-submitting
2. **Step 2 — Event details**: submitter contact info + full event data + Leaflet map pin
3. **Step 3 — Deadlines (optional)**: add submission deadlines; skippable

On success → confirmation screen with submitted event details + "Back to homepage" button.

---

## Entry point — Float button (homepage)

- Position: `fixed top-4 right-4 z-40` — always visible on mobile and desktop
- Icon: `+`
- Tooltip on hover (desktop): "Add a new event" (i18n: `submit.addEventTooltip`)
- `aria-label`: same string as tooltip, for screen readers
- Mobile: no hover — tooltip not shown; `aria-label` still applies
- Does **not** hide when EventDetailView drawer is open (unlike InfoButton)
- Click → navigate to `/<locale>/events/submit` (full page, not modal)

---

## Step 1 — Duplicate check

### Purpose
Let the user verify their event is not already in the system (as a pending submission)
before filling out the form.

### Controls
| Control | Behaviour |
|---|---|
| Text search field | Filters by event name OR slug — client-side on current page results |
| Year filter | Integer; defaults to current year |
| Apply button | Always re-fetches at **page 1** with current filter values |
| Clear button | Resets text and year to defaults, re-fetches at page 1 |
| "Continue anyway" (primary) | Always visible; proceeds to Step 2 regardless of results |

> Apply pressed with all fields empty → re-fetch at page 1 with no filters applied (same
> as the initial load). Pagination never preserves stale offsets when filters change.

### Table columns
| Column | Notes |
|---|---|
| Full event name | Link — opens event detail in new tab |
| Slug | Monospace style |
| Year | |
| Country | |

*(No status column — all rows are pending by definition in this version.)*

### Pagination
- Page size: 10 rows/page
- Controls: Previous / Next buttons + "Page N of M"
- Total count shown: "N events found" (from `meta.total`)
- Previous/Next re-fetch at the new page number without resetting filters

### Data source
```
GET /api/v1/events?status=pending&year=<year>&page=<page>&page_size=10
```
- `meta.total` drives the total count and page calculation
- Text search is client-side on the 10 rows returned for the current page
  (the backend has no `?q=` param; backend search is a future enhancement)
- Year filter and text search both always reset to page 1 on Apply

### Empty & error states
- No results → "No events found. You're likely the first to submit this one —
  continue to Step 2."
- API error → "Couldn't load the events list. You can still continue to Step 2."
  (Continue button still available)
- System has zero pending events at all → "No pending events yet — you're the
  first! Continue to Step 2."

---

## Step 2 — Event details

### Submitter section
| Field | Required | Notes |
|---|---|---|
| Your full name | yes | `submitter.name` |
| Your email | yes | `submitter.email`; simple regex |

### Event section
| Field | Required | Notes |
|---|---|---|
| Full event name | yes | Label: "Full event name" — encourages complete official name |
| Slug | yes | Hint below field: "Short version + year, e.g. MODELS2026 — letters, numbers, hyphens, underscores only." |
| Domain | yes | Select; currently only "Computer Science" |
| Tier | no | Select: A*, A, B, C, Unranked; placeholder "Unranked — I'm not sure"; CORE ranking |
| Start date | yes | |
| End date | yes | Must be ≥ start date |
| Website URL | yes | Must start with http:// or https:// |
| Country | yes | Select from `COUNTRIES` list in `lib/countries.ts` |
| City | yes | Free text |
| Location pin | yes | Leaflet map — drop a pin to set lat/lng; lat/lng shown read-only below map |

### Leaflet map
- Loaded via `dynamic(() => import(...), { ssr: false })` — never server-rendered
- Default view: world map centered at [20, 0], zoom 2
- Selecting a country from the dropdown pans the map to that country's approximate center
- User clicks to place a pin; pin is draggable after placement
- Lat/lng update in real time as pin moves
- Height: 300px desktop / 220px mobile

### Client-side validation (on "Continue to Step 3")
- All required fields filled
- Slug matches `^[A-Za-z0-9_-]+$`
- End date ≥ start date
- Website URL starts with http:// or https://
- Email is valid format
- Lat/lng set (pin dropped on map)

### Navigation
- "Back to Step 1" link — returns to Step 1 with state preserved in wizard component
- "Continue to Step 3" primary button — triggers validation before advancing

---

## Step 3 — Deadlines (optional)

- Empty state: "No deadlines added yet. You can skip this step or add one below."
- "+ Add deadline" button adds a row:
  - Type: select (`abstract | paper | notification | camera_ready | other`)
  - Description: text (e.g. "Research track")
  - Date: date input (required per row)
  - Optional: checkbox
  - Remove: × icon button
- Each row requires: type, description, date
- **"Skip & Submit"** (secondary): submits event with `deadlines: []`
- **"Submit"** (primary): submits with current rows; disabled if any row has a validation error

### On submission
1. `POST /api/v1/events/submit` with the payload from Steps 2 + 3
2. **201** → navigate to success screen
3. **409** → return to Step 2; slug field shows error "This slug is already taken by a
   pending or approved event. Choose a different one."
4. **400** → return to Step 2; per-field error messages shown
5. **429** → banner on Step 3: "Too many submissions. Please wait a moment."
6. Network error → banner on Step 3: "Couldn't reach the server. Check your connection."

---

## Success screen

Shown after 201 response.

| Element | Value |
|---|---|
| Heading | "Event submitted!" |
| Subtext | "Your event is now pending review. A moderator will check it before it appears on the globe." |
| Summary card | Full event name, slug, dates, website (link), status badge "Pending review", submitter name |
| Primary button | "Back to homepage" → `/<locale>/` |

---

## Permissions
Public — no authentication required. Behaviour is identical regardless of the
submitter's existing account role.

---

## Error cases
| Scenario | Expected behaviour |
|---|---|
| Slug already taken (409) | Return to Step 2; slug field error |
| Validation error (400) | Return to Step 2; per-field errors |
| Rate limit (429) | Banner on Step 3 |
| Network error | Banner on Step 3 |
| Step 1 events list fails to load | Warning inline + Continue still works |
| End date < start date | Inline error on end date field (Step 2) |
| Invalid slug characters | Inline error on slug field (Step 2) |
| No pin dropped (lat/lng unset) | Inline error on map section (Step 2) |

---

## Border / corner cases
- Direct navigation to `/events/submit` → wizard starts at Step 1 normally
- Browser refresh mid-wizard → restarts at Step 1, form data lost (no persistence in v1)
- Step 1 is empty (zero pending events) → empty state + Continue still available
- Slug reused from a rejected event → 201 (backend allows it; no special UX needed)
- All deadline rows removed before Submit → submits with `deadlines: []` — allowed
- Both Step 1 API calls fail → warning shown + Continue still available
- Apply pressed with all fields cleared → re-fetch page 1, no filter applied
- Selecting a country pans the map but does not set the pin — user must still click

---

## i18n keys added (en.json namespace: `submit`)
```
submit.addEventButton     — aria-label for the float button
submit.addEventTooltip    — tooltip text: "Add a new event"
submit.step1.*            — step 1 labels, table headers, empty/error states
submit.step2.*            — step 2 field labels, hints, map instruction
submit.step3.*            — step 3 labels, deadline row fields, submit buttons
submit.success.*          — success screen heading, subtext, summary labels, CTA
submit.stepIndicator      — "Step {current} of {total}"
submit.back               — "Back" navigation label
```

---

## Definition of done
- [ ] "+" float button at top-right of homepage, visible on mobile and desktop
- [ ] Tooltip "Add a new event" appears on hover (desktop only)
- [ ] Button navigates to `/<locale>/events/submit` (full page)
- [ ] Step 1 fetches `?status=pending` with server-side pagination (page_size=10)
- [ ] `meta.total` drives page count and "N events found" label
- [ ] Text search filters by name or slug client-side on current page
- [ ] Year filter and Apply always reset to page 1
- [ ] Clear resets fields and re-fetches at page 1
- [ ] "Continue anyway" always proceeds to Step 2
- [ ] Step 1 empty state and API error state handled; Continue still works in error state
- [ ] Step 2 collects all required fields with inline validation on Continue
- [ ] Leaflet map loads without SSR errors
- [ ] Pin placement sets lat/lng; values shown read-only below map
- [ ] Country select pans map to country center
- [ ] Validation blocks Continue if lat/lng not set
- [ ] Step 3 allows adding / removing deadline rows
- [ ] "Skip & Submit" sends `deadlines: []`
- [ ] "Submit" disabled when any deadline row has a validation error
- [ ] `POST /api/v1/events/submit` called with correct payload shape
- [ ] 201 → success screen with submitted event details
- [ ] "Back to homepage" navigates to `/<locale>/`
- [ ] 409 → Step 2 with slug field error
- [ ] 400 → Step 2 with per-field errors
- [ ] 429 → banner on Step 3
- [ ] Network error → banner on Step 3
- [ ] Back navigation preserves wizard state
- [ ] Browser refresh restarts wizard at Step 1
- [ ] All i18n strings present in `en.json`, `pt.json`, `es.json`, `de.json`
- [ ] `pnpm typecheck` passes with no errors
- [ ] `pnpm test` passes
