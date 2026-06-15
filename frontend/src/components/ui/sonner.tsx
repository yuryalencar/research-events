import type { CSSProperties, JSX } from "react"
import { Toaster as SonnerToaster, type ToasterProps } from "sonner"

// --- Component ---

// Toaster wires sonner's toast container into the shadcn/ui theme tokens
// defined in globals.css, so toasts match the app's neutral palette.
// Theme is fixed to "light" for now — no dark mode toggle exists yet.
function Toaster(props: ToasterProps): JSX.Element {
  return (
    <SonnerToaster
      theme="light"
      className="toaster group"
      style={
        {
          "--normal-bg": "var(--popover)",
          "--normal-text": "var(--popover-foreground)",
          "--normal-border": "var(--border)",
        } as CSSProperties
      }
      {...props}
    />
  )
}

// --- Export ---

export { Toaster }
