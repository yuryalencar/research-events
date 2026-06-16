import { describe, it, expect } from "vitest"
import { renderHook, act } from "@testing-library/react"

import { useFilters } from "./useFilters"

const CURRENT_YEAR = 2026

describe("useFilters", () => {
  // --- Initial state ---
  // Spec: "year is always a number, never undefined" (Rules)
  // Spec: "Initial value: current year" (Rules)

  it("initialises draft and applied year to currentYear", () => {
    const { result } = renderHook(() => useFilters(CURRENT_YEAR))

    expect(result.current.draftFilters.year).toBe(CURRENT_YEAR)
    expect(result.current.activeFilters.year).toBe(CURRENT_YEAR)
  })

  it("year in initial draft state is never undefined", () => {
    const { result } = renderHook(() => useFilters(CURRENT_YEAR))

    expect(result.current.draftFilters.year).toBeDefined()
    expect(typeof result.current.draftFilters.year).toBe("number")
  })

  it("year in initial applied state is never undefined", () => {
    const { result } = renderHook(() => useFilters(CURRENT_YEAR))

    expect(result.current.activeFilters.year).toBeDefined()
    expect(typeof result.current.activeFilters.year).toBe("number")
  })

  it("optional filters are undefined in initial state", () => {
    const { result } = renderHook(() => useFilters(CURRENT_YEAR))

    expect(result.current.draftFilters.domain).toBeUndefined()
    expect(result.current.draftFilters.tier).toBeUndefined()
    expect(result.current.draftFilters.country).toBeUndefined()
    expect(result.current.draftFilters.firstDeadlineMonth).toBeUndefined()
  })

  // --- Setters update draft only, not applied ---
  // Spec: "Changing any filter control updates draft state only" (Behaviour)

  it("setYear updates draft year but not applied year", () => {
    const { result } = renderHook(() => useFilters(CURRENT_YEAR))

    act(() => result.current.setYear(2025))

    expect(result.current.draftFilters.year).toBe(2025)
    expect(result.current.activeFilters.year).toBe(CURRENT_YEAR)
  })

  it("setDomain updates draft domain but not applied domain", () => {
    const { result } = renderHook(() => useFilters(CURRENT_YEAR))

    act(() => result.current.setDomain("computer_science"))

    expect(result.current.draftFilters.domain).toBe("computer_science")
    expect(result.current.activeFilters.domain).toBeUndefined()
  })

  it("setTier updates draft tier but not applied tier", () => {
    const { result } = renderHook(() => useFilters(CURRENT_YEAR))

    act(() => result.current.setTier("A*"))

    expect(result.current.draftFilters.tier).toBe("A*")
    expect(result.current.activeFilters.tier).toBeUndefined()
  })

  it("setCountry updates draft country but not applied country", () => {
    const { result } = renderHook(() => useFilters(CURRENT_YEAR))

    act(() => result.current.setCountry("Brazil"))

    expect(result.current.draftFilters.country).toBe("Brazil")
    expect(result.current.activeFilters.country).toBeUndefined()
  })

  it("setFirstDeadlineMonth updates draft month but not applied month", () => {
    const { result } = renderHook(() => useFilters(CURRENT_YEAR))

    act(() => result.current.setFirstDeadlineMonth(6))

    expect(result.current.draftFilters.firstDeadlineMonth).toBe(6)
    expect(result.current.activeFilters.firstDeadlineMonth).toBeUndefined()
  })

  it("year in draft never becomes undefined after setYear", () => {
    const { result } = renderHook(() => useFilters(CURRENT_YEAR))

    act(() => result.current.setYear(2020))

    expect(result.current.draftFilters.year).toBeDefined()
    expect(typeof result.current.draftFilters.year).toBe("number")
  })

  // --- apply ---
  // Spec: "Apply button: copies draft → applied filters → triggers re-fetch" (Behaviour)

  it("apply copies all draft fields to applied", () => {
    const { result } = renderHook(() => useFilters(CURRENT_YEAR))

    act(() => {
      result.current.setYear(2025)
      result.current.setDomain("computer_science")
      result.current.setTier("A")
      result.current.setCountry("Brazil")
      result.current.setFirstDeadlineMonth(6)
    })
    act(() => result.current.apply())

    expect(result.current.activeFilters).toEqual({
      year: 2025,
      domain: "computer_science",
      tier: "A",
      country: "Brazil",
      firstDeadlineMonth: 6,
    })
  })

  it("apply with year only and all optional filters undefined is valid", () => {
    // Spec border case: "Apply with year only (all optional filters cleared) is valid"
    const { result } = renderHook(() => useFilters(CURRENT_YEAR))

    act(() => result.current.apply())

    expect(result.current.activeFilters.year).toBe(CURRENT_YEAR)
    expect(result.current.activeFilters.domain).toBeUndefined()
    expect(result.current.activeFilters.tier).toBeUndefined()
    expect(result.current.activeFilters.country).toBeUndefined()
    expect(result.current.activeFilters.firstDeadlineMonth).toBeUndefined()
  })

  it("year in applied never becomes undefined after apply", () => {
    const { result } = renderHook(() => useFilters(CURRENT_YEAR))

    act(() => result.current.setYear(2020))
    act(() => result.current.apply())

    expect(result.current.activeFilters.year).toBeDefined()
    expect(typeof result.current.activeFilters.year).toBe("number")
  })

  // --- reset ---
  // Spec (revised): "Reset button restores the draft form to defaults but does
  // NOT apply — the user must still click Apply to trigger a re-fetch." (Behaviour)

  it("reset restores draft year to currentYear", () => {
    const { result } = renderHook(() => useFilters(CURRENT_YEAR))

    act(() => result.current.setYear(2020))
    act(() => result.current.reset())

    expect(result.current.draftFilters.year).toBe(CURRENT_YEAR)
  })

  it("reset does not change applied year", () => {
    const { result } = renderHook(() => useFilters(CURRENT_YEAR))

    act(() => result.current.setYear(2020))
    act(() => result.current.apply())
    act(() => result.current.reset())

    // active should still reflect what was applied before the reset
    expect(result.current.activeFilters.year).toBe(2020)
  })

  it("reset clears all optional fields from draft", () => {
    const { result } = renderHook(() => useFilters(CURRENT_YEAR))

    act(() => {
      result.current.setDomain("computer_science")
      result.current.setTier("A")
      result.current.setCountry("Brazil")
      result.current.setFirstDeadlineMonth(6)
    })
    act(() => result.current.reset())

    expect(result.current.draftFilters.domain).toBeUndefined()
    expect(result.current.draftFilters.tier).toBeUndefined()
    expect(result.current.draftFilters.country).toBeUndefined()
    expect(result.current.draftFilters.firstDeadlineMonth).toBeUndefined()
  })

  it("reset does not change applied optional fields", () => {
    const { result } = renderHook(() => useFilters(CURRENT_YEAR))

    act(() => {
      result.current.setDomain("computer_science")
      result.current.setTier("A")
    })
    act(() => result.current.apply())
    act(() => result.current.reset())

    // active should still reflect what was applied before the reset
    expect(result.current.activeFilters.domain).toBe("computer_science")
    expect(result.current.activeFilters.tier).toBe("A")
  })

  it("year in applied never becomes undefined after reset", () => {
    const { result } = renderHook(() => useFilters(CURRENT_YEAR))

    act(() => result.current.reset())

    expect(result.current.activeFilters.year).toBeDefined()
    expect(typeof result.current.activeFilters.year).toBe("number")
  })

  // --- isDirty ---
  // Spec: "A visual indicator shows when applied filters differ from defaults" (Behaviour)

  it("isDirty is false when draft equals applied on init", () => {
    const { result } = renderHook(() => useFilters(CURRENT_YEAR))

    expect(result.current.isDirty).toBe(false)
  })

  it("isDirty is true when draft year differs from applied year", () => {
    const { result } = renderHook(() => useFilters(CURRENT_YEAR))

    act(() => result.current.setYear(2025))

    expect(result.current.isDirty).toBe(true)
  })

  it("isDirty is true when draft has an optional filter that applied does not", () => {
    const { result } = renderHook(() => useFilters(CURRENT_YEAR))

    act(() => result.current.setDomain("computer_science"))

    expect(result.current.isDirty).toBe(true)
  })

  it("isDirty is false immediately after apply", () => {
    const { result } = renderHook(() => useFilters(CURRENT_YEAR))

    act(() => result.current.setYear(2025))
    act(() => result.current.apply())

    expect(result.current.isDirty).toBe(false)
  })

  it("isDirty is false after reset when active was never changed", () => {
    const { result } = renderHook(() => useFilters(CURRENT_YEAR))

    // draft changes but never applied, so active is still at defaults
    act(() => {
      result.current.setYear(2025)
      result.current.setDomain("computer_science")
    })
    act(() => result.current.reset())

    // draft returned to defaults; active was already at defaults → not dirty
    expect(result.current.isDirty).toBe(false)
  })

  it("isDirty is true after reset when active has applied values", () => {
    const { result } = renderHook(() => useFilters(CURRENT_YEAR))

    act(() => result.current.setDomain("computer_science"))
    act(() => result.current.apply())
    // draft now back to defaults, but active still has domain set
    act(() => result.current.reset())

    expect(result.current.isDirty).toBe(true)
  })
})
