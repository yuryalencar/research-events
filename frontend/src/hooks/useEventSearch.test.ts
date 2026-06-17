import { describe, it, expect, vi, beforeEach } from "vitest"
import { renderHook, waitFor, act } from "@testing-library/react"

import { useEventSearch } from "./useEventSearch"
import { listEvents } from "@/lib/api/events"
import type { EventListItem } from "@/types/api"

vi.mock("@/lib/api/events", () => ({
  listEvents: vi.fn(),
}))

// --- Fixtures ---

function makeEvent(overrides: Partial<EventListItem> = {}): EventListItem {
  return {
    id: 1,
    name: "ICSE 2026",
    slug: "ICSE2026",
    country: "Brazil",
    city: "São Paulo",
    latitude: -23.55,
    longitude: -46.63,
    start_date: "2026-05-01",
    end_date: "2026-05-05",
    website_url: "https://example.com",
    domain: "computer_science",
    status: "pending",
    tier: "A*",
    year: 2026,
    created_by: { id: 1, name: "Alice", email: "alice@example.com" },
    last_updated_by: { id: 1, name: "Alice", email: "alice@example.com" },
    deadlines: [],
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    ...overrides,
  }
}

// --- Tests ---

describe("useEventSearch", () => {
  const currentYear = new Date().getFullYear()

  beforeEach(() => {
    vi.mocked(listEvents).mockReset()
  })

  // Spec: "Step 1 fetches GET /api/v1/events?status=pending&year=<year>&page=<page>&page_size=10"
  it("fetches pending events with the current year and page 1 on mount", async () => {
    vi.mocked(listEvents).mockResolvedValue({ data: [], meta: { page: 1, total: 0 } })

    const { result } = renderHook(() => useEventSearch(10))

    await waitFor(() => expect(result.current.isLoading).toBe(false))

    expect(listEvents).toHaveBeenCalledWith({
      status: "pending",
      year: currentYear,
      page: 1,
      page_size: 10,
    })
  })

  // Spec: "Apply pressed (any filter state) → always re-fetch at page 1"
  it("apply() re-fetches at page 1 regardless of current page", async () => {
    vi.mocked(listEvents).mockResolvedValue({ data: [], meta: { page: 1, total: 30 } })

    const { result } = renderHook(() => useEventSearch(10))

    await waitFor(() => expect(result.current.isLoading).toBe(false))

    vi.mocked(listEvents).mockResolvedValue({ data: [], meta: { page: 2, total: 30 } })
    act(() => result.current.goToPage(2))
    await waitFor(() => expect(result.current.page).toBe(2))

    vi.mocked(listEvents).mockResolvedValue({ data: [], meta: { page: 1, total: 30 } })
    act(() => result.current.apply())

    await waitFor(() =>
      expect(listEvents).toHaveBeenLastCalledWith(expect.objectContaining({ page: 1 }))
    )
  })

  // Spec: "All search fields cleared + Apply → re-fetch page 1 with no filter applied"
  it("clear() resets text search and year to defaults then re-fetches at page 1", async () => {
    vi.mocked(listEvents).mockResolvedValue({ data: [], meta: { page: 1, total: 0 } })

    const { result } = renderHook(() => useEventSearch(10))

    await waitFor(() => expect(result.current.isLoading).toBe(false))

    act(() => {
      result.current.setSearchText("ICSE")
      result.current.setYear(2025)
    })

    vi.mocked(listEvents).mockResolvedValue({ data: [], meta: { page: 1, total: 0 } })
    act(() => result.current.clear())

    await waitFor(() => expect(result.current.isLoading).toBe(false))

    expect(result.current.searchText).toBe("")
    expect(result.current.year).toBe(currentYear)
    expect(listEvents).toHaveBeenLastCalledWith(
      expect.objectContaining({ page: 1, year: currentYear })
    )
  })

  // Spec: "Previous/Next re-fetch at the new page number without resetting filters"
  it("goToPage(n) fetches at page n without resetting the year filter", async () => {
    vi.mocked(listEvents).mockResolvedValue({ data: [], meta: { page: 1, total: 30 } })

    const { result } = renderHook(() => useEventSearch(10))

    await waitFor(() => expect(result.current.isLoading).toBe(false))

    act(() => result.current.setYear(2025))
    act(() => result.current.apply())
    await waitFor(() => expect(result.current.isLoading).toBe(false))

    vi.mocked(listEvents).mockResolvedValue({ data: [], meta: { page: 2, total: 30 } })
    act(() => result.current.goToPage(2))

    await waitFor(() => expect(result.current.page).toBe(2))
    expect(listEvents).toHaveBeenLastCalledWith(
      expect.objectContaining({ page: 2, year: 2025 })
    )
  })

  // Spec: "Text search filters by name or slug client-side on current page"
  it("filteredEvents filters by event name case-insensitively", async () => {
    const events = [
      makeEvent({ id: 1, name: "ICSE 2026", slug: "ICSE2026" }),
      makeEvent({ id: 2, name: "MODELS 2026", slug: "MODELS2026" }),
    ]
    vi.mocked(listEvents).mockResolvedValue({ data: events, meta: { page: 1, total: 2 } })

    const { result } = renderHook(() => useEventSearch(10))

    await waitFor(() => expect(result.current.isLoading).toBe(false))

    act(() => result.current.setSearchText("icse"))

    expect(result.current.filteredEvents).toHaveLength(1)
    expect(result.current.filteredEvents[0].name).toBe("ICSE 2026")
  })

  it("filteredEvents filters by slug case-insensitively", async () => {
    const events = [
      makeEvent({ id: 1, name: "ICSE 2026", slug: "ICSE2026" }),
      makeEvent({ id: 2, name: "MODELS 2026", slug: "MODELS2026" }),
    ]
    vi.mocked(listEvents).mockResolvedValue({ data: events, meta: { page: 1, total: 2 } })

    const { result } = renderHook(() => useEventSearch(10))

    await waitFor(() => expect(result.current.isLoading).toBe(false))

    act(() => result.current.setSearchText("models"))

    expect(result.current.filteredEvents).toHaveLength(1)
    expect(result.current.filteredEvents[0].slug).toBe("MODELS2026")
  })

  // Spec: "meta.total drives the total count and page calculation"
  it("totalPages is derived correctly from meta.total and pageSize", async () => {
    vi.mocked(listEvents).mockResolvedValue({ data: [], meta: { page: 1, total: 25 } })

    const { result } = renderHook(() => useEventSearch(10))

    await waitFor(() => expect(result.current.isLoading).toBe(false))

    expect(result.current.totalPages).toBe(3) // Math.ceil(25 / 10)
    expect(result.current.total).toBe(25)
  })

  // Spec: API error → "Couldn't load the events list. Continue button still works."
  it("sets error when the API call fails", async () => {
    vi.mocked(listEvents).mockRejectedValue(new Error("network error"))

    const { result } = renderHook(() => useEventSearch(10))

    await waitFor(() => expect(result.current.isLoading).toBe(false))

    expect(result.current.error).toBeTruthy()
    expect(result.current.events).toEqual([])
  })
})
