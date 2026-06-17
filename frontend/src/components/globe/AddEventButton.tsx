"use client"

import type { JSX } from "react"
import Link from "next/link"
import { PlusIcon } from "lucide-react"
import { useLocale, useTranslations } from "next-intl"

// --- Component ---

// AddEventButton is the fixed floating "+" button at the top-right of the
// globe homepage. It navigates to the event submission wizard (full page).
// A native title tooltip is shown on desktop hover; aria-label covers
// screen readers and mobile (where hover is unavailable).
function AddEventButton(): JSX.Element {
  const t = useTranslations("submit")
  const locale = useLocale()

  return (
    <div className="fixed top-4 right-4 z-40 group flex items-center">
      {/* Tooltip — appears to the left of the button on hover */}
      <span
        aria-hidden="true"
        className="pointer-events-none mr-2 whitespace-nowrap rounded-md border border-border bg-background/90 px-2.5 py-1 text-xs text-foreground shadow-md backdrop-blur-sm opacity-0 transition-opacity duration-150 group-hover:opacity-100"
      >
        {t("addEventTooltip")}
      </span>

      <Link
        href={`/${locale}/events/submit`}
        aria-label={t("addEventTooltip")}
        className="flex size-10 items-center justify-center rounded-full border border-border bg-background/80 text-foreground shadow-lg backdrop-blur-sm transition-colors hover:bg-background"
      >
        <PlusIcon className="size-5" />
      </Link>
    </div>
  )
}

// --- Export ---

export { AddEventButton }
