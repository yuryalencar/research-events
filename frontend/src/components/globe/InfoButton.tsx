"use client"

import { useState } from "react"
import type { JSX } from "react"
import { InfoIcon } from "lucide-react"
import { useTranslations } from "next-intl"

import { InfoModal } from "@/components/globe/InfoModal"
import { useMediaQuery } from "@/hooks/useMediaQuery"

// --- Types ---

interface InfoButtonProps {
  // true when the event-detail Drawer is open on mobile — hides this button
  // so it doesn't compete with the Drawer for bottom-of-screen real estate.
  drawerOpen: boolean
}

// --- Constants ---

// Matches Tailwind's `md` breakpoint — same value used in EventDetailView.
const DESKTOP_QUERY = "(min-width: 768px)"

// --- Component ---

// InfoButton is a fixed floating button at the bottom-left of the globe
// homepage. It owns the open/close state for InfoModal and hides itself on
// mobile when the event-detail Drawer is already open.
function InfoButton({ drawerOpen }: InfoButtonProps): JSX.Element | null {
  const [open, setOpen] = useState(false)
  const t = useTranslations("info")
  const isDesktop = useMediaQuery(DESKTOP_QUERY)

  if (!isDesktop && drawerOpen) {
    return null
  }

  return (
    <>
      <button
        type="button"
        aria-label={t("buttonLabel")}
        onClick={() => setOpen(true)}
        className="fixed bottom-4 left-4 z-40 flex size-10 items-center justify-center rounded-full border border-border bg-background/80 text-foreground shadow-lg backdrop-blur-sm transition-colors hover:bg-background"
      >
        <InfoIcon className="size-5" />
      </button>

      <InfoModal open={open} onOpenChange={setOpen} isDesktop={isDesktop} />
    </>
  )
}

// --- Export ---

export { InfoButton }
export type { InfoButtonProps }
