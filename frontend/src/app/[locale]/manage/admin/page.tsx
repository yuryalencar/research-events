"use client"

import type { JSX } from "react"
import { useTranslations } from "next-intl"

import { ManageDashboard } from "@/components/manage/ManageDashboard"
import { useSessionGuard } from "@/hooks/useSessionGuard"

// --- Component ---

export default function AdminDashboardPage(): JSX.Element {
  const t = useTranslations("manage")
  const { user } = useSessionGuard("admin")

  if (!user) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-background">
        <p className="text-sm text-muted-foreground">{t("dashboard.loading")}</p>
      </main>
    )
  }

  return <ManageDashboard user={user} />
}
