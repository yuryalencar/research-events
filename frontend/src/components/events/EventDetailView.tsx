"use client"

import type { JSX } from "react"
import { useTranslations, useLocale } from "next-intl"
import { useRouter } from "next/navigation"

import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetDescription } from "@/components/ui/sheet"
import { Drawer, DrawerContent, DrawerHeader, DrawerTitle, DrawerDescription } from "@/components/ui/drawer"
import { EventDetailContent } from "@/components/events/EventDetailContent"
import { useMediaQuery } from "@/hooks/useMediaQuery"
import type { EventListItem } from "@/types/api"

// --- Types & Props ---

interface EventDetailViewProps {
  event: EventListItem | null
  onClose: () => void
}

// --- Constants ---

// Matches Tailwind's `md` breakpoint — side panel at md and above, bottom
// card below md (see specs/frontend/globe-homepage.md — "Responsive layout").
const DESKTOP_QUERY = "(min-width: 768px)"

const SESSION_KEY = "deadline_management_event"

// --- Component ---

// EventDetailView shows the selected event's details — as a side panel
// sliding from the right on desktop, or a bottom card on mobile. Both
// layouts share the same content (EventDetailContent) and close
// interaction.
function EventDetailView({ event, onClose }: EventDetailViewProps): JSX.Element | null {
  const t = useTranslations("eventDetail")
  const locale = useLocale()
  const router = useRouter()
  const isDesktop = useMediaQuery(DESKTOP_QUERY)

  if (event === null) {
    return null
  }

  const handleOpenChange = (open: boolean): void => {
    if (!open) onClose()
  }

  const handleManageDeadlines = (ev: EventListItem): void => {
    sessionStorage.setItem(SESSION_KEY, JSON.stringify(ev))
    router.push(`/${locale}/events/${ev.slug}/deadlines`)
  }

  if (isDesktop) {
    return (
      <Sheet open onOpenChange={handleOpenChange}>
        <SheetContent side="right">
          <SheetHeader>
            <SheetTitle className="pr-8">
              {event.name} <span className="text-muted-foreground font-normal">({event.slug})</span>
            </SheetTitle>
            <SheetDescription className="sr-only">{t("detailsDescription")}</SheetDescription>
          </SheetHeader>
          <EventDetailContent event={event} onManageDeadlines={handleManageDeadlines} />
        </SheetContent>
      </Sheet>
    )
  }

  return (
    <Drawer open onOpenChange={handleOpenChange}>
      <DrawerContent>
        <DrawerHeader>
          <DrawerTitle className="pr-8">
            {event.name} <span className="text-muted-foreground font-normal">({event.slug})</span>
          </DrawerTitle>
          <DrawerDescription className="sr-only">{t("detailsDescription")}</DrawerDescription>
        </DrawerHeader>
        <EventDetailContent event={event} onManageDeadlines={handleManageDeadlines} />
      </DrawerContent>
    </Drawer>
  )
}

// --- Export ---

export { EventDetailView }
export type { EventDetailViewProps }
