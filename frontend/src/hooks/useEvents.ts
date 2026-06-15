import { useEffect, useState } from "react"
import { useTranslations } from "next-intl"

import { listEvents } from "@/lib/api/events"
import { handleApiError } from "@/lib/api/errors"
import type { EventListItem } from "@/types/api"

// --- Types ---

interface UseEventsReturn {
  events: EventListItem[]
  isLoading: boolean
}

// --- Hook ---

// useEvents fetches all approved events for the current year on mount, with
// pagination disabled so every matching event is returned in one response
// (see specs/frontend/globe-homepage.md — "Behaviour"). On failure, shows a
// translated toast via handleApiError and leaves events empty.
function useEvents(): UseEventsReturn {
  // 1. State declarations
  const [events, setEvents] = useState<EventListItem[]>([])
  const [isLoading, setIsLoading] = useState(true)

  // 2. Derived values
  const t = useTranslations()

  // 3. Effects
  useEffect(() => {
    let isMounted = true

    listEvents({ pagination: "off" })
      .then((result) => {
        if (isMounted) setEvents(result.data)
      })
      .catch((error: unknown) => {
        handleApiError(error, t)
      })
      .finally(() => {
        if (isMounted) setIsLoading(false)
      })

    return () => {
      isMounted = false
    }
  }, [t])

  // 4. Return
  return { events, isLoading }
}

// --- Export ---

export { useEvents }
export type { UseEventsReturn }
