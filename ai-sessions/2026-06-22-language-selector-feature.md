# Language Selector Feature

**Date:** 2026-06-22
**Status:** Complete

---

## Goal

Add a fixed floating language selector to every page of the frontend, showing the current locale's flag and allowing the user to switch between the 4 supported languages (en, pt, es, de) while preserving the current URL path.

---

## Decisions made

- **Placement:** Fixed bottom-right (`bottom-4 right-4`, `z-40`) — the only free corner (top-left = FilterPanel, top-right = AddEvent/ViewToggle, bottom-left = InfoButton).
- **Flags:** 🇺🇸 English, 🇧🇷 Português, 🇪🇸 Español, 🇩🇪 Deutsch. Brazilian flag for Portuguese (as specified by the user).
- **Current locale in dropdown:** Shown with a blue highlight and a checkmark icon (not omitted).
- **Path preservation:** `router.replace(pathname, { locale })` via next-intl's locale-aware router — `/en/events/my-conf` → `/pt/events/my-conf`.
- **No new i18n keys:** Language names are universal proper nouns, not translated strings.
- **Navigation module:** Created `src/i18n/navigation.ts` exporting `useRouter`, `usePathname`, `Link` via `createNavigation(routing)` — this was missing and is the next-intl v4 canonical way to get locale-aware navigation hooks.
- **Layout placement:** `<LanguageSelector />` added inside `NextIntlClientProvider` in `app/[locale]/layout.tsx` so it renders on every page.

---

## Files created / modified

| File | Action |
|---|---|
| `specs/frontend/language-selector.md` | Created — feature spec |
| `src/i18n/navigation.ts` | Created — locale-aware navigation exports via `createNavigation(routing)` |
| `src/components/ui/LanguageSelector.tsx` | Created — flag button + dropdown client component |
| `src/app/[locale]/layout.tsx` | Modified — added `<LanguageSelector />` inside `NextIntlClientProvider` |

---

## State at end

Feature is complete and typechecks clean (`pnpm typecheck` passes). No Vitest tests were written — the component has no extracted hook or utility logic (pure UI), and CLAUDE.md explicitly excludes component rendering tests from the Vitest scope.

---

## Context to restore

- `src/i18n/navigation.ts` now exists and exports locale-aware hooks. Any future component that needs to **switch locales** or navigate with locale awareness should import from `@/i18n/navigation` (not `next/navigation`).
- Components that only need intra-locale navigation (push, replace without locale change) can continue using `next/navigation` as before.
