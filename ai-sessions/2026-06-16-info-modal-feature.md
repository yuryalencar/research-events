# Info Modal — Floating Info Button

**Date:** 2026-06-16
**Spec:** `specs/frontend/info-modal.md`

---

## Goal

Add a floating `ⓘ` button fixed to the bottom-left of the globe homepage that opens a modal with static project information: app version, pin color legend, event submission flow, moderator contact, author credit, and open-source contribution invite.

---

## Decisions made

### Bottom-left placement
The event-detail Sheet slides in from the right, so bottom-right was ruled out to avoid overlap. Future filter controls will likely occupy the top area. Bottom-left is the cleanest corner that conflicts with nothing existing or planned.

### Button hidden on mobile when Drawer is open — not just invisible
When `selectedEvent !== null` on mobile, the button returns `null` (not just `visibility: hidden`). This prevents it from being keyboard-focusable or read by screen readers while the Drawer has focus. The `drawerOpen` prop flows from `selectedEvent !== null` in `page.tsx`; `InfoButton` reads `useMediaQuery` internally to decide whether to apply the hiding logic.

### `isDesktop` owned by `InfoButton`, passed to `InfoModal`
`InfoButton` already calls `useMediaQuery(DESKTOP_QUERY)` to decide the hide behavior, so it passes `isDesktop` down to `InfoModal` rather than having `InfoModal` call the hook a second time.

### New `ui/dialog.tsx` primitive
The project had `ui/sheet.tsx` (slides from edge) and `ui/drawer.tsx` (bottom sheet) but no centered modal. Created `ui/dialog.tsx` following the same thin-wrapper pattern over `@radix-ui/react-dialog` (already installed). This keeps the modal vocabulary consistent across the codebase.

### `InfoModalContent` extracted as a shared sub-component
Both the Dialog (desktop) and Drawer (mobile) render identical content. Extracting `InfoModalContent` avoids duplicating six sections and keeps `InfoModal` focused on layout selection only.

### Single `overflow-y-auto` on the inner content div
`InfoModalContent` carries `flex-1 min-h-0 overflow-y-auto`. The outer `DialogContent` only constrains height via `max-h-[85vh]` — no `overflow-y-auto` there. Having both would create a redundant scroll context where the Dialog shell scrolls before the inner flex child gets a chance.

### No Vitest tests for this feature
The Vitest scope per CLAUDE.md covers hooks, `lib/api.ts`, pure utilities, and type guards. This feature introduced none of those — only UI components with static content. Phase 3 (Red) was skipped by agreement, verified manually in the browser.

---

## Files changed

| File | Change |
|---|---|
| `src/lib/constants.ts` | Added `APP_VERSION`, `GITHUB_REPO_URL`, `GITHUB_PROFILE_URL`, `ADMIN_EMAIL` |
| `src/messages/en.json` | Added `info` namespace (20 keys) |
| `src/messages/pt.json` | Added `info` namespace (20 keys) |
| `src/messages/es.json` | Added `info` namespace (20 keys) |
| `src/messages/de.json` | Added `info` namespace (20 keys) |
| `src/components/ui/dialog.tsx` | New centered Dialog primitive (wraps `@radix-ui/react-dialog`) |
| `src/components/globe/InfoModal.tsx` | New modal — Dialog (desktop) / Drawer (mobile), 6 content sections |
| `src/components/globe/InfoButton.tsx` | New floating button — owns open state, hides itself on mobile when event Drawer is open |
| `src/app/[locale]/page.tsx` | Added `<InfoButton drawerOpen={selectedEvent !== null} />` |
| `specs/frontend/info-modal.md` | New spec file |

---

## State at end of session

Feature complete. TypeScript clean (`pnpm typecheck` passes, zero errors). No new tests (Vitest scope was empty for this feature).

**To restore context:** The globe homepage and deep-linking are fully implemented (see `2026-06-16-globe-event-deeplink-feature.md`). This session added the info button on top of that. The button lives at `bottom-left z-40`, below the Sheet/Drawer overlay (`z-50`) so it's correctly hidden when any modal is open. The next logical features would be search/filter controls on the globe, or the public event submission form (`/events/submit`).
