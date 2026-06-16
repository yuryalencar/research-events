import { useEffect, useRef, useState } from "react"
import { useTranslations } from "next-intl"

import { listEvents } from "@/lib/api/events"
import { handleApiError } from "@/lib/api/errors"
import type { EventListItem, EventTier, ListEventsParams } from "@/types/api"
import type { EventFilters } from "./useFilters"

// --- Types ---

interface UseEventsReturn {
  events: EventListItem[]
  isLoading: boolean
}

// --- Helpers ---

// toListEventsParams converts camelCase EventFilters (frontend state) to the
// snake_case ListEventsParams the API expects. Only includes optional fields
// when they are explicitly set, so the query string stays clean.
function toListEventsParams(filters: EventFilters): ListEventsParams {
  const params: ListEventsParams = {
    year: filters.year,
    pagination: "off",
  }
  if (filters.domain !== undefined) params.domain = filters.domain
  if (filters.tier !== undefined) params.tier = filters.tier as EventTier
  if (filters.country !== undefined) params.country = filters.country
  if (filters.firstDeadlineMonth !== undefined) params.first_deadline_month = filters.firstDeadlineMonth
  return params
}

// --- Hook ---

// useEvents fetches all approved events matching the given filters, with
// pagination disabled so every matching event is returned in one response
// (see specs/frontend/globe-filters.md). Re-fetches whenever a filter value
// changes. On failure, shows a translated toast and leaves events empty.
function useEvents(filters: EventFilters): UseEventsReturn {
  // 1. State declarations
  const [events, setEvents] = useState<EventListItem[]>([])
  const [isLoading, setIsLoading] = useState(true)

  // 2. Refs — useTranslations() returns a new function reference on every
  // render. Keeping it in a ref means it never appears in the useEffect dep
  // array, so state updates don't trigger spurious re-fetches.
  const t = useTranslations()
  const tRef = useRef(t)
  tRef.current = t

  // 3. Destructure to primitives so the dep array uses stable scalar values
  // instead of the filter object reference (which changes on every render).
  const { year, domain, tier, country, firstDeadlineMonth } = filters

  useEffect(() => {
    let isMounted = true
    setIsLoading(true)

    listEvents(toListEventsParams({ year, domain, tier, country, firstDeadlineMonth }))
      .then((result) => {
        if (isMounted) setEvents(result.data)
      })
      .catch((error: unknown) => {
        handleApiError(error, tRef.current)
      })
      .finally(() => {
        if (isMounted) setIsLoading(false)
      })

    return () => {
      isMounted = false
    }
  }, [year, domain, tier, country, firstDeadlineMonth])

  // 4. Return
  return { events, isLoading }
}

// --- Export ---

export { useEvents }
export type { UseEventsReturn }
