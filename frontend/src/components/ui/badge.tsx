import type { ComponentProps, JSX } from "react"

import { cn } from "@/lib/utils"

// --- Types ---

interface BadgeProps extends ComponentProps<"span"> {
  variant?: "default" | "secondary" | "outline"
}

// --- Component ---
// Badge is used to display an event's tier (e.g. "A*", "A", "B", "C") in the
// detail view — never rendered when tier === "unranked" (see
// specs/frontend/globe-homepage.md).

function Badge({ className, variant = "default", ...props }: BadgeProps): JSX.Element {
  return (
    <span
      data-slot="badge"
      className={cn(
        "inline-flex items-center justify-center rounded-md border px-2 py-0.5 text-xs font-medium w-fit whitespace-nowrap shrink-0",
        variant === "default" && "border-transparent bg-primary text-primary-foreground",
        variant === "secondary" && "border-transparent bg-secondary text-secondary-foreground",
        variant === "outline" && "text-foreground",
        className
      )}
      {...props}
    />
  )
}

// --- Export ---

export { Badge }
export type { BadgeProps }
