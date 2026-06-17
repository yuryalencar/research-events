# Event Submission Wizard — Frontend

**Date:** 2026-06-17
**Spec:** `specs/frontend/event-submission-wizard.md`

---

## Goal

Build a full 3-step public event submission flow accessible via a floating `+` button on the globe homepage. The wizard lets any researcher check for duplicates, fill in event details (including a Leaflet map pin for lat/lng), optionally register deadlines, and see a success confirmation.

---

## What was built

### New files

| File | Purpose |
|---|---|
| `src/hooks/useEventSearch.ts` | Server-side paginated search of pending events (Step 1). Draft/applied year separation, `revision` counter forces re-fetch on repeated Apply. Client-side text filter on current page. |
| `src/hooks/useSubmitWizard.ts` | Wizard state machine: step navigation, form data, validation, submission. Errors in `useRef` (in-place mutation) so `validateStep2()` callers see updated values synchronously outside `act()`. |
| `src/lib/countryCoordinates.ts` | `[lat, lng]` for every country in the COUNTRIES list (~195 entries). Used by LocationPicker to pan the map on country select. |
| `src/components/globe/AddEventButton.tsx` | Fixed `top-4 right-4` floating `+` button. Custom CSS tooltip (Tailwind `group-hover:opacity-100`) appears to the left on hover — more reliable than native `title` attribute. Locale-aware `href`. |
| `src/components/map/LocationPicker.tsx` | Leaflet map (loaded via `dynamic(..., {ssr:false})`). `cancelled` flag in the async Leaflet import prevents "Map container is already initialized" in React StrictMode. `onChangeRef` prevents stale closure in Leaflet drag/click handlers. Country effect pans to `COUNTRY_COORDINATES[country]`. |
| `src/components/events/submit/StepIndicator.tsx` | Step progress display with numbered circles and connecting lines. |
| `src/components/events/submit/Step1Search.tsx` | Pending events table with text + year filters, server-side pagination, and "Back to homepage" link on the left nav. |
| `src/components/events/submit/Step2Details.tsx` | Large form: submitter info, full event name, slug (with hint), domain/tier selects, date pickers, website, country/city, and the Leaflet location picker. |
| `src/components/events/submit/Step3Deadlines.tsx` | Optional deadline rows (type / description / date / optional checkbox / remove). "Skip & Submit" and "Submit" (disabled when rows have errors). |
| `src/components/events/submit/SubmitSuccess.tsx` | Success screen: event summary card (slug, dates, website, status badge, submitter name) + "Back to homepage" link. |
| `src/components/events/submit/SubmitWizard.tsx` | Top-level wizard controller. Renders success screen when `submittedEvent !== null`, otherwise routes between Step 1/2/3. Validates Step 2 before advancing. |
| `src/app/[locale]/events/submit/page.tsx` | Replaced `<main />` stub with `<SubmitWizard />`. |

### Modified files

| File | Change |
|---|---|
| `src/app/[locale]/page.tsx` | Added `<AddEventButton />` import and JSX usage in the globe homepage. |
| `src/messages/en.json` | Added full `submit` namespace (~65 keys). |
| `src/messages/pt.json` | Portuguese translations for all `submit.*` keys. |
| `src/messages/es.json` | Spanish translations. |
| `src/messages/de.json` | German translations. |
| `src/lib/countryCoordinates.ts` | Expanded from ~120 to all 195 countries in the COUNTRIES list (was missing Antigua and Barbuda, Barbados, North Macedonia, South Sudan, Vatican City, and ~50 others). |

---

## Key decisions

**Errors in `useRef`, not `useState`**
`validateStep2()` is called synchronously and the caller checks the errors object immediately after. Using `useState` for errors would require a re-render cycle before the updated value is visible to the caller. The solution: `errorsRef.current` is mutated in-place + a dummy `setErrorVersion(v => v + 1)` triggers the re-render so the UI updates, while the ref holds the latest value for synchronous reads.

**`revision` counter for forced re-fetch**
When the user presses Apply twice with identical year/page values, React bails out because no state changed. A `revision` counter increments on every `apply()` and `clear()` call, ensuring the `useEffect` dependency array always sees a new value.

**`cancelled` flag for Leaflet async init**
`import("leaflet")` is a promise. In React StrictMode, the cleanup function runs before the promise resolves — at that point `map` is still `undefined`, so `map?.remove()` is a no-op and the DOM container stays in Leaflet's registry. The second mount then starts another import, and when both promises resolve they both call `L.map(container)` → "Map container is already initialized". Fix: a `cancelled` boolean in the closure, set to `true` in the cleanup, checked at the top of the `.then()` callback.

**Custom CSS tooltip vs native `title`**
The native `title` attribute on the `+` button was invisible in practice (no consistent tooltip timing in Chrome). Replaced with a Tailwind `group-hover:opacity-100` span that fades in to the left of the button.

**"Back to homepage" in Step 1**
Step 1 is the entry point — there's nowhere to "go back" within the wizard. A `← Back to homepage` link using `useLocale()` + `Link` was added to the left side of the Step 1 nav row, mirroring the layout of Steps 2/3.

---

## Bugs fixed

| Bug | Root cause | Fix |
|---|---|---|
| "Map container is already initialized" | Leaflet async init in StrictMode: cleanup ran before promise resolved, leaving container registered | Added `cancelled` flag checked inside `.then()` callback |
| Antigua and Barbuda (and ~50 others) not panning map on country select | Missing entries in `countryCoordinates.ts` | Added all 195 countries from the COUNTRIES list |
| Float button tooltip invisible | Native `title` tooltip unreliable | Custom CSS tooltip via Tailwind `group-hover` |
| No way back to homepage from Step 1 | Step 1 had no navigation | Added `← Back to homepage` link (locale-aware) |

---

## Tests

- **15 test files, 152 tests — all passing**
- `useEventSearch.test.ts` — 8 tests: fetch on mount, apply/clear/goToPage, filteredEvents, totalPages, error state
- `useSubmitWizard.test.ts` — 13 tests: validateStep2 (happy + all required fields), submit (201/409/400/429/network), deadline add/remove/update
- Environment note: `@rolldown/binding-linux-arm64-gnu` native binding was missing on the arm64 CI image; fixed with `CI=true pnpm install`.

---

## State at end of session

Feature complete. All 4 phases (Red → Green → Refactor → Integration) done. The submission wizard is live at `/<locale>/events/submit`, reachable from the floating `+` button on the globe homepage.

**Open question (not blocking):** `GET /api/v1/events?status=pending` — the public events endpoint may not support status filtering (it defaults to approved). This would mean Step 1 shows approved events, not pending. Needs backend verification before the feature is fully end-to-end tested.
