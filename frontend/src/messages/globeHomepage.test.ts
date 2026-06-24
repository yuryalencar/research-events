import { describe, it, expect } from "vitest"

import en from "./en.json"
import pt from "./pt.json"
import es from "./es.json"
import de from "./de.json"

// Keys for the Globe Homepage feature (specs/frontend/globe-homepage.md):
// the "home" namespace covers the page-level loading state, and
// "eventDetail" covers every label in the side panel / bottom card.
const EXPECTED_HOME_KEYS = [
  "loading",
  "noEvents",
  "noEventsFiltered",
  "cluster.drawerTitle",
  "cluster.eventCount",
].sort()

const EXPECTED_EVENT_DETAIL_KEYS = [
  "detailsDescription",
  "tier",
  "dates",
  "location",
  "website",
  "domain",
  "deadlinesTitle",
  "noUpcomingDeadlines",
  "manageDeadlinesLabel",
  "addedBy",
  "updatedBy",
  "optional",
  "deadlineTypes.abstract",
  "deadlineTypes.paper",
  "deadlineTypes.notification",
  "deadlineTypes.camera_ready",
  "deadlineTypes.other",
  "domains.computer_science",
].sort()

// flattenKeys turns a nested messages object into dot-notation leaf keys,
// e.g. { deadlineTypes: { abstract: "..." } } -> ["deadlineTypes.abstract"].
function flattenKeys(obj: Record<string, unknown>, prefix = ""): string[] {
  return Object.entries(obj).flatMap(([key, value]) => {
    const path = prefix ? `${prefix}.${key}` : key
    return typeof value === "object" && value !== null
      ? flattenKeys(value as Record<string, unknown>, path)
      : [path]
  })
}

describe("home namespace i18n key parity", () => {
  it("en.json (source of truth) has every expected home.* key", () => {
    expect(flattenKeys(en.home ?? {}).sort()).toEqual(EXPECTED_HOME_KEYS)
  })

  it.each([
    ["pt", pt],
    ["es", es],
    ["de", de],
  ])("%s.json mirrors every home.* key from en.json", (_locale, messages) => {
    expect(flattenKeys((messages as typeof en).home ?? {}).sort()).toEqual(EXPECTED_HOME_KEYS)
  })
})

describe("eventDetail namespace i18n key parity", () => {
  it("en.json (source of truth) has every expected eventDetail.* key", () => {
    expect(flattenKeys(en.eventDetail ?? {}).sort()).toEqual(EXPECTED_EVENT_DETAIL_KEYS)
  })

  it.each([
    ["pt", pt],
    ["es", es],
    ["de", de],
  ])("%s.json mirrors every eventDetail.* key from en.json", (_locale, messages) => {
    expect(flattenKeys((messages as typeof en).eventDetail ?? {}).sort()).toEqual(EXPECTED_EVENT_DETAIL_KEYS)
  })
})
