"use client"

import type { JSX } from "react"
import dynamic from "next/dynamic"
import { useTranslations } from "next-intl"

import { EventDetailView } from "@/components/events/EventDetailView"
import { InfoButton } from "@/components/globe/InfoButton"
import { useEvents } from "@/hooks/useEvents"
import { useSelectedEvent } from "@/hooks/useSelectedEvent"

// Globe.gl requires WebGL/browser APIs — must never be server-rendered (see
// CLAUDE.md's Globe.gl/Leaflet rule).
const GlobeView = dynamic(() => import("@/components/globe/GlobeView").then((mod) => mod.GlobeView), {
  ssr: false,
})

// --- Component ---

export default function Page(): JSX.Element {
  const t = useTranslations("home")
  const { events, isLoading } = useEvents()
  const { selectedEvent, selectEvent, closeDetail } = useSelectedEvent(events)

  return (
    <main className="space-bg relative h-screen w-full">
      <GlobeView events={events} selectedEvent={selectedEvent} onPointClick={selectEvent} />

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
            {t("noEvents")}
          </div>
        </div>
      )}

      <InfoButton drawerOpen={selectedEvent !== null} />

      <EventDetailView event={selectedEvent} onClose={closeDetail} />
    </main>
  )
}
