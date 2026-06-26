"use client"

import type { JSX } from "react"
import { useTranslations, useLocale } from "next-intl"

import { useSessionGuard } from "@/hooks/useSessionGuard"
import { UserListPage } from "@/components/manage/users/UserListPage"

// --- Component ---

export default function AdminUsersPage(): JSX.Element {
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

  return (
    <main className="min-h-screen bg-background">
      <UserListPage locale={locale} />
    </main>
  )
}
