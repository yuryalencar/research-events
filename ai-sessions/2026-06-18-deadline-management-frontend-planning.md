# Deadline Management — Frontend Planning

**Date:** 2026-06-18
**Spec:** `specs/frontend/deadline-management.md`

---

## Goal

Design and plan a public deadline management page at `/events/[slug]/deadlines`
where any user can add new deadlines, supersede (update) existing ones, or cancel
them — all on a single page, submitted as one batch. Entry point is a pencil icon
added to the deadline list inside the event drawer/sheet on the homepage.

---

## What was decided

### Feature overview
- **No auth required** — same contributor attribution pattern as event submission (name + email collected once, shared across all operations in the batch)
- **Single-page, no redirects** — all changes tracked locally; one "Submit changes" button sends everything
- **Three operations tracked locally before submit:**
  - **Add** — new deadline card (form with type, description, date, time?, timezone?, is_optional)
  - **Supersede** — pencil icon on existing card makes date/time/timezone editable inline; type/description/is_optional shown read-only (backend rule: supersede only changes date/time/timezone)
  - **Cancel** — X icon marks card with strikethrough + "Pending cancellation" label; Revert button undoes before submit
- **Parallel batch submission** — on Submit, calls `cancelDeadline` + `supersedeDeadline` + `addDeadlines` all in parallel
- **Success state** — same design as `SubmitSuccess` (CheckCircleIcon, event name, summary "X added · Y updated · Z cancelled", "Back to homepage" button)
- **Loading overlay** — full-page spinner while API calls run

### State passing (no GET /events/:id endpoint yet)
- Homepage writes full `EventListItem` to `sessionStorage["deadline_management_event"]` before navigating
- Page reads on mount; if absent or slug mismatch → toast + redirect to `/[locale]`

### Navigation prop pattern
`EventDetailContent` receives `onManageDeadlines: (event: EventListItem) => void` prop.
`EventDetailView` (already `"use client"`) implements it: writes sessionStorage + calls `router.push`.
This keeps `EventDetailContent` as a pure display component.

### Design references
- Bottom nav: `flex justify-between border-t border-border pt-4` — back button left, submit right (same as Step3Deadlines)
- Deadline cards: same `rounded-md border border-border p-4` grid as `DeadlineRow` in Step3Deadlines, extended with time/timezone fields
- Add deadline button: same dashed-border style as Step3Deadlines
- Success: same structure as `SubmitSuccess` component

### Key validation rules
- All validation shown only after a submit attempt (not on blur)
- `date` must not be after `event.end_date`
- `time` if provided: HH:MM format (00:00–23:59)
- Supersede with unchanged values → not counted as a change
- Nothing changed → toast, no API calls

---

## Backend status

**Fully implemented — no backend work needed.** All three endpoints exist:
- `POST /api/v1/events/{id}/deadlines` — add deadlines (batch)
- `PATCH /api/v1/events/{eventId}/deadlines/{deadlineId}/cancel`
- `POST /api/v1/events/{eventId}/deadlines/{deadlineId}/supersede`

API client functions (`addDeadlines`, `cancelDeadline`, `supersedeDeadline`) and
all TypeScript types already exist in `lib/api/events.ts` and `types/api.ts`.

---

## Approved plan (Phase 2)

### New files
| File | Purpose |
|---|---|
| `hooks/useDeadlineManage.ts` | All state: change tracking, validation, submission logic |
| `hooks/useDeadlineManage.test.ts` | ~30 Vitest tests covering state transitions, validation, submission |
| `app/[locale]/events/[slug]/deadlines/page.tsx` | Thin server wrapper |
| `components/events/deadlines/DeadlineManagePage.tsx` | Main client orchestrator (sessionStorage read, redirect, page states) |
| `components/events/deadlines/DeadlineCard.tsx` | Existing deadline card (Default / Supersede-editing / Pending-cancel states) |
| `components/events/deadlines/AddDeadlineCard.tsx` | New deadline form card |
| `components/events/deadlines/DeadlineManageSuccess.tsx` | Success state |

### Modified files
| File | Change |
|---|---|
| `components/events/EventDetailContent.tsx` | Add pencil icon to each deadline item + `onManageDeadlines` prop |
| `components/events/EventDetailView.tsx` | Implement `onManageDeadlines` (sessionStorage write + router.push) |
| `messages/en.json`, `pt.json`, `es.json`, `de.json` | New `deadlines.manage.*` i18n keys |

### Cycle breakdown
1. `useDeadlineManage` — state transitions + `hasChanges`
2. `useDeadlineManage` — validation
3. `useDeadlineManage` — submission
4. `EventDetailContent` + `EventDetailView` — pencil icon, prop, sessionStorage, navigate
5. `DeadlineManageSuccess` — success component
6. `AddDeadlineCard` — new deadline form card
7. `DeadlineCard` — existing deadline card (3 states)
8. `DeadlineManagePage` — orchestrator
9. `page.tsx` — server wrapper + route
10. i18n — all 4 locales

---

## State at end of session

**Phases 0–2 complete. Phase 3 (Red) not yet started.**

- Spec written and approved: `specs/frontend/deadline-management.md`
- Plan written and approved: see cycle breakdown above
- No code written yet

**Resume point:** Start Phase 3 — write failing tests for Cycle 1
(`useDeadlineManage` state transitions + `hasChanges`), then show output
and ask for Green approval.
