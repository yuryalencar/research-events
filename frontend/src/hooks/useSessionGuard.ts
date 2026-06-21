"use client"

import { useState, useEffect } from "react"
import { useRouter } from "next/navigation"
import { useLocale } from "next-intl"

import { validateSession } from "@/lib/api/users"
import type { SessionUser } from "@/lib/api/users"
import { ApiError } from "@/lib/api/client"

// --- Constants ---

// Codes that confirm the session is fully gone — both access and refresh tokens
// are invalid. Any of these after apiPrivateRequest's automatic refresh attempt
// means the user must re-authenticate.
const AUTH_ERROR_CODES = new Set(["TOKEN_MISSING", "TOKEN_EXPIRED", "TOKEN_INVALID"])

// --- Types ---

interface UseSessionGuardReturn {
  user: SessionUser | null
}

// --- Hook ---

// useSessionGuard is the shared session check for every protected /manage page.
//
// Step 1 — synchronous localStorage check: if no user is stored or the stored
//   role doesn't match requiredRole, redirect immediately (no flash).
// Step 2 — set user from localStorage for instant display (avoids loading flash).
// Step 3 — background token validation: call GET /api/v1/users/me. If the token
//   is expired, apiPrivateRequest attempts a refresh automatically. If both
//   tokens are dead, an auth error is thrown and the user is redirected to login.
//
// This means the UI is always responsive (step 2) while a stale-token redirect
// happens silently in the background (step 3) — no blocking spinner needed.
function useSessionGuard(requiredRole: string): UseSessionGuardReturn {
  const locale = useLocale()
  const router = useRouter()
  const [user, setUser] = useState<SessionUser | null>(null)

  useEffect(() => {
    let cancelled = false

    function redirectToLogin(): void {
      localStorage.removeItem("manage_user")
      router.replace(`/${locale}/manage`)
    }

    try {
      const stored = localStorage.getItem("manage_user")
      if (!stored) {
        router.replace(`/${locale}/manage`)
        return
      }

      const parsed = JSON.parse(stored) as SessionUser
      if (parsed.role !== requiredRole) {
        // Redirect to the correct dashboard for this role rather than to login.
        router.replace(`/${locale}/manage/${parsed.role}`)
        return
      }

      // Trust localStorage immediately — gives instant display without a spinner.
      setUser(parsed)

      // Background: validate the token. apiPrivateRequest already retries once
      // with the refresh token on TOKEN_EXPIRED. If that also fails, we get an
      // auth error here and redirect to login.
      void validateSession().catch((err: unknown) => {
        if (cancelled) return
        if (err instanceof ApiError && AUTH_ERROR_CODES.has(err.code)) {
          redirectToLogin()
        }
      })
    } catch {
      redirectToLogin()
    }

    return () => {
      cancelled = true
    }
  }, [locale, router, requiredRole])

  return { user }
}

// --- Export ---

export { useSessionGuard }
export type { UseSessionGuardReturn }
