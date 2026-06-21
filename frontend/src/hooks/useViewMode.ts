import { useState, useEffect, useRef, useCallback } from "react"

import { useMediaQuery } from "@/hooks/useMediaQuery"

// --- Types ---

type ViewMode = "globe" | "table"

interface UseViewModeReturn {
  viewMode: ViewMode
  toggleView: () => void
}

// --- Constants ---

const DESKTOP_QUERY = "(min-width: 768px)"

// --- Hook ---

// useViewMode manages the active view (globe vs table) and enforces the rule
// that the table is desktop-only. When the viewport shrinks below the md
// breakpoint while the table is active, the hook switches back to globe and
// calls onForcedGlobe so the caller can show a toast.
//
// onForcedGlobe is intentionally kept outside this hook — the hook computes
// the transition; the caller decides how to notify the user. This separates
// the state machine from the side effect (toast).
function useViewMode(onForcedGlobe: () => void): UseViewModeReturn {
  const [viewMode, setViewMode] = useState<ViewMode>("globe")
  const isDesktop = useMediaQuery(DESKTOP_QUERY)

  // Track the previous desktop value so we only react to a true→false crossing.
  // A ref avoids adding prevIsDesktop to the effect dependency array, which
  // would cause extra re-runs whenever the ref is updated.
  const prevIsDesktopRef = useRef(isDesktop)

  // Stable ref for the callback so the effect dep array never changes even
  // if the caller passes a new function reference on each render.
  const onForcedGlobeRef = useRef(onForcedGlobe)
  onForcedGlobeRef.current = onForcedGlobe

  useEffect(() => {
    const wasDesktop = prevIsDesktopRef.current
    prevIsDesktopRef.current = isDesktop

    // Only act on a desktop → mobile crossing (true → false).
    // Crossing in the other direction (false → true) requires no action.
    if (wasDesktop && !isDesktop && viewMode === "table") {
      setViewMode("globe")
      onForcedGlobeRef.current()
    }
  }, [isDesktop, viewMode])

  const toggleView = useCallback(() => {
    setViewMode((prev) => (prev === "globe" ? "table" : "globe"))
  }, [])

  return { viewMode, toggleView }
}

// --- Export ---

export { useViewMode }
export type { ViewMode, UseViewModeReturn }
