# Management Event Review — Frontend

**Date:** 2026-06-19
**Phases completed:** 0 (Spec) → 1 (Discovery) → 2 (Plan) → 3 (Red) → 4 (Green) → 5 (Refactor) → 6 (Docs)

---

## Goal

Implement the 3-step review wizard for admin and moderator roles at
`/manage/admin/events/[slug]/review` and `/manage/moderator/events/[slug]/review`.
Mirrors the public submission wizard in structure. Step 1 edits event details, Step 2
makes the approve/reject decision, Step 3 manages deadlines.

---

## Files Created

| File | Description |
|------|-------------|
| `src/hooks/useReviewWizard.ts` | Core wizard state: `buildEventPatch`, `validateFields`, `initFormData` (pure FP functions) + `useReviewWizard` hook managing form data, step, errors, success state |
| `src/hooks/useReviewWizard.test.ts` | 25 unit tests covering `buildEventPatch`, `validateStep1`, `goToStep`, `updateField`, `approve`, `reject`, `clearSuccess` |
| `src/components/manage/review/ReviewWizard.tsx` | Top-level controller: `ManageHeader` + `StepIndicator` (with label prop) + conditional step/success rendering |
| `src/components/manage/review/ReviewStep1Details.tsx` | Editable event form pre-filled from sessionStorage event; `LocationPicker` pre-pinned at current lat/lng; non-editable status badge above form |
| `src/components/manage/review/ReviewStep2Decision.tsx` | Approve/Reject decision buttons (Reject left, Approve right); `ApproveModal` + `RejectModal`; "Review Deadlines →" disabled when status ≠ approved; success state transitions |
| `src/components/manage/review/ReviewStep3Deadlines.tsx` | Reuses `useDeadlineManage` hook; conference URL link; contributor auto-filled from localStorage and hidden from UI |
| `src/components/manage/review/ApproveModal.tsx` | Dialog with optional note textarea; green confirm button |
| `src/components/manage/review/RejectModal.tsx` | Dialog with required reason textarea; confirm disabled until reason.trim().length > 0; red confirm button |
| `src/components/manage/review/ReviewSuccess.tsx` | Approve variant: green CheckCircleIcon + "Manage Deadlines" + "Back to Dashboard". Reject variant: red XCircleIcon + reason shown + "Back to Dashboard" only |
| `src/app/[locale]/manage/admin/events/[slug]/review/page.tsx` | Admin review route — auth guard + sessionStorage event guard + renders `ReviewWizard` |
| `src/app/[locale]/manage/moderator/events/[slug]/review/page.tsx` | Moderator review route — same pattern |
| `specs/frontend/manage-review.md` | Feature spec |

## Files Modified

| File | Change |
|------|--------|
| `src/components/events/submit/StepIndicator.tsx` | Added optional `label?: string` prop to override the default translated indicator text |
| `src/components/manage/EventReviewCard.tsx` | Converted Link to button; writes event to `sessionStorage["manage_review_event"]` before navigating |
| `src/components/map/LocationPicker.tsx` | Fixed async Leaflet init race: added `latLngRef` so init callback reads correct lat/lng after the `import("leaflet")` promise resolves. Also switched tile provider to CartoDB Voyager for English place labels |
| `src/messages/en.json` + `pt.json` + `es.json` + `de.json` | Added `manage.review.*` i18n namespace |

---

## Key Decisions

- **`buildEventPatch` pure function** — only changed fields are sent in the PATCH body;
  unchanged fields are omitted. Keeps the diff minimal and avoids overwriting unrelated
  concurrent edits.
- **`successState` in hook, not local component state** — storing `ReviewSuccessState`
  in `useReviewWizard` lets `ReviewWizard` swap out the whole step area for
  `ReviewSuccess` in one place, without step components needing to know about it.
- **`step3Event = wizard.successState?.event ?? event`** — Step 3 uses the updated event
  from the PATCH response (if the user came from the success screen) rather than the
  original sessionStorage event, so deadline edits reflect the post-approval state.
- **Contributor pre-fill via `useEffect` on mount** — `useDeadlineManage` initializes
  contributor as empty strings; a single `useEffect` in `ReviewStep3Deadlines` calls
  `updateContributor` on mount to inject the reviewer's name/email from localStorage,
  keeping the public hook unmodified.
- **CartoDB Voyager tiles** — switched from the default OSM tile server (which renders
  place labels in local script, e.g. 東京 for Tokyo) to CartoDB Voyager, which renders
  all country and city labels in English globally.

---

## Bugs Fixed

- **Map pin not pre-placed on review form**: `LocationPicker`'s lat/lng sync effect fires
  before the async `import("leaflet")` resolves, so `mapRef.current` is null and the
  early-return guard skips placing the initial marker. Fixed with `latLngRef` — a ref
  that mirrors the lat/lng props and is read inside the async callback after the map is
  created.

---

## State at End of Session

Feature fully implemented and typechecks clean (`pnpm typecheck` passes). Tests pass
(`useReviewWizard.test.ts` — 25 tests). Tile provider updated to CartoDB Voyager for
English map labels.

Known open item: pin-not-restored-on-back bug in the **submission form** (Step 2 → Step 3
→ Back to Step 2) was reported by the user but deferred to the next session.
