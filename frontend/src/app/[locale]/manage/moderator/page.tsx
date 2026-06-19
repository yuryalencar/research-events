"use client"

import type { JSX } from "react"
import { useState, useEffect } from "react"
import { useRouter } from "next/navigation"
import { useLocale, useTranslations } from "next-intl"

import { logout } from "@/lib/api/auth"
import { Badge } from "@/components/ui/badge"

// --- Types ---

interface SessionUser {
  name: string
  role: "admin" | "moderator"
  email: string
}

// --- Component ---

export default function ModeratorDashboardPage(): JSX.Element {
  const t = useTranslations("manage")
  const locale = useLocale()
  const router = useRouter()

  const [user, setUser] = useState<SessionUser | null>(null)

  // Read the stored session from localStorage synchronously after mount.
  // No API call here — token refresh only happens when an authenticated
  // request returns TOKEN_EXPIRED (handled inside apiPrivateRequest).
  useEffect(() => {
    try {
      const stored = localStorage.getItem("manage_user")
      if (!stored) {
        router.replace(`/${locale}/manage`)
        return
      }
      const parsed = JSON.parse(stored) as SessionUser
      if (parsed.role !== "moderator") {
        // Role mismatch — send the user to their correct dashboard.
        router.replace(`/${locale}/manage/${parsed.role}`)
        return
      }
      setUser(parsed)
    } catch {
      localStorage.removeItem("manage_user")
      router.replace(`/${locale}/manage`)
    }
  }, [locale, router])

  async function handleLogout(): Promise<void> {
    try {
      await logout()
    } finally {
      // Always clear local state and redirect — even if the API call fails
      // the session should be considered ended from the client's perspective.
      localStorage.removeItem("manage_user")
      router.replace(`/${locale}/manage`)
    }
  }

  if (!user) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-background">
        <p className="text-sm text-muted-foreground">{t("dashboard.loading")}</p>
      </main>
    )
  }

  return (
    <main className="flex min-h-screen flex-col items-center justify-center gap-2 bg-background p-4">
      <h1 className="text-3xl font-semibold text-foreground">
        {t("dashboard.welcomeHeading", { name: user.name })}
      </h1>
      <Badge variant="secondary" className="capitalize">
        {user.role}
      </Badge>
      <button
        onClick={handleLogout}
        className="mt-6 rounded-md border border-border px-4 py-2 text-sm font-medium text-foreground transition-colors hover:bg-muted"
      >
        {t("dashboard.logoutButton")}
      </button>
    </main>
  )
}
