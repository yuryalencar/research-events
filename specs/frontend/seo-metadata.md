# SEO Metadata

## Description

Adds SEO metadata to the Next.js frontend using the App Router `Metadata` API:
base site metadata on the root layout (title template, description, Open Graph,
Twitter card, hreflang alternates), page-level overrides on public pages, a
`noindex` guard on all private routes, a `robots.txt`, and a `sitemap.xml`.
All user-facing strings go through `next-intl` in all 4 supported locales.

## Behaviour

### Root layout (`app/[locale]/layout.tsx`)

Sets the base metadata inherited by every page under the locale segment:

| Field | Value |
|---|---|
| `metadataBase` | `NEXT_PUBLIC_APP_URL` env var (fallback: `https://research-events.vercel.app`) |
| `title.default` | `"ReSEARCH Events"` (used when no page overrides title) |
| `title.template` | `"%s \| ReSEARCH Events"` (page titles slot into `%s`) |
| `description` | Translated site description from `app.description` i18n key |
| `keywords` | `["research conferences", "software engineering", "computer science", "academic events", "submission deadlines", "call for papers"]` |
| `authors` | `[{ name: "Yury Lima" }]` |
| `openGraph.type` | `"website"` |
| `openGraph.image` | `/logo-with-opensource.png` |
| `twitter.card` | `"summary_large_image"` |
| `alternates.languages` | One entry per locale: `/en`, `/pt`, `/es`, `/de` |

### Per-page metadata

| Route | Title | Description | Robots |
|---|---|---|---|
| `[locale]/` (globe) | *(inherits layout default)* | *(inherits layout)* | index, follow |
| `[locale]/events/submit` | `meta.submitTitle` (translated) | `meta.submitDescription` (translated) | **noindex, nofollow** |
| `[locale]/manage/**` | *(inherits layout)* | *(inherits layout)* | **noindex, nofollow** |

The `manage` noindex is set once in `app/[locale]/manage/layout.tsx` and
cascades to all routes under it — no per-page repetition.

### `robots.txt` (`app/robots.ts`)

```
User-agent: *
Allow: /
Disallow: /en/manage
Disallow: /pt/manage
Disallow: /es/manage
Disallow: /de/manage
Sitemap: https://research-events.vercel.app/sitemap.xml
```

### `sitemap.xml` (`app/sitemap.ts`)

One entry per locale for the public globe homepage:
```
https://research-events.vercel.app/en   (priority 1, daily)
https://research-events.vercel.app/pt   (priority 1, daily)
https://research-events.vercel.app/es   (priority 1, daily)
https://research-events.vercel.app/de   (priority 1, daily)
```

Dynamic event pages are not included yet — they will be added when the event
detail page is implemented.

## Rules

- All user-facing strings (description, page titles) go through `next-intl` — no
  hardcoded English in metadata
- All 4 locales (`en`, `pt`, `es`, `de`) must have every `app.*` and `meta.*`
  key — no missing keys
- `metadataBase` must always be set on the root layout so relative image URLs in
  Open Graph resolve to absolute URLs
- Admin and manage routes must always be `noindex` — they are auth-protected and
  should never appear in search results
- `NEXT_PUBLIC_APP_URL` must be set on Vercel for correct canonical URLs and OG
  image paths; the fallback is the production Vercel URL

## i18n keys added

```
app.description   — site-wide description (all 4 locales)
meta.submitTitle  — title for the submit event page (all 4 locales)
meta.submitDescription — description for the submit event page (all 4 locales)
```

## Files changed

| File | Change |
|---|---|
| `src/app/[locale]/layout.tsx` | Full base metadata via `generateMetadata` |
| `src/app/[locale]/events/submit/page.tsx` | `generateMetadata` with page title + noindex |
| `src/app/[locale]/manage/layout.tsx` | **New** — noindex for all manage/* routes |
| `src/app/robots.ts` | **New** — `robots.txt` generation |
| `src/app/sitemap.ts` | **New** — `sitemap.xml` generation |
| `src/messages/en.json` | `app.description`, `meta.*` keys |
| `src/messages/pt.json` | Same, in Portuguese |
| `src/messages/es.json` | Same, in Spanish |
| `src/messages/de.json` | Same, in German |

## Required env var on Vercel

```
NEXT_PUBLIC_APP_URL = https://research-events.vercel.app
```

Without this the fallback hardcoded URL is used, which is fine for production
but the var should be set explicitly so staging/preview deployments can override it.

## Definition of done

- [ ] `<title>` on the globe page is `"ReSEARCH Events"`
- [ ] `<title>` on the submit page is `"Submit an Event | ReSEARCH Events"`
- [ ] `<meta name="description">` is present on the globe page
- [ ] `<meta property="og:image">` points to the logo (absolute URL)
- [ ] `<link rel="alternate" hreflang="...">` tags present for all 4 locales
- [ ] `GET /robots.txt` returns correct rules with `/*/manage` disallowed
- [ ] `GET /sitemap.xml` returns 4 entries (one per locale)
- [ ] Manage pages have `<meta name="robots" content="noindex, nofollow">`
- [ ] Submit page has `<meta name="robots" content="noindex, nofollow">`
- [ ] All 4 locale files have matching `app.description` and `meta.*` keys
- [ ] `pnpm typecheck` and `pnpm test` pass
