# Admin User Management — Frontend Planning (in progress)

**Date:** 2026-06-26
**Status:** Phase 0 not started — spec interview abandoned mid-session, resume from here.

---

## Goal

Build the frontend UI for admin user management inside the management area. Backed by two endpoints delivered this session:
- `POST /api/v1/admin/users` — register a new admin or moderator
- `PATCH /api/v1/admin/users/{id}/role` — change a user's role (invalidates their session)

---

## Context

Backend commit: `c3fd82c` — both endpoints are live, tested (458 backend tests), and follow the existing auth/rate-limit/audit pattern.

The existing management area lives at:
- `/manage` — login page
- `/manage/admin` — admin dashboard (event review queue)
- `/manage/moderator` — moderator dashboard

---

## Where We Left Off

Phase 0 (spec interview) was started but the user stopped before answering. The following six questions need to be asked in one message when resuming:

1. **Where does this live?** New page `/manage/admin/users`, a tab inside the existing `/manage/admin` dashboard, or a separate nav entry?
2. **User list scope** — show all roles (admin + moderator + contributor) or only admin + moderator?
3. **Register form UX** — inline form, modal/dialog, or separate page (`/manage/admin/users/new`)?
4. **Role change UX** — dropdown on each row, separate promote/downgrade buttons, confirmation modal before submit?
5. **Backend list endpoint** — `GET /api/v1/admin/users` does not exist yet. Need to spec and build it first, or skip the list and build a form-only UI for now?
6. **Pagination** — paginated list or flat (admin count is small)?

---

## Resume Instructions

Start at **Phase 0** — ask all six questions above in a single message, draft the spec, run the approval checklist, then proceed through the standard workflow (Phase 1 → 2 → 3 → 4 → 5 → 6).

If a backend list endpoint is needed, spec and build it first (backend Phase 0–6), then return to the frontend feature.
