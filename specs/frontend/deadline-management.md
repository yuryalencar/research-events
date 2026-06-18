# Deadline Management

## Description
Single-page deadline management view at /events/[slug]/deadlines. Users make
all changes (add / supersede / cancel) on one page and submit everything at
once. No redirects during editing. After a successful submission the same page
renders a success state (same design as the event submission SubmitSuccess).
No auth required; one contributor section (name + email) covers all operations
in the batch.

## Navigation & state passing
- Entry: pencil icon on each active deadline card inside EventDetailContent
  (visible in both the drawer on mobile and the sheet on desktop)
- Before navigating, the homepage writes the full EventListItem to
  sessionStorage under the key "deadline_management_event"
- The page reads that key on mount; if absent or the stored slug does not
  match the URL slug → toast "Please select an event from the homepage
  first" → redirect to /[locale]

## Page states
Three mutually exclusive render states (same URL throughout):
1. Editing   — default; user makes changes
2. Submitting — loading overlay while API calls run
3. Success   — success component after all calls complete

## Page wrapper
- Same layout as the submission wizard: mx-auto max-w-4xl px-4 py-8
- Header: event name as h1 + "Manage Deadlines" subtitle
- Success state uses mx-auto max-w-2xl px-4 py-8 (same as SubmitSuccess)

## Editing state layout
Top to bottom:
1. Header (event name + subtitle)
2. Deadline list — active deadlines from sessionStorage, each as a card
3. "+ Add deadline" dashed button (same style as Step3Deadlines)
4. Contributor section — name (required) + email (required, valid email)
5. Bottom navigation bar — same `flex justify-between border-t border-border pt-4` pattern:
   - Left: "← Back to homepage" secondary button (same class as Step3 back button)
   - Right: "Submit changes" primary button (same class as Step3 submit button)
     disabled when nothing changed or any validation error exists

## Deadline cards — existing deadlines
Adopt the same grid/border pattern as DeadlineRow in Step3Deadlines
(`rounded-md border border-border p-4`) extended with state-specific content.

Three card states:

**Default:**
  Shows: type badge, description, date (+ time + timezone if set), Optional badge
  Right side: pencil icon button + X icon button

**Supersede-editing** (pencil clicked):
  Read-only: type badge, description, Optional badge
  Editable fields (same inputClass as Step3):
    date (date input, required) | time (text, optional, HH:MM) | timezone (text, optional)
  Right side: "Revert" text link — restores original values and exits edit mode

**Pending-cancel** (X clicked):
  Entire card text has line-through styling
  "Pending cancellation" label visible
  Right side: "Revert" button — returns card to Default state

Transition rules:
  - X on Supersede-editing → exit edit + enter Pending-cancel
  - Pencil on Pending-cancel → revert cancel + enter Supersede-editing
  - A card cannot be in both states simultaneously

## New deadline cards (from "+ Add deadline")
Same DeadlineRow grid/border layout as Step3Deadlines, with time and timezone
fields added after date:
  type (select, required) | description (text, required) | date (date, required)
  time (text, optional, HH:MM) | timezone (text, optional) | is_optional (checkbox)
  X button to remove the card before submitting

Empty state when no existing deadlines and no new cards added:
  Same style as Step3: `rounded-md border border-border bg-muted/30 px-4 py-6
  text-center text-sm text-muted-foreground`

## Contributor section
Below the deadline list; above the navigation bar.
Same field style (inputClass) and layout as the submitter section in Step2Details.
Fields: name (required), email (required, valid format).
Applies to all operations submitted in the batch.

## Submit button behavior
- Disabled when: no changes detected OR any validation error exists
- On click with no changes → toast "Nothing to submit. Make a change first."
- On click with validation errors → inline errors shown per field; submit stops
- On click with valid changes → enter Submitting state

## "Changed" detection
A supersede-edit counts as a change only if the resulting date/time/timezone
differs from the original deadline's stored values. Opening edit mode and
saving without changing anything is not counted as a change.
A removed new card is not counted as a change.

## Submission logic
1. Validate all fields. Stop with inline errors if any are invalid.
2. Enter Submitting state (loading overlay, submit shows "Submitting…",
   all controls disabled).
3. Call all endpoints in parallel:
   - Each pending-cancel → cancelDeadline(eventId, id, {submitter})
   - Each supersede-edit with changes → supersedeDeadline(eventId, id, {submitter, date, time?, timezone?})
   - New cards (if any) → addDeadlines(eventId, {submitter, deadlines: [...]}) — one batch call
4. All succeed → render Success state
5. Any fail → error toast + exit Submitting → return to Editing with all pending changes intact

## Validation rules
New deadline cards:
  - type: required
  - description: required
  - date: required; must not be after event.end_date
  - time: if provided, must match HH:MM (00:00–23:59)
  - timezone: if provided, must not be empty string

Supersede-editing cards:
  - date: required; must not be after event.end_date
  - time: if provided, must match HH:MM
  - timezone: if provided, must not be empty string

Contributor:
  - name: required, non-empty
  - email: required, valid email format

Inline errors shown only after a submit attempt; not on blur.

## Success state
Same design as SubmitSuccess (CheckCircleIcon + centered layout + details card):
  - Icon: CheckCircleIcon size-16 text-green-500
  - Heading: "Deadlines updated!"
  - Details card (same `rounded-lg border bg-card` style):
      Event name
      Summary row: "X added · Y updated · Z cancelled"
  - "Back to homepage" primary button → Link to /[locale]

## Loading state
Full-page overlay with centered spinner while API calls run.
All buttons disabled. Submit button label → "Submitting…"

## Error cases
| Scenario | Expected behaviour |
|---|---|
| No sessionStorage / slug mismatch on load | Toast + redirect to /[locale] |
| Submit with no changes | Toast "Nothing to submit. Make a change first." |
| Field validation error on submit | Inline error per field; submit stops |
| Any API call fails | Error toast + exit loading; editing state preserved |
| 409 DEADLINE_ALREADY_INACTIVE | Error toast "One or more deadlines were already modified by someone else." |
| 409 EVENT_NOT_APPROVED | Error toast using existing errors.EVENT_NOT_APPROVED i18n key |
| Network error | Error toast using existing errors.NETWORK_ERROR i18n key |

## Border / corner cases
- Add then immediately Remove a new card → not counted as a change
- Supersede-edit with identical values to original → not counted as a change
- Cancel all active deadlines → allowed; success shows "Z cancelled"
- 0 active deadlines initially → empty state shown; "+ Add deadline" still accessible
- Long description → line-clamp in Default card view, full text in Supersede-editing form

## i18n
All strings in en/pt/es/de under `deadlines.manage.*` namespace

## Responsive
Same page and components for mobile and desktop.
DeadlineRow grid collapses to single column on mobile (same as Step3Deadlines).

## New files
- `app/[locale]/events/[slug]/deadlines/page.tsx`
- `components/events/deadlines/DeadlineManagePage.tsx`      — main orchestrator + page states
- `components/events/deadlines/DeadlineCard.tsx`            — existing deadline card (3 states)
- `components/events/deadlines/AddDeadlineCard.tsx`         — new deadline form card
- `components/events/deadlines/DeadlineManageSuccess.tsx`   — success state
- `hooks/useDeadlineManage.ts`                              — change tracking + submit logic

## Modified files
- `components/events/EventDetailContent.tsx`    — pencil icon + onManageDeadlines prop
- `components/events/EventDetailView.tsx`       — implements onManageDeadlines (sessionStorage + navigate)
- `messages/en.json, pt.json, es.json, de.json` — deadlines.manage.* keys

## No backend work needed
All three endpoints, API client functions (addDeadlines, cancelDeadline,
supersedeDeadline), and TypeScript types already exist.

## Definition of done
- [ ] Pencil icon on each active deadline in EventDetailContent (drawer + sheet)
- [ ] Click stores event in sessionStorage and navigates to /events/[slug]/deadlines
- [ ] Direct access without sessionStorage or slug mismatch → toast + redirect
- [ ] Active deadlines rendered as cards (same grid/border as DeadlineRow)
- [ ] Pencil → Supersede-editing: date/time/timezone editable, type/desc/is_optional read-only
- [ ] Revert on Supersede-editing restores original values and exits edit mode
- [ ] X → Pending-cancel: strikethrough card + "Revert" button
- [ ] Revert on Pending-cancel → Default state
- [ ] X on Supersede-editing → exit edit + Pending-cancel
- [ ] Pencil on Pending-cancel → revert cancel + Supersede-editing
- [ ] "+ Add deadline" appends a new card (same grid + time/timezone fields added)
- [ ] Remove on new card discards it; not counted as a change
- [ ] Contributor section (name + email) with same inputClass styling
- [ ] Submit disabled when nothing changed or validation error present
- [ ] Toast "Nothing to submit" when submit clicked with no changes
- [ ] Inline validation errors shown per field after submit attempt
- [ ] date > event.end_date → inline error
- [ ] time invalid format → inline error
- [ ] Supersede with same values as original → not counted as a change
- [ ] Loading overlay shown during submission; all controls disabled
- [ ] Submit label → "Submitting…" during loading
- [ ] All succeed → success state on same page
- [ ] Any fail → error toast + return to editing with changes intact
- [ ] Success: CheckCircleIcon + heading + details card + change summary + "Back to homepage"
- [ ] Bottom nav: "← Back to homepage" left; "Submit changes" right (same Step3 button classes)
- [ ] All strings in en/pt/es/de under deadlines.manage.*
- [ ] pnpm typecheck passes, pnpm lint passes
