import { describe, it, expect, vi, beforeEach } from "vitest"
import { renderHook, waitFor } from "@testing-library/react"

import { useEvents } from "./useEvents"
import { listEvents } from "@/lib/api/events"
import { handleApiError } from "@/lib/api/errors"
import type { EventListItem } from "@/types/api"

vi.mock("@/lib/api/events", () => ({
  listEvents: vi.fn(),
}))

vi.mock("@/lib/api/errors", () => ({
  handleApiError: vi.fn(),
}))

vi.mock("next-intl", () => ({
  useTranslations: () => (key: string) => key,
}))

const event: EventListItem = {
  id: 1,
  name: "ICSE 2026",
  slug: "icse-2026",
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

const baseFilters = { year: 2026 }

describe("useEvents", () => {
  beforeEach(() => {
    vi.mocked(listEvents).mockReset()
    vi.mocked(handleApiError).mockReset()
  })

  it("starts in a loading state with no events", () => {
    vi.mocked(listEvents).mockReturnValue(new Promise(() => {}))

    const { result } = renderHook(() => useEvents(baseFilters))

    expect(result.current.isLoading).toBe(true)
    expect(result.current.events).toEqual([])
  })

  it("calls listEvents with pagination off and stores the returned events", async () => {
    vi.mocked(listEvents).mockResolvedValue({ data: [event], meta: { page: 1, total: 1 } })

    const { result } = renderHook(() => useEvents(baseFilters))

    await waitFor(() => expect(result.current.isLoading).toBe(false))

    expect(listEvents).toHaveBeenCalledWith({ year: 2026, pagination: "off" })
    expect(result.current.events).toEqual([event])
  })

  it("calls handleApiError and leaves events empty when the fetch fails", async () => {
    vi.mocked(listEvents).mockRejectedValue(new Error("network error"))

    const { result } = renderHook(() => useEvents(baseFilters))

    await waitFor(() => expect(result.current.isLoading).toBe(false))

    expect(handleApiError).toHaveBeenCalled()
    expect(result.current.events).toEqual([])
  })

  it("returns an empty events array when data is empty", async () => {
    vi.mocked(listEvents).mockResolvedValue({ data: [], meta: { page: 1, total: 0 } })

    const { result } = renderHook(() => useEvents(baseFilters))

    await waitFor(() => expect(result.current.isLoading).toBe(false))

    expect(result.current.events).toEqual([])
    expect(handleApiError).not.toHaveBeenCalled()
  })

  // --- Filter params ---
  // Spec: "year is always included in every listEvents call — never omitted" (Rules)

  it("always sends pagination=off regardless of filters", async () => {
    vi.mocked(listEvents).mockResolvedValue({ data: [], meta: { page: 1, total: 0 } })

    const { result } = renderHook(() => useEvents({ year: 2025, domain: "computer_science" }))

    await waitFor(() => expect(result.current.isLoading).toBe(false))

    expect(vi.mocked(listEvents).mock.calls[0][0]).toMatchObject({ pagination: "off" })
  })

  it("includes optional filters in params when set", async () => {
    vi.mocked(listEvents).mockResolvedValue({ data: [], meta: { page: 1, total: 0 } })

    const filters = {
      year: 2025,
      domain: "computer_science",
      tier: "A",
      country: "Brazil",
      firstDeadlineMonth: 6,
    }

    const { result } = renderHook(() => useEvents(filters))

    await waitFor(() => expect(result.current.isLoading).toBe(false))

    expect(listEvents).toHaveBeenCalledWith({
      year: 2025,
      domain: "computer_science",
      tier: "A",
      country: "Brazil",
      first_deadline_month: 6,
      pagination: "off",
    })
  })

  it("omits optional filters from params when undefined", async () => {
    vi.mocked(listEvents).mockResolvedValue({ data: [], meta: { page: 1, total: 0 } })

    const { result } = renderHook(() => useEvents({ year: 2026 }))

    await waitFor(() => expect(result.current.isLoading).toBe(false))

    const calledWith = vi.mocked(listEvents).mock.calls[0][0]
    expect(calledWith).not.toHaveProperty("domain")
    expect(calledWith).not.toHaveProperty("tier")
    expect(calledWith).not.toHaveProperty("country")
    expect(calledWith).not.toHaveProperty("first_deadline_month")
  })

  it("re-fetches when filters change", async () => {
    vi.mocked(listEvents).mockResolvedValue({ data: [], meta: { page: 1, total: 0 } })

    const { result, rerender } = renderHook(
      ({ filters }: { filters: { year: number } }) => useEvents(filters),
      { initialProps: { filters: { year: 2026 } } },
    )

    await waitFor(() => expect(result.current.isLoading).toBe(false))
    expect(listEvents).toHaveBeenCalledTimes(1)

    rerender({ filters: { year: 2025 } })

    await waitFor(() => expect(listEvents).toHaveBeenCalledTimes(2))
    expect(listEvents).toHaveBeenLastCalledWith({ year: 2025, pagination: "off" })
  })
})
