"use client"

import type { JSX } from "react"
import { useTranslations } from "next-intl"

import { EventTableCard } from "@/components/events/EventTableCard"
import type { EventListItem } from "@/types/api"

// --- Types & Props ---

interface EventTableViewProps {
  events: EventListItem[]
  isLoading: boolean
  hasNonDefaultFilters: boolean
}

// --- Component ---

// EventTableView renders the full-screen card-list alternative to the globe.
// It occupies the same layout slot as the globe (full viewport height) and
// handles its own loading and empty states using the same visual language as
// the globe page so the two views feel consistent.
function EventTableView({ events, isLoading, hasNonDefaultFilters }: EventTableViewProps): JSX.Element {
  const t = useTranslations("home")

  if (isLoading) {
    return (
      <div className="flex h-full items-center justify-center">
        <div
          role="status"
          className="flex items-center gap-3 rounded-lg border border-border bg-background/80 px-6 py-4 text-foreground shadow-lg"
        >
          <div aria-hidden="true" className="size-5 animate-spin rounded-full border-2 border-muted border-t-primary" />
          <span>{t("loading")}</span>
        </div>
      </div>
    )
  }

  if (events.length === 0) {
    return (
      <div className="flex h-full items-center justify-center">
        <div
          role="status"
          className="rounded-lg border border-red-400 bg-red-100 px-6 py-4 text-red-900 shadow-lg dark:border-red-900 dark:bg-red-950 dark:text-red-100"
        >
          {hasNonDefaultFilters ? t("noEventsFiltered") : t("noEvents")}
        </div>
      </div>
    )
  }

  return (
    <div className="h-full overflow-y-auto">
      <ul className="mx-auto max-w-3xl flex flex-col gap-3 px-4 py-6 pb-24">
        {events.map((event) => (
          <li key={event.id}>
            <EventTableCard event={event} />
          </li>
        ))}
      </ul>
    </div>
  )
}

// --- Export ---

export { EventTableView }
export type { EventTableViewProps }
