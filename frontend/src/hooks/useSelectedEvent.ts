import { useCallback, useState } from "react"

import type { EventListItem } from "@/types/api"

// --- Types ---

interface UseSelectedEventReturn {
  selectedEvent: EventListItem | null
  selectEvent: (event: EventListItem) => void
  closeDetail: () => void
}

// --- Hook ---

// useSelectedEvent tracks which pin/event is currently highlighted and shown
// in the detail view. Clicking the already-selected event again toggles it
// off (see specs/frontend/globe-homepage.md — "Clicking the currently-
// highlighted pin again ... closes the detail view and removes the
// highlight").
function useSelectedEvent(): UseSelectedEventReturn {
  // 1. State declarations
  const [selectedEvent, setSelectedEvent] = useState<EventListItem | null>(null)

  // 2. Handlers
  const selectEvent = useCallback((event: EventListItem) => {
    setSelectedEvent((current) => (current?.id === event.id ? null : event))
  }, [])

  const closeDetail = useCallback(() => {
    setSelectedEvent(null)
  }, [])

  // 3. Return
  return { selectedEvent, selectEvent, closeDetail }
}

// --- Export ---

export { useSelectedEvent }
export type { UseSelectedEventReturn }
