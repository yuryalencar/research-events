import { useState, useRef, useEffect, useCallback } from "react"

import { listEvents } from "@/lib/api/events"
import type { EventListItem, EventStatus, EventTier, ApiMeta, ListEventsParams } from "@/types/api"

// --- Types ---

interface ReviewFilters {
  status: EventStatus
  tier: EventTier | undefined
  year: number
}

type ReviewPhase = "loading" | "ready" | "error"

interface FetchParams extends ReviewFilters {
  page: number
}

interface UseReviewEventsReturn {
  events: EventListItem[]
  meta: ApiMeta | null
  phase: ReviewPhase
  draftFilters: ReviewFilters
  setDraftStatus: (status: EventStatus) => void
  setDraftTier: (tier: EventTier | undefined) => void
  setDraftYear: (year: number) => void
  apply: () => void
  page: number
  goToPage: (page: number) => void
}

// --- Pure helpers ---

// isOwnEvent returns true when the event's creator matches the given userId.
// Pure: depends only on its arguments — no I/O, no side effects, always the
// same result for the same inputs. Used by the moderator card to gate review.
function isOwnEvent(event: EventListItem, userId: number): boolean {
  return event.created_by.id === userId
}

// --- Hook ---

// useReviewEvents manages the filter, pagination, and fetch state for the
// management dashboard event queue. Draft filters are only applied on apply()
// so that Prev/Next always use the last submitted filter set, not in-progress edits.
function useReviewEvents(initialYear: number): UseReviewEventsReturn {
  const initialFilters: ReviewFilters = { status: "pending", tier: undefined, year: initialYear }

  // 1. State
  const [draftFilters, setDraftFilters] = useState<ReviewFilters>(initialFilters)
  const [fetchParams, setFetchParams] = useState<FetchParams>({ ...initialFilters, page: 1 })
  const [events, setEvents] = useState<EventListItem[]>([])
  const [meta, setMeta] = useState<ApiMeta | null>(null)
  const [phase, setPhase] = useState<ReviewPhase>("loading")

  // Refs mirror the latest filter values so stable useCallback closures can read
  // them without being listed as deps (which would recreate callbacks on every render).
  const draftFiltersRef = useRef<ReviewFilters>(initialFilters)
  const appliedFiltersRef = useRef<ReviewFilters>(initialFilters)

  // 2. Fetch effect — fires on mount and whenever fetchParams changes.
  // fetchParams is always a new object reference from setFetchParams, so the effect
  // re-runs even when individual field values are unchanged (e.g. re-apply same filter).
  useEffect(() => {
    let isMounted = true
    setPhase("loading")

    const params: ListEventsParams = {
      status: fetchParams.status,
      year: fetchParams.year,
      page: fetchParams.page,
      page_size: 30,
      pagination: "on",
    }
    if (fetchParams.tier !== undefined) params.tier = fetchParams.tier

    listEvents(params)
      .then((result) => {
        if (!isMounted) return
        setEvents(result.data)
        setMeta(result.meta)
        setPhase("ready")
      })
      .catch(() => {
        if (!isMounted) return
        setEvents([])
        setMeta(null)
        setPhase("error")
      })

    return () => {
      isMounted = false
    }
  }, [fetchParams])

  // 3. Handlers — all stable references via useCallback with no deps.
  // They read latest values through refs instead of closing over state.

  const setDraftStatus = useCallback((status: EventStatus) => {
    const next = { ...draftFiltersRef.current, status }
    draftFiltersRef.current = next
    setDraftFilters(next)
  }, [])

  const setDraftTier = useCallback((tier: EventTier | undefined) => {
    const next = { ...draftFiltersRef.current, tier }
    draftFiltersRef.current = next
    setDraftFilters(next)
  }, [])

  const setDraftYear = useCallback((year: number) => {
    const next = { ...draftFiltersRef.current, year }
    draftFiltersRef.current = next
    setDraftFilters(next)
  }, [])

  // apply commits the current draft to the applied set and fetches page 1.
  const apply = useCallback(() => {
    const filters = draftFiltersRef.current
    appliedFiltersRef.current = filters
    setFetchParams({ ...filters, page: 1 })
  }, [])

  // goToPage fetches a different page using the already-applied filters,
  // not the in-progress draft — so filters can't change mid-pagination.
  const goToPage = useCallback((page: number) => {
    setFetchParams({ ...appliedFiltersRef.current, page })
  }, [])

  // 4. Return
  return {
    events,
    meta,
    phase,
    draftFilters,
    setDraftStatus,
    setDraftTier,
    setDraftYear,
    apply,
    page: fetchParams.page,
    goToPage,
  }
}

// --- Export ---

export { useReviewEvents, isOwnEvent }
export type { UseReviewEventsReturn, ReviewFilters, ReviewPhase }
