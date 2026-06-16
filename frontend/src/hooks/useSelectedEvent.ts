import { useCallback, useEffect, useRef, useState } from "react"
import { useRouter, useSearchParams, usePathname } from "next/navigation"

import type { EventListItem } from "@/types/api"

// --- Types ---

interface UseSelectedEventReturn {
  selectedEvent: EventListItem | null
  selectEvent: (event: EventListItem) => void
  closeDetail: () => void
}

// --- Hook ---

// useSelectedEvent tracks which pin/event is currently highlighted and shown
// in the detail view. It syncs the selection with the URL via the `?event=`
// query param (router.replace — no history entry) so the page can be shared
// or bookmarked with an event pre-selected.
//
// Clicking an already-selected event toggles it off. On page load, if
// `?event=SLUG` is in the URL and a matching event exists in `events`, it is
// selected automatically — identical to clicking that pin.
function useSelectedEvent(events: EventListItem[]): UseSelectedEventReturn {
  // 1. State declarations
  const [selectedEvent, setSelectedEvent] = useState<EventListItem | null>(null)

  // 2. Next.js navigation — client-only; must be called from a Client Component.
  const router = useRouter()
  const searchParams = useSearchParams()
  const pathname = usePathname()

  // 3. Guards the on-load slug resolution so it runs at most once, preventing
  //    a re-run loop when searchParams or removeEventParam references change.
  const slugResolvedRef = useRef(false)

  // 4. Shared URL helper — removes only the `event` param, preserving others.
  const removeEventParam = useCallback(() => {
    const params = new URLSearchParams(searchParams.toString())
    params.delete("event")
    const qs = params.toString()
    router.replace(qs ? `?${qs}` : pathname)
  }, [router, searchParams, pathname])

  // 5. Handlers
  const selectEvent = useCallback(
    (event: EventListItem) => {
      const isDeselecting = selectedEvent?.id === event.id

      if (isDeselecting) {
        setSelectedEvent(null)
        removeEventParam()
      } else {
        setSelectedEvent(event)
        const params = new URLSearchParams(searchParams.toString())
        params.set("event", event.slug)
        router.replace(`?${params.toString()}`)
      }
    },
    [selectedEvent, searchParams, router, removeEventParam],
  )

  const closeDetail = useCallback(() => {
    setSelectedEvent(null)
    removeEventParam()
  }, [removeEventParam])

  // 6. On-load slug resolution: once events arrive from the API, check if the
  //    URL contains `?event=SLUG` and auto-select the matching event. Runs at
  //    most once (guarded by ref) so it doesn't fight with user interactions.
  useEffect(() => {
    if (slugResolvedRef.current) return
    if (events.length === 0) return // still loading — wait for data

    const slug = searchParams.get("event")
    slugResolvedRef.current = true // mark attempted regardless of outcome

    if (!slug) return

    const found = events.find((e) => e.slug === slug)
    if (found) {
      setSelectedEvent(found)
    } else {
      // Slug present in URL but no matching event — remove silently.
      removeEventParam()
    }
  }, [events, searchParams, removeEventParam])

  // 7. Return
  return { selectedEvent, selectEvent, closeDetail }
}

// --- Export ---

export { useSelectedEvent }
export type { UseSelectedEventReturn }
