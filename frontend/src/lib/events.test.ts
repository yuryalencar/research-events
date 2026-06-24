import { describe, it, expect } from "vitest"

import {
  getTierBadgeLabel,
  shouldShowUpdatedBy,
  getPinColor,
  isEventPast,
  formatDateRange,
  getClusterPinRadius,
  PIN_COLOR_DEFAULT,
  PIN_COLOR_SELECTED,
  PIN_COLOR_PAST,
} from "./events"
import type { EventListItem } from "@/types/api"

const baseEvent: EventListItem = {
  id: 1,
  name: "ICSE 2026",
  slug: "icse-2026",
  country: "Brazil",
  city: "Sao Paulo",
  latitude: -23.55,
  longitude: -46.63,
  start_date: "2026-05-01",
  end_date: "2026-05-05",
  website_url: "https://icse2026.example.com",
  domain: "software_engineering",
  status: "approved",
  tier: "A*",
  year: 2026,
  created_by: { id: 1, name: "Alice", email: "alice@example.com" },
  last_updated_by: { id: 1, name: "Alice", email: "alice@example.com" },
  deadlines: [],
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
}

describe("getTierBadgeLabel", () => {
  it("returns the tier when it is not unranked", () => {
    expect(getTierBadgeLabel("A*")).toBe("A*")
    expect(getTierBadgeLabel("A")).toBe("A")
    expect(getTierBadgeLabel("B")).toBe("B")
    expect(getTierBadgeLabel("C")).toBe("C")
  })

  it("returns null when the tier is unranked", () => {
    expect(getTierBadgeLabel("unranked")).toBeNull()
  })
})

describe("shouldShowUpdatedBy", () => {
  it("returns false when created_by and last_updated_by are the same user", () => {
    const event: EventListItem = { ...baseEvent, created_by: { id: 1, name: "Alice", email: "a@example.com" }, last_updated_by: { id: 1, name: "Alice", email: "a@example.com" } }

    expect(shouldShowUpdatedBy(event)).toBe(false)
  })

  it("returns true when last_updated_by differs from created_by", () => {
    const event: EventListItem = {
      ...baseEvent,
      created_by: { id: 1, name: "Alice", email: "a@example.com" },
      last_updated_by: { id: 2, name: "Bob", email: "b@example.com" },
    }

    expect(shouldShowUpdatedBy(event)).toBe(true)
  })
})

describe("formatDateRange", () => {
  it("returns a single formatted date when start and end are the same day", () => {
    expect(formatDateRange("2026-04-13", "2026-04-13", "en")).toBe("Apr 13, 2026")
  })

  it("condenses the day range when start and end fall in the same month", () => {
    expect(formatDateRange("2026-04-13", "2026-04-19", "en")).toBe("Apr 13–19, 2026")
  })

  it("shows both months when start and end fall in the same year but different months", () => {
    expect(formatDateRange("2026-04-28", "2026-05-02", "en")).toBe("Apr 28 – May 2, 2026")
  })

  it("shows full dates including year for both ends when they span different years", () => {
    expect(formatDateRange("2025-12-28", "2026-01-03", "en")).toBe("Dec 28, 2025 – Jan 3, 2026")
  })

  it("formats dates using the given locale", () => {
    expect(formatDateRange("2026-04-13", "2026-04-19", "pt")).toBe("13 de abr.–19, 2026")
  })
})

describe("isEventPast", () => {
  it("returns true when the end date is before now", () => {
    expect(isEventPast("2026-01-01", new Date("2026-02-01T00:00:00Z"))).toBe(true)
  })

  it("returns false when the end date is after now", () => {
    expect(isEventPast("2026-05-05", new Date("2026-02-01T00:00:00Z"))).toBe(false)
  })
})

describe("getClusterPinRadius", () => {
  it("returns 1.4 for a cluster of exactly 2 events", () => {
    expect(getClusterPinRadius(2)).toBe(1.4)
  })

  it("returns 1.8 for a cluster of 3 events (lower bound of 3–5 range)", () => {
    expect(getClusterPinRadius(3)).toBe(1.8)
  })

  it("returns 1.8 for a cluster of 5 events (upper bound of 3–5 range)", () => {
    expect(getClusterPinRadius(5)).toBe(1.8)
  })

  it("returns 2.4 for a cluster of 6 events (lower bound of 6–10 range)", () => {
    expect(getClusterPinRadius(6)).toBe(2.4)
  })

  it("returns 2.4 for a cluster of 10 events (upper bound of 6–10 range)", () => {
    expect(getClusterPinRadius(10)).toBe(2.4)
  })

  it("returns 3.0 for a cluster of 11 events (lower bound of 11+ range)", () => {
    expect(getClusterPinRadius(11)).toBe(3.0)
  })

  it("returns 3.0 for a cluster of 100 events (clamps at max radius)", () => {
    expect(getClusterPinRadius(100)).toBe(3.0)
  })
})

describe("getPinColor", () => {
  const now = new Date("2026-02-01T00:00:00Z")

  it("returns the selected color when the event is selected", () => {
    expect(getPinColor(1, "2026-05-05", 1, now)).toBe(PIN_COLOR_SELECTED)
  })

  it("returns the default color for an upcoming event that is not selected", () => {
    expect(getPinColor(1, "2026-05-05", 2, now)).toBe(PIN_COLOR_DEFAULT)
  })

  it("returns the default color when no event is selected", () => {
    expect(getPinColor(1, "2026-05-05", null, now)).toBe(PIN_COLOR_DEFAULT)
  })

  it("returns the past color for an event that has already ended", () => {
    expect(getPinColor(1, "2026-01-01", null, now)).toBe(PIN_COLOR_PAST)
  })

  it("returns the selected color for a past event that is currently selected", () => {
    expect(getPinColor(1, "2026-01-01", 1, now)).toBe(PIN_COLOR_SELECTED)
  })
})
