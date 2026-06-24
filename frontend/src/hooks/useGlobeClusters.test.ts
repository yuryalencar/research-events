import { describe, it, expect } from "vitest"
import { renderHook } from "@testing-library/react"

import { useGlobeClusters, isCluster } from "./useGlobeClusters"
import type { EventListItem } from "@/types/api"

// --- Helpers ---

function makeEvent(id: number, lat: number, lng: number): EventListItem {
  return {
    id,
    name: `Event ${id}`,
    slug: `event-${id}`,
    country: "Austria",
    city: "Vienna",
    latitude: lat,
    longitude: lng,
    start_date: "2026-09-01",
    end_date: "2026-09-05",
    website_url: "https://example.com",
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
}

// Vienna and Tokyo are far enough apart that they never cluster at zoom ≥ 3.
const VIENNA = { lat: 48.21, lng: 16.37 }
const TOKYO = { lat: 35.68, lng: 139.69 }

// --- Tests ---

describe("useGlobeClusters", () => {
  it("returns an empty array when events is empty", () => {
    const { result } = renderHook(() => useGlobeClusters([], 1))

    expect(result.current.globePoints).toHaveLength(0)
  })

  it("returns a single event as an individual GlobePoint (not a cluster)", () => {
    const events = [makeEvent(1, VIENNA.lat, VIENNA.lng)]

    const { result } = renderHook(() => useGlobeClusters(events, 1))

    expect(result.current.globePoints).toHaveLength(1)
    expect(isCluster(result.current.globePoints[0])).toBe(false)
  })

  it("groups 2 events at the same lat/lng into a single ClusterPoint at low zoom", () => {
    const events = [
      makeEvent(1, VIENNA.lat, VIENNA.lng),
      makeEvent(2, VIENNA.lat, VIENNA.lng),
    ]

    const { result } = renderHook(() => useGlobeClusters(events, 1))

    expect(result.current.globePoints).toHaveLength(1)
    expect(isCluster(result.current.globePoints[0])).toBe(true)
  })

  it("sets the cluster count to the number of grouped events", () => {
    const events = [
      makeEvent(1, VIENNA.lat, VIENNA.lng),
      makeEvent(2, VIENNA.lat, VIENNA.lng),
      makeEvent(3, VIENNA.lat, VIENNA.lng),
    ]

    const { result } = renderHook(() => useGlobeClusters(events, 1))

    const cluster = result.current.globePoints[0]
    expect(isCluster(cluster)).toBe(true)
    if (isCluster(cluster)) {
      expect(cluster.count).toBe(3)
    }
  })

  it("returns two individual GlobePoints for events on different continents at zoom 5", () => {
    const events = [
      makeEvent(1, VIENNA.lat, VIENNA.lng),
      makeEvent(2, TOKYO.lat, TOKYO.lng),
    ]

    const { result } = renderHook(() => useGlobeClusters(events, 5))

    expect(result.current.globePoints).toHaveLength(2)
    expect(result.current.globePoints.every((p) => !isCluster(p))).toBe(true)
  })

  it("returns the correct EventListItem[] for a given clusterId via getClusterEvents", () => {
    const event1 = makeEvent(1, VIENNA.lat, VIENNA.lng)
    const event2 = makeEvent(2, VIENNA.lat, VIENNA.lng)

    const { result } = renderHook(() => useGlobeClusters([event1, event2], 1))

    const cluster = result.current.globePoints[0]
    expect(isCluster(cluster)).toBe(true)

    if (isCluster(cluster)) {
      const members = result.current.getClusterEvents(cluster.clusterId)
      expect(members).toHaveLength(2)
      expect(members.map((e) => e.id).sort()).toEqual([1, 2])
    }
  })

  it("recomputes globePoints when the events array changes", () => {
    const initial = [makeEvent(1, VIENNA.lat, VIENNA.lng)]
    const updated = [
      makeEvent(1, VIENNA.lat, VIENNA.lng),
      makeEvent(2, VIENNA.lat, VIENNA.lng),
    ]

    const { result, rerender } = renderHook(
      ({ events, zoom }: { events: EventListItem[]; zoom: number }) =>
        useGlobeClusters(events, zoom),
      { initialProps: { events: initial, zoom: 1 } },
    )

    expect(result.current.globePoints).toHaveLength(1)
    expect(isCluster(result.current.globePoints[0])).toBe(false)

    rerender({ events: updated, zoom: 1 })

    expect(result.current.globePoints).toHaveLength(1)
    expect(isCluster(result.current.globePoints[0])).toBe(true)
  })

  it("splits a cluster into individual points when zoom exceeds supercluster maxZoom", () => {
    // Two events at the same location cluster at any zoom ≤ maxZoom (16).
    // Once zoom exceeds maxZoom, supercluster stops clustering and returns
    // individual points — simulating the globe being zoomed in very close.
    const events = [
      makeEvent(1, VIENNA.lat, VIENNA.lng),
      makeEvent(2, VIENNA.lat, VIENNA.lng),
    ]

    const { result, rerender } = renderHook(
      ({ zoom }: { zoom: number }) => useGlobeClusters(events, zoom),
      { initialProps: { zoom: 1 } },
    )

    expect(result.current.globePoints).toHaveLength(1)
    expect(isCluster(result.current.globePoints[0])).toBe(true)

    // zoom 17 exceeds maxZoom (16) — supercluster returns individual points
    rerender({ zoom: 17 })

    expect(result.current.globePoints).toHaveLength(2)
    expect(result.current.globePoints.every((p) => !isCluster(p))).toBe(true)
  })
})
