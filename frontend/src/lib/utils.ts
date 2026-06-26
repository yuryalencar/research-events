import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

// --- Types ---

interface PasswordComplexity {
  minLength: boolean
  hasUppercase: boolean
  hasLowercase: boolean
  hasSpecial: boolean
}

// --- Utilities ---

// cn merges Tailwind class lists, resolving conflicting utility classes
// (e.g. "p-2" vs "p-4") in favor of the last one — the standard shadcn/ui helper.
function cn(...inputs: ClassValue[]): string {
  return twMerge(clsx(inputs))
}

// checkPasswordComplexity evaluates a password against the four rules enforced
// by all password endpoints. Returns a breakdown so UIs can show per-rule feedback.
function checkPasswordComplexity(password: string): PasswordComplexity {
  return {
    minLength: password.length >= 8,
    hasUppercase: /[A-Z]/.test(password),
    hasLowercase: /[a-z]/.test(password),
    hasSpecial: /[^A-Za-z0-9]/.test(password),
  }
}

// --- Export ---

export { cn, checkPasswordComplexity }
export type { PasswordComplexity }
