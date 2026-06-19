# Management Dashboard — Frontend Implementation

**Date:** 2026-06-19
**Phases completed:** 1 (Discovery) → 2 (Plan) → 3 (Red) → 4 (Green) → 5 (Refactor) → 6 (Docs)
**Continued from:** `ai-sessions/2026-06-19-manage-dashboard-planning.md`

---

## Goal

Implement the admin/moderator event review queue at `/manage/admin` and
`/manage/moderator`. Both roles share the same components and layout; the own-event
restriction and review route URL differ between roles.

---

## Files Created

| File | Description |
|------|-------------|
| `src/hooks/useReviewEvents.ts` | Filter + fetch + pagination state hook |
| `src/hooks/useReviewEvents.test.ts` | Unit tests (10 tests) |
| `src/components/manage/ManageHeader.tsx` | Sticky header — Welcome + role (left), avatar with float menu (right) |
| `src/components/manage/EventReviewCard.tsx` | Event card — normal variant (name link + Review button) and greyed-out variant (moderator own-event) |
| `src/components/manage/ManageDashboard.tsx` | Shared dashboard — filter row (status/tier/year + Apply), event list, pagination |

## Files Modified

| File | Change |
|------|--------|
| `src/app/[locale]/manage/admin/page.tsx` | Replaced welcome stub with `ManageDashboard` |
| `src/app/[locale]/manage/moderator/page.tsx` | Replaced welcome stub with `ManageDashboard` |
| `src/app/[locale]/manage/page.tsx` | Added `id: parseInt(claims.sub, 10)` to localStorage on login |
| `src/messages/{en,pt,es,de}.json` | Added `manage.reviewDashboard.*` i18n keys |

---

## Key Decisions

**Endpoint:** `GET /api/v1/events` (public) with `status`, `tier`, `year`, `page`,
`page_size=30`, `pagination=on`. No separate admin endpoint needed.

**Own-event restriction:** `event.created_by.id === sessionUser.id` — requires storing
the user's numeric ID in `localStorage.manage_user`. Sourced from JWT `sub` claim
(`parseInt(claims.sub, 10)`) on login.

**Custom select arrow:** Native `<select>` arrow ignored `pr-*` padding. Fixed with
`appearance-none` + absolutely positioned `ChevronDown` icon from lucide-react.

**Layout centering:** Both header and main content constrained to `max-w-4xl mx-auto`
with identical `px-4 sm:px-6` — ensures "Welcome" aligns with the filter labels below.

**Avatar cursor:** Added `cursor-pointer` explicitly (buttons inside non-interactive
containers don't always inherit pointer cursor in all browsers).

---

## State at End of Session

Feature fully implemented. 225 tests pass, zero type errors. The review page route
(`/manage/[role]/events/[slug]/review`) is a placeholder for the next feature.
