"use client"

import type { JSX } from "react"
import { useTranslations, useLocale } from "next-intl"

import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Drawer, DrawerContent, DrawerDescription, DrawerHeader, DrawerTitle } from "@/components/ui/drawer"
import { useMediaQuery } from "@/hooks/useMediaQuery"
import { formatDateRange } from "@/lib/events"
import type { EventListItem } from "@/types/api"

// --- Types ---

interface ClusterEventDrawerProps {
  events: EventListItem[] | null
  onClose: () => void
  onSelectEvent: (event: EventListItem) => void
}

interface ClusterEventListProps {
  events: EventListItem[]
  locale: string
  onSelectEvent: (event: EventListItem) => void
}

interface EventRowProps {
  event: EventListItem
  locale: string
  onSelect: () => void
}

// --- Component ---

// ClusterEventDrawer opens when the user clicks a cluster pin on the globe.
// It shows a compact, scrollable list of every event in the cluster so the
// user can pick the one they want before the normal EventDetailView opens.
// Rendered as a centered Dialog on desktop, a bottom Drawer on mobile —
// matching the responsive pattern used by EventDetailView and InfoModal.
function ClusterEventDrawer({ events, onClose, onSelectEvent }: ClusterEventDrawerProps): JSX.Element {
  const t = useTranslations("home")
  const locale = useLocale()
  const isDesktop = useMediaQuery("(min-width: 768px)")

  const open = events !== null
  const sorted = events ? [...events].sort((a, b) => a.start_date.localeCompare(b.start_date)) : []
  const title = t("cluster.drawerTitle", { count: sorted.length })

  const handleOpenChange = (next: boolean): void => {
    if (!next) onClose()
  }

  const content = <ClusterEventList events={sorted} locale={locale} onSelectEvent={onSelectEvent} />

  if (isDesktop) {
    return (
      <Dialog open={open} onOpenChange={handleOpenChange}>
        <DialogContent className="max-h-[80vh]">
          <DialogHeader>
            <DialogTitle>{title}</DialogTitle>
            <DialogDescription className="sr-only">{title}</DialogDescription>
          </DialogHeader>
          {content}
        </DialogContent>
      </Dialog>
    )
  }

  return (
    <Drawer open={open} onOpenChange={handleOpenChange}>
      <DrawerContent>
        <DrawerHeader>
          <DrawerTitle>{title}</DrawerTitle>
          <DrawerDescription className="sr-only">{title}</DrawerDescription>
        </DrawerHeader>
        {content}
      </DrawerContent>
    </Drawer>
  )
}

// --- Sub-components ---

function ClusterEventList({ events, locale, onSelectEvent }: ClusterEventListProps): JSX.Element {
  return (
    <div className="flex min-h-0 flex-1 flex-col overflow-y-auto px-4 pb-6">
      {events.map((event) => (
        <EventRow key={event.id} event={event} locale={locale} onSelect={() => onSelectEvent(event)} />
      ))}
    </div>
  )
}

function EventRow({ event, locale, onSelect }: EventRowProps): JSX.Element {
  return (
    <button
      type="button"
      onClick={onSelect}
      className="flex cursor-pointer flex-col gap-0.5 rounded border-b border-border px-2 py-3 text-left transition-colors last:border-0 hover:bg-muted/50"
    >
      <span className="text-sm font-medium text-foreground">{event.name}</span>
      <span className="text-xs text-muted-foreground">
        {formatDateRange(event.start_date, event.end_date, locale)}
      </span>
    </button>
  )
}

// --- Export ---

export { ClusterEventDrawer }
export type { ClusterEventDrawerProps }
