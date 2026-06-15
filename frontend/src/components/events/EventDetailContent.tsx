import type { JSX } from "react"
import { useLocale, useTranslations } from "next-intl"

import { Badge } from "@/components/ui/badge"
import { getTierBadgeLabel, shouldShowUpdatedBy, formatDateRange } from "@/lib/events"
import type { EventListItem } from "@/types/api"

// --- Types & Props ---

interface EventDetailContentProps {
  event: EventListItem
}

// --- Component ---

// EventDetailContent renders the full details for a selected event — shared
// between the desktop side panel and the mobile bottom card (see
// specs/frontend/globe-homepage.md — "Detail view contents").
function EventDetailContent({ event }: EventDetailContentProps): JSX.Element {
  const t = useTranslations("eventDetail")
  const locale = useLocale()

  const tierLabel = getTierBadgeLabel(event.tier)
  const activeDeadlines = event.deadlines.filter((deadline) => deadline.is_active)
  const showUpdatedBy = shouldShowUpdatedBy(event)

  // domain is an extensible string enum (see CLAUDE.md) — fall back to the
  // raw value if a translation for it hasn't been added yet.
  const domainKey = `domains.${event.domain}`
  const domainLabel = t.has(domainKey) ? t(domainKey) : event.domain

  return (
    <div className="flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto px-4 pb-4">
      <dl className="flex flex-col gap-3 text-sm">
        {tierLabel !== null && (
          <div>
            <dt className="text-muted-foreground">{t("tier")}</dt>
            <dd>
              <Badge>{tierLabel}</Badge>
            </dd>
          </div>
        )}
        <div>
          <dt className="text-muted-foreground">{t("dates")}</dt>
          <dd>{formatDateRange(event.start_date, event.end_date, locale)}</dd>
        </div>
        <div>
          <dt className="text-muted-foreground">{t("location")}</dt>
          <dd>
            {event.city}, {event.country}
          </dd>
        </div>
        <div>
          <dt className="text-muted-foreground">{t("website")}</dt>
          <dd>
            <a
              href={event.website_url}
              target="_blank"
              rel="noreferrer"
              className="text-primary underline underline-offset-2"
            >
              {event.website_url}
            </a>
          </dd>
        </div>
        <div>
          <dt className="text-muted-foreground">{t("domain")}</dt>
          <dd>{domainLabel}</dd>
        </div>
      </dl>

      <DeadlineList deadlines={activeDeadlines} />

      <div className="flex flex-col gap-1 text-xs text-muted-foreground">
        <p>{t("addedBy", { name: event.created_by.name })}</p>
        {showUpdatedBy && <p>{t("updatedBy", { name: event.last_updated_by.name })}</p>}
      </div>
    </div>
  )
}

// --- Sub-components ---

interface DeadlineListProps {
  deadlines: EventListItem["deadlines"]
}

function DeadlineList({ deadlines }: DeadlineListProps): JSX.Element {
  const t = useTranslations("eventDetail")

  if (deadlines.length === 0) {
    return <p className="text-sm text-muted-foreground">{t("noUpcomingDeadlines")}</p>
  }

  return (
    <ul className="flex flex-col gap-2">
      {deadlines.map((deadline) => (
        <li key={deadline.id} className="flex flex-col gap-1 rounded-md border p-2 text-sm">
          <div className="flex items-center justify-between gap-2">
            <span className="font-medium">{t(`deadlineTypes.${deadline.type}`)}</span>
            {deadline.is_optional && <Badge variant="secondary">{t("optional")}</Badge>}
          </div>
          <p className="text-muted-foreground">{deadline.description}</p>
          <p>
            {deadline.date}
            {deadline.time ? ` ${deadline.time}` : ""}
            {deadline.timezone ? ` (${deadline.timezone})` : ""}
          </p>
        </li>
      ))}
    </ul>
  )
}

// --- Export ---

export { EventDetailContent }
export type { EventDetailContentProps }
