import { describe, it, expect, vi, beforeEach } from "vitest"
import { renderHook, act } from "@testing-library/react"

import { useSubmitWizard } from "./useSubmitWizard"
import { submitEvent } from "@/lib/api/events"
import { ApiError } from "@/lib/api/client"
import type { SubmitEventResult } from "@/types/api"

vi.mock("@/lib/api/events", () => ({
  submitEvent: vi.fn(),
}))

// --- Fixtures ---

// Complete valid Step 2 data; individual tests override the field under test.
const validStep2Data = {
  submitterName: "Alice",
  submitterEmail: "alice@example.com",
  fullName: "International Conference on Software Engineering 2026",
  slug: "ICSE2026",
  domain: "computer_science",
  tier: "" as const,
  startDate: "2026-05-01",
  endDate: "2026-05-05",
  websiteUrl: "https://icse2026.org",
  country: "Brazil",
  city: "São Paulo",
  latitude: -23.55 as number | null,
  longitude: -46.63 as number | null,
}

const mockSubmitResult: SubmitEventResult = {
  id: 1,
  name: "International Conference on Software Engineering 2026",
  slug: "ICSE2026",
  country: "Brazil",
  city: "São Paulo",
  latitude: -23.55,
  longitude: -46.63,
  start_date: "2026-05-01",
  end_date: "2026-05-05",
  website_url: "https://icse2026.org",
  domain: "computer_science",
  tier: "unranked",
  status: "pending",
  created_by: { id: 1, name: "Alice", email: "alice@example.com" },
  deadlines: [],
  created_at: "2026-01-01T00:00:00Z",
}

// Populate all Step 2 fields in a single act() block.
function fillStep2(result: ReturnType<typeof useSubmitWizard>): void {
  Object.entries(validStep2Data).forEach(([key, value]) =>
    result.updateField(key as Parameters<typeof result.updateField>[0], value)
  )
}

// --- Tests ---

describe("useSubmitWizard", () => {
  beforeEach(() => {
    vi.mocked(submitEvent).mockReset()
  })

  // --- validateStep2 ---
  // Spec: "Client-side validation (on 'Continue to Step 3')"

  it("validateStep2 returns false and sets error when submitter name is empty", () => {
    const { result } = renderHook(() => useSubmitWizard())

    act(() => {
      fillStep2(result.current)
      result.current.updateField("submitterName", "")
    })

    expect(result.current.validateStep2()).toBe(false)
    expect(result.current.errors.submitterName).toBeTruthy()
  })

  it("validateStep2 returns false and sets error when submitter email is empty", () => {
    const { result } = renderHook(() => useSubmitWizard())

    act(() => {
      fillStep2(result.current)
      result.current.updateField("submitterEmail", "")
    })

    expect(result.current.validateStep2()).toBe(false)
    expect(result.current.errors.submitterEmail).toBeTruthy()
  })

  it("validateStep2 returns false and sets error when slug contains invalid characters", () => {
    // Spec: "Slug matches ^[A-Za-z0-9_-]+$"
    const { result } = renderHook(() => useSubmitWizard())

    act(() => {
      fillStep2(result.current)
      result.current.updateField("slug", "ICSE 2026!")
    })

    expect(result.current.validateStep2()).toBe(false)
    expect(result.current.errors.slug).toBeTruthy()
  })

  it("validateStep2 returns false and sets error when end date is before start date", () => {
    // Spec: "End date ≥ start date"
    const { result } = renderHook(() => useSubmitWizard())

    act(() => {
      fillStep2(result.current)
      result.current.updateField("endDate", "2026-04-01")
    })

    expect(result.current.validateStep2()).toBe(false)
    expect(result.current.errors.endDate).toBeTruthy()
  })

  it("validateStep2 returns false and sets error when latitude is null (no pin dropped)", () => {
    // Spec: "Lat/lng must be set (pin dropped on map)"
    const { result } = renderHook(() => useSubmitWizard())

    act(() => {
      fillStep2(result.current)
      result.current.updateField("latitude", null)
    })

    expect(result.current.validateStep2()).toBe(false)
    expect(result.current.errors.location).toBeTruthy()
  })

  it("validateStep2 returns true and clears errors for a fully valid payload", () => {
    const { result } = renderHook(() => useSubmitWizard())

    act(() => fillStep2(result.current))

    expect(result.current.validateStep2()).toBe(true)
    expect(Object.keys(result.current.errors)).toHaveLength(0)
  })

  // --- submit ---
  // Spec: "POST /api/v1/events/submit called with correct payload shape"

  it("submit calls submitEvent with the correct payload shape", async () => {
    vi.mocked(submitEvent).mockResolvedValue(mockSubmitResult)
    const { result } = renderHook(() => useSubmitWizard())

    act(() => fillStep2(result.current))

    await act(async () => {
      await result.current.submit(true) // skipDeadlines = true → deadlines: []
    })

    expect(submitEvent).toHaveBeenCalledWith({
      name: validStep2Data.fullName,
      slug: validStep2Data.slug,
      country: validStep2Data.country,
      city: validStep2Data.city,
      latitude: validStep2Data.latitude,
      longitude: validStep2Data.longitude,
      start_date: validStep2Data.startDate,
      end_date: validStep2Data.endDate,
      website_url: validStep2Data.websiteUrl,
      domain: validStep2Data.domain,
      submitter: {
        name: validStep2Data.submitterName,
        email: validStep2Data.submitterEmail,
      },
      deadlines: [],
    })
  })

  it("submit sets submittedEvent on 201 success", async () => {
    // Spec: "201 → success screen with event details"
    vi.mocked(submitEvent).mockResolvedValue(mockSubmitResult)
    const { result } = renderHook(() => useSubmitWizard())

    act(() => fillStep2(result.current))

    await act(async () => {
      await result.current.submit(true)
    })

    expect(result.current.submittedEvent).toEqual(mockSubmitResult)
  })

  it("submit sets slug error and returns to step 2 on 409", async () => {
    // Spec: "409 → Return to Step 2; slug field shows 'This slug is already taken'"
    vi.mocked(submitEvent).mockRejectedValue(
      new ApiError("EVENT_ALREADY_SUBMITTED", 409, "an event with this slug has already been submitted")
    )
    const { result } = renderHook(() => useSubmitWizard())

    act(() => {
      fillStep2(result.current)
      result.current.goToStep(3)
    })

    await act(async () => {
      await result.current.submit(true)
    })

    expect(result.current.step).toBe(2)
    expect(result.current.errors.slug).toBeTruthy()
  })

  it("submit sets bannerError and returns to step 2 on 400", async () => {
    // Spec: "400 → Return to Step 2; per-field errors shown"
    vi.mocked(submitEvent).mockRejectedValue(
      new ApiError("VALIDATION_ERROR", 400, "website_url must be a valid URL")
    )
    const { result } = renderHook(() => useSubmitWizard())

    act(() => {
      fillStep2(result.current)
      result.current.goToStep(3)
    })

    await act(async () => {
      await result.current.submit(true)
    })

    expect(result.current.step).toBe(2)
    expect(result.current.bannerError).toBeTruthy()
  })

  it("submit sets bannerError and stays on step 3 on 429", async () => {
    // Spec: "429 → banner on Step 3"
    vi.mocked(submitEvent).mockRejectedValue(
      new ApiError("RATE_LIMIT_EXCEEDED", 429, "too many submissions")
    )
    const { result } = renderHook(() => useSubmitWizard())

    act(() => {
      fillStep2(result.current)
      result.current.goToStep(3)
    })

    await act(async () => {
      await result.current.submit(true)
    })

    expect(result.current.step).toBe(3)
    expect(result.current.bannerError).toBeTruthy()
  })

  it("submit sets bannerError and stays on step 3 on network error", async () => {
    // Spec: "Network error → banner on Step 3"
    vi.mocked(submitEvent).mockRejectedValue(
      new ApiError("NETWORK_ERROR", 0, "could not connect to the server")
    )
    const { result } = renderHook(() => useSubmitWizard())

    act(() => {
      fillStep2(result.current)
      result.current.goToStep(3)
    })

    await act(async () => {
      await result.current.submit(true)
    })

    expect(result.current.step).toBe(3)
    expect(result.current.bannerError).toBeTruthy()
  })

  // --- Deadline management ---
  // Spec: "Step 3 allows adding/removing deadline rows"

  it("addDeadline appends a new empty deadline row", () => {
    const { result } = renderHook(() => useSubmitWizard())

    act(() => result.current.addDeadline())

    expect(result.current.formData.deadlines).toHaveLength(1)
    expect(result.current.formData.deadlines[0].type).toBe("")
    expect(result.current.formData.deadlines[0].description).toBe("")
    expect(result.current.formData.deadlines[0].date).toBe("")
  })

  it("removeDeadline removes the correct row by id", () => {
    const { result } = renderHook(() => useSubmitWizard())

    act(() => {
      result.current.addDeadline()
      result.current.addDeadline()
    })

    const idToRemove = result.current.formData.deadlines[0].id

    act(() => result.current.removeDeadline(idToRemove))

    expect(result.current.formData.deadlines).toHaveLength(1)
    expect(result.current.formData.deadlines[0].id).not.toBe(idToRemove)
  })

  it("updateDeadline updates the correct field on the correct row", () => {
    const { result } = renderHook(() => useSubmitWizard())

    act(() => {
      result.current.addDeadline()
      result.current.addDeadline()
    })

    const targetId = result.current.formData.deadlines[1].id

    act(() => result.current.updateDeadline(targetId, "description", "Research track"))

    expect(result.current.formData.deadlines[1].description).toBe("Research track")
    expect(result.current.formData.deadlines[0].description).toBe("") // first row unchanged
  })
})
