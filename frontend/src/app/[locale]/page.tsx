"use client"

import type { JSX } from "react"
import dynamic from "next/dynamic"
import { useEffect, useRef, useState } from "react"
import { useTranslations } from "next-intl"

import { EventDetailView } from "@/components/events/EventDetailView"
import { FilterPanel } from "@/components/globe/FilterPanel"
import { InfoButton } from "@/components/globe/InfoButton"
import { useEvents } from "@/hooks/useEvents"
import { useFilters } from "@/hooks/useFilters"
import { useSelectedEvent } from "@/hooks/useSelectedEvent"

// Globe.gl requires WebGL/browser APIs — must never be server-rendered (see
// CLAUDE.md's Globe.gl/Leaflet rule).
const GlobeView = dynamic(() => import("@/components/globe/GlobeView").then((mod) => mod.GlobeView), {
  ssr: false,
})

// --- Component ---

export default function Page(): JSX.Element {
  const t = useTranslations("home")
  const currentYear = new Date().getFullYear()
  const filters = useFilters(currentYear)
  const { events, isLoading } = useEvents(filters.activeFilters)

  const { activeFilters } = filters
  const hasNonDefaultFilters =
    activeFilters.year !== currentYear ||
    activeFilters.domain !== undefined ||
    activeFilters.tier !== undefined ||
    activeFilters.country !== undefined ||
    activeFilters.firstDeadlineMonth !== undefined
  const { selectedEvent, selectEvent, closeDetail } = useSelectedEvent(events)

  // After filters are applied and the fetch completes, rotate the globe to the
  // first result. The ref skips the initial page-load fetch so the globe only
  // moves when the user explicitly presses Apply.
  const [focusPoint, setFocusPoint] = useState<{ lat: number; lng: number } | null>(null)
  const isFirstLoad = useRef(true)
  useEffect(() => {
    if (isLoading) return
    if (isFirstLoad.current) {
      isFirstLoad.current = false
      return
    }
    if (events.length > 0) {
      setFocusPoint({ lat: events[0].latitude, lng: events[0].longitude })
    }
  }, [events, isLoading])

  return (
    <main className="space-bg relative h-screen w-full">
      <GlobeView events={events} selectedEvent={selectedEvent} onPointClick={selectEvent} focusPoint={focusPoint} />

      {isLoading && (
        <div className="absolute inset-0 flex items-center justify-center">
          <div
            role="status"
            className="flex items-center gap-3 rounded-lg border border-border bg-background/80 px-6 py-4 text-foreground shadow-lg"
          >
            <div aria-hidden="true" className="size-5 animate-spin rounded-full border-2 border-muted border-t-primary" />
            <span>{t("loading")}</span>
          </div>
        </div>
      )}

      {!isLoading && events.length === 0 && (
        <div className="absolute inset-0 flex items-center justify-center">
          <div
            role="status"
            className="rounded-lg border border-red-400 bg-red-100 px-6 py-4 text-red-900 shadow-lg dark:border-red-900 dark:bg-red-950 dark:text-red-100"
          >
            {hasNonDefaultFilters ? t("noEventsFiltered") : t("noEvents")}
          </div>
        </div>
      )}

      <FilterPanel filters={filters} />

      <InfoButton drawerOpen={selectedEvent !== null} />

      <EventDetailView event={selectedEvent} onClose={closeDetail} />
    </main>
  )
}
