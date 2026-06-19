# Management Dashboard — Frontend Planning

**Date:** 2026-06-19
**Phases completed:** 0 (Spec) — approved, ready for Phase 1
**Next session:** start at Phase 1 (Discovery)

---

## Goal

Build the admin/moderator event review queue at `/manage/admin` and `/manage/moderator`.
Both roles share the same layout and components. Replaces the welcome-only stubs left by
the manage-portal feature.

---

## Spec

Full spec: `specs/frontend/manage-dashboard.md`

---

## Key Decisions Made

**Endpoint:** Public `GET /api/v1/events` with `status`, `tier`, `year`, `page`,
`page_size=30`, `pagination=on` params. No separate admin endpoint needed.

**Event name:** Opens `event.website_url` in a new tab (external conference site).

**Review button:** Navigates to `/${locale}/manage/${role}/events/${slug}/review`
(role-aware URL; the review page itself is out of scope for this feature).

**Own-event restriction (moderator only):** When `event.created_by.id === sessionUser.id`,
the entire card is greyed out and the review button is replaced by
"You can't review this submission". Admin role never sees this variant.

**Pagination:** 30 per page, prev/next buttons only, label shows
"Page X of Y (Z events)".

**Avatar menu:** First letter of name in a circle, click opens float menu,
only "Sign out" for now.

**localStorage update:** `manage_user` must also store `id: number` (from JWT `sub`
claim) so the own-event check can compare `event.created_by.id === sessionUser.id`.
This requires a small change to the login page (`/manage/page.tsx`).

---

## Files to Create / Modify (next session)

| File | Action |
|---|---|
| `src/lib/api/events.ts` | Create — `listEvents(params)` |
| `src/lib/api/events.test.ts` | Create — unit tests |
| `src/hooks/useReviewEvents.ts` | Create — filter + fetch + pagination |
| `src/hooks/useReviewEvents.test.ts` | Create — unit tests |
| `src/components/manage/ManageHeader.tsx` | Create — sticky header + avatar menu |
| `src/components/manage/EventReviewCard.tsx` | Create — normal + grey variants |
| `src/components/manage/ManageDashboard.tsx` | Create — filter row + list + pagination |
| `src/app/[locale]/manage/admin/page.tsx` | Modify — wire ManageDashboard |
| `src/app/[locale]/manage/moderator/page.tsx` | Modify — wire ManageDashboard |
| `src/app/[locale]/manage/page.tsx` | Modify — store `id` in localStorage |
| `src/messages/{en,pt,es,de}.json` | Modify — add `manage.reviewDashboard.*` keys |

---

## State at End of Session

Spec written and approved. Committed to `specs/frontend/manage-dashboard.md`.
No implementation started. Resume from Phase 1 (Discovery).
