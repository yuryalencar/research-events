import type { EventListItem, EventTier } from "@/types/api"

// --- Constants ---

// Pin colors for the globe (see specs/frontend/globe-homepage.md — clicking a
// pin highlights it with a distinct color). Both colors are chosen to stand
// out against the blue/green/brown "earth-blue-marble" globe texture and the
// dark starfield background — blue and green are avoided since they blend
// into the globe itself.
const PIN_COLOR_DEFAULT = "#facc15"
const PIN_COLOR_SELECTED = "#ec4899"
const PIN_COLOR_PAST = "#ef4444"

// --- Pure helpers ---

// getTierBadgeLabel returns the tier to display as a badge, or null when the
// event is unranked — the detail view shows no badge in that case.
function getTierBadgeLabel(tier: EventTier): EventTier | null {
  return tier === "unranked" ? null : tier
}

// shouldShowUpdatedBy returns true when the event was last updated by someone
// other than its original submitter, so the detail view can show an
// "Updated by" line in addition to "Added by".
function shouldShowUpdatedBy(event: EventListItem): boolean {
  return event.last_updated_by.id !== event.created_by.id
}

// isEventPast returns true when an event's end date has already passed,
// relative to `now`. The end date is parsed as UTC so the result doesn't
// shift based on the viewer's timezone.
function isEventPast(endDate: string, now: Date): boolean {
  const end = new Date(`${endDate}T00:00:00Z`)
  return end.getTime() < now.getTime()
}

// getPinColor returns the color a pin should be rendered with: highlighted
// when its event is the currently selected one, red when the event has
// already ended, default otherwise.
function getPinColor(eventId: number, endDate: string, selectedEventId: number | null, now: Date): string {
  if (eventId === selectedEventId) return PIN_COLOR_SELECTED
  return isEventPast(endDate, now) ? PIN_COLOR_PAST : PIN_COLOR_DEFAULT
}

// formatDateRange turns two ISO "YYYY-MM-DD" dates into a localized,
// human-readable range — e.g. "Apr 13–19, 2026" when both dates fall in the
// same month, "Apr 13 – May 2, 2026" when they span months in the same year,
// or "Dec 28, 2025 – Jan 3, 2026" when they span years. Dates are parsed as
// UTC so the displayed day never shifts based on the viewer's timezone.
function formatDateRange(startDate: string, endDate: string, locale: string): string {
  const start = new Date(`${startDate}T00:00:00Z`)
  const end = new Date(`${endDate}T00:00:00Z`)

  const fullFormat = new Intl.DateTimeFormat(locale, { month: "short", day: "numeric", year: "numeric", timeZone: "UTC" })

  if (startDate === endDate) {
    return fullFormat.format(start)
  }

  const sameYear = start.getUTCFullYear() === end.getUTCFullYear()
  const sameMonth = sameYear && start.getUTCMonth() === end.getUTCMonth()

  if (sameMonth) {
    const monthDayFormat = new Intl.DateTimeFormat(locale, { month: "short", day: "numeric", timeZone: "UTC" })
    const dayFormat = new Intl.DateTimeFormat(locale, { day: "numeric", timeZone: "UTC" })
    return `${monthDayFormat.format(start)}–${dayFormat.format(end)}, ${start.getUTCFullYear()}`
  }

  if (sameYear) {
    const monthDayFormat = new Intl.DateTimeFormat(locale, { month: "short", day: "numeric", timeZone: "UTC" })
    return `${monthDayFormat.format(start)} – ${fullFormat.format(end)}`
  }

  return `${fullFormat.format(start)} – ${fullFormat.format(end)}`
}

// --- Export ---

export {
  getTierBadgeLabel,
  shouldShowUpdatedBy,
  getPinColor,
  isEventPast,
  formatDateRange,
  PIN_COLOR_DEFAULT,
  PIN_COLOR_SELECTED,
  PIN_COLOR_PAST,
}
