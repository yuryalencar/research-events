import { describe, it, expect, vi, beforeEach } from "vitest"
import { renderHook, act } from "@testing-library/react"

import { useDeadlineManage } from "./useDeadlineManage"
import { cancelDeadline, supersedeDeadline, addDeadlines } from "@/lib/api/events"
import { ApiError } from "@/lib/api/client"
import type { DeadlineResponse, EventListItem, DeadlineType } from "@/types/api"

vi.mock("@/lib/api/events", () => ({
  cancelDeadline: vi.fn(),
  supersedeDeadline: vi.fn(),
  addDeadlines: vi.fn(),
  // keep other exports from the module intact
  listEvents: vi.fn(),
  submitEvent: vi.fn(),
}))

// --- Fixtures ---

function makeDeadline(overrides: Partial<DeadlineResponse> = {}): DeadlineResponse {
  return {
    id: 1,
    type: "abstract",
    description: "Research track",
    date: "2026-09-01",
    time: null,
    timezone: null,
    is_optional: false,
    is_active: true,
    superseded_by_id: null,
    ...overrides,
  }
}

function makeEvent(deadlines: DeadlineResponse[] = []): EventListItem {
  return {
    id: 1,
    name: "ICSE 2026",
    slug: "ICSE2026",
    country: "Brazil",
    city: "São Paulo",
    latitude: -23.55,
    longitude: -46.63,
    start_date: "2026-05-01",
    end_date: "2026-11-30",
    website_url: "https://icse2026.org",
    domain: "computer_science",
    status: "approved",
    tier: "A*",
    year: 2026,
    created_by: { id: 1, name: "Alice", email: "alice@example.com" },
    last_updated_by: { id: 1, name: "Alice", email: "alice@example.com" },
    deadlines,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  }
}

// --- Tests ---

describe("useDeadlineManage", () => {
  // --- startSupersede ---
  // Spec: "Pencil clicked → Supersede-editing: date/time/timezone editable; type/desc/is_optional read-only"

  describe("startSupersede", () => {
    it("sets the deadline mode to superseding", () => {
      const event = makeEvent([makeDeadline({ id: 10 })])
      const { result } = renderHook(() => useDeadlineManage(event))

      act(() => result.current.startSupersede(10))

      expect(result.current.deadlineStates[10].mode).toBe("superseding")
    })

    it("copies original date/time/timezone into editValues", () => {
      const deadline = makeDeadline({ id: 10, date: "2026-09-01", time: "14:00", timezone: "UTC" })
      const event = makeEvent([deadline])
      const { result } = renderHook(() => useDeadlineManage(event))

      act(() => result.current.startSupersede(10))

      expect(result.current.deadlineStates[10].editValues.date).toBe("2026-09-01")
      expect(result.current.deadlineStates[10].editValues.time).toBe("14:00")
      expect(result.current.deadlineStates[10].editValues.timezone).toBe("UTC")
    })

    it("converts null time and timezone to empty strings in editValues", () => {
      // null values in the API response become "" in the form so inputs are controlled
      const deadline = makeDeadline({ id: 10, date: "2026-09-01", time: null, timezone: null })
      const event = makeEvent([deadline])
      const { result } = renderHook(() => useDeadlineManage(event))

      act(() => result.current.startSupersede(10))

      expect(result.current.deadlineStates[10].editValues.time).toBe("")
      expect(result.current.deadlineStates[10].editValues.timezone).toBe("")
    })

    it("reverts a pending-cancel and enters superseding (pencil on pending-cancel card)", () => {
      // Spec: "Pencil on Pending-cancel → revert cancel + enter Supersede-editing"
      const deadline = makeDeadline({ id: 10 })
      const event = makeEvent([deadline])
      const { result } = renderHook(() => useDeadlineManage(event))

      act(() => {
        result.current.cancelDeadlineLocal(10)
        result.current.startSupersede(10)
      })

      expect(result.current.deadlineStates[10].mode).toBe("superseding")
    })

    it("does not affect other deadlines", () => {
      const d1 = makeDeadline({ id: 10 })
      const d2 = makeDeadline({ id: 20 })
      const event = makeEvent([d1, d2])
      const { result } = renderHook(() => useDeadlineManage(event))

      act(() => result.current.startSupersede(10))

      expect(result.current.deadlineStates[20].mode).toBe("default")
    })
  })

  // --- revertSupersede ---
  // Spec: "'Revert' on Supersede-editing restores original values and exits edit mode"

  describe("revertSupersede", () => {
    it("returns the deadline to default mode", () => {
      const event = makeEvent([makeDeadline({ id: 10 })])
      const { result } = renderHook(() => useDeadlineManage(event))

      act(() => {
        result.current.startSupersede(10)
        result.current.revertSupersede(10)
      })

      expect(result.current.deadlineStates[10].mode).toBe("default")
    })
  })

  // --- cancelDeadlineLocal ---
  // Spec: "X clicked → Pending-cancel: strikethrough card + 'Revert' button"

  describe("cancelDeadlineLocal", () => {
    it("sets the deadline mode to pendingCancel", () => {
      const event = makeEvent([makeDeadline({ id: 10 })])
      const { result } = renderHook(() => useDeadlineManage(event))

      act(() => result.current.cancelDeadlineLocal(10))

      expect(result.current.deadlineStates[10].mode).toBe("pendingCancel")
    })

    it("exits superseding and enters pendingCancel (X on supersede-editing card)", () => {
      // Spec: "X on Supersede-editing → exit edit + enter Pending-cancel"
      const event = makeEvent([makeDeadline({ id: 10 })])
      const { result } = renderHook(() => useDeadlineManage(event))

      act(() => {
        result.current.startSupersede(10)
        result.current.cancelDeadlineLocal(10)
      })

      expect(result.current.deadlineStates[10].mode).toBe("pendingCancel")
    })

    it("does not affect other deadlines", () => {
      const d1 = makeDeadline({ id: 10 })
      const d2 = makeDeadline({ id: 20 })
      const event = makeEvent([d1, d2])
      const { result } = renderHook(() => useDeadlineManage(event))

      act(() => result.current.cancelDeadlineLocal(10))

      expect(result.current.deadlineStates[20].mode).toBe("default")
    })
  })

  // --- revertCancel ---
  // Spec: "Revert on Pending-cancel → Default state"

  describe("revertCancel", () => {
    it("returns the deadline to default mode", () => {
      const event = makeEvent([makeDeadline({ id: 10 })])
      const { result } = renderHook(() => useDeadlineManage(event))

      act(() => {
        result.current.cancelDeadlineLocal(10)
        result.current.revertCancel(10)
      })

      expect(result.current.deadlineStates[10].mode).toBe("default")
    })
  })

  // --- updateSupersede ---
  // Spec: "Editable fields (same inputClass as Step3): date | time | timezone"

  describe("updateSupersede", () => {
    it("updates the date field for a superseding deadline", () => {
      const deadline = makeDeadline({ id: 10, date: "2026-09-01" })
      const event = makeEvent([deadline])
      const { result } = renderHook(() => useDeadlineManage(event))

      act(() => {
        result.current.startSupersede(10)
        result.current.updateSupersede(10, "date", "2026-10-15")
      })

      expect(result.current.deadlineStates[10].editValues.date).toBe("2026-10-15")
    })

    it("updates the time field for a superseding deadline", () => {
      const deadline = makeDeadline({ id: 10, time: null })
      const event = makeEvent([deadline])
      const { result } = renderHook(() => useDeadlineManage(event))

      act(() => {
        result.current.startSupersede(10)
        result.current.updateSupersede(10, "time", "14:00")
      })

      expect(result.current.deadlineStates[10].editValues.time).toBe("14:00")
    })

    it("does not mutate another deadline's edit state", () => {
      const d1 = makeDeadline({ id: 10, date: "2026-09-01" })
      const d2 = makeDeadline({ id: 20, date: "2026-10-01" })
      const event = makeEvent([d1, d2])
      const { result } = renderHook(() => useDeadlineManage(event))

      act(() => {
        result.current.startSupersede(10)
        result.current.startSupersede(20)
        result.current.updateSupersede(10, "date", "2026-12-01")
      })

      expect(result.current.deadlineStates[20].editValues.date).toBe("2026-10-01")
    })
  })

  // --- addNewDeadline ---
  // Spec: "'+ Add deadline' appends a new card"

  describe("addNewDeadline", () => {
    it("appends one new empty deadline card", () => {
      const event = makeEvent([])
      const { result } = renderHook(() => useDeadlineManage(event))

      act(() => result.current.addNewDeadline())

      expect(result.current.newDeadlines).toHaveLength(1)
      expect(result.current.newDeadlines[0].type).toBe("")
      expect(result.current.newDeadlines[0].description).toBe("")
      expect(result.current.newDeadlines[0].date).toBe("")
      expect(result.current.newDeadlines[0].time).toBe("")
      expect(result.current.newDeadlines[0].timezone).toBe("")
      expect(result.current.newDeadlines[0].isOptional).toBe(false)
    })

    it("assigns a unique localId to each new deadline card", () => {
      const event = makeEvent([])
      const { result } = renderHook(() => useDeadlineManage(event))

      act(() => {
        result.current.addNewDeadline()
        result.current.addNewDeadline()
      })

      expect(result.current.newDeadlines[0].localId).not.toBe(result.current.newDeadlines[1].localId)
    })

    it("appends multiple cards in order", () => {
      const event = makeEvent([])
      const { result } = renderHook(() => useDeadlineManage(event))

      act(() => {
        result.current.addNewDeadline()
        result.current.addNewDeadline()
        result.current.addNewDeadline()
      })

      expect(result.current.newDeadlines).toHaveLength(3)
    })
  })

  // --- removeNewDeadline ---
  // Spec: "Remove on new card discards it; not counted as a change"

  describe("removeNewDeadline", () => {
    it("removes the correct card by localId", () => {
      const event = makeEvent([])
      const { result } = renderHook(() => useDeadlineManage(event))

      act(() => {
        result.current.addNewDeadline()
        result.current.addNewDeadline()
      })

      const idToRemove = result.current.newDeadlines[0].localId

      act(() => result.current.removeNewDeadline(idToRemove))

      expect(result.current.newDeadlines).toHaveLength(1)
      expect(result.current.newDeadlines[0].localId).not.toBe(idToRemove)
    })
  })

  // --- hasChanges ---
  // Spec: "Submit disabled when nothing changed"; "Changed detection" section

  describe("hasChanges", () => {
    it("returns false when no changes have been made", () => {
      const event = makeEvent([makeDeadline({ id: 10 })])
      const { result } = renderHook(() => useDeadlineManage(event))

      expect(result.current.hasChanges).toBe(false)
    })

    it("returns true when a deadline is pending cancellation", () => {
      const event = makeEvent([makeDeadline({ id: 10 })])
      const { result } = renderHook(() => useDeadlineManage(event))

      act(() => result.current.cancelDeadlineLocal(10))

      expect(result.current.hasChanges).toBe(true)
    })

    it("returns false when a pending cancel is reverted", () => {
      const event = makeEvent([makeDeadline({ id: 10 })])
      const { result } = renderHook(() => useDeadlineManage(event))

      act(() => {
        result.current.cancelDeadlineLocal(10)
        result.current.revertCancel(10)
      })

      expect(result.current.hasChanges).toBe(false)
    })

    it("returns false when entering superseding mode without changing values", () => {
      // Spec: "Opening edit mode and saving without changing anything is not counted as a change"
      const deadline = makeDeadline({ id: 10, date: "2026-09-01", time: null, timezone: null })
      const event = makeEvent([deadline])
      const { result } = renderHook(() => useDeadlineManage(event))

      act(() => result.current.startSupersede(10))

      expect(result.current.hasChanges).toBe(false)
    })

    it("returns true when a supersede-edit changes the date", () => {
      const deadline = makeDeadline({ id: 10, date: "2026-09-01", time: null, timezone: null })
      const event = makeEvent([deadline])
      const { result } = renderHook(() => useDeadlineManage(event))

      act(() => {
        result.current.startSupersede(10)
        result.current.updateSupersede(10, "date", "2026-10-15")
      })

      expect(result.current.hasChanges).toBe(true)
    })

    it("returns true when a supersede-edit changes the time", () => {
      const deadline = makeDeadline({ id: 10, date: "2026-09-01", time: "09:00", timezone: null })
      const event = makeEvent([deadline])
      const { result } = renderHook(() => useDeadlineManage(event))

      act(() => {
        result.current.startSupersede(10)
        result.current.updateSupersede(10, "time", "14:00")
      })

      expect(result.current.hasChanges).toBe(true)
    })

    it("returns false when a supersede-edit is reverted with revertSupersede", () => {
      const deadline = makeDeadline({ id: 10, date: "2026-09-01", time: null, timezone: null })
      const event = makeEvent([deadline])
      const { result } = renderHook(() => useDeadlineManage(event))

      act(() => {
        result.current.startSupersede(10)
        result.current.updateSupersede(10, "date", "2026-10-15")
        result.current.revertSupersede(10)
      })

      expect(result.current.hasChanges).toBe(false)
    })

    it("returns false when edit values are typed back to original values", () => {
      // Spec: "Supersede with same values as original → not counted as a change"
      const deadline = makeDeadline({ id: 10, date: "2026-09-01", time: null, timezone: null })
      const event = makeEvent([deadline])
      const { result } = renderHook(() => useDeadlineManage(event))

      act(() => {
        result.current.startSupersede(10)
        result.current.updateSupersede(10, "date", "2026-10-15")
        result.current.updateSupersede(10, "date", "2026-09-01") // typed back to original
      })

      expect(result.current.hasChanges).toBe(false)
    })

    it("returns true when a new deadline card is added", () => {
      const event = makeEvent([])
      const { result } = renderHook(() => useDeadlineManage(event))

      act(() => result.current.addNewDeadline())

      expect(result.current.hasChanges).toBe(true)
    })

    it("returns false when all added new deadline cards are removed", () => {
      // Spec: "Add then immediately Remove a new card → not counted as a change"
      const event = makeEvent([])
      const { result } = renderHook(() => useDeadlineManage(event))

      act(() => result.current.addNewDeadline())
      const localId = result.current.newDeadlines[0].localId
      act(() => result.current.removeNewDeadline(localId))

      expect(result.current.hasChanges).toBe(false)
    })

    it("returns true when multiple change types are combined", () => {
      const d1 = makeDeadline({ id: 10 })
      const d2 = makeDeadline({ id: 20 })
      const event = makeEvent([d1, d2])
      const { result } = renderHook(() => useDeadlineManage(event))

      act(() => {
        result.current.cancelDeadlineLocal(10)
        result.current.addNewDeadline()
      })

      expect(result.current.hasChanges).toBe(true)
    })
  })

  // --- validate ---
  // Spec: "Validation rules" section; "Inline errors shown only after a submit attempt; not on blur"

  describe("validate", () => {
    // Helper to build a fully valid new deadline in one act().
    function addValidNewDeadline(
      result: ReturnType<typeof renderHook<ReturnType<typeof useDeadlineManage>, EventListItem>>["result"],
      overrides: Partial<Record<"type" | "description" | "date" | "time", string>> = {},
    ): string {
      act(() => result.current.addNewDeadline())
      const localId = result.current.newDeadlines.at(-1)!.localId
      act(() => {
        result.current.updateNewDeadline(localId, "type", (overrides.type ?? "abstract") as DeadlineType)
        result.current.updateNewDeadline(localId, "description", overrides.description ?? "Research track")
        result.current.updateNewDeadline(localId, "date", overrides.date ?? "2026-09-01")
        if (overrides.time !== undefined) {
          result.current.updateNewDeadline(localId, "time", overrides.time)
        }
      })
      return localId
    }

    // Helper: set valid contributor fields.
    function fillContributor(
      result: ReturnType<typeof renderHook<ReturnType<typeof useDeadlineManage>, EventListItem>>["result"],
    ): void {
      act(() => {
        result.current.updateContributor("name", "Alice")
        result.current.updateContributor("email", "alice@example.com")
      })
    }

    // --- contributor validation ---
    // Spec: "Contributor: name: required, non-empty; email: required, valid email format"

    it("returns false and sets contributor.name error when name is empty", () => {
      const { result } = renderHook(() => useDeadlineManage(makeEvent([])))

      act(() => result.current.updateContributor("email", "alice@example.com"))

      expect(result.current.validate()).toBe(false)
      expect(result.current.errors["contributor.name"]).toBeTruthy()
    })

    it("returns false and sets contributor.email error when email is empty", () => {
      const { result } = renderHook(() => useDeadlineManage(makeEvent([])))

      act(() => result.current.updateContributor("name", "Alice"))

      expect(result.current.validate()).toBe(false)
      expect(result.current.errors["contributor.email"]).toBeTruthy()
    })

    it("returns false and sets contributor.email error when email format is invalid", () => {
      const { result } = renderHook(() => useDeadlineManage(makeEvent([])))

      act(() => {
        result.current.updateContributor("name", "Alice")
        result.current.updateContributor("email", "not-an-email")
      })

      expect(result.current.validate()).toBe(false)
      expect(result.current.errors["contributor.email"]).toBeTruthy()
    })

    it("returns true and has no errors when contributor is valid and there are no deadline changes", () => {
      const { result } = renderHook(() => useDeadlineManage(makeEvent([])))

      fillContributor(result)

      expect(result.current.validate()).toBe(true)
      expect(Object.keys(result.current.errors)).toHaveLength(0)
    })

    // --- new deadline card validation ---
    // Spec: "New deadline cards: type required; description required; date required + ≤ end_date; time HH:MM if provided"

    it("returns false and sets type error when new deadline type is empty", () => {
      const { result } = renderHook(() => useDeadlineManage(makeEvent([])))
      fillContributor(result)
      const localId = addValidNewDeadline(result, { type: "" })

      expect(result.current.validate()).toBe(false)
      expect(result.current.errors[`new.${localId}.type`]).toBeTruthy()
    })

    it("returns false and sets description error when new deadline description is empty", () => {
      const { result } = renderHook(() => useDeadlineManage(makeEvent([])))
      fillContributor(result)
      const localId = addValidNewDeadline(result, { description: "" })

      expect(result.current.validate()).toBe(false)
      expect(result.current.errors[`new.${localId}.description`]).toBeTruthy()
    })

    it("returns false and sets date error when new deadline date is empty", () => {
      const { result } = renderHook(() => useDeadlineManage(makeEvent([])))
      fillContributor(result)
      const localId = addValidNewDeadline(result, { date: "" })

      expect(result.current.validate()).toBe(false)
      expect(result.current.errors[`new.${localId}.date`]).toBeTruthy()
    })

    it("returns false and sets date error when new deadline date is after event end_date", () => {
      // Spec: "date: required; must not be after event.end_date" (event end_date: 2026-11-30)
      const { result } = renderHook(() => useDeadlineManage(makeEvent([])))
      fillContributor(result)
      const localId = addValidNewDeadline(result, { date: "2026-12-01" })

      expect(result.current.validate()).toBe(false)
      expect(result.current.errors[`new.${localId}.date`]).toBeTruthy()
    })

    it("returns false and sets time error when new deadline time is not in HH:MM format", () => {
      // Spec: "time: if provided, must match HH:MM (00:00–23:59)"
      const { result } = renderHook(() => useDeadlineManage(makeEvent([])))
      fillContributor(result)
      const localId = addValidNewDeadline(result, { time: "25:00" })

      expect(result.current.validate()).toBe(false)
      expect(result.current.errors[`new.${localId}.time`]).toBeTruthy()
    })

    it("returns true for a valid new deadline with a well-formed time", () => {
      const { result } = renderHook(() => useDeadlineManage(makeEvent([])))
      fillContributor(result)
      addValidNewDeadline(result, { time: "14:30" })

      expect(result.current.validate()).toBe(true)
    })

    it("returns true for a valid new deadline when time is omitted (time is optional)", () => {
      const { result } = renderHook(() => useDeadlineManage(makeEvent([])))
      fillContributor(result)
      addValidNewDeadline(result) // no time override → stays ""

      expect(result.current.validate()).toBe(true)
    })

    // --- supersede card validation ---
    // Spec: "Supersede-editing cards: date required + ≤ end_date; time HH:MM if provided"

    it("returns false and sets supersede date error when supersede date is empty", () => {
      const deadline = makeDeadline({ id: 10, date: "2026-09-01" })
      const { result } = renderHook(() => useDeadlineManage(makeEvent([deadline])))
      fillContributor(result)

      act(() => {
        result.current.startSupersede(10)
        result.current.updateSupersede(10, "date", "")
      })

      expect(result.current.validate()).toBe(false)
      expect(result.current.errors["supersede.10.date"]).toBeTruthy()
    })

    it("returns false and sets supersede date error when supersede date is after event end_date", () => {
      // event end_date: 2026-11-30
      const deadline = makeDeadline({ id: 10, date: "2026-09-01" })
      const { result } = renderHook(() => useDeadlineManage(makeEvent([deadline])))
      fillContributor(result)

      act(() => {
        result.current.startSupersede(10)
        result.current.updateSupersede(10, "date", "2026-12-01")
      })

      expect(result.current.validate()).toBe(false)
      expect(result.current.errors["supersede.10.date"]).toBeTruthy()
    })

    it("returns false and sets supersede time error when time is not in HH:MM format", () => {
      const deadline = makeDeadline({ id: 10, date: "2026-09-01" })
      const { result } = renderHook(() => useDeadlineManage(makeEvent([deadline])))
      fillContributor(result)

      act(() => {
        result.current.startSupersede(10)
        result.current.updateSupersede(10, "time", "9:5")
      })

      expect(result.current.validate()).toBe(false)
      expect(result.current.errors["supersede.10.time"]).toBeTruthy()
    })

    it("returns true for a valid supersede edit with a changed date", () => {
      const deadline = makeDeadline({ id: 10, date: "2026-09-01" })
      const { result } = renderHook(() => useDeadlineManage(makeEvent([deadline])))
      fillContributor(result)

      act(() => {
        result.current.startSupersede(10)
        result.current.updateSupersede(10, "date", "2026-10-15")
      })

      expect(result.current.validate()).toBe(true)
    })

    // --- scoping: only validate deadlines that need it ---

    it("does not add errors for deadlines in default mode", () => {
      const deadline = makeDeadline({ id: 10 })
      const { result } = renderHook(() => useDeadlineManage(makeEvent([deadline])))
      fillContributor(result)
      // deadline 10 stays in default mode — no edit entered

      expect(result.current.validate()).toBe(true)
      expect(result.current.errors["supersede.10.date"]).toBeFalsy()
    })

    it("does not add errors for deadlines in pendingCancel mode (no editable fields)", () => {
      const deadline = makeDeadline({ id: 10 })
      const { result } = renderHook(() => useDeadlineManage(makeEvent([deadline])))
      fillContributor(result)

      act(() => result.current.cancelDeadlineLocal(10))

      expect(result.current.validate()).toBe(true)
      expect(result.current.errors["supersede.10.date"]).toBeFalsy()
    })

    // --- error timing ---
    // Spec: "Inline errors shown only after a submit attempt; not on blur"

    it("exposes no errors before validate() is called", () => {
      const { result } = renderHook(() => useDeadlineManage(makeEvent([])))
      // contributor fields empty, but validate() never called

      expect(Object.keys(result.current.errors)).toHaveLength(0)
    })

    it("clears stale errors and recomputes on repeated validate() calls", () => {
      const { result } = renderHook(() => useDeadlineManage(makeEvent([])))

      act(() => result.current.updateContributor("email", "alice@example.com"))
      // name is empty → first validate fails
      expect(result.current.validate()).toBe(false)
      expect(result.current.errors["contributor.name"]).toBeTruthy()

      // fix name, validate again → error clears
      act(() => result.current.updateContributor("name", "Alice"))
      expect(result.current.validate()).toBe(true)
      expect(result.current.errors["contributor.name"]).toBeFalsy()
    })
  })

  // --- submitChanges ---
  // Spec: "Submission logic" section

  describe("submitChanges", () => {
    const resolvedEvent = makeEvent([])

    beforeEach(() => {
      vi.clearAllMocks()
      vi.mocked(cancelDeadline).mockResolvedValue(resolvedEvent)
      vi.mocked(supersedeDeadline).mockResolvedValue(resolvedEvent)
      vi.mocked(addDeadlines).mockResolvedValue(resolvedEvent)
    })

    function fillContributor(
      result: ReturnType<typeof renderHook<ReturnType<typeof useDeadlineManage>, EventListItem>>["result"],
    ): void {
      act(() => {
        result.current.updateContributor("name", "Alice")
        result.current.updateContributor("email", "alice@example.com")
      })
    }

    it("makes no API calls and stays on editing when validate() fails", async () => {
      // contributor fields empty → validate() returns false
      const { result } = renderHook(() => useDeadlineManage(makeEvent([])))

      await act(async () => {
        await result.current.submitChanges()
      })

      expect(cancelDeadline).not.toHaveBeenCalled()
      expect(supersedeDeadline).not.toHaveBeenCalled()
      expect(addDeadlines).not.toHaveBeenCalled()
      expect(result.current.pageState).toBe("editing")
    })

    it("calls cancelDeadline for each deadline in pendingCancel state", async () => {
      // Spec: "Each pending-cancel → cancelDeadline(eventId, id, {submitter})"
      const event = makeEvent([makeDeadline({ id: 10 }), makeDeadline({ id: 20 })])
      const { result } = renderHook(() => useDeadlineManage(event))
      fillContributor(result)

      act(() => {
        result.current.cancelDeadlineLocal(10)
        result.current.cancelDeadlineLocal(20)
      })

      await act(async () => {
        await result.current.submitChanges()
      })

      const submitter = { name: "Alice", email: "alice@example.com" }
      expect(cancelDeadline).toHaveBeenCalledWith(event.id, 10, { submitter })
      expect(cancelDeadline).toHaveBeenCalledWith(event.id, 20, { submitter })
      expect(cancelDeadline).toHaveBeenCalledTimes(2)
    })

    it("calls supersedeDeadline for each superseding deadline with changed values", async () => {
      // Spec: "Each supersede-edit with changes → supersedeDeadline(eventId, id, {submitter, date, ...})"
      const deadline = makeDeadline({ id: 10, date: "2026-09-01" })
      const event = makeEvent([deadline])
      const { result } = renderHook(() => useDeadlineManage(event))
      fillContributor(result)

      act(() => {
        result.current.startSupersede(10)
        result.current.updateSupersede(10, "date", "2026-10-15")
      })

      await act(async () => {
        await result.current.submitChanges()
      })

      expect(supersedeDeadline).toHaveBeenCalledWith(event.id, 10, {
        submitter: { name: "Alice", email: "alice@example.com" },
        date: "2026-10-15",
      })
    })

    it("omits time and timezone from the supersedeDeadline payload when they are empty", async () => {
      // Spec: time and timezone are optional — absent from payload when not set
      const deadline = makeDeadline({ id: 10, date: "2026-09-01", time: null, timezone: null })
      const event = makeEvent([deadline])
      const { result } = renderHook(() => useDeadlineManage(event))
      fillContributor(result)

      act(() => {
        result.current.startSupersede(10)
        result.current.updateSupersede(10, "date", "2026-10-15")
        // time and timezone remain ""
      })

      await act(async () => {
        await result.current.submitChanges()
      })

      const payload = vi.mocked(supersedeDeadline).mock.calls[0][2]
      expect(payload).not.toHaveProperty("time")
      expect(payload).not.toHaveProperty("timezone")
    })

    it("includes time and timezone in the supersedeDeadline payload when non-empty", async () => {
      const deadline = makeDeadline({ id: 10, date: "2026-09-01", time: null, timezone: null })
      const event = makeEvent([deadline])
      const { result } = renderHook(() => useDeadlineManage(event))
      fillContributor(result)

      act(() => {
        result.current.startSupersede(10)
        result.current.updateSupersede(10, "date", "2026-10-15")
        result.current.updateSupersede(10, "time", "14:00")
        result.current.updateSupersede(10, "timezone", "UTC")
      })

      await act(async () => {
        await result.current.submitChanges()
      })

      expect(supersedeDeadline).toHaveBeenCalledWith(event.id, 10, {
        submitter: { name: "Alice", email: "alice@example.com" },
        date: "2026-10-15",
        time: "14:00",
        timezone: "UTC",
      })
    })

    it("does not call supersedeDeadline for a superseding deadline whose values match the original", async () => {
      // Spec: "Supersede with same values as original → not counted as a change"
      // hasChanges is kept true by the pending-cancel on d1
      const d1 = makeDeadline({ id: 10 })
      const d2 = makeDeadline({ id: 20, date: "2026-09-01", time: null, timezone: null })
      const event = makeEvent([d1, d2])
      const { result } = renderHook(() => useDeadlineManage(event))
      fillContributor(result)

      act(() => {
        result.current.cancelDeadlineLocal(10)
        result.current.startSupersede(20) // enters edit without changing anything
      })

      await act(async () => {
        await result.current.submitChanges()
      })

      expect(supersedeDeadline).not.toHaveBeenCalled()
      expect(cancelDeadline).toHaveBeenCalledTimes(1)
    })

    it("calls addDeadlines once with all new deadline cards as a single batch", async () => {
      // Spec: "New cards (if any) → addDeadlines(eventId, {submitter, deadlines: [...]}) — one batch call"
      const event = makeEvent([])
      const { result } = renderHook(() => useDeadlineManage(event))
      fillContributor(result)

      act(() => {
        result.current.addNewDeadline()
        result.current.addNewDeadline()
      })

      act(() => {
        const [id1, id2] = result.current.newDeadlines.map(d => d.localId)
        result.current.updateNewDeadline(id1, "type", "abstract" as DeadlineType)
        result.current.updateNewDeadline(id1, "description", "Research track")
        result.current.updateNewDeadline(id1, "date", "2026-09-01")
        result.current.updateNewDeadline(id2, "type", "paper" as DeadlineType)
        result.current.updateNewDeadline(id2, "description", "Industry track")
        result.current.updateNewDeadline(id2, "date", "2026-10-01")
      })

      await act(async () => {
        await result.current.submitChanges()
      })

      expect(addDeadlines).toHaveBeenCalledTimes(1)
      const batchInput = vi.mocked(addDeadlines).mock.calls[0][1]
      expect(batchInput.deadlines).toHaveLength(2)
      expect(batchInput.deadlines[0].type).toBe("abstract")
      expect(batchInput.deadlines[1].type).toBe("paper")
    })

    it("does not call addDeadlines when there are no new deadline cards", async () => {
      const event = makeEvent([makeDeadline({ id: 10 })])
      const { result } = renderHook(() => useDeadlineManage(event))
      fillContributor(result)

      act(() => result.current.cancelDeadlineLocal(10))

      await act(async () => {
        await result.current.submitChanges()
      })

      expect(addDeadlines).not.toHaveBeenCalled()
    })

    it("sets pageState to success and successSummary with correct counts when all calls succeed", async () => {
      // Spec: "All succeed → render Success state"
      const d1 = makeDeadline({ id: 10 })
      const d2 = makeDeadline({ id: 20, date: "2026-09-01" })
      const event = makeEvent([d1, d2])
      const { result } = renderHook(() => useDeadlineManage(event))
      fillContributor(result)

      act(() => {
        result.current.cancelDeadlineLocal(10)
        result.current.startSupersede(20)
        result.current.updateSupersede(20, "date", "2026-10-15")
        result.current.addNewDeadline()
      })

      act(() => {
        const localId = result.current.newDeadlines[0].localId
        result.current.updateNewDeadline(localId, "type", "abstract" as DeadlineType)
        result.current.updateNewDeadline(localId, "description", "Track A")
        result.current.updateNewDeadline(localId, "date", "2026-08-01")
      })

      await act(async () => {
        await result.current.submitChanges()
      })

      expect(result.current.pageState).toBe("success")
      expect(result.current.successSummary).toEqual({ added: 1, updated: 1, cancelled: 1 })
    })

    it("sets apiError and returns pageState to editing when any API call rejects", async () => {
      // Spec: "Any fail → error toast + exit Submitting → return to Editing with all pending changes intact"
      vi.mocked(cancelDeadline).mockRejectedValue(
        new ApiError("DEADLINE_ALREADY_INACTIVE", 409, "deadline already inactive"),
      )

      const event = makeEvent([makeDeadline({ id: 10 })])
      const { result } = renderHook(() => useDeadlineManage(event))
      fillContributor(result)

      act(() => result.current.cancelDeadlineLocal(10))

      await act(async () => {
        await result.current.submitChanges()
      })

      expect(result.current.apiError).toBeTruthy()
      expect(result.current.pageState).toBe("editing")
    })

    it("preserves deadlineStates and newDeadlines when an API call fails", async () => {
      // Spec: "return to Editing with all pending changes intact"
      vi.mocked(cancelDeadline).mockRejectedValue(new ApiError("NETWORK_ERROR", 0, "network error"))

      const event = makeEvent([makeDeadline({ id: 10 })])
      const { result } = renderHook(() => useDeadlineManage(event))
      fillContributor(result)

      act(() => {
        result.current.cancelDeadlineLocal(10)
        result.current.addNewDeadline()
      })

      await act(async () => {
        await result.current.submitChanges()
      })

      expect(result.current.deadlineStates[10].mode).toBe("pendingCancel")
      expect(result.current.newDeadlines).toHaveLength(1)
    })
  })
})
