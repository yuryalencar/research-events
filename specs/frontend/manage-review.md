# Management Event Review — Frontend

## Description

A 3-step review wizard at `/manage/admin/events/[slug]/review` and
`/manage/moderator/events/[slug]/review`. Mirrors the submission wizard in structure.
Step 1 edits event data, Step 2 makes the approve/reject decision, Step 3 manages
deadlines.

## Wizard Steps

### Persistent header
`ManageHeader` (sticky, "Welcome {name}" + role + avatar sign-out menu) is rendered
above the `StepIndicator` and step content on all three steps — identical to the
dashboard pages.

### Step 1 — Edit Event Details
- Pre-filled editable form: name, country, city, start date, end date, website URL,
  domain, tier
- Leaflet `LocationPicker` pre-pinned at current lat/lng; dragging updates form state
- Status badge (non-editable pill) shown above the form:
  Pending → yellow, Approved → green, Rejected → red
- Blind review — no submitter or contributor info shown anywhere
- Navigation: "← Back to Dashboard" (small link, returns to role dashboard),
  "Next →" button (validates required fields before advancing)

### Step 2 — Review Decision
- Status badge shown prominently
- "Approve" button → opens ApproveModal (optional note textarea + confirm)
- "Reject" button → opens RejectModal (required reason textarea + confirm;
  confirm disabled until ≥1 non-whitespace character)
- Both modals bundle any field edits from Step 1 into the `event` field of the
  PATCH body — edits are always submitted together with the decision
- Approve success → show ReviewSuccess screen (replaces wizard content)
- Reject success → show ReviewSuccess screen (replaces wizard content)
- "Review Deadlines →" button: enabled when `status === "approved"`, jumps directly
  to Step 3; disabled with message "Approve the event first to manage its deadlines"
  (useful on re-review visits where the event is already approved)
- Navigation: "← Back" (to Step 1)

### ReviewSuccess screen
Shown in place of the wizard steps after a successful approve or reject decision.
Mirrors the pattern of `SubmitSuccess` and `DeadlineManageSuccess`.

- **Approve**: green `CheckCircleIcon`, "Event approved!" title, card showing event
  name + Approved status badge. Buttons: "Manage Deadlines" (advances to Step 3)
  + "Back to Dashboard" (navigates to role dashboard)
- **Reject**: red `XCircleIcon`, "Event rejected" title, card showing event name +
  Rejected status badge + reason text. Button: "Back to Dashboard" only

### Step 3 — Deadlines
- Conference URL shown as a clickable link at the top of the step (opens new tab)
- Same deadline management UI as the public `DeadlineManagePage`:
  existing deadline cards (cancel / supersede) + add new deadlines
- Contributor section **not shown** — name and email are auto-filled from
  `localStorage.manage_user` and sent with API calls transparently
- Navigation: "← Back to Review" (to Step 2), "Done →" (navigates to role dashboard)

## Auth Guard (applied on mount)
- No `manage_user` in localStorage → redirect to `/[locale]/manage`
- `manage_user.role` doesn't match URL role segment → redirect to
  `/[locale]/manage/[stored-role]`
- No event in `sessionStorage["manage_review_event"]` → redirect to role dashboard
  (covers direct URL navigation)

## Event Data Source
- `EventReviewCard` writes the full `EventListItem` to
  `sessionStorage["manage_review_event"]` before navigating to the review URL
- Wizard reads from sessionStorage on mount — no extra API call at load time
- Direct URL navigation without sessionStorage → redirect to role dashboard

## Permissions
- Admin: can review any event
- Moderator: cannot review own submissions — blocked in the dashboard (no Review
  button shown on own-event cards); no additional guard needed on this page
- Both roles use `apiPrivateRequest` — JWT is in an HTTP-only cookie managed by the
  browser

## Error cases

| Scenario | Expected behaviour |
|---|---|
| Not authenticated | Redirect to `/[locale]/manage` |
| Wrong role in URL | Redirect to correct dashboard |
| No event in sessionStorage | Redirect to role dashboard |
| PATCH fails (network / 5xx) | Error toast, stay on Step 2, form preserved |
| PATCH returns 401 | Redirect to `/[locale]/manage` |
| PATCH returns 400 | Show field errors inline on Step 1 fields |
| Deadline API call fails | Error shown inline on Step 3 (same as public deadline page) |

## Border / corner cases
- **Re-review**: page loads with current status in badge; form pre-filled with current
  data; both Approve and Reject always available regardless of current status
- **Already-approved event**: "Review Deadlines →" on Step 2 is enabled immediately;
  user can skip to Step 3 without re-approving
- **Partial edits**: only changed fields sent in PATCH `event` body; unchanged fields
  omitted
- **Reject reason whitespace only**: confirm button stays disabled (trim check)
- **Wizard step navigation**: step state is held in component memory — browser refresh
  returns the user to Step 1

## Definition of done
- [ ] Auth guard: unauthenticated → `/[locale]/manage`
- [ ] Auth guard: wrong role → correct dashboard
- [ ] No sessionStorage → redirect to role dashboard
- [ ] `ManageHeader` visible and sticky across all three wizard steps
- [ ] Status badge shows correct colour for pending / approved / rejected
- [ ] Status badge updates in place after successful approve/reject
- [ ] Step 1: all fields editable (name, country, city, dates, website, domain, tier)
- [ ] Step 1: `LocationPicker` pre-pinned at current lat/lng; dragging updates form state
- [ ] Step 1: "Next →" validates required fields before advancing
- [ ] Step 2: "Approve" opens modal with optional note textarea
- [ ] Step 2: "Reject" opens modal with required reason textarea
- [ ] Step 2: reject confirm disabled until reason has ≥1 non-whitespace character
- [ ] Step 2: Approve success → ReviewSuccess screen with "Manage Deadlines" + "Back to Dashboard"
- [ ] Step 2: Reject success → ReviewSuccess screen with reason shown + "Back to Dashboard" only
- [ ] ReviewSuccess: approve variant shows green CheckCircleIcon + Approved badge
- [ ] ReviewSuccess: reject variant shows red XCircleIcon + Rejected badge + reason
- [ ] ReviewSuccess: "Manage Deadlines" advances to Step 3 (approve only)
- [ ] ReviewSuccess: "Back to Dashboard" navigates to role dashboard
- [ ] Step 2: "Review Deadlines →" disabled when status ≠ "approved" with message
- [ ] Step 2: "Review Deadlines →" jumps to Step 3 when status === "approved"
- [ ] Step 3: conference URL shown as clickable link (opens new tab)
- [ ] Step 3: deadline management UI (existing deadlines + add + cancel + supersede)
- [ ] Step 3: contributor section hidden; name/email auto-filled from localStorage
- [ ] Step 3: "Done →" navigates to role dashboard
- [ ] `EventReviewCard` writes event to `sessionStorage["manage_review_event"]` before
  navigating
- [ ] `StepIndicator` reused (3 steps, matching submission wizard visual style)
- [ ] All UI strings in i18n (en / pt / es / de)
- [ ] Mobile responsive
