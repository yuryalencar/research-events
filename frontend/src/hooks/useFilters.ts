import { useState, useMemo, useCallback } from "react"

// --- Types ---

interface EventFilters {
  year: number | undefined
  domain?: string
  tier?: string
  country?: string
  firstDeadlineMonth?: number
}

interface UseFiltersReturn {
  draftFilters: EventFilters
  activeFilters: EventFilters
  isDirty: boolean
  setYear: (year: number | undefined) => void
  setDomain: (domain: string | undefined) => void
  setTier: (tier: string | undefined) => void
  setCountry: (country: string | undefined) => void
  setFirstDeadlineMonth: (month: number | undefined) => void
  apply: () => void
  reset: () => void
}

// --- Hook ---

// useFilters manages two parallel filter states: draft (what the user is
// currently configuring) and active (what's been applied and sent to the API).
// The globe only re-fetches when apply() or reset() is called, not on every
// individual control change.
function useFilters(currentYear: number): UseFiltersReturn {
  const defaults: EventFilters = useMemo(() => ({ year: currentYear }), [currentYear])

  const [draftFilters, setDraftFilters] = useState<EventFilters>(defaults)
  const [activeFilters, setActiveFilters] = useState<EventFilters>(defaults)

  // isDirty is true when draft diverges from applied — drives the dot indicator
  // on the collapsed toggle button.
  const isDirty = useMemo(
    () =>
      draftFilters.year !== activeFilters.year ||
      draftFilters.domain !== activeFilters.domain ||
      draftFilters.tier !== activeFilters.tier ||
      draftFilters.country !== activeFilters.country ||
      draftFilters.firstDeadlineMonth !== activeFilters.firstDeadlineMonth,
    [draftFilters, activeFilters],
  )

  const setYear = useCallback((year: number | undefined) => {
    setDraftFilters((prev) => ({ ...prev, year }))
  }, [])

  const setDomain = useCallback((domain: string | undefined) => {
    setDraftFilters((prev) => ({ ...prev, domain }))
  }, [])

  const setTier = useCallback((tier: string | undefined) => {
    setDraftFilters((prev) => ({ ...prev, tier }))
  }, [])

  const setCountry = useCallback((country: string | undefined) => {
    setDraftFilters((prev) => ({ ...prev, country }))
  }, [])

  const setFirstDeadlineMonth = useCallback((firstDeadlineMonth: number | undefined) => {
    setDraftFilters((prev) => ({ ...prev, firstDeadlineMonth }))
  }, [])

  // apply promotes draft to active, triggering a re-fetch in useEvents.
  const apply = useCallback(() => {
    setActiveFilters(draftFilters)
  }, [draftFilters])

  // reset restores the draft form to defaults without touching activeFilters.
  // The user must still click Apply to trigger a re-fetch with the reset values.
  const reset = useCallback(() => {
    setDraftFilters(defaults)
  }, [defaults])

  return {
    draftFilters,
    activeFilters,
    isDirty,
    setYear,
    setDomain,
    setTier,
    setCountry,
    setFirstDeadlineMonth,
    apply,
    reset,
  }
}

// --- Export ---

export { useFilters }
export type { EventFilters, UseFiltersReturn }
