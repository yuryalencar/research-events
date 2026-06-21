"use client"

import type { JSX } from "react"
import { useLocale, useTranslations } from "next-intl"

import { UpdatePasswordCard } from "@/components/manage/UpdatePasswordCard"
import { useSessionGuard } from "@/hooks/useSessionGuard"

// --- Component ---

export default function AdminUpdatePasswordPage(): JSX.Element {
  const t = useTranslations("manage")
  const locale = useLocale()
  const { user } = useSessionGuard("admin")

  if (!user) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-background">
        <p className="text-sm text-muted-foreground">{t("dashboard.loading")}</p>
      </main>
    )
  }

  return <UpdatePasswordCard dashboardHref={`/${locale}/manage/admin`} />
}
