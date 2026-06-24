# Globe Event Clustering

## Description

Groups geographically close events into cluster pins on the 3D globe using `supercluster`.
Clicking a cluster opens a compact multi-event drawer immediately; selecting an event from
the drawer closes it and opens the standard `EventDetailView`. Individual isolated events
(not part of any cluster) behave exactly as today. Scope: Globe view only.

---

## Behaviour

- Events within a configurable proximity radius are grouped into a single cluster pin by `supercluster`
- Minimum cluster size is **2 events** — a lone event is never wrapped in a cluster
- Clusters are **dynamic**: as the user zooms into the globe the cluster splits into smaller
  clusters or individual pins when the events are far enough apart at the current zoom level
- Clicking a cluster pin opens the multi-event drawer **immediately** — no zoom-in step
- The multi-event drawer shows a compact, scrollable list of all events in the cluster
- Clicking an event row in the drawer closes the drawer and opens `EventDetailView` for that event
- Clicking an individual pin (not part of a cluster) opens `EventDetailView` directly, same as today
- When active filters change and events reload, any open multi-event drawer closes automatically

---

## Cluster Pin Visual Rules

Cluster pins must be visually distinct from individual pins at a glance.

| Property | Individual pin | Cluster pin |
|---|---|---|
| Color | `#facc15` default · `#ec4899` selected · `#ef4444` past | `#a78bfa` (violet — distinct from all individual colors) |
| Radius | `0.8` (default) · `1.2` (selected) | Scales with event count (see table below) |
| Label (tooltip) | `Event Name (Year)` | `N events` — e.g. `"3 events"` |

**Cluster radius scale:**

| Events in cluster | Point radius |
|---|---|
| 2 | `1.4` |
| 3–5 | `1.8` |
| 6–10 | `2.4` |
| 11+ | `3.0` |

A cluster pin never turns pink (selected color) even if the currently selected event is
one of its members. On cluster split the individual selected event renders as pink as normal.

---

## Multi-Event Drawer

- Opens immediately when a cluster pin is clicked
- Title: `"N events in this area"` (translated via i18n key `home.cluster.drawerTitle`)
- Each row shows: **event name** (primary text) + **date range** (secondary text, muted)
- Rows are sorted by `start_date` ascending
- Drawer is scrollable when the list exceeds the visible area
- Clicking a row: drawer closes → `EventDetailView` opens for the selected event
- Drawer can be dismissed with the close button or by pressing Escape — returns to globe with no event selected

---

## Zoom-Based Cluster Splitting

- Globe altitude is converted to an approximate zoom level to feed into `supercluster`
- When the user pans or zooms, the cluster data is recomputed (debounced — no recompute during active drag)
- The current `supercluster` radius parameter: **`40px`** at zoom level **`1`** (configurable constant, not hardcoded in component logic)
- If a cluster splits while the multi-event drawer is open, the drawer content is **frozen** at
  the time of click — it does not update live. The user can close it and click the now-individual pins.

---

## Info Modal — Legend Update

The `InfoModal` pin legend section already documents yellow, red, and pink pins via
`PinLegendRow`. A **violet row must be added** to explain the cluster pin to users:

| Property | Value |
|---|---|
| Color | `#a78bfa` (exported as `PIN_COLOR_CLUSTER` from `lib/events.ts`) |
| Label key | `info.violetPin.label` |
| Description key | `info.violetPin.description` |

The violet row is inserted **after the pink row** (selected) so the legend reads:
yellow → red → pink → violet.

---

## i18n Keys Required

All four locales (`en`, `pt`, `es`, `de`) must have every key below:

| Key | English value |
|---|---|
| `home.cluster.drawerTitle` | `"{count} events in this area"` |
| `home.cluster.eventCount` | `"{count} events"` (used as pin label/tooltip) |
| `info.violetPin.label` | `"Cluster"` |
| `info.violetPin.description` | `"Multiple events in the same area — click to see the list"` |

---

## Permissions

Public — no authentication required. The globe is accessible to all visitors.

---

## Error Cases

| Scenario | Expected behaviour |
|---|---|
| Cluster event detail fetch fails | Standard API error toast (existing `handleApiError` flow) |
| Zoom changes while multi-event drawer is open | Drawer stays open with frozen list; cluster may have split behind it |
| Filtering changes while multi-event drawer is open | Drawer closes; events reload as normal |
| All events cluster into one pin at low zoom | Correct — user zooms in to split, or opens drawer to pick |
| Single event remains after filtering | Renders as individual pin, not a cluster |
| `supercluster` receives an empty events array | No cluster pins rendered; existing empty-state message shown |

---

## Border / Corner Cases

- **Exactly 2 events that split on zoom** — they become 2 individual pins; if the drawer was open
  it stays frozen showing both; user can close and click individual pins
- **Selected event inside a cluster** — cluster renders in violet (not pink); once the cluster splits
  and the event becomes an individual pin, it renders pink as normal
- **Same event in two clusters** — impossible; `supercluster` guarantees each point belongs to at
  most one cluster at any given zoom level
- **Cluster with past events mixed with future events** — cluster pin is always violet regardless
  of the mix; individual past events only turn red once the cluster splits

---

## Definition of Done

- [ ] Events within the proximity radius are grouped into cluster pins on the globe
- [ ] Minimum cluster size is 2 (a single event is never clustered)
- [ ] Cluster pins are violet (`#a78bfa`) and visually distinct from individual pins
- [ ] Cluster pin radius scales proportionally with event count per the table above
- [ ] Cluster pin tooltip/label shows `"N events"` (i18n)
- [ ] Clicking a cluster opens the multi-event drawer immediately (no zoom step)
- [ ] Multi-event drawer title shows `"N events in this area"` (i18n)
- [ ] Drawer lists all events in the cluster: name + date range, sorted by start date
- [ ] Clicking an event row closes the drawer and opens `EventDetailView`
- [ ] Drawer can be closed/dismissed with no event selected
- [ ] Clusters split dynamically as the user zooms in
- [ ] Cluster data is recomputed on zoom change (debounced — not during active drag)
- [ ] Multi-event drawer content is frozen at click time (does not update on live zoom change)
- [ ] Filter change closes any open multi-event drawer
- [ ] Individual pins (not in any cluster) behave identically to today
- [ ] `PIN_COLOR_CLUSTER` constant (`#a78bfa`) exported from `lib/events.ts`
- [ ] InfoModal legend has a violet `PinLegendRow` inserted after the pink row
- [ ] All four locales have all four new i18n keys with no missing translations
- [ ] Works on mobile viewport (drawer uses existing mobile-compatible drawer component)
