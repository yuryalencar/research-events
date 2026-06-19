"use client"

import type { JSX } from "react"
import { useState, useEffect } from "react"
import { useRouter } from "next/navigation"
import { useLocale, useTranslations } from "next-intl"

import { ManageDashboard } from "@/components/manage/ManageDashboard"
import type { SessionUser } from "@/components/manage/ManageDashboard"

// --- Component ---

export default function ModeratorDashboardPage(): JSX.Element {
  const t = useTranslations("manage")
  const locale = useLocale()
  const router = useRouter()

  const [user, setUser] = useState<SessionUser | null>(null)

  useEffect(() => {
    try {
      const stored = localStorage.getItem("manage_user")
      if (!stored) {
        router.replace(`/${locale}/manage`)
        return
      }
      const parsed = JSON.parse(stored) as SessionUser
      if (parsed.role !== "moderator") {
        router.replace(`/${locale}/manage/${parsed.role}`)
        return
      }
      setUser(parsed)
    } catch {
      localStorage.removeItem("manage_user")
      router.replace(`/${locale}/manage`)
    }
  }, [locale, router])

  if (!user) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-background">
        <p className="text-sm text-muted-foreground">{t("dashboard.loading")}</p>
      </main>
    )
  }

  return <ManageDashboard user={user} />
}
