"use client"

import { useState } from "react"
import type { JSX } from "react"
import { ChevronDownIcon, ChevronUpIcon, PencilIcon } from "lucide-react"
import { useTranslations, useLocale } from "next-intl"
import { useRouter } from "next/navigation"

import { Badge } from "@/components/ui/badge"
import { formatDateRange, getTierBadgeLabel } from "@/lib/events"
import type { DeadlineResponse, EventListItem } from "@/types/api"

// --- Types & Props ---

interface EventTableCardProps {
  event: EventListItem
}

// Shared with EventDetailView — identifies the sessionStorage key used to pass
// event data to the deadline management page without a separate fetch.
const SESSION_KEY = "deadline_management_event"

// --- Component ---

// EventTableCard renders one event as a summary card in the table view. All
// event fields are visible in the collapsed state; expanding reveals the full
// deadline list and a link to the deadline management page.
function EventTableCard({ event }: EventTableCardProps): JSX.Element {
  const [isExpanded, setIsExpanded] = useState(false)
  const t = useTranslations("eventTable")
  const tDetail = useTranslations("eventDetail")
  const locale = useLocale()
  const router = useRouter()

  const tierLabel = getTierBadgeLabel(event.tier)

  // domain is an extensible string enum — fall back to the raw value if no
  // translation key has been added yet (same pattern as EventDetailContent).
  const domainKey = `domains.${event.domain}` as const
  const domainLabel = tDetail.has(domainKey) ? tDetail(domainKey) : event.domain

  const activeDeadlines = event.deadlines.filter((d) => d.is_active)
  const nextDeadline = activeDeadlines
    .slice()
    .sort((a, b) => a.date.localeCompare(b.date))[0] ?? null

  const handleManageDeadlines = (): void => {
    sessionStorage.setItem(SESSION_KEY, JSON.stringify(event))
    router.push(`/${locale}/events/${event.slug}/deadlines`)
  }

  return (
    <div className="rounded-lg border border-border bg-background/80 shadow-sm backdrop-blur-sm">
      {/* --- Summary row (always visible) --- */}
      <div className="flex items-start justify-between gap-4 p-4">
        <div className="flex min-w-0 flex-1 flex-col gap-2">
          {/* Name + year */}
          <div className="flex flex-wrap items-baseline gap-2">
            <h3 className="truncate font-semibold text-foreground">{event.name}</h3>
            <span className="shrink-0 text-sm text-muted-foreground">({event.year})</span>
          </div>

          {/* Location + dates */}
          <div className="flex flex-wrap gap-x-4 gap-y-1 text-sm text-muted-foreground">
            <span>{event.city}, {event.country}</span>
            <span>{formatDateRange(event.start_date, event.end_date, locale)}</span>
          </div>

          {/* Domain + tier badges */}
          <div className="flex flex-wrap gap-2">
            <Badge variant="secondary">{domainLabel}</Badge>
            {tierLabel !== null && <Badge>{tierLabel}</Badge>}
          </div>

          {/* Next deadline summary */}
          <div className="text-sm">
            {nextDeadline !== null ? (
              <span>
                <span className="text-muted-foreground">{t("nextDeadline")}: </span>
                <span className="font-medium">
                  {tDetail(`deadlineTypes.${nextDeadline.type}`)} — {nextDeadline.date}
                </span>
              </span>
            ) : (
              <span className="text-muted-foreground">{t("noUpcomingDeadlines")}</span>
            )}
          </div>
        </div>

        {/* Expand toggle */}
        <button
          type="button"
          aria-label={isExpanded ? t("collapse") : t("expand")}
          aria-expanded={isExpanded}
          onClick={() => setIsExpanded((prev) => !prev)}
          className="mt-1 flex size-8 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
        >
          {isExpanded ? <ChevronUpIcon className="size-4" /> : <ChevronDownIcon className="size-4" />}
        </button>
      </div>

      {/* --- Expanded section --- */}
      {isExpanded && (
        <div className="border-t border-border px-4 pb-4 pt-3">
          <div className="mb-3 flex items-center justify-between">
            <span className="text-sm font-medium">{tDetail("deadlinesTitle")}</span>
            <button
              type="button"
              onClick={handleManageDeadlines}
              className="flex items-center gap-1.5 rounded-md px-2 py-1 text-xs text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
            >
              <PencilIcon className="size-3.5" />
              {t("manageDeadlines")}
            </button>
          </div>

          <DeadlineExpandedList deadlines={event.deadlines} />
        </div>
      )}
    </div>
  )
}

// --- Sub-components ---

interface DeadlineExpandedListProps {
  deadlines: DeadlineResponse[]
}

function DeadlineExpandedList({ deadlines }: DeadlineExpandedListProps): JSX.Element {
  const t = useTranslations("eventDetail")

  const activeDeadlines = deadlines
    .filter((d) => d.is_active)
    .sort((a, b) => a.date.localeCompare(b.date))

  if (activeDeadlines.length === 0) {
    return <p className="text-sm text-muted-foreground">{t("noUpcomingDeadlines")}</p>
  }

  return (
    <ul className="flex flex-col gap-2">
      {activeDeadlines.map((deadline) => {
        const predecessor = deadlines.find((d) => d.superseded_by_id === deadline.id)

        return (
          <li key={deadline.id} className="flex flex-col gap-1 rounded-md border p-2 text-sm">
            <div className="flex items-center justify-between gap-2">
              <span className="font-medium">{t(`deadlineTypes.${deadline.type}`)}</span>
              {deadline.is_optional && <Badge variant="secondary">{t("optional")}</Badge>}
            </div>
            {deadline.description && (
              <p className="text-muted-foreground">{deadline.description}</p>
            )}
            {predecessor ? (
              <p className="flex flex-wrap items-center gap-1.5">
                <span className="text-muted-foreground line-through">
                  {predecessor.date}
                  {predecessor.time ? ` ${predecessor.time}` : ""}
                  {predecessor.timezone ? ` (${predecessor.timezone})` : ""}
                </span>
                <span className="text-muted-foreground">→</span>
                <span>
                  {deadline.date}
                  {deadline.time ? ` ${deadline.time}` : ""}
                  {deadline.timezone ? ` (${deadline.timezone})` : ""}
                </span>
              </p>
            ) : (
              <p>
                {deadline.date}
                {deadline.time ? ` ${deadline.time}` : ""}
                {deadline.timezone ? ` (${deadline.timezone})` : ""}
              </p>
            )}
          </li>
        )
      })}
    </ul>
  )
}

// --- Export ---

export { EventTableCard }
export type { EventTableCardProps }
