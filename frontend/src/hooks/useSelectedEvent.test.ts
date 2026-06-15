import { describe, it, expect } from "vitest"
import { renderHook, act } from "@testing-library/react"

import { useSelectedEvent } from "./useSelectedEvent"
import type { EventListItem } from "@/types/api"

function makeEvent(id: number): EventListItem {
  return {
    id,
    name: `Event ${id}`,
    slug: `event-${id}`,
    country: "Brazil",
    city: "Sao Paulo",
    latitude: -23.55,
    longitude: -46.63,
    start_date: "2026-05-01",
    end_date: "2026-05-05",
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

describe("useSelectedEvent", () => {
  it("starts with no selected event", () => {
    const { result } = renderHook(() => useSelectedEvent())

    expect(result.current.selectedEvent).toBeNull()
  })

  it("selects an event when selectEvent is called", () => {
    const { result } = renderHook(() => useSelectedEvent())
    const event = makeEvent(1)

    act(() => result.current.selectEvent(event))

    expect(result.current.selectedEvent).toEqual(event)
  })

  it("switches to a different event when selectEvent is called with a different event", () => {
    const { result } = renderHook(() => useSelectedEvent())
    const first = makeEvent(1)
    const second = makeEvent(2)

    act(() => result.current.selectEvent(first))
    act(() => result.current.selectEvent(second))

    expect(result.current.selectedEvent).toEqual(second)
  })

  it("deselects when selectEvent is called again with the currently selected event", () => {
    const { result } = renderHook(() => useSelectedEvent())
    const event = makeEvent(1)

    act(() => result.current.selectEvent(event))
    act(() => result.current.selectEvent(event))

    expect(result.current.selectedEvent).toBeNull()
  })

  it("closes the detail view when closeDetail is called", () => {
    const { result } = renderHook(() => useSelectedEvent())
    const event = makeEvent(1)

    act(() => result.current.selectEvent(event))
    act(() => result.current.closeDetail())

    expect(result.current.selectedEvent).toBeNull()
  })
})
