import { useSyncExternalStore } from "react"

// --- Hook ---

// useMediaQuery tracks whether a CSS media query currently matches, updating
// reactively as the viewport is resized. Used to switch the detail view
// between the desktop side panel and the mobile bottom card (see
// specs/frontend/globe-homepage.md — "Responsive layout").
//
// Built on useSyncExternalStore rather than useState+useEffect: matchMedia
// is an external store (the browser's viewport state), and
// useSyncExternalStore is React's purpose-built primitive for subscribing to
// such stores without the cascading-render issues of calling setState inside
// an effect.
function useMediaQuery(query: string): boolean {
  return useSyncExternalStore(
    (onChange) => {
      const mediaQueryList = window.matchMedia(query)
      mediaQueryList.addEventListener("change", onChange)
      return () => mediaQueryList.removeEventListener("change", onChange)
    },
    () => window.matchMedia(query).matches,
    () => false
  )
}

// --- Export ---

export { useMediaQuery }
