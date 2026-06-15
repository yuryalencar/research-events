import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

// --- Utilities ---

// cn merges Tailwind class lists, resolving conflicting utility classes
// (e.g. "p-2" vs "p-4") in favor of the last one — the standard shadcn/ui helper.
function cn(...inputs: ClassValue[]): string {
  return twMerge(clsx(inputs))
}

// --- Export ---

export { cn }
