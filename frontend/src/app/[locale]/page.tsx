"use client"

import type { JSX } from "react"
import dynamic from "next/dynamic"
import { useCallback, useEffect, useRef, useState } from "react"
import { useTranslations } from "next-intl"
import { toast } from "sonner"

import { AddEventButton } from "@/components/globe/AddEventButton"
import { ClusterEventDrawer } from "@/components/globe/ClusterEventDrawer"
import { EventDetailView } from "@/components/events/EventDetailView"
import { EventTableView } from "@/components/events/EventTableView"
import { FilterPanel } from "@/components/globe/FilterPanel"
import { InfoButton } from "@/components/globe/InfoButton"
import { ViewToggleButton } from "@/components/globe/ViewToggleButton"
import { useEvents } from "@/hooks/useEvents"
import { useFilters } from "@/hooks/useFilters"
import { useGlobeClusters } from "@/hooks/useGlobeClusters"
import { useSelectedEvent } from "@/hooks/useSelectedEvent"
import { useViewMode } from "@/hooks/useViewMode"
import type { EventListItem } from "@/types/api"

// Globe.gl requires WebGL/browser APIs — must never be server-rendered (see
// CLAUDE.md's Globe.gl/Leaflet rule).
const GlobeView = dynamic(() => import("@/components/globe/GlobeView").then((mod) => mod.GlobeView), {
  ssr: false,
})

// --- Component ---

export default function Page(): JSX.Element {
  const t = useTranslations("home")
  const tViewToggle = useTranslations("viewToggle")
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

  // zoom is reported by GlobeView via onZoomChange after each camera settle.
  // Initialised to 2, which matches globe.gl's default altitude of ~2.5.
  const [zoom, setZoom] = useState(2)
  const { globePoints, getClusterEvents } = useGlobeClusters(events, zoom)

  // clusterEvents drives the multi-event drawer: null = closed, array = open.
  const [clusterEvents, setClusterEvents] = useState<EventListItem[] | null>(null)

  const { selectedEvent, selectEvent, closeDetail } = useSelectedEvent(events)

  // When the viewport shrinks below md while in table mode, useViewMode calls
  // this: close any open detail panel and show a toast explaining the switch.
  const handleForcedGlobe = useCallback(() => {
    closeDetail()
    toast(tViewToggle("tableUnavailableToast"))
  }, [closeDetail, tViewToggle])

  const { viewMode, toggleView } = useViewMode(handleForcedGlobe)

  // Closing the detail panel on manual toggle keeps the two views independent —
  // a selected event on the globe has no meaning in the table view.
  const handleToggle = useCallback(() => {
    closeDetail()
    setClusterEvents(null)
    toggleView()
  }, [closeDetail, toggleView])

  const handleClusterClick = useCallback(
    (clusterId: number) => {
      setClusterEvents(getClusterEvents(clusterId))
    },
    [getClusterEvents],
  )

  const handleCloseClusterDrawer = useCallback(() => {
    setClusterEvents(null)
  }, [])

  const handleClusterEventSelect = useCallback(
    (event: EventListItem) => {
      setClusterEvents(null)
      selectEvent(event)
    },
    [selectEvent],
  )

  // After filters are applied and the fetch completes, rotate the globe to the
  // first result. The ref skips the initial page-load fetch so the globe only
  // moves when the user explicitly presses Apply.
  const [focusPoint, setFocusPoint] = useState<{ lat: number; lng: number } | null>(null)
  const isFirstLoad = useRef(true)
  useEffect(() => {
    // Close the cluster drawer as soon as a new filter fetch begins so stale
    // cluster content is never shown after the event list changes.
    if (isLoading) {
      setClusterEvents(null)
      return
    }
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
      {viewMode === "globe" ? (
        <>
          <GlobeView
            globePoints={globePoints}
            selectedEvent={selectedEvent}
            onPointClick={selectEvent}
            onClusterClick={handleClusterClick}
            onZoomChange={setZoom}
            focusPoint={focusPoint}
          />

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

          <InfoButton drawerOpen={selectedEvent !== null || clusterEvents !== null} />
          <EventDetailView event={selectedEvent} onClose={closeDetail} />
          <ClusterEventDrawer
            events={clusterEvents}
            onClose={handleCloseClusterDrawer}
            onSelectEvent={handleClusterEventSelect}
          />
        </>
      ) : (
        <EventTableView
          events={events}
          isLoading={isLoading}
          hasNonDefaultFilters={hasNonDefaultFilters}
        />
      )}

      <FilterPanel filters={filters} />
      <AddEventButton />
      <ViewToggleButton viewMode={viewMode} onToggle={handleToggle} />
    </main>
  )
}
