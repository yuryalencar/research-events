# Session: Globe Homepage (Frontend)

**Date:** 2026-06-15
**Status:** Complete. Feature workflow (Phase 0 → 6).

---

## Context

First end-to-end frontend feature built on top of the API client +
error-handling infrastructure: the 3D globe homepage that plots approved
events as pins, with a slide-in/bottom-drawer detail view.

Spec: `specs/frontend/globe-homepage.md`.

---

## What was built

### Globe page (`src/app/[locale]/page.tsx`)
- `GlobeView` (Globe.gl) loaded via `dynamic(..., { ssr: false })` per
  CLAUDE.md's Globe.gl/Leaflet rule.
- `.space-bg` starfield background (`globals.css`, `@layer components`) —
  tiled `radial-gradient` dot pattern + two nebula glows, no image assets.
- Loading overlay: centered, rounded card with a spinner and
  `home.loading` text, shown while `useEvents()` is loading.
- Empty-state overlay: centered, red-bordered card with `home.noEvents`
  text, shown when the fetch completes with zero events. Verified via SSR
  HTML (`curl localhost:3000/en`) that both overlays render inside `<main>`
  as absolutely-positioned siblings of the globe `<div>`, so they paint on
  top of it regardless of DOM order (CSS stacking rules for positioned vs.
  static elements).

### Pins (`src/lib/events.ts`, `src/components/globe/GlobeView.tsx`)
- `PIN_COLOR_DEFAULT` (`#facc15`, yellow), `PIN_COLOR_SELECTED` (`#ec4899`,
  pink) — chosen to stand out against the blue/green/brown globe texture.
- `PIN_COLOR_PAST` (`#ef4444`, red) + new pure helper `isEventPast(endDate,
  now)` — events whose `end_date` has already passed render as red pins
  (selection still takes priority over "past" coloring).
- `getPinColor(eventId, endDate, selectedEventId, now)` — updated signature,
  pure function, fully covered by new tests.
- `PIN_RADIUS`/`SELECTED_PIN_RADIUS` increased (0.4→0.8, 0.6→1.2) for
  visibility.

### Event detail view (`EventDetailView.tsx`, `EventDetailContent.tsx`)
- Desktop: `Sheet` sliding from the right. Mobile (`< md`): `Drawer` sliding
  from the bottom (`useMediaQuery`).
- Title now shows `{event.name} (slug)` in both Sheet and Drawer.
- Fixed Radix/vaul accessibility warning ("Missing `Description` or
  `aria-describedby`") via sr-only `SheetDescription`/`DrawerDescription`
  using new `eventDetail.detailsDescription` key.
- Fixed `Drawer`'s close button — previously inherited `flex-col` from
  `DrawerHeader` and rendered in the middle of the sheet; now an
  absolutely-positioned `DrawerPrimitive.Close` matching `Sheet`'s pattern.
- "CORE Tier" badge moved into the `<dl>` as a stacked `<dt>`/`<dd>` pair
  (label above value), consistent with dates/location/website/domain.
- `domain` (extensible string enum) translated per-locale via
  `eventDetail.domains.<value>`, falling back to the raw value if untranslated
  (`t.has()`).
- `formatDateRange(start, end, locale)` — new pure helper replacing raw
  `YYYY-MM-DD - YYYY-MM-DD` display with locale-aware ranges (e.g.
  "Apr 13–19, 2026", "Dec 28, 2025 – Jan 3, 2026").
- Content wrapper now `flex min-h-0 flex-1 flex-col ... overflow-y-auto` so a
  long deadline list scrolls inside the Sheet/Drawer instead of overflowing
  the viewport.

### Infra
- Fixed a Next.js 16 dev-mode hydration mismatch on `<MetadataWrapper>` by
  adding `generateMetadata` to `src/app/[locale]/layout.tsx` (avoids the
  streaming metadata Suspense boundary).
- Renamed `middleware.ts` → `src/proxy.ts` (Next.js 16 "middleware to proxy"
  convention) and moved it under `src/` (required when using `src/app`),
  fixing the `next-intl` routing import path.
- Seeded 8 approved events around the world (Rio, San Diego, Lisbon, Tokyo,
  Cape Town, Sydney, Berlin, Toronto) via
  `specs/backend/seed-globe-events.curl.sh` — a standalone curl script using
  the public submit + admin review endpoints.

---

## i18n

Added to all 4 locales (`en`/`pt`/`es`/`de`) under `home` and `eventDetail`:
`loading`, `noEvents`, `detailsDescription`, `tier`, `domains.computer_science`
(removed unused `close` key). `globeHomepage.test.ts` keeps key parity across
locales.

---

## Tests

`pnpm test run` → 93 passed. `pnpm typecheck`, `pnpm lint`, `pnpm build` all
clean.

---

## State at end

Feature complete and verified end-to-end (build + SSR HTML inspection of the
loading/empty states). No open follow-ups.
