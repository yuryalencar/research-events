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

describe("useEvents", () => {
  beforeEach(() => {
    vi.mocked(listEvents).mockReset()
    vi.mocked(handleApiError).mockReset()
  })

  it("starts in a loading state with no events", () => {
    vi.mocked(listEvents).mockReturnValue(new Promise(() => {}))

    const { result } = renderHook(() => useEvents())

    expect(result.current.isLoading).toBe(true)
    expect(result.current.events).toEqual([])
  })

  it("calls listEvents with pagination off and stores the returned events", async () => {
    vi.mocked(listEvents).mockResolvedValue({ data: [event], meta: { page: 1, total: 1 } })

    const { result } = renderHook(() => useEvents())

    await waitFor(() => expect(result.current.isLoading).toBe(false))

    expect(listEvents).toHaveBeenCalledWith({ pagination: "off" })
    expect(result.current.events).toEqual([event])
  })

  it("calls handleApiError and leaves events empty when the fetch fails", async () => {
    vi.mocked(listEvents).mockRejectedValue(new Error("network error"))

    const { result } = renderHook(() => useEvents())

    await waitFor(() => expect(result.current.isLoading).toBe(false))

    expect(handleApiError).toHaveBeenCalled()
    expect(result.current.events).toEqual([])
  })

  it("returns an empty events array when data is empty", async () => {
    vi.mocked(listEvents).mockResolvedValue({ data: [], meta: { page: 1, total: 0 } })

    const { result } = renderHook(() => useEvents())

    await waitFor(() => expect(result.current.isLoading).toBe(false))

    expect(result.current.events).toEqual([])
    expect(handleApiError).not.toHaveBeenCalled()
  })
})
