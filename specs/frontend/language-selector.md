# Language Selector

## Description
A fixed floating button visible on every page that shows the current locale's flag and opens
a dropdown to switch between the 4 supported languages.

## Behaviour
- Renders in the root layout so it appears on every page.
- Displays the active locale's flag emoji as a circular button (always visible).
- Clicking opens a small dropdown listing all 4 locales — each with flag + language name.
- The active locale appears in the list with a visual highlight (checkmark or background).
- Clicking a locale navigates to the same path with the new locale prefix (path preserved).
- Clicking outside closes the dropdown without switching.

## Locales and flags
| Locale | Flag | Label     |
|--------|------|-----------|
| en     | 🇺🇸   | English   |
| pt     | 🇧🇷   | Português |
| es     | 🇪🇸   | Español   |
| de     | 🇩🇪   | Deutsch   |

## Placement
Fixed bottom-right (`bottom-4 right-4`, `z-40`) — the only free corner on the globe page.

## Permissions
Public — no authentication required, visible to all visitors.

## Rules
- No new i18n translation keys — language names are universal proper nouns.
- Path is preserved on locale switch using next-intl's `useRouter` + `usePathname`.
- No geocoding or external API calls — purely client-side navigation.

## Error cases
None meaningful — next-intl middleware handles unknown locales via redirect.

## Border / corner cases
- Deep route (`/en/events/my-conf`) → switches to `/pt/events/my-conf` (path preserved).
- Already on the active locale → clicking it in the dropdown is a no-op (navigates to same page).
- SSR: button must render without hydration mismatch — locale is available server-side via next-intl.

## Definition of done
- [ ] Button visible on every page (globe, event detail, submit, admin, manage)
- [ ] Clicking the button opens the dropdown with all 4 locales
- [ ] Active locale is highlighted in the dropdown
- [ ] Clicking a non-active locale navigates to the same path with new locale prefix
- [ ] Page content re-renders in the selected language after switch
- [ ] Clicking outside closes the dropdown
- [ ] No hydration mismatch warnings in the browser console
