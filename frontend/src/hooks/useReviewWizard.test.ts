import { describe, it, expect, vi, beforeEach } from "vitest"
import { renderHook, act } from "@testing-library/react"

import { useReviewWizard, buildEventPatch } from "./useReviewWizard"
import { reviewEvent } from "@/lib/api/admin"
import type { EventListItem, EventTier } from "@/types/api"

vi.mock("@/lib/api/admin", () => ({
  reviewEvent: vi.fn(),
}))

// --- Fixtures ---

const makeEvent = (overrides: Partial<EventListItem> = {}): EventListItem => ({
  id: 7,
  name: "ICSE 2026",
  slug: "icse-2026",
  country: "Canada",
  city: "Toronto",
  latitude: 43.65,
  longitude: -79.38,
  start_date: "2026-05-10",
  end_date: "2026-05-15",
  website_url: "https://icse2026.example.com",
  domain: "software_engineering",
  status: "pending",
  tier: "A*",
  year: 2026,
  created_by: { id: 3, name: "Carol", email: "carol@example.com" },
  last_updated_by: { id: 3, name: "Carol", email: "carol@example.com" },
  deadlines: [],
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
  ...overrides,
})

const makeUser = () =>
  ({ id: 1, name: "Admin User", role: "admin" as const, email: "admin@example.com" })

// --- buildEventPatch ---

describe("buildEventPatch", () => {
  // Spec: "Partial edits: only changed fields sent in PATCH event body; unchanged fields omitted"
  it("returns undefined when no fields differ from the original event", () => {
    const event = makeEvent()
    const result = buildEventPatch(event, {
      name: event.name,
      country: event.country,
      city: event.city,
      startDate: event.start_date,
      endDate: event.end_date,
      websiteUrl: event.website_url,
      domain: event.domain,
      tier: event.tier ?? "",
      latitude: event.latitude,
      longitude: event.longitude,
    })
    expect(result).toBeUndefined()
  })

  // Spec: "only changed fields sent"
  it("returns only the changed name field when only name differs", () => {
    const event = makeEvent()
    const result = buildEventPatch(event, {
      name: "ICSE 2026 Updated",
      country: event.country,
      city: event.city,
      startDate: event.start_date,
      endDate: event.end_date,
      websiteUrl: event.website_url,
      domain: event.domain,
      tier: event.tier ?? "",
      latitude: event.latitude,
      longitude: event.longitude,
    })
    expect(result).toEqual({ name: "ICSE 2026 Updated" })
  })

  // Spec: "only changed fields sent"
  it("returns latitude and longitude when the map pin is moved", () => {
    const event = makeEvent()
    const result = buildEventPatch(event, {
      name: event.name,
      country: event.country,
      city: event.city,
      startDate: event.start_date,
      endDate: event.end_date,
      websiteUrl: event.website_url,
      domain: event.domain,
      tier: event.tier ?? "",
      latitude: 45.0,
      longitude: -75.0,
    })
    expect(result).toEqual({ latitude: 45.0, longitude: -75.0 })
  })

  // Spec: form fields map to snake_case API fields
  it("maps startDate→start_date, endDate→end_date, websiteUrl→website_url", () => {
    const event = makeEvent()
    const result = buildEventPatch(event, {
      name: event.name,
      country: event.country,
      city: event.city,
      startDate: "2026-06-01",
      endDate: "2026-06-05",
      websiteUrl: "https://updated.example.com",
      domain: event.domain,
      tier: event.tier ?? "",
      latitude: event.latitude,
      longitude: event.longitude,
    })
    expect(result).toEqual({
      start_date: "2026-06-01",
      end_date: "2026-06-05",
      website_url: "https://updated.example.com",
    })
  })

  // Spec: tier "" means unset — omit from patch
  it("omits tier from the patch when form tier is an empty string", () => {
    const event = makeEvent({ tier: undefined })
    const result = buildEventPatch(event, {
      name: event.name,
      country: event.country,
      city: event.city,
      startDate: event.start_date,
      endDate: event.end_date,
      websiteUrl: event.website_url,
      domain: event.domain,
      tier: "" as EventTier | "",
      latitude: event.latitude,
      longitude: event.longitude,
    })
    expect(result).toBeUndefined()
  })
})

// --- useReviewWizard ---

describe("useReviewWizard", () => {
  beforeEach(() => {
    vi.mocked(reviewEvent).mockReset()
  })

  // --- validateStep1 ---

  describe("validateStep1", () => {
    // Spec DoD: "Step 1: Next → validates required fields before advancing"
    it("returns true when all required fields are populated", () => {
      const { result } = renderHook(() => useReviewWizard(makeEvent(), makeUser()))
      let valid: boolean = false
      act(() => { valid = result.current.validateStep1() })
      expect(valid).toBe(true)
      expect(result.current.errors).toEqual({})
    })

    it("returns false and sets errors.name when name is empty", () => {
      const { result } = renderHook(() =>
        useReviewWizard(makeEvent({ name: "" }), makeUser())
      )
      let valid: boolean = true
      act(() => { valid = result.current.validateStep1() })
      expect(valid).toBe(false)
      expect(result.current.errors).toHaveProperty("name")
    })

    it("returns false and sets errors.country when country is empty", () => {
      const { result } = renderHook(() =>
        useReviewWizard(makeEvent({ country: "" }), makeUser())
      )
      let valid: boolean = true
      act(() => { valid = result.current.validateStep1() })
      expect(valid).toBe(false)
      expect(result.current.errors).toHaveProperty("country")
    })

    it("returns false and sets errors.city when city is empty", () => {
      const { result } = renderHook(() =>
        useReviewWizard(makeEvent({ city: "" }), makeUser())
      )
      let valid: boolean = true
      act(() => { valid = result.current.validateStep1() })
      expect(valid).toBe(false)
      expect(result.current.errors).toHaveProperty("city")
    })

    it("returns false and sets errors.startDate when start_date is empty", () => {
      const { result } = renderHook(() =>
        useReviewWizard(makeEvent({ start_date: "" }), makeUser())
      )
      let valid: boolean = true
      act(() => { valid = result.current.validateStep1() })
      expect(valid).toBe(false)
      expect(result.current.errors).toHaveProperty("startDate")
    })

    it("returns false and sets errors.endDate when end_date is empty", () => {
      const { result } = renderHook(() =>
        useReviewWizard(makeEvent({ end_date: "" }), makeUser())
      )
      let valid: boolean = true
      act(() => { valid = result.current.validateStep1() })
      expect(valid).toBe(false)
      expect(result.current.errors).toHaveProperty("endDate")
    })

    it("returns false and sets errors.websiteUrl when website_url is empty", () => {
      const { result } = renderHook(() =>
        useReviewWizard(makeEvent({ website_url: "" }), makeUser())
      )
      let valid: boolean = true
      act(() => { valid = result.current.validateStep1() })
      expect(valid).toBe(false)
      expect(result.current.errors).toHaveProperty("websiteUrl")
    })
  })

  // --- goToStep ---

  describe("goToStep", () => {
    it("advances from step 1 to step 2", () => {
      const { result } = renderHook(() => useReviewWizard(makeEvent(), makeUser()))
      expect(result.current.step).toBe(1)
      act(() => { result.current.goToStep(2) })
      expect(result.current.step).toBe(2)
    })

    it("returns from step 2 to step 1", () => {
      const { result } = renderHook(() => useReviewWizard(makeEvent(), makeUser()))
      act(() => { result.current.goToStep(2) })
      act(() => { result.current.goToStep(1) })
      expect(result.current.step).toBe(1)
    })

    it("advances from step 2 to step 3", () => {
      const { result } = renderHook(() => useReviewWizard(makeEvent(), makeUser()))
      act(() => { result.current.goToStep(3) })
      expect(result.current.step).toBe(3)
    })
  })

  // --- updateField ---

  describe("updateField", () => {
    // Spec: form fields editable, form state held locally
    it("updates a string field in formData", () => {
      const { result } = renderHook(() => useReviewWizard(makeEvent(), makeUser()))
      act(() => { result.current.updateField("name", "New Name") })
      expect(result.current.formData.name).toBe("New Name")
    })

    it("updates a numeric field (latitude) in formData", () => {
      const { result } = renderHook(() => useReviewWizard(makeEvent(), makeUser()))
      act(() => { result.current.updateField("latitude", 51.5) })
      expect(result.current.formData.latitude).toBe(51.5)
    })

    it("clears the error key for the updated field", () => {
      const { result } = renderHook(() =>
        useReviewWizard(makeEvent({ name: "" }), makeUser())
      )
      act(() => { result.current.validateStep1() })
      expect(result.current.errors).toHaveProperty("name")

      act(() => { result.current.updateField("name", "ICSE 2026") })
      expect(result.current.errors).not.toHaveProperty("name")
    })
  })

  // --- approve ---

  describe("approve", () => {
    // Spec DoD: "Approve success → ReviewSuccess screen with Manage Deadlines + Back to Dashboard"
    it("calls reviewEvent with action:approve, note, and only changed fields", async () => {
      const event = makeEvent()
      vi.mocked(reviewEvent).mockResolvedValue(makeEvent({ status: "approved" }))

      const { result } = renderHook(() => useReviewWizard(event, makeUser()))

      act(() => { result.current.updateField("city", "Vancouver") })

      await act(async () => { await result.current.approve("Looks good") })

      expect(reviewEvent).toHaveBeenCalledWith(event.id, {
        action: "approve",
        reason: "Looks good",
        event: { city: "Vancouver" },
      })
    })

    it("sets successState with action:approve and the updated event on success", async () => {
      const updated = makeEvent({ status: "approved" })
      vi.mocked(reviewEvent).mockResolvedValue(updated)

      const { result } = renderHook(() => useReviewWizard(makeEvent(), makeUser()))

      await act(async () => { await result.current.approve("") })

      expect(result.current.successState).toEqual({
        event: updated,
        action: "approve",
        reason: undefined,
      })
    })

    it("sets bannerError and keeps step at 2 on failure", async () => {
      vi.mocked(reviewEvent).mockRejectedValue(new Error("server error"))

      const { result } = renderHook(() => useReviewWizard(makeEvent(), makeUser()))
      act(() => { result.current.goToStep(2) })

      await act(async () => { await result.current.approve("") })

      expect(result.current.bannerError).not.toBeNull()
      expect(result.current.step).toBe(2)
      expect(result.current.successState).toBeNull()
    })
  })

  // --- reject ---

  describe("reject", () => {
    // Spec DoD: "Reject success → ReviewSuccess screen with reason shown + Back to Dashboard only"
    it("calls reviewEvent with action:reject and the required reason", async () => {
      const event = makeEvent()
      vi.mocked(reviewEvent).mockResolvedValue(makeEvent({ status: "rejected" }))

      const { result } = renderHook(() => useReviewWizard(event, makeUser()))

      await act(async () => { await result.current.reject("Missing abstract deadline") })

      expect(reviewEvent).toHaveBeenCalledWith(event.id, {
        action: "reject",
        reason: "Missing abstract deadline",
      })
    })

    it("sets successState with action:reject, reason, and updated event on success", async () => {
      const updated = makeEvent({ status: "rejected" })
      vi.mocked(reviewEvent).mockResolvedValue(updated)

      const { result } = renderHook(() => useReviewWizard(makeEvent(), makeUser()))

      await act(async () => { await result.current.reject("Duplicate entry") })

      expect(result.current.successState).toEqual({
        event: updated,
        action: "reject",
        reason: "Duplicate entry",
      })
    })

    it("sets bannerError and keeps step at 2 on failure", async () => {
      vi.mocked(reviewEvent).mockRejectedValue(new Error("network error"))

      const { result } = renderHook(() => useReviewWizard(makeEvent(), makeUser()))
      act(() => { result.current.goToStep(2) })

      await act(async () => { await result.current.reject("Bad event") })

      expect(result.current.bannerError).not.toBeNull()
      expect(result.current.step).toBe(2)
      expect(result.current.successState).toBeNull()
    })
  })

  // --- clearSuccess ---

  describe("clearSuccess", () => {
    // Spec: "Manage Deadlines advances to Step 3" — clearSuccess lets wizard resume
    it("resets successState to null", async () => {
      vi.mocked(reviewEvent).mockResolvedValue(makeEvent({ status: "approved" }))

      const { result } = renderHook(() => useReviewWizard(makeEvent(), makeUser()))

      await act(async () => { await result.current.approve("") })
      expect(result.current.successState).not.toBeNull()

      act(() => { result.current.clearSuccess() })
      expect(result.current.successState).toBeNull()
    })
  })
})
