"use client"

import type { JSX } from "react"
import { TableIcon, GlobeIcon } from "lucide-react"
import { useTranslations } from "next-intl"

import type { ViewMode } from "@/hooks/useViewMode"

// --- Types ---

interface ViewToggleButtonProps {
  viewMode: ViewMode
  onToggle: () => void
}

// --- Component ---

// ViewToggleButton is the fixed floating button that switches between the globe
// and card-list table views. It is desktop-only — the parent renders it inside
// a `hidden md:flex` wrapper so it disappears on resize without waiting for JS.
// Positioning mirrors AddEventButton (fixed top-right, same z-index), placed
// directly below it via top-16.
function ViewToggleButton({ viewMode, onToggle }: ViewToggleButtonProps): JSX.Element {
  const t = useTranslations("viewToggle")

  const isGlobe = viewMode === "globe"
  const tooltip = isGlobe ? t("tableModeTooltip") : t("globeModeTooltip")

  return (
    <div className="fixed top-16 right-4 z-40 hidden md:flex group items-center">
      {/* Tooltip — appears to the left of the button on hover */}
      <span
        aria-hidden="true"
        className="pointer-events-none mr-2 whitespace-nowrap rounded-md border border-border bg-background/90 px-2.5 py-1 text-xs text-foreground shadow-md backdrop-blur-sm opacity-0 transition-opacity duration-150 group-hover:opacity-100"
      >
        {tooltip}
      </span>

      <button
        type="button"
        aria-label={tooltip}
        onClick={onToggle}
        className="flex size-10 items-center justify-center rounded-full border border-border bg-background/80 text-foreground shadow-lg backdrop-blur-sm transition-colors hover:bg-background"
      >
        {isGlobe ? <TableIcon className="size-5" /> : <GlobeIcon className="size-5" />}
      </button>
    </div>
  )
}

// --- Export ---

export { ViewToggleButton }
export type { ViewToggleButtonProps }
