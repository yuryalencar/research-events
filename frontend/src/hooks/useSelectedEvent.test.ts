import { describe, it, expect, vi, beforeEach } from "vitest"
import { renderHook, act } from "@testing-library/react"

// Must be declared before any import that transitively imports next/navigation.
vi.mock("next/navigation", () => ({
  useRouter: vi.fn(),
  useSearchParams: vi.fn(),
  usePathname: vi.fn(),
}))

import { useRouter, useSearchParams, usePathname } from "next/navigation"
import { useSelectedEvent } from "./useSelectedEvent"
import type { EventListItem } from "@/types/api"

// --- Helpers ---

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

// Stable mock router object so the reference doesn't change between renders
// (changing reference would re-trigger effects unnecessarily).
const mockRouter = { replace: vi.fn() }

// Configures the useSearchParams mock for a given query string (e.g. "event=foo").
// Extracted because the URLSearchParams → ReturnType cast is verbose and repeated.
function setSearchParams(init = ""): void {
  vi.mocked(useSearchParams).mockReturnValue(
    new URLSearchParams(init) as unknown as ReturnType<typeof useSearchParams>,
  )
}

// --- Tests ---

describe("useSelectedEvent", () => {
  beforeEach(() => {
    mockRouter.replace = vi.fn()
    vi.mocked(useRouter).mockReturnValue(mockRouter as unknown as ReturnType<typeof useRouter>)
    vi.mocked(useSearchParams).mockReturnValue(
      new URLSearchParams() as unknown as ReturnType<typeof useSearchParams>,
    )
    vi.mocked(usePathname).mockReturnValue("/en")
  })

  // --- Existing state behaviour (unchanged) ---

  it("starts with no selected event", () => {
    const { result } = renderHook(() => useSelectedEvent([]))

    expect(result.current.selectedEvent).toBeNull()
  })

  it("selects an event when selectEvent is called", () => {
    const { result } = renderHook(() => useSelectedEvent([]))
    const event = makeEvent(1)

    act(() => result.current.selectEvent(event))

    expect(result.current.selectedEvent).toEqual(event)
  })

  it("switches to a different event when selectEvent is called with a different event", () => {
    const { result } = renderHook(() => useSelectedEvent([]))
    const first = makeEvent(1)
    const second = makeEvent(2)

    act(() => result.current.selectEvent(first))
    act(() => result.current.selectEvent(second))

    expect(result.current.selectedEvent).toEqual(second)
  })

  it("deselects when selectEvent is called again with the currently selected event", () => {
    const { result } = renderHook(() => useSelectedEvent([]))
    const event = makeEvent(1)

    act(() => result.current.selectEvent(event))
    act(() => result.current.selectEvent(event))

    expect(result.current.selectedEvent).toBeNull()
  })

  it("closes the detail view when closeDetail is called", () => {
    const { result } = renderHook(() => useSelectedEvent([]))
    const event = makeEvent(1)

    act(() => result.current.selectEvent(event))
    act(() => result.current.closeDetail())

    expect(result.current.selectedEvent).toBeNull()
  })

  // --- Cycle 1: URL writes ---
  // Spec: "Clicking a pin → router.replace with ?event=SLUG"
  // Spec: "Closing panel → router.replace removing ?event="

  describe("URL writes", () => {
    it("sets ?event=slug in the URL when selectEvent is called", () => {
      const { result } = renderHook(() => useSelectedEvent([]))
      const event = makeEvent(1) // slug: 'event-1'

      act(() => result.current.selectEvent(event))

      expect(mockRouter.replace).toHaveBeenCalledWith("?event=event-1")
    })

    it("removes ?event= from the URL when closeDetail is called", () => {
      setSearchParams("event=event-1")
      const { result } = renderHook(() => useSelectedEvent([]))

      act(() => result.current.closeDetail())

      // No remaining params → replace with pathname alone
      expect(mockRouter.replace).toHaveBeenCalledWith("/en")
    })

    it("removes ?event= from the URL when the currently-selected event is re-clicked", () => {
      const { result } = renderHook(() => useSelectedEvent([]))
      const event = makeEvent(1)

      act(() => result.current.selectEvent(event))
      act(() => result.current.selectEvent(event)) // toggle off

      // The last replace call should remove the param
      expect(mockRouter.replace).toHaveBeenLastCalledWith("/en")
    })

    it("preserves other query params when adding ?event=slug", () => {
      setSearchParams("foo=bar")
      const { result } = renderHook(() => useSelectedEvent([]))
      const event = makeEvent(1)

      act(() => result.current.selectEvent(event))

      expect(mockRouter.replace).toHaveBeenCalledWith("?foo=bar&event=event-1")
    })

    it("preserves other query params when removing ?event=slug", () => {
      setSearchParams("foo=bar&event=event-1")
      const { result } = renderHook(() => useSelectedEvent([]))

      act(() => result.current.closeDetail())

      expect(mockRouter.replace).toHaveBeenCalledWith("?foo=bar")
    })
  })

  // --- Cycle 2: URL reads (slug resolution on page load) ---
  // Spec: "Once events arrive, find event where event.slug === SLUG"
  // Spec: "Not found → router.replace removing ?event= silently"
  // Spec: "While events still loading → no action"

  describe("URL reads on load", () => {
    it("selects the matching event when ?event=slug is in the URL and events arrive", () => {
      setSearchParams("event=event-1")
      const event = makeEvent(1)

      const { result, rerender } = renderHook(
        ({ events }: { events: EventListItem[] }) => useSelectedEvent(events),
        { initialProps: { events: [] as EventListItem[] } },
      )

      // No selection yet while events are still loading
      expect(result.current.selectedEvent).toBeNull()

      act(() => rerender({ events: [event] }))

      expect(result.current.selectedEvent).toEqual(event)
    })

    it("removes ?event= from the URL silently when the slug matches no event", () => {
      setSearchParams("event=unknown-slug")
      const event = makeEvent(1) // slug: 'event-1', not 'unknown-slug'

      const { result, rerender } = renderHook(
        ({ events }: { events: EventListItem[] }) => useSelectedEvent(events),
        { initialProps: { events: [] as EventListItem[] } },
      )

      act(() => rerender({ events: [event] }))

      expect(result.current.selectedEvent).toBeNull()
      expect(mockRouter.replace).toHaveBeenCalledWith("/en")
    })

    it("takes no action while the events list is empty (still loading)", () => {
      setSearchParams("event=event-1")

      const { result } = renderHook(() => useSelectedEvent([]))

      expect(result.current.selectedEvent).toBeNull()
      expect(mockRouter.replace).not.toHaveBeenCalled()
    })
  })
})
