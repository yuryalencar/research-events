import { useState, useEffect, useCallback, useMemo, useRef } from "react"

import { listEvents } from "@/lib/api/events"
import type { EventListItem } from "@/types/api"

// --- Types ---

interface UseEventSearchReturn {
  events: EventListItem[]
  filteredEvents: EventListItem[] // client-side text filter on current page
  isLoading: boolean
  error: string | null
  page: number
  totalPages: number
  total: number
  searchText: string
  year: number // draft year — only applied to the fetch when apply() or clear() is called
  setSearchText: (text: string) => void
  setYear: (year: number) => void
  apply: () => void // always re-fetches at page 1
  clear: () => void // resets all filters and re-fetches at page 1
  goToPage: (page: number) => void
}

// --- Hook ---

// useEventSearch manages fetching, paginating, and client-side filtering of
// pending events for the duplicate-check step of the submission wizard.
//
// Spec: GET /api/v1/events?status=pending&year=<year>&page=<page>&page_size=<size>
// Text search is client-side on the current page — the backend has no ?q= param.
// apply() and clear() always reset to page 1 so filter changes never leave
// stale page offsets in place.
function useEventSearch(pageSize = 10): UseEventSearchReturn {
  // Stable initial year — captured once on mount so clear() always resets to
  // the year the hook was first rendered, not the current year on each call.
  const initialYearRef = useRef(new Date().getFullYear())

  const [events, setEvents] = useState<EventListItem[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [searchText, setSearchText] = useState("")

  // draftYear: what the user has typed in the year field (not yet fetched).
  const [draftYear, setDraftYear] = useState(initialYearRef.current)

  // appliedYear + revision: what was last committed via apply() / clear().
  // revision increments to force a re-fetch even when appliedYear and page
  // are unchanged (e.g. the user presses Apply twice in a row).
  const [appliedYear, setAppliedYear] = useState(initialYearRef.current)
  const [revision, setRevision] = useState(0)

  useEffect(() => {
    let cancelled = false

    const run = async (): Promise<void> => {
      setIsLoading(true)
      setError(null)
      try {
        const result = await listEvents({
          status: "pending",
          year: appliedYear,
          page,
          page_size: pageSize,
        })
        if (!cancelled) {
          setEvents(result.data)
          setTotal(result.meta.total)
        }
      } catch {
        if (!cancelled) {
          setError("Could not load the events list")
          setEvents([])
        }
      } finally {
        if (!cancelled) {
          setIsLoading(false)
        }
      }
    }

    run()
    return (): void => {
      cancelled = true
    }
  }, [appliedYear, page, pageSize, revision])

  // Spec: "Apply pressed → always re-fetch at page 1"
  const apply = useCallback((): void => {
    setAppliedYear(draftYear)
    setPage(1)
    setRevision(r => r + 1)
  }, [draftYear])

  // Spec: "All search fields cleared + Apply → re-fetch page 1 with no filter"
  const clear = useCallback((): void => {
    const year = initialYearRef.current
    setSearchText("")
    setDraftYear(year)
    setAppliedYear(year)
    setPage(1)
    setRevision(r => r + 1)
  }, [])

  // Spec: "Previous/Next re-fetch at the new page number without resetting filters"
  const goToPage = useCallback((newPage: number): void => {
    setPage(newPage)
  }, [])

  const filteredEvents = useMemo((): EventListItem[] => {
    if (!searchText.trim()) return events
    const lower = searchText.toLowerCase()
    return events.filter(
      e => e.name.toLowerCase().includes(lower) || e.slug.toLowerCase().includes(lower),
    )
  }, [events, searchText])

  const totalPages = useMemo((): number => Math.ceil(total / pageSize), [total, pageSize])

  return {
    events,
    filteredEvents,
    isLoading,
    error,
    page,
    totalPages,
    total,
    searchText,
    year: draftYear,
    setSearchText,
    setYear: setDraftYear,
    apply,
    clear,
    goToPage,
  }
}

// --- Export ---

export { useEventSearch }
export type { UseEventSearchReturn }
