import { describe, it, expect, vi, beforeEach } from "vitest"
import { renderHook, waitFor, act } from "@testing-library/react"

import { useReviewEvents, isOwnEvent } from "./useReviewEvents"
import { listEvents } from "@/lib/api/events"
import type { EventListItem, ApiMeta } from "@/types/api"

vi.mock("@/lib/api/events", () => ({
  listEvents: vi.fn(),
}))

// --- Fixtures ---

const meta: ApiMeta = { page: 1, total: 30 }

const makeEvent = (overrides: Partial<EventListItem> = {}): EventListItem => ({
  id: 1,
  name: "MODELS 2026",
  slug: "models-2026",
  country: "Brazil",
  city: "Sao Paulo",
  latitude: -23.55,
  longitude: -46.63,
  start_date: "2026-09-01",
  end_date: "2026-09-05",
  website_url: "https://models2026.example.com",
  domain: "software_engineering",
  status: "pending",
  tier: "A",
  year: 2026,
  created_by: { id: 10, name: "Alice", email: "alice@example.com" },
  last_updated_by: { id: 10, name: "Alice", email: "alice@example.com" },
  deadlines: [],
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
  ...overrides,
})

// --- useReviewEvents ---

describe("useReviewEvents", () => {
  beforeEach(() => {
    vi.mocked(listEvents).mockReset()
  })

  describe("initial load", () => {
    // Spec: "On initial load, fetch with defaults: status=pending, no tier, year=currentYear, page=1, page_size=30"
    it("fetches with status=pending, current year, page=1, page_size=30, pagination=on", async () => {
      vi.mocked(listEvents).mockResolvedValue({ data: [makeEvent()], meta })

      const { result } = renderHook(() => useReviewEvents(2026))

      expect(result.current.phase).toBe("loading")

      await waitFor(() => expect(result.current.phase).toBe("ready"))

      expect(listEvents).toHaveBeenCalledWith({
        status: "pending",
        year: 2026,
        page: 1,
        page_size: 30,
        pagination: "on",
      })
    })

    it("stores the returned events and meta after a successful fetch", async () => {
      const event = makeEvent()
      vi.mocked(listEvents).mockResolvedValue({ data: [event], meta })

      const { result } = renderHook(() => useReviewEvents(2026))

      await waitFor(() => expect(result.current.phase).toBe("ready"))

      expect(result.current.events).toEqual([event])
      expect(result.current.meta).toEqual(meta)
    })

    it("omits tier from the initial request (default = all)", async () => {
      vi.mocked(listEvents).mockResolvedValue({ data: [], meta: { page: 1, total: 0 } })

      const { result } = renderHook(() => useReviewEvents(2026))

      await waitFor(() => expect(result.current.phase).toBe("ready"))

      const calledWith = vi.mocked(listEvents).mock.calls[0][0]
      expect(calledWith).not.toHaveProperty("tier")
    })
  })

  describe("apply", () => {
    // Spec: "Apply button triggers a fetch with current filter values and resets page to 1"
    it("fetches page 1 with the new filter values when apply is called", async () => {
      vi.mocked(listEvents).mockResolvedValue({ data: [], meta: { page: 1, total: 0 } })

      const { result } = renderHook(() => useReviewEvents(2026))

      await waitFor(() => expect(result.current.phase).toBe("ready"))
      vi.mocked(listEvents).mockClear()

      act(() => {
        result.current.setDraftStatus("approved")
        result.current.setDraftTier("A*")
        result.current.setDraftYear(2025)
      })

      vi.mocked(listEvents).mockResolvedValue({ data: [], meta: { page: 1, total: 0 } })

      act(() => {
        result.current.apply()
      })

      await waitFor(() => expect(result.current.phase).toBe("ready"))

      expect(listEvents).toHaveBeenCalledWith({
        status: "approved",
        tier: "A*",
        year: 2025,
        page: 1,
        page_size: 30,
        pagination: "on",
      })
      expect(result.current.page).toBe(1)
    })

    // Spec: "Filter Apply while already on the filters' page → refetch page 1 regardless"
    it("resets to page 1 even when already on page 1", async () => {
      vi.mocked(listEvents).mockResolvedValue({ data: [], meta: { page: 1, total: 0 } })

      const { result } = renderHook(() => useReviewEvents(2026))

      await waitFor(() => expect(result.current.phase).toBe("ready"))
      vi.mocked(listEvents).mockClear()

      vi.mocked(listEvents).mockResolvedValue({ data: [], meta: { page: 1, total: 0 } })

      act(() => {
        result.current.apply()
      })

      await waitFor(() => expect(result.current.phase).toBe("ready"))

      const calledWith = vi.mocked(listEvents).mock.calls[0][0]
      expect(calledWith).toMatchObject({ page: 1 })
    })

    // Spec: year-filter-from-semantics.md — setDraftYear(undefined) clears year filter
    it("setDraftYear(undefined) clears year in draft and omits year from request on apply", async () => {
      vi.mocked(listEvents).mockResolvedValue({ data: [], meta: { page: 1, total: 0 } })

      const { result } = renderHook(() => useReviewEvents(2026))

      await waitFor(() => expect(result.current.phase).toBe("ready"))
      vi.mocked(listEvents).mockClear()

      act(() => {
        result.current.setDraftYear(undefined)
      })

      vi.mocked(listEvents).mockResolvedValue({ data: [], meta: { page: 1, total: 0 } })

      act(() => {
        result.current.apply()
      })

      await waitFor(() => expect(result.current.phase).toBe("ready"))

      const calledWith = vi.mocked(listEvents).mock.calls[0][0]
      expect(calledWith).not.toHaveProperty("year")
    })

    // Spec: "tier=all → omit tier param from request entirely"
    it("omits tier from the request when draft tier is undefined (all)", async () => {
      vi.mocked(listEvents).mockResolvedValue({ data: [], meta: { page: 1, total: 0 } })

      const { result } = renderHook(() => useReviewEvents(2026))

      await waitFor(() => expect(result.current.phase).toBe("ready"))
      vi.mocked(listEvents).mockClear()

      act(() => {
        result.current.setDraftTier(undefined)
      })

      vi.mocked(listEvents).mockResolvedValue({ data: [], meta: { page: 1, total: 0 } })

      act(() => {
        result.current.apply()
      })

      await waitFor(() => expect(result.current.phase).toBe("ready"))

      const calledWith = vi.mocked(listEvents).mock.calls[0][0]
      expect(calledWith).not.toHaveProperty("tier")
    })
  })

  describe("goToPage", () => {
    // Spec: "Clicking Prev/Next fetches the adjacent page with the current applied filters
    //        (not the draft filter values)"
    it("fetches with applied filters (not draft) and the given page number", async () => {
      vi.mocked(listEvents).mockResolvedValue({ data: [], meta: { page: 1, total: 90 } })

      const { result } = renderHook(() => useReviewEvents(2026))

      await waitFor(() => expect(result.current.phase).toBe("ready"))
      vi.mocked(listEvents).mockClear()

      // Change draft but do NOT apply — goToPage must use applied, not draft
      act(() => {
        result.current.setDraftStatus("approved")
        result.current.setDraftYear(2025)
      })

      vi.mocked(listEvents).mockResolvedValue({ data: [], meta: { page: 2, total: 90 } })

      act(() => {
        result.current.goToPage(2)
      })

      await waitFor(() => expect(result.current.phase).toBe("ready"))

      expect(listEvents).toHaveBeenCalledWith({
        status: "pending",  // applied value, not draft "approved"
        year: 2026,         // applied value, not draft 2025
        page: 2,
        page_size: 30,
        pagination: "on",
      })
      expect(result.current.page).toBe(2)
    })
  })

  describe("error", () => {
    // Spec: "Error state shown when API call fails"
    it("sets phase to error when listEvents rejects", async () => {
      vi.mocked(listEvents).mockRejectedValue(new Error("network error"))

      const { result } = renderHook(() => useReviewEvents(2026))

      await waitFor(() => expect(result.current.phase).toBe("error"))

      expect(result.current.events).toEqual([])
      expect(result.current.meta).toBeNull()
    })
  })
})

// --- isOwnEvent ---

describe("isOwnEvent", () => {
  // Spec: "When event.created_by.id === sessionUser.id and role is moderator → grey card"
  it("returns true when the event was created by the given user", () => {
    const event = makeEvent({ created_by: { id: 42, name: "Bob", email: "bob@example.com" } })
    expect(isOwnEvent(event, 42)).toBe(true)
  })

  // Spec: "Admin role never sees own-event grey variant"
  it("returns false when the event was created by a different user", () => {
    const event = makeEvent({ created_by: { id: 42, name: "Bob", email: "bob@example.com" } })
    expect(isOwnEvent(event, 99)).toBe(false)
  })
})
