import { describe, it, expect } from "vitest"
import { renderHook, act } from "@testing-library/react"

import { useFilters } from "./useFilters"

const CURRENT_YEAR = 2026

describe("useFilters", () => {
  // --- Initial state ---
  // Spec: "Default on page load: current year" (year-filter-from-semantics.md)

  it("initialises draft and applied year to currentYear", () => {
    const { result } = renderHook(() => useFilters(CURRENT_YEAR))

    expect(result.current.draftFilters.year).toBe(CURRENT_YEAR)
    expect(result.current.activeFilters.year).toBe(CURRENT_YEAR)
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

  it("setYear(undefined) clears draft year to undefined", () => {
    // Spec: year-filter-from-semantics.md — clearing year sets it to undefined (all events)
    const { result } = renderHook(() => useFilters(CURRENT_YEAR))

    act(() => result.current.setYear(undefined))

    expect(result.current.draftFilters.year).toBeUndefined()
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

  it("apply with undefined year propagates undefined to applied", () => {
    // Spec: year-filter-from-semantics.md — applying with no year fetches all events
    const { result } = renderHook(() => useFilters(CURRENT_YEAR))

    act(() => result.current.setYear(undefined))
    act(() => result.current.apply())

    expect(result.current.activeFilters.year).toBeUndefined()
  })

  // --- reset ---
  // Spec: "reset() restores year to current year" (year-filter-from-semantics.md)

  it("reset restores draft year to currentYear", () => {
    const { result } = renderHook(() => useFilters(CURRENT_YEAR))

    act(() => result.current.setYear(2020))
    act(() => result.current.reset())

    expect(result.current.draftFilters.year).toBe(CURRENT_YEAR)
  })

  it("reset restores draft year to currentYear when it was cleared to undefined", () => {
    // Spec: year-filter-from-semantics.md — reset() always returns to currentYear default
    const { result } = renderHook(() => useFilters(CURRENT_YEAR))

    act(() => result.current.setYear(undefined))
    act(() => result.current.reset())

    expect(result.current.draftFilters.year).toBe(CURRENT_YEAR)
  })

  it("reset does not change applied year", () => {
    const { result } = renderHook(() => useFilters(CURRENT_YEAR))

    act(() => result.current.setYear(2020))
    act(() => result.current.apply())
    act(() => result.current.reset())

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

    expect(result.current.activeFilters.domain).toBe("computer_science")
    expect(result.current.activeFilters.tier).toBe("A")
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

  it("isDirty is true when year is cleared to undefined", () => {
    // Spec: year-filter-from-semantics.md — undefined year differs from currentYear default
    const { result } = renderHook(() => useFilters(CURRENT_YEAR))

    act(() => result.current.setYear(undefined))

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

    act(() => {
      result.current.setYear(2025)
      result.current.setDomain("computer_science")
    })
    act(() => result.current.reset())

    expect(result.current.isDirty).toBe(false)
  })

  it("isDirty is true after reset when active has applied values", () => {
    const { result } = renderHook(() => useFilters(CURRENT_YEAR))

    act(() => result.current.setDomain("computer_science"))
    act(() => result.current.apply())
    act(() => result.current.reset())

    expect(result.current.isDirty).toBe(true)
  })
})
